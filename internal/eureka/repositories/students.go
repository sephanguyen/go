package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// StudentRepo works with class_students, course_students, course_classes
type StudentRepo struct{}

const findStudentsByCourseIDStmtTpl = `SELECT ARRAY (
	SELECT
		student_id
	FROM
		course_students
	WHERE
		deleted_at IS NULL
		AND course_id = $1
	GROUP BY student_id);`

// FindStudentsByCourseID returns studentIDs array specific by courseID
func (r *StudentRepo) FindStudentsByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error) {
	studentIDs := &pgtype.TextArray{}
	err := database.Select(ctx, db, findStudentsByCourseIDStmtTpl, &courseID).ScanFields(studentIDs)
	if err != nil {
		return nil, err
	}

	return studentIDs, nil
}

const findStudentsByClassIDsStmtTpl = `SELECT ARRAY (
	SELECT
		student_id
	FROM
		class_students
	WHERE
		deleted_at IS NULL
		AND class_id = ANY($1)
	GROUP BY
		student_id);`

// FindStudentsByClassIDs returns studentIDs specific by classIDs
func (r *StudentRepo) FindStudentsByClassIDs(ctx context.Context, db database.QueryExecer, classIDs pgtype.TextArray) (*pgtype.TextArray, error) {
	studentIDs := &pgtype.TextArray{}
	err := database.Select(ctx, db, findStudentsByClassIDsStmtTpl, &classIDs).ScanFields(studentIDs)
	if err != nil {
		return nil, err
	}

	return studentIDs, nil
}

const findClassesByCourseIDStmtTpl = `SELECT ARRAY (
	SELECT
		cm.class_id
	FROM
		class_students cm
	JOIN course_students cs ON
		cm.student_id = cs.student_id
		AND cs.course_id = $1
		AND cs.deleted_at IS NULL
	WHERE cm.deleted_at IS NULL
	GROUP BY
		cm.class_id);`

// FindClassesByCourseID returns only classes related to students
func (r *StudentRepo) FindClassesByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error) {
	classIDs := &pgtype.TextArray{}
	err := database.Select(ctx, db, findClassesByCourseIDStmtTpl, &courseID).ScanFields(classIDs)
	if err != nil {
		return nil, err
	}

	return classIDs, nil
}

// FindStudentsByCourseID returns studentIDs array specific by courseID
func (r *StudentRepo) FindStudentsByCourseLocation(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, locationIDs pgtype.TextArray) (*pgtype.TextArray, error) {
	studentIDs := &pgtype.TextArray{}
	cse := &entities.CourseStudent{}
	csape := &entities.CourseStudentsAccessPath{}
	stmt := fmt.Sprintf(`SELECT ARRAY (
		SELECT
			student_id
		FROM
			%s as cs
		JOIN %s as csap
		USING(course_student_id,student_id,course_id)
		WHERE
			cs.deleted_at IS NULL
			AND csap.deleted_at IS NULL
			AND course_id = $1::TEXT
			AND (
				location_id = ANY($2::TEXT[]) OR ($2 IS NULL)
			)
			
		GROUP BY student_id);`, cse.TableName(), csape.TableName())

	err := database.Select(ctx, db, stmt, &courseID, &locationIDs).ScanFields(studentIDs)
	if err != nil {
		return nil, err
	}

	return studentIDs, nil
}

// FindStudentsByCourseID returns studentIDs array specific by courseID
func (r *StudentRepo) FindStudentsByLocation(ctx context.Context, db database.QueryExecer, locationIDs pgtype.TextArray) (*pgtype.TextArray, error) {
	studentIDs := &pgtype.TextArray{}
	cse := &entities.CourseStudent{}
	csape := &entities.CourseStudentsAccessPath{}
	stmt := fmt.Sprintf(`SELECT ARRAY (
		SELECT DISTINCT
			student_id
		FROM
			%s as cs
		JOIN %s as csap
		USING(course_student_id,student_id,course_id)
		WHERE
			cs.deleted_at IS NULL
			AND csap.deleted_at IS NULL
			AND (
				location_id = ANY($1::TEXT[]) OR ($1 IS NULL)
			)
			
		GROUP BY student_id);`, cse.TableName(), csape.TableName())

	err := database.Select(ctx, db, stmt, &locationIDs).ScanFields(studentIDs)
	if err != nil {
		return nil, err
	}

	return studentIDs, nil
}

type StudentInfo struct {
	StudentID pgtype.Text
	Grade     pgtype.Int2
	CourseIDs pgtype.TextArray
}

func (r *StudentRepo) FilterByGradeBookView(
	ctx context.Context,
	db database.QueryExecer,
	studentIDs,
	studyPlanIDs pgtype.TextArray,
	courseIDs pgtype.TextArray,
	grades pgtype.Int4Array,
	gradeIDs pgtype.TextArray,
	studentName pgtype.Text,
	locationIDs pgtype.TextArray,
	limit,
	offset int64,
) ([]*StudentInfo, error) {
	studentInfos := make([]*StudentInfo, 0)
	var rows pgx.Rows
	var err error

	if len(locationIDs.Elements) > 0 {
		stmt := `
		SELECT
			u.user_id,
			MIN(s.current_grade) AS current_grade,
			ARRAY_AGG(DISTINCT csp.course_id)::TEXT[] AS course_ids
		FROM
			students s
			JOIN users u ON s.student_id = u.user_id
			JOIN student_study_plans ssp ON ssp.student_id = s.student_id
			JOIN course_study_plans csp ON csp.study_plan_id = ssp.master_study_plan_id
			JOIN master_study_plan msp ON ssp.master_study_plan_id = msp.study_plan_id
			JOIN exam_lo el ON msp.learning_material_id = el.learning_material_id
			JOIN course_students_access_paths cap ON csp.course_id = cap.course_id AND u.user_id = cap.student_id
			where ($1::TEXT[] is null or s.student_id = ANY($1::TEXT[]))
			and ($2::TEXT[] is null or ssp.study_plan_id = ANY($2::TEXT[]))
			and ($3::TEXT[] is null or csp.course_id = ANY($3::TEXT[]))
			and ($4::INTEGER[] is null or s.current_grade = ANY($4::INTEGER[]))
			and u.name ilike '%' || $7 || '%'
			and ($8::TEXT[] is null or s.grade_id = any($8::TEXT[]))
			and ($9::TEXT[] is null or cap.location_id = ANY($9::TEXT[]))
		GROUP BY
			u.user_id, u.name
		ORDER BY
			u.name
		LIMIT $5 OFFSET $6;
		`
		rows, err = db.Query(ctx, stmt, &studentIDs, &studyPlanIDs, &courseIDs, &grades, &limit, &offset, &studentName, &gradeIDs, &locationIDs)
	} else {
		stmt := `select u.user_id, min(s.current_grade) current_grade, array_agg(distinct csp.course_id)::text[] from students s
		join users u on s.student_id = u.user_id
		join student_study_plans ssp on ssp.student_id = s.student_id
		join course_study_plans csp on csp.study_plan_id = ssp.master_study_plan_id 
		join master_study_plan msp on ssp.master_study_plan_id = msp.study_plan_id
		join exam_lo el on msp.learning_material_id = el.learning_material_id
		where ($1::TEXT[] is null or s.student_id = ANY($1::TEXT[]))
		and ($2::TEXT[] is null or ssp.study_plan_id = ANY($2::TEXT[]))
		and ($3::TEXT[] is null or csp.course_id = ANY($3::TEXT[]))
		and ($4::INTEGER[] is null or s.current_grade = ANY($4::INTEGER[]))
		and ($8::TEXT[] is null or s.grade_id = any($8::TEXT[]))
		and u.name ilike '%' || $7 || '%'
		group by u.user_id
		order by u.name
		LIMIT $5 OFFSET $6;`

		rows, err = db.Query(ctx, stmt, &studentIDs, &studyPlanIDs, &courseIDs, &grades, &limit, &offset, &studentName, &gradeIDs)
	}

	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	for rows.Next() {
		studentInfo := StudentInfo{}
		err = rows.Scan(&studentInfo.StudentID, &studentInfo.Grade, &studentInfo.CourseIDs)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		studentInfos = append(studentInfos, &studentInfo)
	}

	return studentInfos, nil
}

func (r *StudentRepo) FilterOutDeletedStudentIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentIDs []string,
) ([]string, error) {
	query := `SELECT student_id FROM students WHERE deleted_at IS NULL AND student_id = ANY($1)`

	rows, err := db.Query(ctx, query, &studentIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("StudentsRepo.FilterOutDeletedStudents.Err: %w", err)
	}

	var result []string
	for rows.Next() {
		var studentID string
		if err := rows.Scan(&studentID); err != nil {
			return nil, fmt.Errorf("StudentsRepo.FilterOutDeletedStudents.Scan: %w", err)
		}

		result = append(result, studentID)
	}

	return result, nil
}
