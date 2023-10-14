package communication

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/manabie-com/backend/features/common"
	noti_common "github.com/manabie-com/backend/features/communication/common"
	helper2 "github.com/manabie-com/backend/features/eibanam/communication/helper"

	"github.com/cucumber/godog"
	"github.com/ettle/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func mapFeaturesToStepFuncs(parctx *godog.ScenarioContext, conf *common.Config) {
	parctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		helper := helper2.NewCommunicationHelper(
			connections.BobDB,
			connections.BobConn, // bob
			connections.TomConn,
			connections.YasuoConn, // yasuo
			connections.UserMgmtConn,
			firebaseAddr,
			conf.FirebaseAPIKey,
			conf.BobHasuraAdminURL,
			connections.JSM,
			connections.ShamirConn,
			applicantID,
			connections,
		)
		uriSplit := strings.Split(sc.Uri, ":")
		uri := uriSplit[0]

		switch uri {
		case "communication/stress_test_chat.feature":
			InitStepFuncDynamically(parctx, uri, helper)
			return ctx, nil

		case "communication/create_and_update_notification.feature",
			"communication/create_and_update_notification_with_access_path.feature",
			"communication/create_and_update_scheduled_notification.feature",
			"communication/send_batch_scheduled_notifications.feature",
			"communication/discard_notification.feature",
			"communication/notify_user_unread_notification.feature",
			"communication/send_notification_with_fcm_step_is_failed.feature",
			"communication/send_notification.feature",
			"communication/send_scheduled_notification.feature",
			"communication/get_notifications_by_filter.feature",
			"communication/notification_sync_student_package.feature",
			"communication/job_migrate_student_package_and_package_class.feature",
			"communication/get_tags_by_filter.feature",
			"communication/questionnaire_view.feature",
			"communication/tag_create.feature",
			"communication/tag_check_tag_name_exist.feature",
			"communication/tag_delete.feature",
			"communication/tag_attach_to_notification.feature",
			"communication/multi_tenant_push_notification_by_nats_jet_stream.feature",
			"communication/notification_for_parent_with_multiple_student.feature",
			"communication/async_push_notification_by_nats_jet_stream.feature",
			"communication/questionnaire_create.feature",
			"communication/questionnaire_submit.feature",
			"communication/cron_job_resend_scheduled_notification_after_send_failed.feature",
			"communication/retrieve_info_notification_detail.feature",
			"communication/retrieve_info_notifications.feature",
			"communication/send_notification_with_attachment.feature",
			"communication/set_status_for_user_notifications.feature",
			"communication/update_device_token_v2.feature",
			"communication/mark_user_info_notification_as_read.feature",
			"communication/access_control_for_create_notification.feature",
			"communication/access_control_for_update_notification.feature",
			"communication/access_control_for_get_notification.feature",
			"communication/access_control_for_discard_notification.feature",
			"communication/jprep_sync_student_package.feature",
			"communication/get_audience_for_view_recipient_list.feature",
			"communication/jprep_sync_student_class.feature",
			"communication/tag_import_csv.feature",
			"communication/tag_export_csv.feature",
			"communication/job_migrate_jprep_student_package.feature",
			"communication/send_notification_with_excluded_generic_receiver_ids.feature",
			"communication/questionnaire_download_csv.feature",
			"communication/job_migrate_notification_course_filter.feature",
			"communication/job_migrate_notification_location_filter.feature",
			"communication/job_migrate_notification_class_filter.feature",
			"communication/job_migrate_notification_assignment_return.feature",
			"communication/get_audience_for_draft_scheduled_notification.feature",
			"communication/create_and_update_system_notification.feature",
			"communication/send_email.feature",
			"communication/get_system_notifications.feature",
			"communication/questionnaire_template_create.feature",
			"communication/send_scheduled_notification_by_cronjob.feature",
			"communication/delete_notification.feature",
			"communication/set_status_for_system_notification.feature":
			InitStepFuncDynamicallyV2(parctx, uri, conf)
			ctx = initNotificationCommonState(ctx)
			return ctx, nil

		default:
			return ctx, fmt.Errorf("unknown mapping for files %s", uri)
		}
	})
}

type SuiteConstructor struct{}
type Dependency struct {
	helper      *helper2.CommunicationHelper
	conns       *common.Connections
	commonSuite *common.Suite
}

type DependencyV2 struct {
	notiCommonSuite *noti_common.NotificationSuite
}

// If we have some feature file like abc_xyz.feature
// you need to create a construct method func(c *SuiteConstructor) InitAbcXyz(dep *Dependency,ctx *godog.ScenarioContext)
func InitStepFuncDynamically(parctx *godog.ScenarioContext, uri string, helper *helper2.CommunicationHelper) {
	constructor := &SuiteConstructor{}
	parts := strings.Split(uri, "/")
	filename := parts[len(parts)-1]
	featureName := filename[:len(filename)-len(".feature")]
	featureCamelCase := strcase.ToCamel(featureName)
	caser := cases.Title(language.English, cases.NoLower)
	constructMethod := fmt.Sprintf("Init%s", caser.String(featureCamelCase))

	meth := reflect.ValueOf(constructor).MethodByName(constructMethod)
	if !meth.IsValid() {
		panic(fmt.Sprintf("feature %s has no construct method %s", featureName, constructMethod))
	}
	dep := reflect.ValueOf(&Dependency{helper: helper, conns: connections, commonSuite: newCommonSuite()})
	parCtx := reflect.ValueOf(parctx)
	meth.Call([]reflect.Value{dep, parCtx})
}

func (c *SuiteConstructor) InitScenarioStepMapping(ctx *godog.ScenarioContext, stepsMapping map[string]interface{}) {
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func InitStepFuncDynamicallyV2(parctx *godog.ScenarioContext, uri string, cfg *common.Config) {
	constructor := &SuiteConstructor{}
	parts := strings.Split(uri, "/")
	filename := parts[len(parts)-1]
	featureName := filename[:len(filename)-len(".feature")]
	featureCamelCase := strcase.ToCamel(featureName)
	caser := cases.Title(language.English, cases.NoLower)
	constructMethod := fmt.Sprintf("Init%s", caser.String(featureCamelCase))

	meth := reflect.ValueOf(constructor).MethodByName(constructMethod)
	if !meth.IsValid() {
		panic(fmt.Sprintf("feature %s has no construct method %s", featureName, constructMethod))
	}
	dep := reflect.ValueOf(&DependencyV2{notiCommonSuite: newNotificationCommonSuite(cfg)})
	parCtx := reflect.ValueOf(parctx)
	meth.Call([]reflect.Value{dep, parCtx})
}
