package utils

// nolint
import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func GetAgoraUserID(userID string) string {
	// nolint
	hash := md5.Sum([]byte(userID))
	agoraUserID := hex.EncodeToString(hash[:])
	agoraUserID = strings.ToLower(agoraUserID)

	return agoraUserID
}
