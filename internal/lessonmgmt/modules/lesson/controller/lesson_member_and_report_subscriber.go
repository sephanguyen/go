package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/commands"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

type LessonStudentSubscription struct {
	Logger            *zap.Logger
	JSM               nats.JetStreamManagement
	wrapperConnection *support.WrapperDBConnection
	Env               string
	UnleashClientIns  unleashclient.ClientInstance

	LessonMemberRepo     infrastructure.LessonMemberRepo
	LessonRepo           infrastructure.LessonRepo
	LessonReportRepo     infrastructure.LessonReportRepo
	ReallocationRepo     infrastructure.ReallocationRepo
	LessonCommandHandler commands.LessonCommandHandler
}

func (l *LessonStudentSubscription) Subscribe() error {
	l.Logger.Info("[LessonStudentSubscription]: Subscribing to ",
		zap.String("subject", constants.SubjectStudentPackageV2EventNats),
		zap.String("group", constants.QueueStudentSubscriptionLessonMemberEventNats),
		zap.String("durable", constants.DurableStudentSubscriptionLessonMemberEventNats),
	)
	return multierr.Combine(
		l.subscribeLessonMemberAndReport(),
		l.subscribeRemovingInactiveStudentFromLesson(),
	)
}

func (l *LessonStudentSubscription) subscribeRemovingInactiveStudentFromLesson() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStudentPackageEventNatsV2, constants.DurableStudentCourseDurationEventNats),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverStudentCourseDurationEventNats),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "subscribeRemovingInactiveStudentFromLesson",
	}
	_, err := l.JSM.QueueSubscribe(
		constants.SubjectStudentPackageV2EventNats,
		constants.QueueStudentCourseDurationNats,
		opts,
		l.handleRemovingStudentFromLesson,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectStudentPackageV2EventNats,
			constants.QueueStudentCourseDurationNats,
			err,
		)
	}
	return nil
}

func (l *LessonStudentSubscription) handleRemovingStudentFromLesson(ctx context.Context, msg []byte) (bool, error) {
	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	l.Logger.Info("[LessonStudentSubscription]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectStudentPackageV2EventNats),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var studentPackageEvt npb.EventStudentPackageV2
	err = proto.Unmarshal(msg, &studentPackageEvt)
	if err != nil {
		l.Logger.Error("Failed to parse npb.EventStudentPackageV2: ", zap.Error(err))
		return false, fmt.Errorf("Failed to parse npb.EventStudentPackageV2 :%w", err)
	}
	sp := studentPackageEvt.GetStudentPackage()
	if sp.GetIsActive() {
		studentID := sp.StudentId
		lessonOutOfStudentCourse, err := l.LessonMemberRepo.GetLessonsOutOfStudentCourse(ctx, conn, &user_domain.StudentSubscription{
			CourseID:  sp.Package.CourseId,
			StudentID: studentID,
			StartAt:   sp.Package.StartDate.AsTime(),
			EndAt:     sp.Package.EndDate.AsTime(),
		})
		if err != nil {
			return false, fmt.Errorf("l.LessonMemberRepo.GetLessonsOutOfStudentCourse: %w", err)
		}
		if len(lessonOutOfStudentCourse) > 0 {
			l.Logger.Info("Handling remove inactive student from lessons",
				zap.String("user_id", studentID),
				zap.Any("lesson_id", lessonOutOfStudentCourse),
			)
			if err := database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) (err error) {
				err = l.LessonMemberRepo.SoftDelete(ctx, tx, studentID, lessonOutOfStudentCourse)
				if err != nil {
					return fmt.Errorf("l.LessonMemberRepo.SoftDelete: %w", err)
				}
				isUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ReallocateStudents", l.Env)
				if err != nil {
					return fmt.Errorf("l.connectToUnleash: %w", err)
				}
				if isUnleashToggled {
					studentWithLesson := []string{}
					for _, lessonID := range lessonOutOfStudentCourse {
						studentWithLesson = append(studentWithLesson, studentID, lessonID)
					}
					if err = l.ReallocationRepo.SoftDelete(ctx, tx, studentWithLesson, false); err != nil {
						return fmt.Errorf("l.ReallocationRepo.SoftDelete: %w", err)
					}
					if err = l.ReallocationRepo.CancelIfStudentReallocated(ctx, tx, studentWithLesson); err != nil {
						return fmt.Errorf("l.ReallocationRepo.CancelIfStudentReallocated: %w", err)
					}
				}
				isUnleashTeachingTimeToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_CourseTeachingTime", l.Env)
				if err != nil {
					return fmt.Errorf("l.connectToUnleash: %w", err)
				}
				if isUnleashTeachingTimeToggled {
					if err = l.recomputeLessonCourseBreakTime(ctx, tx, lessonOutOfStudentCourse); err != nil {
						return fmt.Errorf("recomputeLessonCourseBreakTime lessons %s error: %w", strings.Join(lessonOutOfStudentCourse, ", "), err)
					}
				}
				err = l.handleDeleteLessonReport(ctx, tx, lessonOutOfStudentCourse)
				if err != nil {
					return fmt.Errorf("handleDeleteLessonReport: %w", err)
				}
				return nil
			}); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (l *LessonStudentSubscription) subscribeLessonMemberAndReport() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStudentPackageEventNatsV2, constants.DurableStudentSubscriptionLessonMemberEventNats),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverStudentSubscriptionLessonMemberEventNats),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "subscribeLessonMemberAndReport",
	}
	_, err := l.JSM.QueueSubscribe(
		constants.SubjectStudentPackageV2EventNats,
		constants.QueueStudentSubscriptionLessonMemberEventNats,
		opts,
		l.handleDeleteLessonMemberAndReport,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectStudentPackageV2EventNats,
			constants.QueueStudentSubscriptionLessonMemberEventNats,
			err,
		)
	}
	return nil
}

func (l *LessonStudentSubscription) handleDeleteLessonMemberAndReport(ctx context.Context, msg []byte) (bool, error) {
	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	l.Logger.Info("[LessonStudentSubscription]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectStudentPackageV2EventNats),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var studentPackageEvt npb.EventStudentPackageV2
	err = proto.Unmarshal(msg, &studentPackageEvt)
	if err != nil {
		l.Logger.Error("Failed to parse npb.EventStudentPackageV2: ", zap.Error(err))
		return false, fmt.Errorf("Failed to parse npb.EventStudentPackageV2 :%w", err)
	}
	sp := studentPackageEvt.GetStudentPackage()
	if sp.GetIsActive() {
		studentID := sp.StudentId
		lessonIDs, err := l.LessonMemberRepo.GetLessonIDsByStudentCourseRemovedLocation(ctx, conn, sp.Package.CourseId, studentID, []string{sp.Package.LocationId})
		if err != nil {
			return false, fmt.Errorf("lsc.LessonMemberRepo.GetLessonIDsByStudentCourseRemovedLocation: %w", err)
		}
		if len(lessonIDs) > 0 {
			l.Logger.Info(fmt.Sprintf("Start delete lesson_members user_id = %s", studentID),
				zap.Any("lessonIDs", lessonIDs))
			if err := database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) (err error) {
				// Delete lesson_members
				err = l.LessonMemberRepo.SoftDelete(ctx, tx, studentID, lessonIDs)
				if err != nil {
					return fmt.Errorf("lsc.LessonMemberRepo.SoftDelete: %w", err)
				}
				err = l.handleDeleteLessonReport(ctx, tx, lessonIDs)
				if err != nil {
					return fmt.Errorf("handleDeleteLessonReport: %w", err)
				}
				isUnleashTeachingTimeToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_CourseTeachingTime", l.Env)
				if err != nil {
					return fmt.Errorf("l.connectToUnleash: %w", err)
				}
				if isUnleashTeachingTimeToggled {
					if err := l.recomputeLessonCourseBreakTime(ctx, tx, lessonIDs); err != nil {
						return fmt.Errorf("recomputeLessonCourseBreakTime lessons %s error: %w", strings.Join(lessonIDs, ", "), err)
					}
				}
				return nil
			}); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (l *LessonStudentSubscription) handleDeleteLessonReport(ctx context.Context, tx pgx.Tx, lessonIDs []string) error {
	lessonMembers, err := l.LessonMemberRepo.GetLessonMembersInLessons(ctx, tx, lessonIDs)
	if err != nil {
		return fmt.Errorf("lsc.LessonMemberRepo.GetLessonMembersInLessons: %w", err)
	}
	IDs := getLessonIDWithoutStudent(lessonMembers, lessonIDs)
	if len(IDs) > 0 {
		err = l.LessonReportRepo.DeleteReportsBelongToLesson(ctx, tx, IDs)
		if err != nil {
			return fmt.Errorf("lsc.LessonRepo.DeleteReportsBelongToLesson: %w", err)
		}
	}
	return nil
}

func getLessonIDWithoutStudent(lessonMembers []*domain.LessonMember, lessonIDs []string) (ids []string) {
	remainLessonIDs := make([]string, 0, len(lessonMembers))
	for _, lesson := range lessonMembers {
		remainLessonIDs = append(remainLessonIDs, lesson.LessonID)
	}
	for _, lessonID := range lessonIDs {
		if !slices.Contains(remainLessonIDs, lessonID) {
			ids = append(ids, lessonID)
		}
	}
	return ids
}

func (l *LessonStudentSubscription) recomputeLessonCourseBreakTime(ctx context.Context, tx pgx.Tx, lessonIDs []string) error {
	lessons, err := l.LessonRepo.GetLessonByIDs(ctx, tx, lessonIDs)
	if err != nil {
		return fmt.Errorf("l.LessonRepo.GetLessonByIDs: %w", err)
	}
	if err = l.LessonCommandHandler.AddLessonCourseTeachingTime(ctx, tx, lessons, true, "UTC"); err != nil {
		return fmt.Errorf("lesson.AddLessonCourseTeachingTime: %w", err)
	}
	if err = l.LessonRepo.UpdateLessonsTeachingTime(ctx, tx, lessons); err != nil {
		return fmt.Errorf("lessonRepo.UpdateLessonsTeachingTime: %w", err)
	}
	return nil
}
