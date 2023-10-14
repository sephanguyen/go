package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

// StudentSubmissionGradeRepo works with "student_submissions" table
type StudentSubmissionGradeRepo struct{}

// Create create new entities with newly generated ID
func (r *StudentSubmissionGradeRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentSubmissionGrade) error {
	_ = e.ID.Set(idutil.ULIDNow())
	e.Now()

	_, err := database.Insert(ctx, e, db.Exec)
	if err != nil {
		return err
	}

	return nil
}

const retrieveGradeByIDsStmtTpl = `SELECT
	%s
FROM
	student_submission_grades
WHERE
	student_submission_grade_id = ANY($1);`

// RetrieveByIDs returns grades match IDs array
func (r *StudentSubmissionGradeRepo) RetrieveByIDs(ctx context.Context, db database.QueryExecer,
	ids pgtype.TextArray) (entities.StudentSubmissionGrades, error) {
	query := fmt.Sprintf(retrieveGradeByIDsStmtTpl, strings.Join(entities.StudentSubmissionGradeFields, ","))

	results := make(entities.StudentSubmissionGrades, 0, len(ids.Elements))
	if err := database.Select(ctx, db, query, &ids).ScanAll(&results); err != nil {
		return nil, err
	}

	return results, nil
}

const checkSubmissionsStmt = `SELECT
	COUNT(ssg.student_submission_grade_id)
FROM
	student_submission_grades ssg
JOIN student_submissions ss ON
	ss.student_submission_id = ssg.student_submission_id
	AND ss.student_id = $1
	AND ss.status = 'SUBMISSION_STATUS_RETURNED'
WHERE
	ssg.student_submission_grade_id = ANY ($2)
`

// CheckSubmissions check if the submissions are assigned to student match the status
func (r *StudentSubmissionGradeRepo) CheckSubmissions(ctx context.Context, db database.QueryExecer,
	gradeIDs pgtype.TextArray, studentID, status pgtype.Text) (bool, error) {
	var count pgtype.Int8
	err := database.Select(ctx, db, checkSubmissionsStmt, &studentID, &gradeIDs).ScanFields(&count)
	if err != nil {
		return false, err
	}

	return int(count.Int) == len(gradeIDs.Elements), nil
}

const findBySubmissionIDsStmt = `SELECT ssg.%s FROM student_submission_grades ssg 
JOIN student_submissions ss ON 
	ss.student_submission_grade_id = ssg.student_submission_grade_id
WHERE ss.student_submission_id = ANY($1::_TEXT)
 `

func (r *StudentSubmissionGradeRepo) FindBySubmissionIDs(ctx context.Context, db database.QueryExecer, submissionIDs pgtype.TextArray) (*entities.StudentSubmissionGrades, error) {
	e := &entities.StudentSubmissionGrade{}
	fields, _ := e.FieldMap()
	es := &entities.StudentSubmissionGrades{}
	err := database.Select(ctx, db, fmt.Sprintf(findBySubmissionIDsStmt, strings.Join(fields, ",ssg.")), &submissionIDs).ScanAll(es)
	if err != nil {
		return nil, err
	}
	return es, nil
}

func (r *StudentSubmissionGradeRepo) BulkImport(ctx context.Context, db database.QueryExecer, grades []*entities.StudentSubmissionGrade) error {
	e := entities.StudentSubmissionGrade{}
	fields, _ := e.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))
	queueFn := func(b *pgx.Batch, e *entities.StudentSubmissionGrade) {
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			e.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fields)...)
	}

	b := &pgx.Batch{}
	var d pgtype.Timestamptz
	err := d.Set(time.Now())
	if err != nil {
		return fmt.Errorf("cannot set time for grade: %w", err)
	}

	for _, each := range grades {
		err = each.ID.Set(idutil.ULIDNow())
		if err != nil {
			return fmt.Errorf("cannot set id for grade: %w", err)
		}
		each.CreatedAt = d
		each.UpdatedAt = d
		queueFn(b, each)
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

type StudentSubmissionGradeInfo struct {
	StudentSubmissionGradeID pgtype.Text
	StudentSubmissionID      pgtype.Text
	AssignmentID             pgtype.Text
	AssignmentName           pgtype.Text
	StudentID                pgtype.Text
	StudyPlanItemID          pgtype.Text
	CourseID                 pgtype.Text
	StudyPlanID              pgtype.Text
	LearningMaterialID       pgtype.Text
}

func (r *StudentSubmissionGradeRepo) RetrieveInfoByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*StudentSubmissionGradeInfo, error) {
	e := &entities.StudentSubmissionGrade{}
	sse := &entities.StudentSubmission{}
	ae := &entities.Assignment{}
	spie := &entities.StudyPlanItem{}
	stmt := fmt.Sprintf(`
		SELECT 	ssg.student_submission_grade_id, 
				a.assignment_id, 
				a.name as assignment_name , 
				ss.student_submission_id, 
				ss.student_id,
				spi.study_plan_item_id, 
				content_structure ->> 'course_id' as course_id,
				ss.study_plan_id,
				ss.learning_material_id
		FROM %s AS ssg
		JOIN %s As ss
		USING(student_submission_id)
		JOIN %s AS a
		USING(assignment_id)
		JOIN %s as spi
		USING(study_plan_item_id)
		WHERE ssg.student_submission_grade_id = ANY($1::TEXT[])
		AND ssg.deleted_at IS NULL
		AND ss.deleted_at IS NULL
		AND a.deleted_at IS NULL
		and spi.deleted_at IS NULL
	`, e.TableName(), sse.TableName(), ae.TableName(), spie.TableName())

	rows, err := db.Query(ctx, stmt, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*StudentSubmissionGradeInfo
	for rows.Next() {
		var info StudentSubmissionGradeInfo
		if err := rows.Scan(
			&info.StudentSubmissionGradeID,
			&info.AssignmentID,
			&info.AssignmentName,
			&info.StudentSubmissionID,
			&info.StudentID,
			&info.StudyPlanItemID,
			&info.CourseID,
			&info.StudyPlanID,
			&info.LearningMaterialID,
		); err != nil {
			return nil, err
		}
		result = append(result, &info)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err(): %w", rows.Err())
	}
	return result, nil
}
