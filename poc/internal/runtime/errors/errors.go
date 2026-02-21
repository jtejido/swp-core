package errors

import "strings"

const unknownCanonical = "ERR_UNKNOWN"

var aliasToCanonical = map[string]string{
	"INVALID_FRAME":           "ERR_INVALID_FRAME",
	"UNSUPPORTED_VERSION":     "ERR_UNSUPPORTED_VERSION",
	"UNKNOWN_PROFILE":         "ERR_UNKNOWN_PROFILE",
	"INVALID_ENVELOPE":        "ERR_INVALID_ENVELOPE",
	"SECURITY_POLICY":         "ERR_SECURITY_POLICY",
	"UNSUPPORTED_MSG_TYPE":    "ERR_UNSUPPORTED_MSG_TYPE",
	"INVALID_MCP_PAYLOAD":     "ERR_INVALID_MCP_PAYLOAD",
	"INVALID_PROFILE_PAYLOAD": "ERR_INVALID_PROFILE_PAYLOAD",
	"DUPLICATE_MSG_ID":        "ERR_DUPLICATE_MSG_ID",
	"RATE_LIMIT_EXCEEDED":     "ERR_RATE_LIMIT_EXCEEDED",
	"NOT_FOUND":               "ERR_NOT_FOUND",
	"COMPATIBILITY_POLICY":    "ERR_COMPATIBILITY_POLICY",
	"INTERNAL_ERROR":          "ERR_INTERNAL_ERROR",
}

// Canonical returns the ERR_* equivalent for known alias/runtime codes.
// It returns the input unchanged if it is already canonical.
func Canonical(code string) string {
	c := strings.TrimSpace(strings.ToUpper(code))
	if c == "" {
		return ""
	}
	if strings.HasPrefix(c, "ERR_") {
		return c
	}
	if v, ok := aliasToCanonical[c]; ok {
		return v
	}
	return unknownCanonical
}
