package invoicemgmt

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/manabie-com/backend/features/helper"

	"google.golang.org/grpc/metadata"
)

// generateExchangeToken generates an exchange token from an existing token. This should be used when an RLS was enabled.
// The ExchangeToken also adds the resource path to the token therefore, this must be used when dealing with RLS.
func (s *suite) generateExchangeToken(userID, userGroup string, schoolID int64) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", err
	}
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, "manabie-local", schoolID, s.ShamirConn)
	if err != nil {
		return "", err
	}

	return token, nil
}

// contextWithValidVersion append a version to context
func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

// generateValidAuthenticationToken generates a valid auth token
func generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

// generateAuthenticationToken generates a token
func generateAuthenticationToken(sub string, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(b), nil
}

// contextWithToken append the valid auth token to the context.
func contextWithToken(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)

	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.AuthToken)
}
