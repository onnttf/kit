package dingtalk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	client := New(srv.URL)
	if err := client.Send(context.Background(), NewText("done")); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got["msgtype"] != "text" {
		t.Fatalf("payload = %v", got)
	}
}

func TestSendErrors(t *testing.T) {
	if err := New("").Send(context.Background(), NewText("x")); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty webhook error = %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":1,"errmsg":"bad"}`))
	}))
	defer srv.Close()
	if err := New(srv.URL).Send(context.Background(), NewText("x")); !errors.Is(err, ErrUnexpected) {
		t.Fatalf("unexpected response error = %v", err)
	}
}

func TestPayloads(t *testing.T) {
	messages := []Message{
		NewText("text"),
		NewMarkdown("title", "text"),
		NewLink("title", "text", "https://example.com"),
		NewSingleActionCard("title", "text", "go", "https://example.com"),
		NewMultiActionCard("title", "text", []ActionCardButton{{Title: "go", ActionURL: "https://example.com"}}),
		NewFeedCard([]FeedLink{{Title: "go", MessageURL: "https://example.com", PicURL: "https://example.com/a.png"}}),
	}
	for _, msg := range messages {
		data, err := msg.Payload()
		if err != nil {
			t.Fatalf("%T Payload() error = %v", msg, err)
		}
		if !json.Valid(data) || !strings.Contains(string(data), "msgtype") {
			t.Fatalf("%T payload = %s", msg, data)
		}
	}
}

func TestSignedEndpoint(t *testing.T) {
	endpoint, err := New("https://example.com/hook?access_token=x", WithSecret("secret")).endpoint()
	if err != nil {
		t.Fatalf("endpoint() error = %v", err)
	}
	if !strings.Contains(endpoint, "timestamp=") || !strings.Contains(endpoint, "sign=") {
		t.Fatalf("endpoint = %s", endpoint)
	}
}

type failingMessage struct{}

func (failingMessage) Payload() ([]byte, error) {
	return nil, ErrInvalidMessage
}

func TestSendRequestDetailsAndSignedURL(t *testing.T) {
	var method, contentType string
	var query url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		contentType = r.Header.Get("Content-Type")
		query = r.URL.Query()
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	client := New(
		srv.URL+"?access_token=x",
		WithSecret("secret"),
		WithTimeout(time.Second),
	)
	if err := client.Send(context.Background(), NewText("done")); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if method != http.MethodPost || contentType != "application/json" {
		t.Fatalf("request method/content-type = %q/%q", method, contentType)
	}
	if query.Get("access_token") != "x" || query.Get("timestamp") == "" || query.Get("sign") == "" {
		t.Fatalf("signed query = %v", query)
	}
}

func TestSendErrorPaths(t *testing.T) {
	t.Run("nil message", func(t *testing.T) {
		if err := New("https://example.com").Send(context.Background(), nil); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("Send(nil) error = %v", err)
		}
	})

	t.Run("payload error", func(t *testing.T) {
		err := New("https://example.com").Send(context.Background(), failingMessage{})
		if !errors.Is(err, ErrInvalidMessage) {
			t.Fatalf("Send(payload error) = %v", err)
		}
	})

	t.Run("http status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "no", http.StatusInternalServerError)
		}))
		defer srv.Close()
		if err := New(srv.URL).Send(context.Background(), NewText("x")); !errors.Is(err, ErrUnexpected) {
			t.Fatalf("Send(status) error = %v", err)
		}
	})

	t.Run("invalid json response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`not-json`))
		}))
		defer srv.Close()
		err := New(srv.URL).Send(context.Background(), NewText("x"))
		if err == nil || !strings.Contains(err.Error(), "decode response") {
			t.Fatalf("Send(invalid json) error = %v", err)
		}
	})

	t.Run("request failure", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := New("https://example.com").Send(ctx, NewText("x"))
		if err == nil || !strings.Contains(err.Error(), "send request") {
			t.Fatalf("Send(canceled) error = %v", err)
		}
	})

	t.Run("response body read failure", func(t *testing.T) {
		client := New("https://example.com", WithHTTPClient(&http.Client{
			Transport: dingtalkRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(errorReader{}),
					Request:    req,
				}, nil
			}),
		}))
		err := client.Send(context.Background(), NewText("x"))
		if err == nil || !strings.Contains(err.Error(), "read response body") {
			t.Fatalf("Send(read body failure) error = %v", err)
		}
	})
}

func TestPayloadValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{name: "empty text", message: NewText("")},
		{name: "empty markdown title", message: NewMarkdown("", "text")},
		{name: "empty markdown text", message: NewMarkdown("title", "")},
		{name: "empty link title", message: NewLink("", "text", "https://example.com")},
		{name: "empty link text", message: NewLink("title", "", "https://example.com")},
		{name: "empty link url", message: NewLink("title", "text", "")},
		{name: "nil typed text", message: (*Text)(nil)},
		{name: "nil typed markdown", message: (*Markdown)(nil)},
		{name: "nil typed link", message: (*Link)(nil)},
		{name: "nil typed action card", message: (*ActionCard)(nil)},
		{name: "nil typed feed card", message: (*FeedCard)(nil)},
		{
			name: "empty action card title",
			message: NewSingleActionCard(
				"",
				"text",
				"go",
				"https://example.com",
			),
		},
		{
			name: "empty action card text",
			message: NewSingleActionCard(
				"title",
				"",
				"go",
				"https://example.com",
			),
		},
		{name: "empty feed card", message: NewFeedCard(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.message.Payload()
			if !errors.Is(err, ErrInvalidMessage) {
				t.Fatalf("Payload() error = %v", err)
			}
		})
	}
}

func TestPayloadFields(t *testing.T) {
	text := NewText("hello")
	text.At = At{Mobiles: []string{"138"}, UserIDs: []string{"u1"}, All: true}
	data, err := text.Payload()
	if err != nil {
		t.Fatalf("Text Payload() error = %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal(text) error = %v", err)
	}
	if got["msgtype"] != "text" || got["at"] == nil {
		t.Fatalf("text payload = %s", data)
	}

	card := NewMultiActionCard(
		"title",
		"text",
		[]ActionCardButton{{Title: "go", ActionURL: "https://example.com"}},
	)
	card.ButtonVertical = true
	data, err = card.Payload()
	if err != nil {
		t.Fatalf("ActionCard Payload() error = %v", err)
	}
	if !strings.Contains(string(data), `"btnOrientation":"1"`) {
		t.Fatalf("vertical card payload = %s", data)
	}
}

func TestConstructorsCopySlices(t *testing.T) {
	buttons := []ActionCardButton{{Title: "old", ActionURL: "https://example.com"}}
	card := NewMultiActionCard("title", "text", buttons)
	buttons[0].Title = "new"
	if card.Buttons[0].Title != "old" {
		t.Fatalf("NewMultiActionCard did not copy buttons: %+v", card.Buttons)
	}

	links := []FeedLink{{
		Title:      "old",
		MessageURL: "https://example.com",
		PicURL:     "https://example.com/a.png",
	}}
	feed := NewFeedCard(links)
	links[0].Title = "new"
	if feed.Links[0].Title != "old" {
		t.Fatalf("NewFeedCard did not copy links: %+v", feed.Links)
	}
}

type dingtalkRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn dingtalkRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}
