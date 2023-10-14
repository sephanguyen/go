package billing_schedule

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
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	FailCaseNoBillPeriodID = "Fail case: when no billPeriodID"
)

func TestIsBillingSchedulePeriodValidAndReturnBillingSchedulePeriod(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.FailedPrecondition, constant.BillItemHasNoSchedulePeriodID),
			Req: []interface{}{
				entities.Product{},
				&pb.BillingItem{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: when call billPeriodRepo",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.Product{},
				&pb.BillingItem{
					BillingSchedulePeriodId: wrapperspb.String("123"),
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return(entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: when billScheduleID in billPeriod diff to billScheduleID in productInfo",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf("in system does not match billing schedule in product"),
			Req: []interface{}{
				entities.Product{
					BillingScheduleID: pgtype.Text{String: "123", Status: pgtype.Present},
				},
				&pb.BillingItem{
					BillingSchedulePeriodId: wrapperspb.String("123"),
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{String: "1234", Status: pgtype.Present},
				}, nil)
			},
		},
		{
			Name:        "Fail case: when time range in billPeriod conflict with time range in productInfo",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf("has invalid time range"),
			Req: []interface{}{
				entities.Product{
					BillingScheduleID: pgtype.Text{String: "123", Status: pgtype.Present},
					AvailableFrom: pgtype.Timestamptz{
						Time: time.Now().AddDate(-1, 0, 0),
					},
					AvailableUntil: pgtype.Timestamptz{
						Time: time.Now(),
					},
				},
				&pb.BillingItem{
					BillingSchedulePeriodId: wrapperspb.String("123"),
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{String: "123", Status: pgtype.Present},
					StartDate: pgtype.Timestamptz{
						Time: time.Now().AddDate(0, 0, 2),
					},
					EndDate: pgtype.Timestamptz{
						Time: time.Now().AddDate(1, 0, 2),
					},
				}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				entities.Product{
					BillingScheduleID: pgtype.Text{String: "123", Status: pgtype.Present},
					AvailableFrom: pgtype.Timestamptz{
						Time: time.Now().AddDate(-1, 0, 0),
					},
					AvailableUntil: pgtype.Timestamptz{
						Time: time.Now().AddDate(1, 0, 0),
					},
				},
				&pb.BillingItem{
					BillingSchedulePeriodId: wrapperspb.String("123"),
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate",
					ctx,
					mock.Anything,
					mock.Anything,
				).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{String: "123", Status: pgtype.Present},
					StartDate: pgtype.Timestamptz{
						Time: time.Now(),
					},
					EndDate: pgtype.Timestamptz{
						Time: time.Now().AddDate(0, 1, 0),
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			productEntity := testCase.Req.([]interface{})[0].(entities.Product)
			billItem := testCase.Req.([]interface{})[1].(*pb.BillingItem)
			_, err := s.isBillingSchedulePeriodValidAndReturnBillingSchedulePeriod(testCase.Ctx, db, productEntity, billItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}

func TestIsReachLastPeriodOfSchedule(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				time.Now(),
				"1",
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: have more schedule period id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf("upcoming billing last billing schedule period not reached yet"),
			Req: []interface{}{
				time.Now(),
				"1",
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{{}}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				time.Now(),
				"1",
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			startTime := testCase.Req.([]interface{})[0].(time.Time)
			scheduleID := testCase.Req.([]interface{})[1].(string)
			err := s.isReachLastPeriodOfSchedule(testCase.Ctx, db, startTime, scheduleID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}

func TestIsContinuePeriodOfScheduleValid(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				time.Now(),
				time.Now().AddDate(1, 0, 0),
				"1",
				2,
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, constant.ErrDefault)
			},
		},
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf("of periods retrieved and number of"),
			Req: []interface{}{
				time.Now(),
				time.Now().AddDate(1, 0, 0),
				"1",
				2,
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				time.Now(),
				time.Now().AddDate(1, 0, 0),
				"1",
				2,
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{{}, {}}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			startTime := testCase.Req.([]interface{})[0].(time.Time)
			endTime := testCase.Req.([]interface{})[1].(time.Time)
			scheduleID := testCase.Req.([]interface{})[2].(string)
			lenContinuePeriodOfSchedule := testCase.Req.([]interface{})[3].(int)
			err := s.isContinuePeriodOfScheduleValid(testCase.Ctx, db, startTime, endTime, scheduleID, lenContinuePeriodOfSchedule)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}

func TestGetBillingSchedulePeriodByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				"1",
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				"1",
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			scheduleID := testCase.Req.([]interface{})[0].(string)
			_, err := s.GetBillingSchedulePeriodByID(testCase.Ctx, db, scheduleID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}

func TestIsBillingScheduleValid(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{String: "1", Status: pgtype.Present}}},
			},
			Setup: func(ctx context.Context) {
				billScheduleRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedule{}, constant.ErrDefault)
			},
		},
		{
			Name:        "False case: when period is archive",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.FailedPrecondition, "Selected billing schedule is removed or archived"),
			Req: []interface{}{
				utils.OrderItemData{ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{String: "1", Status: pgtype.Present}}},
			},
			Setup: func(ctx context.Context) {
				billScheduleRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedule{IsArchived: pgtype.Bool{
					Bool: true,
				}}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				utils.OrderItemData{ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{String: "1", Status: pgtype.Present}}},
			},
			Setup: func(ctx context.Context) {
				billScheduleRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedule{IsArchived: pgtype.Bool{
					Bool: false,
				}}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			scheduleID := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			err := s.isBillingScheduleValid(testCase.Ctx, db, scheduleID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}

func TestGetAllBillingPeriodsByBillingScheduleID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				"1",
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				"1",
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			scheduleID := testCase.Req.([]interface{})[0].(string)
			_, err := s.GetAllBillingPeriodsByBillingScheduleID(testCase.Ctx, db, scheduleID)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}

func TestCheckScheduleWithProductDisableProRating(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: when empty bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Empty bill item in request"),
			Req: []interface{}{
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.FailedPrecondition, constant.BillItemHasNoSchedulePeriodID),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: when check billPeriodID",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: billItem should be in upcoming",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This bill item should be in upcoming billing"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
						},
					},
					ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{
						String: "1",
					}},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(1, 0, 0)},
				}, nil)
			},
		},
		{
			Name:        "Fail case: billItem should be in present",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This bill item should be in at order billing"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: true,
						},
					},
					ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{
						String: "1",
					}},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
				}, nil)
			},
		},
		{
			Name:        "Fail case: billPeriod is duplicated",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.FailedPrecondition, "Bill item has duplicate billing schedule period ID"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{
						String: "1",
					}},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
				}, nil)
			},
		},
		{
			Name:        "Fail case: proRatedBillItemPeriod have invalid range",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Start date of product is outside of selected billing schedule period"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: true,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
			},
		},
		{
			Name:        "Fail case: check isReachLastPeriodOfSchedule",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: check isContinuePeriodOfScheduleValid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: when multi upcoming billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.FailedPrecondition, "Upcoming billing should only contain one item"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("3"),
							},
							IsUpcoming: true,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("4"),
							},
							IsUpcoming: true,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "3").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 5)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "4").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 4, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 4, 5)},
				}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{{}, {}}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, _, err := s.checkScheduleWithProductDisableProRating(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}

func TestCheckScheduleWithProductNoneDisableProRating(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: when empty bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Empty bill item in request"),
			Req: []interface{}{
				utils.OrderItemData{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        FailCaseNoBillPeriodID,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.FailedPrecondition, constant.BillItemHasNoSchedulePeriodID),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: when check billPeriodID",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: billItem should be in upcoming",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This bill item should be in upcoming billing"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
						},
					},
					ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{
						String: "1",
					}},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(1, 0, 0)},
				}, nil)
			},
		},
		{
			Name:        "Fail case: billItem should be in present",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This bill item should be in at order billing"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						StartDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: true,
						},
					},
					ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{
						String: "1",
					}},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
				}, nil)
			},
		},
		{
			Name:        "Fail case: billPeriod is duplicated",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.FailedPrecondition, "Bill item has duplicate billing schedule period ID"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{BillingScheduleID: pgtype.Text{
						String: "1",
					}},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
				}, nil)
			},
		},
		{
			Name:        "Fail case: proRatedBillItemPeriod have invalid range",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Start date of product is outside of selected billing schedule period"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now()),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: true,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
			},
		},
		{
			Name:        "Fail case: get ratio for proRatingBillItem",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: check isReachLastPeriodOfSchedule",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: check isContinuePeriodOfScheduleValid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: when multi upcoming billing",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.FailedPrecondition, "Upcoming billing should only contain one item"),
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("3"),
							},
							IsUpcoming: true,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("4"),
							},
							IsUpcoming: true,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "3").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 5)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "4").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, 4, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 4, 5)},
				}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{{}, {}}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, nil)
			},
		},
		{
			Name:        "HappyCaseCancel",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
								IsCancelBillItem:        wrapperspb.Bool(true),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
					},
					Order: entities.Order{
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_UPDATE.String(),
							Status: pgtype.Present,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{{}, {}}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, nil)
				billingRatioRepo.On("GetNextRatioByBillingSchedulePeriodIDAndPrevious", ctx, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, _, _, _, err := s.checkScheduleWithProductNoneDisableProRating(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}

func TestCheckScheduleReturnProRatedItemAndMapPeriodInfo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                        *mockDb.Ext
		billingRatioRepo          *mockRepositories.MockBillingRatioRepo
		billScheduleRepo          *mockRepositories.MockBillingScheduleRepo
		billingSchedulePeriodRepo *mockRepositories.MockBillingSchedulePeriodRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case for CheckScheduleReturnProRatedItemAndMapPeriodInfo with none disable pro rating product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
						DisableProRatingFlag: pgtype.Bool{Bool: false},
					},
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{{}, {}}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, nil)
			},
		},
		{
			Name:        "Happy case for disable pro rating product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				utils.OrderItemData{
					OrderItem: &pb.OrderItem{
						EffectiveDate: timestamppb.New(time.Now().AddDate(0, 3, -5)),
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("1"),
							},
							IsUpcoming: false,
						},
						{
							BillingItem: &pb.BillingItem{
								BillingSchedulePeriodId: wrapperspb.String("2"),
							},
							IsUpcoming: false,
						},
					},
					ProductInfo: entities.Product{
						BillingScheduleID: pgtype.Text{
							String: "1",
						},
						AvailableFrom: pgtype.Timestamptz{
							Time: time.Now().AddDate(-1, 0, 0),
						},
						AvailableUntil: pgtype.Timestamptz{
							Time: time.Now().AddDate(1, 0, 0),
						},
						DisableProRatingFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
					},
					IsDisableProRatingFlag: true,
				},
			},
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "1").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(-1, 0, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetByIDForUpdate", ctx, mock.Anything, "2").Once().Return(entities.BillingSchedulePeriod{
					BillingScheduleID: pgtype.Text{
						String: "1",
					},
					BillingDate: pgtype.Timestamptz{Time: time.Now().AddDate(0, -5, 0)},
					StartDate:   pgtype.Timestamptz{Time: time.Now().AddDate(0, 2, 0)},
					EndDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0)},
				}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsByScheduleIDAndStartTimeForUpdate", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{}, nil)
				billingSchedulePeriodRepo.On("GetPeriodIDsInRangeTimeByScheduleID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]pgtype.Text{{}, {}}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
			billScheduleRepo = new(mockRepositories.MockBillingScheduleRepo)
			billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
			testCase.Setup(testCase.Ctx)
			s := &BillingScheduleService{
				BillingRatioRepo:          billingRatioRepo,
				BillingScheduleRepo:       billScheduleRepo,
				BillingSchedulePeriodRepo: billingSchedulePeriodRepo,
			}
			orderItemData := testCase.Req.([]interface{})[0].(utils.OrderItemData)
			_, _, _, _, err := s.CheckScheduleReturnProRatedItemAndMapPeriodInfo(testCase.Ctx, db, orderItemData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billingRatioRepo, billScheduleRepo, billingSchedulePeriodRepo)
		})
	}
}
