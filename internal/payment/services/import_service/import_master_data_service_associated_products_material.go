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

func (s *ImportMasterDataService) importAssociatedProductsMaterial(
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
		"material_id",
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

	type associatedProductsByMaterialWithPositions struct {
		positions                    []int32
		associatedProductsByMaterial []*entities.PackageCourseMaterial
	}

	mapAssociatedProductsWithPositions := make(map[pgtype.Text]associatedProductsByMaterialWithPositions)
	for i, line := range lines[1:] {
		associatedProductByMaterial, err := ConvertToAssociatedProductByMaterialFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError{
				RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf(constant.UnableToParseAssociatedProductsByMaterial, err),
			})
			continue
		}
		v, ok := mapAssociatedProductsWithPositions[associatedProductByMaterial.PackageID]
		if !ok {
			mapAssociatedProductsWithPositions[associatedProductByMaterial.PackageID] = associatedProductsByMaterialWithPositions{
				positions:                    []int32{int32(i)},
				associatedProductsByMaterial: []*entities.PackageCourseMaterial{associatedProductByMaterial},
			}
		} else {
			v.positions = append(v.positions, int32(i))
			v.associatedProductsByMaterial = append(v.associatedProductsByMaterial, associatedProductByMaterial)
			mapAssociatedProductsWithPositions[associatedProductByMaterial.PackageID] = v
		}
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for packageID, associatedProductWithPosition := range mapAssociatedProductsWithPositions {
			upsertErr := s.PackageCourseMaterialRepo.Upsert(ctx, tx, packageID, associatedProductWithPosition.associatedProductsByMaterial)
			if upsertErr != nil {
				for _, position := range associatedProductWithPosition.positions {
					errors = append(errors, &pb.ImportAssociatedProductsResponse_ImportAssociatedProductsError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to create new associated products by material: %s", upsertErr),
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
		log.Printf("Error when importing associated products by material: %s", err.Error())
	}

	return errors, nil
}

func ConvertToAssociatedProductByMaterialFromCsv(line []string) (*entities.PackageCourseMaterial, error) {
	const (
		PackageID = iota
		CourseID
		MaterialID
		AvailableFrom
		AvailableUntil
		IsAddedByDefault
	)

	associatedProductByMaterial := &entities.PackageCourseMaterial{}

	if err := multierr.Combine(
		utils.StringToFormatString("package_id", line[PackageID], false, associatedProductByMaterial.PackageID.Set),
		utils.StringToFormatString("course_id", line[CourseID], false, associatedProductByMaterial.CourseID.Set),
		utils.StringToFormatString("material_id", line[MaterialID], false, associatedProductByMaterial.MaterialID.Set),
		utils.StringToDate("available_from", line[AvailableFrom], false, associatedProductByMaterial.AvailableFrom.Set),
		utils.StringToDate("available_until", line[AvailableUntil], false, associatedProductByMaterial.AvailableUntil.Set),
		utils.StringToBool("is_added_by_default", line[IsAddedByDefault], false, associatedProductByMaterial.IsAddedByDefault.Set),
	); err != nil {
		return nil, err
	}

	return associatedProductByMaterial, nil
}
