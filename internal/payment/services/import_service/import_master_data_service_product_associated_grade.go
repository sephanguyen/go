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

func (s *ImportMasterDataService) importProductAssociatedDataGrade(
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
		"grade_id",
	}
	err = utils.ValidateCsvHeader(len(headerTitles), header, headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	type productGradeWithPositions struct {
		positions     []int32
		productGrades []*entities.ProductGrade
	}

	mapProductGradeWithPositions := make(map[pgtype.Text]productGradeWithPositions)
	for i, line := range lines[1:] {
		productGrade, err := CreateProductGradeFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
				RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf("unable to parse product associated data item: %s", err),
			})
			continue
		}
		v, ok := mapProductGradeWithPositions[productGrade.ProductID]
		if !ok {
			mapProductGradeWithPositions[productGrade.ProductID] = productGradeWithPositions{
				positions:     []int32{int32(i)},
				productGrades: []*entities.ProductGrade{productGrade},
			}
		} else {
			v.positions = append(v.positions, int32(i))
			v.productGrades = append(v.productGrades, productGrade)
			mapProductGradeWithPositions[productGrade.ProductID] = v
		}
	}

	_ = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for productID, productGradesWithPosition := range mapProductGradeWithPositions {
			if err = s.ProductGradeRepo.Upsert(ctx, tx, productID, productGradesWithPosition.productGrades); err != nil {
				for _, position := range productGradesWithPosition.positions {
					errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to create new product grade item: %s", err),
					})
				}
				continue
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error importing product associated data grade")
		}
		return nil
	})

	return errors, nil
}

func CreateProductGradeFromCsv(line []string) (*entities.ProductGrade, error) {
	const (
		ProductID = iota
		GradeID
	)

	productGrade := &entities.ProductGrade{}

	if err := multierr.Combine(
		utils.StringToFormatString("product_id", line[ProductID], false, productGrade.ProductID.Set),
		utils.StringToFormatString("grade_id", line[GradeID], false, productGrade.GradeID.Set),
	); err != nil {
		return nil, err
	}

	return productGrade, nil
}
