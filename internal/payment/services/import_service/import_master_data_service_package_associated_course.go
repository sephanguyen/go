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

func (s *ImportMasterDataService) packageCourseModifier(
	ctx context.Context,
	data []byte) (
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
		"package_id",
		"course_id",
		"mandatory_flag",
		"max_slots_per_course",
		"course_weight",
	}
	err = utils.ValidateCsvHeader(len(headerTitles), lines[0], headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	var errors []*pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError
	type packageCourseWithPositions struct {
		positions      []int32
		packageCourses []*entities.PackageCourse
	}

	mapPackageCourseWithPositions := make(map[pgtype.Text]packageCourseWithPositions)
	for i, line := range lines[1:] {
		var packageCourse entities.PackageCourse
		packageCourse, err = PackageCourseFromCsv(line)
		if err != nil {
			errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
				RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
				Error:     fmt.Sprintf("unable to parse package course item: %s", err),
			})
			continue
		}
		v, ok := mapPackageCourseWithPositions[packageCourse.PackageID]
		if !ok {
			mapPackageCourseWithPositions[packageCourse.PackageID] = packageCourseWithPositions{
				positions:      []int32{int32(i)},
				packageCourses: []*entities.PackageCourse{&packageCourse},
			}
		} else {
			v.positions = append(v.positions, int32(i))
			v.packageCourses = append(v.packageCourses, &packageCourse)
			mapPackageCourseWithPositions[packageCourse.PackageID] = v
		}
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for productID, productCoursesWithPosition := range mapPackageCourseWithPositions {
			if err = s.PackageCourseRepo.Upsert(ctx, tx, productID, productCoursesWithPosition.packageCourses); err != nil {
				for _, position := range productCoursesWithPosition.positions {
					errors = append(errors, &pb.ImportProductAssociatedDataResponse_ImportProductAssociatedDataError{
						RowNumber: position + 2,
						Error:     fmt.Sprintf("unable to import package course item: %s", err),
					})
				}
				continue
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("error importing package course")
		}
		return nil
	})

	return errors, nil
}

func PackageCourseFromCsv(line []string) (productCourse entities.PackageCourse, err error) {
	const (
		ProductID = iota
		CourseID
		MandatoryFlag
		MaxSlotsPerCourse
		CourseWeight
	)

	err = multierr.Combine(
		utils.StringToFormatString("package_id", line[ProductID], false, productCourse.PackageID.Set),
		utils.StringToFormatString("course_id", line[CourseID], false, productCourse.CourseID.Set),
		utils.StringToBool("mandatory_flag", line[MandatoryFlag], false, productCourse.MandatoryFlag.Set),
		utils.StringToInt("max_slots_per_course", line[MaxSlotsPerCourse], false, productCourse.MaxSlotsPerCourse.Set),
		utils.StringToInt("course_weight", line[CourseWeight], false, productCourse.CourseWeight.Set),
	)
	return
}
