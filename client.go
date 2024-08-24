package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	username string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// websocket handler
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(hub *Hub, c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username is required"})
		return
	}

	hub.mu.Lock()
	if _, ok := hub.users[username]; ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken"})
		hub.mu.Unlock()
		return
	}
	hub.mu.Unlock()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), username: username}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
