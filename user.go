package gows

import (
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	Id       string
	Name     string
	RoomName string
	Conn     *websocket.Conn
}

type Message struct {
	Content string
	Sender  string
	SendAt  time.Time
}

func NewMessage(content string, sender string) Message {
	return Message{
		Content: content,
		Sender:  sender,
		SendAt:  time.Now(),
	}
}
