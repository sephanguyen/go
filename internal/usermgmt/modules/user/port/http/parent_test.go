package http

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"

	"github.com/stretchr/testify/assert"
)

type mockDomainParent struct {
	getUsersByExternalIDsFn      func(ctx context.Context, externalUserIDs []string) (entity.Users, error)
	getUsersByEmailsFn           func(ctx context.Context, emails []string) (entity.Users, error)
	getStudentsAccessPathFn      func(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error)
	getTagsByExternalIDsFn       func(ctx context.Context, externalUserIDs []string) (entity.DomainTags, error)
	upsertMultipleWithChildrenFn func(ctx context.Context, option unleash.DomainParentFeatureOption, aggreegateParents ...aggregate.DomainParentWithChildren) ([]aggregate.DomainParent, error)
}

func (m *mockDomainParent) GetUsersByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
	return m.getUsersByExternalIDsFn(ctx, externalUserIDs)
}

func (m *mockDomainParent) GetUsersByEmails(ctx context.Context, emails []string) (entity.Users, error) {
	return m.getUsersByEmailsFn(ctx, emails)
}

func (m *mockDomainParent) GetStudentsAccessPaths(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error) {
	return m.getStudentsAccessPathFn(ctx, studentIDs)
}

func (m *mockDomainParent) GetTagsByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.DomainTags, error) {
	return m.getTagsByExternalIDsFn(ctx, externalUserIDs)
}

func (m *mockDomainParent) UpsertMultipleWithChildren(ctx context.Context, option unleash.DomainParentFeatureOption, aggregateParents ...aggregate.DomainParentWithChildren) ([]aggregate.DomainParent, error) {
	return m.upsertMultipleWithChildrenFn(ctx, option, aggregateParents...)
}

func (m *mockDomainParent) IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool {
	return true
}
func (m *mockDomainParent) IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error) {
	return true, nil
}

func TestDomainParent_toDomainParentWithChildrenAggregate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	parentProfiles := []ParentProfile{
		{
			ExternalUserIDAttr:       field.NewString("external-user-id-1"),
			FirstNameAttr:            field.NewString("first name"),
			LastNameAttr:             field.NewString("last name"),
			EmailAttr:                field.NewString("email@email.com"),
			UserNameAttr:             field.NewString("username"),
			PrimaryPhoneNumberAttr:   field.NewString("0123456789"),
			SecondaryPhoneNumberAttr: field.NewString("0123456789"),
			RemarksAttr:              field.NewString("remarks"),
			ParentTagsAttr: []field.String{
				field.NewString("tag-1"),
				field.NewString("tag-2"),
			},
			GenderAttr: field.NewInt32(1),
			ChildrenAttr: []ParentChildrenPayload{
				{
					StudentEmailAttr: field.NewString("student-email-1"),
					RelationshipAttr: field.NewInt32(1),
				},
			},
		},
	}

	t.Run("happy case", func(t *testing.T) {
		m := mockDomainParent{
			getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
				return entity.Users{mock_usermgmt.User{
					RandomUser: mock_usermgmt.RandomUser{
						UserID:         field.NewString("parent-id"),
						ExternalUserID: field.NewString("external-user-id-1"),
					},
				},
				}, nil
			},
			getUsersByEmailsFn: func(ctx context.Context, emails []string) (entity.Users, error) {
				return entity.Users{entity.EmptyUser{}}, nil
			},
			getStudentsAccessPathFn: func(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error) {
				return entity.DomainUserAccessPaths{entity.DefaultUserAccessPath{}}, nil
			},
			getTagsByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.DomainTags, error) {
				return entity.DomainTags{entity.EmptyDomainTag{}}, nil
			},
		}

		port := DomainParentService{DomainParent: &m}

		domainParentWithChildrenAggregate, err := port.toDomainParentWithChildrenAggregate(ctx, parentProfiles, true)
		for idx, agg := range domainParentWithChildrenAggregate {
			req := parentProfiles[idx]
			assert.Equal(t, req.ExternalUserIDAttr.String(), agg.ExternalUserID().String())
			assert.Equal(t, req.FirstNameAttr.String(), agg.FirstName().String())
			assert.Equal(t, req.LastNameAttr.String(), agg.LastName().String())
			assert.Equal(t, req.EmailAttr.String(), agg.Email().String())
			// with rollback strategy, username should be the same as email
			assert.Equal(t, req.UserNameAttr.String(), agg.UserName().String())
			assert.Equal(t, req.FirstNamePhoneticAttr.String(), agg.FirstNamePhonetic().String())
			assert.Equal(t, req.LastNamePhoneticAttr.String(), agg.LastNamePhonetic().String())
			assert.Equal(t, req.RemarksAttr.String(), agg.Remarks().String())
			assert.Equal(t, agg.LoginEmail().String(), req.EmailAttr.String())
		}
		assert.Equal(t, nil, err)
	})

	t.Run("happy case with username is disabled", func(t *testing.T) {
		m := mockDomainParent{
			getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
				return entity.Users{entity.EmptyUser{}}, nil
			},
			getUsersByEmailsFn: func(ctx context.Context, emails []string) (entity.Users, error) {
				return entity.Users{entity.EmptyUser{}}, nil
			},
			getStudentsAccessPathFn: func(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error) {
				return entity.DomainUserAccessPaths{entity.DefaultUserAccessPath{}}, nil
			},
			getTagsByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.DomainTags, error) {
				return entity.DomainTags{entity.EmptyDomainTag{}}, nil
			},
		}

		port := DomainParentService{DomainParent: &m}

		domainParentWithChildrenAggregate, err := port.toDomainParentWithChildrenAggregate(ctx, parentProfiles, false)
		for idx, agg := range domainParentWithChildrenAggregate {
			req := parentProfiles[idx]
			assert.Equal(t, req.ExternalUserIDAttr.String(), agg.ExternalUserID().String())
			assert.Equal(t, req.FirstNameAttr.String(), agg.FirstName().String())
			assert.Equal(t, req.LastNameAttr.String(), agg.LastName().String())
			assert.Equal(t, req.EmailAttr.String(), agg.Email().String())
			assert.Equal(t, agg.UserName().String(), agg.Email().String())
			assert.Equal(t, req.FirstNamePhoneticAttr.String(), agg.FirstNamePhonetic().String())
			assert.Equal(t, req.LastNamePhoneticAttr.String(), agg.LastNamePhonetic().String())
			assert.Equal(t, req.RemarksAttr.String(), agg.Remarks().String())
			assert.Equal(t, agg.LoginEmail().String(), agg.Email().String())
		}
		assert.Equal(t, nil, err)
	})

	t.Run("external_user_id is null", func(t *testing.T) {
		m := mockDomainParent{
			getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
				return entity.Users{entity.EmptyUser{}}, nil
			},
			getUsersByEmailsFn: func(ctx context.Context, emails []string) (entity.Users, error) {
				return entity.Users{entity.EmptyUser{}}, nil
			},
			getStudentsAccessPathFn: func(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error) {
				return entity.DomainUserAccessPaths{entity.DefaultUserAccessPath{}}, nil
			},
			getTagsByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.DomainTags, error) {
				return entity.DomainTags{entity.EmptyDomainTag{}}, nil
			},
		}

		port := DomainParentService{DomainParent: &m}

		expErr := errcode.Error{
			FieldName: fmt.Sprintf("parents[%d].external_user_id", 0),
			Code:      errcode.MissingMandatory,
		}

		parentProfiles[0].ExternalUserIDAttr = field.NewNullString()
		_, err := port.toDomainParentWithChildrenAggregate(ctx, parentProfiles, false)
		assert.Equal(t, expErr, err)
	})

	t.Run("external_user_id is empty string", func(t *testing.T) {
		m := mockDomainParent{
			getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
				return entity.Users{entity.EmptyUser{}}, nil
			},
			getUsersByEmailsFn: func(ctx context.Context, emails []string) (entity.Users, error) {
				return entity.Users{entity.EmptyUser{}}, nil
			},
			getStudentsAccessPathFn: func(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error) {
				return entity.DomainUserAccessPaths{entity.DefaultUserAccessPath{}}, nil
			},
			getTagsByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.DomainTags, error) {
				return entity.DomainTags{entity.EmptyDomainTag{}}, nil
			},
		}

		port := DomainParentService{DomainParent: &m}

		expErr := errcode.Error{
			FieldName: fmt.Sprintf("parents[%d].external_user_id", 0),
			Code:      errcode.MissingMandatory,
		}

		parentProfiles[0].ExternalUserIDAttr = field.NewString("")
		_, err := port.toDomainParentWithChildrenAggregate(ctx, parentProfiles, false)
		assert.Equal(t, expErr, err)
	})
}
