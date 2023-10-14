package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgx/v4"
)

type ConversationMemberRepo struct{}

func (repo *ConversationMemberRepo) queueUpsert(b *pgx.Batch, item *dto.ConversationMemberPgDTO) {
	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as convo_mem (%s)
			VALUES (%s)
		ON CONFLICT ON CONSTRAINT conversation_member_conversation_id_user_id_un 
		DO UPDATE SET
			status = EXCLUDED.status,
			seen_at = EXCLUDED.seen_at,
			updated_at = EXCLUDED.updated_at
		WHERE convo_mem.deleted_at IS NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)
}

func (repo *ConversationMemberRepo) BulkUpsert(ctx context.Context, db database.Ext, items []domain.ConversationMember) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.BulkUpsert")
	defer span.End()
	b := &pgx.Batch{}
	for i := 0; i < len(items); i++ {
		dto, err := dto.NewConversationMemberDTOFromDomain(&items[i])
		if err != nil {
			return fmt.Errorf("repo.NewConversationMemberPgDTOFromEntity: %w", err)
		}
		repo.queueUpsert(b, dto)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (repo *ConversationMemberRepo) GetConversationMembersByUserID(ctx context.Context, db database.Ext, userID string, conversationIDs []string) ([]*domain.ConversationMember, error) {
	e := &dto.ConversationMemberPgDTO{}
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf(`
		SELECT cm.%s FROM conversation_member cm 
		WHERE cm.user_id = $1 AND cm.conversation_id = ANY($2)
		AND cm.status = 'CONVERSATION_MEMBER_STATUS_ACTIVE' AND cm.deleted_at IS NULL
	`, strings.Join(fields, ",cm."))

	rows, err := db.Query(ctx, query, database.Text(userID), database.TextArray(conversationIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversationMembers := dto.ConversationMemberPgDTOs{}
	for rows.Next() {
		e := &dto.ConversationMemberPgDTO{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		conversationMembers = append(conversationMembers, e)
	}

	return conversationMembers.ToConversationMemberDomain(), nil
}

func (repo *ConversationMemberRepo) CheckMembersExistInConversation(ctx context.Context, db database.Ext, conversationID string, conversationMemberIDs []string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.CheckMembersExistInConversation")
	defer span.End()

	e := &dto.ConversationMemberPgDTO{}
	query := fmt.Sprintf(`
		SELECT cm.user_id
		FROM %s cm
		WHERE cm.conversation_id = $1
		AND cm.deleted_at IS NULL
		AND cm.status = 'CONVERSATION_MEMBER_STATUS_ACTIVE'
		AND cm.user_id = ANY($2::TEXT[])
	`, e.TableName())

	rows, err := db.Query(ctx, query, database.Text(conversationID), database.TextArray(conversationMemberIDs))
	if err != nil {
		return nil, fmt.Errorf("failed Query: %+v", err)
	}

	existedMemberIDs := []string{}
	defer rows.Close()
	for rows.Next() {
		var memID string
		if err = rows.Scan(&memID); err != nil {
			return nil, fmt.Errorf("failed Scan: %+v", err)
		}
		existedMemberIDs = append(existedMemberIDs, memID)
	}

	return existedMemberIDs, nil
}
