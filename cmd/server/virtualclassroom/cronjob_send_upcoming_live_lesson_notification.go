package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const UpcomingLiveLessonNotificationFeatureFlag = "BACKEND_Lesson_UpcomingLiveLessonNotification"
const SwitchDBConnectionToLessonManagementFeatureFlag = "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

/*
upcoming_lessons
  - pickup lessons in the next 24 hours as long as it does not exist in
    live_lesson_sent_notifications (regardless if 15m or 24h notification)

upcoming_lessons_in_15_minutes
  - pickup lessons within the interval of the next 14 to 18 minutes
    14 to 18 minutes are used to give a bit of leeway for the next 15 minute lesson
*/
const GetUpcomingLiveLessonsQuery = `WITH upcoming_lessons AS (
	SELECT l.lesson_id
	FROM lessons l 
	LEFT JOIN live_lesson_sent_notifications llsn
	    ON l.lesson_id = llsn.lesson_id
	    AND llsn.deleted_at IS NULL
	WHERE l.resource_path = $1 
	    AND l.start_time > $2
	    AND l.start_time <= ($2 + INTERVAL '24 hours') 
	    AND l.teaching_medium = 'LESSON_TEACHING_MEDIUM_ONLINE' 
	    AND l.scheduling_status = 'LESSON_SCHEDULING_STATUS_PUBLISHED' 
	    AND l.deleted_at IS NULL
	    AND llsn.lesson_id IS NULL
	),  
	upcoming_lessons_in_15_minutes AS (
		SELECT l.lesson_id
		FROM lessons l 
		LEFT JOIN live_lesson_sent_notifications llsn
			ON l.lesson_id = llsn.lesson_id
			AND llsn.sent_at_interval = '15m'
			AND llsn.deleted_at IS null
		WHERE l.resource_path = $1 
			AND l.start_time >= ($2 + INTERVAL '14 minutes')
			AND l.start_time < ($2 + INTERVAL '18 minutes')
			AND l.teaching_medium = 'LESSON_TEACHING_MEDIUM_ONLINE' 
			AND l.scheduling_status = 'LESSON_SCHEDULING_STATUS_PUBLISHED' 
			AND l.deleted_at IS NULL
			AND llsn.lesson_id IS NULL
	)
	SELECT lesson_id FROM upcoming_lessons
	UNION
	SELECT lesson_id FROM upcoming_lessons_in_15_minutes`

var OrgToInternalUserMap = map[string]string{
	"-2147483629": "01GSX7KMWVTMH0E8NZ6HHZRCS1",
	"-2147483630": "01GSX7KMWVTMH0E8NZ6KDGYXQ8",
	"-2147483631": "01GSX7KMWVTMH0E8NZ6KTD7TFM",
	"-2147483634": "01GSX7KMWVTMH0E8NZ6MC88MHG",
	"-2147483635": "01GSX7KMWWED9ZZ79GDK4C1XYP",
	"-2147483637": "01GSX7KMWWED9ZZ79GDPAE0ME4",
	"-2147483638": "01GSX7KMWWED9ZZ79GDQG1DRX3",
	"-2147483639": "01GSX7KMWWED9ZZ79GDSYYGR4G",
	"-2147483640": "01GSX7KMWWED9ZZ79GDVZA9GHP",
	"-2147483641": "01GSX7KMWWED9ZZ79GDZ45GSHY",
	"-2147483642": "01GSX7KMWWED9ZZ79GDZ7ZX525",
	"-2147483643": "01GSX7KMWWED9ZZ79GDZ8AW30Z",
	"-2147483644": "01GSX7KMWWED9ZZ79GE0JF1THV",
	"-2147483645": "01GSX7KMWWED9ZZ79GE0ZNQTJ5",
	"-2147483646": "01GSX7KMWWED9ZZ79GE1WC0CKA",
	"-2147483647": "01GSX7KMWWED9ZZ79GE2DBHSRW",
	"-2147483648": "01GSX7KMWWED9ZZ79GE4JTJ309",
}

func init() {
	bootstrap.RegisterJob("send_upcoming_live_lesson_notification", sendUpcomingLiveLessonNotification)
}

func sendUpcomingLiveLessonNotification(ctx context.Context, cfg configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	zLogger := zapLogger.Sugar()
	defer zLogger.Sync() //nolint:errcheck

	unleashClient, err := initUnleash(&cfg, zapLogger)
	if err != nil {
		return fmt.Errorf("failed to initialize unleash client: %s", err)
	}
	unleashClient.WaitForUnleashReady()

	isCronJobEnabled, err := unleashClient.IsFeatureEnabled(UpcomingLiveLessonNotificationFeatureFlag, cfg.Common.Environment)
	if err != nil {
		return fmt.Errorf("[isCronJobEnabled] IsFeatureEnabled() failed: %s", err)
	}

	if !isCronJobEnabled {
		zLogger.Infof("cron job %s disabled, skipping execution", UpcomingLiveLessonNotificationFeatureFlag)
		return nil
	}

	bobDB := rsc.DBWith("bob")
	lessonmgmtDB := rsc.DBWith("lessonmgmt")

	jsm := rsc.NATS()

	zLogger.Info("-----START: Send upcoming live lesson notification-----")
	zLogger.Infof("Environment: %s", cfg.Common.Environment)

	organizationIDs := retrieveOrganizationIDs(ctx, zLogger, bobDB)
	orgToUpcomingLessonsMap := retrieveUpcomingLessons(ctx, zLogger, bobDB, lessonmgmtDB, unleashClient, organizationIDs, cfg.Common.Environment)

	publishUpcomingLiveLessonNotificationByOrg(ctx, zapLogger, jsm, orgToUpcomingLessonsMap)

	zLogger.Info("-----END: Send upcoming live lesson notification-----")
	return nil
}

func retrieveOrganizationIDs(ctx context.Context, zLogger *zap.SugaredLogger, db database.QueryExecer) []string {
	organizationIDs := []string{}
	orgQuery := "SELECT organization_id FROM organizations WHERE deleted_at IS NULL;"
	organizations, err := db.Query(ctx, orgQuery)
	if err != nil {
		zLogger.Fatal("Get organizations failed")
	}
	defer organizations.Close()

	for organizations.Next() {
		var organizationID pgtype.Text

		if err := organizations.Scan(&organizationID); err != nil {
			zLogger.Infof("failed to scan an orgs row: %s", err)
			continue
		}
		organizationIDs = append(organizationIDs, organizationID.String)
	}

	return organizationIDs
}

func retrieveUpcomingLessons(ctx context.Context, zLogger *zap.SugaredLogger, bobDB, lessonmgmtDB database.QueryExecer, unleashClient unleashclient.ClientInstance, orgIDs []string, env string) map[string][]string {
	orgToUpcomingLessonsMap := map[string][]string{}
	now := time.Now()

	zLogger.Infof("Retrieving upcoming live lessons, current time: %s", now.String())

	for _, org := range orgIDs {
		userID, hasData := OrgToInternalUserMap[org]
		if !hasData {
			zLogger.Warnf("Organization %s does not have an internal user, skipping lesson query", org)
			continue
		}

		isSwitchToNewDBEnabled, err := unleashClient.IsFeatureEnabledOnOrganization(SwitchDBConnectionToLessonManagementFeatureFlag, env, org)
		if err != nil {
			zLogger.Warnf("[isSwitchToNewDBEnabled] IsFeatureEnabled() failed: %s", err)
			continue
		}
		db := bobDB
		if isSwitchToNewDBEnabled {
			db = lessonmgmtDB
		}

		ctxOrg := getCtxFromOrg(ctx, userID, org)
		lessons, err := db.Query(ctxOrg, GetUpcomingLiveLessonsQuery, database.Text(org), database.Timestamptz(now))
		if err != nil {
			zLogger.Fatal("Get upcoming lessons failed")
		}
		defer lessons.Close()

		for lessons.Next() {
			var lessonID pgtype.Text

			if err := lessons.Scan(&lessonID); err != nil {
				zLogger.Infof("failed to scan an lessons row: %s", err)
				continue
			}

			orgToUpcomingLessonsMap[org] = append(orgToUpcomingLessonsMap[org], lessonID.String)
		}
		zLogger.Infof("Upcoming live lessons from org %s count: %d", org, len(orgToUpcomingLessonsMap[org]))
	}

	return orgToUpcomingLessonsMap
}

func getKeys(m map[string][]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func getCtxFromOrg(ctx context.Context, userID, orgID string) context.Context {
	return interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			DefaultRole:  cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			UserID:       userID,
			ResourcePath: orgID,
		},
	})
}

func initUnleash(c *configurations.Config, logger *zap.Logger) (unleashclient.ClientInstance, error) {
	unleashClientInstance, err := unleashclient.NewUnleashClientInstance(c.UnleashClientConfig.URL,
		c.UnleashClientConfig.AppName,
		c.UnleashClientConfig.APIToken,
		logger)
	if err != nil {
		return nil, err
	}

	err = unleashClientInstance.ConnectToUnleashClient()
	if err != nil {
		return nil, err
	}
	return unleashClientInstance, nil
}

func publishUpcomingLiveLessonNotificationByOrg(ctx context.Context, zapLogger *zap.Logger, jsm nats.JetStreamManagement, orgToLessonIDs map[string][]string) {
	orgIDs := getKeys(orgToLessonIDs)
	for _, orgID := range orgIDs {
		upcomingLessons := orgToLessonIDs[orgID]
		userID := OrgToInternalUserMap[orgID]
		ctxOrg := getCtxFromOrg(ctx, userID, orgID)

		publishUpcomingLiveLessonNotification(ctxOrg, zapLogger, jsm, upcomingLessons)
	}
}

func publishUpcomingLiveLessonNotification(ctx context.Context, zapLogger *zap.Logger, jsm nats.JetStreamManagement, lessonIDs []string) {
	msg, _ := proto.Marshal(&vpb.UpcomingLiveLessonNotificationRequest{
		LessonIds: lessonIDs,
	})

	err := try.Do(func(attempt int) (bool, error) {
		_, err := jsm.PublishContext(ctx, constants.SubjectUpcomingLiveLessonNotification, msg)
		if err == nil {
			return false, nil
		}
		retry := attempt < 5
		if retry {
			time.Sleep(1 * time.Second)
			return true, fmt.Errorf("temporary error jsm.PublishContext: %s", err.Error())
		}
		return false, err
	})
	if err != nil {
		zapLogger.Error("jsm.PublishContext failed", zap.Error(err))
	}
}
