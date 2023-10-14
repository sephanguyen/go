package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	repositories_bob "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type TopicRepo struct{}

func (r *TopicRepo) FindByIDs(ctx context.Context, db database.QueryExecer, topicIDs []string) (map[string]*entities_bob.Topic, error) {
	var topics entities_bob.Topics

	e := entities.Topic{}
	fieldNames := database.GetFieldNames(&e)
	stmt := "SELECT %s FROM %s WHERE topic_id = ANY($1) AND deleted_at IS NULL"
	query := fmt.Sprintf(stmt, strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, query, &topicIDs).ScanAll(&topics)
	if err != nil {
		return nil, err
	}
	topicsMap := make(map[string]*entities_bob.Topic)
	for _, topic := range topics {
		topicsMap[topic.ID.String] = topic

	}
	return topicsMap, nil
}

func (r *TopicRepo) FindSchoolIDs(ctx context.Context, db database.QueryExecer, topicIDs []string) ([]int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.FindSchoolIDs")
	defer span.End()

	query := "SELECT school_id FROM topics WHERE deleted_at IS NULL AND topic_id = ANY($1)"
	pgIDs := database.TextArray(topicIDs)

	schoolIDs := repositories_bob.EnSchoolIDs{}
	err := database.Select(ctx, db, query, &pgIDs).ScanAll(&schoolIDs)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	result := []int32{}
	for _, v := range schoolIDs {
		result = append(result, v.SchoolID)
	}

	return result, nil
}

func (r *TopicRepo) SoftDeleteByPresetStudyPlanWeeklyIDs(ctx context.Context, db database.Ext, pspwIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.SoftDeleteByPresetStudyPlanWeeklyIDs")
	defer span.End()

	query := fmt.Sprintf(`
		UPDATE %s
		SET deleted_at = NOW()
		WHERE topic_id IN (
			SELECT topic_id
			FROM %s
			WHERE preset_study_plan_weekly_id = ANY($1)
				AND deleted_at IS NULL
		)`,
		(&entities.Topic{}).TableName(),
		(&entities.PresetStudyPlanWeekly{}).TableName(),
	)

	cmdTag, err := db.Exec(ctx, query, pspwIDs)
	if err != nil {
		return fmt.Errorf("db.Query: %s", err)
	}
	if cmdTag.RowsAffected() != int64(len(pspwIDs.Elements)) {
		return fmt.Errorf("expect %d row deleted, got %d", len(pspwIDs.Elements), cmdTag.RowsAffected())
	}
	return nil
}

func (r *TopicRepo) Create(ctx context.Context, db database.Ext, plans []*entities_bob.Topic) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.Create")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *entities_bob.Topic) {
		fieldNames := database.GetFieldNames(e)
		placeHolders := database.GeneratePlaceholders(22)

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			e.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
	}

	b := &pgx.Batch{}
	var d pgtype.Timestamptz
	err := d.Set(time.Now())
	if err != nil {
		return fmt.Errorf("cannot set time topics: %w", err)
	}

	for _, each := range plans {
		if each.ID.String == "" {
			err = each.ID.Set(idutil.ULIDNow())
			if err != nil {
				return fmt.Errorf("cannot set id for topics: %w", err)
			}
		}
		each.CreatedAt = d
		each.UpdatedAt = d
		queueFn(b, each)
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

func (r *TopicRepo) FindByID(ctx context.Context, db database.QueryExecer, ID pgtype.Text) (*entities_bob.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.FindByID")
	defer span.End()
	e := &entities_bob.Topic{}

	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE topic_id = $1", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &ID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}

func (rcv *TopicRepo) Update(ctx context.Context, db database.QueryExecer, src *entities_bob.Topic) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Update")
	defer span.End()

	query := "UPDATE topics SET updated_at = now(), attachment_names = $1, attachment_urls = $2, subject = $3, grade = $4, country = $5, name = $6 WHERE topic_id = $7 AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &src.AttachmentNames, &src.AttachmentURLs, &src.Subject, &src.Grade, &src.Country, &src.Name, &src.ID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot update lesson")
	}

	return nil
}

func (r *TopicRepo) UpdateNameByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, newName pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.UpdateNameByLessonID")
	defer span.End()

	query := fmt.Sprintf(`
		UPDATE %s
		SET name = $2
		WHERE topic_id IN (
			SELECT topic_id
			FROM %s
			WHERE lesson_id = $1
				AND deleted_at IS NULL
		)
	`,
		(&entities.Topic{}).TableName(),
		(&entities.PresetStudyPlanWeekly{}).TableName(),
	)
	_, err := db.Exec(ctx, query, lessonID, newName)
	if err != nil {
		return fmt.Errorf("db.Exec: %s", err)
	}
	return nil
}

func (rcv *TopicRepo) SoftDeleteV2(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.SoftDeleteV2")
	defer span.End()

	query := "UPDATE topics SET deleted_at = now(), updated_at = now() WHERE topic_id = ANY($1) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &topicIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete topic")
	}

	return nil
}

func (r *TopicRepo) FindByIDsV2(ctx context.Context, db database.QueryExecer, IDs pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities_bob.Topic, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.FindByIDsV2")
	defer span.End()

	e := &entities_bob.Topic{}

	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE topic_id = ANY($1)", strings.Join(fields, ","), e.TableName())
	if !isAll {
		query += " AND deleted_at IS NULL"
	}
	result := map[pgtype.Text]*entities_bob.Topic{}
	rows, err := db.Query(ctx, query, &IDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities_bob.Topic)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result[c.ID] = c
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return result, nil
}

//SoftDeleteV3 not use ORM and return number of row effected
func (r *TopicRepo) SoftDeleteV3(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "TopicRepo.SoftDeleteV3")
	defer span.End()

	query := "UPDATE topics SET deleted_at = now(), updated_at = now() WHERE topic_id = ANY($1::_TEXT) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &topicIDs)
	if err != nil {
		return 0, err
	}
	return int(cmdTag.RowsAffected()), nil
}
