package users

import (
	"context"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	// spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *UserModifierService) ExchangeCustomToken(ctx context.Context, req *bpb.ExchangeCustomTokenRequest) (*bpb.ExchangeCustomTokenResponse, error) {
	/*resp, err := s.ShamirClient.VerifyTokenV2(ctx, &spb.VerifyTokenRequest{OriginalToken: req.Token})
	if err != nil {
		return nil, fmt.Errorf("ShamirClient.VerifyToken: %w", err)
	}

	var cToken string
	if resp.GetTenantId() != "" {
		tenantClient, err := s.TenantManager.TenantClient(ctx, resp.GetTenantId())
		if err != nil {
			return nil, fmt.Errorf("tenantManager.TenantClient: %w", err)
		}
		cToken, err = tenantClient.CustomToken(ctx, resp.UserId)
		if err != nil {
			return nil, fmt.Errorf("tenantClient.CustomToken: %w", err)
		}
	} else {
		cToken, err = s.FirebaseClient.CustomToken(ctx, resp.UserId)
		if err != nil {
			return nil, fmt.Errorf("firebaseApp.CustomToken: %w", err)
		}
	}

	return &bpb.ExchangeCustomTokenResponse{CustomToken: cToken}, nil*/

	res, err := s.UserMgmtAuthSvc.ExchangeCustomToken(ctx, &upb.ExchangeCustomTokenRequest{Token: req.Token})
	if err != nil {
		return nil, err
	}
	return &bpb.ExchangeCustomTokenResponse{CustomToken: res.CustomToken}, nil
}
