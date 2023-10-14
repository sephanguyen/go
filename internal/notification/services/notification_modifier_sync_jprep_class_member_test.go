package services

import (
	"context"
	"strconv"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
)

func Test_SyncJprepClassMember(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	notificationClassMemberRepo := &mock_repositories.MockNotificationClassMemberRepo{}
	classRepo := &mock_repositories.MockClassRepo{}

	svc := &NotificationModifierService{
		DB:                          mockDB,
		NotificationClassMemberRepo: notificationClassMemberRepo,
		ClassRepo:                   classRepo,
	}

	ctx := context.Background()

	t.Run("happy case join class", func(t *testing.T) {
		req := &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_JoinClass_{
				JoinClass: &pb.EvtClassRoom_JoinClass{
					ClassId:   1,
					UserId:    "student-id",
					UserGroup: pb.USER_GROUP_STUDENT,
				},
			},
		}

		msg := req.GetJoinClass()
		strClassID := strconv.Itoa(int(msg.ClassId))
		courseID := "course-id"

		classRepo.On("FindCourseIDByClassID", ctx, mockDB, strClassID).Once().Return(courseID, nil)

		notificationClassMember := &entities.NotificationClassMember{}
		database.AllNullEntity(notificationClassMember)

		_ = multierr.Combine(
			notificationClassMember.StudentID.Set(msg.UserId),
			notificationClassMember.ClassID.Set(strClassID),
			notificationClassMember.CourseID.Set(courseID),
			notificationClassMember.LocationID.Set(constants.JPREPOrgLocation),
			notificationClassMember.DeletedAt.Set(nil),
		)

		notificationClassMemberRepo.On("Upsert", ctx, mockDB, notificationClassMember).Once().Return(nil)

		err := svc.SyncJprepClassMember(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("happy case leave class", func(t *testing.T) {
		req := &pb.EvtClassRoom{
			Message: &pb.EvtClassRoom_LeaveClass_{
				LeaveClass: &pb.EvtClassRoom_LeaveClass{
					ClassId: 1,
					UserIds: []string{"student-id-1", "student-id-2"},
				},
			},
		}
		msg := req.GetLeaveClass()
		strClassID := strconv.Itoa(int(msg.ClassId))
		courseID := "course-id"

		classRepo.On("FindCourseIDByClassID", ctx, mockDB, strClassID).Once().Return(courseID, nil)

		notificationClassMemberRepo.On("BulkUpsert", ctx, mockDB, mock.Anything).Once().Return(nil)
		err := svc.SyncJprepClassMember(ctx, req)
		assert.NoError(t, err)
	})
}
