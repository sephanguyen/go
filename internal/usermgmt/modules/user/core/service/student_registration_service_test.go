package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	pbu "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func StudentRegistrationServiceMock() (prepareStudentRegistrationServiceMock, StudentRegistrationService) {
	m := prepareStudentRegistrationServiceMock{
		&mock_database.Ext{},
		&mock_database.Tx{},
		&mock_repositories.MockDomainStudentRepo{},
		&mock_repositories.MockDomainLocationRepo{},
		&mock_repositories.MockDomainUserAccessPathRepo{},
		&mock_repositories.MockDomainEnrollmentStatusHistoryRepo{},
	}

	service := StudentRegistrationService{
		DB:                                m.db,
		StudentRepo:                       m.studentRepo,
		LocationRepo:                      m.locationRepo,
		DomainEnrollmentStatusHistoryRepo: m.enrollmentStatusHistoryRepo,
		DomainUserAccessPathRepo:          m.userAccessPathRepo,
	}

	return m, service
}

type prepareStudentRegistrationServiceMock struct {
	db *mock_database.Ext
	tx *mock_database.Tx

	studentRepo                 *mock_repositories.MockDomainStudentRepo
	locationRepo                *mock_repositories.MockDomainLocationRepo
	userAccessPathRepo          *mock_repositories.MockDomainUserAccessPathRepo
	enrollmentStatusHistoryRepo *mock_repositories.MockDomainEnrollmentStatusHistoryRepo
}

type MockDomainEnrollmentStatusHistory struct {
	userID              field.String
	locationID          field.String
	enrollmentStatus    field.String
	startDate           field.Time
	endDate             field.Time
	orderID             field.String
	orderSequenceNumber field.Int32

	entity.DefaultDomainEnrollmentStatusHistory
}

func createMockDomainEnrollmentStatusHistory(studentID, locationID, enrollmentStatus string, startDate, endDate time.Time, orderID string, orderSequenceNumber int32) entity.DomainEnrollmentStatusHistory {
	return &MockDomainEnrollmentStatusHistory{
		userID:              field.NewString(studentID),
		locationID:          field.NewString(locationID),
		enrollmentStatus:    field.NewString(enrollmentStatus),
		startDate:           field.NewTime(startDate),
		endDate:             field.NewTime(endDate),
		orderID:             field.NewString(orderID),
		orderSequenceNumber: field.NewInt32(orderSequenceNumber),
	}
}

func (m *MockDomainEnrollmentStatusHistory) UserID() field.String {
	return m.userID
}

func (m *MockDomainEnrollmentStatusHistory) LocationID() field.String {
	return m.locationID
}

func (m *MockDomainEnrollmentStatusHistory) EnrollmentStatus() field.String {
	return m.enrollmentStatus
}

func (m *MockDomainEnrollmentStatusHistory) StartDate() field.Time {
	return m.startDate
}

func (m *MockDomainEnrollmentStatusHistory) EndDate() field.Time {
	return m.endDate
}

func (m *MockDomainEnrollmentStatusHistory) OrderID() field.String {
	return m.orderID
}

func (m *MockDomainEnrollmentStatusHistory) OrderSequenceNumber() field.Int32 {
	return m.orderSequenceNumber
}

var _ OrderFlowEnrollmentStatusManager = (*MockHandelOrderFlowEnrollmentStatusManager)(nil)

type MockHandelOrderFlowEnrollmentStatusManager struct {
	handleOrderFlowForVoidEnrollmentStatus func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error)
	handleEnrollmentStatusUpdate           func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error)
	handleOrderFlowForTheExistedLocations  func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error)
	handleOrderFlowForTheNewLocation       func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error)
}

func (m *MockHandelOrderFlowEnrollmentStatusManager) HandleVoidEnrollmentStatus(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
	return m.handleOrderFlowForVoidEnrollmentStatus(ctx, db, req)
}

func (m *MockHandelOrderFlowEnrollmentStatusManager) HandleEnrollmentStatusUpdate(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
	return m.handleEnrollmentStatusUpdate(ctx, db, req)
}

func (m *MockHandelOrderFlowEnrollmentStatusManager) HandleExistedLocations(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
	return m.handleOrderFlowForTheExistedLocations(ctx, db, req)
}

func (m *MockHandelOrderFlowEnrollmentStatusManager) HandleForNewLocation(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
	return m.handleOrderFlowForTheNewLocation(ctx, db, req)
}

func mockEnrollmentStatusHistoryStartDateModifier() EnrollmentStatusHistoryStartDateModifierFn {
	return func(_ context.Context, _ libdatabase.Ext, _ DomainEnrollmentStatusHistoryRepo, history entity.DomainEnrollmentStatusHistory) (entity.DomainEnrollmentStatusHistory, error) {
		return history, nil
	}
}

func TestStudentRegistrationService_SyncOrderLogHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	domainEnrollmentStatusHistoryRepo := &mock_repositories.MockDomainEnrollmentStatusHistoryRepo{}
	domainUserAccessPathRepo := &mock_repositories.MockDomainUserAccessPathRepo{}
	unleashClientInstance := &mock_unleash_client.UnleashClientInstance{}
	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)

	zapLogger := ctxzap.Extract(ctx)

	service := StudentRegistrationService{
		Logger:                            zapLogger,
		DB:                                db,
		DomainEnrollmentStatusHistoryRepo: domainEnrollmentStatusHistoryRepo,
		DomainUserAccessPathRepo:          domainUserAccessPathRepo,
		UnleashClient:                     unleashClientInstance,
	}
	service.OrderFlowEnrollmentStatusManager = &HandelOrderFlowEnrollmentStatus{
		Logger:                            service.Logger,
		DomainEnrollmentStatusHistoryRepo: service.DomainEnrollmentStatusHistoryRepo,
		DomainUserAccessPathRepo:          service.DomainUserAccessPathRepo,
		SyncEnrollmentStatusHistory:       service.SyncEnrollmentStatusHistory,
		DeactivateAndReactivateStudents:   service.DeactivateAndReactivateStudents,
	}
	service.EnrollmentStatusHistoryStartDateModifier = mockEnrollmentStatusHistoryStartDateModifier()

	testCases := []TestCase{
		{
			name: "happy case: create new enrollment status",
			ctx:  ctx,
			req: &OrderEventLog{
				OrderStatus:      pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
				OrderType:        pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
				StudentID:        "student-id",
				LocationID:       "Manabie",
				EnrollmentStatus: pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
				StartDate:        time.Now(),
				EndDate:          time.Now().Add(86 * time.Hour),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				unleashClientInstance.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleAutoDeactivateAndReactivateStudentsV2, mock.Anything, mock.Anything).Once().Return(true, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr:  nil,
			expectedResp: false,
		},
		{
			name: "worst case: create new enrollment status with invalid enrollment status",
			ctx:  ctx,
			req: &OrderEventLog{
				OrderStatus:      pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
				OrderType:        pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
				StudentID:        "student-id",
				LocationID:       "Manabie",
				EnrollmentStatus: "invalid",
				StartDate:        time.Now(),
				EndDate:          time.Now().Add(86 * time.Hour),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				unleashClientInstance.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleAutoDeactivateAndReactivateStudentsV2, mock.Anything, mock.Anything).Once().Return(true, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr:  nil,
			expectedResp: false,
		},
		{
			name: "worst case: create new enrollment status with invalid start date after end date",
			ctx:  ctx,
			req: &OrderEventLog{
				OrderStatus:      pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
				OrderType:        pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
				StudentID:        "student-id",
				LocationID:       "Manabie",
				EnrollmentStatus: pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
				StartDate:        time.Now().Add(86 * time.Hour),
				EndDate:          time.Now(),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				unleashClientInstance.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleAutoDeactivateAndReactivateStudentsV2, mock.Anything, mock.Anything).Once().Return(true, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr:  nil,
			expectedResp: false,
		},
		{
			name: "happy case: create new enrollment status with order status submitted and order type new and temporary status",
			ctx:  ctx,
			req: &OrderEventLog{
				OrderStatus:      pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
				OrderType:        pb.OrderType_ORDER_TYPE_NEW.String(),
				StudentID:        "student-id",
				LocationID:       "Manabie",
				EnrollmentStatus: pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
				StartDate:        time.Now().Add(-48 * time.Hour),
				EndDate:          time.Now().Add(240 * time.Hour),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				unleashClientInstance.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleAutoDeactivateAndReactivateStudentsV2, mock.Anything, mock.Anything).Once().Return(true, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr:  nil,
			expectedResp: false,
		},
		{
			name: "happy case: create new enrollment status with order status submitted and order type new and not temporary status",
			ctx:  ctx,
			req: &OrderEventLog{
				OrderStatus:      pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
				OrderType:        pb.OrderType_ORDER_TYPE_NEW.String(),
				StudentID:        "student-id",
				LocationID:       "Manabie",
				EnrollmentStatus: pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
				StartDate:        time.Now().Add(-48 * time.Hour),
				EndDate:          time.Now().Add(240 * time.Hour),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Twice().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				unleashClientInstance.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleAutoDeactivateAndReactivateStudentsV2, mock.Anything, mock.Anything).Once().Return(true, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr:  nil,
			expectedResp: false,
		},
		{
			name: "happy case: create new enrollment status with order status voided and order type enrollment/withdrawal/graduate/loa",
			ctx:  ctx,
			req: &OrderEventLog{
				OrderStatus:      pb.OrderStatus_ORDER_STATUS_VOIDED.String(),
				OrderType:        pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
				StudentID:        "student-id",
				LocationID:       "Manabie",
				EnrollmentStatus: pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
				StartDate:        time.Now().Add(-48 * time.Hour),
				EndDate:          time.Now().Add(240 * time.Hour),
				OrderID:          "order-id",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainEnrollmentStatusHistoryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainUserAccessPathRepo.On("UpsertMultiple", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entity.DomainEnrollmentStatusHistories{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("GetLatestEnrollmentStudentOfLocation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]entity.DomainEnrollmentStatusHistory{
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
					createMockDomainEnrollmentStatusHistory("student-id", "Manabie",
						pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String(),
						time.Now().Add(-40*time.Hour),
						time.Now().Add(200*time.Hour),
						"order-id",
						1,
					),
				}, nil)
				domainEnrollmentStatusHistoryRepo.On("SoftDeleteEnrollments", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("UpdateStudentStatusBasedEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				domainEnrollmentStatusHistoryRepo.On("DeactivateEnrollmentStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				unleashClientInstance.On("IsFeatureEnabledOnOrganization", unleash.FeatureToggleAutoDeactivateAndReactivateStudentsV2, mock.Anything, mock.Anything).Once().Return(true, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr:  nil,
			expectedResp: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.JPREPSchool),
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.setup(testCase.ctx)

			data, err := json.Marshal(testCase.req.(*OrderEventLog))
			assert.Nil(t, err)

			resp, err := service.SyncOrderHandler(testCase.ctx, data)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)

			mock.AssertExpectationsForObjects(t, db, tx)
		})
	}
}

func TestStudentRegistrationService_ReallocateStudentEnrollmentStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	zapLogger := ctxzap.Extract(ctx)

	testCases := []TestCase{
		{
			name: "happy case: reallocate student enrollment status success",
			ctx:  ctx,
			req: &npb.LessonReallocateStudentEnrollmentStatusEvent{
				StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
					{
						StudentId:        idutil.ULIDNow(),
						LocationId:       idutil.ULIDNow(),
						EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().Add(24 * time.Hour)),
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentRegistrationMock, ok := genericMock.(*prepareStudentRegistrationServiceMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentRegistrationMock.studentRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return([]entity.DomainStudent{entity.NullDomainStudent{}}, nil)
				studentRegistrationMock.locationRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				studentRegistrationMock.db.On("Begin", mock.Anything).Return(studentRegistrationMock.tx, nil)
				studentRegistrationMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, studentRegistrationMock.tx, mock.Anything).Once().Return(nil)
				studentRegistrationMock.userAccessPathRepo.On("UpsertMultiple", mock.Anything, studentRegistrationMock.tx, mock.Anything).Once().Return(nil)
				studentRegistrationMock.tx.On("Commit", mock.Anything).Return(nil)
			},
			expectedResp: false,
		},
		{
			name: "can not reallocate student enrollment status: invalid enrollment status",
			ctx:  ctx,
			req: &npb.LessonReallocateStudentEnrollmentStatusEvent{
				StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
					{
						StudentId:        idutil.ULIDNow(),
						LocationId:       idutil.ULIDNow(),
						EnrollmentStatus: npb.StudentEnrollmentStatus(pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED),
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().Add(24 * time.Hour)),
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {

			},
			expectedErr:  fmt.Errorf("ReallocateStudentEnrollmentStatus with invalid status: %s", npb.StudentEnrollmentStatus(pbu.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED).String()),
			expectedResp: false,
		},
		{
			name: "can not get students",
			ctx:  ctx,
			req: &npb.LessonReallocateStudentEnrollmentStatusEvent{
				StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
					{
						StudentId:        idutil.ULIDNow(),
						LocationId:       idutil.ULIDNow(),
						EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().Add(24 * time.Hour)),
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentRegistrationMock, ok := genericMock.(*prepareStudentRegistrationServiceMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentRegistrationMock.studentRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr:  pgx.ErrNoRows,
			expectedResp: false,
		},
		{
			name: "can not get locations",
			ctx:  ctx,
			req: &npb.LessonReallocateStudentEnrollmentStatusEvent{
				StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
					{
						StudentId:        idutil.ULIDNow(),
						LocationId:       idutil.ULIDNow(),
						EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().Add(24 * time.Hour)),
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentRegistrationMock, ok := genericMock.(*prepareStudentRegistrationServiceMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentRegistrationMock.studentRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return([]entity.DomainStudent{entity.NullDomainStudent{}}, nil)
				studentRegistrationMock.locationRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return(entity.DomainLocations{}, pgx.ErrNoRows)
			},
			expectedResp: false,
			expectedErr:  pgx.ErrNoRows,
		},
		{
			name: "invalid start date and end date",
			ctx:  ctx,
			req: &npb.LessonReallocateStudentEnrollmentStatusEvent{
				StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
					{
						StudentId:        idutil.ULIDNow(),
						LocationId:       idutil.ULIDNow(),
						EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().Add(-24 * time.Hour)),
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentRegistrationMock, ok := genericMock.(*prepareStudentRegistrationServiceMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentRegistrationMock.studentRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return([]entity.DomainStudent{entity.NullDomainStudent{}}, nil)
				studentRegistrationMock.locationRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return(entity.DomainLocations{}, pgx.ErrNoRows)
			},
			expectedResp: false,
			expectedErr:  pgx.ErrNoRows,
		},
		{
			name: "create student enrollment status failed",
			ctx:  ctx,
			req: &npb.LessonReallocateStudentEnrollmentStatusEvent{
				StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
					{
						StudentId:        idutil.ULIDNow(),
						LocationId:       idutil.ULIDNow(),
						EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().Add(24 * time.Hour)),
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentRegistrationMock, ok := genericMock.(*prepareStudentRegistrationServiceMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentRegistrationMock.studentRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return([]entity.DomainStudent{entity.NullDomainStudent{}}, nil)
				studentRegistrationMock.locationRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				studentRegistrationMock.db.On("Begin", mock.Anything).Return(studentRegistrationMock.tx, nil)
				studentRegistrationMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, studentRegistrationMock.tx, mock.Anything).Once().Return(fmt.Errorf("can't create student enrollment status"))
				studentRegistrationMock.tx.On("Rollback", mock.Anything).Return(nil)
			},
			expectedResp: false,
			expectedErr:  fmt.Errorf("ReallocateStudentEnrollmentStatus: %w", fmt.Errorf("StudentRegistrationService s.DomainEnrollmentStatusHistoryRepo.Create: %w", fmt.Errorf("can't create student enrollment status"))),
		},
		{
			name: "upsert user access path failed",
			ctx:  ctx,
			req: &npb.LessonReallocateStudentEnrollmentStatusEvent{
				StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
					{
						StudentId:        idutil.ULIDNow(),
						LocationId:       idutil.ULIDNow(),
						EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
						StartDate:        timestamppb.Now(),
						EndDate:          timestamppb.New(time.Now().Add(24 * time.Hour)),
					},
				},
			},
			setupWithMock: func(ctx context.Context, genericMock interface{}) {
				studentRegistrationMock, ok := genericMock.(*prepareStudentRegistrationServiceMock)
				if !ok {
					t.Error("invalid mock")
				}
				studentRegistrationMock.studentRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return([]entity.DomainStudent{entity.NullDomainStudent{}}, nil)
				studentRegistrationMock.locationRepo.On("GetByIDs", mock.Anything, studentRegistrationMock.db, mock.Anything).Once().Return(entity.DomainLocations{entity.NullDomainLocation{}}, nil)
				studentRegistrationMock.db.On("Begin", mock.Anything).Return(studentRegistrationMock.tx, nil)
				studentRegistrationMock.enrollmentStatusHistoryRepo.On("Create", mock.Anything, studentRegistrationMock.tx, mock.Anything).Once().Return(nil)
				studentRegistrationMock.userAccessPathRepo.On("UpsertMultiple", mock.Anything, studentRegistrationMock.tx, mock.Anything).Once().Return(fmt.Errorf("can't upsert user access path"))
				studentRegistrationMock.tx.On("Rollback", mock.Anything).Return(nil)
			},
			expectedResp: false,
			expectedErr:  fmt.Errorf("ReallocateStudentEnrollmentStatus: %w", fmt.Errorf("StudentRegistrationService s.DomainUserAccessPathRepo.UpsertMultiple: %w", fmt.Errorf("can't upsert user access path"))),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
				},
			}

			m, service := StudentRegistrationServiceMock()
			service.Logger = zapLogger

			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.setupWithMock(ctx, &m)

			data, err := json.Marshal(testCase.req.(*npb.LessonReallocateStudentEnrollmentStatusEvent))
			assert.Nil(t, err)

			resp, err := service.ReallocateStudentEnrollmentStatus(testCase.ctx, data)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)

			mock.AssertExpectationsForObjects(t)
		})
	}
}

func TestStudentRegistrationService_SyncOrderHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domainEnrollmentStatusHistoryRepo := new(mock_repositories.MockDomainEnrollmentStatusHistoryRepo)
	domainUserAccessPathRepo := new(mock_repositories.MockDomainUserAccessPathRepo)
	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)

	type fields struct {
		handelOrderFlowEnrollmentStatusManager MockHandelOrderFlowEnrollmentStatusManager
	}

	type args struct {
		data *OrderEventLog
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		setup   func()
		want    bool
		wantErr error
	}{
		{
			name: "submit order with order type is pause will do not thing",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_PAUSE.String(),
				},
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "submit order with order type is update will do not thing",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_UPDATE.String(),
				},
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "submit order with order type is enrollment will update enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleEnrollmentStatusUpdate: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "submit order with order type is withdrawal will update enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleEnrollmentStatusUpdate: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "submit order with order type is graduate will update enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_GRADUATE.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleEnrollmentStatusUpdate: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "submit order with order type is load will update enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_LOA.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleEnrollmentStatusUpdate: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "submit order with order type is new will handle enrollment status with new location",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_NEW.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleOrderFlowForTheNewLocation: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).
					Once().Return(entity.DomainEnrollmentStatusHistories{}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "submit order with order type is custom billing will handle enrollment status with new location",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleOrderFlowForTheExistedLocations: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				domainEnrollmentStatusHistoryRepo.On("GetByStudentIDAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).
					Once().Return(entity.DomainEnrollmentStatusHistories{&repository.EnrollmentStatusHistory{}}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "void order with order type is enrollment will handle void enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_VOIDED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleOrderFlowForVoidEnrollmentStatus: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},

			wantErr: nil,
		},
		{
			name: "void order with order type is withdrawal will handle void enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_VOIDED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleOrderFlowForVoidEnrollmentStatus: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "void order with order type is graduate will handle void enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_VOIDED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_GRADUATE.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleOrderFlowForVoidEnrollmentStatus: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "void order with order type is loa will handle void enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_VOIDED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_LOA.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleOrderFlowForVoidEnrollmentStatus: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "submit order with order type is resume will update enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_RESUME.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleEnrollmentStatusUpdate: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "void order with order type is resume will update enrollment status",
			args: args{
				data: &OrderEventLog{
					OrderStatus: pb.OrderStatus_ORDER_STATUS_VOIDED.String(),
					OrderType:   pb.OrderType_ORDER_TYPE_RESUME.String(),
				},
			},
			fields: fields{
				handelOrderFlowEnrollmentStatusManager: MockHandelOrderFlowEnrollmentStatusManager{
					handleOrderFlowForVoidEnrollmentStatus: func(ctx context.Context, db libdatabase.Ext, req *OrderEventLog) (bool, error) {
						return false, nil
					},
				},
			},
			want: false,
			setup: func() {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StudentRegistrationService{
				DB:                                db,
				DomainEnrollmentStatusHistoryRepo: domainEnrollmentStatusHistoryRepo,
				DomainUserAccessPathRepo:          domainUserAccessPathRepo,
				OrderFlowEnrollmentStatusManager:  &tt.fields.handelOrderFlowEnrollmentStatusManager,
			}
			if tt.setup != nil {
				tt.setup()
			}

			bytes, _ := json.Marshal(tt.args.data)
			got, err := s.SyncOrderHandler(ctx, bytes)
			assert.Equalf(t, tt.want, got, "SyncOrderHandler(%v, %v)", ctx, tt.args.data)
			assert.Equalf(t, tt.wantErr, err, "SyncOrderHandler(%v, %v)", ctx, tt.args.data)
		})
	}
}
