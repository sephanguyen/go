package communication

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MultiTenantPushNotificationByNatsJetStreamSuite struct {
	*common.NotificationSuite
	studentDeviceToken map[int]string
	natsNotification   *ypb.NatsCreateNotificationRequest
}

func (c *SuiteConstructor) InitMultiTenantPushNotificationByNatsJetStream(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &MultiTenantPushNotificationByNatsJetStreamSuite{
		NotificationSuite:  dep.notiCommonSuite,
		studentDeviceToken: make(map[int]string),
	}

	stepsMapping := map[string]interface{}{
		`^"([^"]*)" schools with respective "([^"]*)" for each school login to CMS$`:        s.StaffGrantedRoleLoggedInBackOfficeOfRespectiveOrg,
		`^school admin (\d+) has created (\d+) student with grade, course$`:                 s.schoolAdminHasCreatedStudentWithGradeCourse,
		`^student of school (\d+) login to Learner App$`:                                    s.studentOfSchoolLoginToLearnerApp,
		`^school admin (\d+) push "([^"]*)" notification to student of school admin (\d+)$`: s.schoolAdminPushNotificationToStudentOfSchoolAdmin,
		`^wait to "([^"]*)" notification send$`:                                             s.waitToNotificationSend,
		`^student of school admin (\d+) must not receive notification$`:                     s.studentOfSchoolAdminMustNotReceiveNotification,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *MultiTenantPushNotificationByNatsJetStreamSuite) schoolAdminHasCreatedStudentWithGradeCourse(ctx context.Context, schoolIndex int, studentNum string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	for i, tenantCtx := range commonState.MultiTenants {
		if i == schoolIndex {
			_, err := s.CreatesNumberOfStudents(*tenantCtx, studentNum)
			if err != nil {
				return ctx, err
			}
			_, err = s.CreatesNumberOfCourses(*tenantCtx, studentNum)
			if err != nil {
				return ctx, err
			}
			// tenantCtx = &subCtx
		}
	}
	return ctx, nil
}

func (s *MultiTenantPushNotificationByNatsJetStreamSuite) studentOfSchoolLoginToLearnerApp(ctx context.Context, schoolIndex int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for i, tenantCtx := range commonState.MultiTenants {
		if i == schoolIndex {
			subState := common.StepStateFromContext(*tenantCtx)
			student := subState.Students[0]
			studentToken, err := s.GenerateExchangeTokenCtx(*tenantCtx, student.ID, "student")
			if err != nil {
				return ctx, err
			}
			s.studentDeviceToken[schoolIndex] = studentToken
		}
	}
	return ctx, nil
}

func (s *MultiTenantPushNotificationByNatsJetStreamSuite) schoolAdminPushNotificationToStudentOfSchoolAdmin(ctx context.Context, schoolIndex int, notificationType string, schoolIndex2 int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	receiverIDs := []string{}
	for i, tenantCtx := range commonState.MultiTenants {
		if i == schoolIndex2 {
			subState := common.StepStateFromContext(*tenantCtx)
			for _, student := range subState.Students {
				receiverIDs = append(receiverIDs, student.ID)
			}
		}
	}
	for i, tenantCtx := range commonState.MultiTenants {
		if i == schoolIndex {
			subState := common.StepStateFromContext(*tenantCtx)

			tracingID := uuid.New().String()
			notification := &ypb.NatsCreateNotificationRequest{
				ClientId:       "bdd_testing_client_id",
				SendingMethods: []string{consts.SendingMethodPushNotification},
				Target:         &ypb.NatsNotificationTarget{},
				NotificationConfig: &ypb.NatsPushNotificationConfig{
					Mode:             consts.NotificationModeNotify,
					PermanentStorage: true,
					Notification: &ypb.NatsNotification{
						Title:   fmt.Sprintf("nats notify %v", tracingID),
						Message: "popup message",
						Content: "<h1>hello world</h1>",
					},
					Data: map[string]string{
						"custom_data_type": "eibanam",
					},
				},
				SendTime: &ypb.NatsNotificationSendTime{
					Type: notificationType,
				},
				TracingId: tracingID,
				SchoolId:  subState.CurrentOrganicationID,
			}

			if notificationType == consts.NotificationTypeScheduled {
				notification.SendTime.Time = time.Now().Add(1 * time.Minute).Format(LauoutTimeFormat)
			}

			notification.NotificationConfig.PermanentStorage = true
			notification.Target.ReceivedUserIds = receiverIDs
			notification.TargetGroup = &ypb.NatsNotificationTargetGroup{
				UserGroupFilter: &ypb.NatsNotificationTargetGroup_UserGroupFilter{
					UserGroups: []string{consts.TargetUserGroupStudent},
				},
			}

			s.natsNotification = notification

			// inject resource path for nats js stream
			ctxInjectResourcePath := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: strconv.Itoa(int(subState.CurrentOrganicationID)),
					UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
					UserID:       subState.CurrentStaff.ID,
				},
			})
			data, _ := proto.Marshal(notification)
			err := s.PublishToNats(ctxInjectResourcePath, "Notification.Created", data)
			if err != nil {
				return ctxInjectResourcePath, err
			}
		}
	}
	return ctx, nil
}

func (s *MultiTenantPushNotificationByNatsJetStreamSuite) waitToNotificationSend(ctx context.Context, sendType string) (context.Context, error) {
	if sendType == consts.NotificationTypeScheduled {
		fmt.Printf("\nWaiting for %s notification to be sent...\n", sendType)
		notify := s.natsNotification
		sendTime, _ := time.Parse(LauoutTimeFormat, notify.SendTime.Time)
		if sendTime.After(time.Now()) {
			waitTime := time.Duration(sendTime.Unix()-time.Now().Unix()+60) * time.Second
			time.Sleep(waitTime)
		}
	} else {
		time.Sleep(10 * time.Second)
	}
	return ctx, nil
}

func (s *MultiTenantPushNotificationByNatsJetStreamSuite) studentOfSchoolAdminMustNotReceiveNotification(ctx context.Context, schoolIndex int) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	for i, tenantCtx := range commonState.MultiTenants {
		if i == schoolIndex {
			subState := common.StepStateFromContext(*tenantCtx)
			resp, err := npb.NewInternalServiceClient(s.NotificationMgmtGRPCConn).RetrievePushedNotificationMessages(
				s.ContextWithToken(*tenantCtx, subState.AuthToken),
				&npb.RetrievePushedNotificationMessageRequest{
					DeviceToken: s.studentDeviceToken[schoolIndex],
					Limit:       1,
					Since:       timestamppb.Now(),
				})
			if err != nil {
				return ctx, fmt.Errorf("error when call NotificationModifierService.RetrievePushedNotificationMessages: %w", err)
			}
			if len(resp.Messages) != 0 {
				return ctx, fmt.Errorf("cross-notification push through Nats Jetstream between tenants is prohibited, but users can still do it")
			}
			notifications, err := s.GetNotificationByUser(s.studentDeviceToken[schoolIndex], true)
			if err != nil {
				return ctx, err
			}
			if len(notifications) != 0 {
				return ctx, fmt.Errorf("cross-notification push through Nats Jetstream between tenants is prohibited, but users can still do it")
			}
		}
	}

	return ctx, nil
}
