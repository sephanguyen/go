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

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportProductGroupMapping(ctx context.Context, req *pb.ImportProductGroupMappingRequest) (*pb.ImportProductGroupMappingResponse, error) {
	errors := []*pb.ImportProductGroupMappingResponse_ImportProductGroupMappingError{}

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
		"product_id",
	}
	if err := utils.ValidateCsvHeader(len(headerTitles), header, headerTitles); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	type productGroupMappingWithPositions struct {
		positions           []int32
		productGroupMapping []*entities.ProductGroupMapping
	}

	mapProductGroupMappingWithPositions := make(map[pgtype.Text]productGroupMappingWithPositions)
	for i, line := range lines[1:] {
		productGroupMappingFromCSV, err := ProductGroupMappingFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportProductGroupMappingResponse_ImportProductGroupMappingError{
				RowNumber: int32(i) + 2,
				Error:     fmt.Sprintf("unable to import product group mapping item: %s", err),
			})
			continue
		}
		v, ok := mapProductGroupMappingWithPositions[productGroupMappingFromCSV.ProductGroupID]
		if !ok {
			mapProductGroupMappingWithPositions[productGroupMappingFromCSV.ProductGroupID] = productGroupMappingWithPositions{
				positions:           []int32{int32(i)},
				productGroupMapping: []*entities.ProductGroupMapping{productGroupMappingFromCSV},
			}
		} else {
			v.positions = append(v.positions, int32(i))
			v.productGroupMapping = append(v.productGroupMapping, productGroupMappingFromCSV)
			mapProductGroupMappingWithPositions[productGroupMappingFromCSV.ProductGroupID] = v
		}
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for productGroupID, productGroupMappingWithPosition := range mapProductGroupMappingWithPositions {
			upsertErr := s.ProductGroupMappingRepo.Upsert(ctx, tx, productGroupID, productGroupMappingWithPosition.productGroupMapping)
			if upsertErr != nil {
				for _, position := range productGroupMappingWithPosition.positions {
					errors = append(errors, &pb.ImportProductGroupMappingResponse_ImportProductGroupMappingError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to import product group mapping item: %s", err),
					})
				}
				continue
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error when importing product group mapping: %s", err.Error())
	}

	return &pb.ImportProductGroupMappingResponse{
		Errors: errors,
	}, nil
}

func ProductGroupMappingFromCsv(line []string) (*entities.ProductGroupMapping, error) {
	const (
		ProductGroupID = iota
		ProductID
	)
	productGroupMapping := &entities.ProductGroupMapping{}
	if err := multierr.Combine(
		utils.StringToFormatString("product_group_id", line[ProductGroupID], false, productGroupMapping.ProductGroupID.Set),
		utils.StringToFormatString("product_id", line[ProductID], false, productGroupMapping.ProductID.Set),
	); err != nil {
		return nil, err
	}
	return productGroupMapping, nil
}
