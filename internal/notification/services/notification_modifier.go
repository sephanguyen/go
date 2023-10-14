package services

import (
	"context"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	bobRepo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/infra"
	"github.com/manabie-com/backend/internal/notification/infra/metrics"
	tagRepo "github.com/manabie-com/backend/internal/notification/modules/tagmgmt/repositories"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/domain"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jackc/pgtype"
)

// ConfigProvider
func NewNotificationModifierService(db database.Ext, c configs.StorageConfig, s3Sess client.ConfigProvider, pushNotificationService infra.PushNotificationService, metrics metrics.NotificationMetrics, jsm nats.JetStreamManagement, env string) *NotificationModifierService {
	return &NotificationModifierService{
		DB:                            db,
		JSM:                           jsm,
		StorageConfig:                 c,
		Env:                           env,
		NotificationMetrics:           metrics,
		Uploader:                      s3manager.NewUploader(s3Sess),
		NotificationAudienceRetriever: domain.NewAudienceRetrieverService(env),
		DataRetentionService:          domain.NewDataRetentionService(env),
		InfoNotificationRepo: &repositories.InfoNotificationRepo{
			InfoNotificationSQLBuilder: repositories.InfoNotificationSQLBuilder{},
		},
		InfoNotificationMsgRepo:           &repositories.InfoNotificationMsgRepo{},
		QuestionnaireRepo:                 &repositories.QuestionnaireRepo{},
		QuestionnaireQuestionRepo:         &repositories.QuestionnaireQuestionRepo{},
		QuestionnaireUserAnswer:           &repositories.QuestionnaireUserAnswerRepo{},
		UserNotificationRepo:              &repositories.UsersInfoNotificationRepo{},
		StudentRepo:                       &bobRepo.StudentRepo{},
		StudentParentRepo:                 &bobRepo.StudentParentRepo{},
		UserRepo:                          &repositories.UserRepo{},
		ActivityLogRepo:                   &bobRepo.ActivityLogRepo{},
		PushNotificationService:           pushNotificationService,
		OrganizationRepo:                  &bobRepo.OrganizationRepo{},
		TagRepo:                           &tagRepo.TagRepo{},
		InfoNotificationTagRepo:           &repositories.InfoNotificationTagRepo{},
		NotificationStudentCourseRepo:     &repositories.NotificationStudentCourseRepo{},
		UserDeviceTokenRepo:               &repositories.UserDeviceTokenRepo{},
		LocationRepo:                      &repositories.LocationRepo{},
		InfoNotificationAccessPathRepo:    &repositories.InfoNotificationAccessPathRepo{},
		NotificationClassMemberRepo:       &repositories.NotificationClassMemberRepo{},
		NotificationInternalUserRepo:      &repositories.NotificationInternalUserRepo{},
		ClassRepo:                         &repositories.ClassRepo{},
		NotificationUserRepo:              &repositories.UserRepo{},
		NotificationLocationFilterRepo:    &repositories.NotificationLocationFilterRepo{},
		NotificationCourseFilterRepo:      &repositories.NotificationCourseFilterRepo{},
		NotificationClassFilterRepo:       &repositories.NotificationClassFilterRepo{},
		QuestionnaireTemplateRepo:         &repositories.QuestionnaireTemplateRepo{},
		QuestionnaireTemplateQuestionRepo: &repositories.QuestionnaireTemplateQuestionRepo{},
	}
}

// Deprecated: only for backward compatible with Bob, DO NOT USED
func NewSimpleNotificationModifierService(db database.Ext) *NotificationModifierService {
	return &NotificationModifierService{
		DB:                   db,
		UserNotificationRepo: &repositories.UsersInfoNotificationRepo{},
	}
}

type NotificationModifierService struct {
	DB                      database.Ext
	JSM                     nats.JetStreamManagement
	StorageConfig           configs.StorageConfig
	Env                     string
	PushNotificationService infra.PushNotificationService
	metrics.NotificationMetrics
	Uploader

	// Depend services
	NotificationAudienceRetriever interface {
		FindAudiences(ctx context.Context, db database.QueryExecer, notification *entities.InfoNotification) ([]*entities.Audience, error)
	}
	DataRetentionService interface {
		AssignRetentionNameForUserNotification(ctx context.Context, db database.QueryExecer, userNotifications entities.UserInfoNotifications) (entities.UserInfoNotifications, error)
		AssignIndividualRetentionNamesForNotification(ctx context.Context, db database.QueryExecer, notification *entities.InfoNotification) (*entities.InfoNotification, error)
	}

	// Repositories
	InfoNotificationRepo interface {
		Find(ctx context.Context, db database.QueryExecer, filter *repositories.FindNotificationFilter) (entities.InfoNotifications, error)
		Upsert(ctx context.Context, db database.QueryExecer, n *entities.InfoNotification) (string, error)
		SetStatus(ctx context.Context, db database.QueryExecer, notificationID, status pgtype.Text) error
		SetSentAt(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text) error
		UpdateNotification(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text, attributes map[string]interface{}) error
		DiscardNotification(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text, statues pgtype.TextArray) error
		IsNotificationDeleted(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text) (bool, error)
	}

	InfoNotificationMsgRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, n *entities.InfoNotificationMsg) error
		GetByIDs(ctx context.Context, db database.QueryExecer, notiMsgIDs pgtype.TextArray) (entities.InfoNotificationMsgs, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, notificationMsgIDs []string) error
	}

	QuestionnaireRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, id string) (*entities.Questionnaire, error)
		FindQuestionsByQnID(ctx context.Context, db database.QueryExecer, id string) (entities.QuestionnaireQuestions, error)
		Upsert(ctx context.Context, db database.QueryExecer, questionnaire *entities.Questionnaire) error
		SoftDelete(ctx context.Context, db database.QueryExecer, questionnaireID []string) error
		FindUserAnswers(ctx context.Context, db database.QueryExecer, filter *repositories.FindUserAnswersFilter) (entities.QuestionnaireUserAnswers, error)
	}

	QuestionnaireQuestionRepo interface {
		BulkForceUpsert(ctx context.Context, db database.QueryExecer, items entities.QuestionnaireQuestions) error
		SoftDelete(ctx context.Context, db database.QueryExecer, questionnaireID []string) error
	}

	QuestionnaireUserAnswer interface {
		SoftDelete(ctx context.Context, db database.QueryExecer, answerIDs []string) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.QuestionnaireUserAnswers) error
		SoftDeleteByQuestionnaireID(ctx context.Context, db database.QueryExecer, questionnaireID []string) error
	}

	UserNotificationRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, userInfoNotification []*entities.UserInfoNotification) error
		FindUserIDs(ctx context.Context, db database.QueryExecer, filter repositories.FindUserNotificationFilter) (map[string]entities.UserInfoNotifications, error)
		UpdateUnreadUser(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text, userIDs pgtype.TextArray) error
		Find(ctx context.Context, db database.QueryExecer, filter repositories.FindUserNotificationFilter) (entities.UserInfoNotifications, error)
		SetQuestionnareStatusAndSubmittedAt(ctx context.Context, db database.QueryExecer, userNotificationID string, status string, submittedAt pgtype.Timestamptz) error
		SetStatusByNotificationIDs(ctx context.Context, db database.QueryExecer, userID pgtype.Text, notificationIDs pgtype.TextArray, status pgtype.Text) error
		SetStatus(ctx context.Context, db database.QueryExecer, userID pgtype.Text, notificationID pgtype.TextArray, status pgtype.Text) error
		SoftDeleteByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) error
	}

	StudentRepo interface {
		FindStudents(ctx context.Context, db database.QueryExecer, filter bobRepo.FindStudentFilter) ([]*bobEntities.Student, error)
	}

	StudentParentRepo interface {
		FindParentByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*bobEntities.Parent, error)
		GetStudentParents(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*bobEntities.StudentParent, error)
	}

	UserRepo interface {
		FindUser(ctx context.Context, db database.QueryExecer, filter *repositories.FindUserFilter) ([]*entities.User, map[string]*entities.User, error)
	}

	NotificationUserRepo interface {
		FindUser(ctx context.Context, db database.QueryExecer, filter *repositories.FindUserFilter) ([]*entities.User, map[string]*entities.User, error)
	}

	ActivityLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *bobEntities.ActivityLog) error
	}

	OrganizationRepo interface {
		GetOrganizations(ctx context.Context, db database.Ext) ([]string, error)
	}

	TagRepo interface {
		CheckTagIDsExist(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (bool, error)
	}

	InfoNotificationTagRepo interface {
		GetByNotificationIDs(ctx context.Context, db database.QueryExecer, notificationIDs pgtype.TextArray) (map[string]entities.InfoNotificationsTags, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, ents []*entities.InfoNotificationTag) error
		SoftDelete(ctx context.Context, db database.QueryExecer, filter *repositories.SoftDeleteNotificationTagFilter) error
	}

	NotificationStudentCourseRepo interface {
		Find(ctx context.Context, db database.QueryExecer, filter *repositories.FindNotificationStudentCourseFilter) (entities.NotificationStudentCourses, error)
		BulkCreate(ctx context.Context, db database.QueryExecer, items []*entities.NotificationStudentCourse) error
		Upsert(ctx context.Context, db database.QueryExecer, studentSubscriptionItem *entities.NotificationStudentCourse) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.NotificationStudentCourse) error
		SoftDelete(ctx context.Context, db database.QueryExecer, filter *repositories.SoftDeleteNotificationStudentCourseFilter) error
	}

	UserDeviceTokenRepo interface {
		UpsertUserDeviceToken(ctx context.Context, db database.QueryExecer, u *entities.UserDeviceToken) error
		FindByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) (entities.UserDeviceTokens, error)
	}

	LocationRepo interface {
		GetGrantedLocationsByUserIDAndPermissions(ctx context.Context, db database.QueryExecer, userID string, permission []string) ([]string, map[string]string, error)
		GetLocationAccessPathsByIDs(ctx context.Context, db database.QueryExecer, locationIDs []string) (map[string]string, error)
	}

	InfoNotificationAccessPathRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, notiLocation *entities.InfoNotificationAccessPath) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.InfoNotificationAccessPaths) error
		GetByNotificationIDAndNotInLocationIDs(ctx context.Context, db database.QueryExecer, notificationID string, locationIDs []string) (entities.InfoNotificationAccessPaths, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, filter *repositories.SoftDeleteNotificationAccessPathFilter) error
	}

	NotificationClassMemberRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, notificationClassMemberItem *entities.NotificationClassMember) error
		SoftDeleteByFilter(ctx context.Context, db database.QueryExecer, filter *repositories.NotificationClassMemberFilter) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.NotificationClassMember) error
	}

	NotificationInternalUserRepo interface {
		GetByOrgID(ctx context.Context, db database.QueryExecer, orgID string) (*entities.NotificationInternalUser, error)
	}

	ClassRepo interface {
		FindCourseIDByClassID(ctx context.Context, db database.QueryExecer, classID string) (string, error)
	}

	NotificationLocationFilterRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.NotificationLocationFilters) error
		SoftDeleteByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) error
	}

	NotificationCourseFilterRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.NotificationCourseFilters) error
		SoftDeleteByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) error
	}

	NotificationClassFilterRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.NotificationClassFilters) error
		SoftDeleteByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) error
	}

	QuestionnaireTemplateRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, questionnaireTemplate *entities.QuestionnaireTemplate) error
		CheckIsExistNameAndType(ctx context.Context, db database.QueryExecer, filter *repositories.CheckTemplateNameFilter) (bool, error)
	}

	QuestionnaireTemplateQuestionRepo interface {
		BulkForceUpsert(ctx context.Context, db database.QueryExecer, items entities.QuestionnaireTemplateQuestions) error
		SoftDelete(ctx context.Context, db database.QueryExecer, questionnaireTemplateID []string) error
	}
}
