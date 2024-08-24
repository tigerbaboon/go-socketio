package main

import (
	"encoding/json"
	"log"
	"sync"
)

type Message struct {
	Type      string `json:"type"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	users      map[string]*Client
	rooms      map[string]map[*Client]bool
	mu         sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		users:      make(map[string]*Client),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.users[client.username] = client
			h.broadcastUserList()
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.users, client.username)
				close(client.send)
				h.broadcastUserList()
			}
		case message := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println("Error unmarshalling message:", err)
				continue
			}
			if msg.Type == "message" {
				h.mu.Lock()
				if recipientClient, ok := h.users[msg.Recipient]; ok {
					recipientClient.send <- message
				}
				h.mu.Unlock()
			}
		}
	}
}

func (h *Hub) broadcastUserList() {
	h.mu.Lock()
	defer h.mu.Unlock()
	users := make([]string, 0, len(h.users))
	for username := range h.users {
		users = append(users, username)
	}
	userList, _ := json.Marshal(map[string]interface{}{"type": "userList", "users": users})
	for client := range h.clients {
		client.send <- userList
	}
}
