package ordermgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetrieveStudentEnrollmentStatusByLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db             *mockDb.Ext
		studentService *mockServices.IRetrieveStudentEnrollmentStatusByLocation
	)

	studentID := uuid.New().String()
	locationID := uuid.New().String()
	expectedResp := &pb.RetrieveStudentEnrollmentStatusByLocationResponse{
		StudentStatusPerLocation: []*pb.RetrieveStudentEnrollmentStatusByLocationResponse_StudentStatusPerLocation{
			{
				StudentId:    "",
				LocationId:   "",
				IsEnrollment: true,
			},
		},
	}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: error when checking order permission",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveStudentEnrollmentStatusByLocationRequest{
				StudentLocations: []*pb.RetrieveStudentEnrollmentStatusByLocationRequest_StudentLocation{
					{
						StudentId:  studentID,
						LocationId: locationID,
					},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  fmt.Errorf("error when get student enrollment status: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentService.On("IsStudentEnrolledInLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*pb.RetrieveStudentEnrollmentStatusByLocationResponse_StudentStatusPerLocation{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveStudentEnrollmentStatusByLocationRequest{
				StudentLocations: []*pb.RetrieveStudentEnrollmentStatusByLocationRequest_StudentLocation{
					{
						StudentId:  studentID,
						LocationId: locationID,
					},
				},
			},
			ExpectedResp: &expectedResp,
			Setup: func(ctx context.Context) {
				studentService.On("IsStudentEnrolledInLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*pb.RetrieveStudentEnrollmentStatusByLocationResponse_StudentStatusPerLocation{
					{
						StudentId:    "",
						LocationId:   "",
						IsEnrollment: true,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentService = new(mockServices.IRetrieveStudentEnrollmentStatusByLocation)

			testCase.Setup(testCase.Ctx)
			s := &RetrieveStudentEnrollmentStatusByLocation{
				DB:             db,
				StudentService: studentService,
			}
			req := testCase.Req.(*pb.RetrieveStudentEnrollmentStatusByLocationRequest)
			resp, err := s.RetrieveStudentEnrollmentStatusByLocation(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, resp.StudentStatusPerLocation[0].IsEnrollment, expectedResp.StudentStatusPerLocation[0].IsEnrollment)
			}
		})
	}
}
