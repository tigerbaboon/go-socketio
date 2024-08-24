package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	hub := newHub()
	go hub.run()

	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File("home.html")
	})
	r.GET("/ws", func(c *gin.Context) {
		serveWs(hub, c)
	})

	r.GET("/users", func(c *gin.Context) {
		hub.mu.Lock()
		defer hub.mu.Unlock()
		users := make([]string, 0, len(hub.users))
		for username := range hub.users {
			users = append(users, username)
		}
		c.JSON(http.StatusOK, gin.H{"users": users})
	})

	r.Run(":8080")
}
