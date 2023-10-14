package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DomainParentService struct {
	DomainParent interface {
		GetUsersByExternalIDs(ctx context.Context, externalIDs []string) (entity.Users, error)
		GetUsersByEmails(ctx context.Context, emails []string) (entity.Users, error)
		GetStudentsAccessPaths(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error)
		GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error)
		UpsertMultipleWithChildren(ctx context.Context, option unleash.DomainParentFeatureOption, aggregateParents ...aggregate.DomainParentWithChildren) ([]aggregate.DomainParent, error)
		IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool
		IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error)
	}
	FeatureManager interface {
		FeatureUsernameToParentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainParentFeatureOption) unleash.DomainParentFeatureOption
	}
	UnleashClient unleashclient.ClientInstance
	Env           string
}

func (port *DomainParentService) UpsertParents(c *gin.Context) {
	zapLogger := ctxzap.Extract(c.Request.Context())
	organization, err := interceptors.OrganizationFromContext(c.Request.Context())
	if err != nil {
		zapLogger.Error(err.Error(), zap.Error(err))
		ResponseListErrors(c, []error{err})
		return
	}

	var req UpsertParentRequest
	if err := ParseJSONPayload(c.Request, &req); err != nil {
		zapLogger.Error(err.Error(), zap.Error(err))
		ResponseError(c, err)
		return
	}

	option := unleash.DomainParentFeatureOption{
		DomainUserFeatureOption: unleash.DomainUserFeatureOption{
			EnableIgnoreUpdateEmail: true,
		},
	}
	option = port.FeatureManager.FeatureUsernameToParentFeatureOption(c.Request.Context(), organization, option)

	parentsToUpsert, err := port.toDomainParentWithChildrenAggregate(c.Request.Context(), req.Parents, option.EnableUsername)
	if err != nil {
		zapLogger.Error(err.Error(), zap.Error(err))
		ResponseError(c, err)
		return
	}
	parent, err := port.DomainParent.UpsertMultipleWithChildren(c.Request.Context(), option, parentsToUpsert...)
	if err != nil {
		zapLogger.Error(err.Error(), zap.Error(err))
		ResponseError(c, err)
		return
	}

	data := make([]map[string]interface{}, 0)
	for _, parent := range parent {
		data = append(data, map[string]interface{}{
			"user_id":          parent.UserID().String(),
			"external_user_id": parent.ExternalUserID().String(),
		})
	}

	c.JSON(http.StatusOK, Response{
		Data:    data,
		Code:    20000,
		Message: "success",
	})
}

func (port *DomainParentService) toDomainParentWithChildrenAggregate(ctx context.Context, parentProfiles []ParentProfile, isUserNameEnabled bool) (aggregate.DomainParentWithChildrens, error) {
	userIDs, err := port.toUserIDs(ctx, parentProfiles)
	if err != nil {
		return nil, err
	}

	parentsToUpsert := aggregate.DomainParentWithChildrens{}
	for i, parent := range parentProfiles {
		if err = validateParentChildren(parent, i); err != nil {
			return nil, err
		}

		if userIDs[i] != "" {
			parent.UserIDAttr = field.NewString(userIDs[i])
		}

		if !isUserNameEnabled {
			parent.UserNameAttr = parent.EmailAttr
		}
		parent.LoginEmailAttr = parent.EmailAttr

		// trim external_user_id
		parent.ExternalUserIDAttr = field.NewString(strings.TrimSpace(parent.ExternalUserIDAttr.String()))

		parent.FullNameAttr = field.NewString(utils.CombineFirstNameAndLastNameToFullName(parent.FirstNameAttr.String(), parent.LastNameAttr.String()))
		parent.FullNamePhoneticAttr = field.NewString(utils.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(parent.FirstNamePhoneticAttr.String(), parent.LastNamePhoneticAttr.String()))
		if parent.FullNamePhoneticAttr.String() == "" && !field.IsPresent(parent.FirstNamePhoneticAttr) && !field.IsPresent(parent.LastNamePhoneticAttr) {
			parent.FullNamePhoneticAttr = field.NewNullString()
		}
		studentEmails := make([]string, 0, len(parent.ChildrenAttr))
		for _, child := range parent.ChildrenAttr {
			studentEmails = append(studentEmails, child.StudentEmailAttr.String())
		}

		students, err := port.studentEmailsToUser(ctx, studentEmails, i)
		if err != nil {
			return nil, err
		}

		studentIDs := make([]string, 0, len(students))
		for j, child := range parent.ChildrenAttr {
			for _, student := range students {
				studentIDs = append(studentIDs, student.UserID().String())
				if child.StudentEmailAttr.String() == student.Email().String() {
					parent.ChildrenAttr[j].StudentIDAttr = field.NewString(student.UserID().String())
				}
			}
		}

		userAccessPaths, err := port.toDomainUserAccessPaths(ctx, studentIDs)
		if err != nil {
			return nil, err
		}

		taggedUsers, err := port.toDomainTaggedUsers(ctx, parent.ParentTagsAttr)
		if err != nil {
			return nil, err
		}

		children := entity.DomainStudentParentRelationships{}
		for _, child := range parent.ChildrenAttr {
			children = append(children, child)
		}

		userPhoneNumbers := entity.DomainUserPhoneNumbers{}
		if field.IsPresent(parent.SecondaryPhoneNumberAttr) {
			userPhoneNumbers = append(userPhoneNumbers, toDomainUserPhoneNumbers(parent.SecondaryPhoneNumberAttr.String(), entity.UserPhoneNumberTypeParentSecondaryPhoneNumber))
		}

		if field.IsPresent(parent.PrimaryPhoneNumberAttr) {
			userPhoneNumbers = append(userPhoneNumbers, toDomainUserPhoneNumbers(parent.PrimaryPhoneNumberAttr.String(), entity.UserPhoneNumberTypeParentPrimaryPhoneNumber))
		}

		parentsToUpsert = append(parentsToUpsert, aggregate.DomainParentWithChildren{
			DomainParent: aggregate.DomainParent{
				DomainParent:     parent,
				UserAccessPaths:  userAccessPaths,
				TaggedUsers:      taggedUsers,
				UserPhoneNumbers: userPhoneNumbers,
				IndexAttr:        i,
			},
			Children: children,
		})
	}

	return parentsToUpsert, nil
}

func (port *DomainParentService) toDomainTaggedUsers(ctx context.Context, partnerInternalIDs []field.String) (entity.DomainTaggedUsers, error) {
	domainTaggedUsers := entity.DomainTaggedUsers{}
	externalIDs := make([]string, 0, len(partnerInternalIDs))
	for _, partnerInternalID := range partnerInternalIDs {
		externalIDs = append(externalIDs, partnerInternalID.String())
	}

	tags, err := port.DomainParent.GetTagsByExternalIDs(ctx, externalIDs)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.DomainStudent.GetTagsByExternalIDs"),
		}
	}

	for _, partnerInternalID := range partnerInternalIDs {
		taggedUser := entity.TaggedUserWillBeDelegated{
			HasTagID: entity.EmptyDomainTag{},
		}
		for _, tag := range tags {
			if partnerInternalID.String() == tag.PartnerInternalID().String() {
				taggedUser.HasTagID = tag
				break
			}
		}
		domainTaggedUsers = append(domainTaggedUsers, taggedUser)
	}

	return domainTaggedUsers, nil
}

func (port *DomainParentService) toDomainUserAccessPaths(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error) {
	userAccessPaths := entity.DomainUserAccessPaths{}
	locations, err := port.DomainParent.GetStudentsAccessPaths(ctx, studentIDs)
	if err != nil {
		return nil, err
	}

	for _, location := range locations {
		userAccessPath := entity.UserAccessPathWillBeDelegated{
			HasLocationID:     location,
			HasOrganizationID: location,
		}

		userAccessPaths = append(userAccessPaths, userAccessPath)
	}

	return userAccessPaths, nil
}

func (port *DomainParentService) studentEmailsToUser(ctx context.Context, studentEmails []string, idx int) (entity.Users, error) {
	students, err := port.DomainParent.GetUsersByEmails(ctx, studentEmails)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "s.DomainParent.GetUsersByEmails"),
		}
	}

	if len(students) != len(studentEmails) {
		for i, studentEmail := range studentEmails {
			found := false
			for _, student := range students {
				if studentEmail == student.Email().String() {
					found = true
					break
				}
			}
			if !found {
				return nil, errcode.Error{
					FieldName: fmt.Sprintf("parents[%d].children[%d].student_email", idx, i),
					Code:      errcode.InvalidData,
					Err:       errors.New("student email not found"),
				}
			}
		}
	}

	return students, nil
}

func (port *DomainParentService) toUserIDs(ctx context.Context, parentProfiles []ParentProfile) ([]string, error) {
	externalUserIDs := make([]string, len(parentProfiles))
	for i, s := range parentProfiles {
		trimmedExternalUserID := strings.TrimSpace(s.ExternalUserIDAttr.String())
		if trimmedExternalUserID == "" {
			return nil, errcode.Error{
				FieldName: fmt.Sprintf("parents[%d].external_user_id", i),
				Code:      errcode.MissingMandatory,
			}
		}
		externalUserIDs[i] = strings.TrimSpace(trimmedExternalUserID)
	}

	existingUsers, err := port.DomainParent.GetUsersByExternalIDs(ctx, externalUserIDs)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "s.DomainParent.GetUsersByExternalIDs"),
		}
	}

	userIDs := make([]string, 0)
	for _, externalUserID := range externalUserIDs {
		userID := ""
		for _, user := range existingUsers {
			if externalUserID == user.ExternalUserID().String() {
				userID = user.UserID().String()
			}
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func validateParentChildren(parent ParentProfile, idx int) error {
	if len(parent.ChildrenAttr) == 0 {
		return errcode.Error{
			Code:      errcode.MissingMandatory,
			FieldName: "children",
		}
	}

	for i, child := range parent.ChildrenAttr {
		if child.StudentEmailAttr.String() == "" {
			return errcode.Error{
				Code:      errcode.MissingMandatory,
				FieldName: fmt.Sprintf("parents[%d].children[%d].student_email", idx, i),
			}
		}
		if child.RelationshipAttr.Int32() == 0 {
			return errcode.Error{
				Code:      errcode.MissingMandatory,
				FieldName: fmt.Sprintf("parents[%d].children[%d].relationship", idx, i),
			}
		} else if _, ok := mapStudentParentRelationship[child.RelationshipAttr.Int32()]; !ok {
			return errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("parents[%d].children[%d].relationship", idx, i),
			}
		}
	}

	return nil
}
