package ordermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	studentService "github.com/manabie-com/backend/internal/payment/services/domain_service/student"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type GetOrgLevelStudentStatus struct {
	DB             database.Ext
	StudentService IGetOrgLevelStudentStatus
}

type IGetOrgLevelStudentStatus interface {
	GetEnrolledStatusInOrgByStudentInfo(ctx context.Context, db database.QueryExecer, req *pb.GetOrgLevelStudentStatusRequest) (result []*pb.GetOrgLevelStudentStatusResponse_OrgLevelStudentStatus, err error)
}

func (s *GetOrgLevelStudentStatus) GetOrgLevelStudentStatus(ctx context.Context, req *pb.GetOrgLevelStudentStatusRequest) (res *pb.GetOrgLevelStudentStatusResponse, err error) {
	orgLevelStudentStatus, err := s.StudentService.GetEnrolledStatusInOrgByStudentInfo(ctx, s.DB, req)
	if err != nil {
		err = fmt.Errorf("error when get organization enrollment status by student info: %v", err)
		return
	}
	res = &pb.GetOrgLevelStudentStatusResponse{}
	res.StudentStatus = orgLevelStudentStatus
	return
}

func NewGetOrgLevelStudentStatus(db database.Ext) *GetOrgLevelStudentStatus {
	return &GetOrgLevelStudentStatus{
		DB:             db,
		StudentService: studentService.NewStudentService(),
	}
}
