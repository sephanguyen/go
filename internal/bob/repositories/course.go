package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
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

type CourseRepo struct{}

type CourseQuery struct {
	IDs         []string
	Name        string
	Countries   []string
	Subject     string
	Grade       int
	SchoolIDs   []int
	ClassIDs    []int
	Limit       int
	Offset      int
	Type        string
	Status      string
	LocationIDs []string
	Keyword     string
}

func (r *CourseRepo) QueryRetrieveCourses(q *CourseQuery, isCount bool) (string, []interface{}, error) {
	e := new(entities.Course)
	fields := database.GetFieldNames(e)

	var (
		currentTime                        pgtype.Timestamptz
		countries                          pgtype.TextArray
		name, subject, courseType, keyword pgtype.Text
		schoolIDs, classIDs                pgtype.Int4Array
		grade                              pgtype.Int4
		courseIDs, locationIDs             pgtype.TextArray
	)

	err := multierr.Combine(
		name.Set(nil),
		subject.Set(nil),
		courseType.Set(nil),
		classIDs.Set(nil),
		grade.Set(nil),
		schoolIDs.Set(nil),
		courseIDs.Set(nil),
		countries.Set(nil),
		locationIDs.Set(nil),
		keyword.Set(nil),
	)

	endAt := time.Now()
	err = multierr.Append(err, currentTime.Set(endAt))

	if len(q.IDs) > 0 {
		err = multierr.Append(err, courseIDs.Set(q.IDs))
	}
	if q.Name != "" {
		err = multierr.Append(err, name.Set(q.Name))
	}
	if q.Subject != "" && q.Subject != pb.SUBJECT_NONE.String() {
		err = multierr.Append(err, subject.Set(q.Subject))
	}
	if q.Grade > 0 {
		err = multierr.Append(err, grade.Set(q.Grade))
	}
	if q.Type != pb.COURSE_TYPE_NONE.String() && q.Type != "" {
		err = multierr.Append(err, courseType.Set(q.Type))
	}

	if len(q.SchoolIDs) != 0 {
		err = multierr.Append(err, schoolIDs.Set(q.SchoolIDs))
	}
	if len(q.ClassIDs) != 0 {
		err = multierr.Append(err, classIDs.Set(q.ClassIDs))
	}
	if len(q.Countries) != 0 {
		err = multierr.Append(err, countries.Set(q.Countries))
	}
	if len(q.LocationIDs) > 0 {
		err = multierr.Append(err, locationIDs.Set(q.LocationIDs))
	}
	if len(q.Keyword) > 0 {
		err = multierr.Append(err, keyword.Set(q.Keyword))
	}
	if err != nil {
		return "", nil, err
	}

	args := []interface{}{&countries, &courseIDs, &name, &subject, &grade, &courseType, &currentTime, &classIDs, &schoolIDs}

	commonCond := `c.deleted_at IS NULL AND c.status != 'COURSE_STATUS_INACTIVE'
		AND (ay.status IS NULL OR ay.status = 'ACADEMIC_YEAR_STATUS_ACTIVE')
		AND ($1::text[] IS NULL OR country = ANY($1))
		AND ($2::text[] IS NULL OR c.course_id = ANY($2))
		AND ($3::text IS NULL OR c.name = $3)
		AND ($4::text IS NULL OR c.subject = $4)
		AND ($5::int IS NULL OR c.grade = $5)
		AND ($6::text IS NULL OR c.course_type = $6) `

	var query, cond, joinQuery, condStatus, condLocation, condKeyword string
	condStatus = " AND ($7::timestamptz IS NULL OR $7::timestamptz IS NOT NULL) " // add for default case

	joinQuery += `LEFT JOIN courses_academic_years ca
					ON c.course_id = ca.course_id
				  LEFT JOIN academic_years ay
					ON ay.academic_year_id = ca.academic_year_id`

	if q.Status != "" && !currentTime.Time.IsZero() {
		joinQuery += " LEFT JOIN lessons AS l ON c.course_id = l.course_id AND l.deleted_at is NULL"

		switch q.Status {
		case pb.COURSE_STATUS_ACTIVE.String():
			condStatus = ` AND c.end_date >= $7 AND l.lesson_id IS NOT NULL `
		case pb.COURSE_STATUS_COMPLETED.String():
			condStatus = ` AND c.end_date < $7 AND l.lesson_id IS NOT NULL `
		case pb.COURSE_STATUS_ON_GOING.String():
			condStatus += ` AND l.lesson_id IS NULL `
		}
	}
	if len(q.ClassIDs) > 0 && len(q.SchoolIDs) > 0 {
		joinQuery += " LEFT JOIN courses_classes ON c.course_id = courses_classes.course_id "
		cond = ` AND (($8::int[] IS NOT NULL AND $9::int[] IS NOT NULL
			AND ((courses_classes.class_id = ANY($8) AND courses_classes.deleted_at IS NULL) OR c.school_id = ANY($9)))
			OR ($8::int[] IS NOT NULL AND $9::int[] IS NULL AND courses_classes.class_id = ANY($8) AND courses_classes.deleted_at IS NULL))`
	}
	if len(q.ClassIDs) > 0 && len(q.SchoolIDs) == 0 {
		joinQuery += " INNER JOIN courses_classes ON c.course_id = courses_classes.course_id "
		cond = ` AND ($8::int[] IS NOT NULL AND $9::int[] IS NULL AND courses_classes.class_id = ANY($8) AND courses_classes.deleted_at IS NULL)`
	}

	if len(q.ClassIDs) == 0 {
		cond = ` AND ($8::int[] IS NULL) AND ($9::int[] IS NULL OR c.school_id = ANY($9) )`
	}
	if len(q.LocationIDs) > 0 {
		joinQuery += " INNER JOIN course_access_paths ON c.course_id = course_access_paths.course_id "
		condLocation = ` AND ($10::text[] IS NULL OR course_access_paths.location_id = ANY($10)) and course_access_paths.deleted_at is NULL`
		args = append(args, &locationIDs)
	}
	if len(q.Keyword) > 0 {
		keywordOrder := len(args) + 1
		condKeyword = fmt.Sprintf(` AND ($%d::text IS null OR (lower(c.name) like lower(CONCAT('%%',$%d,'%%'))))`, keywordOrder, keywordOrder)
		args = append(args, &keyword)
	}
	cond = commonCond + condStatus + cond + condLocation + condKeyword

	query = fmt.Sprintf("SELECT DISTINCT c.%s FROM %s c %s WHERE %s ORDER BY c.created_at DESC, c.name ASC, c.course_id DESC", strings.Join(fields, ",c."), e.TableName(), joinQuery, cond)
	if isCount {
		query = fmt.Sprintf("SELECT COUNT(DISTINCT c.course_id) FROM %s c %s WHERE %s", e.TableName(), joinQuery, cond)
	}

	return query, args, nil
}

func (r *CourseRepo) RetrieveCourses(ctx context.Context, db database.QueryExecer, q *CourseQuery) (entities.Courses, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.RetrieveCourses")
	defer span.End()

	query, args, err := r.QueryRetrieveCourses(q, false)

	if err != nil {
		return nil, fmt.Errorf("QueryRetrieveCourses: %w", err)
	}
	if q.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.Limit)
	}
	if q.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	courses := entities.Courses{}
	err = database.Select(ctx, db, query, args...).ScanAll(&courses)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return courses, nil
}

func (r *CourseRepo) CountCourses(ctx context.Context, db database.QueryExecer, q *CourseQuery) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.CountCourses")
	defer span.End()

	query, args, err := r.QueryRetrieveCourses(q, true)
	if err != nil {
		return 0, fmt.Errorf("QueryRetrieveCourses: %w", err)
	}

	var count int
	if err := database.Select(ctx, db, query, args...).ScanFields(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *CourseRepo) RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (entities.Courses, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.RetrieveByIDs")
	defer span.End()

	e := &entities.Course{}
	fieldNames := database.GetFieldNames(e)
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE course_id = ANY($1) AND deleted_at IS NULL`, strings.Join(fieldNames, ", "), e.TableName())

	var courses entities.Courses
	err := database.Select(ctx, db, query, &ids).ScanAll(&courses)
	if err != nil {
		return nil, errors.Wrap(err, "database.Select")
	}

	return courses, nil
}

func (r *CourseRepo) FindSchoolIDsOnCourses(ctx context.Context, db database.QueryExecer, courseIDs []string) ([]int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.FindSchoolIDsOnCourses")
	defer span.End()

	query := "SELECT school_id FROM courses WHERE deleted_at IS NULL AND course_id = ANY($1)"
	pgIDs := database.TextArray(courseIDs)

	schoolIDs := EnSchoolIDs{}
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

func (r *CourseRepo) Upsert(ctx context.Context, db database.Ext, cc []*entities.Course) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities.Course) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT courses_pk DO UPDATE
		SET name = $2, country = $3, subject = $4, grade = $5, display_order = $6, school_id = $7,
		updated_at = $10, deleted_at = $12, start_date = $13, end_date = $14, preset_study_plan_id = $15, icon = $16`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range cc {
		err := multierr.Combine(
			t.CreatedAt.Set(now),
			t.UpdatedAt.Set(now),
		)

		if t.ID.String == "" {
			err = multierr.Append(err, t.ID.Set(idutil.ULIDNow()))
		}

		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("course book not inserted")
		}
	}
	return nil
}

func (r *CourseRepo) FindByID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*entities.Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.FindByID")
	defer span.End()

	e := &entities.Course{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND course_id = $1", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, &courseID).ScanOne(e)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (r *CourseRepo) FindByIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (map[pgtype.Text]*entities.Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.FindByIDs")
	defer span.End()

	e := new(entities.Course)
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND course_id = ANY($1)", strings.Join(fields, ","), e.TableName())

	courses := entities.Courses{}
	err := database.Select(ctx, db, query, &courseIDs).ScanAll(&courses)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	result := map[pgtype.Text]*entities.Course{}
	for _, course := range courses {
		result[course.ID] = course
	}

	return result, nil
}

func (r *CourseRepo) FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.Courses, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.FindByLessonID")
	defer span.End()

	e := &entities.Course{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf(
		`SELECT c.%s 
		FROM courses c 
		JOIN lessons_courses lc
			ON c.course_id = lc.course_id  
		WHERE c.deleted_at IS NULL AND lc.deleted_at IS NULL AND lc.lesson_id = $1`,
		strings.Join(fields, ", c."),
	)

	courses := entities.Courses{}
	err := database.Select(ctx, db, query, &lessonID).ScanAll(&courses)
	if err != nil {
		return nil, err
	}

	return courses, nil
}

func (r *CourseRepo) SoftDelete(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.SoftDelete")
	defer span.End()

	query := "UPDATE courses SET deleted_at = now(), updated_at = now() WHERE course_id = ANY($1) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &courseIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("cannot delete course: %w", pgx.ErrNoRows)
	}

	return nil
}

// UpsertV2 insert on conflict only update name (because in JPREF only send name and id in currently)
func (r *CourseRepo) UpsertV2(ctx context.Context, db database.Ext, cc []*entities.Course) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.UpsertV2")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities.Course) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT courses_pk DO UPDATE
		SET updated_at = $10, name = $2, status = $17, deleted_at = NULL`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range cc {
		err := multierr.Combine(
			t.CreatedAt.Set(now),
			t.UpdatedAt.Set(now),
		)

		if t.ID.String == "" {
			err = multierr.Append(err, t.ID.Set(idutil.ULIDNow()))
		}

		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("course book not inserted")
		}
	}
	return nil
}

type UpdateAcademicYearOpts struct {
	CourseID       string
	AcademicYearID string
}

func (r *CourseRepo) UpdateAcademicYear(ctx context.Context, db database.Ext, cc []*UpdateAcademicYearOpts) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.UpdateAcademicYear")
	defer span.End()

	queue := func(b *pgx.Batch, c *UpdateAcademicYearOpts) {
		query := `INSERT INTO courses_academic_years
					(course_id, academic_year_id, updated_at, created_at)
					VALUES ($1, $2, NOW(), NOW())
					ON CONFLICT(course_id, academic_year_id)
					DO UPDATE SET
					updated_at = NOW(),
					deleted_at = NULL`
		b.Queue(query, database.Text(c.CourseID), database.Text(c.AcademicYearID))
	}

	b := &pgx.Batch{}
	for _, t := range cc {
		queue(b, t)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("err update academicYear: %v-%v: %w", cc[i].CourseID, cc[i].AcademicYearID, err)
		}

		if ct.RowsAffected() != 1 {
			return fmt.Errorf("not found courseID %v to update academicYear", cc[i].CourseID)
		}
	}

	return nil
}

func (r *CourseRepo) GetPresetStudyPlanIDsByCourseIDs(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.GetPresetStudyPlanIDsByCourseIDs")
	defer span.End()

	const query = `
		SELECT preset_study_plan_id
		FROM courses
		WHERE course_id = ANY($1)
			AND deleted_at IS NULL`
	rows, err := db.Query(ctx, query, courseIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %s", err)
	}
	defer rows.Close()

	var pspIDs []string
	for rows.Next() {
		var pspID string
		if err := rows.Scan(&pspID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %s", err)
		}
		pspIDs = append(pspIDs, pspID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %s", err)
	}
	return pspIDs, nil
}

func (r *CourseRepo) UpdateStartAndEndDate(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.UpdateStartAndEndDate")
	defer span.End()

	// Use LEFT JOIN here to handle courses that do not belong to any lessons.
	// In those cases, start_date and end_date are NULL.
	query := `
		WITH t AS (
			SELECT c.course_id AS course_id, MIN(l.start_time) AS start_date, MAX(l.end_time) AS end_date
			FROM courses c
			LEFT JOIN lessons_courses lc 
				ON lc.course_id = c.course_id
				AND lc.deleted_at IS NULL
			LEFT JOIN lessons l 
				ON l.lesson_id = lc.lesson_id
				AND l.deleted_at IS NULL
			WHERE c.course_id = ANY($1)
				AND c.deleted_at IS NULL
			GROUP BY c.course_id
		)
		UPDATE courses c
		SET start_date = t.start_date,
			end_date = t.end_date
		FROM t
		WHERE c.course_id = t.course_id`

	cmdTag, err := db.Exec(ctx, query, courseIDs)
	if err != nil {
		return fmt.Errorf("db.Exec: %s", err)
	}
	if cmdTag.RowsAffected() != int64(len(courseIDs.Elements)) {
		return fmt.Errorf("expect %d row(s) updated, got %d", len(courseIDs.Elements), cmdTag.RowsAffected())
	}
	return nil
}

func (r *CourseRepo) Create(ctx context.Context, db database.QueryExecer, a *entities.Course) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.Create")
	defer span.End()

	now := timeutil.Now()
	_ = a.CreatedAt.Set(now)
	_ = a.UpdatedAt.Set(now)

	if _, err := database.InsertIgnoreConflict(ctx, a, db.Exec); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}
