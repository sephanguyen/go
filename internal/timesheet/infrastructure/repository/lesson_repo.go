package repository

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
)

type LessonRepoImpl struct{}

func (r *LessonRepoImpl) FindLessonsByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepoImpl.Retrieve")
	defer span.End()

	listLesson := &entity.Lessons{}
	lesson := &entity.Lesson{}

	stmt := fmt.Sprintf(`
	SELECT lesson_id, scheduling_status
	FROM %s
	WHERE deleted_at IS NULL
	AND lesson_id = ANY($1::_TEXT);`, lesson.TableName())

	if err := database.Select(ctx, db, stmt, ids).ScanAll(listLesson); err != nil {
		return nil, err
	}

	return *listLesson, nil
}

func (r *LessonRepoImpl) FindAllLessonsByIDsIgnoreDeletedAtCondition(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepoImpl.Retrieve")
	defer span.End()

	listLesson := &entity.Lessons{}
	lesson := &entity.Lesson{}

	stmt := fmt.Sprintf(`
	SELECT lesson_id, scheduling_status
	FROM %s
	WHERE lesson_id = ANY($1::_TEXT);`, lesson.TableName())

	if err := database.Select(ctx, db, stmt, ids).ScanAll(listLesson); err != nil {
		return nil, err
	}

	return *listLesson, nil
}
