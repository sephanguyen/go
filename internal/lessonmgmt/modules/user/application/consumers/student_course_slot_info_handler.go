package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	ppb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

type SubscriberHandler interface {
	Handle(ctx context.Context, msg []byte) (bool, error)
}

type StudentCourseSlotInfoHandler struct {
	Logger *zap.Logger
	DB     database.Ext
	JSM    nats.JetStreamManagement

	UserRepo                          infrastructure.UserRepo
	StudentSubscriptionRepo           infrastructure.StudentSubscriptionRepo
	StudentSubscriptionAccessPathRepo infrastructure.StudentSubscriptionAccessPathRepo
}

func (s *StudentCourseSlotInfoHandler) Handle(ctx context.Context, data []byte) (bool, error) {
	s.Logger.Info("[StudentCourseSlotInfoHandler]: Received message on",
		zap.String("data", string(data)),
		zap.String("subject", constants.SubjectStudentCourseEventSync),
	)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var syncStudentCourses []*ppb.EventSyncStudentPackageCourse
	if err := json.Unmarshal(data, &syncStudentCourses); err != nil {
		errorString := fmt.Sprintf("[%s] Failed to parse ppb.EventSyncStudentPackageCourse: ", constants.DurableLessonSyncStudentCourseSlotInfo)
		s.Logger.Error(errorString, zap.Error(err))
		return false, fmt.Errorf(errorString, "%w", err)
	}

	studentSubInfoList := make(domain.StudentSubscriptions, 0, len(syncStudentCourses))
	studentSubAccessPathList := make(domain.StudentSubscriptionAccessPaths, 0, len(syncStudentCourses))

	for _, studentCourseInfo := range syncStudentCourses {
		studentSubInfo := &domain.StudentSubscription{
			SubscriptionID:    studentCourseInfo.StudentPackageId,
			StudentID:         studentCourseInfo.StudentId,
			CourseID:          studentCourseInfo.CourseId,
			StartAt:           studentCourseInfo.StudentStartDate.AsTime(),
			EndAt:             studentCourseInfo.StudentEndDate.AsTime(),
			PackageType:       studentCourseInfo.PackageType.String(),
			CourseSlot:        studentCourseInfo.CourseSlot.GetValue(),
			CourseSlotPerWeek: studentCourseInfo.CourseSlotPerWeek.GetValue(),
		}

		if err := studentSubInfo.IsValid(); err != nil {
			errorString := fmt.Sprintf("[%s] Student subscription is not valid: ", constants.DurableLessonSyncStudentCourseSlotInfo)
			s.Logger.Error(errorString, zap.Error(err))
			return false, fmt.Errorf(errorString, "%w", err)
		}

		studentSubID, err := s.StudentSubscriptionRepo.GetStudentSubscriptionIDByUniqueIDs(ctx, s.DB, studentSubInfo.SubscriptionID, studentSubInfo.StudentID, studentSubInfo.CourseID)
		if err != nil {
			errorString := fmt.Sprintf("[%s] Failed to fetch student subscription id of package id %s: ", constants.DurableLessonSyncStudentCourseSlotInfo, studentSubInfo.SubscriptionID)
			s.Logger.Error(errorString, zap.Error(err))
			return false, fmt.Errorf(errorString, "%w", err)
		}

		if len(studentSubID) > 0 {
			studentSubInfo.StudentSubscriptionID = studentSubID
		} else {
			newID := idutil.ULIDNow()
			studentSubInfo.StudentSubscriptionID = newID
		}

		studentSubAccessPath := &domain.StudentSubscriptionAccessPath{
			SubscriptionID: studentSubInfo.StudentSubscriptionID,
			LocationID:     studentCourseInfo.LocationId,
		}

		if err := studentSubAccessPath.IsValid(); err != nil {
			errorString := fmt.Sprintf("[%s] Student subscription access path is not valid: ", constants.DurableLessonSyncStudentCourseSlotInfo)
			s.Logger.Error(errorString, zap.Error(err))
			return false, fmt.Errorf(errorString, "%w", err)
		}

		userInfo, err := s.UserRepo.GetUserByUserID(ctx, s.DB, studentSubInfo.StudentID)
		if err != nil {
			errorString := fmt.Sprintf("[%s] Failed to fetch user info of user id %s: ", constants.DurableLessonSyncStudentCourseSlotInfo, studentSubInfo.StudentID)
			s.Logger.Error(errorString, zap.Error(err))
			return false, fmt.Errorf(errorString, "%w", err)
		}
		studentSubInfo.StudentFirstName = userInfo.FirstName
		studentSubInfo.StudentLastName = userInfo.LastName

		studentSubInfoList = append(studentSubInfoList, studentSubInfo)
		studentSubAccessPathList = append(studentSubAccessPathList, studentSubAccessPath)
	}

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.StudentSubscriptionRepo.BulkUpsertStudentSubscription(ctx, tx, studentSubInfoList); err != nil {
			return err
		}

		if err := s.StudentSubscriptionAccessPathRepo.DeleteByStudentSubscriptionIDs(ctx, tx, studentSubAccessPathList.GetSubscriptionIDs()); err != nil {
			return err
		}

		if err := s.StudentSubscriptionAccessPathRepo.BulkUpsertStudentSubscriptionAccessPath(ctx, tx, studentSubAccessPathList); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		errorString := fmt.Sprintf("[%s] Failed to handle message in DB: ", constants.DurableLessonSyncStudentCourseSlotInfo)
		s.Logger.Error(errorString, zap.Error(err))
		return false, fmt.Errorf(errorString, "%w", err)
	}

	return true, nil
}
