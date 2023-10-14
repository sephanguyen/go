package managing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/bob"
)

func (s *suite) tomMustChangeConversationLessonStatusOfThisUserToInactive(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := `SELECT conversation_id
				FROM conversation_lesson
				WHERE lesson_id = $1`

		rows, err := s.tomDB.Query(ctx, query, bob.StepStateFromContext(ctx).CurrentLessonID)
		defer rows.Close()
		if err != nil {
			return err

		}

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
				AND status = 'CONVERSATION_STATUS_INACTIVE'`

		rows, err = s.tomDB.Query(ctx, query, conversationID,
			bob.StepStateFromContext(ctx).CurrentUserID,
			stepState.YasuoStepState.CurrentUserGroup)
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
				"CONVERSATION_STATUS_INACTIVE"))
		}

		stepState.GandalfStateConversationID = conversationID
		return nil
	}
	return GandalfStepStateToContext(ctx, stepState), s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}

func (s *suite) tomMustRecordMessageLeaveLesson(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := `SELECT count(message_id)
				FROM messages
				WHERE conversation_id = $1
				AND user_id = $2
				AND type = $3
				AND message = $4
				AND deleted_at IS NULL`
		rows, err := s.tomDB.Query(ctx, query,
			stepState.GandalfStateConversationID,
			bob.StepStateFromContext(ctx).CurrentUserID,
			"MESSAGE_TYPE_SYSTEM",
			"CODES_MESSAGE_TYPE_LEFT_LESSON")
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
				"CODES_MESSAGE_TYPE_LEFT_LESSON",
				count)
			return errors.New(errorMsg)
		}

		return nil

	}

	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}
