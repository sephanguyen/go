package usermgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	ppb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultInterval = 15000 * time.Millisecond
	defaultTimeout  = 200000 * time.Millisecond
)

type enrollmentAssertion struct {
	expectNewEnrollmentStatus     repository.EnrollmentStatusHistory
	expectCurrentEnrollmentStatus repository.EnrollmentStatusHistory
	expectRemoveEnrollmentStatus  repository.EnrollmentStatusHistory
}

func (s *suite) schoolAdminCreateOrderWithOrderStatusAndOrderType(ctx context.Context, orderType string) (context.Context, error) {
	response := s.Response.(*upb.UpsertStudentResponse)
	if len(response.StudentProfiles) == 0 {
		return ctx, fmt.Errorf("no student found")
	}

	studentID := response.StudentProfiles[0].Id
	locations, err := getLocationsOfUser(ctx, s.BobDBTrace, studentID)
	if err != nil {
		return ctx, err
	}

	orderEventLog := service.OrderEventLog{
		StudentID:           studentID,
		StartDate:           time.Now(),
		OrderID:             "order_id", // set it constantly for testing void order
		OrderSequenceNumber: 1,
	}

	expectedCurrentOne := repository.EnrollmentStatusHistory{
		EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
			UserID:     field.NewString(orderEventLog.StudentID),
			LocationID: field.NewString(orderEventLog.LocationID),
			OrderID:    field.NewString(orderEventLog.OrderID),
			StartDate:  field.NewTime(orderEventLog.StartDate),
		},
	}
	expectedNewOne := expectedCurrentOne
	expectedRemoveOne := expectedCurrentOne

	// this enrollment status history was created by api, so it does not have order id
	expectedCurrentOne.EnrollmentStatusHistoryAttribute.OrderID = field.NewNullString()
	// after submit order, the current enrollment status history will be updated end date
	expectedCurrentOne.EnrollmentStatusHistoryAttribute.EndDate = field.NewTime(orderEventLog.StartDate.Add(-1 * time.Second))

	expectedRemoveOne.DeletedAtAttr = field.NewTime(time.Now())

	switch orderType {
	case "submit graduate order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_GRADUATE.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusGraduated)
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusPotential)
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "submit withdrawal order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_WITHDRAWAL.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusWithdrawn)
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusPotential)
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "submit enrolled order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_ENROLLMENT.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusEnrolled)
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusPotential)
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "submit enrolled order to the location has temporary enrollment status":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_ENROLLMENT.String()

		history, err := findEnrollmentStatusHistoryByStatus(ctx, s.BobDBTrace, studentID, entity.StudentEnrollmentStatusTemporary)
		if err != nil {
			return ctx, errors.Wrap(err, "findEnrollmentStatusHistoryByStatus failed")
		}
		orderEventLog.LocationID = history.LocationID().String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusEnrolled)
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusTemporary)
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "submit new order to the new location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_NEW.String()

		newLocation, err := getNewLocationsForStudent(ctx, s.BobDBTrace, studentID)
		if err != nil {
			return ctx, errors.Wrap(err, "getNewLocationsForStudent failed")
		}
		orderEventLog.LocationID = newLocation.LocationID().String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusPotential)
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "submit new order to the location has temporary enrollment status":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_NEW.String()

		history, err := findEnrollmentStatusHistoryByStatus(ctx, s.BobDBTrace, studentID, entity.StudentEnrollmentStatusTemporary)
		if err != nil {
			return ctx, errors.Wrap(err, "findEnrollmentStatusHistoryByStatus failed")
		}
		orderEventLog.LocationID = history.LocationID().String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusEnrolled)
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusTemporary)
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "submit LOA order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_LOA.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusLOA)
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusPotential)
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "submit resume order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_RESUME.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusEnrolled)
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusPotential)
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "void enrolled order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_VOIDED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_ENROLLMENT.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedRemoveOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusEnrolled)
		expectedRemoveOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "void withdrawal order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_VOIDED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_WITHDRAWAL.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedRemoveOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusWithdrawn)
		expectedRemoveOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "void graduated order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_VOIDED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_GRADUATE.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedRemoveOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusGraduated)
		expectedRemoveOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "void LOA order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_VOIDED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_LOA.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedRemoveOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusLOA)
		expectedRemoveOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "void resume order to the existed location":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_VOIDED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_RESUME.String()
		orderEventLog.LocationID = locations[0].LocationID().String()

		expectedRemoveOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusEnrolled)
		expectedRemoveOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "void order to the location has temporary enrollment status":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_VOIDED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_NEW.String()

		history, err := findEnrollmentStatusHistoryByStatus(ctx, s.BobDBTrace, studentID, entity.StudentEnrollmentStatusTemporary)
		if err != nil {
			return ctx, errors.Wrap(err, "findEnrollmentStatusHistoryByStatus failed")
		}
		orderEventLog.LocationID = history.LocationID().String()

		expectedRemoveOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = field.NewString(entity.StudentEnrollmentStatusPotential)
		expectedRemoveOne.EnrollmentStatusHistoryAttribute.LocationID = field.NewString(orderEventLog.LocationID)

	case "submit new withdrawal order with same location and start date of deleted enrollment status history":
		orderEventLog.OrderStatus = ppb.OrderStatus_ORDER_STATUS_SUBMITTED.String()
		orderEventLog.OrderType = ppb.OrderType_ORDER_TYPE_WITHDRAWAL.String()

		newOrderID := field.NewString(idutil.ULIDNow())
		enrollmentStatus := field.NewString(entity.StudentEnrollmentStatusWithdrawn)

		history, err := findEnrollmentStatusHistoryByStatus(ctx, s.BobDBTrace, studentID, enrollmentStatus.String())
		if err != nil {
			return ctx, errors.Wrap(err, "findEnrollmentStatusHistoryByStatus failed")
		}

		orderEventLog.LocationID = history.LocationID().String()
		orderEventLog.StartDate = history.StartDate().Time()
		orderEventLog.OrderID = newOrderID.String()

		expectedNewOne.EnrollmentStatusHistoryAttribute.OrderID = newOrderID
		expectedNewOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = enrollmentStatus
		expectedNewOne.EnrollmentStatusHistoryAttribute.LocationID = history.LocationID()
		expectedNewOne.EnrollmentStatusHistoryAttribute.StartDate = field.NewTime(orderEventLog.StartDate.Add(time.Microsecond).Add(time.Second))

		expectedCurrentOne.EnrollmentStatusHistoryAttribute.LocationID = history.LocationID()
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.StartDate = history.StartDate()
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EndDate = field.NewNullTime()
		expectedCurrentOne.EnrollmentStatusHistoryAttribute.EnrollmentStatus = enrollmentStatus
		expectedCurrentOne.DeletedAtAttr = field.NewTime(time.Now())

	default:
		return ctx, fmt.Errorf("invalid order type")
	}

	data, err := json.Marshal(orderEventLog)
	if err != nil {
		return ctx, errors.Wrap(err, "json.Marshal failed")
	}
	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectOrderEventLogCreated, data)
	if err != nil {
		return ctx, nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishOrderEventLog JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}

	s.Request1 = orderEventLog
	s.ExpectedData = enrollmentAssertion{
		expectNewEnrollmentStatus:     expectedNewOne,
		expectCurrentEnrollmentStatus: expectedCurrentOne,
		expectRemoveEnrollmentStatus:  expectedRemoveOne,
	}
	return ctx, nil
}

func (s *suite) checkEnrollmentStatusHistoriesOfStudentOrderFlow(ctx context.Context, status string) (context.Context, error) {
	retryCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	orderEventLog := s.Request1.(service.OrderEventLog)
	assertion := s.ExpectedData.(enrollmentAssertion)

	switch status {
	case "be updated":
		err := enrollmentStatusHistoriesOfStudentMustBeUpdated(retryCtx, s.BobDBTrace, orderEventLog, assertion)
		if err != nil {
			return ctx, errors.Wrap(err, "enrollmentStatusHistoriesOfStudentMustBeUpdated failed")
		}

	case "not be updated":
		// the below function will assert request payload when create student with current student in database
		// it included enrollment status histories
		if _, err := s.studentsWereUpsertedSuccessfullyByGRPC(retryCtx); err != nil {
			return ctx, errors.Wrap(err, "s.studentsWereUpsertedSuccessfullyByGRPC failed")
		}

	case "be removed correspondingly":
		err := enrollmentStatusHistoriesOfStudentMustBeRemoved(retryCtx, s.BobDBTrace, orderEventLog, assertion)
		if err != nil {
			return ctx, errors.Wrap(err, "enrollmentStatusHistoriesOfStudentMustBeRemoved failed")
		}

	case "be potential":
		err := enrollmentStatusHistoriesOfStudentMustBeSpecify(retryCtx, s.BobDBTrace, orderEventLog, entity.StudentEnrollmentStatusPotential)
		if err != nil {
			return ctx, errors.Wrap(err, "enrollmentStatusHistoriesOfStudentMustBeSpecify failed")
		}

	default:
		return ctx, fmt.Errorf("invalid status")
	}

	return ctx, nil
}

func (s *suite) studentsWereUpsertedSuccessfully(ctx context.Context) (context.Context, error) {
	request := s.Request.(*upb.UpsertStudentRequest)
	orderEventLog := s.Request1.(service.OrderEventLog)

	// re-assign enrollment status from synced to the order event log for validation s.studentsWereUpsertedSuccessfullyByGRPC
	enrollmentStatus := upb.StudentEnrollmentStatus_value[service.MapOrderTypeAndEnrollmentStatus[orderEventLog.OrderType]]
	for _, enrollmentStatusHistory := range request.StudentProfiles[0].EnrollmentStatusHistories {
		if enrollmentStatusHistory.LocationId == orderEventLog.LocationID {
			enrollmentStatusHistory.EnrollmentStatus = upb.StudentEnrollmentStatus(enrollmentStatus)
			enrollmentStatusHistory.StartDate = timestamppb.New(orderEventLog.StartDate)
		}
	}

	if _, err := s.studentsWereUpsertedSuccessfullyByGRPC(ctx); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) deleteEnrollmentHistoryWithStatusOfStudent(ctx context.Context, status string) (context.Context, error) {
	response := s.Response.(*upb.UpsertStudentResponse)
	if len(response.StudentProfiles) == 0 {
		return ctx, fmt.Errorf("invalid response")
	}

	mapStatus := map[string]string{
		"withdrawn": entity.StudentEnrollmentStatusWithdrawn,
	}

	enrollmentStatus := mapStatus[status]
	if enrollmentStatus == "" {
		return ctx, fmt.Errorf("invalid status: %s", status)
	}

	enrollmentStatusHistory := new(upb.EnrollmentStatusHistory)
	for _, history := range response.StudentProfiles[0].EnrollmentStatusHistories {
		if history.EnrollmentStatus.String() == mapStatus[status] {
			enrollmentStatusHistory = history
		}
	}

	if _, err := s.BobDB.Exec(ctx,
		`
			UPDATE student_enrollment_status_history
				SET deleted_at = NOW()
			WHERE
				student_id = $1 AND
				location_id = $2 AND
				enrollment_status = $3
		`,
		enrollmentStatusHistory.StudentId,
		enrollmentStatusHistory.LocationId,
		enrollmentStatusHistory.EnrollmentStatus,
	); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) simulateTheOrderEvent(ctx context.Context, _ string, numberOfStudents int) (context.Context, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return ctx, err
	}
	createResp := s.Response.(*upb.UpsertStudentResponse)
	students := createResp.StudentProfiles
	enrollmentStatusHistoryRepo := &repository.DomainEnrollmentStatusHistoryRepo{}

	for i := 0; i < numberOfStudents; i++ {
		createdEnrollmentStatus := students[i].EnrollmentStatusHistories[0]
		statusToDeactivate := &repository.EnrollmentStatusHistory{
			EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
				UserID:           field.NewString(students[i].Id),
				LocationID:       field.NewString(createdEnrollmentStatus.LocationId),
				EnrollmentStatus: field.NewString(createdEnrollmentStatus.EnrollmentStatus.String()),
				StartDate:        field.NewTime(createdEnrollmentStatus.StartDate.AsTime()),
			},
		}
		statusToCreate := &repository.EnrollmentStatusHistory{
			EnrollmentStatusHistoryAttribute: repository.EnrollmentStatusHistoryAttribute{
				UserID:              field.NewString(students[i].Id),
				LocationID:          field.NewString(createdEnrollmentStatus.LocationId),
				EnrollmentStatus:    field.NewString(entity.StudentEnrollmentStatusEnrolled),
				StartDate:           field.NewTime(time.Now().Add(-1 * time.Second)),
				EndDate:             field.NewNullTime(),
				OrderID:             field.NewNullString(),
				OrderSequenceNumber: field.NewNullInt32(),
				OrganizationID:      organization.OrganizationID(),
			},
		}
		if err := enrollmentStatusHistoryRepo.DeactivateEnrollmentStatus(ctx, s.BobDBTrace, statusToDeactivate, statusToCreate.StartDate().Time().Add(-1*time.Second)); err != nil {
			return ctx, fmt.Errorf("can not emulate to deactivate status of student: %s", students[i].Id)
		}
		if err := enrollmentStatusHistoryRepo.Create(ctx, s.BobDBTrace, statusToCreate); err != nil {
			return ctx, fmt.Errorf("can not create new status for student: %s", students[i].Id)
		}
	}
	return ctx, nil
}

func getLocationsOfUser(ctx context.Context, db database.QueryExecer, userID string) (entity.DomainLocations, error) {
	rows, err := db.Query(ctx, "select location_id from user_access_paths where user_id = $1;", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, err
	}

	var locations []entity.DomainLocation
	for rows.Next() {
		location := repository.Location{}
		err = rows.Scan(&location.ID)
		if err != nil {
			return nil, err
		}
		locations = append(locations, &location)
	}
	return locations, nil
}

func getNewLocationsForStudent(ctx context.Context, db database.Ext, studentID string) (entity.DomainLocation, error) {
	locations, err := new(repository.DomainLocationRepo).RetrieveLowestLevelLocations(ctx, db, "", 0, 0, nil)
	if err != nil {
		return nil, err
	}

	locationsOfUser, err := getLocationsOfUser(ctx, db, studentID)
	if err != nil {
		return nil, err
	}

	for _, location := range locations {
		if !golibs.InArrayString(location.LocationID().String(), locationsOfUser.LocationIDs()) {
			return location, nil
		}
	}
	return nil, errors.New("no new locations for student")
}

func enrollmentStatusHistoriesOfStudentMustBeUpdated(ctx context.Context, db database.QueryExecer, orderEventLog service.OrderEventLog, assertion enrollmentAssertion) error {
	return TryUntilSuccess(ctx, defaultInterval, func(ctx context.Context) (bool, error) {
		enrollmentStatusHistories, err := getAllEnrollmentStatusHistoriesOfStudent(ctx, db, orderEventLog.StudentID)
		if err != nil {
			return false, errors.Wrap(err, "enrollmentStatusRepo.GetByStudentIDAndLocationID failed")
		}

		existedEnrollmentStatusDB := make(map[string]struct{})
		for _, history := range enrollmentStatusHistories {
			existedEnrollmentStatusDB[getUniqueInfo(history)] = struct{}{}
		}

		uniqueInfoNewEnrollmentStatus := getUniqueInfo(assertion.expectNewEnrollmentStatus)
		uniqueInfoOldEnrollmentStatus := getUniqueInfo(assertion.expectCurrentEnrollmentStatus)

		if _, ok := existedEnrollmentStatusDB[uniqueInfoNewEnrollmentStatus]; !ok {
			return true, fmt.Errorf("data not sync")
		}

		if _, ok := existedEnrollmentStatusDB[uniqueInfoOldEnrollmentStatus]; !ok {
			return true, fmt.Errorf("data not sync")
		}

		return false, nil
	})
}

func enrollmentStatusHistoriesOfStudentMustBeRemoved(ctx context.Context, db database.QueryExecer, orderEventLog service.OrderEventLog, assertion enrollmentAssertion) error {
	return TryUntilSuccess(ctx, defaultInterval, func(ctx context.Context) (bool, error) {
		enrollmentStatusHistories, err := getAllEnrollmentStatusHistoriesOfStudent(ctx, db, orderEventLog.StudentID)
		if err != nil {
			return false, errors.Wrap(err, "getAllEnrollmentStatusHistoriesOfStudent failed")
		}

		existedEnrollmentStatusDB := make(map[string]struct{})
		for _, history := range enrollmentStatusHistories {
			existedEnrollmentStatusDB[getUniqueInfo(history)] = struct{}{}
		}

		softDeletedEnrollmentStatusHistory := getUniqueInfo(assertion.expectRemoveEnrollmentStatus)
		if _, ok := existedEnrollmentStatusDB[softDeletedEnrollmentStatusHistory]; !ok {
			return true, fmt.Errorf("data not sync")
		}

		return false, nil
	})
}

func enrollmentStatusHistoriesOfStudentMustBeSpecify(ctx context.Context, db database.QueryExecer, orderEventLog service.OrderEventLog, enrollmentStatus string) error {
	return TryUntilSuccess(ctx, defaultInterval, func(ctx context.Context) (bool, error) {
		enrollmentStatusRepo := new(repository.DomainEnrollmentStatusHistoryRepo)
		enrollmentStatusHistories, err := enrollmentStatusRepo.GetByStudentIDAndLocationID(ctx, db, orderEventLog.StudentID, orderEventLog.LocationID, true)
		if err != nil {
			return false, errors.Wrap(err, "enrollmentStatusRepo.GetByStudentIDAndLocationID failed")
		}

		isNewEnrollmentStatusCorrected := false
		for _, enrollmentStatusHistory := range enrollmentStatusHistories {
			if enrollmentStatusHistory.LocationID().String() != orderEventLog.LocationID {
				continue
			}

			isNewEnrollmentStatusCorrected = enrollmentStatusHistory.EnrollmentStatus().String() == enrollmentStatus
		}

		if !isNewEnrollmentStatusCorrected {
			return true, fmt.Errorf("data not sync")
		}

		return false, nil
	})
}

func getAllEnrollmentStatusHistoriesOfStudent(ctx context.Context, db database.QueryExecer, studentID string) ([]repository.EnrollmentStatusHistory, error) {
	enrollmentStatusHistories := make([]repository.EnrollmentStatusHistory, 0)
	fieldName, _ := new(repository.EnrollmentStatusHistory).FieldMap()
	query := `
		SELECT %s
		FROM %s
		WHERE student_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.Query(ctx, fmt.Sprintf(query, strings.Join(fieldName, ", "), new(repository.EnrollmentStatusHistory).TableName()), studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, err
	}
	for rows.Next() {
		statusHistory := repository.EnrollmentStatusHistory{}
		_, fields := statusHistory.FieldMap()
		if err := rows.Scan(fields...); err != nil {
			return nil, err
		}
		enrollmentStatusHistories = append(enrollmentStatusHistories, statusHistory)
	}
	return enrollmentStatusHistories, nil
}

func findEnrollmentStatusHistoryByStatus(ctx context.Context, db database.QueryExecer, studentID string, enrollmentStatus string) (*repository.EnrollmentStatusHistory, error) {
	var history repository.EnrollmentStatusHistory
	enrollmentStatusHistories, err := getAllEnrollmentStatusHistoriesOfStudent(ctx, db, studentID)
	if err != nil {
		return nil, errors.Wrap(err, "getAllEnrollmentStatusHistoriesOfStudent failed")
	}
	for _, enrollmentStatusHistory := range enrollmentStatusHistories {
		if enrollmentStatusHistory.EnrollmentStatus().String() == enrollmentStatus {
			history = enrollmentStatusHistory
			break
		}
	}
	if history.UserID().IsEmpty() {
		return nil, errors.New("history not found")
	}
	return &history, nil
}

func getUniqueInfo(history repository.EnrollmentStatusHistory) string {
	return "" +
		history.EnrollmentStatusHistoryAttribute.UserID.String() + "." +
		history.EnrollmentStatusHistoryAttribute.LocationID.String() + "." +
		history.EnrollmentStatusHistoryAttribute.EnrollmentStatus.String() + "." +
		history.EnrollmentStatusHistoryAttribute.StartDate.Time().Format(time.DateOnly) + "." +
		history.EnrollmentStatusHistoryAttribute.EndDate.Time().Format(time.DateOnly) + "." +
		history.EnrollmentStatusHistoryAttribute.OrderID.String() + "." +
		history.DeletedAtAttr.Time().Format(time.DateOnly)
}
