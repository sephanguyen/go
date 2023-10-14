package learnosity

import (
	"encoding/json"
	"time"
)

// FormatUTCTime converts time to UTC, and returns a formatted string in the "YYYYMMDD-HHMM" format.
func FormatUTCTime(time time.Time) string {
	return time.UTC().Format("20060102-1504")
}

// JSONMarshalToString used to encode json to string.
func JSONMarshalToString(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// ContainsKey checks if a map contains a string key.
func ContainsKey(m map[string]any, key string) bool {
	_, ok := m[key]
	return ok
}
