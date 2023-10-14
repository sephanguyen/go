package bob

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) joinClassWithSchoolIdIs(ctx context.Context, number int, role string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	currentUserID := stepState.CurrentUserID
	for number > 0 {
		ctx, err1 := s.aSignedInWithSchool(ctx, role, schoolID)
		ctx, err2 := s.aJoinClassRequest(ctx)
		ctx, err3 := s.aClassCodeInJoinClassRequest(ctx, "valid")
		ctx, err4 := s.userJoinAClass(ctx)
		err := multierr.Combine(err1, err2, err3, err4)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		number--
		if number == 0 {
			break
		}
	}
	if role == "student" {
		stepState.CurrentStudentID = stepState.CurrentUserID
	}
	stepState.CurrentUserID = currentUserID
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aRemoveMemberRequestWithClassInSchoolIdIs(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	classID := int32(0)
	if schoolID > 0 {
		if int32(schoolID) == stepState.CurrentSchoolID {
			classID = stepState.CurrentClassID
		} else {
			err := s.DB.QueryRow(ctx, "SELECT class_id FROM classes WHERE school_id = $1", schoolID).Scan(&classID)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow: %w", err)
			}
		}
	}

	stepState.Request = &pb.RemoveMemberRequest{
		ClassId: classID,
		UserIds: []string{},
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userRemoveMemberFromClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.createClassUpsertedSubscribe(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createClassUpsertedSubscribe: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).RemoveMember(contextWithToken(s, ctx), stepState.Request.(*pb.RemoveMemberRequest))

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) UserRemoveMemberFromClass(ctx context.Context) (context.Context, error) {
	return s.userRemoveMemberFromClass(ctx)
}
func (s *suite) aInRemoveMemberRequest(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if arg1 == "invalid userId" {
		return StepStateToContext(ctx, stepState), nil
	}
	if arg1 == "current teacherId" {
		stepState.Request.(*pb.RemoveMemberRequest).UserIds = append(stepState.Request.(*pb.RemoveMemberRequest).UserIds, stepState.CurrentTeacherID)
		return StepStateToContext(ctx, stepState), nil
	}

	var userID string
	var err error
	if arg1 == "valid userId" {
		if stepState.CurrentStudentID != "" {
			userID = stepState.CurrentStudentID
		} else {
			err = s.DB.QueryRow(ctx, "SELECT user_id FROM class_members WHERE class_id = $1 AND user_group = 'USER_GROUP_STUDENT'", stepState.CurrentClassID).Scan(&userID)
		}
	}
	if arg1 == "valid teacherId" {
		err = s.DB.QueryRow(ctx, "SELECT user_id FROM class_members WHERE class_id = $1 AND user_id != $2 AND user_group = 'USER_GROUP_TEACHER'", stepState.CurrentClassID, stepState.CurrentTeacherID).Scan(&userID)
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if userID != "" {
		userIDs := stepState.Request.(*pb.RemoveMemberRequest).UserIds
		userIDs = append(userIDs, userID)
		stepState.Request.(*pb.RemoveMemberRequest).UserIds = userIDs
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AInRemoveMemberRequest(ctx context.Context, arg1 string) (context.Context, error) {
	return s.aInRemoveMemberRequest(ctx, arg1)
}
func (s *suite) aValidTokenOfCurrentTeacher(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	stepState.AuthToken, err = s.generateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) AValidTokenOfCurrentTeacher(ctx context.Context) (context.Context, error) {
	return s.aValidTokenOfCurrentTeacher(ctx)
}
func (s *suite) bobMustStoreActivityLogsClassMember(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, _ := jwt.ParseString(stepState.AuthToken)
	userID := token.Subject()
	var classID int32
	var memberIDs []string

	count := 0
	status := entities_bob.ClassMemberStatusActive
	actionType := entities_bob.LogActionAddClassMember

	if arg1 == "remove" {
		req := stepState.Request.(*pb.RemoveMemberRequest)
		classID = req.ClassId
		memberIDs = req.UserIds
		status = entities_bob.ClassMemberStatusInactive
		actionType = entities_bob.LogActionRemoveClassMember
	}
	if arg1 == "add" {
		req := stepState.Request.(*pb.AddClassMemberRequest)
		classID = req.ClassId
		memberIDs = req.TeacherIds
	}

	pgUserID := database.Text(userID)
	pgmemberIDs := database.TextArray(memberIDs)
	pgStatus := database.Text(status)
	pgClassID := pgtype.Int4{Int: classID, Status: 2}
	pgActionType := database.Text(actionType)

	sql := "SELECT COUNT(*) FROM class_members WHERE class_id = $1 AND user_id = ANY($2) AND status = $3"
	_ = s.DB.QueryRow(ctx, sql, &pgClassID, &pgmemberIDs, &pgStatus).Scan(&count)
	if count != len(memberIDs) {
		return StepStateToContext(ctx, stepState), errors.Errorf("class members is not exist in class, expected %d got %d", len(memberIDs), count)
	}

	sql = "SELECT count(*) FROM activity_logs WHERE user_id = $1 AND payload->>'class_id'=$2 AND payload->>'class_member_ids'=$3 AND action_type=$4"
	_ = s.DB.QueryRow(ctx, sql, &pgUserID, &pgClassID, &pgmemberIDs, &pgActionType).Scan(&count)

	if count == 0 {
		return StepStateToContext(ctx, stepState), errors.New("actionLog does not match")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) BobMustStoreActivityLogsClassMember(ctx context.Context, arg1 string) (context.Context, error) {
	return s.bobMustStoreActivityLogsClassMember(ctx, arg1)
}
func (s *suite) aRemoveMemberRequestWithClassInSchoolName(ctx context.Context, schoolName string) (context.Context, error) {
	schoolID := s.getSchoolIDByName(ctx, schoolName)
	return s.aRemoveMemberRequestWithClassInSchoolIdIs(ctx, schoolID)
}
func (s *suite) ARemoveMemberRequestWithClassInSchoolName(ctx context.Context, schoolName string) (context.Context, error) {
	return s.aRemoveMemberRequestWithClassInSchoolName(ctx, schoolName)
}
func (s *suite) joinClassWithSchoolName(ctx context.Context, number int, role, schoolName string) (context.Context, error) {
	schoolID := s.getSchoolIDByName(ctx, schoolName)
	return s.joinClassWithSchoolIdIs(ctx, number, role, schoolID)
}
func (s *suite) JoinClassWithSchoolName(ctx context.Context, number int, role, schoolName string) (context.Context, error) {
	return s.joinClassWithSchoolName(ctx, number, role, schoolName)
}
