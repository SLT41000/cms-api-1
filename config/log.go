package config

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func GetLog() *zap.Logger {
	filename := "logs/cmsApi.log"

	// Lumberjack logger (file output)
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    1, // MB
		MaxBackups: 3,
		MaxAge:     28, // days
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
