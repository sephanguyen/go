package agora

// nolint
import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

func GetAgoraRESTAPI(agoraCfg configs.AgoraConfig) string {
	return fmt.Sprintf("%s/%s/%s", agoraCfg.RestAPI, agoraCfg.OrgName, agoraCfg.AppName)
}

// nolint
func GetAgoraCommonHeader() map[string]string {
	header := make(map[string]string)
	header[ContentTypeHeaderKey] = "json"

	return header
}

// nolint
func GetAgoraUserPassword(userID, agoraUserID string) string {
	hash := md5.Sum([]byte(userID + agoraUserID))
	password := strings.ToLower(hex.EncodeToString(hash[:]))
	return password
}
