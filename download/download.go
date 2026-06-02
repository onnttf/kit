package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultMaxBytes = 100 << 20
)

// ErrFileExists is returned when the destination already exists and overwrite is disabled.
var ErrFileExists = errors.New("file already exists")

var (
	// ErrEmptyURL is returned when the source URL is empty.
	ErrEmptyURL = errors.New("url is empty")
	// ErrInvalidScheme is returned for non-HTTP(S) URLs.
	ErrInvalidScheme = errors.New("invalid scheme")
	// ErrEmptyHost is returned when the URL has no host.
	ErrEmptyHost = errors.New("host is empty")
	// ErrUnexpectedStatus is returned for non-200 HTTP responses.
	ErrUnexpectedStatus = errors.New("unexpected http status")
	// ErrResponseBodyTooLarge is returned when the configured size limit is exceeded.
	ErrResponseBodyTooLarge = errors.New("body too large")
)

var getDefaultClient = sync.OnceValue(func() *http.Client {
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: defaultTransport(),
	}
})

type config struct {
	client    *http.Client
	maxBytes  int64
	overwrite bool
}

type Option func(*config)

func WithClient(client *http.Client) Option {
	return func(c *config) {
		if client != nil {
			c.client = client
		}
	}
}

func WithMaxBytes(n int64) Option {
	return func(c *config) {
		if n > 0 {
			c.maxBytes = n
		}
	}
}

func WithOverwrite() Option {
	return func(c *config) {
		c.overwrite = true
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		client := *c.client
		client.Timeout = d
		c.client = &client
	}
}

// GetFile downloads rawURL to name using an atomic temporary file.
// A nil context is treated as context.Background.
func GetFile(ctx context.Context, rawURL, name string, opts ...Option) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := newConfig(opts...)

	if rawURL == "" {
		return ErrEmptyURL
	}
	if err := validateURL(rawURL); err != nil {
		return err
	}

	if !cfg.overwrite {
		if _, err := os.Stat(name); err == nil {
			return ErrFileExists
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := cfg.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close response body: %w", closeErr)
		}
	}()

	if err := checkResponse(resp, cfg.maxBytes); err != nil {
		if _, copyErr := discardLimited(resp.Body, cfg.maxBytes); copyErr != nil {
			return fmt.Errorf("discard response body: %w", copyErr)
		}
		return err
	}

	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(name)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		removeErr := os.Remove(tmpPath)
		if err == nil && removeErr != nil && !os.IsNotExist(removeErr) {
			err = fmt.Errorf("remove temp file: %w", removeErr)
		}
	}()

	if copyErr := copyLimited(tmpFile, resp.Body, cfg.maxBytes); copyErr != nil {
		if closeErr := tmpFile.Close(); closeErr != nil {
			return errors.Join(
				fmt.Errorf("write temp file: %w", copyErr),
				fmt.Errorf("close temp file: %w", closeErr),
			)
		}
		return fmt.Errorf("write temp file: %w", copyErr)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if cfg.overwrite {
		if err := os.Rename(tmpPath, name); err != nil {
			if removeErr := os.Remove(name); removeErr != nil && !os.IsNotExist(removeErr) {
				return fmt.Errorf("remove existing file: %w", removeErr)
			}
			if renameErr := os.Rename(tmpPath, name); renameErr != nil {
				return fmt.Errorf("rename temp file: %w", renameErr)
			}
		}
		return nil
	}

	if err := os.Link(tmpPath, name); err != nil {
		if os.IsExist(err) {
			return ErrFileExists
		}
		return fmt.Errorf("link temp file: %w", err)
	}

	return nil
}

// A nil context is treated as context.Background.
func GetBytes(ctx context.Context, rawURL string, opts ...Option) (data []byte, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := newConfig(opts...)

	if rawURL == "" {
		return nil, ErrEmptyURL
	}
	if err := validateURL(rawURL); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := cfg.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close response body: %w", closeErr)
		}
	}()

	if err := checkResponse(resp, cfg.maxBytes); err != nil {
		if _, copyErr := discardLimited(resp.Body, cfg.maxBytes); copyErr != nil {
			return nil, fmt.Errorf("discard response body: %w", copyErr)
		}
		return nil, err
	}

	data, err = readLimited(resp.Body, cfg.maxBytes)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return data, nil
}

func newConfig(opts ...Option) *config {
	client := *getDefaultClient()
	cfg := &config{
		client:   &client,
		maxBytes: defaultMaxBytes,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidScheme
	}
	if u.Host == "" {
		return ErrEmptyHost
	}
	return nil
}

func defaultTransport() *http.Transport {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		clone := transport.Clone()
		clone.MaxIdleConnsPerHost = 100
		return clone
	}
	return &http.Transport{MaxIdleConnsPerHost: 100}
}

func checkResponse(resp *http.Response, maxBytes int64) error {
	if resp.StatusCode != http.StatusOK {
		return ErrUnexpectedStatus
	}
	if resp.ContentLength > 0 && resp.ContentLength > maxBytes {
		return ErrResponseBodyTooLarge
	}
	return nil
}

func readLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	lr := io.LimitReader(r, maxBytes+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrResponseBodyTooLarge
	}
	return data, nil
}

func copyLimited(dst io.Writer, src io.Reader, maxBytes int64) error {
	lr := io.LimitReader(src, maxBytes+1)
	n, err := io.Copy(dst, lr)
	if err != nil {
		return err
	}
	if n > maxBytes {
		return ErrResponseBodyTooLarge
	}
	return nil
}

func discardLimited(r io.Reader, maxBytes int64) (int64, error) {
	return io.Copy(io.Discard, io.LimitReader(r, maxBytes))
}
