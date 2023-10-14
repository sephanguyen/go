package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deprecated: We don't use this API, but still keep because product requirements might change in the future
func (rcv *TagMgmtModifierService) DeleteTag(ctx context.Context, req *npb.DeleteTagRequest) (*npb.DeleteTagResponse, error) {
	if req.TagId == "" {
		return nil, status.Error(codes.InvalidArgument, "TagID is empty")
	}

	_, err := rcv.TagRepo.FindByID(ctx, rcv.DB, database.Text(req.TagId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("DeleteTag: %v", err))
	}

	err = database.ExecInTx(ctx, rcv.DB, func(ctx context.Context, tx pgx.Tx) error {
		// delete related InfoNotificationTags
		softDeleteFilter := repositories.NewSoftDeleteNotificationTagFilter()
		_ = softDeleteFilter.TagIDs.Set([]string{req.TagId})
		err = rcv.InfoNotificationTagRepo.SoftDelete(ctx, tx, softDeleteFilter)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("DeleteTag.InfoNotificationTagRepo.SoftDelete: %v", err))
		}

		// delete Tags
		err = rcv.TagRepo.SoftDelete(ctx, tx, database.TextArray([]string{req.TagId}))
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("DeleteTag.TagRepo.SoftDelete: %v", err))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &npb.DeleteTagResponse{}, nil
}
