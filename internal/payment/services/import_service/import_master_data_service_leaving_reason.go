package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportLeavingReason(ctx context.Context, req *pb.ImportLeavingReasonRequest) (*pb.ImportLeavingReasonResponse, error) {
	errors := []*pb.ImportLeavingReasonResponse_ImportLeavingReasonError{}

	r := csv.NewReader(bytes.NewReader(req.Payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	header := lines[0]
	headerTitles := []string{
		"leaving_reason_id",
		"name",
		"leaving_reason_type",
		"remark",
		"is_archived",
	}
	if err = utils.ValidateCsvHeader(len(headerTitles), header, headerTitles); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// first line is header
		for i, line := range lines[1:] {
			var leavingReason entities.LeavingReason
			leavingReason, err = LeavingReasonFromCsv(line)
			if err != nil {
				errors = append(errors, &pb.ImportLeavingReasonResponse_ImportLeavingReasonError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf(constant.UnableToParseLeavingReasonItem, err),
				})
				continue
			}
			if leavingReason.LeavingReasonID.Status == pgtype.Null {
				err = s.LeavingReasonRepo.Create(ctx, tx, &leavingReason)
				if err != nil {
					errors = append(errors, &pb.ImportLeavingReasonResponse_ImportLeavingReasonError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to create leaving reason item: %s", err),
					})
				}
			} else {
				err = s.LeavingReasonRepo.Update(ctx, tx, &leavingReason)
				if err != nil {
					errors = append(errors, &pb.ImportLeavingReasonResponse_ImportLeavingReasonError{
						RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to update leaving reason item: %s", err),
					})
				}
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error when importing leaving reason: %s", err.Error())
	}
	return &pb.ImportLeavingReasonResponse{
		Errors: errors,
	}, nil
}

func LeavingReasonFromCsv(line []string) (leavingReason entities.LeavingReason, err error) {
	const (
		LeavingReasonID = iota
		Name
		LeavingReasonType
		Remark
		IsArchived
	)

	err = multierr.Combine(
		utils.StringToFormatString("leaving_reason_id", line[LeavingReasonID], true, leavingReason.LeavingReasonID.Set),
		utils.StringToFormatString("name", line[Name], false, leavingReason.Name.Set),
		utils.StringToLeavingReasonType("leaving_reason_type", line[LeavingReasonType], leavingReason.LeavingReasonType.Set),
		utils.StringToFormatString("remark", line[Remark], true, leavingReason.Remark.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, leavingReason.IsArchived.Set),
	)
	return
}
