package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func NewNotificationReaderService(db database.Ext, env string) *NotificationReaderService {
	return &NotificationReaderService{
		DB:                            db,
		Env:                           env,
		NotificationAudienceRetriever: domain.NewAudienceRetrieverService(env),

		InfoNotificationRepo: &repositories.InfoNotificationRepo{
			InfoNotificationSQLBuilder: repositories.InfoNotificationSQLBuilder{},
		},
		InfoNotificationMsgRepo:        &repositories.InfoNotificationMsgRepo{},
		UserInfoNotificationRepo:       &repositories.UsersInfoNotificationRepo{},
		QuestionnaireRepo:              &repositories.QuestionnaireRepo{},
		InfoNotificationTagRepo:        &repositories.InfoNotificationTagRepo{},
		NotificationLocationFilterRepo: &repositories.NotificationLocationFilterRepo{},
		NotificationCourseFilterRepo:   &repositories.NotificationCourseFilterRepo{},
		NotificationClassFilterRepo:    &repositories.NotificationClassFilterRepo{},
	}
}

type NotificationReaderService struct {
	DB  database.Ext
	Env string

	// Depend services
	NotificationAudienceRetriever interface {
		FindGroupAudiencesWithPaging(ctx context.Context, db database.QueryExecer, notiID string, targetGroup *entities.InfoNotificationTarget, keyword string, includeUserIDs []string, limit, offset int) ([]*entities.Audience, uint32, error)
		FindDraftAudiencesWithPaging(ctx context.Context, db database.QueryExecer, notiID string, targetGroup *entities.InfoNotificationTarget, genericReceiverIds, groupExcludedGenericReceiverIds []string, limit, offset int) ([]*entities.Audience, uint32, error)
	}

	// Repositories
	InfoNotificationRepo interface {
		Find(ctx context.Context, db database.QueryExecer, filter *repositories.FindNotificationFilter) (entities.InfoNotifications, error)
		CountTotalNotificationForStatus(ctx context.Context, db database.QueryExecer, filter *repositories.FindNotificationFilter) (map[string]uint32, error)
	}

	InfoNotificationMsgRepo interface {
		GetByIDs(ctx context.Context, db database.QueryExecer, notiMsgIDs pgtype.TextArray) (entities.InfoNotificationMsgs, error)
		GetByNotificationIDs(ctx context.Context, db database.QueryExecer, notiIDs pgtype.TextArray) (map[string]*entities.InfoNotificationMsg, error)
		GetIDsByTitle(ctx context.Context, db database.QueryExecer, title pgtype.Text) ([]string, error)
	}

	UserInfoNotificationRepo interface {
		Find(ctx context.Context, db database.QueryExecer, filter repositories.FindUserNotificationFilter) (entities.UserInfoNotifications, error)
		CountByStatus(ctx context.Context, db database.QueryExecer, userID pgtype.Text, status pgtype.Text) (int, int, error)
		GetNotificationIDWithFullyQnStatus(ctx context.Context, db database.QueryExecer, notificationIDs pgtype.TextArray, status pgtype.Text) ([]string, error)
	}

	QuestionnaireRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, id string) (*entities.Questionnaire, error)
		FindQuestionsByQnID(ctx context.Context, db database.QueryExecer, id string) (entities.QuestionnaireQuestions, error)
		FindUserAnswers(ctx context.Context, db database.QueryExecer, filter *repositories.FindUserAnswersFilter) (entities.QuestionnaireUserAnswers, error)
		FindQuestionnaireResponders(ctx context.Context, db database.QueryExecer, filter *repositories.FindQuestionnaireRespondersFilter) (uint32, []*repositories.QuestionnaireResponder, error)
		FindQuestionnaireCSVResponders(ctx context.Context, db database.QueryExecer, questionnaireID string) ([]*repositories.QuestionnaireCSVResponder, error)
	}

	InfoNotificationTagRepo interface {
		GetByNotificationIDs(ctx context.Context, db database.QueryExecer, notificationIDs pgtype.TextArray) (map[string]entities.InfoNotificationsTags, error)
		GetNotificationIDsByTagIDs(ctx context.Context, db database.QueryExecer, tagIDs pgtype.TextArray) ([]string, error)
	}

	NotificationLocationFilterRepo interface {
		GetNotificationIDsByLocationIDs(ctx context.Context, db database.QueryExecer, notificationIDs, locationIDs pgtype.TextArray) ([]string, error)
	}

	NotificationCourseFilterRepo interface {
		GetNotificationIDsByCourseIDs(ctx context.Context, db database.QueryExecer, notificationIDs, courseIDs pgtype.TextArray) ([]string, error)
	}

	NotificationClassFilterRepo interface {
		GetNotificationIDsByClassIDs(ctx context.Context, db database.QueryExecer, notificationIDs, classIDs pgtype.TextArray) ([]string, error)
	}
}

func (svc *NotificationReaderService) findSentNotification(ctx context.Context, notificationID string) (*entities.InfoNotification, error) {
	filter := repositories.NewFindNotificationFilter()
	err := filter.NotiIDs.Set([]string{notificationID})
	if err != nil {
		return nil, fmt.Errorf("cannot set notification id for FindNotificationFilter")
	}
	err = filter.Status.Set([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()})
	if err != nil {
		return nil, fmt.Errorf("cannot set notification status for FindNotificationFilter")
	}

	es, err := svc.InfoNotificationRepo.Find(ctx, svc.DB, filter)
	if err != nil {
		return nil, fmt.Errorf("InfoNotificationRepo.Find: %v", err)
	}
	if len(es) == 0 {
		return nil, fmt.Errorf("InfoNotificationRepo.Find: can not find notification with id %v", notificationID)
	}
	noti := es[0]

	return noti, nil
}

func (svc *NotificationReaderService) findNotificationMsg(ctx context.Context, notificationMsgID string) (*entities.InfoNotificationMsg, error) {
	es, err := svc.InfoNotificationMsgRepo.GetByIDs(ctx, svc.DB, database.TextArray([]string{notificationMsgID}))
	if err != nil {
		return nil, fmt.Errorf("InfoNotificationMsgRepo.GetByIDs: %v", err)
	}
	if len(es) == 0 {
		return nil, fmt.Errorf("InfoNotificationMsgRepo.GetByIDs: can not find notification message with id %v", notificationMsgID)
	}
	notiMsg := es[0]

	return notiMsg, nil
}

func (svc *NotificationReaderService) findUserNotification(ctx context.Context, userID string, paging *cpb.Paging, isImportantOnly bool) (entities.UserInfoNotifications, error) {
	filter := repositories.FindUserNotificationFilter{
		UserNotificationIDs: pgtype.TextArray{Status: pgtype.Null},
		UserIDs:             database.TextArray([]string{userID}),
		NotiIDs:             pgtype.TextArray{Status: pgtype.Null},
		UserStatus:          pgtype.TextArray{Status: pgtype.Null},
		Limit:               database.Int8(int64(paging.GetLimit())),
		OffsetTime:          database.Timestamptz(paging.GetOffsetCombined().GetOffsetTime().AsTime()),
		OffsetText:          database.Text(paging.GetOffsetCombined().GetOffsetString()),
		StudentID:           pgtype.Text{Status: pgtype.Null},
		ParentID:            pgtype.Text{Status: pgtype.Null},
		IsImportant:         pgtype.Bool{Status: pgtype.Null},
	}

	var err error
	if paging.GetOffsetCombined().GetOffsetTime() == nil {
		err = multierr.Append(err, filter.OffsetTime.Set(nil))
	}
	if paging.GetOffsetCombined().GetOffsetString() == "" {
		err = multierr.Append(err, filter.OffsetText.Set(nil))
	}

	if isImportantOnly {
		err = multierr.Append(err, filter.IsImportant.Set(true))
	} else {
		err = multierr.Append(err, filter.IsImportant.Set(nil))
	}
	if err != nil {
		return nil, fmt.Errorf("multierr on set filter: %v", err)
	}

	es, err := svc.UserInfoNotificationRepo.Find(ctx, svc.DB, filter)
	if err != nil {
		return nil, fmt.Errorf("UserInfoNotificationRepo.Find: %v", err)
	}
	return es, nil
}
