package communication

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/notification/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	natsJS "github.com/nats-io/nats.go"
)

type UpdateDeviceTokenV2Suite struct {
	*common.NotificationSuite
	studentID   string
	storedToken string
	foundChan   chan struct{}
	subscribes  []*natsJS.Subscription
}

func (c *SuiteConstructor) InitUpdateDeviceTokenV2(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &UpdateDeviceTokenV2Suite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^NotificationMgmt must publish event to user_device_token channel$`: s.notificationMgmtMustPublishEventToUserDeviceTokenChannel,
		`^student try to store device token$`:                                s.studentTryToStoreDeviceToken,
		`^user\'s device token is stored to DB$`:                             s.usersDeviceTokenIsStoredToDB,
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students$`: s.CreatesNumberOfStudents,
		`^student "([^"]*)" logins to Learner App$`: s.StudentLoginsToLearnerApp,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *UpdateDeviceTokenV2Suite) studentTryToStoreDeviceToken(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	s.storedToken = idutil.ULIDNow()
	s.studentID = commonState.Students[0].ID

	s.foundChan = make(chan struct{}, 1)
	err := s.createUserDeviceTokenCreatedSubscription()
	if err != nil {
		return ctx, err
	}

	_, err = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).UpdateUserDeviceToken(
		common.ContextWithToken(ctx, commonState.Students[0].Token),
		&npb.UpdateUserDeviceTokenRequest{
			UserId:            commonState.Students[0].ID,
			DeviceToken:       s.storedToken,
			AllowNotification: true,
		},
	)
	if err != nil {
		return ctx, fmt.Errorf("error when call pb.NewUserServiceClient.UpdateUserDeviceToken: %w", err)
	}
	return ctx, nil
}

func (s *UpdateDeviceTokenV2Suite) usersDeviceTokenIsStoredToDB(ctx context.Context) (context.Context, error) {
	var userID, deviceToken pgtype.Text
	var allowNotification pgtype.Bool
	_ = userID.Set(s.studentID)
	_ = deviceToken.Set(s.storedToken)
	_ = allowNotification.Set(true)
	e := &entities.UserDeviceToken{}
	query := fmt.Sprintf(`SELECT COUNT(*) from %s where user_id=$1 and device_token=$2 and allow_notification=$3`, e.TableName())
	row := s.BobDBConn.QueryRow(ctx, query, &userID, &deviceToken, &allowNotification)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return ctx, err
	}

	if count < 1 {
		return ctx, fmt.Errorf(`notification doesn't store user device token to %s table`, e.TableName())
	}
	return ctx, nil
}

func (s *UpdateDeviceTokenV2Suite) createUserDeviceTokenCreatedSubscription() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{nats.StartTime(time.Now()), nats.AckWait(2 * time.Second), nats.AckExplicit()},
	}

	handleUserDeviceTokenUpdatedEvent := func(ctx context.Context, data []byte) (bool, error) {
		msg := &pb.EvtUserInfo{}
		err := msg.Unmarshal(data)
		if err != nil {
			return true, err
		}

		if s.studentID == msg.UserId && s.storedToken == msg.DeviceToken {
			s.foundChan <- struct{}{}
			return false, nil
		}
		return false, nil
	}

	sub, err := s.JSM.Subscribe(constants.SubjectUserDeviceTokenUpdated, opts, handleUserDeviceTokenUpdatedEvent)
	if err != nil {
		return fmt.Errorf("S.JSM.Subscribe: %v", err)
	}
	s.subscribes = append(s.subscribes, sub.JetStreamSub)
	return nil
}

func (s *UpdateDeviceTokenV2Suite) notificationMgmtMustPublishEventToUserDeviceTokenChannel(ctx context.Context) (context.Context, error) {
	time.Sleep(5 * time.Second)

	timer := time.NewTimer(time.Second * 10)
	defer timer.Stop()

	select {
	case <-s.foundChan:
		return ctx, nil
	case <-timer.C:
		return ctx, errors.New("time out")
	}
}
