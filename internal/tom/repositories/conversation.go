package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

type ConversationRepo struct {
}

func (r *ConversationRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []*core.Conversation) error {
	ctx, span := trace.StartSpan(ctx, "ConversationRepo.BulkUpsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *core.Conversation) {
		fieldNames := []string{"conversation_id", "name", "status", "conversation_type", "created_at", "updated_at", "owner"}
		placeHolders := "$1, $2, $3, $4, $5, $6, $7"
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT conversations_pk
			DO UPDATE SET name = $2, status = $3, updated_at = $6, owner = $7`,
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

func (r *ConversationRepo) FindConversationIdsBySchoolIds(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*core.Conversation, error) {
	ctx, span := trace.StartSpan(ctx, "ConversationRepo.FindConversationIdsBySchoolIds")
	defer span.End()

	cs := make([]*core.Conversation, 0)

	selectStmt := fmt.Sprintf("SELECT conversation_id FROM conversations WHERE owner = any($1::text[]);")

	rows, err := db.Query(ctx, selectStmt, &ids)
	if err != nil {
		return nil, fmt.Errorf("r.DB.QueryEx: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(core.Conversation)
		if err := rows.Scan(database.GetScanFields(c, []string{"conversation_id"})...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w ", err)
		}

		cs = append(cs, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}

	return cs, nil
}

func (r *ConversationRepo) Create(ctx context.Context, db database.QueryExecer, c *core.Conversation) error {
	ctx, span := trace.StartSpan(ctx, "ConversationRepo.Create")
	defer span.End()

	now := time.Now()
	c.UpdatedAt.Set(now)
	c.CreatedAt.Set(now)

	cmdTag, err := Insert(ctx, c, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new conversation")
	}

	return nil
}

func (r *ConversationRepo) FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*core.Conversation, error) {
	ctx, span := trace.StartSpan(ctx, "ConversationRepo.FindByIDLessonID")
	defer span.End()

	c := new(core.Conversation)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf(`SELECT c.%s FROM conversations c
		LEFT JOIN conversation_lesson cl ON cl.conversation_id = c.conversation_id
		WHERE cl.lesson_id = $1 AND cl.deleted_at IS NULL
	`, strings.Join(fields, ", c."))

	row := db.QueryRow(ctx, selectStmt, &lessonID)

	if err := row.Scan(database.GetScanFields(c, fields)...); err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}

	return c, nil
}

func (r *ConversationRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (c *core.Conversation, err error) {
	ctx, span := trace.StartSpan(ctx, "ConversationRepo.FindByID")
	defer span.End()

	c = new(core.Conversation)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE conversation_id = $1", strings.Join(fields, ","), c.TableName())

	row := db.QueryRow(ctx, selectStmt, &id)

	if err := row.Scan(database.GetScanFields(c, fields)...); err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}

	return
}

func (r *ConversationRepo) FindByStudentQuestionID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (c *core.Conversation, err error) {
	ctx, span := trace.StartSpan(ctx, "ConversationRepo.FindByStudentQuestionID")
	defer span.End()

	c = new(core.Conversation)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_question_id = $1", strings.Join(fields, ","), c.TableName())

	row := db.QueryRow(ctx, selectStmt, &id)

	if err := row.Scan(database.GetScanFields(c, fields)...); err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}

	return
}

func (r *ConversationRepo) FindByIDsReturnMapByID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (map[pgtype.Text]core.ConversationFull, error) {
	ctx, span := trace.StartSpan(ctx, "ConversationRepo.FindByIDsReturnMapByID")
	defer span.End()

	staffRoles := database.TextArray(constant.ConversationStaffRoles)

	c := &core.Conversation{}
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf(`SELECT c.%s, is_reply, conversation_students.student_id
		FROM conversations AS c
		LEFT JOIN conversation_students ON conversation_students.conversation_id = c.conversation_id
		LEFT JOIN (
			SELECT c.conversation_id, 
			(cm.role::text = ANY($1) OR messages.message = 'CODES_MESSAGE_TYPE_CREATED_CONVERSATION') AS is_reply FROM messages JOIN conversations c on c.last_message_id = messages.message_id LEFT JOIN conversation_members cm ON messages.user_id = cm.user_id AND c.conversation_id=cm.conversation_id) AS v ON c.conversation_id=v.conversation_id		
		WHERE c.conversation_id = ANY($2)`, strings.Join(fields, ", c."))

	rows, err := db.Query(ctx, selectStmt, &staffRoles, &ids)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	conversations := make(map[pgtype.Text]core.ConversationFull)

	for rows.Next() {
		c := core.Conversation{}
		var (
			isReply   pgtype.Bool
			studentID pgtype.Text
		)
		scanFields := database.GetScanFields(&c, fields)
		scanFields = append(scanFields, &isReply, &studentID)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		conversations[c.ID] = core.ConversationFull{
			Conversation: c,
			// StudentQuestionID: studentQuestionID,
			// ClassID:           classID,
			IsReply:   isReply,
			StudentID: studentID,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}
	return conversations, nil
}

func (r *ConversationRepo) Update(ctx context.Context, db database.QueryExecer, c *core.Conversation) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.Update")
	defer span.End()

	now := time.Now()
	c.UpdatedAt.Set(now)

	cmdTag, err := Update(ctx, c, db.Exec, "conversation_id")
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update conversation")
	}

	return nil
}

func (r *ConversationRepo) SetStatus(ctx context.Context, db database.QueryExecer, cID pgtype.Text, status pgtype.Text) error {
	stmt := "UPDATE conversations SET status = $2, updated_at = NOW() WHERE conversation_id = $1"
	commandTag, err := db.Exec(ctx, stmt, &cID, &status)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("can not update conversation_members")
	}

	return nil
}

func (r *ConversationRepo) SetName(ctx context.Context, db database.QueryExecer, cIDs pgtype.TextArray, name pgtype.Text) error {
	stmt := "UPDATE conversations SET name = $2, updated_at = NOW() WHERE conversation_id = ANY($1)"
	commandTag, err := db.Exec(ctx, stmt, &cIDs, &name)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if commandTag.RowsAffected() != int64(len(cIDs.Elements)) {
		return errors.New("cannot update all conversations")
	}

	return nil
}

const findBySchoolID = `SELECT c.conversation_id, max(m.updated_at) updated_at
FROM conversations AS c
JOIN messages m ON c.conversation_id = m.conversation_id
WHERE c.owner = ANY($1)
	AND c.status = 'CONVERSATION_STATUS_NONE'
	AND ($2::text IS NULL OR c.conversation_id > $2)
GROUP BY c.conversation_id
ORDER BY updated_at, c.conversation_id
LIMIT $3::int`

func (r *ConversationRepo) FindBySchoolIDs(ctx context.Context, db database.QueryExecer, schoolIDs pgtype.TextArray, limit pgtype.Int4, offset pgtype.Text) ([]pgtype.Text, []pgtype.Timestamptz, error) {
	ctx, span := trace.StartSpan(ctx, "ConversationRepo.FindByIDsReturnMapByID")
	defer span.End()

	rows, err := db.Query(ctx, findBySchoolID, schoolIDs, offset, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("r.DB.QueryEx: %w", err)
	}
	defer rows.Close()

	var conversationIDs []pgtype.Text
	var updatedTime []pgtype.Timestamptz

	for rows.Next() {
		var conversationID pgtype.Text
		var updatedAt pgtype.Timestamptz
		if err := rows.Scan(&conversationID, &updatedAt); err != nil {
			return nil, nil, fmt.Errorf("row.Scan: %w", err)
		}
		conversationIDs = append(conversationIDs, conversationID)
		updatedTime = append(updatedTime, updatedAt)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row.Err: %w", err)
	}

	return conversationIDs, updatedTime, nil
}

func (r *ConversationRepo) ListAll(ctx context.Context, db database.QueryExecer, offsetID pgtype.Text, limit uint32, conversationTypesAccepted pgtype.TextArray, schoolID pgtype.Text) ([]*core.Conversation, error) {
	e := &core.Conversation{}
	stmtPtl := `SELECT %s FROM conversations WHERE ($1::TEXT IS NULL OR conversation_id > $1::TEXT) AND conversation_type = ANY($3::_text) AND owner=$4 ORDER BY conversation_id ASC LIMIT $2`
	var result core.Conversations
	fields, _ := e.FieldMap()
	err := database.Select(ctx, db, fmt.Sprintf(stmtPtl, strings.Join(fields, ",")), offsetID, limit, conversationTypesAccepted, schoolID).ScanAll(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

const queryUnjoinedWithLocations = `SELECT c1.%s
	FROM conversations c1 LEFT JOIN conversation_locations cl USING(conversation_id)
	WHERE cl.access_path like ANY($3) AND cl.deleted_at IS NULL
	AND status = 'CONVERSATION_STATUS_NONE'
	AND owner = ANY($1::_TEXT)
	AND c1.conversation_id NOT IN (
		SELECT cm.conversation_id FROM conversation_members cm WHERE cm.user_id=$2::TEXT AND cm.status='CONVERSATION_STATUS_ACTIVE'
	)`

func (r *ConversationRepo) ListConversationUnjoinedInLocations(ctx context.Context, db database.QueryExecer, filter *core.ListConversationUnjoinedFilter) ([]*core.Conversation, error) {
	e := &core.Conversation{}
	var conversations core.Conversations
	fields, _ := e.FieldMap()
	aps := database.FromTextArray(filter.AccessPaths)
	if len(aps) > 20 {
		return nil, fmt.Errorf("too many location filters applied: %d", len(aps))
	}
	likeAps := make([]string, 0, len(aps))
	for _, ap := range aps {
		likeAps = append(likeAps, ap+"%")
	}
	err := database.Select(ctx, db,
		fmt.Sprintf(queryUnjoinedWithLocations, strings.Join(fields, ",c1.")), filter.OwnerIDs, filter.UserID,
		database.TextArray(likeAps),
	).ScanAll(&conversations)
	if err != nil {
		return nil, err
	}
	return conversations, nil
}

const stmtPtl = `SELECT %s 
	FROM conversations
	WHERE status = 'CONVERSATION_STATUS_NONE'
	AND owner = ANY($1::_TEXT) 
	EXCEPT SELECT c.%s FROM conversations c JOIN conversation_members cm USING(conversation_id) WHERE cm.user_id=$2::TEXT AND cm.status='CONVERSATION_STATUS_ACTIVE'
	`

// ListConversationUnjoined list all conversation that the user haven't joined with given user_id and owner_ids
func (r *ConversationRepo) ListConversationUnjoined(ctx context.Context, db database.QueryExecer, filter *core.ListConversationUnjoinedFilter) ([]*core.Conversation, error) {
	e := &core.Conversation{}
	var conversations core.Conversations
	fields, _ := e.FieldMap()
	err := database.Select(ctx, db, fmt.Sprintf(stmtPtl, strings.Join(fields, ","), strings.Join(fields, ",c.")), &filter.OwnerIDs, &filter.UserID).ScanAll(&conversations)
	if err != nil {
		return nil, err
	}
	return conversations, nil
}

// db should be a tx
func (r *ConversationRepo) BulkUpdateResourcePath(ctx context.Context, db database.QueryExecer, convIDs []string, resourcePath string) error {
	updateConv := `
 update conversations c set resource_path = $1
 where c.conversation_id = ANY($2)
 and (c.resource_path is null or length(c.resource_path)=0)
`

	_, err := db.Exec(ctx, updateConv, database.Text(resourcePath), database.TextArray(convIDs))
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	updateMembers := `
update conversation_members cm set resource_path=c.resource_path
from conversations c
where cm.conversation_id=c.conversation_id 
and (cm.resource_path is null or length(cm.resource_path)=0)
and c.conversation_id = ANY($1)
	`
	_, err = db.Exec(ctx, updateMembers, database.TextArray(convIDs))
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	updateMessages := `
update messages m set resource_path=c.resource_path
from conversations c
where m.conversation_id=c.conversation_id 
and (m.resource_path is null or length(m.resource_path)=0)
and c.conversation_id = ANY($1)`
	_, err = db.Exec(ctx, updateMessages, database.TextArray(convIDs))
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}

const countTotalUnreadConversation = `
	SELECT count(*) 
	FROM messages m2 JOIN conversations c ON m2.message_id = c.last_message_id
	JOIN conversation_members cm ON cm.conversation_id = c.conversation_id
	WHERE cm.user_id = $1
	AND m2.deleted_at IS NULL
	AND (cm.seen_at is NULL OR cm.seen_at < m2.created_at)
	AND ($2::text[] IS NULL OR c.conversation_type = ANY($2))
	AND c.status = 'CONVERSATION_STATUS_NONE'
	AND cm.status = 'CONVERSATION_STATUS_ACTIVE'
	AND m2."type" != 'MESSAGE_TYPE_SYSTEM'
	`

const countTotalUnreadConversationV2 = `
	SELECT count(DISTINCT message_id)
	FROM messages m2 JOIN conversations c ON m2.message_id = c.last_message_id
	JOIN conversation_members cm ON cm.conversation_id = c.conversation_id
	JOIN conversation_locations cl ON cl.conversation_id = c.conversation_id 
	WHERE cm.user_id = $1
	AND m2.deleted_at IS NULL
	AND (cm.seen_at is NULL OR cm.seen_at < m2.created_at)
	AND ($2::text[] IS NULL OR c.conversation_type = ANY($2))
	AND cl.location_id = ANY($3)
	AND c.status = 'CONVERSATION_STATUS_NONE'
	AND cm.status = 'CONVERSATION_STATUS_ACTIVE'
	AND m2."type" != 'MESSAGE_TYPE_SYSTEM'
	`

func (r *ConversationRepo) CountUnreadConversations(ctx context.Context, db database.QueryExecer, userID pgtype.Text, msgType pgtype.TextArray, locationIDs pgtype.TextArray, enableChatThreadLaunching bool) (int64, error) {
	var totalUnreadConv int64
	var err error
	if enableChatThreadLaunching {
		err = db.QueryRow(ctx, countTotalUnreadConversationV2, &userID, msgType, locationIDs).Scan(&totalUnreadConv)
	} else {
		err = db.QueryRow(ctx, countTotalUnreadConversation, &userID, msgType).Scan(&totalUnreadConv)
	}

	if err != nil {
		return 0, fmt.Errorf("QueryRow: %w", err)
	}

	return totalUnreadConv, nil
}

const countTotalUnreadConversationWithLocations = `
	SELECT count(*) 
	FROM messages m2 JOIN conversations c ON m2.message_id = c.last_message_id
	JOIN conversation_members cm ON cm.conversation_id = c.conversation_id
	LEFT JOIN conversation_locations cl ON cm.conversation_id=cl.conversation_id
	WHERE cm.user_id = $1
	AND m2.deleted_at IS NULL
	AND (cm.seen_at is NULL OR cm.seen_at < m2.created_at)
	AND ($2::text[] IS NULL OR c.conversation_type = ANY($2))
	AND cl.access_path LIKE ANY($3)
	AND c.status = 'CONVERSATION_STATUS_NONE'
	AND cm.status = 'CONVERSATION_STATUS_ACTIVE'
	AND m2."type" != 'MESSAGE_TYPE_SYSTEM'
	`

const countTotalUnreadConversationWithLocationsV2 = `
	SELECT count(DISTINCT message_id)
	FROM messages m2 JOIN conversations c ON m2.message_id = c.last_message_id
	JOIN conversation_members cm ON cm.conversation_id = c.conversation_id
	LEFT JOIN conversation_locations cl ON cm.conversation_id=cl.conversation_id
	WHERE cm.user_id = $1
	AND m2.deleted_at IS NULL
	AND (cm.seen_at is NULL OR cm.seen_at < m2.created_at)
	AND ((c.conversation_type = 'CONVERSATION_STUDENT' AND cl.access_path LIKE ANY($2)) OR (c.conversation_type = 'CONVERSATION_PARENT' AND cl.access_path LIKE ANY($3)))
	AND c.status = 'CONVERSATION_STATUS_NONE'
	AND cm.status = 'CONVERSATION_STATUS_ACTIVE'
	AND m2."type" != 'MESSAGE_TYPE_SYSTEM'
`

// TODO: This query may be slow, => use Elasticsearch instead, together with epic teacher read status, after this
// elasticsearch document will have info about read status of each conversation member
func (r *ConversationRepo) CountUnreadConversationsByAccessPaths(ctx context.Context, db database.QueryExecer, userID pgtype.Text, msgType pgtype.TextArray, accessPaths pgtype.TextArray) (int64, error) {
	var totalUnreadConv int64
	aps := database.FromTextArray(accessPaths)
	likeAps := make([]string, 0, len(aps))
	for _, ap := range aps {
		likeAps = append(likeAps, ap+"%")
	}
	err := db.QueryRow(ctx, countTotalUnreadConversationWithLocations, userID, msgType, database.TextArray(likeAps)).Scan(&totalUnreadConv)
	if err != nil {
		return 0, fmt.Errorf("QueryRow: %w", err)
	}

	return totalUnreadConv, nil
}

func (r *ConversationRepo) CountUnreadConversationsByAccessPathsV2(ctx context.Context, db database.QueryExecer, userID pgtype.Text, studentAccessPaths pgtype.TextArray, parentAccessPaths pgtype.TextArray) (int64, error) {
	var totalUnreadConv int64
	studentAps := database.FromTextArray(studentAccessPaths)
	likeStudentAps := make([]string, 0, len(studentAps))
	for _, ap := range studentAps {
		likeStudentAps = append(likeStudentAps, ap+"%")
	}
	parentAps := database.FromTextArray(parentAccessPaths)
	likeParentAps := make([]string, 0, len(parentAps))
	for _, ap := range parentAps {
		likeParentAps = append(likeParentAps, ap+"%")
	}
	err := db.QueryRow(ctx, countTotalUnreadConversationWithLocationsV2, userID, database.TextArray(likeStudentAps), database.TextArray(likeParentAps)).Scan(&totalUnreadConv)
	if err != nil {
		return 0, fmt.Errorf("QueryRow: %w", err)
	}

	return totalUnreadConv, nil
}
