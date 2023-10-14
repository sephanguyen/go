package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type ConversationRepo struct{}

func (repo ConversationRepo) UpsertConversation(ctx context.Context, db database.Ext, conversation *domain.Conversation) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.UpsertConversation")
	defer span.End()
	conversationDTO, err := dto.NewConversationDTOFromDomain(conversation)
	if err != nil {
		return err
	}

	fields := database.GetFieldNames(conversationDTO)
	values := database.GetScanFields(conversationDTO, fields)
	pl := database.GeneratePlaceholders(len(fields))
	tableName := conversationDTO.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as convo (%s) 
			VALUES (%s) 
		ON CONFLICT ON CONSTRAINT conversation_pk 
		DO UPDATE SET 
			name = EXCLUDED.name,
			latest_message = EXCLUDED.latest_message, 
			latest_message_sent_time = EXCLUDED.latest_message_sent_time, 
			resource_path = EXCLUDED.resource_path,
			updated_at = EXCLUDED.updated_at
		WHERE convo.deleted_at IS NULL;
		`, tableName, strings.Join(fields, ","), pl)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not upsert conversation")
	}

	return nil
}

func (repo ConversationRepo) UpdateLatestMessage(ctx context.Context, db database.Ext, message *domain.Message) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.UpdateLatestMessage")
	defer span.End()

	msgDTO := dto.NewMessagePgDTOFromDomain(message)

	query := `
		UPDATE conversation
		SET latest_message = $2,
			latest_message_sent_time = $3,
			updated_at = now()
		WHERE conversation_id = $1 and deleted_at IS NULL;
	`

	cmd, err := db.Exec(ctx, query, message.ConversationID, msgDTO.ToJSONB(), database.Timestamptz(message.SentTime))
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not update conversation latest message")
	}

	return nil
}

func (repo *ConversationRepo) FindByIDsAndUserID(ctx context.Context, db database.Ext, userID string, conversationIDs []string) ([]*domain.Conversation, error) {
	e := &dto.ConversationPgDTO{}
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf(`
		SELECT c.%s FROM conversation c 
		JOIN conversation_member cm ON c.conversation_id = cm.conversation_id 
		WHERE cm.user_id = $1 AND c.conversation_id = ANY($2)
		AND c.deleted_at IS NULL AND cm.deleted_at IS null
		ORDER BY c.updated_at DESC
	`, strings.Join(fields, ",c."))

	rows, err := db.Query(ctx, query, database.Text(userID), database.TextArray(conversationIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := dto.ConversationPgDTOs{}
	for rows.Next() {
		e := &dto.ConversationPgDTO{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, e)
	}

	return conversations.ToConversationsDomain()
}

func (repo ConversationRepo) FindByIDs(ctx context.Context, db database.Ext, conversationIDs []string) ([]*domain.Conversation, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.FindById")
	defer span.End()
	e := &dto.ConversationPgDTO{}
	fieldNames := database.GetFieldNames(e)
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s c
		WHERE c.conversation_id = ANY($1::TEXT[])
		AND c.deleted_at IS NULL
	`, strings.Join(fieldNames, ","), e.TableName())

	rows, err := db.Query(ctx, query, database.TextArray(conversationIDs))
	if err != nil {
		return nil, fmt.Errorf("failed db.Query: %+v", err)
	}

	defer rows.Close()

	conversations := dto.ConversationPgDTOs{}
	for rows.Next() {
		c := &dto.ConversationPgDTO{}
		if err = rows.Scan(database.GetScanFields(c, fieldNames)...); err != nil {
			return nil, fmt.Errorf("failed rows.Scan: %+v", err)
		}

		conversations = append(conversations, c)
	}

	return conversations.ToConversationsDomain()
}

func (repo ConversationRepo) FindByID(ctx context.Context, db database.Ext, conversationID string) (*domain.Conversation, error) {
	conversations, err := repo.FindByIDs(ctx, db, []string{conversationID})
	if err != nil {
		return nil, err
	}

	if len(conversations) > 0 {
		return conversations[0], nil
	}

	return nil, fmt.Errorf("not found conversation")
}

func (repo *ConversationRepo) UpdateConversationInfo(ctx context.Context, db database.Ext, conversation *domain.Conversation) error {
	query := `
		UPDATE conversation
		SET "name" = $1, optional_config = $2, updated_at = now()
		WHERE conversation_id = $3 AND deleted_at IS NULL
	`

	cmd, err := db.Exec(ctx, query, database.Text(conversation.Name), database.JSONB(conversation.OptionalConfig), database.Text(conversation.ID))
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not update conversation info")
	}

	return nil
}
