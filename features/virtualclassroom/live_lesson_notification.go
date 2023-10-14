package virtualclassroom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/notification/mock"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// const (
// 	expectedTitleEn = "Live lesson reminder"
// 	expectedTitleJp = "ライブ授業の通知"
// )

func (s *suite) createLiveLessonWithInterval(ctx context.Context, interval string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	date := time.Now()
	message := ""

	switch interval {
	case "24h":
		date = date.Add(24 * time.Hour)
		message = "Live lesson will start in 24 hours from %s on %s."
		if stepState.OrganizationCountry == cpb.Country_COUNTRY_JP.String() {
			message = "24時間後の %s %s にライブ授業が始まります。"
		}
	case "15m":
		date = date.Add(15 * time.Minute)
		message = "Live lesson will start in 15 minutes from %s on %s."
		if stepState.OrganizationCountry == cpb.Country_COUNTRY_JP.String() {
			message = "15分後の %s %s にライブ授業が始まります。"
		}
	default:
		date = date.Add(48 * time.Hour)
	}

	date = date.In(timeutil.Timezone(bpb.Country(cpb.Country_value[stepState.OrganizationCountry])))
	startTime := date.Format("15:04")
	startDate := date.Format("2006/01/02")
	stepState.Notification = &cpb.Notification{
		Message: &cpb.NotificationMessage{
			Content: &cpb.RichText{
				Raw: fmt.Sprintf(message, startTime, startDate),
			},
		},
		ScheduledAt: timestamppb.New(time.Now().Truncate(time.Minute).Add(time.Minute)),
	}

	ctx, err := s.createLiveLessonWithStudentsForDate(ctx, date)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createLiveLessonWithStudentsForDate(ctx context.Context, startDate time.Time) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE)
	req.StartTime = timestamppb.New(startDate.Round(time.Second))
	req.EndTime = timestamppb.New(startDate.Add(1 * time.Hour))

	return s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(StepStateToContext(ctx, stepState), req)
}

func (s *suite) CreateParentAccountsForStudents(ctx context.Context) (context.Context, error) {
	return s.createsNumberOfStudentsWithParentsInfo(ctx, "2")
}

func (s *suite) createsNumberOfStudentsWithParentsInfo(ctx context.Context, numParentReq string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var numParent int
	var err error
	if numParent, err = strconv.Atoi(numParentReq); err != nil {
		return ctx, fmt.Errorf("s.createsNumberOfStudentsWithParentsInfo: %v", err)
	}

	stepState.StudentParent = map[string][]string{}
	// Create parents
	for _, studentID := range stepState.StudentIds {
		for parentIdx := 0; parentIdx < numParent; parentIdx++ {
			newParent, err := s.CommonSuite.CreateParentForStudent(ctx, studentID)
			if err != nil {
				return ctx, fmt.Errorf("s.CreateParentForStudent: %v", err)
			}
			stepState.StudentParent[studentID] = append(stepState.StudentParent[studentID], newParent.UserProfile.UserId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) waitForCronJobToSendLiveLessonNotificationEvent(ctx context.Context) (context.Context, error) {
	time.Sleep(120 * time.Second)
	return ctx, nil
}

func retrievePushedNotification(ctx context.Context, noti *grpc.ClientConn, deviceToken string) (*npb.RetrievePushedNotificationMessageResponse, error) {
	respNoti, err := npb.NewInternalServiceClient(noti).RetrievePushedNotificationMessages(
		ctx,
		&npb.RetrievePushedNotificationMessageRequest{
			DeviceToken: deviceToken,
			Limit:       1,
			Since:       timestamppb.Now(),
		})

	if err != nil {
		return nil, err
	}
	return respNoti, nil
}

func (s *suite) participantsShouldReceiveNotificationsWithProperIntervalMessage(ctx context.Context, interval string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userIDs := []string{}

	// students
	for _, studentID := range stepState.StudentIds {
		userIDs = append(userIDs, studentID)
		// parents
		userIDs = append(userIDs, stepState.StudentParent[studentID]...)
	}

	for _, receiveID := range userIDs {
		var deviceToken string
		row := s.BobDB.QueryRow(ctx, "SELECT device_token FROM user_device_tokens WHERE user_id = $1", receiveID)
		if err := row.Scan(&deviceToken); err != nil {
			return ctx, fmt.Errorf("error finding user device token with userid: %s: %w", receiveID, err)
		}
		resp, err := retrievePushedNotification(ctx, s.NotificationMgmtConn, deviceToken)
		if err != nil {
			return ctx, fmt.Errorf("error when call NotificationModifierService.RetrievePushedNotificationMessages: %w", err)
		}
		if len(resp.Messages) == 0 {
			err = fmt.Errorf("wrong node: user receive id: " + receiveID + ", device_token: " + deviceToken)
			return ctx, err
		}

		hasMatch := false
		// check all of the messages if one of them contains the message that we want
		for _, notification := range resp.Messages {
			if notification.Body == stepState.Notification.Message.Content.Raw {
				hasMatch = true
				break
			}
		}
		if !hasMatch {
			return ctx, fmt.Errorf("body should be %s and wasn't found on the message list", stepState.Notification.Message.Content.Raw)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UpdateDeviceTokenForLeanerUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIds {
		err := s.updateDeviceToken(ctx, studentID)
		if err != nil {
			return ctx, fmt.Errorf("s.updateDeviceToken: %v", err)
		}
		for _, parentID := range stepState.StudentParent[studentID] {
			err = s.updateDeviceToken(ctx, parentID)
			if err != nil {
				return ctx, fmt.Errorf("s.UpdateDeviceToken: %v", err)
			}
		}
	}

	return ctx, nil
}

func (s *suite) updateStudentsCountryTo(ctx context.Context, country string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIds {
		err := s.updateUserCountry(ctx, studentID, country)
		if err != nil {
			return ctx, fmt.Errorf("s.updateUserCountry: %v", err)
		}
	}
	stepState.OrganizationCountry = country
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateUserCountry(ctx context.Context, userID, country string) error {
	updateQuery := fmt.Sprintf("UPDATE users SET country = '%s' WHERE user_id = '%s'", country, userID)
	if _, err := s.BobDBTrace.Exec(ctx, updateQuery); err != nil {
		return fmt.Errorf("db.Exec %v", err)
	}
	return nil
}

func (s *suite) updateDeviceToken(ctx context.Context, userID string) error {
	deviceToken := mock.MockNotificationPusherValidDeviceToken + "-" + idutil.ULIDNow()

	updateQuery := fmt.Sprintf("UPDATE users SET allow_notification = 'true' WHERE user_id = '%s'", userID)
	if _, err := s.BobDBTrace.Exec(ctx, updateQuery); err != nil {
		return fmt.Errorf("db.Exec %v", err)
	}

	insertQuery := "INSERT INTO user_device_tokens(user_id, device_token, allow_notification, created_at, updated_at) VALUES ($1, $2, true, NOW(), NOW())"
	if _, err := s.BobDBTrace.Exec(ctx, insertQuery, userID, deviceToken); err != nil {
		return fmt.Errorf("db.Exec %v", err)
	}

	return nil
}
