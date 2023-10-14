package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/course_mgmt"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImportStudentCClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                    *mockDb.Ext
		studentPackageService *mockServices.IStudentPackageForCourseMgMt
		classService          *mockServices.IClassServiceForCourseMgMt
		subscriptionService   *mockServices.ISubscriptionServiceForCourseMgMte
		tx                    *mockDb.Tx
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportStudentClassesRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, ""),
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_id
1,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - wrong name of column 3 - missing class_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'class_id'"),
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_ideee
1,1,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "invalid file - number of column != 3 - missing class_id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportStudentClassesResponse{
				Errors: []*pb.ImportStudentClassesResponse_ImportStudentClassesError{
					{
						RowNumber: 2,
						Error:     "missing mandatory data: class_id",
					},
				},
			},
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_id
1,1,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "wrong when get map student course",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_id
1,1,1`),
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, constant.ErrDefault)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "wrong when get map class",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_id
1,1,1`),
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				classService.On("GetMapClassWithLocationByClassIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.Class{}, constant.ErrDefault)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "wrong when delete student package class",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportStudentClassesResponse{
				Errors: []*pb.ImportStudentClassesResponse_ImportStudentClassesError{
					{
						RowNumber: 2,
						Error:     "error something",
					},
				},
			},
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_id
1,1,1`),
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				classService.On("GetMapClassWithLocationByClassIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.Class{}, nil)
				studentPackageService.On("DeleteStudentClass",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					[]*npb.EventStudentPackageV2{},
					[]*pb.ImportStudentClassesResponse_ImportStudentClassesError{
						{
							RowNumber: 2,
							Error:     "error something",
						},
					},
				)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "wrong when insert student package class",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportStudentClassesResponse{
				Errors: []*pb.ImportStudentClassesResponse_ImportStudentClassesError{
					{
						RowNumber: 2,
						Error:     "error something",
					},
				},
			},
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_id
1,1,1`),
				IsAddClass: true,
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				classService.On("GetMapClassWithLocationByClassIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.Class{}, nil)
				studentPackageService.On("UpsertStudentClass",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					[]*npb.EventStudentPackageV2{},
					[]*pb.ImportStudentClassesResponse_ImportStudentClassesError{
						{
							RowNumber: 2,
							Error:     "error something",
						},
					},
				)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportStudentClassesResponse{
				Errors: []*pb.ImportStudentClassesResponse_ImportStudentClassesError{},
			},
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_id
1,1,1`),
				IsAddClass: true,
			},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				classService.On("GetMapClassWithLocationByClassIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.Class{}, nil)
				studentPackageService.On("UpsertStudentClass",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					[]*npb.EventStudentPackageV2{},
					[]*pb.ImportStudentClassesResponse_ImportStudentClassesError{},
				)
				subscriptionService.On("PublishStudentClass", mock.Anything, mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Happy case when delete import class",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: &pb.ImportStudentClassesResponse{
				Errors: []*pb.ImportStudentClassesResponse_ImportStudentClassesError{},
			},
			Req: &pb.ImportStudentClassesRequest{
				Payload: []byte(`student_id,course_id,class_id
1,1,1`),
				IsAddClass: false,
			},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				classService.On("GetMapClassWithLocationByClassIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.Class{}, nil)
				studentPackageService.On("DeleteStudentClass",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					[]*npb.EventStudentPackageV2{},
					[]*pb.ImportStudentClassesResponse_ImportStudentClassesError{},
				)
				db.On("Begin", mock.Anything).Return(tx, nil)
				subscriptionService.On("PublishStudentClass", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			studentPackageService = new(mockServices.IStudentPackageForCourseMgMt)
			subscriptionService = new(mockServices.ISubscriptionServiceForCourseMgMte)
			classService = new(mockServices.IClassServiceForCourseMgMt)
			testCase.Setup(testCase.Ctx)
			s := &CourseMgMt{
				DB:                  db,
				StudentPackage:      studentPackageService,
				SubscriptionService: subscriptionService,
				ClassService:        classService,
			}

			resp, err := s.ImportStudentClasses(testCase.Ctx, testCase.Req.(*pb.ImportStudentClassesRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportStudentClassesResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t)
		})
	}
}
