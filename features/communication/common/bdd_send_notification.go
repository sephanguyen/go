package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	noti_repos "github.com/manabie-com/backend/internal/notification/repositories"
	userEntity "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	pbu "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/bxcodec/faker/v3/support/slice"
)

func (s *NotificationSuite) CurrentStaffSendNotification(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &npb.SendNotificationRequest{
		NotificationId: stepState.Notification.NotificationId,
	}
	var err error
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response, stepState.ResponseErr = npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn).SendNotification(s.ContextWithToken(ctx, stepState.CurrentStaff.Token), stepState.Request.(*npb.SendNotificationRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *NotificationSuite) NotificationMgmtMustSendNotificationToUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.CheckReturnStatusCode(ctx, "OK")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return s.checkUserNotifications(ctx)
}

func (s *NotificationSuite) checkUserNotifications(ctx context.Context) (context.Context, error) {
	var (
		expectedUserAfterSend           []string
		expectedIndividualUserAfterSend []string
		err                             error
	)
	stepState := StepStateFromContext(ctx)
	if len(stepState.Notification.ReceiverIds) > 0 {
		// Only NATS uses this
		ctx, expectedUserAfterSend, expectedIndividualUserAfterSend, err = s.GetSendIndividualStudentWithUserGroups(ctx)
		if err != nil {
			return ctx, err
		}
	} else {
		// Composed notification will use this
		ctx, expectedUserAfterSend, expectedIndividualUserAfterSend, err = s.GetSendUserIDsOfComposedNotification(ctx)
		if err != nil {
			return ctx, err
		}
	}

	userNotifications, err := s.GetSendUsersFromDB(ctx)
	if err != nil {
		return ctx, err
	}

	ctx, err = s.checkUserInfoNotificationResponse(ctx, expectedUserAfterSend, expectedIndividualUserAfterSend, userNotifications)

	return ctx, err
}

// This is uses for NATS
func (s *NotificationSuite) GetSendIndividualStudentWithUserGroups(ctx context.Context) (context.Context, []string, []string, error) {
	stepState := StepStateFromContext(ctx)

	notificationSentAt, err := s.getNotificationSentAt(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	targetGroup := stepState.Notification.GetTargetGroup()
	audienceRepo := &noti_repos.AudienceRepo{}

	filter := noti_repos.NewFindIndividualAudienceFilter()
	if len(targetGroup.LocationFilter.LocationIds) == 0 {
		infoNotiAccessPathRepo := &noti_repos.InfoNotificationAccessPathRepo{}
		locationRepo := &noti_repos.LocationRepo{}
		notificationLocationIDs, err := infoNotiAccessPathRepo.GetLocationIDsByNotificationID(ctx, s.BobDBConn, stepState.Notification.NotificationId)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, nil, fmt.Errorf("failed GetNotificationID: %v", err)
		}
		locationIDs, err := locationRepo.GetLowestLocationIDsByIDs(ctx, s.BobDBConn, notificationLocationIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, nil, fmt.Errorf("failed GetNotificationID: %v", err)
		}

		_ = filter.LocationIDs.Set(locationIDs)
	} else {
		_ = filter.LocationIDs.Set(stepState.Notification.TargetGroup.LocationFilter.LocationIds)
	}
	_ = filter.UserIDs.Set(stepState.Notification.ReceiverIds)

	individualStudents, err := audienceRepo.FindIndividualAudiencesByFilter(ctx, s.BobDBConn, filter)
	if err != nil {
		return ctx, nil, nil, fmt.Errorf("audienceRepo.FindStudentAudiencesByFilter: %v", err)
	}

	sendIndividualUserIDs := make([]string, 0)
	studentIDs := []string{}
	for _, student := range individualStudents {
		studentIDs = append(studentIDs, student.StudentID.String)
	}

	studentIDs = golibs.GetUniqueElementStringArray(studentIDs)

	sendUserIDs := make([]string, 0)
	for _, group := range stepState.Notification.TargetGroup.UserGroupFilter.UserGroups {
		switch group {
		case cpb.UserGroup_USER_GROUP_STUDENT:
			sendUserIDs = append(sendUserIDs, studentIDs...)

		case cpb.UserGroup_USER_GROUP_PARENT:
			studentParents, err := s.getStudentParentsWithCreatedAt(ctx, studentIDs, *notificationSentAt)
			if err != nil {
				return ctx, nil, nil, fmt.Errorf("getStudentParentsWithCreatedAt: %v", err)
			}

			totalParentIDs := make([]string, 0, len(studentParents))
			for _, sp := range studentParents {
				totalParentIDs = append(totalParentIDs, sp.ParentID.String)
				if slice.Contains(sendIndividualUserIDs, sp.StudentID.String) {
					sendIndividualUserIDs = append(sendIndividualUserIDs, sp.ParentID.String)
				}
			}

			sendUserIDs = append(sendUserIDs, totalParentIDs...)
		}
	}

	return StepStateToContext(ctx, stepState), sendUserIDs, sendIndividualUserIDs, nil
}

// This is uses for Composed
func (s *NotificationSuite) GetSendUserIDsOfComposedNotification(ctx context.Context) (context.Context, []string, []string, error) {
	stepState := StepStateFromContext(ctx)

	noneSelectType := cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String()
	allSelectType := cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL.String()
	listSelectType := cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST.String()

	var err error
	targetGroup := stepState.Notification.GetTargetGroup()
	audienceRepo := &noti_repos.AudienceRepo{}

	filter := noti_repos.NewFindGroupAudienceFilter()
	if len(targetGroup.LocationFilter.LocationIds) == 0 {
		locationIDs, err := s.GetGrantedLocations(ctx, stepState.Notification.NotificationId)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, nil, fmt.Errorf("failed GetGrantedLocations: %v", err)
		}

		_ = filter.LocationIDs.Set(locationIDs)
	} else {
		_ = filter.LocationIDs.Set(stepState.Notification.TargetGroup.LocationFilter.LocationIds)
	}

	switch targetGroup.CourseFilter.Type.String() {
	case noneSelectType:
		_ = filter.CourseSelectType.Set(noneSelectType)
	case allSelectType:
		_ = filter.CourseSelectType.Set(allSelectType)
	case listSelectType:
		_ = filter.CourseSelectType.Set(listSelectType)
		_ = filter.CourseIDs.Set(stepState.Notification.TargetGroup.CourseFilter.CourseIds)
	}

	switch targetGroup.ClassFilter.Type.String() {
	case noneSelectType:
		_ = filter.ClassSelectType.Set(noneSelectType)
	case allSelectType:
		_ = filter.ClassSelectType.Set(allSelectType)
	case listSelectType:
		_ = filter.ClassSelectType.Set(listSelectType)
		_ = filter.ClassIDs.Set(stepState.Notification.TargetGroup.ClassFilter.ClassIds)
	}

	switch targetGroup.GradeFilter.Type.String() {
	case noneSelectType:
		_ = filter.GradeSelectType.Set(noneSelectType)
	case allSelectType:
		_ = filter.GradeSelectType.Set(allSelectType)
	case listSelectType:
		_ = filter.GradeSelectType.Set(listSelectType)
		_ = filter.GradeIDs.Set(stepState.Notification.TargetGroup.GradeFilter.GradeIds)
	}

	switch targetGroup.SchoolFilter.Type.String() {
	case noneSelectType:
		_ = filter.SchoolSelectType.Set(noneSelectType)
	case allSelectType:
		_ = filter.SchoolSelectType.Set(allSelectType)
	case listSelectType:
		_ = filter.SchoolSelectType.Set(listSelectType)
		_ = filter.SchoolIDs.Set(stepState.Notification.TargetGroup.SchoolFilter.SchoolIds)
	}

	_ = filter.UserGroups.Set(stepState.Notification.TargetGroup.UserGroupFilter.UserGroups)
	_ = filter.ExcludeUserIds.Set(stepState.Notification.ExcludedGenericReceiverIds)
	_ = filter.StudentEnrollmentStatus.Set(pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String())

	audiences, err := audienceRepo.FindGroupAudiencesByFilter(ctx, s.BobDBConn, filter, noti_repos.NewFindAudienceOption())
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, nil, fmt.Errorf("failed FindGroupAudiencesByFilter: %v", err)
	}

	audienceIDs := []string{}
	for _, audience := range audiences {
		audienceIDs = append(audienceIDs, audience.UserID.String)
		childIDs := database.FromTextArray(audience.ChildIDs)
		if len(childIDs) > 0 {
			_ = audience.StudentID.Set(childIDs[0])
			if len(childIDs) > 1 {
				for id := range childIDs {
					if id == 0 {
						continue
					}
					audienceIDs = append(audienceIDs, audience.UserID.String)
				}
			}
		}
	}

	var individualAudiences []*entities.Audience
	if len(stepState.Notification.GenericReceiverIds) > 0 {
		individualFilter := noti_repos.NewFindIndividualAudienceFilter()
		_ = individualFilter.UserIDs.Set(stepState.Notification.GenericReceiverIds)
		_ = individualFilter.EnrollmentStatuses.Set([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()})
		locationIDs, err := s.GetGrantedLocations(ctx, stepState.Notification.NotificationId)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, nil, fmt.Errorf("failed GetGrantedLocations: %v", err)
		}

		_ = individualFilter.LocationIDs.Set(locationIDs)
		individualAudiences, err = audienceRepo.FindIndividualAudiencesByFilter(ctx, s.BobDBConn, individualFilter)
		if err != nil {
			return ctx, nil, nil, fmt.Errorf("audienceRepo.FindIndividualAudiencesByFilter: %v", err)
		}
	}

	sendIndividualUserIDs := make([]string, 0)
	for _, audience := range individualAudiences {
		audienceIDs = append(audienceIDs, audience.UserID.String)
		sendIndividualUserIDs = append(sendIndividualUserIDs, audience.UserID.String)
	}

	return StepStateToContext(ctx, stepState), audienceIDs, sendIndividualUserIDs, nil
}

func (s *NotificationSuite) getNotificationSentAt(ctx context.Context) (*time.Time, error) {
	stepState := StepStateFromContext(ctx)

	var notificationSentAt time.Time
	query := "SELECT sent_at FROM info_notifications WHERE notification_id = $1"
	err := s.BobDBConn.QueryRow(ctx, query, stepState.Notification.NotificationId).Scan(&notificationSentAt)
	if err != nil {
		return nil, fmt.Errorf("getNotificationCreatedTime: %v", err)
	}
	return &notificationSentAt, nil
}

func (s *NotificationSuite) getStudentParentsWithCreatedAt(ctx context.Context, studentIDs []string, createdAt time.Time) ([]*userEntity.StudentParent, error) {
	sp := &userEntity.StudentParent{}
	query := fmt.Sprintf(`
		SELECT %s FROM %s WHERE student_id = ANY($1) AND deleted_at IS NULL AND created_at <= $2`, strings.Join(database.GetFieldNames(sp), ","), sp.TableName())

	rows, err := s.BobDBConn.Query(ctx, query, database.TextArray(studentIDs), database.Timestamptz(createdAt))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	studentParents := make([]*userEntity.StudentParent, 0)
	for rows.Next() {
		e := &userEntity.StudentParent{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		studentParents = append(studentParents, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return studentParents, nil
}

func (s *NotificationSuite) GetSendUsersFromDB(ctx context.Context) ([]*entities.UserInfoNotification, error) {
	stepState := StepStateFromContext(ctx)

	userInfoNotifications := make([]*entities.UserInfoNotification, 0)
	e := &entities.UserInfoNotification{}
	fields := database.GetFieldNames(e)
	queryGetUserInfoNotification := fmt.Sprintf(`SELECT %s FROM %s WHERE notification_id = $1 AND deleted_at IS NULL`, strings.Join(fields, ","), e.TableName())

	rows, err := s.BobDBConn.Query(ctx, queryGetUserInfoNotification, stepState.Notification.NotificationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := &entities.UserInfoNotification{}
		val := database.GetScanFields(e, database.GetFieldNames(e))
		err := rows.Scan(val...)
		if err != nil {
			return nil, err
		}
		userInfoNotifications = append(userInfoNotifications, e)
	}

	return userInfoNotifications, nil
}

func (s *NotificationSuite) checkUserInfoNotificationResponse(ctx context.Context, expectUserAfterSend, expectIndividualUserAfterSend []string, userInfoNotification []*entities.UserInfoNotification) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lenExpectUserAfterSend := len(expectUserAfterSend)
	lenUserActualSend := len(userInfoNotification)
	if lenExpectUserAfterSend != lenUserActualSend {
		if lenExpectUserAfterSend < lenUserActualSend {
			return ctx, fmt.Errorf("expect expect_user_after_send_noti >= actual_send but got expect_user_after_send_noti = %v and actual_send = %v",
				len(expectUserAfterSend),
				len(userInfoNotification),
			)
		}
	}

	mapCountActualSend := make(map[string]int)
	mapCountExpectedSend := make(map[string]int)

	for i := 0; i < lenUserActualSend; i++ {
		mapCountActualSend[userInfoNotification[i].UserID.String]++
	}

	for i := 0; i < lenExpectUserAfterSend; i++ {
		mapCountExpectedSend[expectUserAfterSend[i]]++
	}

	userIDsActualSend := make([]string, 0, lenUserActualSend)
	for i, uin := range userInfoNotification {
		userIDsActualSend = append(userIDsActualSend, userInfoNotification[i].UserID.String)
		if slice.Contains(expectIndividualUserAfterSend, uin.UserID.String) && !uin.IsIndividual.Bool {
			return ctx, fmt.Errorf("expected user id %v to be individual", uin.UserID.String)
		}
	}

	userIDsActualSendUnique := golibs.GetUniqueElementStringArray(userIDsActualSend)

	for _, v := range userIDsActualSendUnique {
		if mapCountActualSend[v] > mapCountExpectedSend[v] {
			return ctx, fmt.Errorf("expect to send %d message to user_id %s, but got %d, notification_id = %s",
				mapCountExpectedSend[v],
				v,
				mapCountActualSend[v],
				stepState.Notification.NotificationId)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
