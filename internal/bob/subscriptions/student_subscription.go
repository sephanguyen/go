package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/support"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type StudentSubscription struct {
	Logger            *zap.Logger
	WrapperConnection *support.WrapperDBConnection
	JSM               nats.JetStreamManagement
	Env               string
	UnleashClientIns  unleashclient.ClientInstance

	StudentSubscriptionRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentSubscriptionItems []*entities.StudentSubscription) error
		DeleteByCourseIDAndStudentID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, studentID pgtype.Text) error
		RetrieveStudentSubscriptionID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text, studentID pgtype.Text, subscriptionID pgtype.Text) (string, error)
	}
	StudentSubscriptionAccessPathRepo interface {
		Upsert(ctx context.Context, db database.Ext, ssAccessPaths []*entities.StudentSubscriptionAccessPath) error
		DeleteByStudentSubscriptionID(ctx context.Context, db database.QueryExecer, studentSubscriptionID pgtype.Text) error
	}
	CourseLocationScheduleRepo interface {
		GetByCourseIDAndLocationID(ctx context.Context, db database.QueryExecer, courseID, locationID string) (*domain.CourseLocationSchedule, error)
	}
	StudentCourseRepo interface {
		GetByStudentCourseID(ctx context.Context, db database.QueryExecer, studentID, courseID, locationID, studentPackageID string) (*user_domain.StudentCourse, error)
	}
	LessonAllocationRepo interface {
		CountPurchasedSlotPerStudentSubscription(ctx context.Context, db database.QueryExecer, freq uint8, startTime, endTime time.Time, courseID, locationID, studentID string) (uint32, error)
	}
}

func (ss *StudentSubscription) Subscribe() error {
	ss.Logger.Info("StudentPackageEvent: subscribing to",
		zap.String("subject", constants.SubjectStudentPackageEventNats),
		zap.String("group", constants.QueueSyncStudentSubscriptionEventNats),
		zap.String("durable", constants.DurableSyncStudentSubscriptionEventNats),
	)
	dayAgo := time.Hour * 24
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStudentPackageEventNats, constants.DurableSyncStudentSubscriptionEventNats),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverStudentSubscriptionEventNats),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
		SkipMsgOlderThan: &dayAgo,
	}

	_, err := ss.JSM.QueueSubscribe(constants.SubjectStudentPackageEventNats,
		constants.QueueSyncStudentSubscriptionEventNats,
		option, ss.handleStudentSubscriptionJobEvent)
	if err != nil {
		return fmt.Errorf("studentSubscription.Subscribe: %w", err)
	}

	return nil
}

func (ss *StudentSubscription) handleStudentSubscriptionJobEvent(ctx context.Context, data []byte) (bool, error) {
	conn, err := ss.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	// set timeout for syncing job
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var studentPackages npb.EventStudentPackage
	if err := proto.Unmarshal(data, &studentPackages); err != nil {
		return false, fmt.Errorf("handleStudentSubscriptionJobEvent proto.Unmarshal: %w", err)
	}

	err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		err = ss.handle(ctx, &studentPackages, tx)
		return err
	})
	if err != nil {
		return true, fmt.Errorf("ss.handleStudentSubscriptionJobEvent: %w", err)
	}

	return false, nil
}

func (ss *StudentSubscription) handle(ctx context.Context, data *npb.EventStudentPackage, tx pgx.Tx) error {
	// contains upsert items from `fatima.student_package`
	var studentSubscriptions []*entities.StudentSubscription
	var studentSubscriptionAccessPaths []*entities.StudentSubscriptionAccessPath
	var err error
	now := time.Now()
	if data == nil || data.StudentPackage == nil || data.StudentPackage.Package == nil {
		return nil
	}

	for i := 0; i < len(data.StudentPackage.Package.CourseIds); i++ {
		// init vars
		studentSubscription := new(entities.StudentSubscription)
		studentSubscriptionAccessPath := new(entities.StudentSubscriptionAccessPath)
		var studentSubscriptionID string
		// retrieve vars from Fatima
		courseID := data.StudentPackage.Package.CourseIds[i]
		studentID := data.StudentPackage.StudentId
		subscriptionID := data.StudentPackage.Package.StudentPackageId
		// retrieve student_subscription_id with course_id, student_id, subscription_id
		// if exists, we assign it with the returned value instead of generate a new one
		studentSubscriptionID, err = ss.StudentSubscriptionRepo.RetrieveStudentSubscriptionID(ctx, tx, database.Text(courseID), database.Text(studentID), database.Text(subscriptionID))
		if err != nil {
			return err
		}
		if len(studentSubscriptionID) == 0 {
			studentSubscriptionID = idutil.ULIDNow()
		}
		// handle lesson_student_subscription
		err = multierr.Combine(
			studentSubscription.StudentSubscriptionID.Set(studentSubscriptionID),
			studentSubscription.CourseID.Set(data.StudentPackage.Package.CourseIds[i]),
			studentSubscription.StudentID.Set(data.StudentPackage.StudentId),
			studentSubscription.StartAt.Set(data.StudentPackage.Package.StartDate.AsTime()),
			studentSubscription.EndAt.Set(data.StudentPackage.Package.EndDate.AsTime()),
			studentSubscription.CreatedAt.Set(now),
			studentSubscription.UpdatedAt.Set(now),
			studentSubscription.SubscriptionID.Set(data.StudentPackage.Package.StudentPackageId),
			studentSubscription.PurchasedSlotTotal.Set(nil),
		)
		if data.StudentPackage.IsActive {
			err = multierr.Append(err, studentSubscription.DeletedAt.Set(nil))
			isUnleashToggled, err := ss.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_OptimizeLessonAllocation", ss.Env)
			if err != nil {
				return fmt.Errorf("l.connectToUnleash: %w", err)
			}
			if isUnleashToggled {
				// calculate total number of lessons
				// package_type = PACKAGE_TYPE_ONE_TIME => total = course_location_schedule.total_no_lessons
				// package_type = PACKAGE_TYPE_SCHEDULED => total = course_location_schedule.frequency * number of weeks
				// package_type = PACKAGE_TYPE_SLOT_BASED => total = student_course.course_slot
				// package_type = PACKAGE_TYPE_FREQUENCY => total = student_course.course_slot_per_week * number of weeks
				var locationID string
				if len(data.StudentPackage.Package.LocationIds) > 0 {
					locationID = data.StudentPackage.Package.LocationIds[0]
				}
				cls, err := ss.CourseLocationScheduleRepo.GetByCourseIDAndLocationID(ctx, tx, courseID, locationID)
				if err != nil && err != domain.ErrorNotFound {
					return fmt.Errorf("ss.CourseLocationScheduleRepo.GetByCourseIDAndLocationID: %v", err)
				}
				if cls != nil {
					var courseSlot, courseSlotPerWeek int
					if cls.IsDefinedByOrder() {
						studentCourse, err := ss.StudentCourseRepo.GetByStudentCourseID(ctx, tx, studentID, courseID, locationID, data.StudentPackage.Package.StudentPackageId)
						if err != nil && err != user_domain.ErrorNotFound {
							return fmt.Errorf("ss.StudentCourseRepo.GetByStudentCourseID: %v", err)
						}
						if studentCourse != nil {
							courseSlot = int(studentCourse.CourseSlot)
							courseSlotPerWeek = int(studentCourse.CourseSlotPerWeek)
						}
					}
					var purchasedSlotTotal uint32

					if cls.IsScheduleWeekly() {
						var freq uint8
						if cls.IsSchedule() {
							freq = uint8(*cls.Frequency)
						} else if cls.IsFrequency() {
							freq = uint8(courseSlotPerWeek)
						}
						purchasedSlotTotal, err = ss.LessonAllocationRepo.CountPurchasedSlotPerStudentSubscription(
							ctx, tx, freq,
							data.StudentPackage.Package.StartDate.AsTime(),
							data.StudentPackage.Package.EndDate.AsTime(),
							courseID, locationID, studentID)
						if err != nil {
							return fmt.Errorf("ss.LessonAllocationRepo.CountPurchasedSlotPerStudentSubscription: %v", err)
						}
					}

					if !cls.IsScheduleWeekly() {
						if cls.IsOneTime() {
							purchasedSlotTotal = uint32(*cls.TotalNoLesson)
						} else if cls.IsSlotBased() {
							purchasedSlotTotal = uint32(courseSlot)
						}
					}
					if err = multierr.Append(err, studentSubscription.PurchasedSlotTotal.Set(purchasedSlotTotal)); err != nil {
						return err
					}
				}
			}
		} else {
			err = multierr.Append(err, studentSubscription.DeletedAt.Set(now))
		}
		if err != nil {
			return err
		}
		studentSubscriptions = append(studentSubscriptions, studentSubscription)
		// handle lesson_student_subscription_access_path
		if len(data.StudentPackage.Package.LocationIds) > 0 {
			for j := 0; j < len(data.StudentPackage.Package.LocationIds); j++ {
				if err := multierr.Combine(
					studentSubscriptionAccessPath.StudentSubscriptionID.Set(studentSubscriptionID),
					studentSubscriptionAccessPath.LocationID.Set(data.StudentPackage.Package.LocationIds[j]),
					studentSubscriptionAccessPath.CreatedAt.Set(now),
					studentSubscriptionAccessPath.UpdatedAt.Set(now),
					studentSubscriptionAccessPath.DeletedAt.Set(nil),
				); err != nil {
					return err
				}
				studentSubscriptionAccessPaths = append(studentSubscriptionAccessPaths, studentSubscriptionAccessPath)
			}
		}
		// delete duplicates for student_subscriptions
		if err := ss.StudentSubscriptionRepo.DeleteByCourseIDAndStudentID(ctx, tx, studentSubscription.CourseID, studentSubscription.StudentID); err != nil {
			return fmt.Errorf("ss.StudentSubscriptionRepo.DeleteByCourseIDAndStudentID: %v", err)
		}
		// delete duplicates for student_subscription_access_path
		if err := ss.StudentSubscriptionAccessPathRepo.DeleteByStudentSubscriptionID(ctx, tx, database.Text(studentSubscription.StudentSubscriptionID.String)); err != nil {
			return fmt.Errorf("ss.StudentSubscriptionAccessPathRepo.DeleteByStudentSubscriptionID: %v", err)
		}
		// bulkInsert array to table `student_subscriptions`
		if err := ss.StudentSubscriptionRepo.BulkUpsert(ctx, tx, studentSubscriptions); err != nil {
			return fmt.Errorf("ss.StudentSubscriptionRepo.BulkUpsert: %v", err)
		}
		// bulkInsert array to table `student_subscription_access_path`
		if len(data.StudentPackage.Package.LocationIds) > 0 {
			ss.Logger.Info("ss.StudentSubscriptionRepo.BulkUpsert",
				zap.String("schoolID", golibs.ResourcePathFromCtx(ctx)))
			if err := ss.StudentSubscriptionAccessPathRepo.Upsert(ctx, tx, studentSubscriptionAccessPaths); err != nil {
				return fmt.Errorf("ss.StudentSubscriptionAccessPathRepo.Upsert: %v", err)
			}
		}
	}
	return err
}
