package usermgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	http_port "github.com/manabie-com/backend/internal/usermgmt/modules/user/port/http"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	pkg_unleash "github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
)

func (s *suite) modifyValidParent(ctx context.Context, condition string, request http_port.UpsertParentRequest, student *entity.LegacyStudent, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch condition {
	case "mandatory fields":
		request.Parents[0].FirstNameAttr = field.NewString(fmt.Sprintf("parent first name %s", id))
	case "with tags":
		tagsIDs, _, err := s.createAmountTags(ctx, 1, entity.ParentTags[0], fmt.Sprint(constants.ManabieSchool))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		tags, err := (&repository.DomainTagRepo{}).GetByIDs(ctx, s.BobPostgresDB, tagsIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.Parents[0].ParentTagsAttr = []field.String{
			tags[0].PartnerInternalID(),
		}
	case "with user phone numbers":
		request.Parents[0].PrimaryPhoneNumberAttr = field.NewString("0987654321")
		request.Parents[0].SecondaryPhoneNumberAttr = field.NewString("0123456789")
	case "existed user in database":
		err := s.addParentDataToCreateParentReq(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if _, err := s.createNewParents(ctx, schoolAdminType); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		parentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).ParentProfiles[0].Parent.UserProfile.UserId
		externalUserIDParent := fmt.Sprintf("external_user_id_%s", parentID)

		_, err = s.BobPostgresDB.Exec(ctx, fmt.Sprintf("UPDATE users SET user_external_id = '%s' WHERE user_id = '%s';", externalUserIDParent, parentID))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.Parents[0].ExternalUserIDAttr = field.NewString(externalUserIDParent)
	case "create multiple parents":
		id1 := newID()
		request.Parents = append(request.Parents, http_port.ParentProfile{
			ExternalUserIDAttr: field.NewString(fmt.Sprintf("external-user-id-parent %s", id1)),
			EmailAttr:          field.NewString(fmt.Sprintf("email-%s@example.com", id1)),
			FirstNameAttr:      field.NewString(fmt.Sprintf("parent first name %s", id1)),
			LastNameAttr:       field.NewString(fmt.Sprintf("parent last name %s", id1)),
			ChildrenAttr: []http_port.ParentChildrenPayload{
				{
					StudentEmailAttr: field.NewString(student.Email.String),
					RelationshipAttr: field.NewInt32(1),
				},
			},
		})
	case "both create and update":
		err := s.addParentDataToCreateParentReq(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if _, err := s.createNewParents(ctx, schoolAdminType); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		parentID := stepState.Response.(*pb.CreateParentsAndAssignToStudentResponse).ParentProfiles[0].Parent.UserProfile.UserId
		externalUserIDParent := fmt.Sprintf("external_user_id_%s", parentID)

		_, err = s.BobPostgresDB.Exec(ctx, fmt.Sprintf("UPDATE users SET user_external_id = '%s' WHERE user_id = '%s';", externalUserIDParent, parentID))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.Parents[0].ExternalUserIDAttr = field.NewString(externalUserIDParent)

		id1 := newID()
		request.Parents = append(request.Parents, http_port.ParentProfile{
			ExternalUserIDAttr: field.NewString(fmt.Sprintf("external-user-id-parent %s", id1)),
			EmailAttr:          field.NewString(fmt.Sprintf("email-%s@example.com", id1)),
			FirstNameAttr:      field.NewString(fmt.Sprintf("parent first name %s", id1)),
			LastNameAttr:       field.NewString(fmt.Sprintf("parent last name %s", id1)),
			ChildrenAttr: []http_port.ParentChildrenPayload{
				{
					StudentEmailAttr: field.NewString(student.Email.String),
					RelationshipAttr: field.NewInt32(1),
				},
			},
		})
	case "external user id with spaces":
		request.Parents[0].ExternalUserIDAttr = field.NewString(request.Parents[0].ExternalUserIDAttr.String() + "       ")
	case "available username":
		request.Parents[0].UserNameAttr = field.NewString("username" + idutil.ULIDNow())
	case "available username with email format":
		request.Parents[0].UserNameAttr = field.NewString("username." + idutil.ULIDNow() + "@manabie.com")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) modifyInvalidParent(ctx context.Context, condition string, request http_port.UpsertParentRequest, student *entity.LegacyStudent, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch condition {
	case "invalid tags":
		tagsIDs, _, err := s.createAmountTags(ctx, 1, entity.StudentTags[0], fmt.Sprint(constants.ManabieSchool))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		tags, err := (&repository.DomainTagRepo{}).GetByIDs(ctx, s.BobPostgresDB, tagsIDs)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.Parents[0].ParentTagsAttr = []field.String{
			tags[0].PartnerInternalID(),
		}
	case "invalid email":
		request.Parents[0].EmailAttr = field.NewString(fmt.Sprintf("invalid-email-%s", id))
	case "missing email":
		request.Parents[0].EmailAttr = field.NewNullString()
	case "missing first_name":
		request.Parents[0].FirstNameAttr = field.NewNullString()
	case "missing last_name":
		request.Parents[0].LastNameAttr = field.NewNullString()
	case "missing external_user_id":
		request.Parents[0].ExternalUserIDAttr = field.NewNullString()
	case "missing children":
		request.Parents[0].ChildrenAttr = nil
	case "children empty":
		request.Parents[0].ChildrenAttr = []http_port.ParentChildrenPayload{}
	case "invalid children email":
		request.Parents[0].ChildrenAttr = []http_port.ParentChildrenPayload{
			{
				StudentEmailAttr: field.NewString(student.Email.String),
				RelationshipAttr: field.NewInt32(1),
			},
			{
				StudentEmailAttr: field.NewString(fmt.Sprintf("%s_not_found@example.com", id)),
				RelationshipAttr: field.NewInt32(1),
			},
		}
	case "children email is null":
		request.Parents[0].ChildrenAttr = []http_port.ParentChildrenPayload{
			{
				StudentEmailAttr: field.NewNullString(),
				RelationshipAttr: field.NewInt32(1),
			},
		}
	case "children email is empty":
		request.Parents[0].ChildrenAttr = []http_port.ParentChildrenPayload{
			{
				StudentEmailAttr: field.NewString(""),
				RelationshipAttr: field.NewInt32(1),
			},
		}
	case "children relationship is missing":
		request.Parents[0].ChildrenAttr = []http_port.ParentChildrenPayload{
			{
				StudentEmailAttr: field.NewString(student.Email.String),
				RelationshipAttr: field.NewNullInt32(),
			},
		}
	case "children relationship is invalid":
		request.Parents[0].ChildrenAttr = []http_port.ParentChildrenPayload{
			{
				StudentEmailAttr: field.NewString(student.Email.String),
				RelationshipAttr: field.NewInt32(10),
			},
		}
	case "username was used by other":
		// user was created in migration file 1-local-init-sql.sql
		request.Parents[0].UserNameAttr = field.NewString("username_existing_01@gmail.com")
	case "username was used by other with upper case":
		// user was created in migration file 1-local-init-sql.sql
		request.Parents[0].UserNameAttr = field.NewString("USERNAME_EXISTING_01@GMAIL.COM")
	case "empty username":
		request.Parents[0].UserNameAttr = field.NewString("  ")
	case "username has special characters":
		request.Parents[0].UserNameAttr = field.NewString("-_-")
	case "external_user_id was used by student":
		request.Parents[0].ExternalUserIDAttr = field.NewString("user_external_id_existing_02")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createParentByOpenAPI(ctx context.Context, check string, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	student, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	id := newID()
	request := http_port.UpsertParentRequest{
		Parents: []http_port.ParentProfile{
			{
				ExternalUserIDAttr: field.NewString(fmt.Sprintf("external-user-id-parent %s", id)),
				UserNameAttr:       field.NewString("username" + idutil.ULIDNow()),
				EmailAttr:          field.NewString(fmt.Sprintf("email-%s@example.com", id)),
				FirstNameAttr:      field.NewString(fmt.Sprintf("parent first name %s", id)),
				LastNameAttr:       field.NewString(fmt.Sprintf("parent last name %s", id)),
				ChildrenAttr: []http_port.ParentChildrenPayload{
					{
						StudentEmailAttr: field.NewString(student.Email.String),
						RelationshipAttr: field.NewInt32(1),
					},
				},
			},
		},
	}

	switch check {
	case "valid":
		ctx, err = s.modifyValidParent(ctx, condition, request, student, id)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "invalid":
		ctx, err = s.modifyInvalidParent(ctx, condition, request, student, id)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	s.Request = request

	ctx, err = s.externalServiceCallUpsertParentsAPI(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) externalServiceCallUpsertParentsAPI(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.setupAPIKey(ctx)
	if err != nil {
		return ctx, errors.Wrap(err, "s.setupAPIKey")
	}

	url := fmt.Sprintf(`http://%s%s`, s.Cfg.UserMgmtRestAddr, constant.DomainParentEndpoint)

	payloadByte, err := json.Marshal(stepState.Request)
	if err != nil {
		return ctx, errors.Wrap(err, "json.Marshal")
	}
	bodyBytes, err := s.makeHTTPRequest(http.MethodPut, url, payloadByte)
	if err != nil {
		return ctx, errors.Wrap(err, "s.makeHTTPRequest")
	}

	if bodyBytes == nil {
		return ctx, fmt.Errorf("body is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentsWereByOpenAPISuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := s.Request.(http_port.UpsertParentRequest)
	resp := s.Response.(http_port.ResponseErrors)

	if resp.Code != 20000 {
		return ctx, fmt.Errorf("message: %s, code: %d", resp.Message, resp.Code)
	}

	parentsResponse, ok := resp.Data.([]interface{})
	if !ok {
		return ctx, fmt.Errorf("resp.Data.([]map[string]interface{}) is invalid")
	}

	for i, reqParent := range req.Parents {
		parentResponse, ok := parentsResponse[i].(map[string]interface{})
		if !ok {
			return ctx, fmt.Errorf("parentsResponse[i].(map[string]interface{}) is invalid")
		}

		userID, ok := parentResponse["user_id"].(string)
		if !ok {
			return ctx, fmt.Errorf("cannot get user_id from parentResponse")
		}

		dbParents, err := (&repository.DomainUserRepo{}).GetByIDs(ctx, s.BobPostgresDB, []string{userID})
		if err != nil {
			return ctx, err
		}

		if len(dbParents) != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("parentsWereSuccessfullyUpserted: cannot find parent by id")
		}
		existingParent := dbParents[0]

		switch {
		case existingParent.Email() != reqParent.Email():
			return ctx, fmt.Errorf(`validateUpsertParents: expected upserted "email": %v but actual is %v`, reqParent.Email(), existingParent.Email())
		case existingParent.FirstName() != reqParent.FirstNameAttr:
			return ctx, fmt.Errorf(`validateUpsertParents: expected upserted "first_name": %v but actual is %v`, reqParent.FirstNameAttr, existingParent.FirstName())
		case existingParent.LastName() != reqParent.LastNameAttr:
			return ctx, fmt.Errorf(`validateUpsertParents: expected upserted "last_name": %v but actual is %v`, reqParent.LastNameAttr, existingParent.LastName())
		case existingParent.UserRole().String() != string(constant.UserRoleParent):
			return ctx, fmt.Errorf(`validateUpsertParents: expected upserted "user_role": %v but actual is %v`, string(constant.UserRoleParent), existingParent.UserRole())
		}
		isEnableUsername, err := isFeatureToggleEnabled(ctx, s.UnleashSuite, pkg_unleash.FeatureToggleUserNameStudentParent)
		if err != nil {
			return ctx, errors.Wrap(err, fmt.Sprintf("Get feature toggle error(%s)", pkg_unleash.FeatureToggleUserNameStudentParent))
		}
		if err := assertUpsertUsername(isEnableUsername,
			assertUsername{
				requestUsername:    reqParent.UserName().String(),
				requestEmail:       reqParent.Email().String(),
				databaseUsername:   existingParent.UserName().String(),
				requestLoginEmail:  userID + "@manabie.com",
				databaseLoginEmail: existingParent.LoginEmail().String(),
			},
		); err != nil {
			return ctx, err
		}

		if reqParent.ExternalUserIDAttr.String() != "" {
			trimmedExternalUserID := strings.TrimSpace(reqParent.ExternalUserIDAttr.String())

			if existingParent.ExternalUserID().String() != trimmedExternalUserID {
				return ctx, fmt.Errorf(`validateUpsertParents: expected upserted "external_user_id": %v but actual is %v`, trimmedExternalUserID, existingParent.ExternalUserID())
			}
		}

		if len(reqParent.ParentTagsAttr) > 0 {
			tagIDsInRequest := reqParent.ParentTagsAttr
			taggedUsers, err := (&repository.DomainTaggedUserRepo{}).GetByUserIDs(ctx, s.BobPostgresDB, []string{userID})
			if err != nil {
				return ctx, fmt.Errorf("validateUpsertParents, DomainTaggedUserRepo.GetByUserIDs %v", err)
			}
			if len(tagIDsInRequest) != len(taggedUsers) {
				return ctx, fmt.Errorf("validateUpsertParents len tags in request expect: %v, actual: %v", len(tagIDsInRequest), len(taggedUsers))
			}
		}

		userPhoneNumbers, err := (&repository.DomainUserPhoneNumberRepo{}).GetByUserIDs(ctx, s.BobPostgresDBTrace, []string{userID})
		if err != nil {
			return ctx, fmt.Errorf("validateUpsertParents, DomainUserPhoneNumberRepo.GetByUserIDs %v", err)
		}

		userPhoneNumbersReq := entity.DomainUserPhoneNumbers{}
		if reqParent.SecondaryPhoneNumberAttr.String() != "" {
			userPhoneNumbersReq = append(userPhoneNumbersReq, &repository.UserPhoneNumber{
				UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
					UserID:      field.NewString(userID),
					PhoneNumber: reqParent.SecondaryPhoneNumberAttr,
					Type:        field.NewString(entity.UserPhoneNumberTypeParentSecondaryPhoneNumber),
				},
			})
		}

		if reqParent.PrimaryPhoneNumberAttr.String() != "" {
			userPhoneNumbersReq = append(userPhoneNumbersReq, &repository.UserPhoneNumber{
				UserPhoneNumberAttribute: repository.UserPhoneNumberAttribute{
					UserID:      field.NewString(userID),
					PhoneNumber: reqParent.PrimaryPhoneNumberAttr,
					Type:        field.NewString(entity.UserPhoneNumberTypeParentPrimaryPhoneNumber),
				},
			})
		}

		if len(userPhoneNumbersReq) != len(userPhoneNumbers) {
			return ctx, fmt.Errorf("validateUpsertParents, expected %v actual %v", len(userPhoneNumbersReq), len(userPhoneNumbers))
		}

		for _, userPhoneNumber := range userPhoneNumbers {
			for _, userPhoneNumberReq := range userPhoneNumbersReq {
				if userPhoneNumberReq.PhoneNumber() == userPhoneNumber.PhoneNumber() {
					err := CompareUserPhoneNumber(userPhoneNumberReq, userPhoneNumber)
					if err != nil {
						return ctx, err
					}
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func CompareUserPhoneNumber(expected entity.DomainUserPhoneNumber, actual entity.DomainUserPhoneNumber) error {
	if expected.UserID().String() != actual.UserID().String() {
		return fmt.Errorf("UserID: exptected %s, actual %s", expected.UserID().String(), actual.UserID().String())
	}

	if expected.Type().String() != actual.Type().String() {
		return fmt.Errorf(" Type: exptected %s, actual %s", expected.Type().String(), actual.Type().String())
	}

	if expected.PhoneNumber().String() != actual.PhoneNumber().String() {
		return fmt.Errorf("PhoneNumber: exptected %s, actual %s", expected.PhoneNumber().String(), actual.PhoneNumber().String())
	}

	return nil
}

func (s *suite) parentsWereCreatedByOpenAPIUnsuccessfully(ctx context.Context, code string, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := s.Response.(http_port.ResponseErrors)
	req := s.Request.(http_port.UpsertParentRequest)

	if resp.Code == 20000 {
		return ctx, fmt.Errorf("expected there are errors but not, message: %s, code: %d", resp.Message, resp.Code)
	}

	if strconv.Itoa(resp.Code) != code {
		return ctx, fmt.Errorf("expected code is %s, actual is %d", code, resp.Code)
	}

	if !strings.Contains(resp.Message, field) {
		return ctx, fmt.Errorf("expected field: %s, but actual message is: %s", field, resp.Message)
	}

	emails := make([]string, 0, len(req.Parents))
	for _, parent := range req.Parents {
		emails = append(emails, parent.EmailAttr.String())
	}
	ctx, err := s.verifyUsersNotInBD(ctx, emails)
	if err != nil {
		return ctx, fmt.Errorf("verifyUsersNotInBD err:%v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
