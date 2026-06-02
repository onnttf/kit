package download

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithClient(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	opt := WithClient(customClient)
	cfg := new(config)
	opt(cfg)
	assert.Equal(t, customClient, cfg.client)
}

func TestWithClient_NilClient(t *testing.T) {
	opt := WithClient(nil)
	cfg := new(config)
	opt(cfg)
	assert.Nil(t, cfg.client)
}

func TestWithMaxBytes(t *testing.T) {
	opt := WithMaxBytes(1024)
	cfg := new(config)
	opt(cfg)
	assert.Equal(t, int64(1024), cfg.maxBytes)
}

func TestWithMaxBytes_Zero(t *testing.T) {
	opt := WithMaxBytes(0)
	cfg := &config{maxBytes: 100}
	opt(cfg)
	assert.Equal(t, int64(100), cfg.maxBytes)
}

func TestWithOverwrite(t *testing.T) {
	opt := WithOverwrite()
	cfg := new(config)
	opt(cfg)
	assert.True(t, cfg.overwrite)
}

func TestWithTimeout(t *testing.T) {
	opt := WithTimeout(30 * time.Second)
	cfg := new(config)
	client := *getDefaultClient()
	cfg.client = &client
	opt(cfg)
	assert.Equal(t, 30*time.Second, cfg.client.Timeout)
}

func TestGetFile(t *testing.T) {
	content := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(content))
		assert.NoError(t, err)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	err := GetFile(context.Background(), server.URL, destPath)
	require.NoError(t, err)
	data, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestGetFile_EmptyURL(t *testing.T) {
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	err := GetFile(context.Background(), "", destPath)
	assert.ErrorIs(t, err, ErrEmptyURL)
}

func TestGetFile_InvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	err := GetFile(context.Background(), "://invalid", destPath)
	assert.Error(t, err)
}

func TestGetFile_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	f, err := os.Create(destPath)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	err = GetFile(context.Background(), "https://example.com/file.txt", destPath)
	assert.ErrorIs(t, err, ErrFileExists)
}

func TestGetFile_Overwrite(t *testing.T) {
	content := "new content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(content))
		assert.NoError(t, err)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(destPath, []byte("old content"), 0o644))

	err := GetFile(context.Background(), server.URL, destPath, WithOverwrite())
	require.NoError(t, err)
	data, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestGetFile_NoOverwriteRace(t *testing.T) {
	content := "new content"
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := os.WriteFile(destPath, []byte("existing"), 0o644); err != nil {
			t.Errorf("write existing file: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(content)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	err := GetFile(context.Background(), server.URL, destPath)
	require.ErrorIs(t, err, ErrFileExists)

	data, readErr := os.ReadFile(destPath)
	require.NoError(t, readErr)
	assert.Equal(t, "existing", string(data))
}

func TestGetFile_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	err := GetFile(context.Background(), server.URL, destPath)
	assert.Error(t, err)
}

func TestGetFile_NilContext(t *testing.T) {
	content := "test"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(content))
		assert.NoError(t, err)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	var ctx context.Context
	err := GetFile(ctx, server.URL, destPath)
	require.NoError(t, err)
}

func TestGetBytes(t *testing.T) {
	content := "test bytes content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(content))
		assert.NoError(t, err)
	}))
	defer server.Close()

	data, err := GetBytes(context.Background(), server.URL)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestGetBytes_EmptyURL(t *testing.T) {
	_, err := GetBytes(context.Background(), "")
	assert.ErrorIs(t, err, ErrEmptyURL)
}

func TestGetBytes_BodyTooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("abcdef")); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	_, err := GetBytes(context.Background(), server.URL, WithMaxBytes(3))
	assert.ErrorIs(t, err, ErrResponseBodyTooLarge)
}

func TestGetBytes_Non200DiscardsLimitedBody(t *testing.T) {
	body := &countingBody{}
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       body,
			}, nil
		}),
	}

	_, err := GetBytes(
		context.Background(),
		"https://example.com/error",
		WithClient(client),
		WithMaxBytes(3),
	)

	assert.ErrorIs(t, err, ErrUnexpectedStatus)
	assert.LessOrEqual(t, body.read, 3)
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		url      string
		hasError bool
	}{
		{"https://example.com", false},
		{"http://example.com", false},
		{"", true},
		{"ftp://example.com", true},
		{"://invalid", true},
		{"https://", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := validateURL(tt.url)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadLimited(t *testing.T) {
	data, err := readLimited(strings.NewReader("abc"), 3)
	require.NoError(t, err)
	assert.Equal(t, []byte("abc"), data)

	_, err = readLimited(strings.NewReader("abcd"), 3)
	assert.ErrorIs(t, err, ErrResponseBodyTooLarge)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type countingBody struct {
	read int
}

func (b *countingBody) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = 'x'
	b.read++
	return 1, nil
}

func (b *countingBody) Close() error {
	return nil
}

var _ io.ReadCloser = (*countingBody)(nil)

func TestCopyLimited(t *testing.T) {
	var dst strings.Builder
	err := copyLimited(&dst, strings.NewReader("abc"), 3)
	require.NoError(t, err)
	assert.Equal(t, "abc", dst.String())

	err = copyLimited(&dst, strings.NewReader("abcd"), 3)
	assert.ErrorIs(t, err, ErrResponseBodyTooLarge)
}

func TestGetFile_LinkErrorWrapsUnexpectedFailure(t *testing.T) {
	err := copyLimited(errWriter{}, strings.NewReader("data"), defaultMaxBytes)
	assert.Error(t, err)
}

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		length   int64
		maxBytes int64
		hasError bool
	}{
		{"ok", http.StatusOK, 100, 200, false},
		{"not ok", http.StatusNotFound, 0, 100, true},
		{"content too large", http.StatusOK, 200, 100, true},
		{"no content length", http.StatusOK, -1, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode:    tt.status,
				ContentLength: tt.length,
			}
			err := checkResponse(resp, tt.maxBytes)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	cfg := newConfig()
	assert.NotNil(t, cfg.client)
	assert.Equal(t, int64(defaultMaxBytes), cfg.maxBytes)
	assert.False(t, cfg.overwrite)
}

func TestNewConfig_DefaultTransportClonesHTTPDefaultTransport(t *testing.T) {
	cfg := newConfig()

	transport, ok := cfg.client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.NotSame(t, http.DefaultTransport, transport)
	assert.NotNil(t, transport.Proxy)
	assert.NotNil(t, transport.DialContext)
	assert.Equal(t, 100, transport.MaxIdleConnsPerHost)
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
