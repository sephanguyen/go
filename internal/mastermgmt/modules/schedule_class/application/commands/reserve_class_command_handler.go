package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	classInfra "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReserveClassCommandHandler struct {
	DB               database.Ext
	ReserveClassRepo infrastructure.ReserveClassRepo
	JSM              nats.JetStreamManagement
	ClassMemberRepo  classInfra.ClassMemberRepo
}

func (rcc *ReserveClassCommandHandler) UpsertReserveClass(ctx context.Context, reserveClass *domain.ReserveClass) (string, *timestamppb.Timestamp, error) {
	var oldClassID string
	var oldEffectiveDate *timestamppb.Timestamp
	err := database.ExecInTx(ctx, rcc.DB, func(ctx context.Context, tx pgx.Tx) error {
		classID, effectiveDate, err1 := rcc.ReserveClassRepo.DeleteOldReserveClass(ctx, tx, reserveClass.StudentPackageID, reserveClass.StudentID, reserveClass.CourseID)
		err2 := rcc.ReserveClassRepo.InsertOne(ctx, tx, reserveClass)
		combineErr := multierr.Combine(err1, err2)
		oldClassID = classID.String
		oldEffectiveDate = &timestamppb.Timestamp{Seconds: effectiveDate.Time.Unix()}
		return combineErr
	})

	if err != nil {
		return oldClassID, oldEffectiveDate, fmt.Errorf("UpsertReserveClass: %w", err)
	}
	return oldClassID, oldEffectiveDate, nil
}

func (rcc *ReserveClassCommandHandler) PublicReserveClassEvt(ctx context.Context, msg *mpb.EvtScheduleClass) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgID, err := rcc.JSM.PublishAsyncContext(ctx, constants.SubjectMasterMgmtReserveClassUpserted, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublicReserveClassEvt JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return nil
}

func (rcc *ReserveClassCommandHandler) CheckWillReserveClass(ctx context.Context, req *mpb.ScheduleStudentClassRequest) (bool, string, error) {
	currentDate := getBeginOfDate(time.Now())
	courseStartDate := getBeginOfDateOnJPZone(req.StartTime)
	courseEndDate := getBeginOfDateOnJPZone(req.EndTime)
	effectiveDate := getBeginOfDateOnJPZone(req.EffectiveDate)
	studentID := req.StudentId
	courseID := req.CourseId

	// effectiveDate must be today or future day and earlier than course end date
	if effectiveDate.Before(currentDate) || effectiveDate.After(courseEndDate) {
		return false, "", fmt.Errorf("invalid effective date")
	}

	// will register class if course has not started or effective date is today
	if courseStartDate.After(currentDate) || effectiveDate.Equal(currentDate) {
		return false, "", nil
	}

	classMembers, err := rcc.ClassMemberRepo.GetByUserAndCourse(ctx, rcc.DB, studentID, courseID)
	if err != nil {
		return false, "", fmt.Errorf("query class members fail: %w", err)
	}

	if cm, ok := classMembers[studentID]; ok {
		classStartDate := getBeginOfDate(cm.StartDate)
		classEndDate := getBeginOfDate(cm.EndDate)

		// class start date is past day or today and class end date is future day or today then class is active
		// will reserve class when class is active
		if !classStartDate.After(currentDate) && !classEndDate.Before(currentDate) {
			return true, cm.ClassID, nil
		}
	}

	return false, "", nil
}

func (rcc *ReserveClassCommandHandler) ReserveStudentClass(ctx context.Context, req *mpb.ScheduleStudentClassRequest, currentClassID string) error {
	effectiveDate := getBeginOfDateOnJPZone(req.EffectiveDate)

	now := time.Now()
	reserveClassBuilder := domain.NewReserveClassBuilder().
		WithStudentID(req.StudentId).
		WithStudentPackageID(req.StudentPackageId).
		WithCourseID(req.CourseId).
		WithClassID(req.ClassId).
		WithEffectiveDate(effectiveDate).
		WithReserveClassRepo(rcc.ReserveClassRepo).
		WithModificationTime(now, now)

	reserveClass, err := reserveClassBuilder.Build()

	if err != nil {
		return fmt.Errorf("build reserve class err: %w", err)
	}

	oldClassID, oldEffectiveDate, err := rcc.UpsertReserveClass(ctx, reserveClass)

	if err != nil {
		return err
	}

	err = rcc.PublicReserveClassEvt(ctx, &mpb.EvtScheduleClass{
		Message: &mpb.EvtScheduleClass_ScheduleClass_{
			ScheduleClass: &mpb.EvtScheduleClass_ScheduleClass{
				ScheduleClassId:           req.ClassId,
				UserId:                    req.StudentId,
				CurrentClassId:            currentClassID,
				EffectiveDate:             &timestamppb.Timestamp{Seconds: effectiveDate.Unix()},
				OldScheduledClassId:       oldClassID,
				OldScheduledEffectiveDate: oldEffectiveDate,
			},
		},
	})

	if err != nil {
		return err
	}

	return err
}

func (rcc *ReserveClassCommandHandler) CancelReserveClass(ctx context.Context, payload CancelReserveClassCommandPayload) error {
	studentPackageID := payload.StudentPackageID
	studentID := payload.StudentID
	courseID := payload.CourseID
	oldScheduledClassID, effectiveDate, err := rcc.ReserveClassRepo.DeleteOldReserveClass(ctx, rcc.DB, studentPackageID, studentID, courseID)
	if err != nil {
		return fmt.Errorf("DeleteOldReserveClass: %w", err)
	}

	if oldScheduledClassID.String == "" {
		return nil
	}

	classMembers, err := rcc.ClassMemberRepo.GetByUserAndCourse(ctx, rcc.DB, studentID, courseID)
	if err != nil {
		return fmt.Errorf("query class members by student %s and course %s fail: %w", studentID, courseID, err)
	}

	var currentActiveClassID string

	if cm, ok := classMembers[studentID]; ok {
		currentActiveClassID = cm.ClassID
	}

	err = rcc.PublicReserveClassEvt(ctx, &mpb.EvtScheduleClass{
		Message: &mpb.EvtScheduleClass_CancelScheduledClass_{
			CancelScheduledClass: &mpb.EvtScheduleClass_CancelScheduledClass{
				UserId:           payload.StudentID,
				ScheduledClassId: oldScheduledClassID.String,
				EffectiveDate:    &timestamppb.Timestamp{Seconds: effectiveDate.Time.Unix()},
				CurrentClassId:   currentActiveClassID,
			},
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (rcc *ReserveClassCommandHandler) DeleteReserveClassesByEffectiveDate(ctx context.Context, date string) error {
	err := rcc.ReserveClassRepo.DeleteByEffectiveDate(ctx, rcc.DB, date)

	if err != nil {
		return fmt.Errorf("delete reserve classes by effective fail: %w", err)
	}

	return nil
}

func getBeginOfDate(t time.Time) time.Time {
	y, m, d := t.In(time.UTC).Date()
	beginningOfDate := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return beginningOfDate
}

func getBeginOfDateOnJPZone(effectiveDate *timestamppb.Timestamp) time.Time {
	// FE send effective date is 00:00 of day on user's timezone => so it's will be 17:00 or 15:00 of yesterday (for VN and JP time zone)
	// add more 9 hour to change effective date to correct date
	t := effectiveDate.AsTime().Add(time.Hour * 9)
	// get begin of date
	return getBeginOfDate(t)
}
