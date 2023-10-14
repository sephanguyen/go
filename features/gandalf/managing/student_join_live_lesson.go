package managing

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/internal/golibs/constants"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	"github.com/pkg/errors"
)

func (s *suite) studentJoinLesson(ctx context.Context) (context.Context, error) {
	ctx, err := s.bobSuite.StudentJoinLesson(ctx)
	if err != nil {
		return ctx, err

	}

	stepState := GandalfStepStateFromContext(ctx)
	stepState.YasuoStepState.CurrentUserGroup = "USER_GROUP_STUDENT"
	return ctx, nil
}

func (s *suite) studentJoinLessonV1(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ctx, err := s.bobSuite.StudentJoinLessonV1(ctx)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err

	}
	stepState.YasuoStepState.CurrentUserGroup = "USER_GROUP_STUDENT"
	return GandalfStepStateToContext(ctx, stepState), nil
}
func (s *suite) aStudentWithValidLesson(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	ctx, err := s.bobSuite.AStudentWithValidLesson(ctx)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err

	}

	pbLiveLessons := make([]*pb_bob.EvtLesson_Lesson, 0, 1)
	pbLiveLessons = append(pbLiveLessons, &pb_bob.EvtLesson_Lesson{
		LessonId: bob.StepStateFromContext(ctx).CurrentLessonID,
	})

	msg := &pb_bob.EvtLesson{
		Message: &pb_bob.EvtLesson_CreateLessons_{
			CreateLessons: &pb_bob.EvtLesson_CreateLessons{
				Lessons: pbLiveLessons,
			},
		},
	}

	data, err := msg.Marshal()
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err
	}

	_, err = s.jsm.PublishContext(ctx, constants.SubjectLessonCreated, data)
	if err != nil {
		return GandalfStepStateToContext(ctx, stepState), err
	}

	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) tomMustRecordNewLessonConversation(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := `SELECT conversation_id
				FROM conversation_lesson
				WHERE lesson_id = $1`

		rows, err := s.tomDB.Query(ctx, query, bob.StepStateFromContext(ctx).CurrentLessonID)
		if err != nil {
			return err

		}
		defer rows.Close()

		var conversationID string
		for rows.Next() {
			err = rows.Scan(&conversationID)
			if err != nil {
				return err

			}
		}

		if conversationID == "" {
			return errors.New(fmt.Sprintf("not found any conversation_lesson where lesson_id = %s",
				bob.StepStateFromContext(ctx).CurrentLessonID))
		}

		query = `SELECT count(conversation_id)
				FROM conversation_members
				WHERE conversation_id = $1
				AND user_id = $2
				AND role = $3
				AND status = 'CONVERSATION_STATUS_ACTIVE'`

		rows, err = s.tomDB.Query(ctx, query, conversationID, bob.StepStateFromContext(ctx).CurrentUserID, stepState.YasuoStepState.CurrentUserGroup)
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

		if count != 1 {
			return errors.New(fmt.Sprintf("number of conversation_member where conversation_id = %s, user_id = %s, role =%s, status = %s must be 1",
				conversationID,
				bob.StepStateFromContext(ctx).CurrentUserID,
				stepState.YasuoStepState.CurrentUserGroup,
				"CONVERSATION_STATUS_ACTIVE"))
		}

		stepState.GandalfStateConversationID = conversationID
		return nil
	}
	return GandalfStepStateToContext(ctx, stepState), s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}

func (s *suite) tomMustStoreMessageJoinLesson(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := `SELECT count(message_id)
				FROM messages
				WHERE conversation_id = $1
				AND target_user = $2
				AND type = $3
				AND message = $4
				AND deleted_at IS NULL`
		rows, err := s.tomDB.Query(ctx, query, stepState.GandalfStateConversationID,
			bob.StepStateFromContext(ctx).CurrentUserID,
			"MESSAGE_TYPE_SYSTEM",
			"CODES_MESSAGE_TYPE_JOINED_LESSON")
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

		if count != 1 {
			errorMsg := fmt.Sprintf("number of message where conversation_id = %s, user_id = %s, type = %s, message = %s, deleted_at is null must be 1, the fact is %d",
				stepState.GandalfStateConversationID,
				bob.StepStateFromContext(ctx).CurrentUserID,
				"MESSAGE_TYPE_SYSTEM",
				"CODES_MESSAGE_TYPE_JOINED_LESSON",
				count)
			return errors.New(errorMsg)
		}

		return nil

	}

	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}
