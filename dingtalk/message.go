package dingtalk

import (
	"encoding/json"
	"slices"
)

const (
	// MsgTypeText is the DingTalk message type for plain text messages.
	MsgTypeText = "text"
	// MsgTypeMarkdown is the DingTalk message type for markdown messages.
	MsgTypeMarkdown = "markdown"
	// MsgTypeLink is the DingTalk message type for link cards.
	MsgTypeLink = "link"
	// MsgTypeActionCard is the DingTalk message type for action cards.
	MsgTypeActionCard = "actionCard"
	// MsgTypeFeedCard is the DingTalk message type for feed cards.
	MsgTypeFeedCard = "feedCard"
)

const (
	// BtnOrientationHorizontal lays action-card buttons out horizontally.
	BtnOrientationHorizontal = "0"

	// BtnOrientationVertical lays action-card buttons out vertically.
	BtnOrientationVertical = "1"
)

// Message is implemented by DingTalk robot message payloads.
type Message interface {
	Payload() ([]byte, error)
}

// At describes users mentioned by a text or markdown message.
type At struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

// TextMsg is a DingTalk plain text message.
type TextMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At At `json:"at"`
}

// NewTextMsg creates a plain text message.
func NewTextMsg(content string) *TextMsg {
	m := &TextMsg{MsgType: MsgTypeText}
	m.Text.Content = content
	return m
}

// WithAtMobiles sets the phone numbers mentioned by the message.
func (m *TextMsg) WithAtMobiles(mobiles []string) *TextMsg {
	m.At.AtMobiles = slices.Clone(mobiles)
	return m
}

// WithIsAtAll sets whether the message mentions all members.
func (m *TextMsg) WithIsAtAll(isAll bool) *TextMsg {
	m.At.IsAtAll = isAll
	return m
}

// Payload marshals the message as DingTalk JSON.
func (m *TextMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

// MarkdownMsg is a DingTalk markdown message.
type MarkdownMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	At At `json:"at"`
}

// NewMarkdownMsg creates a markdown message.
func NewMarkdownMsg(title, text string) *MarkdownMsg {
	m := &MarkdownMsg{MsgType: MsgTypeMarkdown}
	m.Markdown.Title = title
	m.Markdown.Text = text
	return m
}

// WithAtMobiles sets the phone numbers mentioned by the message.
func (m *MarkdownMsg) WithAtMobiles(mobiles []string) *MarkdownMsg {
	m.At.AtMobiles = slices.Clone(mobiles)
	return m
}

// WithIsAtAll sets whether the message mentions all members.
func (m *MarkdownMsg) WithIsAtAll(isAll bool) *MarkdownMsg {
	m.At.IsAtAll = isAll
	return m
}

// Payload marshals the message as DingTalk JSON.
func (m *MarkdownMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

// LinkMsg is a DingTalk link card message.
type LinkMsg struct {
	MsgType string `json:"msgtype"`
	Link    struct {
		Title      string `json:"title"`
		Text       string `json:"text"`
		PicURL     string `json:"picUrl,omitempty"`
		MessageURL string `json:"messageURL"`
	} `json:"link"`
}

// NewLinkMsg creates a link card message.
func NewLinkMsg(title, text, messageURL string) *LinkMsg {
	m := &LinkMsg{MsgType: MsgTypeLink}
	m.Link.Title = title
	m.Link.Text = text
	m.Link.MessageURL = messageURL
	return m
}

// WithPicURL sets the optional picture URL for the link card.
func (m *LinkMsg) WithPicURL(url string) *LinkMsg {
	m.Link.PicURL = url
	return m
}

// Payload marshals the message as DingTalk JSON.
func (m *LinkMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

// ActionCardBtn describes one button in a multi-action card.
type ActionCardBtn struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

// ActionCardMsg is a DingTalk action-card message.
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

// NewSingleActionCard creates an action card with a single action.
func NewSingleActionCard(title, text, singleTitle, singleURL string) *ActionCardMsg {
	m := &ActionCardMsg{MsgType: MsgTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.SingleTitle = singleTitle
	m.ActionCard.SingleURL = singleURL
	return m
}

// NewMultiActionCard creates an action card with multiple buttons.
func NewMultiActionCard(title, text string, btns []ActionCardBtn) *ActionCardMsg {
	m := &ActionCardMsg{MsgType: MsgTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.Btns = slices.Clone(btns)
	return m
}

// WithBtnOrientation sets the button layout when orientation is a known value.
func (m *ActionCardMsg) WithBtnOrientation(orientation string) *ActionCardMsg {
	if orientation == BtnOrientationHorizontal || orientation == BtnOrientationVertical {
		m.ActionCard.BtnOrientation = orientation
	}
	return m
}

// Payload marshals the message as DingTalk JSON.
func (m *ActionCardMsg) Payload() ([]byte, error) {
	return json.Marshal(m)
}

// FeedLink describes one link in a feed-card message.
type FeedLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL"`
}

// FeedCardMsg is a DingTalk feed-card message.
type FeedCardMsg struct {
	MsgType  string `json:"msgtype"`
	FeedCard struct {
		Links []FeedLink `json:"links"`
	} `json:"feedCard"`
}

// NewFeedCardMsg creates a feed-card message.
func NewFeedCardMsg(links []FeedLink) *FeedCardMsg {
	m := &FeedCardMsg{MsgType: MsgTypeFeedCard}
	m.FeedCard.Links = slices.Clone(links)
	return m
}

// Payload marshals the message as DingTalk JSON.
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
