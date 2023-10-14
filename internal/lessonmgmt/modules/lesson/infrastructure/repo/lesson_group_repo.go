package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	media_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"

	"github.com/jackc/pgtype"
)

type LessonGroupRepo struct{}

func (l *LessonGroupRepo) Insert(ctx context.Context, db database.QueryExecer, e *LessonGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroupRepo.Insert")
	defer span.End()

	if err := e.PreInsert(); err != nil {
		return fmt.Errorf("could not pre-insert for new lesson_group %v", err)
	}

	fieldNames, value := e.FieldMap()
	query := fmt.Sprintf(`INSERT INTO lesson_groups (%s) VALUES ($1, $2, $3, $4, $5)`,
		strings.Join(fieldNames, ","))

	_, err := db.Exec(ctx, query, value...)
	return err
}

func (l *LessonGroupRepo) Upsert(ctx context.Context, db database.QueryExecer, e *LessonGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroupRepo.Upsert")
	defer span.End()

	if err := e.PreInsert(); err != nil {
		return fmt.Errorf("could not pre-insert for new lesson_group %v", err)
	}
	fieldNames, value := e.FieldMap()
	query := fmt.Sprintf(`INSERT INTO lesson_groups (%s) VALUES ($1, $2, $3, $4, $5) ON CONFLICT ON CONSTRAINT pk__lesson_groups DO 
						UPDATE SET media_ids = $3, updated_at = $5 `,
		strings.Join(fieldNames, ","))
	_, err := db.Exec(ctx, query, value...)
	return err
}

func (l *LessonGroupRepo) updateMedia(ctx context.Context, db database.QueryExecer, e *LessonGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroupRepo.updateMedia")
	defer span.End()

	if err := e.PreUpdate(); err != nil {
		return fmt.Errorf("could not pre-update for lesson_group %v", err)
	}

	_, err := database.UpdateFields(ctx, e, db.Exec, "lesson_group_id", []string{"media_ids", "updated_at"})
	if err != nil {
		return fmt.Errorf("database.UpdateField: %s", err)
	}

	return nil
}

func (l *LessonGroupRepo) getByIDAndCourseID(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID string) (*LessonGroup, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroup.getByIDAndCourseID")
	defer span.End()

	e := &LessonGroup{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_group_id = $1 AND course_id = $2", strings.Join(fields, ","), e.TableName())
	err := db.QueryRow(ctx, query, &lessonGroupID, &courseID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}

func (l *LessonGroupRepo) getByIDs(ctx context.Context, db database.QueryExecer, lessonGroupID []string) ([]*LessonGroup, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroup.getByIDs")
	defer span.End()
	lg := &LessonGroup{}
	fields, _ := lg.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_group_id = ANY($1)", strings.Join(fields, ","), lg.TableName())
	rows, err := db.Query(ctx, query, lessonGroupID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	res := []*LessonGroup{}
	for rows.Next() {
		lessonGroup := &LessonGroup{}
		if err = rows.Scan(database.GetScanFields(lessonGroup, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		res = append(res, lessonGroup)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return res, nil
}

// GetMedias will returns list of media id of an lesson group
func (l *LessonGroupRepo) ListMediaByLessonArgs(ctx context.Context, db database.QueryExecer, args *domain.ListMediaByLessonArgs) (media_domain.Medias, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroupRepo.ListMediaByLessonArgs")
	defer span.End()
	var (
		lessonGroupID pgtype.Text
		courseID      pgtype.Text
	)

	query := fmt.Sprintf(`
		SELECT lesson_group_id, course_id FROM lessons WHERE lesson_id = $1 AND deleted_at IS NULL `,
	)
	if err := db.QueryRow(ctx, query, &args.LessonID).Scan(&lessonGroupID, &courseID); err != nil {
		return nil, err
	}

	medias := media_infrastructure.Medias{}
	lessonGroupEnt := &LessonGroup{}
	mediaEnt := &media_infrastructure.Media{}
	mediaFields, _ := mediaEnt.FieldMap()
	mediaTableName := mediaEnt.TableName()
	for i := range mediaFields {
		mediaFields[i] = mediaTableName + "." + mediaFields[i]
	}
	query = fmt.Sprintf(`
	SELECT %s
	FROM (
		SELECT UNNEST(media_ids) AS media_id
		FROM %s
		WHERE lesson_group_id=$1 AND course_id=$2
	) AS lg LEFT JOIN %s ON lg.media_id = %s.media_id
	WHERE deleted_at IS NULL AND ($3::TEXT = '' OR media.media_id<$3)
	ORDER BY media.media_id DESC
	LIMIT $4;`, strings.Join(mediaFields, ","), lessonGroupEnt.TableName(), mediaEnt.TableName(), mediaEnt.TableName())
	err := database.Select(ctx, db, query, lessonGroupID, courseID, args.Offset, args.Limit).ScanAll(&medias)
	if err != nil {
		return nil, err
	}
	return medias.ToMediasEntity(), nil
}
