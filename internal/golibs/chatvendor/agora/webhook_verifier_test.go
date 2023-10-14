package agora

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Verify(t *testing.T) {
	t.Parallel()
	dummySecret := "dummy-secret"
	dummyCallID := "dummy-call-id"
	dummyTimestamp := uint64(time.Now().Unix())

	hashedSig := md5.Sum([]byte(fmt.Sprintf("%s%s%v", dummyCallID, dummySecret, dummyTimestamp)))
	security := hex.EncodeToString(hashedSig[:])

	verifier := NewWebhookVerifier()

	t.Run("happy case", func(t *testing.T) {
		securityObj := SecurityInfo{
			CallID:    dummyCallID,
			Timestamp: dummyTimestamp,
			Security:  security,
		}
		securityObjBytes, _ := json.Marshal(securityObj)

		isVerified, err := verifier.Verify(dummySecret, securityObjBytes)
		assert.Equal(t, true, isVerified)
		assert.NoError(t, err)
	})
	t.Run("invalid request", func(t *testing.T) {
		securityObj := SecurityInfo{
			CallID:    "wrong-call-id",
			Timestamp: dummyTimestamp,
			Security:  "wrong-security",
		}
		securityObjBytes, _ := json.Marshal(securityObj)

		isVerified, err := verifier.Verify(dummySecret, securityObjBytes)
		assert.Equal(t, false, isVerified)
		assert.NotNil(t, err)
	})
}
