package ordermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	studentService "github.com/manabie-com/backend/internal/payment/services/domain_service/student"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type RetrieveStudentEnrollmentStatusByLocation struct {
	DB             database.Ext
	StudentService IRetrieveStudentEnrollmentStatusByLocation
}

type IRetrieveStudentEnrollmentStatusByLocation interface {
	IsStudentEnrolledInLocation(ctx context.Context, db database.QueryExecer, req *pb.RetrieveStudentEnrollmentStatusByLocationRequest) (result []*pb.RetrieveStudentEnrollmentStatusByLocationResponse_StudentStatusPerLocation, err error)
}

func (s *RetrieveStudentEnrollmentStatusByLocation) RetrieveStudentEnrollmentStatusByLocation(ctx context.Context, req *pb.RetrieveStudentEnrollmentStatusByLocationRequest) (res *pb.RetrieveStudentEnrollmentStatusByLocationResponse, err error) {
	studentEnrollmentInfo, err := s.StudentService.IsStudentEnrolledInLocation(ctx, s.DB, req)
	if err != nil {
		err = fmt.Errorf("error when get student enrollment status: %v", err)
		return
	}
	res = &pb.RetrieveStudentEnrollmentStatusByLocationResponse{}
	res.StudentStatusPerLocation = studentEnrollmentInfo
	return
}

func NewRetrieveStudentEnrollmentStatusByLocation(db database.Ext) *RetrieveStudentEnrollmentStatusByLocation {
	return &RetrieveStudentEnrollmentStatusByLocation{
		DB:             db,
		StudentService: studentService.NewStudentService(),
	}
}
