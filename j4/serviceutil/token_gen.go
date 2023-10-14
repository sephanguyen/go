package serviceutil

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	"google.golang.org/grpc"
)

type ShamirClient interface {
	GenerateFakeToken(ctx context.Context, in *spb.GenerateFakeTokenRequest, opts ...grpc.CallOption) (*spb.GenerateFakeTokenResponse, error)
}

type TokenGenerator struct {
	ShamirCl spb.InternalServiceClient
}

func NewTokenGenerator(c *infras.ManabieJ4Config, conns *infras.Connections) *TokenGenerator {
	shamirConn := conns.GetGrpcConnByAddr(c.ShamirAddr)
	return &TokenGenerator{
		ShamirCl: spb.NewInternalServiceClient(shamirConn),
	}
}

// GetTokenFromShamir id and schoolID must exist in db and belong to e2e tenant
func (t *TokenGenerator) GetTokenFromShamir(ctx context.Context, id string, schoolID string) (string, error) {
	res, err := t.ShamirCl.GenerateFakeToken(ctx, &spb.GenerateFakeTokenRequest{
		UserId:   id,
		SchoolId: schoolID,
	})
	if err != nil {
		return "", err
	}
	return res.Token, nil
}
