package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type LiveLessonConversationRepo struct{}

func (l *LiveLessonConversationRepo) GetConversationByLessonIDAndConvType(ctx context.Context, db database.QueryExecer, lessonID, convType string) (domain.LiveLessonConversation, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveLessonConversationRepo.GetConversationByLessonIDAndConvType")
	defer span.End()

	conDTO := &LiveLessonConversation{}
	fields, values := conDTO.FieldMap()
	query := fmt.Sprintf(`SELECT %s FROM %s
				WHERE lesson_id = $1 
				AND conversation_type = $2
				AND deleted_at IS NULL
				LIMIT 1 `,
		strings.Join(fields, ","),
		conDTO.TableName(),
	)

	err := db.QueryRow(ctx, query, database.Text(lessonID), database.Text(convType)).Scan(values...)
	if err == pgx.ErrNoRows {
		return domain.LiveLessonConversation{}, domain.ErrNoConversationFound
	} else if err != nil {
		return domain.LiveLessonConversation{}, fmt.Errorf("db.QueryRow: %w", err)
	}

	return conDTO.ToLiveLessonConversationDomain(), nil
}

func (l *LiveLessonConversationRepo) GetConversationIDByExactInfo(ctx context.Context, db database.QueryExecer, lessonID string, participants []string, convType string) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveLessonConversationRepo.GetConversationIDByExactInfo")
	defer span.End()

	query := `SELECT conversation_id 
				FROM live_lesson_conversation
				WHERE lesson_id = $1 
				AND participant_list <@ $2 AND participant_list @> $2
				AND conversation_type = $3
				AND deleted_at IS NULL
			  LIMIT 1 `

	var conversationID pgtype.Text
	err := db.QueryRow(ctx, query, database.Text(lessonID), database.TextArray(participants), database.Text(convType)).Scan(&conversationID)
	if err == pgx.ErrNoRows {
		return "", domain.ErrNoConversationFound
	} else if err != nil {
		return "", fmt.Errorf("db.QueryRow: %w", err)
	}

	return conversationID.String, nil
}

func (l *LiveLessonConversationRepo) UpsertConversation(ctx context.Context, db database.QueryExecer, conversation domain.LiveLessonConversation) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveLessonConversationRepo.UpsertConversation")
	defer span.End()

	conDTO, err := NewLiveLessonConversationDTO(conversation)
	if err != nil {
		return err
	}

	if err := conDTO.PreInsert(); err != nil {
		return err
	}

	fields, _ := conDTO.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))
	args := database.GetScanFields(conDTO, fields)

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT live_lesson_conversation_conversation_id_key 
		DO UPDATE SET participant_list = $4, updated_at = now() `,
		conDTO.TableName(),
		strings.Join(fields, ","),
		placeHolders,
	)
	commandTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return domain.ErrNoConversationCreated
	}

	return nil
}
