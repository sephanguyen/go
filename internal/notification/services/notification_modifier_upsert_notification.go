package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	"github.com/manabie-com/backend/internal/notification/services/validation"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationModifierService) UpsertNotification(ctx context.Context, req *npb.UpsertNotificationRequest) (*npb.UpsertNotificationResponse, error) {
	userInfo := golibs.UserInfoFromCtx(ctx)
	resourcePath := userInfo.ResourcePath
	userID := interceptors.UserIDFromContext(ctx)
	if userID == "" {
		// Support context from nats
		userID = userInfo.UserID
	}
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id doesn't exist in request")
	}

	err := validation.ValidateUpsertNotificationRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validateUpsertNotificationRequest: %v", err))
	}

	createdUserID := ""
	if req.Notification.NotificationId != "" {
		// update existed notification
		notificationEnt, err := svc.findEditableNotification(ctx, req.Notification.NotificationId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		createdUserID = notificationEnt.CreatedUserID.String
	}

	if req.Notification.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED {
		err := validation.ValidateScheduledNotification(req)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	var questionnaire *entities.Questionnaire
	questionnaireQuestions := make(entities.QuestionnaireQuestions, 0)

	if req.Questionnaire != nil {
		questionnaire, err = mappers.PbToQuestionnaireEnt(req.Questionnaire)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "cannot convert PbToQuestionnaireEnt")
		}

		questionnaireQuestions, err = mappers.PbToQuestionnaireQuestionEnts(req.Questionnaire)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "cannot convert PbToQuestionnaireQuestionEnts")
		}
	}

	// upsert infor notification and infor notification msgs table
	infoNotificationMsg, err := mappers.PbToInfoNotificationMsgEnt(req.Notification.Message)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot convert PbToInfoNotificationMsgEnt")
	}

	infoNotification, err := mappers.PbToInfoNotificationEnt(req.Notification)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot convert PbToInfoNotificationEnt")
	}

	// Assign receiver_names for notification
	infoNotification, err = svc.DataRetentionService.AssignIndividualRetentionNamesForNotification(ctx, svc.DB, infoNotification)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot exec DataRetentionService.AssignIndividualRetentionNamesForNotification: %v", err)
	}

	// set latest editor id, In the nats request case, the UserID is empty
	// in nats case, userId is empty so occurred Unique Constraint Violated in SQL error of fk_editor_id
	if len(userID) > 0 {
		_ = infoNotification.EditorID.Set(userID)

		// Only set created_user_id when creating, not updating
		if infoNotification.NotificationID.String == "" {
			_ = infoNotification.CreatedUserID.Set(userID)
		} else {
			_ = infoNotification.CreatedUserID.Set(createdUserID)
		}
	} else {
		_ = infoNotification.EditorID.Set(nil)
		_ = infoNotification.CreatedUserID.Set(nil)
	}

	err = infoNotification.Owner.Set(resourcePath)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot set notification owner")
	}

	// out of range type -> explicitly set notification type to NotificationType_NOTIFICATION_TYPE_COMPOSED
	// we have a handling from nats here so do not explicitly set notification type to NotificationType_NOTIFICATION_TYPE_COMPOSED
	if req.Notification.Type < cpb.NotificationType_NOTIFICATION_TYPE_NONE || req.Notification.Type > cpb.NotificationType_NOTIFICATION_TYPE_NATS_ASYNC {
		err = infoNotification.Type.Set(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "cannot explicitly set notification type: NotificationType_NOTIFICATION_TYPE_COMPOSED")
		}
	}

	// explicitly set notification event to NotificationEvent_NOTIFICATION_EVENT_NONE
	err = infoNotification.Event.Set(cpb.NotificationEvent_NOTIFICATION_EVENT_NONE.String())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cannot explicitly set notification event: NotificationEvent_NOTIFICATION_EVENT_NONE")
	}

	// attach tags to notification
	tagIDs := []string{}
	if len(req.TagIds) > 0 {
		// check & remove duplicate tag id
		tagKeys := make(map[string]bool)
		for _, tagID := range req.TagIds {
			if _, ok := tagKeys[tagID]; !ok {
				tagKeys[tagID] = true
				tagIDs = append(tagIDs, tagID)
			}
		}

		// check exist tag IDs
		isExist, err := svc.TagRepo.CheckTagIDsExist(ctx, svc.DB, database.TextArray(tagIDs))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("UpsertNotification.CheckTagIDsExist: %v", err))
		}
		if !isExist {
			return nil, status.Error(codes.InvalidArgument, "some tags do not exist")
		}
	}

	err = database.ExecInTxWithRetry(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		return svc.upsertNotification(ctx, tx, req, infoNotificationMsg, infoNotification, questionnaire, questionnaireQuestions, tagIDs)
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ExecInTxWithRetry: %v", err))
	}

	resp := &npb.UpsertNotificationResponse{
		NotificationId: infoNotification.NotificationID.String,
	}

	return resp, nil
}

func (svc *NotificationModifierService) upsertNotification(ctx context.Context, tx pgx.Tx, req *npb.UpsertNotificationRequest, infoNotificationMsg *entities.InfoNotificationMsg, infoNotification *entities.InfoNotification, questionnaire *entities.Questionnaire, questionnaireQuestions entities.QuestionnaireQuestions, tagIDs []string) error {
	// upload content change
	generatedURL, _ := generateUploadURL(svc.StorageConfig.Endpoint, svc.StorageConfig.Bucket, req.Notification.Message.Content.Rendered)

	uploadedURL, err := svc.UploadHTMLContent(ctx, req.Notification.Message.Content.Rendered)
	if err != nil {
		return fmt.Errorf("err updateContent: %w", err)
	}
	if uploadedURL != generatedURL {
		return fmt.Errorf("err updateContent: url return does not match")
	}

	err = svc.upsertNotificationQuestionnaire(ctx, tx, infoNotification, questionnaire, questionnaireQuestions)
	if err != nil {
		return fmt.Errorf("svc.upsertNotificationQuestionnaire: %v", err)
	}

	infoNotificationMsg.Content = database.JSONB(&entities.RichText{
		Raw:         req.Notification.Message.Content.Raw,
		RenderedURL: generatedURL,
	})

	err = svc.InfoNotificationMsgRepo.Upsert(ctx, tx, infoNotificationMsg)
	if err != nil {
		return fmt.Errorf("svc.InfoNotificationMsgRepo.Upsert: %v", err)
	}

	_ = infoNotification.NotificationMsgID.Set(infoNotificationMsg.NotificationMsgID)
	notificationID, err := svc.InfoNotificationRepo.Upsert(ctx, tx, infoNotification)
	if err != nil {
		return fmt.Errorf("svc.InfoNotificationRepo.Upsert: %v", err)
	}
	svc.RecordNotificationCreated(1)

	err = svc.attachTagsToNotification(ctx, tx, notificationID, tagIDs)
	if err != nil {
		return fmt.Errorf("svc.attachTagsToNotification: %v", err)
	}

	selectedLocationIDs := make([]string, 0)
	if req.Notification.TargetGroup.LocationFilter != nil && req.Notification.TargetGroup.LocationFilter.Type == consts.TargetGroupSelectTypeList {
		selectedLocationIDs = req.Notification.TargetGroup.LocationFilter.LocationIds
	}
	locationIDs, err := svc.upsertNotificationAccessPath(ctx, tx, infoNotification.NotificationID.String, selectedLocationIDs, infoNotification.CreatedUserID.String)
	if err != nil {
		zapLogger := ctxzap.Extract(ctx)
		zapLogger.Sugar().Errorf("svc.upsertNotificationAccessPath: %v", err)

		return fmt.Errorf("svc.upsertNotificationAccessPath: %v", err)
	}

	// // In case location == SELECTE ALL or NONE, we save the granted location of the current user to target_group.location_filter.location_ids
	// to support getting the location when sending notifications without being affected by Access Control
	targetGroupEnt := &entities.InfoNotificationTarget{}
	err = infoNotification.TargetGroups.AssignTo(targetGroupEnt)
	if err != nil {
		return fmt.Errorf("cannot set target group before update it: %v", err)
	}
	if targetGroupEnt.LocationFilter.Type != consts.TargetGroupSelectTypeList.String() {
		targetGroupEnt.LocationFilter.LocationIDs = locationIDs
		err = svc.InfoNotificationRepo.UpdateNotification(ctx, tx, database.Text(notificationID), map[string]interface{}{
			"target_groups": database.JSONB(targetGroupEnt),
		})

		if err != nil {
			return fmt.Errorf("svc.UpdateNotification with target_groups: %v", err)
		}
	}

	// upsert data for notification filter table support for Advanced Filter
	err = svc.upsertNotificationFilterInfo(ctx, tx, targetGroupEnt, notificationID)
	if err != nil {
		return fmt.Errorf("err svc.upsertNotificationFilterInfo: %v", err)
	}

	return nil
}

func (svc *NotificationModifierService) upsertNotificationQuestionnaire(ctx context.Context, tx pgx.Tx, infoNotification *entities.InfoNotification, questionnaire *entities.Questionnaire, questionnaireQuestions entities.QuestionnaireQuestions) error {
	if questionnaire != nil {
		// Update notification case
		if infoNotification.NotificationID.String != "" {
			currentQuestionnaireID, err := svc.findQuestionnaireID(ctx, tx, infoNotification.NotificationID.String)
			if err != nil {
				return fmt.Errorf("svc.findQuestionnaireID: %v", err)
			}

			if currentQuestionnaireID != "" {
				_ = questionnaire.QuestionnaireID.Set(currentQuestionnaireID)
			}

			// Soft delete questinnaire question if exist in case update notification
			err = svc.QuestionnaireQuestionRepo.SoftDelete(ctx, tx, []string{currentQuestionnaireID})
			if err != nil {
				return fmt.Errorf("svc.QuestionnaireQuestionRepo.SoftDelete: %v", err)
			}
		}

		err := svc.QuestionnaireRepo.Upsert(ctx, tx, questionnaire)
		if err != nil {
			return fmt.Errorf("svc.QuestionnaireRepo.Upsert: %v", err)
		}

		_ = infoNotification.QuestionnaireID.Set(questionnaire.QuestionnaireID.String)
		for _, question := range questionnaireQuestions {
			_ = question.QuestionnaireID.Set(questionnaire.QuestionnaireID.String)
		}

		err = svc.QuestionnaireQuestionRepo.BulkForceUpsert(ctx, tx, questionnaireQuestions)
		if err != nil {
			return fmt.Errorf("svc.QuestionnaireQuestionRepo.BulkUpsert: %v", err)
		}
	} else if infoNotification.NotificationID.String != "" {
		// For case update notification and don't have any questionnaire in request:
		// We are soft delete all questionnaire and questionnaire questions in db if exist
		oldQuestionnaireID, err := svc.findQuestionnaireID(ctx, tx, infoNotification.NotificationID.String)
		if err != nil {
			return fmt.Errorf("svc.GetQuestionnaireID: %v", err)
		}

		if oldQuestionnaireID != "" {
			err = infoNotification.QuestionnaireID.Set(nil)
			if err != nil {
				return fmt.Errorf("infoNotification.QuestionnaireID.Set: %v", err)
			}

			err = svc.QuestionnaireRepo.SoftDelete(ctx, tx, []string{oldQuestionnaireID})
			if err != nil {
				return fmt.Errorf("svc.QuestionnaireRepo.SoftDelete: %v", err)
			}

			err = svc.QuestionnaireQuestionRepo.SoftDelete(ctx, tx, []string{oldQuestionnaireID})
			if err != nil {
				return fmt.Errorf("svc.QuestionnaireQuestionRepo.SoftDelete: %v", err)
			}
		}
	}
	return nil
}

func (svc *NotificationModifierService) attachTagsToNotification(ctx context.Context, tx pgx.Tx, notificationID string, tagIDs []string) error {
	// find all associated info_notification_tags by NotificationID
	ifntTagMap, err := svc.InfoNotificationTagRepo.GetByNotificationIDs(ctx, tx, database.TextArray([]string{notificationID}))
	if err != nil {
		return fmt.Errorf("svc.InfoNotificationTagRepo.GetByNotificationIDs: %v", err)
	}
	// Notification has some Tags attached in DB
	if inftTags, ok := ifntTagMap[notificationID]; ok {
		attachedTagIDs := []string{}
		attachedTagMap := make(map[string]*entities.InfoNotificationTag)
		for _, attachedTag := range inftTags {
			attachedTagIDs = append(attachedTagIDs, attachedTag.TagID.String)
			attachedTagMap[attachedTag.TagID.String] = attachedTag
		}

		// compare current attachedTagIDs with request TagIDs
		arrInsert, arrRemove := utils.CompareTagArrays(tagIDs, attachedTagIDs)

		if len(arrInsert) > 0 {
			insEnts := []*entities.InfoNotificationTag{}
			for _, insTagID := range arrInsert {
				ifntTag := &entities.InfoNotificationTag{}
				now := time.Now()
				var err error
				if val, ok := attachedTagMap[insTagID]; ok { // case Update
					// TODO: write a util func to help assign fields of two same structs type
					err = multierr.Combine(
						ifntTag.NotificationTagID.Set(val.NotificationTagID),
						ifntTag.NotificationID.Set(val.NotificationID),
						ifntTag.TagID.Set(val.TagID),
						ifntTag.CreatedAt.Set(val.CreatedAt),
						ifntTag.UpdatedAt.Set(now),
						ifntTag.DeletedAt.Set(nil), // if info_notifications_tags is already created and soft deleted, Upsert will enable that record again
					)
				} else { // insert
					err = multierr.Combine(
						ifntTag.NotificationTagID.Set(idutil.ULIDNow()),
						ifntTag.NotificationID.Set(notificationID),
						ifntTag.TagID.Set(insTagID),
						ifntTag.CreatedAt.Set(now),
						ifntTag.UpdatedAt.Set(now),
						ifntTag.DeletedAt.Set(nil),
					)
				}
				if err != nil {
					return fmt.Errorf("upsertNotification.multierr.Combine: %v", err)
				}
				insEnts = append(insEnts, ifntTag)
			}
			err := svc.InfoNotificationTagRepo.BulkUpsert(ctx, tx, insEnts)
			if err != nil {
				return fmt.Errorf("svc.InfoNotificationTagRepo.BulkUpsert: %v", err)
			}
		}
		if len(arrRemove) > 0 {
			rmvEntIDs := []string{}
			for _, rmvTagID := range arrRemove {
				if val, ok := attachedTagMap[rmvTagID]; ok {
					rmvEntIDs = append(rmvEntIDs, val.NotificationTagID.String)
				} else {
					return fmt.Errorf("svc.attachTagsToNotification: removing non associated tag ID %s", rmvTagID)
				}
			}
			filterDeleteNotiTag := repositories.NewSoftDeleteNotificationTagFilter()
			_ = filterDeleteNotiTag.NotificationTagIDs.Set(rmvEntIDs)
			err := svc.InfoNotificationTagRepo.SoftDelete(ctx, tx, filterDeleteNotiTag)
			if err != nil {
				return fmt.Errorf("svc.InfoNotificationTagRepo.SoftDelete: %v", err)
			}
		}
	} else {
		insEnts := []*entities.InfoNotificationTag{}
		for _, id := range tagIDs {
			ifntTag := &entities.InfoNotificationTag{}
			now := time.Now()
			err := multierr.Combine(
				ifntTag.NotificationTagID.Set(idutil.ULIDNow()),
				ifntTag.NotificationID.Set(notificationID),
				ifntTag.TagID.Set(id),
				ifntTag.CreatedAt.Set(now),
				ifntTag.UpdatedAt.Set(now),
				ifntTag.DeletedAt.Set(nil),
			)
			if err != nil {
				return fmt.Errorf("upsertNotification.multierr.Combine: %v", err)
			}
			insEnts = append(insEnts, ifntTag)
		}
		err := svc.InfoNotificationTagRepo.BulkUpsert(ctx, tx, insEnts)
		if err != nil {
			return fmt.Errorf("svc.InfoNotificationTagRepo.BulkUpsert: %v", err)
		}
	}

	return nil
}

func (svc *NotificationModifierService) findEditableNotification(ctx context.Context, notificationID string) (*entities.InfoNotification, error) {
	noti, err := svc.findNotificationByID(ctx, svc.DB, notificationID)
	if err != nil {
		isDeleted, err := svc.InfoNotificationRepo.IsNotificationDeleted(ctx, svc.DB, database.Text(notificationID))

		if err != nil {
			return nil, fmt.Errorf("svc.IsDeletedNotification: %w", err)
		}
		if isDeleted {
			return nil, fmt.Errorf("the notification has been deleted, you can no longer edit this notification")
		}
	}

	if noti.Status.String == cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String() {
		return nil, fmt.Errorf("the notification has been sent, you can no longer edit this notification")
	}

	if noti.Status.String != cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String() && noti.Status.String != cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String() {
		return nil, fmt.Errorf("notification %s has status %s and not discardable", noti.NotificationID.String, noti.Status.String)
	}

	return noti, nil
}

func (svc *NotificationModifierService) findQuestionnaireID(ctx context.Context, db database.Ext, notificationID string) (string, error) {
	noti, err := svc.findNotificationByID(ctx, db, notificationID)

	if err != nil {
		return "", fmt.Errorf("svc.FindNotificationByID: %w", err)
	}

	return noti.QuestionnaireID.String, nil
}
