package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/course_mgmt"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCourseMgMt_ManualUpsertStudentCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                    *mockDb.Ext
		studentService        *mockServices.IStudentServiceForCourseMgMt
		studentPackageService *mockServices.IStudentPackageForCourseMgMt
		subscriptionService   *mockServices.ISubscriptionServiceForCourseMgMte
		tx                    *mockDb.Tx
	)

	testcases := []utils.TestCase{
		{
			Name:        "error when get user access path",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.ManualUpsertStudentCourseRequest{
				StudentId: constant.StudentID,
			},
			Setup: func(ctx context.Context) {
				studentService.On("GetMapLocationAccessStudentByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
			},
		},
		{
			Name:        "happy case when none student course update ",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ManualUpsertStudentCourseRequest{
				StudentId: constant.StudentID,
				StudentCourses: []*pb.StudentCourseData{
					{
						StudentPackageId: wrapperspb.String(constant.StudentPackageID),
						CourseId:         "1",
						LocationId:       constant.LocationID,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:        false,
					},
					{
						StudentPackageId: wrapperspb.String(constant.StudentPackageID),
						CourseId:         "2",
						LocationId:       constant.LocationID,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:        false,
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{
					constant.LocationID: 1,
				}, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "duplicate course id ",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.DuplicateCourseByManualError),
			Req: &pb.ManualUpsertStudentCourseRequest{
				StudentId: constant.StudentID,
				StudentCourses: []*pb.StudentCourseData{
					{
						StudentPackageId: wrapperspb.String(constant.StudentPackageID),
						CourseId:         "1",
						LocationId:       constant.LocationID,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:        false,
					},
					{
						StudentPackageId: wrapperspb.String(constant.StudentPackageID),
						CourseId:         "1",
						LocationId:       constant.LocationID,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:        true,
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{
					constant.LocationID: 1,
				}, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "error when update time for student package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.ManualUpsertStudentCourseRequest{
				StudentId: constant.StudentID,
				StudentCourses: []*pb.StudentCourseData{
					{
						StudentPackageId: wrapperspb.String(constant.StudentPackageID),
						CourseId:         constant.CourseID,
						LocationId:       constant.LocationID,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:        true,
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs",
					mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{
					constant.LocationID: 1,
				}, nil)
				studentPackageService.On("UpdateTimeStudentPackageForManualFlow",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "error when insert for student package but user don't have access_path",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, constant.UserCantAccessThisCourse),
			Req: &pb.ManualUpsertStudentCourseRequest{
				StudentId: constant.StudentID,
				StudentCourses: []*pb.StudentCourseData{
					{
						CourseId:   constant.CourseID,
						LocationId: constant.LocationID,
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:  true,
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs",
					mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{
					constant.LocationID: 1,
				}, nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "error when insert for student package but student package service have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.ManualUpsertStudentCourseRequest{
				StudentId: constant.StudentID,
				StudentCourses: []*pb.StudentCourseData{
					{
						CourseId:   constant.CourseID,
						LocationId: constant.LocationID,
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:  true,
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs",
					mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{
					fmt.Sprintf("%v_%v", constant.LocationID, constant.StudentID): 1,
				}, nil)
				studentPackageService.On("UpsertStudentPackageForManualFlow",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "error when publish message",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.ManualUpsertStudentCourseRequest{
				StudentId: constant.StudentID,
				StudentCourses: []*pb.StudentCourseData{
					{
						CourseId:   constant.CourseID,
						LocationId: constant.LocationID,
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:  true,
					},
					{
						StudentPackageId: wrapperspb.String(constant.StudentPackageID),
						CourseId:         "2",
						LocationId:       constant.LocationID,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:        true,
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs",
					mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{
					fmt.Sprintf("%v_%v", constant.LocationID, constant.StudentID): 1,
				}, nil)
				studentPackageService.On("UpsertStudentPackageForManualFlow",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&npb.EventStudentPackage{}, nil)
				studentPackageService.On("UpdateTimeStudentPackageForManualFlow",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&npb.EventStudentPackage{}, nil)
				subscriptionService.On("PublishStudentPackage", mock.Anything, mock.Anything).Return(constant.ErrDefault)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.ManualUpsertStudentCourseRequest{
				StudentId: constant.StudentID,
				StudentCourses: []*pb.StudentCourseData{
					{
						CourseId:   constant.CourseID,
						LocationId: constant.LocationID,
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:  true,
					},
					{
						StudentPackageId: wrapperspb.String(constant.StudentPackageID),
						CourseId:         "2",
						LocationId:       constant.LocationID,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().AddDate(1, 0, 0)),
						IsChanged:        true,
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs",
					mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{
					fmt.Sprintf("%v_%v", constant.LocationID, constant.StudentID): 1,
				}, nil)
				studentPackageService.On("UpsertStudentPackageForManualFlow",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&npb.EventStudentPackage{}, nil)
				studentPackageService.On("UpdateTimeStudentPackageForManualFlow",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&npb.EventStudentPackage{}, nil)
				subscriptionService.On("PublishStudentPackage", mock.Anything, mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			studentService = new(mockServices.IStudentServiceForCourseMgMt)
			studentPackageService = new(mockServices.IStudentPackageForCourseMgMt)
			subscriptionService = new(mockServices.ISubscriptionServiceForCourseMgMte)
			testCase.Setup(testCase.Ctx)
			s := &CourseMgMt{
				DB:                  db,
				StudentPackage:      studentPackageService,
				StudentService:      studentService,
				SubscriptionService: subscriptionService,
			}

			resp, err := s.ManualUpsertStudentCourse(testCase.Ctx, testCase.Req.(*pb.ManualUpsertStudentCourseRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t)
		})
	}
}
