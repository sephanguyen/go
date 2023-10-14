package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	master_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/bxcodec/faker/v3/support/slice"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) adminRetrieveStudentSubscriptionsInLessonmgmt(ctx context.Context, limit, offset int, _lessonDate, keyword, coursers, grades, classIds, locationIds, gradesV2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	filter := &lpb.RetrieveStudentSubscriptionFilter{}
	keywordStr := ""
	if keyword != "" {
		keywordStr = keyword
	}
	if coursers != "" && len(strings.Split(coursers, ",")) > 0 {
		filter.CourseId = stepState.FilterCourseIDs[0:len(strings.Split(coursers, ","))]
	}
	if len(grades) > 0 {
		filter.Grade = strings.Split(grades, ",")
	}
	if classIds != "" && len(strings.Split(classIds, ",")) > 0 {
		for _, v := range strings.Split(classIds, ",") {
			index, err := strconv.Atoi(v)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			filter.ClassId = append(filter.ClassId, stepState.FilterClassIDs[index-1])
		}
	}
	if gradesV2 != "" && len(strings.Split(gradesV2, ",")) > 0 {
		for _, v := range strings.Split(gradesV2, ",") {
			index, err := strconv.Atoi(v)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			filter.GradesV2 = append(filter.GradesV2, stepState.GradeIDs[index-1])
		}
	}
	if locationIds != "" && len(strings.Split(locationIds, ",")) > 0 {
		for _, v := range strings.Split(locationIds, ",") {
			index, err := strconv.Atoi(v)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			filter.LocationId = append(filter.LocationId, stepState.FilterLocationIDs[index-1])
		}
	}

	lessonDate, _ := time.Parse(time.RFC3339, _lessonDate)
	req := &lpb.RetrieveStudentSubscriptionRequest{
		Paging: &cpb.Paging{
			Limit: uint32(limit),
		},
		Keyword:    keywordStr,
		Filter:     filter,
		LessonDate: timestamppb.New(lessonDate),
	}

	if offset > 0 {
		offset := stepState.FilterStudentSubs[offset]
		req = &lpb.RetrieveStudentSubscriptionRequest{
			Paging: &cpb.Paging{
				Limit:  uint32(limit),
				Offset: &cpb.Paging_OffsetString{OffsetString: offset},
			},
			Keyword: keywordStr,
			Filter:  filter,
		}
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewStudentSubscriptionServiceClient(s.LessonMgmtConn).RetrieveStudentSubscription(s.CommonSuite.SignedCtx(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonmgmtMustReturnListStudentSubscriptions(ctx context.Context, total int, fromID, toID string, limit int, next, pre, lessonDate, hasFilter string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*lpb.RetrieveStudentSubscriptionRequest)
	rsp := stepState.Response.(*lpb.RetrieveStudentSubscriptionResponse)
	if total == 0 {
		if rsp.TotalItems != uint32(total) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("must return empty list\n expect total: %d, actual total : %d", total, rsp.TotalItems)
		}
		return StepStateToContext(ctx, stepState), nil
	} else if rsp.TotalItems != uint32(total) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong list return\n expect total: %d, actual: %d", total, rsp.TotalItems)
	}
	if len(rsp.Items) > 0 {
		ctx, err := s.checkClassIDInResponseInLessonmgmt(ctx, rsp.Items)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ctx, err = s.checkRangeDateInResponseInLessonmgmt(ctx, rsp.Items, lessonDate)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	if hasFilter != "" {
		ctx, err := s.resultStudentsSubCorrectCourseInLessonmgmt(ctx, req.Filter.GetCourseId())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with courses: %s", err)
		}

		ctx, err = s.resultStudentsSubCorrectGradeInLessonmgmt(ctx, req.Filter.GetGrade())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with grade: %s", err)
		}

		ctx, err = s.resultStudentsSubCorrectLocationIDsInLessonmgmt(ctx, req.Filter.GetLocationId())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with locationIds: %s", err)
		}

		ctx, err = s.resultStudentsSubCorrectClassIDsInLessonmgmt(ctx, req.Filter.GetClassId())
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with classIds: %s", err)
		}
		ctx, err = s.resultStudentsSubCorrectGradeIDsInLessonmgmt(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with gradeIds: %s", err)
		}
	} else {
		if len(rsp.Items) > 0 {
			if fromID != NIL_VALUE {
				fromIDInt, err := strconv.Atoi(fromID)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}

				if rsp.Items[0].Id != stepState.FilterStudentSubs[fromIDInt] {
					return StepStateToContext(ctx, stepState), fmt.Errorf("wrong first items return \n expect: %s, actual: %s", stepState.FilterStudentSubs[fromIDInt], rsp.Items[0].Id)
				}
			}
			if toID != NIL_VALUE {
				toIDInt, err := strconv.Atoi(toID)
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				if rsp.Items[len(rsp.Items)-1].Id != stepState.FilterStudentSubs[toIDInt] {
					return StepStateToContext(ctx, stepState), fmt.Errorf("wrong last items return \n expect: %s, actual: %s", stepState.FilterStudentSubs[toIDInt], rsp.Items[len(rsp.Items)-1].Id)
				}
			}
		}

		if int(rsp.NextPage.Limit) != limit || int(rsp.PreviousPage.Limit) != limit {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong limit return")
		}
		if next != NIL_VALUE {
			nextInt, err := strconv.Atoi(next)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			if rsp.NextPage.GetOffsetString() != stepState.FilterStudentSubs[nextInt] {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong next offset return \n expect: %s, actual: %s", stepState.FilterStudentSubs[nextInt], rsp.NextPage.GetOffsetString())
			}
		}

		if pre == NIL_VALUE {
			pre = ""
		} else {
			preInt, err := strconv.Atoi(pre)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			pre = stepState.FilterStudentSubs[preInt]
		}
		if rsp.PreviousPage.GetOffsetString() != pre {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong previous offset return \n expect: %s, actual: %s", pre, rsp.PreviousPage.GetOffsetString())
		}
	}
	ctx, err := s.resultCorrectLocationInLessonmgmt(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with access_path: %s", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectLocationInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*lpb.RetrieveStudentSubscriptionResponse)

	for _, item := range rsp.GetItems() {
		for _, locationID := range item.LocationIds {
			if !checkStringExists(locationID, stepState.CenterIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("location_id %s not match", locationID)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectCourseInLessonmgmt(ctx context.Context, courses []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(courses) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*lpb.RetrieveStudentSubscriptionResponse)

	for _, item := range rsp.GetItems() {
		if !checkStringExists(item.CourseId, courses) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("course %s not match with course filter", item.CourseId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectGradeInLessonmgmt(ctx context.Context, grades []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(grades) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*lpb.RetrieveStudentSubscriptionResponse)

	for _, item := range rsp.GetItems() {
		if !checkStringExists(item.Grade, grades) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("grade %s not match with grade filter", item.Grade)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectLocationIDsInLessonmgmt(ctx context.Context, locationIds []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(locationIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	rsp := stepState.Response.(*lpb.RetrieveStudentSubscriptionResponse)

	repo := repositories.StudentSubscriptionAccessPathRepo{}
	studentSubscriptionIDs, err := repo.FindStudentSubscriptionIDsByLocationIDs(ctx, s.BobDB, locationIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot get ClassMemberRepo.FindStudentSubscriptionIDsByLocationIDs: %s", err)
	}

	for _, item := range rsp.GetItems() {
		if !checkStringExists(item.Id, studentSubscriptionIDs) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("student subscription %s in location %s not match with location filter", item.Id, locationIds)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectClassIDsInLessonmgmt(ctx context.Context, classIds []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(classIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}

	rsp := stepState.Response.(*lpb.RetrieveStudentSubscriptionResponse)

	repo := master_repo.ClassMemberRepo{}
	students, err := repo.FindStudentIDWithCourseIDsByClassIDs(ctx, s.BobDB, classIds)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot get ClassMemberRepo.FindStudentIDWithCourseIDsByClassIDs: %s", err)
	}

	for _, item := range rsp.GetItems() {
		if !checkStringExists(item.StudentId, students) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("class %s not match with class filter", item.StudentId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultStudentsSubCorrectGradeIDsInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*lpb.RetrieveStudentSubscriptionResponse)

	for _, item := range rsp.GetItems() {
		if item.GetGradeV2() != "" && !slice.Contains(stepState.GradeIDs, item.GetGradeV2()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("grade %s not match with grade filter", item.Grade)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkClassIDInResponseInLessonmgmt(ctx context.Context, list []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentCourses := make([]*domain.ClassWithCourseStudent, 0, len(list))
	for _, sub := range list {
		sc := &domain.ClassWithCourseStudent{CourseID: sub.CourseId, StudentID: sub.StudentId}
		studentCourses = append(studentCourses, sc)
	}

	repo := master_repo.ClassRepo{}
	result, err := repo.FindByCourseIDsAndStudentIDs(ctx, s.BobDB, studentCourses)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("list result not match with courses: %s", err)
	}

	for _, v := range list {
		isIncorrect := s.checkClassIDIsIncorrectWithCourseAndClassInLessonmgmt(result, v)
		if isIncorrect && v.ClassId != "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("classId: %s in item %s", v.ClassId, v.Id)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkRangeDateInResponseInLessonmgmt(ctx context.Context, list []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription, _lessonDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonDate, _ := time.Parse(time.RFC3339, _lessonDate)
	d := time.Date(lessonDate.Year(), lessonDate.Month(), lessonDate.Day(), 0, 0, 0, 0, time.UTC)
	for _, sub := range list {
		startDate := time.Date(sub.StartDate.AsTime().Year(), sub.StartDate.AsTime().Month(), sub.StartDate.AsTime().Day(), 0, 0, 0, 0, time.UTC)
		endDate := time.Date(sub.EndDate.AsTime().Year(), sub.EndDate.AsTime().Month(), sub.EndDate.AsTime().Day(), 0, 0, 0, 0, time.UTC)

		if startDate.After(d) || endDate.Before(d) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("student subscription duration(student_id:%s,course_id:%s) not aligned with lesson date", sub.StudentId, sub.CourseId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkClassIDIsIncorrectWithCourseAndClassInLessonmgmt(list []*domain.ClassWithCourseStudent, subStudent *lpb.RetrieveStudentSubscriptionResponse_StudentSubscription) bool {
	for _, v := range list {
		if v.CourseID == subStudent.CourseId && v.StudentID == subStudent.StudentId {
			if v.ClassID == subStudent.ClassId {
				return false
			}
		}
	}

	return true
}
