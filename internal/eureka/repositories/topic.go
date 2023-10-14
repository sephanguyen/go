package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

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
	topics := entities.Topics{}

	err := database.Select(ctx, db, fmt.Sprintf("SELECT %s FROM topics WHERE topic_id = ANY($1::_TEXT) AND deleted_at IS NULL ORDER BY display_order ASC", strings.Join(fields, ",")), &ids).ScanAll(&topics)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return topics, nil
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
		WHERE ($1::TEXT[] IS NULL OR bc.book_id = ANY($1::TEXT[])) AND
			($2::TEXT[] IS NULL OR t.topic_id = ANY($2::TEXT[])) AND 
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

	topics := entities.Topics{}
	err := database.Select(ctx, db, query, args...).ScanAll(&topics)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return topics, nil
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

func (r *TopicRepo) BulkImport(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.BulkImport")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities.Topic) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := fmt.Sprintf(`INSERT INTO topics (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT topics_pk DO UPDATE SET
			name = $2,
			country = $3,
			grade = $4,
			subject = $5,
			topic_type = $6,
			updated_at = $7,
			display_order = $11,
			icon_url = $12`, strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, topic := range topics {
		now := time.Now()
		if err := multierr.Combine(
			topic.UpdatedAt.Set(now),
			topic.CreatedAt.Set(now),
			topic.DeletedAt.Set(nil),
			// Always set topic to draft when insert
			// Not allow to update status through upsert func
			topic.Status.Set(cpb.TopicStatus_TOPIC_STATUS_DRAFT.String()),
		); err != nil {
			return fmt.Errorf("failed to set timestamp: %w", err)
		}
		queue(b, topic)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(topics); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("topic not inserted")
		}
	}

	return nil
}

func (r *TopicRepo) BulkUpsertWithoutDisplayOrder(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.BulkUpsertWithoutDisplayOrder")
	defer span.End()

	queueFn := func(b *pgx.Batch, t *entities.Topic) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := fmt.Sprintf(`INSERT INTO topics (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT topics_pk DO UPDATE SET
			name = $2,
			country = $3,
			grade = $4,
			subject = $5,
			topic_type = $6,
			updated_at = $7,
			icon_url = $12`, strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, topic := range topics {
		now := time.Now()
		if err := multierr.Combine(
			topic.UpdatedAt.Set(now),
			topic.CreatedAt.Set(now),
			topic.DeletedAt.Set(nil),
			// Always set topic to draft when insert
			// Not allow to update status through upsert func
			topic.Status.Set(cpb.TopicStatus_TOPIC_STATUS_DRAFT.String()),
		); err != nil {
			return fmt.Errorf("failed to set timestamp: %w", err)
		}
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

func (r *TopicRepo) UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.UpdateTotalLOs")
	defer span.End()

	query := `UPDATE topics SET total_los = sub.total_los, updated_at = $2
	FROM (
		SELECT COUNT(*) AS total_los
		FROM learning_objectives
		WHERE topic_id = $1 AND learning_objectives.deleted_at IS NULL
	) sub
	WHERE topic_id = $1 AND topics.deleted_at IS NULL`

	var now pgtype.Timestamptz
	if err := now.Set(time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to set timestamp: %w", err)
	}

	cmdTag, err := db.Exec(ctx, query, &topicID, &now)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot update total_los for topic: %s", topicID.String)
	}
	return nil
}

func (r *TopicRepo) UpdateStatus(ctx context.Context, db database.Ext, ids pgtype.TextArray, topicStatus pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.UpdateStatus")
	defer span.End()

	query := `UPDATE topics SET status = $1, published_at = $2, updated_at = $3
	WHERE topic_id = ANY($4) AND deleted_at IS NULL`
	var now pgtype.Timestamptz
	if err := now.Set(time.Now()); err != nil {
		return fmt.Errorf("failed to set timestamp: %w", err)
	}

	cmdTag, err := db.Exec(ctx, query, &topicStatus, &now, &now, &ids)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != int64(len(ids.Elements)) {
		return errors.New("cannot update all topic")
	}
	return nil
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

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("row.Err: %w", err)
		}
	}
	return copiedTopics, nil
}

func (r *TopicRepo) FindByIDsV2(ctx context.Context, db database.QueryExecer, ids []string, isAll bool) (map[string]*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.FindByIDsV2")
	defer span.End()

	e := &entities.Topic{}

	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE topic_id = ANY($1)", strings.Join(fields, ","), e.TableName())
	if !isAll {
		query += " AND deleted_at IS NULL"
	}
	result := map[string]*entities.Topic{}
	rows, err := db.Query(ctx, query, database.TextArray(ids))
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities.Topic)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result[c.ID.String] = c
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return result, nil
}

func (r *TopicRepo) SoftDelete(ctx context.Context, db database.QueryExecer, topicIDs []string) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.SoftDelete")
	defer span.End()

	query := "UPDATE topics SET deleted_at = now(), updated_at = now() WHERE topic_id = ANY($1::_TEXT) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, database.TextArray(topicIDs))
	if err != nil {
		return 0, err
	}
	return int(cmdTag.RowsAffected()), nil
}

func (r *TopicRepo) FindByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray) ([]*entities.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.FindByBookIDs")
	defer span.End()

	e := &entities.Topic{}
	fields := database.GetFieldNames(&entities.Topic{})
	query := `SELECT t.%s 
		FROM %s AS t 
		WHERE ($1::TEXT[] IS NULL OR t.chapter_id = ANY($1::TEXT[])) 
		AND t.deleted_at IS NULL
		ORDER BY display_order ASC`

	query = fmt.Sprintf(query, strings.Join(fields, ", t."), e.TableName())

	topics := entities.Topics{}
	err := database.Select(ctx, db, query, &chapterIDs).ScanAll(&topics)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return topics, nil
}
