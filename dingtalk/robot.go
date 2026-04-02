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

var defaultClientOnce sync.Once
var defaultClient *http.Client

func getDefaultClient() *http.Client {
	defaultClientOnce.Do(func() {
		defaultClient = &http.Client{Timeout: 5 * time.Second}
	})
	return defaultClient
}

// Robot is the client for sending messages to DingTalk.
type Robot struct {
	accessToken string
	secret      string
	httpClient  *http.Client
}

// NewRobot creates a Robot instance with the given access token.
//
// Example:
//
//	robot := dingtalk.NewRobot("your-access-token")
func NewRobot(accessToken string) *Robot {
	return &Robot{accessToken: accessToken, httpClient: getDefaultClient()}
}

// WithSecret sets the secret for signature verification.
//
// Example:
//
//	robot := dingtalk.NewRobot("token").WithSecret("SEC...")
func (r *Robot) WithSecret(secret string) *Robot {
	r.secret = secret
	return r
}

// WithClient sets a custom HTTP client.
//
// Example:
//
//	robot := dingtalk.NewRobot("token").WithClient(&http.Client{Timeout: 10*time.Second})
func (r *Robot) WithClient(client *http.Client) *Robot {
	if client != nil {
		r.httpClient = client
	}
	return r
}

// Send posts the message payload to DingTalk using a default context with 5 second timeout.
// For custom timeout or cancellation control, use SendWithContext instead.
//
// Example:
//
//	robot.Send(dingtalk.NewTextMsg("Hello"))
func (r *Robot) Send(msg Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.SendWithContext(ctx, msg)
}

// SendWithContext posts the message payload to DingTalk.
// The context controls the entire HTTP request lifecycle.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	err := robot.SendWithContext(ctx, msg)
func (r *Robot) SendWithContext(ctx context.Context, msg Message) error {
	if r.accessToken == "" {
		return errors.New("access token is empty")
	}
	if r.httpClient == nil {
		return errors.New("http client is nil")
	}
	if msg == nil {
		return errors.New("message is nil")
	}

	payload, err := msg.Payload()
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	if len(payload) == 0 {
		return errors.New("payload is empty")
	}

	timestamp := time.Now().UnixMilli()
	webhookURL := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", r.accessToken)
	if r.secret != "" {
		sign, err := r.calculateSign(timestamp)
		if err != nil {
			return fmt.Errorf("calculate sign: %w", err)
		}
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var dingResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &dingResp); err != nil {
		return fmt.Errorf("unmarshal response: %w, body=%s", err, string(body))
	}
	if dingResp.ErrCode != 0 {
		return fmt.Errorf("unexpected response: errcode=%d, errmsg=%s", dingResp.ErrCode, dingResp.ErrMsg)
	}
	return nil
}

// calculateSign generates the DingTalk message signature.
func (r *Robot) calculateSign(timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, r.secret)
	h := hmac.New(sha256.New, []byte(r.secret))
	h.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return url.QueryEscape(sign), nil
}
