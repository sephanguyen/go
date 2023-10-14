package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) RetrieveListLessonManagement(ctx context.Context, lessonTime, limitStr, offset string) (context.Context, error) {
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
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).
		RetrieveLessons(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminRetrieveLiveLessonManagement(ctx context.Context, lessonTime, limitStr, offset string) (context.Context, error) {
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
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).RetrieveLessons(s.CommonSuite.SignedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobMustReturnListLessonManagement(ctx context.Context, total, fromID, toID, limit, next, pre string) (context.Context, error) {
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

func (s *Suite) bobMustReturnCorrectListLessonManagementWith(
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
	grade, dow,
	schedulingStatus, classes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.bobResultCorrectDateManagement(ctx, dateRange)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with date filter: %s", err)
	}
	ctx, err = s.bobResultCorrectTimeManagement(ctx, timeRange)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with time filter: %s", err)
	}
	ctx, err = s.bobResultCorrectDateOfWeek(ctx, dow)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with date of week filter: %s", err)
	}
	ctx, err = s.bobResultCorrectTeacherMedium(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with teaching medium: %s", err)
	}
	ctx, err = s.bobResultCorrectTeachingMethod(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with teaching method: %s", err)
	}
	ctx, err = s.bobResultCorrectCenter(ctx, centers, locations)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with center: %s", err)
	}
	ctx, err = s.bobResultCorrectSchedulingStatus(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with scheduling status: %s", err)
	}
	ctx, err = s.bobResultCorrectClass(ctx, classes)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result lesson not match with class: %s", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobResultCorrectCenter(ctx context.Context, centers, locations string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var filteredLocations []string
	if len(centers) == 0 && len(locations) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	switch {
	case len(centers) == 0:
		filteredLocations = strings.Split(locations, ",")
	case len(locations) == 0:
		filteredLocations = strings.Split(centers, ",")
	default:
		filteredLocations = s.intersectLocations(centers, locations)
	}

	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	if len(filteredLocations) == 0 && len(rsp.GetItems()) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("should be empty item")
	} else {
		locationIds := make([]string, 0, len(filteredLocations))
		for _, v := range filteredLocations {
			val, _ := strconv.Atoi(strings.Split(v, "-")[1])
			locationIds = append(locationIds, stepState.FilterCenterIDs[val-1])
		}
		for _, lesson := range rsp.GetItems() {
			if !slices.Contains(locationIds, lesson.CenterId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s center %s not match center filter %s", lesson.Id, lesson.CenterId, centers)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobResultCorrectDateManagement(ctx context.Context, dateRange string) (context.Context, error) {
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

func (s *Suite) bobResultCorrectTimeManagement(ctx context.Context, timeRange string) (context.Context, error) {
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

func (s *Suite) bobResultCorrectDateOfWeek(ctx context.Context, dow string) (context.Context, error) {
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

func checkExists(s []int32, e int32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (s *Suite) bobResultCorrectTeachingMethod(ctx context.Context) (context.Context, error) {
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
func (s *Suite) bobResultCorrectTeacherMedium(ctx context.Context) (context.Context, error) {
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

func (s *Suite) bobResultCorrectSchedulingStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	if len(stepState.FilterSchedulingStatuses) > 0 {
		for _, lesson := range rsp.GetItems() {
			if !isContainStatusBob(stepState.FilterSchedulingStatuses, lesson) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s scheduling status %s not match scheduling status filter %s", lesson.Id, lesson.SchedulingStatus, stepState.FilterSchedulingStatuses)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobResultCorrectClass(ctx context.Context, classes string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.RetrieveLessonsResponseV2)
	if classes != "" && len(strings.Split(classes, ",")) > 0 {
		classIds := stepState.FilterClassIDs[0:len(strings.Split(classes, ","))]
		for _, lesson := range rsp.GetItems() {
			if !slices.Contains(classIds, lesson.ClassId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s classId %s not match classIds filter %s", lesson.Id, lesson.ClassId, classIds)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func isContainStatusBob(a []domain.LessonSchedulingStatus, l *bpb.RetrieveLessonsResponseV2_Lesson) bool {
	for _, v := range a {
		if l.SchedulingStatus.String() == string(v) {
			return true
		}
	}
	return false
}

func (s *Suite) adminRetrieveLiveLessonManagementWithFilterOnBob(
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
	dow, scheduling_status,
	classes, gradeV2 string) (context.Context, error) {
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
		val, _ := strconv.Atoi(strings.Split(centers, "-")[1])
		filter.CenterIds = []string{stepState.FilterCenterIDs[val-1]}
	}
	locationIds := make([]string, 0, len(locations))
	if locations != "" {
		locationNames := strings.Split(locations, ",")

		for _, v := range locationNames {
			val, _ := strconv.Atoi(strings.Split(v, "-")[1])
			locationIds = append(locationIds, stepState.FilterCenterIDs[val-1])
		}
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

	if len(gradeV2) > 0 {
		gradesV2 := strings.Split(gradeV2, ",")
		filter.GradesV2 = gradesV2
	}

	if len(scheduling_status) > 0 {
		statusesString := strings.Split(scheduling_status, ",")
		statuses := make([]domain.LessonSchedulingStatus, 0, len(statusesString))
		for _, s := range statusesString {
			statuses = append(statuses, domain.LessonSchedulingStatus(s))
			filter.SchedulingStatus = append(filter.SchedulingStatus, cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[s]))
		}
		stepState.FilterSchedulingStatuses = statuses
	}

	if classes != "" && len(strings.Split(classes, ",")) > 0 {
		filter.ClassIds = stepState.FilterClassIDs[0:len(strings.Split(classes, ","))]
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
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).RetrieveLessons(s.CommonSuite.SignedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
