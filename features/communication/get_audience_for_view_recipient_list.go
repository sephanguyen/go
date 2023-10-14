package communication

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/cucumber/godog"
	"k8s.io/utils/strings/slices"
)

type GetAudienceForViewRecipientListSuite struct {
	*common.NotificationSuite
	TotalItems           uint32
	ViewRecipientListMap map[string]*npb.RetrieveGroupAudienceResponse_Audience
	Keyword              string
	UserNames            []string
	ExcludedRecipientIDs []string
}

func (c *SuiteConstructor) InitGetAudienceForViewRecipientList(dep *DependencyV2, godogCtx *godog.ScenarioContext) {
	s := &GetAudienceForViewRecipientListSuite{
		NotificationSuite:    dep.notiCommonSuite,
		ViewRecipientListMap: make(map[string]*npb.RetrieveGroupAudienceResponse_Audience, 0),
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`:                                                                                                                                     s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^current staff compose notification with "([^"]*)" and "([^"]*)" course and "([^"]*)" grade and "([^"]*)" location and "([^"]*)" class and "([^"]*)" school and "([^"]*)" individuals and "([^"]*)" scheduled time with "([^"]*)" and important is "([^"]*)"$`: s.currentStaffComposeNotification,
		`^current staff upsert and send notification$`:                                                              s.currentStaffSendNotification,
		`^current staff view recipient list popup$`:                                                                 s.currentStaffViewRecipientListPopup,
		`^recipients must be same as data from view recipient list popup$`:                                          s.recipientsMustBeSameAsDataFromViewRecipientListPopup,
		`^school admin add packages data of those courses for each student$`:                                        s.AddPackagesDataOfThoseCoursesForEachStudent,
		`^school admin creates "([^"]*)" courses with "([^"]*)" classes for each course$`:                           s.CreatesNumberOfCoursesWithClass,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                    s.CreatesNumberOfStudentsWithParentsInfo,
		`^current staft search for a keyword and see correct result$`:                                               s.currentStaftSearchForAKeywordAndSeeCorrectResult,
		`^current staff get the audience list with "([^"]*)" and "([^"]*)" and see results display "([^"]*)" rows$`: s.currentStaffGetTheAudienceListWithAndAndSeeResultsDisplayRows,
		`^school admin excluded some recipients from the list$`:                                                     s.schoolAdminExcludedSomeRecipientsFromTheList,
		`^excluded recipients no longer available to view$`:                                                         s.excludedRecipientsNoLongerAvailableToView,
		`^api RetrieveGroupAudience return empty result$`:                                                           s.apiRetrieveGroupAudienceReturnEmptyResult,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *GetAudienceForViewRecipientListSuite) currentStaffComposeNotification(ctx context.Context, userGroups, courseFilter, gradeFilter, locationFilter, classFilter, schoolFilter, individualFilter, scheduledStatus, status string, isImportantStr string) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)

	isImportant := false
	if isImportantStr == "true" {
		isImportant = true
	}

	opts := &common.NotificationWithOpts{
		UserGroups:       userGroups,
		CourseFilter:     courseFilter,
		GradeFilter:      gradeFilter,
		LocationFilter:   locationFilter,
		ClassFilter:      classFilter,
		SchoolFilter:     schoolFilter,
		IndividualFilter: individualFilter,
		ScheduledStatus:  scheduledStatus,
		Status:           status,
		IsImportant:      isImportant,
	}
	var err error
	ctx, stepState.Notification, err = s.GetNotificationWithOptions(ctx, opts)
	if err != nil {
		return common.StepStateToContext(ctx, stepState), fmt.Errorf("GetNotificationWithOptions: %v", err)
	}

	return common.StepStateToContext(ctx, stepState), nil
}

func (s *GetAudienceForViewRecipientListSuite) currentStaffViewRecipientListPopup(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	targetGroup := stepState.Notification.GetTargetGroup()

	request := &npb.RetrieveGroupAudienceRequest{
		TargetGroup: targetGroup,
		Paging:      &cpb.Paging{},
		Keyword:     "",
	}
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveGroupAudience(
		common.ContextWithToken(ctx, stepState.AuthToken),
		request,
	)
	if err != nil {
		return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed RetrieveGroupAudience: %v", err)
	}

	if len(res.Audiences) == 0 || res.TotalItems == 0 {
		return ctx, fmt.Errorf("expected RetrieveGroupAudience to not empty")
	}

	for _, audience := range res.Audiences {
		s.ViewRecipientListMap[audience.UserId] = audience
	}
	s.TotalItems = res.TotalItems
	parentChildMap := make(map[string][]string)
	for _, student := range stepState.Students {
		s.UserNames = append(s.UserNames, student.Name)
		if len(student.Parents) > 0 {
			for _, parent := range student.Parents {
				s.UserNames = append(s.UserNames, parent.Name)
			}
		}
		if user, ok := s.ViewRecipientListMap[student.ID]; ok {
			if student.GradeMaster.Name != user.Grade {
				return ctx, fmt.Errorf("expected grade %s, actual %s", student.GradeMaster.Name, user.Grade)
			}
			for _, parent := range student.Parents {
				if parent.ID == user.UserId {
					if childNames, ok := parentChildMap[parent.ID]; ok {
						childNames = append(childNames, student.Name)
					} else {
						parentChildMap[parent.ID] = []string{student.Name}
					}
				}
			}
		}
	}
	for userID, audience := range s.ViewRecipientListMap {
		if audienceChildNames, ok := parentChildMap[userID]; ok {
			if len(audienceChildNames) != len(audience.ChildNames) {
				return ctx, fmt.Errorf("expected length of child names %d, actual %d", len(audienceChildNames), len(audience.ChildNames))
			}
			if !stringutil.SliceElementsMatch(audienceChildNames, audience.ChildNames) {
				return ctx, fmt.Errorf("child names different: expected %d, actual %d", len(audienceChildNames), len(audience.ChildNames))
			}
		}
	}
	s.Keyword = common.GetRandomKeywordFromStrings(s.UserNames)
	return common.StepStateToContext(ctx, stepState), nil
}

func (s *GetAudienceForViewRecipientListSuite) currentStaffSendNotification(ctx context.Context) (context.Context, error) {
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

func (s *GetAudienceForViewRecipientListSuite) recipientsMustBeSameAsDataFromViewRecipientListPopup(ctx context.Context) (context.Context, error) {
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
	if len(actualUserIDs) < len(s.ViewRecipientListMap) {
		return ctx, fmt.Errorf("expected actual recipients to be >= the view recipient list. actual %d, list %d", len(actualUserIDs), len(s.ViewRecipientListMap))
	}

	for audienceID := range s.ViewRecipientListMap {
		if !slices.Contains(actualUserIDs, audienceID) {
			return ctx, fmt.Errorf("expected user ID %s to be in actual recipient list", audienceID)
		}
	}

	for recipientID, recipient := range s.ViewRecipientListMap {
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

func (s *GetAudienceForViewRecipientListSuite) currentStaftSearchForAKeywordAndSeeCorrectResult(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)

	request := &npb.RetrieveGroupAudienceRequest{
		TargetGroup: stepState.Notification.TargetGroup,
		Paging:      &cpb.Paging{},
		Keyword:     s.Keyword,
	}
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveGroupAudience(
		common.ContextWithToken(ctx, stepState.AuthToken),
		request,
	)
	if err != nil {
		return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed RetrieveGroupAudience: %v", err)
	}

	actualUserNames := []string{}
	for _, audience := range res.Audiences {
		actualUserNames = append(actualUserNames, audience.UserName)
		if !strings.Contains(
			strings.ToLower(audience.UserName),
			strings.ToLower(s.Keyword)) {
			return ctx, fmt.Errorf("unexpected user name %s of %s, keyword %s", audience.UserName, audience.UserId, s.Keyword)
		}
	}
	return common.StepStateToContext(ctx, stepState), nil
}

func (s *GetAudienceForViewRecipientListSuite) currentStaffGetTheAudienceListWithAndAndSeeResultsDisplayRows(ctx context.Context, pageno, limit, expectResult int) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	viewedAudienceIDs := []string{}
	var offset int64 = 0
	i := 1
	for {
		request := &npb.RetrieveGroupAudienceRequest{
			TargetGroup: stepState.Notification.TargetGroup,
			Paging: &cpb.Paging{
				Limit:  uint32(limit),
				Offset: &cpb.Paging_OffsetInteger{OffsetInteger: offset},
			},
			Keyword: "",
		}
		res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveGroupAudience(
			common.ContextWithToken(ctx, stepState.AuthToken),
			request,
		)
		if err != nil {
			return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed RetrieveGroupAudience: %v", err)
		}

		for _, audience := range res.Audiences {
			if !slices.Contains(viewedAudienceIDs, audience.UserId) {
				viewedAudienceIDs = append(viewedAudienceIDs, audience.UserId)
			} else {
				// mean this userID has appeared in previous page
				return ctx, fmt.Errorf("unexpected user %s to be display in result page number %d", audience.UserId, i)
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

func (s *GetAudienceForViewRecipientListSuite) schoolAdminExcludedSomeRecipientsFromTheList(ctx context.Context) (context.Context, error) {
	for userID := range s.ViewRecipientListMap {
		s.ExcludedRecipientIDs = append(s.ExcludedRecipientIDs, userID)
		break
	}
	return ctx, nil
}

func (s *GetAudienceForViewRecipientListSuite) excludedRecipientsNoLongerAvailableToView(ctx context.Context) (context.Context, error) {
	// invalid those excluded recipients by changing their access path
	query := `
		UPDATE user_access_paths SET deleted_at = now()
		WHERE user_id = ANY($1::TEXT[]);
	`
	_, err := s.BobPostgresDBConn.Exec(ctx, query, database.TextArray(s.ExcludedRecipientIDs))
	if err != nil {
		return ctx, fmt.Errorf("failed update user_access_path: %v", err)
	}
	return ctx, nil
}

func (s *GetAudienceForViewRecipientListSuite) apiRetrieveGroupAudienceReturnEmptyResult(ctx context.Context) (context.Context, error) {
	stepState := common.StepStateFromContext(ctx)
	targetGroup := stepState.Notification.GetTargetGroup()

	request := &npb.RetrieveGroupAudienceRequest{
		TargetGroup: targetGroup,
		Paging:      &cpb.Paging{},
		Keyword:     "",
		UserIds:     s.ExcludedRecipientIDs,
	}
	res, err := npb.NewNotificationReaderServiceClient(s.NotificationMgmtGRPCConn).RetrieveGroupAudience(
		common.ContextWithToken(ctx, stepState.AuthToken),
		request,
	)
	if err != nil {
		return common.StepStateToContext(ctx, stepState), fmt.Errorf("failed RetrieveGroupAudience: %v", err)
	}

	if len(res.Audiences) != 0 || res.TotalItems != 0 {
		return ctx, fmt.Errorf("expected RetrieveGroupAudience to be empty")
	}

	return ctx, nil
}
