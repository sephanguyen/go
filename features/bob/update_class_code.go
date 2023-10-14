package bob

import (
	"context"
	"errors"

	"go.uber.org/multierr"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) aStudentInClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.StudentInCurrentClass) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("there are no student in current class")
	}
	stepState.CurrentStudentID = stepState.StudentInCurrentClass[0]
	token, err := s.generateExchangeToken(stepState.CurrentStudentID, entities.UserGroupStudent)
	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) createAClassWithSchoolIdIsAndExpiredAt(ctx context.Context, expiredAt string) (context.Context, error) {
	schoolID := constants.ManabieSchool
	stepState := StepStateFromContext(ctx)
	_, err1 := s.aSignedInWithSchool(ctx, "teacher", schoolID)

	ctx, err2 := s.aCreateClassRequest(ctx)

	ctx, err3 := s.aValidNameInCreateClassRequest(ctx)
	ctx, err4 := s.thisSchoolHasConfigIsIsIs(ctx, "plan_id", "School", "plan_expired_at", expiredAt, "plan_duration", 0)

	err := multierr.Combine(err1, err2, err3, err4)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.Request.(*pb.CreateClassRequest).SchoolId = int32(schoolID)

	s.userCreateAClass(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}

	ctx, err1 = s.bobMustCreateClassFromCreateClassRequest(ctx)
	ctx, err2 = s.classMustHasIs(ctx, "plan_id", "School")
	ctx, err3 = s.classMustHasIs(ctx, "plan_duration", "0")
	ctx, err4 = s.classMustHasIs(ctx, "plan_expired_at", expiredAt)
	ctx, err5 := s.classMustHaveMemberIsAndIsOwnerAndStatus(ctx, 1, "USER_GROUP_TEACHER", "true", "CLASS_MEMBER_STATUS_ACTIVE")
	err = multierr.Combine(err1, err2, err3, err4, err5)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aUpdateClassCodeRequestWithClassId(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	classID := int32(0)
	if arg1 == "valid" {
		classID = stepState.CurrentClassID
	}
	stepState.Request = &pb.UpdateClassCodeRequest{
		ClassId: classID,
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userUpdatesAClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).UpdateClassCode(contextWithToken(s, ctx), stepState.Request.(*pb.UpdateClassCodeRequest))

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustUpdateClassCode(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var classCode string
	err := s.DB.QueryRow(ctx, "SELECT class_code FROM classes WHERE class_id = $1", stepState.CurrentClassID).Scan(&classCode)

	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if classCode == stepState.CurrentClassCode {
		return StepStateToContext(ctx, stepState), errors.New("bob does not update class code")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aSignedInWithSchoolName(ctx context.Context, role, schoolName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	schoolID := s.getSchoolIDByName(ctx, schoolName)
	stepState.CurrentSchoolID = int32(schoolID)
	return s.aSignedInWithSchool(ctx, role, schoolID)
}
func (s *suite) ASignedInWithSchoolName(ctx context.Context, role, schoolName string) (context.Context, error) {
	return s.aSignedInWithSchoolName(ctx, role, schoolName)
}
