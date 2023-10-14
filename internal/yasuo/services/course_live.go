package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	utils "github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	yasuo_repositories "github.com/manabie-com/backend/internal/yasuo/repositories"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (s *CourseService) handlePermissionToActionCourse(ctx context.Context, schoolID int32) error {
	userGroup := interceptors.UserGroupFromContext(ctx)
	if userGroup == "" {
		return nil
	}

	if !isSchoolPortalPermission(userGroup) {
		return status.Error(codes.PermissionDenied, "you do not have permission")
	}
	userID := interceptors.UserIDFromContext(ctx)
	switch userGroup {
	case pb.USER_GROUP_SCHOOL_ADMIN.String():
		{
			schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DBTrace, database.Text(userID))
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return status.Error(codes.NotFound, "cannot find school admin")
				}
				return fmt.Errorf("SchoolAdminRepo.Get: %w", err)
			}

			if schoolAdmin.SchoolID.Int != schoolID {
				return status.Error(codes.PermissionDenied, "cannot handle data of another school")
			}
		}
	}
	return nil
}
func (s *CourseService) validLiveCourseProto(ctx context.Context, req *pb.UpsertLiveCourseRequest) error {
	if req.Name == "" {
		return status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	if req.Grade == "" {
		return status.Error(codes.InvalidArgument, "grade cannot be empty")
	}
	if req.Subject == pb_bob.SUBJECT_NONE {
		return status.Error(codes.InvalidArgument, "subject cannot be empty")
	}
	if req.Country == pb_bob.COUNTRY_NONE {
		return status.Error(codes.InvalidArgument, "country cannot be empty")
	}
	if req.SchoolId == 0 {
		return status.Error(codes.InvalidArgument, "school id cannot be empty")
	}

	if len(req.TeacherIds) > 0 {
		isInSchool, err := s.TeacherRepo.ManyTeacherIsInSchool(ctx, s.DBTrace, database.TextArray(req.TeacherIds), database.Int4(int32(req.SchoolId)))
		if err != nil {
			return fmt.Errorf("TeacherRepo.ManyTeacherIsInSchool: %w", err)
		}
		if !isInSchool {
			return status.Error(codes.InvalidArgument, "cannot add teacher of another school")
		}
	}
	if len(req.ClassIds) > 0 {
		mapClasses, err := s.ClassRepo.FindBySchoolAndID(ctx, s.DBTrace, database.Int4(int32(req.SchoolId)), database.Int4Array(req.ClassIds))
		if err != nil {
			return fmt.Errorf("ClassRepo.FindBySchoolAndID: %w", err)
		}

		if len(mapClasses) != len(req.ClassIds) {
			return status.Error(codes.InvalidArgument, "cannot add class of another school")
		}
	}
	return nil
}

func (s *CourseService) validAndConvertLiveCourseEntity(ctx context.Context, req *pb.UpsertLiveCourseRequest) (*entities.Course, error) {
	err := s.validLiveCourseProto(ctx, req)
	if err != nil {
		return nil, err
	}
	startDate, endDate, err := checkStartAndEndDate(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	err = s.handlePermissionToActionCourse(ctx, int32(req.SchoolId))
	if err != nil {
		return nil, err
	}

	r := &entities.Course{}
	database.AllNullEntity(r)

	if req.Id == "" {
		err = multierr.Combine(
			r.ID.Set(idutil.ULIDNow()),
			r.PresetStudyPlanID.Set(idutil.ULIDNow()),
			r.CourseType.Set(pb_bob.COURSE_TYPE_LIVE.String()),
		)
		if err != nil {
			return nil, fmt.Errorf("multierr.Combine: %w", err)
		}
	} else {
		r, err = s.CourseRepo.FindByID(ctx, s.DBTrace, database.Text(req.Id))
		if err != nil {
			return nil, fmt.Errorf("CourseRepo.FindByID %s: %w", req.Id, err)
		}
		// if r.CourseType.String != pb_bob.COURSE_TYPE_LIVE.String() {
		// return nil, status.Error(codes.InvalidArgument, "cannot upsert a course is not live course")
		//}
		if req.SchoolId != int64(r.SchoolID.Int) {
			return nil, status.Error(codes.InvalidArgument, "school id not match")
		}
	}

	grade, err := i18n.ConvertStringGradeToInt(req.Country, req.Grade)
	if err != nil {
		return nil, fmt.Errorf("utils.ConvertStringGradeToInt: %w", err)
	}

	err = multierr.Combine(
		r.Name.Set(req.Name),
		r.Country.Set(req.Country.String()),
		r.Subject.Set(req.Subject.String()),
		r.Grade.Set(grade),
		r.DisplayOrder.Set(1),
		r.SchoolID.Set(req.SchoolId),
		r.StartDate.Set(startDate),
		r.EndDate.Set(endDate),
		r.TeacherIDs.Set(req.TeacherIds),
		r.DeletedAt.Set(nil),
	)

	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}
	return r, nil
}

func (s *CourseService) UpsertLiveCourse(ctx context.Context, req *pb.UpsertLiveCourseRequest) (*pb.UpsertLiveCourseResponse, error) {
	course, err := s.validAndConvertLiveCourseEntity(ctx, req)
	if err != nil {
		return nil, err
	}

	err = database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		err := s.CourseRepo.Upsert(ctx, tx, []*entities.Course{course})
		if err != nil {
			return fmt.Errorf("CourseRepo.Upsert: %w", err)
		}

		presetStudyPlan := &entities.PresetStudyPlan{}
		database.AllNullEntity(presetStudyPlan)
		err = multierr.Combine(
			presetStudyPlan.Country.Set(course.Country.String),
			presetStudyPlan.ID.Set(course.PresetStudyPlanID.String),
			presetStudyPlan.Name.Set(course.Name.String),
			presetStudyPlan.Grade.Set(course.Grade.Int),
			presetStudyPlan.Subject.Set(course.Subject.String),
			presetStudyPlan.StartDate.Set(course.StartDate.Time),
		)
		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}

		err = s.PresetStudyPlanRepo.Upsert(ctx, tx, presetStudyPlan)
		if err != nil {
			return fmt.Errorf("PresetStudyPlanRepo.Insert: %w", err)
		}
		mapCourseClasses, err := s.CourseClassRepo.FindByCourseID(ctx, tx, course.ID, true)
		if err != nil {
			return fmt.Errorf("CourseClassRepo.FindByCourseIDs: %w", err)
		}

		courseClasses := []*entities.CourseClass{}

		for _, classID := range req.ClassIds {
			courseClass := &entities.CourseClass{}
			database.AllNullEntity(courseClass)
			err = multierr.Combine(
				courseClass.ClassID.Set(classID),
				courseClass.CourseID.Set(course.ID.String),
				courseClass.Status.Set(entities.CourseClassStatusActive),
				courseClass.DeletedAt.Set(nil),
			)
			if err != nil {
				return fmt.Errorf("multierr.Combine: %w", err)
			}

			_, ok := mapCourseClasses[database.Int4(classID)]
			if ok {
				// if classID is exist in course, remove out of mapCourseClasses to do not update status and deleted at
				delete(mapCourseClasses, database.Int4(classID))
			}
			courseClasses = append(courseClasses, courseClass)
		}
		if len(mapCourseClasses) > 0 {
			now := time.Now()
			for _, v := range mapCourseClasses {
				err := multierr.Combine(
					v.DeletedAt.Set(now),
					v.Status.Set(entities.CourseClassStatusInActive),
				)
				if err != nil {
					return fmt.Errorf("multierr.Combine: %v", err)
				}
				courseClasses = append(courseClasses, v)
			}
		}

		err = s.CourseClassRepo.UpsertV2(ctx, tx, courseClasses)
		if err != nil {
			return fmt.Errorf("CourseClassRepo.Upsert: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpsertLiveCourseResponse{
		Id: course.ID.String,
	}, nil
}

func checkStartAndEndDate(srcStartDate, srcEndDate *types.Timestamp) (time.Time, time.Time, error) {
	now := time.Now()
	startDate, err := types.TimestampFromProto(srcStartDate)
	if err != nil {
		return now, now, fmt.Errorf("types.TimestampFromProto: %w", err)
	}
	endDate, err := types.TimestampFromProto(srcEndDate)
	if err != nil {
		return now, now, fmt.Errorf("types.TimestampFromProto: %w", err)
	}
	if startDate.IsZero() || endDate.IsZero() {
		return now, now, status.Error(codes.InvalidArgument, "start date and end date cannot be empty")
	}
	if endDate.Before(startDate) {
		return now, now, status.Error(codes.InvalidArgument, "start date must before end date")
	}
	return startDate, endDate, nil
}

func (s *CourseService) DeleteLiveCourse(ctx context.Context, req *pb.DeleteLiveCourseRequest) (*pb.DeleteLiveCourseResponse, error) {
	courses, err := s.CourseRepo.FindByIDs(ctx, s.DBTrace, database.TextArray(req.CourseIds))
	if err != nil {
		return nil, errors.Wrap(err, "c.CourseRepo.FindByIDs")
	}

	if len(courses) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot find course")
	}

	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled(constant.SwitchDBUnleashKey, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Connect unleash server failed: %s", err))
	}

	if isUnleashToggled {
		err = s.deleteLiveCourseLessonmgmt(ctx, req)
		if err != nil {
			return nil, err
		}

		return &pb.DeleteLiveCourseResponse{}, nil
	}

	presetStudyPlanIDs := []string{}
	for _, v := range courses {
		presetStudyPlanIDs = append(presetStudyPlanIDs, v.PresetStudyPlanID.String)
	}

	err = database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		err := s.CourseRepo.SoftDelete(ctx, tx, database.TextArray(req.CourseIds))
		if err != nil {
			return fmt.Errorf("c.CourseRepo.SoftDelete: %w", err)
		}

		courseClasses, err := s.CourseClassRepo.FindByCourseIDs(ctx, s.DBTrace, database.TextArray(req.CourseIds), false)
		if err != nil {
			return errors.Wrap(err, "c.CourseRepo.FindByIDs")
		}
		if len(courseClasses) > 0 {
			err = s.CourseClassRepo.SoftDelete(ctx, tx, database.TextArray(req.CourseIds))
			if err != nil {
				return fmt.Errorf("c.CourseClassRepo.SoftDelete: %w", err)
			}
		}

		err = s.PresetStudyPlanRepo.SoftDelete(ctx, tx, database.TextArray(presetStudyPlanIDs))
		if err != nil {
			return fmt.Errorf("c.PresetStudyPlanRepo.SoftDelete: %w", err)
		}

		lessons, err := s.LessonRepo.FindByCourseIDs(ctx, s.DBTrace, database.TextArray(req.CourseIds), false)
		if err != nil {
			return errors.Wrap(err, "c.LessonRepo.FindByCourseIDs")
		}
		if len(lessons) != 0 {
			err = s.PresetStudyPlanWeeklyRepo.SoftDeleteByPresetStudyPlanIDs(ctx, tx, database.TextArray(presetStudyPlanIDs))
			if err != nil {
				return fmt.Errorf("c.PresetStudyPlanWeeklyRepo.SoftDeleteByPresetStudyPlanIDs: %w", err)
			}

			err = s.LessonRepo.SoftDeleteByCourseIDs(ctx, tx, database.TextArray(req.CourseIds))
			if err != nil {
				return fmt.Errorf("c.LessonRepo.SoftDeleteByCourseIDs: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.DeleteLiveCourseResponse{}, nil
}

func (s *CourseService) deleteLiveCourseLessonmgmt(ctx context.Context, req *pb.DeleteLiveCourseRequest) error {
	err := database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		err := s.CourseRepo.SoftDelete(ctx, tx, database.TextArray(req.CourseIds))
		if err != nil {
			return fmt.Errorf("c.CourseRepo.SoftDelete: %w", err)
		}

		courseClasses, err := s.CourseClassRepo.FindByCourseIDs(ctx, tx, database.TextArray(req.CourseIds), false)
		if err != nil {
			return errors.Wrap(err, "c.CourseRepo.FindByIDs")
		}
		if len(courseClasses) > 0 {
			err = s.CourseClassRepo.SoftDelete(ctx, tx, database.TextArray(req.CourseIds))
			if err != nil {
				return fmt.Errorf("c.CourseClassRepo.SoftDelete: %w", err)
			}
		}

		lessons, err := s.LessonRepo.FindByCourseIDs(ctx, s.LessonDBTrace, database.TextArray(req.CourseIds), false)
		if err != nil {
			return errors.Wrap(err, "c.LessonRepo.FindByCourseIDs")
		}
		if len(lessons) != 0 {
			err = s.LessonRepo.SoftDeleteByCourseIDs(ctx, s.LessonDBTrace, database.TextArray(req.CourseIds))
			if err != nil {
				return fmt.Errorf("c.LessonRepo.SoftDeleteByCourseIDs: %w", err)
			}
		}
		return nil
	})

	return err
}

func (s *CourseService) convertAttachmentsToString(src []*pb.Attachment) ([]string, []string) {
	attachmentNames := []string{}
	attachmentURLs := []string{}
	for _, att := range src {
		attachmentNames = append(attachmentNames, att.Name)
		attachmentURLs = append(attachmentURLs, att.Url)
	}
	return attachmentNames, attachmentURLs
}

func (s *CourseService) toLessonEnFromLessonPb(
	lesson *pb.CreateLiveLessonRequest_Lesson,
	course_id string,
	liveLessonID, lessonGroupID string,
	lessonStatus cpb.LessonStatus,
	lessonType cpb.LessonType,
	schedulingStatus string,
) (*entities.Lesson, error) {
	e := &entities.Lesson{}
	database.AllNullEntity(e)
	controlSetting := s.toControlSettingLiveLesson(lesson.ControlSettings)
	err := multierr.Combine(
		e.TeacherID.Set(lesson.TeacherId),
		e.LessonID.Set(liveLessonID),
		e.DeletedAt.Set(nil),
		e.CourseID.Set(course_id),
		e.ControlSettings.Set(controlSetting),
		e.LessonGroupID.Set(lessonGroupID),
		e.LessonType.Set(lessonType.String()),
		e.Status.Set(lessonStatus.String()),
		// the input didn't have two fields, so do not check and set default value for stream_learner_counter and learner_ids
		e.StreamLearnerCounter.Set(database.Int4(0)),
		e.LearnerIds.Set(database.JSONB([]byte("{}"))),
		e.SchedulingStatus.Set(schedulingStatus),
		e.IsLocked.Set(false),
	)
	if lessonType != cpb.LessonType_LESSON_TYPE_NONE {
		if lessonType == cpb.LessonType_LESSON_TYPE_ONLINE {
			err = multierr.Append(err, e.TeachingMedium.Set(entities.LessonTeachingMediumOnline))
		} else if lessonType == cpb.LessonType_LESSON_TYPE_OFFLINE {
			err = multierr.Append(err, e.TeachingMedium.Set(entities.LessonTeachingMediumOffline))
		}
	}
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}

func (s *CourseService) toPresetStudyPlanEnFromLessonPb(lesson *pb.CreateLiveLessonRequest_Lesson, presetStudyPlanID string, lessonID string, topicID string) (*entities.PresetStudyPlanWeekly, error) {
	e := &entities.PresetStudyPlanWeekly{}
	database.AllNullEntity(e)

	startDate, endDate, err := checkStartAndEndDate(lesson.StartDate, lesson.EndDate)
	if err != nil {
		return nil, err
	}

	err = multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.PresetStudyPlanID.Set(presetStudyPlanID),
		e.TopicID.Set(topicID),
		e.DeletedAt.Set(nil),
		e.StartDate.Set(startDate),
		e.EndDate.Set(endDate),
		e.LessonID.Set(lessonID),
		e.Week.Set(0),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}
	return e, nil
}

func (s *CourseService) toTopicEnFromLessonPb(lesson *pb.CreateLiveLessonRequest_Lesson, course *entities.Course) (*entities.Topic, error) {
	e := &entities.Topic{}
	database.AllNullEntity(e)

	now := time.Now()
	attachmentNames, attachmentURLs := s.convertAttachmentsToString(lesson.Attachments)
	err := multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.Name.Set(lesson.Name),
		e.Country.Set(course.Country.String),
		e.Grade.Set(course.Grade.Int),
		e.Subject.Set(course.Subject.String),
		e.SchoolID.Set(course.SchoolID.Int),
		e.TopicType.Set(pb_bob.TOPIC_TYPE_LIVE_LESSON.String()),
		e.Status.Set(pb.TOPIC_STATUS_PUBLISHED.String()),
		e.AttachmentNames.Set(attachmentNames),
		e.AttachmentURLs.Set(attachmentURLs),
		e.DisplayOrder.Set(1),
		e.PublishedAt.Set(now),
		e.TotalLOs.Set(0),
		e.ChapterID.Set(nil),
		e.IconURL.Set(nil),
		e.DeletedAt.Set(nil),
		e.EssayRequired.Set(false),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}
	return e, nil
}

func (s *CourseService) updateTimeLiveCourse(ctx context.Context, tx pgx.Tx, course *entities.Course) error {
	courseID := course.ID
	if courseID.String == "" {
		return status.Error(codes.InvalidArgument, "missing course")
	}
	startDate, endDate, err := s.LessonRepo.FindEarlierAndLatestTimeLesson(ctx, tx, courseID)
	if err != nil {
		return fmt.Errorf("LessonRepo.FindEarlierAndLatestTimeLesson: %w", err)
	}

	err = multierr.Combine(
		course.DeletedAt.Set(nil),
	)
	if startDate != nil && endDate != nil {
		err = multierr.Append(err, course.StartDate.Set(&startDate))
		err = multierr.Append(err, course.EndDate.Set(&endDate))
	}
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	err = s.CourseRepo.Upsert(ctx, tx, []*entities.Course{course})
	if err != nil {
		return fmt.Errorf("CourseRepo.UpsertOne: %w", err)
	}

	return nil
}

func (s *CourseService) updateTimeLiveCourseLessonmgmt(ctx context.Context, tx pgx.Tx, course *entities.Course) error {
	courseID := course.ID
	if courseID.String == "" {
		return status.Error(codes.InvalidArgument, "missing course")
	}
	startDate, endDate, err := s.LessonRepo.FindEarlierAndLatestTimeLesson(ctx, tx, courseID)
	if err != nil {
		return fmt.Errorf("LessonRepo.FindEarlierAndLatestTimeLesson: %w", err)
	}

	err = multierr.Combine(
		course.DeletedAt.Set(nil),
	)
	if startDate != nil && endDate != nil {
		err = multierr.Append(err, course.StartDate.Set(&startDate))
		err = multierr.Append(err, course.EndDate.Set(&endDate))
	}
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	err = s.CourseRepo.Upsert(ctx, s.DBTrace, []*entities.Course{course})
	if err != nil {
		return fmt.Errorf("CourseRepo.UpsertOne: %w", err)
	}

	return nil
}

func (s *CourseService) CreateLiveLesson(ctx context.Context, req *pb.CreateLiveLessonRequest) (*pb.CreateLiveLessonResponse, error) {
	lessons := make([]*LiveLessonOpt, 0, len(req.Lessons))
	for _, l := range req.Lessons {
		lessons = append(lessons, &LiveLessonOpt{
			CreateLiveLessonRequest_Lesson: l,
			PresetLessonID:                 idutil.ULIDNow(),
			LessonType:                     cpb.LessonType_LESSON_TYPE_ONLINE,
			LessonStatus:                   cpb.LessonStatus_LESSON_STATUS_NOT_STARTED,
		})
	}
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled(constant.SwitchDBUnleashKey, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Connect unleash server failed: %s", err))
	}

	if isUnleashToggled {
		err = s.createLiveLessonLessonmgmt(
			ctx,
			&CreateLiveLessonOpt{
				CourseID: req.CourseId,
				Lessons:  lessons,
			},
			false,
		)
		if err != nil {
			return nil, err
		}

		return &pb.CreateLiveLessonResponse{}, nil
	}
	err = s.createLiveLesson(ctx, &CreateLiveLessonOpt{
		CourseID: req.CourseId,
		Lessons:  lessons,
	}, false)
	if err != nil {
		return nil, err
	}

	return &pb.CreateLiveLessonResponse{}, nil
}

type LiveLessonOpt struct {
	*pb.CreateLiveLessonRequest_Lesson
	PresetLessonID   string
	LessonType       cpb.LessonType
	LessonStatus     cpb.LessonStatus
	SchedulingStatus string
}

type CreateLiveLessonOpt struct {
	CourseID string
	Lessons  []*LiveLessonOpt
}

func (s *CourseService) createLiveLesson(ctx context.Context, req *CreateLiveLessonOpt, isUpsert bool) error {
	course, err := s.CourseRepo.FindByID(ctx, s.DBTrace, database.Text(req.CourseID))
	if err != nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("c.CourseRepo.FindByID id: %v: %v", req.CourseID, err))
	}

	createNewPresetStudyPlan := course.PresetStudyPlanID.String == ""
	if createNewPresetStudyPlan {
		_ = course.PresetStudyPlanID.Set(idutil.ULIDNow())
	}

	enTopics := []*entities.Topic{}
	enLessons := []*entities.Lesson{}
	enPresetStudyPlanWeeklies := []*entities.PresetStudyPlanWeekly{}
	pbLessons := []*pb_bob.EvtLesson_Lesson{}
	pbLiveLessons := []*pb_bob.EvtLesson_Lesson{}
	for _, v := range req.Lessons {
		lesson, err := s.toLessonEnFromLessonPb(v.CreateLiveLessonRequest_Lesson, course.ID.String, v.PresetLessonID, v.LessonGroup, v.LessonStatus, v.LessonType, v.SchedulingStatus)
		if err != nil {
			return err
		}

		topic, err := s.toTopicEnFromLessonPb(v.CreateLiveLessonRequest_Lesson, course)
		if err != nil {
			return err
		}

		presetStudyPlanWeekly, err := s.toPresetStudyPlanEnFromLessonPb(v.CreateLiveLessonRequest_Lesson, course.PresetStudyPlanID.String, lesson.LessonID.String, topic.ID.String)
		if err != nil {
			return err
		}

		// set data for lesson new schema
		lesson.Name.Set(topic.Name.String)
		lesson.StartTime.Set(presetStudyPlanWeekly.StartDate)
		lesson.EndTime.Set(presetStudyPlanWeekly.EndDate)
		lesson.CenterID.Set(constants.JPREPOrgLocation)
		l := &pb_bob.EvtLesson_Lesson{
			LessonId: lesson.LessonID.String,
			Name:     v.Name,
		}
		pbLessons = append(pbLessons, l)

		if v.LessonType == cpb.LessonType_LESSON_TYPE_ONLINE {
			pbLiveLessons = append(pbLiveLessons, l)
		}

		enLessons = append(enLessons, lesson)
		enTopics = append(enTopics, topic)
		enPresetStudyPlanWeeklies = append(enPresetStudyPlanWeeklies, presetStudyPlanWeekly)
	}
	if len(enTopics) == 0 || len(enPresetStudyPlanWeeklies) == 0 || len(enLessons) == 0 {
		return status.Error(codes.Internal, "can't bulk-insert empty slice")
	}

	err = database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		if createNewPresetStudyPlan {
			presetStudyPlan := &entities.PresetStudyPlan{}
			database.AllNullEntity(presetStudyPlan)
			err = multierr.Combine(
				presetStudyPlan.Country.Set(course.Country.String),
				presetStudyPlan.ID.Set(course.PresetStudyPlanID.String),
				presetStudyPlan.Name.Set(course.Name.String),
				presetStudyPlan.Grade.Set(course.Grade.Int),
				presetStudyPlan.Subject.Set(course.Subject.String),
				presetStudyPlan.StartDate.Set(course.StartDate.Time),
			)
			if err != nil {
				return fmt.Errorf("multierr.Combine: %w", err)
			}

			err = s.PresetStudyPlanRepo.Upsert(ctx, tx, presetStudyPlan)
			if err != nil {
				return fmt.Errorf("PresetStudyPlanRepo.Insert: %w", err)
			}
		}

		err := s.TopicRepo.Create(ctx, tx, enTopics)
		if err != nil {
			return fmt.Errorf("TopicRepo.Create: %w", err)
		}
		if !isUpsert {
			err = s.LessonRepo.Create(ctx, tx, enLessons)
			if err != nil {
				return fmt.Errorf("LessonRepo.Create: %w", err)
			}
		} else {
			err = s.LessonRepo.BulkUpsert(ctx, tx, enLessons)
			if err != nil {
				return fmt.Errorf("LessonRepo.BulkUpsert: %w", err)
			}
		}
		err = s.PresetStudyPlanWeeklyRepo.Create(ctx, tx, enPresetStudyPlanWeeklies)
		if err != nil {
			return fmt.Errorf("PresetStudyPlanWeeklyRepo.Create: %w", err)
		}
		for _, lesson := range enLessons {
			if needUpdateCourseTime(course, lesson) {
				err = s.updateTimeLiveCourse(ctx, tx, course)
				if err != nil {
					return err
				}
			}
		}
		if len(pbLiveLessons) > 0 {
			err = s.PublishLessonEvt(ctx, &pb_bob.EvtLesson{
				Message: &pb_bob.EvtLesson_CreateLessons_{
					CreateLessons: &pb_bob.EvtLesson_CreateLessons{
						Lessons: pbLiveLessons,
					},
				},
			})

			if err != nil {
				return errors.Wrap(err, "rcv.PublishLessonEvt")
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *CourseService) createLiveLessonLessonmgmt(ctx context.Context, req *CreateLiveLessonOpt, isUpsert bool) error {
	course, err := s.CourseRepo.FindByID(ctx, s.DBTrace, database.Text(req.CourseID))
	if err != nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("c.CourseRepo.FindByID id: %v: %v", req.CourseID, err))
	}

	enLessons := []*entities.Lesson{}
	pbLiveLessons := []*pb_bob.EvtLesson_Lesson{}
	for _, v := range req.Lessons {
		lesson, err := s.toLessonEnFromLessonPb(v.CreateLiveLessonRequest_Lesson, course.ID.String, v.PresetLessonID, v.LessonGroup, v.LessonStatus, v.LessonType, v.SchedulingStatus)
		if err != nil {
			return err
		}
		startDate, endDate, err := checkStartAndEndDate(v.CreateLiveLessonRequest_Lesson.StartDate, v.CreateLiveLessonRequest_Lesson.EndDate)
		if err != nil {
			return err
		}

		// set data for lesson new schema
		err = multierr.Combine(
			lesson.Name.Set(v.CreateLiveLessonRequest_Lesson.Name),
			lesson.StartTime.Set(startDate),
			lesson.EndTime.Set(endDate),
			lesson.CenterID.Set(constants.JPREPOrgLocation),
		)
		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}

		l := &pb_bob.EvtLesson_Lesson{
			LessonId: lesson.LessonID.String,
			Name:     v.Name,
		}

		if v.LessonType == cpb.LessonType_LESSON_TYPE_ONLINE {
			pbLiveLessons = append(pbLiveLessons, l)
		}

		enLessons = append(enLessons, lesson)
	}
	if len(enLessons) == 0 {
		return status.Error(codes.Internal, "can't bulk-insert empty slice")
	}

	err = database.ExecInTx(ctx, s.LessonDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		if !isUpsert {
			err = s.LessonRepo.Create(ctx, tx, enLessons)
			if err != nil {
				return fmt.Errorf("LessonRepo.Create: %w", err)
			}
		} else {
			err = s.LessonRepo.BulkUpsert(ctx, tx, enLessons)
			if err != nil {
				return fmt.Errorf("LessonRepo.BulkUpsert: %w", err)
			}
		}
		for _, lesson := range enLessons {
			if needUpdateCourseTime(course, lesson) {
				err = s.updateTimeLiveCourseLessonmgmt(ctx, tx, course)
				if err != nil {
					return err
				}
			}
		}
		if len(pbLiveLessons) > 0 {
			err = s.PublishLessonEvt(ctx, &pb_bob.EvtLesson{
				Message: &pb_bob.EvtLesson_CreateLessons_{
					CreateLessons: &pb_bob.EvtLesson_CreateLessons{
						Lessons: pbLiveLessons,
					},
				},
			})

			if err != nil {
				return errors.Wrap(err, "rcv.PublishLessonEvt")
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func needUpdateCourseTime(course *entities.Course, lesson *entities.Lesson) bool {
	if course.StartDate.Status != pgtype.Present || course.EndDate.Status != pgtype.Present {
		return true
	}
	if lesson.StartTime.Time.Before(course.StartDate.Time) {
		return true
	}
	if lesson.EndTime.Time.After(course.EndDate.Time) {
		return true
	}
	return false
}

func validControlSettingLiveLesson(req *pb.ControlSettingLiveLesson) error {
	if len(req.Lectures) == 0 {
		return status.Error(codes.InvalidArgument, "missing teacher teach")
	}
	if req.PublishStudentVideoStatus == pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_NONE {
		return status.Error(codes.InvalidArgument, "missing publish student video status")
	}
	if req.UnmuteStudentAudioStatus == pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_NONE {
		return status.Error(codes.InvalidArgument, "missing unmute student audio status")
	}
	if req.DefaultView == pb_bob.LIVE_LESSON_VIEW_NONE {
		return status.Error(codes.InvalidArgument, "missing live lesson view")
	}
	return nil
}

func (c *CourseService) toControlSettingLiveLesson(req *pb.ControlSettingLiveLesson) *yasuo_repositories.ControlSettingLiveLesson {
	if req == nil {
		return nil
	}

	return &yasuo_repositories.ControlSettingLiveLesson{
		TeacherObservers:          req.TeacherObversers,
		Lectures:                  req.Lectures,
		DefaultView:               req.DefaultView.String(),
		PublishStudentVideoStatus: req.PublishStudentVideoStatus.String(),
		UnmuteStudentAudioStatus:  req.UnmuteStudentAudioStatus.String(),
	}
}

func (c *CourseService) handleControlSettingLiveLesson(ctx context.Context, req *pb.ControlSettingLiveLesson, schoolID pgtype.Int4) error {
	err := validControlSettingLiveLesson(req)
	if err != nil {
		return err
	}

	isInSchool, err := c.TeacherRepo.ManyTeacherIsInSchool(ctx, c.DBTrace, database.TextArray(append(req.TeacherObversers, req.Lectures...)), schoolID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.NotFound, "cannot find teacher")
		}
		return fmt.Errorf("TeacherRepo.ManyTeacherIsInSchool: %w", err)
	}

	if !isInSchool {
		return status.Error(codes.PermissionDenied, "teacher doesnot is in school")
	}

	return nil
}

type updateLiveLessonV2Request struct {
	*pb.UpdateLiveLessonRequest
	LessonType cpb.LessonType
}

func (s *CourseService) updateLiveLessonV2(ctx context.Context, req *updateLiveLessonV2Request) (*pb.UpdateLiveLessonResponse, error) {
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
	err = database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		attachmentNames, attachmentURLs := s.convertAttachmentsToString(req.Attachments)
		startDate, endDate, err := checkStartAndEndDate(req.StartDate, req.EndDate)
		startDateUpdated := !presetStudyPlanWeekly.StartDate.Time.Equal(startDate)
		updateLesson := canUpdateLessonNameAndTime(lesson, topic, req.Name, presetStudyPlanWeekly, startDate, endDate)
		if canUpdateTopic(topic, req.Name, attachmentNames, attachmentURLs) {
			err = s.TopicRepo.Update(ctx, tx, topic)
			if err != nil {
				return fmt.Errorf("TopicRepo.Update: %w", err)
			}
		}
		if err != nil {
			return fmt.Errorf("checkStartAndEndDate: %w", err)
		}

		if canUpdatePresetStudyPlanWeekly(
			presetStudyPlanWeekly,
			course.PresetStudyPlanID.String,
			startDate,
			endDate,
		) {
			err = s.PresetStudyPlanWeeklyRepo.Update(ctx, tx, presetStudyPlanWeekly)
			if err != nil {
				return fmt.Errorf("PresetStudyPlanWeeklyRepo.Update: %w", err)
			}
		}

		if canUpdateLesson(lesson, req.TeacherId, req.CourseId, req.LessonGroup, req.LessonType) || updateLesson {
			lesson.ControlSettings.Set(s.toControlSettingLiveLesson(req.ControlSettings))

			if startDateUpdated && startDate.After(time.Now().Add(1*time.Minute)) {
				err = s.LiveLessonSentNotificationRepo.SoftDeleteLiveLessonSentNotificationRecord(ctx, tx, lesson.LessonID.String)
				if err != nil {
					return fmt.Errorf("LiveLessonSentNotificationRepo.SoftDeleteLiveLessonSentNotificationRecord(lesson id: %s): %w", lesson.LessonID.String, err)
				}
			}

			err = s.LessonRepo.Update(ctx, tx, lesson)
			if err != nil {
				return fmt.Errorf("LessonRepo.Update: %w", err)
			}
		}
		if needUpdateCourseTime(course, lesson) {
			err = s.updateTimeLiveCourse(ctx, tx, course)
		}
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateLiveLessonResponse{}, nil
}

func (s *CourseService) updateLiveLessonV2Lessonmgmt(ctx context.Context, req *updateLiveLessonV2Request) error {
	lesson, err := s.LessonRepo.FindByID(ctx, s.LessonDBTrace, database.Text(req.LessonId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.NotFound, "cannot find lesson")
		}
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}

	if req.CourseId == "" {
		req.CourseId = lesson.CourseID.String
	}

	course, err := s.CourseRepo.FindByID(ctx, s.DBTrace, database.Text(req.CourseId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.NotFound, "cannot find course")
		}
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}

	err = database.ExecInTx(ctx, s.LessonDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		startDate, endDate, err := checkStartAndEndDate(req.StartDate, req.EndDate)
		startDateUpdated := !lesson.StartTime.Time.Equal(startDate)
		if err != nil {
			return fmt.Errorf("checkStartAndEndDate: %w", err)
		}

		updateLesson := updateLessonNameAndTimeLessonmgmt(lesson, req.GetName(), startDate, endDate)

		if canUpdateLesson(lesson, req.TeacherId, req.CourseId, req.LessonGroup, req.LessonType) || updateLesson {
			lesson.ControlSettings.Set(s.toControlSettingLiveLesson(req.ControlSettings))

			if startDateUpdated && startDate.After(time.Now().Add(1*time.Minute)) {
				err = s.LiveLessonSentNotificationRepo.SoftDeleteLiveLessonSentNotificationRecord(ctx, tx, lesson.LessonID.String)
				if err != nil {
					return fmt.Errorf("LiveLessonSentNotificationRepo.SoftDeleteLiveLessonSentNotificationRecord(lesson id: %s): %w", lesson.LessonID.String, err)
				}
			}

			err = s.LessonRepo.Update(ctx, tx, lesson)
			if err != nil {
				return fmt.Errorf("LessonRepo.Update: %w", err)
			}
		}
		if needUpdateCourseTime(course, lesson) {
			err = s.updateTimeLiveCourseLessonmgmt(ctx, tx, course)
		}
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *CourseService) UpdateLiveLesson(ctx context.Context, req *pb.UpdateLiveLessonRequest) (*pb.UpdateLiveLessonResponse, error) {
	reqV2 := &updateLiveLessonV2Request{UpdateLiveLessonRequest: req, LessonType: cpb.LessonType_LESSON_TYPE_NONE}
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled(constant.SwitchDBUnleashKey, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Connect unleash server failed: %s", err))
	}

	if isUnleashToggled {
		err = s.updateLiveLessonV2Lessonmgmt(ctx, reqV2)
		if err != nil {
			return nil, err
		}
		return &pb.UpdateLiveLessonResponse{}, nil
	}
	return s.updateLiveLessonV2(ctx, reqV2)
}

func (s *CourseService) DeleteLiveLesson(ctx context.Context, req *pb.DeleteLiveLessonRequest) (*pb.DeleteLiveLessonResponse, error) {
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled(constant.SwitchDBUnleashKey, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Connect unleash server failed: %s", err))
	}
	if isUnleashToggled {
		return s.DeleteLiveLessonLessonmgmt(ctx, req)
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

	presetStudyPlanWeeklies, err := s.PresetStudyPlanWeeklyRepo.FindByLessonIDs(ctx, s.DBTrace, database.TextArray(req.LessonIds), false)
	if err != nil {
		return nil, fmt.Errorf("PresetStudyPlanWeeklyRepo.FindByLessonIDs: %w", err)
	}
	if len(presetStudyPlanWeeklies) == 0 {
		return nil, status.Error(codes.NotFound, "cannot find preset study plan weekly")
	}

	topicIDs := []string{}
	presetStudyPlanWeeklyIDs := []string{}
	for id, v := range presetStudyPlanWeeklies {
		topicIDs = append(topicIDs, v.TopicID.String)
		presetStudyPlanWeeklyIDs = append(presetStudyPlanWeeklyIDs, id.String)
	}

	err = database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		err = s.TopicRepo.SoftDeleteV2(ctx, tx, database.TextArray(topicIDs))
		if err != nil {
			return fmt.Errorf("TopicRepo.SoftDeleteV2: %w", err)
		}

		err = s.PresetStudyPlanWeeklyRepo.SoftDelete(ctx, tx, database.TextArray(presetStudyPlanWeeklyIDs))
		if err != nil {
			return fmt.Errorf("PresetStudyPlanWeeklyRepo.SoftDelete: %w", err)
		}

		err = s.LessonRepo.SoftDelete(ctx, tx, database.TextArray(req.LessonIds))
		if err != nil {
			return fmt.Errorf("LessonRepo.SoftDelete: %w", err)
		}

		for _, course := range courses {
			err = s.updateTimeLiveCourse(ctx, tx, course)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.DeleteLiveLessonResponse{}, nil
}

func (s *CourseService) DeleteLiveLessonLessonmgmt(ctx context.Context, req *pb.DeleteLiveLessonRequest) (*pb.DeleteLiveLessonResponse, error) {
	lessons, err := s.LessonRepo.FindByIDs(ctx, s.LessonDBTrace, database.TextArray(req.LessonIds), false)
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

	err = database.ExecInTx(ctx, s.LessonDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		err = s.LessonRepo.SoftDelete(ctx, tx, database.TextArray(req.LessonIds))
		if err != nil {
			return fmt.Errorf("LessonRepo.SoftDelete: %w", err)
		}

		for _, course := range courses {
			err = s.updateTimeLiveCourseLessonmgmt(ctx, tx, course)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.DeleteLiveLessonResponse{}, nil
}

func (s *CourseService) PublishSyncStudentLessons(ctx context.Context, msg *npb.EventSyncUserCourse) error {
	data, _ := proto.Marshal(msg)
	_, err := s.JSM.PublishContext(ctx, constants.SubjectSyncStudentLessons, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishSyncStudentLessons s.JSM.PublishContext failed, %w", err))
	}

	return err
}

func (s *CourseService) PublishLessonEvt(ctx context.Context, msg *pb_bob.EvtLesson) error {
	var msgID string
	data, _ := msg.Marshal()

	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectLessonCreated, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishLessonEvt rcv.JSM.PublishAsyncContext Lesson.Created failed, msgID: %s, %w", msgID, err))
	}

	return err
}

func (s *CourseService) SyncLiveLesson(ctx context.Context, req []*npb.EventMasterRegistration_Lesson) error {
	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled(constant.SwitchDBUnleashKey, s.Env)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Connect unleash server failed: %s", err))
	}

	if isUnleashToggled {
		return s.SyncLiveLessonLessonmgmt(ctx, req)
	}

	mapByCourseID := make(map[string][]*LiveLessonOpt)
	deleteLessonIDs := []string{}

	var errR error
	for _, l := range req {
		switch l.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			lesson, err := s.LessonRepo.FindByID(ctx, s.DBTrace, database.Text(l.LessonId))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				errR = multierr.Append(errR, err)
				continue
			}

			if l.LessonGroup != "" {
				err := s.createLessonGroup(ctx, l.LessonGroup, l.CourseId)
				if err != nil {
					errR = multierr.Append(errR, fmt.Errorf("err createLessonGroup: %w", err))
					continue
				}
			}

			if lesson != nil { // update
				_, err := s.updateLiveLessonV2(ctx, &updateLiveLessonV2Request{
					UpdateLiveLessonRequest: &pb.UpdateLiveLessonRequest{
						LessonId: l.LessonId,
						Name:     l.ClassName,
						StartDate: &types.Timestamp{
							Seconds: l.StartDate.Seconds,
							Nanos:   l.StartDate.Nanos,
						},
						EndDate: &types.Timestamp{
							Seconds: l.EndDate.Seconds,
							Nanos:   l.EndDate.Nanos,
						},
						CourseId:    l.CourseId,
						LessonGroup: l.LessonGroup,
					},
					LessonType: l.LessonType,
				})

				if err != nil {
					errR = multierr.Append(errR, err)
				}

				continue
			}

			mapByCourseID[l.CourseId] = append(mapByCourseID[l.CourseId], &LiveLessonOpt{
				CreateLiveLessonRequest_Lesson: &pb.CreateLiveLessonRequest_Lesson{
					Name: l.ClassName,
					StartDate: &types.Timestamp{
						Seconds: l.StartDate.Seconds,
						Nanos:   l.StartDate.Nanos,
					},
					EndDate: &types.Timestamp{
						Seconds: l.EndDate.Seconds,
						Nanos:   l.EndDate.Nanos,
					},
					LessonGroup: l.LessonGroup,
				},
				PresetLessonID:   l.LessonId,
				LessonType:       l.LessonType,
				LessonStatus:     cpb.LessonStatus_LESSON_STATUS_DRAFT,
				SchedulingStatus: string(entities.LessonSchedulingStatusPublished),
			})
		case npb.ActionKind_ACTION_KIND_DELETED:
			deleteLessonIDs = append(deleteLessonIDs, l.LessonId)
		}
	}

	for courseID, l := range mapByCourseID {
		err := s.createLiveLesson(ctx, &CreateLiveLessonOpt{
			CourseID: courseID,
			Lessons:  l,
		}, true)
		if err != nil {
			errR = multierr.Append(errR, err)
			continue
		}
	}

	if len(deleteLessonIDs) > 0 {
		_, err := s.DeleteLiveLesson(ctx, &pb.DeleteLiveLessonRequest{
			LessonIds: deleteLessonIDs,
		})
		if err != nil {
			errR = multierr.Append(errR, fmt.Errorf("err s.DeleteLiveLesson: %w", err))
			return errR
		}
	}

	return errR
}

func (s *CourseService) SyncLiveLessonLessonmgmt(ctx context.Context, req []*npb.EventMasterRegistration_Lesson) error {
	mapByCourseID := make(map[string][]*LiveLessonOpt)
	deleteLessonIDs := []string{}

	var errR error
	for _, l := range req {
		switch l.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			lesson, err := s.LessonRepo.FindByID(ctx, s.LessonDBTrace, database.Text(l.LessonId))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				errR = multierr.Append(errR, err)
				continue
			}

			if l.LessonGroup != "" {
				err := s.createLessonGroupLessonmgmt(ctx, l.LessonGroup, l.CourseId)
				if err != nil {
					errR = multierr.Append(errR, fmt.Errorf("err createLessonGroup: %w", err))
					continue
				}
			}

			if lesson != nil { // update
				err := s.updateLiveLessonV2Lessonmgmt(ctx, &updateLiveLessonV2Request{
					UpdateLiveLessonRequest: &pb.UpdateLiveLessonRequest{
						LessonId: l.LessonId,
						Name:     l.ClassName,
						StartDate: &types.Timestamp{
							Seconds: l.StartDate.Seconds,
							Nanos:   l.StartDate.Nanos,
						},
						EndDate: &types.Timestamp{
							Seconds: l.EndDate.Seconds,
							Nanos:   l.EndDate.Nanos,
						},
						CourseId:    l.CourseId,
						LessonGroup: l.LessonGroup,
					},
					LessonType: l.LessonType,
				})

				if err != nil {
					errR = multierr.Append(errR, err)
				}

				continue
			}

			mapByCourseID[l.CourseId] = append(mapByCourseID[l.CourseId], &LiveLessonOpt{
				CreateLiveLessonRequest_Lesson: &pb.CreateLiveLessonRequest_Lesson{
					Name: l.ClassName,
					StartDate: &types.Timestamp{
						Seconds: l.StartDate.Seconds,
						Nanos:   l.StartDate.Nanos,
					},
					EndDate: &types.Timestamp{
						Seconds: l.EndDate.Seconds,
						Nanos:   l.EndDate.Nanos,
					},
					LessonGroup: l.LessonGroup,
				},
				PresetLessonID:   l.LessonId,
				LessonType:       l.LessonType,
				LessonStatus:     cpb.LessonStatus_LESSON_STATUS_DRAFT,
				SchedulingStatus: string(entities.LessonSchedulingStatusPublished),
			})
		case npb.ActionKind_ACTION_KIND_DELETED:
			deleteLessonIDs = append(deleteLessonIDs, l.LessonId)
		}
	}

	for courseID, l := range mapByCourseID {
		err := s.createLiveLessonLessonmgmt(
			ctx,
			&CreateLiveLessonOpt{
				CourseID: courseID,
				Lessons:  l,
			},
			false,
		)
		if err != nil {
			errR = multierr.Append(errR, err)
			continue
		}
	}

	if len(deleteLessonIDs) > 0 {
		_, err := s.DeleteLiveLessonLessonmgmt(ctx,
			&pb.DeleteLiveLessonRequest{
				LessonIds: deleteLessonIDs,
			})
		if err != nil {
			errR = multierr.Append(errR, fmt.Errorf("err s.DeleteLiveLesson: %w", err))
			return errR
		}
	}

	return errR
}

func (s *CourseService) handleStudentLessonUpsert(ctx context.Context, studentID string, lessonIDs []string) (added []string, removed []string, err error) {
	lessionsOfStudent, err := s.LessonMemberRepo.Find(ctx, s.DBTrace, database.Text(studentID))
	if err != nil {
		return nil, nil, err
	}

	lessionIds := []string{}
	for _, lessionItem := range lessionsOfStudent {
		lessionIds = append(lessionIds, lessionItem.LessonID.String)
	}
	lessionFromReg := []string{}
	lessionFromReg = append(lessionFromReg, lessonIDs...)
	_, added, removed = utils.Compare(lessionIds, lessionFromReg)

	var errB error

	validIDs, invalidIDs, err := s.LessonRepo.CheckExisted(ctx, s.DBTrace, database.TextArray(added))
	if err != nil {
		return nil, nil, fmt.Errorf("err s.LessonRepo.CheckExisted: %w", err)
	}

	err = s.upsertStudentLesson(ctx, studentID, validIDs)
	if err != nil {
		errB = multierr.Append(errB, fmt.Errorf("s.upsertStudentLesson: %w", err))
		added = nil
	} else {
		added = validIDs
	}

	err = s.deleteStudentLesson(ctx, studentID, removed)
	if err != nil {
		errB = multierr.Append(errB, fmt.Errorf("s.softDeleteStudentLesson: %w", err))
		removed = nil
	}

	if len(invalidIDs) > 0 {
		errB = multierr.Append(errB, fmt.Errorf("studentID %s can not join not existed lessons %v", studentID, invalidIDs))
	}
	return added, removed, errB
}

func (s *CourseService) handleStudentLessonUpsertLessonmgmt(ctx context.Context, studentID string, lessonIDs []string) (added []string, removed []string, err error) {
	lessionsOfStudent, err := s.LessonMemberRepo.Find(ctx, s.LessonDBTrace, database.Text(studentID))
	if err != nil {
		return nil, nil, err
	}

	lessionIds := []string{}
	for _, lessionItem := range lessionsOfStudent {
		lessionIds = append(lessionIds, lessionItem.LessonID.String)
	}
	lessionFromReg := []string{}
	lessionFromReg = append(lessionFromReg, lessonIDs...)
	_, added, removed = utils.Compare(lessionIds, lessionFromReg)

	var errB error

	validIDs, invalidIDs, err := s.LessonRepo.CheckExisted(ctx, s.LessonDBTrace, database.TextArray(added))
	if err != nil {
		return nil, nil, fmt.Errorf("err s.LessonRepo.CheckExisted: %w", err)
	}

	err = s.upsertStudentLessonLessonmgmt(ctx, studentID, validIDs)
	if err != nil {
		errB = multierr.Append(errB, fmt.Errorf("s.upsertStudentLessonLessonmgmt: %w", err))
		added = nil
	} else {
		added = validIDs
	}

	err = s.deleteStudentLessonLessonmgmt(ctx, studentID, removed)
	if err != nil {
		errB = multierr.Append(errB, fmt.Errorf("s.softDeleteStudentLessonLessonmgmt: %w", err))
		removed = nil
	}

	if len(invalidIDs) > 0 {
		errB = multierr.Append(errB, fmt.Errorf("studentID %s can not join not existed lessons %v", studentID, invalidIDs))
	}
	return added, removed, errB
}

func (s *CourseService) SyncStudentLesson(ctx context.Context, req []*npb.EventSyncUserCourse_StudentLesson) error {
	newEvts := make([]*npb.EventSyncUserCourse_StudentLesson, 0, len(req))

	isUnleashToggled, err := s.UnleashClientIns.IsFeatureEnabled(constant.SwitchDBUnleashKey, s.Env)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("Connect unleash server failed: %s", err))
	}

	if isUnleashToggled {
		err = s.SyncStudentLessonLessonmgmt(ctx, req)
		if err != nil {
			return err
		}
		return nil
	}

	var cErr error
	for _, r := range req {
		item := r
		switch item.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			added, removed, err := s.handleStudentLessonUpsert(ctx, item.StudentId, item.LessonIds)
			if err != nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("err handleStudentLessonUpsert: %w", err))
			} else {
				if len(added) > 0 {
					liveLessons, err := s.getLiveLessons(ctx, s.DBTrace, added)
					if err != nil {
						cErr = multierr.Combine(cErr, fmt.Errorf("err s.LessonRepo.GetLiveLessons: %w", err))
					} else {
						newEvts = append(newEvts, &npb.EventSyncUserCourse_StudentLesson{
							StudentId:  item.StudentId,
							ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
							LessonIds:  liveLessons,
						})
					}
				}
				if len(removed) > 0 {
					liveLessons, err := s.getLiveLessons(ctx, s.DBTrace, removed)
					if err != nil {
						cErr = multierr.Combine(cErr, fmt.Errorf("err s.LessonRepo.GetLiveLessons: %w", err))
					} else {
						newEvts = append(newEvts, &npb.EventSyncUserCourse_StudentLesson{
							StudentId:  item.StudentId,
							ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
							LessonIds:  liveLessons,
						})
					}
				}
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			err := s.deleteStudentLesson(ctx, item.StudentId, item.LessonIds)
			if err != nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("s.deleteStudentLesson: %w", err))
			}
			liveLessons, err := s.getLiveLessons(ctx, s.DBTrace, item.LessonIds)
			if err != nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("err s.LessonRepo.GetLiveLessons: %w", err))
			} else {
				newEvts = append(newEvts, &npb.EventSyncUserCourse_StudentLesson{
					StudentId:  item.StudentId,
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					LessonIds:  liveLessons,
				})
			}
		}
	}
	if len(newEvts) > 0 {
		internalSync := &npb.EventSyncUserCourse{
			StudentLessons: newEvts,
		}
		err := s.PublishSyncStudentLessons(ctx, internalSync)
		if err != nil {
			return fmt.Errorf("s.PublishSyncStudentLessons: %w", err)
		}
	}
	if cErr != nil {
		return cErr
	}

	return nil
}

func (s *CourseService) SyncStudentLessonLessonmgmt(ctx context.Context, req []*npb.EventSyncUserCourse_StudentLesson) error {
	newEvts := make([]*npb.EventSyncUserCourse_StudentLesson, 0, len(req))
	var cErr error
	for _, r := range req {
		item := r
		switch item.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			added, removed, err := s.handleStudentLessonUpsertLessonmgmt(ctx, item.StudentId, item.LessonIds)
			if err != nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("err handleStudentLessonUpsert: %w", err))
			} else {
				if len(added) > 0 {
					liveLessons, err := s.getLiveLessons(ctx, s.LessonDBTrace, added)
					if err != nil {
						cErr = multierr.Combine(cErr, fmt.Errorf("err s.LessonRepo.GetLiveLessons: %w", err))
					} else {
						newEvts = append(newEvts, &npb.EventSyncUserCourse_StudentLesson{
							StudentId:  item.StudentId,
							ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
							LessonIds:  liveLessons,
						})
					}
				}
				if len(removed) > 0 {
					liveLessons, err := s.getLiveLessons(ctx, s.LessonDBTrace, removed)
					if err != nil {
						cErr = multierr.Combine(cErr, fmt.Errorf("err s.LessonRepo.GetLiveLessons: %w", err))
					} else {
						newEvts = append(newEvts, &npb.EventSyncUserCourse_StudentLesson{
							StudentId:  item.StudentId,
							ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
							LessonIds:  liveLessons,
						})
					}
				}
			}
		case npb.ActionKind_ACTION_KIND_DELETED:
			err := s.deleteStudentLessonLessonmgmt(ctx, item.StudentId, item.LessonIds)
			if err != nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("s.deleteStudentLesson: %w", err))
			}
			liveLessons, err := s.getLiveLessons(ctx, s.LessonDBTrace, item.LessonIds)
			if err != nil {
				cErr = multierr.Combine(cErr, fmt.Errorf("err s.LessonRepo.GetLiveLessons: %w", err))
			} else {
				newEvts = append(newEvts, &npb.EventSyncUserCourse_StudentLesson{
					StudentId:  item.StudentId,
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					LessonIds:  liveLessons,
				})
			}
		}
	}
	if len(newEvts) > 0 {
		internalSync := &npb.EventSyncUserCourse{
			StudentLessons: newEvts,
		}
		err := s.PublishSyncStudentLessons(ctx, internalSync)
		if err != nil {
			return fmt.Errorf("s.PublishSyncStudentLessons: %w", err)
		}
	}
	if cErr != nil {
		return cErr
	}

	return nil
}

func (s *CourseService) getLiveLessons(ctx context.Context, db database.Ext, lessonIDs []string) (validIDs []string, err error) {
	return s.LessonRepo.GetLiveLessons(ctx, db, database.TextArray(lessonIDs))
}

func (s *CourseService) upsertStudentLesson(ctx context.Context, studentID string, lessonIDs []string) error {
	if len(lessonIDs) == 0 {
		return nil
	}

	now := time.Now()
	b := &pgx.Batch{}
	for _, lessonID := range lessonIDs {
		e := &entities.LessonMember{}
		database.AllNullEntity(e)
		err := multierr.Combine(
			e.LessonID.Set(lessonID),
			e.UserID.Set(studentID),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("err set LessonMember: %w", err)
		}

		s.LessonMemberRepo.UpsertQueue(b, e)
	}

	return database.ExecInTx(ctx, s.DBTrace, func(ctx context.Context, tx pgx.Tx) error {
		r := tx.SendBatch(ctx, b)
		defer r.Close()

		for i := 0; i < b.Len(); i++ {
			_, err := r.Exec()
			if err != nil {
				return fmt.Errorf("r.Exec: %w", err)
			}
		}

		return nil
	})
}

func (s *CourseService) upsertStudentLessonLessonmgmt(ctx context.Context, studentID string, lessonIDs []string) error {
	if len(lessonIDs) == 0 {
		return nil
	}

	now := time.Now()
	b := &pgx.Batch{}
	for _, lessonID := range lessonIDs {
		e := &entities.LessonMember{}
		database.AllNullEntity(e)
		err := multierr.Combine(
			e.LessonID.Set(lessonID),
			e.UserID.Set(studentID),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("err set LessonMember: %w", err)
		}

		s.LessonMemberRepo.UpsertQueue(b, e)
	}

	return database.ExecInTx(ctx, s.LessonDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		r := tx.SendBatch(ctx, b)
		defer r.Close()

		for i := 0; i < b.Len(); i++ {
			_, err := r.Exec()
			if err != nil {
				return fmt.Errorf("r.Exec: %w", err)
			}
		}

		return nil
	})
}

func (s *CourseService) deleteStudentLesson(ctx context.Context, studentID string, lessonIDs []string) error {
	if len(lessonIDs) == 0 {
		return nil
	}

	err := s.LessonMemberRepo.SoftDelete(
		ctx,
		s.DBTrace,
		database.Text(studentID),
		database.TextArray(lessonIDs),
	)
	if err != nil {
		return fmt.Errorf("err s.LessonMemberRepo.SoftDelete: %w", err)
	}

	return nil
}

func (s *CourseService) deleteStudentLessonLessonmgmt(ctx context.Context, studentID string, lessonIDs []string) error {
	if len(lessonIDs) == 0 {
		return nil
	}

	err := s.LessonMemberRepo.SoftDelete(
		ctx,
		s.LessonDBTrace,
		database.Text(studentID),
		database.TextArray(lessonIDs),
	)
	if err != nil {
		return fmt.Errorf("err s.LessonMemberRepo.SoftDeleteLessonmgmt: %w", err)
	}

	return nil
}

func (s *CourseService) createLessonGroup(ctx context.Context, lessonGroupID, courseID string) error {
	l, err := s.LessonGroupRepo.Get(ctx, s.DBTrace, database.Text(lessonGroupID), database.Text(courseID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("s.LessonGroupRepo.Get: %w", err)
	}

	if l != nil {
		return nil
	}

	l = &entities.LessonGroup{}
	database.AllNullEntity(l)
	err = multierr.Combine(
		l.LessonGroupID.Set(lessonGroupID),
		l.CourseID.Set(courseID),
	)
	if err != nil {
		return fmt.Errorf("err set LessonGroup: %w", err)
	}

	err = s.LessonGroupRepo.Create(ctx, s.DBTrace, l)
	if err != nil {
		return fmt.Errorf("s.LessonGroupRepo.Create: %w", err)
	}

	return nil
}

func (s *CourseService) createLessonGroupLessonmgmt(ctx context.Context, lessonGroupID, courseID string) error {
	l, err := s.LessonGroupRepo.Get(ctx, s.LessonDBTrace, database.Text(lessonGroupID), database.Text(courseID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("s.LessonGroupRepo.Get: %w", err)
	}

	if l != nil {
		return nil
	}

	l = &entities.LessonGroup{}
	database.AllNullEntity(l)
	err = multierr.Combine(
		l.LessonGroupID.Set(lessonGroupID),
		l.CourseID.Set(courseID),
	)
	if err != nil {
		return fmt.Errorf("err set LessonGroup: %w", err)
	}

	err = s.LessonGroupRepo.Create(ctx, s.LessonDBTrace, l)
	if err != nil {
		return fmt.Errorf("s.LessonGroupRepo.Create: %w", err)
	}

	return nil
}

func canUpdateTopic(topic *entities.Topic, name string, attachmentNames, attachmentURLs []string) (canUpdate bool) {
	if topic.Name.String != name {
		topic.Name.Set(name)
		canUpdate = true
	}

	var names []string
	topic.AttachmentNames.AssignTo(&names)
	if len(names) != len(attachmentNames) || !utils.EqualStringArray(names, attachmentNames) {
		topic.AttachmentNames.Set(attachmentNames)
		canUpdate = true
	}

	var urls []string
	topic.AttachmentURLs.AssignTo(&urls)
	if len(urls) != len(attachmentURLs) || !utils.EqualStringArray(urls, attachmentURLs) {
		topic.AttachmentURLs.Set(attachmentURLs)
		canUpdate = true
	}

	if canUpdate {
		topic.DeletedAt.Set(nil)
	}

	return
}

func canUpdatePresetStudyPlanWeekly(
	pw *entities.PresetStudyPlanWeekly,
	studyPlanID string,
	startDate time.Time,
	endDate time.Time,
) (canUpdate bool) {
	if pw.PresetStudyPlanID.String != studyPlanID {
		pw.PresetStudyPlanID.Set(studyPlanID)
		canUpdate = true
	}

	if !pw.StartDate.Time.Equal(startDate) {
		pw.StartDate.Set(startDate)
		canUpdate = true
	}

	if !pw.EndDate.Time.Equal(endDate) {
		pw.EndDate.Set(endDate)
		canUpdate = true
	}

	return
}

func canUpdateLesson(
	lesson *entities.Lesson,
	teacherID,
	courseID,
	lessonGroup string,
	lessonType cpb.LessonType,
) (canUpdate bool) {
	if lesson.CourseID.String != courseID {
		lesson.CourseID.Set(courseID)
		canUpdate = true
	}

	if lesson.LessonGroupID.String != lessonGroup {
		lesson.LessonGroupID.Set(lessonGroup)
		canUpdate = true
	}

	if teacherID != "" && lesson.TeacherID.String != teacherID {
		lesson.TeacherID.Set(teacherID)
		canUpdate = true
	}

	if lessonType != cpb.LessonType_LESSON_TYPE_NONE {
		lesson.LessonType.Set(lessonType.String())
		if lessonType == cpb.LessonType_LESSON_TYPE_ONLINE {
			lesson.TeachingMedium.Set(entities.LessonTeachingMediumOnline)
		} else if lessonType == cpb.LessonType_LESSON_TYPE_OFFLINE {
			lesson.TeachingMedium.Set(entities.LessonTeachingMediumOffline)
		}
		canUpdate = true
	}

	return
}

func canUpdateLessonNameAndTime(
	lesson *entities.Lesson,
	topic *entities.Topic, name string,
	pw *entities.PresetStudyPlanWeekly, startDate time.Time, endDate time.Time,
) (canUpdate bool) {
	if topic.Name.String != name {
		lesson.Name.Set(name)
		canUpdate = true
	}
	if !pw.StartDate.Time.Equal(startDate) {
		lesson.StartTime.Set(startDate)
		canUpdate = true
	}

	if !pw.EndDate.Time.Equal(endDate) {
		lesson.EndTime.Set(endDate)
		canUpdate = true
	}
	return
}

func updateLessonNameAndTimeLessonmgmt(
	lesson *entities.Lesson,
	name string,
	startDate, endDate time.Time,
) (canUpdate bool) {
	if lesson.Name.String != name {
		lesson.Name.Set(name)
		canUpdate = true
	}
	if !lesson.StartTime.Time.Equal(startDate) {
		lesson.StartTime.Set(startDate)
		canUpdate = true
	}

	if !lesson.EndTime.Time.Equal(endDate) {
		lesson.EndTime.Set(endDate)
		canUpdate = true
	}
	return
}
