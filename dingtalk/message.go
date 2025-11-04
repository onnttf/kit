package dingtalk

import "encoding/json"

// DingTalk message types.
const (
	MsgTypeText       = "text"
	MsgTypeMarkdown   = "markdown"
	MsgTypeLink       = "link"
	MsgTypeActionCard = "actionCard"
	MsgTypeFeedCard   = "feedCard"
)

// Button orientation for ActionCardMsg.
const (
	BtnOrientationHorizontal = "0" // Horizontal button layout.
	BtnOrientationVertical   = "1" // Vertical button layout.
)

// Message represents the interface implemented by all DingTalk messages.
type Message interface {
	GetPayload() ([]byte, error)
}

// At represents the block specifying users to mention (@).
type At struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

// TextMsg represents a simple plain text message type.
type TextMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At At `json:"at"`
}

// NewTextMsg creates a TextMsg instance with the specified content.
func NewTextMsg(content string) *TextMsg {
	m := &TextMsg{MsgType: MsgTypeText}
	m.Text.Content = content
	return m
}

func (m *TextMsg) WithAtMobiles(mobiles []string) *TextMsg {
	m.At.AtMobiles = mobiles
	return m
}

func (m *TextMsg) WithIsAtAll(isAll bool) *TextMsg {
	m.At.IsAtAll = isAll
	return m
}

func (m *TextMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}

// MarkdownMsg represents a rich Markdown format message type.
type MarkdownMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	At At `json:"at"`
}

// NewMarkdownMsg creates a MarkdownMsg instance with the required title and text content.
func NewMarkdownMsg(title, text string) *MarkdownMsg {
	m := &MarkdownMsg{MsgType: MsgTypeMarkdown}
	m.Markdown.Title = title
	m.Markdown.Text = text
	return m
}

func (m *MarkdownMsg) WithAtMobiles(mobiles []string) *MarkdownMsg {
	m.At.AtMobiles = mobiles
	return m
}

func (m *MarkdownMsg) WithIsAtAll(isAll bool) *MarkdownMsg {
	m.At.IsAtAll = isAll
	return m
}

func (m *MarkdownMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}

// LinkMsg represents a simple link message card type.
type LinkMsg struct {
	MsgType string `json:"msgtype"`
	Link    struct {
		Title      string `json:"title"`
		Text       string `json:"text"`
		PicURL     string `json:"picUrl,omitempty"`
		MessageURL string `json:"messageURL"`
	} `json:"link"`
}

// NewLinkMsg creates a LinkMsg instance with the required title, text, and destination URL.
func NewLinkMsg(title, text, messageURL string) *LinkMsg {
	m := &LinkMsg{MsgType: MsgTypeLink}
	m.Link.Title = title
	m.Link.Text = text
	m.Link.MessageURL = messageURL
	return m
}

// WithPicURL sets the optional picture URL to display on the card.
func (m *LinkMsg) WithPicURL(url string) *LinkMsg {
	m.Link.PicURL = url
	return m
}

func (m *LinkMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}

// ActionCardBtn represents a clickable button within an ActionCardMsg.
type ActionCardBtn struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

// ActionCardMsg represents a message card that can contain one or multiple action buttons.
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

// NewSingleActionCard creates an ActionCardMsg that uses a single action link.
func NewSingleActionCard(title, text, singleTitle, singleURL string) *ActionCardMsg {
	m := &ActionCardMsg{MsgType: MsgTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.SingleTitle = singleTitle
	m.ActionCard.SingleURL = singleURL
	return m
}

// NewMultiActionCard creates an ActionCardMsg that uses multiple buttons.
func NewMultiActionCard(title, text string, btns []ActionCardBtn) *ActionCardMsg {
	m := &ActionCardMsg{MsgType: MsgTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.Btns = btns
	return m
}

// WithBtnOrientation sets button orientation.
func (m *ActionCardMsg) WithBtnOrientation(orientation string) *ActionCardMsg {
	if orientation == BtnOrientationHorizontal || orientation == BtnOrientationVertical {
		m.ActionCard.BtnOrientation = orientation
	}
	return m
}

func (m *ActionCardMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}

// FeedLink represents a single item (link) in a feed card message type.
type FeedLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL"`
}

// FeedCardMsg represents a message card that displays a list of links in a feed format.
type FeedCardMsg struct {
	MsgType  string `json:"msgtype"`
	FeedCard struct {
		Links []FeedLink `json:"links"`
	} `json:"feedCard"`
}

// NewFeedCardMsg creates a FeedCardMsg instance with the provided links.
func NewFeedCardMsg(links []FeedLink) *FeedCardMsg {
	m := &FeedCardMsg{MsgType: MsgTypeFeedCard}
	m.FeedCard.Links = links
	return m
}

func (m *FeedCardMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}
