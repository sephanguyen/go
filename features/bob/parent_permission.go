package bob

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"
	"github.com/lestrrat-go/jwx/jwt"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) aUserSignedInAsAParent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, "parent")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aUserSignedInAsAParentWithSchoolID(ctx context.Context, _ int) (context.Context, error) {
	return s.aUserSignedInAsAParent(ctx)
}

func (s *suite) userCallsAPI(ctx context.Context, apiName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	from := &types.Timestamp{Seconds: timeutil.StartWeekIn(pb.COUNTRY_VN).Unix()}
	to := &types.Timestamp{Seconds: timeutil.EndWeekIn(pb.COUNTRY_VN).Unix()}

	t, _ := jwt.ParseString(stepState.AuthToken)
	studentID := t.Subject()

	switch apiName {
	case "RetrieveLearningProgress":
		if ctx, err := s.callRetrieveLearningProgress(ctx, studentID, from, to); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "RetrieveStat":
		if ctx, err := s.callRetrieveStat(ctx, studentID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "CountTotalLOsFinished":
		if ctx, err := s.callCountTotalLOsFinished(ctx, studentID, from, to); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`invalid API type: "%s"`, apiName)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) callRetrieveLearningProgress(ctx context.Context, studentID string, from *types.Timestamp, to *types.Timestamp) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).RetrieveLearningProgress(s.signedCtx(ctx), &pb.RetrieveLearningProgressRequest{
		StudentId: studentID,
		From:      from,
		To:        to,
	}); stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) callRetrieveStat(ctx context.Context, studentID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.Response, stepState.ResponseErr = epb.NewStudyPlanReaderServiceClient(s.Conn).RetrieveStat(s.signedCtx(ctx), &epb.RetrieveStatRequest{
		StudentId: studentID,
	}); stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) callCountTotalLOsFinished(ctx context.Context, studentID string, from *types.Timestamp, to *types.Timestamp) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.CountTotalLOsFinishedRequest{
		StudentId: studentID,
		From:      from,
		To:        to,
	}
	if stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).CountTotalLOsFinished(s.signedCtx(ctx), req); stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedIn(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if arg1 == "unauthenticated" {
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	}

	if arg1 == "admin" {
		return s.aSignedInAdmin(ctx)
	}

	if arg1 == "student" {
		return s.aSignedInStudent(ctx)
	}

	id := s.newID()
	var (
		userGroup string
		err       error
	)

	if arg1 == "teacher" {
		userGroup = entities_bob.UserGroupTeacher
	}
	if arg1 == "school admin" {
		userGroup = entities_bob.UserGroupSchoolAdmin
	}
	if arg1 == "parent" {
		userGroup = entities_bob.UserGroupParent
	}

	ctx, err = s.aValidUser(ctx, withID(id), withRole(userGroup))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var newGroup string
	switch userGroup {
	case entities_bob.UserGroupTeacher:
		newGroup = consta.RoleTeacher
	case entities_bob.UserGroupStudent:
		newGroup = consta.RoleStudent
	case entities_bob.UserGroupSchoolAdmin:
		newGroup = consta.RoleSchoolAdmin
	default:
		newGroup = consta.RoleStudent
	}
	ctx, err = s.aValidUserInEureka(ctx, id, newGroup, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create user in eureka: %w", err)
	}
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = userGroup
	stepState.AuthToken, err = s.generateExchangeToken(id, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	return StepStateToContext(ctx, stepState), nil
}
