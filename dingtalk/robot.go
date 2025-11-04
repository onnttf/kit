package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Robot represents the client for sending messages to DingTalk.
type Robot struct {
	accessToken string
	secret      string
	httpClient  *http.Client
}

// NewRobot creates a Robot instance with the given access token.
func NewRobot(accessToken string) *Robot {
	defaultClient := &http.Client{Timeout: 5 * time.Second}
	return &Robot{accessToken: accessToken, httpClient: defaultClient}
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

// calculateSign generates the DingTalk message signature.
func (r *Robot) calculateSign(timestamp int64) (string, error) {
	if r.secret == "" {
		return "", fmt.Errorf("secret is empty, cannot calculate sign")
	}
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, r.secret)
	h := hmac.New(sha256.New, []byte(r.secret))
	h.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return url.QueryEscape(sign), nil
}

// Send posts the message payload to DingTalk.
func (r *Robot) Send(msg Message) error {
	if r.accessToken == "" {
		return fmt.Errorf("dingtalk: send: accessToken is empty")
	}
	if r.httpClient == nil {
		return fmt.Errorf("dingtalk: send: httpClient is nil")
	}
	if msg == nil {
		return fmt.Errorf("dingtalk: send: message is nil")
	}

	payload, err := msg.GetPayload()
	if err != nil {
		return fmt.Errorf("dingtalk: send: marshal message failed: %w", err)
	}
	if len(payload) == 0 {
		return fmt.Errorf("dingtalk: send: payload is empty")
	}

	timestamp := time.Now().UnixMilli()
	webhookURL := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", r.accessToken)
	if r.secret != "" {
		sign, err := r.calculateSign(timestamp)
		if err != nil {
			return fmt.Errorf("dingtalk: send: calculate sign failed: %w", err)
		}
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("dingtalk: send: create HTTP request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dingtalk: send: HTTP post failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("dingtalk: send: read response failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk: send: HTTP status error: status=%s, body=%s", resp.Status, string(body))
	}

	var dingResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &dingResp); err != nil {
		return fmt.Errorf("dingtalk: send: unmarshal response failed: %w, body=%s", err, string(body))
	}
	if dingResp.ErrCode != 0 {
		return fmt.Errorf("dingtalk: send: api returned error: errcode=%d, errmsg=%s", dingResp.ErrCode, dingResp.ErrMsg)
	}
	return nil
}
