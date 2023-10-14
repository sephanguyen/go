package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	entities "github.com/manabie-com/backend/internal/tom/domain/lesson"

	"github.com/jackc/pgtype"
)

type PrivateConversationLessonRepo struct {
}

func (r *PrivateConversationLessonRepo) Create(ctx context.Context, db database.QueryExecer, privateConversation *entities.PrivateConversationLesson) error {
	ctx, span := interceptors.StartSpan(ctx, "PrivateConversationLesson.Create")
	defer span.End()
	fieldNames, values := privateConversation.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`,
		privateConversation.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)
	_, err := db.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *PrivateConversationLessonRepo) UpdateLatestStartTime(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, latestStartTime pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.FindAndUpdateLatestCallID")
	defer span.End()
	updateStmt := `
UPDATE private_conversation_lesson
SET 
    latest_start_time = $1 WHERE lesson_id = $2 AND deleted_at IS NULL`
	_, err := db.Exec(ctx, updateStmt, &latestStartTime, &lessonID)
	return err
}
