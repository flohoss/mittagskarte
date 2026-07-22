package snapotter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/flohoss/mittagskarte/pkg/snapotter/api"

	_ "golang.org/x/image/webp"

	ht "github.com/ogen-go/ogen/http"
)

type Result struct {
	Data   []byte
	Name   string
	Width  int
	Height int
}

func newResult(data []byte, name string) (Result, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Result{}, fmt.Errorf("decode image dimensions: %w", err)
	}
	return Result{
		Data:   data,
		Name:   name,
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

type Client struct {
	api    *api.Client
	logger *slog.Logger
}

type noAuth struct{}

func (noAuth) BearerAuth(ctx context.Context, operationName api.OperationName) (api.BearerAuth, error) {
	return api.BearerAuth{}, nil
}

func New(u url.URL, appLogger *slog.Logger) *Client {
	apiClient, err := api.NewClient(u.String(), noAuth{})
	if err != nil {
		return nil
	}
	return &Client{
		api:    apiClient,
		logger: appLogger.WithGroup("snapotter"),
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

func (c *Client) downloadBytes(ctx context.Context, jobId, filename string) ([]byte, error) {
	res, err := c.api.DownloadProcessedImage(ctx, api.DownloadProcessedImageParams{
		JobId:    jobId,
		Filename: filename,
	})
	if err != nil {
		return nil, fmt.Errorf("download request: %w", err)
	}

	r, ok := res.(*api.DownloadProcessedImageOKHeaders)
	if !ok {
		return nil, fmt.Errorf("download returned unexpected response: %T", res)
	}

	data, err := io.ReadAll(r.Response.Data)
	if err != nil {
		return nil, fmt.Errorf("read download response: %w", err)
	}
	return data, nil
}

func (c *Client) downloadJob(ctx context.Context, jobId, downloadUrl string) (Result, error) {
	filename := filepath.Base(downloadUrl)
	data, err := c.downloadBytes(ctx, jobId, filename)
	if err != nil {
		return Result{}, err
	}
	return newResult(data, filename)
}

func (c *Client) downloadFromTool(ctx context.Context, toolResp *api.ToolResponse) (Result, error) {
	jobId, ok := toolResp.GetJobId().Get()
	if !ok {
		return Result{}, fmt.Errorf("response returned no job id")
	}
	downloadUrl, ok := toolResp.GetDownloadUrl().Get()
	if !ok {
		return Result{}, fmt.Errorf("response returned no download url")
	}
	return c.downloadJob(ctx, jobId, downloadUrl)
}

func (c *Client) Setup() error {
	c.logger.Debug("Setting up snapotter client")
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

func (c *Client) ImageToWebp(ctx context.Context, sourcePath string) (Result, error) {
	file, cleanup, err := c.multipartFile(sourcePath)
	if err != nil {
		return Result{}, err
	}
	defer cleanup()

	pipeline, _ := json.Marshal(map[string]any{
		"steps": []map[string]any{
			{"toolId": "smart-crop", "settings": map[string]any{"mode": "trim", "threshold": 10}},
			{"toolId": "optimize-for-web", "settings": map[string]any{"format": "webp", "quality": 85, "maxWidth": 1920}},
		},
	})

	res, err := c.api.ExecutePipeline(ctx, &api.ExecutePipelineReq{
		File:     file,
		Pipeline: string(pipeline),
	})
	if err != nil {
		return Result{}, fmt.Errorf("pipeline request: %w", err)
	}
	c.logger.Debug("Image pipeline completed, downloading result", "sourcePath", sourcePath)

	okResp, ok := res.(*api.ExecutePipelineOK)
	if !ok {
		return Result{}, fmt.Errorf("pipeline returned unexpected response: %T", res)
	}

	jobId, ok := okResp.GetJobId().Get()
	if !ok {
		return Result{}, fmt.Errorf("pipeline returned no job id")
	}
	downloadUrl, ok := okResp.GetDownloadUrl().Get()
	if !ok {
		return Result{}, fmt.Errorf("pipeline returned no download url")
	}

	return c.downloadJob(ctx, jobId, downloadUrl)
}

func (c *Client) PDFToPngPages(ctx context.Context, inputPath, outputDir string, dpi int) ([]string, error) {
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
		"dpi":    dpi,
	})

	res, err := c.api.PdfToImage(ctx, &api.PdfToImageReq{
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
	c.logger.Debug("PDF conversion completed, downloading pages", "inputPath", inputPath, "pageCount", len(pages))
	pagePaths := make([]string, 0, len(pages))
	for i, page := range pages {
		downloadUrl, ok := page.GetDownloadUrl().Get()
		if !ok {
			return nil, fmt.Errorf("page %d returned no download url", i)
		}
		pagePath := filepath.Join(outputDir, fmt.Sprintf("page_%03d.png", i+1))
		data, err := c.downloadBytes(ctx, jobId, filepath.Base(downloadUrl))
		if err != nil {
			return nil, fmt.Errorf("download page %d: %w", i+1, err)
		}
		if err := os.WriteFile(pagePath, data, 0o644); err != nil {
			return nil, fmt.Errorf("write page %d: %w", i+1, err)
		}
		pagePaths = append(pagePaths, pagePath)
	}

	return pagePaths, nil
}

func (c *Client) StitchImagesVertical(ctx context.Context, pagePaths []string, outputPath string) error {
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

	res, err := c.api.StitchImages(ctx, &api.StitchImagesReq{
		File:     files,
		Settings: api.NewOptString(string(settings)),
	})
	if err != nil {
		return fmt.Errorf("stitch images request: %w", err)
	}
	c.logger.Debug("Stitch request completed, downloading result", "pageCount", len(pagePaths))

	toolResp, ok := res.(*api.ToolResponse)
	if !ok {
		return fmt.Errorf("stitch images returned unexpected response: %T", res)
	}

	result, err := c.downloadFromTool(ctx, toolResp)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, result.Data, 0o644)
}
