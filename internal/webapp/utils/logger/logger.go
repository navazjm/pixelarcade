package logger

import (
	"context"
	"fmt"
	"log/slog"
)

// ============================================================================
// Mock slog.Logger for testing purposes
// ============================================================================

type MockHandler struct {
	LogOutput string
}

func (h *MockHandler) HandleLog(record slog.Record) error {
	h.LogOutput = fmt.Sprintf("Level: %s, Message: %s", record.Level, record.Message)
	return nil
}

func (m *MockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (m *MockHandler) Handle(ctx context.Context, r slog.Record) error {
	return nil
}

func (m *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MockHandler{}
}

func (m *MockHandler) WithGroup(name string) slog.Handler {
	return &MockHandler{}
}

func NewMock() *slog.Logger {
	handler := &MockHandler{}
	return slog.New(handler)
}
