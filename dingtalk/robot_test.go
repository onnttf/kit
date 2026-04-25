package dingtalk

import (
	"context"
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