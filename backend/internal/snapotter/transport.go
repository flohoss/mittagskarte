package snapotter

import (
	"context"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/ogen-go/ogen/validate"
)

const (
	RequestTimeout    = 5 * time.Minute
	MaxAttempts       = 3
	InitialRetryDelay = 500 * time.Millisecond
	MaxRetryDelay     = 8 * time.Second
)

type retryTransport struct {
	client    *http.Client
	attempts  int
	baseDelay time.Duration
	maxDelay  time.Duration
	sleep     func(context.Context, time.Duration) error
}

func newRetryTransport(client *http.Client) *retryTransport {
	return &retryTransport{
		client:    client,
		attempts:  MaxAttempts,
		baseDelay: InitialRetryDelay,
		maxDelay:  MaxRetryDelay,
		sleep:     sleepWithContext,
	}
}

func (t *retryTransport) Do(req *http.Request) (*http.Response, error) {
	spool, err := spoolRequest(req)
	if err != nil {
		return nil, err
	}
	defer spool.cleanup()

	var (
		lastErr    error
		lastStatus int
	)
	for attempt := 1; attempt <= t.attempts; attempt++ {
		if attempt > 1 {
			if err := t.sleep(req.Context(), t.backoff(attempt-1)); err != nil {
				return nil, err
			}
		}

		res, err := t.attempt(req)
		if err != nil {
			lastErr, lastStatus = err, 0
			continue
		}
		if !retryableStatus(res.StatusCode) {
			return res, nil
		}

		lastErr, lastStatus = nil, res.StatusCode
		drainAndClose(res)
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, validate.UnexpectedStatusCode(lastStatus)
}

func (t *retryTransport) attempt(req *http.Request) (*http.Response, error) {
	attemptReq := req.Clone(req.Context())
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		attemptReq.Body = body
	}
	return t.client.Do(attemptReq)
}

func (t *retryTransport) backoff(attempt int) time.Duration {
	delay := t.baseDelay << (attempt - 1)
	if delay > t.maxDelay {
		delay = t.maxDelay
	}
	jitter := time.Duration(rand.Int64N(int64(delay/2)+1)) - delay/4
	return delay + jitter
}

func retryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func drainAndClose(res *http.Response) {
	_, _ = io.Copy(io.Discard, res.Body)
	_ = res.Body.Close()
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type requestSpool struct {
	file *os.File
}

func (s *requestSpool) cleanup() {
	if s == nil || s.file == nil {
		return
	}
	name := s.file.Name()
	_ = s.file.Close()
	_ = os.Remove(name)
}

func spoolRequest(req *http.Request) (*requestSpool, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return &requestSpool{}, nil
	}

	file, err := os.CreateTemp("", "snapotter-request-")
	if err != nil {
		return nil, err
	}

	size, err := io.Copy(file, req.Body)
	if err != nil {
		name := file.Name()
		_ = file.Close()
		_ = os.Remove(name)
		return nil, err
	}
	_ = req.Body.Close()

	req.Body = nil
	req.ContentLength = size
	req.GetBody = func() (io.ReadCloser, error) {
		return os.Open(file.Name())
	}

	return &requestSpool{file: file}, nil
}
