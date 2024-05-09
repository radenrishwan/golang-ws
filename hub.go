package gows

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
