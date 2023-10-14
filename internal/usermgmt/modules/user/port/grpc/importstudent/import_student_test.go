package importstudent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	entity_mock "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
)

type DomainStudent interface {
	GetUsersByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.Users, error)
	GetGradesByExternalIDs(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error)
	GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error)
	GetLocationsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainLocations, error)
	GetSchoolsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchools, error)
	GetSchoolCoursesByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error)
	GetPrefecturesByCodes(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error)
	ValidateUpdateSystemAndExternalUserID(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error
	UpsertMultiple(ctx context.Context, option unleash.DomainStudentFeatureOption, studentsToCreate ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error)
	GetEmailWithStudentID(ctx context.Context, studentIDs []string) (map[string]entity.User, error)
	IsFeatureIgnoreInvalidRecordsCSVAndOpenAPIEnabled(organization valueobj.HasOrganizationID) bool
	UpsertMultipleWithErrorCollection(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error)
	IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool
	IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(organization valueobj.HasOrganizationID) bool
	IsDisableAutoDeactivateStudents(organization valueobj.HasOrganizationID) bool
	IsExperimentalBulkInsertEnrollmentStatusHistories(organization valueobj.HasOrganizationID) bool
	IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error)
}

var _ DomainStudent = (*mockDomainStudentService)(nil)

type mockDomainStudentService struct {
	getUsersByExternalIDs                                 func(ctx context.Context, externalUserIDs []string) (entity.Users, error)
	getGradesByExternalIDs                                func(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error)
	getTagsByExternalIDs                                  func(ctx context.Context, externalIDs []string) (entity.DomainTags, error)
	getLocationsByExternalIDs                             func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error)
	getSchoolsByExternalIDs                               func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error)
	getSchoolCoursesByExternalIDs                         func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error)
	getPrefecturesByCodes                                 func(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error)
	upsertMultiple                                        func(ctx context.Context, option unleash.DomainStudentFeatureOption, studentsToCreate ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error)
	validateUpdateSystemAndExternalUserID                 func(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error
	getEmailWithStudentID                                 func(ctx context.Context, studentIDs []string) (map[string]entity.User, error)
	isFeatureIgnoreUpdateEmailEnabled                     func(organization valueobj.HasOrganizationID) bool
	isFeatureIgnoreInvalidRecordsCSVAndOpenAPIEnabled     func(organization valueobj.HasOrganizationID) bool
	upsertMultipleWithErrorCollection                     func(ctx context.Context, studentWithIndexes aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error)
	isFeatureUserNameStudentParentEnabled                 func(organization valueobj.HasOrganizationID) bool
	isFeatureAutoDeactivateAndReactivateStudentsV2Enabled func(organization valueobj.HasOrganizationID) bool
	isDisableAutoDeactivateStudents                       func(organization valueobj.HasOrganizationID) bool
	isExperimentalBulkInsertEnrollmentStatusHistories     func(organization valueobj.HasOrganizationID) bool
}

func (d *mockDomainStudentService) GetUsersByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
	return d.getUsersByExternalIDs(ctx, externalUserIDs)
}

func (d *mockDomainStudentService) GetGradesByExternalIDs(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error) {
	return d.getGradesByExternalIDs(ctx, externalIDs)
}

func (d *mockDomainStudentService) GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
	return d.getTagsByExternalIDs(ctx, externalIDs)
}

func (d *mockDomainStudentService) GetLocationsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
	return d.getLocationsByExternalIDs(ctx, externalIDs)
}

func (d *mockDomainStudentService) GetSchoolsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
	return d.getSchoolsByExternalIDs(ctx, externalIDs)
}

func (d *mockDomainStudentService) GetSchoolCoursesByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
	return d.getSchoolCoursesByExternalIDs(ctx, externalIDs)
}

func (d *mockDomainStudentService) GetPrefecturesByCodes(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error) {
	return d.getPrefecturesByCodes(ctx, codes)
}

func (d *mockDomainStudentService) UpsertMultiple(ctx context.Context, option unleash.DomainStudentFeatureOption, studentsToCreate ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error) {
	return d.upsertMultiple(ctx, option, studentsToCreate...)
}

func (d *mockDomainStudentService) UpsertMultipleWithErrorCollection(ctx context.Context, studentWithIndexes aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
	return d.UpsertMultipleWithErrorCollection(ctx, studentWithIndexes, option)
}

func (d *mockDomainStudentService) ValidateUpdateSystemAndExternalUserID(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error {
	return d.validateUpdateSystemAndExternalUserID(ctx, studentsToUpdate)
}

func (d *mockDomainStudentService) GetEmailWithStudentID(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
	return d.getEmailWithStudentID(ctx, studentIDs)
}

func (d *mockDomainStudentService) IsFeatureIgnoreUpdateEmailEnabled(organization valueobj.HasOrganizationID) bool {
	return d.isFeatureIgnoreUpdateEmailEnabled(organization)
}

func (d *mockDomainStudentService) IsFeatureIgnoreInvalidRecordsCSVAndOpenAPIEnabled(organization valueobj.HasOrganizationID) bool {
	return d.IsFeatureIgnoreInvalidRecordsCSVAndOpenAPIEnabled(organization)
}

func (d *mockDomainStudentService) IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool {
	return d.isFeatureUserNameStudentParentEnabled(organization)
}
func (d *mockDomainStudentService) IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(organization valueobj.HasOrganizationID) bool {
	return d.isFeatureAutoDeactivateAndReactivateStudentsV2Enabled(organization)
}
func (d *mockDomainStudentService) IsDisableAutoDeactivateStudents(organization valueobj.HasOrganizationID) bool {
	return d.isDisableAutoDeactivateStudents(organization)
}
func (d *mockDomainStudentService) IsExperimentalBulkInsertEnrollmentStatusHistories(organization valueobj.HasOrganizationID) bool {
	return d.isExperimentalBulkInsertEnrollmentStatusHistories(organization)
}
func (d *mockDomainStudentService) IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error) {
	return true, nil
}

type mockDomainLocation struct {
	entity.NullDomainLocation
	locationID        field.String
	locationPartnerID field.String
}

func (m *mockDomainLocation) LocationID() field.String {
	return m.locationID
}

func (m *mockDomainLocation) PartnerInternalID() field.String {
	return m.locationPartnerID
}

type mockDomainTag struct {
	tagID             field.String
	partnerInternalID field.String
	tagType           field.String

	entity.EmptyDomainTag
}

func (m *mockDomainTag) TagID() field.String {
	return m.tagID
}

func (m *mockDomainTag) PartnerInternalID() field.String {
	return m.partnerInternalID
}

func (m *mockDomainTag) TagType() field.String {
	return m.tagType
}

type mockDomainGrade struct {
	entity.NullDomainGrade
	gradeID field.String
}

func (m *mockDomainGrade) GradeID() field.String {
	return m.gradeID
}

type mockDomainPrefecture struct {
	entity.DomainPrefecture
	prefectureID field.String
}

func (m *mockDomainPrefecture) PrefectureID() field.String {
	return m.prefectureID
}

func createMockDomainTagWithType(tagID string, tagType pb.UserTagType) entity.DomainTag {
	return &mockDomainTag{
		tagID:             field.NewString(tagID),
		partnerInternalID: field.NewString(fmt.Sprintf("partner-id-%s", tagID)),
		tagType:           field.NewString(tagType.String()),
	}
}

func Test_importCsvValidateTag(t *testing.T) {
	id1 := idutil.ULIDNow()
	id2 := idutil.ULIDNow()
	partnerID1 := field.NewString(fmt.Sprintf("partner-id-%s", id1))
	partnerID2 := field.NewString(fmt.Sprintf("partner-id-%s", id2))

	studentError := fmt.Errorf("tag is only for student")
	parentError := fmt.Errorf("tag is only for parent")

	type args struct {
		role               string
		partnerInternalIDs []string
		tags               entity.DomainTags
	}
	tests := []struct {
		name string
		args func(t *testing.T) args

		wantErr    bool
		inspectErr func(err error, t *testing.T)
	}{
		{
			name: "happy case: valid for student",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleStudent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: []entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_STUDENT),
						createMockDomainTagWithType(id2, pb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT),
					},
				}
			},
			wantErr: false,
		},
		{
			name: "happy case: valid for parent",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleParent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: []entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType(id2, pb.UserTagType_USER_TAG_TYPE_PARENT_DISCOUNT),
					},
				}
			},
			wantErr: false,
		},
		{
			name: "tag is not for student",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleStudent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: []entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType(id2, pb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT),
					},
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, studentError)
			},
		},
		{
			name: "tag is not for parent",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleParent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: []entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_PARENT),
						createMockDomainTagWithType(id2, pb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT),
					},
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, parentError)
			},
		},
		{
			name: "tag is not existed",
			args: func(t *testing.T) args {
				return args{
					role:               constant.RoleStudent,
					partnerInternalIDs: []string{partnerID1.RawValue(), partnerID2.RawValue()},
					tags: []entity.DomainTag{
						createMockDomainTagWithType(id1, pb.UserTagType_USER_TAG_TYPE_STUDENT),
					},
				}
			},
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.Equal(t, err, studentError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			err := importCsvValidateTag(tArgs.role, tArgs.partnerInternalIDs, tArgs.tags)

			if (err != nil) != tt.wantErr {
				t.Fatalf("importCsvValidateTag error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func Test_toUserPhoneNumbers(t *testing.T) {
	type args struct {
		student *StudentCSV
	}
	tests := []struct {
		name string
		args args
		want entity.DomainUserPhoneNumbers
	}{
		{
			name: "without phone number",
			args: args{student: &StudentCSV{}},
			want: entity.DomainUserPhoneNumbers{
				&UserPhoneNumber{
					PhoneTypeAttr:   field.NewString(entity.StudentPhoneNumber),
					PhoneNumberAttr: field.NewUndefinedString(),
				},
				&UserPhoneNumber{
					PhoneTypeAttr:   field.NewString(entity.StudentHomePhoneNumber),
					PhoneNumberAttr: field.NewUndefinedString(),
				},
			},
		},
		{
			name: "has only student phone number",
			args: args{student: &StudentCSV{StudentPhoneNumberAttr: field.NewString("123456789")}},
			want: entity.DomainUserPhoneNumbers{
				&UserPhoneNumber{
					PhoneTypeAttr:   field.NewString(entity.StudentPhoneNumber),
					PhoneNumberAttr: field.NewString("123456789"),
				},
				&UserPhoneNumber{
					PhoneTypeAttr:   field.NewString(entity.StudentHomePhoneNumber),
					PhoneNumberAttr: field.NewUndefinedString(),
				},
			},
		},
		{
			name: "have student phone number and student home phone number",
			args: args{
				student: &StudentCSV{
					StudentPhoneNumberAttr:     field.NewString("123456789"),
					StudentHomePhoneNumberAttr: field.NewString("987654321"),
				},
			},
			want: entity.DomainUserPhoneNumbers{
				&UserPhoneNumber{
					PhoneTypeAttr:   field.NewString(entity.StudentPhoneNumber),
					PhoneNumberAttr: field.NewString("123456789"),
				},
				&UserPhoneNumber{
					PhoneTypeAttr:   field.NewString(entity.StudentHomePhoneNumber),
					PhoneNumberAttr: field.NewString("987654321"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userPhoneNumbers := toUserPhoneNumbers(tt.args.student)
			for idx, wantPhoneNumber := range tt.want {
				assert.Equal(t, wantPhoneNumber.PhoneNumber(), userPhoneNumbers[idx].PhoneNumber())
				assert.Equal(t, wantPhoneNumber.Type(), userPhoneNumbers[idx].Type())
			}
		})
	}
}

func Test_readAndValidatePayload(t *testing.T) {
	type args struct {
		payload []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "payload with size smaller than 5m",
			args: args{
				payload: make([]byte, 1024*1024*4),
			},
			wantErr: nil,
		},
		{
			name: "payload with size equal than 5m",
			args: args{
				payload: make([]byte, 1024*1024*5),
			},
			wantErr: nil,
		},
		{
			name: "payload with size greater than 5m",
			args: args{
				payload: make([]byte, 1024*1024*6),
			},
			wantErr: grpc.InvalidPayloadSizeCSVError{
				RequestSize: 6,
			},
		},
		{
			name: "payload with size equal 0",
			args: args{
				payload: make([]byte, 0),
			},
			wantErr: grpc.InvalidPayloadSizeCSVError{
				RequestSize: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := readAndValidatePayload(tt.args.payload)
			if err != nil {
				assert.Equal(t, tt.wantErr, err)

			}
		})
	}
}

func Test_convertPayloadToImportStudentData(t *testing.T) {
	type args struct {
		payload []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []*StudentCSV
		wantErr error
	}{
		{
			name: "some fields of student",
			args: args{
				payload: []byte(
					"user_id,external_user_id,last_name,first_name,last_name_phonetic,first_name_phonetic,email\n" +
						"9PC,external_user_id,lastname,firstname,lastname_phonetic,firstname_phonetic,studentu001@email.com",
				),
			},
			want: []*StudentCSV{
				{
					IDAttr:                field.NewString("9PC"),
					ExternalUserIDAttr:    field.NewString("external_user_id"),
					LastNameAttr:          field.NewString("lastname"),
					FirstNameAttr:         field.NewString("firstname"),
					FirstNamePhoneticAttr: field.NewString("firstname_phonetic"),
					LastNamePhoneticAttr:  field.NewString("lastname_phonetic"),
					EmailAttr:             field.NewString("studentu001@email.com"),
				},
			},
			wantErr: nil,
		},
		{
			name: "wrong format csv",
			args: args{
				payload: []byte(
					"user_id,external_user_id,\n" +
						"9PC,",
				),
			},
			want: []*StudentCSV(nil),
			wantErr: grpc.InternalError{
				RawErr: errors.Wrap(errors.New("record on line 2: wrong number of fields"), "wrong format csv"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			studentCSVs, err := ConvertPayloadToImportStudentData(tt.args.payload)
			for _, studentCSV := range studentCSVs {
				assert.Equal(t, studentCSV.StudentNote().String(), "")
			}
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}

func TestDomainStudentService_toTaggedUsers(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx      context.Context
		tagField field.String
		user     *StudentCSV
		idx      int
	}

	tests := []struct {
		name    string
		service DomainStudent
		args    args
		want    struct {
			UserIDs []string
			TagIDs  []string
		}
		wantErr error
	}{
		{
			name:    "",
			service: &mockDomainStudentService{},
			args: args{
				ctx:      ctx,
				tagField: field.NewString("1;1"),
				idx:      0,
			},
			wantErr: errcode.Error{
				Code:      errcode.DuplicatedData,
				FieldName: entity.StudentTagsField,
			},
		},
		{
			name: "",
			service: &mockDomainStudentService{
				getTagsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
					return nil, fmt.Errorf("error")
				},
			},
			args: args{
				ctx:      ctx,
				tagField: field.NewString("1"),
				idx:      0,
			},
			wantErr: errcode.Error{
				Code:      errcode.InternalError,
				FieldName: entity.StudentTagsField,
			},
		},
		{
			name: "",
			service: &mockDomainStudentService{
				getTagsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
					return []entity.DomainTag{createMockDomainTagWithType("1", pb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT)}, nil
				},
			},
			args: args{
				ctx:      ctx,
				tagField: field.NewString("1"),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
				idx: 0,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentTagsField,
			},
		},
		{
			name: "",
			service: &mockDomainStudentService{
				getTagsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
					return []entity.DomainTag{createMockDomainTagWithType("1", pb.UserTagType_USER_TAG_TYPE_STUDENT_DISCOUNT)}, nil
				},
			},
			args: args{
				ctx:      ctx,
				tagField: field.NewString("partner-id-1"),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
				idx: 0,
			},
			want: struct {
				UserIDs []string
				TagIDs  []string
			}{
				UserIDs: []string{"user_id"},
				TagIDs:  []string{"1"},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{
				DomainStudent: tt.service,
			}
			got, err := d.toTaggedUsers(tt.args.ctx, tt.args.tagField, tt.args.user, tt.args.idx)
			if tt.wantErr != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
			}
			if tt.wantErr == nil {
				assert.Equal(t, tt.want.UserIDs, got.UserIDs())
				assert.Equal(t, tt.want.TagIDs, got.TagIDs())
			}
		})
	}

}

func TestDomainStudentService_toTaggedUsersV2(t *testing.T) {
	type args struct {
		tagField field.String
		user     *StudentCSV
	}

	tests := []struct {
		name            string
		args            args
		wantTaggedUsers entity.DomainTaggedUsers
		wantTags        entity.DomainTags
	}{
		{
			name: "Happy case: multiple tags",
			args: args{
				tagField: field.NewString("1;2"),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
			},
			wantTaggedUsers: entity.DomainTaggedUsers{
				entity.TaggedUserWillBeDelegated{
					HasUserID: &StudentCSV{
						IDAttr: field.NewString("user_id"),
					},
				},
				entity.TaggedUserWillBeDelegated{
					HasUserID: &StudentCSV{
						IDAttr: field.NewString("user_id"),
					},
				},
			},
			wantTags: entity.DomainTags{
				entity.TagWillBeDelegated{
					HasPartnerInternalID: TagImpl{
						StudentTagAttr: field.NewString("1"),
					},
				},
				entity.TagWillBeDelegated{
					HasPartnerInternalID: TagImpl{
						StudentTagAttr: field.NewString("2"),
					},
				},
			},
		},
		{
			name: "Happy case: single tag",
			args: args{
				tagField: field.NewString("1"),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
			},
			wantTaggedUsers: entity.DomainTaggedUsers{
				entity.TaggedUserWillBeDelegated{
					HasUserID: &StudentCSV{
						IDAttr: field.NewString("user_id"),
					},
				},
			},
			wantTags: entity.DomainTags{
				entity.TagWillBeDelegated{
					HasPartnerInternalID: TagImpl{
						StudentTagAttr: field.NewString("1"),
					},
				},
			},
		},
		{
			name: "Happy case: empty tag",
			args: args{
				tagField: field.NewNullString(),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
			},
			wantTaggedUsers: entity.DomainTaggedUsers{},
			wantTags:        entity.DomainTags{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taggedUsers, tags := toTaggedUsersV2(tt.args.tagField, tt.args.user)

			assert.Equal(t, tt.wantTaggedUsers, taggedUsers)
			assert.Equal(t, tt.wantTags, tags)
		})
	}

}

func TestDomainStudentService_toInternalGrade(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx  context.Context
		user *StudentCSV
		idx  int
	}

	var tests = []struct {
		name    string
		service DomainStudent
		args    args
		want    field.String
		wantErr error
	}{
		{
			name: "",
			service: &mockDomainStudentService{
				getGradesByExternalIDs: func(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error) {
					return nil, fmt.Errorf("error")
				},
			},
			args: args{
				ctx:  ctx,
				user: &StudentCSV{GradeAttr: field.NewString("1")},
				idx:  0,
			},
			wantErr: errcode.Error{
				Code:      errcode.InternalError,
				FieldName: entity.StudentGradeField,
			},
		},
		{
			name: "",
			service: &mockDomainStudentService{
				getGradesByExternalIDs: func(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error) {
					return []entity.DomainGrade{}, nil
				},
			},
			args: args{
				ctx:  ctx,
				user: &StudentCSV{GradeAttr: field.NewString("1")},
				idx:  0,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentGradeField,
			},
		},
		{
			name: "",
			service: &mockDomainStudentService{
				getGradesByExternalIDs: func(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error) {
					return []entity.DomainGrade{
						&mockDomainGrade{gradeID: field.NewString("1")},
					}, nil
				},
			},
			args: args{
				ctx:  ctx,
				user: &StudentCSV{GradeAttr: field.NewString("1")},
				idx:  0,
			},
			want:    field.NewString("1"),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{DomainStudent: tt.service}
			got, err := d.toInternalGrade(tt.args.ctx, tt.args.user, tt.args.idx)
			if tt.wantErr != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
			}
			if tt.wantErr == nil {
				assert.Equal(t, tt.want.String(), got.String())
			}
		})
	}
}

func TestDomainStudentService_toInternalGradeV2(t *testing.T) {
	type args struct {
		user *StudentCSV
	}

	var tests = []struct {
		name    string
		args    args
		want    entity.DomainGrade
		wantErr error
	}{
		{
			name: "Happy case: valid grade",

			args: args{
				user: &StudentCSV{GradeAttr: field.NewString("grade_id")},
			},
			wantErr: nil,
			want: entity.GradeWillBeDelegated{
				HasPartnerInternalID: GradeImpl{
					GradeAttr: field.NewString("grade_id"),
				},
			},
		},

		{
			name: "Happy case: grade is not present",
			args: args{
				user: &StudentCSV{},
			},
			wantErr: nil,
			want:    entity.NullDomainGrade{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toInternalGradeV2(tt.args.user)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDomainStudentService_toContactPreference(t *testing.T) {
	type args struct {
		reference field.String
		idx       int
	}
	tests := []struct {
		name    string
		args    args
		want    field.String
		wantErr error
	}{
		{
			name: "",
			args: args{
				reference: field.NewString("1"),
				idx:       0,
			},
			want: field.NewString(entity.StudentPhoneNumber),
		},
		{
			name: "",
			args: args{
				reference: field.NewString("1000"),
				idx:       0,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentFieldContactPreference,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{}
			got, err := d.toContactPreference(tt.args.reference, tt.args.idx)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.Equalf(t, tt.want, got, "toContactPreference(%v, %v)", tt.args.reference, tt.args.idx)
			}
		})
	}
}

func TestDomainStudentService_toContactPreferenceV2(t *testing.T) {
	type args struct {
		reference field.String
	}
	tests := []struct {
		name    string
		args    args
		want    field.String
		wantErr error
	}{
		{
			name: "Happy case: map to phone number student contact preference",
			args: args{
				reference: field.NewString("1"),
			},
			want: field.NewString(entity.StudentPhoneNumber),
		},
		{
			name: "Happy case: don't map to phone number student contact preference",
			args: args{
				reference: field.NewString("1000"),
			},
			wantErr: entity.InvalidFieldError{
				FieldName:  entity.StudentFieldContactPreference,
				EntityName: entity.StudentEntity,
				Index:      0,
				Reason:     entity.NotMatchingEnum,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toContactPreferenceV2(tt.args.reference, 0)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr == nil {
				assert.Equalf(t, tt.want, got, "toContactPreference(%v)", tt.args.reference)
			}
		})
	}
}

func TestDomainStudentService_toUserAccessPaths(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx      context.Context
		location field.String
		user     *StudentCSV
		idx      int
	}

	tests := []struct {
		name    string
		service DomainStudent
		args    args
		want    struct {
			UserIDs     []string
			LocationIDs []string
		}
		wantErr error
	}{
		{
			name: "",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return nil, fmt.Errorf("error")
				},
			},
			args: args{
				ctx:      ctx,
				location: field.NewString("1"),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
				idx: 0,
			},
			want: struct {
				UserIDs     []string
				LocationIDs []string
			}{
				UserIDs:     []string{"user_id"},
				LocationIDs: []string{"1"},
			},
			wantErr: errcode.Error{
				Code:      errcode.InternalError,
				FieldName: entity.StudentLocationsField,
			},
		},
		{
			name: "",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return []entity.DomainLocation{&mockDomainLocation{locationID: field.NewString("1")}}, nil
				},
			},
			args: args{
				ctx:      ctx,
				location: field.NewString("1"),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
				idx: 0,
			},
			want: struct {
				UserIDs     []string
				LocationIDs []string
			}{
				UserIDs:     []string{"user_id"},
				LocationIDs: []string{"1"},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{
				DomainStudent: tt.service,
			}
			got, err := d.toUserAccessPaths(tt.args.ctx, tt.args.location, tt.args.user, tt.args.idx)
			if tt.wantErr != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
			}
			if tt.wantErr == nil {
				assert.Equal(t, tt.want.UserIDs, got.UserIDs())
				assert.Equal(t, tt.want.LocationIDs, got.LocationIDs())
			}
		})
	}
}

func TestDomainStudentService_toUserAccessPathsV2(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx      context.Context
		location field.String
		user     *StudentCSV
	}

	tests := []struct {
		name string
		args args
		want entity.DomainUserAccessPaths
	}{
		{
			name: "Happy case: single location",
			args: args{
				ctx:      ctx,
				location: field.NewString("partner-id-01"),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
			},
			want: entity.DomainUserAccessPaths{
				entity.UserAccessPathWillBeDelegated{
					HasUserID: &StudentCSV{
						IDAttr: field.NewString("user_id"),
					},
				},
			},
		},
		{
			name: "Happy case: empty location",
			args: args{
				ctx:      ctx,
				location: field.NewNullString(),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
			},
			want: entity.DomainUserAccessPaths{
				entity.UserAccessPathWillBeDelegated{
					HasUserID: &StudentCSV{
						IDAttr: field.NewString("user_id"),
					},
				},
			},
		},
		{
			name: "Happy case: multiple locations",
			args: args{
				ctx:      ctx,
				location: field.NewString("partner-id-01;partner-id-01"),
				user: &StudentCSV{
					IDAttr: field.NewString("user_id"),
				},
			},
			want: entity.DomainUserAccessPaths{
				entity.UserAccessPathWillBeDelegated{
					HasUserID: &StudentCSV{
						IDAttr: field.NewString("user_id"),
					},
				},
				entity.UserAccessPathWillBeDelegated{
					HasUserID: &StudentCSV{
						IDAttr: field.NewString("user_id"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toUserAccessPathsV2(tt.args.location, tt.args.user)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDomainStudentService_toUserAddress(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx  context.Context
		user *StudentCSV
		idx  int
	}

	tests := []struct {
		name    string
		service DomainStudent
		args    args
		want    *UserAddress
		wantErr error
	}{
		{
			name: "",
			service: &mockDomainStudentService{
				getPrefecturesByCodes: func(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error) {
					return nil, fmt.Errorf("error")
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					PrefectureAttr: field.NewString("PrefectureAttr"),
				},
				idx: 0,
			},
			wantErr: errcode.Error{
				Code:      errcode.InternalError,
				FieldName: entity.StudentUserAddressPrefectureField,
			},
		},
		{
			name: "",
			service: &mockDomainStudentService{
				getPrefecturesByCodes: func(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error) {
					return []entity.DomainPrefecture{}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					PrefectureAttr: field.NewString("PrefectureAttr"),
				},
				idx: 0,
			},
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentUserAddressPrefectureField,
			},
		},
		{
			name: "",
			service: &mockDomainStudentService{
				getPrefecturesByCodes: func(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error) {
					return []entity.DomainPrefecture{&mockDomainPrefecture{prefectureID: field.NewString("prefecture-1")}}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					PostalCodeAttr:   field.NewString("PostalCodeAttr"),
					CityAttr:         field.NewString("CityAttr"),
					FirstStreetAttr:  field.NewString("FirstStreetAttr"),
					SecondStreetAttr: field.NewString("SecondStreetAttr"),
					PrefectureAttr:   field.NewString("PrefectureAttr"),
				},
				idx: 0,
			},
			want: &UserAddress{
				AddressTypeAttr:  field.NewString(pb.AddressType_HOME_ADDRESS.String()),
				PostalCodeAttr:   field.NewString("PostalCodeAttr"),
				CityAttr:         field.NewString("CityAttr"),
				FirstStreetAttr:  field.NewString("FirstStreetAttr"),
				SecondStreetAttr: field.NewString("SecondStreetAttr"),
				PrefectureAttr:   field.NewString("prefecture-1"),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{
				DomainStudent: tt.service,
			}
			got, err := d.toUserAddress(tt.args.ctx, tt.args.user, tt.args.idx)
			if tt.wantErr != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
			}
			if tt.wantErr == nil {
				assert.Equal(t, tt.want.AddressType(), got.AddressType())
				assert.Equal(t, tt.want.PostalCode(), got.PostalCode())
				assert.Equal(t, tt.want.City(), got.City())
				assert.Equal(t, tt.want.FirstStreet(), got.FirstStreet())
				assert.Equal(t, tt.want.SecondStreet(), got.SecondStreet())
				assert.Equal(t, tt.want.PrefectureID(), got.PrefectureID())
			}
		})
	}
}

func TestDomainStudentService_toUserAddressV2(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx  context.Context
		user *StudentCSV
	}

	tests := []struct {
		name                string
		args                args
		expectedUserAddress *UserAddress
		expectedPrefecture  *PrefectureImpl
	}{
		{
			name: "Happy case: full data",
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					PostalCodeAttr:   field.NewString("70000"),
					FirstStreetAttr:  field.NewString("12 Ton Dan"),
					SecondStreetAttr: field.NewString("1 Nguyen Du"),
					PrefectureAttr:   field.NewString("35"),
				},
			},
			expectedUserAddress: &UserAddress{
				FirstStreetAttr:  field.NewString("12 Ton Dan"),
				SecondStreetAttr: field.NewString("1 Nguyen Du"),
				PostalCodeAttr:   field.NewString("70000"),
				AddressTypeAttr:  field.NewString(pb.AddressType_HOME_ADDRESS.String()),
			},
			expectedPrefecture: &PrefectureImpl{
				PrefectureCodeAttr: field.NewString("35"),
			},
		},
		{
			name: "Happy case: empty prefecture",
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					PostalCodeAttr: field.NewString("70000"),
				},
			},
			expectedUserAddress: &UserAddress{
				PostalCodeAttr:  field.NewString("70000"),
				AddressTypeAttr: field.NewString(pb.AddressType_HOME_ADDRESS.String()),
			},
			expectedPrefecture: &PrefectureImpl{},
		},
		{
			name: "Happy case: empty data",
			args: args{
				ctx:  ctx,
				user: &StudentCSV{},
			},
			expectedUserAddress: &UserAddress{
				AddressTypeAttr: field.NewString(pb.AddressType_HOME_ADDRESS.String()),
			},
			expectedPrefecture: &PrefectureImpl{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userAddress, prefecture := toUserAddressV2(tt.args.user)
			assert.Equal(t, tt.expectedUserAddress.PostalCodeAttr, userAddress.PostalCode())
			assert.Equal(t, tt.expectedUserAddress.FirstStreetAttr, userAddress.FirstStreet())
			assert.Equal(t, tt.expectedUserAddress.SecondStreetAttr, userAddress.SecondStreet())
			assert.Equal(t, tt.expectedUserAddress.AddressTypeAttr, userAddress.AddressType())
			assert.Equal(t, tt.expectedPrefecture.PrefectureCodeAttr.String(), prefecture.PrefectureCode().String())
		})
	}
}

func TestDomainStudentService_toSchoolHistory(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx  context.Context
		user []byte
		idx  int
	}

	mockSchoolInfo_1 := entity_mock.School{
		RandomSchool: entity_mock.RandomSchool{
			SchoolID:          field.NewString("school-id-1"),
			SchoolLevelID:     field.NewString("school-level-id-1"),
			PartnerInternalID: field.NewString("school-partner-id-1"),
		},
	}
	mockSchoolInfo_2 := entity_mock.School{
		RandomSchool: entity_mock.RandomSchool{
			SchoolID:          field.NewString("school-id-2"),
			SchoolLevelID:     field.NewString("school-level-id-2"),
			PartnerInternalID: field.NewString("school-partner-id-2"),
		},
	}
	mockSchoolCourse_1 := entity_mock.SchoolCourse{
		RandomSchoolCourse: entity_mock.RandomSchoolCourse{
			SchoolCourseID:    field.NewString("school-course-id-1"),
			PartnerInternalID: field.NewString("school-course-partner-id-1"),
		},
	}
	mockSchoolCourse_2 := entity_mock.SchoolCourse{
		RandomSchoolCourse: entity_mock.RandomSchoolCourse{
			SchoolCourseID:    field.NewString("school-course-id-2"),
			PartnerInternalID: field.NewString("school-course-partner-id-2"),
		},
	}
	mockStartDate, _ := time.Parse(constant.DateLayout, "2020/10/10")
	mockEndDate, _ := time.Parse(constant.DateLayout, "2020/10/11")

	tests := []struct {
		name    string
		service DomainStudent
		args    args
		want    entity.DomainSchoolHistories
		wantErr error
	}{
		{
			name: "return correct school histories data",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil
				},
				getSchoolCoursesByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1, mockSchoolCourse_2}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1;school-partner-id-2,school-course-partner-id-1;school-course-partner-id-2,2020/10/10;2020/10/10,2020/10/11;2020/10/11",
				),
				idx: 0,
			},
			want: entity.DomainSchoolHistories{
				&SchoolHistory{
					SchoolIDAttr:       field.NewString("school-id-1"),
					SchoolCourseIDAttr: field.NewString("school-course-id-1"),
					StartDateAttr:      field.NewTime(mockStartDate),
					EndDateAttr:        field.NewTime(mockEndDate),
				},
				&SchoolHistory{
					SchoolIDAttr:       field.NewString("school-id-2"),
					SchoolCourseIDAttr: field.NewString("school-course-id-2"),
					StartDateAttr:      field.NewTime(mockStartDate),
					EndDateAttr:        field.NewTime(mockEndDate),
				},
			},
			wantErr: nil,
		},
		{
			name: "return correct school histories data with empty optional fields",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil
				},
				getSchoolCoursesByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_2}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1;school-partner-id-2,;school-course-partner-id-2,;2020/10/10,;2020/10/11",
				),
				idx: 0,
			},
			want: entity.DomainSchoolHistories{
				&SchoolHistory{
					SchoolIDAttr:       field.NewString("school-id-1"),
					SchoolCourseIDAttr: field.NewNullString(),
					StartDateAttr:      field.NewNullTime(),
					EndDateAttr:        field.NewNullTime(),
				},
				&SchoolHistory{
					SchoolIDAttr:       field.NewString("school-id-2"),
					SchoolCourseIDAttr: field.NewString("school-course-id-2"),
					StartDateAttr:      field.NewTime(mockStartDate),
					EndDateAttr:        field.NewTime(mockEndDate),
				},
			},
			wantErr: nil,
		},
		{
			name: "return correct school histories data with 1 school and empty optional fields",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1,,,",
				),
				idx: 0,
			},
			want: entity.DomainSchoolHistories{
				&SchoolHistory{
					SchoolIDAttr:       field.NewString("school-id-1"),
					SchoolCourseIDAttr: field.NewNullString(),
					StartDateAttr:      field.NewNullTime(),
					EndDateAttr:        field.NewNullTime(),
				},
			},
			wantErr: nil,
		},
		{
			name: "return nil when all fields is empty",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						",,,",
				),
				idx: 0,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "return error when school is empty and school course is not empty",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						",school-course-partner-id-1,,",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolCourseField,
			},
		},
		{
			name: "return error when school is empty and start date is not empty",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						",,2020/10/10,",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolHistoryStartDateField,
			},
		},
		{
			name: "return error when school is empty and end date is not empty",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						",,,2020/10/11",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolHistoryEndDateField,
			},
		},
		{
			name:    "return error when school courses mismatch with school",
			service: &mockDomainStudentService{},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1;school-partner-id-2,school-course-partner-id-1,2020/10/10;2020/10/10,2020/10/11;2020/10/11",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolCourseField,
			},
		},
		{
			name:    "return error when start dates mismatch with school",
			service: &mockDomainStudentService{},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1;school-partner-id-2,school-course-partner-id-1;school-course-partner-id-2,2020/10/10;2020/10/10;,2020/10/11;2020/10/11",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolHistoryStartDateField,
			},
		},
		{
			name:    "return error when end dates mismatch with school",
			service: &mockDomainStudentService{},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1;school-partner-id-2,school-course-partner-id-1;school-course-partner-id-2,2020/10/10;2020/10/10,",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolHistoryEndDateField,
			},
		},
		{
			name: "return error when start date is invalid",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
				getSchoolCoursesByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1,school-course-partner-id-1,2020/10,2020/10/11",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolHistoryStartDateField,
			},
		},
		{
			name: "return error when start date is invalid",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
				getSchoolCoursesByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1,school-course-partner-id-1,2020/10/10,2020/10",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolHistoryEndDateField,
			},
		},
		{
			name: "return error when school does not exist in DB",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1;school-partner-id-2,school-course-partner-id-1;school-course-partner-id-2,2020/10/10;2020/10/10,2020/10/11;2020/10/11",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolField,
			},
		},
		{
			name: "return error when school course does not exist in DB",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil
				},
				getSchoolCoursesByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_2}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1;school-partner-id-2,school-course-partner-id-1;school-course-partner-id-2,2020/10/10;2020/10/10,2020/10/11;2020/10/11",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentSchoolCourseField,
			},
		},
		{
			name: "return error when get school infos failed",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return nil, fmt.Errorf("query school-info error")
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1,school-course-partner-id-1,2020/10/10,2020/10/11",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InternalError,
				FieldName: entity.StudentSchoolField,
				Err:       fmt.Errorf("query school-info error"),
			},
		},
		{
			name: "return error when get school courses failed",
			service: &mockDomainStudentService{
				getSchoolsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
				getSchoolCoursesByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return nil, fmt.Errorf("query school-course error")
				},
			},
			args: args{
				ctx: ctx,
				user: []byte(
					"school,school_course,start_date,end_date\n" +
						"school-partner-id-1,school-course-partner-id-1,2020/10/10,2020/10/11",
				),
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InternalError,
				FieldName: entity.StudentSchoolCourseField,
				Err:       fmt.Errorf("query school-course error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{
				DomainStudent: tt.service,
			}
			studentCSVs, err := ConvertPayloadToImportStudentData(tt.args.user)
			schoolHistories, err := d.toSchoolHistory(tt.args.ctx, studentCSVs[0], tt.args.idx)
			if tt.want != nil {
				if len(schoolHistories) != len(tt.want) {
					panic(fmt.Errorf("expect len: %v - actual len: %v", len(tt.want), len(schoolHistories)))
				}
				for i, school := range schoolHistories {
					assert.Equal(t, tt.want[i].SchoolID(), school.SchoolID())
					assert.Equal(t, tt.want[i].SchoolCourseID(), school.SchoolCourseID())
					assert.Equal(t, tt.want[i].StartDate(), school.StartDate())
					assert.Equal(t, tt.want[i].EndDate(), school.EndDate())
				}
			} else {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
			}
		})
	}
}

func TestDomainStudentService_toSchoolHistoryV2(t *testing.T) {
	type args struct {
		student *StudentCSV
	}
	mockStartDate, _ := time.Parse(constant.DateLayout, "2020/10/10")
	mockEndDate, _ := time.Parse(constant.DateLayout, "2020/10/11")

	tests := []struct {
		name                string
		args                args
		wantSchoolHistories entity.DomainSchoolHistories
		wantSchools         entity.DomainSchools
		wantSchoolCourses   entity.DomainSchoolCourses
		wantErr             error
	}{
		{
			name: "Happy case: full and multiple data school histories data",

			args: args{
				&StudentCSV{
					SchoolAttr:       field.NewString("school-partner-id-1;school-partner-id-2"),
					SchoolCourseAttr: field.NewString("school-course-partner-id-1;school-course-partner-id-2"),
					StartDateAttr:    field.NewString("2020/10/10;2020/10/10"),
					EndDateAttr:      field.NewString("2020/10/11;2020/10/11"),
				},
			},
			wantSchoolHistories: entity.DomainSchoolHistories{
				&SchoolHistory{
					StartDateAttr: field.NewTime(mockStartDate),
					EndDateAttr:   field.NewTime(mockEndDate),
				},
				&SchoolHistory{
					StartDateAttr: field.NewTime(mockStartDate),
					EndDateAttr:   field.NewTime(mockEndDate),
				},
			},
			wantSchools: entity.DomainSchools{
				&SchoolInfoImpl{
					SchoolPartnerInternalIDAttr: field.NewString("school-partner-id-1"),
				},
				&SchoolInfoImpl{
					SchoolPartnerInternalIDAttr: field.NewString("school-partner-id-2"),
				},
			},
			wantSchoolCourses: entity.DomainSchoolCourses{
				&SchoolCourseImpl{
					SchoolCoursePartnerInternalIDAttr: field.NewString("school-course-partner-id-1"),
				},
				&SchoolCourseImpl{
					SchoolCoursePartnerInternalIDAttr: field.NewString("school-course-partner-id-2"),
				},
			},
			wantErr: nil,
		},
		{
			name: "Happy case: full and single data school histories data",

			args: args{
				&StudentCSV{
					SchoolAttr:       field.NewString("school-partner-id-1"),
					SchoolCourseAttr: field.NewString("school-course-partner-id-1"),
					StartDateAttr:    field.NewString("2020/10/10"),
					EndDateAttr:      field.NewString("2020/10/11"),
				},
			},
			wantSchoolHistories: entity.DomainSchoolHistories{
				&SchoolHistory{
					StartDateAttr: field.NewTime(mockStartDate),
					EndDateAttr:   field.NewTime(mockEndDate),
				},
			},

			wantSchools: entity.DomainSchools{
				&SchoolInfoImpl{
					SchoolPartnerInternalIDAttr: field.NewString("school-partner-id-1"),
				},
			},
			wantSchoolCourses: entity.DomainSchoolCourses{
				&SchoolCourseImpl{
					SchoolCoursePartnerInternalIDAttr: field.NewString("school-course-partner-id-1"),
				},
			},

			wantErr: nil,
		},
		{
			name: "Happy case: multiple data school histories data only school",

			args: args{
				&StudentCSV{
					SchoolAttr: field.NewString("school-partner-id-1;school-partner-id-2"),
				},
			},
			wantSchoolHistories: entity.DomainSchoolHistories{
				SchoolHistory{},
				SchoolHistory{},
			},
			wantSchools: entity.DomainSchools{
				&SchoolInfoImpl{
					SchoolPartnerInternalIDAttr: field.NewString("school-partner-id-1"),
				},
				&SchoolInfoImpl{
					SchoolPartnerInternalIDAttr: field.NewString("school-partner-id-2"),
				},
			},
			wantSchoolCourses: entity.DomainSchoolCourses{
				&SchoolCourseImpl{
					SchoolCoursePartnerInternalIDAttr: field.NewString(""),
				},
				&SchoolCourseImpl{
					SchoolCoursePartnerInternalIDAttr: field.NewString(""),
				},
			},
			wantErr: nil,
		},
		{
			name: "Happy case: return nill if all field is empty",

			args: args{
				&StudentCSV{},
			},
			wantSchoolHistories: nil,
			wantErr:             nil,
			wantSchools:         nil,
			wantSchoolCourses:   nil,
		},
		{
			name: "Unhappy case: Don't have school but have school course",

			args: args{
				&StudentCSV{
					SchoolCourseAttr: field.NewString("school-course-partner-id-1"),
					StartDateAttr:    field.NewString("2020/10/10"),
					EndDateAttr:      field.NewString("2020/10/11"),
				},
			},
			wantSchoolHistories: nil,
			wantSchools:         nil,
			wantSchoolCourses:   nil,
			wantErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentSchoolCourseField,
				Index:      0,
				Reason:     entity.NotPresentField,
			},
		},
		{
			name: "Unhappy case: One of school is empty",

			args: args{
				&StudentCSV{
					SchoolAttr:       field.NewString("school-partner-id-1;"),
					SchoolCourseAttr: field.NewString("school-course-partner-id-1;school-course-partner-id-2"),
					StartDateAttr:    field.NewString("2020/10/10"),
					EndDateAttr:      field.NewString("2020/10/11"),
				},
			},
			wantSchoolHistories: nil,
			wantSchools:         nil,
			wantSchoolCourses:   nil,
			wantErr: entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentSchoolField,
				Index:      0,
				Reason:     entity.Empty,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schoolHistories, schoolInfos, schoolCourses, err := toSchoolHistoryV2(tt.args.student, 0)
			if tt.wantSchoolHistories != nil {
				if len(schoolHistories) != len(tt.wantSchoolHistories) {
					panic(fmt.Errorf("expect schoolHistories len: %v - actual len: %v", len(tt.wantSchoolHistories), len(schoolHistories)))
				}
				for i, school := range schoolHistories {
					assert.Equal(t, tt.wantSchoolHistories[i].StartDate(), school.StartDate())
					assert.Equal(t, tt.wantSchoolHistories[i].EndDate(), school.EndDate())
				}
			}

			if tt.wantSchools != nil {
				if len(schoolInfos) != len(tt.wantSchools) {
					panic(fmt.Errorf("expect schoolInfos len: %v - actual len: %v", len(tt.wantSchools), len(schoolInfos)))
				}
				for i, school := range schoolInfos {
					assert.Equal(t, tt.wantSchools[i].PartnerInternalID(), school.PartnerInternalID())
				}
			}

			if tt.wantSchoolCourses != nil {
				if len(schoolCourses) != len(tt.wantSchoolCourses) {
					panic(fmt.Errorf("expect schoolCourses len: %v - actual len: %v", len(tt.wantSchoolCourses), len(schoolCourses)))
				}
				for i, school := range schoolCourses {
					assert.Equal(t, tt.wantSchoolCourses[i].PartnerInternalID(), school.PartnerInternalID())
				}
			}

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}

func TestDomainStudentService_toEnrollmentStatusHistories(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx  context.Context
		user *StudentCSV
		idx  int
	}

	startDate := field.NewNullTime()
	_ = startDate.UnmarshalCSV("2020/10/10")

	tests := []struct {
		name    string
		service DomainStudent
		args    args
		want    entity.DomainEnrollmentStatusHistories
		wantErr error
	}{
		{
			name: "can not get location by external id",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return nil, fmt.Errorf("3rr0f")
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					LocationAttr:         field.NewString("location-id"),
					EnrollmentStatusAttr: field.NewString("1"),
					StatusStartDateAttr:  field.NewString("2020/10/10"),
				},
				idx: 1,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InternalError,
				Err:       fmt.Errorf("3rr0f"),
				FieldName: entity.StudentLocationsField,
				Index:     1,
			},
		},
		{
			name: "len location is not match with external id",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return []entity.DomainLocation{
						&mockDomainLocation{
							locationID:        field.NewString("locationID"),
							locationPartnerID: field.NewString("location-id"),
						},
						&mockDomainLocation{
							locationID:        field.NewString("locationID1"),
							locationPartnerID: field.NewString("location-id1"),
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					LocationAttr:         field.NewString("location-id"),
					EnrollmentStatusAttr: field.NewString("1"),
					StatusStartDateAttr:  field.NewString("2020/10/10"),
				},
				idx: 1,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				Err:       errcode.ErrUserLocationsAreInvalid,
				FieldName: entity.StudentLocationsField,
				Index:     1,
			},
		},
		{
			name: "parse status start date failed",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return []entity.DomainLocation{
						&mockDomainLocation{
							locationID:        field.NewString("locationID"),
							locationPartnerID: field.NewString("location-id"),
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					LocationAttr:         field.NewString("location-id"),
					EnrollmentStatusAttr: field.NewString("1"),
					StatusStartDateAttr:  field.NewString("10/10/2020"),
				},
				idx: 0,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				Err:       field.NewNullTime().Ptr().UnmarshalCSV("10/10/2020"),
				FieldName: entity.StudentFieldEnrollmentStatusStartDate,
				Index:     0,
			},
		},
		{
			name: "parse enrollment status failed",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return []entity.DomainLocation{
						&mockDomainLocation{
							locationID:        field.NewString("locationID"),
							locationPartnerID: field.NewString("location-id"),
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					LocationAttr:         field.NewString("location-id"),
					EnrollmentStatusAttr: field.NewString("--"),
					StatusStartDateAttr:  field.NewString("2020/10/10"),
				},
				idx: 2,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				Err:       field.NewNullInt16().Ptr().UnmarshalCSV("--"),
				FieldName: entity.StudentFieldEnrollmentStatus,
				Index:     2,
			},
		},
		{
			name: "bad case: some status, one location",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return []entity.DomainLocation{
						&mockDomainLocation{
							locationID:        field.NewString("location-id-1"),
							locationPartnerID: field.NewString("location-id-1"),
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					LocationAttr:         field.NewString("location-id-1;"),
					EnrollmentStatusAttr: field.NewString("1;1"),
					StatusStartDateAttr:  field.NewString("2020/10/10;"),
				},
				idx: 2,
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				Err:       errcode.ErrUserLocationsAreInvalid,
				FieldName: entity.StudentLocationsField,
				Index:     2,
			},
		},
		{
			name: "happy case: some locations, one status",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return []entity.DomainLocation{
						&mockDomainLocation{
							locationID:        field.NewString("location-id-1"),
							locationPartnerID: field.NewString("location-id-1"),
						},
						&mockDomainLocation{
							locationID:        field.NewString("location-id-2"),
							locationPartnerID: field.NewString("location-id-2"),
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					LocationAttr:         field.NewString("location-id-1;location-id-2"),
					EnrollmentStatusAttr: field.NewString("1;"),
					StatusStartDateAttr:  field.NewString("2020/10/10;"),
				},
				idx: 0,
			},
			want: entity.DomainEnrollmentStatusHistories{
				EnrollmentStatusHistory{
					EnrollmentStatusAttr: field.NewString(entity.StudentEnrollmentStatusPotential),
					LocationIDAttr:       field.NewString("location-id-1"),
					StartDateAttr:        startDate,
				},
				EnrollmentStatusHistory{
					EnrollmentStatusAttr: field.NewNullString(),
					LocationIDAttr:       field.NewString("location-id-2"),
					StartDateAttr:        field.NewNullTime(),
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: get locations with order are not same as input",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return []entity.DomainLocation{
						&mockDomainLocation{
							locationID:        field.NewString("location-id-1"),
							locationPartnerID: field.NewString("location-id-1"),
						},
						&mockDomainLocation{
							locationID:        field.NewString("location-id-2"),
							locationPartnerID: field.NewString("location-id-2"),
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					LocationAttr:         field.NewString("location-id-2;location-id-1"),
					EnrollmentStatusAttr: field.NewString("1;1"),
					StatusStartDateAttr:  field.NewString("2020/10/10;"),
				},
				idx: 0,
			},
			want: entity.DomainEnrollmentStatusHistories{
				EnrollmentStatusHistory{
					EnrollmentStatusAttr: field.NewString(entity.StudentEnrollmentStatusPotential),
					LocationIDAttr:       field.NewString("location-id-2"),
					StartDateAttr:        startDate,
				},
				EnrollmentStatusHistory{
					EnrollmentStatusAttr: field.NewString(entity.StudentEnrollmentStatusPotential),
					LocationIDAttr:       field.NewString("location-id-1"),
					StartDateAttr:        field.NewNullTime(),
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: without enrollment status",
			args: args{
				ctx:  ctx,
				user: &StudentCSV{},
				idx:  0,
			},
			want:    entity.DomainEnrollmentStatusHistories{},
			wantErr: nil,
		},
		{
			name: "happy case",
			service: &mockDomainStudentService{
				getLocationsByExternalIDs: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return []entity.DomainLocation{
						&mockDomainLocation{
							locationID:        field.NewString("locationID"),
							locationPartnerID: field.NewString("location-id"),
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				user: &StudentCSV{
					LocationAttr:         field.NewString("location-id"),
					EnrollmentStatusAttr: field.NewString("1"),
					StatusStartDateAttr:  field.NewString("2020/10/10"),
				},
				idx: 0,
			},
			want: entity.DomainEnrollmentStatusHistories{
				EnrollmentStatusHistory{
					EnrollmentStatusAttr: field.NewString(entity.StudentEnrollmentStatusPotential),
					LocationIDAttr:       field.NewString("locationID"),
					StartDateAttr:        startDate,
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{
				DomainStudent: tt.service,
			}
			got, err := d.toEnrollmentStatusHistories(tt.args.ctx, tt.args.user, tt.args.idx)
			if tt.wantErr != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, wantErr, e)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDomainStudentService_toEnrollmentStatusHistoriesV2(t *testing.T) {
	type args struct {
		user *StudentCSV
	}
	startDate := field.NewNullTime()
	_ = startDate.UnmarshalCSV("2020/10/10")

	t.Run("Happy case: multiple enrollment status histories but missing enrollment status", func(t *testing.T) {
		user := &StudentCSV{
			LocationAttr:         field.NewString("partner-id-01;partner-id-02"),
			EnrollmentStatusAttr: field.NewString("1;2"),
			StatusStartDateAttr:  field.NewString("2020/10/10"),
		}

		wantEnrollmentStatusHistories := entity.DomainEnrollmentStatusHistories{
			EnrollmentStatusHistory{
				EnrollmentStatusAttr: field.NewString(entity.StudentEnrollmentStatusPotential),
				StartDateAttr:        startDate,
			},
			EnrollmentStatusHistory{
				EnrollmentStatusAttr: field.NewString(entity.StudentEnrollmentStatusEnrolled),
				StartDateAttr:        field.NewTime(time.Now()),
			},
		}
		wantLocations := entity.DomainLocations{
			entity.LocationWillBeDelegated{
				HasPartnerInternalID: LocationImpl{
					LocationPartnerInternalAttr: field.NewString("partner-id-01"),
				},
			},
			entity.LocationWillBeDelegated{
				HasPartnerInternalID: LocationImpl{
					LocationPartnerInternalAttr: field.NewString("partner-id-02"),
				},
			},
		}

		enrollmentStatusHistories, domainLocations, err := toEnrollmentStatusHistoriesV2(user, 0)
		assert.Nil(t, err)

		for i := range enrollmentStatusHistories {
			assert.Equal(t, wantEnrollmentStatusHistories[i].EnrollmentStatus(), enrollmentStatusHistories[i].EnrollmentStatus())
			assert.Equal(t, wantEnrollmentStatusHistories[i].LocationID(), enrollmentStatusHistories[i].LocationID())
			assert.Equal(t, wantEnrollmentStatusHistories[i].StartDate().Time().Format(constant.DateLayout), enrollmentStatusHistories[i].StartDate().Time().Format(constant.DateLayout))
		}
		assert.Equal(t, wantLocations, domainLocations)

	})

	t.Run("Happy case: single and full data enrollment status histories", func(t *testing.T) {
		user := &StudentCSV{
			LocationAttr:         field.NewString("partner-id-01"),
			EnrollmentStatusAttr: field.NewString("1"),
			StatusStartDateAttr:  field.NewString("2020/10/10"),
		}

		wantEnrollmentStatusHistories := entity.DomainEnrollmentStatusHistories{
			EnrollmentStatusHistory{
				EnrollmentStatusAttr: field.NewString(entity.StudentEnrollmentStatusPotential),
				StartDateAttr:        startDate,
			},
		}
		wantLocations := entity.DomainLocations{
			entity.LocationWillBeDelegated{
				HasPartnerInternalID: LocationImpl{
					LocationPartnerInternalAttr: field.NewString("partner-id-01"),
				},
			},
		}

		enrollmentStatusHistories, domainLocations, err := toEnrollmentStatusHistoriesV2(user, 0)
		assert.Nil(t, err)
		assert.Equal(t, wantEnrollmentStatusHistories, enrollmentStatusHistories)
		assert.Equal(t, wantLocations, domainLocations)
	})

	t.Run("Happy case: multiple enrollment status histories but missing location id", func(t *testing.T) {
		user := &StudentCSV{
			EnrollmentStatusAttr: field.NewString("1"),
			StatusStartDateAttr:  field.NewString("2020/10/10"),
		}
		wantEnrollmentStatusHistories := entity.DomainEnrollmentStatusHistories{
			EnrollmentStatusHistory{
				EnrollmentStatusAttr: field.NewString(entity.StudentEnrollmentStatusPotential),
				StartDateAttr:        startDate,
			},
		}
		wantLocations := entity.DomainLocations{
			entity.LocationWillBeDelegated{
				HasPartnerInternalID: LocationImpl{
					LocationPartnerInternalAttr: field.NewString(""),
				},
			},
		}

		enrollmentStatusHistories, domainLocations, err := toEnrollmentStatusHistoriesV2(user, 0)
		assert.Nil(t, err)
		assert.Equal(t, wantEnrollmentStatusHistories, enrollmentStatusHistories)
		assert.Equal(t, wantLocations, domainLocations)
	})

	t.Run("Happy case: multiple enrollment status histories but missing enrollment status", func(t *testing.T) {
		user := &StudentCSV{
			LocationAttr:        field.NewString("partner-id-01"),
			StatusStartDateAttr: field.NewString("2020/10/10"),
		}
		wantEnrollmentStatusHistories := entity.DomainEnrollmentStatusHistories{
			EnrollmentStatusHistory{
				StartDateAttr:        startDate,
				EnrollmentStatusAttr: field.NewNullString(),
			},
		}
		wantLocations := entity.DomainLocations{
			entity.LocationWillBeDelegated{
				HasPartnerInternalID: LocationImpl{
					LocationPartnerInternalAttr: field.NewString("partner-id-01"),
				},
			},
		}

		enrollmentStatusHistories, domainLocations, err := toEnrollmentStatusHistoriesV2(user, 0)
		assert.Nil(t, err)
		assert.Equal(t, wantEnrollmentStatusHistories, enrollmentStatusHistories)
		assert.Equal(t, wantLocations, domainLocations)
	})

	t.Run("Unhappy case: throw error when UnmarshalCSV start date failed with wrong format", func(t *testing.T) {
		user := &StudentCSV{
			LocationAttr:         field.NewString("partner-id-01"),
			StatusStartDateAttr:  field.NewString("10/10/2020"),
			EnrollmentStatusAttr: field.NewString("1"),
		}

		wantErr := entity.InvalidFieldError{
			FieldName:  entity.StudentFieldEnrollmentStatusStartDate,
			Index:      0,
			EntityName: entity.StudentEntity,
			Reason:     entity.FailedUnmarshal,
		}
		enrollmentStatusHistories, domainLocations, err := toEnrollmentStatusHistoriesV2(user, 0)

		assert.Nil(t, enrollmentStatusHistories)
		assert.Nil(t, domainLocations)
		assert.Equal(t, wantErr.Error(), err.Error())

	})

	t.Run("Unhappy case: throw error when UnmarshalCSV enrollment_status failed with not number", func(t *testing.T) {
		user := &StudentCSV{
			LocationAttr:         field.NewString("partner-id-01"),
			StatusStartDateAttr:  field.NewString("2000/11/11"),
			EnrollmentStatusAttr: field.NewString("ac"),
		}

		wantErr := entity.InvalidFieldError{
			FieldName:  entity.StudentFieldEnrollmentStatus,
			Index:      0,
			EntityName: entity.StudentEntity,
			Reason:     entity.FailedUnmarshal,
		}
		enrollmentStatusHistories, domainLocations, err := toEnrollmentStatusHistoriesV2(user, 0)

		assert.Nil(t, enrollmentStatusHistories)
		assert.Nil(t, domainLocations)
		assert.Equal(t, wantErr.Error(), err.Error())

	})
}

func TestDomainStudentService_validateUserID(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx      context.Context
		student  *StudentCSV
		csvIndex int
	}
	tests := []struct {
		name    string
		service DomainStudent
		args    args
		wantErr error
	}{
		{
			name: "validate new student",
			args: args{
				ctx:      ctx,
				student:  &StudentCSV{},
				csvIndex: 0,
			},
			wantErr: nil,
		},
		{
			name: "validate update student successfully",
			service: &mockDomainStudentService{
				validateUpdateSystemAndExternalUserID: func(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error {
					return nil
				},
			},
			args: args{
				ctx: ctx,
				student: &StudentCSV{
					IDAttr: field.NewString(idutil.ULIDNow()),
				},
				csvIndex: 1,
			},
			wantErr: nil,
		},
		{
			name: "validate update student failed",
			service: &mockDomainStudentService{
				validateUpdateSystemAndExternalUserID: func(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error {
					return errcode.Error{
						Code:      errcode.UpdateFieldFail,
						FieldName: string(entity.UserFieldExternalUserID),
					}
				},
			},
			args: args{
				ctx: ctx,
				student: &StudentCSV{
					IDAttr: field.NewString(idutil.ULIDNow()),
				},
				csvIndex: 2,
			},
			wantErr: errcode.Error{
				Code:      errcode.UpdateFieldFail,
				FieldName: string(entity.UserFieldExternalUserID),
				Index:     2,
			},
		},
		{
			name: "validate update student failed with internal error",
			service: &mockDomainStudentService{
				validateUpdateSystemAndExternalUserID: func(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error {
					return fmt.Errorf("error")
				},
			},
			args: args{
				ctx: ctx,
				student: &StudentCSV{
					IDAttr: field.NewString(idutil.ULIDNow()),
				},
				csvIndex: 3,
			},
			wantErr: errcode.Error{
				Code:  errcode.InternalError,
				Index: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{DomainStudent: tt.service}
			err := d.validateUserID(tt.args.ctx, tt.args.student, tt.args.csvIndex)
			if tt.wantErr != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.wantErr.(errcode.Error)
				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
				assert.Equal(t, e.Index, wantErr.Index)
			}
		})
	}
}

func Test_toBirthDay(t *testing.T) {
	validTime := time.Date(2000, 01, 01, 0, 0, 0, 0, time.UTC)
	testCases := []struct {
		name          string
		birthDayStr   field.String
		csvIndex      int
		expectedDate  field.Date
		expectedError error
	}{
		{
			name:          "Valid birth day",
			birthDayStr:   field.NewString("2000/01/01"),
			csvIndex:      0,
			expectedDate:  field.NewDate(validTime),
			expectedError: nil,
		},
		{
			name:          "Invalid birth day format",
			birthDayStr:   field.NewString("2000/1/1"),
			csvIndex:      1,
			expectedDate:  field.NewNullDate(),
			expectedError: errcode.Error{Code: errcode.InvalidData, FieldName: entity.StudentBirthdayField, Index: 1},
		},
		{
			name:          "Invalid birth day format",
			birthDayStr:   field.NewString("01-01-2000"),
			csvIndex:      1,
			expectedDate:  field.NewNullDate(),
			expectedError: errcode.Error{Code: errcode.InvalidData, FieldName: entity.StudentBirthdayField, Index: 1},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			date, err := toBirthDay(tt.birthDayStr, tt.csvIndex)
			assert.Equal(t, tt.expectedDate, date)
			if tt.expectedError != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.expectedError.(errcode.Error)
				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
				assert.Equal(t, e.Index, wantErr.Index)
			}
		})
	}
}

func Test_toBirthDayV2(t *testing.T) {
	validTime := time.Date(2000, 01, 01, 0, 0, 0, 0, time.UTC)
	expectedDate := field.NewDate(validTime)
	t.Run("happy case: valid birth date", func(t *testing.T) {
		date, err := toBirthDayV2(field.NewString("2000/01/01"), 0)

		assert.Equal(t, expectedDate, date)
		assert.Nil(t, err)
	})

	t.Run("unhappy case: invalid birth day format - 2000/1/1", func(t *testing.T) {
		_, err := toBirthDayV2(field.NewString("2000/1/1"), 0)
		expectErr := entity.InvalidFieldError{
			FieldName:  entity.StudentBirthdayField,
			EntityName: entity.StudentEntity,
			Index:      0,
			Reason:     entity.FailedUnmarshal,
		}
		assert.Equal(t, expectErr.Error(), err.Error())
	})

	t.Run("unhappy case: invalid birth day format - 2000/1/1", func(t *testing.T) {
		_, err := toBirthDayV2(field.NewString("2000/1/1"), 0)
		expectErr := entity.InvalidFieldError{
			FieldName:  entity.StudentBirthdayField,
			EntityName: entity.StudentEntity,
			Index:      0,
			Reason:     entity.FailedUnmarshal,
		}
		assert.Equal(t, expectErr.Error(), err.Error())
	})
}

func Test_toExternalUserID(t *testing.T) {
	testCases := []struct {
		name    string
		student *StudentCSV
		want    field.String
	}{
		{
			name: "only white space",
			student: &StudentCSV{
				ExternalUserIDAttr: field.NewString(" "),
			},
			want: field.NewNullString(),
		},
		{
			name: "white space and text",
			student: &StudentCSV{
				ExternalUserIDAttr: field.NewString(" test "),
			},
			want: field.NewString("test"),
		},
		{
			name: "empty string",
			student: &StudentCSV{
				ExternalUserIDAttr: field.NewString(""),
			},
			want: field.NewNullString(),
		},
		{
			name:    "don't present external user id",
			student: &StudentCSV{},
			want:    field.NewNullString(),
		},
		{
			name: "new line and text",
			student: &StudentCSV{
				ExternalUserIDAttr: field.NewString(`text

				
				`),
			},
			want: field.NewString("text"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			externalUserID := toExternalUserID(tt.student)
			assert.Equal(t, tt.want, externalUserID)

		})
	}
}

func Test_toGender(t *testing.T) {
	testCases := []struct {
		name           string
		gender         field.String
		csvIndex       int
		expectedGender field.String
		expectedError  error
	}{
		{
			name:           "Valid male gender enum",
			gender:         field.NewString("1"),
			csvIndex:       0,
			expectedGender: field.NewString(constant.UserGenderMale),
			expectedError:  nil,
		},
		{
			name:           "Valid empty gender",
			gender:         field.NewNullString(),
			csvIndex:       1,
			expectedGender: field.NewNullString(),
			expectedError:  nil,
		},
		{
			name:           "invalid gender enum",
			gender:         field.NewString("-"),
			csvIndex:       2,
			expectedGender: field.NewNullString(),
			expectedError: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentGenderField,
				Index:     2,
			},
		},
		{
			name:           "out of gender range",
			gender:         field.NewString("3"),
			csvIndex:       2,
			expectedGender: field.NewNullString(),
			expectedError: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentGenderField,
				Index:     2,
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToGender(tt.gender, tt.csvIndex)
			assert.Equal(t, tt.expectedGender, got)
			if tt.expectedError != nil || err != nil {
				e, _ := err.(errcode.Error)
				wantErr, _ := tt.expectedError.(errcode.Error)
				assert.Equal(t, e.Code, wantErr.Code)
				assert.Equal(t, e.FieldName, wantErr.FieldName)
				assert.Equal(t, e.Index, wantErr.Index)
			}
		})
	}
}

func Test_ToGenderV2(t *testing.T) {
	t.Parallel()
	t.Run("happy case: valid male gender enum", func(t *testing.T) {
		got, err := ToGenderV2(field.NewString("1"), 0)
		assert.Equal(t, field.NewString("MALE"), got)
		assert.Nil(t, err)
	})

	t.Run("happy case: valid empty gender", func(t *testing.T) {
		got, err := ToGenderV2(field.NewNullString(), 0)
		assert.Equal(t, field.NewNullString(), got)
		assert.Nil(t, err)
	})

	t.Run("unhappy case: invalid gender enum", func(t *testing.T) {
		got, err := ToGenderV2(field.NewString("-"), 0)
		expectedErr := entity.InvalidFieldError{
			FieldName:  entity.StudentGenderField,
			EntityName: entity.StudentEntity,
			Index:      0,
			Reason:     entity.FailedUnmarshal,
		}
		assert.Equal(t, field.NewNullString(), got)
		assert.Equal(t, expectedErr.Error(), err.Error())
	})

	t.Run("unhappy case: out of gender range", func(t *testing.T) {
		got, err := ToGenderV2(field.NewString("3"), 0)
		expectedErr := entity.InvalidFieldError{
			FieldName:  entity.StudentGenderField,
			EntityName: entity.StudentEntity,
			Index:      0,
			Reason:     entity.NotMatchingEnum,
		}
		assert.Equal(t, field.NewNullString(), got)
		assert.Equal(t, expectedErr.Error(), err.Error())
	})
}

func TestDomainStudentService_fillExistedEmailOfUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type args struct {
		ctx      context.Context
		students []aggregate.DomainStudent
	}
	tests := []struct {
		name    string
		service DomainStudent
		args    args
		setup   func()
		want    []aggregate.DomainStudent
		wantErr error
	}{
		{
			name: "fill old emails to all users have id",
			service: &mockDomainStudentService{
				getEmailWithStudentID: func(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
					return map[string]entity.User{
						"user_id-1": entity_mock.User{
							RandomUser: entity_mock.RandomUser{
								Email:      field.NewString("email-1"),
								LoginEmail: field.NewString("login-email-1"),
							},
						},
						"user_id-2": entity_mock.User{
							RandomUser: entity_mock.RandomUser{
								Email:      field.NewString("email-2"),
								LoginEmail: field.NewString("login-email-2"),
							},
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: &StudentCSV{
							IDAttr:    field.NewString("user_id-1"),
							EmailAttr: field.NewString("edited-email-1"),
						},
					},
					{
						DomainStudent: &StudentCSV{
							IDAttr:    field.NewString("user_id-2"),
							EmailAttr: field.NewString("edited-email-2"),
						},
					},
				},
			},
			want: []aggregate.DomainStudent{
				{
					DomainStudent: &StudentCSV{
						IDAttr:         field.NewString("user_id-1"),
						EmailAttr:      field.NewString("email-1"),
						LoginEmailAttr: field.NewString("login-email-1"),
						UserNameAttr:   field.NewString("email-1"),
					},
				},
				{
					DomainStudent: &StudentCSV{
						IDAttr:         field.NewString("user_id-2"),
						EmailAttr:      field.NewString("email-2"),
						LoginEmailAttr: field.NewString("login-email-2"),
						UserNameAttr:   field.NewString("email-2"),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fill only one user has id",
			service: &mockDomainStudentService{
				getEmailWithStudentID: func(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
					return map[string]entity.User{
						"user_id-1": entity_mock.User{
							RandomUser: entity_mock.RandomUser{
								Email:      field.NewString("email-1"),
								LoginEmail: field.NewString("login-email-1"),
							},
						},
					}, nil
				},
			},
			args: args{
				ctx: ctx,
				students: []aggregate.DomainStudent{
					{
						DomainStudent: &StudentCSV{
							IDAttr:    field.NewNullString(),
							EmailAttr: field.NewString("email-2"),
						},
					},
					{
						DomainStudent: &StudentCSV{
							IDAttr:    field.NewString("user_id-1"),
							EmailAttr: field.NewNullString(),
						},
					},
				},
			},
			want: []aggregate.DomainStudent{
				{
					DomainStudent: &StudentCSV{
						IDAttr:    field.NewNullString(),
						EmailAttr: field.NewString("email-2"),
					},
				},
				{
					DomainStudent: &StudentCSV{
						IDAttr:         field.NewString("user_id-1"),
						EmailAttr:      field.NewString("email-1"),
						LoginEmailAttr: field.NewString("login-email-1"),
						UserNameAttr:   field.NewString("email-1"),
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DomainStudentService{DomainStudent: tt.service}
			got, err := d.fillExistedEmailOfUsers(tt.args.ctx, tt.args.students)
			assert.Equalf(t, tt.want, got, "fillExistedEmailOfUsers(%v, %v)", tt.args.ctx, tt.args.students)
			assert.Equalf(t, tt.wantErr, err, "fillExistedEmailOfUsers(%v, %v)", tt.args.ctx, tt.args.students)
		})
	}
}

func Test_validateCSVEnrollmentStatusFields(t *testing.T) {
	tests := []struct {
		name             string
		location         field.String
		enrollmentStatus field.String
		statusStartDate  field.String
		csvIndex         int
		expectedError    error
	}{
		{
			name:             "valid enrollment status",
			location:         field.NewString("location_1;location_2;location_3"),
			enrollmentStatus: field.NewString("enrolled;;"),
			statusStartDate:  field.NewNullString(),
			csvIndex:         0,
			expectedError:    nil,
		},
		{
			name:             "valid enrollment status",
			location:         field.NewString("location_1;location_2;location_3"),
			enrollmentStatus: field.NewString("enrolled;not enrolled;enrolled"),
			statusStartDate:  field.NewString("2022-01-01;2022-02-01;"),
			csvIndex:         0,
			expectedError:    nil,
		},
		{
			name:             "length of location and status start date are not equal",
			location:         field.NewString("location_1;location_2;location_3"),
			enrollmentStatus: field.NewString("enrolled;;enrolled"),
			statusStartDate:  field.NewString("2022-01-01;2022-03-01"),
			csvIndex:         0,
			expectedError: errcode.Error{
				Code:      errcode.InvalidData,
				Err:       fmt.Errorf("length of location and status start date are not equal"),
				FieldName: entity.StudentFieldEnrollmentStatusStartDate,
				Index:     0,
			},
		},
		{
			name:             "length of location and enrollment status are not equal",
			location:         field.NewString("location_1;location_2;location_3"),
			enrollmentStatus: field.NewString("enrolled"),
			statusStartDate:  field.NewNullString(),
			csvIndex:         1,
			expectedError: errcode.Error{
				Code:      errcode.InvalidData,
				Err:       fmt.Errorf("length of location and enrollment status are not equal"),
				FieldName: entity.StudentFieldEnrollmentStatus,
				Index:     1,
			},
		},
		{
			name:             "empty enrollment status",
			location:         field.NewString("San Francisco"),
			enrollmentStatus: field.NewNullString(),
			statusStartDate:  field.NewString("2022-01-01"),
			csvIndex:         2,
			expectedError:    nil,
		},
		{
			name:             "without enrollment status",
			location:         field.NewNullString(),
			enrollmentStatus: field.NewNullString(),
			statusStartDate:  field.NewNullString(),
			csvIndex:         3,
			expectedError:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCSVEnrollmentStatusFields(test.location, test.enrollmentStatus, test.statusStartDate, test.csvIndex)
			if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.expectedError) {
				t.Errorf("For location=%s, enrollmentStatus=%s, statusStartDate=%s, csvIndex=%d, expected error=%v, but got %v", test.location, test.enrollmentStatus, test.statusStartDate, test.csvIndex, test.expectedError, err)
			}
		})
	}
}
