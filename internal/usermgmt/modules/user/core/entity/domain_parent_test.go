package entity

import (
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type RandomParent struct {
	NullDomainParent
	UserNameAttr  field.String
	UserIDAttr    field.String
	EmailAttr     field.String
	FirstNameAttr field.String
	LastNameAttr  field.String
}

func (m *RandomParent) UserID() field.String {
	return m.UserIDAttr
}

func (m *RandomParent) Email() field.String {
	return m.EmailAttr
}

func (m *RandomParent) UserName() field.String {
	return m.UserNameAttr
}

func (m *RandomParent) FirstName() field.String {
	return m.FirstNameAttr
}

func (m *RandomParent) LastName() field.String {
	return m.LastNameAttr
}

func TestValidParent(t *testing.T) {
	type args struct {
		parent           DomainParent
		isEnableUsername bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "bad case: validate parent with username",
			args: args{
				parent: &RandomParent{
					EmailAttr:     field.NewString(strings.ToLower(fmt.Sprintf("email_%s@manabie.com", idutil.ULIDNow()))),
					FirstNameAttr: field.NewString("first_name"),
					LastNameAttr:  field.NewString("last_name"),
				},
				isEnableUsername: true,
			},
			wantErr: MissingMandatoryFieldError{
				FieldName:  string(UserFieldUserName),
				EntityName: UserEntity,
				Index:      -1,
			},
		},
		{
			name: "happy case: validate parent with username",
			args: args{
				parent: &RandomParent{
					EmailAttr:     field.NewString(strings.ToLower(fmt.Sprintf("email_%s@manabie.com", idutil.ULIDNow()))),
					UserNameAttr:  field.NewString("username"),
					FirstNameAttr: field.NewString("first_name"),
					LastNameAttr:  field.NewString("last_name"),
				},
				isEnableUsername: true,
			},
			wantErr: nil,
		},
		{
			name: "happy case: validate normal parent",
			args: args{
				parent: &RandomParent{
					EmailAttr:     field.NewString(strings.ToLower(fmt.Sprintf("email_%s@manabie.com", idutil.ULIDNow()))),
					UserNameAttr:  field.NewString("username"),
					FirstNameAttr: field.NewString("first_name"),
					LastNameAttr:  field.NewString("last_name"),
				},
				isEnableUsername: true,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidParent(tt.args.parent, tt.args.isEnableUsername); err != tt.wantErr {
				t.Errorf("ValidStudent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
