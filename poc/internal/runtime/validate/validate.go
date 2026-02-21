package validate

import (
	"fmt"
	"strings"
)

func RequireNonEmpty(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s required", fieldName)
	}
	return nil
}

func Severity(severity string) error {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("invalid severity")
	}
}

func Traceparent(traceparent string) error {
	tp := strings.TrimSpace(traceparent)
	if tp == "" {
		return fmt.Errorf("traceparent required")
	}
	parts := strings.Split(tp, "-")
	if len(parts) != 4 {
		return fmt.Errorf("invalid traceparent format")
	}
	return nil
}

func HasCorrelation(msgID, taskID, rpcID []byte) bool {
	return len(msgID) > 0 || len(taskID) > 0 || len(rpcID) > 0
}
