package repositories

import (
	"context"
	"fmt"
	"strings"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LessonGroupRepo struct{}

func (r *LessonGroupRepo) Create(ctx context.Context, db database.QueryExecer, e *entities_bob.LessonGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroupRepo.Create")
	defer span.End()

	if err := e.PreInsert(); err != nil {
		return fmt.Errorf("could not pre-insert for new lesson_group %v", err)
	}

	fieldNames, value := e.FieldMap()
	const placeHolders = "$1, $2, $3, $4, $5"

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT (lesson_group_id, course_id)
			DO UPDATE SET updated_at = NOW()`,
		e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	cmdTag, err := db.Exec(ctx, query, value...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new LessonGroup")
	}

	return nil
}

func (r *LessonGroupRepo) Get(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text) (*entities_bob.LessonGroup, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroup.Get")
	defer span.End()

	e := &entities_bob.LessonGroup{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_group_id = $1 AND course_id = $2", strings.Join(fields, ","), e.TableName())
	err := db.QueryRow(ctx, query, &lessonGroupID, &courseID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}

func (r *LessonGroupRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities_bob.LessonGroup) error {
	b := &pgx.Batch{}
	e := &entities_bob.LessonGroup{}
	currentTime := timeutil.Now().UTC()

	for _, item := range items {
		fieldNames, value := item.FieldMap()
		const placeHolders = "$1, $2, $3, $4, $5"

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT (lesson_group_id, course_id)
			DO UPDATE SET media_ids = $3, updated_at = NOW()`,
			e.TableName(), strings.Join(fieldNames, ","), placeHolders)

		if item.CreatedAt.Status != pgtype.Present && item.UpdatedAt.Status != pgtype.Present {
			b.Queue(query, append(value[:3], currentTime, currentTime)...)
		} else {
			b.Queue(query, value...)
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

func (r *LessonGroupRepo) UpdateMedias(ctx context.Context, db database.QueryExecer, e *entities_bob.LessonGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroupRepo.UpdateMedias")
	defer span.End()

	err := e.PreUpdate()
	if err != nil {
		return fmt.Errorf("e.PreUpdate: %s", err)
	}

	cmd, err := database.UpdateFields(ctx, e, db.Exec, "lesson_group_id", []string{"media_ids", "updated_at"})
	if err != nil {
		return fmt.Errorf("database.UpdateField: %s", err)
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("expect 1 rows affected, got %d", cmd.RowsAffected())
	}
	return nil
}

// GetMedias will returns list of media id of an lesson group
func (r *LessonGroupRepo) GetMedias(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text, limit pgtype.Int4, offset pgtype.Text) (entities_bob.Medias, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroup.GetMedias")
	defer span.End()

	medias := entities_bob.Medias{}
	lessonGroupEnt := &entities_bob.LessonGroup{}
	mediaEnt := &entities_bob.Media{}
	mediaFields, _ := mediaEnt.FieldMap()
	mediaTableName := mediaEnt.TableName()
	for i := range mediaFields {
		mediaFields[i] = mediaTableName + "." + mediaFields[i]
	}
	query := fmt.Sprintf(`
	SELECT %s
	FROM (
		SELECT UNNEST(media_ids) AS media_id
		FROM %s
		WHERE lesson_group_id=$1 AND course_id=$2
	) AS lg LEFT JOIN %s ON lg.media_id = %s.media_id
	WHERE deleted_at IS NULL AND ($3::TEXT IS NULL OR media.media_id<$3)
	ORDER BY media.media_id DESC
	LIMIT $4;`, strings.Join(mediaFields, ","), lessonGroupEnt.TableName(), mediaEnt.TableName(), mediaEnt.TableName())

	err := database.Select(ctx, db, query, lessonGroupID, courseID, offset, limit.Get()).ScanAll(&medias)
	if err != nil {
		return nil, err
	}

	return medias, nil
}
