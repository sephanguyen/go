package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/stretchr/testify/assert"
)

func TestNotificationInternalUserRepo_GetByOrgID(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		sampleUserID := idutil.ULIDNow()

		agoraUserID := GetAgoraUserID(sampleUserID)

		hash := md5.Sum([]byte(sampleUserID))
		expectedAgoraUserID := strings.ToLower(hex.EncodeToString(hash[:]))
		assert.Equal(t, expectedAgoraUserID, agoraUserID)
	})
}
