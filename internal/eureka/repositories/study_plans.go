package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.opencensus.io/trace"
	"go.uber.org/multierr"
)

type StudyPlanRepo struct {
}

func (r *StudyPlanRepo) Insert(ctx context.Context, db database.QueryExecer, studyPlan *entities.StudyPlan) (pgtype.Text, error) {
	now := timeutil.Now()
	err := multierr.Combine(
		studyPlan.CreatedAt.Set(now),
		studyPlan.UpdatedAt.Set(now),
	)
	var id pgtype.Text

	if err != nil {
		return id, fmt.Errorf("error set time: %w", err)
	}

	if err := database.InsertReturning(ctx, studyPlan, db, "study_plan_id", &id); err != nil {
		return id, fmt.Errorf("error insert: %w", err)
	}

	return id, nil
}

func (r *StudyPlanRepo) QueueUpsertStudyPlan(b *pgx.Batch, item *entities.StudyPlan) {
	fieldNames := database.GetFieldNames(item)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`INSERT INTO %s (%s)VALUES (%s) 
	ON CONFLICT ON CONSTRAINT study_plans_pk DO UPDATE SET
		name = $3,
		study_plan_type = $4,
		updated_at = $6,
		school_id = $8,
		deleted_at = NULL `, item.TableName(), strings.Join(fieldNames, ","), placeHolders)
	scanFields := database.GetScanFields(item, fieldNames)
	b.Queue(query, scanFields...)
}

func (r *StudyPlanRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error {
	b := &pgx.Batch{}
	for _, item := range items {
		r.QueueUpsertStudyPlan(b, item)
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

func (r *StudyPlanRepo) BulkUpdateByMaster(ctx context.Context, db database.QueryExecer, item *entities.StudyPlan) error {
	query := fmt.Sprintf(` UPDATE %s
		SET updated_at = NOW(), name = $2, track_school_progress = $3, grades = $4, status = $5
		WHERE (study_plan_id = $1 OR master_study_plan_id = $1) AND deleted_at IS NULL
		`,
		item.TableName(),
	)
	if _, err := db.Exec(ctx, query, &item.ID, &item.Name, &item.TrackSchoolProgress, &item.Grades, &item.Status); err != nil {
		return err
	}

	return nil
}

const bulkCopyStmt = `INSERT
	INTO
	study_plans (study_plan_id,
	master_study_plan_id,
	"name",
	study_plan_type,
	school_id,
	course_id,
	book_id,
	created_at,
	updated_at,
	deleted_at,
	track_school_progress,
	grades,
	status
)
SELECT
	generate_ulid() AS study_plan_id,
	$1::TEXT AS master_study_plan_id,
	name,
	study_plan_type,
	school_id,
	course_id,
	book_id,
	created_at,
	updated_at,
	deleted_at,
	track_school_progress,
	grades,
	status
FROM
	study_plans sp
WHERE
	sp.study_plan_id = $1 RETURNING study_plan_id,
	master_study_plan_id;`

func (r *StudyPlanRepo) QueueBulkCopy(b *pgx.Batch, studyPlanID pgtype.Text) {
	b.Queue(bulkCopyStmt, studyPlanID)
}

func (r *StudyPlanRepo) BulkCopy(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) ([]string, []string, error) {
	b := &pgx.Batch{}
	for _, item := range studyPlanIDs.Elements {
		r.QueueBulkCopy(b, item)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	newStudyPlanIDs := make([]string, len(studyPlanIDs.Elements))
	orgStudyPlanIDs := make([]string, len(studyPlanIDs.Elements))
	for i := 0; i < b.Len(); i++ {
		var originalStudyPlanID, newStudyPlanID string
		row := result.QueryRow()
		if err := row.Scan(&newStudyPlanID, &originalStudyPlanID); err != pgx.ErrNoRows && err != nil {
			return orgStudyPlanIDs, newStudyPlanIDs, err
		}
		newStudyPlanIDs[i] = newStudyPlanID
		orgStudyPlanIDs[i] = originalStudyPlanID
	}
	return orgStudyPlanIDs, newStudyPlanIDs, nil
}

func (r *StudyPlanRepo) FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) (*entities.StudyPlan, error) {
	studyPlan := &entities.StudyPlan{}
	fieldNames := database.GetFieldNames(studyPlan)

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE study_plan_id = $1`, strings.Join(fieldNames, ", "), studyPlan.TableName())
	err := database.Select(ctx, db, query, &studyPlanID).ScanOne(studyPlan)
	if err != nil {
		return nil, err
	}
	return studyPlan, nil
}

func (r *StudyPlanRepo) FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlan, error) {
	studyPlan := &entities.StudyPlan{}
	fieldNames := database.GetFieldNames(studyPlan)

	var plans entities.StudyPlans
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE study_plan_id = ANY($1)`, strings.Join(fieldNames, ", "), studyPlan.TableName())
	err := database.Select(ctx, db, query, &ids).ScanAll(&plans)
	if err != nil {
		return nil, err
	}
	return plans, nil
}

const findDependStudyPlanStmt = `SELECT
	study_plan_id
FROM
	%s
WHERE
	master_study_plan_id = ANY($1)
	AND deleted_at IS NULL`

func (r *StudyPlanRepo) FindDependStudyPlan(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]string, error) {
	studyPlan := &entities.StudyPlan{}

	query := fmt.Sprintf(findDependStudyPlanStmt, studyPlan.TableName())
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studyPlanIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		studyPlanIDs = append(studyPlanIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return studyPlanIDs, nil
}

func (r *StudyPlanRepo) SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	query := `UPDATE study_plans SET deleted_at = NOW() WHERE deleted_at IS NULL AND study_plan_id = ANY($1)`
	_, err := db.Exec(ctx, query, &ids)
	if err != nil {
		return err
	}
	return nil
}

type RetrieveStudyPlanByCourseArgs struct {
	CourseID      pgtype.Text
	Limit         uint32
	StudyPlanName pgtype.Text
	StudyPlanID   pgtype.Text
}

const retrieveStudyPlanByCourseIDStmtTpl = `SELECT 
	sp.%s 
FROM study_plans AS sp JOIN course_study_plans AS csp USING(study_plan_id) 
WHERE csp.course_id = $1
	AND (($2::text IS NULL AND $3::text IS NULL) OR ($2::text collate japanese_collation, $3::text) < (sp.name, sp.study_plan_id))
	AND csp.deleted_at IS NULL
	AND sp.deleted_at IS NULL
	AND sp.status = 'STUDY_PLAN_STATUS_ACTIVE'
ORDER BY sp.name collate japanese_collation, sp.study_plan_id ASC, sp.created_at DESC
LIMIT $4`

func (r *StudyPlanRepo) RetrieveByCourseID(ctx context.Context, db database.QueryExecer, args *RetrieveStudyPlanByCourseArgs) ([]*entities.StudyPlan, error) {
	e := &entities.StudyPlan{}
	var es entities.StudyPlans
	fields, _ := e.FieldMap()
	query := fmt.Sprintf(retrieveStudyPlanByCourseIDStmtTpl, strings.Join(fields, ", sp."))
	if err := database.Select(ctx, db, query, &args.CourseID, &args.StudyPlanName, &args.StudyPlanID, &args.Limit).ScanAll(&es); err != nil {
		return nil, err
	}
	return es, nil
}

type RetrieveStudyPlanIdentityResponse struct {
	StudyPlanID        pgtype.Text
	StudentID          pgtype.Text
	LearningMaterialID pgtype.Text
	StudyPlanItemID    pgtype.Text
}

func (r *StudyPlanRepo) RetrieveStudyPlanIdentity(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*RetrieveStudyPlanIdentityResponse, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanRepo.RetrieveStudyPlanIdentity")
	defer span.End()

	query := `
		SELECT study_plan_id, student_id, lm_id, study_plan_item_id
		FROM retrieve_study_plan_identity($1::_TEXT)
		`

	res := []*RetrieveStudyPlanIdentityResponse{}
	rows, err := db.Query(ctx, query, &studyPlanItemIDs)
	if err != nil {
		return nil, fmt.Errorf("StudyPlanRepo.RetrieveStudyPlanIdentity.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r RetrieveStudyPlanIdentityResponse
		if err := rows.Scan(&r.StudyPlanID, &r.StudentID, &r.LearningMaterialID, &r.StudyPlanItemID); err != nil {
			return nil, fmt.Errorf("StudyPlanRepo.RetrieveStudyPlanIdentity.Scan: %w", err)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("StudyPlanRepo.RetrieveStudyPlanIdentity.Err: %w", err)
		}
		res = append(res, &r)
	}

	return res, nil
}

// Recursive delete item and it's child items with given root study_plan_id (for now root study_plan_id belongs to courses)
/*
* @param: 	study_plan_id 	string 	(from course_study_plans)
* @return:	study_plan_ids []string	(list student_study_plans)
 */
func (r *StudyPlanRepo) RecursiveSoftDeleteStudyPlanByStudyPlanIDInCourse(ctx context.Context, db database.QueryExecer, courseStudyPlanID pgtype.Text) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudyPlanRepo.RecursiveSoftDeleteStudyPlanByStudyPlanIDInCourse")
	defer span.End()

	query := `
		WITH RECURSIVE study_plan_recurs (study_plan_id, master_study_plan_id, deleted_at) AS (
			SELECT sp1.study_plan_id,
						sp1.master_study_plan_id,
						sp1.deleted_at
			FROM study_plans sp1
			WHERE sp1.study_plan_id = $1
				AND sp1.master_study_plan_id IS NULL

			UNION ALL

			SELECT sp2.study_plan_id,
						sp2.master_study_plan_id,
						sp2.deleted_at
			FROM study_plans as sp2
							JOIN study_plan_recurs spr ON spr.study_plan_id = sp2.master_study_plan_id
		)
			UPDATE study_plans SET deleted_at = now() WHERE study_plan_id IN (SELECT spr.study_plan_id FROM study_plan_recurs AS spr) RETURNING study_plan_id
	`

	rows, err := db.Query(ctx, query, &courseStudyPlanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studyPlanIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		studyPlanIDs = append(studyPlanIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return studyPlanIDs, nil
}

type StudyPlanBook struct {
	StudyPlanID pgtype.Text
	BookID      pgtype.Text
}

func (r *StudyPlanRepo) BulkUpdateBook(ctx context.Context, db database.QueryExecer, spbs []*StudyPlanBook) error {
	e := &entities.StudyPlan{}
	queue := func(b *pgx.Batch, spb *StudyPlanBook) {
		query := fmt.Sprintf(`
		UPDATE %s
		SET book_id = $2, updated_at = NOW()
		WHERE study_plan_id = $1
		OR master_study_plan_id = $1 
		`, e.TableName())
		b.Queue(query, &spb.StudyPlanID, &spb.BookID)
	}

	b := &pgx.Batch{}
	for _, spb := range spbs {
		queue(b, spb)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(spbs); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() == 0 {
			return fmt.Errorf("course book not inserted")
		}
	}
	return nil
}

func (r *StudyPlanRepo) RetrieveMasterByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.StudyPlan, error) {
	studyPlan := &entities.StudyPlan{}
	fieldNames := database.GetFieldNames(studyPlan)
	studyPlans := entities.StudyPlans{}
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE book_id = ANY($1::TEXT[]) AND deleted_at IS NULL AND master_study_plan_id IS NULL`, strings.Join(fieldNames, ", "), studyPlan.TableName())
	err := database.Select(ctx, db, query, &bookIDs).ScanAll(&studyPlans)
	if err != nil {
		return nil, err
	}
	return studyPlans, nil
}

func (r *StudyPlanRepo) RetrieveByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.StudyPlan, error) {
	ctx, span := interceptors.StartSpan(ctx, "RetrieveByBookIDs.RetrieveByBookIDs")
	defer span.End()
	studyPlan := &entities.StudyPlan{}
	fieldNames := database.GetFieldNames(studyPlan)
	studyPlans := entities.StudyPlans{}
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE book_id = ANY($1::TEXT[]) AND deleted_at IS NULL`, strings.Join(fieldNames, ", "), studyPlan.TableName())
	err := database.Select(ctx, db, query, &bookIDs).ScanAll(&studyPlans)
	if err != nil {
		return nil, err
	}
	return studyPlans, nil
}

func (r *StudyPlanRepo) RetrieveMasterByCourseIDs(ctx context.Context, db database.QueryExecer, studyPlanType pgtype.Text, courseIDs pgtype.TextArray) ([]*entities.StudyPlan, error) {
	studyPlan := &entities.StudyPlan{}
	fieldNames := database.GetFieldNames(studyPlan)
	studyPlans := entities.StudyPlans{}
	query := fmt.Sprintf(`SELECT %s FROM %s
			 WHERE deleted_at IS NULL 
			 AND course_id = ANY($1::_TEXT) 
			 AND master_study_plan_id IS NULL 
			 AND study_plan_type = $2::TEXT`, strings.Join(fieldNames, ", "), studyPlan.TableName())
	err := database.Select(ctx, db, query, &courseIDs, &studyPlanType).ScanAll(&studyPlans)
	if err != nil {
		return nil, err
	}
	return studyPlans, nil
}

func (r *StudyPlanRepo) RetrieveCombineStudent(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.StudyPlanCombineStudentID, error) {
	ctx, span := trace.StartSpan(ctx, "StudyPlanRepo.RetrieveCombineStudent")
	defer span.End()

	sp := &entities.StudyPlan{}
	fields := database.GetFieldNames(sp)
	selectStmt := fmt.Sprintf(`SELECT sp.%s, ssp.student_id FROM study_plans AS sp
	LEFT JOIN student_study_plans ssp USING (study_plan_id)
	WHERE sp.book_id= ANY($1::_TEXT)
	AND ssp.deleted_at IS NULL
	AND sp.deleted_at IS NULL`, strings.Join(fields, ", sp."))

	rows, err := db.Query(ctx, selectStmt, &bookIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	sStudyPlans := make([]*entities.StudyPlanCombineStudentID, 0)

	for rows.Next() {
		spTemp := entities.StudyPlan{}
		var (
			studentID pgtype.Text
		)
		scanFields := database.GetScanFields(&spTemp, fields)
		scanFields = append(scanFields, &studentID)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		sStudyPlan := entities.StudyPlanCombineStudentID{
			StudyPlan: spTemp,
			StudentID: studentID,
		}
		sStudyPlans = append(sStudyPlans, &sStudyPlan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}
	return sStudyPlans, nil
}

type StudyPlanItemInfoArgs struct {
	BookIDs       pgtype.TextArray
	LoIDs         pgtype.TextArray
	AssignmentIDs pgtype.TextArray
}
type StudyPlanItemInfo struct {
	entities.StudyPlanItem

	StudyPlanID       pgtype.Text
	MasterStudyPlanID pgtype.Text
	BookID            pgtype.Text
	CourseID          pgtype.Text
}

func (r *StudyPlanRepo) RetrieveStudyPlanItemInfo(ctx context.Context, db database.QueryExecer, args StudyPlanItemInfoArgs) ([]*StudyPlanItemInfo, error) {
	spe := &entities.StudyPlan{}
	spie := &entities.StudyPlanItem{}
	fieldNames := database.GetFieldNames(spie)
	for i := range fieldNames {
		fieldNames[i] = "spi." + fieldNames[i]
	}
	stmt := fmt.Sprintf(`
	SELECT %s, sp.study_plan_id, sp.book_id, sp.course_id, sp.master_study_plan_id
	FROM %s as sp
	LEFT OUTER JOIN (
		SELECT * FROM %s
		WHERE deleted_at IS NULL
		AND ($1::TEXT[] IS NULL OR content_structure ->> 'lo_id' = ANY($1::TEXT[]))
		AND ($2::TEXT[] IS NULL OR content_structure ->> 'assignment_id' = ANY($2::TEXT[]))
	) as spi
	USING(study_plan_id)
	WHERE ($3::TEXT[] IS NULL OR sp.book_id = ANY($3::TEXT[]))
	AND sp.deleted_at IS NULL
	ORDER BY sp.master_study_plan_id DESC
	`, strings.Join(fieldNames, ", "), spe.TableName(), spie.TableName())
	rows, err := db.Query(ctx, stmt, &args.LoIDs, &args.AssignmentIDs, &args.BookIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*StudyPlanItemInfo
	for rows.Next() {
		var (
			info StudyPlanItemInfo
		)
		fields, _ := info.StudyPlanItem.FieldMap()
		if err := rows.Scan(append(database.GetScanFields(&info.StudyPlanItem, fields), &info.StudyPlanID, &info.BookID, &info.CourseID, &info.MasterStudyPlanID)...); err != nil {
			return nil, err
		}
		result = append(result, &info)
	}
	return result, nil
}

type ListIndividualStudyPlanArgs struct {
	StudentID pgtype.Text
	Limit     uint32
	Status    string
	// fields used for pagination query
	Offset             pgtype.Timestamptz
	LearningMaterialID pgtype.Text
	CourseIDs          pgtype.TextArray
}

type IndividualStudyPlanItem struct {
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	StudentID          pgtype.Text
	AvailableFrom      pgtype.Timestamptz
	AvailableTo        pgtype.Timestamptz
	StartDate          pgtype.Timestamptz
	EndDate            pgtype.Timestamptz
	Status             pgtype.Text
	SchoolDate         pgtype.Timestamptz
	Score              pgtype.Int2
	CompletedAt        pgtype.Timestamptz
	Type               pgtype.Text
}

const listActiveStudyPlanItems = `
SELECT ispi.%s FROM list_individual_study_plan_item() ispi
	JOIN study_plans sp using(study_plan_id)
WHERE
	ispi.student_id = $1
	AND ispi.start_date < NOW()
	AND (($2::timestamp IS NULL OR $3::text IS NULL) OR ((ispi.start_date, ispi.learning_material_id) > ($2, $3)))
	AND (ispi.end_date IS NULL OR ispi.end_date >= NOW())
	AND ispi.completed_at IS NULL
    AND sp.course_id = ANY($5::TEXT[])
	AND sp.deleted_at IS NULL
ORDER BY
	ispi.start_date ASC,
	ispi.lm_display_order ASC,
	ispi.learning_material_id ASC
LIMIT $4;
`
const listCompletedStudyPlanItems = `
SELECT ispi.%s FROM list_individual_study_plan_item() ispi
	JOIN study_plans sp using(study_plan_id)
WHERE
	ispi.student_id = $1
	AND ispi.start_date < NOW()
	AND (($2::timestamp IS NULL OR $3::text IS NULL) OR ((ispi.start_date, ispi.learning_material_id) < ($2, $3)))
	AND (ispi.end_date IS NULL OR ispi.end_date >= NOW())
	AND ispi.completed_at IS NOT NULL
    AND sp.course_id = ANY($5::TEXT[])
	AND sp.deleted_at IS NULL
ORDER BY
	ispi.start_date DESC,
	ispi.learning_material_id DESC
LIMIT $4;
`

const listOverdueStudyPlanItems = `
SELECT ispi.%s FROM list_individual_study_plan_item() ispi
	JOIN study_plans sp using(study_plan_id)
WHERE
    ispi.student_id = $1
	AND (ispi.start_date, ispi.learning_material_id) < ($2, $3)
	AND ispi.end_date < NOW()
	AND ispi.completed_at IS NULL
    AND sp.course_id = ANY($5::TEXT[])
	AND sp.deleted_at IS NULL
ORDER BY
	ispi.start_date DESC,
	ispi.learning_material_id DESC
LIMIT $4;
`

func (r *StudyPlanRepo) ListIndividualStudyPlanItems(ctx context.Context, db database.QueryExecer, args *ListIndividualStudyPlanArgs) ([]*IndividualStudyPlanItem, error) {
	var query string
	var selectFieldsItem = []string{
		"study_plan_id", "learning_material_id", "student_id", "available_from", "available_to", "start_date", "end_date", "status", "school_date", "scorce", "type", "completed_at",
	}
	switch args.Status {
	case "active":
		query = fmt.Sprintf(listActiveStudyPlanItems, strings.Join(selectFieldsItem, ", ispi."))
	case "completed":
		query = fmt.Sprintf(listCompletedStudyPlanItems, strings.Join(selectFieldsItem, ", ispi."))
	case "overdue":
		query = fmt.Sprintf(listOverdueStudyPlanItems, strings.Join(selectFieldsItem, ", ispi."))
	}

	rows, err := db.Query(ctx, query, &args.StudentID, &args.Offset, &args.LearningMaterialID, &args.Limit, &args.CourseIDs)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	defer rows.Close()
	var res []*IndividualStudyPlanItem
	for rows.Next() {
		item := new(IndividualStudyPlanItem)
		if err := rows.Scan(
			&item.StudyPlanID,
			&item.LearningMaterialID,
			&item.StudentID,
			&item.AvailableFrom,
			&item.AvailableTo,
			&item.StartDate,
			&item.EndDate,
			&item.Status,
			&item.SchoolDate,
			&item.Score,
			&item.Type,
			&item.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("database.Scan: %w", err)
		}
		res = append(res, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db.Error: %w", err)
	}
	return res, nil
}

type StudentStudyPlanItem struct {
	LearningMaterialID  pgtype.Text
	TopicID             pgtype.Text
	Name                pgtype.Text
	Type                pgtype.Text
	DisplayOrder        pgtype.Int2
	CreatedAt           pgtype.Timestamptz
	UpdatedAt           pgtype.Timestamptz
	DeletedAt           pgtype.Timestamptz
	StartDate           pgtype.Timestamptz
	EndDate             pgtype.Timestamptz
	CompletedAt         pgtype.Timestamptz
	SchoolDate          pgtype.Timestamptz
	AvailableFrom       pgtype.Timestamptz
	AvailableTo         pgtype.Timestamptz
	StudyPlanItemStatus pgtype.Text
	BookID              pgtype.Text
}
type ListStudentToDoItemArgs struct {
	StudentID   pgtype.Text
	StudyPlanID pgtype.Text

	// fields used for pagination query
	Limit   uint32
	TopicID pgtype.Text
}
type TopicProgress struct {
	TopicID         pgtype.Text
	CompletedSPItem pgtype.Int2
	TotalSPItem     pgtype.Int2
	AverageScore    pgtype.Int2
	Name            pgtype.Text
	IconURL         pgtype.Text
	DisplayOrder    pgtype.Int2
}

func (r *StudyPlanRepo) ListStudentToDoItem(ctx context.Context, db database.QueryExecer, args *ListStudentToDoItemArgs) ([]*StudentStudyPlanItem, []*TopicProgress, error) {
	var selectFieldsItem = []string{
		"learning_material_id", "isp.topic_id", "name", "type", "lm.display_order", "created_at", "lm.updated_at", "deleted_at", "start_date",
		"end_date", "school_date", "completed_at", "status", "available_from", "available_to", "book_id",
	}

	var selectFieldsTopicProgress = []string{
		"topic_id", "completed_sp_item", "total_sp_item", "average_score", "name", "icon_url", "display_order",
	}

	StudentToDoItemStmt := fmt.Sprintf(`
	SELECT
	%s
	FROM individual_study_plan_fn() isp
	INNER JOIN learning_material lm using (learning_material_id)
	LEFT JOIN get_student_completion_learning_material() gsl using(student_id, study_plan_id, learning_material_id)
	WHERE
		isp.student_id = $1
		AND study_plan_id = $2
		AND ($3::TEXT is null or isp.topic_id > $3::TEXT)
	ORDER BY
		topic_id ASC,
		lm.display_order ASC
	LIMIT $4
	`, strings.Join(selectFieldsItem, ", "))

	TopicProgressStmt := fmt.Sprintf(`
	SELECT
	%s
	FROM get_student_topic_progress()
	INNER JOIN topics using (topic_id)
	WHERE 
	student_id = $1
	AND study_plan_id = $2
	AND topic_id = ANY($3::_TEXT)
	`, strings.Join(selectFieldsTopicProgress, ", "))

	rows, err := db.Query(ctx, StudentToDoItemStmt, &args.StudentID, &args.StudyPlanID, &args.TopicID, &args.Limit)
	if err != nil {
		return nil, nil, fmt.Errorf("listStudentToDoItem.Database.Select: %w", err)
	}
	defer rows.Close()
	var items []*StudentStudyPlanItem
	for rows.Next() {
		item := new(StudentStudyPlanItem)
		if err := rows.Scan(
			&item.LearningMaterialID,
			&item.TopicID,
			&item.Name,
			&item.Type,
			&item.DisplayOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
			&item.StartDate,
			&item.EndDate,
			&item.SchoolDate,
			&item.CompletedAt,
			&item.StudyPlanItemStatus,
			&item.AvailableFrom,
			&item.AvailableTo,
			&item.BookID,
		); err != nil {
			return nil, nil, fmt.Errorf("listStudentToDoItem.database.Scan: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("db.Error: %w", err)
	}

	var StudentTopicProgress []*TopicProgress
	if len(items) > 0 {
		topicIDs := make([]string, 0, len(items))
		for _, item := range items {
			topicIDs = append(topicIDs, item.TopicID.String)
		}

		topicIDs = golibs.Uniq(topicIDs)
		rows, err := db.Query(ctx, TopicProgressStmt, &args.StudentID, &args.StudyPlanID, topicIDs)
		if err != nil {
			return nil, nil, fmt.Errorf("listTopicProgress.Database.Select: %w", err)
		}

		for rows.Next() {
			topicProgress := new(TopicProgress)
			if err := rows.Scan(
				&topicProgress.TopicID,
				&topicProgress.CompletedSPItem,
				&topicProgress.TotalSPItem,
				&topicProgress.AverageScore,
				&topicProgress.Name,
				&topicProgress.IconURL,
				&topicProgress.DisplayOrder,
			); err != nil {
				return nil, nil, fmt.Errorf("listTopicProgress.database.Scan: %w", err)
			}

			StudentTopicProgress = append(StudentTopicProgress, topicProgress)
		}
		if err := rows.Err(); err != nil {
			return nil, nil, fmt.Errorf("db.Error: %w", err)
		}
	}

	return items, StudentTopicProgress, nil
}

type ListStudentStudyPlansArgs struct {
	StudentIDs pgtype.TextArray
	CourseID   pgtype.Text
	Limit      uint32
	Offset     pgtype.Text
	Search     pgtype.Text
	Status     pgtype.Text
	BookIDs    pgtype.TextArray
	Grades     pgtype.Int4Array
}

type StudentStudyPlan struct {
	entities.StudyPlan
	StudentID pgtype.Text
}

const listStudentStudyPlansStmtTpl = `SELECT distinct
	i.%s, s.student_id
FROM
	%s AS i
INNER JOIN individual_study_plan_fn() s ON
	i.study_plan_id = s.study_plan_id
WHERE
	s.student_id = ANY($1::TEXT[])
	AND ($2::TEXT IS NULL
	OR i.course_id = $2)
	AND ($3::TEXT IS NULL
	OR i.study_plan_id < $3)
	AND i.deleted_at IS NULL
	AND ($5::TEXT IS NULL OR i.name ilike ('%%' || $5::TEXT || '%%'))
	AND ($6::TEXT IS NULL or i.status = $6::TEXT)
	AND ($7::TEXT[] IS NULL or i.book_id = any($7::TEXT[]))
	AND ($8::INT[] IS NULL or exists (SELECT * from unnest(i.grades) where unnest = any($8::INT[])))
ORDER BY
	i.study_plan_id DESC
LIMIT $4`

func (r *StudyPlanRepo) ListStudentStudyPlans(ctx context.Context, db database.QueryExecer, args *ListStudentStudyPlansArgs) ([]*StudentStudyPlan, error) {
	var e entities.StudyPlan
	var items []*StudentStudyPlan
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(listStudentStudyPlansStmtTpl,
		strings.Join(selectFields, ", i."), e.TableName())

	rows, err := db.Query(ctx, query, &args.StudentIDs, &args.CourseID, &args.Offset, &args.Limit, &args.Search, &args.Status, &args.BookIDs, &args.Grades)
	if err != nil {
		return nil, fmt.Errorf("StudyPlanRepo.ListStudentStudyPlans: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		item := new(StudentStudyPlan)
		if err := rows.Scan(
			&item.ID,
			&item.MasterStudyPlan,
			&item.Name,
			&item.StudyPlanType,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
			&item.SchoolID,
			&item.CourseID,
			&item.BookID,
			&item.Status,
			&item.TrackSchoolProgress,
			&item.Grades,
			&item.StudentID,
		); err != nil {
			return nil, fmt.Errorf("StudyPlanRepo.ListStudentStudyPlans.Rows.Scan: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}
