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
	// defaultMaxBytes is the default maximum response body size (100 MB).
	defaultMaxBytes = 100 << 20
)

// ErrFileExists is returned when the target file already exists and overwrite is not enabled.
var ErrFileExists = errors.New("file already exists")

// ErrEmptyURL indicates a request URL is empty.
var ErrEmptyURL = errors.New("url is empty")

// ErrInvalidScheme indicates an unsupported URL scheme.
var ErrInvalidScheme = errors.New("invalid scheme")

// ErrEmptyHost indicates a request URL has an empty host.
var ErrEmptyHost = errors.New("host is empty")

// ErrUnexpectedStatus indicates an unexpected HTTP status code.
var ErrUnexpectedStatus = errors.New("unexpected http status")

// ErrResponseBodyTooLarge indicates the response body exceeds the maximum allowed size.
var ErrResponseBodyTooLarge = errors.New("body too large")

var getDefaultClient = sync.OnceValue(func() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
	}
})

// config holds the configuration for a download operation.
type config struct {
	client    *http.Client
	maxBytes  int64
	overwrite bool
}

// Option configures a download operation.
type Option func(*config)

// WithClient sets a custom HTTP client.
func WithClient(client *http.Client) Option {
	return func(c *config) {
		if client != nil {
			c.client = client
		}
	}
}

// WithMaxBytes sets the maximum allowed response body size.
func WithMaxBytes(n int64) Option {
	return func(c *config) {
		if n > 0 {
			c.maxBytes = n
		}
	}
}

// WithOverwrite allows overwriting the target file if it exists.
func WithOverwrite() Option {
	return func(c *config) {
		c.overwrite = true
	}
}

// WithTimeout sets the HTTP client's timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		client := *c.client
		client.Timeout = d
		c.client = &client
	}
}

// GetFile downloads content from the URL to the given path.
// The download is atomic: content is first written to a temporary file and then renamed.
// The parent directory is created if it does not exist.
// If ctx is nil, context.Background() is used.
func GetFile(ctx context.Context, rawURL, name string, opts ...Option) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := newConfig(opts...)

	if rawURL == "" {
		return ErrEmptyURL
	}
	if err := validateURL(rawURL); err != nil {
		return fmt.Errorf("%w", err)
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
	defer resp.Body.Close()

	if err := checkResponse(resp, cfg.maxBytes); err != nil {
		io.Copy(io.Discard, resp.Body)
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
	defer os.Remove(tmpPath)

	if err := copyLimited(tmpFile, resp.Body, cfg.maxBytes); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
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

	if err := os.Rename(tmpPath, name); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// GetBytes downloads content from the URL and returns it as a byte slice.
// If ctx is nil, context.Background() is used.
func GetBytes(ctx context.Context, rawURL string, opts ...Option) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := newConfig(opts...)

	if rawURL == "" {
		return nil, ErrEmptyURL
	}
	if err := validateURL(rawURL); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := cfg.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp, cfg.maxBytes); err != nil {
		io.Copy(io.Discard, resp.Body)
		return nil, err
	}

	data, err := readLimited(resp.Body, cfg.maxBytes)
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
		return fmt.Errorf("%w", ErrInvalidScheme)
	}
	if u.Host == "" {
		return fmt.Errorf("%w", ErrEmptyHost)
	}
	return nil
}

func checkResponse(resp *http.Response, maxBytes int64) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w", ErrUnexpectedStatus)
	}
	if resp.ContentLength > 0 && resp.ContentLength > maxBytes {
		return fmt.Errorf("%w", ErrResponseBodyTooLarge)
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
		return nil, fmt.Errorf("%w", ErrResponseBodyTooLarge)
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
		return fmt.Errorf("%w", ErrResponseBodyTooLarge)
	}
	return nil
}
