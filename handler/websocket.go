package handler

import (
	"log"
	"mainPackage/config"
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

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Failed to read username:", err)
		return
	}
	username := strings.TrimSpace(string(msg))
	if username == "" {
		log.Println("Empty username from client")
		return
	}

	log.Println("✅ WebSocket connected:", username)

	connMutex.Lock()
	userConnections[username] = conn
	log.Printf("🌐 Users connected: %v\n", getUsernames())
	connMutex.Unlock()

	go func() {
		unreadNotis, err := GetUnreadNotificationsByUsername(username)
		if err != nil {
			log.Println("❌ Failed to get unread notis:", err)
			return
		}
		for _, n := range unreadNotis {
			if err := conn.WriteJSON(n); err != nil {
				log.Println("❌ Failed to send noti to", username, ":", err)
				break
			}
		}
	}()

	defer func() {
		connMutex.Lock()
		delete(userConnections, username)
		log.Printf("❌ Disconnected: %s\n", username)
		connMutex.Unlock()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Println("Read error for", username, ":", err)
			break
		}
	}
}

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
		for username, conn := range userConnections {
			if err := conn.WriteJSON(noti); err != nil {
				log.Println("❌ Failed to send to", username, ":", err)
				conn.Close()
				delete(userConnections, username)
			}
		}
		return
	}

	conn, exists := userConnections[noti.Recipient]
	if !exists {
		log.Println("🔕", noti.Recipient, "is not connected")
		return
	}

	if err := conn.WriteJSON(noti); err != nil {
		log.Println("❌ Send failed to", noti.Recipient, ":", err)
		conn.Close()
		delete(userConnections, noti.Recipient)
	}
}

func GetUnreadNotificationsByUsername(username string) ([]model.Notification, error) {
	conn, ctx, cancel := config.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT id, "caseId", "caseType", "caseDetail", recipient, sender, message, "eventType", "createdAt", read, "redirectUrl"
		FROM notifications
		WHERE recipient = $1 AND read = false
		ORDER BY "createdAt" ASC
	`, username)
	if err != nil {
		log.Println("❌ Failed to query unread notifications:", err)
		return nil, err
	}
	defer rows.Close()

	var notis []model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.CaseID, &n.CaseType, &n.CaseDetail, &n.Recipient, &n.Sender, &n.Message, &n.EventType, &n.CreatedAt, &n.Read, &n.RedirectURL); err != nil {
			log.Println("❌ Scan error:", err)
			continue
		}
		notis = append(notis, n)
	}
	return notis, nil
}
