package gows

import (
	"log"

	"github.com/gorilla/websocket"
)

type Room struct {
	Name    string
	Users   map[string]*User
	Enter   chan User
	Leave   chan *User
	Message chan Message
}

func NewRoom(name string) Room {
	return Room{
		Name:    name,
		Users:   make(map[string]*User),
		Enter:   make(chan User),
		Leave:   make(chan *User),
		Message: make(chan Message),
	}
}

// websocket pooling
func (r *Room) Run() {
	Logger.Info("Running room pooling", "NAME", r.Name)
	for {
		select {
		case user := <-r.Enter:
			r.Users[user.Id] = &user

			for _, usr := range r.Users {
				usr.Conn.WriteMessage(websocket.TextMessage, []byte(user.Name+" joined the room"))
			}
		case user := <-r.Leave:
			delete(r.Users, user.Id)
			for _, usr := range r.Users {
				usr.Conn.WriteMessage(websocket.TextMessage, []byte(user.Name+" left the room"))
			}

			user.Conn.Close()
		case msg := <-r.Message:
			for _, user := range r.Users {
				user.Conn.WriteMessage(websocket.TextMessage, []byte(msg.Sender+": "+msg.Content))
			}

			log.Println(msg)
		}
	}
}
