package services

import (
	"context"
	"fmt"
	"strings"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
)

func toPbLesson(src *repositories.LessonWithTime, teacher *pb.BasicProfile) *pb.Lesson {
	status := pb.LESSON_STATUS_NOT_STARTED
	if src.Lesson.StartTime.Time.Unix() >= timeutil.Now().Unix() {
		status = pb.LESSON_STATUS_IN_PROGRESS
	}
	if src.Lesson.EndTime.Time.Unix() < timeutil.Now().Unix() || (src.Lesson.EndAt.Status == pgtype.Present) {
		status = pb.LESSON_STATUS_COMPLETED
	}

	l := &pb.Lesson{
		LessonId:                 src.Lesson.LessonID.String,
		CourseId:                 src.Lesson.CourseID.String,
		PresetStudyPlanWeeklyIds: "",
		Topic: &pb.Topic{
			Name:        src.Lesson.Name.String,
			Attachments: []*pb.Attachment{},
		},
		StartTime: &types.Timestamp{Seconds: src.Lesson.StartTime.Time.Unix()},
		EndTime:   &types.Timestamp{Seconds: src.Lesson.EndTime.Time.Unix()},
		Status:    status,
	}
	if teacher != nil {
		l.Teacher = []*pb.BasicProfile{teacher}
	}

	return l
}

func toPbLessons(lessons []*repositories.LessonWithTime, teacherMap map[string]*pb.BasicProfile) []*pb.Lesson {
	pbLessons := make([]*pb.Lesson, 0, len(lessons))
	for _, lesson := range lessons {
		teacher := teacherMap[lesson.Lesson.TeacherID.String]

		pbLesson := toPbLesson(lesson, teacher)
		pbLessons = append(pbLessons, pbLesson)
	}

	return pbLessons
}

func (c *CourseService) RetrieveLiveLesson(ctx context.Context, req *pb.RetrieveLiveLessonRequest) (*pb.RetrieveLiveLessonResponse, error) {
	courseIDs := database.TextArray(req.CourseIds)

	from, err := database.TimestamptzFromProto(req.From)
	if err != nil {
		return nil, toStatusError(err)
	}
	to, err := database.TimestamptzFromProto(req.To)
	if err != nil {
		return nil, toStatusError(err)
	}
	var limit, page int32
	if req.Pagination != nil {
		limit = req.Pagination.Limit
		page = req.Pagination.Page
	}

	userID := interceptors.UserIDFromContext(ctx)
	group, err := c.UserRepo.UserGroup(ctx, c.DB, database.Text(userID))
	if err != nil {
		return nil, toStatusError(err)
	}

	var (
		lessons []*repositories.LessonWithTime
		total   pgtype.Int8
	)
	isUnleashToggled, err := c.UnleashClientIns.IsFeatureEnabled("BACKEND_Lesson_HandleShowOnlyPublishStatusForEndpointListLessonForTeacherStudent", c.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}
	schedulingStatus := pgtype.Text{Status: pgtype.Null}
	if isUnleashToggled {
		if err := schedulingStatus.Set(entities_bob.LessonSchedulingStatusPublished); err != nil {
			return nil, ToStatusError(err)
		}
	}
	switch group {
	case constant.UserGroupStudent:
		if c.Cfg.Common.Environment == "prod" {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			specificCourses, err := c.ConfigRepo.RetrieveWithResourcePath(ctx, c.DB, database.Text(pb.COUNTRY_MASTER.String()), database.Text("lesson"), database.TextArray([]string{"specificCourseIDsForLesson"}), database.Text(resourcePath))
			if err == nil && len(specificCourses) > 0 {
				courseIDs = database.TextArray(strings.Split(specificCourses[0].Value.String, ","))
			}
		}

		lessons, total, err = c.LessonRepo.FindLessonJoined(ctx, c.DB, database.Text(userID), &courseIDs, from, to, limit, page, schedulingStatus)
	default:
		lessons, total, err = c.LessonRepo.FindLessonWithTime(ctx, c.DB, &courseIDs, from, to, limit, page, schedulingStatus)
	}

	if err != nil {
		return nil, toStatusError(err)
	}

	teacherIDs := make([]string, 0, len(lessons))
	for _, lesson := range lessons {
		teacherIDs = append(teacherIDs, lesson.Lesson.TeacherID.String)
	}

	teacherProfilesMap, err := c.getTeacherProfile(ctx, teacherIDs)
	if err != nil {
		return nil, err
	}

	pbLessons := toPbLessons(lessons, teacherProfilesMap)

	return &pb.RetrieveLiveLessonResponse{
		Lessons: pbLessons,
		Total:   int32(total.Int),
	}, nil
}

func (c *CourseService) RetrieveCoursesByIDs(ctx context.Context, req *pb.RetrieveCoursesByIDsRequest) (*pb.RetrieveCoursesResponse, error) {
	courses, err := c.CourseRepo.RetrieveByIDs(ctx, c.DB, database.TextArray(req.Ids))
	if err != nil {
		return nil, toStatusError(err)
	}

	return c.coursesDecoration(ctx, courses)
}

func (c *CourseService) coursesDecoration(ctx context.Context, courses []*entities_bob.Course) (*pb.RetrieveCoursesResponse, error) {
	var teacherIDs, courseIDs, bookIDs []string
	for _, course := range courses {
		ids := make([]string, 0, len(course.TeacherIDs.Elements))
		err := course.TeacherIDs.AssignTo(&ids)
		if err != nil {
			return nil, err
		}
		teacherIDs = append(teacherIDs, ids...)

		courseIDs = append(courseIDs, course.ID.String)
	}

	teacherProfilesMap, err := c.getTeacherProfile(ctx, teacherIDs)
	if err != nil {
		return nil, err
	}

	courseClassMap, err := c.CourseClassRepo.FindByCourseIDs(ctx, c.DB, database.TextArray(courseIDs))
	if err != nil {
		return nil, fmt.Errorf("CourseClassRepo.FindByCourseIDs: %w", err)
	}

	courseBookMap, err := c.CourseBookRepo.FindByCourseIDs(ctx, c.DB, courseIDs)
	if err != nil {
		return nil, fmt.Errorf("CourseBookRepo.FindByCourseIDs: %w", err)
	}

	for _, v := range courseBookMap {
		bookIDs = append(bookIDs, v...)
	}

	bookChapterMap, err := c.BookChapterRepo.FindByBookIDs(ctx, c.DB, bookIDs)
	if err != nil {
		return nil, fmt.Errorf("BookRepo.FindByIDs: %w", err)
	}

	pbCourses, err := c.toCoursesPb(ctx, courses, teacherProfilesMap, courseClassMap, courseBookMap, bookChapterMap)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &pb.RetrieveCoursesResponse{
		Courses: pbCourses,
		Total:   int32(len(pbCourses)),
	}, nil
}

func (c *CourseService) getTeacherProfile(ctx context.Context, teacherIDs []string) (map[string]*pb.BasicProfile, error) {
	if len(teacherIDs) == 0 {
		return nil, nil
	}

	teacherProfilesMap := make(map[string]*pb.BasicProfile, len(teacherIDs))
	teachers, err := c.UserRepo.Retrieve(ctx, c.DB, database.TextArray(teacherIDs))
	if err != nil {
		return nil, fmt.Errorf("UserRepo.Retrieve: %w", err)
	}

	for _, teacher := range teachers {
		basicProfile := toBasicProfile(teacher)
		teacherProfilesMap[teacher.ID.String] = basicProfile
	}

	return teacherProfilesMap, nil
}
