package entities

import (
	"time"

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
)

type ActivityLog struct {
	ID           pgtype.Text `sql:"activity_log_id,pk"`
	UserID       pgtype.Text `sql:"user_id"`
	ActionType   pgtype.Text `sql:"action_type"`
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	Payload      pgtype.JSONB
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text `sql:"resource_path"`
	RequestAt    pgtype.Timestamptz
	Status       pgtype.Text `sql:"status"`
	FinishedAt   pgtype.Timestamptz
}

func (a *ActivityLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"activity_log_id", "user_id", "action_type", "updated_at", "created_at", "deleted_at", "resource_path", "request_at", "payload", "status", "finished_at",
		}, []interface{}{
			&a.ID, &a.UserID, &a.ActionType, &a.UpdatedAt, &a.CreatedAt, &a.DeletedAt, &a.ResourcePath, &a.RequestAt, &a.Payload, &a.Status, &a.FinishedAt,
		}
}

func (a *ActivityLog) TableName() string {
	return "activity_logs"
}

func ToActivityLog(userID, actionType string, payload string, resourcePath string, requestAt time.Time, status string, finishedAt time.Time) (*ActivityLog, error) {
	a := &ActivityLog{}
	now := time.Now()

	err := multierr.Combine(
		a.ID.Set(idutil.ULIDNow()),
		a.UserID.Set(userID),
		a.ActionType.Set(actionType),
		a.Payload.Set(payload),
		a.CreatedAt.Set(now),
		a.UpdatedAt.Set(now),
		a.RequestAt.Set(requestAt),
		a.ResourcePath.Set(resourcePath),
		a.Status.Set(status),
		a.FinishedAt.Set(finishedAt),
	)

	return a, err
}
