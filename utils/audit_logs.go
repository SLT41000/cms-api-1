package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func GetLog() *zap.Logger {
	maxSizeStr := os.Getenv("LOG_MaxSize")
	maxSize, _ := strconv.Atoi(maxSizeStr)
	MaxBackupsStr := os.Getenv("LOG_MaxBackups")
	MaxBackups, _ := strconv.Atoi(MaxBackupsStr)
	maxAgeStr := os.Getenv("LOG_MaxAge")
	maxAge, _ := strconv.Atoi(maxAgeStr)
	// Lumberjack logger (file output)
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   os.Getenv("LOG_Filename"),
		MaxSize:    maxSize, // MB
		MaxBackups: MaxBackups,
		MaxAge:     maxAge, // days
		Compress:   false,
	})

	// Console output (stdout)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Encoder configuration
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:     "T",
		LevelKey:    "L",
		NameKey:     "N",
		CallerKey:   "C",
		MessageKey:  "M",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encoderCfg)

	// Combine file and console outputs
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriter, zapcore.DebugLevel),
		zapcore.NewCore(encoder, consoleWriter, zapcore.DebugLevel),
	)

	logger := zap.New(core, zap.AddCaller())
	// logger.Info("Logger initialized", zap.String("output", filename))

	return logger
}

func InsertAuditLogs(
	ctx *gin.Context,
	conn *pgx.Conn,
	orgId string,
	username string,
	txId string,
	uniqueId string,
	mainFunc string,
	subFunc string,
	nameFunc string,
	action string,
	status int,
	timeStart time.Time,
	newData any,
	resData any,
	message string,
) error {
	AUDIT_LOGS_ALLOW := GetAuditLogsAllow()

	// Step 1: Check if action is allowed
	allowedStatuses, ok := AUDIT_LOGS_ALLOW[action]
	if !ok {
		return fmt.Errorf("invalid action: %s", action)
	}

	// Step 2: Check if this status is allowed for that action
	isAllowed := false
	for _, s := range allowedStatuses {
		if s == status {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		log.Printf("Skip audit log — action '%s' with status %d not allowed", action, status)
		return nil
	}
	now := time.Now()
	duration := now.Sub(timeStart).Seconds()
	duration, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", duration), 64)
	log.Print("-----InsertAuditLogs----")
	// Step 1: Get old data if uniqueId provided
	var oldData string = "{}"
	if uniqueId != "" && (action == "update" || action == "delete" || action == "view") {
		err := conn.QueryRow(ctx, `
			SELECT "newData"
			FROM audit_logs
			WHERE "uniqueId" = $1
			AND "action" in ( 'update', 'create') 
			ORDER BY "createdAt" DESC
			LIMIT 1
		`, uniqueId).Scan(&oldData)
		if err != nil && err.Error() != "no rows in result set" {
			return fmt.Errorf("query old data failed: %v", err)
		}
	}

	// Step 2: Convert newData/resData to JSON text
	newDataJSON, _ := json.Marshal(newData)
	resDataJSON, _ := json.Marshal(resData)
	log.Print(string(newDataJSON))
	log.Print(string(resDataJSON))
	// Step 3: Insert new record
	query := `
		INSERT INTO audit_logs (
			"orgId", "username", "txId", "uniqueId",
			"mainFunc", "subFunc", "nameFunc",
			"action", "status", "duration",
			"newData", "oldData", "resData", "message"
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`
	log.Print(query)
	_, err := conn.Exec(ctx, query,
		orgId, username, txId, uniqueId,
		mainFunc, subFunc, nameFunc,
		action, status, duration,
		string(newDataJSON), string(oldData), string(resDataJSON), message,
	)

	log.Print(err)
	if err != nil {
		return fmt.Errorf("insert audit failed: %v", err)
	}

	return nil
}

// GetAuditLogsAllow reads AUDIT_LOGS_ALLOW and returns map[action][]int{allowedStatuses}
func GetAuditLogsAllow() map[string][]int {
	allowStr := os.Getenv("AUDIT_LOGS_ALLOW")
	allowMap := make(map[string][]int)

	// Example format:
	// AUDIT_LOGS_ALLOW=create:0,-1,-2;update:0,-1,-2;search:-1,-2
	for _, entry := range strings.Split(allowStr, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 {
			continue
		}

		action := strings.TrimSpace(parts[0])
		statusList := strings.Split(parts[1], ",")
		var statuses []int
		for _, s := range statusList {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if val, err := strconv.Atoi(s); err == nil {
				statuses = append(statuses, val)
			}
		}
		allowMap[action] = statuses
	}

	return allowMap
}

func WriteConsole(level string, function string, msg string, args ...interface{}) {
	// Global enable/disable
	if strings.ToLower(os.Getenv("CONSOLE_LOG")) != "true" {
		return
	}

	// Normalize level to uppercase for comparison
	level = strings.ToUpper(level)

	// Parse allowed levels
	allowedEnv := os.Getenv("CONSOLE_LOG_ALLOW")
	if allowedEnv != "" {
		allowedLevels := strings.Split(strings.ToUpper(allowedEnv), ",")
		isAllowed := false
		for _, allowed := range allowedLevels {
			if strings.TrimSpace(allowed) == level {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return // 🚫 Level not allowed
		}
	}

	// Format message safely
	formatted := msg
	if len(args) > 0 {
		// If msg has no % directives, just append the args
		if strings.Contains(msg, "%") {
			formatted = fmt.Sprintf(msg, args...)
		} else {
			formatted = fmt.Sprintf("%s | %v", msg, args)
		}
	}

	// Add timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Colorized prefix by log level
	var prefix, color string
	switch level {
	case "INFO":
		color = "\033[1;34m" // blue
		prefix = "[INFO]"
	case "WARN", "WARNING":
		color = "\033[1;33m" // yellow
		prefix = "[WARN]"
	case "ERROR":
		color = "\033[1;31m" // red
		prefix = "[ERROR]"
	case "DEBUG":
		color = "\033[0;36m" // cyan
		prefix = "[DEBUG]"
	default:
		color = "\033[0;37m" // gray
		prefix = "[LOG]"
	}

	// Print colored output
	fmt.Printf("%s%s %s [%s] %s\033[0m\n", color, prefix, timestamp, function, formatted)
}
