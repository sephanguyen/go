package consumers

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	masterMgmt "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ScheduleClassHandler struct {
	Logger            *zap.Logger
	BobDB             database.Ext
	WrapperConnection *support.WrapperDBConnection
	JSM               nats.JetStreamManagement
	LessonRepo        infrastructure.LessonRepo
	LessonMemberRepo  infrastructure.LessonMemberRepo
	ClassMemberRepo   masterMgmt.ClassMemberRepo
	LessonReportRepo  infrastructure.LessonReportRepo
}

func (s *ScheduleClassHandler) Handle(ctx context.Context, msg []byte) (bool, error) {
	s.Logger.Info("[ScheduleClassHandler]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectMasterMgmtReserveClassUpserted),
	)

	scheduleClassEvent := &mpb.EvtScheduleClass{}

	if err := proto.Unmarshal(msg, scheduleClassEvent); err != nil {
		return false, err
	}

	switch scheduleClassEvent.Message.(type) {
	case *mpb.EvtScheduleClass_ScheduleClass_:
		return s.handleScheduleClassEvent(ctx, scheduleClassEvent)
	case *mpb.EvtScheduleClass_CancelScheduledClass_:
		return s.cancelScheduledClassEvent(ctx, scheduleClassEvent)
	default:
		s.Logger.Info(fmt.Sprintf("[ScheduleClassHandler]: schedule class event type not supported %T", scheduleClassEvent.Message))
	}

	return true, nil
}

func (s *ScheduleClassHandler) handleScheduleClassEvent(ctx context.Context, eventData *mpb.EvtScheduleClass) (bool, error) {
	conn, err := s.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}

	scheduleClassEvent := eventData.GetScheduleClass()
	scheduleClassID := scheduleClassEvent.GetScheduleClassId()
	studentID := scheduleClassEvent.GetUserId()
	effectiveDate := scheduleClassEvent.GetEffectiveDate().AsTime()
	currentClassID := scheduleClassEvent.GetCurrentClassId()
	oldScheduledClassID := scheduleClassEvent.GetOldScheduledClassId()
	oldEffectiveDate := scheduleClassEvent.GetOldScheduledEffectiveDate().AsTime()

	now := time.Now()

	mapClassMember, err := s.ClassMemberRepo.GetByClassIDAndUserIDs(ctx, s.BobDB, currentClassID, []string{studentID})
	if err != nil {
		// retry to get class members
		return true, fmt.Errorf("GetByClassIDAndUserIDs fail: %w", err)
	}
	classMember := mapClassMember[studentID]
	if classMember == nil {
		return false, fmt.Errorf("not found student %s in class %s", studentID, currentClassID)
	}

	queryLessonCurrentClass := &domain.QueryLesson{ClassID: currentClassID, StartTime: &effectiveDate, EndTime: &classMember.EndDate}
	queryLessonAddStudent := &domain.QueryLesson{ClassID: scheduleClassID, StartTime: &effectiveDate, EndTime: &classMember.EndDate}
	queryLessonOldSchedule := &domain.QueryLesson{ClassID: oldScheduledClassID, StartTime: &oldEffectiveDate, EndTime: &classMember.EndDate}

	lessonsCurrentClass, err := s.LessonRepo.GetLessonsTeachingModelGroupByClassIdWithDuration(ctx, conn, queryLessonCurrentClass)

	if err != nil {
		return false, fmt.Errorf("get lessons of student %s on current class %s: %w", studentID, currentClassID, err)
	}

	lessonsOldSchedule, err := s.LessonRepo.GetLessonsTeachingModelGroupByClassIdWithDuration(ctx, conn, queryLessonOldSchedule)

	if err != nil {
		return false, fmt.Errorf("get lessons of student %s on old scheduled class %s: %w", studentID, oldScheduledClassID, err)
	}

	lessonsAddStudent, err := s.LessonRepo.GetLessonsTeachingModelGroupByClassIdWithDuration(ctx, conn, queryLessonAddStudent)

	if err != nil {
		return false, fmt.Errorf("get lessons of student %s on schedule class %s: %w", studentID, scheduleClassID, err)
	}

	lessonMembersAdded := sliceutils.Map(lessonsAddStudent, func(l *domain.Lesson) *domain.LessonMember {
		return &domain.LessonMember{LessonID: l.LessonID, StudentID: studentID, CourseID: l.CourseID, UpdatedAt: now}
	})

	lessonMembersRemoved := sliceutils.Map(lessonsCurrentClass, func(l *domain.Lesson) *domain.LessonMember {
		return &domain.LessonMember{LessonID: l.LessonID, StudentID: studentID, CourseID: l.CourseID, UpdatedAt: now}
	})

	lessonMembersRemoved = append(lessonMembersRemoved, sliceutils.Map(lessonsOldSchedule, func(l *domain.Lesson) *domain.LessonMember {
		return &domain.LessonMember{LessonID: l.LessonID, StudentID: studentID, CourseID: l.CourseID, UpdatedAt: now}
	})...)

	err = s.upsertLessonMembers(ctx, conn, lessonMembersAdded, lessonMembersRemoved)

	// retry upsert lesson members
	if err != nil {
		return true, fmt.Errorf("upsert lesson members fail: %w", err)
	}

	return true, nil
}

func (s *ScheduleClassHandler) cancelScheduledClassEvent(ctx context.Context, eventData *mpb.EvtScheduleClass) (bool, error) {
	conn, err := s.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}

	cancelScheduledClassEvent := eventData.GetCancelScheduledClass()
	scheduledClassID := cancelScheduledClassEvent.GetScheduledClassId()
	studentID := cancelScheduledClassEvent.GetUserId()
	effectiveDate := cancelScheduledClassEvent.GetEffectiveDate().AsTime()
	currentClassID := cancelScheduledClassEvent.GetCurrentClassId()

	now := time.Now()

	mapClassMember, err := s.ClassMemberRepo.GetByClassIDAndUserIDs(ctx, s.BobDB, currentClassID, []string{studentID})
	if err != nil {
		// retry to get class members
		return true, fmt.Errorf("GetByClassIDAndUserIDs fail: %w", err)
	}
	classMember := mapClassMember[studentID]
	if classMember == nil {
		return false, fmt.Errorf("not found student %s in class %s", studentID, currentClassID)
	}

	queryLessonAddStudent := &domain.QueryLesson{ClassID: currentClassID, StartTime: &effectiveDate, EndTime: &classMember.EndDate}
	queryLessonRemoveStudent := &domain.QueryLesson{ClassID: scheduledClassID, StartTime: &effectiveDate, EndTime: &classMember.EndDate}

	lessonsAddStudent, err := s.LessonRepo.GetLessonsTeachingModelGroupByClassIdWithDuration(ctx, conn, queryLessonAddStudent)

	if err != nil {
		return false, fmt.Errorf("get lessons of student %s on current active class %s: %w", studentID, currentClassID, err)
	}

	lessonsRemoveStudent, err := s.LessonRepo.GetLessonsTeachingModelGroupByClassIdWithDuration(ctx, conn, queryLessonRemoveStudent)

	if err != nil {
		return false, fmt.Errorf("get lessons of student %s on scheduled class %s: %w", studentID, scheduledClassID, err)
	}

	lessonMembersAdded := sliceutils.Map(lessonsAddStudent, func(l *domain.Lesson) *domain.LessonMember {
		return &domain.LessonMember{LessonID: l.LessonID, StudentID: studentID, CourseID: l.CourseID, UpdatedAt: now}
	})

	lessonMembersRemoved := sliceutils.Map(lessonsRemoveStudent, func(l *domain.Lesson) *domain.LessonMember {
		return &domain.LessonMember{LessonID: l.LessonID, StudentID: studentID, CourseID: l.CourseID, UpdatedAt: now}
	})

	err = s.upsertLessonMembers(ctx, conn, lessonMembersAdded, lessonMembersRemoved)

	// retry upsert lesson members
	if err != nil {
		return true, fmt.Errorf("upsert lesson members fail: %w", err)
	}

	return true, nil
}

func (s *ScheduleClassHandler) upsertLessonMembers(ctx context.Context, db database.Ext, lessonMembersAdded, lessonMembersRemoved []*domain.LessonMember) error {
	lessonIdsCleanReport := sliceutils.Map(lessonMembersRemoved, func(lm *domain.LessonMember) string {
		return lm.LessonID
	})
	return database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		err := s.LessonMemberRepo.DeleteLessonMembers(ctx, tx, lessonMembersRemoved)

		if err != nil {
			return fmt.Errorf("remove lesson members fail: %w", err)
		}

		if len(lessonIdsCleanReport) > 0 {
			if err := s.LessonReportRepo.DeleteLessonReportWithoutStudent(ctx, tx, lessonIdsCleanReport); err != nil {
				return fmt.Errorf("clean lesson reports fail: %w", err)
			}
		}

		if len(lessonMembersAdded) > 0 {
			if err := s.LessonMemberRepo.InsertLessonMembers(ctx, tx, lessonMembersAdded); err != nil {
				return fmt.Errorf("insert lesson members fail: %w", err)
			}
		}

		return nil
	})
}
