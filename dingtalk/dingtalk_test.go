package dingtalk

import "testing"

const (
	accessToken = ""
	secret      = ""
)

func TestSendTextMsg(t *testing.T) {
	if accessToken == "" {
		t.Skip("access token is empty")
	}
	msg := NewTextMsg("Hello, DingTalk!")
	err := NewRobot(accessToken).WithSecret(secret).Send(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestSendMarkdownMsg(t *testing.T) {
	if accessToken == "" {
		t.Skip("access token is empty")
	}
	msg := NewMarkdownMsg("Hi", "## Hello\nWorld")
	err := NewRobot(accessToken).WithSecret(secret).Send(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestSendLinkMsg(t *testing.T) {
	if accessToken == "" {
		t.Skip("access token is empty")
	}
	msg := NewLinkMsg("Title", "Description", "https://example.com").
		WithPicURL("https://example.com/image.png")
	err := NewRobot(accessToken).WithSecret(secret).Send(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestSendSingleActionCardMsg(t *testing.T) {
	if accessToken == "" {
		t.Skip("access token is empty")
	}
	msg := NewSingleActionCard("Title", "Text", "Click me", "https://example.com")
	err := NewRobot(accessToken).WithSecret(secret).Send(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestSendMultiActionCardMsg(t *testing.T) {
	if accessToken == "" {
		t.Skip("access token is empty")
	}
	msg := NewMultiActionCard("Title", "Text", []ActionCardBtn{
		{Title: "Button1", ActionURL: "https://example.com/1"},
		{Title: "Button2", ActionURL: "https://example.com/2"},
	}).WithBtnOrientation(BtnOrientationVertical)
	err := NewRobot(accessToken).WithSecret(secret).Send(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestSendFeedCardMsg(t *testing.T) {
	if accessToken == "" {
		t.Skip("access token is empty")
	}
	msg := NewFeedCardMsg([]FeedLink{
		{Title: "Link1", MessageURL: "https://example.com/1", PicURL: "https://example.com/1.png"},
		{Title: "Link2", MessageURL: "https://example.com/2", PicURL: "https://example.com/2.png"},
	})
	err := NewRobot(accessToken).WithSecret(secret).Send(msg)
	if err != nil {
		t.Error(err)
	}
}
