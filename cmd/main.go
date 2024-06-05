package main

import (
	"gows"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var hub = gows.NewHub()

func main() {
	slog.SetDefault(gows.Logger)
	mux := http.NewServeMux()

	go hub.Run()

	mux.HandleFunc("GET /", hc)
	mux.HandleFunc("GET /ws", upgradeHandler)

	gows.Logger.Info("Running server...", "PORT", 9999)
	err := http.ListenAndServe(":9999", mux)
	if err != nil {
		log.Fatalln("Error while listening server", "ERROR", err)
	}
}

func hc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func upgradeHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		gows.Logger.Error("Error while upgrading client connection", "ERROR", err)
		return
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
			gows.Logger.Error("Error while reading message", "ERROR", err)
			hub.Unregister <- &user
			conn.Close()
			break
		}

		gows.Logger.Info("Message received", "MESSAGE", string(msg), "USER", user.Name)
		hub.Rooms[user.RoomName].Message <- gows.NewMessage(string(msg), user.Name)
	}
}
