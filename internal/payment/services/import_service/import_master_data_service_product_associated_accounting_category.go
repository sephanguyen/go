package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

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

func (s *ImportMasterDataService) importProductAssociatedDataAccountingCategory(
	ctx context.Context,
	data []byte) (
	[]*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError,
	error,
) {
	errors := []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{}
	r := csv.NewReader(bytes.NewReader(data))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	header := lines[0]
	if len(header) != 2 {
		return nil, status.Error(codes.InvalidArgument, "csv file invalid format - number of columns should be 2")
	}

	headerTitles := []string{
		"product_id",
		"accounting_category_id",
	}

	err = utils.ValidateCsvHeader(len(headerTitles), header, headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	type productAccountingCategoryWithPositions struct {
		positions                   []int32
		productAccountingCategories []*entities.ProductAccountingCategory
	}
	mapProductAccountingCategoryWithPositions := make(map[pgtype.Text]productAccountingCategoryWithPositions)
	for i, line := range lines[1:] {
		productAccountingCategory, err := CreateProductAccountingCategoryFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
				RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf("unable to parse product associated data item: %s", err),
			})
			continue
		}
		v, ok := mapProductAccountingCategoryWithPositions[productAccountingCategory.ProductID]
		if !ok {
			mapProductAccountingCategoryWithPositions[productAccountingCategory.ProductID] = productAccountingCategoryWithPositions{
				positions:                   []int32{int32(i)},
				productAccountingCategories: []*entities.ProductAccountingCategory{productAccountingCategory},
			}
		} else {
			v.positions = append(v.positions, int32(i))
			v.productAccountingCategories = append(v.productAccountingCategories, productAccountingCategory)
			mapProductAccountingCategoryWithPositions[productAccountingCategory.ProductID] = v
		}
	}
	_ = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for productID, productAccountingCategoryWithPosition := range mapProductAccountingCategoryWithPositions {
			if err = s.ProductAccountingCategoryRepo.Upsert(ctx, tx, productID, productAccountingCategoryWithPosition.productAccountingCategories); err != nil {
				for _, position := range productAccountingCategoryWithPosition.positions {
					errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to create new product accounting category item: %s", err),
					})
				}
				continue
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error importing product associated data accounting category")
		}
		return nil
	})
	return errors, nil
}

func CreateProductAccountingCategoryFromCsv(line []string) (*entities.ProductAccountingCategory, error) {
	const (
		ProductID = iota
		AccountingCategoryID
	)

	productAccountingCategory := &entities.ProductAccountingCategory{}

	if err := multierr.Combine(
		utils.StringToFormatString("product_id", line[ProductID], false, productAccountingCategory.ProductID.Set),
		utils.StringToFormatString("accounting_category_id", line[AccountingCategoryID], false, productAccountingCategory.AccountingCategoryID.Set),
	); err != nil {
		return nil, err
	}

	return productAccountingCategory, nil
}
