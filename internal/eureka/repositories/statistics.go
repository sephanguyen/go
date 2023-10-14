package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type StatisticsRepo struct {
}

type StudentTopicProgress struct {
	StudentID        pgtype.Text
	StudyPlanID      pgtype.Text
	ChapterID        pgtype.Text
	TopicID          pgtype.Text
	TopicName        pgtype.Text
	CompletedSPItems pgtype.Int2
	TotalSpItems     pgtype.Int2
	AverageScore     pgtype.Int2
}

func (s *StatisticsRepo) GetStudentTopicProgress(ctx context.Context, db database.QueryExecer, studentID, studyPlanID pgtype.Text) ([]*StudentTopicProgress, error) {
	ctx, span := interceptors.StartSpan(ctx, "StatisticsRepo.GetStudentTopicProgress")
	defer span.End()

	query := `
	SELECT student_id, study_plan_id, chapter_id, topic_id, completed_sp_item, total_sp_item, average_score
	FROM get_student_topic_progress() 
		WHERE student_id = $1
		AND study_plan_id = $2
	`

	resp := make([]*StudentTopicProgress, 0)
	rows, err := db.Query(ctx, query, &studentID, &studyPlanID)
	if err != nil {
		return nil, fmt.Errorf("StatisticsRepo.GetStudentTopicProgress.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		r := new(StudentTopicProgress)
		err := rows.Scan(&r.StudentID, &r.StudyPlanID, &r.ChapterID, &r.TopicID, &r.CompletedSPItems, &r.TotalSpItems, &r.AverageScore)
		if err != nil {
			return nil, fmt.Errorf("StatisticsRepo.GetStudentTopicProgress.Scan: %w", err)
		}

		resp = append(resp, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("StatisticsRepo.GetStudentTopicProgress.Err: %w", err)
	}

	return resp, nil
}

type StudentChapterProgress struct {
	StudentID    pgtype.Text
	StudyPlanID  pgtype.Text
	ChapterID    pgtype.Text
	AverageScore pgtype.Int2
}

func (s *StatisticsRepo) GetStudentChapterProgress(ctx context.Context, db database.QueryExecer, studentID, studyPlanID pgtype.Text) ([]*StudentChapterProgress, error) {
	ctx, span := interceptors.StartSpan(ctx, "StatisticsRepo.GetStudentChapterProgress")
	defer span.End()

	query := `
		SELECT student_id, study_plan_id, chapter_id, average_score
		FROM get_student_chapter_progress()
		WHERE student_id = $1
		AND study_plan_id = $2
	`

	resp := make([]*StudentChapterProgress, 0)
	rows, err := db.Query(ctx, query, &studentID, &studyPlanID)
	if err != nil {
		return nil, fmt.Errorf("StatisticsRepo.GetStudentChapterProgress.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		r := new(StudentChapterProgress)
		err := rows.Scan(&r.StudentID, &r.StudyPlanID, &r.ChapterID, &r.AverageScore)
		if err != nil {
			return nil, fmt.Errorf("StatisticsRepo.GetStudentChapterProgress.Scan: %w", err)
		}
		resp = append(resp, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("StatisticsRepo.GetStudentChapterProgress.Err: %w", err)
	}

	return resp, nil
}

type LearningMaterialProgress struct {
	StudyPlanID         pgtype.Text
	BookID              pgtype.Text
	ChapterID           pgtype.Text
	ChapterDisplayOrder pgtype.Int2
	TopicID             pgtype.Text
	TopicDisplayOrder   pgtype.Int2
	LearningMaterialID  pgtype.Text
	LmDisplayOrder      pgtype.Int2
	Type                pgtype.Text
	Name                pgtype.Text
	AvailableFrom       pgtype.Timestamptz
	AvailableTo         pgtype.Timestamptz
	StartDate           pgtype.Timestamptz
	EndDate             pgtype.Timestamptz
	IsCompleted         pgtype.Bool
	CompletedAt         pgtype.Timestamptz
	Status              pgtype.Text
	SchoolDate          pgtype.Timestamptz
	HighestScore        pgtype.Int2
}

// TODO: validate if the transaction is closed, we need commit drop or not ?
// TODO: join with function student_study_plans instead of join course_study_plans and course_students when that function is completed
const createTempLmProgressQuery = `
create temp table temp_learning_material_progress on commit drop as 
select 
	student_id,
	study_plan_id,
	learning_material_id, 
	type, 
	name, 
	book_id, 
	chapter_id, 
	lalm.chapter_display_order, 
	lalm.topic_id, 
	lalm.topic_display_order, 
	lalm.lm_display_order,
	lalm.available_from,
    lalm.available_to,
    lalm.start_date,
    lalm.end_date,
	gsclm.completed_at is not null as is_completed,
    gsclm.completed_at,
    lalm.status,
    lalm.school_date,
	((mgs.graded_points * 1.0 / mgs.total_points) * 100)::smallint as highest_score
from list_available_learning_material() lalm
join course_study_plans csp using (study_plan_id)
join course_students cs using (course_id, student_id)
join learning_material lm using (learning_material_id)
left join get_student_completion_learning_material() gsclm using (student_id, study_plan_id, learning_material_id)
left join max_graded_score() mgs using (student_id, study_plan_id, learning_material_id)
where student_id = $1::TEXT
and (study_plan_id = $2::TEXT or $2::TEXT is null)
and course_id = $3::TEXT;
`

const createTempTopicProgressQuery = `
create temp table temp_topic_progress on commit drop as
select student_id,
       study_plan_id,
       lmp.chapter_id,
       topic_id,
       t."name" as topic_name,
	   sum(is_completed::int)::smallint as completed_sp_item,
       count(*)::smallint  as total_sp_item,
       (avg(highest_score))::smallint as average_score
from temp_learning_material_progress lmp
join topics t using (topic_id)
group by student_id, study_plan_id, lmp.chapter_id, topic_id, t.name;
`

const createTempChapterProgressQuery = `
create temp table temp_chapter_progress on commit drop as
select student_id,
		study_plan_id,
		chapter_id,
		avg(average_score)::smallint as average_score
	from temp_topic_progress
	group by student_id, study_plan_id, chapter_id;
`

func (s *StatisticsRepo) GetStudentProgress(ctx context.Context, db database.QueryExecer, studentID, studyPlanID, courseID pgtype.Text) ([]*LearningMaterialProgress, []*StudentTopicProgress, []*StudentChapterProgress, error) {
	ctx, span := interceptors.StartSpan(ctx, "StatisticsRepo.GetStudentProgress")
	defer span.End()

	_, err := db.Exec(ctx, createTempLmProgressQuery, &studentID, &studyPlanID, &courseID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.CreateTempBookTree: %w", err)
	}

	_, err = db.Exec(ctx, createTempTopicProgressQuery)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.CreateTempTopicProgress: %w", err)
	}

	_, err = db.Exec(ctx, createTempChapterProgressQuery)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.CreateTempChapterProgress: %w", err)
	}

	studyPlanTreeQuery := `
		select study_plan_id, book_id, chapter_id, chapter_display_order, topic_id, topic_display_order, learning_material_id, lm_display_order, type, name, 
		available_from, available_to, start_date, end_date, is_completed, completed_at, status, school_date, highest_score
		from temp_learning_material_progress
	`

	studyPlanTreeResp := make([]*LearningMaterialProgress, 0)
	studyPlanTreeRows, err := db.Query(ctx, studyPlanTreeQuery)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.BookTreeQuery: %w", err)
	}
	defer studyPlanTreeRows.Close()

	for studyPlanTreeRows.Next() {
		r := new(LearningMaterialProgress)
		err := studyPlanTreeRows.Scan(&r.StudyPlanID, &r.BookID, &r.ChapterID, &r.ChapterDisplayOrder, &r.TopicID, &r.TopicDisplayOrder, &r.LearningMaterialID, &r.LmDisplayOrder, &r.Type, &r.Name,
			&r.AvailableFrom, &r.AvailableTo, &r.StartDate, &r.EndDate, &r.IsCompleted, &r.CompletedAt, &r.Status, &r.SchoolDate, &r.HighestScore)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.BookTreeScan: %w", err)
		}
		studyPlanTreeResp = append(studyPlanTreeResp, r)
	}

	if err := studyPlanTreeRows.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.BookTreeRowsErr: %w", err)
	}

	topicQuery := `
		select student_id, study_plan_id, chapter_id, topic_id, topic_name, completed_sp_item, total_sp_item, average_score
		from temp_topic_progress
	`

	topicResp := make([]*StudentTopicProgress, 0)
	topicRows, err := db.Query(ctx, topicQuery)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.TopicQuery: %w", err)
	}
	defer topicRows.Close()

	for topicRows.Next() {
		r := new(StudentTopicProgress)
		err := topicRows.Scan(&r.StudentID, &r.StudyPlanID, &r.ChapterID, &r.TopicID, &r.TopicName, &r.CompletedSPItems, &r.TotalSpItems, &r.AverageScore)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.TopicScan: %w", err)
		}
		topicResp = append(topicResp, r)
	}

	if err := topicRows.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.TopicRowsErr: %w", err)
	}

	chapterQuery := `
		select student_id, study_plan_id, chapter_id, average_score
		from temp_chapter_progress
	`

	chapterResp := make([]*StudentChapterProgress, 0)
	chapterRows, err := db.Query(ctx, chapterQuery)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.ChapterQuery: %w", err)
	}
	defer chapterRows.Close()

	for chapterRows.Next() {
		r := new(StudentChapterProgress)
		err := chapterRows.Scan(&r.StudentID, &r.StudyPlanID, &r.ChapterID, &r.AverageScore)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.ChapterScan: %w", err)
		}
		chapterResp = append(chapterResp, r)
	}

	if err := chapterRows.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("StatisticsRepo.GetStudentProgress.ChapterRowsErr: %w", err)
	}

	return studyPlanTreeResp, topicResp, chapterResp, nil
}
