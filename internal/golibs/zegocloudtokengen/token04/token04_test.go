package token04

import (
	"testing"
)

func Test_GenerateToken03(t *testing.T) {
	var appID uint32 = 1
	userID := "demo"
	serverSecret := "fa94dd0f974cf2e293728a526b028271"
	var effectiveTimeInSeconds int64 = 3600
	var payload string = ""

	token, err := GenerateToken04(appID, userID, serverSecret, effectiveTimeInSeconds, payload)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(token)
}
