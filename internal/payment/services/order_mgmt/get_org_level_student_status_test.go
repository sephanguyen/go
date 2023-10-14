package ordermgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOrgLevelStudentStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db             *mockDb.Ext
		studentService *mockServices.IGetOrgLevelStudentStatus
	)

	studentID := uuid.New().String()
	expectedResp := &pb.GetOrgLevelStudentStatusResponse{
		StudentStatus: []*pb.GetOrgLevelStudentStatusResponse_OrgLevelStudentStatus{},
	}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: error when get student enrollment status by student ID",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.GetOrgLevelStudentStatusRequest{
				StudentInfo: []*pb.GetOrgLevelStudentStatusRequestStudentInfo{
					{
						StudentId: studentID,
					},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  fmt.Errorf("error when get organization enrollment status by student info: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentService.On("GetEnrolledStatusInOrgByStudentInfo", mock.Anything, mock.Anything, mock.Anything).Return([]*pb.GetOrgLevelStudentStatusResponse_OrgLevelStudentStatus{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.GetOrgLevelStudentStatusRequest{
				StudentInfo: []*pb.GetOrgLevelStudentStatusRequestStudentInfo{
					{
						StudentId: studentID,
					},
				},
			},
			ExpectedResp: &expectedResp,
			Setup: func(ctx context.Context) {
				studentService.On("GetEnrolledStatusInOrgByStudentInfo", mock.Anything, mock.Anything, mock.Anything).Return([]*pb.GetOrgLevelStudentStatusResponse_OrgLevelStudentStatus{}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentService = new(mockServices.IGetOrgLevelStudentStatus)

			testCase.Setup(testCase.Ctx)
			s := &GetOrgLevelStudentStatus{
				DB:             db,
				StudentService: studentService,
			}
			req := testCase.Req.(*pb.GetOrgLevelStudentStatusRequest)
			_, err := s.GetOrgLevelStudentStatus(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
