package handler

import (
	"context"
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
	"github.com/jackc/pgx/v5"
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

// ---------- Query User Profile ----------

func getUserProfileFromDB(ctx context.Context, dbConn *pgx.Conn, orgId, username string) (*model.UserConnectionInfo, error) {
	log.Printf("Database: Querying for user '%s' in organization '%s'", username, orgId)

	var userProfile model.UserConnectionInfo
	var roleID string
	var distIdListsJSON []byte
	var GrpID []string

	// 1) ลองอ่านจาก user_connections ก่อน (เก็บ grpId เป็น array อยู่แล้ว)
	connectionQuery := `
        SELECT "empId", "username", "orgId", "deptId", "commId", "stnId", "roleId", "grpId", "distIdLists", COALESCE("ip", '') as ip
        FROM user_connections
        WHERE "orgId" = $1 AND "username" = $2
        LIMIT 1;
    `
	err := dbConn.QueryRow(ctx, connectionQuery, orgId, username).Scan(
		&userProfile.ID, &userProfile.Username, &userProfile.OrgID,
		&userProfile.DeptID, &userProfile.CommID, &userProfile.StnID,
		&roleID, &GrpID, &userProfile.DistIdLists, &userProfile.Ip, // scan array -> []string
	)
	if err == nil {
		userProfile.RoleID = roleID
		userProfile.GrpID = GrpID
		log.Printf("Database: Found existing connection for '%s'", username)
		return &userProfile, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("ERROR: Failed to query user connections for '%s': %v", username, err)
		return nil, err
	}

	// 2) ไม่เจอใน user_connections -> ไปอ่านจาก um_users + um_user_with_groups
	//    รวมหลายแถวของ grpId ให้เป็น array ด้วย array_agg(DISTINCT ...)
	query := `
        SELECT 
          COALESCE(u."empId"::text, '')  AS "empId",
          u."username",
          COALESCE(u."orgId"::text, '')  AS "orgId",
          COALESCE(u."deptId"::text, '') AS "deptId",
          COALESCE(u."commId"::text, '') AS "commId",
          COALESCE(u."stnId"::text, '')  AS "stnId",
          COALESCE(u."roleId"::text, '') AS "roleId",
          COALESCE(array_agg(DISTINCT ug."grpId"::text) FILTER (WHERE ug."grpId" IS NOT NULL), '{}') AS "grpIds",
          COALESCE(uar."distIdLists", '[]'::jsonb) AS "distIdLists"
        FROM um_users u
        LEFT JOIN um_user_with_groups ug 
               ON u."username" = ug."username"
        LEFT JOIN um_user_with_area_response uar 
               ON u."username" = uar."username"
        WHERE u."orgId"::text = $1 
          AND u."username" = $2 
          AND u."active" = true
        GROUP BY u."empId", u."username", u."orgId", u."deptId", u."commId", u."stnId", u."roleId", uar."distIdLists"
        LIMIT 1;
    `
	err = dbConn.QueryRow(ctx, query, orgId, username).Scan(
		&userProfile.ID, &userProfile.Username, &userProfile.OrgID,
		&userProfile.DeptID, &userProfile.CommID, &userProfile.StnID,
		&roleID, &GrpID, &distIdListsJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found or is not active")
		}
		log.Printf("ERROR: Failed to query user profile for '%s': %v", username, err)
		return nil, err
	}

	userProfile.RoleID = roleID
	userProfile.GrpID = GrpID

	// distIdLists จากตาราง uar เป็น jsonb -> unmarshal เป็น []string
	if len(distIdListsJSON) > 0 {
		if err := json.Unmarshal(distIdListsJSON, &userProfile.DistIdLists); err != nil {
			log.Printf("WARNING: Failed to parse distIdLists for user '%s': %v", username, err)
			userProfile.DistIdLists = []string{}
		}
	}

	return &userProfile, nil
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

	connInfo, err := getUserProfileFromDB(ctx, dbConn, regMsg.OrgID, regMsg.Username)

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
		if _, _, err := wsConn.ReadMessage(); err != nil {
			break
		}
	}
}

// ---------- Broadcast ----------

func BroadcastNotification(noti model.Notification) {
	connMutex.Lock()
	defer connMutex.Unlock()

	log.Printf("📢 Broadcasting notification ID: %s", noti.ID)
	sentTo := make(map[string]bool)

	for _, connInfo := range userConnections {
		if sentTo[connInfo.ID] {
			continue
		}
		// if connInfo.OrgID != noti.OrgID {
		// 	continue
		// }

		var isTrueBroadcast bool
		if noti.Recipients != nil {
			isTrueBroadcast = len(*noti.Recipients) == 0

			if isTrueBroadcast {
				log.Printf("  🚀 True broadcast! Sending to EmpID: %s (OrgID: %s)", connInfo.ID, connInfo.OrgID)
				if err := connInfo.Conn.WriteJSON(noti); err != nil {
					log.Printf("    ❌ Failed to send to EmpID %s: %v", connInfo.ID, err)
				}
				sentTo[connInfo.ID] = true
				continue
			}

			for _, recipient := range *noti.Recipients {
				shouldReceive := false

				// รองรับการส่ง value หลายค่าในหนึ่ง rule ด้วย comma
				values := strings.Split(recipient.Value, ",")
				for _, value := range values {
					value = strings.TrimSpace(value)

					switch strings.ToLower(recipient.Type) {
					case "orgid":
						shouldReceive = connInfo.OrgID == value
					case "empid":
						shouldReceive = connInfo.ID == value
					case "roleid":
						shouldReceive = connInfo.RoleID == value
					case "deptid":
						shouldReceive = connInfo.DeptID == value
					case "stnid":
						shouldReceive = connInfo.StnID == value
					case "commid":
						shouldReceive = connInfo.CommID == value
					case "username":
						shouldReceive = connInfo.Username == value
					case "grpid":
						// เช็คสมาชิกใน array
						shouldReceive = contains(connInfo.GrpID, value)
					case "provid":
						shouldReceive = checkUserInProvince(connInfo.DistIdLists, value)
					case "distid":
						shouldReceive = checkUserInDistrict(connInfo.DistIdLists, value)
					}

					if shouldReceive {
						break
					}
				}

				if shouldReceive {
					log.Printf("  🚀 Match found! Sending to EmpID: %s (Rule: %s:%s)", connInfo.ID, recipient.Type, recipient.Value)
					var socketPayload model.Notification
					if noti.EventType == "hidden" {
						socketPayload = model.Notification{ //if you not want to use noti triger in front end use this
							EventType:  noti.EventType,
							Additional: noti.Additional,
							Event:      noti.Event,
						}
					} else {
						now := time.Now()
						socketPayload = model.Notification{
							ID:          noti.ID,
							OrgID:       noti.OrgID, // ใช้ orgId จาก input แทนที่จะใช้ orgId[0]
							SenderType:  noti.SenderType,
							Sender:      noti.Sender,
							SenderPhoto: noti.SenderPhoto,
							Message:     noti.Message,
							EventType:   noti.EventType,
							RedirectUrl: noti.RedirectUrl,
							Data:        noti.Data,
							CreatedAt:   &now, // ใช้เวลาปัจจุบันเสมอ ไม่รับจาก input
							CreatedBy:   noti.CreatedBy,
							ExpiredAt:   noti.ExpiredAt,
							Additional:  noti.Additional,
							Event:       noti.Event,
						}
					}
					b, _ := json.Marshal(socketPayload)
					log.Printf("Socket Payload: %s", b)
					if err := connInfo.Conn.WriteJSON(socketPayload); err != nil {
						log.Printf("    ❌ Failed to send to EmpID %s: %v", connInfo.ID, err)
					}
					sentTo[connInfo.ID] = true
					break
				}
				log.Printf("✅ Broadcasting finished for notification ID: %s", noti.ID)
			}
		}
	}

}
