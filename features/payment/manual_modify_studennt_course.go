package payment

import (
	"context"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) prepareDataForInsertStudentCourse(ctx context.Context) (context.Context, error) {
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
	studentCourses := make([]*pb.StudentCourseData, 0, 3)
	for _, courseID = range courseIDs {
		studentCourses = append(studentCourses, &pb.StudentCourseData{
			CourseId:   courseID,
			LocationId: locationID,
			StartDate:  timestamppb.New(startDate),
			EndDate:    timestamppb.New(endDate),
			IsChanged:  true,
		})
	}
	stepState.Request = &pb.ManualUpsertStudentCourseRequest{
		StudentId:      userID,
		StudentCourses: studentCourses,
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) manualModifyStudentCourse(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewCourseServiceClient(s.PaymentConn).ManualUpsertStudentCourse(contextWithToken(ctx), stepState.Request.(*pb.ManualUpsertStudentCourseRequest))
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
