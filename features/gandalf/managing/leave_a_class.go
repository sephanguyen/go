package managing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) currentStudentJoinToClassByUsingThisClassCode(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.userJoinAClass(ctx)
	ctx, err2 := s.bobSuite.ReturnsStatusCode(ctx, "OK")
	ctx, err3 := s.bobSuite.ClassMustHaveMemberIsAndIsOwnerAndStatus(ctx, 1, "USER_GROUP_TEACHER", "true", "CLASS_MEMBER_STATUS_ACTIVE")
	ctx, err4 := s.bobSuite.ClassMustHaveMemberIsAndIsOwnerAndStatus(ctx, 1, "USER_GROUP_STUDENT", "false", "CLASS_MEMBER_STATUS_ACTIVE")
	ctx, err5 := s.bobSuite.BobMustPushMsgSubjectToNats(ctx, "JoinClass", constants.SubjectClassUpserted)
	return ctx, multierr.Combine(err1, err2, err3, err4, err5)
}

func (s *suite) tomMustRecordMessageLeaveClassOnThisClassConversation(ctx context.Context, kicked string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := `SELECT user_id
				FROM conversation_members
				WHERE conversation_id = $1 and status = 'CONVERSATION_STATUS_INACTIVE'`
		rows, err := s.tomDB.Query(ctx, query, stepState.GandalfStateConversationID)
		defer rows.Close()
		if err != nil {
			return err

		}
		var listUserID = make([]string, 0)
		for rows.Next() {
			var userID string
			err = rows.Scan(&userID)
			if err != nil {
				return err

			}
			if userID != "" {
				listUserID = append(listUserID, userID)
			}
		}
		if len(listUserID) == 0 {
			return errors.New(fmt.Sprintf("not found any conversation_members where conversation_id = %s and status = 'CONVERSATION_STATUS_INACTIVE'", stepState.GandalfStateConversationID))
		}
		var message = "CODES_MESSAGE_LEFT_CLASS"
		if kicked == "true" {
			message = "CODES_MESSAGE_REMOVED_FROM_CLASS"
		}
		query = `SELECT count(conversation_id)
				FROM messages
				WHERE user_id = ANY ($1)
				AND conversation_id = $2
				AND type = $3
				AND message = $4
				AND deleted_at IS NULL`
		rows, err = s.tomDB.Query(ctx, query, listUserID, stepState.GandalfStateConversationID, "MESSAGE_TYPE_SYSTEM", message)
		if err != nil {
			return err
		}
		defer rows.Close()
		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err

			}
		}
		if count != len(listUserID) {
			return errors.New("Tom does not store message leave class")
		}
		return nil
	}
	return GandalfStepStateToContext(ctx, stepState), s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}

func (s *suite) currentTeacherJoinToClassByUsingThisClassCode(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.userJoinAClass(ctx)
	ctx, err2 := s.bobSuite.ReturnsStatusCode(ctx, "OK")
	ctx, err3 := s.bobSuite.ClassMustHaveMemberIsAndIsOwnerAndStatus(ctx, 2, "USER_GROUP_TEACHER", "true", "CLASS_MEMBER_STATUS_ACTIVE")
	ctx, err4 := s.bobSuite.BobMustPushMsgSubjectToNats(ctx, "JoinClass", constants.SubjectClassUpserted)

	return ctx, multierr.Combine(err1, err2, err3, err4)
}

func (s *suite) eurekaMustRemoveClassMember(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		var err error
		for _, studentID := range stepState.BobStepState.ClassStudentsID {
			if golibs.InArrayString(studentID, stepState.BobStepState.UserIDsLeaveClass) {
				ctx, err = s.checkStudentIsRemovedInEurekaClass(ctx, studentID, bob.StepStateFromContext(ctx).CurrentClassID)
				if err != nil {
					return err

				}
			}
		}
		return nil
	}
	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}

func (s *suite) checkStudentIsRemovedInEurekaClass(ctx context.Context, studentID string, classID int32) (context.Context, error) {
	query := `SELECT count(student_id)
					FROM class_students
					WHERE class_id = $1 
					AND student_id = $2
					AND deleted_at IS NOT NULL`

	rows, err := s.eurekaDB.Query(ctx, query, strconv.FormatInt(int64(classID), 10), studentID)
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

	if count != 1 {
		return ctx, errors.New(fmt.Sprintf("Eureka does not update class_students where class_id = %d, student_id = %s, deleted_at is not null", classID, studentID))
	}

	return ctx, nil
}

func (s *suite) getUserIDsLeaveClass(ctx context.Context) []string {
	switch v := bob.StepStateFromContext(ctx).Request.(type) {
	case *pb.LeaveClassRequest:
		return []string{bob.StepStateFromContext(ctx).CurrentUserID}
	case *pb.RemoveMemberRequest:
		return v.UserIds
	}
	return nil
}
