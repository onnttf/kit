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

var (
	ErrEmptyURL = errors.New("download: url is empty")

	ErrEmptyName = errors.New("download: file name is empty")

	ErrFileExists = errors.New("download: file exists")

	ErrInvalidScheme = errors.New("download: invalid scheme")

	ErrEmptyHost = errors.New("download: host is empty")

	ErrUnexpectedStatus = errors.New("download: unexpected http status")

	ErrResponseBodyTooLarge = errors.New("download: body too large")
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
		if d <= 0 {
			return
		}
		client := *c.client
		client.Timeout = d
		c.client = &client
	}
}

func GetFile(ctx context.Context, rawURL, name string, opts ...Option) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := newConfig(opts...)

	if rawURL == "" {
		return ErrEmptyURL
	}
	if name == "" {
		return ErrEmptyName
	}
	if err := validateURL(rawURL); err != nil {
		return err
	}

	if !cfg.overwrite {
		if _, err := os.Stat(name); err == nil {
			return ErrFileExists
		}
	}

	resp, err := openResponse(ctx, cfg, rawURL)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close response body: %w", closeErr)
		}
	}()

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

	resp, err := openResponse(ctx, cfg, rawURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close response body: %w", closeErr)
		}
	}()

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
		if opt == nil {
			continue
		}
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

func openResponse(ctx context.Context, cfg *config, rawURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := cfg.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if err := checkResponse(resp, cfg.maxBytes); err != nil {
		if _, copyErr := io.Copy(io.Discard, io.LimitReader(resp.Body, cfg.maxBytes)); copyErr != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				return nil, errors.Join(
					fmt.Errorf("discard response body: %w", copyErr),
					fmt.Errorf("close response body: %w", closeErr),
				)
			}
			return nil, fmt.Errorf("discard response body: %w", copyErr)
		}
		if closeErr := resp.Body.Close(); closeErr != nil {
			return nil, errors.Join(err, fmt.Errorf("close response body: %w", closeErr))
		}
		return nil, err
	}

	return resp, nil
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
