package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type ExamLORepo struct{}

func (r *ExamLORepo) Insert(ctx context.Context, db database.QueryExecer, e *entities.ExamLO) error {
	ctx, span := interceptors.StartSpan(ctx, "ExamLORepo.Insert")
	defer span.End()
	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	return nil
}

func (r *ExamLORepo) Update(ctx context.Context, db database.QueryExecer, e *entities.ExamLO) error {
	ctx, span := interceptors.StartSpan(ctx, "ExamLORepo.Update")
	defer span.End()

	updateFields := []string{
		"name",
		"instruction",
		"grade_to_pass",
		"manual_grading",
		"time_limit",
		"updated_at",
		"maximum_attempt",
		"approve_grading",
		"grade_capping",
		"review_option",
	}

	if _, err := database.UpdateFields(ctx, e, db.Exec, "learning_material_id", updateFields); err != nil {
		return fmt.Errorf("database.UpdateFields: %w", err)
	}
	return nil
}

func (r *ExamLORepo) ListByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ExamLO, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLORepo.ListByIDs")
	defer span.End()
	els := &entities.ExamLOs{}
	e := &entities.ExamLO{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = ANY($1::_TEXT) AND deleted_at IS NULL", strings.Join(database.GetFieldNames(e), ", "), e.TableName())
	if err := database.Select(ctx, db, query, ids).ScanAll(els); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return els.Get(), nil
}

func (r *ExamLORepo) ListExamLOBaseByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ExamLOBase, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLORepo.ListExamLOBaseByIDs")
	defer span.End()

	bases := make([]*entities.ExamLOBase, 0)
	e := &entities.ExamLO{}

	stmt := fmt.Sprintf(`
		SELECT el.%s, array_length(qs.quiz_external_ids, 1) as total_question
		FROM %s el
		LEFT JOIN quiz_sets qs ON qs.lo_id = el.learning_material_id AND qs.deleted_at IS NULL
		WHERE el.learning_material_id = ANY($1::_TEXT) AND el.deleted_at IS NULL
	`,
		strings.Join(database.GetFieldNames(e), ", el."),
		e.TableName(),
	)

	rows, err := db.Query(ctx, stmt, ids)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		b := &entities.ExamLOBase{}

		_, values := b.ExamLO.FieldMap()
		values = append(values, &b.TotalQuestion)

		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		bases = append(bases, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return bases, nil
}

func (r *ExamLORepo) ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.ExamLO, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLORepo.ListByTopicIDs")
	defer span.End()
	examLOs := &entities.ExamLOs{}
	e := &entities.ExamLO{}
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(e), ","), e.TableName())
	if err := database.Select(ctx, db, query, topicIDs).ScanAll(examLOs); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return *examLOs, nil
}

func (r *ExamLORepo) BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.ExamLO) error {
	err := dbeureka.BulkUpsert(ctx, db, bulkInsertQuery, items)
	if err != nil {
		return fmt.Errorf("ExamLORepo database.BulkInsert error: %s", err.Error())
	}
	return nil
}

func (r *ExamLORepo) GetScores(ctx context.Context, db database.QueryExecer, courseIDs, studyPlanIDs, studentIDs pgtype.TextArray, getGradeToPassScore pgtype.Bool) ([]*entities.ExamLoScore, error) {
	examLoScores := make([]*entities.ExamLoScore, 0)
	stmt := `
	select sp.course_id,
		msp.study_plan_id,
		sp.name,
		ssp.student_id,
		u.name,
		s.current_grade,
		s.grade_id,
		el.learning_material_id,
		el.name,
		elp.total_point,
		elp.graded_point,
		elp.passed,
		elp.total_attempts,
		elp.status,
		(el.grade_to_pass is not null) as is_grade_to_pass,
		el.review_option,
		check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.end_date,isp.end_date) end_date,
		c.display_order,
		t.display_order,
		lm.display_order,
		count(*) over (partition by ssp.student_id, msp.study_plan_id) as total_exam_los,
		count(*) filter (where elp.status is not null ) over (partition by ssp.student_id, msp.study_plan_id) as total_completed_exam_los,
		count(*) filter (where el.grade_to_pass is not null) over (partition by ssp.student_id, msp.study_plan_id) as total_grade_to_pass,
		count(*) filter (where elp.passed and (el.review_option = 'EXAM_LO_REVIEW_OPTION_IMMEDIATELY' or (now()>check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.end_date,isp.end_date)))) over (partition by ssp.student_id, msp.study_plan_id) as total_passed
		from study_plans sp
		join student_study_plans ssp on ssp.study_plan_id  = sp.study_plan_id 
	    Join books b on b.book_id = sp.book_id
	    JOIN chapters c on c.book_id = b.book_id
	    JOIN topics t on t.chapter_id= c.chapter_id
	    JOIN learning_material lm USING (topic_id)		
	    JOIN exam_lo el on lm.learning_material_id = el.learning_material_id
		JOIN master_study_plan msp on msp.study_plan_id = sp.master_study_plan_id and msp.learning_material_id = el.learning_material_id 
		LEFT JOIN individual_study_plan isp on isp.study_plan_id = msp.study_plan_id and isp.learning_material_id = msp.learning_material_id and isp.student_id = ssp.student_id 
		join users u on ssp.student_id = u.user_id
		join students s on u.user_id = s.student_id
		left join (
					select distinct on (student_id, study_plan_id, learning_material_id) student_id,
			                                                                     study_plan_id,
			                                                                     learning_material_id,
			                                                                     -- graded point is calculated
			                                                                     -- if all submission are fails choose latest score
			                                                                     -- if a submission is passed choose grade_to_pass from exam_lo setting
			                                                                     -- if a submission is passed from 2nd
			                                                                     -- ex : s1 failed, s2 pass, s3 pass
			                                                                     -- we will calculate s2 = pass * count(s1,s2,s3) > 1
			                                                                     coalesce(NULLIF(
			                                                                                      (e.grade_to_pass *
			                                                                                       (result = 'EXAM_LO_SUBMISSION_PASSED')::integer *
			                                                                                       (count(*) over (
			                                                                                           partition by student_id,
			                                                                                               study_plan_id,
			                                                                                               learning_material_id
			                                                                                           ) > 1)::integer *
			                                                                                       e.grade_capping::integer)::smallint,
			                                                                                      0)
			                                                                         , s.graded_point)                  as graded_point,
			                                                                     total_point,
			                                                                     status,
			                                                                     (result = 'EXAM_LO_SUBMISSION_PASSED') as passed,
			                                                                     count(*) over (partition by student_id,
			                                                                         study_plan_id,
			                                                                         learning_material_id)::smallint    as total_attempts
			from get_exam_lo_scores() s
			         join exam_lo e using (learning_material_id)
			-- order by the submissions are passed -> then latest
			order by student_id, study_plan_id, learning_material_id, (result = 'EXAM_LO_SUBMISSION_PASSED') desc, (status = 'SUBMISSION_STATUS_RETURNED') desc, s.created_at desc
		) elp on elp.student_id = ssp.student_id and elp.study_plan_id = msp.study_plan_id and elp.learning_material_id = el.learning_material_id
	where ($1::TEXT[] IS NULL OR sp.course_id = any($1::TEXT[]))
	AND ($2::TEXT[] IS NULL OR msp.study_plan_id = any($2::TEXT[]))
	AND ($3::TEXT[] IS NULL OR ssp.student_id = any($3::TEXT[]))
	AND msp.deleted_at is null AND isp.deleted_at is null	
	AND ((isp.updated_at IS NULL and msp.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE') or (isp.updated_at >= msp.updated_at and isp.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE') or (isp.updated_at <= msp.updated_at and msp.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'));
	`

	rows, err := db.Query(ctx, stmt, &courseIDs, &studyPlanIDs, &studentIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	for rows.Next() {
		examLoScore := entities.ExamLoScore{}
		err = rows.Scan(&examLoScore.CourseID, &examLoScore.StudyPlanID, &examLoScore.StudyPlanName, &examLoScore.StudentID,
			&examLoScore.StudentName, &examLoScore.Grade, &examLoScore.GradeID, &examLoScore.LearningMaterialID, &examLoScore.ExamLOName,
			&examLoScore.TotalPoint, &examLoScore.GradePoint, &examLoScore.PassedExamLo, &examLoScore.TotalAttempts, &examLoScore.Status,
			&examLoScore.IsGradeToPass, &examLoScore.ReviewOption, &examLoScore.DueDate, &examLoScore.ChapterDisplayOrder, &examLoScore.TopicDisplayOrder,
			&examLoScore.LmDisplayOrder, &examLoScore.TotalExamLOs, &examLoScore.TotalCompletedExamLOs, &examLoScore.TotalGradeToPass, &examLoScore.TotalPassed)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		examLoScores = append(examLoScores, &examLoScore)
	}

	return examLoScores, nil
}

func (r *ExamLORepo) UpsertGradeBookSetting(ctx context.Context, db database.QueryExecer, item *entities.GradeBookSetting) error {
	stmt := `INSERT INTO grade_book_setting (%s) VALUES (%s)
	ON CONFLICT ON CONSTRAINT grade_book_setting_pk
	DO UPDATE SET setting = $1, updated_by = $2, updated_at = $3`
	fieldNames := database.GetFieldNames(item)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	scanFields := database.GetScanFields(item, fieldNames)
	query := fmt.Sprintf(stmt, strings.Join(fieldNames, ","), placeHolders)
	_, err := db.Exec(ctx, query, scanFields...)
	if err != nil {
		return err
	}

	return nil
}

func (r *ExamLORepo) Get(ctx context.Context, db database.QueryExecer, learningMaterialID pgtype.Text) (*entities.ExamLO, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExamLORepo.Get")
	defer span.End()

	var result entities.ExamLO
	fields, _ := result.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE deleted_at IS NULL AND learning_material_id = $1::TEXT;`, strings.Join(fields, ", "), result.TableName())

	if err := database.Select(ctx, db, stmt, learningMaterialID).ScanOne(&result); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return &result, nil
}
