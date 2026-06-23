package dingtalk

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTextMessage(t *testing.T) {
	msg := NewTextMessage("Hello World")

	assert.Equal(t, MessageTypeText, msg.Type)
	assert.Equal(t, "Hello World", msg.Text.Content)
}

func TestTextMessage_WithAtMobiles(t *testing.T) {
	msg := NewTextMessage("Hello")
	mobiles := []string{"13800138000", "13900139000"}
	msg.WithAtMobiles(mobiles)
	mobiles[0] = "changed"

	assert.Equal(t, []string{"13800138000", "13900139000"}, msg.At.AtMobiles)
	assert.False(t, msg.At.IsAtAll)
}

func TestTextMessage_WithIsAtAll(t *testing.T) {
	msg := NewTextMessage("Hello")
	msg.WithIsAtAll(true)

	assert.True(t, msg.At.IsAtAll)
}

func TestTextMessage_Payload(t *testing.T) {
	msg := NewTextMessage("Hello")
	msg.WithAtMobiles([]string{"13800138000"})

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MessageTypeText, result["msgtype"])
	assert.Equal(t, "Hello", result["text"].(map[string]any)["content"])
}

func TestNewMarkdownMessage(t *testing.T) {
	msg := NewMarkdownMessage("Title", "## Content")

	assert.Equal(t, MessageTypeMarkdown, msg.Type)
	assert.Equal(t, "Title", msg.Markdown.Title)
	assert.Equal(t, "## Content", msg.Markdown.Text)
}

func TestMarkdownMessage_WithAtMobiles(t *testing.T) {
	msg := NewMarkdownMessage("Title", "Content")
	mobiles := []string{"13800138000"}
	msg.WithAtMobiles(mobiles)
	mobiles[0] = "changed"

	assert.Equal(t, []string{"13800138000"}, msg.At.AtMobiles)
}

func TestMarkdownMessage_WithIsAtAll(t *testing.T) {
	msg := NewMarkdownMessage("Title", "Content")
	msg.WithIsAtAll(true)

	assert.True(t, msg.At.IsAtAll)
}

func TestMarkdownMessage_Payload(t *testing.T) {
	msg := NewMarkdownMessage("Title", "Content")

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MessageTypeMarkdown, result["msgtype"])
	assert.Equal(t, "Title", result["markdown"].(map[string]any)["title"])
}

func TestNewLinkMessage(t *testing.T) {
	msg := NewLinkMessage("Title", "Description", "https://example.com")

	assert.Equal(t, MessageTypeLink, msg.Type)
	assert.Equal(t, "Title", msg.Link.Title)
	assert.Equal(t, "Description", msg.Link.Text)
	assert.Equal(t, "https://example.com", msg.Link.MessageURL)
}

func TestLinkMessage_WithPicURL(t *testing.T) {
	msg := NewLinkMessage("Title", "Description", "https://example.com")
	msg.WithPicURL("https://example.com/pic.jpg")

	assert.Equal(t, "https://example.com/pic.jpg", msg.Link.PicURL)
}

func TestLinkMessage_Payload(t *testing.T) {
	msg := NewLinkMessage("Title", "Description", "https://example.com")

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MessageTypeLink, result["msgtype"])
	assert.Equal(t, "Title", result["link"].(map[string]any)["title"])
}

func TestNewSingleActionCardMessage(t *testing.T) {
	msg := NewSingleActionCardMessage("Title", "Text", "Click Me", "https://example.com/click")

	assert.Equal(t, MessageTypeActionCard, msg.Type)
	assert.Equal(t, "Title", msg.ActionCard.Title)
	assert.Equal(t, "Text", msg.ActionCard.Text)
	assert.Equal(t, "Click Me", msg.ActionCard.SingleTitle)
	assert.Equal(t, "https://example.com/click", msg.ActionCard.SingleURL)
}

func TestNewMultiActionCardMessage(t *testing.T) {
	btns := []ActionCardButton{
		{Title: "Button1", ActionURL: "https://example.com/1"},
		{Title: "Button2", ActionURL: "https://example.com/2"},
	}

	msg := NewMultiActionCardMessage("Title", "Text", btns)
	btns[0].Title = "Changed"

	assert.Equal(t, MessageTypeActionCard, msg.Type)
	assert.Equal(t, "Title", msg.ActionCard.Title)
	assert.Equal(t, "Text", msg.ActionCard.Text)
	assert.Len(t, msg.ActionCard.Btns, 2)
	assert.Equal(t, "Button1", msg.ActionCard.Btns[0].Title)
}

func TestActionCardMessage_WithButtonOrientation(t *testing.T) {
	msg := NewSingleActionCardMessage("Title", "Text", "Click", "https://example.com")

	msg.WithButtonOrientation(OrientationHorizontal)
	assert.Equal(t, OrientationHorizontal, msg.ActionCard.Orientation)

	msg.WithButtonOrientation(OrientationVertical)
	assert.Equal(t, OrientationVertical, msg.ActionCard.Orientation)

	msg.WithButtonOrientation("invalid")
	assert.Equal(t, OrientationVertical, msg.ActionCard.Orientation)
}

func TestActionCardMessage_Payload(t *testing.T) {
	msg := NewSingleActionCardMessage("Title", "Text", "Click", "https://example.com")

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MessageTypeActionCard, result["msgtype"])
}

func TestNewFeedCardMessage(t *testing.T) {
	links := []FeedLink{
		{Title: "Link1", MessageURL: "https://example.com/1", PicURL: "https://example.com/pic1.jpg"},
		{Title: "Link2", MessageURL: "https://example.com/2", PicURL: "https://example.com/pic2.jpg"},
	}

	msg := NewFeedCardMessage(links)
	links[0].Title = "Changed"

	assert.Equal(t, MessageTypeFeedCard, msg.Type)
	assert.Len(t, msg.FeedCard.Links, 2)
	assert.Equal(t, "Link1", msg.FeedCard.Links[0].Title)
}

func TestFeedCardMessage_Payload(t *testing.T) {
	links := []FeedLink{
		{Title: "Link1", MessageURL: "https://example.com/1", PicURL: "https://example.com/pic1.jpg"},
	}

	msg := NewFeedCardMessage(links)

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MessageTypeFeedCard, result["msgtype"])
}
