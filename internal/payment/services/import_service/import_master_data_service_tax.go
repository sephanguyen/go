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

func (s *ImportMasterDataService) ImportTax(ctx context.Context, req *pb.ImportTaxRequest) (*pb.ImportTaxResponse, error) {
	errors := []*pb.ImportTaxResponse_ImportTaxError{}

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
		"tax_id",
		"name",
		"tax_percentage",
		"tax_category",
		"default_flag",
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
			tax, err := CreateTaxEntityFromCsv(line)
			if err != nil {
				errors = append(errors, &pb.ImportTaxResponse_ImportTaxError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse tax item: %s", err),
				})
				continue
			}
			if tax.TaxID.Get() == nil {
				err := s.TaxRepo.Create(ctx, tx, tax)
				if err != nil {
					errors = append(errors, &pb.ImportTaxResponse_ImportTaxError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to create new tax item: %s", err),
					})
				}
			} else {
				err := s.TaxRepo.Update(ctx, tx, tax)
				if err != nil {
					errors = append(errors, &pb.ImportTaxResponse_ImportTaxError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to update tax item: %s", err),
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
		log.Printf("Error when importing tax: %s", err.Error())
	}
	return &pb.ImportTaxResponse{
		Errors: errors,
	}, nil
}

func getTaxColumnName(i int) string {
	mandatoryColumns := []string{
		"tax_id",
		"name",
		"tax_percentage",
		"tax_category",
		"default_flag",
		"is_archived",
	}
	return mandatoryColumns[i]
}

func CreateTaxEntityFromCsv(line []string) (*entities.Tax, error) {
	const (
		TaxID = iota
		Name
		TaxPercentage
		TaxCategory
		DefaultFlag
		IsArchived
	)

	mandatory := []int{Name, TaxPercentage, TaxCategory, DefaultFlag, IsArchived}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", getTaxColumnName(colPosition))
	}

	tax := &entities.Tax{}

	if err := multierr.Combine(
		utils.StringToFormatString("tax_id", line[TaxID], true, tax.TaxID.Set),
		utils.StringToFormatString("name", line[Name], false, tax.Name.Set),
		utils.StringToInt("tax_percentage", line[TaxPercentage], false, tax.TaxPercentage.Set),
		utils.StringToTaxCategory("tax_category", line[TaxCategory], tax.TaxCategory.Set),
		utils.StringToBool("default_flag", line[DefaultFlag], false, tax.DefaultFlag.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, tax.IsArchived.Set),
	); err != nil {
		return nil, err
	}

	return tax, nil
}
