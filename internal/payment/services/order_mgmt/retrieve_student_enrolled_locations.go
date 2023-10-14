package ordermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	studentService "github.com/manabie-com/backend/internal/payment/services/domain_service/student"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type RetrieveStudentEnrolledLocations struct {
	DB             database.Ext
	StudentService IRetrieveStudentEnrolledLocations
}

type IRetrieveStudentEnrolledLocations interface {
	GetStudentEnrolledLocationsByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (result []*pb.RetrieveStudentEnrolledLocationsResponse_StudentStatusPerLocation, err error)
}

func (s *RetrieveStudentEnrolledLocations) RetrieveStudentEnrolledLocations(ctx context.Context, req *pb.RetrieveStudentEnrolledLocationsRequest) (res *pb.RetrieveStudentEnrolledLocationsResponse, err error) {
	studentEnrollmentInfo, err := s.StudentService.GetStudentEnrolledLocationsByStudentID(ctx, s.DB, req.StudentId)
	if err != nil {
		err = fmt.Errorf("error when get student enrollment status by student ID: %v", err)
		return
	}
	res = &pb.RetrieveStudentEnrolledLocationsResponse{}
	res.StudentId = req.StudentId
	res.StudentStatusPerLocation = studentEnrollmentInfo
	return
}

func NewRetrieveStudentEnrolledLocations(db database.Ext) *RetrieveStudentEnrolledLocations {
	return &RetrieveStudentEnrolledLocations{
		DB:             db,
		StudentService: studentService.NewStudentService(),
	}
}
