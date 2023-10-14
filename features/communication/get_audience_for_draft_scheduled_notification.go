package communication

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"k8s.io/utils/strings/slices"
)

type GetAudienceForDraftScheduledNotificationSuite struct {
	*common.NotificationSuite
	TotalItems              uint32
	ViewRecipientListDetail map[string]*npb.RetrieveDraftAudienceResponse_Audience
	UserNames               []string
	ExcludedRecipientIDs    []string
	resAudiences            *npb.RetrieveDraftAudienceResponse
}

func (c *SuiteConstructor) InitGetAudienceForDraftScheduledNotification(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &GetAudienceForDraftScheduledNotificationSuite{
		NotificationSuite:       dep.notiCommonSuite,
		ViewRecipientListDetail: make(map[string]*npb.RetrieveDraftAudienceResponse_Audience, 0),
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" courses with "([^"]*)" classes for each course$`:                                           s.CreatesNumberOfCoursesWithClass,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^school admin add packages data of those courses for each student$`:                                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^current staff upsert notification to "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.CurrentStaffUpsertNotificationWithFilter,
		`^returns "([^"]*)" status code$`:                                                                                          s.CheckReturnStatusCode,
		`^current staff upsert and send notification$`:                                                                             s.currentStaffSendNotification,
		`^current staff view recipient list in detail page of draft or scheduled notification$`:                                    s.currentStaffViewRecipientOfDraftOrScheduledNotification,
		`^recipients must be same as data from view recipient list in detail page$`:                                                s.recipientsMustBeSameAsDataFromViewRecipient,
		`^current staff get the audience list in detail page with "([^"]*)" and "([^"]*)" and see results display "([^"]*)" rows$`: s.currentStaffGetTheAudienceListWithAndAndSeeResultsDisplayRows,
		`^school admin creates "([^"]*)" students with the same parent$`:                                                           s.CreatesNumberOfStudentsWithSameParentsInfo,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}
func (s *GetAudienceForDraftScheduledNotificationSuite) currentStaffViewRecipientOfDraftOrScheduledNotification(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)

	request := &npb.RetrieveDraftAudienceRequest{
		NotificationId: stepState.Notification.NotificationId,
		Paging:         &cpb.Paging{},
	}
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveDraftAudience(
		common.ContextWithToken(ctx, stepState.AuthToken),
		request,
	)
	if err != nil {
		return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed RetrieveGroupAudience: %v", err)
	}

	if len(res.Audiences) == 0 || res.TotalItems == 0 {
		return ctx, fmt.Errorf("expected RetrieveGroupAudience to not empty")
	}

	s.resAudiences = res

	for _, audience := range res.Audiences {
		s.ViewRecipientListDetail[audience.UserId] = audience
	}
	s.TotalItems = res.TotalItems
	parentChildMap := make(map[string]string)
	for _, student := range stepState.Students {
		s.UserNames = append(s.UserNames, student.Name)
		if len(student.Parents) > 0 {
			for _, parent := range student.Parents {
				s.UserNames = append(s.UserNames, parent.Name)
			}
		}
		if user, ok := s.ViewRecipientListDetail[student.ID]; ok {
			if student.GradeMaster.Name != user.Grade {
				return ctx, fmt.Errorf("expected grade %s, actual %s", student.GradeMaster.Name, user.Grade)
			}
			for _, parent := range student.Parents {
				if parent.ID == user.UserId {
					if _, ok := parentChildMap[parent.ID]; !ok {
						parentChildMap[parent.ID] = student.Name
					}
				}
			}
		}
	}
	for userID, audience := range s.ViewRecipientListDetail {
		if audienceChildName, ok := parentChildMap[userID]; ok {
			if audienceChildName != audience.ChildName {
				return ctx, fmt.Errorf("expected equal child name %s, actual %s", audienceChildName, audience.ChildName)
			}
		}
	}
	return common.StepStateToContext(ctx, stepState), nil
}

func (s *GetAudienceForDraftScheduledNotificationSuite) currentStaffSendNotification(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	upsertRequest := &npb.UpsertNotificationRequest{
		Notification: stepState.Notification,
	}
	notiClient := npb.NewNotificationModifierServiceClient(s.NotificationMgmtGRPCConn)
	upsertResponse, err := notiClient.UpsertNotification(
		ctx,
		upsertRequest,
	)
	if err != nil {
		return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed UpsertNotification: %v", err)
	}
	stepState.Notification.NotificationId = upsertResponse.NotificationId

	sendRequest := &npb.SendNotificationRequest{
		NotificationId: upsertResponse.NotificationId,
	}
	_, err = notiClient.SendNotification(
		ctx,
		sendRequest,
	)
	if err != nil {
		return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed SendNotification: %v", err)
	}
	return common.StepStateToContext(ctx, stepState), nil
}

func (s *GetAudienceForDraftScheduledNotificationSuite) recipientsMustBeSameAsDataFromViewRecipient(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	e := &entities.UserInfoNotification{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf(`
		SELECT %s
		FROM users_info_notifications ufn
		WHERE ufn.notification_id = $1
	`, strings.Join(fields, ","))
	rows, err := s.BobDBConn.Query(ctx, query, stepState.Notification.NotificationId)
	if err != nil {
		return common.StepStateToContext(ctx, stepState), nil
	}
	defer rows.Close()

	actualUserIDs := []string{}
	actualUsersInfoNotiMap := make(map[string]*entities.UserInfoNotification, 0)
	for rows.Next() {
		item := &entities.UserInfoNotification{}
		err = rows.Scan(database.GetScanFields(item, database.GetFieldNames(item))...)
		if err != nil {
			return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed scan %v", err)
		}
		actualUsersInfoNotiMap[item.UserID.String] = item
		actualUserIDs = append(actualUserIDs, item.UserID.String)
	}

	// what user see in View Recipient List popup must be same as the recipients list receive noti
	if len(actualUserIDs) < len(s.ViewRecipientListDetail) {
		return ctx, fmt.Errorf("expected actual recipients to be >= the view recipient list. actual %d, list %d, notification_id: %s", len(actualUserIDs), len(s.ViewRecipientListDetail), stepState.Notification.NotificationId)
	}

	for audienceID := range s.ViewRecipientListDetail {
		if !slices.Contains(actualUserIDs, audienceID) {
			return ctx, fmt.Errorf("expected user ID %s to be in actual recipient list", audienceID)
		}
	}

	for recipientID, recipient := range s.ViewRecipientListDetail {
		if userNoti, ok := actualUsersInfoNotiMap[recipientID]; ok {
			if recipient.UserId == userNoti.StudentID.String &&
				recipient.UserName != userNoti.StudentName.String {
				return ctx, fmt.Errorf("expected recipient %s to have user name %s, actual %s", recipientID, recipient.UserName, userNoti.StudentName.String)
			}
			if recipient.UserId == userNoti.ParentID.String &&
				recipient.UserName != userNoti.ParentName.String {
				return ctx, fmt.Errorf("expected recipient %s to have user name %s, actual %s", recipientID, recipient.UserName, userNoti.ParentName.String)
			}
		} else {
			return ctx, fmt.Errorf("expected recipient %s in actual user notification list", recipientID)
		}
	}
	return common.StepStateToContext(ctx, stepState), nil
}

func (s *GetAudienceForDraftScheduledNotificationSuite) currentStaffGetTheAudienceListWithAndAndSeeResultsDisplayRows(ctx context.Context, pageno, limit, expectResult int) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	// RetrieveDraftAudience allows returning duplicate parent UserID if parent has multiple students
	viewedAudienceIDsMap := make(map[string]bool)
	var offset int64
	i := 1
	for {
		request := &npb.RetrieveDraftAudienceRequest{
			NotificationId: stepState.Notification.NotificationId,
			Paging: &cpb.Paging{
				Limit:  uint32(limit),
				Offset: &cpb.Paging_OffsetInteger{OffsetInteger: offset},
			},
		}
		res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveDraftAudience(
			common.ContextWithToken(ctx, stepState.AuthToken),
			request,
		)
		if err != nil {
			return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed RetrieveGroupAudience: %v", err)
		}

		for _, audience := range res.Audiences {
			switch audience.UserGroup {
			case cpb.UserGroup_USER_GROUP_STUDENT:
				if hasSeenStudent := viewedAudienceIDsMap[audience.UserId]; !hasSeenStudent {
					viewedAudienceIDsMap[audience.UserId] = true
				} else {
					return ctx, fmt.Errorf("unexpected student user %s to be display in result page number %d", audience.UserId, i)
				}
			case cpb.UserGroup_USER_GROUP_PARENT:
				if hasSeenParentWithChildren := viewedAudienceIDsMap[audience.UserId+audience.ChildName]; !hasSeenParentWithChildren {
					viewedAudienceIDsMap[audience.UserId+audience.ChildName] = true
				} else {
					return ctx, fmt.Errorf("unexpected parent user %s to be display in result page number %d", audience.UserId, i)
				}
			}
		}

		if i == pageno {
			if len(res.Audiences) != expectResult {
				return ctx, fmt.Errorf("expected page %d will have %d results, actual %d", pageno, expectResult, len(res.Audiences))
			}
			break
		}

		if res.NextPage.GetOffsetInteger() == request.Paging.GetOffsetInteger() { // there's no next page
			break
		}
		i++
		offset = res.NextPage.GetOffsetInteger()
	}
	return common.StepStateToContext(ctx, stepState), nil
}
