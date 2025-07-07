package handler

import (
	"log"
	"mainPackage/model"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var userConnections = make(map[string]*websocket.Conn)
var connMutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler godoc
// @Summary WebSocket endpoint for real-time notifications by username
// @Description Opens a WebSocket connection and listens for a username from the client to register for real-time notifications.
// @Tags Notifications
// @Param username query string true "Username used to establish WebSocket connection"
// @Success 101 "Switching Protocols (WebSocket Upgrade)"
// @Failure 400 {object} map[string]string "Missing or invalid username query parameter"
// @Failure 500 {object} map[string]string "Internal server error or WebSocket upgrade failed"
// @Router /api/v1/notifications/ws [get]
func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Read initial message from client to get username
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Failed to read username:", err)
		return
	}
	username := strings.TrimSpace(string(msg)) // ตัดช่องว่างรอบข้าง
	if username == "" {
		log.Println("Empty username from client")
		return
	}

	log.Println("✅ WebSocket connected, username:", username)

	connMutex.Lock()
	userConnections[username] = conn
	log.Printf("🌐 Current connected users: %v\n", getUsernames())
	connMutex.Unlock()

	defer func() {
		connMutex.Lock()
		delete(userConnections, username)
		log.Printf("❌ WebSocket disconnected: %s\n", username)
		log.Printf("🌐 Current connected users: %v\n", getUsernames())
		connMutex.Unlock()
	}()

	// Keep the connection alive or handle ping/pong
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Println("Read error for user", username, ":", err)
			break
		}
	}
}

// helper function to get usernames from map
func getUsernames() []string {
	keys := make([]string, 0, len(userConnections))
	for k := range userConnections {
		keys = append(keys, k)
	}
	return keys
}

func SendNotificationToRecipient(noti model.Notification) {
	connMutex.Lock()
	defer connMutex.Unlock()

	if noti.Recipient == "all" {
		// Broadcast to all connected users
		for username, conn := range userConnections {
			err := conn.WriteJSON(noti)
			if err != nil {
				log.Println("❌ Failed to send to", username, ":", err)
				conn.Close()
				delete(userConnections, username)
			}
		}
		return
	}

	// Send to a specific recipient by username
	conn, exists := userConnections[noti.Recipient]
	if !exists {
		log.Println("🔕 Recipient", noti.Recipient, "is not connected")
		return
	}

	err := conn.WriteJSON(noti)
	if err != nil {
		log.Println("❌ Failed to send to", noti.Recipient, ":", err)
		conn.Close()
		delete(userConnections, noti.Recipient)
	}
}
