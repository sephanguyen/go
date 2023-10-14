package domain

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDataRetention_AssignRetentionName(t *testing.T) {
	mockDB := testutil.NewMockDB()
	userRepo := &mock_repositories.MockUserRepo{}

	svc := &DataRetentionService{
		UserRepo: userRepo,
	}

	users := []*entities.User{
		{
			UserID: database.Text("user-id-1"),
			Name:   database.Text("user-name-1"),
		},
		{
			UserID: database.Text("user-id-2"),
			Name:   database.Text("user-name-2"),
		},
	}

	mapUserIDAndUser := map[string]*entities.User{
		"user-id-1": users[0],
		"user-id-2": users[1],
	}

	userStudents := []*entities.User{
		{
			UserID: database.Text("student-id-2"),
			Name:   database.Text("student-name-2"),
		},
	}

	mapStudentIDAndUser := map[string]*entities.User{
		"student-id-2": userStudents[0],
	}

	userNotifiationsReq := []*entities.UserInfoNotification{
		{
			UserID:    database.Text("user-id-1"),
			UserGroup: database.Text("USER_GROUP_STUDENT"),
			ParentID:  database.Text(""),
			StudentID: database.Text("user-id-1"),
		},
		{
			UserID:    database.Text("user-id-2"),
			UserGroup: database.Text("USER_GROUP_PARENT"),
			ParentID:  database.Text("user-id-2"),
			StudentID: database.Text("student-id-2"),
		},
	}

	userNotifiationsRes := []*entities.UserInfoNotification{
		{
			UserID:      database.Text("user-id-1"),
			UserGroup:   database.Text("USER_GROUP_STUDENT"),
			ParentID:    database.Text(""),
			StudentID:   database.Text("user-id-1"),
			StudentName: database.Text("user-name-1"),
			ParentName:  database.Text(""),
		},
		{
			UserID:      database.Text("user-id-2"),
			UserGroup:   database.Text("USER_GROUP_PARENT"),
			ParentID:    database.Text("user-id-2"),
			StudentID:   database.Text("student-id-2"),
			StudentName: database.Text("student-name-2"),
			ParentName:  database.Text("user-name-2"),
		},
	}

	ctx := context.Background()
	t.Run("happy case", func(t *testing.T) {
		findUserFilter := &repositories.FindUserFilter{
			UserIDs: database.TextArray([]string{"user-id-1", "user-id-2"}),
		}
		userRepo.On("FindUser", ctx, mock.Anything, findUserFilter).Once().Return(users, mapUserIDAndUser, nil)
		findUserStudentFilter := &repositories.FindUserFilter{
			UserIDs: database.TextArray([]string{"student-id-2"}),
		}
		userRepo.On("FindUser", ctx, mock.Anything, findUserStudentFilter).Once().Return(userStudents, mapStudentIDAndUser, nil)
		userNotis, err := svc.AssignRetentionNameForUserNotification(ctx, mockDB.DB, userNotifiationsReq)
		assert.Nil(t, err)
		assert.Equal(t, userNotifiationsRes[0], userNotis[0])
		assert.Equal(t, userNotifiationsRes[1], userNotis[1])
	})
}

func TestDataRetention_AssignIndividualRetentionNamesForNotification(t *testing.T) {
	mockDB := testutil.NewMockDB()
	userRepo := &mock_repositories.MockUserRepo{}

	svc := &DataRetentionService{
		UserRepo: userRepo,
	}

	users := []*entities.User{
		{
			UserID: database.Text("individual_1"),
			Name:   database.Text("user-name-1"),
		},
		{
			UserID: database.Text("individual_2"),
			Name:   database.Text("user-name-2"),
		},
	}

	mapUserIDAndUser := map[string]*entities.User{
		"individual_1": users[0],
		"individual_2": users[1],
	}

	ctx := context.Background()
	t.Run("happy case", func(t *testing.T) {
		notification := utils.GenSampleNotification()
		individualIDs := []string{"individual_1", "individual_2"}
		notification.GenericReceiverIds = append(notification.GenericReceiverIds, individualIDs...)
		infoNotification, _ := mappers.PbToInfoNotificationEnt(notification)

		findUserFilter := &repositories.FindUserFilter{
			UserIDs: database.TextArray([]string{"individual_1", "individual_2"}),
		}
		userRepo.On("FindUser", ctx, mock.Anything, findUserFilter).Once().Return(users, mapUserIDAndUser, nil)

		notiActual, err := svc.AssignIndividualRetentionNamesForNotification(ctx, mockDB.DB, infoNotification)
		assert.Nil(t, err)
		receiverNames := database.FromTextArray(notiActual.GenericReceiverIDs)
		assert.Equal(t, "individual_1", receiverNames[0])
		assert.Equal(t, "individual_2", receiverNames[1])
	})

	t.Run("backward compatible with ReceiverIDs", func(t *testing.T) {
		notification := utils.GenSampleNotification()
		notification.Type = cpb.NotificationType_NOTIFICATION_TYPE_NATS_ASYNC
		individualIDs := []string{"individual_1", "individual_2"}
		notification.ReceiverIds = append(notification.ReceiverIds, individualIDs...)
		infoNotification, _ := mappers.PbToInfoNotificationEnt(notification)

		findUserFilter := &repositories.FindUserFilter{
			UserIDs: database.TextArray([]string{"individual_1", "individual_2"}),
		}
		userRepo.On("FindUser", ctx, mock.Anything, findUserFilter).Once().Return(users, mapUserIDAndUser, nil)

		notiActual, err := svc.AssignIndividualRetentionNamesForNotification(ctx, mockDB.DB, infoNotification)
		assert.Nil(t, err)
		receiverNames := database.FromTextArray(notiActual.ReceiverIDs)
		assert.Equal(t, "individual_1", receiverNames[0])
		assert.Equal(t, "individual_2", receiverNames[1])

		emptyNames := database.FromTextArray(notiActual.GenericReceiverIDs)
		assert.Nil(t, emptyNames)
	})
}
