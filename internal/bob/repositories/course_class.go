package repositories

import (
	"context"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type CourseClassRepo struct{}

func (r *CourseClassRepo) Find(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (mapCourseIDsByClassID map[pgtype.Int4]pgtype.TextArray, err error) {
	query := `SELECT course_id, class_id
		FROM courses_classes
		WHERE class_id = ANY($1) AND status = $2`

	var pgStatus pgtype.Text
	_ = pgStatus.Set(entities_bob.CourseClassStatusActive)

	rows, err := db.Query(ctx, query, &ids, &pgStatus)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}

	defer rows.Close()

	result := make(map[pgtype.Int4]pgtype.TextArray)
	for rows.Next() {
		var (
			courseID pgtype.Text
			classID  pgtype.Int4
		)

		if err := rows.Scan(&courseID, &classID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		var (
			pgCourseIDs = result[classID]
			courseIDs   []string
		)

		_ = pgCourseIDs.AssignTo(&courseIDs)
		courseIDs = append(courseIDs, courseID.String)
		_ = pgCourseIDs.Set(courseIDs)

		result[classID] = pgCourseIDs
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return result, nil
}

func (r *CourseClassRepo) FindClassInCourse(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray, classIDs pgtype.Int4Array) (mapClassIDByCourseID map[pgtype.Text]pgtype.Int4Array, err error) {
	query := `SELECT course_id, class_id
	FROM courses_classes
	WHERE course_id = ANY($1) AND class_id = ANY($2)  AND status = $3`

	var pgStatus pgtype.Text
	_ = pgStatus.Set(entities_bob.CourseClassStatusActive)

	rows, err := db.Query(ctx, query, &courseIDs, &classIDs, &pgStatus)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}

	defer rows.Close()

	result := make(map[pgtype.Text]pgtype.Int4Array)
	for rows.Next() {
		var (
			courseID pgtype.Text
			classID  pgtype.Int4
		)

		if err := rows.Scan(&courseID, &classID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		var (
			pgClassIDs = result[courseID]
			classIDs   []int32
		)

		_ = pgClassIDs.AssignTo(&classIDs)
		classIDs = append(classIDs, classID.Int)
		_ = pgClassIDs.Set(classIDs)

		result[courseID] = pgClassIDs
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return result, nil
}

func (r *CourseClassRepo) FindByCourseIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (mapClassIDByCourseID map[pgtype.Text]pgtype.Int4Array, err error) {
	query := `SELECT course_id, ARRAY_AGG(class_id)
	FROM courses_classes
	WHERE course_id = ANY($1) AND status = $2
	GROUP BY course_id`

	var pgStatus pgtype.Text
	_ = pgStatus.Set(entities_bob.CourseClassStatusActive)

	rows, err := db.Query(ctx, query, &ids, &pgStatus)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}

	defer rows.Close()

	result := make(map[pgtype.Text]pgtype.Int4Array)
	for rows.Next() {
		var (
			courseID pgtype.Text
			classIDs pgtype.Int4Array
		)

		if err := rows.Scan(&courseID, &classIDs); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		result[courseID] = classIDs
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return result, nil
}
