package dingtalk

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTextMsg(t *testing.T) {
	msg := NewTextMsg("Hello World")

	assert.Equal(t, MsgTypeText, msg.MsgType)
	assert.Equal(t, "Hello World", msg.Text.Content)
}

func TestTextMsg_WithAtMobiles(t *testing.T) {
	msg := NewTextMsg("Hello")
	msg.WithAtMobiles([]string{"13800138000", "13900139000"})

	assert.Equal(t, []string{"13800138000", "13900139000"}, msg.At.AtMobiles)
	assert.False(t, msg.At.IsAtAll)
}

func TestTextMsg_WithIsAtAll(t *testing.T) {
	msg := NewTextMsg("Hello")
	msg.WithIsAtAll(true)

	assert.True(t, msg.At.IsAtAll)
}

func TestTextMsg_Payload(t *testing.T) {
	msg := NewTextMsg("Hello")
	msg.WithAtMobiles([]string{"13800138000"})

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MsgTypeText, result["msgtype"])
	assert.Equal(t, "Hello", result["text"].(map[string]any)["content"])
}

func TestNewMarkdownMsg(t *testing.T) {
	msg := NewMarkdownMsg("Title", "## Content")

	assert.Equal(t, MsgTypeMarkdown, msg.MsgType)
	assert.Equal(t, "Title", msg.Markdown.Title)
	assert.Equal(t, "## Content", msg.Markdown.Text)
}

func TestMarkdownMsg_WithAtMobiles(t *testing.T) {
	msg := NewMarkdownMsg("Title", "Content")
	msg.WithAtMobiles([]string{"13800138000"})

	assert.Equal(t, []string{"13800138000"}, msg.At.AtMobiles)
}

func TestMarkdownMsg_WithIsAtAll(t *testing.T) {
	msg := NewMarkdownMsg("Title", "Content")
	msg.WithIsAtAll(true)

	assert.True(t, msg.At.IsAtAll)
}

func TestMarkdownMsg_Payload(t *testing.T) {
	msg := NewMarkdownMsg("Title", "Content")

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MsgTypeMarkdown, result["msgtype"])
	assert.Equal(t, "Title", result["markdown"].(map[string]any)["title"])
}

func TestNewLinkMsg(t *testing.T) {
	msg := NewLinkMsg("Title", "Description", "https://example.com")

	assert.Equal(t, MsgTypeLink, msg.MsgType)
	assert.Equal(t, "Title", msg.Link.Title)
	assert.Equal(t, "Description", msg.Link.Text)
	assert.Equal(t, "https://example.com", msg.Link.MessageURL)
}

func TestLinkMsg_WithPicURL(t *testing.T) {
	msg := NewLinkMsg("Title", "Description", "https://example.com")
	msg.WithPicURL("https://example.com/pic.jpg")

	assert.Equal(t, "https://example.com/pic.jpg", msg.Link.PicURL)
}

func TestLinkMsg_Payload(t *testing.T) {
	msg := NewLinkMsg("Title", "Description", "https://example.com")

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MsgTypeLink, result["msgtype"])
	assert.Equal(t, "Title", result["link"].(map[string]any)["title"])
}

func TestNewSingleActionCard(t *testing.T) {
	msg := NewSingleActionCard("Title", "Text", "Click Me", "https://example.com/click")

	assert.Equal(t, MsgTypeActionCard, msg.MsgType)
	assert.Equal(t, "Title", msg.ActionCard.Title)
	assert.Equal(t, "Text", msg.ActionCard.Text)
	assert.Equal(t, "Click Me", msg.ActionCard.SingleTitle)
	assert.Equal(t, "https://example.com/click", msg.ActionCard.SingleURL)
}

func TestNewMultiActionCard(t *testing.T) {
	btns := []ActionCardBtn{
		{Title: "Button1", ActionURL: "https://example.com/1"},
		{Title: "Button2", ActionURL: "https://example.com/2"},
	}

	msg := NewMultiActionCard("Title", "Text", btns)

	assert.Equal(t, MsgTypeActionCard, msg.MsgType)
	assert.Equal(t, "Title", msg.ActionCard.Title)
	assert.Equal(t, "Text", msg.ActionCard.Text)
	assert.Len(t, msg.ActionCard.Btns, 2)
}

func TestActionCardMsg_WithBtnOrientation(t *testing.T) {
	msg := NewSingleActionCard("Title", "Text", "Click", "https://example.com")

	msg.WithBtnOrientation(BtnOrientationHorizontal)
	assert.Equal(t, BtnOrientationHorizontal, msg.ActionCard.BtnOrientation)

	msg.WithBtnOrientation(BtnOrientationVertical)
	assert.Equal(t, BtnOrientationVertical, msg.ActionCard.BtnOrientation)

	msg.WithBtnOrientation("invalid")
	assert.Equal(t, BtnOrientationVertical, msg.ActionCard.BtnOrientation)
}

func TestActionCardMsg_Payload(t *testing.T) {
	msg := NewSingleActionCard("Title", "Text", "Click", "https://example.com")

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MsgTypeActionCard, result["msgtype"])
}

func TestNewFeedCardMsg(t *testing.T) {
	links := []FeedLink{
		{Title: "Link1", MessageURL: "https://example.com/1", PicURL: "https://example.com/pic1.jpg"},
		{Title: "Link2", MessageURL: "https://example.com/2", PicURL: "https://example.com/pic2.jpg"},
	}

	msg := NewFeedCardMsg(links)

	assert.Equal(t, MsgTypeFeedCard, msg.MsgType)
	assert.Len(t, msg.FeedCard.Links, 2)
	assert.Equal(t, "Link1", msg.FeedCard.Links[0].Title)
}

func TestFeedCardMsg_Payload(t *testing.T) {
	links := []FeedLink{
		{Title: "Link1", MessageURL: "https://example.com/1", PicURL: "https://example.com/pic1.jpg"},
	}

	msg := NewFeedCardMsg(links)

	payload, err := msg.Payload()
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, MsgTypeFeedCard, result["msgtype"])
}

func TestMessageInterface(t *testing.T) {
	var _ Message = NewTextMsg("test")
	var _ Message = NewMarkdownMsg("title", "text")
	var _ Message = NewLinkMsg("title", "text", "url")
	var _ Message = NewSingleActionCard("title", "text", "btn", "url")
	var _ Message = NewMultiActionCard("title", "text", nil)
	var _ Message = NewFeedCardMsg(nil)
}

func TestMsgTypeConstants(t *testing.T) {
	assert.Equal(t, "text", MsgTypeText)
	assert.Equal(t, "markdown", MsgTypeMarkdown)
	assert.Equal(t, "link", MsgTypeLink)
	assert.Equal(t, "actionCard", MsgTypeActionCard)
	assert.Equal(t, "feedCard", MsgTypeFeedCard)
}

func TestBtnOrientationConstants(t *testing.T) {
	assert.Equal(t, "0", BtnOrientationHorizontal)
	assert.Equal(t, "1", BtnOrientationVertical)
}