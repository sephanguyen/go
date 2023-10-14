package common

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
)

func (s *NotificationSuite) ContextWithToken(ctx context.Context, authtoken string) context.Context {
	return contextWithToken(ctx, authtoken)
}

func (s *NotificationSuite) ContextWithTokenAndTimeOut(ctx context.Context, token string) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	return contextWithToken(newCtx, token), cancel
}

func (s *NotificationSuite) generateExchangeToken(userID, userGroup string, orgID int64) (string, error) {
	firebaseToken, err := s.CommunicationHelper.GenerateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", fmt.Errorf("error when create generateValidAuthenticationToken: %v", err)
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.ApplicantID, orgID, s.ShamirGRPCConn, helper.NewAuthUserListener(context.Background(), s.AuthDBConn))
	if err != nil {
		return "", fmt.Errorf("error when create exchange token: %v", err)
	}
	return token, nil
}
