package helpers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	f_helper "github.com/manabie-com/backend/features/helper"
)

func (helper *ConversationMgmtHelper) generateAuthenticationToken(sub string, template string) (string, error) {
	resp, err := http.Get("http://" + helper.FirebaseAddress + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot generate new user token, err: %v", err)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(bodyResp), nil
}

func (helper *ConversationMgmtHelper) GenerateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return helper.generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

func (helper *ConversationMgmtHelper) GenerateExchangeTokenCtx(ctx context.Context, userID, userGroup string) (string, error) {
	firebaseToken, err := helper.generateAuthenticationToken(userID, "templates/phone.template")
	if err != nil {
		return "", err
	}
	rp := intResourcePathFromCtx(ctx)
	token, err := f_helper.ExchangeToken(firebaseToken, userID, userGroup, helper.ApplicantID, rp, helper.ShamirGRPCConn)
	if err != nil {
		return "", fmt.Errorf("failed to generate exchange token: %w", err)
	}
	return token, nil
}
