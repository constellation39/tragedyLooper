package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New creates a new zap logger.
func New() *zap.Logger {
	var core zapcore.Core
	var config zap.Config

	// Use production logger only if LOG_ENV is explicitly set to "production"
	if strings.ToLower(os.Getenv("LOG_ENV")) == "production" {
		config = zap.NewProductionConfig()
		// Configure lumberjack for log rotation
		logRotator := &lumberjack.Logger{
			Filename:   "logs/app.log", // Log file path
			MaxSize:    100,            // Max size in megabytes before rotation
			MaxBackups: 3,              // Max number of old log files to keep
			MaxAge:     28,             // Max number of days to retain old log files
			Compress:   true,           // Compress old log files
		}

		// Create a core that writes to both console and file
		consoleEncoder := zapcore.NewConsoleEncoder(config.EncoderConfig)
		fileEncoder := zapcore.NewJSONEncoder(config.EncoderConfig)
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), config.Level),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(logRotator), config.Level),
		)
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Human-readable, colored level
		config.EncoderConfig.EncodeCaller = relativeCallerEncoder           // Use custom relative path encoder
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(config.EncoderConfig), zapcore.AddSync(os.Stdout), config.Level)
	}

	// Allow overriding log level from environment variable
	level, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		switch strings.ToLower(level) {
		case "debug":
			config.Level.SetLevel(zap.DebugLevel)
		case "info":
			config.Level.SetLevel(zap.InfoLevel)
		case "warn":
			config.Level.SetLevel(zap.WarnLevel)
		case "error":
			config.Level.SetLevel(zap.ErrorLevel)
		}
	}

	// AddCallerSkip(1) is important to make sure the caller is reported correctly
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger
}

// custom caller encoder to get relative path
func relativeCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to short caller if CWD is not available.
		zapcore.ShortCallerEncoder(caller, enc)
		return
	}
	relPath, err := filepath.Rel(cwd, caller.File)
	if err != nil {
		// Fallback to full caller if relative path calculation fails.
		zapcore.FullCallerEncoder(caller, enc)
		return
	}
	enc.AppendString(fmt.Sprintf("%s:%d", relPath, caller.Line))
}
