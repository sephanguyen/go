package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func Test_getIndexFromDomainError(t *testing.T) {
	tests := []struct {
		name string
		args string
		want int
	}{
		{
			name: "happy case: index is 0",
			args: "student[0].phone_number.contact_preference invalid",
			want: 0,
		},
		{
			name: "happy case: index is 2",
			args: "student[2].phone_number.contact_preference invalid",
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := GetIndexFromMessageError(tt.args)
			assert.Equal(t, tt.want, index)

		})
	}
}

func Test_getFieldFromMessageError(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "happy case: field is contact_preference",
			args: "student[0].phone_number.contact_preference invalid",
			want: "contact_preference",
		},
		{
			name: "happy case: field is email",
			args: "student[2].email is existing",
			want: "email",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := GetFieldFromMessageError(tt.args)
			assert.Equal(t, tt.want, index)

		})
	}
}

func Test_ToPbErrorMessageImport(t *testing.T) {
	tests := []struct {
		name string
		args errcode.DomainError
		want *upb.ErrorMessage
	}{
		{
			name: "happy case: field is FieldEnrollmentStatusHistoryStartDate",
			args: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				Index:      2,
				FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
			},
			want: &upb.ErrorMessage{
				FieldName: "start_date",
				Code:      errcode.InvalidData,
				Index:     2 + 2,
			},
		},
		{
			name: "happy case: field is email",
			args: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				Index:      1,
				FieldName:  string(entity.UserFieldEmail),
			},
			want: &upb.ErrorMessage{
				FieldName: "email",
				Code:      errcode.InvalidData,
				Index:     1 + 2,
			},
		},
		{
			name: "happy case: field is StudentSchoolHistoryEndDateField",
			args: entity.DuplicatedFieldError{
				EntityName:      entity.StudentEntity,
				Index:           1,
				DuplicatedField: string(entity.StudentSchoolHistoryEndDateField),
			},
			want: &upb.ErrorMessage{
				FieldName: "end_date",
				Code:      errcode.DuplicatedData,
				Index:     1 + 2,
			},
		},
		{
			name: "happy case: field is StudentSchoolHistoryEndDateField",
			args: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				Index:      1,
				FieldName:  string(entity.StudentFieldContactPreference),
			},
			want: &upb.ErrorMessage{
				FieldName: "contact_preference",
				Code:      errcode.InvalidData,
				Index:     1 + 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ToPbErrorMessageImport(tt.args)
			assert.Equal(t, tt.want, err)
		})
	}
}

func Test_ToPbErrorMessageBackOffice(t *testing.T) {
	tests := []struct {
		name string
		args errcode.DomainError
		want *upb.ErrorMessage
	}{
		{
			name: "happy case: field is FieldEnrollmentStatusHistoryStartDate",
			args: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					Index:      2,
					FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
			want: &upb.ErrorMessage{
				FieldName: "start_date",
				Code:      errcode.InvalidData,
				Index:     2,
				Error:     "student[2].enrollment_status_histories[0].start_date invalid",
			},
		},
		{
			name: "happy case: field is email",
			args: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				Index:      1,
				FieldName:  string(entity.UserFieldEmail),
			},
			want: &upb.ErrorMessage{
				FieldName: "email",
				Code:      errcode.InvalidData,
				Index:     1,
				Error:     "student[1].email invalid",
			},
		},
		{
			name: "happy case: field is StudentSchoolHistoryEndDateField",
			args: entity.DuplicatedFieldError{
				EntityName:      entity.StudentEntity,
				Index:           1,
				DuplicatedField: string(entity.StudentSchoolHistoryEndDateField),
			},
			want: &upb.ErrorMessage{
				FieldName: "end_date",
				Code:      errcode.DuplicatedData,
				Index:     1,
				Error:     "student[1].end_date duplicated",
			},
		},
		{
			name: "happy case: field is StudentSchoolHistoryEndDateField",
			args: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				Index:      1,
				FieldName:  string(entity.StudentFieldContactPreference),
			},
			want: &upb.ErrorMessage{
				FieldName: "contact_preference",
				Code:      errcode.InvalidData,
				Index:     1,
				Error:     "student[1].phone_number.contact_preference invalid",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			err := ToPbErrorMessageBackOffice(tt.args)
			assert.Equal(t, tt.want, err)
		})
	}
}
