package service

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *InternalService) UpdateStudentCourse(ctx context.Context, req *pb.UpdateStudentCourseRequest) (res *pb.UpdateStudentCourseResponse, err error) {
	/*	startTime := req.To.AsTime().Truncate(24 * time.Hour)
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
			err = s.packageService.UpdateStudentPackageByOrderAndStudentCourseByStartTime(ctx, tx, startTime)
			return
		})
		if err != nil {
			res = &pb.UpdateStudentCourseResponse{
				Successful: false,
			}
			return
		}*/

	res = &pb.UpdateStudentCourseResponse{
		Successful: true,
	}
	return
}
