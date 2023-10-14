package domain

import (
	"context"
	"fmt"
	"sync"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	bobRepo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pbu "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/bxcodec/faker/v3/support/slice"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"k8s.io/utils/strings/slices"
)

type EurekaCourseReader interface {
	ListStudentIDsByCourse(ctx context.Context, courseIDs []string, schoolID int32) ([]entities.StudentCourses, error)
}

type AudienceRetrieverService struct {
	Env          string
	AudienceRepo interface {
		FindGroupAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *repositories.FindGroupAudienceFilter, opts *repositories.FindAudienceOption) ([]*entities.Audience, error)
		CountGroupAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *repositories.FindGroupAudienceFilter, opts *repositories.FindAudienceOption) (uint32, error)
		FindIndividualAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *repositories.FindIndividualAudienceFilter) ([]*entities.Audience, error)
		FindDraftAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *repositories.FindDraftAudienceFilter, opts *repositories.FindAudienceOption) ([]*entities.Audience, error)
		CountDraftAudiencesByFilter(ctx context.Context, db database.QueryExecer, filter *repositories.FindDraftAudienceFilter, opts *repositories.FindAudienceOption) (uint32, error)
	}

	StudentParentRepo interface {
		FindParentByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*bobEntities.Parent, error)
		GetStudentParents(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*bobEntities.StudentParent, error)
	}

	InfoNotificationAccessPathRepo interface {
		GetLocationIDsByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) ([]string, error)
	}

	LocationRepo interface {
		GetLowestLocationIDsByIDs(ctx context.Context, db database.QueryExecer, locationIDs []string) ([]string, error)
		GetLowestGrantedLocationsByUserIDAndPermissions(ctx context.Context, db database.QueryExecer, userID string, permissions []string) ([]string, map[string]string, error)
	}

	UserRepo interface {
		FindUser(ctx context.Context, db database.QueryExecer, filter *repositories.FindUserFilter) ([]*entities.User, map[string]*entities.User, error)
	}

	GradeRepo interface {
		GetGradesByOrg(ctx context.Context, db database.QueryExecer, orgID string) (map[string]string, error)
	}

	NotificationInternalUserRepo interface {
		GetByOrgID(ctx context.Context, db database.QueryExecer, orgID string) (*entities.NotificationInternalUser, error)
	}
}

func NewAudienceRetrieverService(env string) *AudienceRetrieverService {
	return &AudienceRetrieverService{
		AudienceRepo: &repositories.AudienceRepo{
			AudienceSQLBuilder: repositories.AudienceSQLBuilder{},
		},
		StudentParentRepo:              &bobRepo.StudentParentRepo{},
		InfoNotificationAccessPathRepo: &repositories.InfoNotificationAccessPathRepo{},
		LocationRepo:                   &repositories.LocationRepo{},
		UserRepo:                       &repositories.UserRepo{},
		GradeRepo:                      &repositories.GradeRepo{},
		NotificationInternalUserRepo:   &repositories.NotificationInternalUserRepo{},
		Env:                            env,
	}
}

// This function is using
// Separates Group Target with Individual Target and Individual Student Target by User Groups
// Group Target will take User Group into account to calculate audiences
// Individual Target (receiver_ids, generic_receiver_ids) will ignore User Group selection
func (s *AudienceRetrieverService) FindAudiences(ctx context.Context, db database.QueryExecer, notification *entities.InfoNotification) ([]*entities.Audience, error) {
	// nolint
	targetGroup, err := notification.GetTargetGroup()
	if err != nil {
		return nil, fmt.Errorf("failed GetTargetGroup: %v", err)
	}
	notificationID, notificationType := notification.NotificationID.String, notification.Type.String
	excludedGenericReceiverIds := database.FromTextArray(notification.ExcludedGenericReceiverIDs)
	genericReceiverIDs := database.FromTextArray(notification.GenericReceiverIDs)
	studentReceiverIDs := database.FromTextArray(notification.ReceiverIDs)

	// -- START FIND RECEIVERS BY TARGET GROUP --
	groupAudiences := []*entities.Audience{}
	if !utils.CheckNoneSelectTargetGroup(targetGroup) {
		filter, err := s.makeGroupAudienceFilter(ctx, db, notificationID, notificationType, targetGroup, excludedGenericReceiverIds)
		if err != nil {
			return nil, fmt.Errorf("failed makeGroupAudienceFilter: %v", err)
		}

		// rawGroupAudiences can be both student and parent with their childs
		groupAudiences, err = s.AudienceRepo.FindGroupAudiencesByFilter(ctx, db, filter, repositories.NewFindAudienceOption())
		if err != nil {
			return nil, fmt.Errorf("failed FindGroupAudiencesByFilter: %v", err)
		}
	}
	groupAudienceUserIDs := make([]string, 0)
	for _, audience := range groupAudiences {
		groupAudienceUserIDs = append(groupAudienceUserIDs, audience.UserID.String)
	}
	// -- END FIND RECEIVERS BY TARGET GROUP --

	// -- START FIND RECEIVERS BY GENERIC RECEIVER --
	// Remove duplicate individual target IDs from group target IDs
	individualGenericIDs := []string{}
	for _, genericReceiverID := range genericReceiverIDs {
		if !slices.Contains(groupAudienceUserIDs, genericReceiverID) {
			individualGenericIDs = append(individualGenericIDs, genericReceiverID)
		}
	}

	individualGenericAudiences := make([]*entities.Audience, 0)
	if len(individualGenericIDs) > 0 {
		individualFilter, err := s.makeIndividualAudienceFilter(ctx, db, notificationType, individualGenericIDs)
		if err != nil {
			return nil, fmt.Errorf("failed makeIndividualAudienceFilter: %v", err)
		}

		individualGenericAudiences, err = s.AudienceRepo.FindIndividualAudiencesByFilter(ctx, db, individualFilter)
		if err != nil {
			return nil, fmt.Errorf("failed FindIndividualAudiencesByFilter for individual: %v", err)
		}
	}
	// -- END FIND RECEIVERS BY GENERIC RECEIVER --

	// -- This section supports for Async Notification --
	// -- START FIND RECEIVERS BY TARGET RECEIVER (student/parent) (depend on target_group.user_group_filter) --
	// Remove duplicate individual student IDs from group target IDs
	individualStudentIDs := []string{}
	for _, studentReceiverID := range studentReceiverIDs {
		if !slices.Contains(groupAudienceUserIDs, studentReceiverID) {
			individualStudentIDs = append(individualStudentIDs, studentReceiverID)
		}
	}

	individualAudiencesWithUserGroups := make([]*entities.Audience, 0)
	if len(individualStudentIDs) > 0 {
		rawIndividualAudiencesWithUserGroups, err := s.findIndividualAudiencesByStudentIDAndUserGroups(ctx, db, notificationType, individualStudentIDs, targetGroup.UserGroupFilter)
		if err != nil {
			return nil, fmt.Errorf("failed findIndividualAudiencesByStudentIDAndUserGroups: %v", err)
		}

		// Remove duplicate individual audiences by user groups from individual generic audiences
		for _, audience := range rawIndividualAudiencesWithUserGroups {
			if !slices.Contains(individualGenericIDs, audience.UserID.String) {
				individualAudiencesWithUserGroups = append(individualAudiencesWithUserGroups, audience)
			}
		}
	}
	// -- END FIND RECEIVERS BY TARGET RECEIVER (student/parent) --

	resAudiences := make([]*entities.Audience, 0, len(groupAudiences)+len(individualGenericAudiences)+len(individualAudiencesWithUserGroups))
	resAudiences = append(resAudiences, groupAudiences...)
	resAudiences = append(resAudiences, individualGenericAudiences...)
	resAudiences = append(resAudiences, individualAudiencesWithUserGroups...)
	return resAudiences, nil
}

// Audience Selector
// This function only use to find GroupAudience, NOT IndividualAudience
func (s *AudienceRetrieverService) FindGroupAudiencesWithPaging(ctx context.Context, db database.QueryExecer, notificationID string, targetGroup *entities.InfoNotificationTarget, keyword string, includeUserIDs []string, limit, offset int) ([]*entities.Audience, uint32, error) {
	// Must have atleast 1 filter enabled
	if utils.CheckNoneSelectTargetGroup(targetGroup) {
		return nil, uint32(0), nil
	}

	// This function is only used for Composed notification.
	notificationType := cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String()
	filter, err := s.makeGroupAudienceFilter(ctx, db, notificationID, notificationType, targetGroup, nil)
	if err != nil {
		return nil, uint32(0), fmt.Errorf("failed makeGroupAudienceFilter: %v", err)
	}
	_ = filter.IncludeUserIds.Set(includeUserIDs)
	_ = filter.Keyword.Set(keyword)
	_ = filter.Limit.Set(limit)
	_ = filter.Offset.Set(offset)

	opts := repositories.NewFindAudienceOption()
	opts.OrderByName = consts.AscendingOrder
	opts.IsGetName = true

	audiences := []*entities.Audience{}
	total := uint32(0)

	wg := sync.WaitGroup{}
	errChan := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		audiences, err = s.AudienceRepo.FindGroupAudiencesByFilter(ctx, db, filter, opts)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		total, err = s.AudienceRepo.CountGroupAudiencesByFilter(ctx, db, filter, opts)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var goErr error
	for er := range errChan {
		if er != nil {
			goErr = multierr.Append(goErr, er)
		}
	}
	if goErr != nil {
		return nil, uint32(0), fmt.Errorf("failed AudienceRepo: %v", goErr)
	}

	audiences, err = s.setExtensionInfoForAudiences(ctx, db, audiences)
	if err != nil {
		return nil, uint32(0), fmt.Errorf("failed set setExtensionInfoForAudiences: %v", err)
	}

	return audiences, total, nil
}

// Audience Selector
// This function only use to find both of GroupAudience, and IndividualAudience
func (s *AudienceRetrieverService) FindDraftAudiencesWithPaging(ctx context.Context, db database.QueryExecer, notificationID string, targetGroup *entities.InfoNotificationTarget, genericReceiverIds, groupExcludedGenericReceiverIds []string, limit, offset int) ([]*entities.Audience, uint32, error) {
	// Ensure empty filter case should be 0 recipient
	if utils.CheckNoneSelectTargetGroup(targetGroup) && len(genericReceiverIds) == 0 {
		return nil, 0, nil
	}

	var err error
	groupFilter := repositories.NewFindGroupAudienceFilter()
	// This function is only used for Composed notification.
	notificationType := cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String()
	if !utils.CheckNoneSelectTargetGroup(targetGroup) {
		groupFilter, err = s.makeGroupAudienceFilter(ctx, db, notificationID, notificationType, targetGroup, groupExcludedGenericReceiverIds)
		if err != nil {
			return nil, uint32(0), fmt.Errorf("failed makeGroupAudienceFilter: %v", err)
		}
	}

	individualFilter := repositories.NewFindIndividualAudienceFilter()
	if len(genericReceiverIds) > 0 {
		individualFilter, err = s.makeIndividualAudienceFilter(ctx, db, notificationType, genericReceiverIds)
		if err != nil {
			return nil, uint32(0), fmt.Errorf("failed set individualFilter: %v", err)
		}
	}

	filter := repositories.NewFindDraftAudienceFilter()
	filter.GroupFilter = groupFilter
	filter.IndividualFilter = individualFilter
	filter.Limit = database.Int8(int64(limit))
	filter.Offset = database.Int8(int64(offset))

	opts := repositories.NewFindAudienceOption()
	opts.OrderByName = consts.AscendingOrder
	opts.IsGetName = true

	audiences := []*entities.Audience{}
	total := uint32(0)

	wg := sync.WaitGroup{}
	errChan := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		audiences, err = s.AudienceRepo.FindDraftAudiencesByFilter(ctx, db, filter, opts)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		total, err = s.AudienceRepo.CountDraftAudiencesByFilter(ctx, db, filter, opts)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var goErr error
	for er := range errChan {
		if er != nil {
			goErr = multierr.Append(goErr, er)
		}
	}
	if goErr != nil {
		return nil, uint32(0), fmt.Errorf("failed AudienceRepo: %v", goErr)
	}

	audiences, err = s.setExtensionInfoForAudiences(ctx, db, audiences)
	if err != nil {
		return nil, uint32(0), fmt.Errorf("failed set setExtensionInfoForAudiences: %v", err)
	}

	return audiences, total, nil
}

func (s *AudienceRetrieverService) findIndividualAudiencesByStudentIDAndUserGroups(ctx context.Context, db database.QueryExecer, notificationType string, individualStudentIDs []string, userFilter entities.InfoNotificationTarget_UserGroupFilter) ([]*entities.Audience, error) {
	individualFilter, err := s.makeIndividualAudienceFilter(ctx, db, notificationType, individualStudentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed set individualFilter: %v", err)
	}

	var individualStudentAudiences []*entities.Audience
	individualStudentAudiences, err = s.AudienceRepo.FindIndividualAudiencesByFilter(ctx, db, individualFilter)
	if err != nil {
		return nil, fmt.Errorf("find individual student %v", err)
	}
	for _, individualStudent := range individualStudentAudiences {
		err = individualStudent.IsIndividual.Set(true)
		if err != nil {
			return nil, fmt.Errorf("failed set student IsIndividual: %v", err)
		}
	}

	// right now these Audiences only have student ID, current grade OR grade ID, course IDs
	studentIDs := make([]string, 0, len(individualStudentAudiences))
	for _, individualStudentAudience := range individualStudentAudiences {
		studentIDs = append(studentIDs, individualStudentAudience.StudentID.String)
	}

	results := make([]*entities.Audience, 0)
	for _, group := range userFilter.UserGroups {
		switch group {
		case cpb.UserGroup_USER_GROUP_STUDENT.String():
			for _, studentAudience := range individualStudentAudiences {
				err = multierr.Combine(
					studentAudience.UserID.Set(studentAudience.StudentID),
					studentAudience.ParentID.Set(nil),
					studentAudience.UserGroup.Set(cpb.UserGroup_USER_GROUP_STUDENT.String()),
				)
				if err != nil {
					return nil, fmt.Errorf("failed combine for studentAudience: %v", err)
				}
			}
			results = append(results, individualStudentAudiences...)

		case cpb.UserGroup_USER_GROUP_PARENT.String():
			// for now we still use Bob repo because Notification is still depend on Bob's Parent entity
			studentParents, err := s.StudentParentRepo.GetStudentParents(ctx, db, database.TextArray(studentIDs))

			if err != nil {
				return nil, fmt.Errorf("StudentParentRepo.GetStudentParents: %v", err)
			}

			for _, stuParent := range studentParents {
				parentAudience := &entities.Audience{}
				err = multierr.Combine(
					parentAudience.UserID.Set(stuParent.ParentID),
					parentAudience.ParentID.Set(stuParent.ParentID),
					parentAudience.StudentID.Set(stuParent.StudentID),
					parentAudience.ChildIDs.Set([]string{stuParent.StudentID.String}),
					parentAudience.UserGroup.Set(cpb.UserGroup_USER_GROUP_PARENT.String()),
					parentAudience.IsIndividual.Set(true),
				)
				if slice.Contains(studentIDs, stuParent.StudentID.String) {
					err = multierr.Append(err, parentAudience.IsIndividual.Set(true))
				}
				if err != nil {
					return nil, fmt.Errorf("failed combine for parentAudience: %v", err)
				}
				results = append(results, parentAudience)
			}
		}
	}

	return results, nil
}

func (s *AudienceRetrieverService) makeGroupAudienceFilter(ctx context.Context, db database.QueryExecer, notificationID, notificationType string, targetGroup *entities.InfoNotificationTarget, excludedGenericReceiverIds []string) (*repositories.FindGroupAudienceFilter, error) {
	var err error
	courseFilter := targetGroup.CourseFilter
	gradeFilter := targetGroup.GradeFilter
	locationFilter := targetGroup.LocationFilter
	classFilter := targetGroup.ClassFilter
	schoolFilter := targetGroup.SchoolFilter

	filter := repositories.NewFindGroupAudienceFilter()

	// In case Immediate notification, context will be of the current user.
	// In case Scheduled notification, context will be of the EditedUserID.
	// If EditedUserID == CreatedUserID, meaning the notification owner is sending the notification.
	// If EditedUserID != CreatedUserID, meaning someone else is sending the notification (ex. admin role)
	if locationFilter.Type == consts.TargetGroupSelectTypeList.String() {
		err = multierr.Combine(
			filter.LocationIDs.Set(locationFilter.LocationIDs),
		)
		if err != nil {
			return nil, fmt.Errorf("failed case listSelectType set locationIDs: %v", err)
		}
	} else {
		var locationIDs []string
		if notificationID != "" {
			// If found SOME locationIDs, this means 2 cases, in either cases, we keep using the values of locationIDs
			// 1. CreatedUser's granted access paths had not changed
			// 2. CreatedUser's granted access paths had changed, but got some in common locations
			// ex. granted [1,2,3] , changed to [2,3,4]
			// => locationIDs = [2,3, some descendance of [2,3] if exists,..]
			locationIDs, err = s.LocationRepo.GetLowestLocationIDsByIDs(ctx, db, locationFilter.LocationIDs)
			if err != nil {
				return nil, fmt.Errorf("failed GetLowestLocationIDsByIDs: %v", err)
			}
		} else {
			locationIDs, err = s.getAssignedLocationsToQueryRecipients(ctx, db)
			if err != nil {
				return nil, fmt.Errorf("failed getAssignedLocationsFromUserGrantedLocations: %v", err)
			}
		}

		// This case appeared when the granted location of them is changed -> can't get any location.
		// If found NO locationIDs, this means CreatedUser's granted locations had changed
		// and have no common with the old granted locations.
		// ex. granted [1,2,3] , changed to [4,5,6]
		if len(locationIDs) == 0 {
			// Then we use the current notification location_ids from LocationFilter (saved it at UpsertNotification API).
			// If we don't, the repo will resulting recipients of locations [4,5,6] instead, this is wrong behavior.
			// locationIDs will be same as notification locations access path but without descendant locations, but in this case the user's granted locations are already changed,
			// using access path locations is enough to make sure the repo will query 0 recipients
			locationIDs = locationFilter.LocationIDs
		}

		err = filter.LocationIDs.Set(locationIDs)
		if err != nil {
			return nil, fmt.Errorf("failed case noneSelectType set locationIDs: %v", err)
		}
	}

	selectedCourseIDs, selectedType := utils.CourseTargetGroupToCourseFilter(courseFilter)
	err = multierr.Combine(
		filter.CourseIDs.Set(selectedCourseIDs),
		filter.CourseSelectType.Set(selectedType),
	)
	if err != nil {
		return nil, fmt.Errorf("failed set CourseFilter: %v", err)
	}

	selectedClassIDs, selectedType := utils.ClassTargetGroupToClassFilter(classFilter)
	err = multierr.Combine(
		filter.ClassIDs.Set(selectedClassIDs),
		filter.ClassSelectType.Set(selectedType),
	)
	if err != nil {
		return nil, fmt.Errorf("failed set ClassFilter: %v", err)
	}

	selectedGradeIDs, selectedType := utils.GradeTargetGroupToGradeFilter(gradeFilter)
	err = multierr.Combine(
		filter.GradeIDs.Set(selectedGradeIDs),
		filter.GradeSelectType.Set(selectedType),
	)
	if err != nil {
		return nil, fmt.Errorf("failed set GradeFilter: %v", err)
	}

	selectedSchoolIDs, selectedType := utils.SchoolTargetGroupToSchoolFilter(schoolFilter)
	err = multierr.Combine(
		filter.SchoolIDs.Set(selectedSchoolIDs),
		filter.SchoolSelectType.Set(selectedType),
	)
	if err != nil {
		return nil, fmt.Errorf("failed set SchoolFilter: %v", err)
	}

	err = filter.UserGroups.Set(targetGroup.UserGroupFilter.UserGroups)
	if err != nil {
		return nil, fmt.Errorf("faield set UserGroup: %v", err)
	}

	// Only check enrollment status when notification is COMPOSED
	if notificationType == cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String() {
		filter.StudentEnrollmentStatus = database.Text(pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String())
	}

	if len(excludedGenericReceiverIds) > 0 {
		_ = filter.ExcludeUserIds.Set(excludedGenericReceiverIds)
	}

	return filter, nil
}

func (s *AudienceRetrieverService) makeIndividualAudienceFilter(ctx context.Context, db database.QueryExecer, notificationType string, individualTargetIDs []string) (*repositories.FindIndividualAudienceFilter, error) {
	// LOCAL - STAG
	if s.Env == consts.StagingEnv || s.Env == consts.LocalEnv {
		individualFilterNew := repositories.NewFindIndividualAudienceFilter()
		_ = individualFilterNew.UserIDs.Set(individualTargetIDs)

		// Ignore check location in case send individual
		individualFilterNew.LocationIDs = pgtype.TextArray{
			Elements:   nil,
			Dimensions: nil,
			Status:     pgtype.Undefined,
		}

		if notificationType == cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String() {
			// Only check location when notification is COMPOSED
			grantedLocationIDs, err := s.getAssignedLocationsToQueryRecipients(ctx, db)
			if err != nil {
				return nil, fmt.Errorf("failed getAssignedLocationsFromUserGrantedLocations(individual):%v", err)
			}
			_ = individualFilterNew.LocationIDs.Set(grantedLocationIDs)

			// Only check enrollment status when notification is COMPOSED
			_ = individualFilterNew.EnrollmentStatuses.Set([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()})
		}

		return individualFilterNew, nil
	}

	// UAT - PROD, TODO: UPDATE THIS LOGIC WHEN STABLE
	grantedLocationIDs, err := s.getAssignedLocationsToQueryRecipients(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed getAssignedLocationsFromUserGrantedLocations(individual):%v", err)
	}
	individualFilter := repositories.NewFindIndividualAudienceFilter()
	_ = individualFilter.LocationIDs.Set(grantedLocationIDs)
	_ = individualFilter.UserIDs.Set(individualTargetIDs)
	// Only check enrollment status when notification is COMPOSED
	if notificationType == cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String() {
		_ = individualFilter.EnrollmentStatuses.Set([]string{pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()})
	}

	return individualFilter, nil
}

// cases: NATs, FindAudiencesWithPaging, query Individual recipients
func (s *AudienceRetrieverService) getAssignedLocationsToQueryRecipients(ctx context.Context, db database.QueryExecer) ([]string, error) {
	var locationIDs []string
	var err error
	userInfo := golibs.UserInfoFromCtx(ctx)
	userID := userInfo.UserID
	if userID == "" {
		return nil, fmt.Errorf("cannot get userID from context")
	}

	// Change to use user.user.read in there only for LOCAL/STAG, need to update for UAT/PROD if work well.
	if s.Env == consts.LocalEnv || s.Env == consts.StagingEnv {
		userPermission := []string{
			consts.UserReadPermission,
		}
		locationIDs, _, err = s.LocationRepo.GetLowestGrantedLocationsByUserIDAndPermissions(ctx, db, userID, userPermission)
		if err != nil {
			return nil, fmt.Errorf("failed GetLowestGrantedLocationsByUserIDAndPermissions: %v", err)
		}
	} else {
		notificationPermissions := []string{
			consts.NotificationWritePermission,
			consts.NotificationOwnerPermission,
		}
		locationIDs, _, err = s.LocationRepo.GetLowestGrantedLocationsByUserIDAndPermissions(ctx, db, userID, notificationPermissions)
		if err != nil {
			return nil, fmt.Errorf("failed GetLowestGrantedLocationsByUserIDAndPermissions: %v", err)
		}
	}

	return locationIDs, nil
}

// Set grade, child name for each audience support for FindGroupAudiences and FindDraftAudiences
func (s *AudienceRetrieverService) setExtensionInfoForAudiences(ctx context.Context, db database.QueryExecer, audiences []*entities.Audience) ([]*entities.Audience, error) {
	org, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed OrganizationFromContext: %v", err)
	}
	gradeMap, err := s.GradeRepo.GetGradesByOrg(ctx, db, org.OrganizationID().String())
	if err != nil {
		return nil, fmt.Errorf("failed GetGradesByOrg: %v", err)
	}

	childIDs := []string{}
	for _, audience := range audiences {
		childIDs = append(childIDs, database.FromTextArray(audience.ChildIDs)...)
	}

	userFilter := repositories.NewFindUserFilter()
	_ = userFilter.UserIDs.Set(childIDs)
	_, userNameMap, err := s.UserRepo.FindUser(ctx, db, userFilter)
	if err != nil {
		return nil, fmt.Errorf("failed FindUser: %v", err)
	}

	for _, audience := range audiences {
		childIDs := database.FromTextArray(audience.ChildIDs)
		childNames := []string{}
		for _, childID := range childIDs {
			if child, ok := userNameMap[childID]; ok {
				childNames = append(childNames, child.Name.String)
			}
		}
		err = audience.ChildNames.Set(childNames)
		if err != nil {
			return nil, fmt.Errorf("failed set ChildNames of parent %v, err: %v", audience.UserID, err)
		}

		if grade, ok := gradeMap[audience.GradeID.String]; ok {
			err = audience.GradeName.Set(grade)
			if err != nil {
				return nil, fmt.Errorf("failed set Grade of student %v, err: %v", audience.UserID, err)
			}
		}
	}

	return audiences, nil
}
