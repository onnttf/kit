package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBytes(t *testing.T) {
	var header string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Get("X-Test")
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	got, err := Bytes(context.Background(), srv.URL, WithHeader("X-Test", "yes"))
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}
	if string(got) != "ok" || header != "yes" {
		t.Fatalf("Bytes() = %q, header %q", got, header)
	}
}

func TestBytesErrors(t *testing.T) {
	if _, err := Bytes(context.Background(), "ftp://example.com"); !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("invalid url error = %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("too large"))
	}))
	defer srv.Close()
	if _, err := Bytes(context.Background(), srv.URL, WithMaxBytes(2)); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("too large error = %v", err)
	}
}

func TestFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file"))
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "nested", "out.txt")
	if err := File(context.Background(), srv.URL, path); err != nil {
		t.Fatalf("File() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "file" {
		t.Fatalf("file data = %q", data)
	}
	if err := File(context.Background(), srv.URL, path); !errors.Is(err, ErrExists) {
		t.Fatalf("File(existing) error = %v", err)
	}
	if err := File(context.Background(), srv.URL, path, WithOverwrite(true)); err != nil {
		t.Fatalf("File(overwrite) error = %v", err)
	}
}

func TestClientReuseAndStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no", http.StatusTeapot)
	}))
	defer srv.Close()

	client := New(WithTimeout(time.Second))
	if _, err := client.Bytes(context.Background(), srv.URL); !errors.Is(err, ErrUnexpectedStatus) {
		t.Fatalf("Bytes(status) error = %v", err)
	}
}

func TestBytesDefaultOptionsAndInvalidURLs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	got, err := New(
		WithHeader("", "ignored"),
		WithTimeout(0),
		WithMaxBytes(0),
	).Bytes(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Bytes() error = %v", err)
	}
	if string(got) != "ok" {
		t.Fatalf("Bytes() = %q", got)
	}

	for _, rawURL := range []string{"", "://bad", "mailto:a@example.com", "http://"} {
		t.Run(rawURL, func(t *testing.T) {
			_, err := Bytes(context.Background(), rawURL)
			if !errors.Is(err, ErrInvalidURL) {
				t.Fatalf("Bytes(%q) error = %v", rawURL, err)
			}
		})
	}
}

func TestBytesStatusAndRequestFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, strings.Repeat("x", 8192), http.StatusTeapot)
	}))
	defer srv.Close()
	if _, err := Bytes(context.Background(), srv.URL); !errors.Is(err, ErrUnexpectedStatus) {
		t.Fatalf("Bytes(status) error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Bytes(ctx, "https://example.com"); err == nil || !strings.Contains(err.Error(), "download: request") {
		t.Fatalf("Bytes(canceled) error = %v", err)
	}
}

func TestBytesAllowsExactlyMaxBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("abc"))
	}))
	defer srv.Close()

	got, err := Bytes(context.Background(), srv.URL, WithMaxBytes(3))
	if err != nil {
		t.Fatalf("Bytes(exact max) error = %v", err)
	}
	if string(got) != "abc" {
		t.Fatalf("Bytes(exact max) = %q", got)
	}
}

func TestFileEdgeCases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("0123456789"))
	}))
	defer srv.Close()

	if err := File(context.Background(), srv.URL, ""); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("File(empty path) error = %v", err)
	}

	tooLarge := filepath.Join(t.TempDir(), "too-large.txt")
	if err := File(context.Background(), srv.URL, tooLarge, WithMaxBytes(3)); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("File(too large) error = %v", err)
	}
	if _, err := os.Stat(tooLarge); !os.IsNotExist(err) {
		t.Fatalf("too-large target exists, stat err = %v", err)
	}

	path := filepath.Join(t.TempDir(), "out.txt")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile(old) error = %v", err)
	}
	if err := File(context.Background(), srv.URL, path, WithOverwrite(true)); err != nil {
		t.Fatalf("File(overwrite existing) error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(overwritten) error = %v", err)
	}
	if string(data) != "0123456789" {
		t.Fatalf("overwritten data = %q", data)
	}
}

func TestCustomHTTPClientAndTimeout(t *testing.T) {
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "example.com" {
			return nil, fmt.Errorf("unexpected host %s", req.URL.Host)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("custom")),
			Request:    req,
		}, nil
	})
	client := New(WithHTTPClient(&http.Client{Transport: transport}), WithTimeout(time.Second))
	got, err := client.Bytes(context.Background(), "https://example.com/file")
	if err != nil {
		t.Fatalf("Bytes(custom client) error = %v", err)
	}
	if string(got) != "custom" {
		t.Fatalf("Bytes(custom client) = %q", got)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
