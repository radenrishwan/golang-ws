package main

import (
	"gows"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var hub = gows.NewHub()

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

	user := gows.User{
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

		hub.Rooms[user.RoomName].Message <- gows.NewMessage(string(msg), user.Name)
	}
}
