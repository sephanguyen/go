package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// StudentSubmissionRepo works with "student_submissions" table
type StudentSubmissionRepo struct{}

// Create create new entities with newly generated ID
func (r *StudentSubmissionRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentSubmission) error {
	e.ID.Set(idutil.ULIDNow())
	e.Now()

	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return err
	}

	return nil
}

// Get returns a single submission
func (r *StudentSubmissionRepo) Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.StudentSubmission, error) {
	e := &entities.StudentSubmission{}
	fields := strings.Join(database.GetFieldNames(e), ",")

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_submission_id = $1 AND deleted_at IS NULL", fields, e.TableName())
	err := database.Select(ctx, db, stmt, &id).ScanOne(e)

	return e, err
}

// List

// StudentSubmissionFilter used in List
type StudentSubmissionFilter struct {
	Limit uint

	OffsetID pgtype.Text

	StudentIDs,
	Statuses pgtype.TextArray

	StartDate,
	EndDate,
	CreatedAt pgtype.Timestamptz

	AssignmentName pgtype.Text
	CourseID       pgtype.Text
	ClassIDs       pgtype.TextArray

	LocationIDs pgtype.TextArray

	StudentName pgtype.Text
}

const listStmtTpl1 = `SELECT
	student_latest_submissions.%s
FROM
	student_latest_submissions
INNER JOIN study_plan_items ON
	student_latest_submissions.study_plan_item_id = study_plan_items.study_plan_item_id
INNER JOIN assignments ON
	assignments.assignment_id = student_latest_submissions.assignment_id
WHERE
	($3::_text IS NULL OR student_latest_submissions.student_id = ANY($3))
	AND ($4::text IS NULL OR assignments.name LIKE $4)
	AND ($5::text IS NULL OR student_latest_submissions.student_submission_id < $5)
	AND ($6::timestamp IS NULL OR student_latest_submissions.created_at < $6)
	AND ($7::_text IS NULL OR student_latest_submissions.status = ANY($7))
	AND study_plan_items.start_date BETWEEN $1 AND $2
	AND study_plan_items.deleted_at IS NULL
	AND student_latest_submissions.deleted_at IS NULL
  AND assignments.type != $8
ORDER BY
	student_latest_submissions.created_at DESC, student_latest_submissions.student_submission_id DESC
LIMIT %d`

const listWithCourseStmtTpl1 = `SELECT
	student_latest_submissions.%s
FROM
	student_latest_submissions
INNER JOIN study_plan_items ON
	student_latest_submissions.study_plan_item_id = study_plan_items.study_plan_item_id
INNER JOIN assignments ON
	assignments.assignment_id = student_latest_submissions.assignment_id
JOIN study_plans ON
	study_plans.study_plan_id = study_plan_items.study_plan_id
WHERE
	($3::_text IS NULL OR student_latest_submissions.student_id = ANY($3))
	AND ($4::text IS NULL OR assignments.name LIKE $4)
	AND ($6::text IS NULL OR student_latest_submissions.student_submission_id < $6)
	AND ($7::timestamp IS NULL OR student_latest_submissions.created_at < $7)
	AND ($8::_text IS NULL OR student_latest_submissions.status = ANY($8))
	AND study_plan_items.start_date BETWEEN $1 AND $2
	AND study_plan_items.deleted_at IS NULL
	AND ($5::text IS NULL OR study_plans.course_id = $5)
	AND student_latest_submissions.deleted_at IS NULL
  AND assignments.type != $9
ORDER BY
	student_latest_submissions.created_at DESC, student_latest_submissions.student_submission_id DESC
LIMIT %d`

// List returns a list of submissions matching the filter
func (r *StudentSubmissionRepo) List(ctx context.Context, db database.QueryExecer, filter *StudentSubmissionFilter) (entities.StudentSubmissions, error) {
	var tpl string
	var args []interface{}

	if filter.CourseID.Status == pgtype.Null {
		tpl = listStmtTpl1
		args = []interface{}{
			&filter.StartDate,
			&filter.EndDate,
			&filter.StudentIDs,
			&filter.AssignmentName,
			&filter.OffsetID,
			&filter.CreatedAt,
			&filter.Statuses,
			epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
		}
	} else {
		tpl = listWithCourseStmtTpl1
		args = []interface{}{
			&filter.StartDate,
			&filter.EndDate,
			&filter.StudentIDs,
			&filter.AssignmentName,
			&filter.CourseID,
			&filter.OffsetID,
			&filter.CreatedAt,
			&filter.Statuses,
			epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
		}
	}

	e := &entities.StudentSubmission{}
	listStmt := fmt.Sprintf(tpl, strings.Join(database.GetFieldNames(e), ",student_latest_submissions."), filter.Limit)
	results := make(entities.StudentSubmissions, 0, int(filter.Limit))

	if err := database.Select(ctx, db, listStmt, args...).ScanAll(&results); err != nil {
		return nil, err
	}

	return results, nil
}

// List returns a list of submissions matching the filter
func (r *StudentSubmissionRepo) ListV2(ctx context.Context, db database.QueryExecer, filter *StudentSubmissionFilter) (entities.StudentSubmissions, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubmissionRepo.ListV2")
	defer span.End()

	results := make(entities.StudentSubmissions, 0, int(filter.Limit))

	stmt, args := GetListV2Statement(filter)

	if err := database.Select(ctx, db, stmt, args...).ScanAll(&results); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return results, nil
}

const listStmtTpl3 = `
	SELECT DISTINCT
		sls.%s, 
		sp.course_id,
		check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.start_date,isp.start_date) as start_date,  
		check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.end_date,isp.end_date) end_date
	FROM student_latest_submissions sls
	JOIN master_study_plan msp
	USING(study_plan_id, learning_material_id)
	LEFT JOIN individual_study_plan isp
	USING(student_id, study_plan_id, learning_material_id)
	JOIN study_plans sp
	USING(study_plan_id)
	JOIN course_students_access_paths csap
	USING(student_id, course_id)
	JOIN course_students cs
	USING (student_id, course_id)
	JOIN %s lm
	USING(learning_material_id)
	LEFT JOIN class_students css
	USING(student_id)
	WHERE
		NOW() BETWEEN cs.start_at AND cs.end_at
		AND ($1::TIMESTAMP IS NULL OR $2::TIMESTAMP IS NULL OR check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.start_date,isp.start_date) BETWEEN $1 AND $2)
		AND ($3::_TEXT IS NULL OR sls.student_id = ANY($3))
		AND ($4::TEXT IS NULL OR sp.course_id = $4)
		AND ($5::TEXT IS NULL OR sls.student_submission_id < $5)
		AND ($6::TIMESTAMP IS NULL OR sls.created_at < $6)
		AND ($7::_TEXT IS NULL OR sls.status = ANY($7))
		AND ($8::_TEXT IS NULL OR csap.location_id = ANY($8))
		AND lm.type != $9
		AND ($10::_TEXT IS NULL OR css.class_id = ANY($10))

		AND sls.deleted_at IS NULL
		AND msp.deleted_at IS NULL
		AND isp.deleted_at IS NULL
		AND sp.deleted_at IS NULL
		AND csap.deleted_at IS NULL
		AND cs.deleted_at IS NULL
		AND css.deleted_at IS NULL
		AND lm.deleted_at IS NULL
		AND css.deleted_at IS NULL
	ORDER BY
		sls.created_at DESC, sls.student_submission_id DESC
	LIMIT %d
`

type StudentSubmissionInfo struct {
	entities.StudentSubmission
	CourseID  pgtype.Text
	StartDate pgtype.Timestamptz
	EndDate   pgtype.Timestamptz
}

// List returns a list of submissions matching the filter
func (r *StudentSubmissionRepo) ListV3(ctx context.Context, db database.QueryExecer, filter *StudentSubmissionFilter) ([]*StudentSubmissionInfo, error) {
	tpl := listStmtTpl3
	args := []interface{}{
		&filter.StartDate,
		&filter.EndDate,
		&filter.StudentIDs,
		&filter.CourseID,
		&filter.OffsetID,
		&filter.CreatedAt,
		&filter.Statuses,
		&filter.LocationIDs,
		sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String(),
		&filter.ClassIDs,
	}
	lm := "learning_material"
	if filter.AssignmentName.Status == pgtype.Present {
		args = append(args, &filter.AssignmentName)
		lm = "filter_rls_search_name_lm_fn($11::TEXT)"
	}
	e := &entities.StudentSubmission{}
	listStmt := fmt.Sprintf(tpl, strings.Join(database.GetFieldNames(e), ",sls."), lm, filter.Limit)
	results := make([]*StudentSubmissionInfo, 0, int(filter.Limit))
	rows, err := db.Query(ctx, listStmt, args...)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	for rows.Next() {
		info := &StudentSubmissionInfo{}
		_, values := info.StudentSubmission.FieldMap()
		values = append(values, &info.CourseID, &info.StartDate, &info.EndDate)
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		results = append(results, info)
	}
	return results, nil
}

const listStmtTpl4 = `
	SELECT DISTINCT
		sls.%s, 
		sp.course_id,
		check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.start_date,isp.start_date) as start_date,  
		check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.end_date,isp.end_date) end_date
	FROM student_latest_submissions sls JOIN master_study_plan msp 	USING(study_plan_id, learning_material_id)
	LEFT JOIN individual_study_plan isp USING(student_id, study_plan_id, learning_material_id)
	JOIN study_plans sp	USING(study_plan_id)
	JOIN course_students_access_paths csap USING(student_id, course_id)
	LEFT JOIN class_students css USING(student_id)
	JOIN assignment lm USING(learning_material_id)
	JOIN users us ON sls.student_id = us.user_id
	WHERE
		EXISTS (SELECT 1 FROM course_students cs WHERE cs.course_id = sp.course_id 
				AND cs.student_id = sls.student_id
				AND NOW() BETWEEN cs.start_at AND cs.end_at
				AND cs.deleted_at IS NULL)
		AND ($1::TIMESTAMP IS NULL OR $2::TIMESTAMP IS NULL OR check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.start_date,isp.start_date) BETWEEN $1 AND $2)
		AND ($3::_TEXT IS NULL OR sls.student_id = ANY($3))
		AND ($4::TEXT IS NULL OR sp.course_id = $4)
		AND ($5::TEXT IS NULL OR sls.student_submission_id < $5)
		AND ($6::TIMESTAMP IS NULL OR sls.created_at < $6)
		AND ($7::_TEXT IS NULL OR sls.status = ANY($7))
		AND ($8::_TEXT IS NULL OR csap.location_id = ANY($8))
		AND lm.type != $9
		AND ($10::TEXT IS NULL OR lm.name ILIKE '%%' || $10 || '%%')
		AND ($11::_TEXT IS NULL OR css.class_id = ANY($11))
		AND ($12::TEXT IS NULL OR us.name ILIKE '%%' || $12 || '%%')

		AND sls.deleted_at IS NULL
		AND msp.deleted_at IS NULL
		AND isp.deleted_at IS NULL
		AND sp.deleted_at IS NULL
		AND csap.deleted_at IS NULL
		AND css.deleted_at IS NULL
		AND lm.deleted_at IS NULL
		AND css.deleted_at IS NULL
		AND us.deleted_at IS NULL   
	ORDER BY
		sls.created_at DESC, sls.student_submission_id DESC
	LIMIT %d
`

func (r *StudentSubmissionRepo) ListV4(ctx context.Context, db database.QueryExecer, filter *StudentSubmissionFilter) ([]*StudentSubmissionInfo, error) {
	tpl := listStmtTpl4
	args := []interface{}{
		&filter.StartDate,
		&filter.EndDate,
		&filter.StudentIDs,
		&filter.CourseID,
		&filter.OffsetID,
		&filter.CreatedAt,
		&filter.Statuses,
		&filter.LocationIDs,
		sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String(),
		&filter.AssignmentName,
		&filter.ClassIDs,
		&filter.StudentName,
	}

	e := &entities.StudentSubmission{}
	listStmt := fmt.Sprintf(tpl, strings.Join(database.GetFieldNames(e), ",sls."), filter.Limit)
	results := make([]*StudentSubmissionInfo, 0, int(filter.Limit))
	rows, err := db.Query(ctx, listStmt, args...)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	for rows.Next() {
		info := &StudentSubmissionInfo{}
		_, values := info.StudentSubmission.FieldMap()
		values = append(values, &info.CourseID, &info.StartDate, &info.EndDate)
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		results = append(results, info)
	}
	return results, nil
}

const retrieveByIDsStmtTpl = `SELECT
	%s
FROM
	(
	SELECT
		((array_agg(
			student_submissions.* 
			ORDER BY 
				student_submissions.created_at DESC, student_submissions.student_submission_id DESC
		))[1]).*
	FROM
		student_submissions
	JOIN study_plan_items ON
		student_submissions.study_plan_item_id = study_plan_items.study_plan_item_id
		AND study_plan_items.study_plan_item_id = ANY($1)
		AND student_submissions.deleted_at IS NULL
	GROUP BY
		student_submissions.study_plan_item_id
	ORDER BY
		student_submissions.study_plan_item_id DESC ) AS submissions;`

// RetrieveByStudyPlanItemIDs returns student_submissions match study_plan_item_id
func (r *StudentSubmissionRepo) RetrieveByStudyPlanItemIDs(
	ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray,
) (entities.StudentSubmissions, error) {
	e := &entities.StudentSubmission{}
	retrieveStmt := fmt.Sprintf(retrieveByIDsStmtTpl, strings.Join(database.GetFieldNames(e), ","))

	results := make(entities.StudentSubmissions, 0, len(studyPlanItemIDs.Elements))
	err := database.Select(ctx, db, retrieveStmt, studyPlanItemIDs).ScanAll(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

const updateGradeStatusStmt = `UPDATE
	student_submissions
SET
	updated_at = NOW(),
	student_submission_grade_id = $1,
	status = $2,
	editor_id = $3
WHERE
	student_submission_id = $4 
	AND deleted_at IS NULL;`

// UpdateGradeStatus updates correspond grading result
func (r *StudentSubmissionRepo) UpdateGradeStatus(ctx context.Context, db database.QueryExecer,
	id, gradeID, userChangeStatusID, status pgtype.Text) error {
	_, err := db.Exec(ctx, updateGradeStatusStmt, &gradeID, &status, &userChangeStatusID, &id)
	if err != nil {
		return err
	}

	return nil
}

func (r *StudentSubmissionRepo) BulkUpdateStatus(ctx context.Context, db database.QueryExecer, editorID, status pgtype.Text, grades []*entities.StudentSubmissionGrade) error {
	queueFn := func(b *pgx.Batch, editorID, status pgtype.Text, e *entities.StudentSubmissionGrade) {
		query := "UPDATE student_submissions SET updated_at=now(),student_submission_grade_id=$1, status=$2, editor_id=$3 WHERE student_submission_id=$4"
		b.Queue(query, e.ID, status, editorID, e.StudentSubmissionID)
	}
	b := &pgx.Batch{}
	for _, each := range grades {
		queueFn(b, editorID, status, each)
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

// DeleteByStudyPlanItemIDs updates deleted_at = now()
func (r *StudentSubmissionRepo) DeleteByStudyPlanItemIDs(
	ctx context.Context, db database.QueryExecer,
	studyPlanItemIDs pgtype.TextArray, deletedBy pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubmissionRepo.DeleteByStudyPlanItemIDs")
	defer span.End()

	a := &entities.StudentSubmission{}
	query := fmt.Sprintf("UPDATE %s SET deleted_at = now(), deleted_by = $1 WHERE study_plan_item_id = ANY($2)", a.TableName())
	commandTag, err := db.Exec(ctx, query, &deletedBy, &studyPlanItemIDs)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no raw affected, failed delete study plan item")
	}
	return nil
}

const findBySubmissionIDStmt = "SELECT %s FROM %s WHERE student_submission_id = ANY($1)"

func (r *StudentSubmissionRepo) FindBySubmissionIDs(ctx context.Context, db database.QueryExecer, submissionIDs pgtype.TextArray) (*entities.StudentSubmissions, error) {
	e := entities.StudentSubmissions{}
	submission := entities.StudentSubmission{}
	fieldNames := database.GetFieldNames(&submission)
	query := fmt.Sprintf(findBySubmissionIDStmt, strings.Join(fieldNames, " ,"), submission.TableName())
	err := database.Select(ctx, db, query, submissionIDs).ScanAll(&e)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return &e, nil
}

func (r *StudentSubmissionRepo) RetrieveByStudyPlanIdentities(ctx context.Context, db database.QueryExecer, identities []*StudyPlanItemIdentity) ([]*StudentSubmissionInfo, error) {
	ss := &entities.StudentSubmission{}
	fields, _ := ss.FieldMap()
	sp := &entities.StudyPlan{}
	stmt := fmt.Sprintf(`
SELECT
  ss.%s,
	sp.course_id,
	check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.start_date, isp.start_date) AS start_date,  
	check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.end_date, isp.end_date) AS end_date
FROM
  UNNEST($1::_TEXT, $2::_TEXT, $3::_TEXT) spi(student_id, study_plan_id, learning_material_id)
  JOIN %s ss
		USING(student_id, study_plan_id, learning_material_id)
  JOIN master_study_plan msp
		USING(study_plan_id, learning_material_id)
	LEFT JOIN individual_study_plan isp
		USING(student_id, study_plan_id, learning_material_id)
  JOIN %s sp
		USING(study_plan_id)
`, strings.Join(fields, ", ss."), ss.TableName(), sp.TableName())

	studyPlanIDs, lmIDs, studentIDs := make([]pgtype.Text, 0), make([]pgtype.Text, 0), make([]pgtype.Text, 0)
	for _, identity := range identities {
		studyPlanIDs = append(studyPlanIDs, identity.StudyPlanID)
		lmIDs = append(lmIDs, identity.LearningMaterialID)
		studentIDs = append(studentIDs, identity.StudentID)
	}

	rows, err := db.Query(ctx, stmt, studentIDs, studyPlanIDs, lmIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	submissions := make([]*StudentSubmissionInfo, 0)
	for rows.Next() {
		submission := &StudentSubmissionInfo{}
		_, values := submission.FieldMap()
		values = append(values, &submission.CourseID, &submission.StartDate, &submission.EndDate)
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		submissions = append(submissions, submission)
	}

	return submissions, nil
}

func GetListV2Statement(filter *StudentSubmissionFilter) (string, []interface{}) {
	// common args
	args := []interface{}{
		&filter.OffsetID,
		&filter.StudentIDs,
		&filter.Statuses,
		&filter.CreatedAt,
		epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String(),
		&filter.AssignmentName,
	}

	var joinTables, cond strings.Builder

	// common tables
	joinTables.WriteString(` JOIN assignments a ON a.assignment_id = sls.assignment_id
		JOIN study_plans sp ON sp.study_plan_id = sls.study_plan_id
		JOIN course_students cs ON cs.course_id = sp.course_id AND cs.student_id = sls.student_id
	`)

	// common cond
	cond.WriteString(` AND sls.deleted_at IS NULL
		AND ($1::text IS NULL OR sls.student_submission_id < $1)
		AND ($2::_text IS NULL OR sls.student_id = ANY ($2))
		AND ($3::_text IS NULL OR sls.status = ANY ($3))
		AND ($4::timestamp IS NULL OR sls.created_at < $4)
		AND a.deleted_at IS NULL
		AND a.type <> $5
		AND ($6::text IS NULL OR a.name ILIKE '%' || $6 || '%')
		AND sp.deleted_at IS NULL
		AND cs.deleted_at IS NULL
	`)

	// get by start_date
	if filter.StartDate.Status != pgtype.Null && filter.EndDate.Status != pgtype.Null {
		fmt.Fprintf(&cond, ` AND
			EXISTS(SELECT 1
					 FROM study_plan_items spi
					WHERE spi.deleted_at IS NULL
					  AND spi.study_plan_item_id = sls.study_plan_item_id
					  AND spi.start_date BETWEEN $%d AND $%d)
		`, len(args)+1, len(args)+2)
		args = append(args, &filter.StartDate, &filter.EndDate)
	}

	// get by course_id
	if filter.CourseID.Status != pgtype.Null {
		// if location_id is empty, get data on course_students table
		if len(filter.LocationIDs.Elements) == 0 {
			fmt.Fprintf(&cond, ` AND cs.course_id = $%d`, len(args)+1)
			args = append(args, &filter.CourseID)
		} else { // get data on course_students_access_paths table
			fmt.Fprintf(&cond, ` AND
				EXISTS(SELECT 1
						 FROM course_students_access_paths csap
						WHERE csap.deleted_at IS NULL
						  AND csap.course_student_id = cs.course_student_id
						  AND csap.course_id = $%d
						  AND csap.location_id = ANY ($%d))
			`, len(args)+1, len(args)+2)
			args = append(args, &filter.CourseID, &filter.LocationIDs)
		}

		// get by class ids
		if len(filter.ClassIDs.Elements) != 0 {
			fmt.Fprintf(&cond, ` AND
				EXISTS(SELECT 1
						 FROM class_students cls
						WHERE cls.deleted_at IS NULL
						  AND cls.student_id = sls.student_id
						  AND cls.class_id = ANY ($%d))
			`, len(args)+1)
			args = append(args, &filter.ClassIDs)
		}
	} else if len(filter.LocationIDs.Elements) != 0 { // get all matched courses on course_students_access_paths
		fmt.Fprintf(&cond, ` AND
			EXISTS(SELECT 1
					 FROM course_students_access_paths csap
					WHERE csap.deleted_at IS NULL
					  AND csap.course_student_id = cs.course_student_id
					  AND csap.location_id = ANY ($%d))
		`, len(args)+1)
		args = append(args, &filter.LocationIDs)
	}

	// get by student name
	if filter.StudentName.Status != pgtype.Null {
		fmt.Fprintf(&cond, ` AND
			EXISTS(SELECT 1
					 FROM users u
					WHERE u.deleted_at IS NULL
					  AND u.user_id = sls.student_id
					  AND u.name ILIKE '%%' || $%d || '%%')
		`, len(args)+1)
		args = append(args, &filter.StudentName)
	}

	e := &entities.StudentSubmission{}
	stmt := fmt.Sprintf("SELECT sls.%s FROM student_latest_submissions sls %s WHERE 1 = 1 %s ORDER BY sls.student_submission_id DESC LIMIT %d;",
		strings.Join(database.GetFieldNames(e), ", sls."),
		joinTables.String(),
		cond.String(),
		filter.Limit)

	return stmt, args
}
