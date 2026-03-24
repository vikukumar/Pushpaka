package basemodel

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm/logger"
)

// CustomLogger is a minimal GORM logger that hides the full SQL query.
type CustomLogger struct {
	LogLevel logger.LogLevel
	AppEnv   string
}

// LogMode implements logger.Interface.
func (l *CustomLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &CustomLogger{
		LogLevel: level,
		AppEnv:   l.AppEnv,
	}
}

// Info implements logger.Interface.
func (l *CustomLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		log.Printf(msg, data...)
	}
}

// Warn implements logger.Interface.
func (l *CustomLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		log.Printf(msg, data...)
	}
}

// Error implements logger.Interface.
func (l *CustomLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		log.Printf(msg, data...)
	}
}

// Trace implements logger.Interface.
func (l *CustomLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	//file := utils.FileWithLineNum()

	// Determine status message
	status := "SUCCESS"
	if err != nil && err != logger.ErrRecordNotFound {
		status = fmt.Sprintf("ERROR: %v", err)
	}

	// For slow queries, add context
	if elapsed > 200*time.Millisecond && status == "SUCCESS" {
		status = fmt.Sprintf("SLOW SQL (%v)", elapsed)
	}

	// In debug mode (non-production), show the file and line number
	if l.AppEnv != "production" {
		log.Printf("[Pushpaka DB] %s | %s", elapsed, status)
	} else {
		// In production, only log errors or slow queries
		if err != nil && err != logger.ErrRecordNotFound {
			log.Printf("[Pushpaka DB] %s | %s", elapsed, status)
		} else if elapsed > 500*time.Millisecond {
			log.Printf("[Pushpaka DB] SLOW | %s", elapsed)
		}
	}
}

// NewCustomLogger creates a new instance of our minimal logger.
func NewCustomLogger(appEnv string) logger.Interface {
	level := logger.Warn
	if appEnv != "production" {
		level = logger.Info
	}
	return &CustomLogger{
		LogLevel: level,
		AppEnv:   appEnv,
	}
}
