package lessonmgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) UserCreatesASetOfLessonsFor(ctx context.Context, lessonDates string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"
	localTimezone := LoadLocalLocation()

	if len(lessonDates) > 0 {
		lessonDatesList := strings.Split(lessonDates, ",")

		for _, lessonDate := range lessonDatesList {
			convertedDate, err := time.Parse(timeLayout, lessonDate)
			if err != nil {
				return ctx, fmt.Errorf("parse datetime error: %w", err)
			}
			convertedDate = convertedDate.In(localTimezone)

			req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE)
			req.StartTime = timestamppb.New(convertedDate)
			req.EndTime = timestamppb.New(convertedDate.Add(2 * time.Hour))

			if len(stepState.ClassroomIDs) > 0 {
				req.ClassroomIds = append(req.ClassroomIds, stepState.ClassroomIDs...)
			}

			if ctx, err := s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserRetrievesLessonsOnFromTo(ctx context.Context, calendarView, fromDate, toDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	localTimezone := LoadLocalLocation()

	convertedFromDate, err := time.Parse(timeLayout, fromDate)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error for from date: %w", err)
	}

	convertedToDate, err := time.Parse(timeLayout, toDate)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error for to date: %w", err)
	}

	convertedFromDate = convertedFromDate.In(localTimezone)
	convertedToDate = convertedToDate.In(localTimezone)

	req := &lpb.RetrieveLessonsOnCalendarRequest{
		CalendarView: lpb.CalendarView(lpb.CalendarView_value[calendarView]),
		LocationId:   stepState.CenterIDs[len(stepState.CenterIDs)-1],
		FromDate:     timestamppb.New(convertedFromDate),
		ToDate:       timestamppb.New(convertedToDate),
		Timezone:     localTimezone.String(),
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewLessonReaderServiceClient(s.LessonMgmtConn).RetrieveLessonsOnCalendar(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) LesonsRetrievedAreWithinTheDateRangeTo(ctx context.Context, fromDate, toDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	localTimezone := LoadLocalLocation()

	convertedFromDate, err := time.Parse(timeLayout, fromDate)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error for from date: %w", err)
	}

	convertedToDate, err := time.Parse(timeLayout, toDate)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error for to date: %w", err)
	}

	convertedFromDate = convertedFromDate.In(localTimezone)
	convertedToDate = convertedToDate.In(localTimezone)

	res := stepState.Response.(*lpb.RetrieveLessonsOnCalendarResponse)
	failedDates := make([]time.Time, 0, len(res.Items))

	for _, lessonItem := range res.Items {
		startTime := lessonItem.StartTime.AsTime()

		if startTime.Before(convertedFromDate) && startTime.After(convertedToDate) {
			failedDates = append(failedDates, startTime)
		}
	}

	if len(failedDates) > 0 {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("retrieved dates are not within the range %v to %v: %v", convertedFromDate, convertedToDate, failedDates)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheLessonsFirstAndLastDateMatchesWithAnd(ctx context.Context, fromDate, toDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	localTimezone := LoadLocalLocation()

	convertedFromDate, err := time.Parse(timeLayout, fromDate)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error for from date: %w", err)
	}

	convertedToDate, err := time.Parse(timeLayout, toDate)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error for to date: %w", err)
	}

	fromDateString := convertedFromDate.In(localTimezone).Format(timeLayout)
	toDateString := convertedToDate.In(localTimezone).Format(timeLayout)

	res := stepState.Response.(*lpb.RetrieveLessonsOnCalendarResponse)
	firstDate := res.Items[0].StartTime.AsTime().Format(timeLayout)
	lastDate := res.Items[len(res.Items)-1].StartTime.AsTime().Format(timeLayout)

	if !strings.EqualFold(firstDate, fromDateString) {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("the first lesson date %v is not equal to the from date %v", firstDate, fromDateString)
	}

	if !strings.EqualFold(lastDate, toDateString) {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("the last lesson date %v is not equal to the to date %v", lastDate, toDateString)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserRetrievesLessonsOnWithFilter(ctx context.Context, calendarView, students, teachers, courses, classes, noneAssignedTeacher string) (context.Context, error) {
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	localTimezone := LoadLocalLocation()
	filter := &lpb.RetrieveLessonsOnCalendarRequest_Filter{
		NoneAssignedTeacherLessons: noneAssignedTeacher == "true",
	}
	now := time.Now().In(localTimezone)

	stepState.NoneAssignedTeacher = filter.NoneAssignedTeacherLessons

	if len(strings.TrimSpace(teachers)) > 0 {
		teachersCount := len(strings.Split(teachers, ","))
		stepState.FilterTeachersCount = teachersCount
		filter.TeacherIds = stepState.FilterTeacherIDs[0:teachersCount]
	}

	if len(strings.TrimSpace(students)) > 0 {
		studentCount := len(strings.Split(students, ","))
		stepState.FilterStudentsCount = studentCount
		filter.StudentIds = stepState.FilterStudentIDs[0:studentCount]
	}

	if len(strings.TrimSpace(courses)) > 0 {
		coursesCount := len(strings.Split(courses, ","))
		stepState.FilterCoursesCount = coursesCount
		filter.CourseIds = stepState.FilterCourseIDs[0:coursesCount]
	}

	if len(strings.TrimSpace(classes)) > 0 {
		classesCount := len(strings.Split(classes, ","))
		stepState.FilterClassesCount = classesCount
		filter.ClassIds = stepState.FilterClassIDs[0:classesCount]
	}

	req := &lpb.RetrieveLessonsOnCalendarRequest{
		CalendarView: lpb.CalendarView(lpb.CalendarView_value[calendarView]),
		LocationId:   stepState.FilterCenterIDs[0],
		FromDate:     timestamppb.New(now.Add(-24 * time.Hour)),
		ToDate:       timestamppb.New(now.Add(30 * 24 * time.Hour)),
		Timezone:     localTimezone.String(),
		Filter:       filter,
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewLessonReaderServiceClient(s.LessonMgmtConn).RetrieveLessonsOnCalendar(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) LessonsRetrievedOnCalendarWithFilterAreCorrect(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.lessonsResultHasCorrectStudents(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with students: %s", err)
	}

	ctx, err = s.lessonsResultHasCorrectCourses(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with courses: %s", err)
	}

	ctx, err = s.lessonsResultHasCorrectTeachers(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with teachers: %s", err)
	}

	ctx, err = s.lessonsResultHasCorrectClass(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with class: %s", err)
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with class: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonsResultHasCorrectClass(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	classCount := stepState.FilterClassesCount
	resp := stepState.Response.(*lpb.RetrieveLessonsOnCalendarResponse)

	if classCount > 0 {
		classIDs := stepState.FilterClassIDs[0:classCount]
		for _, lesson := range resp.GetItems() {
			if !slices.Contains(classIDs, lesson.ClassId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson ID %s class ID %s not match classIDs filter %s", lesson.LessonId, lesson.ClassId, classIDs)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonsResultHasCorrectCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseCount := stepState.FilterCoursesCount
	resp := stepState.Response.(*lpb.RetrieveLessonsOnCalendarResponse)

	if courseCount > 0 {
		courseIDs := stepState.FilterCourseIDs[0:courseCount]
		for _, lesson := range resp.GetItems() {
			// check in lesson course ID first
			if !slices.Contains(courseIDs, lesson.CourseId) {
				for _, student := range lesson.GetLessonMembers() {
					// check in each student in the lesson if any course ID match
					if !slices.Contains(courseIDs, student.CourseId) {
						return StepStateToContext(ctx, stepState), fmt.Errorf("lesson ID %s course ID %s not match courseIDs filter %s", lesson.LessonId, lesson.CourseId, courseIDs)
					}
				}
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonsResultHasCorrectStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentCount := stepState.FilterStudentsCount

	if studentCount > 0 {
		resp := stepState.Response.(*lpb.RetrieveLessonsOnCalendarResponse)

		studentIDs := stepState.FilterStudentIDs[0:studentCount]
		for _, lesson := range resp.GetItems() {
			if lesson.TeachingMethod == cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL {
				students := sliceutils.FilterWithReferenceList(studentIDs,
					lesson.GetLessonMembers(),
					func(studentIDs []string, item *lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonMember) bool {
						return slices.Contains(studentIDs, item.StudentId)
					},
				)

				if len(students) == 0 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("lesson ID %s does not contain any studentIDs filter %s", lesson.LessonId, studentIDs)
				}
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonsResultHasCorrectTeachers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	teacherCount := stepState.FilterTeachersCount
	noneAssignedTeacher := stepState.NoneAssignedTeacher

	if teacherCount > 0 && !noneAssignedTeacher {
		resp := stepState.Response.(*lpb.RetrieveLessonsOnCalendarResponse)
		teacherIDs := stepState.FilterTeacherIDs[0:teacherCount]
		for _, lesson := range resp.GetItems() {
			teachers := sliceutils.FilterWithReferenceList(teacherIDs,
				lesson.GetLessonTeachers(),
				func(teacherIDs []string, item *lpb.RetrieveLessonsOnCalendarResponse_Lesson_LessonTeacher) bool {
					return slices.Contains(teacherIDs, item.TeacherId)
				},
			)

			if len(teachers) == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson ID %s does not contain any teacherIDs filter %s", lesson.LessonId, teacherIDs)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aListOfLessonsAreExisting(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.aListOfLessonsManagementOfSchoolAreExistedInDB(StepStateToContext(ctx, stepState), constant.ManabieSchool)
}
