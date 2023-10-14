package services

import (
	"context"
	"fmt"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	yasuo_constant "github.com/manabie-com/backend/internal/yasuo/constant"
	"github.com/manabie-com/backend/internal/yasuo/utils"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var CMSSchoolPlusPermissionControl = []string{constant.RoleSchoolAdmin, constant.RoleTeacher}

type CourseAbac struct {
	*CourseService
}

func (rcv *CourseAbac) UpsertCourses(ctx context.Context, req *pb.UpsertCoursesRequest) (*pb.UpsertCoursesResponse, error) {
	user, err := rcv.getAdminInfo(ctx)
	if err != nil {
		return nil, err
	}

	// is array has all unique value
	courseIDs := make([]string, 0, len(req.Courses))
	schoolIDs := make([]int32, 0, len(req.Courses))
	for _, each := range req.Courses {
		courseIDs = append(courseIDs, each.Id)
		schoolIDs = append(schoolIDs, each.SchoolId)
	}
	if !isArrayUpsertHasUniqueValue(courseIDs) {
		return nil, status.Error(codes.InvalidArgument, "duplicate IDs sent")
	}

	if err = rcv.checkCanUserDoCRUDOnSchoolEntities(ctx, user, schoolIDs); err != nil {
		return nil, err
	}

	return rcv.CourseService.UpsertCourses(ctx, req)
}

func (rcv *CourseAbac) DeleteCourses(ctx context.Context, req *pb.DeleteCoursesRequest) (*pb.DeleteCoursesResponse, error) {
	err := rcv.validateSchoolPermission(ctx, req.CourseIds, rcv.CourseRepo.FindSchoolIDsOnCourses)
	if err != nil {
		return nil, err
	}

	userGroup := interceptors.UserGroupFromContext(ctx)
	if !isSchoolPortalPermission(userGroup) {
		courses, err := rcv.CourseRepo.FindByIDs(ctx, rcv.DBTrace, database.TextArray(req.CourseIds))
		if err != nil {
			return nil, errors.Wrap(err, "s.CourseRepo.FindByIDs")
		}
		if len(req.CourseIds) != len(courses) {
			return nil, status.Error(codes.InvalidArgument, "course not found")
		}

		for _, course := range courses {
			if course.CourseType.String == pb_bob.COURSE_TYPE_LIVE.String() {
				return nil, status.Error(codes.PermissionDenied, "teacher do not have permission to delete live course")
			}
		}
	}

	return rcv.CourseService.DeleteCourses(ctx, req)
}

func (rcv *CourseAbac) getAdminInfo(ctx context.Context) (*bob_entities.User, error) {
	adminID := interceptors.UserIDFromContext(ctx)
	return rcv.UserRepo.Get(ctx, rcv.DBTrace, database.Text(adminID))
}

func isArrayHasUniqueValue(arr []string) bool {
	existedMap := make(map[string]bool)
	for _, each := range arr {
		if _, ok := existedMap[each]; ok {
			return false
		}
		existedMap[each] = true
	}
	return true
}

func isArrayUpsertHasUniqueValue(arr []string) bool {
	existedMap := make(map[string]bool)
	for _, each := range arr {
		if _, ok := existedMap[each]; ok && each != "" {
			return false
		}
		existedMap[each] = true
	}
	return true
}

func hasAbsolutePermission(userGroup string) bool {
	return userGroup == constant.UserGroupAdmin
}

func hasPermissionIfSameSchool(userGroup string) bool {
	return userGroup == constant.UserGroupSchoolAdmin ||
		userGroup == constant.UserGroupTeacher
}

func (rcv *CourseAbac) checkCanUserDoCRUDOnSchoolEntities(ctx context.Context, user *bob_entities.User, schoolIds []int32) error {
	if hasAbsolutePermission(user.Group.String) {
		return nil
	}

	if hasPermissionIfSameSchool(user.Group.String) {
		if user.Group.String == constant.UserGroupTeacher {
			_, err := rcv.TeacherRepo.GetTeacherHasSchoolIDs(ctx, rcv.DBTrace, user.ID.String, schoolIds)
			if err != nil {
				return status.Error(codes.PermissionDenied, "teacher not found")
			}
		}

		if user.Group.String == constant.UserGroupSchoolAdmin {
			// if user is school admin. Force all request course to be in same school id
			schoolAdmin, err := rcv.SchoolAdminRepo.Get(ctx, rcv.DBTrace, user.ID)
			if err != nil {
				return status.Error(codes.PermissionDenied, "school admin not found")
			}

			// school in request is different than admin school
			if utils.IsArrayMatch(len(schoolIds), func(i int) bool {
				return schoolIds[i] != schoolAdmin.SchoolID.Int
			}) {
				return status.Error(codes.PermissionDenied, "can not action on entity of other school")
			}
		}

		return nil
	}

	return status.Error(codes.PermissionDenied, "do not have permission")
}

func (rcv *CourseAbac) validateSchoolPermission(ctx context.Context, reqIDs []string, getSchoolIDFn func(context.Context, database.QueryExecer, []string) ([]int32, error)) error {
	if len(reqIDs) == 0 {
		return status.Error(codes.InvalidArgument, "Empty request")
	}

	user, err := rcv.getAdminInfo(ctx)
	if err != nil {
		return err
	}

	// is array has all unique value
	if !isArrayHasUniqueValue(reqIDs) {
		return status.Error(codes.InvalidArgument, "duplicate IDs sent")
	}

	// Is id in request valid
	schoolIDs, err := getSchoolIDFn(ctx, rcv.DBTrace, reqIDs)
	if err != nil {
		return errors.Wrap(err, "rcv.FindSchoolIDs")
	}

	if len(schoolIDs) != len(reqIDs) { // maybe some courses are deleted
		return status.Error(codes.NotFound, "Elements not found")
	}

	if err = rcv.checkCanUserDoCRUDOnSchoolEntities(ctx, user, schoolIDs); err != nil {
		return err
	}

	return nil
}

func (s *CourseAbac) CreateLiveLesson(ctx context.Context, req *pb.CreateLiveLessonRequest) (*pb.CreateLiveLessonResponse, error) {
	course, err := s.CourseRepo.FindByID(ctx, s.DBTrace, database.Text(req.CourseId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("c.CourseRepo.FindByID: %v", err))
	}

	err = s.validCreateLessonRequest(ctx, req, course.SchoolID.Int)
	if err != nil {
		return nil, err
	}

	return s.CourseService.CreateLiveLesson(ctx, req)
}

func (s *CourseAbac) UpdateLiveLesson(ctx context.Context, req *pb.UpdateLiveLessonRequest) (*pb.UpdateLiveLessonResponse, error) {
	if req.LessonId == "" {
		return nil, status.Error(codes.InvalidArgument, "lesson id cannot be empty")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if req.TeacherId == "" {
		return nil, status.Error(codes.InvalidArgument, "teacher ids cannot be empty")
	}
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled(yasuo_constant.SwitchDBUnleashKey, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Connect unleash server failed: %s", err))
	}
	if isUnleashToggled {
		return s.UpdateLiveLessonLessonmgmt(ctx, req)
	}

	lesson, err := s.LessonRepo.FindByID(ctx, s.DBTrace, database.Text(req.LessonId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "cannot find lesson")
		}
		return nil, fmt.Errorf("LessonRepo.FindByID: %w", err)
	}

	if req.CourseId == "" {
		req.CourseId = lesson.CourseID.String
	}

	course, err := s.CourseRepo.FindByID(ctx, s.DBTrace, database.Text(req.CourseId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "cannot find course")
		}
		return nil, fmt.Errorf("LessonRepo.FindByID: %w", err)
	}

	if req.ControlSettings != nil {
		err = s.handleControlSettingLiveLesson(ctx, req.ControlSettings, course.SchoolID)
		if err != nil {
			return nil, err
		}
	}

	presetStudyPlanWeekly, err := s.PresetStudyPlanWeeklyRepo.FindByLessonID(ctx, s.DBTrace, database.Text(req.LessonId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "cannot find preset study plan weekly")
		}
		return nil, fmt.Errorf("PresetStudyPlanWeeklyRepo.FindByLessonID: %w", err)
	}

	if course.ID.String != lesson.CourseID.String {
		course.PresetStudyPlanID = presetStudyPlanWeekly.PresetStudyPlanID
	}

	topic, err := s.TopicRepo.FindByID(ctx, s.DBTrace, presetStudyPlanWeekly.TopicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "cannot find topic")
		}
		return nil, fmt.Errorf("TopicRepo.FindByID: %w", err)
	}

	err = s.handlePermissionToActionCourse(ctx, topic.SchoolID.Int)
	if err != nil {
		return nil, err
	}

	isInSchool, err := s.TeacherRepo.ManyTeacherIsInSchool(ctx, s.DBTrace, database.TextArray([]string{req.TeacherId}), topic.SchoolID)
	if err != nil {
		return nil, fmt.Errorf("TeacherRepo.ManyTeacherIsInSchool: %w", err)
	}

	if !isInSchool {
		return nil, status.Error(codes.InvalidArgument, "cannot add teacher of another school")
	}

	return s.CourseService.UpdateLiveLesson(ctx, req)
}

func (rcv *CourseAbac) UpdateLiveLessonLessonmgmt(ctx context.Context, req *pb.UpdateLiveLessonRequest) (*pb.UpdateLiveLessonResponse, error) {
	lesson, err := rcv.LessonRepo.FindByID(ctx, rcv.LessonDBTrace, database.Text(req.LessonId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "cannot find lesson")
		}
		return nil, fmt.Errorf("LessonRepo.FindByID: %w", err)
	}

	if req.CourseId == "" {
		req.CourseId = lesson.CourseID.String
	}

	course, err := rcv.CourseRepo.FindByID(ctx, rcv.DBTrace, database.Text(req.CourseId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "cannot find course")
		}
		return nil, fmt.Errorf("LessonRepo.FindByID: %w", err)
	}

	if req.ControlSettings != nil {
		err = rcv.handleControlSettingLiveLesson(ctx, req.ControlSettings, course.SchoolID)
		if err != nil {
			return nil, err
		}
	}

	err = rcv.handlePermissionToActionCourse(ctx, course.SchoolID.Int)
	if err != nil {
		return nil, err
	}

	isInSchool, err := rcv.TeacherRepo.ManyTeacherIsInSchool(ctx, rcv.DBTrace, database.TextArray([]string{req.TeacherId}), course.SchoolID)
	if err != nil {
		return nil, fmt.Errorf("TeacherRepo.ManyTeacherIsInSchool: %w", err)
	}

	if !isInSchool {
		return nil, status.Error(codes.InvalidArgument, "cannot add teacher of another school")
	}

	return rcv.CourseService.UpdateLiveLesson(ctx, req)
}

func (s *CourseAbac) validCreateLessonRequest(ctx context.Context, req *pb.CreateLiveLessonRequest, schoolID int32) error {
	if req.CourseId == "" {
		return status.Error(codes.InvalidArgument, "course id cannot be empty")
	}
	if len(req.Lessons) == 0 {
		return status.Error(codes.InvalidArgument, "lessons cannot be empty")
	}

	teacherIDs := []string{}
	for _, v := range req.Lessons {
		if len(v.TeacherId) == 0 {
			return status.Error(codes.InvalidArgument, "teacher ids cannot be empty")
		}
		if v.Name == "" {
			return status.Error(codes.InvalidArgument, "name cannot be empty")
		}

		if v.ControlSettings != nil { // add to avoid break client
			err := s.handleControlSettingLiveLesson(ctx, v.ControlSettings, database.Int4(schoolID))
			if err != nil {
				return err
			}
		}

		teacherIDs = append(teacherIDs, v.TeacherId)
	}

	isInSchool, err := s.TeacherRepo.ManyTeacherIsInSchool(ctx, s.DBTrace, database.TextArray(teacherIDs), database.Int4(schoolID))
	if err != nil {
		return fmt.Errorf("TeacherRepo.ManyTeacherIsInSchool: %w", err)
	}
	if !isInSchool {
		return status.Error(codes.InvalidArgument, "cannot add teacher of another school")
	}
	return nil
}

func isSchoolPortalPermission(group string) bool {
	return group == constant.RoleSchoolAdmin
}

func (s *CourseAbac) DeleteLiveCourse(ctx context.Context, req *pb.DeleteLiveCourseRequest) (*pb.DeleteLiveCourseResponse, error) {
	if len(req.CourseIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "course ids cannot empty")
	}

	courses, err := s.CourseRepo.FindByIDs(ctx, s.DBTrace, database.TextArray(req.CourseIds))
	if err != nil {
		return nil, errors.Wrap(err, "c.CourseRepo.FindByIDs")
	}

	if len(courses) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot find course")
	}

	schoolIDs := map[int32][]string{}
	presetStudyPlanIDs := []string{}
	for _, v := range courses {
		if schoolIDs[v.SchoolID.Int] != nil {
			schoolIDs[v.SchoolID.Int] = []string{v.ID.String}
		} else {
			schoolIDs[v.SchoolID.Int] = append(schoolIDs[v.SchoolID.Int], v.ID.String)
		}

		presetStudyPlanIDs = append(presetStudyPlanIDs, v.PresetStudyPlanID.String)
	}

	for schoolID := range schoolIDs {
		err = s.handlePermissionToActionCourse(ctx, schoolID)
		if err != nil {
			return nil, err
		}
	}

	return s.CourseService.DeleteLiveCourse(ctx, req)
}

func (s *CourseAbac) DeleteLiveLesson(ctx context.Context, req *pb.DeleteLiveLessonRequest) (*pb.DeleteLiveLessonResponse, error) {
	if len(req.LessonIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing lesson ids")
	}

	lessons, err := s.LessonRepo.FindByIDs(ctx, s.DBTrace, database.TextArray(req.LessonIds), false)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.FindByIDs: %w", err)
	}
	if len(lessons) == 0 {
		return nil, status.Error(codes.NotFound, "cannot find lessons")
	}
	courseIDs := []string{}
	for _, v := range lessons {
		courseIDs = append(courseIDs, v.CourseID.String)
	}
	courses, err := s.CourseRepo.FindByIDs(ctx, s.DBTrace, database.TextArray(courseIDs))
	if err != nil {
		return nil, fmt.Errorf("CourseRepo.FindByIDs: %w", err)
	}

	enCourse := map[int32][]string{}
	schoolIDs := []int32{}
	for courseID, v := range courses {
		_, ok := enCourse[v.SchoolID.Int]
		if !ok {
			enCourse[v.SchoolID.Int] = []string{courseID.String}
			schoolIDs = append(schoolIDs, v.SchoolID.Int)
		} else {
			enCourse[v.SchoolID.Int] = append(enCourse[v.SchoolID.Int], courseID.String)
		}
	}

	for _, v := range schoolIDs {
		err = s.handlePermissionToActionCourse(ctx, v)
		if err != nil {
			return nil, err
		}
	}

	return s.CourseService.DeleteLiveLesson(ctx, req)
}

func (rcv *CourseAbac) DeleteLiveLessonLessonmgmt(ctx context.Context, req *pb.DeleteLiveLessonRequest) (*pb.DeleteLiveLessonResponse, error) {
	if len(req.LessonIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing lesson ids")
	}

	lessons, err := rcv.LessonRepo.FindByIDs(ctx, rcv.LessonDBTrace, database.TextArray(req.LessonIds), false)
	if err != nil {
		return nil, fmt.Errorf("LessonRepo.FindByIDs: %w", err)
	}
	if len(lessons) == 0 {
		return nil, status.Error(codes.NotFound, "cannot find lessons")
	}
	courseIDs := []string{}
	for _, v := range lessons {
		courseIDs = append(courseIDs, v.CourseID.String)
	}
	courses, err := rcv.CourseRepo.FindByIDs(ctx, rcv.DBTrace, database.TextArray(courseIDs))
	if err != nil {
		return nil, fmt.Errorf("CourseRepo.FindByIDs: %w", err)
	}

	enCourse := map[int32][]string{}
	schoolIDs := []int32{}
	for courseID, v := range courses {
		_, ok := enCourse[v.SchoolID.Int]
		if !ok {
			enCourse[v.SchoolID.Int] = []string{courseID.String}
			schoolIDs = append(schoolIDs, v.SchoolID.Int)
		} else {
			enCourse[v.SchoolID.Int] = append(enCourse[v.SchoolID.Int], courseID.String)
		}
	}

	for _, v := range schoolIDs {
		err = rcv.handlePermissionToActionCourse(ctx, v)
		if err != nil {
			return nil, err
		}
	}

	return rcv.CourseService.DeleteLiveLesson(ctx, req)
}
