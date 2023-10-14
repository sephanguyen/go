package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type ExamLOSubmissionRepo struct{}

type StudyPlanItemIdentity struct {
	StudentID          pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
}

func (r *ExamLOSubmissionRepo) ListByStudyPlanItemIdentities(ctx context.Context, db database.QueryExecer, studyPlanItemIdentities []*StudyPlanItemIdentity) ([]*entities.ExamLOSubmission, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.ListByStudyPlanItemIdentities")
	defer span.End()
	elss := &entities.ExamLOSubmissions{}
	els := &entities.ExamLOSubmission{}

	args := make([]interface{}, 0, 3*len(studyPlanItemIdentities))
	for _, studyPlanItemIdentity := range studyPlanItemIdentities {
		args = append(args, studyPlanItemIdentity.StudentID, studyPlanItemIdentity.StudyPlanID, studyPlanItemIdentity.LearningMaterialID)
	}

	var placeHolders string

	for i := 0; i < len(studyPlanItemIdentities); i++ {
		placeHolders += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		if i != len(studyPlanItemIdentities)-1 {
			placeHolders += ", "
		}
	}
	placeHolders = "(" + placeHolders + ")"

	query := fmt.Sprintf("SELECT %s FROM %s WHERE (student_id, study_plan_id, learning_material_id) IN %s AND deleted_at IS NULL", strings.Join(database.GetFieldNames(els), ", "), els.TableName(), placeHolders)
	if err := database.Select(ctx, db, query, args...).ScanAll(elss); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return elss.Get(), nil
}

const listExamLOSubmissionWithDatesStmtTpl = `
	SELECT els.%s, 
		COALESCE(isp.start_date, msp.start_date) AS start_date,
		COALESCE(isp.end_date, msp.end_date) AS end_date,
		COALESCE(isp.available_from, msp.available_from) AS available_from,
		COALESCE(isp.available_to, msp.available_to) AS available_to
	FROM exam_lo_submission els
	LEFT JOIN master_study_plan msp ON
		els.study_plan_id = msp.study_plan_id AND els.learning_material_id = msp.learning_material_id
	LEFT JOIN individual_study_plan isp ON
		els.study_plan_id = isp.study_plan_id AND els.student_id = isp.student_id AND els.learning_material_id = isp.learning_material_id
	WHERE 
	(els.student_id, els.study_plan_id, els.learning_material_id) IN %s 
	AND els.deleted_at IS NULL
	AND msp.deleted_at IS NULL
	AND isp.deleted_at IS NULL
`

func (r *ExamLOSubmissionRepo) ListExamLOSubmissionWithDates(ctx context.Context, db database.QueryExecer, studyPlanItemIdentities []*StudyPlanItemIdentity) (res []*ExtendedExamLOSubmission, _ error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.ListExamLOSubmissionWithDates")
	defer span.End()
	els := &entities.ExamLOSubmission{}

	args := make([]interface{}, 0, 3*len(studyPlanItemIdentities))
	for _, studyPlanItemIdentity := range studyPlanItemIdentities {
		args = append(args, studyPlanItemIdentity.StudentID, studyPlanItemIdentity.StudyPlanID, studyPlanItemIdentity.LearningMaterialID)
	}

	var placeHolders string
	for i := 0; i < len(studyPlanItemIdentities); i++ {
		placeHolders += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		if i != len(studyPlanItemIdentities)-1 {
			placeHolders += ", "
		}
	}
	placeHolders = "(" + placeHolders + ")"

	listStmt := fmt.Sprintf(listExamLOSubmissionWithDatesStmtTpl, strings.Join(database.GetFieldNames(els), ",els."), placeHolders)

	rows, err := db.Query(ctx, listStmt, args...)
	if err != nil {
		return nil, fmt.Errorf("ExamLOSubmissionRepo.ListExamLOSubmissionWithDates.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		e := &ExtendedExamLOSubmission{}
		fields := database.GetScanFields(&e.ExamLOSubmission, database.GetFieldNames(&e.ExamLOSubmission))
		fields = append(fields, &e.StartDate, &e.EndDate, &e.AvailableFrom, &e.AvailableTo)
		if err := rows.Scan(fields...); err != nil {
			return nil, fmt.Errorf("ExamLOSubmissionRepo.ListExamLOSubmissionWithDates.Scan: %w", err)
		}

		res = append(res, e)
	}

	return res, nil
}

type ExtendedExamLOSubmission struct {
	entities.ExamLOSubmission
	StartDate     pgtype.Timestamptz
	EndDate       pgtype.Timestamptz
	AvailableFrom pgtype.Timestamptz
	AvailableTo   pgtype.Timestamptz
	CourseID      pgtype.Text
	CorrectorID   pgtype.Text
}

// ExamLOSubmissionFilter used in List
type ExamLOSubmissionFilter struct {
	Limit uint

	OffsetID,
	CourseID,
	StudentName,
	CorrectorID,
	ExamName,
	SubmissionID pgtype.Text

	StudentIDs,
	Statuses,
	LocationIDs,
	ClassIDs pgtype.TextArray

	StartDate,
	EndDate,
	SubmittedStartDate,
	SubmittedEndDate,
	UpdatedStartDate,
	UpdatedEndDate,
	CreatedAt pgtype.Timestamptz
}

const listExamLOSubmissionsStmtTpl = `
SELECT *
  FROM (
    SELECT DISTINCT ON (student_id, study_plan_id, learning_material_id)
        els.%s,
        COALESCE(isp.start_date, msp.start_date) AS start_date,
        COALESCE(isp.end_date, msp.end_date) AS end_date,
        sp.course_id AS course_id,
		am.teacher_id AS corrector_id
    FROM
        exam_lo_submission els
    INNER JOIN study_plans sp ON
        els.study_plan_id = sp.study_plan_id
    INNER JOIN course_students ON
        sp.course_id = course_students.course_id AND els.student_id = course_students.student_id
    %s
    LEFT JOIN master_study_plan msp ON
        els.study_plan_id = msp.study_plan_id AND els.learning_material_id = msp.learning_material_id
    LEFT JOIN individual_study_plan isp ON
        els.study_plan_id = isp.study_plan_id AND els.student_id = isp.student_id AND
        els.learning_material_id = isp.learning_material_id
	LEFT JOIN allocate_marker am ON am.learning_material_id =  els.learning_material_id and am.student_id = els.student_id and am.study_plan_id = els.study_plan_id
    WHERE
		($13::text IS NULL OR els.submission_id = $13)
        AND ($4::text IS NULL OR sp.course_id = $4)
		AND ($3::_text IS NULL OR els.student_id = ANY($3))
        AND ($7::_text IS NULL OR els.status = ANY($7))
        AND ($5::text IS NULL OR els.submission_id < $5)
        AND ($6::timestamp IS NULL OR els.created_at < $6)
		AND ($8::text is NULL OR am.teacher_id = $8)
        AND (isp.study_plan_id IS NULL OR isp.start_date BETWEEN $1 AND $2)
        AND (isp.study_plan_id IS NOT NULL OR msp.start_date BETWEEN $1 AND $2)
		AND ($9::timestamp IS NULL OR $10::timestamp IS NULL OR els.created_at BETWEEN $9 AND $10)
		AND ($11::timestamp IS NULL OR $12::timestamp IS NULL OR els.updated_at BETWEEN $11 AND $12)
        AND el.manual_grading IS TRUE
        AND msp.deleted_at IS NULL
        AND isp.deleted_at IS NULL
        AND sp.deleted_at IS NULL
        AND els.deleted_at IS NULL
    ORDER BY els.student_id, els.study_plan_id, els.learning_material_id, els.created_at DESC, els.submission_id DESC
) as tmp
ORDER BY tmp.created_at DESC, tmp.submission_id desc
LIMIT %d;
`
const studentNameFilterJoin = `
	JOIN filter_rls_search_name_user_fn($%d::text) AS usf ON
		els.student_id = usf.user_id`

const examNameFilterJoin = `
	JOIN filter_rls_search_name_exam_lo_fn($%d::text) AS el ON
		els.learning_material_id = el.learning_material_id`

const examJoin = `
	JOIN exam_lo el ON
		els.learning_material_id = el.learning_material_id`

func (r *ExamLOSubmissionRepo) List(ctx context.Context, db database.QueryExecer, filter *ExamLOSubmissionFilter) (res []*ExtendedExamLOSubmission, _ error) {
	args := []interface{}{
		&filter.StartDate,
		&filter.EndDate,
		&filter.StudentIDs,
		&filter.CourseID,
		&filter.OffsetID,
		&filter.CreatedAt,
		&filter.Statuses,
		&filter.CorrectorID,
		&filter.SubmittedStartDate,
		&filter.SubmittedEndDate,
		&filter.UpdatedStartDate,
		&filter.UpdatedEndDate,
		&filter.SubmissionID,
	}

	var joinQueries []string
	if filter.StudentName.Status == pgtype.Present {
		args = append(args, &filter.StudentName)
		joinQueries = append(joinQueries, fmt.Sprintf(studentNameFilterJoin, len(args)))
	}
	if filter.ExamName.Status == pgtype.Present {
		args = append(args, &filter.ExamName)
		joinQueries = append(joinQueries, fmt.Sprintf(examNameFilterJoin, len(args)))
	} else {
		// avoid double JOIN exam_lo table
		joinQueries = append(joinQueries, examJoin)
	}

	e := &entities.ExamLOSubmission{}
	listStmt := fmt.Sprintf(listExamLOSubmissionsStmtTpl, strings.Join(database.GetFieldNames(e), ",els."), strings.Join(joinQueries, "\n"), filter.Limit)
	rows, err := db.Query(ctx, listStmt, args...)
	if err != nil {
		return nil, fmt.Errorf("ExamLOSubmissionRepo.List.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		e := &ExtendedExamLOSubmission{}
		fields := database.GetScanFields(e, database.GetFieldNames(e))
		fields = append(fields, &e.StartDate, &e.EndDate, &e.CourseID, &e.CorrectorID)
		if err := rows.Scan(fields...); err != nil {
			return nil, fmt.Errorf("ExamLOSubmissionRepo.List.Scan: %w", err)
		}

		res = append(res, e)
	}

	return res, nil
}

type GetExamLOSubmissionArgs struct {
	SubmissionID      pgtype.Text
	ShuffledQuizSetID pgtype.Text
}

func (r *ExamLOSubmissionRepo) Get(ctx context.Context, db database.QueryExecer, args *GetExamLOSubmissionArgs) (*entities.ExamLOSubmission, error) {
	examLOSubmission := &entities.ExamLOSubmission{}

	stmt := `SELECT %s 
	FROM %s 
	WHERE deleted_at IS NULL 
	AND ($1::TEXT IS NULL OR submission_id = $1)
	AND ($2::TEXT IS NULL OR shuffled_quiz_set_id = $2);`

	stmt = fmt.Sprintf(stmt, strings.Join(database.GetFieldNames(examLOSubmission), ","), examLOSubmission.TableName())
	if err := database.Select(ctx, db, stmt, args.SubmissionID, args.ShuffledQuizSetID).ScanOne(examLOSubmission); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return examLOSubmission, nil
}

func (r *ExamLOSubmissionRepo) GetTotalGradedPoint(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (pgtype.Int4, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.GetTotalGradedPoint")
	defer span.End()

	var result pgtype.Int4

	stmt := `
    SELECT COALESCE(SUM(X.point), 0)::INT AS total_graded_point
      FROM (SELECT DISTINCT ON (Y.quiz_id) Y.quiz_id, Y.point
              FROM (SELECT quiz_id, point, created_at
                      FROM exam_lo_submission_answer
                     WHERE submission_id = $1::TEXT
                       AND deleted_at IS NULL
                    UNION ALL
                    SELECT quiz_id, point, created_at
                      FROM exam_lo_submission_score
                     WHERE submission_id = $1::TEXT
                       AND deleted_at IS NULL
                   ) Y
             ORDER BY Y.quiz_id, Y.created_at DESC
           ) X;
	`

	if err := database.Select(ctx, db, stmt, submissionID).ScanFields(&result); err != nil {
		return database.Int4(0), fmt.Errorf("database.Select: %w", err)
	}

	return result, nil
}

type ExamLOSubmissionWithGrade struct {
	entities.ExamLOSubmission
	TotalGradePoint pgtype.Int2
}

const listTotalGradePointsStmtTmpl = `
	SELECT exam_lo_submission.%s, els.graded_point AS total_graded_point
	FROM exam_lo_submission
	JOIN get_exam_lo_scores() els ON exam_lo_submission.submission_id = els.submission_id
	WHERE els.submission_id = ANY($1::_TEXT)`

func (r *ExamLOSubmissionRepo) ListTotalGradePoints(ctx context.Context, db database.QueryExecer, submissionIDs pgtype.TextArray) (res []*ExamLOSubmissionWithGrade, _ error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.ListTotalGradePoints")
	defer span.End()

	e := &entities.ExamLOSubmission{}
	listStmt := fmt.Sprintf(listTotalGradePointsStmtTmpl, strings.Join(database.GetFieldNames(e), ",exam_lo_submission."))

	rows, err := db.Query(ctx, listStmt, submissionIDs)
	if err != nil {
		return nil, fmt.Errorf("ExamLOSubmissionRepo.ListTotalGradePoints.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		ewg := &ExamLOSubmissionWithGrade{}
		_, fields := ewg.ExamLOSubmission.FieldMap()
		fields = append(fields, &ewg.TotalGradePoint)
		if err := rows.Scan(fields...); err != nil {
			return nil, fmt.Errorf("ExamLOSubmissionRepo.ListTotalGradePoints.Scan: %w", err)
		}

		res = append(res, ewg)
	}

	return res, nil
}

const deleteSubmissionQuery = `UPDATE %s SET deleted_at = now() WHERE submission_id = $1::TEXT AND deleted_at IS NULL`

func (r *ExamLOSubmissionRepo) Delete(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.DeleteSubmissionBySubmissionId")
	defer span.End()

	e := entities.ExamLOSubmission{}
	query := fmt.Sprintf(deleteSubmissionQuery, e.TableName())
	cmdTag, err := db.Exec(ctx, query, submissionID)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func (r *ExamLOSubmissionRepo) GetLatestSubmissionID(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (pgtype.Text, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.GetLatestSubmissionID")
	defer span.End()

	var result pgtype.Text

	stmt := `
    SELECT submission_id
      FROM exam_lo_submission ELM
           JOIN (SELECT student_id, study_plan_id, learning_material_id
                   FROM exam_lo_submission
                  WHERE submission_id = $1::TEXT) TMP
               USING (student_id, study_plan_id, learning_material_id)
     WHERE deleted_at IS NULL
    ORDER BY ELM.created_at DESC
    LIMIT 1;
	`

	if err := database.Select(ctx, db, stmt, submissionID).ScanFields(&result); err != nil {
		return result, fmt.Errorf("database.Select: %w", err)
	}

	return result, nil
}

func (r *ExamLOSubmissionRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmission) error {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.Update")
	defer span.End()

	if _, err := database.UpdateFields(ctx, e, db.Exec, "submission_id", []string{
		"status",
		"result",
		"teacher_feedback",
		"teacher_id",
		"marked_at",
		"updated_at",
	}); err != nil {
		return fmt.Errorf("database.UpdateFields: %w", err)
	}

	return nil
}

func (r *ExamLOSubmissionRepo) Insert(ctx context.Context, db database.QueryExecer, e *entities.ExamLOSubmission) error {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.Insert")
	defer span.End()

	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}

	return nil
}

func (r *ExamLOSubmissionRepo) GetLatestExamLOSubmission(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text) (entities.ExamLOSubmission, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.GetLatestExamLOSubmission")
	defer span.End()

	e := entities.ExamLOSubmission{}
	stmt := fmt.Sprintf(`
    SELECT %s
    FROM %s ELM
    JOIN (SELECT student_id, study_plan_id, learning_material_id
        FROM exam_lo_submission
        WHERE submission_id = $1::TEXT) TMP
    USING (student_id, study_plan_id, learning_material_id)
    WHERE deleted_at IS NULL
    ORDER BY ELM.created_at DESC
	LIMIT 1;`, strings.Join(database.GetFieldNames(&e), ","), e.TableName())

	if err := database.Select(ctx, db, stmt, submissionID).ScanOne(&e); err != nil {
		return entities.ExamLOSubmission{}, fmt.Errorf("database.Select: %w", err)
	}
	return e, nil
}

func (r *ExamLOSubmissionRepo) UpdateExamSubmissionTotalPointsWithResult(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text, newTotalPoints pgtype.Int4, newExamResult pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.UpdateExamSubmissionTotalPointsWithResult")
	defer span.End()

	e := entities.ExamLOSubmission{}

	// somehow we need to trigger update_max_score_exam_lo_once_exam_lo_submission_status_change so we run this query
	// where we just set the status to the current value which forces an update causing the trigger to fire
	stmt := fmt.Sprintf(`UPDATE %s SET total_point = $1, result = $2, status = status, updated_at = now() WHERE submission_id = $3`, e.TableName())

	_, err := db.Exec(ctx, stmt, newTotalPoints, newExamResult, submissionID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *ExamLOSubmissionRepo) UpdateExamSubmissionTotalPoints(ctx context.Context, db database.QueryExecer, submissionID pgtype.Text, newTotalPoints pgtype.Int4) error {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.UpdateExamSubmissionTotalPoints")
	defer span.End()

	e := entities.ExamLOSubmission{}
	// see UpdateExamSubmissionTotalPointsWithResult on why we update status = status
	stmt := fmt.Sprintf(`UPDATE %s SET total_point = $1, status = status, updated_at = now() WHERE submission_id = $2`, e.TableName())

	_, err := db.Exec(ctx, stmt, newTotalPoints, submissionID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

type BulkUpdateApproveRejectArgs struct {
	SubmissionIDs pgtype.TextArray
	Status        pgtype.Text
	LastAction    pgtype.Text
	LastActionAt  pgtype.Timestamptz
	LastActionBy  pgtype.Text
	StatusCond    pgtype.TextArray
	UpdatedAt     pgtype.Timestamptz
}

func (r *ExamLOSubmissionRepo) BulkUpdateApproveReject(ctx context.Context, db database.QueryExecer, args *BulkUpdateApproveRejectArgs) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.BulkUpdateApproveReject")
	defer span.End()

	stmt := `
    UPDATE exam_lo_submission
       SET updated_at = $1,
           status = $2,
           last_action = $3,
           last_action_at = $4,
           last_action_by = $5,
           result = sub.result
      FROM (SELECT submission_id,
                   CASE
                    WHEN $2 = 'SUBMISSION_STATUS_RETURNED' THEN
                       CASE
                        WHEN info.grade_to_pass IS NULL THEN 'EXAM_LO_SUBMISSION_COMPLETED'
                        ELSE CASE
                              WHEN info.total_graded_point >= info.grade_to_pass THEN 'EXAM_LO_SUBMISSION_PASSED'
                              ELSE 'EXAM_LO_SUBMISSION_FAILED'
                             END
                       END
                    WHEN $2 = 'SUBMISSION_STATUS_IN_PROGRESS' THEN 'EXAM_LO_SUBMISSION_WAITING_FOR_GRADE'
                   END AS result
              FROM (SELECT els.submission_id,
                           (SELECT grade_to_pass FROM exam_lo WHERE learning_material_id = els.learning_material_id) AS grade_to_pass,
                           sum(coalesce(elss.point, elsa.point))::smallint AS total_graded_point
                      FROM exam_lo_submission els
                          JOIN exam_lo_submission_answer elsa USING (submission_id)
                          LEFT JOIN exam_lo_submission_score elss USING (submission_id, quiz_id)
                    WHERE els.deleted_at IS NULL
                      AND submission_id = ANY($6::TEXT[])
                      AND status = ANY($7::TEXT[])
                    GROUP BY els.submission_id) info
           ) sub
    WHERE sub.submission_id = exam_lo_submission.submission_id;
	`

	cmdTag, err := db.Exec(ctx, stmt, args.UpdatedAt, args.Status, args.LastAction, args.LastActionAt, args.LastActionBy,
		args.SubmissionIDs, args.StatusCond)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}

	return int(cmdTag.RowsAffected()), nil
}

func (r *ExamLOSubmissionRepo) GetInvalidIDsByBulkApproveReject(ctx context.Context, db database.QueryExecer, submissionIDs pgtype.TextArray, statusCond pgtype.TextArray) (pgtype.TextArray, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLOSubmissionRepo.GetInvalidIDsByBulkApproveReject")
	defer span.End()

	var results pgtype.TextArray

	stmt := `
    SELECT ARRAY (
        SELECT submission_id
          FROM exam_lo_submission
         WHERE deleted_at IS NULL
           AND submission_id = ANY($1::TEXT[])
           AND status <> ANY($2::TEXT[])
    );
	`

	if err := database.Select(ctx, db, stmt, submissionIDs, statusCond).ScanFields(&results); err != nil {
		return results, fmt.Errorf("database.Select: %w", err)
	}

	return results, nil
}
