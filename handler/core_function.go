package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mainPackage/model"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

func getTimeNowBangkok() time.Time {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	return time.Now().In(loc)
}

func displayBangkokTime(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	return t.In(loc).Format("2006-01-02 15:04:05")
}

func displayTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
