package managing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/yasuo"
	"github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	tomlesson "github.com/manabie-com/backend/internal/tom/domain/lesson"
)

func (s *suite) tomMustRecordNewConversationAndRecordNewConversation_lesson(ctx context.Context) (context.Context, error) {
	yasuoState := yasuo.StepStateFromContext(ctx)

	process := func() error {
		var listConversation []entities.Conversation
		for _, lessonName := range yasuoState.CurrentLessonNames {
			var conversation entities.Conversation
			fields := database.GetFieldNames(&conversation)
			querystm := fmt.Sprintf("SELECT %s FROM conversations WHERE name=$1", strings.Join(fields, ","))
			row := tomDB.QueryRow(ctx, querystm, lessonName)
			err := row.Scan(database.GetScanFields(&conversation, fields)...)
			if err != nil {
				return fmt.Errorf("error in tom subscribe lesson_event: %s", err.Error())
			}
			listConversation = append(listConversation, conversation)
		}
		for _, lessonId := range yasuoState.CurrentLessonIDs {
			var conversationLesson tomlesson.ConversationLesson
			fields := database.GetFieldNames(&conversationLesson)
			querystm := fmt.Sprintf("SELECT %s FROM conversation_lesson WHERE lesson_id=$1", strings.Join(fields, ","))
			row := tomDB.QueryRow(ctx, querystm, lessonId)
			err := row.Scan(database.GetScanFields(&conversationLesson, fields)...)
			if err != nil {
				return fmt.Errorf("error in tom subscribe lesson_event: %s", err.Error())
			}
		}
		return nil
	}
	return ctx, s.ExecuteWithRetry(process, 2*time.Second, 10)
}

// func (s *suite) aGenerateSchool(ctx context.Context) (context.Context, error) {
// 	err := s.yasuoSuite.ARandomNumber()
// 	if err != nil {
// 		return ctx, err

// 	}
// 	return s.yasuoSuite.AGenerateSchool(ctx)
// }
