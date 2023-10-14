package usermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) createRandomStaff(ctx context.Context, staffProfileOption string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	temCtx := s.signedIn(context.Background(), OrgIDFromCtx(ctx), StaffRoleSchoolAdmin)
	req, err := s.generateCreateStaffRequest(temCtx, "full field valid", "valid locations")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateACreateStaffProfile: %w", err)
	}

	switch staffProfileOption {
	case "user group was granted teacher role", "user group was granted school admin role", "empty user_group_ids", "invalid user_group_ids":
		userGroupIDs, err := s.initUserGroups(temCtx, staffProfileOption)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		req.Staff.UserGroupIds = userGroupIDs
	case "external_user_id-non-existed":
		req.Staff.ExternalUserId = fmt.Sprintf("other-external-user-id+%s", newID())
	case "empty external_user_id":
		break
	}

	createStaffResponse, createStaffResponseErr := pb.NewStaffServiceClient(s.UserMgmtConn).CreateStaff(temCtx, req)
	// if create staff request is failed, we don't need to assign staff id to update staff request
	if createStaffResponseErr == nil {
		stepState.CurrentTeacherID = createStaffResponse.GetStaff().GetStaffId()
		// store created user stepState.SrcUser for old email checking
		stepState.SrcUser = &entity.LegacyUser{
			ID:         database.Text(createStaffResponse.GetStaff().GetStaffId()),
			Email:      database.Text(createStaffResponse.Staff.GetEmail()),
			LoginEmail: database.Text(createStaffResponse.Staff.GetEmail()),
		}
	}
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aProfileOfStaffWithSpecificData(ctx context.Context, name, userGroupType, requestMissingType, locationKind string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := newID()
	createStaffRequest := stepState.Request.(*pb.CreateStaffRequest)
	email := createStaffRequest.GetStaff().GetEmail()
	existedEmail := ""
	externalUserID := createStaffRequest.GetStaff().GetExternalUserId()
	staffPhoneNumber := []*pb.StaffPhoneNumber{
		{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
		{PhoneNumber: "12345679", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER},
	}
	gender := pb.Gender_FEMALE
	birthday := timestamppb.New(time.Now().Add(-87600 * time.Hour))
	userNameField := &pb.UserNameFields{
		LastName:          fmt.Sprintf("last_name-%s", id),
		FirstName:         fmt.Sprintf("first_3name-%s", id),
		LastNamePhonetic:  fmt.Sprintf("last_name_phonetic+%s", id),
		FirstNamePhonetic: fmt.Sprintf("first_name_phonetic+%s", id),
	}
	startDate := timestamppb.New(time.Now())
	endDate := timestamppb.New(time.Now().Add(86 * time.Hour))
	username := createStaffRequest.GetStaff().GetUsername()

	temCtx := s.signedIn(context.Background(), OrgIDFromCtx(ctx), StaffRoleSchoolAdmin)
	if err := s.BobDBTrace.DB.QueryRow(temCtx, "SELECT email FROM users WHERE email IS NOT NULL LIMIT 1").Scan(&existedEmail); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can not find existed email: %w", err)
	}

	existedExternalUserID := ""
	if err := s.BobDBTrace.QueryRow(temCtx, "SELECT user_external_id FROM users WHERE user_external_id IS NOT NULL LIMIT 1").Scan(&existedExternalUserID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can not find existed user_external_id: %w", err)
	}

	tagIDs, _, err := s.createAmountTags(ctx, 3, pb.UserTagType_USER_TAG_TYPE_STAFF.String(), fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return nil, fmt.Errorf("createAmountTags: %w", err)
	}
	staffTagIds := []string{}
	staffTagIds = append(staffTagIds, tagIDs[0])

	switch requestMissingType {
	case "email-existed":
		email = existedEmail
	case "email-non-existed":
		email = fmt.Sprintf("other+%s+%s", strings.ToLower(newID()), existedEmail)
	case "email-empty":
		email = ""
	case "email-non-changed":
		email = createStaffRequest.GetStaff().GetEmail()
	case "empty-optional-field":
		staffPhoneNumber = []*pb.StaffPhoneNumber{}
		gender = pb.Gender_NONE
		birthday = nil
	case "empty-optional-field-only-primary-phone-number":
		staffPhoneNumber = []*pb.StaffPhoneNumber{
			{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
		}
	case "duplicated-phone-number":
		staffPhoneNumber = []*pb.StaffPhoneNumber{
			{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
			{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER},
		}
	case "phone-number-has-wrong-type":
		staffPhoneNumber = []*pb.StaffPhoneNumber{
			{PhoneNumber: "123456789", PhoneNumberType: 3},
		}
	case "phone-numbers-have-same-type":
		staffPhoneNumber = []*pb.StaffPhoneNumber{
			{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
			{PhoneNumber: "123456789", PhoneNumberType: pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
		}
	case "start-date-is-less-than-end-date":
		startDate = timestamppb.New(time.Now().Add(86 * time.Hour))
		endDate = timestamppb.New(time.Now())
	case "empty-name-and-empty-first&last-name":
		userNameField.LastName = ""
		userNameField.FirstName = ""
	case "add more tags":
		staffTagIds = append(staffTagIds, tagIDs[1])
	case "remove all tags":
		staffTagIds = []string{}
	case "non existing tag":
		staffTagIds = []string{idutil.ULIDNow()}
	case "wrong tag type":
		wrongTagIDs, _, err := s.createAmountTags(ctx, 3, pb.UserTagType_USER_TAG_TYPE_PARENT.String(), fmt.Sprint(constants.ManabieSchool))
		if err != nil {
			return nil, fmt.Errorf("createAmountTags: %w", err)
		}
		staffTagIds = wrongTagIDs
	case "external_user_id-existed":
		externalUserID = existedExternalUserID
	case "external_user_id-existed and space":
		externalUserID = fmt.Sprintf(" %s ", existedExternalUserID)
	case "external_user_id-non-existed":
		externalUserID = fmt.Sprintf("other+%s+%s", newID(), existedExternalUserID)
	case "external_user_id-non-existed and space":
		externalUserID = fmt.Sprintf(" other+%s+%s ", newID(), existedExternalUserID)
	case "available username":
		username = fmt.Sprintf("username%s", strings.ToLower(newID()))
	case "available username with email format":
		username = fmt.Sprintf("username_%s@manabie.com", strings.ToLower(newID()))
	case "empty username":
		username = ""
	case "username has spaces":
		username = fmt.Sprintf("username %s", strings.ToLower(newID()))
	case "username has special characters":
		username = "invalid_username"
	case "existing username":
		// user was created in migration file 1-local-init-sql.sql
		username = "username_existing_01@gmail.com"
	case "existing username and upper case":
		// user was created in migration file 1-local-init-sql.sql
		username = "USERNAME_EXISTING_01@GMAIL.COM"
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("this condition is not supported: %v", requestMissingType)
	}

	userGroupIDs, err := s.initUserGroups(temCtx, userGroupType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	locationIDs := []string{}
	// stepState.ExistingLocations = [0:manabie 1:jprep 2:manabie]
	switch locationKind {
	case "add more valid locations":
		locationIDs = append(locationIDs, stepState.ExistingLocations[2].LocationID.String)
	case "add more invalid locations":
		locationIDs = append(locationIDs, stepState.ExistingLocations[1].LocationID.String)
	case "remove all locations":
		locationIDs = []string{}
	case "invalid locations":
		locationIDs = []string{newID()}
	}

	stepState.Request = &pb.UpdateStaffRequest{
		Staff: &pb.UpdateStaffRequest_StaffProfile{
			StaffId:          stepState.CurrentTeacherID,
			Name:             name,
			Email:            email,
			UserGroupIds:     userGroupIDs,
			LocationIds:      locationIDs,
			StaffPhoneNumber: staffPhoneNumber,
			Gender:           gender,
			Birthday:         birthday,
			WorkingStatus:    pb.StaffWorkingStatus_AVAILABLE,
			StartDate:        startDate,
			EndDate:          endDate,
			Remarks:          "staff remarks",
			UserNameFields:   userNameField,
			TagIds:           staffTagIds,
			ExternalUserId:   externalUserID,
			Username:         username,
		},
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) profileOfStaffMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stepState.ResponseErr: %w", stepState.ResponseErr)
	}

	reqProfile := stepState.Request.(*pb.UpdateStaffRequest).Staff
	currentUserID := reqProfile.StaffId

	user := new(entity.LegacyUser)
	fieldName, values := user.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM users WHERE user_id = $1", strings.Join(fieldName, ","))
	err := s.BobDBTrace.QueryRow(ctx, query, &currentUserID).Scan(values...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error while querying user: %s: %w", currentUserID, err)
	}

	var staffInDB *entity.Staff
	if staffInDB, err = checkStaffInStaffTable(ctx, s.BobDB, currentUserID); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrapf(err, "can not find staff %s record in staff table", currentUserID)
	}
	userRepo := &repository.UserRepo{}
	userInDB, err := userRepo.Retrieve(ctx, s.BobDB, database.TextArray([]string{currentUserID}))
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "cannot user in database")
	}
	var isEnableUsernameToggle bool
	isEnableUsernameToggle, err = isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleStaffUsername)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrapf(err, "can not get feature toggle (%s)", pkg_unleash.FeatureToggleStaffUsername)
	}
	if err := assertUpsertUsername(isEnableUsernameToggle,
		assertUsername{
			requestUsername:    reqProfile.Username,
			requestEmail:       reqProfile.Email,
			databaseUsername:   staffInDB.UserName.String,
			requestLoginEmail:  reqProfile.Email,
			databaseLoginEmail: staffInDB.LoginEmail.String,
		},
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch {
	case staffInDB.WorkingStatus.String != reqProfile.WorkingStatus.String():
		return StepStateToContext(ctx, stepState), fmt.Errorf("staff working status was not updated correctly, DB: %v -  req: %v", staffInDB.WorkingStatus.String, reqProfile.WorkingStatus.String())
	case !checkMatchDate(reqProfile.StartDate, staffInDB.StartDate.Time):
		return StepStateToContext(ctx, stepState), fmt.Errorf("staff start date was not updated correctly, DB: %v -  req: %v", staffInDB.StartDate.Time, reqProfile.StartDate.AsTime())
	case !checkMatchDate(reqProfile.EndDate, staffInDB.EndDate.Time):
		return StepStateToContext(ctx, stepState), fmt.Errorf("staff start date was not updated correctly, DB: %v -  req: %v", staffInDB.EndDate.Time, reqProfile.EndDate.AsTime())
	case userInDB[0].Remarks.String != reqProfile.Remarks:
		return StepStateToContext(ctx, stepState), fmt.Errorf("remarks was not updated correctly, DB: %v -  req: %v", userInDB[0].Remarks.String, reqProfile.Remarks)
	case userInDB[0].UserRole.String != string(constant.UserRoleStaff):
		return StepStateToContext(ctx, stepState), fmt.Errorf("user_role expect: %s, actual: %s", constant.UserRoleStaff, userInDB[0].UserRole.String)
	}
	firstName, lastName := helper.SplitNameToFirstNameAndLastName(reqProfile.Name)
	fullname := reqProfile.Name
	firstNamePhonetic := ""
	lastNamePhonetic := ""
	fullnamePhonetic := ""
	if reqProfile.UserNameFields != nil {
		firstName = reqProfile.UserNameFields.FirstName
		lastName = reqProfile.UserNameFields.LastName
		fullname = helper.CombineFirstNameAndLastNameToFullName(reqProfile.UserNameFields.FirstName, reqProfile.UserNameFields.LastName)
		firstNamePhonetic = reqProfile.UserNameFields.FirstNamePhonetic
		lastNamePhonetic = reqProfile.UserNameFields.LastNamePhonetic
		fullnamePhonetic = helper.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(reqProfile.UserNameFields.FirstNamePhonetic, reqProfile.UserNameFields.LastNamePhonetic)
	}
	switch {
	case fullname != user.FullName.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`expected updated "FullName": %v but actual is %v`, fullname, user.FullName.String)
	case firstName != user.FirstName.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`expected updated "FirstName": %v but actual is %v`, firstName, user.FirstName.String)
	case lastName != user.LastName.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`expected updated "LastName": %v but actual is %v`, lastName, user.LastName.String)
	case fullnamePhonetic != user.FullNamePhonetic.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`expected updated "FullnamePhonetic": %v but actual is %v`, fullnamePhonetic, user.FullNamePhonetic.String)
	case firstNamePhonetic != user.FirstNamePhonetic.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`expected updated "FirstNamePhonetic": %v but actual is %v`, firstNamePhonetic, user.FirstNamePhonetic.String)
	case lastNamePhonetic != user.LastNamePhonetic.String:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`expected updated "LastNamePhonetic": %v but actual is %v`, lastNamePhonetic, user.LastNamePhonetic.String)
	}
	//TODO: add query to check phone number
	if reqProfile.Gender != 0 && reqProfile.Gender.String() != user.Gender.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("staff profile gender was not updated correctly, DB: %v -  req: %v", user.Gender, reqProfile.Gender)
	}

	if reqProfile.Birthday != nil {
		birthdayInDB := user.Birthday.Time.Format("2006-01-02")
		birthdayInReq := reqProfile.Birthday.AsTime().Format("2006-01-02")

		if birthdayInReq != birthdayInDB {
			return StepStateToContext(ctx, stepState), fmt.Errorf("staff profile birthday was not updated correctly, DB: %v -  req: %v", birthdayInDB, birthdayInReq)
		}
	}

	if strings.TrimSpace(reqProfile.ExternalUserId) != user.ExternalUserID.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("staff profile external_user_id was not updated correctly, DB: %v -  req: %v", user.ExternalUserID.String, reqProfile.ExternalUserId)
	}

	if len(reqProfile.UserGroupIds) == 0 {
		if user.Group.String != constant.UserGroupTeacher {
			return ctx, errors.New("must be teacher")
		}
	}

	if err := checkLegacyUserGroup(ctx, s.BobDBTrace, user.GetUID(), user.Group.String); err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrapf(err, "checkLegacyUserGroup for %s", user.GetUID())
	}

	// check user group of staff must be updated in db
	if err := s.checkUserGroupMustExitedInDB(ctx, currentUserID, reqProfile.GetUserGroupIds(), user.ResourcePath.String); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("checkUserGroupMustExitedInDB: %w", err)
	}

	if err := s.validateLocationStored(ctx, currentUserID, reqProfile.LocationIds); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// check after update with new email, old email must be not existed in DB
	newEmail := user.Email.String
	oldEmail := stepState.SrcUser.(*entity.LegacyUser).GetEmail()
	if newEmail != reqProfile.Email {
		return StepStateToContext(ctx, stepState), fmt.Errorf(`expected updated "Email": %v but actual is %v`, reqProfile.Email, newEmail)
	}
	ctx, err = checkOldEmailMustBeNotExisted(ctx, s.BobDBTrace, newEmail, oldEmail)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "checkOldEmailMustBeNotexisted")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) staffUpdateProfileCreatedProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewStaffServiceClient(s.UserMgmtConn).UpdateStaff(ctx, stepState.Request.(*pb.UpdateStaffRequest))

	return StepStateToContext(ctx, stepState), nil
}

func checkOldEmailMustBeNotExisted(ctx context.Context, db database.QueryExecer, newEmail, oldEmail string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	countedEmailUsed := 0
	if newEmail != oldEmail {
		// count this email is existed in users tables
		counted := 0
		if err := db.QueryRow(ctx, `select count(*) from users where email= $1`, oldEmail).Scan(&counted); err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "query email in users")
		}
		countedEmailUsed += counted

		// count this email is existed in usr_email tables
		if err := db.QueryRow(ctx, `select count(*) from usr_email where email= $1`, oldEmail).Scan(&counted); err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "query email in usr_email")
		}
		countedEmailUsed += counted
	}

	// after updating old email must be not existed in db
	if countedEmailUsed != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("email %s was used", oldEmail)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) initUserGroups(ctx context.Context, userGroupType string) ([]string, error) {
	switch userGroupType {
	case "invalid user_group_ids":
		return []string{newID()}, nil
	case "empty user_group_ids":
		return []string{}, nil
	case "user group was granted school admin role":
		roleWithLocation := RoleWithLocation{
			RoleName:    constant.RoleSchoolAdmin,
			LocationIDs: []string{constants.ManabieOrgLocation},
		}
		resp, err := CreateUserGroup(ctx, s.BobDBTrace, s.UserMgmtConn, nil, []RoleWithLocation{roleWithLocation})
		if err != nil {
			return nil, err
		}

		return []string{resp.UserGroupId}, nil
	case "user group was granted teacher role":
		roleWithLocation := RoleWithLocation{
			RoleName:    constant.RoleTeacher,
			LocationIDs: []string{s.ExistingLocations[0].LocationID.String},
		}
		resp, err := CreateUserGroup(ctx, s.BobDBTrace, s.UserMgmtConn, nil, []RoleWithLocation{roleWithLocation})
		if err != nil {
			return nil, err
		}
		return []string{resp.UserGroupId}, nil
	default:
		return nil, nil
	}
}
