package service

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func TestImportStudentCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                    *mockDb.Ext
		studentService        *mockServices.IStudentServiceForCourseMgMt
		courseService         *mockServices.ICourseServiceForCourseMgMt
		studentPackageService *mockServices.IStudentPackageForCourseMgMt
		subscriptionService   *mockServices.ISubscriptionServiceForCourseMgMte
		tx                    *mockDb.Tx
	)

	startDate := time.Now()
	endDate := startDate.AddDate(1, 0, 0)

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportStudentCoursesRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - mismatched number of fields in header and content",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, ""),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(`student_id,course_id,location_id,start_date,end_date
1,1,1,`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of column != 5 - missing end_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 5"),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(`student_id,course_id,location_id,start_date
1,1,1,iui`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file - update student course with missing mandatory column",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(`student_id,course_id,location_id,start_date,end_date
1,1,1,,`),
			},
			ExpectedResp: &pb.ImportStudentCoursesResponse{
				Errors: []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to parse student course: missing mandatory data: start_date"),
					},
				},
			},
			Setup: func(ctx context.Context) {},
		},
		{
			Name: "parsing valid file - update student course with wrong start date column",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(`student_id,course_id,location_id,start_date,end_date
1,1,1,3232,3232332`),
			},
			ExpectedResp: &pb.ImportStudentCoursesResponse{
				Errors: []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("error parsing start_date"),
					},
				},
			},
			Setup: func(ctx context.Context) {},
		},
		{
			Name: "parsing valid file - update student course with wrong unique studentID and courseID",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(fmt.Sprintf(`student_id,course_id,location_id,start_date,end_date
1,1,1,%s,%s
1,1,1,%s,%s`, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339), startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))),
			},
			ExpectedResp: &pb.ImportStudentCoursesResponse{
				Errors: []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("duplicate student course with student id"),
					},
				},
			},
			Setup: func(ctx context.Context) {},
		},
		{
			Name: "fail when map student id with location access",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(fmt.Sprintf(`student_id,course_id,location_id,start_date,end_date
1,1,1,%s,%s`, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))),
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, constant.ErrDefault)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "fail when map course id with location access",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(fmt.Sprintf(`student_id,course_id,location_id,start_date,end_date
1,1,1,%s,%s`, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))),
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
				courseService.On("GetMapLocationAccessCourseForCourseIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, constant.ErrDefault)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "fail when map student course id with student package access",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(fmt.Sprintf(`student_id,course_id,location_id,start_date,end_date
1,1,1,%s,%s`, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))),
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
				courseService.On("GetMapLocationAccessCourseForCourseIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, constant.ErrDefault)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "fail when upsert student course",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(fmt.Sprintf(`student_id,course_id,location_id,start_date,end_date
1,1,1,%s,%s`, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))),
			},
			ExpectedResp: &pb.ImportStudentCoursesResponse{
				Errors: []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("error something"),
					},
				},
			},
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
				courseService.On("GetMapLocationAccessCourseForCourseIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentPackageService.On("UpsertStudentPackage",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*npb.EventStudentPackage{},
					[]*pb.ImportStudentCoursesResponse_ImportStudentCoursesError{
						{
							RowNumber: 2,
							Error:     fmt.Sprintf("error something"),
						},
					},
				)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportStudentCoursesRequest{
				Payload: []byte(fmt.Sprintf(`student_id,course_id,location_id,start_date,end_date
1,1,1,%s,%s`, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))),
			},
			ExpectedResp: &pb.ImportStudentCoursesResponse{
				Errors: []*pb.ImportStudentCoursesResponse_ImportStudentCoursesError{},
			},
			Setup: func(ctx context.Context) {
				tx.On("Commit", mock.Anything).Return(nil)
				studentService.On("GetMapLocationAccessStudentByStudentIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
				courseService.On("GetMapLocationAccessCourseForCourseIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
				studentPackageService.On("GetMapStudentCourseWithStudentPackageIDByIDs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.StudentPackageAccessPath{}, nil)
				studentPackageService.On("UpsertStudentPackage",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]*npb.EventStudentPackage{},
					[]*pb.ImportStudentCoursesResponse_ImportStudentCoursesError{},
				)
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
			courseService = new(mockServices.ICourseServiceForCourseMgMt)
			testCase.Setup(testCase.Ctx)
			s := &CourseMgMt{
				DB:                  db,
				StudentPackage:      studentPackageService,
				StudentService:      studentService,
				SubscriptionService: subscriptionService,
				CourseService:       courseService,
			}

			resp, err := s.ImportStudentCourses(testCase.Ctx, testCase.Req.(*pb.ImportStudentCoursesRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				expectedResp := testCase.ExpectedResp.(*pb.ImportStudentCoursesResponse)
				for i, err := range resp.Errors {
					assert.Equal(t, expectedResp.Errors[i].RowNumber, err.RowNumber)
					assert.Contains(t, err.Error, expectedResp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t)
		})
	}
}
