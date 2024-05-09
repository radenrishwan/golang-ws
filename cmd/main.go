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
	Id   string
	Name string
	Conn *websocket.Conn
}

type Room struct {
	Name    string
	Users   map[string]*User
	Enter   chan *User
	Leave   chan *User
	Message chan string
}

func NewRoom() Room {
	return Room{
		Users:   make(map[string]*User),
		Enter:   make(chan *User),
		Leave:   make(chan *User),
		Message: make(chan string),
	}
}

var rooms = NewRoom()

func Pool() {
	log.Println("Running websocket pooling")
	for {
		select {
		case user := <-rooms.Enter:
			rooms.Users[user.Id] = user

			log.Println("an user joined the room")
			for _, user := range rooms.Users {
				user.Conn.WriteMessage(websocket.TextMessage, []byte(user.Id+" joined the room"))
			}
		case user := <-rooms.Leave:
			delete(rooms.Users, user.Id)
			for _, user := range rooms.Users {
				user.Conn.WriteMessage(websocket.TextMessage, []byte(user.Id+" left the room"))
			}

			user.Conn.Close()
		case msg := <-rooms.Message:
			for _, user := range rooms.Users {
				user.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
			}

			log.Println(msg)
		}
	}
}

func main() {
	mux := http.NewServeMux()

	go Pool()

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

	conn.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))

	user := User{
		Id:   time.Now().String(),
		Name: time.Now().String(),
		Conn: conn,
	}

	rooms.Users[user.Id] = &user
	rooms.Enter <- &user
	log.Println("current user on room : ", len(rooms.Users))
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("error while reading message:", err)
			rooms.Leave <- &user
			conn.Close()
			break
		}

		log.Printf("getting message from %s: %s\n", user.Id, string(msg))

		rooms.Message <- string(msg)
	}
}
