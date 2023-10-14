package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockKafka "github.com/manabie-com/backend/mock/golibs/kafka"
	mockNats "github.com/manabie-com/backend/mock/golibs/nats"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSubscriptionService_publishOrderEventLog(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db  *mockDb.Ext
		jsm *mockNats.JetStreamManagement
	)
	msgID := "1"
	args := []interface{}{ctx, entities.Order{}, entities.Student{}}
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when publish async context",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishOrderEventLog JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, constant.ErrDefault)),
			Req:         args,
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         args,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			jsm = &mockNats.JetStreamManagement{}
			s := &SubscriptionService{
				JSM: jsm,
			}
			testCase.Setup(testCase.Ctx)

			orderReq := testCase.Req.([]interface{})[1].(entities.Order)
			studentReq := testCase.Req.([]interface{})[2].(entities.Student)
			err := s.publishOrderEventLog(testCase.Ctx, orderReq, studentReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, jsm)
		})
	}
}

func TestSubscriptionService_publishOrderWithProductInfoLog(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db  *mockDb.Ext
		jsm *mockNats.JetStreamManagement
	)
	msgID := "1"
	args := []interface{}{ctx, entities.Order{}, []entities.StudentProduct{}}
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when publish async context",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishOrderWithProductInfoLog JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, constant.ErrDefault)),
			Req:         args,
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         args,
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			jsm = &mockNats.JetStreamManagement{}
			s := &SubscriptionService{
				JSM: jsm,
			}
			testCase.Setup(testCase.Ctx)

			orderReq := testCase.Req.([]interface{})[1].(entities.Order)
			studentProducts := testCase.Req.([]interface{})[2].([]entities.StudentProduct)

			err := s.publishOrderWithProductInfoLog(testCase.Ctx, orderReq, studentProducts)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, jsm)
		})
	}
}

func TestSubscriptionService_PublishStudentPackage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		jsm                     *mockNats.JetStreamManagement
		db                      *mockDb.Ext
		studentPackageClassRepo *mockRepositories.MockStudentPackageClassRepo
	)

	msgID := "1"
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when publish async context",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageEventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, constant.ErrDefault)),
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
				studentPackageClassRepo.On("GetByStudentPackageID", ctx, mock.Anything, mock.Anything).Return(entities.StudentPackageClass{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			jsm = &mockNats.JetStreamManagement{}
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			db = new(mockDb.Ext)
			s := &SubscriptionService{
				JSM:                     jsm,
				DB:                      db,
				StudentPackageClassRepo: studentPackageClassRepo,
			}
			testCase.Setup(testCase.Ctx)
			err := s.PublishStudentPackage(testCase.Ctx, []*npb.EventStudentPackage{{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					Package: &npb.EventStudentPackage_Package{
						CourseIds:   []string{""},
						LocationIds: []string{""},
					},
				},
				LocationIds: []string{""},
			}})

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, jsm)
		})
	}
}

func TestSubscriptionService_publishStudentCourseSyncEvent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db  *mockDb.Ext
		jsm *mockNats.JetStreamManagement
	)
	msgID := "1"
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when publish async context",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentCourseEventSync JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, constant.ErrDefault)),
			Req:         []*pb.EventSyncStudentPackageCourse{},
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         []*pb.EventSyncStudentPackageCourse{},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			jsm = &mockNats.JetStreamManagement{}
			s := &SubscriptionService{
				JSM: jsm,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.([]*pb.EventSyncStudentPackageCourse)
			err := s.publishStudentCourseSyncEvent(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, jsm)
		})
	}
}

func TestSubscriptionService_PublishStudentPackageForCreateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		jsm                     *mockNats.JetStreamManagement
		db                      *mockDb.Ext
		studentPackageClassRepo *mockRepositories.MockStudentPackageClassRepo
	)
	msgID := "1"
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when publish async context",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageEventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, constant.ErrDefault)),
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
				studentPackageClassRepo.On("GetByStudentPackageID", ctx, mock.Anything, mock.Anything).Return(entities.StudentPackageClass{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			jsm = &mockNats.JetStreamManagement{}
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			db = new(mockDb.Ext)
			s := &SubscriptionService{
				JSM:                     jsm,
				StudentPackageClassRepo: studentPackageClassRepo,
				DB:                      db,
			}
			testCase.Setup(testCase.Ctx)
			err := s.PublishStudentPackageForCreateOrder(testCase.Ctx, []*npb.EventStudentPackage{{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					Package: &npb.EventStudentPackage_Package{
						CourseIds:   []string{""},
						LocationIds: []string{""},
					},
				},
				LocationIds: []string{""},
			}})

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, jsm)
		})
	}
}

func TestSubscriptionService_PublishStudentClass(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		jsm *mockNats.JetStreamManagement
	)
	msgID := "1"
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when publish async context",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nats.HandlePushMsgFail(ctx, fmt.Errorf("SubjectStudentPackageV2EventNats JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, constant.ErrDefault)),
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return(msgID, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			jsm = &mockNats.JetStreamManagement{}
			s := &SubscriptionService{
				JSM: jsm,
			}
			testCase.Setup(testCase.Ctx)
			err := s.PublishStudentClass(testCase.Ctx, []*npb.EventStudentPackageV2{{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					Package: &npb.EventStudentPackageV2_PackageV2{},
				},
			}})

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, jsm)
		})
	}
}

func TestSubscriptionService_PublishNotification(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		kafka                *mockKafka.KafkaManagement
		userRepo             *mockRepositories.MockUserRepo
		orderRepo            *mockRepositories.MockOrderRepo
		notificationDateRepo *mockRepositories.MockNotificationDateRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when publishing context",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf("error when publish failed kafka PublishNotification: %v", constant.ErrDefault),
			Req: []interface{}{
				&payload.UpsertSystemNotification{
					ReferenceID: constant.OrderID,
					Content:     nil,
					URL:         constant.StudentDetailPath,
					ValidFrom:   time.Time{},
					Recipients:  nil,
					IsDeleted:   false,
				},
			},
			Setup: func(ctx context.Context) {
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&payload.UpsertSystemNotification{
					ReferenceID: constant.OrderID,
					Content:     nil,
					URL:         constant.StudentDetailPath,
					ValidFrom:   time.Time{},
					Recipients:  nil,
					IsDeleted:   false,
				},
			},
			Setup: func(ctx context.Context) {
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			Name:        "must not pass if missing reference_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				&payload.UpsertSystemNotification{},
			},
			Setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			kafka = &mockKafka.KafkaManagement{}
			userRepo = new(mockRepositories.MockUserRepo)
			orderRepo = new(mockRepositories.MockOrderRepo)
			notificationDateRepo = new(mockRepositories.MockNotificationDateRepo)
			s := &SubscriptionService{
				Kafka: kafka,
				Config: configs.CommonConfig{
					Environment: "local",
				},
				NotificationDateRepo: notificationDateRepo,
				userRepo:             userRepo,
				orderRepo:            orderRepo,
			}
			testCase.Setup(testCase.Ctx)
			upsertNotificationData := testCase.Req.([]interface{})[0].(*payload.UpsertSystemNotification)
			err := s.PublishNotification(testCase.Ctx, upsertNotificationData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, kafka, userRepo, orderRepo, notificationDateRepo)
		})
	}
}

func TestSubscriptionService_PublishNotification_ProductionEnvironment(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		kafka                *mockKafka.KafkaManagement
		userRepo             *mockRepositories.MockUserRepo
		orderRepo            *mockRepositories.MockOrderRepo
		notificationDateRepo *mockRepositories.MockNotificationDateRepo
	)
	testcases := []utils.TestCase{
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&payload.UpsertSystemNotification{
					ReferenceID: constant.OrderID,
					Content: []payload.SystemNotificationContent{
						{
							Language: "en",
							Text:     "text_en",
						},
						{
							Language: "vi",
							Text:     "text_vi",
						},
					},
					URL:        constant.StudentDetailPath,
					ValidFrom:  time.Now(),
					Recipients: []payload.SystemNotificationRecipient{},
					IsDeleted:  false,
				},
			},
			Setup: func(ctx context.Context) {
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			kafka = &mockKafka.KafkaManagement{}
			userRepo = new(mockRepositories.MockUserRepo)
			orderRepo = new(mockRepositories.MockOrderRepo)
			notificationDateRepo = new(mockRepositories.MockNotificationDateRepo)
			s := &SubscriptionService{
				Kafka: kafka,
				Config: configs.CommonConfig{
					Environment: "prod",
				},
				NotificationDateRepo: notificationDateRepo,
				userRepo:             userRepo,
				orderRepo:            orderRepo,
			}
			testCase.Setup(testCase.Ctx)
			upsertNotificationData := testCase.Req.([]interface{})[0].(*payload.UpsertSystemNotification)
			err := s.PublishNotification(testCase.Ctx, upsertNotificationData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, kafka)
		})
	}
}

func TestSubscriptionService_Publish(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		kafka                   *mockKafka.KafkaManagement
		userRepo                *mockRepositories.MockUserRepo
		orderRepo               *mockRepositories.MockOrderRepo
		notificationDateRepo    *mockRepositories.MockNotificationDateRepo
		db                      *mockDb.Ext
		jsm                     *mockNats.JetStreamManagement
		studentPackageClassRepo *mockRepositories.MockStudentPackageClassRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when publishing order event log",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.MessageSyncData{
					OrderType:                 0,
					Order:                     entities.Order{},
					Student:                   entities.Student{},
					StudentCourseMessage:      nil,
					StudentPackages:           nil,
					StudentProducts:           nil,
					SystemNotificationMessage: nil,
				},
			},
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Return("msg-id", constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when publishing order with product info log",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.MessageSyncData{
					OrderType:                 0,
					Order:                     entities.Order{},
					Student:                   entities.Student{},
					StudentCourseMessage:      nil,
					StudentPackages:           nil,
					StudentProducts:           nil,
					SystemNotificationMessage: nil,
				},
			},
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("msg-id", nil)
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("msg-id", constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when publishing notification",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.MessageSyncData{
					OrderType: 0,
					Order: entities.Order{
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_RESUME.String(),
							Status: pgtype.Present,
						},
					},
					Student:              entities.Student{},
					StudentCourseMessage: nil,
					StudentPackages:      nil,
					StudentProducts:      nil,
					SystemNotificationMessage: &payload.UpsertSystemNotification{
						ReferenceID: constant.OrderID,
						Content:     nil,
						URL:         "",
						ValidFrom:   time.Time{},
						Recipients:  nil,
						IsDeleted:   false,
						Status:      "",
					},
				},
			},
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("msg-id", nil)
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("msg-id", nil)
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when publishing student package for create order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.MessageSyncData{
					OrderType: 0,
					Order: entities.Order{
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_RESUME.String(),
							Status: pgtype.Present,
						},
					},
					Student:              entities.Student{},
					StudentCourseMessage: nil,
					StudentPackages: []*npb.EventStudentPackage{{
						StudentPackage: &npb.EventStudentPackage_StudentPackage{
							Package: &npb.EventStudentPackage_Package{
								CourseIds:   []string{""},
								LocationIds: []string{""},
							},
						},
						LocationIds: []string{""},
					}},
					StudentProducts: nil,
					SystemNotificationMessage: &payload.UpsertSystemNotification{
						ReferenceID: constant.OrderID,
						Content:     nil,
						URL:         "",
						ValidFrom:   time.Time{},
						Recipients:  nil,
						IsDeleted:   false,
						Status:      "",
					},
				},
			},
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("msg-id", nil)
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("msg-id", nil)
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return("msg-id", nil)
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return("msg-id", nil)
				studentPackageClassRepo.On("GetByStudentPackageID", ctx, mock.Anything, mock.Anything).Return(entities.StudentPackageClass{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.MessageSyncData{
					OrderType: 0,
					Order: entities.Order{
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_RESUME.String(),
							Status: pgtype.Present,
						},
					},
					Student:              entities.Student{},
					StudentCourseMessage: nil,
					StudentPackages: []*npb.EventStudentPackage{{
						StudentPackage: &npb.EventStudentPackage_StudentPackage{
							Package: &npb.EventStudentPackage_Package{
								CourseIds:   []string{""},
								LocationIds: []string{""},
							},
						},
						LocationIds: []string{""},
					}},
					StudentProducts: nil,
					SystemNotificationMessage: &payload.UpsertSystemNotification{
						ReferenceID: constant.OrderID,
						Content:     nil,
						URL:         "",
						ValidFrom:   time.Time{},
						Recipients:  nil,
						IsDeleted:   false,
						Status:      "",
					},
				},
			},
			Setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("msg-id", nil)
				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("msg-id", nil)
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return("msg-id", nil)
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).Return("msg-id", nil)
				studentPackageClassRepo.On("GetByStudentPackageID", ctx, mock.Anything, mock.Anything).Return(entities.StudentPackageClass{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			kafka = &mockKafka.KafkaManagement{}
			db = new(mockDb.Ext)
			userRepo = new(mockRepositories.MockUserRepo)
			orderRepo = new(mockRepositories.MockOrderRepo)
			notificationDateRepo = new(mockRepositories.MockNotificationDateRepo)
			jsm = &mockNats.JetStreamManagement{}
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			s := &SubscriptionService{
				DB:    db,
				Kafka: kafka,
				Config: configs.CommonConfig{
					Environment: "local",
				},
				NotificationDateRepo:    notificationDateRepo,
				userRepo:                userRepo,
				orderRepo:               orderRepo,
				JSM:                     jsm,
				StudentPackageClassRepo: studentPackageClassRepo,
			}
			testCase.Setup(testCase.Ctx)
			messageData := testCase.Req.([]interface{})[0].(utils.MessageSyncData)
			err := s.Publish(testCase.Ctx, db, messageData)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, kafka, userRepo, orderRepo, notificationDateRepo)
		})
	}
}

func TestSubscriptionService_ToNotificationMessage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		kafka                   *mockKafka.KafkaManagement
		userRepo                *mockRepositories.MockUserRepo
		orderRepo               *mockRepositories.MockOrderRepo
		notificationDateRepo    *mockRepositories.MockNotificationDateRepo
		gradeRepo               *mockRepositories.MockGradeRepo
		db                      *mockDb.Ext
		jsm                     *mockNats.JetStreamManagement
		studentPackageClassRepo *mockRepositories.MockStudentPackageClassRepo
	)
	testcases := []utils.TestCase{
		{
			Name: "Happy case: order_type == resume and order_status = voided",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Order{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
					OrderStatus: pgtype.Text{
						String: pb.OrderStatus_ORDER_STATUS_VOIDED.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: &payload.UpsertSystemNotification{
				ReferenceID: constant.OrderID,
				IsDeleted:   true,
			},
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Happy case: order_type == loa and order_status = submmited",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Order{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
					OrderStatus: pgtype.Text{
						String: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: &payload.UpsertSystemNotification{
				ReferenceID: constant.OrderID,
				IsDeleted:   true,
			},
			Setup: func(ctx context.Context) {
				notificationDateRepo.On("GetByOrderType", mock.Anything, mock.Anything, mock.Anything).Return(entities.NotificationDate{
					NotificationDate: pgtype.Int4{
						Int: 15,
					},
				}, nil)
				gradeRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Grade{
					Name: constant.DefaultGrade,
				}, nil)
				userRepo.On("GetUserIDsByRoleNamesAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{
					constant.UserID,
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			kafka = &mockKafka.KafkaManagement{}
			db = new(mockDb.Ext)
			userRepo = new(mockRepositories.MockUserRepo)
			orderRepo = new(mockRepositories.MockOrderRepo)
			notificationDateRepo = new(mockRepositories.MockNotificationDateRepo)
			gradeRepo = new(mockRepositories.MockGradeRepo)
			jsm = &mockNats.JetStreamManagement{}
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			s := &SubscriptionService{
				DB:    db,
				Kafka: kafka,
				Config: configs.CommonConfig{
					Environment: "local",
				},
				NotificationDateRepo:    notificationDateRepo,
				userRepo:                userRepo,
				orderRepo:               orderRepo,
				JSM:                     jsm,
				StudentPackageClassRepo: studentPackageClassRepo,
				gradeRepo:               gradeRepo,
			}
			testCase.Setup(testCase.Ctx)
			orderReq := testCase.Req.([]interface{})[0].(entities.Order)
			studentReq := testCase.Req.([]interface{})[1].(entities.Student)
			notificationDataReq := testCase.Req.([]interface{})[2].(utils.UpsertSystemNotificationData)
			_, err := s.ToNotificationMessage(testCase.Ctx, db, orderReq, studentReq, notificationDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, kafka, userRepo, orderRepo, notificationDateRepo, gradeRepo)
		})
	}
}

func TestSubscriptionService_ToNotificationMessageForCreate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		kafka                   *mockKafka.KafkaManagement
		userRepo                *mockRepositories.MockUserRepo
		orderRepo               *mockRepositories.MockOrderRepo
		notificationDateRepo    *mockRepositories.MockNotificationDateRepo
		gradeRepo               *mockRepositories.MockGradeRepo
		db                      *mockDb.Ext
		jsm                     *mockNats.JetStreamManagement
		studentPackageClassRepo *mockRepositories.MockStudentPackageClassRepo
		loaEndDate              = time.Now().AddDate(0, 1, 0)
		loc                     = time.FixedZone("", 0)
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when missing student detail path",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when missing student detail path"),
			Req: []interface{}{
				entities.Order{
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: "",
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        "Fail case: Error when getting notification date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when get notification date by order type: %v", constant.ErrDefault),
			Req: []interface{}{
				entities.Order{
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
				notificationDateRepo.On("GetByOrderType", mock.Anything, mock.Anything, mock.Anything).Return(entities.NotificationDate{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when getting grade by grade_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.Order{
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
				notificationDateRepo.On("GetByOrderType", mock.Anything, mock.Anything, mock.Anything).Return(entities.NotificationDate{}, nil)
				gradeRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Grade{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when getting userIDs by role names and location id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.Order{
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
				notificationDateRepo.On("GetByOrderType", mock.Anything, mock.Anything, mock.Anything).Return(entities.NotificationDate{}, nil)
				gradeRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Grade{}, nil)
				userRepo.On("GetUserIDsByRoleNamesAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: order_type == LOA",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				entities.Order{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
					StudentFullName: pgtype.Text{
						String: constant.StudentName,
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           loaEndDate,
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: &payload.UpsertSystemNotification{
				ReferenceID: constant.OrderID,
				Content: []payload.SystemNotificationContent{
					{
						Language: string(utils.EnCode),
						Text:     fmt.Sprintf(utils.EngLOANotificationContentTemp, constant.DefaultGrade.String, constant.StudentName, constant.LocationName),
					},
					{
						Language: string(utils.JpCode),
						Text:     fmt.Sprintf(utils.JpLOANotificationContentTemp, constant.LocationName, constant.DefaultGrade.String, constant.StudentName),
					},
				},
				URL:       constant.StudentDetailPath,
				ValidFrom: time.Date(loaEndDate.Year(), loaEndDate.Month()-1, 15, 0, 0, 0, 0, loc),
				Recipients: []payload.SystemNotificationRecipient{
					{
						UserID: constant.UserID,
					},
				},
				IsDeleted: false,
			},
			Setup: func(ctx context.Context) {
				notificationDateRepo.On("GetByOrderType", mock.Anything, mock.Anything, mock.Anything).Return(entities.NotificationDate{
					NotificationDate: pgtype.Int4{
						Int: 15,
					},
				}, nil)
				gradeRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Grade{
					Name: constant.DefaultGrade,
				}, nil)
				userRepo.On("GetUserIDsByRoleNamesAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{
					constant.UserID,
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get latest LOA order by student id and location id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				entities.Order{
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_RESUME.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: &payload.UpsertSystemNotification{},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetLatestOrderByStudentIDAndLocationIDAndOrderType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: order_type == resume",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Order{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_RESUME.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: &payload.UpsertSystemNotification{
				ReferenceID: constant.OrderID,
				IsDeleted:   true,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetLatestOrderByStudentIDAndLocationIDAndOrderType", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					LOAEndDate: pgtype.Timestamptz{
						Time: time.Now().AddDate(0, 1, 1),
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: order_type == withdraw",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				entities.Order{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
					},
				},
				entities.Student{CurrentGrade: pgtype.Int2{
					Int:    10,
					Status: pgtype.Present,
				}},
				utils.UpsertSystemNotificationData{
					StudentDetailPath: constant.StudentDetailPath,
					LocationName:      constant.LocationName,
					EndDate:           time.Now().AddDate(0, 1, 0),
					StartDate:         time.Now(),
				},
			},
			ExpectedResp: &payload.UpsertSystemNotification{
				ReferenceID: constant.OrderID,
				IsDeleted:   true,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetLatestOrderByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					LOAEndDate: pgtype.Timestamptz{
						Time: time.Now().AddDate(0, 1, 1),
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
					OrderStatus: pgtype.Text{
						String: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			kafka = &mockKafka.KafkaManagement{}
			db = new(mockDb.Ext)
			userRepo = new(mockRepositories.MockUserRepo)
			orderRepo = new(mockRepositories.MockOrderRepo)
			notificationDateRepo = new(mockRepositories.MockNotificationDateRepo)
			gradeRepo = new(mockRepositories.MockGradeRepo)
			jsm = &mockNats.JetStreamManagement{}
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			s := &SubscriptionService{
				DB:    db,
				Kafka: kafka,
				Config: configs.CommonConfig{
					Environment: "local",
				},
				NotificationDateRepo:    notificationDateRepo,
				userRepo:                userRepo,
				orderRepo:               orderRepo,
				JSM:                     jsm,
				StudentPackageClassRepo: studentPackageClassRepo,
				gradeRepo:               gradeRepo,
			}
			testCase.Setup(testCase.Ctx)
			orderReq := testCase.Req.([]interface{})[0].(entities.Order)
			studentReq := testCase.Req.([]interface{})[1].(entities.Student)
			notificationDataReq := testCase.Req.([]interface{})[2].(utils.UpsertSystemNotificationData)
			resp, err := s.toNotificationMessageForCreate(testCase.Ctx, db, orderReq, studentReq, notificationDataReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp.(*payload.UpsertSystemNotification), resp)
			}

			mock.AssertExpectationsForObjects(t, kafka, userRepo, orderRepo, notificationDateRepo, gradeRepo)
		})
	}
}

func TestSubscriptionService_ToNotificationMessageForVoid(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		kafka                   *mockKafka.KafkaManagement
		userRepo                *mockRepositories.MockUserRepo
		orderRepo               *mockRepositories.MockOrderRepo
		notificationDateRepo    *mockRepositories.MockNotificationDateRepo
		gradeRepo               *mockRepositories.MockGradeRepo
		db                      *mockDb.Ext
		jsm                     *mockNats.JetStreamManagement
		studentPackageClassRepo *mockRepositories.MockStudentPackageClassRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				entities.Order{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_LOA.String(),
					},
					OrderStatus: pgtype.Text{
						String: pb.OrderStatus_ORDER_STATUS_VOIDED.String(),
					},
				},
			},
			ExpectedResp: &payload.UpsertSystemNotification{
				ReferenceID: constant.OrderID,
				IsDeleted:   true,
			},
			Setup: func(ctx context.Context) {},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			kafka = &mockKafka.KafkaManagement{}
			db = new(mockDb.Ext)
			userRepo = new(mockRepositories.MockUserRepo)
			orderRepo = new(mockRepositories.MockOrderRepo)
			notificationDateRepo = new(mockRepositories.MockNotificationDateRepo)
			gradeRepo = new(mockRepositories.MockGradeRepo)
			jsm = &mockNats.JetStreamManagement{}
			studentPackageClassRepo = new(mockRepositories.MockStudentPackageClassRepo)
			s := &SubscriptionService{
				DB:    db,
				Kafka: kafka,
				Config: configs.CommonConfig{
					Environment: "local",
				},
				NotificationDateRepo:    notificationDateRepo,
				userRepo:                userRepo,
				orderRepo:               orderRepo,
				JSM:                     jsm,
				StudentPackageClassRepo: studentPackageClassRepo,
				gradeRepo:               gradeRepo,
			}
			testCase.Setup(testCase.Ctx)
			orderReq := testCase.Req.([]interface{})[0].(entities.Order)
			resp := s.toNotificationMessageForVoid(testCase.Ctx, orderReq)

			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedResp.(*payload.UpsertSystemNotification), resp)
			} else {
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, kafka, userRepo, orderRepo, notificationDateRepo, gradeRepo)
		})
	}
}
