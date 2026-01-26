package logger

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Level string

const (
	INFO  Level = "info"
	WARN  Level = "warn"
	ERROR Level = "error"
)

type LogEvent struct {
	Level     Level             `json:"level"`
	Message   string            `json:"message"`
	Source    string            `json:"source"`
	Timestamp string            `json:"timestamp"`
	Fields    map[string]string `json:"fields,omitempty"`
}

type LogEmitter struct {
	ctx context.Context
}

func NewLogEmitter(ctx context.Context) *LogEmitter {
	return &LogEmitter{ctx: ctx}
}

func (e *LogEmitter) emit(level Level, msg string, source string, fields map[string]string) {
	runtime.EventsEmit(e.ctx, "log:event", LogEvent{
		Level:     level,
		Message:   msg,
		Source:    source,
		Fields:    fields,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Public API
func (e *LogEmitter) Info(msg, source string, fields map[string]string) {
	e.emit(INFO, msg, source, fields)
}

func (e *LogEmitter) Warn(msg, source string, fields map[string]string) {
	e.emit(WARN, msg, source, fields)
}

func (e *LogEmitter) Error(msg, source string, fields map[string]string) {
	e.emit(ERROR, msg, source, fields)
}
