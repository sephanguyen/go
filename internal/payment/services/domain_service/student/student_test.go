package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestStudentService_GetStudentAndNameByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db          *mockDb.Ext
		studentRepo *mockRepositories.MockStudentRepo
		userRepo    *mockRepositories.MockUserRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get by id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when checking student id: %v", constant.ErrDefault),
			Req:         constant.StudentID,
			Setup: func(ctx context.Context) {
				studentRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Student{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when student grade is empty",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "can't create order because this student have empty grade"),
			Req:         constant.StudentID,
			Setup: func(ctx context.Context) {
				studentRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					GradeID: pgtype.Text{
						Status: pgtype.Null,
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get student by id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when checking student id: %v", constant.ErrDefault),
			Req:         constant.StudentID,
			Setup: func(ctx context.Context) {
				studentRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					GradeID: pgtype.Text{
						Status: pgtype.Present,
					},
				}, nil)
				userRepo.On("GetStudentByIDForUpdate", ctx, db, constant.StudentID).Return(entities.User{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         constant.StudentID,
			Setup: func(ctx context.Context) {
				studentRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					GradeID: pgtype.Text{
						Status: pgtype.Present,
					},
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, nil)
				userRepo.On("GetStudentByIDForUpdate", ctx, db, constant.StudentID).Return(entities.User{
					Name: pgtype.Text{
						String: constant.StudentName,
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			userRepo = new(mockRepositories.MockUserRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentRepo: studentRepo,
				userRepo:    userRepo,
			}
			student, studentName, err := s.GetStudentAndNameByID(testCase.Ctx, db, testCase.Req.(string))

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, constant.StudentName, studentName)
				assert.Equal(t, testCase.Req.(string), student.StudentID.String)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo)
		})
	}
}

func TestStudentService_validateStudentStatusFromStudentEntity(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db          *mockDb.Ext
		studentRepo *mockRepositories.MockStudentRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when enrollment status is null",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatus, nil),
			Req: entities.Student{
				EnrollmentStatus: pgtype.Text{Status: pgtype.Null},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when enrollment status is none",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatus, nil),
			Req: entities.Student{
				EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE.String()},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when enrollment status is already enrolled",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatusAlreadyEnrolled, nil),
			Req: entities.Student{
				EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: entities.Student{
				EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			testCase.Setup(testCase.Ctx)
			err := validateStudentStatusFromStudentEntity(testCase.Req.(entities.Student))

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo)
		})
	}
}

func TestStudentService_ValidateStudentStatusForOrderTypeEnrollment(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentRepo                        *mockRepositories.MockStudentRepo
		userRepo                           *mockRepositories.MockUserRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)
	const errorHeader string = "Error when checking student enrollment status: "
	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when enrollment status is null",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatus,
				nil,
			),
			Req: entities.Student{
				EnrollmentStatus: pgtype.Text{Status: pgtype.Null},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{}, nil)
			},
		},
		{
			Name: "Fail case: Error when enrollment status is none",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatus,
				nil,
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE.String()}}, nil)
			},
		},
		{
			Name: "Fail case: Error when enrollment status is already enrolled",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusAlreadyEnrolled,
				nil,
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()}}, nil)
			},
		},
		{
			Name: "Fail case: Error when enrollment status is LOA",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusOnLOA,
				nil,
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA.String()}}, nil)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: entities.Student{
				EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentEnrollmentStatusHistory{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			userRepo = new(mockRepositories.MockUserRepo)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentRepo:                        studentRepo,
				userRepo:                           userRepo,
				studentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}
			err := s.ValidateStudentStatusForOrderType(testCase.Ctx, db, pb.OrderType_ORDER_TYPE_ENROLLMENT, testCase.Req.(entities.Student), mock.Anything, time.Now())

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo)
		})
	}
}

func TestStudentService_ValidateStudentStatusForOrderTypeWithdrawal(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentRepo                        *mockRepositories.MockStudentRepo
		userRepo                           *mockRepositories.MockUserRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)
	const errorHeader string = "Error when checking student enrollment status: "
	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when enrollment status is potential",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL),
				},
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()},
					}, nil)
			},
		},
		{
			Name: "Fail case: Error when enrollment status is withdrawn",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN),
				},
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String()},
					}, nil)
			},
		},
		{
			Name: "Fail case: Error when enrollment status is graduated",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED),
				},
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String()},
					}, nil)
			},
		},
		{
			Name: "Fail case: Error when enrollment status is temporary",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY),
				},
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String()},
					}, nil)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()},
					}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			userRepo = new(mockRepositories.MockUserRepo)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentRepo:                        studentRepo,
				userRepo:                           userRepo,
				studentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}
			err := s.ValidateStudentStatusForOrderType(testCase.Ctx, db, pb.OrderType_ORDER_TYPE_WITHDRAWAL, testCase.Req.(entities.Student), mock.Anything, time.Now())

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo)
		})
	}
}

func TestStudentService_ValidateStudentStatusForOrderTypeLOA(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentRepo                        *mockRepositories.MockStudentRepo
		userRepo                           *mockRepositories.MockUserRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when enrollment status is potential",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL),
				},
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String()},
					}, nil)
			},
		},
		{
			Name: "Fail case: Error when enrollment status is withdrawn",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN),
				},
			),
			Req: entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String()},
					}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         entities.Student{},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String()},
					}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			userRepo = new(mockRepositories.MockUserRepo)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentRepo:                        studentRepo,
				userRepo:                           userRepo,
				studentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}
			err := s.ValidateStudentStatusForOrderType(testCase.Ctx, db, pb.OrderType_ORDER_TYPE_LOA, testCase.Req.(entities.Student), mock.Anything, time.Now())

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo)
		})
	}
}

func TestStudentService_GetMapLocationAccessStudentByStudentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	t.Run("Happy case", func(t *testing.T) {
		userAccessPathRepo := &mockRepositories.MockUserAccessPathRepo{}
		userAccessPathRepo.On("GetUserAccessPathByUserIDs", ctx, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
		studentService := StudentService{userAccessRepo: userAccessPathRepo}
		db := new(mockDb.Ext)
		_, err := studentService.GetMapLocationAccessStudentByStudentIDs(ctx, db, []string{"1"})
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(t, db, userAccessPathRepo)
	})

	t.Run("Error when get from repo", func(t *testing.T) {
		userAccessPathRepo := &mockRepositories.MockUserAccessPathRepo{}
		userAccessPathRepo.On("GetUserAccessPathByUserIDs", ctx, mock.Anything, mock.Anything).Return(map[string]interface{}{}, constant.ErrDefault)
		studentService := StudentService{userAccessRepo: userAccessPathRepo}
		db := new(mockDb.Ext)
		_, err := studentService.GetMapLocationAccessStudentByStudentIDs(ctx, db, []string{"1"})
		require.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, db, userAccessPathRepo)
	})
}

func TestStudentService_IsAllowedToOrderEnrollmentRequiredProducts(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)
	const errorHeader string = "Error when checking student enrollment status: "

	studentID := uuid.New().String()
	locationID := uuid.New().String()

	testcases := []utils.TestCase{
		{
			Name:        "Happy case: empty enrollment status history",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetLatestStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentEnrollmentStatusHistory{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: enrollment status history found",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetLatestStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentEnrollmentStatusHistory{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}
			_, err := s.IsAllowedToOrderEnrollmentRequiredProducts(testCase.Ctx, db, studentID, locationID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentEnrollmentStatusHistoryRepo)
		})
	}
}

func TestStudentService_IsEnrolledWithStudentAndLocation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentRepo                        *mockRepositories.MockStudentRepo
		userRepo                           *mockRepositories.MockUserRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)
	const errorHeader string = "Error when checking student enrollment status: "
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get enrollment status",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.RetrieveStudentEnrollmentStatusByLocationRequest{
				StudentLocations: []*pb.RetrieveStudentEnrollmentStatusByLocationRequest_StudentLocation{
					{
						StudentId:  "student_id",
						LocationId: "location_id",
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: Error when get enrollment status true",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.RetrieveStudentEnrollmentStatusByLocationRequest{
				StudentLocations: []*pb.RetrieveStudentEnrollmentStatusByLocationRequest_StudentLocation{
					{
						StudentId:  "student_id",
						LocationId: "location_id",
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{
							String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String(),
						},
					}, nil)
			},
		},
		{
			Name:        "Happy case: Error when get enrollment status false",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.RetrieveStudentEnrollmentStatusByLocationRequest{
				StudentLocations: []*pb.RetrieveStudentEnrollmentStatusByLocationRequest_StudentLocation{
					{
						StudentId:  "student_id",
						LocationId: "location_id",
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						EnrollmentStatus: pgtype.Text{
							String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA.String(),
						},
					}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			userRepo = new(mockRepositories.MockUserRepo)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentRepo:                        studentRepo,
				userRepo:                           userRepo,
				studentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}
			_, err := s.IsStudentEnrolledInLocation(testCase.Ctx, db, testCase.Req.(*pb.RetrieveStudentEnrollmentStatusByLocationRequest))

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo)
		})
	}
}

func TestStudentService_GetStudentEnrolledLocationsByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentRepo                        *mockRepositories.MockStudentRepo
		userRepo                           *mockRepositories.MockUserRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)
	const errorHeader string = "Error when checking student enrollment status: "
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get enrollment status",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         "student_id",
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListStudentEnrollmentStatusHistoryByStudentID", ctx, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get current enrollment status",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req:         "student_id",
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListStudentEnrollmentStatusHistoryByStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{
						{
							LocationID: pgtype.Text{
								String: "location_id",
							},
							EnrollmentStatus: pgtype.Text{
								String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String(),
							},
							StartDate: pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
						},
					}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						LocationID: pgtype.Text{
							String: "location_id",
						},
						EnrollmentStatus: pgtype.Text{
							String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String(),
						},
						StartDate: pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: Error when get enrollment status true",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         "student_id",
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListStudentEnrollmentStatusHistoryByStudentID", ctx, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{
						{
							LocationID: pgtype.Text{
								String: "location_id",
							},
							EnrollmentStatus: pgtype.Text{
								String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String(),
							},
							StartDate: pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
						},
					}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						LocationID: pgtype.Text{
							String: "location_id",
						},
						EnrollmentStatus: pgtype.Text{
							String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
						},
						StartDate: pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					}, nil)
			},
		},
		{
			Name:        "Happy case: Error when get enrollment status false",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req:         "student_id",
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListStudentEnrollmentStatusHistoryByStudentID", ctx, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{
						{
							LocationID: pgtype.Text{
								String: "location_id",
							},
							EnrollmentStatus: pgtype.Text{
								String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String(),
							},
							StartDate: pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, -1, 0)},
						},
					}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetCurrentStatusByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					entities.StudentEnrollmentStatusHistory{
						LocationID: pgtype.Text{
							String: "location_id",
						},
						EnrollmentStatus: pgtype.Text{
							String: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
						},
						StartDate: pgtype.Timestamptz{Status: pgtype.Present, Time: time.Now().AddDate(0, 1, 0)},
					}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			userRepo = new(mockRepositories.MockUserRepo)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentRepo:                        studentRepo,
				userRepo:                           userRepo,
				studentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}
			_, err := s.GetStudentEnrolledLocationsByStudentID(testCase.Ctx, db, testCase.Req.(string))

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo)
		})
	}
}

func TestStudentService_GetEnrolledStatusInOrgByStudentInfo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentRepo                        *mockRepositories.MockStudentRepo
		userRepo                           *mockRepositories.MockUserRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
		studentProductRepo                 *mockRepositories.MockStudentProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when check enrollment status by student ID",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.GetOrgLevelStudentStatusRequest{
				StudentInfo: []*pb.GetOrgLevelStudentStatusRequestStudentInfo{
					{
						StudentId: "student_id_1",
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStudentEnrollmentStatusByStudentID", ctx, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: Error when IsEnrolledInOrg is true",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.GetOrgLevelStudentStatusRequest{
				StudentInfo: []*pb.GetOrgLevelStudentStatusRequestStudentInfo{
					{
						StudentId: "student_id_1",
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStudentEnrollmentStatusByStudentID", ctx, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when check is EnrolledInOrg",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &pb.GetOrgLevelStudentStatusRequest{
				StudentInfo: []*pb.GetOrgLevelStudentStatusRequestStudentInfo{
					{
						StudentId: "student_id_1",
						StudentProductId: &wrapperspb.StringValue{
							Value: "student_product_id_1",
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStudentEnrollmentStatusByStudentID", ctx, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{
						{
							StudentID: pgtype.Text{
								String: "student_id",
								Status: pgtype.Present,
							},
						},
					}, nil)
				studentProductRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &pb.GetOrgLevelStudentStatusRequest{
				StudentInfo: []*pb.GetOrgLevelStudentStatusRequestStudentInfo{
					{
						StudentId: "student_id_1",
						StudentProductId: &wrapperspb.StringValue{
							Value: "student_product_id_1",
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStudentEnrollmentStatusByStudentID", ctx, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{
						{
							StudentID: pgtype.Text{
								String: "student_id",
								Status: pgtype.Present,
							},
						},
					}, nil)
				studentProductRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			userRepo = new(mockRepositories.MockUserRepo)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentRepo:                        studentRepo,
				userRepo:                           userRepo,
				studentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
				studentProductRepo:                 studentProductRepo,
			}
			_, err := s.GetEnrolledStatusInOrgByStudentInfo(testCase.Ctx, db, testCase.Req.(*pb.GetOrgLevelStudentStatusRequest))

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo, userRepo, studentEnrollmentStatusHistoryRepo, studentRepo)
		})
	}
}

func TestStudentService_CheckIsEnrolledInOrgByStudentIDAndTime(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                                 *mockDb.Ext
		studentRepo                        *mockRepositories.MockStudentRepo
		userRepo                           *mockRepositories.MockUserRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
		studentProductRepo                 *mockRepositories.MockStudentProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get list enrolled status by student id and time",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				constant.StudentID,
				time.Now(),
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				constant.StudentID,
				time.Now(),
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				constant.StudentID,
				time.Now(),
			},
			Setup: func(ctx context.Context) {
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(
					[]*entities.StudentEnrollmentStatusHistory{{
						StudentID: pgtype.Text{
							String: constant.StudentID,
							Status: pgtype.Present,
						},
					}}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentRepo = new(mockRepositories.MockStudentRepo)
			userRepo = new(mockRepositories.MockUserRepo)
			studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentService{
				studentRepo:                        studentRepo,
				userRepo:                           userRepo,
				studentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
				studentProductRepo:                 studentProductRepo,
			}
			studentIDReq := testCase.Req.([]interface{})[0].(string)
			timeReq := testCase.Req.([]interface{})[1].(time.Time)
			_, err := s.CheckIsEnrolledInOrgByStudentIDAndTime(testCase.Ctx, db, studentIDReq, timeReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentRepo, userRepo, studentEnrollmentStatusHistoryRepo, studentRepo)
		})
	}
}
