package payment

import (
	"context"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) prepareDataForGetOrgLevelStudentStatusValidRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		locationID string
		userID     string

		req pb.CreateOrderRequest
		err error
	)
	defaultPrepareDataSettings := PrepareDataForCreatingOrderSettings{
		insertTax:             true,
		insertDiscount:        true,
		insertStudent:         true,
		insertMaterial:        true,
		insertProductPrice:    true,
		insertProductLocation: true,
		insertLocation:        false,
		insertProductGrade:    true,
		insertFee:             false,
		insertProductDiscount: true,
		insertMaterialUnique:  false,
	}
	_,
		_,
		locationID,
		_,
		userID,
		err = s.insertAllDataForInsertOrder(ctx, defaultPrepareDataSettings, OneTimeMaterial)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = s.insertStudentEnrollmentStatusHistory(ctx, userID, locationID, upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req.StudentId = userID
	req.LocationId = locationID
	req.OrderType = pb.OrderType_ORDER_TYPE_NEW
	stepState.Request = &req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertStudentEnrollmentStatusHistory(ctx context.Context, studentID string, locationID string, enrollmentStatus string) error {
	insertStudentEnrollmentStatusHistoryStmt := `INSERT INTO student_enrollment_status_history (student_id, location_id, enrollment_status, start_date) VALUES ($1, $2, $3, now())`

	_, err := s.FatimaDBTrace.Exec(ctx, insertStudentEnrollmentStatusHistoryStmt, studentID, locationID, enrollmentStatus)
	if err != nil {
		return fmt.Errorf("can not insert product_discount, err: %s", err)
	}

	return nil
}

func (s *suite) getOrgEnrollmentStatus(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	reqCreateOrder := stepState.Request.(*pb.CreateOrderRequest)
	req := &pb.GetOrgLevelStudentStatusRequest{
		StudentInfo: []*pb.GetOrgLevelStudentStatusRequestStudentInfo{
			{
				StudentId: reqCreateOrder.StudentId,
			},
		},
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	resp, err := client.GetOrgLevelStudentStatus(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkDataWhenGetOrgEnrollmentStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetOrgLevelStudentStatusResponse)

	if len(resp.StudentStatus) == 0 {
		return ctx, fmt.Errorf("error response: empty data with valid data")
	}

	for _, studentStatus := range resp.StudentStatus {
		if !studentStatus.IsEnrolledInOrg {
			return ctx, fmt.Errorf("expect IsEnrolledInOrg is true")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudentEnrolledLocation(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	reqCreateOrder := stepState.Request.(*pb.CreateOrderRequest)
	req := &pb.RetrieveStudentEnrolledLocationsRequest{
		StudentId: reqCreateOrder.StudentId,
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	resp, err := client.RetrieveStudentEnrolledLocations(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkDataWhenGetStudentEnrolledLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.RetrieveStudentEnrolledLocationsResponse)

	if len(resp.StudentStatusPerLocation) == 0 {
		return ctx, fmt.Errorf("error response: empty data with valid data")
	}

	for _, studentStatus := range resp.StudentStatusPerLocation {
		if studentStatus.HasScheduledChangeOfStatusInLocation {
			return ctx, fmt.Errorf("expect HasScheduledChangeOfStatusInLocation is false")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudentEnrollmentStatusByLocation(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	reqCreateOrder := stepState.Request.(*pb.CreateOrderRequest)
	req := &pb.RetrieveStudentEnrollmentStatusByLocationRequest{
		StudentLocations: []*pb.RetrieveStudentEnrollmentStatusByLocationRequest_StudentLocation{
			{
				StudentId:  reqCreateOrder.StudentId,
				LocationId: reqCreateOrder.LocationId,
			},
		},
	}

	client := pb.NewOrderServiceClient(s.PaymentConn)
	stepState.RequestSentAt = time.Now()
	resp, err := client.RetrieveStudentEnrollmentStatusByLocation(contextWithToken(ctx), req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkDataWhenGetStudentEnrollmentStatusByLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.RetrieveStudentEnrollmentStatusByLocationResponse)

	if len(resp.StudentStatusPerLocation) == 0 {
		return ctx, fmt.Errorf("error response: empty data with valid data")
	}

	for _, studentStatus := range resp.StudentStatusPerLocation {
		if !studentStatus.IsEnrollment {
			return ctx, fmt.Errorf("expect HasScheduledChangeOfStatusInLocation is true")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
