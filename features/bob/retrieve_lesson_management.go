package bob

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aSignedInWithRandomID(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RandSchoolID = rand.Intn(2999) + 1
	s.aSignedInWithSchool(ctx, role, stepState.RandSchoolID)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminRetrieveLiveLessonManagement(ctx context.Context, lessonTime, limitStr, offset string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &bpb.RetrieveLessonsRequestV2{
		Paging: &cpb.Paging{
			Limit: uint32(limit),
		},
		LessonTime:  bpb.LessonTime(bpb.LessonTime_value[lessonTime]),
		CurrentTime: timestamppb.Now(),
	}

	if offset != NIL_VALUE {
		offset = stepState.Random + "_" + offset
		req = &bpb.RetrieveLessonsRequestV2{
			Paging: &cpb.Paging{
				Limit:  uint32(limit),
				Offset: &cpb.Paging_OffsetString{OffsetString: offset},
			},
			LessonTime:  bpb.LessonTime(bpb.LessonTime_value[lessonTime]),
			CurrentTime: timestamppb.Now(),
		}
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.Conn).RetrieveLessons(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bobMustReturnListLessonManagement(ctx context.Context, total, fromID, toID, limit, next, pre string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	if total != NIL_VALUE {
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
	if len(rsp.Items) > 0 {
		if rsp.Items[0].Id != (stepState.Random + "_" + fromID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong first items return \n expect: %s, actual: %s", (stepState.Random + "_" + fromID), rsp.Items[0].Id)
		}

		if rsp.Items[len(rsp.Items)-1].Id != (stepState.Random + "_" + toID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong items return \n expect: %s, actual: %s", (stepState.Random + "_" + toID), rsp.Items[len(rsp.Items)-1].Id)
		}

		if rsp.NextPage.GetOffsetString() != (stepState.Random + "_" + next) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong next offset return \n expect: %s, actual: %s", (stepState.Random + "_" + next), rsp.NextPage.GetOffsetString())
		}
	}

	if int(rsp.NextPage.Limit) != limitInt || int(rsp.PreviousPage.Limit) != limitInt {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong limit return")
	}

	if pre == NIL_VALUE {
		pre = ""
	} else {
		pre = stepState.Random + "_" + pre
	}

	if rsp.PreviousPage.GetOffsetString() != pre {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong previous offset return \n expect: %s, actual: %s", pre, rsp.PreviousPage.GetOffsetString())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfLessonManagementOfSchoolAreExistedInDB(ctx context.Context, lessonTime, strFromID, strToID string) (context.Context, error) {
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
	courseRepo := repositories.CourseRepo{}
	for i := fromID; i < toID; i++ {
		duration := time.Duration(i)
		startTime := timeutil.Now().Add(duration * time.Hour)
		if lessonTime == bpb.LessonTime_LESSON_TIME_PAST.String() {
			startTime = timeutil.Now().Add(-duration * time.Hour)
		}
		courseIDs := []string{}
		teacherIDs := []string{}
		school := stepState.RandSchoolID
		for j := 0; j < 2; j++ {
			courseID := fmt.Sprintf("course_%s_%d_%d", stepState.Random, i, j)
			timeAny := database.Timestamptz(time.Now())
			courseName := courseID
			course := &entities_bob.Course{}
			database.AllNullEntity(course)
			err := multierr.Combine(
				course.ID.Set(courseID),
				course.Name.Set(courseName),
				course.CreatedAt.Set(timeAny),
				course.UpdatedAt.Set(timeAny),
				course.DeletedAt.Set(nil),
				course.Grade.Set(3),
				course.StartDate.Set(time.Now().Add(2*time.Hour)),
				course.SchoolID.Set(school),
				course.Status.Set("COURSE_STATUS_NONE"),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			err = courseRepo.Create(ctx, s.DB, course)

			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err with school %d: %s", school, err)
			}
			courseIDs = append(courseIDs, courseID)
		}
		for j := 0; j < 2; j++ {
			teacherID := fmt.Sprintf("teacher_%s_%d_%d", stepState.Random, i, j)
			ctx, err := s.aValidTeacherProfileWithId(ctx, teacherID, int32(school))
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
			lesson.StartTime.Set(startTime),
			lesson.EndTime.Set(startTime),
			lesson.TeachingMethod.Set(database.Text(cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL.String())),
			lesson.TeachingMedium.Set(database.Text(cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE.String())),
			lesson.ClassID.Set(classID),
			lesson.CenterID.Set(idutil.ULIDNow()),
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
		updateResourcePath := "UPDATE lessons SET resource_path = $1 WHERE lesson_id = $2"
		_, err = s.DB.Exec(ctx, updateResourcePath, database.Text(fmt.Sprint(school)), database.Text(lessonID))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for j := 0; j < 2; j++ {
			sql := `insert into lessons_teachers 
		(lesson_id, teacher_id, created_at)
		VALUES ($1, $2, $3)`
			_, err = s.DB.Exec(ctx, sql, lesson.LessonID, database.Text(teacherIDs[j]), database.Timestamptz(time.Now()))
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			stepState.FilterTeacherIDs = append(stepState.FilterTeacherIDs, teacherIDs[j])
		}

		if i%2 == 0 {
			for j := 0; j < 2; j++ {
				timeAny := database.Timestamptz(time.Now())
				sql := `insert into lessons_courses
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

func (s *suite) aListOfLessonsManagementOfSchoolAreExistedInDB(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// init teacher
	teacherIDs := []string{}
	for i := 1; i <= 2; i++ {
		teacherID := s.newID()
		ctx, err := s.aValidTeacherProfileWithId(ctx, teacherID, int32(schoolID))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init teacher%s", err)
		}
		teacherIDs = append(teacherIDs, teacherID)
	}
	stepState.FilterTeacherIDs = teacherIDs
	// init courses
	courseID := s.newID()
	stepState.FilterCourseIDs = []string{courseID}
	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	err := multierr.Combine(
		course.ID.Set(courseID),
		course.Name.Set("name-"+courseID),
		course.CreatedAt.Set(time.Now()),
		course.UpdatedAt.Set(time.Now()),
		course.DeletedAt.Set(nil),
		course.Grade.Set(3),
		course.StartDate.Set(time.Now()),
		course.StartDate.Set(time.Now().Add(2*time.Hour)),
		course.SchoolID.Set(schoolID),
		course.Status.Set("COURSE_STATUS_NONE"),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	courseRepo := repositories.CourseRepo{}
	err = courseRepo.Create(ctx, s.DB, course)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init course err: %s", err)
	}
	// create students
	stepState.FilterStudentIDs = []string{s.newID(), s.newID(), s.newID()}
	for _, id := range stepState.FilterStudentIDs {
		if ctx, err := s.createStudentWithName(ctx, id, fmt.Sprintf("student name %s", id)); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentSubscriptionId, subscriptionId := s.newID(), s.newID()
		sql := `INSERT INTO lesson_student_subscriptions
		(student_subscription_id, course_id, student_id, subscription_id, start_at, end_at, created_at, updated_at, resource_path)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
		_, err := s.DB.Exec(ctx, sql, database.Text(studentSubscriptionId), database.Text(courseID), id, subscriptionId, time.Now(), time.Now(), time.Now(), time.Now(), database.Text(fmt.Sprint(schoolID)))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_student_subscriptions err: %s", err)
		}
	}
	// generate 22 lessons
	for i := 1; i <= 22; i++ {
		duration := time.Duration(i)
		startTime := timeutil.Now().Add(duration * time.Hour)
		if i > 11 {
			startTime = timeutil.Now().Add(-duration * time.Hour)
		}
		centerId := fmt.Sprintf("center-%d", rand.Intn(3)+1)
		s.genLessonManagement(ctx, teacherIDs, schoolID, courseID, startTime, centerId)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) genLessonManagement(ctx context.Context, teacherIDs []string, schoolID int, courseID string, startTime time.Time, centerId string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lesson := &entities_bob.Lesson{}
	database.AllNullEntity(lesson)
	lessonID := s.newID()
	lessonName := fmt.Sprintf("Lesson name management %s", lessonID)
	classID := idutil.ULIDNow()

	err := multierr.Combine(
		lesson.LessonID.Set(lessonID),
		lesson.Name.Set(lessonName),
		lesson.CourseID.Set(courseID),
		lesson.TeacherID.Set(teacherIDs[0]),
		lesson.CreatedAt.Set(time.Now()),
		lesson.UpdatedAt.Set(timeutil.Now()),
		lesson.LessonType.Set(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
		lesson.Status.Set(cpb.LessonStatus_LESSON_STATUS_NOT_STARTED.String()),
		lesson.StreamLearnerCounter.Set(database.Int4(0)),
		lesson.LearnerIds.Set(database.JSONB([]byte("{}"))),
		lesson.StartTime.Set(startTime),
		lesson.EndTime.Set(startTime),
		lesson.TeachingMethod.Set(database.Text(cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL.String())),
		lesson.TeachingMedium.Set(database.Text(cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE.String())),
		lesson.SchedulingStatus.Set("LESSON_SCHEDULING_STATUS_PUBLISHED"),
		lesson.ClassID.Set(classID),
		lesson.CenterID.Set(centerId),
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
	updateResourcePath := "UPDATE lessons SET resource_path = $1 WHERE lesson_id = $2"
	_, err = s.DB.Exec(ctx, updateResourcePath, database.Text(fmt.Sprint(schoolID)), database.Text(lessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, teacher := range teacherIDs {
		sql := `INSERT INTO lessons_teachers
		(lesson_id, teacher_id, created_at, resource_path)
		VALUES ($1, $2, $3, $4)`
		_, err = s.DB.Exec(ctx, sql, database.Text(lessonID), database.Text(teacher), time.Now(), database.Text(fmt.Sprint(schoolID)))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lessons_teachers: %s", err)
		}
	}
	// insert lessons_member

	for i := 0; i < 3; i++ {
		sql := `INSERT INTO lesson_members
		(lesson_id, user_id, created_at, updated_at, resource_path)
		VALUES ($1, $2, $3, $4, $5)`
		learnerId := stepState.FilterStudentIDs[i]
		_, err = s.DB.Exec(ctx, sql, database.Text(lessonID), database.Text(learnerId), time.Now(), time.Now(), database.Text(fmt.Sprint(schoolID)))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init lesson_members: %s", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminRetrieveLiveLessonManagementWithFilter(
	ctx context.Context,
	lessonTime,
	keyWord,
	dateRange,
	timeRange,
	teachers,
	students,
	coursers,
	centers,
	locations,
	grade,
	dow string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	filter := &bpb.RetrieveLessonsFilterV2{}
	if len(dateRange) > 0 {
		if lessonTime == bpb.LessonTime_LESSON_TIME_PAST.String() {
			stepState.FilterFromDate = time.Now().Add(-24 * time.Hour)
			stepState.FilterToDate = time.Now()
		} else {
			stepState.FilterFromDate = time.Now()
			stepState.FilterToDate = time.Now().Add(24 * time.Hour)
		}
		filter.FromDate = timestamppb.New(stepState.FilterFromDate)
		filter.ToDate = timestamppb.New(stepState.FilterToDate)
	}
	if len(timeRange) > 0 {
		if lessonTime == bpb.LessonTime_LESSON_TIME_PAST.String() {
			stepState.FilterFromTime = time.Now().Add(-3 * time.Hour)
			stepState.FilterToTime = time.Now()
			filter.FromTime = durationpb.New(time.Duration(time.Now().Hour()-3) * time.Hour)
			filter.ToTime = durationpb.New(time.Duration(time.Now().Hour()) * time.Hour)
		} else {
			stepState.FilterFromTime = time.Now()
			stepState.FilterToTime = time.Now().Add(3 * time.Hour)
			filter.FromTime = durationpb.New(time.Duration(time.Now().Hour()) * time.Hour)
			filter.ToTime = durationpb.New(time.Duration(time.Now().Hour()+3) * time.Hour)
		}
	}
	if len(dow) > 0 {
		dows := strings.Split(dow, ",")
		ints := make([]cpb.DateOfWeek, len(dows))
		for i, s := range dows {
			val, _ := strconv.Atoi(s)
			ints[i] = cpb.DateOfWeek(cpb.DateOfWeek_value[cpb.DateOfWeek_name[int32(val)]])
		}
		filter.DateOfWeeks = ints
		filter.TimeZone = "UTC"
	}
	if teachers != "" && len(strings.Split(teachers, ",")) > 0 {
		filter.TeacherIds = stepState.FilterTeacherIDs[0:len(strings.Split(teachers, ","))]
	}
	if students != "" && len(strings.Split(students, ",")) > 0 {
		filter.StudentIds = stepState.FilterStudentIDs[0:len(strings.Split(students, ","))]
	}
	if coursers != "" && len(strings.Split(coursers, ",")) > 0 {
		filter.CourseIds = stepState.FilterCourseIDs[0:len(strings.Split(coursers, ","))]
	}
	if centers != "" {
		filter.CenterIds = []string{centers}
	}
	var locationIds []string
	if locations != "" {
		locationIds = strings.Split(locations, ",")
	}
	if len(grade) > 0 {
		grades := strings.Split(grade, ",")
		ints := make([]int32, len(grades))
		for i, s := range grades {
			val, _ := strconv.Atoi(s)
			ints[i] = int32(val)
		}
		filter.Grades = ints
	}

	req := &bpb.RetrieveLessonsRequestV2{
		Paging: &cpb.Paging{
			Limit: uint32(20),
		},
		LessonTime:  bpb.LessonTime(bpb.LessonTime_value[lessonTime]),
		CurrentTime: timestamppb.Now(),
		Filter:      filter,
		Keyword:     keyWord,
		LocationIds: locationIds,
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.Conn).RetrieveLessons(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bobMustReturnCorrectListLessonManagementWith(
	ctx context.Context,
	lessonTime,
	keyWord,
	dateRange,
	timeRange,
	teachers,
	students,
	coursers,
	centers,
	locations,
	grade, dow string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.resultCorrectDateManagement(ctx, dateRange)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with date filter: %s", err)
	}
	ctx, err = s.resultCorrectTimeManagement(ctx, timeRange)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with time filter: %s", err)
	}
	ctx, err = s.resultCorrectDateOfWeek(ctx, dow)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with date of week filter: %s", err)
	}
	ctx, err = s.resultCorrectTeacherMedium(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with teaching medium: %s", err)
	}
	ctx, err = s.resultCorrectTeachingMethod(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with teaching method: %s", err)
	}
	ctx, err = s.resultCorrectCenter(ctx, centers, locations)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with center: %s", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) resultCorrectCenter(ctx context.Context, centers, locations string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var filteredLocations []string
	if len(centers) == 0 && len(locations) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	if len(centers) == 0 {
		filteredLocations = strings.Split(locations, ",")
	} else if len(locations) == 0 {
		filteredLocations = strings.Split(centers, ",")
	} else {
		filteredLocations = s.intersectLocations(centers, locations)
	}
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	if len(filteredLocations) == 0 && len(rsp.GetItems()) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("should be empty item")
	} else {
		for _, lesson := range rsp.GetItems() {
			if !contain(filteredLocations, lesson.CenterId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s center %s not match center filter %s", lesson.Id, lesson.CenterId, centers)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) resultCorrectDateManagement(ctx context.Context, dateRange string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(dateRange) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	for _, lesson := range rsp.GetItems() {
		if stepState.FilterFromDate.After(lesson.EndTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match from_date filter", lesson.Id)
		}
		if stepState.FilterToDate.Before(lesson.StartTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match to_date filter", lesson.Id)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) resultCorrectTimeManagement(ctx context.Context, timeRange string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(timeRange) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	for _, lesson := range rsp.GetItems() {
		if stepState.FilterFromTime.After(lesson.EndTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match from_time filter", lesson.Id)
		}
		if stepState.FilterToTime.Before(lesson.StartTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match to_time filter", lesson.Id)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func checkExists(s []int32, e int32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (s *suite) intersectLocations(_centers, _locations string) []string {
	centers := strings.Split(_centers, ",")
	locations := strings.Split(_locations, ",")
	res := make([]string, 0)
	for i := range centers {
		if contain(locations, centers[i]) {
			res = append(res, centers[i])
		}
	}
	return res
}

func contain(arr []string, str string) bool {
	lowerStr := strings.ToLower(str)
	for _, v := range arr {
		if strings.ToLower(v) == lowerStr {
			return true
		}
	}
	return false
}

func (s *suite) resultCorrectDateOfWeek(ctx context.Context, dow string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(dow) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	for _, lesson := range rsp.GetItems() {
		dows := strings.Split(dow, ",")
		dowInts := make([]int32, len(dows))
		for i, s := range dows {
			val, _ := strconv.Atoi(s)
			dowInts[i] = int32(val)
		}
		if !checkExists(dowInts, int32(lesson.StartTime.AsTime().Weekday())) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match date of week filter, expect %d, actual %d", lesson.Id, dowInts, int(lesson.StartTime.AsTime().Weekday()))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) resultCorrectTeachingMethod(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	for _, lesson := range rsp.GetItems() {
		// if field is not null, check for inputs, else ignore
		if len(lesson.TeachingMethod.String()) > 0 {
			if lesson.TeachingMethod.String() != cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP.String() &&
				lesson.TeachingMethod.String() != cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s teaching_method %s not match teaching_method '%s' and teaching_method '%s'",
					lesson.Id, lesson.TeachingMethod,
					cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP.String(),
					cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL.String())
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) resultCorrectTeacherMedium(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	for _, lesson := range rsp.GetItems() {
		// if field is not null, check for inputs, else ignore
		if len(lesson.TeachingMedium.String()) > 0 {
			if lesson.TeachingMedium.String() != cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_HYBRID.String() &&
				lesson.TeachingMedium.String() != cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE.String() &&
				lesson.TeachingMedium.String() != cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s teaching medium %s not match teaching medium '%s', '%s' and '%s'",
					lesson.Id, lesson.TeachingMedium,
					cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_HYBRID.String(),
					cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE.String(),
					cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE.String())
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
