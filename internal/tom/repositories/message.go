package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type MessageRepo struct {
}

func (r *MessageRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, messages []*domain.Message) error {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.BulkUpsert")
	defer span.End()

	fieldNames := database.GetFieldNames(&domain.Message{})
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	queueFn := func(b *pgx.Batch, e *domain.Message) error {
		now := time.Now()
		err := multierr.Combine(
			e.UpdatedAt.Set(now),
			e.CreatedAt.Set(now),
		)
		if err != nil {
			return err
		}
		query := fmt.Sprintf(`INSERT INTO messages (%s) VALUES (%s)`,
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
		return nil
	}

	b := &pgx.Batch{}

	for _, m := range messages {
		err := queueFn(b, m)
		if err != nil {
			return fmt.Errorf("MessageRepo.BulkUpsert queueFn: %w", err)
		}
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

func (r *MessageRepo) Create(ctx context.Context, db database.QueryExecer, m *domain.Message) error {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.Create")
	defer span.End()

	now := time.Now()
	m.UpdatedAt.Set(now)
	m.CreatedAt.Set(now)

	cmdTag, err := Insert(ctx, m, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new conversation")
	}

	return nil
}

func (r *MessageRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.FindByID")
	defer span.End()

	m := &domain.Message{}
	fields := database.GetFieldNames(m)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE message_id = $1", strings.Join(fields, ","), m.TableName())

	row := db.QueryRow(ctx, selectStmt, &id)

	if err := row.Scan(database.GetScanFields(m, fields)...); err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}

	return m, nil
}

func (r *MessageRepo) GetLastMessageEachUserConversation(ctx context.Context, db database.QueryExecer, userID, status pgtype.Text, limit uint, endAt pgtype.Timestamptz, locationIDs pgtype.TextArray, enableChatThreadLaunching bool) ([]*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.GetLastMessageEachUserConversation")
	defer span.End()

	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return nil, err
	}
	args := []interface{}{&status, &userID, &endAt, &limit, resourcePath, locationIDs}

	m := &domain.Message{}
	fields := database.GetFieldNames(m)
	var selectStmt string
	if enableChatThreadLaunching {
		selectStmt = fmt.Sprintf(`
			SELECT %s FROM %s m 
			JOIN (
				SELECT DISTINCT last_message_id FROM conversations c 
				LEFT JOIN conversation_locations cl ON c.conversation_id  = cl.conversation_id 
				WHERE (c.conversation_type <> 'CONVERSATION_LESSON' OR c.conversation_type IS NULL) AND c.status = $1
				AND c.conversation_id IN (SELECT conversation_id FROM conversation_members cm WHERE cm.user_id = $2 AND cm.status = 'CONVERSATION_STATUS_ACTIVE')
				AND cl.location_id = ANY($6) AND cl.deleted_at IS NULL AND c.resource_path = $5 AND cl.resource_path = $5
			) AS latest
			ON m.message_id  = latest.last_message_id
			WHERE m.deleted_at IS NULL AND m.resource_path = $5 AND m.created_at < $3
			ORDER BY m.created_at DESC
			LIMIT $4
		`, strings.Join(fields, ", m."), m.TableName())
	} else {
		selectStmt = strings.Replace(`
			WITH root AS (
				SELECT l.location_id FROM locations l JOIN location_types lt  
				ON l.location_type = lt.location_type_id 
					WHERE lt."name" LIKE '%org%' AND lt.resource_path = $5 AND l.resource_path = $5
					AND l.deleted_at IS NULL AND lt.deleted_at IS NULL
			)
			SELECT m.:FIELD
			FROM messages AS m
			JOIN (
				SELECT last_message_id FROM conversations c 
				LEFT JOIN conversation_locations cl USING (conversation_id) 
				CROSS JOIN root
				WHERE (c.conversation_type <> 'CONVERSATION_LESSON' OR c.conversation_type IS NULL) AND c.status = $1
				AND c.conversation_id in (select conversation_id FROM conversation_members cm WHERE cm.user_id = $2 AND cm.status = 'CONVERSATION_STATUS_ACTIVE')
				AND (cl.location_id = ANY($6) OR root.location_id = ANY($6))
				ORDER BY last_message_id
			) AS latest
				ON m.message_id = latest.last_message_id AND m.created_at < $3
			ORDER BY m.created_at DESC
			LIMIT $4
		`, ":FIELD", strings.Join(fields, ", m."), 1)
	}

	rows, err := db.Query(ctx, selectStmt, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var messages []*domain.Message

	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(database.GetScanFields(m, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}
	return messages, nil
}

// Deprecated
func (r *MessageRepo) FindAllMessageByConversation(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, limit uint, endAt pgtype.Timestamptz) ([]*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.FindAllMessageByConversation")
	defer span.End()

	fields := database.GetFieldNames(&domain.Message{})
	selectStmt := fmt.Sprintf(`
SELECT %s
FROM messages
WHERE conversation_id = $1 AND created_at < $2
ORDER BY created_at DESC
LIMIT $3`, strings.Join(fields, ","))

	rows, err := db.Query(ctx, selectStmt, &conversationID, &endAt, &limit)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx(")
	}
	defer rows.Close()

	messages := []*domain.Message{}
	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(database.GetScanFields(m, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}
	return messages, nil
}

func (r *MessageRepo) CountMessagesSince(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, since *pgtype.Timestamptz) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.CountMessagesSince")
	defer span.End()

	args := []interface{}{&conversationID}
	query := "SELECT COUNT(*) FROM messages WHERE conversation_id = $1"
	if since != nil {
		args = append(args, since)
		query += " AND created_at > $2"
	}

	row := db.QueryRow(ctx, query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "row.Scan")
	}
	return count, nil
}

func (r *MessageRepo) GetLatestMessageByConversation(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.GetLatestMessageByConversation")
	defer span.End()

	m := new(domain.Message)
	fields := database.GetFieldNames(m)

	query := fmt.Sprintf(`
SELECT %s
FROM %s
WHERE message_id = (
	SELECT last_message_id
	FROM %s
	WHERE conversation_id = $1
);`, strings.Join(fields, ","), m.TableName(), (&domain.Conversation{}).TableName())
	row := db.QueryRow(ctx, query, &conversationID)

	if err := row.Scan(database.GetScanFields(m, fields)...); err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}

	return m, nil
}

// Latest message queried from latest_message_field
func (r *MessageRepo) getNonSystemLastMessageByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, limit uint, endAt pgtype.Timestamptz) ([]*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.GetLastMessageByConversationIDs")
	defer span.End()

	m := &domain.Message{}
	fields := database.GetFieldNames(m)
	selectStmt := fmt.Sprintf(`
SELECT %s
FROM %s AS m 
JOIN (
		SELECT conversation_id, last_message_id AS message_id
		FROM %s
		WHERE conversation_id = ANY($1)
	) latest
	USING(message_id, conversation_id)
WHERE m.created_at < $2
ORDER BY m.created_at DESC
LIMIT $3`, strings.Join(fields, ","), m.TableName(), (&domain.Conversation{}).TableName())

	rows, err := db.Query(ctx, selectStmt, &conversationIDs, &endAt, &limit)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	messages := []*domain.Message{}

	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(database.GetScanFields(m, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}
	return messages, nil
}

// Latest message query from timestamp each conversation id
func (r *MessageRepo) GetLastMessageByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, limit uint, endAt pgtype.Timestamptz, includeSystemMsg bool) ([]*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.GetLastMessageByConversationIDs")
	defer span.End()

	if !includeSystemMsg {
		return r.getNonSystemLastMessageByConversationIDs(ctx, db, conversationIDs, limit, endAt)
	}

	m := &domain.Message{}
	fields := database.GetFieldNames(m)
	selectStmt := fmt.Sprintf(`
SELECT %s
FROM messages m1
    INNER JOIN (
        SELECT max(m2.created_at) AS created_at,
            m2.conversation_id
        FROM messages m2
        WHERE m2.conversation_id = ANY($1)
        GROUP BY m2.conversation_id
    ) AS latest USING(
        created_at,
        conversation_id
    )
WHERE m1.created_at < $2
ORDER BY m1.created_at DESC
LIMIT $3
`, strings.Join(fields, ","))

	rows, err := db.Query(ctx, selectStmt, &conversationIDs, &endAt, &limit)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	messages := []*domain.Message{}

	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(database.GetScanFields(m, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}
	return messages, nil
}

func (r *MessageRepo) SoftDelete(ctx context.Context, db database.QueryExecer, userID, id pgtype.Text) error {
	stmt := `UPDATE messages
			SET updated_at = NOW(), deleted_at = NOW(), deleted_by = $2
			WHERE message_id = $1 AND deleted_at IS NULL`
	cmd, err := db.Exec(ctx, stmt, &id, &userID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("not found message to delete")
	}

	return nil
}

func (r *MessageRepo) FindLessonMessages(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, args *domain.FindMessagesArgs) ([]*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.FindLessonMessages")
	defer span.End()

	fields := database.GetFieldNames(&domain.Message{})
	var rows pgx.Rows
	if args.IncludeSystemMsg {
		selectStmt := fmt.Sprintf(`
SELECT m.%s
FROM messages m
INNER JOIN conversation_lesson cl USING(conversation_id)
WHERE cl.conversation_id=$1
AND m.created_at > cl.latest_start_time
AND m.created_at < $2::timestamptz
AND ($3::text[] IS NULL OR m.message = ANY($3))
AND ($4::text[] IS NULL OR NOT (m.message = ANY($4)))
ORDER BY m.created_at DESC
LIMIT $5`, strings.Join(fields, ", m."))
		subRows, err := db.Query(ctx, selectStmt, &conversationID, &args.EndAt, &args.IncludeMessageTypes, &args.ExcludeMessagesTypes, &args.Limit)
		if err != nil {
			return nil, fmt.Errorf("db.Query: %w", err)
		}
		rows = subRows
	} else {
		selectStmt := fmt.Sprintf(`
SELECT m.%s
FROM messages m
INNER JOIN conversation_lesson cl USING(conversation_id)
WHERE cl.conversation_id=$1
AND m.created_at > cl.latest_start_time
AND m.created_at < $2::timestamptz
AND m.type != 'MESSAGE_TYPE_SYSTEM'
ORDER BY m.created_at DESC
LIMIT $3`, strings.Join(fields, ", m."))
		subRows, err := db.Query(ctx, selectStmt, &conversationID, &args.EndAt, &args.Limit)
		if err != nil {
			return nil, fmt.Errorf("db.Query: %w", err)
		}
		rows = subRows
	}

	defer rows.Close()

	messages := []*domain.Message{}
	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(database.GetScanFields(m, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return messages, nil
}

func (r *MessageRepo) FindPrivateLessonMessages(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, args *domain.FindMessagesArgs) ([]*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.FindPrivateLessonMessages")
	defer span.End()

	fields := database.GetFieldNames(&domain.Message{})

	selectStmt := fmt.Sprintf(`
		SELECT m.%s
		FROM messages m
		INNER JOIN conversations cv USING(conversation_id)
		INNER JOIN private_conversation_lesson cl USING(conversation_id)
		WHERE cv.conversation_type = 'CONVERSATION_LESSON_PRIVATE'
		AND cv.status = 'CONVERSATION_STATUS_NONE'
		AND cl.conversation_id=$1
		AND m.created_at > cl.latest_start_time
		AND m.created_at < $2::timestamptz
		AND m.deleted_at IS NULL
		AND m.type != 'MESSAGE_TYPE_SYSTEM'
		ORDER BY m.created_at DESC
		LIMIT $3`,
		strings.Join(fields, ", m."),
	)

	rows, err := db.Query(ctx, selectStmt, &conversationID, &args.EndAt, &args.Limit)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	messages, err := r.getMessagesFromDBRows(rows)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepo) FindMessages(ctx context.Context, db database.QueryExecer, args *domain.FindMessagesArgs) ([]*domain.Message, error) {
	ctx, span := interceptors.StartSpan(ctx, "MessageRepo.FindAllMessageByConversation")
	defer span.End()

	if args.IncludeMessageTypes.Status == pgtype.Present && args.ExcludeMessagesTypes.Status == pgtype.Present {
		return nil, fmt.Errorf("invalid arguments: IncludeMessageTypes and ExcludeMessagesTypes can't be co-exist")
	}

	fields := database.GetFieldNames(&domain.Message{})
	var rows pgx.Rows
	if args.IncludeSystemMsg {
		selectStmt := fmt.Sprintf(`
SELECT %s
FROM messages
WHERE conversation_id = $1
AND created_at < $2
AND ($3::text[] IS NULL OR message = ANY($3))
AND ($4::text[] IS NULL OR NOT (message = ANY($4)))
ORDER BY created_at DESC
LIMIT $5`, strings.Join(fields, ","))

		subRows, err := db.Query(ctx, selectStmt, &args.ConversationID, &args.EndAt, &args.IncludeMessageTypes, &args.ExcludeMessagesTypes, &args.Limit)
		if err != nil {
			return nil, fmt.Errorf("db.Query: %w", err)
		}
		rows = subRows
	} else {
		selectStmt := fmt.Sprintf(`
SELECT %s
FROM messages
WHERE conversation_id = $1
AND created_at < $2
AND type != 'MESSAGE_TYPE_SYSTEM'
ORDER BY created_at DESC
LIMIT $3`, strings.Join(fields, ","))

		subRows, err := db.Query(ctx, selectStmt, &args.ConversationID, &args.EndAt, &args.Limit)
		if err != nil {
			return nil, fmt.Errorf("db.Query: %w", err)
		}
		rows = subRows
	}

	defer rows.Close()

	messages := []*domain.Message{}
	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(database.GetScanFields(m, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return messages, nil
}

func (r *MessageRepo) getMessagesFromDBRows(rows pgx.Rows) ([]*domain.Message, error) {
	fields := database.GetFieldNames(&domain.Message{})
	messages := []*domain.Message{}

	for rows.Next() {
		m := &domain.Message{}
		if err := rows.Scan(database.GetScanFields(m, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return messages, nil
}
