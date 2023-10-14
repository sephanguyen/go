package domain

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

type DataRetentionService struct {
	Env      string
	UserRepo interface {
		FindUser(ctx context.Context, db database.QueryExecer, filter *repositories.FindUserFilter) ([]*entities.User, map[string]*entities.User, error)
	}
}

func NewDataRetentionService(env string) *DataRetentionService {
	return &DataRetentionService{
		Env:      env,
		UserRepo: &repositories.UserRepo{},
	}
}

func (s *DataRetentionService) AssignRetentionNameForUserNotification(ctx context.Context, db database.QueryExecer, userNotifications entities.UserInfoNotifications) (entities.UserInfoNotifications, error) {
	userIDs := make([]string, 0)
	studentIDs := make([]string, 0)
	for _, userNotification := range userNotifications {
		userIDs = append(userIDs, userNotification.UserID.String)
		if userNotification.StudentID.String != "" && userNotification.UserGroup.String == cpb.UserGroup_USER_GROUP_PARENT.String() {
			studentIDs = append(studentIDs, userNotification.StudentID.String)
		}
	}

	findUserFilter := &repositories.FindUserFilter{
		UserIDs: database.TextArray(userIDs),
	}
	_, mapUserIDAndUser, err := s.UserRepo.FindUser(ctx, db, findUserFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to find users in domain.DataRetentionService.AssignRetentionName: %v", err)
	}

	findUserStudentFilter := &repositories.FindUserFilter{
		UserIDs: database.TextArray(studentIDs),
	}
	_, mapStudentIDAndUser, err := s.UserRepo.FindUser(ctx, db, findUserStudentFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to find user students in domain.DataRetentionService.AssignRetentionName: %v", err)
	}

	for _, userNotification := range userNotifications {
		userNotification.StudentName = database.Text("")
		userNotification.ParentName = database.Text("")

		userName := ""
		if user, ok := mapUserIDAndUser[userNotification.UserID.String]; ok {
			userName = user.Name.String
		}

		if userNotification.UserGroup.String == cpb.UserGroup_USER_GROUP_STUDENT.String() {
			userNotification.StudentName = database.Text(userName)
		} else if userNotification.UserGroup.String == cpb.UserGroup_USER_GROUP_PARENT.String() {
			userNotification.ParentName = database.Text(userName)

			studentName := ""
			if user, ok := mapStudentIDAndUser[userNotification.StudentID.String]; ok {
				studentName = user.Name.String
			}
			userNotification.StudentName = database.Text(studentName)
		}
	}

	return userNotifications, nil
}

func (s *DataRetentionService) AssignIndividualRetentionNamesForNotification(ctx context.Context, db database.QueryExecer, notification *entities.InfoNotification) (*entities.InfoNotification, error) {
	var err error
	var receivers []*entities.User
	// backward compatible with current squads using NATs with old ReceiverIDs
	if len(notification.ReceiverIDs.Elements) > 0 &&
		notification.Type.String == cpb.NotificationType_NOTIFICATION_TYPE_NATS_ASYNC.String() {
		receivers, _, err = s.UserRepo.FindUser(ctx, db, &repositories.FindUserFilter{UserIDs: notification.ReceiverIDs})
	} else {
		// Get receiver_names from generic_receiver_ids
		receivers, _, err = s.UserRepo.FindUser(ctx, db, &repositories.FindUserFilter{UserIDs: notification.GenericReceiverIDs})
	}
	if err != nil {
		return nil, fmt.Errorf("svc.NotificationUserRepo.Get: %v", err)
	}

	receiverNames := make([]string, 0)
	for _, receiver := range receivers {
		receiverNames = append(receiverNames, receiver.Name.String)
	}

	notification.ReceiverNames = database.TextArray(receiverNames)

	return notification, nil
}
