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

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportDiscount(ctx context.Context, req *pb.ImportDiscountRequest) (*pb.ImportDiscountResponse, error) {
	errors := []*pb.ImportDiscountResponse_ImportDiscountError{}

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
		"discount_id",
		"name",
		"discount_type",
		"discount_amount_type",
		"discount_amount_value",
		"recurring_valid_duration",
		"available_from",
		"available_until",
		"remarks",
		"is_archived",
		"student_tag_id_validation",
		"parent_tag_id_validation",
		"discount_tag_id",
	}
	err = utils.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// first line is header
		for i, line := range lines[1:] {
			discount, err := DiscountFromCsv(line, headerTitles)
			if err != nil {
				errors = append(errors, &pb.ImportDiscountResponse_ImportDiscountError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf(constant.UnableToParseDiscountItem, err),
				})
				continue
			}
			if discount.DiscountID.Get() == nil {
				err := s.DiscountRepo.Create(ctx, tx, discount)
				if err != nil {
					errors = append(errors, &pb.ImportDiscountResponse_ImportDiscountError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to create new discount item: %s", err),
					})
				}
			} else {
				err := s.DiscountRepo.Update(ctx, tx, discount)
				if err != nil {
					errors = append(errors, &pb.ImportDiscountResponse_ImportDiscountError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to update discount item: %s", err),
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
		log.Printf("Error when importing discount: %s", err.Error())
	}
	return &pb.ImportDiscountResponse{
		Errors: errors,
	}, nil
}

func DiscountFromCsv(line []string, columnNames []string) (*entities.Discount, error) {
	const (
		DiscountID = iota
		Name
		DiscountType
		DiscountAmountType
		DiscountAmountValue
		RecurringValidDuration
		AvailableFrom
		AvailableUntil
		Remarks
		IsArchived
		StudentTagIDValidation
		ParentTagIDValidation
		DiscountTagID
	)

	mandatory := []int{
		Name,
		DiscountType,
		DiscountAmountType,
		DiscountAmountValue,
		AvailableFrom,
		AvailableUntil,
		IsArchived,
	}

	tagToCheck := []int{
		StudentTagIDValidation,
		ParentTagIDValidation,
	}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
	}

	areConflictDiscountTag, err := checkConflictDiscountTag(line, tagToCheck)
	if areConflictDiscountTag {
		return nil, err
	}

	discount := &entities.Discount{}

	if err := multierr.Combine(
		utils.StringToFormatString("discount_id", line[DiscountID], true, discount.DiscountID.Set),
		utils.StringToFormatString("name", line[Name], false, discount.Name.Set),
		utils.StringToDiscountType("discount_type", line[DiscountType], discount.DiscountType.Set),
		utils.StringToDiscountAmountType("discount_amount_type", line[DiscountAmountType], discount.DiscountAmountType.Set),
		utils.StringToFloat("discount_amount_value", line[DiscountAmountValue], false, discount.DiscountAmountValue.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, discount.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, discount.IsArchived.Set),
		utils.StringToDate("available_from", line[AvailableFrom], false, discount.AvailableFrom.Set),
		utils.StringToDate("available_until", line[AvailableUntil], false, discount.AvailableUntil.Set),
		utils.StringToInt("recurring_valid_duration", line[RecurringValidDuration], true, discount.RecurringValidDuration.Set),
		utils.StringToFormatString("student_tag_id_validation", line[StudentTagIDValidation], true, discount.StudentTagIDValidation.Set),
		utils.StringToFormatString("parent_tag_id_validation", line[ParentTagIDValidation], true, discount.ParentTagIDValidation.Set),
		utils.StringToFormatString("discount_tag_id", line[DiscountTagID], true, discount.DiscountTagID.Set),
	); err != nil {
		return nil, err
	}

	return discount, nil
}
