package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_fatima "github.com/manabie-com/backend/mock/fatima/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_locationRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	existStudentID    = "1111"
	nonExistStudentID = "0"
)

func TestUpsertStudentCoursePackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now()
	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	studentRepo := new(mock_repositories.MockStudentRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	fatimaClient := new(mock_fatima.SubscriptionModifierServiceClient)
	service := UserModifierService{
		DB:           db,
		StudentRepo:  studentRepo,
		FatimaClient: fatimaClient,
		LocationRepo: locationRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case: add 1 package and update 1 package",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: existStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 01",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "course 02",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID: "location-id",
						Name:       "center",
					},
				}, nil).Twice()

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(&entity.LegacyStudent{}, nil)
				fatimaClient.On("AddStudentPackageCourse", mock.Anything, mock.Anything).Once().Return(&fpb.AddStudentPackageCourseResponse{}, nil)
				fatimaClient.On("EditTimeStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.EditTimeStudentPackageResponse{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: upsert multiple student course packages",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: existStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 02",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 03",
						},
						StartTime:   timestamppb.New(now.Add(-24 * time.Hour)),
						EndTime:     timestamppb.New(now.Add(-24 * time.Hour).Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "course 01",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Times(3).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(3).Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID: "location-id",
						Name:       "center",
					},
				}, nil).Times(3)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(&entity.LegacyStudent{}, nil)
				fatimaClient.On("AddStudentPackageCourse", mock.Anything, mock.Anything).Twice().Return(&fpb.AddStudentPackageCourseResponse{}, nil)
				fatimaClient.On("EditTimeStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.EditTimeStudentPackageResponse{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "validate request fail: studentID empty",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: "",
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 01",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = UpsertStudentCoursePackage.validRequest: studentID cannot be empty"),
		},
		{
			name: "validate request fail: start date after end date",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: existStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 01",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(-time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = UpsertStudentCoursePackage.validRequest: package profile start date must before end date"),
		},
		{
			name: "validate request fail: package profile id empty",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: existStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(-time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = UpsertStudentCoursePackage.validRequest: package profile id cannot be empty"),
		},
		// {
		//	name: "validate request fail: locationIds in package profile empty",
		//	ctx:  ctx,
		//	req: &pb.UpsertStudentCoursePackageRequest{
		//		StudentId: existStudentID,
		//		StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		//			{
		//				Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
		//					CourseId: "course 01",
		//				},
		//				StartTime:   timestamppb.New(now),
		//				EndTime:     timestamppb.New(now.Add(time.Hour)),
		//				LocationIds: []string{},
		//			},
		//		},
		//	},
		//	expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = UpsertStudentCoursePackage.validRequest: locationIDs cannot be empty"),
		// },
		// {
		//	name: "validate request fail: locationIds in package profile nil",
		//	ctx:  ctx,
		//	req: &pb.UpsertStudentCoursePackageRequest{
		//		StudentId: existStudentID,
		//		StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		//			{
		//				Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
		//					CourseId: "course 01",
		//				},
		//				StartTime:   timestamppb.New(now),
		//				EndTime:     timestamppb.New(now.Add(time.Hour)),
		//				LocationIds: nil,
		//			},
		//		},
		//	},
		//	expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = UpsertStudentCoursePackage.validRequest: locationIDs cannot be empty"),
		// },
		{
			name: "cannot find student: studentRepo return error",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: existStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 01",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID: "location-id",
						Name:       "center",
					},
				}, nil).Once()

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = StudentRepo.Find, studentID %s: %w", existStudentID, puddle.ErrClosedPool),
		},
		{
			name: "cannot find student",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: nonExistStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 01",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID: "location-id",
						Name:       "center",
					},
				}, nil).Once()

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(nil, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
			expectedErr: fmt.Errorf("rpc error: code = InvalidArgument desc = cannot find student with id: %s", nonExistStudentID),
		},
		{
			name: "call fatima service fail: add package return error",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: nonExistStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 01",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID: "location-id",
						Name:       "center",
					},
				}, nil).Once()

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(&entity.LegacyStudent{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				fatimaClient.On("AddStudentPackageCourse", mock.Anything, mock.Anything).Once().Return(nil, grpc.ErrServerStopped)
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = s.FatimaClient.AddStudentPackageCourse: %s", grpc.ErrServerStopped),
		},
		{
			name: "call fatima service fail: edit package return error",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: nonExistStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "course 01",
						},
						StartTime:   timestamppb.New(now),
						EndTime:     timestamppb.New(now.Add(time.Hour)),
						LocationIds: []string{constants.ManabieOrgLocation},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, mock.Anything, false).Return([]*domain.Location{
					{
						LocationID: "location-id",
						Name:       "center",
					},
				}, nil).Once()

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(&entity.LegacyStudent{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				fatimaClient.On("EditTimeStudentPackage", mock.Anything, mock.Anything).Once().Return(nil, grpc.ErrServerStopped)
			},
			expectedErr: fmt.Errorf("rpc error: code = Internal desc = s.FatimaClient.EditTimeStudentPackage: %s", grpc.ErrServerStopped),
		},
		{
			name: "happy case: student package with course id",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: existStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
							CourseId: "course 01",
						},
						StartTime: timestamppb.New(now),
						EndTime:   timestamppb.New(now.Add(time.Hour)),
						StudentPackageExtra: []*pb.StudentPackageExtra{
							{
								LocationId: constants.ManabieOrgLocation,
								ClassId:    "class-id-1",
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(&entity.LegacyStudent{}, nil)
				fatimaClient.On("AddStudentPackageCourse", mock.Anything, mock.Anything).Once().Return(&fpb.AddStudentPackageCourseResponse{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: student package with student package id",
			ctx:  ctx,
			req: &pb.UpsertStudentCoursePackageRequest{
				StudentId: existStudentID,
				StudentPackageProfiles: []*pb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
					{
						Id: &pb.UpsertStudentCoursePackageRequest_StudentPackageProfile_StudentPackageId{
							StudentPackageId: "student package id 1",
						},
						StartTime: timestamppb.New(now),
						EndTime:   timestamppb.New(now.Add(time.Hour)),
						StudentPackageExtra: []*pb.StudentPackageExtra{
							{
								LocationId: constants.ManabieOrgLocation,
								ClassId:    "class-id-1",
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studentRepo.On("Find", ctx, db, mock.Anything).Once().Return(&entity.LegacyStudent{}, nil)
				fatimaClient.On("EditTimeStudentPackage", mock.Anything, mock.Anything).Once().Return(&fpb.EditTimeStudentPackageResponse{}, nil)
			},
			expectedErr: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			_, err := service.UpsertStudentCoursePackage(testCase.ctx, testCase.req.(*pb.UpsertStudentCoursePackageRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo, fatimaClient)
		})
	}
}
