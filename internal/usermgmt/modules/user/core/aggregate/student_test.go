package aggregate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
)

func TestValidStudent(t *testing.T) {
	type args struct {
		student          DomainStudent
		isEnableUsername bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "bad case: validate student with username",
			args: args{
				student: DomainStudent{
					DomainStudent: &mock.Student{
						RandomStudent: mock.RandomStudent{
							Email:            field.NewString(strings.ToLower(fmt.Sprintf("email_%s@manabie.com", idutil.ULIDNow()))),
							FirstName:        field.NewString("first_name"),
							LastName:         field.NewString("last_name"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
						},
					},
					IndexAttr: 0,
				},
				isEnableUsername: true,
			},
			wantErr: entity.MissingMandatoryFieldError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			},
		},
		{
			name: "happy case: validate student with username",
			args: args{
				student: DomainStudent{
					DomainStudent: &mock.Student{
						RandomStudent: mock.RandomStudent{
							Email:            field.NewString(strings.ToLower(fmt.Sprintf("email_%s@manabie.com", idutil.ULIDNow()))),
							UserName:         field.NewString("username"),
							FirstName:        field.NewString("first_name"),
							LastName:         field.NewString("last_name"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
						},
					},
				},
				isEnableUsername: true,
			},
			wantErr: nil,
		},
		{
			name: "happy case: validate normal student",
			args: args{
				student: DomainStudent{
					DomainStudent: &mock.Student{
						RandomStudent: mock.RandomStudent{
							Email:            field.NewString(strings.ToLower(fmt.Sprintf("email_%s@manabie.com", idutil.ULIDNow()))),
							FirstName:        field.NewString("first_name"),
							LastName:         field.NewString("last_name"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
						},
					},
				},
				isEnableUsername: false,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidStudent(tt.args.student, tt.args.isEnableUsername); err != tt.wantErr {
				t.Errorf("ValidStudent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
