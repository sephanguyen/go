package services

import (
	"context"
	"fmt"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	CMSSchoolPlusPermissionControl = []string{constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher}
	SchoolPortalPermissionControl  = []string{constant.RoleSchoolAdmin, constant.RoleHQStaff}
)

type CourseServiceABAC struct {
	*CourseService
}

func (rcv *CourseServiceABAC) UpsertLOs(ctx context.Context, req *pb.UpsertLOsRequest) (*pb.UpsertLOsResponse, error) {
	return rcv.CourseService.UpsertLOs(ctx, req)
}

func (rcv *CourseServiceABAC) getAdminInfo(ctx context.Context) (*entities_bob.User, error) {
	adminID := interceptors.UserIDFromContext(ctx)
	return rcv.UserRepo.Get(ctx, rcv.DB, database.Text(adminID))
}

func hasAbsolutePermission(group string) bool {
	return group == constant.UserGroupAdmin
}

func hasPermissionIfSameSchool(group string) bool {
	return group == constant.UserGroupSchoolAdmin ||
		group == constant.RoleTeacher
}

func (rcv *CourseServiceABAC) checkCanUserDoCRUDOnSchoolEntities(ctx context.Context, user *entities_bob.User, schoolIDs []int32) error {
	if hasAbsolutePermission(user.Group.String) {
		return nil
	}

	if hasPermissionIfSameSchool(user.Group.String) {
		if user.Group.String == constant.RoleTeacher {
			_, err := rcv.TeacherRepo.GetTeacherHasSchoolIDs(ctx, rcv.DB, user.ID.String, schoolIDs)
			if err != nil {
				return status.Error(codes.PermissionDenied, fmt.Errorf("teacher not found %w", err).Error())
			}
		}

		if user.Group.String == constant.UserGroupSchoolAdmin {
			// if user is school admin. Force all request course to be in same school id
			schoolAdmin, err := rcv.SchoolAdminRepo.Get(ctx, rcv.DB, user.ID)
			if err != nil {
				return status.Error(codes.PermissionDenied, fmt.Errorf("school admin not found %w", err).Error())
			}

			// school in request is different than admin school
			for _, each := range schoolIDs {
				if each != schoolAdmin.SchoolID.Int {
					return status.Error(codes.PermissionDenied, "can not action on entity of other school")
				}
			}
		}

		return nil
	}

	return status.Error(codes.PermissionDenied, "do not have permission")
}

func (rcv *CourseServiceABAC) findStudentValidCourse(ctx context.Context, userID string) ([]string, error) {
	var validCourseIDs []string
	classes, err := rcv.ClassRepo.FindJoined(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return nil, toStatusError(err)
	}
	classIDs := make([]int32, 0, len(classes))
	for _, class := range classes {
		classIDs = append(classIDs, class.ID.Int)
	}
	courseMapByClass, err := rcv.CourseClassRepo.Find(ctx, rcv.DB, database.Int4Array(classIDs))
	if err != nil {
		return validCourseIDs, toStatusError(err)
	}
	for _, courseIDByClass := range courseMapByClass {
		var courseIDs []string
		courseIDByClass.AssignTo(&courseIDs)
		validCourseIDs = append(validCourseIDs, courseIDs...)
	}

	courseIDs, err := rcv.LessonMemberRepo.CourseAccessible(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return nil, fmt.Errorf("err rcv.LessonMemberRepo.CourseAccessible: %w", err)
	}
	validCourseIDs = append(validCourseIDs, courseIDs...)

	return validCourseIDs, nil
}

func (rcv *CourseServiceABAC) RetrieveLiveLesson(ctx context.Context, req *pb.RetrieveLiveLessonRequest) (*pb.RetrieveLiveLessonResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	group, err := rcv.UserRepo.UserGroup(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return nil, toStatusError(err)
	}
	var validCourseIDs []string

	if group != constant.UserGroupStudent {
		courses, err := rcv.CourseRepo.RetrieveCourses(ctx, rcv.DB, &repositories.CourseQuery{
			IDs: req.CourseIds,
		})
		if err != nil {
			return nil, toStatusError(err)
		}

		for _, course := range courses {
			validCourseIDs = append(validCourseIDs, course.ID.String)
		}
	} else {
		courseIDs, err := rcv.findStudentValidCourse(ctx, userID)
		if err != nil {
			return nil, toStatusError(err)
		}
		validCourseIDs = append(validCourseIDs, courseIDs...)
	}
	validCourseMap := make(map[string]bool)
	for _, id := range validCourseIDs {
		validCourseMap[id] = true
	}

	for _, id := range req.CourseIds {
		isValid, ok := validCourseMap[id]
		if !(isValid && ok) {
			return nil, status.Error(codes.PermissionDenied, "do not have permission")
		}
	}

	if len(req.CourseIds) == 0 {
		req.CourseIds = validCourseIDs
	}

	return rcv.CourseService.RetrieveLiveLesson(ctx, req)
}
