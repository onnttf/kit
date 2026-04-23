package download

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGetFile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	err := GetFile(context.Background(), server.URL, destPath)
	if err != nil {
		t.Fatalf("GetFile() error = %v", err)
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("got %q, want %q", string(data), "hello world")
	}
}

func TestGetFile_EmptyURL(t *testing.T) {
	err := GetFile(context.Background(), "", "/tmp/test.txt")
	if err == nil {
		t.Error("expected error for empty URL")
	}
	if err != nil && err.Error() != "url is empty" {
		t.Errorf("got error %q, want 'url is empty'", err)
	}
}

func TestGetFile_InvalidURL(t *testing.T) {
	err := GetFile(context.Background(), "ftp://example.com/file.txt", "/tmp/test.txt")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	if err != nil && err.Error() != "invalid scheme: scheme=ftp" {
		t.Errorf("got error %q", err)
	}
}

func TestGetFile_InvalidURL2(t *testing.T) {
	err := GetFile(context.Background(), "://example.com/file.txt", "/tmp/test.txt")
	if err == nil {
		t.Error("expected error for empty scheme")
	}
}

func TestGetFile_FileAlreadyExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "exists.txt")

	// Create file first
	if err := os.WriteFile(destPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := GetFile(context.Background(), server.URL, destPath)
	if err == nil {
		t.Error("expected error when file exists")
	}
	if err != nil && err.Error() != "file already exists" {
		t.Errorf("got error %q", err.Error())
	}
}

func TestGetFile_Overwrite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("new content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "overwrite.txt")

	// Create file first
	if err := os.WriteFile(destPath, []byte("old content"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := GetFile(context.Background(), server.URL, destPath, WithOverwrite())
	if err != nil {
		t.Fatalf("GetFile() with overwrite error = %v", err)
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "new content" {
		t.Errorf("got %q, want %q", string(data), "new content")
	}
}

func TestGetFile_CreateDirectory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("content"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "nested", "dir", "test.txt")

	err := GetFile(context.Background(), server.URL, destPath)
	if err != nil {
		t.Fatalf("GetFile() error = %v", err)
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "content" {
		t.Errorf("got %q, want %q", string(data), "content")
	}
}

func TestGetFile_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	err := GetFile(context.Background(), server.URL, destPath)
	if err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestGetFile_ContentLengthExceedsMax(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.Write(make([]byte, 1000))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	err := GetFile(context.Background(), server.URL, destPath, WithMaxBytes(500))
	if err == nil {
		t.Error("expected error when content length exceeds max")
	}
}

func TestGetFile_WithCustomClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("custom client"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	client := &http.Client{}
	err := GetFile(context.Background(), server.URL, destPath, WithClient(client))
	if err != nil {
		t.Fatalf("GetFile() error = %v", err)
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "custom client" {
		t.Errorf("got %q, want %q", string(data), "custom client")
	}
}

func TestGetFile_WithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("timeout test"))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.txt")

	err := GetFile(context.Background(), server.URL, destPath, WithTimeout(5e9)) // 5 seconds
	if err != nil {
		t.Fatalf("GetFile() error = %v", err)
	}
}

func TestGetBytes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bytes content"))
	}))
	defer server.Close()

	data, err := GetBytes(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("GetBytes() error = %v", err)
	}
	if string(data) != "bytes content" {
		t.Errorf("got %q, want %q", string(data), "bytes content")
	}
}

func TestGetBytes_EmptyURL(t *testing.T) {
	_, err := GetBytes(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestGetBytes_InvalidURL(t *testing.T) {
	_, err := GetBytes(context.Background(), "http://")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestGetBytes_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	_, err := GetBytes(context.Background(), server.URL)
	if err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestGetBytes_ContentLengthExceedsMax(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "2000")
		w.Write(make([]byte, 2000))
	}))
	defer server.Close()

	_, err := GetBytes(context.Background(), server.URL, WithMaxBytes(1000))
	if err == nil {
		t.Error("expected error when content length exceeds max")
	}
}

func TestGetBytes_ContentExceedsMax(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No Content-Length header, but body exceeds limit
		w.Write(make([]byte, 2000))
	}))
	defer server.Close()

	_, err := GetBytes(context.Background(), server.URL, WithMaxBytes(1000))
	if err == nil {
		t.Error("expected error when content exceeds max bytes")
	}
}

func TestGetBytes_WithClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("custom"))
	}))
	defer server.Close()

	client := &http.Client{}
	data, err := GetBytes(context.Background(), server.URL, WithClient(client))
	if err != nil {
		t.Fatalf("GetBytes() error = %v", err)
	}
	if string(data) != "custom" {
		t.Errorf("got %q, want %q", string(data), "custom")
	}
}

func TestWithMaxBytes_Invalid(t *testing.T) {
	// Negative value should be ignored
	opt := WithMaxBytes(-100)
	cfg := &config{maxBytes: 1000}
	opt(cfg)
	if cfg.maxBytes != 1000 {
		t.Errorf("expected 1000, got %d", cfg.maxBytes)
	}

	// Zero should be ignored
	opt = WithMaxBytes(0)
	opt(cfg)
	if cfg.maxBytes != 1000 {
		t.Errorf("expected 1000, got %d", cfg.maxBytes)
	}
}

func TestWithClient_Nil(t *testing.T) {
	opt := WithClient(nil)
	cfg := &config{client: &http.Client{}}
	opt(cfg)
	// nil should be ignored, keep existing client
	if cfg.client == nil {
		t.Error("client should not be nil")
	}
}

func TestNewConfig_Defaults(t *testing.T) {
	cfg := newConfig()
	if cfg.client == nil {
		t.Error("client should not be nil")
	}
	if cfg.maxBytes != defaultMaxBytes {
		t.Errorf("maxBytes = %d, want %d", cfg.maxBytes, defaultMaxBytes)
	}
	if cfg.overwrite {
		t.Error("overwrite should be false by default")
	}
}

func TestNewConfig_WithOptions(t *testing.T) {
	client := &http.Client{}
	cfg := newConfig(
		WithClient(client),
		WithMaxBytes(1024),
		WithOverwrite(),
	)

	if cfg.client != client {
		t.Error("client not set correctly")
	}
	if cfg.maxBytes != 1024 {
		t.Errorf("maxBytes = %d, want 1024", cfg.maxBytes)
	}
	if !cfg.overwrite {
		t.Error("overwrite should be true")
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://example.com/file.txt", false},
		{"valid http", "http://example.com/file.txt", false},
		{"invalid scheme", "ftp://example.com", true},
		{"empty host", "http://", true},
		{"empty url", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
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
		wantErr  bool
	}{
		{"ok", http.StatusOK, 100, 200, false},
		{"non-200", 404, 100, 200, true},
		{"length exceeds", http.StatusOK, 300, 200, true},
		{"length zero", http.StatusOK, 0, 200, false},
		{"length negative", http.StatusOK, -1, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode:    tt.status,
				ContentLength: tt.length,
			}
			err := checkResponse(resp, tt.maxBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadLimited(t *testing.T) {
	// Create valid hex encoded data
	original := make([]byte, 50)
	rand.Read(original)
	encoded := make([]byte, hex.EncodedLen(len(original)))
	hex.Encode(encoded, original)

	data, err := readLimited(hex.NewDecoder(bytes.NewReader(encoded)), 100)
	if err != nil {
		t.Fatalf("readLimited() error = %v", err)
	}
	if int64(len(data)) > 100 {
		t.Errorf("read more than maxBytes, len = %d", len(data))
	}
}

func TestReadLimited_Exceeds(t *testing.T) {
	// Read more than limit
	reader := io.LimitReader(hex.NewDecoder(rand.Reader), 200)
	_, err := readLimited(reader, 100)
	if err == nil {
		t.Error("expected error when exceeds max bytes")
	}
}

func TestCopyLimited(t *testing.T) {
	content := "test content"
	reader := &limitTestReader{data: []byte(content)}
	buf := &testWriter{}

	err := copyLimited(buf, reader, 1000)
	if err != nil {
		t.Fatalf("copyLimited() error = %v", err)
	}

	if string(buf.data) != content {
		t.Errorf("got %q, want %q", string(buf.data), content)
	}
}

func TestCopyLimited_Exceeds(t *testing.T) {
	// Reader that returns more than max
	reader := &exceedReader{limit: 200}
	buf := &testWriter{}

	err := copyLimited(buf, reader, 100)
	if err == nil {
		t.Error("expected error when exceeds max bytes")
	}
}

func TestCopyLimited_FromReader(t *testing.T) {
	content := "hello world"
	reader := &limitTestReader{data: []byte(content)}
	buf := &testWriter{}

	err := copyLimited(buf, reader, 100)
	if err != nil {
		t.Fatalf("copyLimited() error = %v", err)
	}
}

// testWriter implements io.Writer for testing.
type testWriter struct {
	data []byte
}

func (tw *testWriter) Write(p []byte) (int, error) {
	tw.data = append(tw.data, p...)
	return len(p), nil
}

// exceedReader is a reader that returns more data than the specified limit.
type exceedReader struct {
	limit  int64
	called int
}

func (er *exceedReader) Read(p []byte) (int, error) {
	er.called++
	if er.called == 1 {
		// Return more than limit in first read
		size := int(er.limit + 1)
		if size > len(p) {
			size = len(p)
		}
		for i := 0; i < size; i++ {
			p[i] = 'x'
		}
		return size, nil
	}
	return 0, io.EOF
}

// limitTestReader implements io.Reader for testing.
type limitTestReader struct {
	data   []byte
	offset int
}

func (r *limitTestReader) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func BenchmarkGetFile(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("benchmark data"))
	}))
	defer server.Close()

	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destPath := filepath.Join(tmpDir, "bench", "test.txt")
		_ = GetFile(context.Background(), server.URL, destPath)
		os.RemoveAll(filepath.Join(tmpDir, "bench"))
	}
}

func BenchmarkGetBytes(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("benchmark data"))
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetBytes(context.Background(), server.URL)
	}
}
