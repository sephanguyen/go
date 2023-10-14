package communication

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/godogutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func (s *suite) hasCreatedAScheduledNotification(ctx context.Context, arg1 string) (context.Context, error) {
	var err error
	ctx, err = godogutil.MultiErrChain(
		ctx,
		s.fillsScheduledNotificationInformation,
		s.clicksButton, "",
		s.seesNewScheduledNotificationOnCMS,
	)

	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) hasOpenedEditorFullscreenDialogOfScheduledNotification(ctx context.Context, arg1 string) (context.Context, error) {
	// FE behavior
	return ctx, nil
}

func (s *suite) isAtPageOnCMS(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	// FE behavior
	return ctx, nil
}

func (s *suite) seesScheduledNotificationHasBeenSavedToDraftNotification(ctx context.Context, arg1 string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	st.notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT
	var err error
	ctx, err = godogutil.MultiErrChain(ctx,
		s.upsertNotification, st.notification,
		s.storeNotificationSuccessfully,
	)

	return StepStateToContext(ctx, st), err
}

func (s *suite) selectsStatus(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	// FE behavior
	return ctx, nil
}
func (s *suite) statusOfScheduledNotificationIsUpdatedToSent(ctx context.Context) (context.Context, error) {
	st := StepStateFromContext(ctx)
	st.notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SENT
	StepStateToContext(ctx, st)
	return s.storeNotificationSuccessfully(ctx)
}

func (s *suite) waitsForScheduledNotificationToBeSentOnTime(ctx context.Context, arg1 string) (context.Context, error) {
	st := StepStateFromContext(ctx)
	// plus 4s to wait to process send notification
	waitTime := time.Until(st.notification.ScheduledAt.AsTime()) + time.Duration(4*time.Second)
	time.Sleep(waitTime)
	return StepStateToContext(ctx, st), nil
}
