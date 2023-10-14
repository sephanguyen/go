package services

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (rcv *TagMgmtModifierService) UpsertTag(ctx context.Context, req *npb.UpsertTagRequest) (*npb.UpsertTagResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "TagName is empty")
	}
	if req.TagId == "" {
		req.TagId = idutil.ULIDNow()
	}

	req.Name = strings.TrimSpace(req.Name)

	isExist, err := rcv.TagRepo.DoesTagNameExist(ctx, rcv.DB, database.Text(req.Name))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if isExist {
		return nil, status.Error(codes.InvalidArgument, "TagName is exist")
	}

	now := timeutil.Now()
	tag := new(entities.Tag)
	err = multierr.Combine(
		tag.TagID.Set(req.TagId),
		tag.TagName.Set(req.Name),
		tag.CreatedAt.Set(now),
		tag.UpdatedAt.Set(now),
		tag.DeletedAt.Set(nil),
		tag.IsArchived.Set(false),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = rcv.TagRepo.Upsert(ctx, rcv.DB, tag)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &npb.UpsertTagResponse{TagId: req.TagId}, nil
}
