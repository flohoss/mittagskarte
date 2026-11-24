package snapotter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestTransport(client *http.Client, sleeps *[]time.Duration) *retryTransport {
	t := newRetryTransport(client)
	t.sleep = func(_ context.Context, d time.Duration) error {
		if sleeps != nil {
			*sleeps = append(*sleeps, d)
		}
		return nil
	}
	return t
}

func TestRetryTransportRetriesTransientFailures(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
			hijackAndClose(w)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	res, err := newTestTransport(server.Client(), nil).Do(newGetRequest(t, server.URL))
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer res.Body.Close()

	if got := attempts.Load(); got != 3 {
		t.Fatalf("attempt count, got %d want 3", got)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status, got %d want 200", res.StatusCode)
	}
}

func TestRetryTransportReplaysUploadBody(t *testing.T) {
	t.Parallel()

	const payload = "menu-bytes"

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if string(body) != payload {
			t.Errorf("attempt %d body, got %q want %q", attempts.Load()+1, string(body), payload)
		}
		if attempts.Add(1) < 2 {
			hijackAndClose(w)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, strings.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	res, err := newTestTransport(server.Client(), nil).Do(req)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer res.Body.Close()

	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempt count, got %d want 2", got)
	}
}

func TestRetryTransportGivesUpAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		hijackAndClose(w)
	}))
	defer server.Close()

	if _, err := newTestTransport(server.Client(), nil).Do(newGetRequest(t, server.URL)); err == nil {
		t.Fatal("expected error after exhausting attempts, got nil")
	}
	if got := attempts.Load(); got != MaxAttempts {
		t.Fatalf("attempt count, got %d want %d", got, MaxAttempts)
	}
}

func TestRetryTransportRetriesRetryableStatuses(t *testing.T) {
	t.Parallel()

	for _, status := range []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if attempts.Add(1) == 1 {
					w.WriteHeader(status)
					return
				}
				_, _ = w.Write([]byte("ok"))
			}))
			defer server.Close()

			res, err := newTestTransport(server.Client(), nil).Do(newGetRequest(t, server.URL))
			if err != nil {
				t.Fatalf("Do returned error: %v", err)
			}
			defer res.Body.Close()

			if got := attempts.Load(); got != 2 {
				t.Fatalf("attempt count, got %d want 2", got)
			}
			if res.StatusCode != http.StatusOK {
				t.Fatalf("status, got %d want 200", res.StatusCode)
			}
		})
	}
}

func TestRetryTransportReturnsRetryableStatusAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := newTestTransport(server.Client(), nil).Do(newGetRequest(t, server.URL))
	if err == nil {
		t.Fatal("expected error after exhausting attempts, got nil")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Fatalf("error should mention the status, got %v", err)
	}
	if got := attempts.Load(); got != MaxAttempts {
		t.Fatalf("attempt count, got %d want %d", got, MaxAttempts)
	}
}

func TestRetryTransportDoesNotRetryClientErrors(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	res, err := newTestTransport(server.Client(), nil).Do(newGetRequest(t, server.URL))
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	defer res.Body.Close()

	if got := attempts.Load(); got != 1 {
		t.Fatalf("attempt count, got %d want 1", got)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status, got %d want 400", res.StatusCode)
	}
}

func TestRetryTransportBackoffGrowsAndIsCapped(t *testing.T) {
	t.Parallel()

	transport := newRetryTransport(&http.Client{})

	first := transport.backoff(1)
	if first < InitialRetryDelay*3/4 || first > InitialRetryDelay*5/4 {
		t.Fatalf("first backoff outside expected jitter range: %v", first)
	}

	second := transport.backoff(2)
	if second <= first {
		t.Fatalf("backoff should grow, got %v after %v", second, first)
	}

	for attempt := 3; attempt <= 10; attempt++ {
		if got := transport.backoff(attempt); got > MaxRetryDelay*5/4 {
			t.Fatalf("backoff for attempt %d exceeds cap: %v", attempt, got)
		}
	}
}

func TestRetryTransportStopsWhenContextIsCancelled(t *testing.T) {
	t.Parallel()

	transport := newRetryTransport(&http.Client{})
	transport.attempts = 5

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:1", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Body = io.NopCloser(strings.NewReader("payload"))

	_, err = transport.Do(req)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestSpoolRequestRemovesTempFileOnCleanup(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://example.test", strings.NewReader("payload"))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	spool, err := spoolRequest(req)
	if err != nil {
		t.Fatalf("spoolRequest returned error: %v", err)
	}
	path := spool.file.Name()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("spool file should exist: %v", err)
	}

	spool.cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("spool file should be removed, got %v", err)
	}
}

func newGetRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	return req
}

func hijackAndClose(w http.ResponseWriter) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	_ = conn.Close()
}

func TestRetryTransportHonorsClientTimeout(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
		_, _ = w.Write([]byte("too late"))
	}))
	defer func() {
		close(release)
		server.Close()
	}()

	transport := newTestTransport(&http.Client{Timeout: 50 * time.Millisecond}, nil)
	transport.attempts = 1

	start := time.Now()
	_, err := transport.Do(newGetRequest(t, server.URL))
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("timeout took too long: %v", elapsed)
	}
}
