package usermodadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type UserModuleAdapter struct {
	*user.Module
}

func (u *UserModuleAdapter) CheckTeacherIDs(ctx context.Context, ids []string) error {
	res, err := u.UserGRPCService.GetTeachers(ctx, &lpb.GetTeachersRequest{
		TeacherIds: ids,
	})
	if err != nil {
		return fmt.Errorf("got error when get teachers: %w", err)
	}

	actualIDs := make(map[string]bool)
	for _, teacher := range res.Teachers {
		actualIDs[teacher.Id] = true
	}
	for _, expected := range ids {
		if _, ok := actualIDs[expected]; !ok {
			return fmt.Errorf("teacher id %s not exist", expected)
		}
	}

	return nil
}

func (u *UserModuleAdapter) CheckStudentCourseSubscriptions(ctx context.Context, lessonDate time.Time, studentIDWithCourseID ...string) error {
	if len(studentIDWithCourseID)%2 != 0 {
		return fmt.Errorf("missing course id of student %s", studentIDWithCourseID[len(studentIDWithCourseID)-1])
	}

	req := make([]*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription, 0, len(studentIDWithCourseID))
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		req = append(req, &lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
			StudentId: studentIDWithCourseID[i],
			CourseId:  studentIDWithCourseID[i+1],
		})
	}
	// HACK: when have full data we will check all student course with location same lesson center
	res, err := u.StudentSubscriptionGRPCLessonmgmtService.GetStudentCourseSubscriptions(
		ctx,
		&lpb.GetStudentCourseSubscriptionsRequest{Subscriptions: req},
	)
	if err != nil {
		return fmt.Errorf("got error when get student course subscriptions: %w", err)
	}

	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		studentId := studentIDWithCourseID[i]
		courseId := studentIDWithCourseID[i+1]
		exist := false
		for _, item := range res.Items {
			if item.StudentId == studentId && item.CourseId == courseId {
				exist = true
				break
			}
		}
		if !exist {
			return fmt.Errorf("subscription of student %s with course %s not exist", studentId, courseId)
		}
	}
	for _, item := range res.Items {
		startAt := golibs.TimestamppbToTime(item.StartDate)
		endAt := golibs.TimestamppbToTime(item.EndDate)
		if startAt.After(lessonDate) || endAt.Before(lessonDate) {
			return fmt.Errorf("student subscription duration(student_id:%s,course_id:%s) not aligned with lesson date", item.StudentId, item.CourseId)
		}
	}
	return nil
}

func (u *UserModuleAdapter) GetUserGroup(ctx context.Context, userID string) (string, error) {
	res, err := u.UserGRPCService.GetUserGroup(ctx, &lpb.GetUserGroupRequest{UserId: userID})
	if err != nil {
		return "", fmt.Errorf("got error when get user group of user %s: %w", userID, err)
	}

	return res.UserGroup, err
}
