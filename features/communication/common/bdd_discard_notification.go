package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func (s *NotificationSuite) CurrentStaffDiscardsNotification(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).DiscardNotification(s.ContextWithToken(ctx, stepState.CurrentStaff.Token), &npb.DiscardNotificationRequest{
		NotificationId: stepState.Notification.NotificationId,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) CurrentStaffDeletesNotification(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).DeleteNotification(s.ContextWithToken(ctx, stepState.CurrentStaff.Token), &npb.DeleteNotificationRequest{
		NotificationId: stepState.Notification.NotificationId,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) NotificationIsDiscarded(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.CheckReturnStatusCode(ctx, "OK")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	noti := &entities.InfoNotification{}
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1`, strings.Join(database.GetFieldNames(noti), ","), noti.TableName())

	// use postgres account bypass rls
	err = database.Select(ctx, s.BobPostgresDBConn, query, database.Text(stepState.Notification.NotificationId)).ScanOne(noti)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if noti.Status.String != cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification status is %v but got %v", cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD.String(), noti.Status.String)
	}

	if noti.DeletedAt.Status == pgtype.Null {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification is deleted")
	}

	notiMsg := &entities.InfoNotificationMsg{}
	query = fmt.Sprintf(`SELECT %s FROM %s WHERE notification_msg_id = $1`, strings.Join(database.GetFieldNames(notiMsg), ","), notiMsg.TableName())
	err = database.Select(ctx, s.BobDBConn, query, noti.NotificationMsgID.String).ScanOne(notiMsg)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if notiMsg.DeletedAt.Status == pgtype.Null {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification msg is deleted")
	}

	if noti.QuestionnaireID.String != "" {
		qn := &entities.Questionnaire{}
		query = fmt.Sprintf(`SELECT %s FROM %s WHERE questionnaire_id = $1`, strings.Join(database.GetFieldNames(qn), ","), qn.TableName())
		err = database.Select(ctx, s.BobDBConn, query, noti.QuestionnaireID.String).ScanOne(qn)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if qn.DeletedAt.Status == pgtype.Null {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect questionnaire is deleted")
		}

		qnQuestion := &entities.QuestionnaireQuestion{}
		qnQuestions := entities.QuestionnaireQuestions{}
		query = fmt.Sprintf(`SELECT %s FROM %s WHERE questionnaire_id = $1 AND deleted_at IS NULL`, strings.Join(database.GetFieldNames(qnQuestion), ","), qnQuestion.TableName())
		err = database.Select(ctx, s.BobDBConn, query, noti.QuestionnaireID.String).ScanAll(&qnQuestions)
		if err != nil && err != pgx.ErrNoRows {
			return StepStateToContext(ctx, stepState), err
		}
		if len(qnQuestions) > 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification access path is deleted")
		}
	}

	notiAccessPath := &entities.InfoNotificationAccessPath{}
	notiAccessPaths := entities.InfoNotificationAccessPaths{}
	query = fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1 AND deleted_at IS NULL`, strings.Join(database.GetFieldNames(notiAccessPath), ","), notiAccessPath.TableName())
	err = database.Select(ctx, s.BobDBConn, query, database.Text(stepState.Notification.NotificationId)).ScanAll(&notiAccessPaths)
	if err != nil && err != pgx.ErrNoRows {
		return StepStateToContext(ctx, stepState), err
	}
	if len(notiAccessPaths) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification access path is deleted")
	}

	notiTag := &entities.InfoNotificationTag{}
	notiTags := entities.InfoNotificationsTags{}
	query = fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1 AND deleted_at IS NULL`, strings.Join(database.GetFieldNames(notiTag), ","), notiTag.TableName())
	err = database.Select(ctx, s.BobDBConn, query, database.Text(stepState.Notification.NotificationId)).ScanAll(&notiTags)
	if err != nil && err != pgx.ErrNoRows {
		return StepStateToContext(ctx, stepState), err
	}
	if len(notiTags) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect notification tags is deleted")
	}

	if stepState.Notification.TargetGroup.LocationFilter.Type == *consts.TargetGroupSelectTypeList.Enum() {
		countLocationFilter := 0
		queryCountLocationFilter := `SELECT count(*) FROM notification_location_filter WHERE notification_id = $1 AND deleted_at is NULL`
		err := s.BobDBConn.QueryRow(ctx, queryCountLocationFilter, stepState.Notification.NotificationId).Scan(&countLocationFilter)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if countLocationFilter != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected count location filter == 0, but got: %d", countLocationFilter)
		}
	}

	if stepState.Notification.TargetGroup.CourseFilter.Type == *consts.TargetGroupSelectTypeList.Enum() {
		countCourseFilter := 0
		queryCountCourseFilter := `SELECT count(*) FROM notification_course_filter WHERE notification_id = $1 AND deleted_at is NULL`
		err := s.BobDBConn.QueryRow(ctx, queryCountCourseFilter, stepState.Notification.NotificationId).Scan(&countCourseFilter)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if countCourseFilter != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected count course filter == 0, but got: %d", countCourseFilter)
		}
	}

	if stepState.Notification.TargetGroup.ClassFilter.Type == *consts.TargetGroupSelectTypeList.Enum() {
		countClassFilter := 0
		queryCountClassFilter := `SELECT count(*) FROM notification_class_filter WHERE notification_id = $1 AND deleted_at is NULL`
		err := s.BobDBConn.QueryRow(ctx, queryCountClassFilter, stepState.Notification.NotificationId).Scan(&countClassFilter)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if countClassFilter != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected count class filter == 0, but got: %d", countClassFilter)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
