package entities

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

const (
	LogActionTypeOnline                            = "online"
	LogActionTypeOffline                           = "offline"
	LogActionTypeDisableOrder                      = "disable_order"
	LogActionTypeCreateManualOrder                 = "create_manual_order"
	LogActionUsePromotionCode                      = "use_promotion_code"
	LogActionUseActivationCode                     = "use_activation_code"
	LogActionTransitionSuccessCodOrder             = "transition_success_cod_order"
	LogActionAddClassMember                        = "add_class_member"
	LogActionRemoveClassMember                     = "remove_class_member"
	LogActionTypeUpsertTopics                      = "upsert_topics"
	LogActionTypeUpsertLos                         = "upsert_los"
	LogActionTypeMergeSchool                       = "merge_school"
	LogActionTypeExtendSubscription                = "extend_subscription"
	LogActionTypeCreateNotification                = "create_notification"
	LogActionTypeAddTeacherToSchool                = "add_teacher_to_school"
	LogActionTypeRemoveTeacherFromSchoolFromSchool = "remove_teacher_from_school"
	LogActionTypeUpsertCourses                     = "upsert_courses"
	LogActionTypeDeleteCourses                     = "delete_courses"
	LogActionTypeUpsertChapters                    = "upsert_chapters"
	LogActionTypeDeleteChapters                    = "delete_chapters"
	LogActionTypeDeleteTopics                      = "delete_topics"
	LogActionTypeDeleteLos                         = "delete_los"
	LogActionTypePublish                           = "publish_streaming"
	LogActionTypeUnpublish                         = "unpublish_streaming"
)

type ActivityLog struct {
	ID         pgtype.Text `sql:"activity_log_id,pk"`
	UserID     pgtype.Text `sql:"user_id"`
	ActionType pgtype.Text `sql:"action_type"`
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	Payload    pgtype.JSONB
	DeletedAt  pgtype.Timestamptz
}

func (t *ActivityLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"activity_log_id", "user_id", "action_type", "updated_at", "created_at", "deleted_at", "payload",
		}, []interface{}{
			&t.ID, &t.UserID, &t.ActionType, &t.UpdatedAt, &t.CreatedAt, &t.DeletedAt, &t.Payload,
		}
}

func (t *ActivityLog) TableName() string {
	return "activity_logs"
}

type ActivityLogs []*ActivityLog

func (u *ActivityLogs) Add() database.Entity {
	e := &ActivityLog{}
	*u = append(*u, e)

	return e
}

func ToActivityLog(userID, actionType string, payload map[string]interface{}) (*ActivityLog, error) {
	a := &ActivityLog{}
	now := time.Now()
	marshaled, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal payload %w", err)
	}

	err = multierr.Combine(
		a.ID.Set(idutil.ULIDNow()),
		a.UserID.Set(userID),
		a.ActionType.Set(actionType),
		a.Payload.Set(string(marshaled)),
		a.CreatedAt.Set(now),
		a.UpdatedAt.Set(now),
		a.DeletedAt.Set(nil),
	)

	return a, err
}
