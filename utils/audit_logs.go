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

	if _, ok := AUDIT_LOGS_ALLOW[action]; !ok {
		return fmt.Errorf("invalid action: %s", action)
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

// GetAuditLogsAllow reads the AUDIT_LOGS_ALLOW env variable and returns a map for validation
func GetAuditLogsAllow() map[string]struct{} {
	allowStr := os.Getenv("AUDIT_LOGS_ALLOW")
	allowMap := make(map[string]struct{})

	for _, action := range strings.Split(allowStr, ",") {
		action = strings.TrimSpace(action)
		if action != "" {
			allowMap[action] = struct{}{}
		}
	}

	return allowMap
}
