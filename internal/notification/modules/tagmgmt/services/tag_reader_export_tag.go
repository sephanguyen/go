package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *TagMgmtReaderService) ExportTags(ctx context.Context, req *npb.ExportTagsRequest) (*npb.ExportTagsResponse, error) {
	tagFilter := repositories.NewFindTagFilter()
	_ = tagFilter.IsArchived.Set(nil)
	allTags, _, err := svc.TagRepo.FindByFilter(ctx, svc.DB, tagFilter)
	if err != nil {
		return &npb.ExportTagsResponse{}, status.Error(codes.Internal, fmt.Sprintf("failed FindByFilter: %v", err))
	}

	exportColumns := []exporter.ExportColumnMap{}

	allowedHeaders := strings.Split(consts.AllowTagCSVHeaders, "|")
	for _, col := range allowedHeaders {
		exportColumns = append(exportColumns, exporter.ExportColumnMap{
			DBColumn:  col,
			CSVColumn: col,
		})
	}

	exportableTags := sliceutils.Map(allTags, func(tag *entities.Tag) database.Entity {
		return tag
	})

	str, err := exporter.ExportBatch(exportableTags, exportColumns)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed ExportBatch: %v", err))
	}

	return &npb.ExportTagsResponse{
		Data: exporter.ToCSV(str),
	}, nil
}
