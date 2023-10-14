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

/* ImportPackageDiscountCourseMapping function that imports package_discount_course_mappings records from a CSV file.
*  this will be used for validating combo discount along with package course combination
 */
func (s *ImportMasterDataService) ImportPackageDiscountCourseMapping(ctx context.Context, req *pb.ImportPackageDiscountCourseMappingRequest) (*pb.ImportPackageDiscountCourseMappingResponse, error) {
	errors := []*pb.ImportPackageDiscountCourseMappingResponse_ImportPackageDiscountCourseMappingError{}

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
		"package_id",
		"course_combination_ids",
		"discount_tag_id",
		"is_archived",
		"product_group_id",
	}

	if err := utils.ValidateCsvHeader(len(headerTitles), header, headerTitles); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	type packageDiscountCourseMappingSetting struct {
		positions                    []int32
		packageDiscountCourseMapping []*entities.PackageDiscountCourseMapping
	}

	mapPackageDiscountCourseMappingSetting := make(map[pgtype.Text]packageDiscountCourseMappingSetting)

	for i, line := range lines[1:] {
		packageDiscountCourseMappingFromCSV, err := PackageDiscountCourseMappingFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportPackageDiscountCourseMappingResponse_ImportPackageDiscountCourseMappingError{
				RowNumber: int32(i) + 2,
				Error:     fmt.Sprintf("unable to parse package discount course mapping: %s", err),
			})
			continue
		}

		// Check if the packageID exists in the map, if not, create a new entry for it.
		v, ok := mapPackageDiscountCourseMappingSetting[packageDiscountCourseMappingFromCSV.PackageID]
		if !ok {
			mapPackageDiscountCourseMappingSetting[packageDiscountCourseMappingFromCSV.PackageID] = packageDiscountCourseMappingSetting{
				positions:                    []int32{int32(i)},
				packageDiscountCourseMapping: []*entities.PackageDiscountCourseMapping{packageDiscountCourseMappingFromCSV},
			}
		} else {
			// If the packageID exists, append the position and package discount course mapping to the existing entry.
			v.positions = append(v.positions, int32(i))
			v.packageDiscountCourseMapping = append(v.packageDiscountCourseMapping, packageDiscountCourseMappingFromCSV)
			mapPackageDiscountCourseMappingSetting[packageDiscountCourseMappingFromCSV.PackageID] = v
		}
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for packageID, packageDiscountCourseMappingWithPosition := range mapPackageDiscountCourseMappingSetting {
			upsertErr := s.PackageDiscountCourseMappingRepo.Upsert(ctx, tx, packageID, packageDiscountCourseMappingWithPosition.packageDiscountCourseMapping)
			if upsertErr != nil {
				for _, position := range packageDiscountCourseMappingWithPosition.positions {
					errors = append(errors, &pb.ImportPackageDiscountCourseMappingResponse_ImportPackageDiscountCourseMappingError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to import package discount course mapping: %s", upsertErr),
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
		log.Printf("Error when importing package discount course mapping: %s", err.Error())
	}

	return &pb.ImportPackageDiscountCourseMappingResponse{
		Errors: errors,
	}, nil
}

// PackageDiscountCourseMappingFromCsv function that helps convert csv line format and map to an entity
func PackageDiscountCourseMappingFromCsv(line []string) (*entities.PackageDiscountCourseMapping, error) {
	const (
		PackageID = iota
		CourseCombination
		DiscountTagID
		IsArchived
		ProductGroupID
	)

	packageDiscountCourseMapping := &entities.PackageDiscountCourseMapping{}

	if err := multierr.Combine(
		utils.StringToFormatString("package_id", line[PackageID], false, packageDiscountCourseMapping.PackageID.Set),
		utils.StringToFormatString("course_combination_ids_ids", line[CourseCombination], false, packageDiscountCourseMapping.CourseCombinationIDs.Set),
		utils.StringToFormatString("discount_tag_id", line[DiscountTagID], false, packageDiscountCourseMapping.DiscountTagID.Set),
		utils.StringToBool("is_archived", line[IsArchived], true, packageDiscountCourseMapping.IsArchived.Set),
		utils.StringToFormatString("product_group_id", line[ProductGroupID], false, packageDiscountCourseMapping.ProductGroupID.Set),
	); err != nil {
		return nil, err
	}

	return packageDiscountCourseMapping, nil
}
