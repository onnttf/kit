package dingtalk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var (
	ErrInvalidInput = errors.New("dingtalk: invalid input")
	ErrUnexpected   = errors.New("dingtalk: unexpected response")
)

type Client struct {
	webhook    string
	secret     string
	httpClient *http.Client
}

type Option func(*Client)

var defaultClient = sync.OnceValue(func() *http.Client {
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: defaultTransport(),
	}
})

func New(webhook string, opts ...Option) *Client {
	base := *defaultClient()
	c := &Client{webhook: webhook, httpClient: &base}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

func WithSecret(secret string) Option {
	return func(c *Client) {
		c.secret = secret
	}
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

func (c *Client) Send(ctx context.Context, msg Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.webhook == "" {
		return fmt.Errorf("%w: empty webhook", ErrInvalidInput)
	}
	if msg == nil {
		return fmt.Errorf("%w: nil message", ErrInvalidInput)
	}

	body, err := msg.Payload()
	if err != nil {
		return err
	}

	endpoint, err := c.endpoint()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("dingtalk: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dingtalk: send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if err != nil {
		return fmt.Errorf("dingtalk: read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: status=%q body=%q", ErrUnexpected, resp.Status, string(data))
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("dingtalk: decode response: %w", err)
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("%w: code=%d message=%q", ErrUnexpected, result.ErrCode, result.ErrMsg)
	}
	return nil
}

func (c *Client) endpoint() (string, error) {
	if c.secret == "" {
		return c.webhook, nil
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sign, err := sign(timestamp, c.secret)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(c.webhook)
	if err != nil {
		return "", fmt.Errorf("%w: invalid webhook", ErrInvalidInput)
	}
	q := u.Query()
	q.Set("timestamp", timestamp)
	q.Set("sign", sign)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func sign(timestamp, secret string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(timestamp + "\n" + secret)); err != nil {
		return "", fmt.Errorf("dingtalk: sign message: %w", err)
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
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
