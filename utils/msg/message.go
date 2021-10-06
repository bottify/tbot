package msg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const kMsgType_Text = "text"
const kMsgType_Image = "image"

type MessageData struct {
	Text interface{} `json:"text,omitempty"`
	File interface{} `json:"file,omitempty"`
}

type MessageSegment struct {
	Type string      `json:"type"`
	Data MessageData `json:"data"`
}

type Message struct {
	Segments []MessageSegment
}

func New() *Message {
	return &Message{}
}

func (m *Message) Text(text string) *Message {
	m.Segments = append(m.Segments, MessageSegment{
		Type: kMsgType_Text,
		Data: MessageData{
			Text: text,
		},
	})
	return m
}

func (m *Message) Image(filename string) *Message {
	m.Segments = append(m.Segments, MessageSegment{
		Type: kMsgType_Image,
		Data: MessageData{
			File: filename,
		},
	})
	return m
}

func (m *Message) ImageBytes(rawdata []byte) *Message {
	data := base64.StdEncoding.EncodeToString(rawdata)
	return m.Image(fmt.Sprintf("base64://%s", data))
}

func (m *Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Segments)
}
