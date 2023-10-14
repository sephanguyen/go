package bob

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	types "github.com/gogo/protobuf/types"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/manabie-com/backend/internal/golibs/i18n"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) userUpdatesProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()

	err := s.createUserDeviceTokenCreatedSubscription(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createUserDeviceTokenCreatedSubscription: %v", err)
	}

	stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).UpdateProfile(s.signedCtx(ctx), stepState.Request.(*pb.UpdateProfileRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserUpdatesProfile(ctx context.Context) (context.Context, error) {
	return s.userUpdatesProfile(ctx)
}
func (s *suite) aValidUpdatesProfileRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	num := rand.Int()
	stepState.Request = &pb.UpdateProfileRequest{
		Name:             fmt.Sprintf("valid-student-%d", num),
		Grade:            "G11",
		TargetUniversity: fmt.Sprintf("target-university-%d", num),
		Avatar:           "http://valid-avatar",
		Birthday:         &types.Timestamp{Seconds: time.Now().Unix()},
		Biography:        "sort biography",
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AValidUpdatesProfileRequest(ctx context.Context) (context.Context, error) {
	return s.aValidUpdatesProfileRequest(ctx)
}
func (s *suite) bobMustRecordsStudentsProfileUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}

	studentRsp, err := pb.NewStudentClient(s.Conn).GetStudentProfile(s.signedCtx(ctx), &pb.GetStudentProfileRequest{
		StudentIds: []string{
			stepState.CurrentStudentID,
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("GetStudentProfile: %w", err)
	}
	if len(studentRsp.Datas) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("don't have any profile")
	}
	student := studentRsp.Datas[0].Profile
	studentReq := stepState.Request.(*pb.UpdateProfileRequest)
	if student.Name != studentReq.Name {
		return StepStateToContext(ctx, stepState), errors.New("name was not updated")
	}
	localGradeMap := i18n.InGradeMap[student.Country]
	clientGradeMap := i18n.OutGradeMap[student.Country]
	reqGrade := localGradeMap[studentReq.Grade]
	expectedGrade := clientGradeMap[reqGrade]

	if student.Grade != expectedGrade {
		return StepStateToContext(ctx, stepState), errors.New("grade was not updated")
	}

	if student.Avatar != studentReq.Avatar {
		return StepStateToContext(ctx, stepState), errors.New("avatar was not updated")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) BobMustRecordsStudentsProfileUpdate(ctx context.Context) (context.Context, error) {
	return s.bobMustRecordsStudentsProfileUpdate(ctx)
}
func (s *suite) bobMustNotUpdateStudentsProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), errors.New("bob must update student profile")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hisOwnedStudentUUID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return ctx, err
	}
	stepState.CurrentStudentID = t.Subject()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) HisOwnedStudentUUID(ctx context.Context) (context.Context, error) {
	return s.hisOwnedStudentUUID(ctx)
}

func (s *suite) ARandomNumber(ctx context.Context) (context.Context, error) {
	return s.aRandomNumber(ctx)
}

func (s *suite) anInvalidStudentUUID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = "invalid-student-UUID"
	return StepStateToContext(ctx, stepState), nil
}
