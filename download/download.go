package download

import (
	"bytes"
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

const defaultMaxBytes = 100 << 20

var (
	ErrInvalidURL       = errors.New("download: invalid url")
	ErrInvalidPath      = errors.New("download: invalid path")
	ErrUnexpectedStatus = errors.New("download: unexpected status")
	ErrTooLarge         = errors.New("download: too large")
	ErrExists           = errors.New("download: file exists")
)

type Client struct {
	httpClient *http.Client
	maxBytes   int64
	headers    http.Header
	overwrite  bool
}

type Option func(*Client)

var defaultHTTPClient = sync.OnceValue(func() *http.Client {
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: defaultTransport(),
	}
})

func New(opts ...Option) *Client {
	base := *defaultHTTPClient()
	c := &Client{
		httpClient: &base,
		maxBytes:   defaultMaxBytes,
		headers:    make(http.Header),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d <= 0 {
			return
		}
		client := *c.httpClient
		client.Timeout = d
		c.httpClient = &client
	}
}

func WithMaxBytes(n int64) Option {
	return func(c *Client) {
		if n > 0 {
			c.maxBytes = n
		}
	}
}

func WithHeader(key, value string) Option {
	return func(c *Client) {
		if key != "" {
			c.headers.Set(key, value)
		}
	}
}

func WithOverwrite(v bool) Option {
	return func(c *Client) {
		c.overwrite = v
	}
}

func Bytes(ctx context.Context, rawURL string, opts ...Option) ([]byte, error) {
	return New(opts...).Bytes(ctx, rawURL)
}

func File(ctx context.Context, rawURL, path string, opts ...Option) error {
	return New(opts...).File(ctx, rawURL, path)
}

func (c *Client) Bytes(ctx context.Context, rawURL string) ([]byte, error) {
	resp, err := c.open(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var buf bytes.Buffer
	if err := copyLimited(&buf, resp.Body, c.maxBytes); err != nil {
		return nil, fmt.Errorf("download: read response body: %w", err)
	}
	return buf.Bytes(), nil
}

func (c *Client) File(ctx context.Context, rawURL, path string) (err error) {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPath)
	}
	if !c.overwrite {
		if _, err := os.Stat(path); err == nil {
			return ErrExists
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("download: stat file: %w", err)
		}
	}

	resp, err := c.open(ctx, rawURL)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("download: create directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("download: create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		removeErr := os.Remove(tmpPath)
		if err == nil && removeErr != nil && !os.IsNotExist(removeErr) {
			err = fmt.Errorf("download: remove temp file: %w", removeErr)
		}
	}()

	if err := copyLimited(tmp, resp.Body, c.maxBytes); err != nil {
		closeErr := tmp.Close()
		if closeErr != nil {
			return errors.Join(
				fmt.Errorf("download: write temp file: %w", err),
				fmt.Errorf("download: close temp file: %w", closeErr),
			)
		}
		return fmt.Errorf("download: write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("download: close temp file: %w", err)
	}

	if c.overwrite {
		if err := os.Rename(tmpPath, path); err != nil {
			if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
				return fmt.Errorf("download: remove existing file: %w", removeErr)
			}
			if renameErr := os.Rename(tmpPath, path); renameErr != nil {
				return fmt.Errorf("download: rename temp file: %w", renameErr)
			}
		}
		return nil
	}

	if err := os.Link(tmpPath, path); err != nil {
		if os.IsExist(err) {
			return ErrExists
		}
		return fmt.Errorf("download: link temp file: %w", err)
	}
	return nil
}

func (c *Client) open(ctx context.Context, rawURL string) (*http.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validateURL(rawURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("download: create request: %w", err)
	}
	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.CopyN(io.Discard, resp.Body, min(c.maxBytes, 4<<10))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%w: status=%q", ErrUnexpectedStatus, resp.Status)
	}
	return resp, nil
}

func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%w: url=%q", ErrInvalidURL, rawURL)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: scheme=%q", ErrInvalidURL, u.Scheme)
	}
	return nil
}

func copyLimited(dst io.Writer, src io.Reader, maxBytes int64) error {
	limited := &io.LimitedReader{R: src, N: maxBytes + 1}
	n, err := io.Copy(dst, limited)
	if err != nil {
		return err
	}
	if n > maxBytes {
		return ErrTooLarge
	}
	return nil
}

func defaultTransport() *http.Transport {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		return transport.Clone()
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
