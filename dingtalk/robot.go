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
	"sync"
	"time"
)

var (
	// ErrUnexpectedStatus indicates that DingTalk returned a non-200 HTTP status.
	ErrUnexpectedStatus = errors.New("unexpected http status")
	// ErrUnexpectedResponse indicates that DingTalk returned a non-zero error code.
	ErrUnexpectedResponse = errors.New("unexpected response")
)

var getDefaultClient = sync.OnceValue(func() *http.Client {
	return &http.Client{
		Timeout:   5 * time.Second,
		Transport: defaultTransport(),
	}
})

// Robot sends messages to a DingTalk robot webhook.
type Robot struct {
	accessToken string
	secret      string
	httpClient  *http.Client
}

func NewRobot(accessToken string) *Robot {
	return &Robot{accessToken: accessToken, httpClient: getDefaultClient()}
}

func (r *Robot) WithSecret(secret string) *Robot {
	r.secret = secret
	return r
}

func (r *Robot) WithClient(client *http.Client) *Robot {
	if client != nil {
		r.httpClient = client
	}
	return r
}

// Send posts msg using a background context with the default timeout.
func (r *Robot) Send(msg Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.SendWithContext(ctx, msg)
}

// A nil context is treated as context.Background.
func (r *Robot) SendWithContext(ctx context.Context, msg Message) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if r.accessToken == "" {
		return errors.New("send dingtalk message: access token is empty")
	}
	if r.httpClient == nil {
		return errors.New("send dingtalk message: http client is nil")
	}
	if msg == nil {
		return errors.New("send dingtalk message: message is nil")
	}

	payload, err := msg.Payload()
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	if len(payload) == 0 {
		return errors.New("send dingtalk message: payload is empty")
	}

	timestamp := time.Now().UnixMilli()
	values := url.Values{}
	values.Set("access_token", r.accessToken)
	if r.secret != "" {
		sign, err := r.calculateSign(timestamp)
		if err != nil {
			return fmt.Errorf("calculate sign: %w", err)
		}
		values.Set("timestamp", fmt.Sprintf("%d", timestamp))
		values.Set("sign", sign)
	}
	webhookURL := "https://oapi.dingtalk.com/robot/send?" + values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close response body: %w", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status=%d", ErrUnexpectedStatus, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var dingResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &dingResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	if dingResp.ErrCode != 0 {
		return fmt.Errorf("%w: errcode=%d, errmsg=%s", ErrUnexpectedResponse, dingResp.ErrCode, dingResp.ErrMsg)
	}
	return nil
}

func (r *Robot) calculateSign(timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, r.secret)
	h := hmac.New(sha256.New, []byte(r.secret))
	h.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return sign, nil
}

func defaultTransport() *http.Transport {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		clone := transport.Clone()
		clone.MaxIdleConnsPerHost = 100
		return clone
	}
	return &http.Transport{MaxIdleConnsPerHost: 100}
}
