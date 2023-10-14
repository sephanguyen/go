package communication

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) cancelsToDiscard(ctx context.Context, arg1 string) (context.Context, error) {
	return ctx, nil
}

func (s *suite) discardNotification(ctx context.Context, notificationID string) error {
	_, err := ypb.NewNotificationModifierServiceClient(s.yasuoConn).DiscardNotification(contextWithToken(ctx, s.getToken(schoolAdmin)), &ypb.DiscardNotificationRequest{NotificationId: notificationID})
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) confirmsToDiscard(ctx context.Context, arg1 string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	err := s.discardNotification(ctx, st.notification.NotificationId)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, st), nil
}

func (s *suite) hasCreatedNotifications(ctx context.Context, role string, num int, notiTypeArg string) (context.Context, error) {
	var err error
	st := StepStateFromContext(ctx)
	switch notiTypeArg {
	case "draft":
		ctx, err = s.hasCreatedADraftNotification(ctx, role)
		if err != nil {
			return ctx, err
		}
		st.draftNotification = st.notification
	case "scheduled":
		ctx, err = s.hasCreatedAScheduledNotification(ctx, role)
		if err != nil {
			return ctx, err
		}
		st.scheduledNotification = st.notification
	default:
		return ctx, fmt.Errorf("not support notification type %s", notiTypeArg)
	}
	return StepStateToContext(ctx, st), nil
}

func (s *suite) hasOpenedEditorFullscreenDialogOfNotification(ctx context.Context, role, notiTypeArg string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	switch notiTypeArg {
	case "draft":
		st.notification = st.draftNotification
	case "scheduled":
		st.notification = st.scheduledNotification
	default:
		return ctx, fmt.Errorf("not support notification type %s", notiTypeArg)
	}
	return StepStateToContext(ctx, st), nil
}

func (s *suite) isDeletedNotification(ctx context.Context, notificationID string) (bool, error) {
	query := `SELECT count(*) FROM info_notifications WHERE notification_id = $1 AND deleted_at IS NOT NULL;`
	var cnt pgtype.Int8
	err := s.bobDB.QueryRow(ctx, query, database.Text(notificationID)).Scan(&cnt)
	if err != nil {
		return false, err
	}
	return cnt.Int == 1, nil
}

func (s *suite) seesNotificationHasBeenDeletedOnCMS(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	notificationID := st.notification.NotificationId
	isDeleted, err := s.isDeletedNotification(ctx, st.notification.NotificationId)
	if err != nil {
		return ctx, err
	}
	if !isDeleted {
		return ctx, fmt.Errorf("expect notification id %s is deleted but it is not", notificationID)
	}
	return StepStateToContext(ctx, st), nil
}

func (s *suite) stillSeesNotificationOnCMS(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	notificationID := st.notification.NotificationId
	isDeleted, err := s.isDeletedNotification(ctx, st.notification.NotificationId)
	if err != nil {
		return ctx, err
	}
	if isDeleted {
		return ctx, fmt.Errorf("expect notification id %s is not deleted but it is", notificationID)
	}
	return StepStateToContext(ctx, st), nil
}
