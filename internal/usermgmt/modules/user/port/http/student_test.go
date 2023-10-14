package http

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"

	"github.com/stretchr/testify/assert"
)

type mockDomainStudent struct {
	getUsersByExternalIDsFn         func(ctx context.Context, externalUserIDs []string) (entity.Users, error)
	getGradesByExternalIDsFn        func(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error)
	getTagsByExternalIDsFn          func(ctx context.Context, externalIDs []string) (entity.DomainTags, error)
	getLocationsByExternalIDsFn     func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error)
	getSchoolsByExternalIDsFn       func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error)
	getSchoolCoursesByExternalIDsFn func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error)
	getPrefecturesByCodesFn         func(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error)
	getEmailWithStudentID           func(ctx context.Context, studentIDs []string) (map[string]entity.User, error)

	upsertMultipleWithErrorCollection func(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error)

	isFeatureIgnoreUpdateEmailEnabled                     func(organization valueobj.HasOrganizationID) bool
	isFeatureUserNameStudentParentEnabled                 func(organization valueobj.HasOrganizationID) bool
	isFeatureIgnoreInvalidRecordsOpenAPIEnabled           func(organization valueobj.HasOrganizationID) bool
	isFeatureAutoDeactivateAndReactivateStudentsV2Enabled func(organization valueobj.HasOrganizationID) bool
	isDisableAutoDeactivateStudents                       func(organization valueobj.HasOrganizationID) bool
	isExperimentalBulkInsertEnrollmentStatusHistories     func(organization valueobj.HasOrganizationID) bool
}

func (m *mockDomainStudent) GetUsersByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
	return m.getUsersByExternalIDsFn(ctx, externalUserIDs)
}
func (m *mockDomainStudent) GetGradesByExternalIDs(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error) {
	return m.getGradesByExternalIDsFn(ctx, externalIDs)
}
func (m *mockDomainStudent) GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
	return m.getTagsByExternalIDsFn(ctx, externalIDs)
}
func (m *mockDomainStudent) GetLocationsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
	return m.getLocationsByExternalIDsFn(ctx, externalIDs)
}
func (m *mockDomainStudent) GetSchoolsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
	return m.getSchoolsByExternalIDsFn(ctx, externalIDs)
}
func (m *mockDomainStudent) GetSchoolCoursesByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
	return m.getSchoolCoursesByExternalIDsFn(ctx, externalIDs)
}
func (m *mockDomainStudent) GetPrefecturesByCodes(ctx context.Context, externalIDs []string) ([]entity.DomainPrefecture, error) {
	return m.getPrefecturesByCodesFn(ctx, externalIDs)
}
func (m *mockDomainStudent) GetEmailWithStudentID(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
	return m.getEmailWithStudentID(ctx, studentIDs)
}
func (m *mockDomainStudent) UpsertMultiple(ctx context.Context, option unleash.DomainStudentFeatureOption, studentsToCreate ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error) {
	return nil, nil
}
func (m *mockDomainStudent) UpsertMultipleWithErrorCollection(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
	return m.upsertMultipleWithErrorCollection(ctx, domainStudents, option)
}
func (m *mockDomainStudent) IsFeatureIgnoreUpdateEmailEnabled(organization valueobj.HasOrganizationID) bool {
	return m.isFeatureIgnoreUpdateEmailEnabled(organization)
}
func (m *mockDomainStudent) IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool {
	return m.isFeatureUserNameStudentParentEnabled(organization)
}
func (m *mockDomainStudent) IsFeatureIgnoreInvalidRecordsOpenAPIEnabled(organization valueobj.HasOrganizationID) bool {
	return m.isFeatureIgnoreInvalidRecordsOpenAPIEnabled(organization)
}
func (m *mockDomainStudent) IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(organization valueobj.HasOrganizationID) bool {
	return m.isFeatureAutoDeactivateAndReactivateStudentsV2Enabled(organization)
}
func (d *mockDomainStudent) IsDisableAutoDeactivateStudents(organization valueobj.HasOrganizationID) bool {
	return d.isDisableAutoDeactivateStudents(organization)
}
func (d *mockDomainStudent) IsExperimentalBulkInsertEnrollmentStatusHistories(organization valueobj.HasOrganizationID) bool {
	return d.isExperimentalBulkInsertEnrollmentStatusHistories(organization)
}
func (d *mockDomainStudent) IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error) {
	return true, nil
}
func TestDomainStudent_toDomainStudentsAgg(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type args struct {
		studentProfiles  []StudentProfile
		isEnableUsername bool
	}
	testCases := []struct {
		name         string
		ctx          context.Context
		args         args
		setupMock    func() mockDomainStudent
		expectedResp []aggregate.DomainStudent
		expectedErr  error
	}{
		{
			name: "happy case",
			ctx:  ctx,
			args: args{
				studentProfiles: []StudentProfile{
					{
						ExternalUserID: field.NewString("external-user-id-1"),
						Grade:          field.NewString("grade-partner-internal-id"),
						FirstName:      field.NewString("first-name"),
						LastName:       field.NewString("last-name"),
						Email:          field.NewString("student-email@manabie.com"),
						UserName:       field.NewString("username1"),
						EnrollmentStatusHistories: []EnrollmentStatusHistoryPayload{
							{
								EnrollmentStatus: field.NewInt16(1),
								Location:         field.NewString("location-partner-internal-id"),
							},
						},
					},
				},
				isEnableUsername: true,
			},
			setupMock: func() mockDomainStudent {
				m := mockDomainStudent{
					getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
						return entity.Users{mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								UserID:         field.NewString("user-id"),
								ExternalUserID: field.NewString("external-user-id-1"),
							},
						},
						}, nil
					},
					getGradesByExternalIDsFn: func(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error) {
						return []entity.DomainGrade{mock_usermgmt.Grade{
							RandomGrade: mock_usermgmt.RandomGrade{
								GradeID:           field.NewString("grade-id"),
								PartnerInternalID: field.NewString("grade-partner-internal-id"),
							},
						}}, nil
					},

					getTagsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
						return entity.DomainTags{entity.EmptyDomainTag{}}, nil
					},
					getLocationsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
						return entity.DomainLocations{mock_usermgmt.Location{
							LocationIDAttr:        field.NewString("location-id"),
							PartnerInternalIDAttr: field.NewString("location-partner-internal-id"),
						}}, nil
					},
					getSchoolsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
						return entity.DomainSchools{entity.DefaultDomainSchool{}}, nil
					},
					getSchoolCoursesByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
						return entity.DomainSchoolCourses{entity.DefaultDomainSchoolCourse{}}, nil
					},
					getPrefecturesByCodesFn: func(ctx context.Context, externalIDs []string) ([]entity.DomainPrefecture, error) {
						return []entity.DomainPrefecture{entity.DefaultDomainPrefecture{}}, nil
					},
				}
				return m
			},
			expectedResp: []aggregate.DomainStudent{
				{
					DomainStudent: DomainStudentImpl{StudentProfile: StudentProfile{
						UserID:     field.NewString("user-id"),
						GradeID:    field.NewString("grade-id"),
						FirstName:  field.NewString("first-name"),
						LastName:   field.NewString("last-name"),
						FullName:   field.NewString("last-name first-name"),
						Email:      field.NewString("student-email@manabie.com"),
						UserName:   field.NewString("username1"),
						LoginEmail: field.NewString("student-email@manabie.com"),
					},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "external_user_id is null",
			ctx:  ctx,
			args: args{
				studentProfiles: []StudentProfile{
					{
						Grade:     field.NewString("grade-partner-internal-id"),
						FirstName: field.NewString("first-name"),
						LastName:  field.NewString("last-name"),
						Email:     field.NewString("student-email@manabie.com"),
						UserName:  field.NewString("username1"),
						EnrollmentStatusHistories: []EnrollmentStatusHistoryPayload{
							{
								EnrollmentStatus: field.NewInt16(1),
								Location:         field.NewString("location-partner-internal-id"),
							},
						},
					},
				},
			},
			setupMock: func() mockDomainStudent {
				m := mockDomainStudent{}
				return m
			},
			expectedErr: errcode.Error{
				FieldName: fmt.Sprintf("students[%d].external_user_id", 0),
				Code:      errcode.MissingMandatory,
			},
		},
		{
			name: "external_user_id is empty string",
			ctx:  ctx,
			args: args{
				studentProfiles: []StudentProfile{
					{
						ExternalUserID: field.NewString(""),
						Grade:          field.NewString("grade-partner-internal-id"),
						FirstName:      field.NewString("first-name"),
						LastName:       field.NewString("last-name"),
						Email:          field.NewString("student-email@manabie.com"),
						UserName:       field.NewString("username1"),
						EnrollmentStatusHistories: []EnrollmentStatusHistoryPayload{
							{
								EnrollmentStatus: field.NewInt16(1),
								Location:         field.NewString("location-partner-internal-id"),
							},
						},
					},
				},
				isEnableUsername: true,
			},
			setupMock: func() mockDomainStudent {
				m := mockDomainStudent{}
				return m
			},
			expectedErr: errcode.Error{
				FieldName: fmt.Sprintf("students[%d].external_user_id", 0),
				Code:      errcode.MissingMandatory,
			},
		},
		{
			name: "return correct students when username is disabled",
			ctx:  ctx,
			args: args{
				studentProfiles: []StudentProfile{
					{
						ExternalUserID: field.NewString("external-user-id-1"),
						Grade:          field.NewString("grade-partner-internal-id"),
						FirstName:      field.NewString("first-name"),
						LastName:       field.NewString("last-name"),
						Email:          field.NewString("student-email@manabie.com"),
						UserName:       field.NewString("username1"),
						EnrollmentStatusHistories: []EnrollmentStatusHistoryPayload{
							{
								EnrollmentStatus: field.NewInt16(1),
								Location:         field.NewString("location-partner-internal-id"),
							},
						},
					},
				},
				isEnableUsername: false,
			},
			setupMock: func() mockDomainStudent {
				m := mockDomainStudent{
					getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
						return entity.Users{mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								UserID:         field.NewString("user-id"),
								ExternalUserID: field.NewString("external-user-id-1"),
							},
						},
						}, nil
					},
					getGradesByExternalIDsFn: func(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error) {
						return []entity.DomainGrade{mock_usermgmt.Grade{
							RandomGrade: mock_usermgmt.RandomGrade{
								GradeID:           field.NewString("grade-id"),
								PartnerInternalID: field.NewString("grade-partner-internal-id"),
							},
						}}, nil
					},

					getTagsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
						return entity.DomainTags{entity.EmptyDomainTag{}}, nil
					},
					getLocationsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
						return entity.DomainLocations{mock_usermgmt.Location{
							LocationIDAttr:        field.NewString("location-id"),
							PartnerInternalIDAttr: field.NewString("location-partner-internal-id"),
						}}, nil
					},
					getSchoolsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
						return entity.DomainSchools{entity.DefaultDomainSchool{}}, nil
					},
					getSchoolCoursesByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
						return entity.DomainSchoolCourses{entity.DefaultDomainSchoolCourse{}}, nil
					},
					getPrefecturesByCodesFn: func(ctx context.Context, externalIDs []string) ([]entity.DomainPrefecture, error) {
						return []entity.DomainPrefecture{entity.DefaultDomainPrefecture{}}, nil
					},
				}
				return m
			},
			expectedResp: []aggregate.DomainStudent{
				{
					DomainStudent: DomainStudentImpl{StudentProfile: StudentProfile{
						UserID:     field.NewString("user-id"),
						GradeID:    field.NewString("grade-id"),
						FirstName:  field.NewString("first-name"),
						LastName:   field.NewString("last-name"),
						FullName:   field.NewString("last-name first-name"),
						Email:      field.NewString("student-email@manabie.com"),
						UserName:   field.NewString("student-email@manabie.com"),
						LoginEmail: field.NewString("student-email@manabie.com"),
					},
					},
				},
			},
			expectedErr: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockDomainStudent := testCase.setupMock()
			sd := DomainStudentService{
				DomainStudent: &mockDomainStudent,
			}
			students, err := sd.ToDomainStudentsAgg(testCase.ctx, testCase.args.studentProfiles, testCase.args.isEnableUsername)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				for idx, student := range students {
					req := testCase.expectedResp[idx]
					assert.Equal(t, req.UserID().String(), student.UserID().String())
					assert.Equal(t, req.FirstName().String(), student.FirstName().String())
					assert.Equal(t, req.LastName().String(), student.LastName().String())
					assert.Equal(t, req.FullName().String(), student.FullName().String())
					assert.Equal(t, req.Email().String(), student.Email().String())
					assert.Equal(t, req.GradeID().String(), student.GradeID().String())
					assert.Equal(t, req.UserName().String(), student.UserName().String())
					assert.Equal(t, req.LoginEmail().String(), student.LoginEmail().String())
				}
			}

		})
	}
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
		service *mockDomainStudent
		args    args
		want    []aggregate.DomainStudent
		wantErr error
	}{
		{
			name: "fill old emails to all users have id",
			service: &mockDomainStudent{
				getEmailWithStudentID: func(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
					return map[string]entity.User{
						"user_id-1": mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								Email:      field.NewString("email-1"),
								LoginEmail: field.NewString("login-email-1"),
							},
						},
						"user_id-2": mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
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
						DomainStudent: DomainStudentImpl{
							NullDomainStudent: entity.NullDomainStudent{},
							StudentProfile: StudentProfile{
								UserID: field.NewString("user_id-1"),
								Email:  field.NewString("edited-email-1"),
							},
						},
					},
					{
						DomainStudent: DomainStudentImpl{
							NullDomainStudent: entity.NullDomainStudent{},
							StudentProfile: StudentProfile{
								UserID: field.NewString("user_id-2"),
								Email:  field.NewString("edited-email-2"),
							},
						},
					},
				},
			},
			want: []aggregate.DomainStudent{
				{
					DomainStudent: DomainStudentImpl{
						NullDomainStudent: entity.NullDomainStudent{},
						StudentProfile: StudentProfile{
							UserID:     field.NewString("user_id-1"),
							Email:      field.NewString("email-1"),
							LoginEmail: field.NewString("login-email-1"),
							UserName:   field.NewString("email-1"),
						},
					},
				},
				{
					DomainStudent: DomainStudentImpl{
						NullDomainStudent: entity.NullDomainStudent{},
						StudentProfile: StudentProfile{
							UserID:     field.NewString("user_id-2"),
							Email:      field.NewString("email-2"),
							LoginEmail: field.NewString("login-email-2"),
							UserName:   field.NewString("email-2"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "fill only one user has id",
			service: &mockDomainStudent{
				getEmailWithStudentID: func(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
					return map[string]entity.User{
						"user_id-1": mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
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
						DomainStudent: DomainStudentImpl{
							NullDomainStudent: entity.NullDomainStudent{},
							StudentProfile: StudentProfile{
								UserID: field.NewNullString(),
								Email:  field.NewString("email-2"),
							},
						},
					},
					{
						DomainStudent: DomainStudentImpl{
							NullDomainStudent: entity.NullDomainStudent{},
							StudentProfile: StudentProfile{
								UserID: field.NewString("user_id-1"),
								Email:  field.NewNullString(),
							},
						},
					},
				},
			},
			want: []aggregate.DomainStudent{
				{
					DomainStudent: DomainStudentImpl{
						NullDomainStudent: entity.NullDomainStudent{},
						StudentProfile: StudentProfile{
							UserID: field.NewNullString(),
							Email:  field.NewString("email-2"),
						},
					},
				},
				{
					DomainStudent: DomainStudentImpl{
						NullDomainStudent: entity.NullDomainStudent{},
						StudentProfile: StudentProfile{
							UserID:     field.NewString("user_id-1"),
							Email:      field.NewString("email-1"),
							LoginEmail: field.NewString("login-email-1"),
							UserName:   field.NewString("email-1"),
						},
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := DomainStudentService{DomainStudent: tt.service}
			got, err := port.fillExistedEmailOfUsers(tt.args.ctx, tt.args.students)
			assert.Equalf(t, tt.want, got, "GetEmailWithStudentID(%v, %v)", tt.args.ctx, tt.args.students)
			assert.Equalf(t, tt.wantErr, err, "GetEmailWithStudentID(%v, %v)", tt.args.ctx, tt.args.students)
		})
	}
}

func TestDomainStudentService_toDomainSchoolHistories(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type args struct {
		ctx             context.Context
		schoolHistories []byte
	}
	mockSchoolInfo_1 := mock.School{
		RandomSchool: mock.RandomSchool{
			SchoolID:          field.NewString("school-id-1"),
			PartnerInternalID: field.NewString("school-partner-id-1"),
		},
	}
	mockSchoolInfo_2 := mock.School{
		RandomSchool: mock.RandomSchool{
			SchoolID:          field.NewString("school-id-2"),
			PartnerInternalID: field.NewString("school-partner-id-2"),
		},
	}
	mockSchoolCourse_1 := mock.SchoolCourse{
		RandomSchoolCourse: mock.RandomSchoolCourse{
			SchoolCourseID:    field.NewString("school-course-id-1"),
			PartnerInternalID: field.NewString("school-course-partner-id-1"),
		},
	}
	mockSchoolCourse_2 := mock.SchoolCourse{
		RandomSchoolCourse: mock.RandomSchoolCourse{
			SchoolCourseID:    field.NewString("school-course-id-2"),
			PartnerInternalID: field.NewString("school-course-partner-id-2"),
		},
	}
	mockStartDate, _ := time.Parse(constant.DateLayout, "2020/10/10")
	mockEndDate, _ := time.Parse(constant.DateLayout, "2020/10/11")

	tests := []struct {
		name    string
		service *mockDomainStudent
		args    args
		want    entity.DomainSchoolHistories
		wantErr error
	}{
		{
			name: "return correct school histories",
			service: &mockDomainStudent{
				getSchoolsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil
				},
				getSchoolCoursesByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1, mockSchoolCourse_2}, nil
				},
			},
			args: args{
				ctx: ctx,
				schoolHistories: []byte(`{
					"students": [
						{
							"school_histories": [
								{
									"school": "school-partner-id-1",
									"school_course": "school-course-partner-id-1",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								},
								{
									"school": "school-partner-id-2",
									"school_course": "school-course-partner-id-2",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								}
							]
						}
					]
				}`),
			},
			want: entity.DomainSchoolHistories{
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-1"),
						SchoolCourseID: field.NewString("school-course-id-1"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-2"),
						SchoolCourseID: field.NewString("school-course-id-2"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "return correct school histories when school is empty",
			service: &mockDomainStudent{
				getSchoolsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
				getSchoolCoursesByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1, mockSchoolCourse_2}, nil
				},
			},
			args: args{
				ctx: ctx,
				schoolHistories: []byte(`{
					"students": [
						{
							"school_histories": [
								{
									"school": "school-partner-id-1",
									"school_course": "school-course-partner-id-1",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								},
								{
									"school_course": "school-course-partner-id-2",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								}
							]
						}
					]
				}`),
			},
			want: entity.DomainSchoolHistories{
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-1"),
						SchoolCourseID: field.NewString("school-course-id-1"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewNullString(),
						SchoolCourseID: field.NewString("school-course-id-2"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "return correct school histories when school does not exist in DB",
			service: &mockDomainStudent{
				getSchoolsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
				getSchoolCoursesByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1, mockSchoolCourse_2}, nil
				},
			},
			args: args{
				ctx: ctx,
				schoolHistories: []byte(`{
					"students": [
						{
							"school_histories": [
								{
									"school": "school-partner-id-1",
									"school_course": "school-course-partner-id-1",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								},
								{
									"school": "school-partner-id-2",
									"school_course": "school-course-partner-id-2",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								}
							]
						}
					]
				}`),
			},

			want: entity.DomainSchoolHistories{
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-1"),
						SchoolCourseID: field.NewString("school-course-id-1"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewNullString(),
						SchoolCourseID: field.NewString("school-course-id-2"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "return correct school histories when school and school course are duplicated",
			service: &mockDomainStudent{
				getSchoolsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1}, nil
				},
				getSchoolCoursesByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				schoolHistories: []byte(`{
					"students": [
						{
							"school_histories": [
								{
									"school": "school-partner-id-1",
									"school_course": "school-course-partner-id-1",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								},
								{
									"school": "school-partner-id-1",
									"school_course": "school-course-partner-id-1",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								}
							]
						}
					]
				}`),
			},
			want: entity.DomainSchoolHistories{
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-1"),
						SchoolCourseID: field.NewString("school-course-id-1"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-1"),
						SchoolCourseID: field.NewString("school-course-id-1"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "return correct school histories when course is empty",
			service: &mockDomainStudent{
				getSchoolsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil
				},
				getSchoolCoursesByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				schoolHistories: []byte(`{
					"students": [
						{
							"school_histories": [
								{
									"school": "school-partner-id-1",
									"school_course": "school-course-partner-id-1",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								},
								{
									"school": "school-partner-id-2",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								}
							]
						}
					]
				}`),
			},
			want: entity.DomainSchoolHistories{
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-1"),
						SchoolCourseID: field.NewString("school-course-id-1"),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-2"),
						SchoolCourseID: field.NewNullString(),
						StartDate:      field.NewTime(mockStartDate),
						EndDate:        field.NewTime(mockEndDate),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "return error when school course does not exist in DB",
			service: &mockDomainStudent{
				getSchoolsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
					return entity.DomainSchools{mockSchoolInfo_1, mockSchoolInfo_2}, nil
				},
				getSchoolCoursesByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
					return entity.DomainSchoolCourses{mockSchoolCourse_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				schoolHistories: []byte(`{
					"students": [
						{
							"school_histories": [
								{
									"school": "school-partner-id-1",
									"school_course": "school-course-partner-id-1",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								},
								{
									"school": "school-partner-id-2",
									"school_course": "school-course-partner-id-2",
									"start_date": "2020/10/10",
									"end_date": "2020/10/11"
								}
							]
						}
					]
				}`),
			},
			want: nil,
			wantErr: errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: "school_histories.school_course",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := DomainStudentService{DomainStudent: tt.service}
			httpReq := &http.Request{
				Body: ioutil.NopCloser(bytes.NewReader(tt.args.schoolHistories)),
			}
			var req UpsertStudentsRequest
			ParseJSONPayload(httpReq, &req)
			schoolHistories, err := port.toDomainSchoolHistories(tt.args.ctx, req.Students[0].SchoolHistories)
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

func TestDomainStudentService_toDomainEnrollmentStatusHistories(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	type args struct {
		ctx                       context.Context
		enrollmentStatusHistories []byte
	}
	mockLocation_1 := mock.Location{
		LocationIDAttr:        field.NewString("location-id-1"),
		PartnerInternalIDAttr: field.NewString("location-partner-id-1"),
	}
	mockLocation_2 := mock.Location{
		LocationIDAttr:        field.NewString("location-id-2"),
		PartnerInternalIDAttr: field.NewString("location-partner-id-2"),
	}
	mockStartDate, _ := time.Parse(constant.DateLayout, "2020/10/10")

	tests := []struct {
		name                string
		service             *mockDomainStudent
		args                args
		want                entity.DomainEnrollmentStatusHistories
		wantErr             error
		expectedAccessPaths entity.DomainUserAccessPaths
	}{
		{
			name: "return correct enrollment status histories",
			service: &mockDomainStudent{
				getLocationsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return entity.DomainLocations{mockLocation_1, mockLocation_2}, nil
				},
			},
			args: args{
				ctx: ctx,
				enrollmentStatusHistories: []byte(`{
					"students": [
						{
							"enrollment_status_histories": [
								{
									"enrollment_status": 1,
									"location": "location-partner-id-1",
									"start_date": "2020/10/10"
								},
								{
									"enrollment_status": 2,
									"location": "location-partner-id-2",
									"start_date": "2020/10/10"
								}
							]
						}
					]
				}`),
			},
			want: entity.DomainEnrollmentStatusHistories{
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
						LocationID:       field.NewString("location-id-1"),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
						LocationID:       field.NewString("location-id-2"),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
			},
			wantErr: nil,
			expectedAccessPaths: entity.DomainUserAccessPaths{
				mock.UserAccessPath{
					RandomUserAccessPath: mock.RandomUserAccessPath{
						LocationID: field.NewString("location-id-1"),
					},
				},
				mock.UserAccessPath{
					RandomUserAccessPath: mock.RandomUserAccessPath{
						LocationID: field.NewString("location-id-2"),
					},
				},
			},
		},
		{
			name: "return correct enrollment status histories when location and status are empty",
			service: &mockDomainStudent{
				getLocationsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return entity.DomainLocations{mockLocation_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				enrollmentStatusHistories: []byte(`{
					"students": [
						{
							"enrollment_status_histories": [
								{
									"enrollment_status": 1,
									"location": "location-partner-id-1",
									"start_date": "2020/10/10"
								},
								{
									"start_date": "2020/10/10"
								}
							]
						}
					]
				}`),
			},
			want: entity.DomainEnrollmentStatusHistories{
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
						LocationID:       field.NewString("location-id-1"),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewNullString(),
						LocationID:       field.NewNullString(),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
			},
			wantErr: nil,
			expectedAccessPaths: entity.DomainUserAccessPaths{
				mock.UserAccessPath{
					RandomUserAccessPath: mock.RandomUserAccessPath{
						LocationID: field.NewString("location-id-1"),
					},
				},
			},
		},
		{
			name: "return correct enrollment status histories when location does not exist in DB",
			service: &mockDomainStudent{
				getLocationsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return entity.DomainLocations{mockLocation_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				enrollmentStatusHistories: []byte(`{
					"students": [
						{
							"enrollment_status_histories": [
								{
									"enrollment_status": 1,
									"location": "location-partner-id-1",
									"start_date": "2020/10/10"
								},
								{
									"enrollment_status": 2,
									"location": "location-partner-id-2",
									"start_date": "2020/10/10"
								}
							]
						}
					]
				}`),
			},
			want: entity.DomainEnrollmentStatusHistories{
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
						LocationID:       field.NewString("location-id-1"),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
						LocationID:       field.NewNullString(),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
			},
			wantErr: nil,
			expectedAccessPaths: entity.DomainUserAccessPaths{
				mock.UserAccessPath{
					RandomUserAccessPath: mock.RandomUserAccessPath{
						LocationID: field.NewString("location-id-1"),
					},
				},
			},
		},
		{
			name: "return correct enrollment status histories when locations are duplicated ",
			service: &mockDomainStudent{
				getLocationsByExternalIDsFn: func(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
					return entity.DomainLocations{mockLocation_1}, nil
				},
			},
			args: args{
				ctx: ctx,
				enrollmentStatusHistories: []byte(`{
					"students": [
						{
							"enrollment_status_histories": [
								{
									"enrollment_status": 1,
									"location": "location-partner-id-1",
									"start_date": "2020/10/10"
								},
								{
									"enrollment_status": 1,
									"location": "location-partner-id-1",
									"start_date": "2020/10/10"
								}
							]
						}
					]
				}`),
			},
			want: entity.DomainEnrollmentStatusHistories{
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
						LocationID:       field.NewString("location-id-1"),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
						LocationID:       field.NewString("location-id-1"),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
			},
			wantErr: nil,
			expectedAccessPaths: entity.DomainUserAccessPaths{
				mock.UserAccessPath{
					RandomUserAccessPath: mock.RandomUserAccessPath{
						LocationID: field.NewString("location-id-1"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := DomainStudentService{DomainStudent: tt.service}
			httpReq := &http.Request{
				Body: ioutil.NopCloser(bytes.NewReader(tt.args.enrollmentStatusHistories)),
			}
			var req UpsertStudentsRequest
			ParseJSONPayload(httpReq, &req)
			enrollmentStatusHistories, domainUserAccessPaths, err := port.toDomainEnrollmentStatusHistories(tt.args.ctx, req.Students[0].EnrollmentStatusHistories, entity.DomainUserAccessPaths{})
			if tt.want != nil {
				if len(enrollmentStatusHistories) != len(tt.want) {
					panic(fmt.Errorf("expect len: %v - actual len: %v", len(tt.want), len(enrollmentStatusHistories)))
				}
				for i, status := range enrollmentStatusHistories {
					assert.Equal(t, tt.want[i].EnrollmentStatus(), status.EnrollmentStatus())
					assert.Equal(t, tt.want[i].LocationID(), status.LocationID())
					assert.Equal(t, tt.want[i].StartDate(), status.StartDate())
				}

				if len(domainUserAccessPaths) != len(tt.expectedAccessPaths) {
					panic(fmt.Errorf("expect len: %v - actual len: %v", len(tt.expectedAccessPaths), len(domainUserAccessPaths)))
				}

				for i, accessPaths := range domainUserAccessPaths {
					assert.Equal(t, tt.expectedAccessPaths[i].LocationID(), accessPaths.LocationID())
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

func TestDomainStudentService_toDomainEnrollmentStatusHistoriesV2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	t.Parallel()

	type args struct {
		ctx                       context.Context
		enrollmentStatusHistories []EnrollmentStatusHistoryPayload
	}
	location1 := mock.Location{
		LocationIDAttr:        field.NewString("location-id-1"),
		PartnerInternalIDAttr: field.NewString("location-partner-id-1"),
	}
	location2 := mock.Location{
		LocationIDAttr:        field.NewString("location-id-2"),
		PartnerInternalIDAttr: field.NewString("location-partner-id-2"),
	}
	mockStartDate, _ := time.Parse(constant.DateLayout, "2020/10/10")

	tests := []struct {
		name                           string
		args                           args
		expectedDomainEnrollmentStatus entity.DomainEnrollmentStatusHistories
		expectedDomainLocations        entity.DomainLocations
	}{
		{
			name: "happy case: return 1 domain enrollment status and 1 domain user access path ",
			args: args{
				ctx: ctx,
				enrollmentStatusHistories: []EnrollmentStatusHistoryPayload{
					{
						EnrollmentStatus: field.NewInt16(1),
						Location:         location1.PartnerInternalIDAttr,
						StartDate:        field.NewDate(mockStartDate),
					},
				},
			},
			expectedDomainEnrollmentStatus: entity.DomainEnrollmentStatusHistories{
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
			},
			expectedDomainLocations: entity.DomainLocations{
				mock.Location{
					PartnerInternalIDAttr: location1.PartnerInternalIDAttr,
				},
			},
		},
		{
			name: "happy case: return multiple domain enrollment status and 1 domain user access path ",
			args: args{
				ctx: ctx,
				enrollmentStatusHistories: []EnrollmentStatusHistoryPayload{
					{
						EnrollmentStatus: field.NewInt16(1),
						Location:         location1.PartnerInternalIDAttr,
						StartDate:        field.NewDate(mockStartDate),
					},
					{
						EnrollmentStatus: field.NewInt16(2),
						Location:         location2.PartnerInternalIDAttr,
						StartDate:        field.NewDate(mockStartDate),
					},
				},
			},
			expectedDomainEnrollmentStatus: entity.DomainEnrollmentStatusHistories{
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusPotential),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
				mock.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(constant.StudentEnrollmentStatusEnrolled),
						StartDate:        field.NewTime(mockStartDate),
					},
				},
			},
			expectedDomainLocations: entity.DomainLocations{
				mock.Location{
					PartnerInternalIDAttr: location1.PartnerInternalIDAttr,
				},
				mock.Location{
					PartnerInternalIDAttr: location2.PartnerInternalIDAttr,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			domainEnrollmentStatusHistories, domainLocations := toEnrollmentStatusHistoriesV2(tt.args.enrollmentStatusHistories)
			if len(domainEnrollmentStatusHistories) != len(tt.expectedDomainEnrollmentStatus) {
				panic(fmt.Errorf("expect len: %v - actual len: %v", len(tt.expectedDomainEnrollmentStatus), len(domainEnrollmentStatusHistories)))
			}
			for i, status := range domainEnrollmentStatusHistories {
				assert.Equal(t, tt.expectedDomainEnrollmentStatus[i].EnrollmentStatus(), status.EnrollmentStatus())
				assert.Equal(t, tt.expectedDomainEnrollmentStatus[i].StartDate(), status.StartDate())
			}

			if len(domainLocations) != len(tt.expectedDomainLocations) {
				panic(fmt.Errorf("expect len: %v - actual len: %v", len(tt.expectedDomainLocations), len(domainLocations)))
			}

			for i, location := range domainLocations {
				assert.Equal(t, tt.expectedDomainLocations[i].PartnerInternalID(), location.PartnerInternalID())
			}
		})
	}
}

func TestDomainStudentService_toSchoolHistoryV2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Parallel()
	type args struct {
		ctx             context.Context
		schoolHistories []SchoolHistoryPayload
	}
	school1 := mock.School{
		RandomSchool: mock.RandomSchool{
			PartnerInternalID: field.NewString("school-partner-1"),
		},
	}
	school2 := mock.School{
		RandomSchool: mock.RandomSchool{
			PartnerInternalID: field.NewString("school-partner-2"),
		},
	}
	schoolCourse1 := mock.SchoolCourse{
		RandomSchoolCourse: mock.RandomSchoolCourse{
			PartnerInternalID: field.NewString("school-course-partner-1"),
		},
	}
	schoolCourse2 := mock.SchoolCourse{
		RandomSchoolCourse: mock.RandomSchoolCourse{
			PartnerInternalID: field.NewString("school-course-partner-2"),
		},
	}
	mockStartDate, _ := time.Parse(constant.DateLayout, "2020/10/10")
	mockEndDate, _ := time.Parse(constant.DateLayout, "2020/11/11")

	tests := []struct {
		name                          string
		args                          args
		expectedDomainSchoolHistories entity.DomainSchoolHistories
		expectedDomainSchoolsInfo     entity.DomainSchools
		expectedDomainSchoolCourses   entity.DomainSchoolCourses
	}{
		{
			name: "happy case: return 1 school history",
			args: args{
				ctx: ctx,
				schoolHistories: []SchoolHistoryPayload{
					{
						School:       school1.PartnerInternalID(),
						SchoolCourse: schoolCourse1.PartnerInternalID(),
						StartDate:    field.NewDate(mockStartDate),
						EndDate:      field.NewDate(mockEndDate),
					},
				},
			},
			expectedDomainSchoolHistories: entity.DomainSchoolHistories{
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						StartDate: field.NewTime(mockStartDate),
						EndDate:   field.NewTime(mockEndDate),
					},
				},
			},
			expectedDomainSchoolsInfo: entity.DomainSchools{
				mock.School{
					RandomSchool: mock.RandomSchool{
						PartnerInternalID: school1.PartnerInternalID(),
					},
				},
			},
			expectedDomainSchoolCourses: entity.DomainSchoolCourses{
				mock.SchoolCourse{
					RandomSchoolCourse: mock.RandomSchoolCourse{
						PartnerInternalID: schoolCourse1.PartnerInternalID(),
					},
				},
			},
		},
		{
			name: "happy case: return multiple school history",
			args: args{
				ctx: ctx,
				schoolHistories: []SchoolHistoryPayload{
					{
						School:       school1.PartnerInternalID(),
						SchoolCourse: schoolCourse1.PartnerInternalID(),
						StartDate:    field.NewDate(mockStartDate),
						EndDate:      field.NewDate(mockEndDate),
					},
					{
						School:       school2.PartnerInternalID(),
						SchoolCourse: schoolCourse2.PartnerInternalID(),
						StartDate:    field.NewDate(mockStartDate),
						EndDate:      field.NewDate(mockEndDate),
					},
				},
			},
			expectedDomainSchoolHistories: entity.DomainSchoolHistories{
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						StartDate: field.NewTime(mockStartDate),
						EndDate:   field.NewTime(mockEndDate),
					},
				},
				mock.SchoolHistory{
					RandomSchoolHistory: mock.RandomSchoolHistory{
						StartDate: field.NewTime(mockStartDate),
						EndDate:   field.NewTime(mockEndDate),
					},
				},
			},
			expectedDomainSchoolsInfo: entity.DomainSchools{
				mock.School{
					RandomSchool: mock.RandomSchool{
						PartnerInternalID: school1.PartnerInternalID(),
					},
				},
				mock.School{
					RandomSchool: mock.RandomSchool{
						PartnerInternalID: school2.PartnerInternalID(),
					},
				},
			},
			expectedDomainSchoolCourses: entity.DomainSchoolCourses{
				mock.SchoolCourse{
					RandomSchoolCourse: mock.RandomSchoolCourse{
						PartnerInternalID: schoolCourse1.PartnerInternalID(),
					},
				},
				mock.SchoolCourse{
					RandomSchoolCourse: mock.RandomSchoolCourse{
						PartnerInternalID: schoolCourse2.PartnerInternalID(),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			domainSchoolHistories, domainSchoolInfos, domainSchoolCourses := toSchoolHistoryV2(tt.args.schoolHistories)
			if len(domainSchoolHistories) != len(tt.expectedDomainSchoolHistories) {
				panic(fmt.Errorf("expect len: %v - actual len: %v", len(tt.expectedDomainSchoolHistories), len(domainSchoolHistories)))
			}
			for i, schoolHistory := range domainSchoolHistories {
				assert.Equal(t, tt.expectedDomainSchoolHistories[i].StartDate(), schoolHistory.StartDate())
				assert.Equal(t, tt.expectedDomainSchoolHistories[i].EndDate(), schoolHistory.EndDate())
			}

			if len(domainSchoolInfos) != len(tt.expectedDomainSchoolsInfo) {
				panic(fmt.Errorf("expect len: %v - actual len: %v", len(tt.expectedDomainSchoolsInfo), len(domainSchoolInfos)))
			}
			for i, location := range domainSchoolInfos {
				assert.Equal(t, tt.expectedDomainSchoolsInfo[i].PartnerInternalID(), location.PartnerInternalID())
			}

			if len(domainSchoolCourses) != len(tt.expectedDomainSchoolCourses) {
				panic(fmt.Errorf("expect len: %v - actual len: %v", len(tt.expectedDomainSchoolCourses), len(domainSchoolCourses)))
			}
			for i, location := range domainSchoolCourses {
				assert.Equal(t, tt.expectedDomainSchoolCourses[i].PartnerInternalID(), location.PartnerInternalID())
			}
		})
	}
}

func TestDomainStudentService_upsertStudentWithErrorCollections(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Parallel()
	type args struct {
		ctx      context.Context
		students []StudentProfile
	}
	startDate, _ := time.Parse(constant.DateLayout, "2020/10/10")
	endDate, _ := time.Parse(constant.DateLayout, "2020/11/11")
	studentOnlyMandatory := StudentProfile{
		ExternalUserID: field.NewString("partner-student-1"),
		LastName:       field.NewString("last name"),
		FirstName:      field.NewString("first name"),
		Email:          field.NewString("example@manabie.com"),
		GradeID:        field.NewString("grade-1"),
		EnrollmentStatusHistories: []EnrollmentStatusHistoryPayload{
			{
				EnrollmentStatus: field.NewInt16(1),
				Location:         field.NewString("location-1"),
				StartDate:        field.NewDate(startDate),
				EndDate:          field.NewDate(endDate),
			},
		},
	}

	studentMissingName := StudentProfile{
		ExternalUserID: field.NewString("partner-student-2"),
		LastName:       field.NewString("last name"),
		FirstName:      field.NewString("first name"),
		Email:          field.NewString("example2@manabie.com"),
		GradeID:        field.NewString("grade-2"),
		EnrollmentStatusHistories: []EnrollmentStatusHistoryPayload{
			{
				EnrollmentStatus: field.NewInt16(2),
				Location:         field.NewString("location-2"),
				StartDate:        field.NewDate(startDate),
				EndDate:          field.NewDate(endDate),
			},
		},
	}

	aggregateStudent := aggregate.DomainStudent{
		DomainStudent: &mock_usermgmt.Student{
			RandomStudent: mock_usermgmt.RandomStudent{
				UserID:         field.NewString("student-1"),
				GradeID:        field.NewString("grade-1"),
				Email:          field.NewString("example@manabie.com"),
				LastName:       field.NewString("last name"),
				FirstName:      field.NewString("first name"),
				ExternalUserID: field.NewString("partner-student-1"),
				UserName:       field.NewString("username01"),
			},
		},
	}
	tests := []struct {
		name                   string
		domainStudent          mockDomainStudent
		args                   args
		expectedError          []error
		expectedDomainStudents aggregate.DomainStudents
	}{
		{
			name: "happy case: create 1 student when disable toggle ignore update mail",
			args: args{
				students: []StudentProfile{
					studentOnlyMandatory,
				},
			},
			domainStudent: mockDomainStudent{
				getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
					return entity.Users{entity.EmptyUser{}}, nil
				},
				isFeatureIgnoreUpdateEmailEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				isFeatureUserNameStudentParentEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				upsertMultipleWithErrorCollection: func(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
					return aggregate.DomainStudents{aggregateStudent}, nil
				},
			},
			expectedError:          nil,
			expectedDomainStudents: aggregate.DomainStudents{aggregateStudent},
		},
		{
			name: "happy case: create 1 student when enable toggle ignore update mail",
			args: args{
				students: []StudentProfile{
					studentOnlyMandatory,
				},
			},
			domainStudent: mockDomainStudent{
				getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
					return entity.Users{entity.EmptyUser{}}, nil
				},
				isFeatureIgnoreUpdateEmailEnabled: func(organization valueobj.HasOrganizationID) bool {
					return true
				},
				getEmailWithStudentID: func(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
					result := make(map[string]entity.User)
					return result, nil
				},
				isFeatureUserNameStudentParentEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				upsertMultipleWithErrorCollection: func(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
					return aggregate.DomainStudents{aggregateStudent}, nil
				},
			},
			expectedError:          nil,
			expectedDomainStudents: aggregate.DomainStudents{aggregateStudent},
		},
		{
			name: "happy case: create 1 student when enable toggle ignore update mail and enable toggle username",
			args: args{
				students: []StudentProfile{
					studentOnlyMandatory,
				},
			},
			domainStudent: mockDomainStudent{
				getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
					return entity.Users{entity.EmptyUser{}}, nil
				},
				isFeatureIgnoreUpdateEmailEnabled: func(organization valueobj.HasOrganizationID) bool {
					return true
				},
				getEmailWithStudentID: func(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
					result := make(map[string]entity.User)
					return result, nil
				},
				isFeatureUserNameStudentParentEnabled: func(organization valueobj.HasOrganizationID) bool {
					return true
				},
				upsertMultipleWithErrorCollection: func(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
					return aggregate.DomainStudents{
						aggregate.DomainStudent{
							DomainStudent: &mock_usermgmt.Student{
								RandomStudent: mock_usermgmt.RandomStudent{
									UserID:         field.NewString("student-1"),
									GradeID:        field.NewString("grade-1"),
									Email:          field.NewString("example@manabie.com"),
									LastName:       field.NewString("last name"),
									FirstName:      field.NewString("first name"),
									ExternalUserID: field.NewString("partner-student-1"),
									UserName:       field.NewString("example@manabie.com"),
								},
							},
						},
					}, nil
				},
			},
			expectedError: nil,
			expectedDomainStudents: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:         field.NewString("student-1"),
							GradeID:        field.NewString("grade-1"),
							Email:          field.NewString("example@manabie.com"),
							LastName:       field.NewString("last name"),
							FirstName:      field.NewString("first name"),
							ExternalUserID: field.NewString("partner-student-1"),
							UserName:       field.NewString("example@manabie.com"),
						},
					},
				},
			},
		},
		{
			name: "happy case: update 1 student when disable toggle ignore update mail",
			args: args{
				students: []StudentProfile{
					studentOnlyMandatory,
				},
			},
			domainStudent: mockDomainStudent{
				getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
					return entity.Users{mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							UserID:         field.NewString("student-1"),
							Email:          field.NewString("example@manabie.com"),
							ExternalUserID: field.NewString("partner-student-1"),
						},
					},
					}, nil
				},
				isFeatureIgnoreUpdateEmailEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				isFeatureUserNameStudentParentEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				upsertMultipleWithErrorCollection: func(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
					return aggregate.DomainStudents{aggregateStudent}, nil
				},
			},
			expectedError:          nil,
			expectedDomainStudents: aggregate.DomainStudents{aggregateStudent},
		},
		{
			name: "unhappy case: create 1 student error",
			args: args{
				students: []StudentProfile{
					studentOnlyMandatory,
				},
			},
			domainStudent: mockDomainStudent{
				getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
					return entity.Users{entity.EmptyUser{}}, nil
				},
				isFeatureIgnoreUpdateEmailEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				isFeatureUserNameStudentParentEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				upsertMultipleWithErrorCollection: func(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
					return aggregate.DomainStudents{}, []error{
						entity.ExistingDataError{
							FieldName:  string(entity.UserFieldExternalUserID),
							EntityName: entity.StudentEntity,
							Index:      0,
						},
					}
				},
			},
			expectedError: []error{entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.StudentEntity,
				Index:      0,
			}},
			expectedDomainStudents: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: create 1 student and update 1 student error",
			args: args{
				students: []StudentProfile{
					studentOnlyMandatory,
					studentMissingName,
				},
			},
			domainStudent: mockDomainStudent{
				getUsersByExternalIDsFn: func(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
					return entity.Users{entity.EmptyUser{}}, nil
				},
				isFeatureIgnoreUpdateEmailEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				isFeatureUserNameStudentParentEnabled: func(organization valueobj.HasOrganizationID) bool {
					return false
				},
				upsertMultipleWithErrorCollection: func(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
					return aggregate.DomainStudents{}, []error{
						entity.ExistingDataError{
							FieldName:  string(entity.UserFieldExternalUserID),
							EntityName: entity.StudentEntity,
							Index:      0,
						},
						entity.MissingMandatoryFieldError{
							FieldName:  string(entity.UserFieldFirstName),
							EntityName: entity.StudentEntity,
							Index:      1,
						},
					}
				},
			},
			expectedError: []error{
				entity.ExistingDataError{
					FieldName:  string(entity.UserFieldExternalUserID),
					EntityName: entity.StudentEntity,
					Index:      0,
				},
				entity.MissingMandatoryFieldError{
					FieldName:  string(entity.UserFieldFirstName),
					EntityName: entity.StudentEntity,
					Index:      1,
				},
			},
			expectedDomainStudents: aggregate.DomainStudents{},
		},
	}
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.ManabieSchool),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			port := DomainStudentService{
				DomainStudent: &tt.domainStudent,
			}
			option := unleash.DomainStudentFeatureOption{
				DomainUserFeatureOption: unleash.DomainUserFeatureOption{
					EnableIgnoreUpdateEmail: false,
					EnableUsername:          true,
				},
				EnableAutoDeactivateAndReactivateStudentV2:            true,
				DisableAutoDeactivateAndReactivateStudent:             false,
				EnableExperimentalBulkInsertEnrollmentStatusHistories: false,
			}
			students, listErrors := port.upsertStudentWithErrorCollections(tt.args.ctx, tt.args.students, option)

			assert.Equal(t, len(tt.expectedError), len(listErrors))
			assert.Equal(t, len(tt.expectedDomainStudents), len(students))
			for idx, err := range tt.expectedError {
				assert.Equal(t, err, listErrors[idx])
			}

			for idx, student := range tt.expectedDomainStudents {
				assert.Equal(t, student.UserID(), students[idx].UserID())
				assert.Equal(t, student.FullName(), students[idx].FullName())
				assert.Equal(t, student.UserName(), students[idx].UserName())
				assert.Equal(t, student.Email(), students[idx].Email())
			}
		})
	}

}
