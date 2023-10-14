package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

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

type lineWithNumber struct {
	number int
	line   []string
}

func (s *ImportMasterDataService) ImportProductPrice(ctx context.Context, req *pb.ImportProductPriceRequest) (*pb.ImportProductPriceResponse, error) {
	errors := []*pb.ImportProductPriceResponse_ImportProductPriceError{}

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
		"product_id",
		"billing_schedule_period_id",
		"quantity",
		"price",
		"price_type",
	}
	if err := utils.ValidateCsvHeader(len(headerTitles), header, headerTitles); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	mapProductIDWithProductPriceLines := make(map[string][]lineWithNumber)

	// first line is header
	for i, line := range lines[1:] {
		// line[0] = product_id
		if v, exist := mapProductIDWithProductPriceLines[strings.TrimSpace(line[0])]; exist {
			mapProductIDWithProductPriceLines[strings.TrimSpace(line[0])] = append(v, lineWithNumber{i + 2, trimSpaces(line)})
		} else {
			mapProductIDWithProductPriceLines[strings.TrimSpace(line[0])] = []lineWithNumber{{i + 2, trimSpaces(line)}}
		}
	}

	missingDefaultPriceProductIDs := make([]string, 0, len(mapProductIDWithProductPriceLines))
	for productID, ppLines := range mapProductIDWithProductPriceLines {
		hasDefaultPrice := false
		for _, line := range ppLines {
			// line[4] = price_type
			if line.line[4] == pb.ProductPriceType_DEFAULT_PRICE.String() {
				hasDefaultPrice = true
			}
		}
		if !hasDefaultPrice {
			missingDefaultPriceProductIDs = append(missingDefaultPriceProductIDs, productID)
		}
	}
	if len(missingDefaultPriceProductIDs) > 0 {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error when import product prices without DEFAULT_PRICE value with product_ids: %s", strings.Join(missingDefaultPriceProductIDs, ",")))
	}

	_ = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for productID, ppLines := range mapProductIDWithProductPriceLines {
			if productID == "" {
				for _, v := range ppLines {
					errors = append(errors, &pb.ImportProductPriceResponse_ImportProductPriceError{
						RowNumber: int32(v.number),
						Error:     "product_id is empty",
					})
				}
			} else {
				productID := productID
				if err := s.ProductPriceRepo.DeleteByProductID(ctx, tx, pgtype.Text{Status: pgtype.Present, String: productID}); err != nil {
					for _, v := range ppLines {
						errors = append(errors, &pb.ImportProductPriceResponse_ImportProductPriceError{
							RowNumber: int32(v.number),
							Error:     "something wrong when delete product_price",
						})
					}
					continue
				}
				for _, v := range ppLines {
					var productPrice entities.ProductPrice
					productPrice, err := ProductPriceFromCsv(v.line)
					if err != nil {
						errors = append(errors, &pb.ImportProductPriceResponse_ImportProductPriceError{
							RowNumber: int32(v.number),
							Error:     fmt.Sprintf("unable to parse product_price item: %s", err),
						})
						continue
					}

					if err := s.ProductPriceRepo.Create(ctx, tx, &productPrice); err != nil {
						errors = append(errors, &pb.ImportProductPriceResponse_ImportProductPriceError{
							RowNumber: int32(v.number),
							Error:     fmt.Sprintf("unable to create new product_price item: %s", err),
						})
					}
				}
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error importing product price")
		}
		return nil
	})

	sort.Slice(errors, func(i, j int) bool {
		return errors[i].RowNumber < errors[j].RowNumber
	})

	return &pb.ImportProductPriceResponse{
		Errors: errors,
	}, nil
}

func ProductPriceFromCsv(line []string) (productPrice entities.ProductPrice, err error) {
	const (
		ProductID = iota
		BillingSchedulePeriodID
		Quantity
		Price
		PriceType
	)

	if err = multierr.Combine(
		utils.StringToFormatString("product_id", line[ProductID], false, productPrice.ProductID.Set),
		utils.StringToFormatString("billing_schedule_period_id", line[BillingSchedulePeriodID], true, productPrice.BillingSchedulePeriodID.Set),
		utils.StringToInt("quantity", line[Quantity], true, productPrice.Quantity.Set),
		utils.StringToFloat("price", line[Price], false, productPrice.Price.Set),
		utils.StringToFormatString("price_type", line[PriceType], false, productPrice.PriceType.Set),
	); err != nil {
		return productPrice, err
	}

	return productPrice, nil
}

func trimSpaces(line []string) []string {
	for i, v := range line {
		line[i] = strings.TrimSpace(v)
	}
	return line
}
