package usermgmt

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *suite) updateParentData(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.UpdateParentsAndFamilyRelationshipRequest)

	for _, parentProfile := range req.ParentProfiles {
		switch field {
		case "full fields":
			parentProfile.ExternalUserId = fmt.Sprintf("external-user-id-%s", newID())
		case "available username":
			parentProfile.Username = fmt.Sprintf("username%s", newID())
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createSubscriptionUpdateParents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleUpdateParents := func(ctx context.Context, data []byte) (bool, error) {
		evtUser := &pb.EvtUser{}
		if err := proto.Unmarshal(data, evtUser); err != nil {
			return false, err
		}

		switch req := stepState.Request.(type) {
		case *pb.UpdateParentsAndFamilyRelationshipRequest:
			switch msg := evtUser.Message.(type) {
			case *pb.EvtUser_ParentAssignedToStudent_:
				foundParentID := func(evtParentID string, parents []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile) bool {
					for _, parent := range parents {
						if evtParentID == parent.Id {
							return true
						}
					}
					return false
				}(msg.ParentAssignedToStudent.ParentId, req.ParentProfiles)
				if req.StudentId == msg.ParentAssignedToStudent.StudentId && foundParentID {
					stepState.FoundChanForJetStream <- evtUser.Message
					return true, nil
				}
			}
		}
		return false, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectUserUpdated, opts, handleUpdateParents)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createSubscriptionUpdateParents: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateNewParents(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	ctx, err = s.createSubscriptionUpdateParents(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createSubscriptionUpdateParents: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewUserModifierServiceClient(s.UserMgmtConn).UpdateParentsAndFamilyRelationship(ctx, stepState.Request.(*pb.UpdateParentsAndFamilyRelationshipRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentsDataToUpdate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.addParentDataToUpdateParentReq(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addParentDataToUpdateParentReq(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	err := s.addParentDataToCreateParentReq(ctx)
	if err != nil {
		return err
	}

	if _, err := s.createNewParents(ctx, schoolAdminType); err != nil {
		return err
	}
	if stepState.ResponseErr != nil {
		return stepState.ResponseErr
	}

	tagIDs, _, err := s.createAmountTags(ctx, amountSampleTestElement, pb.UserTagType_USER_TAG_TYPE_PARENT_DISCOUNT.String(), fmt.Sprint(constants.ManabieSchool))
	if err != nil {
		return err
	}

	parentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).ParentProfiles[0].Parent.UserProfile.UserId
	studentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).StudentId

	stepState.ParentIDs = append(stepState.ParentIDs, parentID)
	stepState.ParentPassword = stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).ParentProfiles[0].ParentPassword
	profiles := []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
		{
			Id:           parentID,
			Email:        fmt.Sprintf("updated-%v@example.com", newID()),
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
			TagIds:       tagIDs,
			Username:     "username" + newID(),
			UserNameFields: &pb.UserNameFields{
				FirstName:         fmt.Sprintf("first_name + edited %v", parentID),
				LastName:          fmt.Sprintf("last_name + edited %v", parentID),
				FirstNamePhonetic: fmt.Sprintf("first_name_phonetic + edited %v", parentID),
				LastNamePhonetic:  fmt.Sprintf("last_name_phonetic + edited %v", parentID),
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
			Remarks: "parent-remark + edited",
		},
	}

	stepState.TagIDs = tagIDs
	stepState.Request = &pb.UpdateParentsAndFamilyRelationshipRequest{
		SchoolId:       constants.ManabieSchool,
		StudentId:      studentID,
		ParentProfiles: profiles,
	}

	return nil
}

func (s *suite) parentsWereUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	if err := s.validateUpdatedParentInfo(ctx); err != nil {
		return ctx, fmt.Errorf("validateUpdatedParentInfo: %s", err.Error())
	}

	select {
	case <-stepState.FoundChanForJetStream:
		return StepStateToContext(ctx, stepState), nil
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) validateUpdatedParentInfo(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.UpdateParentsAndFamilyRelationshipRequest)
	res := stepState.Response.(*pb.UpdateParentsAndFamilyRelationshipResponse)

	parentRepo := repository.ParentRepo{}

	if len(req.ParentProfiles) != len(res.ParentProfiles) {
		return errors.New("number of parents in request is different with response")
	}

	parentInRespMap := map[string]*pb.UpdateParentsAndFamilyRelationshipResponse_ParentProfile{}
	parentIDs := make([]string, 0, len(res.ParentProfiles))
	// parentIdsParam := new(pgtype.TextArray)

	for _, parentProfile := range res.ParentProfiles {
		parentInRespMap[parentProfile.Parent.UserProfile.UserId] = parentProfile
		parentIDs = append(parentIDs, parentProfile.Parent.UserProfile.UserId)
	}

	parentIdsParam := database.TextArray(parentIDs)

	parentEntities, err := parentRepo.GetByIds(ctx, s.BobDB, parentIdsParam)

	if err != nil {
		return err
	}

	updatedParentsIDMap := map[string]*entity.Parent{}
	updatedParentsEmailMap := map[string]*entity.Parent{}

	for _, parentEntity := range parentEntities {
		updatedParentsIDMap[parentEntity.ID.String] = parentEntity
		updatedParentsEmailMap[parentEntity.Email.String] = parentEntity
	}

	for _, parentInReq := range req.ParentProfiles {
		var updatedRelationship string

		updatedParent, exists := updatedParentsEmailMap[parentInReq.Email]
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
			updatedParent.ID,
		)

		err := row.Scan(&updatedRelationship)
		if err != nil {
			return errors.WithMessage(err, "rows.Scan relationship")
		}

		if parentInReq.Relationship.String() != updatedRelationship {
			return fmt.Errorf(`expected updated "relationship": %v but actual is %v`, parentInReq.Relationship, updatedRelationship)
		}

		// Updated parent data should have values equal to request values
		if parentInReq.Id != updatedParent.ID.String {
			return fmt.Errorf(`expected updated "id": %v but actual is %v`, parentInReq.Id, updatedParent.ID.String)
		}
		if parentInReq.Email != updatedParent.Email.String {
			return fmt.Errorf(`expected updated "email": %v but actual is %v`, parentInReq.Email, updatedParent.Email.String)
		}
		isEnableUsername, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleUserNameStudentParent)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Get feature toggle error(%s)", pkg_unleash.FeatureToggleUserNameStudentParent))
		}
		if err := assertUpsertUsername(isEnableUsername,
			assertUsername{
				requestUsername:    parentInReq.Username,
				requestEmail:       parentInReq.Email,
				databaseUsername:   updatedParent.UserName.String,
				requestLoginEmail:  updatedParent.ID.String + constant.LoginEmailPostfix,
				databaseLoginEmail: updatedParent.LoginEmail.String,
			},
		); err != nil {
			return err
		}
		if req.SchoolId != updatedParent.SchoolID.Int {
			return fmt.Errorf(`expected updated "school_id": %v but actual is %v`, req.SchoolId, updatedParent.SchoolID.Int)
		}
		if parentInReq.UserNameFields.FirstName != updatedParent.FirstName.String {
			return fmt.Errorf(`expected updated "FirstName": %v but actual is %v`, parentInReq.UserNameFields.FirstName, updatedParent.FirstName.String)
		}
		if parentInReq.UserNameFields.LastName != updatedParent.LastName.String {
			return fmt.Errorf(`expected updated "LastName": %v but actual is %v`, parentInReq.UserNameFields.LastName, updatedParent.LastName.String)
		}
		if parentInReq.UserNameFields.LastNamePhonetic != updatedParent.LastNamePhonetic.String {
			return fmt.Errorf(`expected updated "LastNamePhonetic": %v but actual is %v`, parentInReq.UserNameFields.LastNamePhonetic, updatedParent.LastNamePhonetic.String)
		}
		if parentInReq.UserNameFields.FirstNamePhonetic != updatedParent.FirstNamePhonetic.String {
			return fmt.Errorf(`expected updated "FirstNamePhonetic": %v but actual is %v`, parentInReq.UserNameFields.FirstNamePhonetic, updatedParent.FirstNamePhonetic.String)
		}
		if helper.CombineFirstNameAndLastNameToFullName(parentInReq.UserNameFields.FirstName, parentInReq.UserNameFields.LastName) != updatedParent.FullName.String {
			return fmt.Errorf(`expected updated "FullName": %v but actual is %v`, helper.CombineFirstNameAndLastNameToFullName(parentInReq.UserNameFields.FirstName, parentInReq.UserNameFields.LastName), updatedParent.FullName.String)
		}
		if helper.CombineFirstNameAndLastNameToFullName(parentInReq.UserNameFields.FirstNamePhonetic, parentInReq.UserNameFields.LastNamePhonetic) != updatedParent.FullNamePhonetic.String {
			return fmt.Errorf(`expected updated "FullNamePhonetic": %v but actual is %v`, helper.CombineFirstNameAndLastNameToFullName(parentInReq.UserNameFields.FirstNamePhonetic, parentInReq.UserNameFields.LastNamePhonetic), updatedParent.FullNamePhonetic.String)
		}
		if parentInReq.ExternalUserId != updatedParent.ExternalUserID.String {
			return fmt.Errorf(`expected updated "external_user_id": %v but actual is %v`, parentInReq.ExternalUserId, updatedParent.ExternalUserID.String)
		}
		if err := s.validateUserTags(ctx, updatedParent.ID.String, parentInReq.TagIds); err != nil {
			return errors.WithMessage(err, "validateUserTags")
		}
		// Parent should login successfully with new email
		email := parentInReq.Email
		if isEnableUsername {
			email = updatedParent.ID.String + constant.LoginEmailPostfix
		}
		if err := s.loginIdentityPlatform(ctx, auth.LocalTenants[constants.ManabieSchool], email, stepState.ParentPassword); err != nil {
			return errors.WithMessage(err, "loginIdentityPlatform")
		}
	}
	// Compare locations between student and parents
	var studentLocationIDs, parentLocationIDs []string
	stmtStudentLocation := `
		SELECT location_id
		FROM user_access_paths
		WHERE user_id = $1 and deleted_at is null
	`
	stmtParentLocation := `
		SELECT location_id
		FROM user_access_paths
		WHERE user_id = ANY($1) and deleted_at is null
	`

	rows, err := s.BobDBTrace.Query(ctx, stmtStudentLocation, res.StudentId)
	if err != nil {
		return errors.Wrap(err, "query user_access_path get location ID from studentID")
	}
	defer rows.Close()

	for rows.Next() {
		var studentLocationID string

		err := rows.Scan(&studentLocationID)
		if err != nil {
			return errors.WithMessage(err, "rows.Scan get student location IDs")
		}
		studentLocationIDs = append(studentLocationIDs, studentLocationID)
	}

	rows, err = s.BobDBTrace.Query(ctx, stmtParentLocation, parentIdsParam)
	if err != nil {
		return errors.Wrap(err, "query user_access_path get location ID from parentID")
	}
	defer rows.Close()

	for rows.Next() {
		var parentLocationID string

		err := rows.Scan(&parentLocationID)
		if err != nil {
			return errors.WithMessage(err, "rows.Scan get parent location IDs")
		}
		parentLocationIDs = append(parentLocationIDs, parentLocationID)
	}

	// Check the length of locations per parent to compare whether each parent has equal location data with the student.
	lengthLocationsPerParent := len(parentLocationIDs) / len(parentIDs)

	if len(studentLocationIDs) != lengthLocationsPerParent {
		return fmt.Errorf(`expected parent location : %v but actual is %v`, lengthLocationsPerParent, len(studentLocationIDs))
	}

	return nil
}

func (s *suite) parentDataToUpdateHasEmptyOrInvalid(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.UpdateParentsAndFamilyRelationshipRequest)
	switch field {
	case "student_id not exist":
		req.StudentId = newID()
	case "email already exist":
		err := s.addParentDataToCreateParentReq(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if ctx, err := s.createNewParents(ctx, schoolAdminType); err != nil {
			return ctx, err
		}

		req := stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)
		oldParentProfiles := req.ParentProfiles

		err = s.addParentDataToUpdateParentReq(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		updateParentProfiles := stepState.Request.(*pb.UpdateParentsAndFamilyRelationshipRequest).ParentProfiles
		for _, updateProfile := range updateParentProfiles {
			updateProfile.Email = oldParentProfiles[0].Email
		}
	case "external_user_id already exist":
		err := s.addParentDataToCreateParentReq(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		req := stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)
		req.ParentProfiles[0].ExternalUserId = fmt.Sprintf("external-user-id-%s", req.ParentProfiles[0].Name)

		if ctx, err := s.createNewParents(ctx, schoolAdminType); err != nil {
			return ctx, err
		}

		oldParentProfiles := req.ParentProfiles
		err = s.addParentDataToUpdateParentReq(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		updateParentProfiles := stepState.Request.(*pb.UpdateParentsAndFamilyRelationshipRequest).ParentProfiles
		for _, updateProfile := range updateParentProfiles {
			updateProfile.ExternalUserId = oldParentProfiles[0].ExternalUserId
		}
	case "external_user_id re-update":
		err := s.addParentDataToCreateParentReq(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		req := stepState.Request.(*pb.CreateParentsAndAssignToStudentRequest)
		req.ParentProfiles[0].ExternalUserId = fmt.Sprintf("external-user-id-%s", req.ParentProfiles[0].Name)

		if ctx, err := s.createNewParents(ctx, schoolAdminType); err != nil {
			return ctx, err
		}

		res := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse)

		updateParentProfile := updateParentReq(req.StudentId, res.GetParentProfiles()[0].Parent.UserProfile.UserId)

		for _, updateProfile := range updateParentProfile.ParentProfiles {
			updateProfile.ExternalUserId = fmt.Sprintf("external-user-id-%s", newID())
		}
		stepState.Request = updateParentProfile
	case "empty username":
		for _, profile := range req.ParentProfiles {
			profile.Username = ""
		}
	case "username has spaces":
		for _, profile := range req.ParentProfiles {
			profile.Username = "  "
		}
	case "username has special characters":
		for _, profile := range req.ParentProfiles {
			profile.Username = "-.-"
		}
	case "existing username":
		// because getUsernameByUsingOther will override the request
		// so we need to save the request first
		request := s.Request

		for _, profile := range req.ParentProfiles {
			existingUsername, err := s.getUsernameByUsingOther(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "getUsernameByUsingOther")
			}
			profile.Username = existingUsername
		}

		s.Request = request
	case "existing username and upper case":
		// because getUsernameByUsingOther will override the request
		// so we need to save the request first
		request := s.Request

		for _, profile := range req.ParentProfiles {
			existingUsername, err := s.getUsernameByUsingOther(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "getUsernameByUsingOther")
			}
			profile.Username = strings.ToUpper(existingUsername)
		}

		s.Request = request
	}

	for _, parentProfile := range req.ParentProfiles {
		switch field {
		case "id":
			parentProfile.Id = ""
		case "email":
			parentProfile.Email = ""
		case "last name":
			parentProfile.UserNameFields.LastName = ""
		case "first name":
			parentProfile.UserNameFields.FirstName = ""
		case "parent_id not exist":
			parentProfile.Id = newID()
		case "parentPhoneNumber invalid":
			parentProfile.ParentPhoneNumbers = []*pb.ParentPhoneNumber{
				{PhoneNumber: "123", PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER},
			}
		case "relationship":
			max := int32(math.MinInt32)
			for _, value := range pb.FamilyRelationship_value {
				if value > max {
					max = value
				}
			}
			parentProfile.Relationship = pb.FamilyRelationship(max + 1)
		case "tag not exist":
			parentProfile.TagIds = append(parentProfile.TagIds, idutil.ULIDNow())
		case "tag for only student":
			tagIDs, _, err := s.createTagsType(ctx, studentType)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			parentProfile.TagIds = tagIDs
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) cannotUpdateParents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return ctx, errors.New("expected response has err but actual is nil")
	}
	return ctx, nil
}

func updateParentReq(studentID string, parentID string) *pb.UpdateParentsAndFamilyRelationshipRequest {
	profiles := []*pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
		{
			Id:           parentID,
			Email:        fmt.Sprintf("%v@example.com", parentID),
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
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

	req := &pb.UpdateParentsAndFamilyRelationshipRequest{
		ParentProfiles: profiles,
		StudentId:      studentID,
	}

	return req
}
func (s *suite) editParentDataToBlank(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.UpdateParentsAndFamilyRelationshipRequest)

	for _, parentProfile := range req.ParentProfiles {
		switch field {
		case "last name phonetic":
			parentProfile.UserNameFields.LastNamePhonetic = ""
		case "first name phonetic":
			parentProfile.UserNameFields.FirstNamePhonetic = ""
		case "parent primary phone number":
			parentProfile.ParentPhoneNumbers = []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "0987654321",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER,
				},
			}
		case "parent secondary phone number":
			parentProfile.ParentPhoneNumbers = []*pb.ParentPhoneNumber{
				{
					PhoneNumber:     "0987654321",
					PhoneNumberType: pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER,
				},
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
