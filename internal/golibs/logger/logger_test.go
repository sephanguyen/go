package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type writeSyncSpy struct {
	*bytes.Buffer
}

func (s writeSyncSpy) Sync() error {
	return nil
}

type structuredLog struct {
	Severity string `json:"severity"`
	Time     string `json:"time"`
	Caller   string `json:"caller"`
	Msg      string `json:"msg"`
	Error    string `json:"error"`
}

// NewStructuredLog parse one row of raw, but structured, log.
// It cannot parse more than one row.
func NewStructuredLog(rawlog string) (*structuredLog, error) {
	out := &structuredLog{}
	err := json.Unmarshal([]byte(rawlog), out)
	return out, err
}

// assertLog asserts output from ws is equal to that of expectedLog.
// Only Severity, Msg, and Error are compared.
// Time and Caller are only checked to be non-nil, as they are non-deterministic.
func assertLog(ws *writeSyncSpy, expectedLog structuredLog) error {
	msg := ws.String()
	ws.Reset() // resets the buffer
	slog, err := NewStructuredLog(msg)
	if err != nil {
		return err
	}
	if slog.Time == "" {
		return fmt.Errorf("log time should not be empty")
	}
	if slog.Caller == "" {
		return fmt.Errorf("log caller should not be empty")
	}
	if slog.Severity != expectedLog.Severity {
		return fmt.Errorf("expected log severity %q, got %q", expectedLog.Severity, slog.Severity)
	}
	if slog.Msg != expectedLog.Msg {
		return fmt.Errorf("expected log msg %q, got %q", expectedLog.Msg, slog.Msg)
	}
	if slog.Error != expectedLog.Error {
		return fmt.Errorf("expected log error %q, got %q", expectedLog.Error, slog.Error)
	}
	return nil
}
