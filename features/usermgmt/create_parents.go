package usermgmt

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	common "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *suite) createParentEventNatsJSSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleCreateUser := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &pb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := stepState.Request.(type) {
		case *pb.CreateParentsAndAssignToStudentRequest:
			switch msg := evtUser.Message.(type) {
			case *pb.EvtUser_CreateParent_:
				if req.StudentId == msg.CreateParent.StudentId {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		}
		return false, nil
	}
	subs, err := s.JSM.Subscribe(constants.SubjectUserCreated, opts, handleCreateUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createParentEventNatsJSSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createNewParents(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, account)

	stepState.RequestSentAt = time.Now()

	ctx, err := s.createParentEventNatsJSSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createParentEventNatsJSSubscription: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(ctx, stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createMultipleNewParents(ctx context.Context, account, times string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)
	res := &pb.CreateParentsAndAssignToStudentResponse{}
	parentProfiles := req.ParentProfiles

	for index := range parentProfiles {
		stepState.RequestSentAt = time.Now()
		req.ParentProfiles = []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{parentProfiles[index]}

		ctx, err := s.createParentEventNatsJSSubscription(ctx)
		if err != nil {
			return ctx, fmt.Errorf("s.createParentEventNatsJSSubscription: %w", err)
		}
		tmpRes, err := pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(ctx, req)
		if err != nil {
			stepState.ResponseErr = err
		}

		if tmpRes != nil {
			res.StudentId = tmpRes.StudentId
			res.ParentProfiles = append(res.ParentProfiles, tmpRes.ParentProfiles...)
		}
	}

	req.ParentProfiles = parentProfiles
	stepState.Request = req
	stepState.Response = res

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createNewParentsWithInvalidResourcePath(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	ctx, err := s.createParentEventNatsJSSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createParentEventNatsJSSubscription: %w", err)
	}
	ctx = s.signedIn(ctx, constants.JPREPSchool, account)
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(ctx, stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newParentsData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.addParentDataToCreateParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addParentDataToCreateParentReq(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	parentID := newID()
	stepState.ParentIDs = append(stepState.ParentIDs, parentID)
	profiles := []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
		{
			Name:         fmt.Sprintf("user-%v", parentID),
			CountryCode:  common.Country_COUNTRY_VN,
			PhoneNumber:  fmt.Sprintf("phone-number-%v", parentID),
			Email:        fmt.Sprintf("%v@example.com", parentID),
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
			Password:     fmt.Sprintf("password-%v", parentID),
			Username:     fmt.Sprintf("username%v", parentID),
			UserNameFields: &pb.UserNameFields{
				FirstName:         fmt.Sprintf("first-name-%v", parentID),
				LastName:          fmt.Sprintf("last-name-%v", parentID),
				FirstNamePhonetic: fmt.Sprintf("first-name-phonetic-%v", parentID),
				LastNamePhonetic:  fmt.Sprintf("last-name-phonetic-%v", parentID),
			},
			ParentPhoneNumbers: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "12345678",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
					PhoneNumberId:   "",
				},
				{
					PhoneNumber:     "098765432",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
					PhoneNumberId:   "",
				},
			},
			Remarks: "parent-remark",
		},
	}
	student, err := s.createStudent(ctx)
	if err != nil {
		return err
	}

	stepState.Request = &pb.CreateParentsAndAssignToStudentRequest{
		SchoolId:       constants.ManabieSchool,
		StudentId:      student.ID.String,
		ParentProfiles: profiles,
	}

	return nil
}

func NewCreateParentReqWithOnlyParentInfo(createdStudent *entity.LegacyStudent) *pb.CreateParentsAndAssignToStudentRequest {
	parentID := newID()
	profiles := []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
		{
			Name:         fmt.Sprintf("user-%v", parentID),
			CountryCode:  common.Country_COUNTRY_VN,
			PhoneNumber:  fmt.Sprintf("phone-number-%v", parentID),
			Email:        fmt.Sprintf("%v@example.com", parentID),
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
			Password:     fmt.Sprintf("password-%v", parentID),
			Username:     fmt.Sprintf("username%v", parentID),
			UserNameFields: &pb.UserNameFields{
				FirstName:         fmt.Sprintf("first-name-%v", parentID),
				LastName:          fmt.Sprintf("last-name-%v", parentID),
				FirstNamePhonetic: fmt.Sprintf("first-name-phonetic-%v", parentID),
				LastNamePhonetic:  fmt.Sprintf("last-name-phonetic-%v", parentID),
			},
			ParentPhoneNumbers: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "12345678",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
					PhoneNumberId:   "",
				},
				{
					PhoneNumber:     "098765432",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
					PhoneNumberId:   "",
				},
			},
		},
	}

	req := &pb.CreateParentsAndAssignToStudentRequest{
		ParentProfiles: profiles,
		StudentId:      createdStudent.ID.String,
	}

	return req
}

func (s *suite) newMultipleParentsData(ctx context.Context, times string, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	intTimes, err := strconv.Atoi(times)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("strconv.ParseInt: %w", err)
	}
	req, err := s.addMultipleParentDataToCreateParentReq(ctx, intTimes)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	for _, parentProfile := range req.ParentProfiles {
		switch condition {
		case "without first name phonetic":
			parentProfile.UserNameFields.FirstNamePhonetic = ""
		case "without last name phonetic":
			parentProfile.UserNameFields.LastNamePhonetic = ""
		case "without parent primary phone number":
			parentProfile.ParentPhoneNumbers = parentProfile.ParentPhoneNumbers[:0]
		case "without parent secondary phone number":
			parentProfile.ParentPhoneNumbers = parentProfile.ParentPhoneNumbers[:len(parentProfile.ParentPhoneNumbers)-1]
		case "without parent phone number":
			parentProfile.ParentPhoneNumbers = nil
		case "without remark":
			parentProfile.Remarks = ""
		case "available username":
			parentProfile.Username = "username" + newID()
		}
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addMultipleParentDataToCreateParentReq(ctx context.Context, times int) (*pb.CreateParentsAndAssignToStudentRequest, error) {
	stepState := StepStateFromContext(ctx)
	profiles := []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{}

	for i := 0; i < times; i++ {
		parentID := fmt.Sprint(times) + newID()
		stepState.ParentIDs = append(stepState.ParentIDs, parentID)

		newProfile := pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
			Name:           "",
			ExternalUserId: fmt.Sprintf("external-user-id-%v", parentID),
			CountryCode:    common.Country_COUNTRY_VN,
			PhoneNumber:    fmt.Sprintf("phone-number-%v", parentID),
			Email:          fmt.Sprintf("%v@example.com", parentID),
			Relationship:   pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
			Password:       fmt.Sprintf("password-%v", parentID),
			Username:       fmt.Sprintf("username%v", parentID),
			UserNameFields: &pb.UserNameFields{
				FirstName:         fmt.Sprintf("first-name-%v", parentID),
				LastName:          fmt.Sprintf("last-name-%v", parentID),
				FirstNamePhonetic: fmt.Sprintf("first-name-phonetic-%v", parentID),
				LastNamePhonetic:  fmt.Sprintf("last-name-phonetic-%v", parentID),
			},
			ParentPhoneNumbers: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "12345678",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
					PhoneNumberId:   "",
				},
				{
					PhoneNumber:     "098765432",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
					PhoneNumberId:   "",
				},
			},
		}
		profiles = append(profiles, &newProfile)
	}

	student, err := s.createStudent(ctx)
	if err != nil {
		return nil, err
	}

	req := &pb.CreateParentsAndAssignToStudentRequest{
		SchoolId:       constants.ManabieSchool,
		StudentId:      student.ID.String,
		ParentProfiles: profiles,
	}

	return req, nil
}

func (s *suite) newParentsWereCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	if err := s.validCreatedParentInfo(ctx); err != nil {
		return ctx, fmt.Errorf("validCreatedParentInfo: %s", err.Error())
	}

	select {
	case <-stepState.FoundChanForJetStream:
		return StepStateToContext(ctx, stepState), nil
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s suite) validCreatedParentInfo(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	currentResourcePath := golibs.ResourcePathFromCtx(ctx)

	req := stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)
	res := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse)

	if len(req.ParentProfiles) != len(res.ParentProfiles) {
		return errors.New("number of parents in request is different with response")
	}

	tagIDs := []string{}
	parentIDs := []string{}

	for _, parentProfile := range res.ParentProfiles {
		parentIDs = append(parentIDs, parentProfile.Parent.UserProfile.UserId)
		tagIDs = append(tagIDs, parentProfile.TagIds...)
	}

	parentIdsParam := database.TextArray(parentIDs)

	stmt := `
		SELECT 
			users.user_id,
			users.email,
			users.username,
			users.name,
			users.first_name,
			users.last_name,
			users.first_name_phonetic,
			users.last_name_phonetic,
			users.full_name_phonetic,
			users.country,
			users.phone_number,
			parents.school_id,
			parents.resource_path,
			users.user_external_id,
			users.login_email,
			users.user_role
		FROM users
		JOIN parents ON users.user_id = parents.parent_id  
		JOIN users_groups ON users.user_id = users_groups.user_id
		WHERE users.user_id = ANY($1)
		`
	rows, err := s.BobDBTrace.Query(ctx, stmt, parentIdsParam)
	if err != nil {
		return errors.Wrap(err, "query parents")
	}
	defer rows.Close()

	createdParentsIDMap := map[string]*entity.Parent{}
	createdParentsEmailMap := map[string]*entity.Parent{}
	createdParentIDWithTags := map[string]map[string]struct{}{}

	for rows.Next() {
		createdParent := &entity.Parent{}
		err := rows.Scan(
			&createdParent.ID,
			&createdParent.Email,
			&createdParent.UserName,
			&createdParent.GivenName,
			&createdParent.FirstName,
			&createdParent.LastName,
			&createdParent.FirstNamePhonetic,
			&createdParent.LastNamePhonetic,
			&createdParent.FullNamePhonetic,
			&createdParent.Country,
			&createdParent.PhoneNumber,
			&createdParent.SchoolID,
			&createdParent.ResourcePath,
			&createdParent.ExternalUserID,
			&createdParent.LoginEmail,
			&createdParent.UserRole,
		)
		if err != nil {
			return errors.WithMessage(err, "rows.Scan created parent")
		}
		createdParentsIDMap[createdParent.ID.String] = createdParent
		createdParentsEmailMap[createdParent.Email.String] = createdParent
	}

	// Check tagIDs
	rows, err = s.BobDBTrace.Query(
		ctx,
		`
			SELECT user_id, tag_id
			FROM tagged_user
			WHERE user_id = ANY($1) 
				AND tag_id = ANY($2)
				AND deleted_at IS NULL
		`,
		database.TextArray(parentIDs),
		database.TextArray(tagIDs),
	)
	if err != nil {
		return errors.Wrap(err, "query parents")
	}
	defer rows.Close()
	for rows.Next() {
		parentID := ""
		tagID := ""

		if err := rows.Scan(&parentID, &tagID); err != nil {
			return errors.WithMessage(err, "rows.Scan created parent")
		}

		if _, ok := createdParentIDWithTags[parentID]; !ok {
			createdParentIDWithTags[parentID] = map[string]struct{}{}
		}
		createdParentIDWithTags[parentID][tagID] = struct{}{}
	}
	// TODO: add code check user_phone_number
	for _, parentInReq := range req.ParentProfiles {
		var createdRelationship string

		createdParent, exists := createdParentsEmailMap[parentInReq.Email]
		if !exists {
			return fmt.Errorf(`can't found user with email: "%v"`, parentInReq.Email)
		}

		row := s.BobDBTrace.QueryRow(
			ctx,
			`
				SELECT relationship 
				FROM student_parents 
				WHERE student_id = $1 AND parent_id = $2
			`,
			res.StudentId,
			createdParent.ID,
		)

		if err := row.Scan(&createdRelationship); err != nil {
			return errors.WithMessage(err, "rows.Scan relationship")
		}

		if parentInReq.Relationship.String() != createdRelationship {
			return fmt.Errorf(`expected inserted "relationship": %v but actual is %v`, parentInReq.Relationship, createdRelationship)
		}

		// Inserted parent data should have values equal to request values
		if parentInReq.Email != createdParent.Email.String {
			return fmt.Errorf(`expected inserted "email": %v but actual is %v`, parentInReq.Email, createdParent.Email.String)
		}
		isEnableUsername, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleUserNameStudentParent)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Get feature toggle error(%s)", pkg_unleash.FeatureToggleUserNameStudentParent))
		}
		if err := assertUpsertUsername(isEnableUsername,
			assertUsername{
				requestUsername:    parentInReq.Username,
				requestEmail:       parentInReq.Email,
				databaseUsername:   createdParent.UserName.String,
				requestLoginEmail:  createdParent.ID.String + constant.LoginEmailPostfix,
				databaseLoginEmail: createdParent.LoginEmail.String,
			},
		); err != nil {
			return err
		}
		if helper.CombineFirstNameAndLastNameToFullName(parentInReq.UserNameFields.FirstName, parentInReq.UserNameFields.LastName) != createdParent.GivenName.String {
			return fmt.Errorf(`expected inserted "given_name": %v but actual is %v`, helper.CombineFirstNameAndLastNameToFullName(parentInReq.UserNameFields.FirstName, parentInReq.UserNameFields.LastName), createdParent.GivenName.String)
		}
		if strings.TrimSpace(helper.CombineFirstNameAndLastNameToFullName(parentInReq.UserNameFields.FirstNamePhonetic, parentInReq.UserNameFields.LastNamePhonetic)) != createdParent.FullNamePhonetic.String {
			return fmt.Errorf(`expected inserted "FullNamePhonetic": %v but actual is %v`, helper.CombineFirstNameAndLastNameToFullName(parentInReq.UserNameFields.FirstNamePhonetic, parentInReq.UserNameFields.LastNamePhonetic), createdParent.FullNamePhonetic.String)
		}
		if parentInReq.UserNameFields.FirstName != createdParent.FirstName.String {
			return fmt.Errorf(`expected inserted "FirstName": %v but actual is %v`, parentInReq.UserNameFields.FirstName, createdParent.FirstName.String)
		}
		if parentInReq.UserNameFields.LastName != createdParent.LastName.String {
			return fmt.Errorf(`expected inserted "LastName": %v but actual is %v`, parentInReq.UserNameFields.LastName, createdParent.LastName.String)
		}
		if parentInReq.UserNameFields.FirstNamePhonetic != createdParent.FirstNamePhonetic.String {
			return fmt.Errorf(`expected inserted "FirstNamePhonetic": %v but actual is %v`, parentInReq.UserNameFields.FirstNamePhonetic, createdParent.FirstNamePhonetic.String)
		}
		if parentInReq.UserNameFields.LastNamePhonetic != createdParent.LastNamePhonetic.String {
			return fmt.Errorf(`expected inserted "LastNamePhonetic": %v but actual is %v`, parentInReq.UserNameFields.LastNamePhonetic, createdParent.LastNamePhonetic.String)
		}
		if parentInReq.CountryCode.String() != createdParent.Country.String {
			return fmt.Errorf(`expected inserted "country": %v but actual is %v`, parentInReq.CountryCode, createdParent.Country.String)
		}
		if parentInReq.PhoneNumber != createdParent.PhoneNumber.String {
			return fmt.Errorf(`expected inserted "phone_number": '%v' but actual is '%v'`, parentInReq.PhoneNumber, createdParent.PhoneNumber.String)
		}
		if req.SchoolId != createdParent.SchoolID.Int {
			return fmt.Errorf(`expected inserted "school_id": %v but actual is %v`, req.SchoolId, createdParent.SchoolID.Int)
		}

		if createdParent.ResourcePath.String != currentResourcePath {
			return fmt.Errorf(`expected "resource_path": %v but actual is %v`, currentResourcePath, createdParent.ResourcePath.String)
		}

		if createdParent.ExternalUserID.String != parentInReq.ExternalUserId {
			return fmt.Errorf(`expected "external_user_id": %v but actual is %v`, parentInReq.ExternalUserId, createdParent.ExternalUserID.String)
		}

		if createdParent.UserRole.String != string(constant.UserRoleParent) {
			return fmt.Errorf(`expected "user_role": %v but actual is %v`, constant.UserRoleParent, createdParent.UserRole.String)
		}
		if len(parentInReq.TagIds) > 0 {
			parentID := createdParent.ID.String

			tagIDs, ok := createdParentIDWithTags[parentID]
			if !ok {
				return fmt.Errorf("expected parent %s have %v tags but does not have anything", parentID, tagIDs)
			}

			for _, tagID := range parentInReq.TagIds {
				if _, ok := tagIDs[tagID]; !ok {
					return fmt.Errorf("expected parent %s have %v tag but does not have anything", parentID, tagID)
				}
			}
		}

		if err := s.validateUsersHasUserGroupWithRole(ctx, []string{createdParent.ID.String}, currentResourcePath, constant.RoleParent); err != nil {
			return errors.WithMessage(err, "validateUserHasUserGroupWithRole")
		}
		email := parentInReq.Email
		if isEnableUsername {
			email = createdParent.ID.String + constant.LoginEmailPostfix
		}
		if err := s.loginIdentityPlatform(ctx, auth.LocalTenants[constants.ManabieSchool], email, parentInReq.Password); err != nil {
			return errors.Wrap(err, "loginIdentityPlatform")
		}
	}
	if err := s.validParentAccessPath(ctx, req.StudentId); err != nil {
		return errors.Wrap(err, "validParentAccessPath")
	}

	return nil
}

func (s *suite) parentDataHasEmptyOrInvalid(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.addParentDataToCreateParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)
	switch field {
	case "studentID empty":
		req.StudentId = ""
	case "studentID not exist":
		req.StudentId = newID()
	}

	for _, parentProfile := range req.ParentProfiles {
		parentProfile.PhoneNumber = ""
		switch field {
		case "empty email":
			parentProfile.Email = ""
		case "username":
			parentProfile.Email = ""
		case "password":
			parentProfile.Password = ""
		case "name":
			parentProfile.Name = ""
			parentProfile.UserNameFields = nil // feature flag was turned off
		case "first name":
			parentProfile.UserNameFields.FirstName = ""
		case "last name":
			parentProfile.UserNameFields.LastName = ""
		case "country code":
			parentProfile.CountryCode = common.Country(999999)
		case "parentPhoneNumber":
			parentProfile.ParentPhoneNumbers = []*pb.ParentPhoneNumber{
				{PhoneNumber: "abc", PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER},
			}
		case "relationship":
			max := int32(math.MinInt32)
			for _, value := range pb.FamilyRelationship_value {
				if value > max {
					max = value
				}
			}
			parentProfile.Relationship = pb.FamilyRelationship(max + 1)
		case "tag for only student":
			tagIDs, _, err := s.createTagsType(ctx, studentType)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			parentProfile.TagIds = tagIDs
		case "empty username":
			parentProfile.Username = ""
		case "username has spaces":
			parentProfile.Username = "  "
		case "username has special characters":
			parentProfile.Username = "_"
		case "existing username":
			// because getUsernameByUsingOther will override the request
			// so we need to save the request first
			request := s.Request

			existingUsername, err := s.getUsernameByUsingOther(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "getUsernameByUsingOther")
			}

			s.Request = request
			parentProfile.Username = existingUsername
		case "existing username and upper case":
			// because getUsernameByUsingOther will override the request
			// so we need to save the request first
			request := s.Request

			existingUsername, err := s.getUsernameByUsingOther(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "getUsernameByUsingOther")
			}

			s.Request = request
			parentProfile.Username = strings.ToUpper(existingUsername)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) inOrganizationCreateParent(ctx context.Context, signedUser string, orgOrdinal, userOrdinal int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Create student first
	student, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := NewCreateParentReqWithOnlyParentInfo(student)

	stepState.Request1 = req

	resp, err := pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response1 = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) inOrganizationCreateParentWithTheSameAsParent(ctx context.Context, signedUser string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// create student and parent in organization 2
	ctx = s.signedIn(ctx, constants.TestingSchool, signedUser)

	// Create student first
	student, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := NewCreateParentReqWithOnlyParentInfo(student)

	req.ParentProfiles[0].Email = stepState.Request1.(*pb.CreateParentsAndAssignToStudentRequest).ParentProfiles[0].Email

	stepState.Request2 = req

	resp, err := pb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response2 = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentWillBeCreatedSuccessfullyAndBelongedToOrganization(ctx context.Context, arg1, arg2 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	isEnableUsername, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleUserNameStudentParent)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, fmt.Sprintf("Get feature toggle error(%s)", pkg_unleash.FeatureToggleUserNameStudentParent))
	}
	resp1 := stepState.Response1.(*pb.CreateParentsAndAssignToStudentResponse)
	parentEmail1 := resp1.ParentProfiles[0].Parent.UserProfile.Email
	if isEnableUsername {
		parentEmail1 = resp1.ParentProfiles[0].Parent.UserProfile.UserId + constant.LoginEmailPostfix
	}
	err = LoginIdentityPlatform(ctx, s.Cfg.FirebaseAPIKey, auth.LocalTenants[constants.ManabieSchool], parentEmail1, resp1.ParentProfiles[0].ParentPassword)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp2 := stepState.Response2.(*pb.CreateParentsAndAssignToStudentResponse)
	parentEmail2 := resp2.ParentProfiles[0].Parent.UserProfile.Email
	if isEnableUsername {
		parentEmail2 = resp2.ParentProfiles[0].Parent.UserProfile.UserId + constant.LoginEmailPostfix
	}
	err = LoginIdentityPlatform(ctx, s.Cfg.FirebaseAPIKey, auth.LocalTenants[constants.TestingSchool], parentEmail2, resp2.ParentProfiles[0].ParentPassword)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) assignTagsToParentData(ctx context.Context, tagType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)

	tagIDs, _, err := s.createAmountTags(ctx, amountSampleTestElement, tagType, fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, parentProfile := range req.GetParentProfiles() {
		parentProfile.TagIds = tagIDs
	}

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validParentAccessPath(ctx context.Context, studentID string) error {
	studentAccessPaths, err := (&repository.UserAccessPathRepo{}).FindLocationIDsFromUserID(ctx, s.BobDBTrace, studentID)
	if err != nil {
		return err
	}

	parentStmt := `SELECT location_id FROM user_access_paths
	WHERE user_id IN (
		SELECT parent_id FROM student_parents
		WHERE student_id = $1
	)`
	rows, err := s.BobDBTrace.Query(ctx, parentStmt, studentID)
	if err != nil {
		return errors.Wrap(err, "query user_access_path get location ID from parentID")
	}

	parentLocationIDs := []string{}
	for rows.Next() {
		var parentLocationID string

		err := rows.Scan(&parentLocationID)
		if err != nil {
			return errors.WithMessage(err, "rows.Scan get parent location IDs")
		}
		parentLocationIDs = append(parentLocationIDs, parentLocationID)
	}

	for _, locationID := range studentAccessPaths {
		if !golibs.InArrayString(locationID, parentLocationIDs) {
			return errors.Wrap(err, "student location and parent location didn't match")
		}
	}

	return nil
}

func createParentReq(studentID string) *pb.CreateParentsAndAssignToStudentRequest {
	parentID := newID()
	profiles := []*pb.CreateParentsAndAssignToStudentRequest_ParentProfile{
		{
			Name:         fmt.Sprintf("user-%v", parentID),
			CountryCode:  common.Country_COUNTRY_VN,
			PhoneNumber:  fmt.Sprintf("phone-number-%v", parentID),
			Email:        fmt.Sprintf("%v@example.com", parentID),
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
			Password:     fmt.Sprintf("password-%v", parentID),
			Username:     fmt.Sprintf("username%v", parentID),
			UserNameFields: &pb.UserNameFields{
				FirstName:         fmt.Sprintf("first-name-%v", parentID),
				LastName:          fmt.Sprintf("last-name-%v", parentID),
				FirstNamePhonetic: fmt.Sprintf("first-name-phonetic-%v", parentID),
				LastNamePhonetic:  fmt.Sprintf("last-name-phonetic-%v", parentID),
			},
			ParentPhoneNumbers: []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "12345678",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
					PhoneNumberId:   "",
				},
				{
					PhoneNumber:     "098765432",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
					PhoneNumberId:   "",
				},
			},
		},
	}

	req := &pb.CreateParentsAndAssignToStudentRequest{
		ParentProfiles: profiles,
		StudentId:      studentID,
	}

	return req
}
