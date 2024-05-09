package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

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
	log.Println("Running websocket pooling")
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

type Hub struct {
	Rooms      map[string]*Room
	Register   chan User
	Unregister chan *User
}

func (h *Hub) Run() {
	for {
		select {
		case user := <-h.Register:
			// check if room is not exists
			if _, ok := h.Rooms[user.RoomName]; !ok {
				room := NewRoom(user.RoomName)

				h.Rooms[user.RoomName] = &room
				go h.Rooms[user.RoomName].Run()
			}

			h.Rooms[user.RoomName].Enter <- user
			// TODO: check if user is already in the room
		case user := <-h.Unregister:
			h.Rooms[user.RoomName].Leave <- user

			// if room is empty, delete it
			if len(h.Rooms[user.RoomName].Users) == 0 {
				delete(h.Rooms, user.RoomName)
			}
		}
	}
}

func NewHub() Hub {
	return Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan User),
		Unregister: make(chan *User),
	}
}

var hub = NewHub()

func main() {
	mux := http.NewServeMux()

	go hub.Run()

	mux.HandleFunc("GET /ws", upgradeHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func upgradeHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		w.Write([]byte(err.Error()))
	}

	roomName := r.URL.Query().Get("room")
	name := r.URL.Query().Get("name")

	user := User{
		Id:       time.Now().String(),
		Name:     name,
		RoomName: roomName,
		Conn:     conn,
	}

	// add user to room and add into room
	hub.Register <- user
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("error while reading message:", err)
			hub.Unregister <- &user
			conn.Close()
			break
		}

		log.Printf("getting message from %s: %s", user.Name, string(msg))

		hub.Rooms[user.RoomName].Message <- NewMessage(string(msg), user.Name)
	}
}
