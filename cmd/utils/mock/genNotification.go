package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	mediaModuleRepo "github.com/manabie-com/backend/internal/notification/modules/media/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/queries"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure/repo"
	tagModuleRepo "github.com/manabie-com/backend/internal/notification/modules/tagmgmt/repositories"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/domain"

	"github.com/spf13/cobra"
)

func genNotificationRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"info_notification":               &repositories.InfoNotificationRepo{},
		"info_notification_msg":           &repositories.InfoNotificationMsgRepo{},
		"info_notification_tag":           &repositories.InfoNotificationTagRepo{},
		"user_info_notification":          &repositories.UsersInfoNotificationRepo{},
		"questionnaire":                   &repositories.QuestionnaireRepo{},
		"questionnaire_question":          &repositories.QuestionnaireQuestionRepo{},
		"questionnaire_user_answer":       &repositories.QuestionnaireUserAnswerRepo{},
		"user_device_token":               &repositories.UserDeviceTokenRepo{},
		"notification_student_course":     &repositories.NotificationStudentCourseRepo{},
		"info_notification_access_path":   &repositories.InfoNotificationAccessPathRepo{},
		"location":                        &repositories.LocationRepo{},
		"notification_class_member":       &repositories.NotificationClassMemberRepo{},
		"grade":                           &repositories.GradeRepo{},
		"notification_internal_user":      &repositories.NotificationInternalUserRepo{},
		"audience":                        &repositories.AudienceRepo{},
		"user":                            &repositories.UserRepo{},
		"class":                           &repositories.ClassRepo{},
		"notification_location_filter":    &repositories.NotificationLocationFilterRepo{},
		"notification_course_filter":      &repositories.NotificationCourseFilterRepo{},
		"notification_class_filter":       &repositories.NotificationClassFilterRepo{},
		"questionnaire_template":          &repositories.QuestionnaireTemplateRepo{},
		"questionnaire_template_question": &repositories.QuestionnaireTemplateQuestionRepo{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "notification", repos)

	tagmgmtRepos := map[string]interface{}{
		"tag": &tagModuleRepo.TagRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "modules/tagmgmt/repositories"), "notification/modules/tagmgmt", tagmgmtRepos)

	mediaRepos := map[string]interface{}{
		"media": &mediaModuleRepo.MediaRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "modules/media/repositories"), "notification/modules/media", mediaRepos)

	systemNotificationRepos := map[string]interface{}{
		"system_notification":           &repo.SystemNotificationRepo{},
		"system_notification_recipient": &repo.SystemNotificationRecipientRepo{},
		"system_notification_content":   &repo.SystemNotificationContentRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "modules/system_notification/infrastructure/repo"), "notification/modules/system_notification", systemNotificationRepos)
	structs := map[string][]interface{}{
		"internal/notification/services/domain": {
			&domain.AudienceRetrieverService{},
			&domain.DataRetentionService{},
		},
		"internal/notification/modules/system_notification/application/commands": {
			&commands.SystemNotificationCommandHandler{},
			&commands.SystemNotificationRecipientCommandHandler{},
			&commands.SystemNotificationContentHandler{},
		},
		"internal/notification/modules/system_notification/application/queries": {
			&queries.SystemNotificationQueryHandler{},
		},
	}

	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}

	interfaces := map[string][]string{
		"internal/notification/services": {
			"Uploader",
		},
		"internal/notification/infra": {
			"PushNotificationService",
		},
		"internal/notification/infra/metrics": {
			"NotificationMetrics",
		},
	}
	return tools.GenMockInterfaces(interfaces)
}

func newGenNotificationCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notification [../../mock/notification]",
		Short: "Generate mock repositories for notification",
		Args:  cobra.ExactArgs(1),
		RunE:  genNotificationRepo,
	}
}
