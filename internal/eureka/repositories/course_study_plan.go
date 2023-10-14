package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type CourseStudyPlanRepo struct {
}

type ListCourseStatisticItemsArgsV2 struct {
	CourseID    pgtype.Text
	StudyPlanID pgtype.Text
	ClassID     pgtype.TextArray
}

type CourseStatisticItemV2 struct {
	ContentStructure    entities.ContentStructure
	StudentID           string
	RootStudyPlanItemID string
	StudyPlanItemID     string
	Status              string
	CompletedAt         pgtype.Timestamptz
	Score               float32
	LearningMaterialID  string
}

func (r *CourseStudyPlanRepo) QueueUpsertCourseStudyPlan(b *pgx.Batch, studyPlan *entities.CourseStudyPlan) {
	fieldNames := database.GetFieldNames(studyPlan)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT course_study_plans_pk DO UPDATE SET
		updated_at = $4`, studyPlan.TableName(), strings.Join(fieldNames, ","), placeHolders)
	scanFields := database.GetScanFields(studyPlan, fieldNames)
	b.Queue(query, scanFields...)
}

func (r *CourseStudyPlanRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, courseStudyPlans []*entities.CourseStudyPlan) error {
	b := &pgx.Batch{}
	for _, studyPlan := range courseStudyPlans {
		r.QueueUpsertCourseStudyPlan(b, studyPlan)
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

// FindByCourseIDs return all master study plan of course
func (r *CourseStudyPlanRepo) FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) ([]*entities.CourseStudyPlan, error) {
	e := &entities.CourseStudyPlan{}
	fields, _ := e.FieldMap()

	selectFields := make([]string, 0)
	for _, f := range fields {
		selectFields = append(selectFields, "csp"+"."+f)
	}

	query := fmt.Sprintf(
		`
		SELECT %s
		FROM %s AS csp
		JOIN study_plans AS sp
			ON sp.study_plan_id = csp.study_plan_id
			AND sp.master_study_plan_id IS NULL
			AND sp.deleted_at IS NULL
		WHERE csp.course_id = ANY($1) AND csp.deleted_at IS NULL
	`,
		strings.Join(selectFields, ", "), e.TableName(),
	)

	rows, err := db.Query(ctx, query, &courseIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result []*entities.CourseStudyPlan
	for rows.Next() {
		e := &entities.CourseStudyPlan{}
		_, values := e.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("[Scan Error]:%v", err)
		}
		result = append(result, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[Other Error]:%v", err)
	}
	return result, nil
}

type ListCourseStudyPlansArgs struct {
	CourseIDs pgtype.TextArray
	BookIDs   pgtype.TextArray
}

// ListCourseStudyPlans return all master study plan of course
func (r *CourseStudyPlanRepo) ListCourseStudyPlans(ctx context.Context, db database.QueryExecer, args *ListCourseStudyPlansArgs) ([]*entities.CourseStudyPlan, error) {
	e := &entities.CourseStudyPlan{}
	fields, _ := e.FieldMap()

	selectFields := make([]string, 0)
	for _, f := range fields {
		selectFields = append(selectFields, "csp"+"."+f)
	}

	query := fmt.Sprintf(
		`
		SELECT %s
		FROM %s AS csp
		JOIN study_plans AS sp
			ON sp.study_plan_id = csp.study_plan_id
			AND sp.master_study_plan_id IS NULL
			AND sp.deleted_at IS NULL
		WHERE ($1::_TEXT IS NULL OR csp.course_id = ANY($1)) AND 
		($2::_TEXT IS NULL OR sp.book_id = ANY($2)) AND
		csp.deleted_at IS NULL
	`,
		strings.Join(selectFields, ", "), e.TableName(),
	)

	rows, err := db.Query(ctx, query, &args.CourseIDs, &args.BookIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result []*entities.CourseStudyPlan
	for rows.Next() {
		e := &entities.CourseStudyPlan{}
		_, values := e.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("[Scan Error]:%v", err)
		}
		result = append(result, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("[Other Error]:%v", err)
	}
	return result, nil
}

func (r *CourseStudyPlanRepo) DeleteCourseStudyPlanBy(ctx context.Context, db database.QueryExecer, req *entities.CourseStudyPlan) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudyPlanRepo.DeleteCourseStudyPlanBy")
	defer span.End()

	query := `
		UPDATE course_study_plans AS csp SET deleted_at = NOW() WHERE csp.course_id = $1 AND csp.study_plan_id = $2 ;
	`
	_, err := db.Exec(ctx, query, &req.CourseID, &req.StudyPlanID)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

type ListCourseStatisticItemsArgs struct {
	CourseID    pgtype.Text
	StudyPlanID pgtype.Text
	ClassID     pgtype.Text
}

type CourseStatisticItem struct {
	ContentStructure    entities.ContentStructure
	StudentID           string
	RootStudyPlanItemID string
	StudyPlanItemID     string
	Status              string
	CompletedAt         pgtype.Timestamptz
	Score               float32
}

func (r *CourseStudyPlanRepo) ListCourseStatisticItems(ctx context.Context, db database.QueryExecer, args *ListCourseStatisticItemsArgs) ([]*CourseStatisticItem, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudyPlanRepo.ListCourseStatisticItems")
	defer span.End()

	const queryTmpl = `
WITH study_plan_item_tmp AS
  (SELECT spi.content_structure,
          spi.copy_study_plan_item_id AS root_study_plan_item_id,
          ssp.student_id,
          spi.study_plan_item_id,
          spi.status,
          spi.completed_at,
          spi.content_structure ->> 'book_id' AS book_id,
          spi.content_structure ->> 'chapter_id' AS chapter_id,
          spi.content_structure ->> 'topic_id' AS topic_id,
          COALESCE(NULLIF(spi.content_structure ->> 'assignment_id', ''), spi.content_structure ->> 'lo_id') AS lo_id
   FROM course_study_plans csp
   JOIN study_plans sp ON sp.course_id = csp.course_id AND sp.master_study_plan_id = csp.study_plan_id
   JOIN student_study_plans ssp ON ssp.study_plan_id = sp.study_plan_id
   JOIN course_students cst on cst.course_id = csp.course_id and cst.student_id = ssp.student_id
   JOIN study_plan_items spi ON spi.study_plan_id = sp.study_plan_id
   LEFT JOIN course_classes cc ON cc.course_id = csp.course_id AND cc.class_id = $3
   LEFT JOIN class_students cs ON cs.student_id = ssp.student_id AND cs.class_id = $3
   WHERE csp.course_id = $1::TEXT
     AND sp.master_study_plan_id = $2::TEXT
     AND ($3::TEXT IS NULL OR ($3::TEXT = cs.class_id AND cs.deleted_at IS NULL AND cc.deleted_at IS NULL))
     AND csp.deleted_at IS NULL
     AND sp.deleted_at IS NULL
     AND spi.deleted_at IS NULL
     AND ssp.deleted_at IS NULL
     AND cst.deleted_at IS NULL
     AND spi.content_structure IS NOT NULL )
SELECT spi.content_structure,
       spi.root_study_plan_item_id,
       spi.student_id,
       spi.study_plan_item_id,
       spi.status,
       spi.completed_at
FROM study_plan_item_tmp spi
JOIN books b ON b.book_id = spi.book_id
JOIN chapters c ON c.chapter_id = spi.chapter_id
JOIN topics t ON t.topic_id = spi.topic_id
LEFT JOIN topics_assignments ta ON ta.topic_id = spi.topic_id AND ta.assignment_id = spi.lo_id
LEFT JOIN topics_learning_objectives tlo ON tlo.topic_id = spi.topic_id AND tlo.lo_id = spi.lo_id
WHERE b.deleted_at IS NULL
  AND c.deleted_at IS NULL
  AND t.deleted_at IS NULL
  AND (ta.deleted_at IS NULL OR tlo.deleted_at IS NULL)
ORDER BY b.book_id,
         c.display_order,
         t.display_order,
         coalesce(ta.display_order, tlo.display_order)
`
	rows, err := db.Query(ctx, queryTmpl, args.CourseID, args.StudyPlanID, args.ClassID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result []*CourseStatisticItem
	for rows.Next() {
		e := &CourseStatisticItem{}

		var contentStructure pgtype.JSONB
		if err := rows.Scan(&contentStructure, &e.RootStudyPlanItemID, &e.StudentID, &e.StudyPlanItemID, &e.Status, &e.CompletedAt); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		if err := contentStructure.AssignTo(&e.ContentStructure); err != nil {
			return nil, fmt.Errorf("contentStructure.AssignTo: %w", err)
		}
		result = append(result, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return result, nil
}

func (r *CourseStudyPlanRepo) ListCourseStatisticItemsV2(ctx context.Context, db database.QueryExecer, args *ListCourseStatisticItemsArgsV2) ([]*CourseStatisticItemV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudyPlanRepo.ListTopicStatistic")
	defer span.End()

	const queryTmpl = `
WITH student_ids AS (
	SELECT distinct(cst.student_id)
	FROM course_study_plans csp
	JOIN course_students cst USING (course_id)
	LEFT JOIN course_classes cc USING (course_id)
	LEFT JOIN class_students cs USING (student_id, class_id)
	WHERE csp.study_plan_id = $2::TEXT
	AND csp.course_id = $1::TEXT
	AND ($3::TEXT[] IS NULL OR cs.class_id = ANY($3::TEXT[]))
	AND csp.deleted_at IS NULL
	AND cst.deleted_at IS NULL
	AND cs.deleted_at IS NULL
),
study_plan_item_tmp AS
  (SELECT spi.content_structure,
          spi.copy_study_plan_item_id AS root_study_plan_item_id,
          ssp.student_id,
          spi.study_plan_item_id,
          spi.status,
          spi.completed_at,
          spi.content_structure ->> 'book_id' AS book_id,
          spi.content_structure ->> 'chapter_id' AS chapter_id,
          spi.content_structure ->> 'topic_id' AS topic_id,
          COALESCE(NULLIF(spi.content_structure ->> 'assignment_id', ''), spi.content_structure ->> 'lo_id') AS learning_material_id
   FROM student_study_plans ssp
   JOIN student_ids sid on sid.student_id = ssp.student_id
   JOIN study_plan_items spi ON spi.study_plan_id = ssp.study_plan_id
   WHERE coalesce(ssp.master_study_plan_id, ssp.study_plan_id) = $2::TEXT
	AND spi.deleted_at IS NULL
)
SELECT spi.content_structure,
       spi.root_study_plan_item_id,
       spi.student_id,
       spi.study_plan_item_id,
       spi.status,
       spi.completed_at,
       spi.learning_material_id
FROM study_plan_item_tmp spi
JOIN books b ON b.book_id = spi.book_id
JOIN chapters c ON c.chapter_id = spi.chapter_id
JOIN topics t ON t.topic_id = spi.topic_id
LEFT JOIN topics_assignments ta ON ta.topic_id = spi.topic_id AND ta.assignment_id = spi.learning_material_id
LEFT JOIN topics_learning_objectives tlo ON tlo.topic_id = spi.topic_id AND tlo.lo_id = spi.learning_material_id
WHERE b.deleted_at IS NULL
  AND c.deleted_at IS NULL
  AND t.deleted_at IS NULL
  AND (ta.deleted_at IS NULL OR tlo.deleted_at IS NULL)
ORDER BY b.book_id,
         c.display_order,
         t.display_order,
         coalesce(ta.display_order, tlo.display_order)`

	rows, err := db.Query(ctx, queryTmpl, args.CourseID, args.StudyPlanID, args.ClassID)
	if err != nil {
		return nil, fmt.Errorf("db.Query error: %w", err)
	}
	defer rows.Close()

	var result []*CourseStatisticItemV2
	for rows.Next() {
		e := &CourseStatisticItemV2{}

		var contentStructure pgtype.JSONB
		if err := rows.Scan(&contentStructure, &e.RootStudyPlanItemID, &e.StudentID, &e.StudyPlanItemID, &e.Status, &e.CompletedAt, &e.LearningMaterialID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		if err := contentStructure.AssignTo(&e.ContentStructure); err != nil {
			return nil, fmt.Errorf("contentStructure.AssignTo: %w", err)
		}
		result = append(result, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return result, nil
}

type ListCourseStatisticItemsArgsV3 struct {
	CourseID    pgtype.Text
	StudyPlanID pgtype.Text
	ClassID     pgtype.TextArray
	StudentIDs  pgtype.TextArray
	TagIDs      pgtype.TextArray
	LocationIDs pgtype.TextArray
}

type ItemStatistic struct {
	LearningMaterialID  string
	TopicID             string
	TotalAssignStudent  pgtype.Int4
	CompletedStudent    pgtype.Int4
	AverageScore        pgtype.Int4
	AverageScoreRaw     pgtype.Int4
	Type                string
	ChapterDisplayOrder pgtype.Int2
	TopicDisplayOrder   pgtype.Int2
	LmDisplayOrder      pgtype.Int4
}

type LearningMaterialStatistic struct {
	LearningMaterialID string
	TopicID            string
	TotalAssignStudent pgtype.Int4
	CompletedStudent   pgtype.Int4
	AverageScore       pgtype.Int4
	AverageScoreRaw    pgtype.Int4
}

type TopicStatistic struct {
	TopicID            string
	TotalAssignStudent pgtype.Int4
	CompletedStudent   pgtype.Int4
	AverageScore       pgtype.Int4
}

const courseStatisticQuery = `
	WITH ss as (
		SELECT
			bt.learning_material_id,
			bt.topic_id,
			sp.study_plan_id,
			cs.student_id,
			bt.chapter_display_order,
			bt.topic_display_order,
			bt.lm_display_order,
			(
				CASE
					WHEN (msp.updated_at >= isp.updated_at) THEN msp.status
					WHEN ((msp.updated_at IS NULL) OR (isp.updated_at IS NULL)) THEN COALESCE(msp.status, isp.status)
					ELSE isp.status
				END
			) AS status
		FROM study_plans sp
		JOIN course_students cs on cs.course_id = sp.course_id
		JOIN UNNEST($3::TEXT[]) WITH ORDINALITY AS si(student_id) on cs.student_id=si.student_id
		JOIN book_tree_fn() bt USING (book_id)
		LEFT JOIN master_study_plan msp
			on sp.study_plan_id = msp.study_plan_id and bt.learning_material_id = msp.learning_material_id
		LEFT JOIN student_study_plans ssp
			on sp.study_plan_id = ssp.master_study_plan_id and cs.student_id = ssp.student_id
		LEFT JOIN individual_study_plan isp
			on sp.study_plan_id = isp.study_plan_id and cs.student_id = isp.student_id and bt.learning_material_id = isp.learning_material_id
		WHERE
			sp.course_id=$1::TEXT
			AND sp.study_plan_id=$2::TEXT
			AND sp.master_study_plan_id IS NULL
			AND (ssp.student_id IS NOT NULL OR sp.study_plan_type = 'STUDY_PLAN_TYPE_COURSE')
			AND now() BETWEEN cs.start_at AND cs.end_at
			AND (
				$4::TEXT[] IS NULL OR EXISTS (
					SELECT 1
					FROM tagged_user tu
					WHERE tu.user_id = cs.student_id AND tu.tag_id = ANY($4::TEXT[]) AND tu.deleted_at IS NULL
				)
			)
			AND msp.available_from IS NOT NULL AND msp.available_to IS NOT NULL
	)
	SELECT
		ss.student_id,
		ss.study_plan_id,
		ss.topic_id,
		ss.status,
		ss.learning_material_id,
		ss.chapter_display_order,
		ss.topic_display_order,
		ss.lm_display_order,
		(
			select completed_at is not null
			from get_student_completion_learning_material() cl
			where cl.study_plan_id = ss.study_plan_id and
					cl.learning_material_id = ss.learning_material_id and
					cl.student_id = ss.student_id
			limit 1
		) as is_completed,
		(
			select coalesce((sc.graded_points * 1.0 / sc.total_points) * 100, null)::smallint
			from max_graded_score() sc
			where sc.study_plan_id=ss.study_plan_id and
					sc.learning_material_id = ss.learning_material_id and
					sc.student_id = ss.student_id
			limit 1
		) as scorce
	FROM ss
`

func (r *CourseStudyPlanRepo) ListCourseStatisticV3(ctx context.Context, db database.QueryExecer, args *ListCourseStatisticItemsArgsV3) ([]*TopicStatistic, []*LearningMaterialStatistic, error) {
	topics := []*TopicStatistic{}
	lms := []*LearningMaterialStatistic{}

	{
		// Query for learning_material_statistic
		query := fmt.Sprintf(`
			SELECT
				learning_material_id,
				topic_id,
				sum((status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE')::int)::integer as total_assign_student,
				sum((status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE' and is_completed = true)::int)::integer as completed_student,
				coalesce(avg(scorce), -1)::integer as average_score,
				avg(scorce)::integer as average_score_raw
			FROM (%s) course_stats
			GROUP BY learning_material_id, topic_id, chapter_display_order,topic_display_order,lm_display_order
			order by chapter_display_order,topic_display_order,lm_display_order
		`, courseStatisticQuery)

		rows, err := db.Query(ctx, query, &args.CourseID, &args.StudyPlanID, &args.StudentIDs, &args.TagIDs)
		if err != nil {
			return nil, nil, err
		}

		defer rows.Close()

		for rows.Next() {
			e := &LearningMaterialStatistic{}
			if err := rows.Scan(&e.LearningMaterialID, &e.TopicID, &e.TotalAssignStudent, &e.CompletedStudent, &e.AverageScore, &e.AverageScoreRaw); err != nil {
				return nil, nil, fmt.Errorf("LearningMaterialStatistic rows.Scan: %w", err)
			}
			lms = append(lms, e)
		}

		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
	}

	{
		// Query for topic_statistic
		query := fmt.Sprintf(`
			WITH course_stat AS (
				%s
			),
			lm_stat AS (
				SELECT
					topic_id,
					coalesce(avg(scorce), -1)::integer as average_score,
					avg(scorce)::integer as average_score_raw
				FROM course_stat
				GROUP BY learning_material_id, topic_id
			)
			SELECT
				topic_id,
				coalesce((
						SELECT avg(ls.average_score_raw) as topic_avg_score
						FROM lm_stat ls
						WHERE ls.average_score != -1
						and ls.topic_id = topic_stat.topic_id
						GROUP BY ls.topic_id
					)::smallint,
					-1
				)::integer as avg_score,
				sum((active_lm > 0 and active_lm = completed_lm)::int)::integer as topic_completed_student,
				sum((active_lm > 0)::int)::integer as topic_assign_student
			FROM (
				SELECT
					topic_id,
					student_id,
					sum((status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE')::int) as active_lm,
					sum((status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE' and is_completed = true)::int) as completed_lm
				FROM course_stat
				GROUP BY topic_id, student_id
			) topic_stat
			GROUP BY topic_id
		`, courseStatisticQuery)

		rows, err := db.Query(ctx, query, &args.CourseID, &args.StudyPlanID, &args.StudentIDs, &args.TagIDs)
		if err != nil {
			return nil, nil, err
		}

		defer rows.Close()

		for rows.Next() {
			e := &TopicStatistic{}
			if err := rows.Scan(&e.TopicID, &e.AverageScore, &e.CompletedStudent, &e.TotalAssignStudent); err != nil {
				return nil, nil, fmt.Errorf("TopicStatistic rows.Scan: %w", err)
			}
			topics = append(topics, e)
		}

		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
	}

	return topics, lms, nil
}

const courseStatisticQueryV4 = `
WITH course_stats AS (
	WITH ss as (
		SELECT
			bt.learning_material_id,
			bt.topic_id,
			sp.study_plan_id,
			cs.student_id,
			bt.chapter_display_order,
			bt.topic_display_order,
			bt.lm_display_order,
			check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.available_from, isp.available_from) as available_from,
			check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.available_to, isp.available_to) as available_to,
			check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.end_date, isp.end_date) as end_date,
			(
				CASE
					WHEN (msp.updated_at >= isp.updated_at) THEN msp.status
					WHEN ((msp.updated_at IS NULL) OR (isp.updated_at IS NULL)) THEN COALESCE(msp.status, isp.status)
					ELSE isp.status
				END
			) AS status,
			el.review_option,
			al.is_required_grade,
			al.learning_material_id as assignment_id,
			ta.require_correctness,
			ta.learning_material_id as task_assignment_id
		FROM study_plans sp
		JOIN book_tree_fn() bt USING (book_id)
		LEFT JOIN master_study_plan msp
			on sp.study_plan_id = msp.study_plan_id and bt.learning_material_id = msp.learning_material_id
		LEFT JOIN student_study_plans ssp
			on sp.study_plan_id = ssp.master_study_plan_id 
		JOIN course_students cs 
			on cs.course_id = sp.course_id and ssp.student_id = cs.student_id 
		JOIN UNNEST($3::TEXT[]) WITH ORDINALITY AS si(student_id) on cs.student_id=si.student_id
		LEFT JOIN individual_study_plan isp
			on sp.study_plan_id = isp.study_plan_id and cs.student_id = isp.student_id and bt.learning_material_id = isp.learning_material_id
		LEFT JOIN exam_lo el 
			on el.learning_material_id = bt.learning_material_id
		LEFT JOIN assignment al 
			on al.learning_material_id = bt.learning_material_id
		LEFT JOIN task_assignment ta
			on ta.learning_material_id = bt.learning_material_id 
		WHERE
			sp.course_id=$1::TEXT
			AND sp.study_plan_id=$2::TEXT
			AND sp.master_study_plan_id IS NULL
			AND (ssp.student_id IS NOT NULL OR sp.study_plan_type = 'STUDY_PLAN_TYPE_COURSE')
			AND now() BETWEEN cs.start_at AND cs.end_at
			AND (
				$4::TEXT[] IS NULL OR EXISTS (
					SELECT 1
					FROM tagged_user tu
					WHERE tu.user_id = cs.student_id AND tu.tag_id = ANY($4::TEXT[]) AND tu.deleted_at IS NULL
				)
			)
	)
	SELECT
		ss.student_id,
		ss.study_plan_id,
		ss.topic_id,
		ss.status,
		ss.learning_material_id,
		ss.chapter_display_order,
		ss.topic_display_order,
		ss.lm_display_order,
		(
			select completed_at is not null
			from get_student_completion_learning_material() cl
			where cl.study_plan_id = ss.study_plan_id and
					cl.learning_material_id = ss.learning_material_id and
					cl.student_id = ss.student_id
			limit 1
		) as is_completed,
		(
			select max_percentage
			from max_score_submission mss
			where mss.study_plan_id = ss.study_plan_id
			and mss.learning_material_id = ss.learning_material_id
			and mss.student_id = ss.student_id
			and ss.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE'
			and (ss.review_option is null or ss.review_option = 'EXAM_LO_REVIEW_OPTION_IMMEDIATELY' or (now()>ss.end_date))
			and (ss.assignment_id is null or ss.is_required_grade is true)
			and (ss.task_assignment_id is null or ss.require_correctness is true)
		) as percentage
	FROM ss
	WHERE ss.available_from is not null and ss.available_to is not null
)
(SELECT
	learning_material_id,
	topic_id,
	sum((status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE')::int)::integer as total_assign_student,
	sum((status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE' and is_completed = true)::int)::integer as completed_student,
	coalesce(avg(percentage), -1)::integer as average_score,
	avg(percentage)::integer as average_score_raw,
	'LEARNING_MATERIAL' as type,
	chapter_display_order,
	topic_display_order,
	lm_display_order
FROM course_stats
GROUP BY learning_material_id, topic_id, chapter_display_order,topic_display_order,lm_display_order) 
UNION
(SELECT
	'NONE' learning_material_id,
	topic_id,
	sum((active_lm > 0)::int)::integer as total_assign_student,
	sum((active_lm > 0 and active_lm = completed_lm)::int)::integer as completed_student,
	coalesce((
			SELECT avg(ls.average_score_raw) as topic_avg_score
			FROM (
				SELECT
				topic_id,
				coalesce(avg(percentage), -1)::integer as average_score,
				avg(percentage)::integer as average_score_raw
				FROM course_stats
				GROUP BY learning_material_id, topic_id
				) ls
			WHERE ls.average_score != -1
			and ls.topic_id = topic_stat.topic_id
			GROUP BY ls.topic_id
		)::smallint,
		-1
	)::integer as average_score,
	0 as average_score_raw,
	'TOPIC' as type,
	chapter_display_order,
	topic_display_order,
	0
FROM (
	SELECT
		topic_id,
		student_id,
		chapter_display_order,
		topic_display_order,
		sum((status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE')::int) as active_lm,
		sum((status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE' and is_completed = true)::int) as completed_lm
	FROM course_stats
	GROUP BY topic_id, student_id,chapter_display_order,topic_display_order
) topic_stat
GROUP BY topic_id,chapter_display_order,topic_display_order)
order by chapter_display_order,topic_display_order,lm_display_order
`

func (r *CourseStudyPlanRepo) ListCourseStatisticV4(ctx context.Context, db database.QueryExecer, args *ListCourseStatisticItemsArgsV3) ([]*TopicStatistic, []*LearningMaterialStatistic, error) {
	topics := []*TopicStatistic{}
	lms := []*LearningMaterialStatistic{}
	itemStatistics := []*ItemStatistic{}

	rows, err := db.Query(ctx, courseStatisticQueryV4, &args.CourseID, &args.StudyPlanID, &args.StudentIDs, &args.TagIDs)

	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	for rows.Next() {
		e := &ItemStatistic{}
		if err := rows.Scan(&e.LearningMaterialID, &e.TopicID, &e.TotalAssignStudent, &e.CompletedStudent, &e.AverageScore, &e.AverageScoreRaw, &e.Type, &e.ChapterDisplayOrder, &e.TopicDisplayOrder, &e.LmDisplayOrder); err != nil {
			return nil, nil, fmt.Errorf("ItemStatistic rows.Scan: %w", err)
		}
		itemStatistics = append(itemStatistics, e)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	for _, item := range itemStatistics {
		if item.Type == "LEARNING_MATERIAL" {
			lms = append(lms, &LearningMaterialStatistic{
				LearningMaterialID: item.LearningMaterialID,
				TopicID:            item.TopicID,
				TotalAssignStudent: item.TotalAssignStudent,
				CompletedStudent:   item.CompletedStudent,
				AverageScore:       item.AverageScore,
				AverageScoreRaw:    item.AverageScoreRaw,
			})
		} else {
			topics = append(topics, &TopicStatistic{
				TopicID:            item.TopicID,
				TotalAssignStudent: item.TotalAssignStudent,
				CompletedStudent:   item.CompletedStudent,
				AverageScore:       item.AverageScore,
			})
		}
	}

	return topics, lms, nil
}
