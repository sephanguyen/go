package service

import (
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestValidateUserPhoneNumbers(t *testing.T) {
	type args struct {
		userPhoneNumbers entity.DomainUserPhoneNumbers
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "invalid phone number with invalid student phone number",
			args: args{
				userPhoneNumbers: []entity.DomainUserPhoneNumber{
					mock_usermgmt.NewUserPhoneNumber("0", entity.UserPhoneNumberTypeStudentPhoneNumber),
					mock_usermgmt.NewUserPhoneNumber("0", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
				},
			},
			wantErr: entity.InvalidFieldError{
				EntityName: entity.UserEntity,
				FieldName:  entity.StudentFieldStudentPhoneNumber,
				Index:      0,
			},
		},
		{
			name: "invalid phone number with invalid home phone number",
			args: args{
				userPhoneNumbers: []entity.DomainUserPhoneNumber{
					mock_usermgmt.NewUserPhoneNumber("0900000000", entity.UserPhoneNumberTypeStudentPhoneNumber),
					mock_usermgmt.NewUserPhoneNumber("0", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
				},
			},
			wantErr: entity.InvalidFieldError{
				EntityName: entity.UserEntity,
				FieldName:  entity.StudentFieldHomePhoneNumber,
				Index:      0,
			},
		},
		{
			name: "invalid phone number with invalid primary phone number",
			args: args{
				userPhoneNumbers: []entity.DomainUserPhoneNumber{
					mock_usermgmt.NewUserPhoneNumber("0", constant.ParentPrimaryPhoneNumber),
				},
			},
			wantErr: entity.InvalidFieldError{
				EntityName: entity.UserEntity,
				FieldName:  entity.StudentFieldPrimaryPhoneNumber,
				Index:      0,
			},
		},
		{
			name: "invalid phone number with invalid second phone number",
			args: args{
				userPhoneNumbers: []entity.DomainUserPhoneNumber{
					mock_usermgmt.NewUserPhoneNumber("0", constant.ParentSecondaryPhoneNumber),
				},
			},
			wantErr: entity.InvalidFieldError{
				EntityName: entity.UserEntity,
				FieldName:  entity.StudentFieldSecondaryPhoneNumber,
				Index:      0,
			},
		},
		{
			name: "invalid phone number with duplicated home phone number",
			args: args{
				userPhoneNumbers: []entity.DomainUserPhoneNumber{
					mock_usermgmt.NewUserPhoneNumber("0123123421412", entity.UserPhoneNumberTypeStudentPhoneNumber),
					mock_usermgmt.NewUserPhoneNumber("0123123421412", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
				},
			},
			wantErr: entity.DuplicatedFieldError{
				EntityName:      entity.UserEntity,
				DuplicatedField: entity.StudentFieldHomePhoneNumber,
				Index:           0,
			},
		},
		{
			name: "valid phone number with 2 empty phone number",
			args: args{
				userPhoneNumbers: []entity.DomainUserPhoneNumber{
					mock_usermgmt.NewUserPhoneNumber("", entity.UserPhoneNumberTypeStudentPhoneNumber),
					mock_usermgmt.NewUserPhoneNumber("", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
				},
			},
			wantErr: nil,
		},
		{
			name: "valid phone number with 2 diff phone number",
			args: args{
				userPhoneNumbers: []entity.DomainUserPhoneNumber{
					mock_usermgmt.NewUserPhoneNumber("0123123421412", entity.UserPhoneNumberTypeStudentPhoneNumber),
					mock_usermgmt.NewUserPhoneNumber("0123123421413", entity.UserPhoneNumberTypeStudentHomePhoneNumber),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			err := ValidateUserPhoneNumbers(tt.args.userPhoneNumbers, 0)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}
