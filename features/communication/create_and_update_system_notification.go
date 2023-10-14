package communication

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"k8s.io/utils/strings/slices"
)

type CreateAndUpdateSystemNotificationSuite struct {
	*common.NotificationSuite
	UserIDs                         []string
	UpsertSystemNotificationPayload payload.UpsertSystemNotification
}

func (c *SuiteConstructor) InitCreateAndUpdateSystemNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &CreateAndUpdateSystemNotificationSuite{
		NotificationSuite: dep.notiCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^publish event from kafka$`:                                                                s.publishEventFromKafka,
		`^an "([^"]*)" upsert system notification kafka payload$`:                                   s.anUpsertSystemNotificationKafkaPayload,
		`^system notification data must be "([^"]*)"$`:                                              s.systemNotificationDataMustBe,
		`^some staffs with random roles and granted organization location of current organization$`: s.CreateSomeStaffsWithSomeRolesAndGrantedOrgLevelLocationOfCurrentOrganization,
		`^admin update sent system notification to deleted$`:                                        s.adminUpdateSentSystemNotificationToDeleted,
		`^admin update content of system notification$`:                                             s.adminUpdateContentOfSystemNotification,
	}
	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *CreateAndUpdateSystemNotificationSuite) anUpsertSystemNotificationKafkaPayload(ctx context.Context, payloadType string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	switch payloadType {
	case "valid":
		now := time.Now()
		recipients := []payload.SystemNotificationRecipient{}
		i := common.RandRangeIn(1, 3)
		selectedStaffID := commonState.Organization.Staffs[i].ID
		recipients = append(recipients, payload.SystemNotificationRecipient{
			UserID: selectedStaffID,
		})
		s.UserIDs = append(s.UserIDs, selectedStaffID)
		s.UpsertSystemNotificationPayload = payload.UpsertSystemNotification{
			ReferenceID: idutil.ULIDNow(),
			Content: []payload.SystemNotificationContent{
				{
					Language: "en",
					Text:     "text_en",
				},
				{
					Language: "vi",
					Text:     "text_vi",
				},
			},
			URL:        "https://manabie.com",
			ValidFrom:  now,
			Recipients: recipients,
			Status:     payload.SystemNotificationStatusNew,
		}
	case "invalid":
		s.UpsertSystemNotificationPayload = payload.UpsertSystemNotification{
			ReferenceID: idutil.ULIDNow(),
			Content:     []payload.SystemNotificationContent{},
			URL:         "https://manabie.com",
			Status:      payload.SystemNotificationStatusNew,
		}
	}
	return ctx, nil
}

func (s *CreateAndUpdateSystemNotificationSuite) publishEventFromKafka(ctx context.Context) (context.Context, error) {
	data, err := json.Marshal(s.UpsertSystemNotificationPayload)
	if err != nil {
		return ctx, fmt.Errorf("failed Marshal: %+v", err)
	}
	err = s.PublishToKafka(ctx, constants.SystemNotificationUpsertingTopic, data)
	if err != nil {
		return ctx, fmt.Errorf("failed PublishToKafka: %+v", err)
	}
	return ctx, nil
}

func (s *CreateAndUpdateSystemNotificationSuite) adminUpdateSentSystemNotificationToDeleted(ctx context.Context) (context.Context, error) {
	s.UpsertSystemNotificationPayload.IsDeleted = true
	// to make sure that the service still work even if we don't send other data
	s.UpsertSystemNotificationPayload.Content = nil
	s.UpsertSystemNotificationPayload.Recipients = nil
	s.UpsertSystemNotificationPayload.URL = ""
	return s.publishEventFromKafka(ctx)
}

func (s *CreateAndUpdateSystemNotificationSuite) adminUpdateContentOfSystemNotification(ctx context.Context) (context.Context, error) {
	s.UpsertSystemNotificationPayload.Content = []payload.SystemNotificationContent{
		{Language: "jp", Text: "どらえもん"},
		{Language: "vi", Text: "Bạn có một thông báo"},
	}
	return s.publishEventFromKafka(ctx)
}

func (s *CreateAndUpdateSystemNotificationSuite) systemNotificationDataMustBe(ctx context.Context, dataMustBe string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	expectEvent := s.UpsertSystemNotificationPayload
	actualEvent := &model.SystemNotification{}
	err := doRetry(func() (bool, error) {
		query := fmt.Sprintf(`
			SELECT %s
			FROM system_notifications sn
			WHERE sn.reference_id = $1 AND sn.resource_path = $2 AND sn.deleted_at IS NULL;
		`, strings.Join(database.GetFieldNames(actualEvent), ","))
		err := s.NotificationMgmtDBConn.QueryRow(ctx, query,
			database.Text(expectEvent.ReferenceID),
			database.Text(commonState.CurrentResourcePath)).
			Scan(database.GetScanFields(actualEvent, database.GetFieldNames(actualEvent))...)
		if err != nil {
			if dataMustBe == "not created" {
				if errors.Is(err, pgx.ErrNoRows) {
					// found no data, retry to check if data would appear
					return true, nil
				}
				// return with unexpected error
				return false, fmt.Errorf("failed scan: %v", err)
			}
			// else return with no error as we expected no data be created
			return false, nil
		}

		err = compareExpectEventAndActualEvent(expectEvent, actualEvent)
		if err != nil {
			return true, err
		}

		queryRecipients := `
		SELECT user_id
		FROM system_notification_recipients snr
		JOIN system_notifications sn ON sn.system_notification_id = snr.system_notification_id
		WHERE sn.reference_id = $1 AND sn.resource_path = $2;
	`
		rows, err := s.NotificationMgmtDBConn.Query(ctx, queryRecipients,
			database.Text(expectEvent.ReferenceID),
			database.Text(commonState.CurrentResourcePath),
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return true, fmt.Errorf("found no data, retrying")
			}
			return false, fmt.Errorf("failed query rows Recipient: %v", err)
		}

		defer rows.Close()

		for rows.Next() {
			var userID string
			err = rows.Scan(&userID)
			if err != nil {
				return false, fmt.Errorf("failed scan row Recipient: %v", err)
			}

			if !slices.Contains(s.UserIDs, userID) {
				return false, fmt.Errorf("unexpected user id found %s, system notification %s", userID, actualEvent.SystemNotificationID.String)
			}
		}

		queryContent := `
			SELECT snc.system_notification_content_id, snc.language, snc.text
			FROM system_notification_contents snc
			JOIN system_notifications sn ON sn.system_notification_id = snc.system_notification_id
			WHERE sn.reference_id = $1 AND snc.resource_path = $2 AND snc.deleted_at IS NULL;
		`
		rows, err = s.NotificationMgmtDBConn.Query(ctx, queryContent,
			database.Text(expectEvent.ReferenceID),
			database.Text(commonState.CurrentResourcePath),
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return true, fmt.Errorf("found no data, retrying")
			}
			return false, fmt.Errorf("failed query rows Content: %v", err)
		}

		defer rows.Close()
		mapContentVersion := make(map[string]string, 0)
		var sncID string
		for rows.Next() {
			var sncLanguage, sncText string
			err = rows.Scan(&sncID, &sncLanguage, &sncText)
			if err != nil {
				return false, fmt.Errorf("failed scan row Content: %v", err)
			}
			if _, found := mapContentVersion[sncLanguage]; found {
				return false, fmt.Errorf("found duplicate of content %s [language %s and text %s]", sncID, sncLanguage, sncText)
			}
			mapContentVersion[sncLanguage] = sncText
		}
		if len(mapContentVersion) == 0 {
			return false, fmt.Errorf("not found content data, reference ID %s", expectEvent.ReferenceID)
		}
		for _, expectContent := range expectEvent.Content {
			if lngText, exist := mapContentVersion[expectContent.Language]; exist {
				if lngText != expectContent.Text {
					return false, fmt.Errorf("expected content %s to have text '%s', found '%s'", sncID, expectContent.Text, lngText)
				}
			} else {
				// maybe upsert logic for SystemNotificationContent is not finished yet,
				// so the data we queried is from the old content, not yet deleted
				// => retry to see if it would update new data
				return true, fmt.Errorf("found old data, retrying for new SystemNotificationContent data")
			}
		}
		return false, nil
	})
	if err != nil {
		return ctx, err
	}

	return common.StepStateToContext(ctx, commonState), nil
}

func compareExpectEventAndActualEvent(expectEvent payload.UpsertSystemNotification, actualEvent *model.SystemNotification) error {
	if actualEvent.SystemNotificationID.Status != pgtype.Present {
		return fmt.Errorf("error System Notification missing ID. reference ID: %v", expectEvent.ReferenceID)
	}

	if actualEvent.URL.String != expectEvent.URL {
		return fmt.Errorf("error System Notification stored incorrect URL data. reference ID: %v", expectEvent.ReferenceID)
	}

	if actualEvent.ValidFrom.Time.Unix() != expectEvent.ValidFrom.Unix() {
		return fmt.Errorf("error System Notification stored incorrect ValidFrom data. reference ID: %v", expectEvent.ReferenceID)
	}

	if expectEvent.IsDeleted && actualEvent.DeletedAt.Status == pgtype.Null {
		return fmt.Errorf("error System Notification stored incorrect DeletedAt data. reference ID: %v", expectEvent.ReferenceID)
	}

	return nil
}
