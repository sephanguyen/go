package bob

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/manabie-com/backend/features/helper"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/gogo/protobuf/types"
	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aListOfLessonsAreExistedInDBOfWithStartTimeAndEndTime(ctx context.Context, lesson_opt, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	courseID := "course-live-teacher-1"
	courseID2 := "course-live-teacher-1"
	if lesson_opt == "above teacher and belong to multy course" {
		courseID = "course-live-teacher-5"
		courseID2 = "course-live-teacher-6"
	}
	if lesson_opt == "above teacher and belong to single course" {
		courseID = "course-live-teacher-4"
		courseID2 = "course-live-teacher-4"
	}

	courseID += stepState.Random
	courseID2 += stepState.Random
	classID := idutil.ULIDNow()

	// create lesson group
	lg := &entities_bob.LessonGroup{}
	database.AllNullEntity(lg)
	lg.MediaIDs = database.TextArray(stepState.MediaIDs)
	lg.CourseID = database.Text(courseID)
	err = (&bob_repo.LessonGroupRepo{}).Create(ctx, s.DB, lg)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.LessonGroupRepo.Create: %w", err)
	}

	lesson := &entities_bob.Lesson{}
	database.AllNullEntity(lesson)

	err = multierr.Combine(
		lesson.LessonID.Set(s.newID()),
		lesson.CourseID.Set(courseID),
		lesson.TeacherID.Set(stepState.CurrentTeacherID),
		lesson.CreatedAt.Set(timeutil.Now()),
		lesson.UpdatedAt.Set(timeutil.Now()),
		lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
		lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
		lesson.StreamLearnerCounter.Set(database.Int4(0)),
		lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
		lesson.StartTime.Set(startDate),
		lesson.EndTime.Set(endDate),
		lesson.LessonGroupID.Set(lg.LessonGroupID),
		lesson.ClassID.Set(classID),
		lesson.SchedulingStatus.Set(entities_bob.LessonSchedulingStatusPublished),
		lesson.CenterID.Set(constants.ManabieOrgLocation),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err := lesson.Normalize(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
	}

	cmdTag, err := database.Insert(ctx, lesson, s.DBPostgres.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if cmdTag.RowsAffected() != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
	}

	if lesson_opt == "above teacher and belong to multy course" {
		sql := `INSERT INTO lessons_courses
				(lesson_id, course_id, created_at)
				VALUES ($1, $2, $4), ($1, $3, $4)`
		_, err = s.DB.Exec(ctx, sql, lesson.LessonID, database.Text(courseID),
			database.Text(courseID2),
			database.Timestamptz(time.Now()))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_course, err = %v", err)
		}
	}

	stepState.CurrentLessonID = lesson.LessonID.String

	if err := (&bob_repo.LessonRepo{}).UpsertLessonMembers(ctx, db, lesson.LessonID, database.TextArray(stepState.StudentIds)); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) genPresetStudyPlanWeekly(ctx context.Context, startDate, endDate time.Time, lessonID, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var topicID string
	if err := s.DB.QueryRow(ctx, "SELECT topic_id FROM topics LIMIT 1").Scan(&topicID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var presetStudyPlanID string
	if err := s.DB.QueryRow(ctx, "SELECT preset_study_plan_id FROM courses c WHERE c.course_id =$1", courseID).Scan(&presetStudyPlanID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	presetStudyPlanWeekly := &entities_bob.PresetStudyPlanWeekly{}
	database.AllNullEntity(presetStudyPlanWeekly)

	week := rand.Intn(10000)
	err := multierr.Combine(
		presetStudyPlanWeekly.ID.Set(s.newID()),
		presetStudyPlanWeekly.StartDate.Set(startDate),
		presetStudyPlanWeekly.EndDate.Set(endDate),
		presetStudyPlanWeekly.PresetStudyPlanID.Set(presetStudyPlanID),
		presetStudyPlanWeekly.TopicID.Set(topicID),
		presetStudyPlanWeekly.Week.Set(week),
		presetStudyPlanWeekly.CreatedAt.Set(timeutil.Now()),
		presetStudyPlanWeekly.UpdatedAt.Set(timeutil.Now()),
		presetStudyPlanWeekly.LessonID.Set(lessonID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	cmdTag, err := database.Insert(ctx, presetStudyPlanWeekly, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if cmdTag.RowsAffected() != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert preset study plan weekly")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentRetrieveLiveLessonWithStartTimeAndEndTime(ctx context.Context, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.CurrentUserID = stepState.CurrentTeacherID
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		From: &types.Timestamp{Seconds: startDate.Unix()},
		To:   &types.Timestamp{Seconds: endDate.Unix()},
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveLiveLesson(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentRetrieveLiveLessonByCourseWithStartTimeAndEndTime(ctx context.Context, courseID, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.CurrentUserID = stepState.CurrentTeacherID
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	courseID += stepState.Random

	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		CourseIds: []string{courseID},
		From:      &types.Timestamp{Seconds: startDate.Unix()},
		To:        &types.Timestamp{Seconds: endDate.Unix()},
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveLiveLesson(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) teacherRetrieveLiveLessonWithStartTimeAndEndTime(ctx context.Context, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	token, err := s.generateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		From: &types.Timestamp{Seconds: startDate.Unix()},
		To:   &types.Timestamp{Seconds: endDate.Unix()},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveLiveLesson(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) teacherRetrieveLiveLessonByCourseWithStartTimeAndEndTime(ctx context.Context, courseID, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	courseID += stepState.Random

	token, err := s.generateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		CourseIds: []string{courseID},
		From:      &types.Timestamp{Seconds: startDate.Unix()},
		To:        &types.Timestamp{Seconds: endDate.Unix()},
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveLiveLesson(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnLessonsMustHaveCorrectTopic(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	for _, lesson := range rsp.Lessons {
		query := `SELECT COUNT(*) FROM preset_study_plans_weekly pspw WHERE pspw.preset_study_plan_weekly_id = $1 AND pspw.topic_id = $2`
		if lesson.Topic == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson does not return topic")
		}
		var count int
		if err := s.DB.QueryRow(ctx, query, &lesson.PresetStudyPlanWeeklyIds, &lesson.Topic.Id).Scan(&count); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if count != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't find plan contain topic")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnLessonsMustHaveCorrectTeacherProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	for _, lesson := range rsp.Lessons {
		if len(lesson.Teacher) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson return does not contain teacher")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnLessonsMustHaveCorrectStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	for _, lesson := range rsp.Lessons {
		status := lesson.Status
		if lesson.EndTime.Nanos < int32(time.Now().Nanosecond()) && status == pb.LESSON_STATUS_COMPLETED {
			continue
		}
		if lesson.StartTime.Nanos >= int32(time.Now().Nanosecond()) && status == pb.LESSON_STATUS_IN_PROGRESS {
			continue
		}
		if lesson.StartTime.Nanos < int32(time.Now().Nanosecond()) && status == pb.LESSON_STATUS_NOT_STARTED {
			continue
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s return wrong status: %s", lesson.LessonId, status.String())
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) checkUserClassIDInLesson(ctx context.Context, lesson *pb.Lesson) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(lesson.UserClassIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	log.Print(lesson)
	var count int64
	query := `SELECT COUNT(*) FROM courses_classes cc LEFT JOIN class_members cm ON cc.class_id = cm.class_id
		WHERE cc.course_id = $1 AND cm.user_id = $2 AND cm.class_id = ANY($3)`
	if err := s.DB.QueryRow(ctx, query, lesson.CourseId, &stepState.CurrentUserID, &lesson.UserClassIds).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("class_id return does not belong to lesson")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnLessons(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	if len(rsp.Lessons) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob must return lessons")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentRetrieveLiveLessonWithInvalidTimePeriod(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		From: &types.Timestamp{Seconds: timeutil.Now().Add(-72 * time.Hour).Unix()},
		To:   &types.Timestamp{Seconds: timeutil.Now().Add(-49 * time.Hour).Unix()},
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveLiveLesson(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) teacherRetrieveLiveLessonWithInvalidTimePeriod(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.generateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		From: &types.Timestamp{Seconds: timeutil.Now().Add(-72 * time.Hour).Unix()},
		To:   &types.Timestamp{Seconds: timeutil.Now().Add(-49 * time.Hour).Unix()},
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveLiveLesson(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnEmptyLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	if len(rsp.Lessons) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob must return empty lessons")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobReturnResultLiveLessonForStudent(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if result == "empty" {
		return s.bobMustReturnEmptyLiveLesson(ctx)
	}
	if result == "correct" {
		return s.bobMustReturnCorrectLiveLessonsForStudent(ctx)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobReturnResultLiveLessonForTeacher(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if result == "empty" {
		return s.bobMustReturnEmptyLiveLesson(ctx)
	}
	if result == "correct" {
		return s.bobMustReturnCorrectLiveLessonsForTeacher(ctx)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnCorrectUserClassIdInLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	courseMapIDs := make(map[string]bool)
	for _, lesson := range rsp.Lessons {
		found, ok := courseMapIDs[lesson.CourseId]
		if found && ok {
			continue
		}
		courseMapIDs[lesson.CourseId] = true
		ctx, err := s.checkUserClassIDInLesson(ctx, lesson)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectCourseInLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.RetrieveLiveLessonRequest)
	if len(req.CourseIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	for _, lesson := range rsp.GetLessons() {
		if !contains(req.CourseIds, lesson.CourseId) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match course retrieve\n%s\n%s", lesson.LessonId, req.CourseIds, lesson.CourseId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectLessonMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	lessonIDs := []string{}
	for _, lesson := range rsp.Lessons {
		lessonIDs = append(lessonIDs, lesson.LessonId)
		ctx, err := s.checkUserClassIDInLesson(ctx, lesson)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
	}
	t, _ := jwt.ParseString(stepState.AuthToken)
	query := `SELECT result_lesson
	FROM UNNEST($1::TEXT[]) AS result_lesson
	LEFT JOIN lesson_members lm ON result_lesson=lm.lesson_id AND lm.user_id =$2
	WHERE lm.lesson_id IS NULL`
	rows, err := s.DB.Query(ctx, query, database.TextArray(lessonIDs), t.Subject())
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	defer rows.Close()

	wrongLessons := []string{}
	for rows.Next() {
		var wrongLesson string
		if err := rows.Scan(&wrongLesson); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Scan :%v", err)
		}
		wrongLessons = append(wrongLessons, wrongLesson)
	}
	if len(wrongLessons) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student not a member of these lesson %s", wrongLessons)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnCorrectLiveLessonsForStudent(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.bobMustReturnLessons(ctx)
	ctx, err2 := s.returnLessonsMustHaveCorrectStatus(ctx)
	ctx, err3 := s.returnLessonsMustHaveCorrectTeacherProfile(ctx)
	ctx, err4 := s.returnCorrectUserClassIdInLesson(ctx)
	ctx, err5 := s.resultCorrectLessonMember(ctx)
	err := multierr.Combine(err1, err2, err3, err4, err5)
	return ctx, err
}
func (s *suite) bobMustReturnCorrectLiveLessonsForTeacher(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.bobMustReturnLessons(ctx)
	ctx, err2 := s.returnLessonsMustHaveCorrectStatus(ctx)
	ctx, err3 := s.returnLessonsMustHaveCorrectTeacherProfile(ctx)
	ctx, err4 := s.returnCorrectUserClassIdInLesson(ctx)
	ctx, err5 := s.resultCorrectCourseInLesson(ctx)
	err := multierr.Combine(err1, err2, err3, err4, err5)
	return ctx, err
}
func (s *suite) currentStudentAssignedToAboveLessons(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	t, _ := jwt.ParseString(stepState.AuthToken)

	lessonMember := &entities_bob.LessonMember{}
	database.AllNullEntity(lessonMember)
	if err := multierr.Combine(
		lessonMember.LessonID.Set(stepState.CurrentLessonID),
		lessonMember.UserID.Set(t.Subject()),
		lessonMember.CreatedAt.Set(now),
		lessonMember.UpdatedAt.Set(now),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if _, err := database.Insert(ctx, lessonMember, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) adminRetrieveLiveLesson(ctx context.Context, limitStr, offset string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	req := &bpb.RetrieveLessonsRequest{
		Paging: &cpb.Paging{
			Limit: uint32(limit),
		},
	}

	if offset != NIL_VALUE {
		offset = stepState.Random + "_" + offset
		req = &bpb.RetrieveLessonsRequest{
			Paging: &cpb.Paging{
				Limit:  uint32(limit),
				Offset: &cpb.Paging_OffsetString{OffsetString: offset},
			},
		}
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewLessonReaderServiceClient(s.Conn).RetrieveLessons(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfLessonOfSchoolAreExistedInDB(ctx context.Context, strFromID, strToID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now().Add(1000 * 24 * time.Hour)
	stepState.TimeRandom = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
	fromID, err := strconv.Atoi((strings.Split(strFromID, "_"))[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson with outline init param")
	}
	toID, err := strconv.Atoi((strings.Split(strToID, "_"))[1])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson with outline init param")
	}
	for i := fromID; i < toID; i++ {
		courseIDs := []string{}
		presetStudyPlanIDs := []string{}
		teacherIDs := []string{}
		for j := 0; j < 2; j++ {
			courseID := fmt.Sprintf("course_%s_%d_%d", stepState.Random, i, j)
			timeAny := database.Timestamptz(time.Now())
			courseName := courseID
			school := stepState.RandSchoolID
			grade := database.Int2(10)
			sql := `INSERT INTO courses
		(course_id, name, grade, created_at, updated_at, school_id)
		VALUES ($1, $2, $3, $4, $5, $6)`
			_, err := s.DB.Exec(ctx, sql, database.Text(courseID), courseName, grade, timeAny, timeAny, school)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err with school %d: %s", school, err)
			}
			courseIDs = append(courseIDs, courseID)

			sql = `INSERT INTO preset_study_plans
		(preset_study_plan_id, name, grade, subject, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
			_, err = s.DB.Exec(ctx, sql, database.Text("StudyPlan - "+courseID), courseName, grade, "MATH", timeAny, timeAny)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init preset_study_plans err: %s", err)
			}
			presetStudyPlanIDs = append(presetStudyPlanIDs, "StudyPlan - "+courseID)
		}
		for j := 0; j < 2; j++ {
			teacherID := fmt.Sprintf("teacher_%s_%d_%d", stepState.Random, i, j)
			ctx, err := s.aValidTeacherProfileWithId(ctx, teacherID, int32(stepState.RandSchoolID))
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init teacher%s", err)
			}
			teacherIDs = append(teacherIDs, teacherID)
		}
		lesson := &entities_bob.Lesson{}
		database.AllNullEntity(lesson)
		lessonID := fmt.Sprintf("%s_id_%d", stepState.Random, i)
		lessonName := fmt.Sprintf("name_%d", i)
		classID := idutil.ULIDNow()

		err := multierr.Combine(
			lesson.LessonID.Set(lessonID),
			lesson.Name.Set(lessonName),
			lesson.CourseID.Set(courseIDs[0]),
			lesson.TeacherID.Set(teacherIDs[0]),
			lesson.CreatedAt.Set(stepState.TimeRandom.Add(time.Duration(i)*time.Microsecond)),
			lesson.UpdatedAt.Set(timeutil.Now()),
			lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
			lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
			lesson.StreamLearnerCounter.Set(database.Int4(0)),
			lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
			lesson.StartTime.Set(timeutil.Now()),
			lesson.EndTime.Set(timeutil.Now()),
			lesson.ClassID.Set(classID),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson err: %s", err)
		}
		if err := lesson.Normalize(); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
		}

		cmdTag, err := database.Insert(ctx, lesson, s.DB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
		if cmdTag.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
		}

		for j := 0; j < 2; j++ {
			sql := `INSERT INTO lessons_teachers 
		(lesson_id, teacher_id, created_at)
		VALUES ($1, $2, $3)`
			_, err = s.DB.Exec(ctx, sql, lesson.LessonID, database.Text(teacherIDs[j]), database.Timestamptz(time.Now()))
			if err != nil {
				return StepStateToContext(ctx, stepState), err

			}
		}

		if i%2 == 0 {
			for j := 0; j < 2; j++ {
				timeAny := database.Timestamptz(time.Now())
				sql := `INSERT INTO lessons_courses
		(lesson_id, course_id, created_at)
		VALUES ($1, $2, $3)`
				_, err = s.DB.Exec(ctx, sql, lesson.LessonID, database.Text(courseIDs[j]), timeAny)
				if err != nil {
					return StepStateToContext(ctx, stepState), err

				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bobMustReturnListLesson(ctx context.Context, total, fromID, toID, limit, next, pre string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	rsp := stepState.Response.(*bpb.RetrieveLessonsResponse)

	if total != "nil" {
		totalInt, err := strconv.Atoi(total)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("err parse expect value")
		}
		if totalInt == 0 {
			if rsp.TotalLesson != uint32(totalInt) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("must return empty list\n expect total: %d, actual total : %d", totalInt, rsp.TotalLesson)
			}
			return StepStateToContext(ctx, stepState), nil
		} else if rsp.TotalLesson != uint32(totalInt) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong list return\n expect total: %d, actual: %d", totalInt, rsp.TotalLesson)
		}
	}

	if rsp.Items[0].Id != (stepState.Random + "_" + fromID) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong first items return \n expect: %s, actual: %s", (stepState.Random + "_" + fromID), rsp.Items[0].Id)
	}

	if rsp.Items[len(rsp.Items)-1].Id != (stepState.Random + "_" + toID) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong items return \n expect: %s, actual: %s", (stepState.Random + "_" + toID), rsp.Items[len(rsp.Items)-1].Id)
	}

	if int(rsp.NextPage.Limit) != limitInt || int(rsp.PreviousPage.Limit) != limitInt {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong limit return")
	}

	if rsp.NextPage.GetOffsetString() != (stepState.Random + "_" + next) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong next offset return \n expect: %s, actual: %s", (stepState.Random + "_" + next), rsp.NextPage.GetOffsetString())
	}

	if pre == "nil" {
		pre = ""
	} else {
		pre = stepState.Random + "_" + pre
	}

	if rsp.PreviousPage.GetOffsetString() != pre {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong previous offset return \n expect: %s, actual: %s", pre, rsp.PreviousPage.GetOffsetString())
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aListOfLessonsOfSchoolAreExistedInDB(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	teacherID := s.newID()
	ctx, err := s.aValidTeacherProfileWithId(ctx, teacherID, int32(schoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init teacher%s", err)
	}

	presetStudyPlanIDs := []string{}
	randonString := s.newID()
	stepState.FilterFromTime = time.Now()
	stepState.FilterToTime = time.Now()
	for i := 1; i < 4; i++ {
		courseID := fmt.Sprintf("course_%d_%s", i, randonString)
		courseName := "name-" + courseID
		sql := `INSERT INTO courses
		(course_id, name, grade, created_at, updated_at, school_id)
		VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := s.DB.Exec(ctx, sql, database.Text(courseID), courseName, 10, time.Now(), time.Now(), schoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err: %s", err)
		}
		stepState.FilterCourseIDs = append(stepState.FilterCourseIDs, courseID)

		presetStudyPlanID := "StudyPlan - " + courseID
		sql = `INSERT INTO preset_study_plans
		(preset_study_plan_id, name, grade, subject, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = s.DB.Exec(ctx, sql, presetStudyPlanID, courseName, 10, "MATH", time.Now(), time.Now())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init preset_study_plans err: %s", err)
		}
		presetStudyPlanIDs = append(presetStudyPlanIDs, presetStudyPlanID)
	}

	lessonStatus := []cpb.LessonStatus{cpb.LessonStatus_LESSON_STATUS_COMPLETED, cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS, cpb.LessonStatus_LESSON_STATUS_NOT_STARTED}
	for i, course := range stepState.FilterCourseIDs {
		for _, status := range lessonStatus {
			courseIDs := make([]string, 3)
			addCourseIDs := [][]string{}
			copy(courseIDs, stepState.FilterCourseIDs)
			courseIDs = append(courseIDs[:i], courseIDs[i+1:]...)
			addCourseIDs = append(addCourseIDs, []string{})
			addCourseIDs = append(addCourseIDs, []string{courseIDs[0]})
			addCourseIDs = append(addCourseIDs, []string{courseIDs[1]})
			addCourseIDs = append(addCourseIDs, []string{courseIDs[0], courseIDs[1]})
			for _, addCourseID := range addCourseIDs {
				s.genLesson(ctx, course, teacherID, addCourseID, status, presetStudyPlanIDs[i], schoolID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) genLesson(ctx context.Context, courseID, teacherID string, addCourseIDs []string, status cpb.LessonStatus, presetStudyPlanID string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	addCourseIDs = append(addCourseIDs, courseID)

	lesson := &entities_bob.Lesson{}
	database.AllNullEntity(lesson)
	lessonID := s.newID()
	lessonName := fmt.Sprintf("name_%s", lessonID)
	classID := idutil.ULIDNow()

	err := multierr.Combine(
		lesson.LessonID.Set(lessonID),
		lesson.Name.Set(lessonName),
		lesson.CourseID.Set(courseID),
		lesson.TeacherID.Set(teacherID),
		lesson.CreatedAt.Set(time.Now()),
		lesson.UpdatedAt.Set(timeutil.Now()),
		lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
		lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
		lesson.StreamLearnerCounter.Set(database.Int4(0)),
		lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
		lesson.StartTime.Set(timeutil.Now()),
		lesson.EndTime.Set(timeutil.Now()),
		lesson.ClassID.Set(classID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson err: %s", err)
	}

	if err := lesson.Normalize(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson.Normalize err: %s", err)
	}

	cmdTag, err := database.Insert(ctx, lesson, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if cmdTag.RowsAffected() != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson")
	}

	for _, course := range addCourseIDs {
		sql := `INSERT INTO lessons_courses
		(lesson_id, course_id, created_at)
		VALUES ($1, $2, $3)`
		_, err = s.DB.Exec(ctx, sql, lessonID, course, time.Now())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_course: %s", err)
		}

		topicID := s.newID()
		topicName := database.Text("Topic - " + lessonName + " - " + courseID)
		sql = `INSERT INTO topics
		(topic_id, name, grade, subject, topic_type, created_at, updated_at, school_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err := s.DB.Exec(ctx, sql, topicID, topicName, 10, "MATH", "TOPIC_TYPE_LIVE_LESSON", time.Now(), time.Now(), schoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init topic: %s", err)
		}

		startDate := database.Timestamptz(time.Now().Add(-1 * time.Hour))
		endDate := database.Timestamptz(time.Now().Add(1 * time.Hour))
		switch status {
		case cpb.LessonStatus_LESSON_STATUS_COMPLETED:
			startDate = database.Timestamptz(time.Now().Add(-10 * time.Hour))
			endDate = database.Timestamptz(time.Now().Add(-8 * time.Hour))
			if startDate.Time.Before(stepState.FilterFromTime) {
				stepState.FilterFromTime = startDate.Time
			}
		case cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS:
		case cpb.LessonStatus_LESSON_STATUS_NOT_STARTED:
			startDate = database.Timestamptz(time.Now().Add(8 * time.Hour))
			endDate = database.Timestamptz(time.Now().Add(10 * time.Hour))
			if endDate.Time.After(stepState.FilterToTime) {
				stepState.FilterToTime = endDate.Time
			}
		}
		sql = `INSERT INTO preset_study_plans_weekly
		(preset_study_plan_weekly_id, preset_study_plan_id, topic_id, week, created_at, updated_at, lesson_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		_, err = s.DB.Exec(ctx, sql, s.newID(), presetStudyPlanID, topicID, 1, time.Now(), time.Now(), lessonID, startDate, endDate)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init preset_study_plans_weekly: %s", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) adminRetrieveLiveLessonWithFilter(ctx context.Context, numCourse, status, from, to, keyWord string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.FilterCourseIDs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson not init yet")
	}

	filter := &bpb.RetrieveLessonsFilter{}

	if len(numCourse) > 0 {
		if numCourse == "1" {
			filter.CourseIds = stepState.FilterCourseIDs[0:1]
		}
		if numCourse == "2" {
			filter.CourseIds = stepState.FilterCourseIDs[0:2]
		}
	}

	if len(status) > 0 {
		if strings.Contains(status, "Completed") {
			filter.LessonStatus = append(filter.LessonStatus, cpb.LessonStatus_LESSON_STATUS_COMPLETED)
		}
		if strings.Contains(status, "InProgress") {
			filter.LessonStatus = append(filter.LessonStatus, cpb.LessonStatus_LESSON_STATUS_IN_PROGRESS)
		}
		if strings.Contains(status, "NotStarted") {
			filter.LessonStatus = append(filter.LessonStatus, cpb.LessonStatus_LESSON_STATUS_NOT_STARTED)
		}
	}
	stepState.FilterFromTime = stepState.FilterFromTime.Add(1 * time.Hour)
	stepState.FilterToTime = time.Now()
	if len(from) > 0 {
		filter.StartTime = timestamppb.New(stepState.FilterFromTime)
	}

	if len(to) > 0 {
		filter.EndTime = timestamppb.New(stepState.FilterToTime)
	}

	req := &bpb.RetrieveLessonsRequest{
		Paging: &cpb.Paging{
			Limit: uint32(20),
		},
		Filter:  filter,
		Keyword: keyWord,
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewLessonReaderServiceClient(s.Conn).RetrieveLessons(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnCorrectListLessonWith(ctx context.Context, numCourse, status, from, to, keyWord string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.resultCorrectCourse(ctx, numCourse)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with course filter: %s", err)
	}
	ctx, err = s.resultCorrectStatus(ctx, status)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with status filter: %s", err)
	}
	ctx, err = s.resultCorrectTime(ctx, from, to)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with time filter: %s", err)
	}
	ctx, err = s.resultCorrectKeyword(ctx, keyWord)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with keyword: %s", err)
	}
	ctx, err = s.resultCorrectLessonType(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with lesson type: %s", err)
	}
	ctx, err = s.resultCorrectClassID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with class id: %s", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectCourse(ctx context.Context, numCourse string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(numCourse) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponse)
	if numCourse == "1" {
		for _, lesson := range rsp.GetItems() {
			if !contains(lesson.CourseIds, stepState.FilterCourseIDs[0]) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match course filter", lesson.Id)
			}
		}
	}
	if numCourse == "2" {
		for _, lesson := range rsp.GetItems() {
			if !contains(lesson.CourseIds, stepState.FilterCourseIDs[0]) && !contains(lesson.CourseIds, stepState.FilterCourseIDs[1]) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match course filter", lesson.Id)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(status) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		if strings.Contains(status, "Completed") &&
			lesson.EndTime.AsTime().Before(time.Now()) {
			continue
		}
		if strings.Contains(status, "InProgress") &&
			lesson.StartTime.AsTime().Before(time.Now()) &&
			lesson.EndTime.AsTime().After(time.Now()) {
			continue
		}
		if strings.Contains(status, "NotStarted") &&
			lesson.StartTime.AsTime().After(time.Now()) {
			continue
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match status filter", lesson.Id)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectTime(ctx context.Context, from, to string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(from) == 0 && len(to) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponse)
	if len(from) > 0 {
		for _, lesson := range rsp.GetItems() {
			if lesson.EndTime.AsTime().Before(stepState.FilterFromTime) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match fromtime filter", lesson.Id)
			}
		}
	}
	if len(to) > 0 {
		for _, lesson := range rsp.GetItems() {
			if lesson.StartTime.AsTime().After(stepState.FilterToTime) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match totime filter", lesson.Id)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectKeyword(ctx context.Context, keyWord string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(keyWord) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		if !strings.Contains(strings.ToLower(lesson.Name), strings.ToLower(keyWord)) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s name %s not match keyword %s", lesson.Id, lesson.Name, keyWord)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectLessonType(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*bpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		// if field is not null, check for inputs, else ignore
		if len(lesson.LessonType.String()) > 0 {
			if lesson.LessonType.String() != cpb.LessonType_LESSON_TYPE_OFFLINE.String() &&
				lesson.LessonType.String() != cpb.LessonType_LESSON_TYPE_ONLINE.String() &&
				lesson.LessonType.String() != cpb.LessonType_LESSON_TYPE_NONE.String() &&
				lesson.LessonType.String() != cpb.LessonType_LESSON_TYPE_HYBRID.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s lesson_type %s not match lesson_type '%s', '%s', '%s' and '%s'",
					lesson.Id, lesson.LessonType,
					cpb.LessonType_LESSON_TYPE_OFFLINE.String(),
					cpb.LessonType_LESSON_TYPE_ONLINE.String(),
					cpb.LessonType_LESSON_TYPE_NONE.String(),
					cpb.LessonType_LESSON_TYPE_HYBRID.String())
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectClassID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponse)
	for _, lesson := range rsp.GetItems() {
		if len(lesson.ClassId) > 0 {
			//check if classID can be parsed, if not, invalid uuid
			_, err := uuid.Parse(lesson.ClassId)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s class_id %s return incorrect class_id: %s", lesson.Id, lesson.ClassId, err)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
