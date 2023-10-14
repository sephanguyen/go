package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/payment/mockdata"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) prepareDataForImportStudentClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var (
		locationID string
		userID     string
		courseIDs  []string
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
	request := &pb.ImportStudentCoursesRequest{
		Payload: []byte(csvFile),
	}
	stepState.RequestSentAt = time.Now()
	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	response, responseErr := pb.NewCourseServiceClient(s.PaymentConn).ImportStudentCourses(contextWithToken(ctx), request)
	if responseErr != nil {
		fmt.Println(responseErr.Error())
		return StepStateToContext(ctx, stepState), responseErr
	}
	if len(response.Errors) > 0 {
		err = convertErrorResToErr(response.Errors)
		return StepStateToContext(ctx, stepState), err
	}
	classCsvFile := "student_id,course_id,class_id"
	for _, courseID := range courseIDs {
		classID, err := mockdata.InsertOneClass(ctx, s.FatimaDBTrace, courseID, locationID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		classCsvFile += fmt.Sprintf("\n%s,%s,%s", userID, courseID, classID)
	}
	stepState.Request = &pb.ImportStudentClassesRequest{
		Payload:    []byte(classCsvFile),
		IsAddClass: true,
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importStudentClassForInsert(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	req := stepState.Request.(*pb.ImportStudentClassesRequest)
	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.PaymentConn).ImportStudentClasses(contextWithToken(ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(stepState.Response.(*pb.ImportStudentClassesResponse).Errors) > 0 {
		err = convertStudentClassErrorResToErr(stepState.Response.(*pb.ImportStudentClassesResponse).Errors)
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importStudentClassForDelete(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	req := stepState.Request.(*pb.ImportStudentClassesRequest)
	req.IsAddClass = true
	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.PaymentConn).ImportStudentClasses(contextWithToken(ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(stepState.Response.(*pb.ImportStudentClassesResponse).Errors) > 0 {
		err = convertStudentClassErrorResToErr(stepState.Response.(*pb.ImportStudentClassesResponse).Errors)
		return StepStateToContext(ctx, stepState), err
	}
	req.IsAddClass = false
	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.PaymentConn).ImportStudentClasses(contextWithToken(ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(stepState.Response.(*pb.ImportStudentClassesResponse).Errors) > 0 {
		err = convertStudentClassErrorResToErr(stepState.Response.(*pb.ImportStudentClassesResponse).Errors)
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func convertStudentClassErrorResToErr(errors []*pb.ImportStudentClassesResponse_ImportStudentClassesError) error {
	var errorMessage string
	for _, coursesError := range errors {
		errorMessage += fmt.Sprintf("Row %v with error message %v \n", coursesError.RowNumber, coursesError.Error)
	}
	return fmt.Errorf(errorMessage)
}
