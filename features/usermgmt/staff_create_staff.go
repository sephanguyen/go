package usermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) generateACreateStaffProfile(ctx context.Context, requestMissingType string, locationKind string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request, stepState.ResponseErr = s.generateCreateStaffRequest(ctx, requestMissingType, locationKind)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateCreateStaffRequest(ctx context.Context, requestMissingType string, locationKind string) (*pb.CreateStaffRequest, error) {
	num := rand.Int()
	staff := &pb.CreateStaffRequest_StaffProfile{
		Name:        fmt.Sprintf("create_staff+%d", num),
		Email:       fmt.Sprintf("thanhdanh.nguyen+%d@manabie.com", num),
		Country:     cpb.Country_COUNTRY_VN,
		PhoneNumber: "",
		UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
		StaffPhoneNumber: []*pb.StaffPhoneNumber{
			{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
			{PhoneNumber: "987654321", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER},
		},
		Gender:        pb.Gender_MALE,
		Birthday:      timestamppb.New(time.Now().Add(-87600 * 10 * time.Hour)),
		WorkingStatus: pb.StaffWorkingStatus_AVAILABLE,
		StartDate:     timestamppb.New(time.Now()),
		EndDate:       timestamppb.New(time.Now().Add(87600 * time.Hour)),
		Remarks:       "Hello remarks",
		UserNameFields: &pb.UserNameFields{
			LastName:          fmt.Sprintf("last_name+%d", num),
			FirstName:         fmt.Sprintf("first_name+%d", num),
			LastNamePhonetic:  fmt.Sprintf("last_name_phonetic+%d", num),
			FirstNamePhonetic: fmt.Sprintf("first_name_phonetic+%d", num),
		},
		Username: fmt.Sprintf("username%d", num),
	}
	userGroupIDs, err := s.initUserGroups(ctx, requestMissingType)
	if err != nil {
		return &pb.CreateStaffRequest{Staff: staff}, err
	}
	if userGroupIDs != nil {
		staff.UserGroupIds = userGroupIDs
	}

	tagIDs, _, err := s.createAmountTags(ctx, 3, pb.UserTagType_USER_TAG_TYPE_STAFF.String(), fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return nil, fmt.Errorf("createAmountTags: %w", err)
	}
	if tagIDs != nil {
		staff.TagIds = tagIDs
	}

	switch requestMissingType {
	case "full field valid", "user group was granted teacher role", "user group was granted school admin role", "empty user_group_ids", "invalid user_group_ids":
		// Do nothing, initial data already satisfied these conditions
		break

	case "full field valid with only primary phone number":
		staff.StaffPhoneNumber = []*pb.StaffPhoneNumber{{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER}}

	case "empty email":
		staff.Email = ""

	case "empty country":
		staff.Country = cpb.Country_COUNTRY_NONE

	case "empty user_group":
		staff.UserGroup = pb.UserGroup_USER_GROUP_NONE

	case "empty optional field":
		staff.StaffPhoneNumber = []*pb.StaffPhoneNumber{}
		staff.Gender = pb.Gender_NONE
		staff.StartDate = nil
		staff.EndDate = nil
		staff.Birthday = nil
		staff.UserGroupIds = []string{}
		staff.Remarks = ""
		staff.ExternalUserId = ""
		staff.TagIds = []string{}

	case "wrong type phone number":
		staff.StaffPhoneNumber = []*pb.StaffPhoneNumber{{PhoneNumber: "123456789", PhoneNumberType: 3}}

	case "add two phone number have a same":
		staff.StaffPhoneNumber = []*pb.StaffPhoneNumber{
			{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
			{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER},
		}

	case "wrong type gender":
		staff.Gender = 4

	case "duplicated email":
		// create staff
		roleWithLocationTeacher := RoleWithLocation{
			RoleName:    constant.RoleTeacher,
			LocationIDs: []string{constants.ManabieOrgLocation},
		}
		resp, err := CreateStaff(ctx, s.BobDBTrace, s.UserMgmtConn, nil, []RoleWithLocation{roleWithLocationTeacher}, []string{s.ExistingLocations[0].ToLocationEntity().LocationID})
		if err != nil {
			return nil, fmt.Errorf("CreateStaff: %w", err)
		}
		// find user to get existed email
		userRepo := repository.UserRepo{}
		currentStaff, err := userRepo.Get(ctx, s.BobDBTrace, database.Text(resp.Staff.StaffId))
		if err != nil {
			return nil, fmt.Errorf("userRepo.GetProfile: %w", err)
		}
		staff.Email = currentStaff.Email.String

	case "existed external user id":
		// create staff
		roleWithLocationTeacher := RoleWithLocation{
			RoleName:    constant.RoleTeacher,
			LocationIDs: []string{constants.ManabieOrgLocation},
		}
		resp, err := CreateStaff(ctx, s.BobDBTrace, s.UserMgmtConn, nil, []RoleWithLocation{roleWithLocationTeacher}, []string{s.ExistingLocations[0].ToLocationEntity().LocationID})
		if err != nil {
			return nil, fmt.Errorf("CreateStaff: %w", err)
		}
		// find user to get existed external user id
		userRepo := repository.UserRepo{}
		currentStaff, err := userRepo.Get(ctx, s.BobDBTrace, database.Text(resp.Staff.StaffId))
		if err != nil {
			return nil, fmt.Errorf("userRepo.GetProfile: %w", err)
		}
		staff.ExternalUserId = currentStaff.ExternalUserID.String

	case "non existed external user id":
		staff.ExternalUserId = fmt.Sprintf("external_user_id+%d", num)

	case "non existed external user id with space":
		staff.ExternalUserId = fmt.Sprintf(" external_user_id+%d ", num)

	case "empty name and valid first and last name":
		staff.Name = ""
	case "valid name and empty first and last name":
		staff.UserNameFields = nil
	case "empty name and empty first and last name":
		staff.Name = ""
		staff.UserNameFields = nil
	case "non existing tag":
		staff.TagIds = []string{idutil.ULIDNow()}
	case "wrong tag type":
		tagIDs, _, err := s.createAmountTags(ctx, 3, pb.UserTagType_USER_TAG_TYPE_PARENT.String(), fmt.Sprint(constants.ManabieSchool))
		if err != nil {
			return nil, fmt.Errorf("createAmountTags: %w", err)
		}
		staff.TagIds = tagIDs
	case "available username":
		staff.Username = fmt.Sprintf("username" + idutil.ULIDNow())
	case "available username with email format":
		staff.Username = fmt.Sprintf("username+%s@manabie.com", idutil.ULIDNow())
	case "empty username":
		staff.Username = ""
	case "username has spaces":
		staff.Username = "Username has spaces"
	case "username has special characters":
		staff.Username = "Invalid_Username"
	case "existing username":
		// user was created in migration file 1-local-init-sql.sql
		staff.Username = "username_existing_01@gmail.com"
	case "existing username and upper case":
		// user was created in migration file 1-local-init-sql.sql
		staff.Username = "USERNAME_EXISTING_01@GMAIL.COM"
	default:
		return nil, fmt.Errorf("unsupported condition")
	}

	locationIDs := make([]string, 0)
	switch locationKind {
	case "empty locations":
		locationIDs = []string{}
	case "valid locations":
		// values stepState.ExistingLocations are [manabie, jprep] in features/usermgmt/util.go
		locationIDs = []string{s.ExistingLocations[0].LocationID.String}
	case "invalid locations":
		locationIDs = []string{s.ExistingLocations[1].LocationID.String}
	default:
		return nil, fmt.Errorf("unsupported location condition")
	}
	staff.LocationIds = locationIDs
	staff.OrganizationId = fmt.Sprint(constants.ManabieSchool)

	return &pb.CreateStaffRequest{Staff: staff}, nil
}

func (s *suite) createStaffAccount(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewStaffServiceClient(s.UserMgmtConn).CreateStaff(ctx, stepState.Request.(*pb.CreateStaffRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkStaffInIdentityPlatform(ctx context.Context, email, staffID, userGroup string, schoolID int64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	tenantClient, err := s.TenantManager.TenantClient(ctx, auth.LocalTenants[int(schoolID)])
	if err != nil {
		return ctx, errors.Wrap(err, "tenantClient()")
	}

	user, err := tenantClient.GetUser(ctx, staffID)
	if err != nil {
		return ctx, errors.Wrap(err, "tenantClient.GetUser")
	}

	if user.GetEmail() != email {
		return ctx, fmt.Errorf("email not match expected %s, got %s", email, user.GetEmail())
	}
	claims := utils.CustomUserClaims(userGroup, staffID, schoolID)
	if fmt.Sprint(user.GetCustomClaims()) != fmt.Sprint(claims) {
		return ctx, fmt.Errorf("custom claims not match expected %v, got %v", user.GetCustomClaims(), claims)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkStaffInDB(ctx context.Context, req *pb.CreateStaffRequest_StaffProfile) error {
	teacherRepo := repository.TeacherRepo{}
	schoolAdminRepo := repository.SchoolAdminRepo{}
	userRepo := repository.UserRepo{}

	userEns, err := userRepo.GetByEmail(ctx, s.BobDBTrace, database.TextArray([]string{req.Email}))
	if err != nil {
		return fmt.Errorf("GetByEmail with email: %s: %w", req.Email, err)
	}
	if len(userEns) == 0 {
		return fmt.Errorf("staff not found %s", req.Email)
	}
	currentStaffResourcePath := golibs.ResourcePathFromCtx(ctx)
	currentStaffResourcePathInt64, err := strconv.ParseInt(currentStaffResourcePath, 10, 64)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseInt")
	}

	for _, userEn := range userEns {
		if _, err := checkStaffInStaffTable(ctx, s.BobDB, userEn.GetUID()); err != nil {
			return errors.Wrapf(err, "can not find staff %s record in staff table", userEn.GetUID())
		}

		ctx, err = s.checkStaffInIdentityPlatform(ctx, req.Email, userEn.GetUID(), userEn.Group.String, currentStaffResourcePathInt64)
		if err != nil {
			return errors.Wrap(err, "s.checkStaffInIdentityPlatform")
		}
		firstName, lastName := helper.SplitNameToFirstNameAndLastName(req.Name)
		fullname := req.Name
		firstNamePhonetic := ""
		lastNamePhonetic := ""
		fullnamePhonetic := ""
		if req.UserNameFields != nil {
			firstName = req.UserNameFields.FirstName
			lastName = req.UserNameFields.LastName
			fullname = helper.CombineFirstNameAndLastNameToFullName(req.UserNameFields.FirstName, req.UserNameFields.LastName)
			firstNamePhonetic = req.UserNameFields.FirstNamePhonetic
			lastNamePhonetic = req.UserNameFields.LastNamePhonetic
			fullnamePhonetic = helper.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(req.UserNameFields.FirstNamePhonetic, req.UserNameFields.LastNamePhonetic)
		}
		var isEnableUsernameToggle bool
		isEnableUsernameToggle, err = isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleStaffUsername)
		if err != nil {
			return errors.Wrapf(err, "can not get feature toggle (%s)", pkg_unleash.FeatureToggleStaffUsername)
		}
		if err := assertUpsertUsername(isEnableUsernameToggle,
			assertUsername{
				requestUsername:    req.Username,
				requestEmail:       req.Email,
				databaseUsername:   userEn.UserName.String,
				requestLoginEmail:  req.Email,
				databaseLoginEmail: userEn.LoginEmail.String,
			},
		); err != nil {
			return err
		}

		switch {
		case userEn.Email.String != req.Email:
			return fmt.Errorf("email not match expected %s, got %s", req.Email, userEn.Email.String)
		case userEn.Country.String != req.Country.String():
			return fmt.Errorf("country not match expected %s, got %s", req.Country, userEn.Country.String)
		case fullname != userEn.FullName.String:
			return fmt.Errorf(`expected inserted "FullName": %v but actual is %v`, fullname, userEn.FullName.String)
		case firstName != userEn.FirstName.String:
			return fmt.Errorf(`expected inserted "FirstName": %v but actual is %v`, firstName, userEn.FirstName.String)
		case lastName != userEn.LastName.String:
			return fmt.Errorf(`expected inserted "LastName": %v but actual is %v`, lastName, userEn.LastName.String)
		case fullnamePhonetic != userEn.FullNamePhonetic.String:
			return fmt.Errorf(`expected inserted "FullnamePhonetic": %v but actual is %v`, fullnamePhonetic, userEn.FullNamePhonetic.String)
		case firstNamePhonetic != userEn.FirstNamePhonetic.String:
			return fmt.Errorf(`expected inserted "FirstNamePhonetic": %v but actual is %v`, firstNamePhonetic, userEn.FirstNamePhonetic.String)
		case lastNamePhonetic != userEn.LastNamePhonetic.String:
			return fmt.Errorf(`expected inserted "LastNamePhonetic": %v but actual is %v`, lastNamePhonetic, userEn.LastNamePhonetic.String)
		}
		query := `SELECT count(*) FROM users WHERE email=$1 and resource_path = $2;`
		var count int
		if err := s.BobDBTrace.QueryRow(ctx, query, &req.Email, &currentStaffResourcePath).Scan(&count); err != nil {
			return fmt.Errorf("error counting staff:%w", err)
		}
		if count != 1 {
			return fmt.Errorf("expecting 1 got %d currentStaffResourcePath %s", count, currentStaffResourcePath)
		}

		roles, err := userRepo.GetUserRoles(ctx, s.BobDBTrace, userEn.ID)
		if err != nil {
			return errors.Wrap(err, "userRepo.GetUserRoles")
		}

		for _, role := range roles {
			_, err = teacherRepo.FindByID(ctx, s.BobDBTrace, userEn.ID)
			foundTeacher := err == nil

			_, err = schoolAdminRepo.Get(ctx, s.BobDBTrace, userEn.ID)
			foundSchoolAdmin := err == nil

			switch role.RoleName.Get() {
			case constant.RoleTeacher:
				if !foundTeacher || foundSchoolAdmin {
					return errors.Errorf("create staff %s with role %s was not inserted correctly", userEn.GetUID(), constant.RoleTeacher)
				}

			case constant.RoleSchoolAdmin:
				if foundTeacher || !foundSchoolAdmin {
					return errors.Errorf("create staff %s with role %s was not inserted correctly", userEn.GetUID(), constant.RoleSchoolAdmin)
				}
			}
		}

		if err := multierr.Combine(
			checkLegacyUserGroup(ctx, s.BobDBTrace, userEn.GetUID(), userEn.Group.String),
			s.validateLocationStored(ctx, userEn.ID.String, req.GetLocationIds()),
		); err != nil {
			return err
		}

		// check user group for staff must be created in db
		if err := s.checkUserGroupMustExitedInDB(ctx, userEn.GetUID(), req.UserGroupIds, currentStaffResourcePath); err != nil {
			return fmt.Errorf("checkUserGroupMustExitedInDB: %w", err)
		}
		// check tags for staff must be created in db
		if len(req.TagIds) > 0 {
			if err := s.checkStaffTagMustExitedInDB(ctx, userEn.GetUID(), req.TagIds); err != nil {
				return fmt.Errorf("checkStaffTagMustExitedInDB: %w", err)
			}
		}
		staffRepo := repository.StaffRepo{}
		staffInDB, err := staffRepo.Find(ctx, s.BobDBTrace, userEn.ID)
		if err != nil {
			return fmt.Errorf("staffRepo.Find: %w", err)
		}
		switch {
		case staffInDB.WorkingStatus.String != req.WorkingStatus.String():
			return fmt.Errorf("checkStaffInDB working_status dont match, expect: %s, actual: %s", req.WorkingStatus.String(), staffInDB.WorkingStatus.String)
		case !checkMatchDate(req.StartDate, staffInDB.StartDate.Time):
			return fmt.Errorf("checkStaffInDB start_date dont match, expect: %s, actual: %s", req.StartDate.AsTime(), staffInDB.StartDate.Time)
		case !checkMatchDate(req.EndDate, staffInDB.EndDate.Time):
			return fmt.Errorf("checkStaffInDB end_date dont match, expect: %s, actual: %s", req.EndDate.AsTime(), staffInDB.EndDate.Time)
		case userEn.Remarks.String != req.Remarks:
			return fmt.Errorf("checkStaffInDB remarks don't match, expect: %s, actual: %s", req.Remarks, userEn.Remarks.String)
		case userEn.ExternalUserID.String != strings.TrimSpace(req.ExternalUserId):
			return fmt.Errorf("checkStaffInDB ExternalUserID don't match, expect: %s, actual: %s", req.ExternalUserId, userEn.ExternalUserID.String)
		case userEn.UserRole.String != string(constant.UserRoleStaff):
			return fmt.Errorf("checkStaffInDB user_role expect: %s, actual: %s", constant.UserRoleStaff, userEn.UserRole.String)
		}
	}

	return nil
}

// check legacy user group
func checkLegacyUserGroup(ctx context.Context, db database.QueryExecer, userID, currentUserGroup string) error {
	userGroup, err := getLegacyUserGroup(ctx, db, userID)
	if err != nil {
		return err
	}

	if currentUserGroup != userGroup {
		return fmt.Errorf("expect user group is %s but got %s", currentUserGroup, userGroup)
	}
	return nil
}

// find legacy user group
func getLegacyUserGroup(ctx context.Context, db database.QueryExecer, userID string) (string, error) {
	query := `
	SELECT ARRAY_AGG(group_id)
	FROM users_groups
	WHERE user_id = $1
		AND status = $2
		AND is_origin = $3
	`

	userGroups := []string{}
	err := db.QueryRow(ctx, query, userID, entity.UserGroupStatusActive, true).Scan(&userGroups)
	if err != nil {
		return "", fmt.Errorf("can not find legacy user group of user %s", userID)
	}
	if len(userGroups) != 1 {
		return "", fmt.Errorf("user %s must have single legacy user groups", userID)
	}
	return userGroups[0], nil
}

func (s *suite) newStaffAccountWasCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.CreateStaffRequest)

	err := s.checkStaffInDB(
		ctx,
		req.Staff,
	)
	if err != nil {
		return ctx, fmt.Errorf("checkStaffInDB: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUserGroupMustExitedInDB(ctx context.Context, userID string, userGroupIDs []string, resourcePath string) error {
	userGroupsMember := new(entity.UserGroupMember)
	userGroupMembers := make([]*entity.UserGroupMember, 0)
	fieldName, _ := userGroupsMember.FieldMap()
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE user_id = $1 AND user_group_id = ANY($2) AND deleted_at IS NULL`, strings.Join(fieldName, ","), userGroupsMember.TableName())

	rows, err := s.BobDBTrace.Query(ctx, query, &userID, userGroupIDs)
	if err != nil {
		return fmt.Errorf("error querying user_groups_member for user: %s: %w", userID, err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(database.GetScanFields(userGroupsMember, database.GetFieldNames(userGroupsMember))...)
		if err != nil {
			return fmt.Errorf("error scanning user_groups_member for user: %s: %w", userID, err)
		}
		userGroupMembers = append(userGroupMembers, userGroupsMember)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error in row user_groups_member for user: %s: %w", userID, err)
	}

	if len(userGroupIDs) != len(userGroupMembers) {
		return fmt.Errorf("userGroupMember not match expected %v, got %v", len(userGroupIDs), len(userGroupMembers))
	}

	// map for check user group id later
	mapUserGroupIDs := make(map[string]struct{})
	for _, userGroupID := range userGroupIDs {
		mapUserGroupIDs[userGroupID] = struct{}{}
	}

	for _, userGroupMember := range userGroupMembers {
		if userGroupMember.UserID.Get() != userID {
			return fmt.Errorf("userGroupMember user_id not match expected %s, got %s", userID, userGroupMember.UserID.Get())
		}

		if _, ok := mapUserGroupIDs[userGroupMember.UserGroupID.String]; !ok {
			return fmt.Errorf("userGroupMember not match expected %v, got %v", userGroupIDs, userGroupMember.UserGroupID.Get())
		}

		if userGroupMember.ResourcePath.Get() != resourcePath {
			return fmt.Errorf("userGroupMember resource_path not match expected %s, got %s", resourcePath, userGroupMember.ResourcePath.Get())
		}
	}

	return nil
}

func (s *suite) checkStaffTagMustExitedInDB(ctx context.Context, userID string, tagIDsInRequest []string) error {
	domainTaggedUserRepo := repository.DomainTaggedUserRepo{}
	mapExistTags := map[string]struct{}{}

	taggedUsers, err := domainTaggedUserRepo.GetByUserIDs(ctx, s.BobPostgresDB, []string{userID})
	if err != nil {
		return fmt.Errorf("checkStaffTagMustExitedInDB, DomainTaggedUserRepo.GetByUserIDs %v", err)
	}

	if len(tagIDsInRequest) != len(taggedUsers) {
		return fmt.Errorf("checkStaffTagMustExitedInDB len tags in request expect: %v, actual: %v", len(tagIDsInRequest), len(taggedUsers))
	}

	for _, taggedUser := range taggedUsers {
		mapExistTags[taggedUser.TagID().String()] = struct{}{}
	}

	for _, tagID := range tagIDsInRequest {
		if _, ok := mapExistTags[tagID]; !ok {
			return fmt.Errorf("staff %s is missing %s tag", userID, tagID)
		}
	}

	return nil
}

func checkStaffInStaffTable(ctx context.Context, db database.QueryExecer, staffID string) (*entity.Staff, error) {
	return new(repository.StaffRepo).FindByID(ctx, db, database.Text(staffID))
}

func createStaffReq(userGroupIDs []string, locationIDs []string) *pb.CreateStaffRequest {
	randomULID := idutil.ULIDNow()
	staff := &pb.CreateStaffRequest{
		Staff: &pb.CreateStaffRequest_StaffProfile{
			Name:        fmt.Sprintf("staff+%s", randomULID),
			Email:       fmt.Sprintf("staff+%s@gmail.com", randomULID),
			Country:     cpb.Country_COUNTRY_VN,
			PhoneNumber: "",
			UserGroup:   pb.UserGroup_USER_GROUP_TEACHER,
			StaffPhoneNumber: []*pb.StaffPhoneNumber{
				{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
			},
			Gender:         pb.Gender_MALE,
			Birthday:       timestamppb.New(time.Now().Add(-87600 * 10 * time.Hour)),
			WorkingStatus:  pb.StaffWorkingStatus_AVAILABLE,
			StartDate:      timestamppb.New(time.Now()),
			EndDate:        timestamppb.New(time.Now().Add(87600 * time.Hour)),
			Remarks:        "Hello remarks",
			UserGroupIds:   userGroupIDs,
			LocationIds:    locationIDs,
			ExternalUserId: fmt.Sprintf("external_user_id+%s", randomULID),
			Username:       fmt.Sprintf("username%s", randomULID),
		},
	}

	return staff
}
