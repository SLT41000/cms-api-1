package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func callAPI(url string, method string, data map[string]interface{}) (string, error) {
	var reqBody io.Reader

	// If data is provided and method is not GET, encode as JSON
	if data != nil && method != "GET" {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal data: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// Create HTTP request
	log.Print(url)
	log.Print(reqBody)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers if sending JSON
	if method != "GET" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+os.Getenv("METTTER_TOKEN"))
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(respBody), nil
}

func getEnvList(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return []string{}
	}
	return strings.Split(val, ",")
}

func mapStatus(code string) string {
	if contains(getEnvList("CONV_SCHEDULE"), code) {
		return "NEW"
	}
	if contains(getEnvList("CONV_NEW"), code) {
		return "NEW"
	}
	if contains(getEnvList("CONV_ASSIGNED"), code) {
		return "ASSIGNED"
	}
	if contains(getEnvList("CONV_ACKNOWLEDGE"), code) {
		return "ACKNOWLEDGE"
	}
	if contains(getEnvList("CONV_INPROGRESS"), code) {
		return "INPROGRESS"
	}
	if contains(getEnvList("CONV_DONE"), code) {
		return "DONE"
	}
	if contains(getEnvList("CONV_CLOSED"), code) {
		return "CLOSED"
	}
	if contains(getEnvList("CONV_CANCEL"), code) {
		return "CANCEL"
	}
	return ""
}

func GetCaseStatusMap() map[string]string {
	return map[string]string{
		"NEW":         os.Getenv("NEW"),
		"ASSIGNED":    os.Getenv("ASSIGNED"),
		"ACKNOWLEDGE": os.Getenv("ACKNOWLEDGE"),
		"INPROGRESS":  os.Getenv("INPROGRESS"),
		"DONE":        os.Getenv("DONE"),
		"CANCEL":      os.Getenv("CANCEL"),
		"CLOSED":      os.Getenv("REQUESTCLOSE"),
	}
}

func GetPriorityMap() map[string]int {
	return map[string]int{
		"CRITICAL": getEnvAsInt("CRITICAL", 1),
		"HIGH":     getEnvAsInt("HIGH", 3),
		"MEDIUM":   getEnvAsInt("MEDIUM", 6),
		"LOW":      getEnvAsInt("LOW", 9),
	}
}

func GetPriorityName_TXT(value int) string {
	switch {
	case value <= getEnvAsInt("CRITICAL", 1):
		return "CRITICAL"
	case value <= getEnvAsInt("HIGH", 3):
		return "HIGH"
	case value <= getEnvAsInt("MEDIUM", 6):
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// helper แปลง ENV → int ถ้าไม่มีให้ใช้ default
func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func ConvertStatusList(input string) string {
	// Split by comma
	parts := strings.Split(input, ",")

	// Trim spaces and wrap each item with quotes
	for i := range parts {
		parts[i] = fmt.Sprintf("'%s'", strings.TrimSpace(parts[i]))
	}

	// Join back with commas
	return strings.Join(parts, ", ")
}

func LoadSLAChangeMap() map[string]string {
	mapping := make(map[string]string)
	changeStr := os.Getenv("MONITOR_SLA_CHANGE")

	// Split ด้วย comma
	pairs := strings.Split(changeStr, ",")
	for _, p := range pairs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// แยกด้วย "->"
		parts := strings.Split(p, "->")
		if len(parts) == 2 {
			from := strings.TrimSpace(parts[0])
			to := strings.TrimSpace(parts[1])
			mapping[from] = to
		}
	}
	return mapping
}

func RecheckSLA(currentStatus string) string {
	mapping := LoadSLAChangeMap()
	if next, ok := mapping[currentStatus]; ok {
		return next
	}
	return currentStatus
}

func CalSLA(timelines []model.CaseResponderCustom) []model.CaseResponderCustom {
	if len(timelines) == 0 {
		return timelines
	}

	for i := range timelines {
		if i == 0 {
			timelines[i].Duration = 0
		} else {
			diff := timelines[i].CreatedAt.Sub(timelines[i-1].CreatedAt)
			timelines[i].Duration = int64(diff.Seconds())
		}
	}

	return timelines
}

// GetQueryParams flattens query parameters from Gin context into map[string]string
func GetQueryParams(c *gin.Context) map[string]string {
	flat := make(map[string]string)

	// Query params
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			flat[key] = values[0]
		}
	}

	// Path params (like /stations/:id)
	for _, param := range c.Params {
		flat[param.Key] = param.Value
	}

	return flat
}

func AddPreviousSLA(workflow []model.WorkflowNode) []model.WorkflowNode {
	// Step 1: separate nodes and connections
	nodeMap := make(map[string]*model.WorkflowNode)
	var connections []map[string]interface{}

	for i := range workflow {
		n := &workflow[i]
		if n.Section == "nodes" && n.NodeId != "" {
			nodeMap[n.NodeId] = n
		} else if n.Section == "connections" {
			// connections is stored as array in n.Data
			if arr, ok := n.Data.([]interface{}); ok {
				for _, c := range arr {
					if conn, ok := c.(map[string]interface{}); ok {
						connections = append(connections, conn)
					}
				}
			}
		}
	}

	// Step 2: iterate through each connection (source -> target)
	for _, conn := range connections {
		sourceId, _ := conn["source"].(string)
		targetId, _ := conn["target"].(string)

		sourceNode, srcOk := nodeMap[sourceId]
		targetNode, tgtOk := nodeMap[targetId]
		if !srcOk || !tgtOk {
			continue
		}

		// Extract source and target configs
		srcCfg := extractConfig(sourceNode.Data)
		tgtCfg := extractConfig(targetNode.Data)

		// Get source SLA and set target's sla_
		if slaVal, ok := srcCfg["sla"]; ok {
			tgtCfg["sla_"] = slaVal
		} else {
			// if no SLA in previous node, default sla_ = "0"
			tgtCfg["sla_"] = "0"
		}

		// Save config back into target node
		targetNode.Data = setConfig(targetNode.Data, tgtCfg)
	}

	return workflow
}

func extractConfig(data interface{}) map[string]interface{} {
	if d1, ok := data.(map[string]interface{}); ok {
		if d2, ok := d1["data"].(map[string]interface{}); ok {
			if cfg, ok := d2["config"].(map[string]interface{}); ok {
				return cfg
			}
		}
	}
	return map[string]interface{}{}
}

func setConfig(data interface{}, newCfg map[string]interface{}) interface{} {
	if d1, ok := data.(map[string]interface{}); ok {
		if d2, ok := d1["data"].(map[string]interface{}); ok {
			d2["config"] = newCfg
			d1["data"] = d2
			return d1
		}
	}
	return data
}

func getTimeNow() time.Time {
	timeZone := os.Getenv("TIME_ZONE")
	if timeZone == "" {
		timeZone = "Asia/Bangkok"
	}

	loc, err := time.LoadLocation(timeZone)
	if err != nil || loc == nil {
		fmt.Printf("⚠️ Cannot load timezone '%s': %v → use UTC\n", timeZone, err)
		return time.Now().UTC()
	}

	now := time.Now().In(loc)
	fmt.Printf("✅ Timezone: %s → Now: %v\n", timeZone, now)
	return now
}

//	func convertTime(dt *time.Time) time.Time {
//		//loc := time.FixedZone(os.Getenv("TIME_ZONE"), 0) // +7 ชั่วโมง
//		loc, _ := time.LoadLocation(os.Getenv("TIME_ZONE"))
//		return dt.In(loc)
//	}
func getTimeNowUTC() time.Time {
	return time.Now().UTC()
}

// func getTimeNowBangkok() time.Time {
// 	loc, _ := time.LoadLocation("Asia/Bangkok")
// 	return time.Now().In(loc)
// }

// func displayBangkokTime(t time.Time) string {
// 	loc, _ := time.LoadLocation("Asia/Bangkok")
// 	return t.In(loc).Format("2006-01-02 15:04:05")
// }

// func displayTime(t time.Time) string {
// 	return t.Format("2006-01-02 15:04:05")
// }

func genNotiCustom(
	c context.Context,
	conn *pgx.Conn,
	orgId string,
	createdBy string,
	senderName string,
	senderPhoto string,
	eventType string,
	data []model.Data,
	message string,
	recipients []model.Recipient,
	redirectUrl string,
	senderType string,
	event string,
	additional ...*json.RawMessage,
) error {

	user, err := utils.GetUserByUsername(c, conn, orgId, senderName)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	if user == nil {
		log.Printf("User not found")
	} else {
		if user.Photo != nil {
			senderPhoto = *user.Photo
		} else {
			senderPhoto = ""
		}
	}

	// เตรียม request ชุดเดียว
	now := time.Now().Add(24 * time.Hour)
	req := model.NotificationCreateRequest{
		OrgID:       orgId,
		SenderType:  senderType,
		Sender:      senderName,
		SenderPhoto: senderPhoto,
		Message:     message,
		EventType:   eventType,
		RedirectUrl: redirectUrl,
		Data:        &data,
		Recipients:  &recipients,
		CreatedBy:   createdBy,
		ExpiredAt:   &now, // default TTL 24 ชม.
		Event:       &event,
	}
	if len(additional) > 0 && additional[0] != nil {
		req.Additional = *additional[0]
		log.Printf("Additional data set: %s", string(*additional[0]))
	} else {
		log.Println("No additional data provided")
	}
	// ยิงเข้า CoreNotifications
	created, err := CoreNotifications(c, []model.NotificationCreateRequest{req})
	if err != nil {
		return err
	}
	if len(created) == 0 {
		return fmt.Errorf("no notifications were created")
	}

	// ใช้ตัวที่ DB สร้างจริง (มี id/createdAt) เพื่อ log
	if b, merr := json.MarshalIndent(created[0], "", "  "); merr == nil {
		log.Println(string(b))
	}

	// ถ้า CoreNotifications ยัง "ไม่" broadcast ภายใน ให้เปิดบรรทัดนี้
	// go BroadcastNotification(created[0])

	return nil
}

func CoreNotifications(ctx context.Context, inputs []model.NotificationCreateRequest) ([]model.Notification, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("notification array cannot be empty")
	}

	conn, ctx, cancel := utils.ConnectDB()
	defer cancel()
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var createdNotifications []model.Notification
	now := time.Now()
	for _, input := range inputs {
		noti := model.Notification{
			OrgID:       input.OrgID, // ใช้ orgId จาก input แทนที่จะใช้ orgId[0]
			SenderType:  input.SenderType,
			Sender:      input.Sender,
			SenderPhoto: input.SenderPhoto,
			Message:     input.Message,
			EventType:   input.EventType,
			RedirectUrl: input.RedirectUrl,
			Data:        input.Data,
			CreatedAt:   &now, // ใช้เวลาปัจจุบันเสมอ ไม่รับจาก input
			CreatedBy:   input.CreatedBy,
			ExpiredAt:   input.ExpiredAt,
			Recipients:  input.Recipients,
			Additional:  input.Additional,
			Event:       input.Event,
		}

		recipientsJSON, err := json.Marshal(noti.Recipients)
		if err != nil {
			return nil, fmt.Errorf("failed to process recipients: %w", err)
		}
		dataJSON, err := json.Marshal(noti.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to process custom data: %w", err)
		}

		if noti.EventType != "hidden" {
			err = tx.QueryRow(ctx, `
			INSERT INTO notifications 
			("orgId", "senderType", "sender", "senderPhoto", "message", "eventType", "redirectUrl", "createdAt", "createdBy", "expiredAt", "recipients", "data")
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING "id"
		`, noti.OrgID, noti.SenderType, noti.Sender, noti.SenderPhoto, noti.Message,
				noti.EventType, noti.RedirectUrl, noti.CreatedAt, noti.CreatedBy, noti.ExpiredAt, string(recipientsJSON), dataJSON).Scan(&noti.ID)

			if err != nil {
				return nil, fmt.Errorf("database insert failed: %w", err)
			}

			log.Printf("Database (Tx): Queued insert for notification ID: %d", noti.ID)
		} else {
			noti.ID = generate6DigitID()
			// Support json hidden type
			// 			{
			//   "EVENT": "DASHBOARD",
			//   "eventType": "hidden",
			//   "recipients": [
			//     {
			//       "type": "provId",
			//       "value": "c"
			//     }
			//   ],
			//   "additionalJson": {
			//     "type": "CASE-SUMMARY",
			//     "title_en": "Work Order Summary",
			//     "title_th": "สรุปใบสั่งงาน",
			//     "data": [
			//       {
			//         "total_en": "Total",
			//         "total_th": "ทั้งหมด",
			//         "val": 376
			//       },
			//       {
			//         "g1_en": "Censor",
			//         "g1_th": "เซ็นเซอร์",
			//         "val": 188
			//       },
			//       {
			//         "g2_en": "CCTV",
			//         "g2_th": "กล้อง",
			//         "val": 112
			//       },
			//       {
			//         "g3_en": "Traffic",
			//         "g3_th": "การจราจร",
			//         "val": 79
			//       }
			//     ]
			//   }
			// }
		}
		// Broadcast async
		notiCopy := noti

		// Change noti to ESB
		//go BroadcastNotification(notiCopy)

		payloadMap, err := StructToMap(notiCopy)
		if err != nil {
			log.Println("convert struct to map failed:", err)

		}
		res, err := callAPI(os.Getenv("METTTER_SERVER")+"/welcome/v1/notification/create", "POST", payloadMap)
		if err != nil {
			log.Printf("❌ Send to ESB : %v", err)
		} else {
			log.Printf("✅ Send to ESB : %v", res)
		}

		//utils.SendKafkaJSONMessage([]string{os.Getenv("ESB_SERVER")}, os.Getenv("ESB_NOTIFICATIONS"), "noti", payloadMap)

		createdNotifications = append(createdNotifications, noti)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return createdNotifications, nil
}

func generate6DigitID() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(90000000) + 10000000 // generates 100000–999999
}

func StructToMap(data interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	tmp, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(tmp, &result)
	return result, err
}

func ConvertDate(dateStr string) (string, error) {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}

func ConvertDateSafe(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Try RFC3339 format: 2025-11-11T05:20:00.000Z
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t.Format("2006-01-02")
	}

	// Try Go default time format: 2025-11-11 05:20:00 +0000 UTC
	layout := "2006-01-02 15:04:05 -0700 MST"
	if t, err := time.Parse(layout, dateStr); err == nil {
		return t.Format("2006-01-02")
	}

	// Return empty string if all parsing fails
	return ""
}
