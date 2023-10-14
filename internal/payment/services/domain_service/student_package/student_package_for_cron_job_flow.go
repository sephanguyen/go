package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetStudentPackagesForCronJob
// Get student packages with end_at is in 3 latest days
func (s *StudentPackageService) GetStudentPackagesForCronJob(ctx context.Context, db database.QueryExecer) (
	studentPackages []entities.StudentPackages,
	err error,
) {
	studentPackages, err = s.StudentPackageRepo.GetStudentPackagesForCronjobByDay(ctx, db, 3)
	if err != nil {
		err = status.Errorf(codes.Internal, fmt.Sprintf("error when getting student packages for cron job: %s", err.Error()))
		return
	}
	return
}

func (s *StudentPackageService) UpsertStudentPackageDataForCronjob(ctx context.Context,
	db database.QueryExecer,
	studentPackage entities.StudentPackages,
) (eventMessage *npb.EventStudentPackage, currentStudentPackageOrder *entities.StudentPackageOrder, err error) {
	var (
		newStudentPackage entities.StudentPackages
		newStudentCourse  entities.StudentCourse
		action            = pb.StudentPackageActions_STUDENT_PACKAGE_ACTION_UPSERT.String()
		flow              = "Cronjob update student package"
	)

	currentStudentPackageOrder, err = s.StudentPackageOrderService.SetCurrentStudentPackageOrderByTimeAndStudentPackageID(ctx, db, studentPackage.ID.String)
	if err != nil {
		err = status.Errorf(codes.Internal, fmt.Sprintf("error when set current student package order with student_package_id=%s: %s", studentPackage.ID.String, err.Error()))
		return
	}
	if currentStudentPackageOrder == nil {
		return
	}
	// There is new current student package order for this student package and need to be upsert
	if currentStudentPackageOrder.StartAt.Status == pgtype.Present &&
		currentStudentPackageOrder.StartAt.Time.After(studentPackage.EndAt.Time) {
		newStudentPackage, newStudentCourse, eventMessage, err = s.convertStudentPackageDataByStudentPackageOrder(ctx, db, *currentStudentPackageOrder)
		if err != nil {
			err = status.Errorf(codes.Internal, fmt.Sprintf("error UpsertStudentPackageDataForCronjob.convertStudentPackageDataByStudentPackageOrder with student_package_id=%s: %s", studentPackage.ID.String, err.Error()))
			return
		}
		err = utils.GroupErrorFunc(
			s.StudentPackageRepo.Upsert(ctx, db, &newStudentPackage),
			s.StudentCourseRepo.UpsertStudentCourse(ctx, db, newStudentCourse),
			s.writeStudentPackageLog(ctx, db, &studentPackage, newStudentCourse.CourseID.String, action, flow),

			currentStudentPackageOrder.IsExecutedByCronJob.Set(true),
			s.StudentPackageOrderService.UpdateExecuteStatus(ctx, db, *currentStudentPackageOrder),
		)
		if err != nil {
			err = status.Errorf(codes.Internal, fmt.Sprintf("error upsert student package data by cron job with student_package_id=%s and student_package_order_id=%s: %s", studentPackage.ID.String, currentStudentPackageOrder.ID.String, err.Error()))
			return
		}
	}
	return
}
