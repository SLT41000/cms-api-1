package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
