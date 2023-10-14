package agora

import (
	"fmt"
	"time"

	"github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/chatTokenBuilder"
)

// Ref: https://docs.agora.io/en/agora-chat/develop/authentication

func (a *agoraClientImpl) GetAppToken() (string, error) {
	timeExpire := time.Now().Add(TokenAppExpire * time.Second).Unix()
	appToken, err := chatTokenBuilder.BuildChatAppToken(a.AppID, a.PrimaryCertificate, uint32(timeExpire))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", appToken), nil
}

func (a *agoraClientImpl) GetUserToken(userID string) (string, uint64, error) {
	timeExpire := time.Now().Add(TokenUserExpire * time.Second).Unix()
	userToken, err := chatTokenBuilder.BuildChatUserToken(a.AppID, a.PrimaryCertificate, userID, uint32(timeExpire))
	if err != nil {
		return "", 0, err
	}

	return userToken, uint64(timeExpire), nil
}

func (a *agoraClientImpl) GetAppKey() string {
	return a.OrgName + "#" + a.AppName
}
