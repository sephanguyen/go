package users

import (
	"context"
	"fmt"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
)

// ExchangeToken calls shamir after doing basic validation
func (s *UserModifierService) ExchangeToken(ctx context.Context, req *bpb.ExchangeTokenRequest) (*bpb.ExchangeTokenResponse, error) {
	resp, err := s.ShamirClient.VerifyToken(ctx, &spb.VerifyTokenRequest{
		OriginalToken: req.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("s.ShamirClient.VerifyToken: %v", err)
	}

	// exchange token
	newToken, err := s.ShamirClient.ExchangeToken(ctx, &spb.ExchangeTokenRequest{
		NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
			Applicant: s.ApplicantID,
			UserId:    resp.UserId,
		},
		OriginalToken: req.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("s.ShamirClient.ExchangeToken: %v", err)
	}

	return &bpb.ExchangeTokenResponse{
		Token: newToken.NewToken,
	}, nil
}
