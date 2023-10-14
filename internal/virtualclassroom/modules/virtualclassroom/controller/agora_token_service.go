package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/agoratokenbuilder"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type AgoraTokenService struct {
	AgoraCfg configurations.AgoraConfig
}

func (a *AgoraTokenService) GenerateAgoraStreamToken(referenceID string, userID string, role domain.AgoraRole) (string, error) {
	expireTimestamp := uint32(time.Now().UTC().Unix() + 21600)

	// Using lesson id / channel id as channel name for agora token
	token, err := agoratokenbuilder.BuildStreamToken(a.AgoraCfg.AppID, a.AgoraCfg.Cert, referenceID, userID, domain.AgoraRoleMap[role], expireTimestamp)
	if err != nil {
		return "", fmt.Errorf("error generate Agora token: %v", err)
	}

	return token, nil
}

func (a *AgoraTokenService) BuildRTMToken(referenceID string, userID string) (string, error) {
	expireTimestamp := uint32(time.Now().UTC().Unix() + 3600)

	rtmToken, err := agoratokenbuilder.BuildRTMToken(a.AgoraCfg.AppID, a.AgoraCfg.Cert, userID+referenceID, expireTimestamp)
	if err != nil {
		return "", fmt.Errorf("error generate Agora RTM token: %v", err)
	}

	return rtmToken, nil
}

func (a *AgoraTokenService) BuildRTMTokenByUserID(userID string) (string, error) {
	expireTimestamp := uint32(time.Now().UTC().Unix() + 3600)

	rtmToken, err := agoratokenbuilder.BuildRTMToken(a.AgoraCfg.AppID, a.AgoraCfg.Cert, userID, expireTimestamp)
	if err != nil {
		return "", fmt.Errorf("error generate Agora RTM token: %v", err)
	}

	return rtmToken, nil
}
