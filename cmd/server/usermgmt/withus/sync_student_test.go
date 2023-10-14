package withus

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	mock_fatima "github.com/manabie-com/backend/mock/fatima/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
)

func TestSJICReaderToUnicodeUTF8Reader(t *testing.T) {
	sjicCSVFile, err := os.Open("./../testdata/SJIC-W2-D6L_users20230227.tsv")
	if err != nil {
		t.Fatal(err)
	}

	sjicCSVReader := SJISReaderToUnicodeUTF8Reader(sjicCSVFile)
	sjicCSVData, err := sjicCSVReader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	unicodeCSVFile, err := os.Open("./../testdata/UnicodeUTF8-W2-D6L_users20230227.tsv")
	if err != nil {
		t.Fatal(err)
	}

	unicodeCSVReader := csv.NewReader(unicodeCSVFile)
	unicodeCSVReader.Comma = '\t'

	unicodeCSVData, err := unicodeCSVReader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	for i := range sjicCSVData {
		assert.DeepEqual(t, unicodeCSVData[i], sjicCSVData[i])
	}
}

func Test_mapToStudentCourses(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	domainStudent := &mock_usermgmt.Student{
		RandomStudent: mock_usermgmt.RandomStudent{
			UserID: field.NewString(idutil.ULIDNow()),
			Email:  field.NewString("test@manabie.com"),
		},
	}
	studentPackage := &repository.StudentPackage{
		StudentIDAttr: domainStudent.UserID(),
		StartDateAttr: field.NewTime(startTime),
		EndDateAttr:   field.NewTime(endTime),
	}
	course1 := &repository.Course{
		CourseIDAttr:        field.NewString(idutil.ULIDNow()),
		CoursePartnerIDAttr: field.NewString(idutil.ULIDNow()),
	}
	course2 := &repository.Course{
		CourseIDAttr:        field.NewString(idutil.ULIDNow()),
		CoursePartnerIDAttr: field.NewString(idutil.ULIDNow()),
	}
	studentAggregate := aggregate.DomainStudent{
		DomainStudent: domainStudent,
		UserAccessPaths: entity.DomainUserAccessPaths{
			mock_usermgmt.UserAccessPath{
				RandomUserAccessPath: mock_usermgmt.RandomUserAccessPath{
					LocationID: field.NewString("location-id-1"),
				},
			},
		},
	}
	testCases := []struct {
		name                 string
		ctx                  context.Context
		reqCourseIDs         []string
		aggregateStudent     aggregate.DomainStudent
		setup                func(ctx context.Context) *service.DomainStudent
		expectedErr          error
		domainStudentCourses entity.DomainStudentCourses
	}{
		{
			name:             "happy case: partnerCourseIDs empty",
			ctx:              ctx,
			aggregateStudent: studentAggregate,
			setup: func(ctx context.Context) *service.DomainStudent {
				db := &mock_database.Ext{}
				fatimaClient := &mock_fatima.SubscriptionModifierServiceClient{}
				studentPackageRepo := &mock_repositories.MockDomainStudentPackageRepo{}

				studentPackageRepo.On("GetByStudentIDs", ctx, db, []string{studentAggregate.UserID().String()}).Once().Return(entity.DomainStudentPackages{studentPackage}, nil)
				fatimaClient.On("EditTimeStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.EditTimeStudentPackageResponse{}, nil)

				return &service.DomainStudent{
					DB:             db,
					StudentPackage: studentPackageRepo,
					FatimaClient:   fatimaClient,
				}
			},
			expectedErr: nil,
		},
		{
			name:             "happy case: remove 1 existed course and add new 1 course",
			ctx:              ctx,
			aggregateStudent: studentAggregate,
			reqCourseIDs:     []string{course1.CourseID().String()},
			setup: func(ctx context.Context) *service.DomainStudent {
				db := &mock_database.Ext{}
				fatimaClient := &mock_fatima.SubscriptionModifierServiceClient{}
				studentPackageRepo := &mock_repositories.MockDomainStudentPackageRepo{}
				courseRepo := &mock_repositories.MockDomainCourseRepo{}

				courseRepo.On("GetByCoursePartnerIDs", ctx, db, []string{course1.CourseID().String()}).Once().Return(entity.DomainCourses{course1}, nil)
				studentPackageRepo.On("GetByStudentIDs", ctx, db, []string{studentAggregate.UserID().String()}).Once().Return(entity.DomainStudentPackages{studentPackage}, nil)
				studentPackageRepo.On("GetByStudentCourseAndLocationIDs", ctx, db,
					studentAggregate.UserID().String(), course1.CourseID().String(), mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				fatimaClient.On("EditTimeStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.EditTimeStudentPackageResponse{}, nil)

				return &service.DomainStudent{
					DB:             db,
					StudentPackage: studentPackageRepo,
					FatimaClient:   fatimaClient,
					CourseRepo:     courseRepo,
				}
			},
			expectedErr: nil,
			domainStudentCourses: entity.DomainStudentCourses{
				entity.StudentCourseWillBeDelegated{
					DomainStudentCourseAttribute: withUsCourse{
						courseID: course1.CourseID(),
						startAt:  field.NewTime(startTime),
						endAt:    field.NewTime(endTime),
					},
				},
			},
		},
		{
			name:             "happy case: add more course",
			ctx:              ctx,
			aggregateStudent: studentAggregate,
			reqCourseIDs:     []string{course1.CourseID().String(), course2.CourseID().String()},
			setup: func(ctx context.Context) *service.DomainStudent {
				db := &mock_database.Ext{}
				fatimaClient := &mock_fatima.SubscriptionModifierServiceClient{}
				studentPackageRepo := &mock_repositories.MockDomainStudentPackageRepo{}
				courseRepo := &mock_repositories.MockDomainCourseRepo{}

				courseRepo.On("GetByCoursePartnerIDs", ctx, db, []string{course1.CourseID().String(), course2.CourseID().String()}).Once().Return(entity.DomainCourses{course1, course2}, nil)
				studentPackageRepo.On("GetByStudentIDs", ctx, db, []string{studentAggregate.UserID().String()}).Once().Return(entity.DomainStudentPackages{studentPackage}, nil)
				studentPackageRepo.On("GetByStudentCourseAndLocationIDs", ctx, db,
					studentAggregate.UserID().String(), course1.CourseID().String(), mock.Anything).Once().Return(entity.DomainStudentPackages{}, nil)
				studentPackageRepo.On("GetByStudentCourseAndLocationIDs", ctx, db,
					studentAggregate.UserID().String(), course2.CourseID().String(), mock.Anything).Once().Return(entity.DomainStudentPackages{studentPackage}, nil)
				fatimaClient.On("EditTimeStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.EditTimeStudentPackageResponse{}, nil)

				return &service.DomainStudent{
					DB:             db,
					StudentPackage: studentPackageRepo,
					FatimaClient:   fatimaClient,
					CourseRepo:     courseRepo,
				}
			},
			expectedErr: nil,
			domainStudentCourses: entity.DomainStudentCourses{
				entity.StudentCourseWillBeDelegated{
					DomainStudentCourseAttribute: withUsCourse{
						courseID: course1.CourseID(),
						startAt:  field.NewTime(startTime),
						endAt:    field.NewTime(endTime),
					},
				},
				entity.StudentCourseWillBeDelegated{
					DomainStudentCourseAttribute: withUsCourse{
						courseID:         course2.CourseID(),
						studentPackageID: studentPackage.PackageID(),
						startAt:          field.NewTime(startTime),
						endAt:            field.NewTime(endTime),
					},
				},
			},
		},
		{
			name:             "unhappy case: invalid course id",
			ctx:              ctx,
			aggregateStudent: studentAggregate,
			reqCourseIDs:     []string{course1.CourseID().String(), course2.CourseID().String()},
			setup: func(ctx context.Context) *service.DomainStudent {
				db := &mock_database.Ext{}
				fatimaClient := &mock_fatima.SubscriptionModifierServiceClient{}
				studentPackageRepo := &mock_repositories.MockDomainStudentPackageRepo{}
				courseRepo := &mock_repositories.MockDomainCourseRepo{}

				courseRepo.On("GetByCoursePartnerIDs", ctx, db, []string{course1.CourseID().String(), course2.CourseID().String()}).Once().Return(entity.DomainCourses{course1}, nil)

				return &service.DomainStudent{
					DB:             db,
					StudentPackage: studentPackageRepo,
					FatimaClient:   fatimaClient,
					CourseRepo:     courseRepo,
				}
			},
			expectedErr: fmt.Errorf("invalid course ids: %s", []string{course1.CourseID().String(), course2.CourseID().String()}),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)

			studentService := tt.setup(tt.ctx)

			courses, err := mapToStudentCourses(tt.ctx, studentService, tt.reqCourseIDs, tt.aggregateStudent)
			fmt.Println(len(courses))
			if tt.expectedErr != nil || err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, len(tt.domainStudentCourses), len(courses))
				for idx, course := range tt.domainStudentCourses {
					assert.Equal(t, course.CourseID().String(), courses[idx].CourseID().String())
					assert.Equal(t, course.StudentPackageID().String(), courses[idx].StudentPackageID().String())
					assert.Equal(t, course.StartAt().Time(), courses[idx].StartAt().Time())
					assert.Equal(t, course.EndAt().Time(), courses[idx].EndAt().Time())
				}
			}
		})
	}
}
