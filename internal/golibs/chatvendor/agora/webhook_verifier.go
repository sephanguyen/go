package agora

// The signature in the callback request used to confirm whether this callback is sent from the Chat server.
// The signature is the MD5 hash of the {callId} + {secret} + {timestamp} string, where the value of secret can be found on Agora Console.
// Ref: https://docs.agora.io/en/agora-chat/reference/callbacks-events?platform=web

import (
	// nolint
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

type SecurityInfo struct {
	CallID    string `json:"callId"`
	Timestamp uint64 `json:"timestamp"`
	Security  string `json:"security"`
}

type WebhookVerifier interface {
	Verify(secret string, payload []byte) (bool, error)
}

type agoraWebhookVerifierImpl struct{}

func NewWebhookVerifier() WebhookVerifier {
	return &agoraWebhookVerifierImpl{}
}

func (a *agoraWebhookVerifierImpl) Verify(secret string, payload []byte) (bool, error) {
	securityInfo := new(SecurityInfo)
	err := json.Unmarshal(payload, &securityInfo)
	if err != nil {
		return false, fmt.Errorf("cannot verify request: failed parse security object [%v]", err)
	}

	timestampStr := strconv.FormatUint(securityInfo.Timestamp, 10)
	rawSignature := fmt.Sprintf("%s%s%s", securityInfo.CallID, secret, timestampStr)
	// nolint
	hashedSignature := md5.Sum([]byte(rawSignature))
	signature := hex.EncodeToString(hashedSignature[:])

	if signature != securityInfo.Security {
		return false, fmt.Errorf("cannot verify request: invalid request")
	}

	return true, nil
}
