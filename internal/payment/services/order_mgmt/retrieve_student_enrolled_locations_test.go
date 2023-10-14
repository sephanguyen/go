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
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetrieveStudentEnrolledLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db             *mockDb.Ext
		studentService *mockServices.IRetrieveStudentEnrolledLocations
	)

	studentID := uuid.New().String()
	expectedResp := &pb.RetrieveStudentEnrolledLocationsResponse{
		StudentId: "",
		StudentStatusPerLocation: []*pb.RetrieveStudentEnrolledLocationsResponse_StudentStatusPerLocation{
			{
				LocationId:                           "location_id",
				StudentStatus:                        upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
				HasScheduledChangeOfStatusInLocation: true,
			},
		},
	}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: error when get student enrollment status by student ID",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveStudentEnrolledLocationsRequest{
				StudentId: studentID,
			},
			ExpectedResp: nil,
			ExpectedErr:  fmt.Errorf("error when get student enrollment status by student ID: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentEnrolledLocationsByStudentID", mock.Anything, mock.Anything, mock.Anything).Return([]*pb.RetrieveStudentEnrolledLocationsResponse_StudentStatusPerLocation{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveStudentEnrolledLocationsRequest{
				StudentId: studentID,
			},
			ExpectedResp: &expectedResp,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentEnrolledLocationsByStudentID", mock.Anything, mock.Anything, mock.Anything).Return([]*pb.RetrieveStudentEnrolledLocationsResponse_StudentStatusPerLocation{
					{
						LocationId:                           "location_id",
						StudentStatus:                        upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
						HasScheduledChangeOfStatusInLocation: true,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			studentService = new(mockServices.IRetrieveStudentEnrolledLocations)

			testCase.Setup(testCase.Ctx)
			s := &RetrieveStudentEnrolledLocations{
				DB:             db,
				StudentService: studentService,
			}
			req := testCase.Req.(*pb.RetrieveStudentEnrolledLocationsRequest)
			resp, err := s.RetrieveStudentEnrolledLocations(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, resp.StudentStatusPerLocation[0].HasScheduledChangeOfStatusInLocation, expectedResp.StudentStatusPerLocation[0].HasScheduledChangeOfStatusInLocation)
			}
		})
	}
}
