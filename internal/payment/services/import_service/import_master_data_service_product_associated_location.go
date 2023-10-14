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

type productLocationWithPositions struct {
	positions        []int32
	productLocations []*entities.ProductLocation
}

func (s *ImportMasterDataService) importProductLocationModifier(ctx context.Context, data []byte) (
	[]*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError,
	error,
) {
	r := csv.NewReader(bytes.NewReader(data))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	headerTitles := []string{
		"product_id",
		"location_id",
	}
	err = utils.ValidateCsvHeader(len(headerTitles), lines[0], headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	var errors []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError

	mapProductLocationWithPositions := make(map[pgtype.Text]productLocationWithPositions)
	for i, line := range lines[1:] {
		var productLocation entities.ProductLocation
		productLocation, err = ReadProductLocationFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
				RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf("unable to parse product location item: %s", err),
			})
			continue
		}
		v, ok := mapProductLocationWithPositions[productLocation.ProductID]
		if !ok {
			mapProductLocationWithPositions[productLocation.ProductID] = productLocationWithPositions{
				positions:        []int32{int32(i)},
				productLocations: []*entities.ProductLocation{&productLocation},
			}
			continue
		}
		v.positions = append(v.positions, int32(i))
		v.productLocations = append(v.productLocations, &productLocation)
		mapProductLocationWithPositions[productLocation.ProductID] = v
	}
	_ = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for productID, productLocationsWithPosition := range mapProductLocationWithPositions {
			err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
				return s.ProductLocationRepo.Replace(ctx, tx, productID, productLocationsWithPosition.productLocations)
			})
			if err = s.ProductLocationRepo.Replace(ctx, tx, productID, productLocationsWithPosition.productLocations); err != nil {
				for _, position := range productLocationsWithPosition.positions {
					errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
						RowNumber: position + 2, // position = 0 <=> line number 2 in csv file
						Error:     fmt.Sprintf("unable to import product location item: %s", err),
					})
				}
				continue
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error importing product location")
		}
		return nil
	})

	return errors, nil
}

func ReadProductLocationFromCsv(line []string) (productLocation entities.ProductLocation, err error) {
	const (
		ProductID = iota
		LocationID
	)

	err = multierr.Combine(
		utils.StringToFormatString("product_id", line[ProductID], false, productLocation.ProductID.Set),
		utils.StringToFormatString("location_id", line[LocationID], false, productLocation.LocationID.Set),
	)
	return
}
