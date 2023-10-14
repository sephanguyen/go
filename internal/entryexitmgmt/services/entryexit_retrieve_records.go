package services

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/entryexitmgmt/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *EntryExitModifierService) RetrieveEntryExitRecords(ctx context.Context, req *eepb.RetrieveEntryExitRecordsRequest) (*eepb.RetrieveEntryExitRecordsResponse, error) {
	limit, offset := s.getLimitOffset(req)

	filter := repositories.RetrieveEntryExitRecordFilter{
		StudentID:    database.Text(req.StudentId),
		RecordFilter: req.RecordFilter,
		Limit:        limit,
		Offset:       offset,
	}
	entryExitRecords, err := s.StudentEntryExitRecordsRepo.RetrieveRecordsByStudentID(ctx, s.DB, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(entryExitRecords) == 0 {
		return &eepb.RetrieveEntryExitRecordsResponse{}, nil
	}

	responseItems := make([]*eepb.EntryExitRecord, 0, len(entryExitRecords))
	for _, ee := range entryExitRecords {
		responseItems = append(responseItems, &eepb.EntryExitRecord{
			EntryexitId: ee.ID.Int,
			EntryAt:     timestamppb.New(ee.EntryAt.Time),
			ExitAt:      timestamppb.New(ee.ExitAt.Time),
		})
	}

	return &eepb.RetrieveEntryExitRecordsResponse{
		EntryExitRecords: responseItems,
		NextPage:         s.getNextPaging(limit, offset),
	}, nil
}

func (s *EntryExitModifierService) getLimitOffset(req *eepb.RetrieveEntryExitRecordsRequest) (limit, offset pgtype.Int8) {
	limit = database.Int8(constant.PageLimit)
	offset = database.Int8(0)

	if req.Paging != nil && req.Paging.Limit != 0 {
		_ = limit.Set(req.Paging.Limit)
		_ = offset.Set(req.Paging.GetOffsetInteger())
	}

	return limit, offset
}

func (s *EntryExitModifierService) getNextPaging(limit, offset pgtype.Int8) *cpb.Paging {
	return &cpb.Paging{
		Limit: uint32(limit.Int),
		Offset: &cpb.Paging_OffsetInteger{
			OffsetInteger: limit.Int + offset.Int,
		},
	}
}
