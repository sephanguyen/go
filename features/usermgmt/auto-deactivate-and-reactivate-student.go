package usermgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	activated   = "activated"
	deactivated = "deactivated"
)

var (
	mapStatusBDDToEnrollmentStatus = map[string]pb.StudentEnrollmentStatus{
		"TEMPORARY":     pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
		"LOA":           pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA,
		"ENROLLED":      pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
		"POTENTIAL":     pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
		"WITHDRAWN":     pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN,
		"GRADUATED":     pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED,
		"NON-POTENTIAL": pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL,
	}
)

func (s *suite) upsertStudentWithEnrollmentStatuses(ctx context.Context, activeStatusAmount, activeStatus, inActiveStatusAmount, inActiveStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	orgID := OrgIDFromCtx(ctx)
	uid := idutil.ULIDNow()
	studentProfile := &pb.StudentProfileV2{
		FirstName: "FirstName" + uid,
		LastName:  "LastName" + uid,
		Email:     uid + "student@email.com",
		Username:  uid + "username",
		GradeId:   fmt.Sprintf("%d_grade_01", orgID),
		Password:  "123456",
		StudentPhoneNumbers: &pb.StudentPhoneNumbers{
			ContactPreference: pb.StudentContactPreference_STUDENT_HOME_PHONE_NUMBER,
		},
	}
	enrollmentStatusHistories := []*pb.EnrollmentStatusHistory{}
	activeAmount, _ := strconv.Atoi(activeStatusAmount)

	activeStatuses := []string{}
	if activeStatus != "" {
		activeStatuses = strings.Split(activeStatus, ",")
	}

	for i, status := range activeStatuses {
		if mapStatusBDDToEnrollmentStatus[status] == pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE {
			continue
		}
		startDate := timestamppb.New(time.Now().AddDate(0, 0, -i))
		enrollmentStatusHistory := &pb.EnrollmentStatusHistory{
			LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[len(enrollmentStatusHistories)+1]),
			EnrollmentStatus: mapStatusBDDToEnrollmentStatus[status],
			StartDate:        startDate,
		}
		enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistory)
	}

	inActiveStatuses := []string{}
	if inActiveStatus != "" {
		inActiveStatuses = strings.Split(inActiveStatus, ",")
	}
	for i, status := range inActiveStatuses {
		if mapStatusBDDToEnrollmentStatus[status] == pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE {
			continue
		}
		startDate := timestamppb.New(time.Now().AddDate(0, 0, 1+i))
		enrollmentStatusHistory := &pb.EnrollmentStatusHistory{
			LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[len(enrollmentStatusHistories)+1]),
			EnrollmentStatus: mapStatusBDDToEnrollmentStatus[status],
			StartDate:        startDate,
		}
		enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistory)
	}

	switch activeStatus {
	case "non-withdrawn":
		for i := 0; i < activeAmount; i++ {
			enrollmentStatusHistory := &pb.EnrollmentStatusHistory{
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[len(enrollmentStatusHistories)+1]),
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
				StartDate:        timestamppb.Now(),
			}
			enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistory)
		}
	case "withdrawn":
		for i := 0; i < activeAmount; i++ {
			startDate := timestamppb.Now()
			if i >= 1 {
				startDate = timestamppb.New(time.Now().AddDate(0, 0, -1))
			}
			enrollmentStatusHistory := &pb.EnrollmentStatusHistory{
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[len(enrollmentStatusHistories)+1]),
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN,
				StartDate:        startDate,
			}
			enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistory)
		}
	case "non-withdrawn and withdrawn":
		for i := 0; i < activeAmount; i++ {
			enrollmentStatusHistory := []*pb.EnrollmentStatusHistory{
				{
					LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[len(enrollmentStatusHistories)+1]),
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL,
					StartDate:        timestamppb.Now(),
				},
				{
					LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[len(enrollmentStatusHistories)+2]),
					EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN,
					StartDate:        timestamppb.Now(),
				},
			}
			enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistory...)
		}
	}
	inActiveAmount, _ := strconv.Atoi(inActiveStatusAmount)
	switch inActiveStatus {
	case "non-withdrawn":
		for i := 0; i < inActiveAmount; i++ {
			enrollmentStatusHistory := &pb.EnrollmentStatusHistory{
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[len(enrollmentStatusHistories)+1]),
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
				StartDate:        timestamppb.New(time.Now().AddDate(0, 0, 1)),
			}
			enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistory)
		}
	case "withdrawn":
		for i := 0; i < inActiveAmount; i++ {
			enrollmentStatusHistory := &pb.EnrollmentStatusHistory{
				LocationId:       fmt.Sprintf("%d_%s", orgID, s.LocationIDs[len(enrollmentStatusHistories)+1]),
				EnrollmentStatus: pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN,
				StartDate:        timestamppb.New(time.Now().AddDate(0, 0, 1)),
			}
			enrollmentStatusHistories = append(enrollmentStatusHistories, enrollmentStatusHistory)
		}
	}
	studentProfile.EnrollmentStatusHistories = enrollmentStatusHistories
	studentProfiles := []*pb.StudentProfileV2{studentProfile}

	req := &pb.UpsertStudentRequest{
		StudentProfiles: studentProfiles,
	}

	resp, err := pb.NewStudentServiceClient(s.UserMgmtConn).UpsertStudent(ctx, req)
	if err != nil {
		return ctx, fmt.Errorf("UpsertStudent: %v", err)
	}
	stepState.Request = req
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) assertUserActivation(ctx context.Context, isDeactivated string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpsertStudentRequest)
	resp := stepState.Response.(*pb.UpsertStudentResponse)
	enrollmentStatusHistoriesReq := req.StudentProfiles[0].EnrollmentStatusHistories

	userID := resp.StudentProfiles[0].Id

	if err := try.Do(func(attempt int) (retry bool, err error) {
		deactivatedAt := field.NewNullTime()
		query := "SELECT deactivated_at FROM users WHERE user_id = $1"
		row := s.BobDBTrace.QueryRow(ctx, query, userID)
		err = row.Scan(&deactivatedAt)
		if err == nil {
			latestStartDate := enrollmentStatusHistoriesReq[0].StartDate.AsTime()
			for _, enrollmentStatus := range enrollmentStatusHistoriesReq {
				if !enrollmentStatus.StartDate.AsTime().After(time.Now()) && enrollmentStatus.StartDate.AsTime().After(latestStartDate) {
					latestStartDate = enrollmentStatus.StartDate.AsTime()
				}
			}

			if isDeactivated == activated && field.IsPresent(deactivatedAt) {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("student is not activated : %v", deactivatedAt.Time().String())
			}
			if isDeactivated == deactivated && latestStartDate.Truncate(time.Hour).String() != deactivatedAt.Time().Truncate(time.Hour).String() {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("student deactivatedAt is not latest withdrawn start date : %v, latestStartDate: %v, deactivatedAt: %v", deactivatedAt.Time().String(), latestStartDate.String(), deactivatedAt.Time().String())
			}
		}
		return false, errors.Wrap(err, "can not get deactivated_at of user: %v")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentWithActiveStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	activeStatus := "withdrawn"

	if status == activated {
		activeStatus = "non-withdrawn"
	}
	if status != deactivated && status != activated {
		activeStatus = status
	}
	if _, err := s.upsertStudentWithEnrollmentStatuses(ctx, "1", activeStatus, "0", "withdrawn"); err != nil {
		return ctx, fmt.Errorf("upsertStudentWithEnrollmentStatuses error : %v", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) syncOrderToDeactivateAndReactivateStudents(ctx context.Context, orderFunction string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.UpsertStudentResponse)
	studentID := resp.StudentProfiles[0].Id
	orderStatus := "ORDER_STATUS_SUBMITTED"
	orderType := "ORDER_TYPE_ENROLLMENT"
	enrollmentStatus := "STUDENT_ENROLLMENT_STATUS_ENROLLED"
	locationID := resp.StudentProfiles[0].EnrollmentStatusHistories[0].LocationId
	switch orderFunction {
	case "Order":
		orderType = "ORDER_TYPE_NEW"
		enrollmentStatus = "STUDENT_ENROLLMENT_STATUS_TEMPORARY"
		for _, id := range s.LocationIDs {
			if locationID != id {
				locationID = fmt.Sprintf("%d_%s", OrgIDFromCtx(ctx), id)
			}
		}
	case "Graduate Order":
		enrollmentStatus = "STUDENT_ENROLLMENT_STATUS_GRADUATED"
		orderType = "ORDER_TYPE_GRADUATE"

	case "Withdrawal Request":
		enrollmentStatus = "STUDENT_ENROLLMENT_STATUS_WITHDRAWN"
		orderType = "ORDER_TYPE_WITHDRAWAL"
	}
	orderEventLog := service.OrderEventLog{
		OrderStatus:      orderStatus,
		OrderType:        orderType,
		StudentID:        studentID,
		LocationID:       locationID,
		EnrollmentStatus: enrollmentStatus,
		StartDate:        time.Now(),
	}
	enrollmentStatusHistories := []*pb.EnrollmentStatusHistory{
		{
			LocationId:       locationID,
			EnrollmentStatus: pb.StudentEnrollmentStatus(pb.StudentEnrollmentStatus_value[enrollmentStatus]),
			StartDate:        timestamppb.New(orderEventLog.StartDate),
		},
	}
	stepState.Request = &pb.UpsertStudentRequest{
		StudentProfiles: []*pb.StudentProfileV2{
			{
				EnrollmentStatusHistories: enrollmentStatusHistories,
			},
		},
	}
	data, err := json.Marshal(orderEventLog)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectOrderEventLogCreated, data)
	if err != nil {
		return StepStateToContext(ctx, stepState), nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishOrderEventLog JSM.PublishContext failed, msgID: %s, %w", msgID, err))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertStudentWithStatusWillBe(ctx context.Context, statusWillBe string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	activeStatus := "withdrawn"
	deactivatedAt := database.TimestamptzNull(time.Time{})
	if statusWillBe == activated {
		activeStatus = "non-withdrawn"
		deactivatedAt = database.Timestamptz(time.Now())
	}

	if statusWillBe != deactivated && statusWillBe != activated {
		activeStatus = statusWillBe
	}
	if _, err := s.upsertStudentWithEnrollmentStatuses(ctx, "1", activeStatus, "0", "withdrawn"); err != nil {
		return ctx, fmt.Errorf("upsertStudentWithEnrollmentStatuses error : %v", err)
	}

	resp := stepState.Response.(*pb.UpsertStudentResponse)
	studentID := resp.StudentProfiles[0].Id
	query := "UPDATE users SET deactivated_at = $1 where user_id = $2"
	_, err := s.BobDBTrace.Exec(ctx, query, deactivatedAt, studentID)
	if err != nil {
		return ctx, fmt.Errorf("can not update deactivated_at of user: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}
