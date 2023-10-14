package common

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/features/conversationmgmt/common/entities"
	"github.com/manabie-com/backend/features/conversationmgmt/common/helpers"
	const_conversationmgmt "github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upbv2 "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *ConversationMgmtSuite) StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg(ctx context.Context, staffGrantedRole string) (context.Context, error) {
	organization, err := s.CreateNewOrganization(helpers.OrganizationTypeNew)
	if err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.CurrentResourcePath = fmt.Sprint(organization.ID)
	stepState.CurrentOrganicationID = organization.ID

	// Assigned resource path for ctx
	ctx = contextWithResourcePath(ctx, strconv.Itoa(int(organization.ID)))

	// Create grade masters
	grades, err := s.CreateGradeMasterForOrgazination(ctx, strconv.Itoa(int(organization.ID)))
	if err != nil {
		return ctx, err
	}
	stepState.GradeMasters = grades

	// Create schools 5 school with same school level
	schools, err := s.CreateSchoolsForOrganization(ctx, strconv.Itoa(int(organization.ID)))
	if err != nil {
		return ctx, err
	}
	stepState.Schools = schools

	if staffGrantedRole == "" {
		staffGrantedRole = helpers.StaffGrantedRoleSchoolAdmin
	}

	// Send state to context for create new user
	ctx = StepStateToContext(ctx, stepState)
	// Create a staff with granted role
	ctx, err = s.signedAsAccount(ctx, staffGrantedRole, int64(organization.ID), []string{organization.DefaultLocation.ID})
	if err != nil {
		return ctx, err
	}

	stepState = StepStateFromContext(ctx)

	// Assigned user_id for ctx
	ctx = contextWithUserID(ctx, stepState.CurrentUserID)

	organization.Staffs = []*entities.Staff{
		{
			ID:                 stepState.CurrentUserID,
			Token:              stepState.AuthToken,
			GrandtedRoles:      stepState.CurrentGrandtedRoles,
			UserGroup:          stepState.CurrentUserGroup,
			OrganizationIDs:    []int32{stepState.CurrentOrganicationID},
			GrantedLocationIDs: []string{organization.DefaultLocation.ID},
		},
	}

	stepState.Organization = organization
	stepState.CurrentStaff = organization.Staffs[0]

	// Assign AuthToken to ctx
	ctx = s.ContextWithToken(ctx, stepState.CurrentStaff.Token)
	// fmt.Printf("\n\nTOKEN: [%s]\n\n", stepState.CurrentStaff.Token)

	ctx, err = s.insertCurrentStaffIntoInternalUser(ctx)
	if err != nil {
		return ctx, fmt.Errorf("err insertCurrentStaffIntoInternalUser: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfCurrentOrg(ctx context.Context, staffGrantedRole string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.Organization == nil || stepState.CurrentOrganicationID == 0 {
		return ctx, fmt.Errorf("current organization not exist in your test, please create it first")
	}

	if staffGrantedRole == "" {
		staffGrantedRole = helpers.StaffGrantedRoleSchoolAdmin
	}

	// Create a staff with granted role
	ctx, err := s.signedAsAccount(ctx, staffGrantedRole, int64(stepState.CurrentOrganicationID), []string{stepState.Organization.DefaultLocation.ID})
	if err != nil {
		return ctx, err
	}

	stepState = StepStateFromContext(ctx)

	// Assigned user_id for ctx
	ctx = contextWithUserID(ctx, stepState.CurrentUserID)

	newStaff := &entities.Staff{
		ID:                 stepState.CurrentUserID,
		Token:              stepState.AuthToken,
		UserGroup:          stepState.CurrentUserGroup,
		GrandtedRoles:      stepState.CurrentGrandtedRoles,
		OrganizationIDs:    []int32{stepState.CurrentOrganicationID},
		GrantedLocationIDs: []string{stepState.Organization.DefaultLocation.ID},
	}

	stepState.Organization.Staffs = append(stepState.Organization.Staffs, newStaff)
	stepState.CurrentStaff = newStaff

	// Assign AuthToken to ctx
	ctx = s.ContextWithToken(ctx, stepState.CurrentStaff.Token)

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) StaffGrantedRoleAndLocationsLoggedInBackOfficeOfCurrentOrg(ctx context.Context, staffGrantedRole string, idxsLocStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	grantedLocationIDs := make([]string, 0)

	if idxsLocStr == "" {
		return ctx, fmt.Errorf("please choice descendant locations index")
	}

	idxsLocsStr := strings.Split(idxsLocStr, ",")
	for _, idxLocStr := range idxsLocsStr {
		idxLoc, err := strconv.Atoi(idxLocStr)
		if err != nil {
			return ctx, fmt.Errorf("can't convert descendant location index: %v", err)
		}
		if idxLoc <= 0 || idxLoc > helpers.NumberOfNewCenterLocationCreated {
			return ctx, fmt.Errorf("index descendant location out of range")
		}
		grantedLocationIDs = append(grantedLocationIDs, stepState.Organization.DescendantLocations[idxLoc-1].ID)
	}

	if stepState.Organization == nil || stepState.CurrentOrganicationID == 0 {
		return ctx, fmt.Errorf("current organization not exist in your test, please create it first")
	}

	if staffGrantedRole == "" {
		staffGrantedRole = helpers.StaffGrantedRoleSchoolAdmin
	}

	// Create a staff with granted role
	ctx, err := s.signedAsAccount(ctx, staffGrantedRole, int64(stepState.CurrentOrganicationID), grantedLocationIDs)
	if err != nil {
		return ctx, err
	}

	stepState = StepStateFromContext(ctx)

	// Assigned user_id for ctx
	ctx = contextWithUserID(ctx, stepState.CurrentUserID)

	newStaff := &entities.Staff{
		ID:                 stepState.CurrentUserID,
		Token:              stepState.AuthToken,
		UserGroup:          stepState.CurrentUserGroup,
		GrandtedRoles:      stepState.CurrentGrandtedRoles,
		OrganizationIDs:    []int32{stepState.CurrentOrganicationID},
		GrantedLocationIDs: grantedLocationIDs,
	}

	stepState.Organization.Staffs = append(stepState.Organization.Staffs, newStaff)
	stepState.CurrentStaff = newStaff

	// Assign AuthToken to ctx
	ctx = s.ContextWithToken(ctx, stepState.CurrentStaff.Token)

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) StaffGrantedRoleLoggedInBackOfficeOfRespectiveOrg(ctx context.Context, numOrg, staffGrantedRole string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	intNumOrg, _ := strconv.Atoi(numOrg)
	for i := 0; i < intNumOrg; i++ {
		emptyCtx := context.Background()
		tenantCtx, err := s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg(emptyCtx, staffGrantedRole)
		if err != nil {
			return ctx, fmt.Errorf("failed create multi tenant context")
		}
		stepState.MultiTenants = append(stepState.MultiTenants, &tenantCtx)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) StaffGrantedRoleOfJprepOrganizationWithDefaultLocationLoggedInBackOffice(ctx context.Context, staffGrantedRole string) (context.Context, error) {
	organization, err := s.CreateNewOrganization(helpers.OrganizationTypeJPREP)
	if err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.CurrentResourcePath = fmt.Sprint(organization.ID)
	stepState.CurrentOrganicationID = organization.ID

	// Assigned resource path for ctx
	ctx = contextWithResourcePath(ctx, strconv.Itoa(int(organization.ID)))

	grades, err := s.CreateGradeMasterForOrgazination(ctx, strconv.Itoa(int(organization.ID)))
	if err != nil {
		return ctx, err
	}
	stepState.GradeMasters = grades

	// Create schools 5 school with same school level
	schools, err := s.CreateSchoolsForOrganization(ctx, strconv.Itoa(int(organization.ID)))
	if err != nil {
		return ctx, err
	}
	stepState.Schools = schools

	if staffGrantedRole == "" {
		staffGrantedRole = helpers.StaffGrantedRoleSchoolAdmin
	}

	// Send state to context for create new user
	ctx = StepStateToContext(ctx, stepState)
	// Create a staff with granted role
	ctx, err = s.signedAsAccount(ctx, staffGrantedRole, int64(organization.ID), []string{organization.DefaultLocation.ID})
	if err != nil {
		return ctx, err
	}

	stepState = StepStateFromContext(ctx)

	// Assigned user_id for ctx
	ctx = contextWithUserID(ctx, stepState.CurrentUserID)

	organization.Staffs = []*entities.Staff{
		{
			ID:                 stepState.CurrentUserID,
			Token:              stepState.AuthToken,
			GrandtedRoles:      stepState.CurrentGrandtedRoles,
			UserGroup:          stepState.CurrentUserGroup,
			OrganizationIDs:    []int32{stepState.CurrentOrganicationID},
			GrantedLocationIDs: []string{organization.DefaultLocation.ID},
		},
	}

	stepState.Organization = organization
	stepState.CurrentStaff = organization.Staffs[0]

	// Assign AuthToken to ctx
	ctx = s.ContextWithToken(ctx, stepState.CurrentStaff.Token)

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) StaffGrantedRoleOfManabieOrganizationWithDefaultLocationLoggedInBackOffice(ctx context.Context, staffGrantedRole string) (context.Context, error) {
	organization, err := s.CreateNewOrganization(helpers.OrganizationTypeManabie)
	if err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.CurrentResourcePath = fmt.Sprint(organization.ID)
	stepState.CurrentOrganicationID = organization.ID

	// Assigned resource path for ctx
	ctx = contextWithResourcePath(ctx, strconv.Itoa(int(organization.ID)))

	grades, err := s.CreateGradeMasterForOrgazination(ctx, strconv.Itoa(int(organization.ID)))
	if err != nil {
		return ctx, err
	}
	stepState.GradeMasters = grades

	// Create schools 5 school with same school level
	schools, err := s.CreateSchoolsForOrganization(ctx, strconv.Itoa(int(organization.ID)))
	if err != nil {
		return ctx, err
	}
	stepState.Schools = schools

	if staffGrantedRole == "" {
		staffGrantedRole = helpers.StaffGrantedRoleSchoolAdmin
	}

	// Send state to context for create new user
	ctx = StepStateToContext(ctx, stepState)
	// Create a staff with granted role
	ctx, err = s.signedAsAccount(ctx, staffGrantedRole, int64(organization.ID), []string{organization.DefaultLocation.ID})
	if err != nil {
		return ctx, err
	}

	stepState = StepStateFromContext(ctx)

	// Assigned user_id for ctx
	ctx = contextWithUserID(ctx, stepState.CurrentUserID)

	organization.Staffs = []*entities.Staff{
		{
			ID:                 stepState.CurrentUserID,
			Token:              stepState.AuthToken,
			GrandtedRoles:      stepState.CurrentGrandtedRoles,
			UserGroup:          stepState.CurrentUserGroup,
			OrganizationIDs:    []int32{stepState.CurrentOrganicationID},
			GrantedLocationIDs: []string{organization.DefaultLocation.ID},
		},
	}

	stepState.Organization = organization
	stepState.CurrentStaff = organization.Staffs[0]

	// Assign AuthToken to ctx
	ctx = s.ContextWithToken(ctx, stepState.CurrentStaff.Token)

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) insertCurrentStaffIntoInternalUser(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	_, err := s.BobDBConn.Exec(ctx, `
		INSERT INTO notification_internal_user 
			(user_id, is_system, created_at, updated_at, deleted_at, resource_path)
		VALUES 
			($1, TRUE, now(), now(), NULL, autofillresourcepath())
	`, stepState.CurrentStaff.ID)
	if err != nil {
		return ctx, fmt.Errorf("err insert notification_internal_user: %v", err)
	}

	return ctx, nil
}

func (s *ConversationMgmtSuite) AdminUpdateCurrentStaffGrantedLocationsTo(ctx context.Context, locationFilter string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	locations := make([]string, 0)
	switch locationFilter {
	case "random":
		for i, location := range stepState.Organization.DescendantLocations {
			// nolint
			if i == 0 || rand.Intn(2) == 1 {
				locations = append(locations, location.ID)
			}
		}
	default:
		idxsLocsStr := strings.Split(locationFilter, ",")
		for _, idxLocStr := range idxsLocsStr {
			if idxLocStr == "default" {
				locations = append(locations, stepState.Organization.DefaultLocation.ID)
				continue
			}

			idxLoc, err := strconv.Atoi(idxLocStr)
			if err != nil {
				return ctx, fmt.Errorf("can't convert descendant location index: %v", err)
			}
			if idxLoc <= 0 || idxLoc > helpers.NumberOfNewCenterLocationCreated {
				return ctx, fmt.Errorf("index descendant location out of range")
			}
			locations = append(locations, stepState.Organization.DescendantLocations[idxLoc-1].ID)
		}
	}

	// currently our BDD only support 1 user group per staff
	query := `
		SELECT ug.user_group_id, ug.user_group_name, gr.role_id
		FROM user_group ug
			JOIN user_group_member ugm ON ugm.user_group_id = ug.user_group_id
			JOIN granted_role gr ON gr.user_group_id = ug.user_group_id 
		WHERE ugm.user_id = $1
		GROUP BY ug.user_group_id, ug.user_group_name, gr.role_id
	`
	var userGroupID, userGroupName, roleID string
	err := s.BobPostgresDBConn.QueryRow(ctx, query, stepState.CurrentUserID).Scan(&userGroupID, &userGroupName, &roleID)
	if err != nil {
		return ctx, fmt.Errorf("failed scan: %v", err)
	}

	if userGroupID == "" || userGroupName == "" || roleID == "" {
		return ctx, fmt.Errorf("invalid user_group_id, user_group_name or role_id")
	}

	req := &upbv2.UpdateUserGroupRequest{
		UserGroupId:   userGroupID,
		UserGroupName: userGroupName,
		RoleWithLocations: []*upbv2.RoleWithLocations{
			{
				RoleId:      roleID,
				LocationIds: locations,
			},
		},
	}

	adminToken := stepState.Organization.Staffs[0].Token
	_, err = upbv2.NewUserGroupMgmtServiceClient(s.UserMgmtGRPCConn).UpdateUserGroup(
		contextWithToken(ctx, adminToken),
		req,
	)

	if err != nil {
		return ctx, fmt.Errorf("failed UpdateUserGroup: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) CreateSomeStaffsWithSomeRolesAndGrantedOrgLevelLocationOfCurrentOrganization(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rolePool := []string{helpers.StaffGrantedRoleTeacher, helpers.StaffGrantedRoleCentreLead, helpers.StaffGrantedRoleHQStaff}

	// create 3 staffs
	for i := 0; i < len(rolePool); i++ {
		// use different context to make sure the main context will always from admin's
		tmpCtx := ctx
		// Create a staff with granted role
		tmpCtx, err := s.signedAsAccount(tmpCtx, rolePool[i], int64(stepState.CurrentOrganicationID), []string{stepState.Organization.DefaultLocation.ID})
		if err != nil {
			return tmpCtx, fmt.Errorf("failed create staff number %d", i+1)
		}

		tmpStepState := StepStateFromContext(tmpCtx)

		newStaff := &entities.Staff{
			ID:                 tmpStepState.CurrentUserID,
			Token:              tmpStepState.AuthToken,
			UserGroup:          tmpStepState.CurrentUserGroup,
			GrandtedRoles:      tmpStepState.CurrentGrandtedRoles,
			OrganizationIDs:    []int32{tmpStepState.CurrentOrganicationID},
			GrantedLocationIDs: []string{tmpStepState.Organization.DefaultLocation.ID},
		}

		// append the newly created staff to the main StepState
		stepState.Organization.Staffs = append(stepState.Organization.Staffs, newStaff)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *ConversationMgmtSuite) StaffWithRoleTeacher(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.Organization == nil || stepState.CurrentOrganicationID == 0 {
		return ctx, fmt.Errorf("current organization not exist in your test, please create it first")
	}

	grantedLocations := []string{}

	for _, location := range stepState.Organization.DescendantLocations {
		grantedLocations = append(grantedLocations, location.ID)
	}

	userGroupTeacherID, err := s.createUserGroupWithRoleNames(ctx, []string{constant.RoleTeacher}, grantedLocations, int64(stepState.CurrentOrganicationID))
	if err != nil {
		return ctx, fmt.Errorf("failed createUserGroupWithRoleNames: %+v", err)
	}

	userName := const_conversationmgmt.UserNameConditionToCreateAgoraUser + "-" + idutil.ULIDNow() + "@manabie.com"
	createStaffRequest := &upbv2.CreateStaffRequest{
		Staff: &upbv2.CreateStaffRequest_StaffProfile{
			Name:           userName,
			Email:          userName,
			Country:        cpb.Country_COUNTRY_VN,
			UserGroupIds:   []string{userGroupTeacherID},
			UserGroup:      upbv2.UserGroup_USER_GROUP_TEACHER,
			OrganizationId: fmt.Sprint(stepState.CurrentOrganicationID),
			LocationIds:    grantedLocations,
		},
	}

	resp, err := upbv2.NewStaffServiceClient(s.UserMgmtGRPCConn).CreateStaff(ctx, createStaffRequest)
	if err != nil {
		return ctx, fmt.Errorf("failed CreateStaff: %+v", err)
	}

	stepState.CurrentUserID = resp.Staff.StaffId
	token, err := s.generateExchangeToken(resp.Staff.StaffId, upbv2.UserGroup_USER_GROUP_TEACHER.String(), int64(stepState.CurrentOrganicationID))
	if err != nil {
		return ctx, fmt.Errorf("failed generateExchangeToken: %+v", err)
	}

	stepState.AuthToken = token
	stepState.CurrentUserGroup = upbv2.UserGroup_USER_GROUP_TEACHER.String()
	stepState.CurrentGrandtedRoles = []string{constant.RoleTeacher}

	// Assigned user_id for ctx
	ctx = contextWithUserID(ctx, stepState.CurrentUserID)

	newStaff := &entities.Staff{
		ID:                 stepState.CurrentUserID,
		Token:              stepState.AuthToken,
		UserGroup:          stepState.CurrentUserGroup,
		GrandtedRoles:      stepState.CurrentGrandtedRoles,
		OrganizationIDs:    []int32{stepState.CurrentOrganicationID},
		GrantedLocationIDs: []string{stepState.Organization.DefaultLocation.ID},
	}

	stepState.Organization.Staffs = append(stepState.Organization.Staffs, newStaff)
	stepState.CurrentStaff = newStaff

	// Assign AuthToken to ctx
	ctx = s.ContextWithToken(ctx, stepState.CurrentStaff.Token)

	return StepStateToContext(ctx, stepState), nil
}
