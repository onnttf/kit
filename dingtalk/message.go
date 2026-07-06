package dingtalk

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrInvalidMessage = errors.New("dingtalk: invalid message")

type Message interface {
	Payload() ([]byte, error)
}

type At struct {
	Mobiles []string `json:"atMobiles,omitempty"`
	UserIDs []string `json:"atUserIds,omitempty"`
	All     bool     `json:"isAtAll,omitempty"`
}

type Text struct {
	Content string
	At      At
}

func NewText(content string) *Text {
	return &Text{Content: content}
}

func (m *Text) Payload() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("%w: nil text message", ErrInvalidMessage)
	}
	if m.Content == "" {
		return nil, fmt.Errorf("%w: empty text content", ErrInvalidMessage)
	}
	return json.Marshal(struct {
		MsgType string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
		At At `json:"at,omitempty"`
	}{
		MsgType: "text",
		Text: struct {
			Content string `json:"content"`
		}{Content: m.Content},
		At: m.At,
	})
}

type Markdown struct {
	Title string
	Text  string
	At    At
}

func NewMarkdown(title, text string) *Markdown {
	return &Markdown{Title: title, Text: text}
}

func (m *Markdown) Payload() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("%w: nil markdown message", ErrInvalidMessage)
	}
	if m.Title == "" || m.Text == "" {
		return nil, fmt.Errorf("%w: empty markdown field", ErrInvalidMessage)
	}
	return json.Marshal(struct {
		MsgType  string `json:"msgtype"`
		Markdown struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		} `json:"markdown"`
		At At `json:"at,omitempty"`
	}{
		MsgType: "markdown",
		Markdown: struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		}{Title: m.Title, Text: m.Text},
		At: m.At,
	})
}

type Link struct {
	Title      string
	Text       string
	MessageURL string
	PicURL     string
}

func NewLink(title, text, messageURL string) *Link {
	return &Link{Title: title, Text: text, MessageURL: messageURL}
}

func (m *Link) Payload() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("%w: nil link message", ErrInvalidMessage)
	}
	if m.Title == "" || m.Text == "" || m.MessageURL == "" {
		return nil, fmt.Errorf("%w: empty link field", ErrInvalidMessage)
	}
	return json.Marshal(struct {
		MsgType string `json:"msgtype"`
		Link    Link   `json:"link"`
	}{MsgType: "link", Link: *m})
}

type ActionCardButton struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

type ActionCard struct {
	Title          string
	Text           string
	SingleTitle    string
	SingleURL      string
	Buttons        []ActionCardButton
	ButtonVertical bool
}

func NewSingleActionCard(title, text, singleTitle, singleURL string) *ActionCard {
	return &ActionCard{Title: title, Text: text, SingleTitle: singleTitle, SingleURL: singleURL}
}

func NewMultiActionCard(title, text string, buttons []ActionCardButton) *ActionCard {
	return &ActionCard{Title: title, Text: text, Buttons: append([]ActionCardButton(nil), buttons...)}
}

func (m *ActionCard) Payload() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("%w: nil action card message", ErrInvalidMessage)
	}
	if m.Title == "" || m.Text == "" {
		return nil, fmt.Errorf("%w: empty action card field", ErrInvalidMessage)
	}
	hideAvatar := "0"
	btnOrientation := "0"
	if m.ButtonVertical {
		btnOrientation = "1"
	}
	card := struct {
		Title           string             `json:"title"`
		Text            string             `json:"text"`
		SingleTitle     string             `json:"singleTitle,omitempty"`
		SingleURL       string             `json:"singleURL,omitempty"`
		Buttons         []ActionCardButton `json:"btns,omitempty"`
		HideAvatar      string             `json:"hideAvatar"`
		ButtonOriention string             `json:"btnOrientation"`
	}{
		Title:           m.Title,
		Text:            m.Text,
		SingleTitle:     m.SingleTitle,
		SingleURL:       m.SingleURL,
		Buttons:         m.Buttons,
		HideAvatar:      hideAvatar,
		ButtonOriention: btnOrientation,
	}
	return json.Marshal(struct {
		MsgType    string `json:"msgtype"`
		ActionCard any    `json:"actionCard"`
	}{MsgType: "actionCard", ActionCard: card})
}

type FeedLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL"`
}

type FeedCard struct {
	Links []FeedLink
}

func NewFeedCard(links []FeedLink) *FeedCard {
	return &FeedCard{Links: append([]FeedLink(nil), links...)}
}

func (m *FeedCard) Payload() ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("%w: nil feed card message", ErrInvalidMessage)
	}
	if len(m.Links) == 0 {
		return nil, fmt.Errorf("%w: empty feed links", ErrInvalidMessage)
	}
	return json.Marshal(struct {
		MsgType  string `json:"msgtype"`
		FeedCard struct {
			Links []FeedLink `json:"links"`
		} `json:"feedCard"`
	}{
		MsgType: "feedCard",
		FeedCard: struct {
			Links []FeedLink `json:"links"`
		}{Links: m.Links},
	})
}
