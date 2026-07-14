package snapotter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flohoss/mittagskarte/pkg/snapotter/api"

	ht "github.com/ogen-go/ogen/http"
)

type Client struct {
	api *api.Client
}

type noAuth struct{}

func (noAuth) BearerAuth(ctx context.Context, operationName api.OperationName) (api.BearerAuth, error) {
	return api.BearerAuth{}, nil
}

func New(u url.URL) *Client {
	apiClient, err := api.NewClient(u.String(), noAuth{})
	if err != nil {
		return nil
	}
	return &Client{
		api: apiClient,
	}
}

func (c *Client) multipartFile(path string) (ht.MultipartFile, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return ht.MultipartFile{}, nil, fmt.Errorf("open file: %w", err)
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return ht.MultipartFile{}, nil, fmt.Errorf("stat file: %w", err)
	}
	return ht.MultipartFile{
		Name: filepath.Base(path),
		File: f,
		Size: stat.Size(),
	}, func() { f.Close() }, nil
}

func (c *Client) download(jobId, downloadUrl, outputPath string) error {
	res, err := c.api.DownloadProcessedImage(context.Background(), api.DownloadProcessedImageParams{
		JobId:    jobId,
		Filename: filepath.Base(downloadUrl),
	})
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}

	r, ok := res.(*api.DownloadProcessedImageOKHeaders)
	if !ok {
		return fmt.Errorf("download returned unexpected response: %T", res)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, r.Response.Data); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}

func (c *Client) downloadFromTool(toolResp *api.ToolResponse, outputPath string) error {
	jobId, ok := toolResp.GetJobId().Get()
	if !ok {
		return fmt.Errorf("response returned no job id")
	}
	downloadUrl, ok := toolResp.GetDownloadUrl().Get()
	if !ok {
		return fmt.Errorf("response returned no download url")
	}
	return c.download(jobId, downloadUrl, outputPath)
}

func (c *Client) Setup() error {
	res, err := c.api.ListFeatures(context.Background())
	if err != nil {
		return fmt.Errorf("list features: %w", err)
	}

	okResp, ok := res.(*api.ListFeaturesOK)
	if !ok {
		return fmt.Errorf("list features returned unexpected response: %T", res)
	}

	for _, bundle := range okResp.GetBundles() {
		id, _ := bundle.GetID().Get()
		if id != "face-detection" {
			continue
		}
		status, _ := bundle.GetStatus().Get()
		if status == api.ListFeaturesOKBundlesItemStatusInstalled {
			return nil
		}

		installRes, err := c.api.InstallFeature(context.Background(), api.InstallFeatureParams{
			BundleId: id,
		})
		if err != nil {
			return fmt.Errorf("install feature %s: %w", id, err)
		}
		if _, ok := installRes.(*api.InstallFeatureAccepted); !ok {
			return fmt.Errorf("install feature %s returned unexpected response: %T", id, installRes)
		}

		for range 60 {
			time.Sleep(5 * time.Second)
			checkRes, err := c.api.ListFeatures(context.Background())
			if err != nil {
				continue
			}
			checkOk, ok := checkRes.(*api.ListFeaturesOK)
			if !ok {
				continue
			}
			for _, b := range checkOk.GetBundles() {
				bid, _ := b.GetID().Get()
				if bid != id {
					continue
				}
				bs, _ := b.GetStatus().Get()
				if bs == api.ListFeaturesOKBundlesItemStatusInstalled {
					return nil
				}
				if bs == api.ListFeaturesOKBundlesItemStatusNotInstalled {
					return fmt.Errorf("feature %s installation failed", id)
				}
				break
			}
		}
		return fmt.Errorf("feature %s installation timed out", id)
	}

	return nil
}

func (c *Client) ProcessFileToWebp(sourcePath, outputPath string) error {
	if isPDFFile(sourcePath) {
		return c.pdfToWebp(sourcePath, outputPath)
	}
	return c.imageToWebp(sourcePath, outputPath)
}

func (c *Client) imageToWebp(sourcePath, outputPath string) error {
	file, cleanup, err := c.multipartFile(sourcePath)
	if err != nil {
		return err
	}
	defer cleanup()

	pipeline, _ := json.Marshal(map[string]any{
		"steps": []map[string]any{
			{"toolId": "smart-crop", "settings": map[string]any{"mode": "trim", "threshold": 10}},
			{"toolId": "optimize-for-web", "settings": map[string]any{"format": "webp", "quality": 85, "maxWidth": 1920}},
		},
	})

	res, err := c.api.ExecutePipeline(context.Background(), &api.ExecutePipelineReq{
		File:     file,
		Pipeline: string(pipeline),
	})
	if err != nil {
		return fmt.Errorf("pipeline request: %w", err)
	}

	okResp, ok := res.(*api.ExecutePipelineOK)
	if !ok {
		return fmt.Errorf("pipeline returned unexpected response: %T", res)
	}

	jobId, ok := okResp.GetJobId().Get()
	if !ok {
		return fmt.Errorf("pipeline returned no job id")
	}
	downloadUrl, ok := okResp.GetDownloadUrl().Get()
	if !ok {
		return fmt.Errorf("pipeline returned no download url")
	}

	return c.download(jobId, downloadUrl, outputPath)
}

func (c *Client) pdfToWebp(inputPath, outputPath string) error {
	tmpDir, err := os.MkdirTemp("", "pdf2webp-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pagePaths, err := c.PDFToPngPages(inputPath, tmpDir)
	if err != nil {
		return fmt.Errorf("convert pdf to images: %w", err)
	}

	if len(pagePaths) == 1 {
		return c.imageToWebp(pagePaths[0], outputPath)
	}

	stitchedPath := filepath.Join(tmpDir, "stitched.png")
	if err := c.StitchImagesVertical(pagePaths, stitchedPath); err != nil {
		return fmt.Errorf("stitch pdf pages: %w", err)
	}
	return c.imageToWebp(stitchedPath, outputPath)
}

func isPDFFile(sourcePath string) bool {
	if strings.EqualFold(filepath.Ext(sourcePath), ".pdf") {
		return true
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 512)
	readBytes, err := file.Read(header)
	if err != nil || readBytes == 0 {
		return false
	}

	return http.DetectContentType(header[:readBytes]) == "application/pdf"
}

func (c *Client) PDFToPngPages(inputPath, outputDir string) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	file, cleanup, err := c.multipartFile(inputPath)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	settings, _ := json.Marshal(map[string]any{
		"format": "png",
		"dpi":    300,
	})

	res, err := c.api.PdfToImage(context.Background(), &api.PdfToImageReq{
		File:     file,
		Settings: api.NewOptString(string(settings)),
	})
	if err != nil {
		return nil, fmt.Errorf("pdf to image request: %w", err)
	}

	okResp, ok := res.(*api.PdfToImageOK)
	if !ok {
		return nil, fmt.Errorf("pdf to image returned unexpected response: %T", res)
	}

	jobId, ok := okResp.GetJobId().Get()
	if !ok {
		return nil, fmt.Errorf("pdf to image returned no job id")
	}

	pages := okResp.GetPages()
	pagePaths := make([]string, 0, len(pages))
	for i, page := range pages {
		downloadUrl, ok := page.GetDownloadUrl().Get()
		if !ok {
			return nil, fmt.Errorf("page %d returned no download url", i)
		}
		pagePath := filepath.Join(outputDir, fmt.Sprintf("page_%03d.png", i+1))
		if err := c.download(jobId, downloadUrl, pagePath); err != nil {
			return nil, fmt.Errorf("download page %d: %w", i+1, err)
		}
		pagePaths = append(pagePaths, pagePath)
	}

	return pagePaths, nil
}

func (c *Client) StitchImagesVertical(pagePaths []string, outputPath string) error {
	files := make([]ht.MultipartFile, 0, len(pagePaths))
	cleanups := make([]func(), 0, len(pagePaths))
	defer func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	}()

	for _, pagePath := range pagePaths {
		file, cleanup, err := c.multipartFile(pagePath)
		if err != nil {
			return err
		}
		files = append(files, file)
		cleanups = append(cleanups, cleanup)
	}

	settings, _ := json.Marshal(map[string]any{
		"direction": "vertical",
		"format":    "png",
	})

	res, err := c.api.StitchImages(context.Background(), &api.StitchImagesReq{
		File:     files,
		Settings: api.NewOptString(string(settings)),
	})
	if err != nil {
		return fmt.Errorf("stitch images request: %w", err)
	}

	toolResp, ok := res.(*api.ToolResponse)
	if !ok {
		return fmt.Errorf("stitch images returned unexpected response: %T", res)
	}

	return c.downloadFromTool(toolResp, outputPath)
}
