package token03

import (
	"testing"
)

func Test_GenerateToken03(t *testing.T) {
	var appID uint32 = 1
	roomID := "demo"
	userID := "demo"
	serverSecret := "fa94dd0f974cf2e293728a526b028271"
	var effectiveTimeInSeconds int64 = 3600
	privilege := make(map[int]int)
	privilege[PrivilegeKeyLogin] = PrivilegeEnable
	privilege[PrivilegeKeyPublish] = PrivilegeDisable

	token, err := GenerateToken03(appID, roomID, userID, privilege, serverSecret, effectiveTimeInSeconds)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(token)
}
