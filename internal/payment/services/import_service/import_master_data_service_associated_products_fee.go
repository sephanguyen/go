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

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) importAssociatedProductsFee(
	ctx context.Context,
	data []byte) (
	[]*pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError,
	error,
) {
	var errors []*pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError
	r := csv.NewReader(bytes.NewReader(data))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	header := lines[0]
	headerTitles := []string{
		"package_id",
		"course_id",
		"fee_id",
		"available_from",
		"available_until",
		"is_added_by_default",
	}
	err = utils.ValidateCsvHeader(
		len(headerTitles),
		header,
		headerTitles,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	type associatedProductsByFeeWithPositions struct {
		positions               []int32
		associatedProductsByFee []*entities.PackageCourseFee
	}

	mapAssociatedProductsWithPositions := make(map[pgtype.Text]associatedProductsByFeeWithPositions)
	for i, line := range lines[1:] {
		associatedProductByFee, err := ConvertToAssociatedProductByFeeFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError{
				RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByFee, err),
			})
			continue
		}
		v, ok := mapAssociatedProductsWithPositions[associatedProductByFee.PackageID]
		if !ok {
			mapAssociatedProductsWithPositions[associatedProductByFee.PackageID] = associatedProductsByFeeWithPositions{
				positions:               []int32{int32(i)},
				associatedProductsByFee: []*entities.PackageCourseFee{associatedProductByFee},
			}
		} else {
			v.positions = append(v.positions, int32(i))
			v.associatedProductsByFee = append(v.associatedProductsByFee, associatedProductByFee)
			mapAssociatedProductsWithPositions[associatedProductByFee.PackageID] = v
		}
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for packageID, associatedProductWithPosition := range mapAssociatedProductsWithPositions {
			upsertErr := s.PackageCourseFeeRepo.Upsert(ctx, tx, packageID, associatedProductWithPosition.associatedProductsByFee)
			if upsertErr != nil {
				for _, position := range associatedProductWithPosition.positions {
					errors = append(errors, &pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to create new associated products by fee: %s", upsertErr),
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
		log.Printf("Error when importing associated products by fee: %s", err.Error())
	}

	return errors, nil
}

func ConvertToAssociatedProductByFeeFromCsv(line []string) (*entities.PackageCourseFee, error) {
	const (
		PackageID = iota
		CourseID
		FeeID
		AvailableFrom
		AvailableUntil
		IsAddedByDefault
	)

	associatedProductByFee := &entities.PackageCourseFee{}

	if err := multierr.Combine(
		utils.StringToFormatString("package_id", line[PackageID], false, associatedProductByFee.PackageID.Set),
		utils.StringToFormatString("course_id", line[CourseID], false, associatedProductByFee.CourseID.Set),
		utils.StringToFormatString("fee_id", line[FeeID], false, associatedProductByFee.FeeID.Set),
		utils.StringToDate("available_from", line[AvailableFrom], false, associatedProductByFee.AvailableFrom.Set),
		utils.StringToDate("available_until", line[AvailableUntil], false, associatedProductByFee.AvailableUntil.Set),
		utils.StringToBool("is_added_by_default", line[IsAddedByDefault], false, associatedProductByFee.IsAddedByDefault.Set),
	); err != nil {
		return nil, err
	}

	return associatedProductByFee, nil
}
