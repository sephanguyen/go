package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomEntityService struct {
	DB database.Ext
}

func (c *CustomEntityService) ExecuteCustomEntity(ctx context.Context, req *mpb.ExecuteCustomEntityRequest) (*mpb.ExecuteCustomEntityResponse, error) {
	_, err := c.DB.Exec(ctx, req.Sql)
	if err != nil {
		return &mpb.ExecuteCustomEntityResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}
	return &mpb.ExecuteCustomEntityResponse{
		Success: true,
		Error:   "",
	}, nil
}
