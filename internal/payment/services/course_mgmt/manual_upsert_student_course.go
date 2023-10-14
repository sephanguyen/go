package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CourseMgMt) ManualUpsertStudentCourse(ctx context.Context, req *pb.ManualUpsertStudentCourseRequest) (res *pb.ManualUpsertStudentCourseResponse, err error) {
	var (
		mapUserAccess map[string]interface{}
		studentID     string
		mapCourse     map[string]interface{}
		defaultValue  interface{}
	)
	studentID = req.StudentId
	mapUserAccess, err = s.StudentService.GetMapLocationAccessStudentByStudentIDs(ctx, s.DB, []string{studentID})
	if err != nil {
		err = status.Errorf(codes.Internal, "when get map user access path have err %v", err.Error())
		return
	}
	var events []*npb.EventStudentPackage
	mapCourse = map[string]interface{}{}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		for _, studentCourse := range req.StudentCourses {
			if !studentCourse.IsChanged {
				mapCourse[studentCourse.CourseId] = defaultValue
				continue
			}

			if _, ok := mapCourse[studentCourse.CourseId]; ok {
				err = status.Errorf(codes.FailedPrecondition, constant.DuplicateCourseByManualError)
				return
			}

			var event *npb.EventStudentPackage
			if studentCourse.StudentPackageId != nil {
				event, err = s.StudentPackage.UpdateTimeStudentPackageForManualFlow(ctx, tx, studentID, studentCourse)
				if err != nil {
					return
				}
				events = append(events, event)
				continue
			}

			key := fmt.Sprintf("%v_%v", studentCourse.LocationId, studentID)
			if _, ok := mapUserAccess[key]; !ok {
				err = status.Errorf(codes.FailedPrecondition, constant.UserCantAccessThisCourse)
				return
			}
			event, err = s.StudentPackage.UpsertStudentPackageForManualFlow(ctx, tx, studentID, studentCourse)
			if err != nil {
				return
			}
			events = append(events, event)
		}
		return
	})
	if err != nil {
		return
	}
	if len(events) == 0 {
		res = &pb.ManualUpsertStudentCourseResponse{
			Successful: true,
		}
		return
	}
	err = s.SubscriptionService.PublishStudentPackage(ctx, events)
	if err != nil {
		return
	}
	res = &pb.ManualUpsertStudentCourseResponse{
		Successful: true,
	}
	return
}
