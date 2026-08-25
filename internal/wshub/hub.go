package wshub

import (
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type HubStruct struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan interface{}
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	mu         sync.Mutex
}

var Hub *HubStruct

func InitWSHub() {
	Hub = &HubStruct{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
	go Hub.run()
}

func (h *HubStruct) run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Println("New WebSocket client connected")
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Println("WebSocket client disconnected")
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteJSON(message)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastMessage sends a JSON object to all connected clients
func BroadcastMessage(msg interface{}) {
	if Hub != nil {
		Hub.broadcast <- msg
	}
}
