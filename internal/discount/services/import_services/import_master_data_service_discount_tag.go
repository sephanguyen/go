package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportDiscountTag(ctx context.Context, req *pb.ImportDiscountTagRequest) (*pb.ImportDiscountTagResponse, error) {
	errors := []*pb.ImportDiscountTagResponse_ImportDiscountTagError{}

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
		"discount_tag_id",
		"discount_tag_name",
		"selectable",
		"is_archived",
	}
	err = utils.ValidateCsvHeader(len(headerTitles), header, headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for i, line := range lines[1:] {
			discountTag, err := DiscountTagFromCsv(line, headerTitles)
			if err != nil {
				errors = append(errors, &pb.ImportDiscountTagResponse_ImportDiscountTagError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse discount tag item: %s", err.Error()),
				})
				continue
			}
			if discountTag.DiscountTagID.Get() == nil {
				err = s.DiscountTagRepo.Create(ctx, tx, discountTag)
				if err != nil {
					errors = append(errors, &pb.ImportDiscountTagResponse_ImportDiscountTagError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to create new discount tag item: %s", err),
					})
				}
			} else {
				err = s.DiscountTagRepo.Update(ctx, tx, discountTag)
				if err != nil {
					errors = append(errors, &pb.ImportDiscountTagResponse_ImportDiscountTagError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to update discount tag item: %s", err),
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
		log.Printf("Error when importing discount tag: %s", err.Error())
	}
	return &pb.ImportDiscountTagResponse{
		Errors: errors,
	}, nil
}

func DiscountTagFromCsv(line []string, columnNames []string) (*entities.DiscountTag, error) {
	const (
		DiscountTagID = iota
		DiscountTagName
		Selectable
		IsArchived
	)
	mandatory := []int{
		DiscountTagName,
		Selectable,
		IsArchived,
	}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
	}

	discountTag := &entities.DiscountTag{}
	if err := multierr.Combine(
		utils.StringToFormatString("discount_tag_id", line[DiscountTagID], true, discountTag.DiscountTagID.Set),
		utils.StringToFormatString("discount_tag_name", line[DiscountTagName], false, discountTag.DiscountTagName.Set),
		utils.StringToBool("selectable", line[Selectable], false, discountTag.Selectable.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, discountTag.IsArchived.Set),
		discountTag.CreatedAt.Set(nil),
		discountTag.UpdatedAt.Set(nil),
	); err != nil {
		return nil, err
	}
	return discountTag, nil
}
