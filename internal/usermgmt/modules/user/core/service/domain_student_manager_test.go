package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainStudentValidationManagerMock() (prepareStudentValidationManagerMock, StudentValidationManager) {
	m := prepareStudentValidationManagerMock{
		&mock_repositories.MockDomainUserRepo{},
		&mock_repositories.MockDomainUserGroupRepo{},
		&mock_repositories.MockDomainLocationRepo{},
		&mock_repositories.MockDomainGradeRepo{},
		&mock_repositories.MockDomainSchoolRepo{},
		&mock_repositories.MockDomainSchoolCourseRepo{},
		&mock_repositories.MockDomainPrefectureRepo{},
		&mock_repositories.MockDomainTagRepo{},
		&mock_repositories.MockDomainInternalConfigurationRepo{},
		&mock_repositories.MockDomainEnrollmentStatusHistoryRepo{},
		&mock_repositories.MockDomainStudentRepo{},
	}

	service := StudentValidationManager{
		UserRepo:                    m.userRepo,
		UserGroupRepo:               m.userGroupRepo,
		LocationRepo:                m.locationRepo,
		GradeRepo:                   m.gradeRepo,
		SchoolRepo:                  m.schoolRepo,
		SchoolCourseRepo:            m.schoolCourseRepo,
		PrefectureRepo:              m.prefectureRepo,
		TagRepo:                     m.tagRepo,
		InternalConfigurationRepo:   m.internalConfigurationRepo,
		EnrollmentStatusHistoryRepo: m.enrollmentStatusHistoryRepo,
		StudentRepo:                 m.studentRepo,
	}
	return m, service
}

type prepareStudentValidationManagerMock struct {
	userRepo                    *mock_repositories.MockDomainUserRepo
	userGroupRepo               *mock_repositories.MockDomainUserGroupRepo
	locationRepo                *mock_repositories.MockDomainLocationRepo
	gradeRepo                   *mock_repositories.MockDomainGradeRepo
	schoolRepo                  *mock_repositories.MockDomainSchoolRepo
	schoolCourseRepo            *mock_repositories.MockDomainSchoolCourseRepo
	prefectureRepo              *mock_repositories.MockDomainPrefectureRepo
	tagRepo                     *mock_repositories.MockDomainTagRepo
	internalConfigurationRepo   *mock_repositories.MockDomainInternalConfigurationRepo
	enrollmentStatusHistoryRepo *mock_repositories.MockDomainEnrollmentStatusHistoryRepo
	studentRepo                 *mock_repositories.MockDomainStudentRepo
}

func Test_FullyValidate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.ManabieSchool),
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	defer cancel()

	domainStudentToCreate := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			GradeID:        field.NewNullString(),
			Email:          field.NewString("test@manabie.com"),
			Gender:         field.NewString(upb.Gender_FEMALE.String()),
			FirstName:      field.NewString("test first name"),
			LastName:       field.NewString("test last name"),
			ExternalUserID: field.NewString("external-user-id"),
			CurrentGrade:   field.NewInt16(1),
			UserName:       field.NewString("username"),
		},
	}
	domainStudentToUpdate := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			GradeID:        field.NewNullString(),
			Email:          field.NewString("test@manabie.com"),
			Gender:         field.NewString(upb.Gender_FEMALE.String()),
			FirstName:      field.NewString("test first name"),
			LastName:       field.NewString("test last name"),
			ExternalUserID: field.NewString("external-user-id"),
			CurrentGrade:   field.NewInt16(1),
			UserName:       field.NewString("username"),
			UserID:         field.NewString("user-id"),
		},
	}
	domainGrade := &mock_usermgmt.Grade{
		RandomGrade: mock_usermgmt.RandomGrade{
			GradeID:           field.NewString("grade-id-1"),
			PartnerInternalID: field.NewString("partner-internal-id"),
			Name:              field.NewString("grade name"),
		},
	}
	domainTag := &mock_usermgmt.Tag{
		TagIDAttr:             field.NewString(idutil.ULIDNow()),
		PartnerInternalIDAttr: field.NewString(idutil.ULIDNow()),
		TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
	}
	domainSchool := mock_usermgmt.School{
		RandomSchool: mock_usermgmt.RandomSchool{
			SchoolID:          field.NewString("school-id-1"),
			PartnerInternalID: field.NewString("school-internal-id-1"),
			SchoolLevelID:     field.NewString("school-level-1"),
			IsArchived:        field.NewBoolean(false),
		},
	}
	domainSchoolCourse := mock_usermgmt.SchoolCourse{
		RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
			SchoolID:          field.NewString("school-id-1"),
			SchoolCourseID:    field.NewString("school-course-id-1"),
			PartnerInternalID: field.NewString("school-course-internal-id-1"),
			IsArchived:        field.NewBoolean(false),
		},
	}
	domainLocation := mock_usermgmt.Location{
		LocationIDAttr:        field.NewString("location-id-1"),
		PartnerInternalIDAttr: field.NewString("location-internal-id-1"),
	}
	domainEnrollmentStatusHistory := mock_usermgmt.EnrollmentStatusHistory{
		RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
			EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusPotential),
		},
	}
	domainPrefecture := mock_usermgmt.Prefecture{
		RandomPrefecture: mock_usermgmt.RandomPrefecture{
			PrefectureID:   field.NewString(idutil.ULIDNow()),
			PrefectureCode: field.NewString("19"),
		},
	}

	type args struct {
		studentWithIndexes aggregate.DomainStudents
		isEnableUsername   bool
	}

	tests := []struct {
		name                 string
		args                 args
		wantErr              []error
		setupWithMock        func(ctx context.Context, mockInterface interface{})
		wantStudentsToCreate aggregate.DomainStudents
		wantStudentsToUpdate aggregate.DomainStudents
	}{
		{
			name: "happy case: pass all validation for creating student",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToCreate,
						Grade:         domainGrade,
						Tags:          entity.DomainTags{domainTag},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						Prefecture: domainPrefecture,
						IndexAttr:  0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.prefectureRepo.On("GetByPrefectureCodes", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainPrefectures{domainPrefecture}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{},
			wantStudentsToCreate: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:        domainGrade.GradeID(),
							Email:          field.NewString("test@manabie.com"),
							Gender:         field.NewString(upb.Gender_FEMALE.String()),
							FirstName:      field.NewString("test first name"),
							LastName:       field.NewString("test last name"),
							ExternalUserID: field.NewString("external-user-id"),
							CurrentGrade:   field.NewInt16(1),
							UserName:       field.NewString("username"),
						},
					},
					Tags: entity.DomainTags{
						domainTag,
					},
				},
			},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "happy case: pass all validation for updating student",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToUpdate,
						Grade:         domainGrade,
						Tags:          entity.DomainTags{domainTag},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						Prefecture: domainPrefecture,
						IndexAttr:  0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.prefectureRepo.On("GetByPrefectureCodes", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainPrefectures{domainPrefecture}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr:              []error{},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:        domainGrade.GradeID(),
							Email:          field.NewString("test@manabie.com"),
							Gender:         field.NewString(upb.Gender_FEMALE.String()),
							FirstName:      field.NewString("test first name"),
							LastName:       field.NewString("test last name"),
							ExternalUserID: field.NewString("external-user-id"),
							CurrentGrade:   field.NewInt16(1),
							UserName:       field.NewString("username"),
							UserID:         field.NewString("user-id"),
						},
					},
					Tags: entity.DomainTags{
						domainTag,
					},
				},
			},
		},
		{
			name: "unhappy case: failed validation of grade",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToCreate,
						Grade: &mock_usermgmt.Grade{
							RandomGrade: mock_usermgmt.RandomGrade{
								GradeID:           field.NewString("grade-id-1"),
								PartnerInternalID: field.NewString("partner-internal-id-no-exist"),
								Name:              field.NewString("grade name"),
							},
						},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)

			},
			wantErr: []error{
				entity.NotFoundError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentGradeField,
					Index:      0,
					FieldValue: "partner-internal-id-no-exist",
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: can not get tag by partner internal ids",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToCreate,
						Grade: &mock_usermgmt.Grade{
							RandomGrade: mock_usermgmt.RandomGrade{
								GradeID:           field.NewString("grade-id-1"),
								PartnerInternalID: field.NewString("partner-internal-id"),
								Name:              field.NewString("grade name"),
							},
						},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{}, pgx.ErrNoRows,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)

			},
			wantErr: []error{
				errors.WithStack(pgx.ErrNoRows),
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: can not get school info by partner internal ids",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToCreate,
						Grade: &mock_usermgmt.Grade{
							RandomGrade: mock_usermgmt.RandomGrade{
								GradeID:           field.NewString("grade-id-1"),
								PartnerInternalID: field.NewString("partner-internal-id"),
								Name:              field.NewString("grade name"),
							},
						},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, pgx.ErrNoRows,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)

			},
			wantErr: []error{
				errors.WithStack(pgx.ErrNoRows),
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: can not get school course by partner internal ids",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToCreate,
						Grade: &mock_usermgmt.Grade{
							RandomGrade: mock_usermgmt.RandomGrade{
								GradeID:           field.NewString("grade-id-1"),
								PartnerInternalID: field.NewString("partner-internal-id"),
								Name:              field.NewString("grade name"),
							},
						},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, pgx.ErrNoRows,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr:              []error{pgx.ErrNoRows},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of tag",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToCreate,
						Grade: &mock_usermgmt.Grade{
							RandomGrade: mock_usermgmt.RandomGrade{
								GradeID:           field.NewString("grade-id-1"),
								PartnerInternalID: field.NewString("partner-internal-id"),
								Name:              field.NewString("grade name"),
							},
						},
						Tags: entity.DomainTags{&mock_usermgmt.Tag{
							PartnerInternalIDAttr: field.NewString("partner-internal-id-non-existed"),
						}},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.NotFoundError{
					FieldName:  entity.StudentTagsField,
					EntityName: entity.StudentEntity,
					Index:      0,
					FieldValue: "partner-internal-id-non-existed",
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of gender",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								Email:          field.NewString("test@manabie.com"),
								Gender:         field.NewString(upb.Gender_NONE.String()),
								FirstName:      field.NewString("test first name"),
								LastName:       field.NewString("test last name"),
								ExternalUserID: field.NewString("external-user-id"),
								CurrentGrade:   field.NewInt16(1),
							},
						},
						Grade: domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.InvalidFieldError{
					EntityName: entity.UserEntity,
					Index:      0,
					FieldName:  entity.StudentGenderField,
					Reason:     entity.NotMatchingEnum,
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of last name and first name",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								Email:          field.NewString("test@manabie.com"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString(""),
								LastName:       field.NewString(""),
								ExternalUserID: field.NewString("external-user-id"),
								CurrentGrade:   field.NewInt16(1),
							},
						},
						Grade: domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.MissingMandatoryFieldError{
					EntityName: entity.UserEntity,
					Index:      0,
					FieldName:  string(entity.UserFieldFirstName),
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of email (wrong format)",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								Email:          field.NewString("wrong-email-format"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString("first name"),
								LastName:       field.NewString("last name"),
								ExternalUserID: field.NewString("external-user-id"),
								CurrentGrade:   field.NewInt16(1),
							},
						},
						Grade: domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.InvalidFieldError{
					EntityName: entity.UserEntity,
					Index:      0,
					FieldName:  string(entity.UserFieldEmail),
					Reason:     entity.NotMatchingPattern,
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of external user id not unique",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								Email:          field.NewString("emailv1@gmail.com"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString("first name"),
								LastName:       field.NewString("last name"),
								ExternalUserID: field.NewString("external-user-id-existed"),
								CurrentGrade:   field.NewInt16(1),
								UserName:       field.NewString("username"),
							},
						},
						Grade: domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser:      entity.EmptyUser{},
								Email:          field.NewString("emailv1@gmail.com"),
								ExternalUserID: field.NewString("external-user-id-existed"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.ExistingDataError{
					FieldName:  string(entity.UserFieldExternalUserID),
					EntityName: entity.StudentEntity,
					Index:      0,
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of username (duplicated)",
			args: args{

				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								UserName:       field.NewString("email@gmail.com"),
								Email:          field.NewString("emailv1@gmail.com"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString("first name"),
								LastName:       field.NewString("last name"),
								ExternalUserID: field.NewString(""),
								CurrentGrade:   field.NewInt16(1),
							},
						},
						Grade: domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								UserName:       field.NewString("email@gmail.com"),
								Email:          field.NewString("email2@gmail.com"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString("first name"),
								LastName:       field.NewString("last name"),
								ExternalUserID: field.NewString(""),
								CurrentGrade:   field.NewInt16(1),
							},
						},
						Grade: domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 1,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser:      entity.EmptyUser{},
								Email:          field.NewString("emailv1@gmail.com"),
								ExternalUserID: field.NewString("external-user-id-existed"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.DuplicatedFieldError{
					DuplicatedField: string(entity.UserFieldUserName),
					EntityName:      entity.StudentEntity,
					Index:           1,
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:        domainGrade.GradeID(),
							UserName:       field.NewString("email@gmail.com"),
							Email:          field.NewString("emailv1@gmail.com"),
							Gender:         field.NewString(upb.Gender_MALE.String()),
							FirstName:      field.NewString("first name"),
							LastName:       field.NewString("last name"),
							ExternalUserID: field.NewString(""),
							CurrentGrade:   field.NewInt16(1),
						},
					},
					Grade: domainGrade,
				},
			},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of username (already existed)",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToCreate,
						Grade:         domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							UserName: domainStudentToCreate.UserName(),
						},
					}}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.ExistingDataError{
					FieldName:  string(entity.UserFieldUserName),
					EntityName: entity.StudentEntity,
					Index:      0,
				},
			},
		},
		{
			name: "happy case: pass all validation for updating student with username unchanged",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToUpdate,
						Grade:         domainGrade,
						Tags:          entity.DomainTags{domainTag},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						Prefecture: domainPrefecture,
						IndexAttr:  0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							UserID:   domainStudentToUpdate.UserID(),
							UserName: domainStudentToUpdate.UserName(),
						},
					}}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.prefectureRepo.On("GetByPrefectureCodes", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainPrefectures{domainPrefecture}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr:              []error{},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:        domainGrade.GradeID(),
							Email:          field.NewString("test@manabie.com"),
							Gender:         field.NewString(upb.Gender_FEMALE.String()),
							FirstName:      field.NewString("test first name"),
							LastName:       field.NewString("test last name"),
							ExternalUserID: field.NewString("external-user-id"),
							CurrentGrade:   field.NewInt16(1),
							UserName:       field.NewString("username"),
							UserID:         field.NewString("user-id"),
						},
					},
					Tags: entity.DomainTags{
						domainTag,
					},
				},
			},
		},
		{
			name: "unhappy case: failed validation of email (duplicated)",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								Email:          field.NewString("emailv1@gmail.com"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString("first name"),
								LastName:       field.NewString("last name"),
								ExternalUserID: field.NewString(""),
								CurrentGrade:   field.NewInt16(1),
								UserName:       field.NewString("username01"),
							},
						},
						Grade: domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								Email:          field.NewString("emailv1@gmail.com"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString("first name"),
								LastName:       field.NewString("last name"),
								ExternalUserID: field.NewString(""),
								CurrentGrade:   field.NewInt16(1),
								UserName:       field.NewString("username02"),
							},
						},
						Grade: domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 1,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser:      entity.EmptyUser{},
								Email:          field.NewString("emailv1@gmail.com"),
								ExternalUserID: field.NewString("external-user-id-existed"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.DuplicatedFieldError{
					DuplicatedField: string(entity.UserFieldEmail),
					EntityName:      entity.StudentEntity,
					Index:           1,
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:        domainGrade.GradeID(),
							Email:          field.NewString("emailv1@gmail.com"),
							Gender:         field.NewString(upb.Gender_MALE.String()),
							FirstName:      field.NewString("first name"),
							LastName:       field.NewString("last name"),
							ExternalUserID: field.NewString(""),
							CurrentGrade:   field.NewInt16(1),
							UserName:       field.NewString("username01"),
						},
					},
					Grade: domainGrade,
				},
			},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of email (already existed)",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToCreate,
						Grade:         domainGrade,
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email: domainStudentToCreate.Email(),
						},
					}}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.ExistingDataError{
					FieldName:  string(entity.UserFieldEmail),
					EntityName: entity.StudentEntity,
					Index:      0,
				},
			},
		},
		{
			name: "happy case: pass all validation for updating student with email unchanged",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: domainStudentToUpdate,
						Grade:         domainGrade,
						Tags:          entity.DomainTags{domainTag},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						Locations: entity.DomainLocations{
							domainLocation,
						},
						Prefecture: domainPrefecture,
						IndexAttr:  0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							UserID: domainStudentToUpdate.UserID(),
							Email:  domainStudentToUpdate.Email(),
						},
					}}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.prefectureRepo.On("GetByPrefectureCodes", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainPrefectures{domainPrefecture}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr:              []error{},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:        domainGrade.GradeID(),
							Email:          field.NewString("test@manabie.com"),
							Gender:         field.NewString(upb.Gender_FEMALE.String()),
							FirstName:      field.NewString("test first name"),
							LastName:       field.NewString("test last name"),
							ExternalUserID: field.NewString("external-user-id"),
							CurrentGrade:   field.NewInt16(1),
							UserName:       field.NewString("username"),
							UserID:         field.NewString("user-id"),
						},
					},
					Tags: entity.DomainTags{
						domainTag,
					},
				},
			},
		},
		{
			name: "unhappy case: failed validation of location (duplicated)",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								Email:          field.NewString("email@gmail.com"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString("first name"),
								LastName:       field.NewString("last name"),
								ExternalUserID: field.NewString("external-user-id"),
								CurrentGrade:   field.NewInt16(1),
								UserName:       field.NewString("username01"),
							},
						},
						Grade: domainGrade,
						Locations: entity.DomainLocations{
							mock_usermgmt.Location{
								LocationIDAttr:        field.NewString("location-id"),
								PartnerInternalIDAttr: field.NewString("partner-internal-id"),
							},
							mock_usermgmt.Location{
								LocationIDAttr:        field.NewString("location-id"),
								PartnerInternalIDAttr: field.NewString("partner-internal-id"),
							},
						},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
							domainEnrollmentStatusHistory,
						},
						IndexAttr: 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser:      entity.EmptyUser{},
								Email:          field.NewString("emailv1@gmail.com"),
								ExternalUserID: field.NewString("external-user-id-existed"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.DuplicatedFieldError{
					DuplicatedField: string(entity.StudentLocationsField),
					Index:           0,
					EntityName:      entity.StudentEntity,
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
		{
			name: "unhappy case: failed validation of enrollment status histories (missing enrollment status history and location)",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								GradeID:        field.NewNullString(),
								Email:          field.NewString("email@gmail.com"),
								Gender:         field.NewString(upb.Gender_MALE.String()),
								FirstName:      field.NewString("first name"),
								LastName:       field.NewString("last name"),
								ExternalUserID: field.NewString("external-user-id"),
								CurrentGrade:   field.NewInt16(1),
								UserName:       field.NewString("userName"),
							},
						},
						Grade:                     domainGrade,
						Locations:                 entity.DomainLocations{},
						EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{},
						IndexAttr:                 0,
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentValidationManagerMock, ok := genericMock.(prepareStudentValidationManagerMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentValidationManagerMock.gradeRepo.On("GetAll", ctx, &mock_database.Ext{}).Once().Return(
					[]entity.DomainGrade{domainGrade}, nil,
				)
				studentValidationManagerMock.tagRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainTags{domainTag}, nil,
				)
				studentValidationManagerMock.schoolRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchools{domainSchool}, nil,
				)
				studentValidationManagerMock.schoolCourseRepo.On("GetByPartnerInternalIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainSchoolCourses{domainSchoolCourse}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByUserNames", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByEmailsInsensitiveCase", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser: entity.EmptyUser{},
								UserID:    field.NewString("user-id-1"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.userRepo.On("GetByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{
						mock_usermgmt.User{
							RandomUser: mock_usermgmt.RandomUser{
								EmptyUser:      entity.EmptyUser{},
								Email:          field.NewString("emailv1@gmail.com"),
								ExternalUserID: field.NewString("external-user-id-existed"),
							},
						},
					}, nil,
				)
				studentValidationManagerMock.userGroupRepo.On("FindUserGroupByRoleName", ctx, &mock_database.Ext{}, constant.RoleStudent).Once().Return(
					entity.UserGroupWillBeDelegated{}, nil,
				)
				studentValidationManagerMock.studentRepo.On("GetUsersByExternalUserIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.Users{}, nil,
				)
				studentValidationManagerMock.locationRepo.On("RetrieveLowestLevelLocations", ctx, &mock_database.Ext{}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(
					entity.DomainLocations{domainLocation}, nil,
				)
				studentValidationManagerMock.internalConfigurationRepo.On("GetByKey", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.NullDomainConfiguration{}, nil,
				)
				studentValidationManagerMock.enrollmentStatusHistoryRepo.On("GetByStudentIDs", ctx, &mock_database.Ext{}, mock.Anything).Once().Return(
					entity.DomainEnrollmentStatusHistories{}, nil,
				)
			},
			wantErr: []error{
				entity.MissingMandatoryFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      0,
				},
			},
			wantStudentsToCreate: aggregate.DomainStudents{},
			wantStudentsToUpdate: aggregate.DomainStudents{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			m, s := DomainStudentValidationManagerMock()
			tt.setupWithMock(ctx, m)
			studentsToCreate, studentsToUpdate, errorCollection := s.FullyValidate(ctx, &mock_database.Ext{}, tt.args.studentWithIndexes, true)
			t.Log("errors", errorCollection)

			if len(tt.wantStudentsToCreate) != len(studentsToCreate) {
				assert.Fail(t, "FullyValidate() got = %v, want %v", len(studentsToCreate), len(tt.wantStudentsToCreate))
			} else {
				for idx, student := range tt.wantStudentsToCreate {
					assert.Equal(t, student.GradeID().String(), studentsToCreate[idx].GradeID().String())
					assert.Equal(t, student.Gender().String(), studentsToCreate[idx].Gender().String())
					assert.Equal(t, student.FirstName().String(), studentsToCreate[idx].FirstName().String())
					assert.Equal(t, student.LastName().String(), studentsToCreate[idx].LastName().String())
					assert.Equal(t, student.ExternalUserID().String(), studentsToCreate[idx].ExternalUserID().String())
					assert.Equal(t, student.Email().String(), studentsToCreate[idx].Email().String())
					assert.Equal(t, student.UserName().String(), studentsToCreate[idx].UserName().String())
					assert.Equal(t, studentsToCreate[idx].LoginEmail().String(), studentsToCreate[idx].UserID().String()+constant.LoginEmailPostfix)
				}
			}

			if len(tt.wantStudentsToUpdate) != len(studentsToUpdate) {
				assert.Fail(t, "FullyValidate() got = %v, want %v", len(studentsToUpdate), len(tt.wantStudentsToUpdate))
			} else {
				for idx, student := range tt.wantStudentsToUpdate {
					assert.Equal(t, student.GradeID().String(), studentsToUpdate[idx].GradeID().String())
					assert.Equal(t, student.Gender().String(), studentsToUpdate[idx].Gender().String())
					assert.Equal(t, student.FirstName().String(), studentsToUpdate[idx].FirstName().String())
					assert.Equal(t, student.LastName().String(), studentsToUpdate[idx].LastName().String())
					assert.Equal(t, student.ExternalUserID().String(), studentsToUpdate[idx].ExternalUserID().String())
					assert.Equal(t, student.Email().String(), studentsToUpdate[idx].Email().String())
					assert.Equal(t, student.UserName().String(), studentsToUpdate[idx].UserName().String())
					assert.Equal(t, student.UserID().String(), studentsToUpdate[idx].UserID().String())
				}
			}

			if len(tt.wantErr) == 0 {
				assert.Equal(t, 0, len(errorCollection))
			} else {
				assert.Equal(t, len(tt.wantErr), len(errorCollection))
				for idx, err := range errorCollection {
					assert.Equal(t, tt.wantErr[idx].Error(), err.Error())
				}
			}
		})
	}
}

func Test_validationGrade(t *testing.T) {
	t.Parallel()

	type args struct {
		student              aggregate.DomainStudent
		mapPartnerIDAndGrade map[string]entity.DomainGrade
	}

	tests := []struct {
		name            string
		args            args
		wantErr         error
		wantDomainGrade entity.DomainGrade
	}{
		{
			name: "happy case: pass validation of grade",
			args: args{
				student: aggregate.DomainStudent{
					Grade: &mock_usermgmt.Grade{
						RandomGrade: mock_usermgmt.RandomGrade{
							PartnerInternalID: field.NewString("partner-internal-id"),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndGrade: map[string]entity.DomainGrade{
					"partner-internal-id": mock_usermgmt.Grade{
						RandomGrade: mock_usermgmt.RandomGrade{
							PartnerInternalID: field.NewString("partner-internal-id"),
						},
					},
				},
			},
			wantErr: nil,
			wantDomainGrade: mock_usermgmt.Grade{
				RandomGrade: mock_usermgmt.RandomGrade{
					PartnerInternalID: field.NewString("partner-internal-id"),
				},
			},
		},
		{
			name: "unhappy case: grade not exist",
			args: args{
				student: aggregate.DomainStudent{
					Grade: &mock_usermgmt.Grade{
						RandomGrade: mock_usermgmt.RandomGrade{
							PartnerInternalID: field.NewString("partner-internal-id-no-exist"),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndGrade: map[string]entity.DomainGrade{
					"partner-internal-id": mock_usermgmt.Grade{
						RandomGrade: mock_usermgmt.RandomGrade{
							PartnerInternalID: field.NewString("partner-internal-id"),
						},
					},
				},
			},
			wantErr: entity.NotFoundError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentGradeField,
				Index:      0,
				FieldValue: "partner-internal-id-no-exist",
			},
			wantDomainGrade: nil,
		},
		{
			name: "unhappy case: grade is empty",
			args: args{
				student: aggregate.DomainStudent{
					Grade: &mock_usermgmt.Grade{
						RandomGrade: mock_usermgmt.RandomGrade{
							PartnerInternalID: field.NewNullString(),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndGrade: map[string]entity.DomainGrade{
					"partner-internal-id": mock_usermgmt.Grade{
						RandomGrade: mock_usermgmt.RandomGrade{
							PartnerInternalID: field.NewString("partner-internal-id"),
						},
					},
				},
			},
			wantErr: entity.MissingMandatoryFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentGradeField,
				Index:      0,
			},
			wantDomainGrade: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domainGrade, err := validateGrade(tt.args.student, tt.args.mapPartnerIDAndGrade)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.wantDomainGrade, domainGrade)

			}

		})
	}
}

func Test_validationTag(t *testing.T) {
	t.Parallel()

	type args struct {
		student            aggregate.DomainStudent
		mapPartnerIDAndTag map[string]entity.DomainTag
	}

	tests := []struct {
		name          string
		args          args
		wantErr       error
		wantDomainTag entity.DomainTags
	}{
		{
			name: "happy case: pass validation of tag",
			args: args{
				student: aggregate.DomainStudent{
					Tags: entity.DomainTags{
						&mock_usermgmt.Tag{
							PartnerInternalIDAttr: field.NewString("partner-internal-id"),
							TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndTag: map[string]entity.DomainTag{
					"partner-internal-id": mock_usermgmt.Tag{
						PartnerInternalIDAttr: field.NewString("partner-internal-id"),
						TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
						TagIDAttr:             field.NewString("tag-id"),
					},
				},
			},
			wantErr: nil,
			wantDomainTag: entity.DomainTags{
				mock_usermgmt.Tag{
					PartnerInternalIDAttr: field.NewString("partner-internal-id"),
					TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
					TagIDAttr:             field.NewString("tag-id"),
				},
			},
		},
		{
			name: "unhappy case: duplicated tag",
			args: args{
				student: aggregate.DomainStudent{
					Tags: entity.DomainTags{
						&mock_usermgmt.Tag{
							PartnerInternalIDAttr: field.NewString("partner-internal-id"),
							TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
						},
						&mock_usermgmt.Tag{
							PartnerInternalIDAttr: field.NewString("partner-internal-id"),
							TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndTag: map[string]entity.DomainTag{
					"partner-internal-id": mock_usermgmt.Tag{
						PartnerInternalIDAttr: field.NewString("partner-internal-id"),
						TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
					},
				},
			},
			wantErr: entity.DuplicatedFieldError{
				DuplicatedField: entity.StudentTagsField,
				EntityName:      entity.StudentEntity,
				Index:           0,
			},
			wantDomainTag: entity.DomainTags{
				mock_usermgmt.Tag{
					PartnerInternalIDAttr: field.NewString("partner-internal-id"),
					TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
				},
			},
		},
		{
			name: "unhappy case: tag not exist",
			args: args{
				student: aggregate.DomainStudent{
					Tags: entity.DomainTags{
						&mock_usermgmt.Tag{
							PartnerInternalIDAttr: field.NewString("partner-internal-id-non-existed"),
							TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndTag: map[string]entity.DomainTag{
					"partner-internal-id": &mock_usermgmt.Tag{
						PartnerInternalIDAttr: field.NewString("partner-internal-id"),
						TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
					},
				},
			},
			wantErr: entity.NotFoundError{
				FieldName:  entity.StudentTagsField,
				EntityName: entity.StudentEntity,
				Index:      0,
				FieldValue: "partner-internal-id-non-existed",
			},
			wantDomainTag: nil,
		},
		{
			name: "unhappy case: partner id empty",
			args: args{
				student: aggregate.DomainStudent{
					Tags: entity.DomainTags{
						&mock_usermgmt.Tag{
							PartnerInternalIDAttr: field.NewString("non-existed-partner-id"),
							TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndTag: map[string]entity.DomainTag{
					"partner-internal-id": &mock_usermgmt.Tag{
						PartnerInternalIDAttr: field.NewString("partner-internal-id"),
						TagTypeAttr:           field.NewString(entity.UserTagTypeStudent),
					},
				},
			},
			wantErr: entity.NotFoundError{
				FieldName:  entity.StudentTagsField,
				EntityName: entity.StudentEntity,
				Index:      0,
				FieldValue: "non-existed-partner-id",
			},
			wantDomainTag: nil,
		},
		{
			name: "unhappy case: wrong type of student tag",
			args: args{
				student: aggregate.DomainStudent{
					Tags: entity.DomainTags{
						&mock_usermgmt.Tag{
							PartnerInternalIDAttr: field.NewString("partner-internal-id"),
							TagTypeAttr:           field.NewString(entity.UserTagTypeParent),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndTag: map[string]entity.DomainTag{
					"partner-internal-id": &mock_usermgmt.Tag{
						PartnerInternalIDAttr: field.NewString("partner-internal-id"),
						TagTypeAttr:           field.NewString(entity.UserTagTypeParent),
					},
				},
			},
			wantErr: entity.InvalidFieldError{
				FieldName:  entity.StudentTagsField,
				EntityName: entity.StudentEntity,
				Index:      0,
				Reason:     entity.NotMatchingConstants,
			},
			wantDomainTag: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			domainTag, err := validateTag(tt.args.student, tt.args.mapPartnerIDAndTag)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
			assert.Equal(t, tt.wantDomainTag, domainTag)
		})
	}
}

func Test_validateUserNameForCreating(t *testing.T) {
	t.Parallel()

	type args struct {
		student            aggregate.DomainStudent
		mapUserNameAndUser map[string]entity.User
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: pass validation of username creating",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserName: field.NewString("validUsername"),
							UserID:   field.NewNullString(),
						},
					},
					IndexAttr: 0,
				},
				mapUserNameAndUser: map[string]entity.User{
					"username": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							UserName: field.NewString("username"),
							UserID:   field.NewString("user-id"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: there is no username in map",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserName: field.NewString("validUsername"),
							UserID:   field.NewNullString(),
						},
					},
					IndexAttr: 0,
				},
				mapUserNameAndUser: map[string]entity.User{},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: username already exist",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserName: field.NewString("validUsername"),
							UserID:   field.NewNullString(),
						},
					},
				},
				mapUserNameAndUser: map[string]entity.User{
					"validUsername": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							UserName: field.NewString("validUsername"),
							UserID:   field.NewString("user-id"),
						},
					},
				},
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserNameForCreating(tt.args.student, tt.args.mapUserNameAndUser)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateUserNameForUpdating(t *testing.T) {
	t.Parallel()

	type args struct {
		student            aggregate.DomainStudent
		mapUserNameAndUser map[string]entity.User
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: pass validation of username creating",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserName: field.NewString("validUsername"),
							UserID:   field.NewString("user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapUserNameAndUser: map[string]entity.User{
					"validUsername": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							UserName: field.NewString("validUsername"),
							UserID:   field.NewString("user-id"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: username already exist in other user",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserName: field.NewString("validUsername"),
							UserID:   field.NewString("user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapUserNameAndUser: map[string]entity.User{
					"validUsername": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							UserName: field.NewString("validUsername"),
							UserID:   field.NewString("user-id-2"),
						},
					},
				},
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserNameForUpdating(tt.args.student, tt.args.mapUserNameAndUser)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateEmailForCreating(t *testing.T) {
	t.Parallel()

	type args struct {
		student         aggregate.DomainStudent
		mapEmailAndUser map[string]entity.User
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: pass validation of email creating",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:  field.NewString("valid-email@gmail.com"),
							UserID: field.NewString(""),
						},
					},
					IndexAttr: 0,
				},
				mapEmailAndUser: map[string]entity.User{
					"email": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:  field.NewString("email"),
							UserID: field.NewString("user-id"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: there is no email in map",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:  field.NewString("valid-email@gmail.com"),
							UserID: field.NewString(""),
						},
					},
					IndexAttr: 0,
				},
				mapEmailAndUser: map[string]entity.User{},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: email already exist",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:  field.NewString("valid-email@gmail.com"),
							UserID: field.NewString(""),
						},
					},
					IndexAttr: 0,
				},
				mapEmailAndUser: map[string]entity.User{
					"valid-email@gmail.com": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:  field.NewString("valid-email@gmail.com"),
							UserID: field.NewString("user-id"),
						},
					},
				},
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldEmail),
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmailForCreating(tt.args.student, tt.args.mapEmailAndUser)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateEmailForUpdating(t *testing.T) {
	t.Parallel()

	type args struct {
		student         aggregate.DomainStudent
		mapEmailAndUser map[string]entity.User
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: pass validation of external_user_id creating",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:  field.NewString("valid-email@gmail.com"),
							UserID: field.NewString("user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapEmailAndUser: map[string]entity.User{
					"valid-email@gmail.com": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:  field.NewString("valid-email@gmail.com"),
							UserID: field.NewString("user-id"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: external_user_id already exist in other user",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:  field.NewString("valid-email@gmail.com"),
							UserID: field.NewString("user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapEmailAndUser: map[string]entity.User{
					"valid-email@gmail.com": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:  field.NewString("valid-email@gmail.com"),
							UserID: field.NewString("user-id-2"),
						},
					},
				},
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldEmail),
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmailForUpdating(tt.args.student, tt.args.mapEmailAndUser)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateExternalUserIDForCreating(t *testing.T) {
	t.Parallel()

	type args struct {
		student                  aggregate.DomainStudent
		mapExternalUserIDAndUser map[string]entity.User
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: pass validation of external_user_id creating",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapExternalUserIDAndUser: map[string]entity.User{
					"external-user-id-2": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id-2"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: external_user_id already exist in other user",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapExternalUserIDAndUser: map[string]entity.User{
					"external-user-id": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id-2"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
				},
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExternalUserIDForCreating(tt.args.student, tt.args.mapExternalUserIDAndUser)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateExternalUserIDForUpdating(t *testing.T) {
	t.Parallel()

	type args struct {
		reqStudent               aggregate.DomainStudent
		mapUserIDAndUser         map[string]entity.User
		mapExternalUserIDAndUser map[string]entity.User
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: existing external_user_id is not changed when updating",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapUserIDAndUser: map[string]entity.User{
					"user-id": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "happy case: existing external_user_id is empty when updating",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapUserIDAndUser: map[string]entity.User{
					"user-id": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString(""),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: Do not allow to change external_user_id when updating with existing external_user_id",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id-edited"),
						},
					},
					IndexAttr: 0,
				},
				mapUserIDAndUser: map[string]entity.User{
					"user-id": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
				},
			},
			wantErr: entity.UpdateFieldError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
		{
			name: "unhappy case: Do not allow to change external_user_id other user",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
					IndexAttr: 0,
				},
				mapUserIDAndUser: map[string]entity.User{
					"user-id": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id"),
							ExternalUserID: field.NewString(""),
						},
					},
				},
				mapExternalUserIDAndUser: map[string]entity.User{
					"external-user-id": mock_usermgmt.User{
						RandomUser: mock_usermgmt.RandomUser{
							Email:          field.NewString("valid-email@gmail.com"),
							UserID:         field.NewString("user-id-existing"),
							ExternalUserID: field.NewString("external-user-id"),
						},
					},
				},
			},
			wantErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			err := validateExternalUserIDForUpdating(tt.args.reqStudent, tt.args.mapUserIDAndUser, tt.args.mapExternalUserIDAndUser, nil)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_removeDuplicatedExternalUserID(t *testing.T) {
	studentWithIndexes := aggregate.DomainStudents{
		aggregate.DomainStudent{
			DomainStudent: &mock_usermgmt.Student{
				RandomStudent: mock_usermgmt.RandomStudent{
					Email:            field.NewString("emailv1@gmail.com"),
					Gender:           field.NewString(upb.Gender_MALE.String()),
					FirstName:        field.NewString("first name"),
					LastName:         field.NewString("last name"),
					ExternalUserID:   field.NewString("external-user-id-v1"),
					CurrentGrade:     field.NewInt16(1),
					EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
				},
			},
			IndexAttr: 0,
		},
		aggregate.DomainStudent{
			DomainStudent: &mock_usermgmt.Student{
				RandomStudent: mock_usermgmt.RandomStudent{
					Email:            field.NewString("emailv2@gmail.com"),
					Gender:           field.NewString(upb.Gender_MALE.String()),
					FirstName:        field.NewString("first name"),
					LastName:         field.NewString("last name"),
					ExternalUserID:   field.NewString("external-user-id-v2"),
					CurrentGrade:     field.NewInt16(1),
					EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
				},
			},
			IndexAttr: 1,
		},
	}

	type args struct {
		studentWithIndexes aggregate.DomainStudents
	}

	tests := []struct {
		name         string
		args         args
		wantErr      []error
		wantStudents aggregate.DomainStudents
	}{
		{
			name: "happy case: no duplicated external_user_id",
			args: args{
				studentWithIndexes: studentWithIndexes,
			},
			wantErr:      nil,
			wantStudents: studentWithIndexes,
		},
		{
			name: "unhappy case: there is a duplicated external_user_id",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								UserName:         field.NewString("userNameV1"),
								Email:            field.NewString("emailv1@gmail.com"),
								Gender:           field.NewString(upb.Gender_MALE.String()),
								FirstName:        field.NewString("first name"),
								LastName:         field.NewString("last name"),
								ExternalUserID:   field.NewString("external-user-id-duplicated"),
								CurrentGrade:     field.NewInt16(1),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						IndexAttr: 0,
					},
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								UserName:         field.NewString("userNameV2"),
								Email:            field.NewString("emailv2@gmail.com"),
								Gender:           field.NewString(upb.Gender_MALE.String()),
								FirstName:        field.NewString("first name"),
								LastName:         field.NewString("last name"),
								ExternalUserID:   field.NewString("external-user-id-duplicated"),
								CurrentGrade:     field.NewInt16(1),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						IndexAttr: 1,
					},
				},
			},
			wantErr: []error{
				entity.DuplicatedFieldError{
					DuplicatedField: string(entity.UserFieldExternalUserID),
					Index:           1,
					EntityName:      entity.StudentEntity,
				},
			},
			wantStudents: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserName:         field.NewString("userNameV1"),
							Email:            field.NewString("emailv1@gmail.com"),
							Gender:           field.NewString(upb.Gender_MALE.String()),
							FirstName:        field.NewString("first name"),
							LastName:         field.NewString("last name"),
							ExternalUserID:   field.NewString("external-user-id-duplicated"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					IndexAttr: 0,
				},
			},
		},
		{
			name: "happy case: all students have empty external_user_id",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								Email:            field.NewString("emailv1@gmail.com"),
								Gender:           field.NewString(upb.Gender_MALE.String()),
								FirstName:        field.NewString("first name"),
								LastName:         field.NewString("last name"),
								ExternalUserID:   field.NewString(""),
								CurrentGrade:     field.NewInt16(1),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						IndexAttr: 0,
					},
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								Email:            field.NewString("emailv2@gmail.com"),
								Gender:           field.NewString(upb.Gender_MALE.String()),
								FirstName:        field.NewString("first name"),
								LastName:         field.NewString("last name"),
								ExternalUserID:   field.NewString(""),
								CurrentGrade:     field.NewInt16(1),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						IndexAttr: 1,
					},
				},
			},
			wantErr: []error{},

			wantStudents: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:            field.NewString("emailv1@gmail.com"),
							Gender:           field.NewString(upb.Gender_MALE.String()),
							FirstName:        field.NewString("first name"),
							LastName:         field.NewString("last name"),
							ExternalUserID:   field.NewString(""),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					IndexAttr: 0,
				},
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:            field.NewString("emailv2@gmail.com"),
							Gender:           field.NewString(upb.Gender_MALE.String()),
							FirstName:        field.NewString("first name"),
							LastName:         field.NewString("last name"),
							ExternalUserID:   field.NewString(""),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					IndexAttr: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			students, errors := removeDuplicatedExternalUserID(tt.args.studentWithIndexes)

			if len(tt.wantErr) != len(errors) {
				t.Errorf("removeDuplicatedExternalUserID() got = %v, want %v", errors, tt.wantErr)
			} else {
				for i, err := range errors {
					assert.Equal(t, tt.wantErr[i].Error(), err.Error())
				}
			}
			assert.Equal(t, tt.wantStudents, students)

		})
	}
}

func Test_removeDuplicatedUserName(t *testing.T) {
	studentWithIndexes := aggregate.DomainStudents{
		aggregate.DomainStudent{
			DomainStudent: &mock_usermgmt.Student{
				RandomStudent: mock_usermgmt.RandomStudent{
					UserName:         field.NewString("username1"),
					Gender:           field.NewString(upb.Gender_MALE.String()),
					FirstName:        field.NewString("first name"),
					LastName:         field.NewString("last name"),
					ExternalUserID:   field.NewString("external-user-id-v1"),
					CurrentGrade:     field.NewInt16(1),
					EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
				},
			},
			IndexAttr: 0,
		},
		aggregate.DomainStudent{
			DomainStudent: &mock_usermgmt.Student{
				RandomStudent: mock_usermgmt.RandomStudent{
					UserName:         field.NewString("username2"),
					Gender:           field.NewString(upb.Gender_MALE.String()),
					FirstName:        field.NewString("first name"),
					LastName:         field.NewString("last name"),
					ExternalUserID:   field.NewString("external-user-id-v2"),
					CurrentGrade:     field.NewInt16(1),
					EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
				},
			},
			IndexAttr: 1,
		},
	}

	type args struct {
		studentWithIndexes aggregate.DomainStudents
	}

	tests := []struct {
		name         string
		args         args
		wantErr      []error
		wantStudents aggregate.DomainStudents
	}{
		{
			name: "happy case: no duplicated username",
			args: args{
				studentWithIndexes: studentWithIndexes,
			},
			wantErr:      nil,
			wantStudents: studentWithIndexes,
		},
		{
			name: "unhappy case: there is a duplicated username",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								UserName:         field.NewString("username"),
								Gender:           field.NewString(upb.Gender_MALE.String()),
								FirstName:        field.NewString("first name"),
								LastName:         field.NewString("last name"),
								ExternalUserID:   field.NewString("external-user-id-v1"),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						IndexAttr: 0,
					},
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								UserName:         field.NewString("username"),
								Gender:           field.NewString(upb.Gender_MALE.String()),
								FirstName:        field.NewString("first name"),
								LastName:         field.NewString("last name"),
								ExternalUserID:   field.NewString("external-user-id-v2"),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						IndexAttr: 1,
					},
				},
			},
			wantErr: []error{
				entity.DuplicatedFieldError{
					DuplicatedField: string(entity.UserFieldUserName),
					Index:           1,
					EntityName:      entity.StudentEntity,
				},
			},
			wantStudents: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserName:         field.NewString("username"),
							Gender:           field.NewString(upb.Gender_MALE.String()),
							FirstName:        field.NewString("first name"),
							LastName:         field.NewString("last name"),
							ExternalUserID:   field.NewString("external-user-id-v1"),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					IndexAttr: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			students, errors := removeDuplicatedUserName(tt.args.studentWithIndexes)

			if len(tt.wantErr) != len(errors) {
				t.Errorf("removeDuplicatedUserName() got = %v, want %v", errors, tt.wantErr)
			} else {
				for i, err := range errors {
					assert.Equal(t, tt.wantErr[i].Error(), err.Error())
				}
			}
			assert.Equal(t, tt.wantStudents, students)
		})
	}
}

func Test_removeDuplicatedEmail(t *testing.T) {
	studentWithIndexes := aggregate.DomainStudents{
		aggregate.DomainStudent{
			DomainStudent: &mock_usermgmt.Student{
				RandomStudent: mock_usermgmt.RandomStudent{
					Email:            field.NewString("emailv1@gmail.com"),
					Gender:           field.NewString(upb.Gender_MALE.String()),
					FirstName:        field.NewString("first name"),
					LastName:         field.NewString("last name"),
					ExternalUserID:   field.NewString("external-user-id-v1"),
					CurrentGrade:     field.NewInt16(1),
					EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
				},
			},
			IndexAttr: 0,
		},
		aggregate.DomainStudent{
			DomainStudent: &mock_usermgmt.Student{
				RandomStudent: mock_usermgmt.RandomStudent{
					Email:            field.NewString("emailv2@gmail.com"),
					Gender:           field.NewString(upb.Gender_MALE.String()),
					FirstName:        field.NewString("first name"),
					LastName:         field.NewString("last name"),
					ExternalUserID:   field.NewString("external-user-id-v2"),
					CurrentGrade:     field.NewInt16(1),
					EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
				},
			},
			IndexAttr: 1,
		},
	}

	type args struct {
		studentWithIndexes aggregate.DomainStudents
	}

	tests := []struct {
		name         string
		args         args
		wantErr      []error
		wantStudents aggregate.DomainStudents
	}{
		{
			name: "happy case: no duplicated email",
			args: args{
				studentWithIndexes: studentWithIndexes,
			},
			wantErr:      nil,
			wantStudents: studentWithIndexes,
		},
		{
			name: "unhappy case: there is a duplicated email",
			args: args{
				studentWithIndexes: aggregate.DomainStudents{
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								Email:            field.NewString("emailv1@gmail.com"),
								Gender:           field.NewString(upb.Gender_MALE.String()),
								FirstName:        field.NewString("first name"),
								LastName:         field.NewString("last name"),
								ExternalUserID:   field.NewString("external-user-id-v1"),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						IndexAttr: 0,
					},
					aggregate.DomainStudent{
						DomainStudent: &mock_usermgmt.Student{
							RandomStudent: mock_usermgmt.RandomStudent{
								Email:            field.NewString("emailv1@gmail.com"),
								Gender:           field.NewString(upb.Gender_MALE.String()),
								FirstName:        field.NewString("first name"),
								LastName:         field.NewString("last name"),
								ExternalUserID:   field.NewString("external-user-id-v2"),
								EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
							},
						},
						IndexAttr: 1,
					},
				},
			},
			wantErr: []error{
				entity.DuplicatedFieldError{
					DuplicatedField: string(entity.UserFieldEmail),
					Index:           1,
					EntityName:      entity.StudentEntity,
				},
			},
			wantStudents: aggregate.DomainStudents{
				aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							Email:            field.NewString("emailv1@gmail.com"),
							Gender:           field.NewString(upb.Gender_MALE.String()),
							FirstName:        field.NewString("first name"),
							LastName:         field.NewString("last name"),
							ExternalUserID:   field.NewString("external-user-id-v1"),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					IndexAttr: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			students, errors := removeDuplicatedEmail(tt.args.studentWithIndexes)

			if len(tt.wantErr) != len(errors) {
				t.Errorf("removeDuplicatedExternalUserID() got = %v, want %v", errors, tt.wantErr)
			} else {
				for i, err := range errors {
					assert.Equal(t, tt.wantErr[i].Error(), err.Error())
				}
			}
			assert.Equal(t, tt.wantStudents, students)
		})
	}
}

func Test_validateSchoolHistories(t *testing.T) {
	t.Parallel()

	type args struct {
		student                     aggregate.DomainStudent
		mapPartnerIDAndSchool       map[string]entity.DomainSchool
		mapPartnerIDAndSchoolCourse map[string]entity.DomainSchoolCourse
	}

	startDate := time.Now()
	endDate := time.Now().Add(24 * time.Hour)

	schoolHistory := mock_usermgmt.SchoolHistory{
		RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
			StartDate: field.NewTime(startDate),
			EndDate:   field.NewTime(endDate),
		},
	}
	schoolInfo1 := mock_usermgmt.School{
		RandomSchool: mock_usermgmt.RandomSchool{
			SchoolID:          field.NewString("school-id-1"),
			PartnerInternalID: field.NewString("school-internal-id-1"),
			SchoolLevelID:     field.NewString("school-level-1"),
			IsArchived:        field.NewBoolean(false),
		},
	}
	schoolInfo2 := mock_usermgmt.School{
		RandomSchool: mock_usermgmt.RandomSchool{
			SchoolID:          field.NewString("school-id-2"),
			PartnerInternalID: field.NewString("school-internal-id-2"),
			SchoolLevelID:     field.NewString("school-level-2"),
			IsArchived:        field.NewBoolean(false),
		},
	}
	schoolInfoDuplicatedSchoolLevel1 := mock_usermgmt.School{
		RandomSchool: mock_usermgmt.RandomSchool{
			SchoolID:          field.NewString("school-id-2"),
			PartnerInternalID: field.NewString("school-internal-id-2"),
			SchoolLevelID:     field.NewString("school-level-1"),
		},
	}
	schoolCourse1 := mock_usermgmt.SchoolCourse{
		RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
			SchoolID:          field.NewString("school-id-1"),
			SchoolCourseID:    field.NewString("school-course-id-1"),
			PartnerInternalID: field.NewString("school-course-internal-id-1"),
			IsArchived:        field.NewBoolean(false),
		},
	}
	schoolCourse2 := mock_usermgmt.SchoolCourse{
		RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
			SchoolID:          field.NewString("school-id-2"),
			SchoolCourseID:    field.NewString("school-course-id-2"),
			PartnerInternalID: field.NewString("school-course-internal-id-2"),
			IsArchived:        field.NewBoolean(false),
		},
	}

	tests := []struct {
		name                      string
		args                      args
		wantErr                   error
		wantDomainSchoolHistories entity.DomainSchoolHistories
	}{
		{
			name: "happy case: pass validation of school histories",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1},
					SchoolCourses:   entity.DomainSchoolCourses{schoolCourse1},
					IndexAttr:       0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": schoolInfo1,
				},
				mapPartnerIDAndSchoolCourse: map[string]entity.DomainSchoolCourse{
					"school-course-internal-id-1": schoolCourse1,
				},
			},
			wantErr: nil,
			wantDomainSchoolHistories: entity.DomainSchoolHistories{
				mock_usermgmt.SchoolHistory{
					RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-1"),
						SchoolCourseID: field.NewString("school-course-id-1"),
						StartDate:      field.NewTime(startDate),
						EndDate:        field.NewTime(endDate),
					},
				},
			},
		},
		{
			name: "happy case: pass validation of multiple school histories",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory, schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1, schoolInfo2},
					SchoolCourses:   entity.DomainSchoolCourses{schoolCourse1, schoolCourse2},
					IndexAttr:       0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": schoolInfo1,
					"school-internal-id-2": schoolInfo2,
				},
				mapPartnerIDAndSchoolCourse: map[string]entity.DomainSchoolCourse{
					"school-course-internal-id-1": schoolCourse1,
					"school-course-internal-id-2": schoolCourse2,
				},
			},
			wantErr: nil,
			wantDomainSchoolHistories: entity.DomainSchoolHistories{
				mock_usermgmt.SchoolHistory{
					RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-1"),
						SchoolCourseID: field.NewString("school-course-id-1"),
						StartDate:      field.NewTime(startDate),
						EndDate:        field.NewTime(endDate),
					},
				},
				mock_usermgmt.SchoolHistory{
					RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
						SchoolID:       field.NewString("school-id-2"),
						SchoolCourseID: field.NewString("school-course-id-2"),
						StartDate:      field.NewTime(startDate),
						EndDate:        field.NewTime(endDate),
					},
				},
			},
		},
		{
			name: "unhappy case: start date after end date",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{
						mock_usermgmt.SchoolHistory{
							RandomSchoolHistory: mock_usermgmt.RandomSchoolHistory{
								StartDate: field.NewTime(endDate),
								EndDate:   field.NewTime(startDate),
							},
						},
					},
					SchoolInfos:   entity.DomainSchools{schoolInfo1},
					SchoolCourses: entity.DomainSchoolCourses{schoolCourse1},
					IndexAttr:     0,
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  entity.StudentSchoolHistoryStartDateField,
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.StartDateAfterEndDate,
				},
				NestedFieldName: entity.StudentSchoolHistoryField,
				NestedIndex:     0,
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: can not find school",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory},
					SchoolInfos: entity.DomainSchools{
						mock_usermgmt.School{
							RandomSchool: mock_usermgmt.RandomSchool{
								SchoolID:          field.NewString("school-id-1"),
								PartnerInternalID: field.NewString("school-internal-id-non-existed"),
								SchoolLevelID:     field.NewString("school-level-1"),
							},
						},
					},
					SchoolCourses: entity.DomainSchoolCourses{schoolCourse1},
					IndexAttr:     0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id": schoolInfo1,
				},
			},
			wantErr: entity.NotFoundError{
				FieldName:  entity.StudentSchoolField,
				EntityName: entity.StudentEntity,
				Index:      0,
				FieldValue: "school-internal-id-non-existed",
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: school id empty",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory},
					SchoolInfos: entity.DomainSchools{
						schoolInfo1,
						mock_usermgmt.School{
							RandomSchool: mock_usermgmt.RandomSchool{
								SchoolID:          field.NewString("school-id-1"),
								PartnerInternalID: field.NewString(""),
								SchoolLevelID:     field.NewString("school-level-1"),
							},
						},
					},
					SchoolCourses: entity.DomainSchoolCourses{schoolCourse1},
					IndexAttr:     0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id": schoolInfo1,
				},
			},
			wantErr: entity.NotFoundError{
				FieldName:  entity.StudentSchoolField,
				EntityName: entity.StudentEntity,
				Index:      0,
				FieldValue: "school-internal-id-1",
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: school is archived",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1},
					SchoolCourses:   entity.DomainSchoolCourses{schoolCourse1},
					IndexAttr:       0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": mock_usermgmt.School{
						RandomSchool: mock_usermgmt.RandomSchool{
							SchoolID:          field.NewString("school-id-1"),
							PartnerInternalID: field.NewString("school-internal-id-1"),
							SchoolLevelID:     field.NewString("school-level-1"),
							IsArchived:        field.NewBoolean(true),
						},
					},
				},
			},
			wantErr: entity.InvalidFieldError{
				FieldName:  entity.StudentSchoolField,
				EntityName: entity.StudentEntity,
				Index:      0,
				Reason:     entity.Archived,
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: duplicated school levels",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory, schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1, schoolInfoDuplicatedSchoolLevel1},
					SchoolCourses:   entity.DomainSchoolCourses{schoolCourse1, schoolCourse2},
					IndexAttr:       0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": schoolInfo1,
					"school-internal-id-2": schoolInfoDuplicatedSchoolLevel1,
				},
				mapPartnerIDAndSchoolCourse: map[string]entity.DomainSchoolCourse{
					"school-course-internal-id-1": schoolCourse1,
					"school-course-internal-id-2": schoolCourse2,
				},
			},
			wantErr: entity.InvalidFieldError{
				FieldName:  entity.StudentSchoolField,
				EntityName: entity.StudentEntity,
				Index:      0,
				Reason:     entity.AlreadyRegistered,
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: can not find school course",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1},
					SchoolCourses: entity.DomainSchoolCourses{
						mock_usermgmt.SchoolCourse{
							RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
								PartnerInternalID: field.NewString("school-course-internal-id-non-existed"),
							},
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": schoolInfo1,
				},
				mapPartnerIDAndSchoolCourse: map[string]entity.DomainSchoolCourse{
					"school-course-internal-id-1": schoolCourse1,
				},
			},
			wantErr: entity.NotFoundError{
				FieldName:  entity.StudentSchoolCourseField,
				EntityName: entity.StudentEntity,
				Index:      0,
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: duplicated school info",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory, schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1, schoolInfo1},
					SchoolCourses:   entity.DomainSchoolCourses{schoolCourse1, schoolCourse2},
					IndexAttr:       0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": schoolInfo1,
				},
				mapPartnerIDAndSchoolCourse: map[string]entity.DomainSchoolCourse{
					"school-course-internal-id-1": schoolCourse1,
				},
			},
			wantErr: entity.DuplicatedFieldError{
				DuplicatedField: entity.StudentSchoolField,
				EntityName:      entity.StudentEntity,
				Index:           0,
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: school course is archived",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1},
					SchoolCourses:   entity.DomainSchoolCourses{schoolCourse1},
					IndexAttr:       0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": schoolInfo1,
				},
				mapPartnerIDAndSchoolCourse: map[string]entity.DomainSchoolCourse{
					"school-course-internal-id-1": mock_usermgmt.SchoolCourse{
						RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
							SchoolID:          field.NewString("school-id-1"),
							SchoolCourseID:    field.NewString("school-course-id-1"),
							PartnerInternalID: field.NewString("school-course-internal-id-1"),
							IsArchived:        field.NewBoolean(true),
						},
					},
				},
			},
			wantErr: entity.InvalidFieldError{
				FieldName:  entity.StudentSchoolCourseField,
				EntityName: entity.StudentEntity,
				Index:      0,
				Reason:     entity.Archived,
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: duplicated school course",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory, schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1, schoolInfo2},
					SchoolCourses:   entity.DomainSchoolCourses{schoolCourse1, schoolCourse1},
					IndexAttr:       0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": schoolInfo1,
					"school-internal-id-2": schoolInfo2,
				},
				mapPartnerIDAndSchoolCourse: map[string]entity.DomainSchoolCourse{
					"school-course-internal-id-1": schoolCourse1,
				},
			},
			wantErr: entity.DuplicatedFieldError{
				DuplicatedField: entity.StudentSchoolCourseField,
				EntityName:      entity.StudentEntity,
				Index:           0,
			},
			wantDomainSchoolHistories: nil,
		},
		{
			name: "unhappy case: school course invalid (school_id does not equal with school info)",
			args: args{
				student: aggregate.DomainStudent{
					SchoolHistories: entity.DomainSchoolHistories{schoolHistory},
					SchoolInfos:     entity.DomainSchools{schoolInfo1},
					SchoolCourses:   entity.DomainSchoolCourses{schoolCourse1},
					IndexAttr:       0,
				},
				mapPartnerIDAndSchool: map[string]entity.DomainSchool{
					"school-internal-id-1": schoolInfo1,
				},
				mapPartnerIDAndSchoolCourse: map[string]entity.DomainSchoolCourse{
					"school-course-internal-id-1": mock_usermgmt.SchoolCourse{
						RandomSchoolCourse: mock_usermgmt.RandomSchoolCourse{
							SchoolID:          field.NewString("school-id-tem"),
							SchoolCourseID:    field.NewString("school-course-id-1"),
							PartnerInternalID: field.NewString("school-course-internal-id-1"),
							IsArchived:        field.NewBoolean(false),
						},
					},
				},
			},
			wantErr: entity.InvalidFieldError{
				FieldName:  entity.StudentSchoolCourseField,
				EntityName: entity.StudentEntity,
				Index:      0,
				Reason:     entity.NotMatching,
			},
			wantDomainSchoolHistories: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			domainSchoolHistories, err := validateSchoolHistories(tt.args.student, tt.args.mapPartnerIDAndSchool, tt.args.mapPartnerIDAndSchoolCourse)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				for idx, v := range tt.wantDomainSchoolHistories {
					assert.True(t, v.SchoolID().Equal(domainSchoolHistories[idx].SchoolID()))
					assert.True(t, v.SchoolCourseID().Equal(domainSchoolHistories[idx].SchoolCourseID()))
					assert.True(t, v.StartDate().Time().Equal(domainSchoolHistories[idx].StartDate().Time()))
					assert.True(t, v.EndDate().Time().Equal(domainSchoolHistories[idx].EndDate().Time()))
				}
			}
		})
	}
}

func Test_validateEnrollmentStatusHistoriesForCreating(t *testing.T) {
	t.Parallel()

	type args struct {
		studentWithIndex aggregate.DomainStudent
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: pass validation of enrollment status histories for creating",
			args: args{
				studentWithIndex: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusNonPotential)),
								StartDate:        field.NewTime(time.Now()),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: missing enrollment status histories for creating",
			args: args{
				studentWithIndex: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{},
					IndexAttr:                 0,
				},
			},
			wantErr: entity.MissingMandatoryFieldError{
				FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
		{
			name: "unhappy case: create student with enrollment status temporary without activated enrollment status for creating",
			args: args{
				studentWithIndex: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusTemporary)),
								StartDate:        field.NewTime(time.Now()),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.MissingActivatedEnrollmentStatus,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},
		{
			name: "happy case: create student with enrollment status temporary and activated enrollment status for creating",
			args: args{
				studentWithIndex: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusTemporary)),
								StartDate:        field.NewTime(time.Now()),
								LocationID:       field.NewString("location-id-1"),
							},
						},
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusPotential)),
								StartDate:        field.NewTime(time.Now()),
								LocationID:       field.NewString("location-id-2"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: create student with enrollment status enrolled and start date is future date for creating",
			args: args{
				studentWithIndex: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusEnrolled)),
								StartDate:        field.NewTime(time.Now().Add(time.Hour * 24 * 7)),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: create student with enrollment status potential and start date is zero date for creating",
			args: args{
				studentWithIndex: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusPotential)),
								StartDate:        field.Time(field.NewNullDate()),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			err := validateEnrollmentStatusHistoriesForCreating(tt.args.studentWithIndex)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateEnrollmentStatusHistories(t *testing.T) {
	t.Parallel()

	type args struct {
		reqStudent aggregate.DomainStudent
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: enrollment status and location match",
			args: args{
				reqStudent: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusPotential)),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: enrollment status smaller than locations",
			args: args{
				reqStudent: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:     field.NewString("user-id-1"),
								LocationID: field.NewString("location-id-1"),
							},
						},
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusPotential)),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.Empty,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},
		{
			name: "unhappy case: locations smaller than enrollment status",
			args: args{
				reqStudent: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusNonPotential)),
								StartDate:        field.NewTime(time.Now()),
								LocationID:       field.NewString("location-id-1"),
							},
						},
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusNonPotential)),
								StartDate:        field.NewTime(time.Now()),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.StudentLocationsField),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.Empty,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     1,
			},
		},
		{
			name: "unhappy case: location is empty but enrollment status is not",
			args: args{
				reqStudent: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusPotential)),
								LocationID:       field.NewString(""),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.StudentLocationsField),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.Empty,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},
		{
			name: "unhappy case: enrollment status is empty but location is not empty",
			args: args{
				reqStudent: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewNullString(),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.Empty,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},

		{
			name: "unhappy case: create student with enrollment status potential and start date is future date for creating",
			args: args{
				reqStudent: aggregate.DomainStudent{
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(string(entity.StudentEnrollmentStatusPotential)),
								StartDate:        field.NewTime(time.Now().Add(time.Hour * 24 * 7)),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.StartDateAfterCurrentDate,
				},
				NestedIndex:     0,
				NestedFieldName: entity.EnrollmentStatusHistories,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			err := validateEnrollmentStatusHistories(tt.args.reqStudent, false)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateEnrollmentStatusHistoriesForUpdating(t *testing.T) {
	t.Parallel()

	type args struct {
		reqStudent                        aggregate.DomainStudent
		existingEnrollmentStatusHistories entity.DomainEnrollmentStatusHistories
		isOrderFlow                       bool
	}

	tests := []struct {
		name      string
		args      args
		wantErr   error
		isERPFlow bool
	}{
		{
			name: "happy case: allow to update enrollment status histories Potential to Enrolled",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:   field.NewNullString(),
							Email:     field.NewString("test@manabie.com"),
							Gender:    field.NewString(upb.Gender_FEMALE.String()),
							FirstName: field.NewString("test first name"),
							LastName:  field.NewString("test last name"),
							UserID:    field.NewString("user-id-1"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
								LocationID:       field.NewString("location-id-1"),
								StartDate:        field.NewTime(time.Now()),
							},
						},
					},
					IndexAttr: 0,
				},
				existingEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id-1"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusPotential),
							LocationID:       field.NewString("location-id-1"),
							StartDate:        field.NewTime(time.Now()),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: skip validation if enrollment status histories is empty && locations is empty",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:   field.NewNullString(),
							Email:     field.NewString("test@manabie.com"),
							Gender:    field.NewString(upb.Gender_FEMALE.String()),
							FirstName: field.NewString("test first name"),
							LastName:  field.NewString("test last name"),
							UserID:    field.NewString("user-id-1"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{},
					IndexAttr:                 0,
				},
				existingEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id-1"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusPotential),
							LocationID:       field.NewString("location-id-1"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: skip validation if enrollment status histories does not change",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:   field.NewNullString(),
							Email:     field.NewString("test@manabie.com"),
							Gender:    field.NewString(upb.Gender_FEMALE.String()),
							FirstName: field.NewString("test first name"),
							LastName:  field.NewString("test last name"),
							UserID:    field.NewString("user-id-1"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
				existingEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id-1"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
							LocationID:       field.NewString("location-id-1"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: new location is added, no existing enrollment status histories at new location",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:   field.NewNullString(),
							Email:     field.NewString("test@manabie.com"),
							Gender:    field.NewString(upb.Gender_FEMALE.String()),
							FirstName: field.NewString("test first name"),
							LastName:  field.NewString("test last name"),
							UserID:    field.NewString("user-id-1"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
				existingEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id-1"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
							LocationID:       field.NewString("location-id-2"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: skip validation if req enrollment status history and existing enrollment status history are temporary",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:   field.NewNullString(),
							Email:     field.NewString("test@manabie.com"),
							Gender:    field.NewString(upb.Gender_FEMALE.String()),
							FirstName: field.NewString("test first name"),
							LastName:  field.NewString("test last name"),
							UserID:    field.NewString("user-id-1"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusTemporary),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
				existingEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id-1"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusTemporary),
							LocationID:       field.NewString("location-id-1"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: cannot add new enrollment status histories with temporary status without at least on activated status",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:   field.NewNullString(),
							Email:     field.NewString("test@manabie.com"),
							Gender:    field.NewString(upb.Gender_FEMALE.String()),
							FirstName: field.NewString("test first name"),
							LastName:  field.NewString("test last name"),
							UserID:    field.NewString("user-id-1"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusTemporary),
								LocationID:       field.NewString("location-id-2"),
							},
						},
					},
					IndexAttr: 0,
				},
				existingEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id-1"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusTemporary),
							LocationID:       field.NewString("location-id-1"),
						},
					}, mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id-1"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusPotential),
							LocationID:       field.NewString("location-id-2"),
							StartDate:        field.NewTime(time.Now().AddDate(0, 0, 1)),
						},
					},
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.MissingActivatedEnrollmentStatus,
				},
				NestedIndex:     0,
				NestedFieldName: entity.EnrollmentStatusHistories,
			},
		},
		{
			name: "unhappy case: non-ERP status can not be changed to others at Order flow",
			args: args{
				reqStudent: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							GradeID:   field.NewNullString(),
							Email:     field.NewString("test@manabie.com"),
							Gender:    field.NewString(upb.Gender_FEMALE.String()),
							FirstName: field.NewString("test first name"),
							LastName:  field.NewString("test last name"),
							UserID:    field.NewString("user-id-1"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								UserID:           field.NewString("user-id-1"),
								EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusPotential),
								LocationID:       field.NewString("location-id-1"),
							},
						},
					},
					IndexAttr: 0,
				},
				existingEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
					mock_usermgmt.EnrollmentStatusHistory{
						RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
							UserID:           field.NewString("user-id-1"),
							EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusEnrolled),
							LocationID:       field.NewString("location-id-1"),
							StartDate:        field.NewTime(time.Now()),
						},
					},
				},
				isOrderFlow: true,
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.ChangingNonERPStatusToOtherStatusAtOrderFlow,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			err := validateEnrollmentStatusHistoriesForUpdating(tt.args.reqStudent, tt.args.existingEnrollmentStatusHistories, tt.args.isOrderFlow)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateEntityEnrollmentStatusHistoryForUpdating(t *testing.T) {
	t.Parallel()

	now := time.Now()

	type args struct {
		activatedOrLatestEnrollmentStatus entity.DomainEnrollmentStatusHistory
		reqEnrollmentStatus               entity.DomainEnrollmentStatusHistory
		studentIndex                      int
		nestedIndex                       int
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "happy case: no change in req and db",
			args: args{
				activatedOrLatestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: Cannot change Non-Potential to any status",
			args: args{
				activatedOrLatestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-39*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.ChangingNonPotentialToOtherStatus,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},
		{
			name: "unhappy case: Cannot update when req start date is smaller DB start date",
			args: args{
				activatedOrLatestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-100*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.ChangingStartDateWithoutChangingStatus,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},
		{
			name: "happy case: Can change without compare millisecond in start_date",
			args: args{
				activatedOrLatestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC).Add(2*time.Millisecond),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
			},
			wantErr: nil,
		},
		{
			name: "happy case: can update with start date req is now, change enrollment and existed record in DB",
			args: args{
				activatedOrLatestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},

				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
					startDate:        field.NewTime(time.Now()),
				},
			},
			wantErr: nil,
		},
		{
			name: "unhappy case: start_date is diff but enrollment status is the same from client (don't allow update start_date if enrollment status don't change)",
			args: args{
				activatedOrLatestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},

				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2001, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.ChangingStartDateWithoutChangingStatus,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},
		{
			name: "unhappy case: enrollment status is diff but start_date is the same from client",
			args: args{
				activatedOrLatestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},

				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},
			},
			wantErr: entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
					EntityName: entity.StudentEntity,
					Index:      0,
					Reason:     entity.ChangingStatusWithoutChangingStartDate,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     0,
			},
		},
		{
			name: "happy case: Can change any status to temporary",
			args: args{
				activatedOrLatestEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
					now.Add(-40*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
				reqEnrollmentStatus: createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
					upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
					now.Add(30*time.Hour),
					now.Add(200*time.Hour),
					"order-id",
					1,
				),
			},
			wantErr: nil,
		},
		{
			name: "happy case: can update when start_date is now in req, enrollmentStatus is changed",
			args: args{
				activatedOrLatestEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()),
					startDate:        field.NewTime(time.Date(2000, time.October, 11, 11, 11, 11, 11, time.UTC)),
				},

				reqEnrollmentStatus: &MockDomainEnrollmentStatusHistory{
					userID:           field.NewString("student-id"),
					locationID:       field.NewString("Manabie"),
					enrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
					startDate:        field.NewTime(time.Now()),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			err := validateEntityEnrollmentStatusHistoryForUpdating(tt.args.activatedOrLatestEnrollmentStatus, tt.args.reqEnrollmentStatus, tt.args.studentIndex, tt.args.nestedIndex)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_validateLocations(t *testing.T) {
	t.Parallel()

	type args struct {
		student                       aggregate.DomainStudent
		mapPartnerIDAndLowestLocation map[string]entity.DomainLocation
	}

	tests := []struct {
		name                          string
		args                          args
		wantErr                       error
		wantDomainLocation            entity.DomainLocations
		wantEnrollmentStatusHistories entity.DomainEnrollmentStatusHistories
	}{
		{
			name: "happy case: pass all validation of location",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:           field.NewString("user-id"),
							GradeID:          field.NewString("grade-id"),
							Email:            field.NewString("test@manabie.com"),
							Gender:           field.NewString(upb.Gender_FEMALE.String()),
							FirstName:        field.NewString("test first name"),
							LastName:         field.NewString("test last name"),
							ExternalUserID:   field.NewString("external-user-id"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					Locations: entity.DomainLocations{
						mock_usermgmt.Location{
							PartnerInternalIDAttr: field.NewString("partner-id"),
						},
					},
					EnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
						mock_usermgmt.EnrollmentStatusHistory{
							RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
								EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusPotential),
							},
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndLowestLocation: map[string]entity.DomainLocation{
					"partner-id": mock_usermgmt.Location{
						LocationIDAttr:        field.NewString("location-id"),
						PartnerInternalIDAttr: field.NewString("partner-id"),
					},
				},
			},
			wantErr: nil,
			wantDomainLocation: entity.DomainLocations{
				mock_usermgmt.Location{
					NullDomainLocation:    entity.NullDomainLocation{},
					LocationIDAttr:        field.NewString("location-id"),
					PartnerInternalIDAttr: field.NewString("partner-id"),
				},
			},
			wantEnrollmentStatusHistories: entity.DomainEnrollmentStatusHistories{
				mock_usermgmt.EnrollmentStatusHistory{
					RandomEnrollmentStatusHistory: mock_usermgmt.RandomEnrollmentStatusHistory{
						EnrollmentStatus: field.NewString(entity.StudentEnrollmentStatusPotential),
						LocationID:       field.NewString("location-id"),
					},
				},
			},
		},
		{
			name: "unhappy case: locations is duplicated",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:           field.NewString("user-id"),
							GradeID:          field.NewString("grade-id"),
							Email:            field.NewString("test@manabie.com"),
							Gender:           field.NewString(upb.Gender_FEMALE.String()),
							FirstName:        field.NewString("test first name"),
							LastName:         field.NewString("test last name"),
							ExternalUserID:   field.NewString("external-user-id"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					Locations: entity.DomainLocations{
						mock_usermgmt.Location{
							PartnerInternalIDAttr: field.NewString("partner-id"),
						},
						mock_usermgmt.Location{
							PartnerInternalIDAttr: field.NewString("partner-id"),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndLowestLocation: map[string]entity.DomainLocation{
					"partner-id": mock_usermgmt.Location{
						LocationIDAttr:        field.NewString("location-id"),
						PartnerInternalIDAttr: field.NewString("partner-id"),
					},
				},
			},
			wantErr: entity.DuplicatedFieldError{
				DuplicatedField: string(entity.StudentLocationsField),
				Index:           0,
				EntityName:      entity.StudentEntity,
			},
			wantDomainLocation:            nil,
			wantEnrollmentStatusHistories: nil,
		},
		{
			name: "one of location_id is empty",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:           field.NewString("user-id"),
							GradeID:          field.NewString("grade-id"),
							Email:            field.NewString("test@manabie.com"),
							Gender:           field.NewString(upb.Gender_FEMALE.String()),
							FirstName:        field.NewString("test first name"),
							LastName:         field.NewString("test last name"),
							ExternalUserID:   field.NewString("external-user-id"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					Locations: entity.DomainLocations{
						mock_usermgmt.Location{
							PartnerInternalIDAttr: field.NewString("partner-id"),
						},
						mock_usermgmt.Location{
							PartnerInternalIDAttr: field.NewString(""),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndLowestLocation: map[string]entity.DomainLocation{
					"partner-id": mock_usermgmt.Location{
						LocationIDAttr:        field.NewString("location-id"),
						PartnerInternalIDAttr: field.NewString("partner-id"),
					},
				},
			},
			wantErr: entity.InvalidFieldError{
				FieldName:  entity.StudentLocationsField,
				EntityName: entity.StudentEntity,
				Index:      0,
				Reason:     entity.Empty,
			},
			wantDomainLocation:            nil,
			wantEnrollmentStatusHistories: nil,
		},
		{
			name: "unhappy: location not found in system by partner_internal_id",
			args: args{
				student: aggregate.DomainStudent{
					DomainStudent: &mock_usermgmt.Student{
						RandomStudent: mock_usermgmt.RandomStudent{
							UserID:           field.NewString("user-id"),
							GradeID:          field.NewString("grade-id"),
							Email:            field.NewString("test@manabie.com"),
							Gender:           field.NewString(upb.Gender_FEMALE.String()),
							FirstName:        field.NewString("test first name"),
							LastName:         field.NewString("test last name"),
							ExternalUserID:   field.NewString("external-user-id"),
							CurrentGrade:     field.NewInt16(1),
							EnrollmentStatus: field.NewString(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()),
						},
					},
					Locations: entity.DomainLocations{
						mock_usermgmt.Location{
							PartnerInternalIDAttr: field.NewString("partner-id-v1"),
						},
						mock_usermgmt.Location{
							PartnerInternalIDAttr: field.NewString("parent-id-v2"),
						},
					},
					IndexAttr: 0,
				},
				mapPartnerIDAndLowestLocation: map[string]entity.DomainLocation{
					"partner-id": mock_usermgmt.Location{
						LocationIDAttr:        field.NewString("location-id"),
						PartnerInternalIDAttr: field.NewString("partner-id"),
					},
				},
			},
			wantErr: entity.NotFoundError{
				FieldName:  entity.StudentLocationsField,
				EntityName: entity.StudentEntity,
				Index:      0,
			},
			wantDomainLocation:            nil,
			wantEnrollmentStatusHistories: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			domainLocations, enrollmentStatusHistories, err := validateLocations(tt.args.student, tt.args.mapPartnerIDAndLowestLocation)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
			if tt.wantDomainLocation != nil {
				for idx, domainLocation := range domainLocations {
					assert.Equal(t, tt.wantDomainLocation[idx].LocationID(), domainLocation.LocationID())
					assert.Equal(t, tt.wantDomainLocation[idx].PartnerInternalID(), domainLocation.PartnerInternalID())
				}
			} else {
				assert.Nil(t, domainLocations)
			}

			if tt.wantEnrollmentStatusHistories != nil {
				for idx, enrollmentStatusHistory := range enrollmentStatusHistories {
					assert.Equal(t, tt.wantEnrollmentStatusHistories[idx].EnrollmentStatus(), enrollmentStatusHistory.EnrollmentStatus())
					assert.Equal(t, tt.wantEnrollmentStatusHistories[idx].LocationID(), enrollmentStatusHistory.LocationID())
				}
			} else {
				assert.Nil(t, enrollmentStatusHistories)
			}
		})
	}
}

func TestStudentValidationManager_validateUserAddress(t *testing.T) {
	t.Parallel()

	type args struct {
		student                        aggregate.DomainStudent
		mapPrefectureCodeAndPrefecture map[string]entity.DomainPrefecture
	}

	mockDomainPrefecture := &mock_usermgmt.Prefecture{
		RandomPrefecture: mock_usermgmt.RandomPrefecture{
			PrefectureCode: field.NewString("19"),
			PrefectureID:   field.NewString(idutil.ULIDNow()),
		},
	}
	mockDomainUserAddress := &mock_usermgmt.UserAddress{
		RandomUserAddress: mock_usermgmt.RandomUserAddress{
			UserAddressID: field.NewString(idutil.ULIDNow()),
		},
	}
	tests := []struct {
		name                      string
		args                      args
		expectedErr               error
		expectedDomainUserAddress entity.DomainUserAddress
	}{
		{
			name: "happy case: have prefecture",
			args: args{
				student: aggregate.DomainStudent{
					Prefecture:  mockDomainPrefecture,
					UserAddress: mockDomainUserAddress,
					IndexAttr:   0,
				},
				mapPrefectureCodeAndPrefecture: map[string]entity.DomainPrefecture{
					"19": mockDomainPrefecture,
				},
			},
			expectedErr: nil,
			expectedDomainUserAddress: &mock_usermgmt.UserAddress{
				RandomUserAddress: mock_usermgmt.RandomUserAddress{
					PrefectureID:  mockDomainPrefecture.PrefectureID(),
					UserAddressID: mockDomainUserAddress.UserAddressID(),
				},
			},
		},
		{
			name: "happy case: empty prefecture",
			args: args{
				student: aggregate.DomainStudent{
					UserAddress: mockDomainUserAddress,
					IndexAttr:   0,
				},
				mapPrefectureCodeAndPrefecture: map[string]entity.DomainPrefecture{
					"19": mockDomainPrefecture,
				},
			},
			expectedErr: nil,
			expectedDomainUserAddress: &mock_usermgmt.UserAddress{
				RandomUserAddress: mock_usermgmt.RandomUserAddress{
					UserAddressID: mockDomainUserAddress.UserAddressID(),
				},
			},
		},
		{
			name: "unhappy case: prefecture invalid",
			args: args{
				student: aggregate.DomainStudent{
					Prefecture:  mockDomainPrefecture,
					UserAddress: mockDomainUserAddress,
					IndexAttr:   0,
				},
				mapPrefectureCodeAndPrefecture: map[string]entity.DomainPrefecture{
					"20": mockDomainPrefecture,
				},
			},
			expectedErr: entity.NotFoundError{
				FieldName:  entity.StudentUserAddressPrefectureField,
				EntityName: entity.StudentEntity,
				Index:      0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)

			domainUserAddress, err := validateUserAddress(tt.args.student, tt.args.mapPrefectureCodeAndPrefecture)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expectedDomainUserAddress.PrefectureID(), domainUserAddress.PrefectureID())
				assert.Equal(t, tt.expectedDomainUserAddress.UserAddressID(), domainUserAddress.UserAddressID())
			}
		})
	}
}

func TestStudentValidationManager_validatePhoneNumbers(t *testing.T) {
	t.Parallel()

	type args struct {
		student aggregate.DomainStudent
	}

	phoneNumber1 := mock_usermgmt.NewUserPhoneNumber("0987654321", entity.UserPhoneNumberTypeStudentPhoneNumber)
	phoneNumber2 := mock_usermgmt.NewUserPhoneNumber("0123456789", entity.UserPhoneNumberTypeStudentHomePhoneNumber)
	phoneNumberInvalid := mock_usermgmt.NewUserPhoneNumber("098765;4321", entity.UserPhoneNumberTypeStudentPhoneNumber)
	tests := []struct {
		name                           string
		args                           args
		expectedErr                    error
		expectedDomainUserPhoneNumbers entity.DomainUserPhoneNumbers
	}{
		{
			name: "happy case: 1 phone_number",
			args: args{
				student: aggregate.DomainStudent{
					UserPhoneNumbers: entity.DomainUserPhoneNumbers{phoneNumber1},
					IndexAttr:        0,
				},
			},
			expectedErr:                    nil,
			expectedDomainUserPhoneNumbers: entity.DomainUserPhoneNumbers{phoneNumber1},
		},
		{
			name: "happy case: 2 phone_number",
			args: args{
				student: aggregate.DomainStudent{
					UserPhoneNumbers: entity.DomainUserPhoneNumbers{phoneNumber1, phoneNumber2},
					IndexAttr:        0,
				},
			},
			expectedErr:                    nil,
			expectedDomainUserPhoneNumbers: entity.DomainUserPhoneNumbers{phoneNumber1, phoneNumber2},
		},
		{
			name: "unhappy case: duplicated phone_number",
			args: args{
				student: aggregate.DomainStudent{
					UserPhoneNumbers: entity.DomainUserPhoneNumbers{phoneNumber1, phoneNumber1},
					IndexAttr:        0,
				},
			},
			expectedErr: entity.DuplicatedFieldError{
				DuplicatedField: entity.StudentFieldStudentPhoneNumber,
				EntityName:      entity.StudentEntity,
				Index:           0,
			},
		},
		{
			name: "unhappy case: phone_number invalid type",
			args: args{
				student: aggregate.DomainStudent{
					UserPhoneNumbers: entity.DomainUserPhoneNumbers{phoneNumberInvalid},
					IndexAttr:        0,
				},
			},
			expectedErr: entity.InvalidFieldError{
				FieldName:  entity.StudentFieldStudentPhoneNumber,
				EntityName: entity.StudentEntity,
				Index:      0,
				Reason:     entity.NotMatchingPattern,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)

			domainUserPhoneNumbers, err := validatePhoneNumbers(tt.args.student)
			assert.Equal(t, len(tt.expectedDomainUserPhoneNumbers), len(domainUserPhoneNumbers))
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				for idx, phoneNumber := range tt.expectedDomainUserPhoneNumbers {
					assert.Equal(t, phoneNumber.PhoneNumber(), domainUserPhoneNumbers[idx].PhoneNumber())
					assert.Equal(t, phoneNumber.Type(), domainUserPhoneNumbers[idx].Type())
				}
			}
		})
	}
}

func TestStudentValidationManager_GetMapUserNameAndUserByUserName(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	managerMock, manager := DomainStudentValidationManagerMock()

	student1 := mock_usermgmt.User{
		RandomUser: mock_usermgmt.RandomUser{
			UserName: field.NewString("student1"),
		},
	}

	student2 := mock_usermgmt.User{
		RandomUser: mock_usermgmt.RandomUser{
			UserName: field.NewString("student2"),
		},
	}

	type args struct {
		db        database.QueryExecer
		usernames []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]entity.User
		setup   func()
		wantErr error
	}{
		{
			name: "happy case: get map username and user by username with 2 username but only 1 user found",
			args: args{
				usernames: []string{"student1", "student2"},
			},
			setup: func() {
				managerMock.userRepo.On("GetByUserNames", mock.Anything, mock.Anything, []string{"student1", "student2"}).
					Once().Return(entity.Users{student1}, nil)
			},
			want: map[string]entity.User{
				"student1": student1,
			},
			wantErr: nil,
		},
		{
			name: "happy case: get map username and user by username with 2 username and 2 user found",
			args: args{
				usernames: []string{"student1", "student2"},
			},
			setup: func() {
				managerMock.userRepo.On("GetByUserNames", mock.Anything, mock.Anything, []string{"student1", "student2"}).
					Once().Return(entity.Users{student1, student2}, nil)
			},
			want: map[string]entity.User{
				"student1": student1,
				"student2": student2,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := manager.GetMapUserNameAndUserByUserName(ctx, new(mock_database.Ext), tt.args.usernames)
			assert.Equalf(t, tt.want, got, "GetMapUserNameAndUserByUserName(%v, %v, %v)", ctx, tt.args.db, tt.args.usernames)
			assert.Equalf(t, tt.wantErr, err, "GetMapUserNameAndUserByUserName(%v, %v, %v)", ctx, tt.args.db, tt.args.usernames)
		})
	}
}
