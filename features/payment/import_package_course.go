package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) anPackageCoursesValidRequestPayloadWith(ctx context.Context, rowCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.insertSomePackages(ctx)
	if err != nil {
		fmt.Printf("error when insert package %v\n", err.Error())
		return StepStateToContext(ctx, stepState), err
	}
	existingPackages, err := s.selectAllPackages(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	headerTitles := []string{
		"package_id",
		"course_id",
		"mandatory_flag",
		"max_slots_per_course",
		"course_weight",
	}
	headerText := strings.Join(headerTitles, ",")
	validRow1 := fmt.Sprintf("%s,Course-%s,0,2,3", existingPackages[0].PackageID.String, idutil.ULIDNow())
	validRow2 := fmt.Sprintf("%s,Course-%s,1,2,4", existingPackages[0].PackageID.String, idutil.ULIDNow())
	validRow3 := fmt.Sprintf("%s,Course-%s,1,2,3", existingPackages[1].PackageID.String, idutil.ULIDNow())
	validRow4 := fmt.Sprintf("%s,Course-%s,0,2,4", existingPackages[1].PackageID.String, idutil.ULIDNow())
	invalidEmptyRow1 := fmt.Sprintf(",Course-%s,0,2,3", idutil.ULIDNow())
	invalidEmptyRow2 := fmt.Sprintf("%s,,0,2,3", existingPackages[0].PackageID.String)
	invalidValueRow1 := fmt.Sprintf("%s,Course-%s,df,2,3", existingPackages[0].PackageID.String, idutil.ULIDNow())
	invalidValueRow2 := fmt.Sprintf("%s,Course-%s,0,sd,2", existingPackages[0].PackageID.String, idutil.ULIDNow())
	invalidValueRow3 := fmt.Sprintf("sd,Course-%s,0,2,3", idutil.ULIDNow())
	invalidValueRow4 := fmt.Sprintf("%s,Course-%s,,2,3", existingPackages[0].PackageID.String, idutil.ULIDNow())
	invalidValueRow5 := fmt.Sprintf("%s,Course-%s,0,2,", existingPackages[0].PackageID.String, idutil.ULIDNow())

	stepState.ValidCsvRows = []string{}
	stepState.InvalidCsvRows = []string{}
	switch rowCondition {
	case "all valid rows":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, validRow1, validRow2)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2}
	case "empty value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, invalidEmptyRow1, invalidEmptyRow2)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2}
	case "invalid value row":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s`, headerText, invalidValueRow1, invalidValueRow2)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		}
		stepState.InvalidCsvRows = []string{invalidValueRow1, invalidValueRow2}
	case "valid and invalid rows":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s
			%s`, headerText, validRow1, validRow2, validRow3, validRow4, invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, invalidValueRow5)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		}
		stepState.ValidCsvRows = []string{validRow1, validRow2, validRow3, validRow4}
		stepState.InvalidCsvRows = []string{invalidEmptyRow1, invalidEmptyRow2, invalidValueRow1, invalidValueRow2, invalidValueRow3, invalidValueRow4, invalidValueRow5}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anPackageCoursesInvalidRequestPayload(ctx context.Context, invalidFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	headerTitles := []string{
		"package_id",
		"course_id",
		"mandatory_flag",
		"max_slots_per_course",
		"course_weight",
	}
	headerTitleMisfield := []string{
		"package_id",
		"course_id",
		"mandatory_flag",
		"max_slots_per_course",
	}
	headerTitlesWrongNameOfID := []string{
		"package_idabc",
		"course_id",
		"mandatory_flag",
		"max_slots_per_course",
		"course_weight",
	}
	headerText := strings.Join(headerTitles, ",")
	headerTitleMisfieldText := strings.Join(headerTitleMisfield, ",")
	headerTitlesWrongNameOfIDText := strings.Join(headerTitlesWrongNameOfID, ",")
	validRow1 := fmt.Sprintf("1,Course-%s,0,3", idutil.ULIDNow())
	switch invalidFormat {
	case "no data":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		}
	case "header only":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload:                   []byte(headerText),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		}
	case "number of column is not equal 5":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s`, headerTitleMisfieldText, validRow1)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		}
	case "wrong package_id column name in header":
		stepState.Request = &pb.ImportProductAssociatedDataRequest{
			Payload: []byte(fmt.Sprintf(`%s
				%s`, headerTitlesWrongNameOfIDText, validRow1)),
			ProductAssociatedDataType: pb.ProductAssociatedDataType_PRODUCT_ASSOCIATED_DATA_TYPE_COURSE,
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importingPackageCourses(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewImportMasterDataServiceClient(s.PaymentConn).
		ImportProductAssociatedData(contextWithToken(ctx), stepState.Request.(*pb.ImportProductAssociatedDataRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theValidPackageCoursesLinesAreImportedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	allProductCourses, err := s.selectAllPackageCourses(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	const (
		ProductID = iota
		CourseID
		MandatoryFlag
		MaxSlot
		CourseWeight
	)
	for _, row := range stepState.ValidCsvRows {
		var productCourse entities.PackageCourse
		values := strings.Split(row, ",")

		err = multierr.Combine(
			utils.StringToFormatString("package_id", values[ProductID], false, productCourse.PackageID.Set),
			utils.StringToFormatString("course_id", values[CourseID], false, productCourse.CourseID.Set),
			utils.StringToBool("mandatory_flag", values[MandatoryFlag], false, productCourse.MandatoryFlag.Set),
			utils.StringToInt("course_weight", values[CourseWeight], false, productCourse.CourseWeight.Set),
			utils.StringToInt("max_slots_per_course", values[MaxSlot], false, productCourse.CourseWeight.Set),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		findingProduct := foundPackageCourses(productCourse, allProductCourses)
		if findingProduct.PackageID.Get() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("not found product in list")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theInvalidPackageCoursesLinesAreReturnedWithError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.ImportProductAssociatedDataRequest)
	reqSplit := strings.Split(string(req.Payload), "\n")
	resp := stepState.Response.(*pb.ImportProductAssociatedDataResponse)
	for _, row := range stepState.InvalidCsvRows {
		found := false
		for _, e := range resp.Errors {
			if strings.TrimSpace(reqSplit[e.RowNumber-1]) == row {
				found = true
				break
			}
		}
		if !found {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid line is not returned in response")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectAllPackageCourses(ctx context.Context) ([]*entities.PackageCourse, error) {
	var allEntities []*entities.PackageCourse
	const getPackageCourses = `
SELECT package_id, course_id, mandatory_flag, course_weight, max_slots_per_course, created_at 
FROM package_course`
	rows, err := s.FatimaDBTrace.Query(
		ctx,
		getPackageCourses,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query package_course")
	}
	defer rows.Close()
	for rows.Next() {
		var i entities.PackageCourse
		if err := rows.Scan(
			&i.PackageID,
			&i.CourseID,
			&i.MandatoryFlag,
			&i.CourseWeight,
			&i.MaxSlotsPerCourse,
			&i.CreatedAt,
		); err != nil {
			return nil, errors.WithMessage(err, "rows.Scan Product")
		}
		allEntities = append(allEntities, &i)
	}
	return allEntities, nil
}

func foundPackageCourses(productCourseNeedFinding entities.PackageCourse, productCourseList []*entities.PackageCourse) (finding entities.PackageCourse) {
	for i, productCourse := range productCourseList {
		if productCourseNeedFinding.PackageID == productCourse.PackageID &&
			productCourseNeedFinding.CourseID == productCourse.CourseID &&
			productCourseNeedFinding.MandatoryFlag == productCourse.MandatoryFlag &&
			productCourseNeedFinding.MaxSlotsPerCourse == productCourse.MaxSlotsPerCourse &&
			productCourseNeedFinding.CourseWeight == productCourse.CourseWeight {
			finding = *productCourseList[i]
			break
		}
	}
	return finding
}
