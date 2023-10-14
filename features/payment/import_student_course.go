package payment

import (
	"context"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataForImportStudentCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		locationID string
		userID     string
		courseIDs  []string
		err        error
	)
	ctx, err = s.subscribeStudentCourseEventSync(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defaultOptionPrepareData := optionToPrepareDataForCreateOrderPackageOneTime{
		insertStudent:                    true,
		insertPackage:                    false,
		insertPackageCourse:              false,
		insertCourse:                     true,
		insertProductPrice:               false,
		insertProductLocation:            false,
		insertLocation:                   false,
		insertProductGrade:               false,
		insertPackageQuantityTypeMapping: false,
		insertCourseAccessLocation:       true,
		insertUserAccessLocation:         true,
	}
	_,
		locationID,
		_,
		courseIDs,
		userID,
		err = s.insertAllDataForInsertOrderPackageOneTime(ctx, defaultOptionPrepareData)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	startDate := time.Now()
	endDate := startDate.AddDate(1, 0, 0)
	csvFile := "student_id,course_id,location_id,start_date,end_date"
	for _, courseID = range courseIDs {
		csvFile += fmt.Sprintf("\n%s,%s,%s,%s,%s", userID, courseID, locationID, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))
	}
	stepState.Request = &pb.ImportStudentCoursesRequest{
		Payload: []byte(csvFile),
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importStudentCourse(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.PaymentConn).ImportStudentCourses(contextWithToken(ctx), stepState.Request.(*pb.ImportStudentCoursesRequest))
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(stepState.Response.(*pb.ImportStudentCoursesResponse).Errors) > 0 {
		err = convertErrorResToErr(stepState.Response.(*pb.ImportStudentCoursesResponse).Errors)
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.PaymentConn).ImportStudentCourses(contextWithToken(ctx), stepState.Request.(*pb.ImportStudentCoursesRequest))
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(stepState.Response.(*pb.ImportStudentCoursesResponse).Errors) > 0 {
		err = convertErrorResToErr(stepState.Response.(*pb.ImportStudentCoursesResponse).Errors)
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func convertErrorResToErr(errors []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError) error {
	var errorMessage string
	for _, coursesError := range errors {
		errorMessage += fmt.Sprintf("Row %v with error message %v \n", coursesError.RowNumber, coursesError.Error)
	}
	return fmt.Errorf(errorMessage)
}
