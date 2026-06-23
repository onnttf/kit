package dingtalk

import (
	"encoding/json"
	"slices"
)

const (
	MessageTypeText = "text"

	MessageTypeMarkdown = "markdown"

	MessageTypeLink = "link"

	MessageTypeActionCard = "actionCard"

	MessageTypeFeedCard = "feedCard"
)

const (
	OrientationHorizontal = "0"

	OrientationVertical = "1"
)

type Message interface {
	Payload() ([]byte, error)
}

type At struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

type TextMessage struct {
	Type string `json:"msgtype"`
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
	At At `json:"at"`
}

func NewTextMessage(content string) *TextMessage {
	m := &TextMessage{Type: MessageTypeText}
	m.Text.Content = content
	return m
}

func (m *TextMessage) WithAtMobiles(mobiles []string) *TextMessage {
	m.At.AtMobiles = slices.Clone(mobiles)
	return m
}

func (m *TextMessage) WithIsAtAll(isAll bool) *TextMessage {
	m.At.IsAtAll = isAll
	return m
}

func (m *TextMessage) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type MarkdownMessage struct {
	Type     string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	At At `json:"at"`
}

func NewMarkdownMessage(title, text string) *MarkdownMessage {
	m := &MarkdownMessage{Type: MessageTypeMarkdown}
	m.Markdown.Title = title
	m.Markdown.Text = text
	return m
}

func (m *MarkdownMessage) WithAtMobiles(mobiles []string) *MarkdownMessage {
	m.At.AtMobiles = slices.Clone(mobiles)
	return m
}

func (m *MarkdownMessage) WithIsAtAll(isAll bool) *MarkdownMessage {
	m.At.IsAtAll = isAll
	return m
}

func (m *MarkdownMessage) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type LinkMessage struct {
	Type string `json:"msgtype"`
	Link struct {
		Title      string `json:"title"`
		Text       string `json:"text"`
		PicURL     string `json:"picUrl,omitempty"`
		MessageURL string `json:"messageURL"`
	} `json:"link"`
}

func NewLinkMessage(title, text, messageURL string) *LinkMessage {
	m := &LinkMessage{Type: MessageTypeLink}
	m.Link.Title = title
	m.Link.Text = text
	m.Link.MessageURL = messageURL
	return m
}

func (m *LinkMessage) WithPicURL(url string) *LinkMessage {
	m.Link.PicURL = url
	return m
}

func (m *LinkMessage) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type ActionCardButton struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

type ActionCardMessage struct {
	Type       string `json:"msgtype"`
	ActionCard struct {
		Title       string             `json:"title"`
		Text        string             `json:"text"`
		SingleTitle string             `json:"singleTitle,omitempty"`
		SingleURL   string             `json:"singleURL,omitempty"`
		Orientation string             `json:"btnOrientation,omitempty"`
		Btns        []ActionCardButton `json:"btns,omitempty"`
	} `json:"actionCard"`
}

func NewSingleActionCardMessage(title, text, singleTitle, singleURL string) *ActionCardMessage {
	m := &ActionCardMessage{Type: MessageTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.SingleTitle = singleTitle
	m.ActionCard.SingleURL = singleURL
	return m
}

func NewMultiActionCardMessage(title, text string, btns []ActionCardButton) *ActionCardMessage {
	m := &ActionCardMessage{Type: MessageTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.Btns = slices.Clone(btns)
	return m
}

func (m *ActionCardMessage) WithButtonOrientation(orientation string) *ActionCardMessage {
	if orientation == OrientationHorizontal || orientation == OrientationVertical {
		m.ActionCard.Orientation = orientation
	}
	return m
}

func (m *ActionCardMessage) Payload() ([]byte, error) {
	return json.Marshal(m)
}

type FeedLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL"`
}

type FeedCardMessage struct {
	Type     string `json:"msgtype"`
	FeedCard struct {
		Links []FeedLink `json:"links"`
	} `json:"feedCard"`
}

func NewFeedCardMessage(links []FeedLink) *FeedCardMessage {
	m := &FeedCardMessage{Type: MessageTypeFeedCard}
	m.FeedCard.Links = slices.Clone(links)
	return m
}

func (m *FeedCardMessage) Payload() ([]byte, error) {
	return json.Marshal(m)
}

var (
	_ Message = (*TextMessage)(nil)
	_ Message = (*MarkdownMessage)(nil)
	_ Message = (*LinkMessage)(nil)
	_ Message = (*ActionCardMessage)(nil)
	_ Message = (*FeedCardMessage)(nil)
)
