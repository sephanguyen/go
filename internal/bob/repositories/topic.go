package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type TopicRepo struct{}

func (r *TopicRepo) RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.RetrieveByIDs")
	defer span.End()

	fields := database.GetFieldNames(&entities.Topic{})
	query := fmt.Sprintf("SELECT %s FROM topics WHERE topic_id = ANY($1) AND topics.deleted_at IS NULL ORDER BY display_order ASC", strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []*entities.Topic
	for rows.Next() {
		p := new(entities.Topic)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *TopicRepo) RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...QueryEnhancer) (*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.RetrieveByID")
	defer span.End()

	e := new(entities.Topic)
	fields, values := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM topics WHERE topic_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","))
	for _, e := range enhancers {
		e(&query)
	}
	if err := db.QueryRow(ctx, query, &id).Scan(values...); err != nil {
		return nil, errors.Wrap(err, "db.QueryRow.Scan")
	}

	return e, nil
}

func (r *TopicRepo) Retrieve(ctx context.Context, db database.QueryExecer, country, subject, topicType, status string, grade int) ([]*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.Retrieve")
	defer span.End()

	fields := database.GetFieldNames(&entities.Topic{})
	var args []interface{}
	query := fmt.Sprintf("SELECT %s FROM topics WHERE topics.deleted_at IS NULL", strings.Join(fields, ","))
	if country != "" {
		if strings.Index(query, "WHERE") > -1 {
			query += fmt.Sprintf(" AND country = $%d", len(args)+1)
		} else {
			query += fmt.Sprintf(" WHERE country = $1")
		}
		args = append(args, country)
	}
	if subject != "" {
		if strings.Index(query, "WHERE") > -1 {
			query += fmt.Sprintf(" AND subject = $%d", len(args)+1)
		} else {
			query += fmt.Sprintf(" WHERE subject = $1")
		}
		args = append(args, subject)
	}
	if topicType != "" {
		if strings.Index(query, "WHERE") > -1 {
			query += fmt.Sprintf(" AND topic_type = $%d", len(args)+1)
		} else {
			query += fmt.Sprintf(" WHERE topic_type = $1")
		}
		args = append(args, topicType)
	}
	if grade != -1 {
		if strings.Index(query, "WHERE") > -1 {
			query += fmt.Sprintf(" AND grade = $%d", len(args)+1)
		} else {
			query += fmt.Sprintf(" WHERE grade = $1")
		}
		args = append(args, grade)
	}
	if status != "" {
		if strings.Index(query, "WHERE") > -1 {
			query += fmt.Sprintf(" AND status = $%d", len(args)+1)
		} else {
			query += fmt.Sprintf(" WHERE status = $1")
		}
		args = append(args, status)
	}
	query += " ORDER BY display_order ASC"

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []*entities.Topic
	for rows.Next() {
		p := new(entities.Topic)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *TopicRepo) BulkImport(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.BulkImport")
	defer span.End()

	queueFn := func(b *pgx.Batch, t *entities.Topic) {
		fieldNames := []string{"topic_id", "name", "country", "grade", "subject", "topic_type", "created_at", "updated_at", "status", "display_order", "icon_url"}
		placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11"
		query := "INSERT INTO topics (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT topics_pk DO UPDATE SET name = $2, country = $3, grade = $4, subject = $5, topic_type = $6, updated_at = $8, display_order = $10, icon_url = $11"
		if t.ChapterID.Status == pgtype.Present {
			fieldNames = append(fieldNames, "chapter_id")
			placeHolders += fmt.Sprintf(", $%d", len(fieldNames))
			query += fmt.Sprintf(", chapter_id = $%d", len(fieldNames))
		}
		if t.SchoolID.Status == pgtype.Present {
			fieldNames = append(fieldNames, "school_id")
			placeHolders += fmt.Sprintf(", $%d", len(fieldNames))
			query += fmt.Sprintf(", school_id = $%d", len(fieldNames))
		}

		stmt := fmt.Sprintf(query, strings.Join(fieldNames, ","), placeHolders)
		b.Queue(stmt, database.GetScanFields(t, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, topic := range topics {
		now := time.Now()
		_ = topic.UpdatedAt.Set(now)
		_ = topic.CreatedAt.Set(now)
		topic.DeletedAt.Set(nil)
		/*Always set topic to draft when insert
		  Not allow to update status through upsert func*/
		topic.Status.Set(pb.TOPIC_STATUS_DRAFT.String())
		queueFn(b, topic)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(topics); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("topic not inserted")
		}
	}

	return nil
}

func (r *TopicRepo) GetTopicFromLoId(ctx context.Context, db database.QueryExecer, loID pgtype.Text) (*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.GetTopicFromLoId")
	defer span.End()

	fieldNames := []string{"topic_id", "name", "country", "grade", "subject", "topic_type", "created_at", "updated_at"}

	fmts := "SELECT a.%s FROM topics a INNER JOIN " +
		"(SELECT topic_id FROM learning_objectives WHERE lo_id=$1 AND learning_objectives.deleted_at IS NULL) b USING(topic_id) " +
		"WHERE a.deleted_at IS NULL"
	query := fmt.Sprintf(fmts, strings.Join(fieldNames, ", a."))
	rows := db.QueryRow(ctx, query, &loID)
	p := new(entities.Topic)
	if err := rows.Scan(database.GetScanFields(p, fieldNames)...); err != nil {
		return nil, errors.Wrap(err, "rows.Scan")
	}
	return p, nil
}

func (r *TopicRepo) UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.UpdateTotalLOs")
	defer span.End()

	query := `UPDATE topics SET total_los = sub.total_los, updated_at = $2
	FROM (
		SELECT COUNT(*) AS total_los FROM learning_objectives WHERE topic_id = $1 AND learning_objectives.deleted_at IS NULL
	) sub
	WHERE topic_id = $1 AND topics.deleted_at IS NULL`

	var now pgtype.Timestamptz
	now.Set(time.Now().UTC())

	cmdTag, err := db.Exec(ctx, query, &topicID, &now)
	if err != nil {
		return errors.Wrap(err, "db.Exec")
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.Errorf("cannot update total_los for topic: %s", topicID.String)
	}
	return nil
}

func (r *TopicRepo) UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.UpdateLODisplayOrderCounter")
	defer span.End()
	e := entities.Topic{}
	updateLODisplayOrderCounterStmtPlt := fmt.Sprintf(
		`UPDATE %s SET lo_display_order_counter = lo_display_order_counter + $1::INT, updated_at = now() WHERE topic_id = $2::TEXT`,
		e.TableName(),
	)

	if _, err := db.Exec(ctx, updateLODisplayOrderCounterStmtPlt, number, topicID); err != nil {
		return err
	}

	return nil
}

func (r *TopicRepo) FindByChapterIds(ctx context.Context, db database.QueryExecer, chapterIds []string) ([]*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.FindByChapterIds")
	defer span.End()

	fields := database.GetFieldNames(&entities.Topic{})
	query := fmt.Sprintf("SELECT %s FROM topics WHERE chapter_id = ANY($1) AND deleted_at IS NULL ORDER BY display_order ASC", strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, &chapterIds)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []*entities.Topic
	for rows.Next() {
		p := new(entities.Topic)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *TopicRepo) FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.FindByBookIDs")
	defer span.End()

	fields := database.GetFieldNames(&entities.Topic{})
	args := []interface{}{&bookIDs, &topicIDs}
	query := `SELECT t.%s 
		FROM topics AS t 
		INNER JOIN chapters AS c ON t.chapter_id = c.chapter_id 
		INNER JOIN books_chapters AS bc ON bc.chapter_id = c.chapter_id 
		WHERE ($1::TEXT[] IS NULL OR bc.book_id = ANY($1)) AND
			($2::TEXT[] IS NULL OR t.topic_id = ANY($2)) AND 
			bc.deleted_at IS NULL AND 
			t.deleted_at IS NULL AND 
			c.deleted_at IS NULL 
		ORDER BY bc.book_id, c.display_order, t.display_order, t.updated_at `

	if limit.Status != pgtype.Null && offset.Status != pgtype.Null {
		query += `LIMIT $3::INT 
		OFFSET $4::INT`
		args = append(args, &limit, &offset)
	}
	query = fmt.Sprintf(query, strings.Join(fields, ", t."))

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("r.DB.QueryEx: %w", err)
	}
	defer rows.Close()

	var pp []*entities.Topic
	for rows.Next() {
		p := new(entities.Topic)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return pp, nil
}

// Create creates new topic with generated ULID
func (r *TopicRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.Topic) (string, error) {
	now := timeutil.Now()
	topicID := idutil.ULID(now)

	err := multierr.Combine(
		e.ID.Set(topicID),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	e.DeletedAt.Set(nil)
	if err != nil {
		return "", err
	}

	cmdTag, err := database.Insert(ctx, e, db.Exec)
	if err == nil && cmdTag.RowsAffected() != 1 {
		return "", ErrUnAffected
	}

	return topicID, nil
}

func (r *TopicRepo) DuplicateTopics(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray, newChapterIDs pgtype.TextArray) ([]*entities.CopiedTopic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.DuplicateTopics")
	defer span.End()

	var now pgtype.Timestamptz
	now.Set(time.Now().UTC())

	queue := func(b *pgx.Batch, chapterID pgtype.Text, newChapterID pgtype.Text) {
		e := &entities.Topic{}
		topicFieldNames := database.GetFieldNames(e)
		selectFields := golibs.Replace(
			topicFieldNames,
			[]string{"topic_id", "copied_topic_id", "chapter_id", "created_at", "updated_at"},
			[]string{"uuid_generate_v4()", "topic_id", "$1", "$2", "$2"},
		)
		query := `
		INSERT INTO
			%s (%s)
		SELECT
			%s
		FROM
			%s
		WHERE
			chapter_id = $3 
		RETURNING topic_id ,
			copied_topic_id; `
		stmt := fmt.Sprintf(query, e.TableName(), strings.Join(topicFieldNames, ", "), strings.Join(selectFields, ", "), e.TableName())
		b.Queue(stmt, &newChapterID, &now, &chapterID)
	}

	b := &pgx.Batch{}

	for i, chapterID := range chapterIDs.Elements {
		queue(b, chapterID, newChapterIDs.Elements[i])
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	var copiedTopics []*entities.CopiedTopic

	for i := 0; i < len(chapterIDs.Elements); i++ {
		rows, err := batchResults.Query()
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			copiedTopic := &entities.CopiedTopic{}
			if err := rows.Scan(&copiedTopic.ID, &copiedTopic.CopyFromID); err != nil {
				return nil, err
			}
			copiedTopics = append(copiedTopics, copiedTopic)
		}
	}
	return copiedTopics, nil
}

func (r *TopicRepo) BulkUpsertWithoutDisplayOrder(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.BulkUpsertWithoutDisplayOrder")
	defer span.End()

	queueFn := func(b *pgx.Batch, t *entities.Topic) {
		fieldNames := []string{"topic_id", "name", "country", "grade", "subject", "topic_type", "created_at", "updated_at", "status", "display_order", "icon_url"}
		placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11"
		query := "INSERT INTO topics (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT topics_pk DO UPDATE SET name = $2, country = $3, grade = $4, subject = $5, topic_type = $6, updated_at = $8, icon_url = $11"
		if t.ChapterID.Status == pgtype.Present {
			fieldNames = append(fieldNames, "chapter_id")
			placeHolders += fmt.Sprintf(", $%d", len(fieldNames))
			query += fmt.Sprintf(", chapter_id = $%d", len(fieldNames))
		}
		if t.SchoolID.Status == pgtype.Present {
			fieldNames = append(fieldNames, "school_id")
			placeHolders += fmt.Sprintf(", $%d", len(fieldNames))
			query += fmt.Sprintf(", school_id = $%d", len(fieldNames))
		}

		stmt := fmt.Sprintf(query, strings.Join(fieldNames, ","), placeHolders)
		b.Queue(stmt, database.GetScanFields(t, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, topic := range topics {
		now := time.Now()
		_ = topic.UpdatedAt.Set(now)
		_ = topic.CreatedAt.Set(now)
		topic.DeletedAt.Set(nil)
		/*Always set topic to draft when insert
		  Not allow to update status through upsert func*/
		topic.Status.Set(pb.TOPIC_STATUS_DRAFT.String())
		queueFn(b, topic)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(topics); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("student event log not inserted")
		}
	}

	return nil
}

func (r *TopicRepo) RetrieveBookTopic(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.BookTopic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.RetrieveBookTopic")
	defer span.End()
	stmt := `
	SELECT t.%s, book_id FROM topics t
	JOIN books_chapters USING (chapter_id)
	WHERE t.topic_id = ANY($1::_TEXT)
	  AND t.deleted_at IS NULL
      AND books_chapters.deleted_at IS NULL`

	topic := &entities.Topic{}
	fields := database.GetFieldNames(topic)
	selectStmt := fmt.Sprintf(stmt, strings.Join(fields, ", t."))

	rows, err := db.Query(ctx, selectStmt, &topicIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	bTopics := make([]*entities.BookTopic, 0)

	for rows.Next() {
		topicTemp := entities.Topic{}
		var (
			bookID pgtype.Text
		)
		scanFields := database.GetScanFields(&topicTemp, fields)
		scanFields = append(scanFields, &bookID)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		bTopic := entities.BookTopic{
			Topic:  topicTemp,
			BookID: bookID,
		}
		bTopics = append(bTopics, &bTopic)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}

	return bTopics, nil
}
