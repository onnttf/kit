package dingtalk

import (
	"encoding/json"
	"slices"
)

const (
	MsgTypeText       = "text"
	MsgTypeMarkdown   = "markdown"
	MsgTypeLink       = "link"
	MsgTypeActionCard = "actionCard"
	MsgTypeFeedCard   = "feedCard"
)

const (
	BtnOrientationHorizontal = "0"
	BtnOrientationVertical   = "1"
)

// Message is implemented by DingTalk robot message payloads.
type Message interface {
	Payload() ([]byte, error)
}

type At struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

type TextMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At At `json:"at"`
}

func NewTextMsg(content string) *TextMsg {
	m := &TextMsg{MsgType: MsgTypeText}
	m.Text.Content = content
	return m
}

func (m *TextMsg) WithAtMobiles(mobiles []string) *TextMsg {
	m.At.AtMobiles = slices.Clone(mobiles)
	return m
}

func (m *TextMsg) WithIsAtAll(isAll bool) *TextMsg {
	m.At.IsAtAll = isAll
	return m
}

func (m *TextMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type MarkdownMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	At At `json:"at"`
}

func NewMarkdownMsg(title, text string) *MarkdownMsg {
	m := &MarkdownMsg{MsgType: MsgTypeMarkdown}
	m.Markdown.Title = title
	m.Markdown.Text = text
	return m
}

func (m *MarkdownMsg) WithAtMobiles(mobiles []string) *MarkdownMsg {
	m.At.AtMobiles = slices.Clone(mobiles)
	return m
}

func (m *MarkdownMsg) WithIsAtAll(isAll bool) *MarkdownMsg {
	m.At.IsAtAll = isAll
	return m
}

func (m *MarkdownMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type LinkMsg struct {
	MsgType string `json:"msgtype"`
	Link    struct {
		Title      string `json:"title"`
		Text       string `json:"text"`
		PicURL     string `json:"picUrl,omitempty"`
		MessageURL string `json:"messageURL"`
	} `json:"link"`
}

func NewLinkMsg(title, text, messageURL string) *LinkMsg {
	m := &LinkMsg{MsgType: MsgTypeLink}
	m.Link.Title = title
	m.Link.Text = text
	m.Link.MessageURL = messageURL
	return m
}

func (m *LinkMsg) WithPicURL(url string) *LinkMsg {
	m.Link.PicURL = url
	return m
}

func (m *LinkMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type ActionCardBtn struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

type ActionCardMsg struct {
	MsgType    string `json:"msgtype"`
	ActionCard struct {
		Title          string          `json:"title"`
		Text           string          `json:"text"`
		SingleTitle    string          `json:"singleTitle,omitempty"`
		SingleURL      string          `json:"singleURL,omitempty"`
		BtnOrientation string          `json:"btnOrientation,omitempty"`
		Btns           []ActionCardBtn `json:"btns,omitempty"`
	} `json:"actionCard"`
}

func NewSingleActionCard(title, text, singleTitle, singleURL string) *ActionCardMsg {
	m := &ActionCardMsg{MsgType: MsgTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.SingleTitle = singleTitle
	m.ActionCard.SingleURL = singleURL
	return m
}

func NewMultiActionCard(title, text string, btns []ActionCardBtn) *ActionCardMsg {
	m := &ActionCardMsg{MsgType: MsgTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.Btns = slices.Clone(btns)
	return m
}

func (m *ActionCardMsg) WithBtnOrientation(orientation string) *ActionCardMsg {
	if orientation == BtnOrientationHorizontal || orientation == BtnOrientationVertical {
		m.ActionCard.BtnOrientation = orientation
	}
	return m
}

func (m *ActionCardMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type FeedLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL"`
}

type FeedCardMsg struct {
	MsgType  string `json:"msgtype"`
	FeedCard struct {
		Links []FeedLink `json:"links"`
	} `json:"feedCard"`
}

func NewFeedCardMsg(links []FeedLink) *FeedCardMsg {
	m := &FeedCardMsg{MsgType: MsgTypeFeedCard}
	m.FeedCard.Links = slices.Clone(links)
	return m
}

func (m *FeedCardMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

var (
	_ Message = (*TextMsg)(nil)
	_ Message = (*MarkdownMsg)(nil)
	_ Message = (*LinkMsg)(nil)
	_ Message = (*ActionCardMsg)(nil)
	_ Message = (*FeedCardMsg)(nil)
)
