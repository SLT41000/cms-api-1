package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// --- WebSocket Connection Management ---

var (
	// เก็บ connection ที่ online อยู่ ณ ตอนนี้ key = empId
	userConnections = make(map[string]*model.UserConnectionInfo)
	connMutex       = &sync.Mutex{}
)

// อนุญาตทุก origin (โปรดปรับสำหรับ production)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ---------- Helpers: พื้นที่/จังหวัด ----------

func checkUserInProvince(userDistIdLists []string, provId string) bool {
	if len(userDistIdLists) == 0 {
		return false
	}

	dbConn, ctx, cancel := utils.ConnectDB()
	if dbConn == nil {
		return false
	}
	defer cancel()
	defer dbConn.Close(ctx)

	placeholders := make([]string, len(userDistIdLists))
	args := make([]interface{}, len(userDistIdLists)+1)
	args[0] = provId

	for i, distId := range userDistIdLists {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = distId
	}

	var count int
	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM area_districts 
		WHERE "provId" = $1 AND "distId" IN (%s)
	`, strings.Join(placeholders, ","))

	if err := dbConn.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		log.Printf("ERROR: Failed to check user in province %s: %v", provId, err)
		return false
	}
	return count > 0
}

func checkUserInDistrict(userDistIdLists []string, distId string) bool {
	for _, userDistId := range userDistIdLists {
		if userDistId == distId {
			return true
		}
	}
	return false
}

func contains(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

// ---------- Upsert Connection ----------

func upsertUserConnectionToDB(userInfo *model.UserConnectionInfo) error {
	dbConn, ctx, cancel := utils.ConnectDB()
	if dbConn == nil {
		log.Println("ERROR: Failed to get DB connection for upsert.")
		return errors.New("could not connect to database")
	}
	defer cancel()
	defer dbConn.Close(ctx)

	// หมายเหตุ: คอลัมน์ "grpId" และ "distIdLists" ใน user_connections ควรเป็น text[]/varchar[]
	query := `
    INSERT INTO user_connections ("empId", "username", "orgId", "deptId", "commId", "stnId", "roleId", "grpId", "distIdLists", "connectedAt","ip")
    VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10,NULLIF($11, ''))
    ON CONFLICT ("empId") DO UPDATE SET
        "username"    = EXCLUDED."username",
        "orgId"       = EXCLUDED."orgId",
        "deptId"      = EXCLUDED."deptId",
        "commId"      = EXCLUDED."commId",
        "stnId"       = EXCLUDED."stnId",
        "roleId"      = EXCLUDED."roleId",
        "grpId"       = EXCLUDED."grpId",
        "distIdLists" = EXCLUDED."distIdLists",
        "connectedAt" = EXCLUDED."connectedAt",
		"ip" = EXCLUDED."ip";
    `

	// ส่ง []string ตรง ๆ (pgx v5 encode เป็น array ให้ ถ้าคอลัมน์เป็น text[]/varchar[])
	var grpIDParam any
	if len(userInfo.GrpID) == 0 {
		grpIDParam = nil
	} else {
		grpIDParam = userInfo.GrpID
	}

	var distListsParam any
	if len(userInfo.DistIdLists) == 0 {
		distListsParam = nil
	} else {
		distListsParam = userInfo.DistIdLists
	}
	clientAddr := userInfo.Conn.RemoteAddr().String()
	clientIP, _, err_ip := net.SplitHostPort(clientAddr)
	if err_ip != nil {
		clientIP = clientAddr
	}
	_, err := dbConn.Exec(ctx, query,
		userInfo.ID, userInfo.Username, userInfo.OrgID,
		userInfo.DeptID, userInfo.CommID, userInfo.StnID,
		userInfo.RoleID, grpIDParam, distListsParam, time.Now(), clientIP,
	)
	if err != nil {
		log.Printf("ERROR: Failed to upsert user connection to DB for EmpID %s: %v", userInfo.ID, err)
	} else {
		log.Printf("Database: Successfully upserted connection for EmpID %s", userInfo.ID)
	}
	return err
}

// ---------- Remove Connection ----------

func removeUserConnectionFromDB(userID string) {
	dbConn, ctx, cancel := utils.ConnectDB()
	if dbConn == nil {
		log.Printf("ERROR: Failed to get DB connection to remove user %s", userID)
		return
	}
	defer cancel()
	defer dbConn.Close(ctx)

	if _, err := dbConn.Exec(ctx, `DELETE FROM user_connections WHERE "empId" = $1`, userID); err != nil {
		log.Printf("ERROR: Failed to remove user connection from DB for EmpID %s: %v", userID, err)
	} else {
		log.Printf("Database: Successfully removed connection for EmpID %s", userID)
	}
}

// ---------- WebSocket Handler ----------

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
	// อ่าน registration message แรกจาก client
	_, msg, err := wsConn.ReadMessage()
	if err != nil {
		log.Println("Failed to read registration message:", err)
		return
	}

	var regMsg model.RegistrationMessage

	if err := json.Unmarshal(msg, &regMsg); err != nil {
		log.Println("Invalid registration message format:", err)
		_ = wsConn.WriteJSON(gin.H{"error": "invalid registration format"})
		return
	}

	subscribeFailureResponse := model.SubscribeResponse{
		EVENT:    "SUBSCRIBE-FAILURE",
		Msg:      "user subscribe Failure",
		OrgId:    regMsg.OrgID,
		Username: regMsg.Username,
	}

	// (แนะนำ) validate ค่าว่างเบื้องต้น
	if strings.TrimSpace(regMsg.OrgID) == "" || strings.TrimSpace(regMsg.Username) == "" {
		_ = wsConn.WriteJSON(gin.H{"error": "orgId and username are required"})
		wsConn.WriteJSON(subscribeFailureResponse)
		return
	}

	dbConn, ctx, cancel := utils.ConnectDB()
	if dbConn == nil {
		_ = wsConn.WriteJSON(gin.H{"error": "could not connect to the database"})
		wsConn.WriteJSON(subscribeFailureResponse)
		return
	}

	connInfo, err := utils.GetUserProfileFromDB(ctx, dbConn, regMsg.OrgID, regMsg.Username)

	dbConn.Close(ctx)
	cancel()

	if err != nil {
		log.Printf("User registration failed for '%s': %v", regMsg.Username, err)
		_ = wsConn.WriteJSON(gin.H{"error": "user not found or invalid credentials"})
		wsConn.WriteJSON(subscribeFailureResponse)
		return
	}
	connInfo.Conn = wsConn

	connMutex.Lock()
	userConnections[connInfo.ID] = connInfo
	connMutex.Unlock()

	go func() {
		if err := upsertUserConnectionToDB(connInfo); err != nil {
			response := model.SubscribeResponse{
				EVENT:    "SUBSCRIBE-FAILURE",
				Msg:      err.Error(),
				OrgId:    regMsg.OrgID,
				Username: regMsg.Username,
			}

			connInfo.Conn.WriteJSON(response)

		} else {
			response := model.SubscribeResponse{
				EVENT:    "SUBSCRIBE-SUCCESS",
				Msg:      "user subscribe success",
				OrgId:    regMsg.OrgID,
				Username: regMsg.Username,
			}

			connInfo.Conn.WriteJSON(response)

		}
	}()

	defer func() {
		response := model.SubscribeResponse{
			EVENT:    "SUBSCRIBE-KICK",
			Msg:      "user was subscribe fron other client",
			OrgId:    regMsg.OrgID,
			Username: regMsg.Username,
		}

		connInfo.Conn.WriteJSON(response)
		connMutex.Lock()
		delete(userConnections, connInfo.ID)
		connMutex.Unlock()

		go removeUserConnectionFromDB(connInfo.ID)
		log.Printf("❌ Disconnected: EmpID=%s", connInfo.ID)
	}()

	// keep-alive: รอจน connection ปิด
	for {
		// if _, _, err := wsConn.ReadMessage(); err != nil {
		// 	break
		// }
		for {
			_, msg, err := wsConn.ReadMessage()
			if err != nil {
				log.Printf("Connection closed for user %s: %v", regMsg.Username, err)
				break
			}

			var event map[string]interface{}
			if err := json.Unmarshal(msg, &event); err != nil {
				log.Println("Invalid message format:", err)
				continue
			}

			evt, ok := event["EVENT"].(string)
			if !ok {
				log.Println("Message missing EVENT field")
				continue
			}

			switch evt {
			case "DASHBOARD":
				orgID, _ := event["orgId"].(string)
				username, _ := event["username"].(string)

				log.Printf("Received DASHBOARD event from %s/%s", orgID, username)

				dbConn, ctx, cancel := utils.ConnectDB()
				if dbConn == nil {
					_ = wsConn.WriteJSON(gin.H{"error": "could not connect to the database"})
					wsConn.WriteJSON(subscribeFailureResponse)
					return
				}

				recipients := []model.Recipient{
					{Type: "username", Value: username},
				}
				err = CoreDashboard(ctx, dbConn, orgID, username, recipients, true, true, true)
				if err != nil {
					log.Printf("Dashboard notification error: %v", err)
				}

				dbConn.Close(ctx)
				cancel()

			case "OTHER_EVENT":
				// handle event อื่น ๆ
			default:
				log.Println("Unknown event:", evt)
			}
		}
	}
}

// ---------- Broadcast ----------

func BroadcastNotification(noti model.Notification) {
	connMutex.Lock()
	defer connMutex.Unlock()

	log.Printf("📢 Broadcasting notification ID: %d", noti.ID)
	sentTo := make(map[string]bool)

	for _, connInfo := range userConnections {
		if sentTo[connInfo.ID] {
			continue
		}

		var shouldSend bool
		if noti.Recipients == nil || len(*noti.Recipients) == 0 {
			// True broadcast
			shouldSend = true
		} else {
			// Check each recipient rule
			for _, recipient := range *noti.Recipients {
				values := strings.Split(recipient.Value, ",")
				for _, value := range values {
					value = strings.TrimSpace(value)
					switch strings.ToLower(recipient.Type) {
					case "orgid":
						shouldSend = connInfo.OrgID == value
					case "empid":
						shouldSend = connInfo.ID == value
					case "roleid":
						shouldSend = connInfo.RoleID == value
					case "deptid":
						shouldSend = connInfo.DeptID == value
					case "stnid":
						shouldSend = connInfo.StnID == value
					case "commid":
						shouldSend = connInfo.CommID == value
					case "username":
						shouldSend = connInfo.Username == value
					case "grpid":
						shouldSend = contains(connInfo.GrpID, value)
					case "provid":
						shouldSend = checkUserInProvince(connInfo.DistIdLists, value)
					case "distid":
						shouldSend = checkUserInDistrict(connInfo.DistIdLists, value)
					}
					if shouldSend {
						break
					}
				}
				if shouldSend {
					break
				}
			}
		}

		if !shouldSend {
			continue
		}

		// -------------------- Send JSON --------------------
		if strings.ToLower(noti.EventType) == "hidden" {
			// Hidden notification
			socketPayload := ToHidden(noti)
			b, _ := json.Marshal(socketPayload)
			log.Printf("Hidden notification payload: %s", b)
			if err := connInfo.Conn.WriteJSON(socketPayload); err != nil {
				log.Printf("❌ Failed to send hidden notification to EmpID %s: %v", connInfo.ID, err)
			}
		} else {
			// Normal notification
			now := time.Now()
			socketPayload := model.Notification{
				ID:          noti.ID,
				OrgID:       noti.OrgID,
				SenderType:  noti.SenderType,
				Sender:      noti.Sender,
				SenderPhoto: noti.SenderPhoto,
				Message:     noti.Message,
				EventType:   noti.EventType,
				RedirectUrl: noti.RedirectUrl,
				Data:        noti.Data,
				CreatedAt:   &now,
				CreatedBy:   noti.CreatedBy,
				ExpiredAt:   noti.ExpiredAt,
				Additional:  noti.Additional,
				Event:       noti.Event,
			}
			b, _ := json.Marshal(socketPayload)
			log.Printf("Normal notification payload: %s", b)
			if err := connInfo.Conn.WriteJSON(socketPayload); err != nil {
				log.Printf("❌ Failed to send normal notification to EmpID %s: %v", connInfo.ID, err)
			}
		}

		sentTo[connInfo.ID] = true
		log.Printf("✅ Broadcasting finished for notification ID: %d", noti.ID)
	}
}

func ToHidden(n model.Notification) model.HiddenNotification {
	return model.HiddenNotification{
		Event:      n.Event,      // optional original event name
		ID:         n.ID,         // optional DB ID
		EventType:  n.EventType,  // should be "hidden"
		Additional: n.Additional, // dashboard/summary JSON
	}
}
