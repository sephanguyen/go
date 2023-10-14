package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportProductGroup(ctx context.Context, req *pb.ImportProductGroupRequest) (*pb.ImportProductGroupResponse, error) {
	errors := []*pb.ImportProductGroupResponse_ImportProductGroupError{}

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
		"product_group_id",
		"group_name",
		"group_tag",
		"discount_type",
		"is_archived",
	}
	if err := utils.ValidateCsvHeader(len(headerTitles), header, headerTitles); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	_ = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		hasError := false
		for i, line := range lines[1:] {
			productGroup, err := ProductGroupFromCsv(line)
			if err != nil {
				errors = append(errors, &pb.ImportProductGroupResponse_ImportProductGroupError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse product group item: %s", err),
				})
				hasError = true
				continue
			}

			if productGroupCheck, _ := s.ProductGroupRepo.GetByID(ctx, tx, productGroup.ProductGroupID.String); productGroupCheck.ProductGroupID.Status != pgtype.Present {
				if err = s.ProductGroupRepo.Create(ctx, tx, productGroup); err != nil {
					errors = append(errors, &pb.ImportProductGroupResponse_ImportProductGroupError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to create new product group item: %s", err),
					})
					hasError = true
				}
			} else {
				if err := s.ProductGroupRepo.Update(ctx, tx, productGroup); err != nil {
					errors = append(errors, &pb.ImportProductGroupResponse_ImportProductGroupError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to update product group item: %s", err),
					})
					hasError = true
				}
			}
		}
		if hasError {
			return fmt.Errorf("error importing group setting")
		}
		return nil
	})

	return &pb.ImportProductGroupResponse{
		Errors: errors,
	}, nil
}

func ProductGroupFromCsv(line []string) (*entities.ProductGroup, error) {
	const (
		ProductGroupID = iota
		GroupName
		GroupTag
		DiscountType
		IsArchived
	)
	productGroup := &entities.ProductGroup{}
	if err := multierr.Combine(
		utils.StringToFormatString("product_group_id", line[ProductGroupID], true, productGroup.ProductGroupID.Set),
		utils.StringToFormatString("group_name", line[GroupName], false, productGroup.GroupName.Set),
		utils.StringToFormatString("group_tag", line[GroupTag], true, productGroup.GroupTag.Set),
		utils.StringToFormatString("discount_type", line[DiscountType], false, productGroup.DiscountType.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, productGroup.IsArchived.Set),
	); err != nil {
		return nil, err
	}
	return productGroup, nil
}
