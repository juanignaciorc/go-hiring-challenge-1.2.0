package logz

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type ctxKey string

const (
	keyLogger    ctxKey = "logger"
	keyRequestID ctxKey = "request_id"
)

type Logger interface {
	With(fields Fields) Logger
	Info(msg string, fields Fields)
	Error(msg string, fields Fields)
}

type Fields map[string]any

type stdLogger struct {
	base   *log.Logger
	fields Fields
	mu     sync.Mutex
}

func New() Logger {
	return &stdLogger{base: log.New(os.Stdout, "", 0), fields: Fields{}}
}

func (l *stdLogger) With(fields Fields) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	merged := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &stdLogger{base: l.base, fields: merged}
}

func (l *stdLogger) log(level, msg string, fields Fields) {
	l.mu.Lock()
	defer l.mu.Unlock()
	payload := make(Fields, len(l.fields)+len(fields)+2)
	for k, v := range l.fields {
		payload[k] = v
	}
	for k, v := range fields {
		payload[k] = v
	}
	payload["level"] = level
	payload["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	payload["msg"] = msg
	b, _ := json.Marshal(payload)
	l.base.Println(string(b))
}

func (l *stdLogger) Info(msg string, fields Fields)  { l.log("info", msg, fields) }
func (l *stdLogger) Error(msg string, fields Fields) { l.log("error", msg, fields) }

// Context helpers

func IntoContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, keyLogger, logger)
}

func FromContext(ctx context.Context) Logger {
	if v := ctx.Value(keyLogger); v != nil {
		if lg, ok := v.(Logger); ok {
			return lg
		}
	}
	return New()
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

func RequestIDFromContext(ctx context.Context) string {
	if v := ctx.Value(keyRequestID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
