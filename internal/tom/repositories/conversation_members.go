package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ConversationMemberRepo struct {
}

func (rcv *ConversationMemberRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, conversationMembers []*entities.ConversationMembers) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.BulkUpsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *entities.ConversationMembers) {
		fieldNames := database.GetFieldNames(e)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT conversation_statuses__user_id__conversation_id_un 
		DO UPDATE SET updated_at = $9, status = $5, role = $4`,
			e.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, c := range conversationMembers {
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

func (rcv *ConversationMemberRepo) Create(ctx context.Context, db database.QueryExecer, c *entities.ConversationMembers) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.Create")
	defer span.End()

	now := time.Now()
	c.UpdatedAt.Set(now)
	c.CreatedAt.Set(now)

	fields := []string{"conversation_statuses_id", "user_id", "conversation_id", "role", "status", "seen_at", "last_notify_at", "created_at", "updated_at"}
	placeHolders := generateInsertPlaceholders(len(fields))

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) "+
		"ON CONFLICT ON CONSTRAINT conversation_statuses__user_id__conversation_id_un "+
		"DO UPDATE SET updated_at = $9, status = $5, role = $4;", c.TableName(), strings.Join(fields, ", "), placeHolders)
	args := database.GetScanFields(c, fields)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return errors.Wrap(err, "db.Exec")
	}
	return nil
}

func (rcv *ConversationMemberRepo) Update(ctx context.Context, db database.QueryExecer, c *entities.ConversationMembers) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.Update")
	defer span.End()

	now := time.Now()
	c.UpdatedAt.Set(now)
	c.CreatedAt.Set(now)

	cmdTag, err := Update(ctx, c, db.Exec, "conversation_statuses_id")
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new conversation")
	}

	return nil
}

func (rcv *ConversationMemberRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (c *entities.ConversationMembers, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.FindByID")
	defer span.End()

	c = new(entities.ConversationMembers)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE conversation_statuses_id = $1", strings.Join(fields, ","), c.TableName())

	row := db.QueryRow(ctx, selectStmt, &id)

	if err := row.Scan(database.GetScanFields(c, fields)...); err != nil {
		return nil, err
	}

	return
}

func (rcv *ConversationMemberRepo) Find(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, role, status pgtype.Text) (*entities.ConversationMembers, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.Find")
	defer span.End()

	c := new(entities.ConversationMembers)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s "+
		"FROM %s "+
		"WHERE conversation_id = $1 AND role = $2 AND status = $3", strings.Join(fields, ","), c.TableName())

	row := db.QueryRow(ctx, selectStmt, &conversationID, &role, &status)

	if err := row.Scan(database.GetScanFields(c, fields)...); err != nil {
		return nil, err
	}

	return c, nil
}

func (rcv *ConversationMemberRepo) FindByConversationIDsAndRoles(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, roles pgtype.TextArray) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.FindByConversationIDsAndRoles")
	defer span.End()

	c := &entities.ConversationMembers{}
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf(`SELECT %s FROM %s WHERE 
		conversation_id = ANY($1)
		AND role = ANY($2)
		AND status = 'CONVERSATION_STATUS_ACTIVE'`, strings.Join(fields, ","), c.TableName(),
	)

	rows, err := db.Query(ctx, selectStmt, &conversationIDs, &roles)
	if err != nil {
		return nil, errors.Wrap(err, "db.QueryEx")
	}
	defer rows.Close()

	conversationMembersMap := make(map[string][]string)

	for rows.Next() {
		c := &entities.ConversationMembers{}
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		conversationMembersMap[c.ConversationID.String] = append(conversationMembersMap[c.ConversationID.String], c.UserID.String)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}

	return conversationMembersMap, nil
}

func (rcv *ConversationMemberRepo) FindByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (mapConversationID map[pgtype.Text][]*entities.ConversationMembers, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.FindByConversationIDs")
	defer span.End()

	c := &entities.ConversationMembers{}
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE conversation_id = ANY($1)", strings.Join(fields, ","), c.TableName())

	rows, err := db.Query(ctx, selectStmt, &conversationIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.QueryEx")
	}
	defer rows.Close()

	conversations := make(map[pgtype.Text][]*entities.ConversationMembers)

	for rows.Next() {
		c := &entities.ConversationMembers{}
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		conversations[c.ConversationID] = append(conversations[c.ConversationID], c)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}

	return conversations, nil
}

func (rcv *ConversationMemberRepo) FindByConversationIDAndStatus(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, status pgtype.Text) (map[pgtype.Text]entities.ConversationMembers, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.FindByConversationID")
	defer span.End()

	c := &entities.ConversationMembers{}
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM conversation_members WHERE conversation_id = $1 AND ($2::text IS NULL or status=$2)", strings.Join(fields, ","))

	rows, err := db.Query(ctx, selectStmt, &conversationID, status)
	if err != nil {
		return nil, errors.Wrap(err, "db.QueryEx")
	}
	defer rows.Close()

	conversations := make(map[pgtype.Text]entities.ConversationMembers)

	for rows.Next() {
		c := entities.ConversationMembers{}
		if err := rows.Scan(database.GetScanFields(&c, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		conversations[c.UserID] = c
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}

	return conversations, nil
}

func (rcv *ConversationMemberRepo) FindByConversationID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (map[pgtype.Text]entities.ConversationMembers, error) {
	return rcv.FindByConversationIDAndStatus(ctx, db, conversationID, database.Text(entities.ConversationStatusActive))
}

func (rcv *ConversationMemberRepo) FindUnseenSince(ctx context.Context, db database.QueryExecer, t pgtype.Timestamptz) ([]*entities.ConversationMembers, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.FindUnseenSince")
	defer span.End()

	c := &entities.ConversationMembers{}
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE seen_at IS NULL OR seen_at <= $1", strings.Join(fields, ","), c.TableName())

	rows, err := db.Query(ctx, selectStmt, &t)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	var cc []*entities.ConversationMembers
	for rows.Next() {
		c := new(entities.ConversationMembers)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		cc = append(cc, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row.Err")
	}

	return cc, nil
}

func (rcv *ConversationMemberRepo) GetSeenAt(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*pgtype.Timestamptz, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.GetSeenAt")
	defer span.End()

	seenAt := new(pgtype.Timestamptz)

	query := "SELECT seen_at FROM conversation_members WHERE conversation_statuses_id = $1 FOR SHARE"
	row := db.QueryRow(ctx, query, &id)
	if err := row.Scan(seenAt); err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}
	if seenAt.Status != pgtype.Present {
		return nil, nil
	}
	return seenAt, nil
}

func (rcv *ConversationMemberRepo) SetSeenAt(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text, seenAt pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.SetSeenAt")
	defer span.End()

	return rcv.updateField(ctx, db, conversationID, userID, "seen_at", &seenAt)
}

func (rcv *ConversationMemberRepo) SetNotifyAt(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text, notifyAt pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.SetSeenAt")
	defer span.End()

	return rcv.updateField(ctx, db, conversationID, userID, "last_notify_at", &notifyAt)
}

func (rcv *ConversationMemberRepo) SetStatus(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userIDs pgtype.TextArray, status pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.SetStatus")
	defer span.End()
	stmt := "UPDATE conversation_members SET status = $3, updated_at = NOW() WHERE conversation_id = $1 AND user_id = ANY($2)"
	commandTag, err := db.Exec(ctx, stmt, &conversationID, &userIDs, &status)
	if err != nil {
		return errors.Wrap(err, "db.ExecEx")
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("can not update conversation_members")
	}

	return nil
}

func (rcv *ConversationMemberRepo) SetStatusByConversationAndUserIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, userIDs pgtype.TextArray, status pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.SetStatusByConversationAndUserIDs")
	defer span.End()

	stmt := "UPDATE conversation_members SET status = $3, updated_at = NOW() WHERE conversation_id = ANY($1) AND user_id = ANY($2)"
	commandTag, err := db.Exec(ctx, stmt, &conversationIDs, &userIDs, &status)
	if err != nil {
		return errors.Wrap(err, "db.ExecEx")
	}

	if commandTag.RowsAffected() == 0 {
		return errors.New("can not update conversation_members")
	}

	return nil
}

func (rcv *ConversationMemberRepo) updateField(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text, field string, value interface{}) error {
	stmt := "UPDATE conversation_members SET " + field + " = $3, updated_at = NOW() WHERE user_id = $1 AND conversation_id = $2"
	commandTag, err := db.Exec(ctx, stmt, &userID, &conversationID, value)
	if err != nil {
		return errors.Wrap(err, "db.ExecEx")
	}

	if commandTag.RowsAffected() != 1 {
		return errors.New("can not update conversation_members")
	}

	return nil
}

func (rcv *ConversationMemberRepo) FindByCIDAndUserID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text) (c *entities.ConversationMembers, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.FindByCIDAndUserID")
	defer span.End()

	c = new(entities.ConversationMembers)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND conversation_id = $2", strings.Join(fields, ","), c.TableName())

	row := db.QueryRow(ctx, selectStmt, &userID, &conversationID)

	if err := row.Scan(database.GetScanFields(c, fields)...); err != nil {
		return nil, err
	}

	return
}

func (rcv *ConversationMemberRepo) UserGroup(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.UserGroup")
	defer span.End()

	c := new(entities.ConversationMembers)
	selectStmt := fmt.Sprintf("SELECT role FROM %s WHERE user_id = $1", c.TableName())

	var userGroup pgtype.Text
	row := db.QueryRow(ctx, selectStmt, &userID)
	if err := row.Scan(&userGroup); err != nil {
		return "", err
	}

	return userGroup.String, nil
}

func (rcv *ConversationMemberRepo) SetStatusByConversationID(ctx context.Context, db database.QueryExecer, conversationID, status pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.InactiveMembersOfLessonID")
	defer span.End()

	stmt := `
		UPDATE conversation_members
		SET status = $1, updated_at = NOW()
		WHERE conversation_id = $2
	`

	commandTag, err := db.Exec(ctx, stmt, &status, &conversationID)
	if err != nil {
		return fmt.Errorf("db.ExecEx: %v", err)
	}
	if commandTag.RowsAffected() == 0 {
		return errors.New("can not update conversation_members")
	}

	return nil
}

func (rcv *ConversationMemberRepo) FindByCIDsAndUserID(ctx context.Context, db database.QueryExecer, cIDs pgtype.TextArray, userID pgtype.Text) (ret []*entities.ConversationMembers, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.FindByCIDsAndUserID")
	defer span.End()

	ret = make([]*entities.ConversationMembers, 0, len(cIDs.Elements))
	fields := database.GetFieldNames(&entities.ConversationMembers{})
	selectStmt := fmt.Sprintf("SELECT %s FROM conversation_members WHERE user_id = $1 AND conversation_id = ANY($2)", strings.Join(fields, ","))

	rows, err := db.Query(ctx, selectStmt, &userID, &cIDs)
	if err != nil {
		err = fmt.Errorf("db.Query: %w", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := &entities.ConversationMembers{}
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		ret = append(ret, c)
	}
	return
}

func (rcv *ConversationMemberRepo) FindUserIDConversationIDsMapByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationMemberRepo.FindByConversationIDsAndUserIDs")
	defer span.End()

	fields := database.GetFieldNames(&entities.ConversationMembers{})
	selectStmt := fmt.Sprintf(`SELECT cm.%s FROM conversation_members cm
	INNER JOIN conversation_students cs
		ON cm.conversation_id = cs.conversation_id 
	WHERE cm.user_id = ANY($1)
		AND cm.status = 'CONVERSATION_STATUS_ACTIVE'
		AND cs.deleted_at IS NULL`, strings.Join(fields, ",cm."))

	rows, err := db.Query(ctx, selectStmt, &userIDs)
	if err != nil {
		err = fmt.Errorf("db.Query: %w", err)
		return nil, err
	}
	defer rows.Close()

	ret := make(map[string][]string)

	for rows.Next() {
		c := &entities.ConversationMembers{}
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		ret[c.UserID.String] = append(ret[c.UserID.String], c.ConversationID.String)
	}
	return ret, nil
}
