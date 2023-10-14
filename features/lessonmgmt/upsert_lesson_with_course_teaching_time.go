package lessonmgmt

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

func (s *Suite) RegisterSomeCourseTeachingTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	args := []string{}
	addedCourseIDs := make(map[string]bool)
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		courseID := stepState.StudentIDWithCourseID[i+1]
		if _, ok := addedCourseIDs[courseID]; ok {
			continue
		}
		addedCourseIDs[courseID] = true
		prepTime, _ := rand.Int(rand.Reader, big.NewInt(300))
		breakTime, _ := rand.Int(rand.Reader, big.NewInt(60))
		args = append(args, fmt.Sprintf("('%s',%d, %d)", courseID, prepTime, breakTime))
	}
	if len(args) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no have courses to register teaching time")
	}

	query := fmt.Sprintf("INSERT INTO course_teaching_time(course_id, preparation_time, break_time) VALUES %s ON CONFLICT DO NOTHING;", strings.Join(args, ","))
	if _, err := s.BobDB.Exec(ctx, query); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert course teaching time: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheLessonsHaveCorrectCourseTeachingTimeInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// get lesson info
	lessonQuery := `select lesson_id, COALESCE(preparation_time,0), COALESCE(break_time,0) from lessons 
	  where scheduler_id = (select scheduler_id from lessons 
						  where lesson_id = $1) order by start_time asc`
	rows, err := s.BobDB.Query(ctx, lessonQuery, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Scan: %w", err)
	}
	defer rows.Close()

	lessonIDs := []string{}
	lessonMap := make(map[string][]int32)
	for rows.Next() {
		l := &domain.Lesson{}
		err := rows.Scan(
			&l.LessonID,
			&l.PreparationTime,
			&l.BreakTime,
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Scan: %w", err)
		}
		lessonMap[l.LessonID] = []int32{l.PreparationTime, l.BreakTime}
		lessonIDs = append(lessonIDs, l.LessonID)
	}

	//  get expect lesson course's info
	expectedLessonquery := `SELECT l.lesson_id,
            COALESCE((
              case when l.teaching_method = 'LESSON_TEACHING_METHOD_GROUP'
              then (select ctt.preparation_time from course_teaching_time ctt 
                  where ctt.course_id = l.course_id and ctt.deleted_at is null
                  group by ctt.course_id, ctt.break_time order by ctt.break_time desc, ctt.preparation_time desc limit 1)
              else (select ctt.preparation_time 
                  from course_teaching_time ctt
                  join lessons_courses lc on ctt.course_id = lc.course_id
                  where lc.lesson_id = l.lesson_id and ctt.deleted_at is null
                  group by ctt.course_id, ctt.break_time order by ctt.break_time desc, ctt.preparation_time desc limit 1)
              end
            ),0) AS preparation_time,
            COALESCE((
              case when l.teaching_method = 'LESSON_TEACHING_METHOD_GROUP'
              then (select ctt.break_time from course_teaching_time ctt 
                  where ctt.course_id = l.course_id and ctt.deleted_at is null
                  group by ctt.course_id, ctt.break_time order by ctt.break_time desc, ctt.preparation_time desc limit 1)
              else (select ctt.break_time 
                  from course_teaching_time ctt
                  join lessons_courses lc on ctt.course_id = lc.course_id
                  where lc.lesson_id = l.lesson_id and ctt.deleted_at is null
                  group by ctt.course_id, ctt.break_time order by ctt.break_time desc, ctt.preparation_time desc limit 1)
              end
            ),0) AS break_time
            FROM lessons l WHERE lesson_id = any($1)`

	rows, err = s.BobDBTrace.Query(ctx, expectedLessonquery, lessonIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("get expected lessons: %w", err)
	}
	defer rows.Close()

	expectedLessonsMap := make(map[string][]int32)
	for rows.Next() {
		l := &domain.Lesson{}
		err := rows.Scan(
			&l.LessonID,
			&l.PreparationTime,
			&l.BreakTime,
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Scan: %w", err)
		}
		expectedLessonsMap[l.LessonID] = []int32{l.PreparationTime, l.BreakTime}
	}

	//  compare with course
	for lessonID, teachingTime := range lessonMap {
		expectedTeachingTime, ok := expectedLessonsMap[lessonID]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson(%s) doesn't have course teaching time info", lessonID)
		}
		if teachingTime[0] != expectedTeachingTime[0] || teachingTime[1] != expectedTeachingTime[1] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson(%s) doesn't correct teaching time. \nexpect %v but got %v", lessonID, teachingTime, expectedTeachingTime)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
