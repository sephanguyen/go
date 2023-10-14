package managing

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/yasuo"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) aClassWithSchoolNameAndExpiredAt(ctx context.Context, schoolName, schoolExpiredDate string) (context.Context, error) {
	ctx, err := s.bobSuite.CreateAClassWithSchoolNameAndExpiredAt(ctx, schoolName, schoolExpiredDate)
	if err != nil {
		return ctx, err

	}

	stepState := GandalfStepStateFromContext(ctx)
	stepState.BobStepState.ClassOwnersID = stepState.BobStepState.ClassOwnersID[:0]
	stepState.BobStepState.ClassOwnersID = append(stepState.BobStepState.ClassOwnersID, bob.StepStateFromContext(ctx).CurrentTeacherID)
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) aClassWithThisConfigCreatedByCurrentTeacher(ctx context.Context, planIDKey, planIDValue, planExpiredAtKey, planExpiredAtValue, planDurationKey, planDurationValue string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	stepState.BobStepState.ClassOwnersID = stepState.BobStepState.ClassOwnersID[:0]
	stepState.BobStepState.ClassOwnersID = append(stepState.BobStepState.ClassOwnersID, bob.StepStateFromContext(ctx).CurrentTeacherID)
	ctx, err1 := s.bobSuite.UserCreateAClass(ctx)
	ctx, err2 := s.bobSuite.ReturnsStatusCode(ctx, "OK")
	ctx, err3 := s.bobSuite.BobMustCreateClassFromCreateClassRequest(ctx)
	ctx, err4 := s.bobSuite.ClassMustHasIs(ctx, planIDKey, planIDValue)
	ctx, err5 := s.bobSuite.ClassMustHasIs(ctx, planExpiredAtKey, planExpiredAtValue)
	ctx, err6 := s.bobSuite.ClassMustHasIs(ctx, planDurationKey, planDurationValue)
	ctx, err7 := s.bobSuite.ClassMustHaveMemberIsAndIsOwnerAndStatus(ctx, 1, "USER_GROUP_TEACHER", "true", "CLASS_MEMBER_STATUS_ACTIVE")
	ctx, err8 := s.bobSuite.BobMustPushMsgSubjectToNats(ctx, "CreateClass", constants.SubjectClassUpserted)
	return ctx, multierr.Combine(err1, err2, err3, err4, err5, err6, err7, err8)
}

func (s *suite) aUserSignedInWithSchoolName(ctx context.Context, role string, schoolName string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ctx, err := s.bobSuite.ASignedInWithSchoolName(ctx, role, schoolName)
	if err != nil {
		return ctx, err

	}
	if role == "student" {
		stepState.BobStepState.CurrentStudentId = bob.StepStateFromContext(ctx).CurrentStudentID
	}
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) userJoinAClass(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ctx, err := s.bobSuite.UserJoinAClass(ctx)
	if err != nil {
		return ctx, err

	}
	if stepState.BobStepState.CurrentStudentId != "" {
		stepState.BobStepState.ClassStudentsID = append(stepState.BobStepState.ClassStudentsID,
			stepState.BobStepState.CurrentStudentId)
	} else {
		stepState.BobStepState.ClassOwnersID = append(stepState.BobStepState.ClassOwnersID,
			bob.StepStateFromContext(ctx).CurrentTeacherID)
	}

	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUserIsInConversation(ctx context.Context, userID, role, conversationID, status string) (context.Context, error) {
	query := `SELECT COUNT(conversation_id)
			FROM conversation_members
			WHERE conversation_id = $1
			AND status = $2
			AND role = $3
			AND user_id = $4`
	rows, err := s.tomDB.Query(ctx, query, conversationID, status, role, userID)
	if err != nil {
		return ctx, err
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return ctx, err
		}
	}
	if count == 0 {
		return ctx, errors.New(fmt.Sprintf(`not found any conversation_member where conversation_id = %s and status = %s and role = %s and user_id = %s`, conversationID, "CONVERSATION_STATUS_ACTIVE", role, userID))
	}
	return ctx, nil
}

func (s *suite) tomMustRecordMessageJoinClassOfCurrentUser(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := `SELECT count(message_id) 
				FROM messages 
				WHERE conversation_id = $1 
				AND message = $2 
				AND type = $3 
				AND user_id = $4
                AND deleted_at IS NULL`
		rows, err := s.tomDB.Query(ctx, query, stepState.GandalfStateConversationID,
			"CODES_MESSAGE_JOINED_CLASS",
			"MESSAGE_TYPE_SYSTEM", bob.StepStateFromContext(ctx).CurrentUserID)
		defer rows.Close()
		if err != nil {
			return err

		}
		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err

			}
		}
		if count == 0 {
			return errors.New(fmt.Sprintf("tom must create message with conversation_id = %v, message =%v, type = %v, user_id = %v",
				stepState.GandalfStateConversationID,
				"CODES_MESSAGE_JOINED_CLASS", "MESSAGE_TYPE_SYSTEM",
				bob.StepStateFromContext(ctx).CurrentUserID))
		}
		return nil
	}
	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}

func (s *suite) eurekaMustAddNewClassMember(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	if len(stepState.BobStepState.ClassStudentsID) == 0 {
		return ctx, nil
	}
	joinClassResp := bob.StepStateFromContext(ctx).Response.(*pb.JoinClassResponse)
	classID := joinClassResp.ClassId
	mainProcess := func() error {
		var err error
		for _, studentID := range stepState.BobStepState.ClassStudentsID {
			ctx, err = s.checkStudentIsInEurekaClass(ctx, strconv.FormatInt(int64(classID), 10), studentID)
			if err != nil {
				return err

			}
		}
		return nil
	}
	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}

func (s *suite) checkStudentIsInEurekaClass(ctx context.Context, classID, studentID string) (context.Context, error) {
	query := `SELECT COUNT(student_id)
			FROM class_students
			WHERE class_id = $1 
			AND student_id = $2
			AND deleted_at IS NULL`

	rows, err := s.eurekaDB.Query(ctx, query, classID, studentID)
	defer rows.Close()
	if err != nil {
		return ctx, err

	}

	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return ctx, err

		}
	}

	if count == 0 {
		return ctx, errors.New(fmt.Sprintf("not found any class_students where class_id = %s and student_id = %s", classID, studentID))
	}

	return ctx, nil
}

func (s *suite) tomMustRecordMessageWithType(ctx context.Context, message, typeMessage string) (context.Context, error) {
	yasuoState := yasuo.StepStateFromContext(ctx)
	stepState := GandalfStepStateFromContext(ctx)

	process := func() error {
		for _, conversationID := range stepState.YasuoStepState.CurrentConversationIDs {
			querystm := fmt.Sprintf("SELECT %s FROM messages WHERE message = $1 AND type = $2 AND conversation_id = $3 AND user_id = $4",
				strings.Join(database.GetFieldNames(&entities.Message{}), ","))
			row := s.tomDB.QueryRow(ctx, querystm, &message, &typeMessage, conversationID, yasuoState.CurrentTeacherID)
			var message entities.Message
			err := row.Scan(database.GetScanFields(&message, database.GetFieldNames(&message))...)
			if err != nil {
				return err

			}
		}
		return nil
	}
	return ctx, s.ExecuteWithRetry(process, 2*time.Second, 10)
}
