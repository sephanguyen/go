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

func (s *ImportMasterDataService) ImportAccountingCategory(ctx context.Context, req *pb.ImportAccountingCategoryRequest) (*pb.ImportAccountingCategoryResponse, error) {
	errors := []*pb.ImportAccountingCategoryResponse_ImportAccountingCategoryError{}

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
		"accounting_category_id",
		"name",
		"remarks",
		"is_archived",
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
			accountingCategory, err := AccountingCategoryFromCsv(line, headerTitles)
			if err != nil {
				errors = append(errors, &pb.ImportAccountingCategoryResponse_ImportAccountingCategoryError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse accounting category item: %s", err),
				})
				continue
			}
			if accountingCategory.AccountingCategoryID.Get() == nil {
				err := s.AccountingCategoryRepo.Create(ctx, tx, accountingCategory)
				if err != nil {
					errors = append(errors, &pb.ImportAccountingCategoryResponse_ImportAccountingCategoryError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to create new accounting category item: %s", err),
					})
				}
			} else {
				err := s.AccountingCategoryRepo.Update(ctx, tx, accountingCategory)
				if err != nil {
					errors = append(errors, &pb.ImportAccountingCategoryResponse_ImportAccountingCategoryError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to update accounting category item: %s", err),
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
		log.Printf("Error when importing service accounting category: %s", err.Error())
	}
	return &pb.ImportAccountingCategoryResponse{
		Errors: errors,
	}, nil
}

func AccountingCategoryFromCsv(line []string, columnNames []string) (*entities.AccountingCategory, error) {
	const (
		AccountingCategoryID = iota
		Name
		Remarks
		IsArchived
	)

	mandatory := []int{Name, IsArchived}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
	}

	accountingCategory := &entities.AccountingCategory{}

	if err := multierr.Combine(
		utils.StringToFormatString("accounting_category_id", line[AccountingCategoryID], true, accountingCategory.AccountingCategoryID.Set),
		utils.StringToFormatString("name", line[Name], false, accountingCategory.Name.Set),
		utils.StringToFormatString("remarks", line[Remarks], true, accountingCategory.Remarks.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, accountingCategory.IsArchived.Set),
	); err != nil {
		return nil, err
	}

	return accountingCategory, nil
}
