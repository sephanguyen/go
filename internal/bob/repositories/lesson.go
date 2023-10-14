package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LessonRepo struct{}

type LessonFilter struct {
	LessonID  pgtype.TextArray
	TeacherID pgtype.TextArray
	CourseID  pgtype.TextArray
}
type LessonJoinedV2Filter struct {
	UserID               pgtype.Text
	CourseIDs            pgtype.TextArray
	BlacklistedCourseIDs pgtype.TextArray
	StartDate            *pgtype.Timestamptz
	EndDate              *pgtype.Timestamptz
}
type LessonWithTime struct {
	Lesson                entities.Lesson
	PresetStudyPlanWeekly entities.PresetStudyPlanWeekly
}

const DistinctKeyword string = "distinct"

// Deprecated:
// Find will return lesson have lesson_type = 'LESSON_TYPE_ONLINE'
func (l *LessonRepo) Find(ctx context.Context, db database.QueryExecer, filter *LessonFilter) ([]*entities.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Find")
	defer span.End()

	lesson := &entities.Lesson{}
	fields := database.GetFieldNames(lesson)
	query := fmt.Sprintf(`SELECT %s FROM %s
		WHERE ($1::text[] IS NULL OR lesson_id = ANY($1))
			AND ($2::text[] IS NULL OR teacher_id = ANY($2))
			AND ($3::text[] IS NULL OR course_id = ANY($3))
			AND lesson_type = 'LESSON_TYPE_ONLINE'
			AND deleted_at IS NULL`, strings.Join(fields, ","), lesson.TableName())

	rows, err := db.Query(ctx, query, &filter.LessonID, &filter.TeacherID, &filter.CourseID)
	if err != nil {
		return nil, fmt.Errorf("db.Query :%v", err)
	}
	defer rows.Close()

	var lessons []*entities.Lesson
	for rows.Next() {
		e := &entities.Lesson{}
		_, values := e.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan :%v", err)
		}
		lessons = append(lessons, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err :%v", err)
	}

	return lessons, nil
}

// FindByID is copied from yasuo.repositories.Lesson.FindByID.
func (l *LessonRepo) FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindByID")
	defer span.End()

	lesson := &entities.Lesson{}
	fields, values := lesson.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM lessons
		WHERE lesson_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return lesson, nil
}

func (l *LessonRepo) FindLessonWithTime(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*LessonWithTime, pgtype.Int8, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindLessonWithTime")
	defer span.End()

	args := []interface{}{courseIDs, startDate, endDate, schedulingStatus}
	lesson := &entities.Lesson{}
	lessonFields := database.GetFieldNames(lesson)

	query := fmt.Sprintf(`SELECT DISTINCT ls.%s, COUNT(*) OVER() AS total 
		FROM lessons ls
		LEFT JOIN lessons_courses lc ON lc.lesson_id = ls.lesson_id and lc.deleted_at IS NULL
		WHERE (ls.course_id = ANY($1) OR lc.course_id = ANY($1))
			AND (($2::timestamptz IS NULL OR $3::timestamptz IS NULL)
				OR (start_time <= $3 AND end_time >= $2))
			AND lesson_type = 'LESSON_TYPE_ONLINE'
			AND ls.deleted_at IS NULL
			AND ($4::text IS NULL OR ls.scheduling_status = $4)
		ORDER BY start_time, end_time ASC`,
		strings.ReplaceAll(strings.Join(lessonFields, ", ls."), "ls.course_id", "CASE WHEN ls.course_id = ANY($1) THEN ls.course_id ELSE lc.course_id END as course_id"))

	query, args = database.AddPagingQuery(query, limit, page, args...)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, pgtype.Int8{}, err
	}

	defer rows.Close()

	var lessons []*LessonWithTime
	var total pgtype.Int8
	for rows.Next() {
		e := &LessonWithTime{}
		lessonValue := database.GetScanFields(&e.Lesson, lessonFields)
		lessonValue = append(lessonValue, &total)
		if err := rows.Scan(lessonValue...); err != nil {
			return nil, total, fmt.Errorf("rows.Scan :%v", err)
		}
		lessons = append(lessons, e)
	}

	if err := rows.Err(); err != nil {
		return nil, total, fmt.Errorf("db.Query :%v", err)
	}

	return lessons, total, nil
}

func (l *LessonRepo) FindLessonWithTimeAndLocations(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, schedulingStatus pgtype.Text) ([]*LessonWithTime, pgtype.Int8, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindLessonWithTimeAndLocations")
	defer span.End()

	args := []interface{}{courseIDs, startDate, endDate, locationIDs, schedulingStatus}
	lesson := &entities.Lesson{}
	lessonFields := database.GetFieldNames(lesson)
	query := fmt.Sprintf(`SELECT DISTINCT ls.%s, COUNT(*) OVER() AS total 
		FROM lessons ls
		LEFT JOIN lessons_courses lc ON lc.lesson_id = ls.lesson_id and lc.deleted_at IS NULL 
		WHERE (ls.course_id = ANY($1) OR lc.course_id = ANY($1))
			AND (($2::timestamptz IS NULL OR $3::timestamptz IS NULL)
				OR (start_time <= $3 AND end_time >= $2))
			AND lesson_type = 'LESSON_TYPE_ONLINE' 
			AND ls.deleted_at IS NULL 
			AND ls.center_id = ANY($4)
			AND ($5::text IS NULL OR ls.scheduling_status = $5)
		ORDER BY start_time, end_time ASC`,
		strings.ReplaceAll(strings.Join(lessonFields, ", ls."), "ls.course_id", "CASE WHEN ls.course_id = ANY($1) THEN ls.course_id ELSE lc.course_id END as course_id"))

	query, args = database.AddPagingQuery(query, limit, page, args...)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, pgtype.Int8{}, err
	}

	defer rows.Close()

	var lessons []*LessonWithTime
	var total pgtype.Int8
	for rows.Next() {
		e := &LessonWithTime{}
		lessonValue := database.GetScanFields(&e.Lesson, lessonFields)
		lessonValue = append(lessonValue, &total)
		if err := rows.Scan(lessonValue...); err != nil {
			return nil, total, fmt.Errorf("rows.Scan :%v", err)
		}
		lessons = append(lessons, e)
	}

	if err := rows.Err(); err != nil {
		return nil, total, fmt.Errorf("db.Query :%v", err)
	}

	return lessons, total, nil
}

func (l *LessonRepo) FindLessonJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*LessonWithTime, pgtype.Int8, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindLessonJoined")
	defer span.End()

	args := []interface{}{&userID, courseIDs, startDate, endDate, schedulingStatus}
	lesson := &entities.Lesson{}
	lessonFields := database.GetFieldNames(lesson)

	query := fmt.Sprintf(`
		SELECT DISTINCT ls.%s, COUNT(*) OVER() AS total
		FROM lessons ls
		JOIN lesson_members lm ON ls.lesson_id = lm.lesson_id
		LEFT JOIN lessons_courses lc ON lc.lesson_id = ls.lesson_id and lc.deleted_at IS NULL
		WHERE (ls.course_id = ANY($2) OR lc.course_id = ANY($2))
		  AND lm.user_id = $1 AND lm.deleted_at IS NULL
		  AND (($3::timestamptz IS NULL
				OR $4::timestamptz IS NULL)
			   OR (start_time <= $4
				   AND end_time >= $3))
		  AND lesson_type = 'LESSON_TYPE_ONLINE'
		  AND ls.deleted_at IS NULL
		  AND ($5::text IS NULL OR ls.scheduling_status = $5)
		ORDER BY start_time, end_time ASC`,
		strings.Join(lessonFields, ", ls."))

	query, args = database.AddPagingQuery(query, limit, page, args...)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, pgtype.Int8{}, err
	}

	defer rows.Close()

	var lessons []*LessonWithTime
	var total pgtype.Int8
	for rows.Next() {
		e := &LessonWithTime{}
		lessonValue := database.GetScanFields(&e.Lesson, lessonFields)
		lessonValue = append(lessonValue, &total)
		if err := rows.Scan(lessonValue...); err != nil {
			return nil, total, fmt.Errorf("rows.Scan :%v", err)
		}
		lessons = append(lessons, e)
	}

	if err := rows.Err(); err != nil {
		return nil, total, fmt.Errorf("db.Query :%v", err)
	}
	return lessons, total, nil
}

func (l *LessonRepo) FindLessonJoinedV2(ctx context.Context, db database.QueryExecer, filter *LessonJoinedV2Filter, limit int32, page int32) ([]*LessonWithTime, pgtype.Int8, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindLessonJoinedV2")
	defer span.End()

	args := []interface{}{&filter.UserID, &filter.CourseIDs, &filter.BlacklistedCourseIDs, filter.StartDate, filter.EndDate}
	lesson := &entities.Lesson{}
	lessonFields := database.GetFieldNames(lesson)

	query := fmt.Sprintf(`
		SELECT DISTINCT ls.%s, COUNT(*) OVER() AS total
		FROM lessons ls
		JOIN lesson_members lm ON ls.lesson_id = lm.lesson_id
		LEFT JOIN lessons_courses lc ON lc.lesson_id = ls.lesson_id and lc.deleted_at IS NULL
		WHERE (ls.course_id = ANY($2) OR lc.course_id = ANY($2))
		  AND NOT (ls.course_id = ANY($3))
		  AND lm.user_id = $1 AND lm.deleted_at IS NULL
		  AND (($4::timestamptz IS NULL
				OR $5::timestamptz IS NULL)
			   OR (start_time <= $5
				   AND end_time >= $4))
		  AND lesson_type = 'LESSON_TYPE_ONLINE'
		  AND ls.deleted_at IS NULL
		ORDER BY start_time, end_time ASC`,
		strings.Join(lessonFields, ", ls."))

	query, args = database.AddPagingQuery(query, limit, page, args...)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, pgtype.Int8{}, err
	}

	defer rows.Close()

	var lessons []*LessonWithTime
	var total pgtype.Int8
	for rows.Next() {
		e := &LessonWithTime{}
		lessonValue := database.GetScanFields(&e.Lesson, lessonFields)
		lessonValue = append(lessonValue, &total)
		if err := rows.Scan(lessonValue...); err != nil {
			return nil, total, fmt.Errorf("rows.Scan :%v", err)
		}
		lessons = append(lessons, e)
	}

	if err := rows.Err(); err != nil {
		return nil, total, fmt.Errorf("db.Query :%v", err)
	}
	return lessons, total, nil
}
func (l *LessonRepo) FindLessonJoinedWithLocations(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, schedulingStatus pgtype.Text) ([]*LessonWithTime, pgtype.Int8, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindLessonJoinedWithLocations")
	defer span.End()

	args := []interface{}{&userID, courseIDs, startDate, endDate, locationIDs, schedulingStatus}
	lesson := &entities.Lesson{}
	lessonFields := database.GetFieldNames(lesson)

	query := fmt.Sprintf(`
		SELECT DISTINCT ls.%s, COUNT(*) OVER() AS total
		FROM lessons ls
		JOIN lesson_members lm ON ls.lesson_id = lm.lesson_id
		LEFT JOIN lessons_courses lc ON lc.lesson_id = ls.lesson_id and lc.deleted_at IS NULL 
		WHERE (ls.course_id = ANY($2) OR lc.course_id = ANY($2))
		  AND lm.user_id = $1 AND lm.deleted_at IS NULL
		  AND (($3::timestamptz IS NULL
				OR $4::timestamptz IS NULL)
			   OR (start_time <= $4
				   AND end_time >= $3))
		  AND lesson_type = 'LESSON_TYPE_ONLINE'
		  AND ls.deleted_at IS NULL 
		  AND ls.center_id = ANY($5) 
		  AND ($6::text IS NULL OR ls.scheduling_status = $6)
		ORDER BY start_time, end_time ASC`,
		strings.Join(lessonFields, ", ls."))

	query, args = database.AddPagingQuery(query, limit, page, args...)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, pgtype.Int8{}, err
	}

	defer rows.Close()

	var lessons []*LessonWithTime
	var total pgtype.Int8
	for rows.Next() {
		e := &LessonWithTime{}
		lessonValue := database.GetScanFields(&e.Lesson, lessonFields)
		lessonValue = append(lessonValue, &total)
		if err := rows.Scan(lessonValue...); err != nil {
			return nil, total, fmt.Errorf("rows.Scan :%v", err)
		}
		lessons = append(lessons, e)
	}

	if err := rows.Err(); err != nil {
		return nil, total, fmt.Errorf("db.Query :%v", err)
	}

	return lessons, total, nil
}

type ListLessonArgs struct {
	Limit            uint32
	LessonID         pgtype.Text
	SchoolID         pgtype.Int4
	Courses          pgtype.TextArray
	StartTime        pgtype.Timestamptz
	EndTime          pgtype.Timestamptz
	StatusNotStarted pgtype.Text
	StatusInProcess  pgtype.Text
	StatusCompleted  pgtype.Text
	KeyWord          pgtype.Text
}

const retrieveLessonFilterWithCourseQuery = `
WITH filter_lesson AS (
	SELECT distinct l.lesson_id, l."name", l.start_time , l.end_time, l.created_at, l.lesson_type, l.class_id 
	FROM lessons l 
		LEFT JOIN lessons_courses lc ON l.lesson_id = lc.lesson_id
		JOIN courses c ON COALESCE(lc.course_id, l.course_id) = c.course_id
	WHERE l.deleted_at IS NULL AND lc.deleted_at IS NULL AND c.deleted_at IS NULL
		AND ($9::int IS NULL OR c.school_id = $9)
		AND ($10::text[] IS NULL OR c.course_id = ANY($10))
		AND ($3::timestamptz IS NULL OR l.start_time >= $3::timestamptz)
		AND ($4::timestamptz IS NULL OR l.start_time <= $4::timestamptz)
		AND ( ($5::text IS NULL AND $6::text IS NULL AND $7::text IS NULL)
			 OR ($5::text IS NOT NULL AND l.end_time < NOW())
			 OR ($6::text IS NOT NULL AND NOW() BETWEEN l.start_time AND l.end_time)
			 OR ($7::text IS NOT NULL AND l.start_time > NOW())
			)
		AND ($8::text IS null OR (nospace(l."name") ILIKE nospace(CONCAT('%',$8,'%'))))
	)
`

const retrieveLessonFilterWithOutCourseQuery = `
WITH filter_lesson AS (
	SELECT l.lesson_id, l."name", l.start_time , l.end_time, l.created_at, l.lesson_type, l.class_id 
	FROM lessons l 
		JOIN courses c ON l.course_id = c.course_id
	WHERE l.deleted_at IS NULL AND c.deleted_at IS NULL
		AND ($9::int IS NULL OR c.school_id = $9)
		AND ($3::timestamptz IS NULL OR l.start_time >= $3::timestamptz)
		AND ($4::timestamptz IS NULL OR l.start_time <= $4::timestamptz)
		AND ( ($5::text IS NULL AND $6::text IS NULL AND $7::text IS NULL)
			 OR ($5::text IS NOT NULL AND l.end_time < NOW())
			 OR ($6::text IS NOT NULL AND NOW() BETWEEN l.start_time AND l.end_time)
			 OR ($7::text IS NOT NULL AND l.start_time > NOW())
			)
		AND ($8::text IS null OR (nospace(l."name") ILIKE nospace(CONCAT('%',$8,'%'))))
	)
`

func (l *LessonRepo) Retrieve(ctx context.Context, db database.QueryExecer, args *ListLessonArgs) ([]*entities.Lesson, uint32, string, uint32, error) {
	retrieveLessonQuery := `
, count_result as (
	select count(*) as total from filter_lesson
)
, previous_sort as(
	select fl.lesson_id, fl.created_at,
			COUNT(*) OVER() AS total
	from filter_lesson fl
	where $1::text is not NULL
			and (fl.created_at, fl.lesson_id) > ((SELECT created_at FROM lessons WHERE lesson_id = $1 LIMIT 1), $1)
	order by fl.created_at ASC, fl.lesson_id ASC
	LIMIT $2
)
, previous as (
select ps.lesson_id, ps.total
from previous_sort ps
order by ps.created_at desc 
limit 1
)
select fl.lesson_id, fl.created_at, fl."name", fl.start_time, fl.end_time, fl.lesson_type, fl.class_id, 
		p.lesson_id AS pre_offset, p.total AS pre_total, count_result.total
from filter_lesson fl
	left join count_result on TRUE
	left join previous p on TRUE
where $1::text IS NULL
			OR (fl.created_at, fl.lesson_id) < ((SELECT created_at FROM lessons WHERE lesson_id = $1 LIMIT 1), $1)
order by fl.created_at DESC, fl.lesson_id DESC
LIMIT $2
`
	lessons := entities.Lessons{}
	fields := []string{"lesson_id", "created_at", "name", "start_time", "end_time", "lesson_type", "class_id"}
	var query string
	var rows pgx.Rows
	var err error
	if args.Courses.Status != pgtype.Present || len(args.Courses.Elements) == 0 {
		query = retrieveLessonFilterWithOutCourseQuery + retrieveLessonQuery
		rows, err = db.Query(ctx, query, &args.LessonID, &args.Limit, args.StartTime, args.EndTime,
			args.StatusCompleted, args.StatusInProcess, args.StatusNotStarted, args.KeyWord, &args.SchoolID)
	} else {
		query = retrieveLessonFilterWithCourseQuery + retrieveLessonQuery
		rows, err = db.Query(ctx, query, &args.LessonID, &args.Limit, args.StartTime, args.EndTime,
			args.StatusCompleted, args.StatusInProcess, args.StatusNotStarted, args.KeyWord, &args.SchoolID,
			args.Courses)
	}

	if err != nil {
		return nil, 0, "", 0, err
	}
	defer rows.Close()

	var total pgtype.Int8
	var pre_offset pgtype.Text
	var pre_total pgtype.Int8
	for rows.Next() {
		lesson := lessons.Add()
		scanFields := append(database.GetScanFields(lesson, fields), &pre_offset, &pre_total, &total)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, 0, "", 0, errors.Wrap(err, "rows.Scan")
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "rows.Err")
	}

	return lessons, uint32(total.Int), pre_offset.String, uint32(pre_total.Int), nil
}

func (l *LessonRepo) FindPreviousPageOffset(ctx context.Context, db database.QueryExecer, args *ListLessonArgs) (string, error) {
	retrievePreviousLessonQuery := `
	select ft.lesson_id, ft.created_at,
			COUNT(*) OVER() AS total
	from filter_time ft
	where $1::text IS NULL
			OR (ft.created_at, ft.lesson_id) > ((SELECT created_at FROM lessons WHERE lesson_id = $1 LIMIT 1), $1)
	order by ft.created_at ASC, ft.lesson_id ASC
	LIMIT $2`
	lessons := entities.Lessons{}
	fields := []string{"lesson_id", "created_at"}
	query := retrieveLessonFilterWithCourseQuery + retrievePreviousLessonQuery
	rows, err := db.Query(ctx, query, &args.LessonID, &args.Limit, &args.SchoolID,
		args.Courses, args.StartTime, args.EndTime,
		args.StatusCompleted, args.StatusInProcess, args.StatusNotStarted,
		args.KeyWord)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var total pgtype.Int8
	for rows.Next() {
		lesson := lessons.Add()
		scanFields := append(database.GetScanFields(lesson, fields), &total)
		if err := rows.Scan(scanFields...); err != nil {
			return "", errors.Wrap(err, "rows.Scan")
		}
	}
	if err := rows.Err(); err != nil {
		return "", errors.Wrap(err, "rows.Err")
	}

	if uint32(len(lessons)) < args.Limit {
		return "", nil
	}

	return lessons[len(lessons)-1].LessonID.String, nil
}

func (l *LessonRepo) CountLesson(ctx context.Context, db database.QueryExecer, args *ListLessonArgs) (int64, error) {
	lesson := &entities.Lesson{}
	fields := []string{"lesson_id"}
	query := fmt.Sprintf(`
		WITH ft as (
			SELECT l.%s
			FROM %s l 
				LEFT JOIN lessons_courses lc ON (l.lesson_id =lc.lesson_id AND lc.deleted_at IS NULL)
				LEFT JOIN courses c ON COALESCE(lc.course_id, l.course_id) = c.course_id 
			WHERE ($1::int IS NULL OR c.school_id = $1) AND l.deleted_at IS NULL)
		, ls as (SELECT DISTINCT(lesson_id) FROM ft)
		SELECT COUNT(lesson_id)
		FROM ls
	`, strings.Join(fields, ", "), lesson.TableName())
	var count pgtype.Int8
	if err := database.Select(ctx, db, query, &args.SchoolID).ScanFields(&count); err != nil {
		return 0, err
	}

	return count.Int, nil
}

func (l *LessonRepo) GetTeacherIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error) {
	var teacherIDs []string

	query := `SELECT teacher_id 
			FROM lessons_teachers
			WHERE lesson_id = $1 AND deleted_at IS NULL`

	rows, err := db.Query(ctx, query, &lessonID)
	if err != nil {
		return pgtype.TextArray{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return pgtype.TextArray{}, fmt.Errorf("rows.Scan :%v", err)
		}
		teacherIDs = append(teacherIDs, id.String)
	}
	return database.TextArray(teacherIDs), nil
}

func (l *LessonRepo) GetTeachersOfLessons(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) ([]*entities.LessonsTeachers, error) {
	lessonTeachers := &entities.LessonsTeachers{}
	fields := database.GetFieldNames(lessonTeachers)
	query := `SELECT teacher_id, lesson_id
	FROM lessons_teachers
	WHERE lesson_id = ANY($1) AND deleted_at IS NULL`
	rows, err := db.Query(ctx, query, &lessonIDs)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var teachers []*entities.LessonsTeachers
	for rows.Next() {
		teacher := &entities.LessonsTeachers{}
		if err := rows.Scan(database.GetScanFields(teacher, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		teachers = append(teachers, teacher)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return teachers, nil
}

func (l *LessonRepo) GetCourseIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetCourseIDsOfLesson")
	defer span.End()

	query := `
		SELECT course_id 
		FROM lessons_courses
		WHERE lesson_id = $1 
			AND deleted_at IS NULL`

	rows, err := db.Query(ctx, query, &lessonID)
	if err != nil {
		return pgtype.TextArray{}, fmt.Errorf("db.Query: %s", err)
	}
	defer rows.Close()

	var courseIDs []string
	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return pgtype.TextArray{}, fmt.Errorf("rows.Scan :%v", err)
		}
		courseIDs = append(courseIDs, id.String)
	}
	return database.TextArray(courseIDs), nil
}

func (l *LessonRepo) GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error) {
	var learnerIDs []string

	query := `SELECT user_id 
			FROM lesson_members
			WHERE lesson_id = $1 AND deleted_at IS NULL`

	rows, err := db.Query(ctx, query, &lessonID)
	if err != nil {
		return pgtype.TextArray{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return pgtype.TextArray{}, fmt.Errorf("rows.Scan :%v", err)
		}
		learnerIDs = append(learnerIDs, id.String)
	}
	return database.TextArray(learnerIDs), nil
}

func (l *LessonRepo) EndLiveLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, endTime pgtype.Timestamptz) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.EndLiveLesson")
	defer span.End()

	cmdTag, err := db.Exec(ctx, `UPDATE lessons SET end_at =$1 WHERE lesson_id =$2`, &endTime, &lessonID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("EndLiveLesson: Can't update lesson end time")
	}
	return nil
}

const stmtGetStreamingLearnersTpl = `SELECT %s FROM %s WHERE lesson_id = $1`

func (l *LessonRepo) GetStreamingLearners(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, queryEnhancers ...QueryEnhancer) ([]string, error) {
	query := stmtGetStreamingLearnersTpl
	e := &entities.Lesson{}
	fields, _ := e.FieldMap()
	for _, ehc := range queryEnhancers {
		ehc(&query)
	}
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ","), e.TableName()), lessonID).ScanOne(e)
	if err != nil {
		return nil, fmt.Errorf("GetStreamingLearners: %w", err)
	}
	ids := database.FromTextArray(e.LearnerIds)
	return ids, nil
}

const stmtIncreaseNumberOfStreamingTpl = `UPDATE lessons SET stream_learner_counter = stream_learner_counter+1, learner_ids = array_append(learner_ids, $1) WHERE lesson_id = $2 AND stream_learner_counter < $3 AND NOT($1 = ANY(learner_ids))`

func (l *LessonRepo) IncreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID pgtype.Text, maximumLearnerStreamings int) error {
	commandTag, err := db.Exec(ctx, stmtIncreaseNumberOfStreamingTpl, learnerID, lessonID, maximumLearnerStreamings)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrUnAffected
	}
	return nil
}

const stmtDecreaseNumberOfStreamingTpl = `UPDATE lessons SET stream_learner_counter = stream_learner_counter-1, learner_ids = array_remove(learner_ids, $1) WHERE lesson_id =$2 AND stream_learner_counter >0 AND $1 = ANY(learner_ids)`

func (l *LessonRepo) DecreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID pgtype.Text) error {
	commandTag, err := db.Exec(ctx, stmtDecreaseNumberOfStreamingTpl, learnerID, lessonID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return ErrUnAffected
	}
	return nil
}

func (l *LessonRepo) Create(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error) {
	err := lesson.PreInsert()
	if err != nil {
		return nil, fmt.Errorf("could not pre-insert for new lesson %v", err)
	}

	fieldNames, args := lesson.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		lesson.TableName(),
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return nil, err
	}

	return lesson, nil
}

func (l *LessonRepo) Update(ctx context.Context, db database.Ext, lesson *entities.Lesson) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Update")
	defer span.End()

	err := lesson.PreUpdate()
	if err != nil {
		return fmt.Errorf("lesson.PreUpdate: %s", err)
	}

	updatedFields := []string{
		"name",
		"start_time",
		"end_time",
		"lesson_group_id",
		"lesson_type",
		"teacher_id",
		"course_id",
		"updated_at",
		"stream_learner_counter",
		"learner_ids",
	}

	cmd, err := database.UpdateFields(ctx, lesson, db.Exec, "lesson_id", updatedFields)
	if err != nil {
		return fmt.Errorf("database.Update: %s", err)
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("expect 1 row affected, got %d", cmd.RowsAffected())
	}
	return nil
}

// UpsertLessonTeachers also deletes all rows belonging to lessonID before upserting.
func (l *LessonRepo) UpsertLessonTeachers(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpsertLessonTeachers")
	defer span.End()

	var now pgtype.Timestamptz
	if err := now.Set(time.Now()); err != nil {
		return fmt.Errorf("now.Set(time.Now()): %s", err)
	}

	b := &pgx.Batch{}
	b.Queue(`UPDATE lessons_teachers SET deleted_at = $2 WHERE lesson_id = $1`, lessonID, now)
	l.queueUpsertLessonTeacher(b, lessonID, teacherIDs, now)
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return nil
}

func (l *LessonRepo) queueUpsertLessonTeacher(b *pgx.Batch, lessonID pgtype.Text, teacherIDs pgtype.TextArray, t pgtype.Timestamptz) {
	queueFn := func(b *pgx.Batch, teacherID pgtype.Text) {
		query := `
			INSERT INTO lessons_teachers (lesson_id, teacher_id) VALUES ($1, $2)
			ON CONFLICT ON CONSTRAINT lessons_teachers_pk 
				DO UPDATE SET created_at = $3, deleted_at = NULL`
		b.Queue(query, lessonID, teacherID, t)
	}

	for _, teacherID := range teacherIDs.Elements {
		queueFn(b, teacherID)
	}
}

// UpsertLessonMembers also deletes all rows belonging to lessonID before upserting.
func (l *LessonRepo) UpsertLessonMembers(ctx context.Context, db database.Ext, lessonID pgtype.Text, userIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpsertLessonMembers")
	defer span.End()

	var now pgtype.Timestamptz
	err := now.Set(time.Now())
	if err != nil {
		return fmt.Errorf("now.Set(time.Now()): %w", err)
	}

	b := &pgx.Batch{}
	b.Queue(fmt.Sprintf(`UPDATE %s SET deleted_at = $2 WHERE lesson_id = $1`, (&entities.LessonMember{}).TableName()), lessonID, now)
	l.queueUpsertLessonMember(b, lessonID, userIDs, now)
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return nil
}

func (l *LessonRepo) queueUpsertLessonMember(b *pgx.Batch, lessonID pgtype.Text, userIDs pgtype.TextArray, t pgtype.Timestamptz) {
	lmRepo := &LessonMemberRepo{}
	for _, userID := range userIDs.Elements {
		lm := &entities.LessonMember{}
		database.AllNullEntity(lm)
		lm.LessonID = lessonID
		lm.UserID = userID
		lm.CreatedAt = t
		lm.UpdatedAt = t
		lmRepo.UpsertQueue(b, lm)
	}
}

// UpsertLessonCourses also deletes all rows belonging to lessonID before upserting.
func (l *LessonRepo) UpsertLessonCourses(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpsertLessonCourses")
	defer span.End()

	var now pgtype.Timestamptz
	err := now.Set(time.Now())
	if err != nil {
		return fmt.Errorf("now.Set(time.Now()): %w", err)
	}

	b := &pgx.Batch{}
	b.Queue(`UPDATE lessons_courses SET deleted_at = $2 WHERE lesson_id = $1`, lessonID, now)
	l.queueUpsertLessonCourse(b, lessonID, courseIDs, now)
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return nil
}

func (l *LessonRepo) queueUpsertLessonCourse(b *pgx.Batch, lessonID pgtype.Text, courseIDs pgtype.TextArray, t pgtype.Timestamptz) {
	queueFn := func(b *pgx.Batch, courseID pgtype.Text) {
		query := `
			INSERT INTO lessons_courses (lesson_id, course_id) VALUES ($1, $2)
			ON CONFLICT ON CONSTRAINT lessons_courses_pk
				DO UPDATE SET created_at = $3, deleted_at = NULL`
		b.Queue(query, lessonID, courseID, t)
	}

	for _, courseID := range courseIDs.Elements {
		queueFn(b, courseID)
	}
}

func (l *LessonRepo) FindEarliestAndLatestTimeLessonByCourses(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (*entities.CourseAvailableRanges, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindEarliestAndLatestTimeLessonByCourses")
	defer span.End()

	lesson := &entities.Lesson{}
	query := fmt.Sprintf(`
		SELECT lc.course_id, MIN(l.start_time), MAX(l.end_time) FROM %s l 
		JOIN lessons_courses lc ON l.lesson_id = lc.lesson_id 
		WHERE lc.course_id = any($1) 
				AND l.deleted_at IS NULL
				AND lc.deleted_at IS NULL
		GROUP BY lc.course_id`,
		lesson.TableName(),
	)
	rows, err := db.Query(ctx, query, &courseIDs)
	if err != nil {
		return nil, fmt.Errorf("err db.Query: %w", err)
	}
	defer rows.Close()

	res := &entities.CourseAvailableRanges{}
	for rows.Next() {
		var id pgtype.Text
		var startDate, endDate pgtype.Timestamptz
		if err := rows.Scan(&id, &startDate, &endDate); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		res.Add(&entities.CourseAvailableRange{
			ID:        id,
			StartDate: startDate,
			EndDate:   endDate,
		})
	}

	return res, nil
}

func (l *LessonRepo) Delete(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Delete")
	defer span.End()

	query := "UPDATE lessons SET deleted_at = now(), updated_at = now() WHERE lesson_id = ANY($1) AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &lessonIDs)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonRepo) DeleteLessonMembers(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.DeleteLessonMembers")
	defer span.End()

	query := "UPDATE lesson_members SET deleted_at = now() WHERE lesson_id = $1 AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonRepo) DeleteLessonTeachers(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.DeleteLessonTeachers")
	defer span.End()

	query := "UPDATE lessons_teachers SET deleted_at = now() WHERE lesson_id = $1 AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonRepo) DeleteLessonCourses(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.DeleteLessonCourses")
	defer span.End()

	query := "UPDATE lessons_courses SET deleted_at = now() WHERE lesson_id = $1 AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonRepo) UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpdateLessonRoomState")
	defer span.End()

	query := `UPDATE lessons SET room_state = $2
			WHERE lesson_id = $1 AND deleted_at IS NULL`

	_, err := db.Exec(ctx, query, &lessonID, &state)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonRepo) UpdateRoomID(ctx context.Context, db database.QueryExecer, lessonID, roomID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpdateRoomID")
	defer span.End()

	query := "UPDATE lessons SET updated_at = now(), room_id = $1, status = 'LESSON_STATUS_NOT_STARTED' WHERE lesson_id = $2"
	cmdTag, err := db.Exec(ctx, query, &roomID, &lessonID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot update lesson")
	}

	return nil
}

func (l *LessonRepo) GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GrantRecordingPermission")
	defer span.End()

	state, err := recordingState.MarshalJSON()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`update lessons set room_state = coalesce(room_state || '%s', '%s')
	where lesson_id = $1 and (coalesce(room_state->'recording'->'is_recording', 'false') = 'false');`, string(state), string(state))
	_, err = db.Exec(ctx, query, &lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonRepo) StopRecording(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.StopRecording")
	defer span.End()

	state, err := recordingState.MarshalJSON()
	if err != nil {
		return err
	}
	escapeCreator := fmt.Sprintf("\"%s\"", creator.String)
	query := fmt.Sprintf(`update lessons set room_state = coalesce(room_state || '%s', '%s')
	where lesson_id = $1 and (coalesce(room_state->'recording'->'creator', '%s') = '%s');`, string(state), string(state), escapeCreator, escapeCreator)
	_, err = db.Exec(ctx, query, &lessonID)
	if err != nil {
		return err
	}

	return nil
}
