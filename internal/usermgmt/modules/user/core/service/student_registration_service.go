package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	MapOrderTypeAndEnrollmentStatus = map[string]string{
		pb.OrderType_ORDER_TYPE_ENROLLMENT.String(): entity.StudentEnrollmentStatusEnrolled,
		pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(): entity.StudentEnrollmentStatusWithdrawn,
		pb.OrderType_ORDER_TYPE_GRADUATE.String():   entity.StudentEnrollmentStatusGraduated,
		pb.OrderType_ORDER_TYPE_LOA.String():        entity.StudentEnrollmentStatusLOA,
		pb.OrderType_ORDER_TYPE_RESUME.String():     entity.StudentEnrollmentStatusEnrolled,
	}
)

type OrderEventLog struct {
	OrderStatus         string    `json:"order_status"`
	OrderType           string    `json:"order_type"`
	StudentID           string    `json:"student_id"`
	LocationID          string    `json:"location_id"`
	EnrollmentStatus    string    `json:"enrollment_status"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
	OrderID             string    `json:"order_id"`
	OrderSequenceNumber int32     `json:"order_sequence_number"`
}

type SyncOrderRequest struct {
	OrderEventLog *OrderEventLog `json:"order_event_log"`

	entity.DefaultDomainEnrollmentStatusHistory
}

func NewOrderLogRequest(orderEventLog *OrderEventLog) entity.DomainEnrollmentStatusHistory {
	return &SyncOrderRequest{OrderEventLog: orderEventLog}
}

func (req *SyncOrderRequest) UserID() field.String {
	return field.NewString(req.OrderEventLog.StudentID)
}

func (req *SyncOrderRequest) LocationID() field.String {
	return field.NewString(req.OrderEventLog.LocationID)
}

func (req *SyncOrderRequest) EnrollmentStatus() field.String {
	return field.NewString(req.OrderEventLog.EnrollmentStatus)
}

func (req *SyncOrderRequest) StartDate() field.Time {
	if req.OrderEventLog.StartDate.IsZero() {
		return field.NewTime(time.Now())
	}
	return field.NewTime(req.OrderEventLog.StartDate)
}

func (req *SyncOrderRequest) EndDate() field.Time {
	if req.OrderEventLog.EndDate.IsZero() {
		return field.NewNullTime()
	}
	return field.NewTime(req.OrderEventLog.EndDate)
}

func (req *SyncOrderRequest) OrderID() field.String {
	return field.NewString(req.OrderEventLog.OrderID)
}

func (req *SyncOrderRequest) OrderSequenceNumber() field.Int32 {
	return field.NewInt32(req.OrderEventLog.OrderSequenceNumber)
}

type MapUserAccessPath struct {
	OrderEventLog *OrderEventLog `json:"order_event_log"`

	entity.DefaultUserAccessPath
}

func NewUserAccessPath(orderEventLog *OrderEventLog) entity.DomainUserAccessPath {
	return &MapUserAccessPath{OrderEventLog: orderEventLog}
}

func (req *MapUserAccessPath) UserID() field.String {
	return field.NewString(req.OrderEventLog.StudentID)
}

func (req *MapUserAccessPath) LocationID() field.String {
	return field.NewString(req.OrderEventLog.LocationID)
}

type ReallocateStudentEnrollmentStatus struct {
	Evt *npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo

	entity.DefaultDomainEnrollmentStatusHistory
}

func (req *ReallocateStudentEnrollmentStatus) UserID() field.String {
	return field.NewString(req.Evt.StudentId)
}

func (req *ReallocateStudentEnrollmentStatus) LocationID() field.String {
	return field.NewString(req.Evt.LocationId)
}

func (req *ReallocateStudentEnrollmentStatus) EnrollmentStatus() field.String {
	return field.NewString(req.Evt.EnrollmentStatus.String())
}

func (req *ReallocateStudentEnrollmentStatus) StartDate() field.Time {
	if req.Evt.StartDate.AsTime().IsZero() {
		return field.NewTime(time.Now())
	}
	return field.NewTime(req.Evt.StartDate.AsTime())
}

func (req *ReallocateStudentEnrollmentStatus) EndDate() field.Time {
	if req.Evt.EndDate.AsTime().IsZero() {
		return field.NewNullTime()
	}
	return field.NewTime(req.Evt.EndDate.AsTime())
}

func (req *ReallocateStudentEnrollmentStatus) OrderID() field.String {
	return field.NewNullString()
}

func (req *ReallocateStudentEnrollmentStatus) OrderSequenceNumber() field.Int32 {
	return field.NewNullInt32()
}

func NewReallocateStudentEnrollmentStatus(evt *npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo) *ReallocateStudentEnrollmentStatus {
	return &ReallocateStudentEnrollmentStatus{
		Evt: evt,
	}
}

type ReallocateStudentEnrollmentStatusEvent struct {
	*npb.LessonReallocateStudentEnrollmentStatusEvent
}

func (evt *ReallocateStudentEnrollmentStatusEvent) LocationIDs() []string {
	mapIDs := map[string]struct{}{}
	for _, studentStatus := range evt.StudentEnrollmentStatus {
		mapIDs[studentStatus.LocationId] = struct{}{}
	}
	ids := []string{}
	for id := range mapIDs {
		ids = append(ids, id)
	}

	return ids
}

func (evt *ReallocateStudentEnrollmentStatusEvent) StudentIDs() []string {
	mapIDs := map[string]struct{}{}
	for _, studentStatus := range evt.StudentEnrollmentStatus {
		mapIDs[studentStatus.StudentId] = struct{}{}
	}
	ids := []string{}
	for id := range mapIDs {
		ids = append(ids, id)
	}

	return ids
}

func (evt *ReallocateStudentEnrollmentStatusEvent) validStatus() error {
	for _, v := range evt.StudentEnrollmentStatus {
		if v.EnrollmentStatus.String() != entity.StudentEnrollmentStatusTemporary {
			return fmt.Errorf("ReallocateStudentEnrollmentStatus with invalid status: %s", v.EnrollmentStatus.String())
		}
	}

	return nil
}

type DomainEnrollmentStatusHistoryRepo interface {
	Create(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryToCreate entity.DomainEnrollmentStatusHistory) error
	GetByStudentIDAndLocationID(ctx context.Context, db database.QueryExecer, studentID, locationID string, getCurrent bool) (entity.DomainEnrollmentStatusHistories, error)
	GetLatestEnrollmentStudentOfLocation(ctx context.Context, db database.QueryExecer, studentID, locationID string) ([]entity.DomainEnrollmentStatusHistory, error)
	SoftDeleteEnrollments(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryToCreate entity.DomainEnrollmentStatusHistory) error
	DeactivateEnrollmentStatus(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryToCreate entity.DomainEnrollmentStatusHistory, endDateReq time.Time) error
	GetByStudentID(ctx context.Context, db database.QueryExecer, studentID string, getCurrent bool) ([]entity.DomainEnrollmentStatusHistory, error)
	GetByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (entity.DomainEnrollmentStatusHistories, error)
	Update(ctx context.Context, db database.QueryExecer, enrollStatusHistDB, enrollStatusHistReq entity.DomainEnrollmentStatusHistory) error
	GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate(ctx context.Context, db database.QueryExecer, enrollmentStatusHistoryReq entity.DomainEnrollmentStatusHistory) ([]entity.DomainEnrollmentStatusHistory, error)
	BulkInsert(ctx context.Context, db database.QueryExecer, reqEnrollmentStatusHistoriesToCreate entity.DomainEnrollmentStatusHistories) error
	UpdateStudentStatusBasedEnrollmentStatus(ctx context.Context, db database.QueryExecer, studentIDs, deactivateEnrollmentStatuses []string) error
	GetSameStartDateEnrollmentStatusHistory(ctx context.Context, db database.QueryExecer, _e entity.DomainEnrollmentStatusHistory) (entity.DomainEnrollmentStatusHistories, error)
}

type DomainUserAccessPathRepo interface {
	UpsertMultiple(ctx context.Context, db database.QueryExecer, userAccessPaths ...entity.DomainUserAccessPath) error
	SoftDeleteByUserIDAndLocationIDs(ctx context.Context, db database.QueryExecer, userID, organizationID string, locationIDs []string) error
	GetByUserID(ctx context.Context, db database.QueryExecer, userID field.String) (entity.DomainUserAccessPaths, error)
}

type StudentRegistrationService struct {
	DB                                database.Ext
	Logger                            *zap.Logger
	Env                               string
	DomainEnrollmentStatusHistoryRepo DomainEnrollmentStatusHistoryRepo
	DomainUserAccessPathRepo          DomainUserAccessPathRepo
	UnleashClient                     unleashclient.ClientInstance
	StudentRepo                       interface {
		GetByIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]entity.DomainStudent, error)
	}
	LocationRepo interface {
		GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) (entity.DomainLocations, error)
	}
	OrderFlowEnrollmentStatusManager         OrderFlowEnrollmentStatusManager
	EnrollmentStatusHistoryStartDateModifier EnrollmentStatusHistoryStartDateModifierFn
}

func (s *StudentRegistrationService) SyncOrderHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req *OrderEventLog
	if err := json.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("handleSyncOrder json.Unmarshal: %w", err)
	}
	// Set current datetime to adapt with user logic
	if req.StartDate.UTC().Before(time.Now().UTC()) {
		req.StartDate = time.Now().UTC().Truncate(time.Second)
	}

	switch req.OrderStatus {
	case pb.OrderStatus_ORDER_STATUS_SUBMITTED.String():
		switch req.OrderType {
		// when order type is pause/update -> do nothing
		case pb.OrderType_ORDER_TYPE_PAUSE.String(),
			pb.OrderType_ORDER_TYPE_UPDATE.String():
			return false, nil
		// when order type is enrollment/withdrawal/graduate/LOA -> just add record to enrollment status history
		case pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
			pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
			pb.OrderType_ORDER_TYPE_GRADUATE.String(),
			pb.OrderType_ORDER_TYPE_LOA.String(),
			pb.OrderType_ORDER_TYPE_RESUME.String():
			var (
				retry bool
				err   error
			)
			if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
				retry, err = s.OrderFlowEnrollmentStatusManager.HandleEnrollmentStatusUpdate(ctx, tx, req)
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				s.Logger.Error("OrderFlowEnrollmentStatusManager.HandleEnrollmentStatusUpdate",
					zap.String("StudentID", req.StudentID),
					zap.String("LocationID", req.LocationID),
					zap.String("OrderStatus", req.OrderStatus),
					zap.String("OrderType", req.OrderType),
					zap.String("StartDate", req.StartDate.String()),
					zap.String("EndDate", req.EndDate.String()),
					zap.Error(err),
				)
				return retry, err
			}
			return false, nil
		// when order type is new/custom we check location is new or not
		case pb.OrderType_ORDER_TYPE_NEW.String(),
			pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String():
			// check location of student is exists (new or existed)

			var retry bool
			if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
				enrollmentStatusHistories, err := s.DomainEnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID(ctx, tx, req.StudentID, req.LocationID, false)
				if err != nil {
					s.Logger.Error("DomainEnrollmentStatusHistoryRepo.GetByStudentIDAndLocationID",
						zap.String("StudentID", req.StudentID),
						zap.String("LocationID", req.LocationID),
						zap.String("OrderStatus", req.OrderStatus),
						zap.String("OrderType", req.OrderType),
						zap.String("StartDate", req.StartDate.String()),
						zap.String("EndDate", req.EndDate.String()),
						zap.Error(err),
					)
					// return true to retry
					retry = true
					return err
				}
				if len(enrollmentStatusHistories) == 0 {
					// new location
					if retry, err = s.OrderFlowEnrollmentStatusManager.HandleForNewLocation(ctx, tx, req); err != nil {
						s.Logger.Error("OrderFlowEnrollmentStatusManager.HandleForNewLocation",
							zap.String("StudentID", req.StudentID),
							zap.String("LocationID", req.LocationID),
							zap.String("OrderStatus", req.OrderStatus),
							zap.String("OrderType", req.OrderType),
							zap.String("StartDate", req.StartDate.String()),
							zap.String("EndDate", req.EndDate.String()),
							zap.Error(err),
						)
						return err
					}
				} else {
					// existed location
					if retry, err = s.OrderFlowEnrollmentStatusManager.HandleExistedLocations(ctx, tx, req); err != nil {
						s.Logger.Error("OrderFlowEnrollmentStatusManager.HandleExistedLocations",
							zap.String("StudentID", req.StudentID),
							zap.String("LocationID", req.LocationID),
							zap.String("OrderStatus", req.OrderStatus),
							zap.String("OrderType", req.OrderType),
							zap.String("StartDate", req.StartDate.String()),
							zap.String("EndDate", req.EndDate.String()),
							zap.Error(err),
						)
						return err
					}
				}

				return nil
			}); err != nil {
				return retry, err
			}
			return false, nil
		}
	case pb.OrderStatus_ORDER_STATUS_VOIDED.String():
		switch req.OrderType {
		case pb.OrderType_ORDER_TYPE_ENROLLMENT.String(),
			pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
			pb.OrderType_ORDER_TYPE_GRADUATE.String(),
			pb.OrderType_ORDER_TYPE_LOA.String(),
			pb.OrderType_ORDER_TYPE_NEW.String(),
			pb.OrderType_ORDER_TYPE_RESUME.String():

			var (
				retry bool
				err   error
			)
			if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
				if retry, err = s.OrderFlowEnrollmentStatusManager.HandleVoidEnrollmentStatus(ctx, tx, req); err != nil {
					s.Logger.Error("OrderFlowEnrollmentStatusManager.HandleVoidEnrollmentStatus",
						zap.String("StudentID", req.StudentID),
						zap.String("LocationID", req.LocationID),
						zap.String("OrderStatus", req.OrderStatus),
						zap.String("OrderType", req.OrderType),
						zap.String("StartDate", req.StartDate.String()),
						zap.String("EndDate", req.EndDate.String()),
						zap.Error(err),
					)
					return err
				}
				return nil
			}); err != nil {
				return retry, err
			}
			return false, nil
		default:
			return false, nil
		}
	}
	return false, nil
}

func (s *StudentRegistrationService) ReallocateStudentEnrollmentStatus(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var event *npb.LessonReallocateStudentEnrollmentStatusEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return false, fmt.Errorf("ReallocateStudentEnrollmentStatus json.Unmarshal: %w", err)
	}

	reallocateStudentEnrollmentStatusEvent := ReallocateStudentEnrollmentStatusEvent{event}
	if err := reallocateStudentEnrollmentStatusEvent.validStatus(); err != nil {
		return false, err
	}

	studentIDs := reallocateStudentEnrollmentStatusEvent.StudentIDs()
	existingStudents, err := s.StudentRepo.GetByIDs(ctx, s.DB, studentIDs)
	if err != nil {
		return false, err
	}
	if len(existingStudents) != len(studentIDs) {
		return false, fmt.Errorf("invalid student ids: %s", studentIDs)
	}

	locationIDs := reallocateStudentEnrollmentStatusEvent.LocationIDs()
	existingLocations, err := s.LocationRepo.GetByIDs(ctx, s.DB, locationIDs)
	if err != nil {
		return false, err
	}
	if len(existingLocations) != len(locationIDs) {
		return false, fmt.Errorf("invalid location ids: %s", locationIDs)
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for _, studentEnrollmentStatus := range reallocateStudentEnrollmentStatusEvent.StudentEnrollmentStatus {
			if studentEnrollmentStatus.StartDate.AsTime().After(studentEnrollmentStatus.EndDate.AsTime()) {
				return fmt.Errorf("ReallocateStudentEnrollmentStatus: start_date must be before end_date")
			}

			reallocateStudent := NewReallocateStudentEnrollmentStatus(studentEnrollmentStatus)
			if err = s.Create(ctx, tx, reallocateStudent); err != nil {
				return fmt.Errorf("ReallocateStudentEnrollmentStatus: %w", err)
			}
			if err := s.UpsertMultipleAccessPath(ctx, tx, reallocateStudent); err != nil {
				return fmt.Errorf("ReallocateStudentEnrollmentStatus: %w", err)
			}
		}
		return nil
	})

	return false, err
}

func (s *StudentRegistrationService) SyncEnrollmentStatusHistory(ctx context.Context, db database.Ext, req *OrderEventLog, enrollmentStatus string) error {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "OrganizationFromContext")
	}

	orderLogSync := OrderEventLog{
		OrderStatus:         req.OrderStatus,
		OrderType:           req.OrderType,
		StudentID:           req.StudentID,
		LocationID:          req.LocationID,
		EnrollmentStatus:    enrollmentStatus,
		StartDate:           req.StartDate,
		EndDate:             req.EndDate,
		OrderID:             req.OrderID,
		OrderSequenceNumber: req.OrderSequenceNumber,
	}

	hadDate := !req.StartDate.IsZero() && !req.EndDate.IsZero()

	if hadDate && req.StartDate.After(req.EndDate) {
		s.Logger.Error("StudentRegistrationService.SyncEnrollmentStatusHistory",
			zap.String("StartDate", req.StartDate.String()),
		)
		return nil
	}

	switch enrollmentStatus {
	case entity.StudentEnrollmentStatusEnrolled,
		entity.StudentEnrollmentStatusWithdrawn,
		entity.StudentEnrollmentStatusGraduated,
		entity.StudentEnrollmentStatusTemporary,
		entity.StudentEnrollmentStatusPotential,
		entity.StudentEnrollmentStatusLOA:
		enrollmentStatusHistory := NewOrderLogRequest(&orderLogSync)
		userAccessPath := NewUserAccessPath(&orderLogSync)

		domainUserAccessPath := entity.UserAccessPathWillBeDelegated{
			HasUserID:         userAccessPath,
			HasLocationID:     userAccessPath,
			HasOrganizationID: organization,
		}

		enrollmentStatusHistory, err = s.EnrollmentStatusHistoryStartDateModifier(ctx, db, s.DomainEnrollmentStatusHistoryRepo, enrollmentStatusHistory)
		if err != nil {
			return err
		}

		if err := s.Create(ctx, db, enrollmentStatusHistory); err != nil {
			return fmt.Errorf("StudentRegistrationService.SyncEnrollmentStatusHistory sync order log : %w", err)
		}

		if err := s.UpsertMultipleAccessPath(ctx, db, domainUserAccessPath); err != nil {
			return fmt.Errorf("StudentRegistrationService.SyncEnrollmentStatusHistory sync order log access Path : %w", err)
		}

		if err := s.DeactivateAndReactivateStudents(ctx, db, []string{orderLogSync.StudentID}); err != nil {
			return err
		}

	default:
		s.Logger.Warn("StudentRegistrationService.SyncEnrollmentStatusHistory",
			zap.String("EnrollmentStatus", req.EnrollmentStatus),
		)
		return nil
	}

	return nil
}

func (s *StudentRegistrationService) DeactivateAndReactivateStudents(ctx context.Context, db database.Ext, studentIDs []string) error {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "OrganizationFromContext")
	}
	enrollmentStatuses := []string{upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String()}
	isEnableAutoDeactivateAndReactivateStudentsV2, err := s.UnleashClient.IsFeatureEnabledOnOrganization(unleash.FeatureToggleAutoDeactivateAndReactivateStudentsV2, s.Env, organization.OrganizationID().String())
	if err != nil {
		isEnableAutoDeactivateAndReactivateStudentsV2 = false
	}
	if isEnableAutoDeactivateAndReactivateStudentsV2 {
		enrollmentStatuses = []string{
			upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String(),
			upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String(),
			upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
		}
	}
	if err := s.DomainEnrollmentStatusHistoryRepo.UpdateStudentStatusBasedEnrollmentStatus(ctx, db, studentIDs, enrollmentStatuses); err != nil {
		return fmt.Errorf("DeactivateAndReactivateStudentsManager.DeactivateAndReactivateStudents sync order log error : %w", err)
	}
	return nil
}

func (s *StudentRegistrationService) Create(ctx context.Context, tx database.Ext, enrollmentStatusHistoryToCreate entity.DomainEnrollmentStatusHistory) error {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "OrganizationFromContext")
	}

	enrollmentStatusHistory := entity.EnrollmentStatusHistoryWillBeDelegated{
		EnrollmentStatusHistory: enrollmentStatusHistoryToCreate,
		HasUserID:               enrollmentStatusHistoryToCreate,
		HasLocationID:           enrollmentStatusHistoryToCreate,
		HasOrganizationID:       organization,
	}

	err = s.DomainEnrollmentStatusHistoryRepo.Create(ctx, tx, enrollmentStatusHistory)
	if err != nil {
		return fmt.Errorf("StudentRegistrationService s.DomainEnrollmentStatusHistoryRepo.Create: %w", err)
	}

	return nil
}

func (s *StudentRegistrationService) UpsertMultipleAccessPath(ctx context.Context, tx database.Ext, userAccessPathToCreate entity.DomainUserAccessPath) error {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "OrganizationFromContext")
	}

	userAccessPath := entity.UserAccessPathWillBeDelegated{
		HasUserID:         userAccessPathToCreate,
		HasLocationID:     userAccessPathToCreate,
		HasOrganizationID: organization,
	}

	err = s.DomainUserAccessPathRepo.UpsertMultiple(ctx, tx, []entity.DomainUserAccessPath{userAccessPath}...)
	if err != nil {
		return fmt.Errorf("StudentRegistrationService s.DomainUserAccessPathRepo.UpsertMultiple: %w", err)
	}

	return nil
}
