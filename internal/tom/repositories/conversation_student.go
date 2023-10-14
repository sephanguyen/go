package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	sentities "github.com/manabie-com/backend/internal/tom/domain/support"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type ConversationStudentRepo struct {
}

func (r *ConversationStudentRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []*sentities.ConversationStudent) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.BulkUpsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *sentities.ConversationStudent) {
		fieldNames := []string{"id", "conversation_id", "student_id", "conversation_type", "created_at", "updated_at", "deleted_at"}
		placeHolders := "$1, $2, $3, $4, $5, $6, $7"

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT conversation_students_pk 
		DO UPDATE SET updated_at = $6`,
			e.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, c := range conversations {
		queueFn(b, c)
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

func (r *ConversationStudentRepo) UpdateSearchIndexTime(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, insertTime pgtype.Timestamptz) error {
	query := `
UPDATE conversation_students SET search_index_time = $1 WHERE conversation_id = ANY($2) AND deleted_at IS NULL`
	ctx, span := interceptors.StartSpan(ctx, "ConversationStudentRepo.UpdateSearchIndexTime")
	defer span.End()

	_, err := db.Exec(ctx, query, insertTime, conversationIDs)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}
func (r *ConversationStudentRepo) FindSearchIndexTime(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (map[string]pgtype.Timestamptz, error) {
	query := `select conversation_id,search_index_time from conversation_students where conversation_id = ANY($1) and deleted_at IS NULL`
	ctx, span := interceptors.StartSpan(ctx, "ConversationStudentRepo.FindSearchIndexTime")
	defer span.End()

	rows, err := db.Query(ctx, query, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	idTimeMap := make(map[string]pgtype.Timestamptz)

	for rows.Next() {
		var (
			convID          pgtype.Text
			searchIndexTime pgtype.Timestamptz
		)
		if err := rows.Scan(&convID, &searchIndexTime); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		idTimeMap[convID.String] = searchIndexTime
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}
	if len(idTimeMap) != len(database.FromTextArray(conversationIDs)) {
		return nil, fmt.Errorf("want %d items after searching for search index time, has %d", len(database.FromTextArray(conversationIDs)), len(idTimeMap))
	}
	return idTimeMap, nil
}

func (r *ConversationStudentRepo) FindByConversationIDs(
	ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (conversationStudentMap map[pgtype.Text]*sentities.ConversationStudent, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationStudentRepo.FindByConversationIDs")
	defer span.End()

	c := &sentities.ConversationStudent{}
	fields := database.GetFieldNames(c)

	query := fmt.Sprintf(`SELECT %s 
		FROM conversation_students 
		WHERE conversation_id = ANY($1) AND deleted_at IS NULL`, strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	conversations := make(map[pgtype.Text]*sentities.ConversationStudent)

	for rows.Next() {
		c := &sentities.ConversationStudent{}
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		conversations[c.ConversationID] = c
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}

	return conversations, nil
}

func (r *ConversationStudentRepo) FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, conversationType pgtype.Text) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationStudentRepo.FindByStudentID")
	defer span.End()

	query := `SELECT conversation_id 
		FROM conversation_students 
		WHERE student_id = ANY($1) 
		AND ($2::text IS NULL OR conversation_type = $2)`
	rows, err := db.Query(ctx, query, studentIDs, conversationType)
	if err != nil {
		return nil, fmt.Errorf("db.QueryEx: %w", err)
	}
	defer rows.Close()

	var conversationIDs []string

	for rows.Next() {
		var conversationID string
		if err := rows.Scan(&conversationID); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		conversationIDs = append(conversationIDs, conversationID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}

	return conversationIDs, nil
}

func (r *ConversationStudentRepo) FindByStaffIDs(ctx context.Context, db database.QueryExecer, staffIDs pgtype.TextArray) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationStudentRepo.FindByStudentID")
	defer span.End()

	query := `SELECT cs.conversation_id 
		FROM conversation_students cs
		INNER JOIN conversation_members cm
		ON cs.conversation_id = cm.conversation_id
		WHERE cm.user_id = ANY($1)
		AND cs.deleted_at IS NULL
		AND cm.status = 'CONVERSATION_STATUS_ACTIVE'`
	rows, err := db.Query(ctx, query, staffIDs)
	if err != nil {
		return nil, fmt.Errorf("db.QueryEx: %w", err)
	}
	defer rows.Close()

	var conversationIDs []string

	for rows.Next() {
		var conversationID string
		if err := rows.Scan(&conversationID); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		conversationIDs = append(conversationIDs, conversationID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}

	return conversationIDs, nil
}
