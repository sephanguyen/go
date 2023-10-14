package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentService struct {
	studentRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, studentID string) (entities.Student, error)
	}
	userRepo interface {
		GetStudentByIDForUpdate(ctx context.Context, db database.QueryExecer, id string) (entities.User, error)
	}
	userAccessRepo interface {
		GetUserAccessPathByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (mapUserAccess map[string]interface{}, err error)
	}
	studentEnrollmentStatusHistoryRepo interface {
		GetLatestStatusByStudentIDAndLocationID(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (entities.StudentEnrollmentStatusHistory, error)
		GetCurrentStatusByStudentIDAndLocationID(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (entities.StudentEnrollmentStatusHistory, error)
		GetListStudentEnrollmentStatusHistoryByStudentID(
			ctx context.Context,
			db database.QueryExecer,
			studentID string,
		) (
			[]*entities.StudentEnrollmentStatusHistory,
			error,
		)
		GetListEnrolledStudentEnrollmentStatusByStudentID(ctx context.Context, db database.QueryExecer, StudentID string) ([]*entities.StudentEnrollmentStatusHistory, error)
		GetListEnrolledStatusByStudentIDAndTime(ctx context.Context, db database.QueryExecer, StudentID string, time2 time.Time) ([]*entities.StudentEnrollmentStatusHistory, error)
	}
	studentProductRepo interface {
		GetByID(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.StudentProduct, error)
	}
}

func (s *StudentService) GetStudentAndNameByID(ctx context.Context, db database.QueryExecer, studentID string) (
	student entities.Student,
	studentName string,
	err error,
) {
	student, err = s.studentRepo.GetByIDForUpdate(ctx, db, studentID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when checking student id: %v", err.Error())
		return
	}

	if student.GradeID.Status != pgtype.Present {
		err = status.Errorf(codes.FailedPrecondition, "can't create order because this student have empty grade")
		return
	}

	user, err := s.userRepo.GetStudentByIDForUpdate(ctx, db, studentID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when checking student id: %v", err.Error())
	}
	studentName = user.Name.String

	return
}

func (s *StudentService) GetMapLocationAccessStudentByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (mapLocationAccessStudent map[string]interface{}, err error) {
	mapLocationAccessStudent, err = s.userAccessRepo.GetUserAccessPathByUserIDs(ctx, db, studentIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get user access path by student id s with err: %v", err.Error())
	}
	return
}

func (s *StudentService) IsEnrolledInLocation(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (bool, error) {
	studentID := orderItemData.StudentInfo.StudentID.String
	locationID := orderItemData.Order.LocationID.String

	latestEnrollmentStatusHistory, err := s.studentEnrollmentStatusHistoryRepo.GetCurrentStatusByStudentIDAndLocationID(ctx, db, studentID, locationID)
	if err != nil {
		return false, nil
	}
	if latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String() {
		return true, nil
	}

	return false, nil
}

func (s *StudentService) IsEnrolledInOrg(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (bool, error) {
	studentID := orderItemData.StudentInfo.StudentID.String
	now := time.Now()

	enrolledStudentEnrollmentStatusHistoryList, err := s.studentEnrollmentStatusHistoryRepo.GetListEnrolledStatusByStudentIDAndTime(ctx, db, studentID, now)
	if err != nil {
		return false, err
	}
	if len(enrolledStudentEnrollmentStatusHistoryList) != 0 {
		return true, nil
	}
	return false, nil
}

func (s *StudentService) ValidateStudentStatusForOrderType(ctx context.Context, db database.QueryExecer, orderType pb.OrderType, student entities.Student, locationID string, effectedDate time.Time) (err error) {
	emptyEnrollmentHistory := entities.StudentEnrollmentStatusHistory{}
	latestEnrollmentStatusHistory, err := s.studentEnrollmentStatusHistoryRepo.GetCurrentStatusByStudentIDAndLocationID(ctx, db, student.StudentID.String, locationID)

	if latestEnrollmentStatusHistory == emptyEnrollmentHistory {
		return validateStudentStatusFromStudentEntity(student)
	}

	switch orderType {
	case pb.OrderType_ORDER_TYPE_ENROLLMENT:
		if latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE.String() {
			return utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatus, nil)
		}

		if latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String() {
			return utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatusAlreadyEnrolled, nil)
		}

		if latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA.String() {
			return utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatusOnLOA, nil)
		}
	case pb.OrderType_ORDER_TYPE_WITHDRAWAL, pb.OrderType_ORDER_TYPE_GRADUATE:
		if latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String() ||
			latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String() ||
			latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String() ||
			latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String() {
			return utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", latestEnrollmentStatusHistory.EnrollmentStatus.String),
				},
			)
		}
	case pb.OrderType_ORDER_TYPE_LOA:
		if latestEnrollmentStatusHistory.EnrollmentStatus.String != upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String() {
			return utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", latestEnrollmentStatusHistory.EnrollmentStatus.String),
				},
			)
		}
	case pb.OrderType_ORDER_TYPE_RESUME:
		if latestEnrollmentStatusHistory.EnrollmentStatus.String != upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA.String() {
			return utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.InvalidStudentEnrollmentStatusUnavailable,
				&errdetails.DebugInfo{
					Detail: fmt.Sprintf("Student enrollment status is %s", latestEnrollmentStatusHistory.EnrollmentStatus.String),
				},
			)
		}
	}

	return
}

func validateStudentStatusFromStudentEntity(student entities.Student) (err error) {
	if student.EnrollmentStatus.Status == pgtype.Null {
		return utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatus, nil)
	}

	if student.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NONE.String() {
		return utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatus, nil)
	}

	if student.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String() {
		return utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatusAlreadyEnrolled, nil)
	}

	if student.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA.String() {
		return utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InvalidStudentEnrollmentStatusOnLOA, nil)
	}
	return
}

func (s *StudentService) IsAllowedToOrderEnrollmentRequiredProducts(ctx context.Context, db database.QueryExecer, studentID string, locationID string) (isAllowedToOrder bool, err error) {
	var latestEnrollmentStatusHistory entities.StudentEnrollmentStatusHistory
	emptyEnrollmentHistory := entities.StudentEnrollmentStatusHistory{}

	latestEnrollmentStatusHistory, _ = s.studentEnrollmentStatusHistoryRepo.GetLatestStatusByStudentIDAndLocationID(ctx, db, studentID, locationID)

	if latestEnrollmentStatusHistory == emptyEnrollmentHistory {
		return false, nil
	}

	if latestEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String() {
		return true, nil
	}

	return false, nil
}

func (s *StudentService) IsStudentEnrolledInLocation(ctx context.Context, db database.QueryExecer, req *pb.RetrieveStudentEnrollmentStatusByLocationRequest) (result []*pb.RetrieveStudentEnrollmentStatusByLocationResponse_StudentStatusPerLocation, err error) {
	for _, studentLocation := range req.StudentLocations {
		studentEnrollmentInfo := pb.RetrieveStudentEnrollmentStatusByLocationResponse_StudentStatusPerLocation{
			StudentId:    studentLocation.StudentId,
			LocationId:   studentLocation.LocationId,
			IsEnrollment: false,
		}

		studentEnrollmentStatusHistory, err := s.studentEnrollmentStatusHistoryRepo.GetCurrentStatusByStudentIDAndLocationID(ctx, db, studentLocation.StudentId, studentLocation.LocationId)
		if err != nil {
			if err == pgx.ErrNoRows {
				result = append(result, &studentEnrollmentInfo)
				continue
			}
			return nil, err
		}
		if studentEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA.String() || studentEnrollmentStatusHistory.EnrollmentStatus.String == upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String() {
			studentEnrollmentInfo.IsEnrollment = true
		}
		result = append(result, &studentEnrollmentInfo)
	}
	return
}

func (s *StudentService) GetStudentEnrolledLocationsByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (result []*pb.RetrieveStudentEnrolledLocationsResponse_StudentStatusPerLocation, err error) {
	studentEnrollmentStatusHistoryList, err := s.studentEnrollmentStatusHistoryRepo.GetListStudentEnrollmentStatusHistoryByStudentID(ctx, db, studentID)
	if err != nil {
		return nil, err
	}
	for _, studentEnrollmentStatusHistory := range studentEnrollmentStatusHistoryList {
		studentEnrollmentStatusCurrent, err := s.studentEnrollmentStatusHistoryRepo.GetCurrentStatusByStudentIDAndLocationID(ctx, db, studentID, studentEnrollmentStatusHistory.LocationID.String)
		if err != nil {
			return nil, err
		}
		studentStatusPerLocation := pb.RetrieveStudentEnrolledLocationsResponse_StudentStatusPerLocation{
			LocationId:                           studentEnrollmentStatusHistory.LocationID.String,
			StudentStatus:                        studentEnrollmentStatusCurrent.EnrollmentStatus.String,
			HasScheduledChangeOfStatusInLocation: false,
		}

		if studentEnrollmentStatusHistory.StartDate.Time.After(time.Now()) {
			studentStatusPerLocation.HasScheduledChangeOfStatusInLocation = true
		}

		result = append(result, &studentStatusPerLocation)
	}
	return
}

func (s *StudentService) GetEnrolledStatusInOrgByStudentInfo(ctx context.Context, db database.QueryExecer, req *pb.GetOrgLevelStudentStatusRequest) (result []*pb.GetOrgLevelStudentStatusResponse_OrgLevelStudentStatus, err error) {
	for _, studentInfo := range req.StudentInfo {
		studentOrgEnrollmentStatus := pb.GetOrgLevelStudentStatusResponse_OrgLevelStudentStatus{
			StudentId:       studentInfo.StudentId,
			IsEnrolledInOrg: false,
		}
		enrolledStudentEnrollmentStatusList, err := s.studentEnrollmentStatusHistoryRepo.GetListEnrolledStudentEnrollmentStatusByStudentID(ctx, db, studentInfo.StudentId)
		if err != nil {
			return nil, err
		}

		if len(enrolledStudentEnrollmentStatusList) != 0 {
			studentOrgEnrollmentStatus.IsEnrolledInOrg = true
			if studentInfo.StudentProductId != nil {
				isEnrolledInOrg, err := s.CheckIsEnrolledInOrg(ctx, db, studentInfo.StudentProductId.GetValue(), enrolledStudentEnrollmentStatusList)
				if err != nil {
					return nil, err
				}
				studentOrgEnrollmentStatus.StudentProductId = studentInfo.StudentProductId
				studentOrgEnrollmentStatus.IsEnrolledInOrg = isEnrolledInOrg
			}
		}
		result = append(result, &studentOrgEnrollmentStatus)
	}
	return
}

func (s *StudentService) CheckIsEnrolledInOrg(
	ctx context.Context,
	db database.QueryExecer,
	studentProductID string,
	studentEnrollmentStatusList []*entities.StudentEnrollmentStatusHistory,
) (isEnrolledInOrg bool, err error) {
	studentProduct, err := s.studentProductRepo.GetByID(ctx, db, studentProductID)
	if err != nil {
		return
	}

	if studentProduct.RootStudentProductID.Status == pgtype.Present {
		studentProduct, err = s.studentProductRepo.GetByID(ctx, db, studentProduct.RootStudentProductID.String)
		if err != nil {
			return
		}
	}

	for _, studentEnrollmentStatus := range studentEnrollmentStatusList {
		if studentEnrollmentStatus.StartDate.Time.Before(studentProduct.CreatedAt.Time) &&
			(studentEnrollmentStatus.EndDate.Status != pgtype.Present ||
				studentEnrollmentStatus.EndDate.Time.After(studentProduct.CreatedAt.Time)) {
			isEnrolledInOrg = true
			return
		}
	}
	return
}

// CheckIsEnrolledInOrgByStudentIDAndTime ...
// To check if student is enrolled in org or not in range Time
func (s *StudentService) CheckIsEnrolledInOrgByStudentIDAndTime(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	time time.Time,
) (isEnrolledInOrg bool, err error) {
	enrolledStudentEnrollmentStatusList, err := s.studentEnrollmentStatusHistoryRepo.GetListEnrolledStatusByStudentIDAndTime(ctx, db, studentID, time)
	if err != nil {
		return false, err
	}
	if len(enrolledStudentEnrollmentStatusList) > 0 {
		isEnrolledInOrg = true
		return
	}
	return
}

func NewStudentService() *StudentService {
	return &StudentService{
		studentRepo:                        &repositories.StudentRepo{},
		userRepo:                           &repositories.UserRepo{},
		userAccessRepo:                     &repositories.UserAccessPathRepo{},
		studentEnrollmentStatusHistoryRepo: &repositories.StudentEnrollmentStatusHistoryRepo{},
		studentProductRepo:                 &repositories.StudentProductRepo{},
	}
}
