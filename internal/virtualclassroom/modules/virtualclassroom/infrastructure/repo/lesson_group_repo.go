package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type LessonGroupRepo struct{}

func (l *LessonGroupRepo) Insert(ctx context.Context, db database.QueryExecer, e *LessonGroupDTO) error {
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

func (l *LessonGroupRepo) Upsert(ctx context.Context, db database.QueryExecer, e *LessonGroupDTO) error {
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

func (l *LessonGroupRepo) GetByIDAndCourseID(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID string) (*LessonGroupDTO, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonGroup.GetByIDAndCourseID")
	defer span.End()

	e := &LessonGroupDTO{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_group_id = $1 AND course_id = $2", strings.Join(fields, ","), e.TableName())
	err := db.QueryRow(ctx, query, &lessonGroupID, &courseID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}
