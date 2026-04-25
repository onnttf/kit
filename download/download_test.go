package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	err := GetFile(context.Background(), server.URL, destPath)
	require.NoError(t, err)
	_, err = os.Stat(destPath)
	assert.NoError(t, err)
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
	f, _ := os.Create(destPath)
	f.Close()
	err := GetFile(context.Background(), "https://example.com/file.txt", destPath)
	assert.ErrorIs(t, err, ErrFileExists)
}

func TestGetFile_Overwrite(t *testing.T) {
	content := "new content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	f, _ := os.Create(destPath)
	f.WriteString("old content")
	f.Close()

	err := GetFile(context.Background(), server.URL, destPath, WithOverwrite())
	require.NoError(t, err)
}

func TestGetFile_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")
	err := GetFile(nil, server.URL, destPath)
	require.NoError(t, err)
}

func TestGetBytes(t *testing.T) {
	content := "test bytes content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
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

func TestErrors(t *testing.T) {
	assert.Equal(t, "file already exists", ErrFileExists.Error())
	assert.Equal(t, "url is empty", ErrEmptyURL.Error())
	assert.Equal(t, "invalid scheme", ErrInvalidScheme.Error())
	assert.Equal(t, "host is empty", ErrEmptyHost.Error())
	assert.Equal(t, "unexpected http status", ErrUnexpectedStatus.Error())
	assert.Equal(t, "body too large", ErrResponseBodyTooLarge.Error())
}

func TestNewConfig(t *testing.T) {
	cfg := newConfig()
	assert.NotNil(t, cfg.client)
	assert.Equal(t, int64(defaultMaxBytes), cfg.maxBytes)
	assert.False(t, cfg.overwrite)
}