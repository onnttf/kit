package dingtalk

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRobot(t *testing.T) {
	robot := NewRobot("test_token")
	assert.Equal(t, "test_token", robot.accessToken)
	assert.NotNil(t, robot.httpClient)
}

func TestNewRobot_DefaultTransportClonesHTTPDefaultTransport(t *testing.T) {
	robot := NewRobot("test_token")

	transport, ok := robot.httpClient.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.NotSame(t, http.DefaultTransport, transport)
	assert.NotNil(t, transport.Proxy)
	assert.NotNil(t, transport.DialContext)
	assert.Equal(t, 100, transport.MaxIdleConnsPerHost)
}

func TestRobot_WithSecret(t *testing.T) {
	robot := NewRobot("test_token")
	result := robot.WithSecret("test_secret")
	assert.Equal(t, "test_secret", robot.secret)
	assert.Same(t, robot, result)
}

func TestRobot_WithClient(t *testing.T) {
	robot := NewRobot("test_token")
	customClient := &http.Client{Timeout: 10 * time.Second}
	result := robot.WithClient(customClient)
	assert.Equal(t, customClient, robot.httpClient)
	assert.Same(t, robot, result)
}

func TestRobot_WithClient_NilClient(t *testing.T) {
	robot := NewRobot("test_token")
	originalClient := robot.httpClient
	result := robot.WithClient(nil)
	assert.Same(t, originalClient, robot.httpClient)
	assert.Same(t, robot, result)
}

func TestRobot_SendWithContext_EmptyToken(t *testing.T) {
	robot := NewRobot("")
	err := robot.SendWithContext(context.Background(), NewTextMsg("Hello"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access token is empty")
}

func TestRobot_SendWithContext_NilClient(t *testing.T) {
	robot := NewRobot("test_token")
	robot.httpClient = nil
	err := robot.SendWithContext(context.Background(), NewTextMsg("Hello"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http client is nil")
}

func TestRobot_SendWithContext_NilMessage(t *testing.T) {
	robot := NewRobot("test_token")
	err := robot.SendWithContext(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message is nil")
}

func TestRobot_SendWithContext_Success(t *testing.T) {
	var gotMethod, gotContentType string
	var gotBody []byte

	robot := NewRobot("test_token").WithClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotMethod = req.Method
			gotContentType = req.Header.Get("Content-Type")
			var err error
			gotBody, err = io.ReadAll(req.Body)
			assert.NoError(t, err)
			return jsonResponse(http.StatusOK, `{"errcode":0,"errmsg":"ok"}`), nil
		}),
	})

	err := robot.SendWithContext(context.Background(), NewTextMsg("Hello"))

	assert.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "application/json;charset=utf-8", gotContentType)
	assert.Contains(t, string(gotBody), `"msgtype":"text"`)
	assert.Contains(t, string(gotBody), `"content":"Hello"`)
}

func TestRobot_SendWithContext_DingTalkError(t *testing.T) {
	robot := NewRobot("test_token").WithClient(&http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{"errcode":310000,"errmsg":"keywords not in content"}`), nil
		}),
	})

	err := robot.SendWithContext(context.Background(), NewTextMsg("Hello"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "errcode=310000")
}

func TestRobot_SendWithContext_HTTPErrorIncludesBody(t *testing.T) {
	robot := NewRobot("test_token").WithClient(&http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusBadGateway, `bad gateway`), nil
		}),
	})

	err := robot.SendWithContext(context.Background(), NewTextMsg("Hello"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status=502")
	assert.Contains(t, err.Error(), "bad gateway")
}

func TestRobot_CalculateSign(t *testing.T) {
	robot := NewRobot("test_token")
	robot.secret = "test_secret"
	sign, err := robot.calculateSign(1234567890000)
	assert.NoError(t, err)
	assert.NotEmpty(t, sign)
}

func TestMessagePayload(t *testing.T) {
	msg := NewTextMsg("Hello")
	payload, err := msg.Payload()
	assert.NoError(t, err)
	assert.NotEmpty(t, payload)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
