package services

import (
	"bytes"
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/services/mappers"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/services/validations"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *TagMgmtModifierService) ImportTags(ctx context.Context, req *npb.ImportTagsRequest) (*npb.ImportTagsResponse, error) {
	var errMsg string
	sc := scanner.NewCSVScanner(bytes.NewReader(req.Payload))
	if len(sc.GetRow()) == 0 {
		return &npb.ImportTagsResponse{}, status.Error(codes.InvalidArgument, " No data in CSV file")
	}

	// validate CSV headers, only allow tag_id, tag_name, is_archived
	_, err := validations.ValidateCSVHeaders(sc.Head)
	if err != nil {
		return &npb.ImportTagsResponse{}, status.Error(codes.InvalidArgument, err.Error())
	}

	tagData := []*entities.Tag{}
	mapTagNameAndRow := make(map[string]int)
	mapTagIDAndRow := make(map[string]int)
	upsertTagIDs := []string{}

	for sc.Scan() {
		currRow := sc.GetCurRow()
		tag, err := mappers.CSVRowToTag(sc)
		if err != nil {
			return &npb.ImportTagsResponse{}, status.Error(codes.InvalidArgument, fmt.Sprintf(" Error at row number %d: %v", currRow, err))
		}

		// if found duplicate Tag Name in CSV, immediate return
		if rowNum := mapTagNameAndRow[tag.TagName.String]; rowNum > 0 {
			return &npb.ImportTagsResponse{}, status.Error(codes.InvalidArgument, fmt.Sprintf(" Error at row number %d: Tag Name duplicated in CSV file.", currRow))
		}
		mapTagNameAndRow[tag.TagName.String] = currRow

		if tag.TagID.String == "" {
			_ = tag.TagID.Set(idutil.ULIDNow())
		} else {
			// if found duplicate Tag ID in CSV, immediate return
			if rowNum := mapTagIDAndRow[tag.TagID.String]; rowNum > 0 {
				return &npb.ImportTagsResponse{}, status.Error(codes.InvalidArgument, fmt.Sprintf(" Error at row number %d: Tag ID duplicated in CSV file.", currRow))
			}
			mapTagIDAndRow[tag.TagID.String] = currRow
			upsertTagIDs = append(upsertTagIDs, tag.TagID.String)
		}
		tagData = append(tagData, tag)
	}

	if len(tagData) == 0 {
		return &npb.ImportTagsResponse{}, status.Error(codes.InvalidArgument, " No data in CSV file")
	}

	// find tag ids not exist
	notExistTagIDs, err := svc.TagRepo.FindTagIDsNotExist(ctx, svc.DB, database.TextArray(upsertTagIDs))
	if err != nil {
		return &npb.ImportTagsResponse{}, status.Error(codes.Internal, fmt.Sprintf("failed FindTagIDsNotExist: %v", err))
	}
	for _, tagID := range notExistTagIDs {
		if rowNum := mapTagIDAndRow[tagID]; rowNum > 0 {
			errMsg += fmt.Sprintf(" Error at row number %d: Tag ID not exists.", rowNum)
		}
	}

	// find duplicate tag names
	duplicatedTagNames, err := svc.TagRepo.FindDuplicateTagNames(ctx, svc.DB, tagData)
	if err != nil {
		return &npb.ImportTagsResponse{}, status.Error(codes.InvalidArgument, fmt.Sprintf("failed FindDuplicateTagNames: %v", err))
	}
	for _, dbTagName := range duplicatedTagNames {
		if rowNum := mapTagNameAndRow[dbTagName]; rowNum > 0 {
			errMsg += fmt.Sprintf(" Error at row number %d: Tag Name duplicated.", rowNum)
		}
	}

	if len(errMsg) > 0 {
		return &npb.ImportTagsResponse{}, status.Error(codes.InvalidArgument, errMsg)
	}

	err = svc.TagRepo.BulkUpsert(ctx, svc.DB, tagData)
	if err != nil {
		return &npb.ImportTagsResponse{}, status.Error(codes.Internal, fmt.Sprintf("failed BulkUpsert: %v", err))
	}
	return &npb.ImportTagsResponse{}, nil
}
