package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

// --- WebSocket Connection Management ---

var (
	// userConnections stores active connections in memory. Key is the user's Employee ID (empId).
	userConnections = make(map[string]*model.UserConnectionInfo)
	// connMutex protects concurrent access to the userConnections map.
	connMutex = &sync.Mutex{}
)

// upgrader upgrades HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development purposes.
		// For production, implement proper origin checks.
		return true
	},
}

// getUserProfileFromDB fetches user details from the database to populate the connection info.
func getUserProfileFromDB(ctx context.Context, dbConn *pgx.Conn, orgId, username string) (*model.UserConnectionInfo, error) {
	log.Printf("Database: Querying for user '%s' in organization '%s'", username, orgId)

	var userProfile model.UserConnectionInfo
	var roleID string

	query := `
		SELECT "empId", "username", "orgId", "deptId", "commId", "stnId", "roleId" 
		FROM um_users 
		WHERE "orgId" = $1 AND "username" = $2 AND "active" = true 
		LIMIT 1
	`
	err := dbConn.QueryRow(ctx, query, orgId, username).Scan(
		&userProfile.ID, &userProfile.Username, &userProfile.OrgID,
		&userProfile.DeptID, &userProfile.CommID, &userProfile.StnID,
		&roleID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found or is not active")
		}
		log.Printf("ERROR: Failed to query user profile for '%s': %v", username, err)
		return nil, err
	}

	userProfile.RoleID = roleID // ใช้ string ตรง ๆ ไม่ใช่ slice
	return &userProfile, nil
}

// upsertUserConnectionToDB inserts or updates a user's connection status in the database.
func upsertUserConnectionToDB(userInfo *model.UserConnectionInfo) error {
	dbConn, ctx, cancel := config.ConnectDB()
	if dbConn == nil {
		log.Println("ERROR: Failed to get DB connection for upsert.")
		return errors.New("could not connect to database")
	}
	defer cancel()
	defer dbConn.Close(ctx)

	query := `
	INSERT INTO user_connections ("empId", "username", "orgId", "deptId", "commId", "stnId", "roleId", "connectedAt")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT ("empId") DO UPDATE SET
		"username" = EXCLUDED."username",
		"orgId" = EXCLUDED."orgId",
		"deptId" = EXCLUDED."deptId",
		"commId" = EXCLUDED."commId",
		"stnId" = EXCLUDED."stnId",
		"roleId" = EXCLUDED."roleId",
		"connectedAt" = EXCLUDED."connectedAt";
	`

	_, err := dbConn.Exec(ctx, query,
		userInfo.ID, userInfo.Username, userInfo.OrgID,
		userInfo.DeptID, userInfo.CommID, userInfo.StnID,
		userInfo.RoleID, time.Now())

	if err != nil {
		log.Printf("ERROR: Failed to upsert user connection to DB for EmpID %s: %v", userInfo.ID, err)
	} else {
		log.Printf("Database: Successfully upserted connection for EmpID %s", userInfo.ID)
	}
	return err
}

// removeUserConnectionFromDB removes a user's connection status from the database.
func removeUserConnectionFromDB(userID string) {
	dbConn, ctx, cancel := config.ConnectDB()
	if dbConn == nil {
		log.Printf("ERROR: Failed to get DB connection to remove user %s", userID)
		return
	}
	defer cancel()
	defer dbConn.Close(ctx)

	_, err := dbConn.Exec(ctx, `DELETE FROM user_connections WHERE "empId" = $1`, userID)
	if err != nil {
		log.Printf("ERROR: Failed to remove user connection from DB for EmpID %s: %v", userID, err)
	} else {
		log.Printf("Database: Successfully removed connection for EmpID %s", userID)
	}
}

// WebSocketHandler godoc
// @Summary WebSocket endpoint for real-time notifications
// @Description Establishes a WebSocket connection. The client must send a JSON message with `orgId` and `username` to register the session.
// @Tags Notifications
// @security ApiKeyAuth
// @Success 101 "Switching Protocols"
// @Failure 400 "Bad Request (invalid registration message)"
// @Failure 404 "Not Found (User not found)"
// @Failure 500 "Internal Server Error"
// @Router /api/v1/notifications/register [get]
func WebSocketHandler(c *gin.Context) {
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer wsConn.Close()

	// 1. Read initial registration message from client.
	_, msg, err := wsConn.ReadMessage()
	if err != nil {
		log.Println("Failed to read registration message:", err)
		return
	}

	var regMsg model.RegistrationMessage
	if err := json.Unmarshal(msg, &regMsg); err != nil {
		log.Println("Invalid registration message format:", err)
		wsConn.WriteJSON(gin.H{"error": "invalid registration format"})
		return
	}

	dbConn, ctx, cancel := config.ConnectDB()
	if dbConn == nil {
		wsConn.WriteJSON(gin.H{"error": "could not connect to the database"})
		return
	}

	connInfo, err := getUserProfileFromDB(ctx, dbConn, regMsg.OrgID, regMsg.Username)

	dbConn.Close(ctx)
	cancel()

	if err != nil {
		log.Printf("User registration failed for '%s': %v", regMsg.Username, err)
		wsConn.WriteJSON(gin.H{"error": "user not found or invalid credentials"})
		return
	}
	connInfo.Conn = wsConn

	connMutex.Lock()
	userConnections[connInfo.ID] = connInfo
	connMutex.Unlock()

	go upsertUserConnectionToDB(connInfo)

	defer func() {
		connMutex.Lock()
		delete(userConnections, connInfo.ID)
		connMutex.Unlock()

		go removeUserConnectionFromDB(connInfo.ID)
		log.Printf("❌ Disconnected: EmpID=%s", connInfo.ID)
	}()

	// Keep-alive loop: exit when connection closes.
	for {
		if _, _, err := wsConn.ReadMessage(); err != nil {
			break
		}
	}
}

// BroadcastNotification sends a notification to relevant connected users.
func BroadcastNotification(noti model.Notification) {
	connMutex.Lock()
	defer connMutex.Unlock()

	log.Printf("📢 Broadcasting notification ID: %s", noti.ID)
	sentTo := make(map[string]bool)

	for _, connInfo := range userConnections {
		if sentTo[connInfo.ID] {
			continue
		}
		if connInfo.OrgID != noti.OrgID {
			continue
		}

		isTrueBroadcast := noti.Recipients == nil || len(noti.Recipients) == 0

		if isTrueBroadcast {
			log.Printf("  🚀 True broadcast! Sending to EmpID: %s (OrgID: %s)", connInfo.ID, connInfo.OrgID)
			if err := connInfo.Conn.WriteJSON(noti); err != nil {
				log.Printf("    ❌ Failed to send to EmpID %s: %v", connInfo.ID, err)
			}
			sentTo[connInfo.ID] = true
			continue
		}

		for _, recipient := range noti.Recipients {
			shouldReceive := false

			switch strings.ToLower(recipient.Type) {
			case "orgid":
				shouldReceive = connInfo.OrgID == recipient.Value
			case "empid":
				shouldReceive = connInfo.ID == recipient.Value
			case "roleid":
				shouldReceive = connInfo.RoleID == recipient.Value
			case "deptid":
				shouldReceive = connInfo.DeptID == recipient.Value
			case "stnid":
				shouldReceive = connInfo.StnID == recipient.Value
			case "commid":
				shouldReceive = connInfo.CommID == recipient.Value
			case "username":
				shouldReceive = connInfo.Username == recipient.Value
			}

			if shouldReceive {
				log.Printf("  🚀 Match found! Sending to EmpID: %s (Rule: %s:%s)", connInfo.ID, recipient.Type, recipient.Value)

				payloadToSend := noti

				if err := connInfo.Conn.WriteJSON(payloadToSend); err != nil {
					log.Printf("    ❌ Failed to send to EmpID %s: %v", connInfo.ID, err)
				}
				sentTo[connInfo.ID] = true
				break
			}
		}
	}
	log.Printf("✅ Broadcasting finished for notification ID: %s", noti.ID)
}
