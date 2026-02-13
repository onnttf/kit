package dingtalk

import "encoding/json"

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

// Message is the interface implemented by all DingTalk messages.
type Message interface {
	GetPayload() ([]byte, error)
}

// At specifies users to mention (@).
type At struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

// TextMsg is a simple plain text message.
type TextMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At At `json:"at"`
}

// NewTextMsg creates a TextMsg instance with the specified content.
//
// Example:
//
//	msg := dingtalk.NewTextMsg("Hello, world!")
func NewTextMsg(content string) *TextMsg {
	m := &TextMsg{MsgType: MsgTypeText}
	m.Text.Content = content
	return m
}

// WithAtMobiles sets the mobile numbers to mention.
//
// Example:
//
//	msg := dingtalk.NewTextMsg("Hello").WithAtMobiles([]string{"13800138000"})
func (m *TextMsg) WithAtMobiles(mobiles []string) *TextMsg {
	m.At.AtMobiles = mobiles
	return m
}

// WithIsAtAll sets whether to mention all members.
//
// Example:
//
//	msg := dingtalk.NewTextMsg("Hello").WithIsAtAll(true)
func (m *TextMsg) WithIsAtAll(isAll bool) *TextMsg {
	m.At.IsAtAll = isAll
	return m
}

func (m *TextMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}

// MarkdownMsg is a rich Markdown format message.
type MarkdownMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	At At `json:"at"`
}

// NewMarkdownMsg creates a MarkdownMsg instance with the required title and text content.
//
// Example:
//
//	msg := dingtalk.NewMarkdownMsg("Title", "## Hello\nWorld")
func NewMarkdownMsg(title, text string) *MarkdownMsg {
	m := &MarkdownMsg{MsgType: MsgTypeMarkdown}
	m.Markdown.Title = title
	m.Markdown.Text = text
	return m
}

// WithAtMobiles sets the mobile numbers to mention.
func (m *MarkdownMsg) WithAtMobiles(mobiles []string) *MarkdownMsg {
	m.At.AtMobiles = mobiles
	return m
}

// WithIsAtAll sets whether to mention all members.
func (m *MarkdownMsg) WithIsAtAll(isAll bool) *MarkdownMsg {
	m.At.IsAtAll = isAll
	return m
}

func (m *MarkdownMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}

// LinkMsg is a simple link message card.
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
//
// Example:
//
//	msg := dingtalk.NewLinkMsg("Title", "Description", "https://example.com")
func NewLinkMsg(title, text, messageURL string) *LinkMsg {
	m := &LinkMsg{MsgType: MsgTypeLink}
	m.Link.Title = title
	m.Link.Text = text
	m.Link.MessageURL = messageURL
	return m
}

// WithPicURL sets the optional picture URL to display on the card.
//
// Example:
//
//	msg := dingtalk.NewLinkMsg("Title", "Desc", "https://example.com").WithPicURL("https://example.com/image.png")
func (m *LinkMsg) WithPicURL(url string) *LinkMsg {
	m.Link.PicURL = url
	return m
}

func (m *LinkMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}

// ActionCardBtn is a clickable button within an ActionCardMsg.
type ActionCardBtn struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

// ActionCardMsg is a message card that can contain one or multiple action buttons.
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
//
// Example:
//
//	msg := dingtalk.NewSingleActionCard("Title", "Text", "Click", "https://example.com")
func NewSingleActionCard(title, text, singleTitle, singleURL string) *ActionCardMsg {
	m := &ActionCardMsg{MsgType: MsgTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.SingleTitle = singleTitle
	m.ActionCard.SingleURL = singleURL
	return m
}

// NewMultiActionCard creates an ActionCardMsg that uses multiple buttons.
//
// Example:
//
//	msg := dingtalk.NewMultiActionCard("Title", "Text", []dingtalk.ActionCardBtn{
//	    {Title: "Button1", ActionURL: "https://example.com/1"},
//	    {Title: "Button2", ActionURL: "https://example.com/2"},
//	})
func NewMultiActionCard(title, text string, btns []ActionCardBtn) *ActionCardMsg {
	m := &ActionCardMsg{MsgType: MsgTypeActionCard}
	m.ActionCard.Title = title
	m.ActionCard.Text = text
	m.ActionCard.Btns = btns
	return m
}

// WithBtnOrientation sets button orientation.
//
// Example:
//
//	msg := dingtalk.NewMultiActionCard("Title", "Text", btns).WithBtnOrientation(dingtalk.BtnOrientationVertical)
func (m *ActionCardMsg) WithBtnOrientation(orientation string) *ActionCardMsg {
	if orientation == BtnOrientationHorizontal || orientation == BtnOrientationVertical {
		m.ActionCard.BtnOrientation = orientation
	}
	return m
}

func (m *ActionCardMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}

// FeedLink is a single item in a feed card message.
type FeedLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL"`
}

// FeedCardMsg is a message card that displays a list of links in a feed format.
type FeedCardMsg struct {
	MsgType  string `json:"msgtype"`
	FeedCard struct {
		Links []FeedLink `json:"links"`
	} `json:"feedCard"`
}

// NewFeedCardMsg creates a FeedCardMsg instance with the provided links.
//
// Example:
//
//	msg := dingtalk.NewFeedCardMsg([]dingtalk.FeedLink{
//	    {Title: "Link1", MessageURL: "https://example.com/1", PicURL: "https://example.com/1.png"},
//	})
func NewFeedCardMsg(links []FeedLink) *FeedCardMsg {
	m := &FeedCardMsg{MsgType: MsgTypeFeedCard}
	m.FeedCard.Links = links
	return m
}

func (m *FeedCardMsg) GetPayload() ([]byte, error) {
	return json.Marshal(m)
}
