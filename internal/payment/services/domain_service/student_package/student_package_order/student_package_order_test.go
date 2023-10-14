package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStudentPackageOrderPackageService_CheckTimeForStudentPackage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo

		now                = time.Now()
		startDateOfCurrent = now
		endDateOfCurrent   = now.AddDate(0, 1, 0)

		startDateOfPast = now.AddDate(0, -4, 0)
		endDateOfPast   = now.AddDate(0, -3, 0)

		startDateOfFuture = now.AddDate(0, 4, 0)
		endDateOfFuture   = now.AddDate(0, 5, 0)
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get student package orders by student_package_id",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfCurrent,
				endDateOfCurrent,
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Errorf(codes.Internal, "error when get student package orders by student_package_id = %v: %v", constant.StudentPackageID, constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when student package order is overlap",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfCurrent,
				endDateOfCurrent,
			},
			ExpectedResp: nil,
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.DuplicateCourses, &errdetails.DebugInfo{
					Detail: fmt.Sprintf("wrong start time in this student package id %v, student package end date %v, start time %v",
						constant.ErrDefault, endDateOfCurrent, startDateOfCurrent),
				},
			),
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: startDateOfCurrent,
						},
						EndAt: pgtype.Timestamptz{
							Time: endDateOfCurrent,
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: CurrentStudentPackage when there are no student package order",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfCurrent,
				endDateOfCurrent,
			},
			ExpectedResp: entities.CurrentStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{}, nil)
			},
		},
		{
			Name: "Happy case: CurrentStudentPackage when order date is within new student package time range",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfCurrent,
				endDateOfCurrent,
			},
			ExpectedResp: entities.CurrentStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: startDateOfPast,
						},
						EndAt: pgtype.Timestamptz{
							Time: endDateOfPast,
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: CurrentStudentPackage when now < newSPOTimeRange < currentSPOTTimeRange",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				now.AddDate(0, 2, 0),
				now.AddDate(0, 3, 0),
			},
			ExpectedResp: entities.CurrentStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: startDateOfFuture,
						},
						EndAt: pgtype.Timestamptz{
							Time: endDateOfFuture,
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: FutureStudentPackage when now < currentSPOTTimeRange < newSPOTimeRange",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfFuture,
				endDateOfFuture,
			},
			ExpectedResp: entities.FutureStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: now.AddDate(0, 2, 0),
						},
						EndAt: pgtype.Timestamptz{
							Time: now.AddDate(0, 3, 0),
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: FutureStudentPackage when newSPOTimeRange < now < currentSPOTimeRange",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfPast,
				endDateOfPast,
			},
			ExpectedResp: entities.PastStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: startDateOfFuture,
						},
						EndAt: pgtype.Timestamptz{
							Time: endDateOfFuture,
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: FutureStudentPackage when now ∈ currentSPOTimeRange < newSPOTimeRange ",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfFuture,
				endDateOfFuture,
			},
			ExpectedResp: entities.FutureStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: startDateOfCurrent,
						},
						EndAt: pgtype.Timestamptz{
							Time: endDateOfCurrent,
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: PastStudentPackage when newSPOTimeRange < now ∈ currentSPOTimeRange",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfPast,
				endDateOfPast,
			},
			ExpectedResp: entities.PastStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: startDateOfCurrent,
						},
						EndAt: pgtype.Timestamptz{
							Time: endDateOfCurrent,
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: PastStudentPackage when newSPOTimeRange < currentSPOTimeRange < now",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				startDateOfPast,
				endDateOfPast,
			},
			ExpectedResp: entities.PastStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: now.AddDate(0, -2, 0),
						},
						EndAt: pgtype.Timestamptz{
							Time: now.AddDate(0, -1, 0),
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: CurrentStudentPackage when currentSPOTimeRange < newSPOTimeRange < now",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				now.AddDate(0, -2, 0),
				now.AddDate(0, -1, 0),
				startDateOfPast,
				endDateOfPast,
			},
			ExpectedResp: entities.CurrentStudentPackage,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						StartAt: pgtype.Timestamptz{
							Time: startDateOfPast,
						},
						EndAt: pgtype.Timestamptz{
							Time: endDateOfPast,
						},
						IsCurrentStudentPackage: pgtype.Bool{
							Status: pgtype.Present,
							Bool:   true,
						},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageIDReq := testCase.Req.([]interface{})[0].(string)
			startTimeReq := testCase.Req.([]interface{})[1].(time.Time)
			endTimeReq := testCase.Req.([]interface{})[2].(time.Time)

			resp, err := s.GetPositionForStudentPackageByTime(testCase.Ctx, db, studentPackageIDReq, startTimeReq, endTimeReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_GetCurrentStudentPackageOrderByStudentPackageID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get get student package orders by student package id",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
			},
			ExpectedResp: &entities.StudentPackageOrder{
				ID: pgtype.Text{String: constant.StudentPackageOrderID},
				IsCurrentStudentPackage: pgtype.Bool{
					Bool:   true,
					Status: pgtype.Present,
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						ID: pgtype.Text{String: constant.StudentPackageOrderID},
						IsCurrentStudentPackage: pgtype.Bool{
							Bool:   true,
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageIDReq := testCase.Req.([]interface{})[0].(string)

			resp, err := s.GetCurrentStudentPackageOrderByStudentPackageID(testCase.Ctx, db, studentPackageIDReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_InsertStudentPackageOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when reset current position",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
				entities.CurrentStudentPackage,
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("ResetCurrentPosition", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when upsert student package order",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
				entities.FutureStudentPackage,
			},
			ExpectedResp: nil,
			ExpectedErr:  status.Errorf(codes.Internal, "insert student package order have error : %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(status.Errorf(codes.Internal, "insert student package order have error : %v", constant.ErrDefault))
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
				entities.FutureStudentPackage,
			},
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageOrderReq := testCase.Req.([]interface{})[0].(entities.StudentPackageOrder)
			positionStudentPackageOrderReq := testCase.Req.([]interface{})[1].(entities.StudentPackagePosition)

			err := s.InsertStudentPackageOrder(testCase.Ctx, db, studentPackageOrderReq, positionStudentPackageOrderReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_GetStudentPackageOrderByStudentPackageIDAndTime(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
		newStartTime            = time.Now()
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get student package orders by student package id",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				newStartTime,
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				newStartTime,
			},
			ExpectedResp: &entities.StudentPackageOrder{
				ID: pgtype.Text{String: constant.StudentPackageOrderID},
				IsCurrentStudentPackage: pgtype.Bool{
					Bool:   true,
					Status: pgtype.Present,
				},
				StartAt: pgtype.Timestamptz{Status: pgtype.Present, Time: newStartTime.AddDate(0, -1, 0)},
				EndAt:   pgtype.Timestamptz{Status: pgtype.Present, Time: newStartTime.AddDate(0, 1, 0)},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetStudentPackageOrdersByStudentPackageID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentPackageOrder{
					{
						ID: pgtype.Text{String: constant.StudentPackageOrderID},
						IsCurrentStudentPackage: pgtype.Bool{
							Bool:   true,
							Status: pgtype.Present,
						},
						StartAt: pgtype.Timestamptz{Status: pgtype.Present, Time: newStartTime.AddDate(0, -1, 0)},
						EndAt:   pgtype.Timestamptz{Status: pgtype.Present, Time: newStartTime.AddDate(0, 1, 0)},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageIDReq := testCase.Req.([]interface{})[0].(string)
			time := testCase.Req.([]interface{})[1].(time.Time)

			resp, err := s.GetStudentPackageOrderByStudentPackageIDAndTime(testCase.Ctx, db, studentPackageIDReq, time)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_GetStudentPackageOrderByStudentPackageIDAndOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get student package order by student_package_id and order_id",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				constant.OrderID,
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageID,
				constant.OrderID,
			},
			ExpectedResp: &entities.StudentPackageOrder{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetByStudentPackageIDAndOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageIDReq := testCase.Req.([]interface{})[0].(string)
			orderIDReq := testCase.Req.([]interface{})[1].(string)

			resp, err := s.GetStudentPackageOrderByStudentPackageIDAndOrderID(testCase.Ctx, db, studentPackageIDReq, orderIDReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_DeleteStudentPackageOrderByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when deleting student package order by id",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("SoftDeleteByID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("SoftDeleteByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageOrderIDReq := testCase.Req.([]interface{})[0].(string)

			err := s.DeleteStudentPackageOrderByID(testCase.Ctx, db, studentPackageOrderIDReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_RevertStudentPackageOrderByStudentPackageOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when revert student package order by id",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("RevertByID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("RevertByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageOrderIDReq := testCase.Req.([]interface{})[0].(string)

			err := s.RevertStudentPackageOrderByStudentPackageOrderID(testCase.Ctx, db, studentPackageOrderIDReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_UpdateStudentPackageOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when update student package order",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageOrderReq := testCase.Req.([]interface{})[0].(entities.StudentPackageOrder)

			err := s.UpdateStudentPackageOrder(testCase.Ctx, db, studentPackageOrderReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_UpdateExecuteStatus(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when update student package order",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("UpdateExecuteStatus", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("UpdateExecuteStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageOrderReq := testCase.Req.([]interface{})[0].(entities.StudentPackageOrder)

			err := s.UpdateExecuteStatus(testCase.Ctx, db, studentPackageOrderReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_UpdateExecuteError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when update student package order",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("UpdateExecuteError", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentPackageOrder{},
			},
			ExpectedResp: nil,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("UpdateExecuteError", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageOrderReq := testCase.Req.([]interface{})[0].(entities.StudentPackageOrder)

			err := s.UpdateExecuteError(testCase.Ctx, db, studentPackageOrderReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}

func TestStudentPackageOrderPackageService_GetStudentPackageOrderByStudentPackageOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		studentPackageOrderRepo *mockRepositories.MockStudentPackageOrderRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get student package order by student_package_id and order_id",
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
			},
			ExpectedResp: nil,
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				constant.StudentPackageOrderID,
			},
			ExpectedResp: &entities.StudentPackageOrder{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				studentPackageOrderRepo.On("GetByStudentPackageOrderID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.StudentPackageOrder{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentPackageOrderRepo = new(mockRepositories.MockStudentPackageOrderRepo)
			testCase.Setup(testCase.Ctx)

			s := &StudentPackageOrderService{
				studentPackageOrderRepo: studentPackageOrderRepo,
			}

			studentPackageOrderIDReq := testCase.Req.([]interface{})[0].(string)

			resp, err := s.GetStudentPackageOrderByStudentPackageOrderID(testCase.Ctx, db, studentPackageOrderIDReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentPackageOrderRepo)
		})
	}
}
