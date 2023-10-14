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

func (s *ImportMasterDataService) importProductAssociatedDataDiscount(
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
		"discount_id",
	}
	err = utils.ValidateCsvHeader(len(headerTitles), header, headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	type discountAssocationWithPositions struct {
		positions        []int32
		productDiscounts []*entities.ProductDiscount
	}

	mapProductDiscountWithPositions := make(map[pgtype.Text]discountAssocationWithPositions)

	for i, line := range lines[1:] {
		productDiscount, err := ProductDiscountFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
				RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf("unable to parse product discount item: %s", err),
			})
			continue
		}

		v, ok := mapProductDiscountWithPositions[productDiscount.ProductID]
		if !ok {
			mapProductDiscountWithPositions[productDiscount.ProductID] = discountAssocationWithPositions{
				positions:        []int32{int32(i)},
				productDiscounts: []*entities.ProductDiscount{productDiscount},
			}
		} else {
			v.positions = append(v.positions, int32(i))
			v.productDiscounts = append(v.productDiscounts, productDiscount)
			mapProductDiscountWithPositions[productDiscount.ProductID] = v
		}
	}

	_ = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for productID, productdiscountAssocationWithPosition := range mapProductDiscountWithPositions {
			if err = s.ProductDiscountRepo.Upsert(ctx, tx, productID, productdiscountAssocationWithPosition.productDiscounts); err != nil {
				for _, position := range productdiscountAssocationWithPosition.positions {
					errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to import product discount item: %s", err),
					})
				}
				continue
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error import product discount")
		}
		return nil
	})

	return errors, nil
}

func ProductDiscountFromCsv(line []string) (*entities.ProductDiscount, error) {
	const (
		ProductID = iota
		DiscountID
	)
	productDiscount := &entities.ProductDiscount{}
	if err := multierr.Combine(
		utils.StringToFormatString("product_id", line[ProductID], false, productDiscount.ProductID.Set),
		utils.StringToFormatString("discount_id", line[DiscountID], false, productDiscount.DiscountID.Set),
	); err != nil {
		return nil, err
	}
	return productDiscount, nil
}
