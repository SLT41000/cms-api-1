package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func ConnectDB() (*pgx.Conn, context.Context, context.CancelFunc) {
	logger := GetLog()
	var username string = os.Getenv("DB_USER")
	var password string = os.Getenv("DB_PASS")
	var host string = os.Getenv("DB_HOST")
	var database string = os.Getenv("DB_NAME")
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s", username, password, host, database)
	logger.Debug("Connection String : " + connStr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		cancel()
		logger.DPanic("Unable to connect to database : " + err.Error())
	}

	return conn, ctx, cancel
}

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
