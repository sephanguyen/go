package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

func (svc *NotificationModifierService) SyncJprepStudentPackage(ctx context.Context, data []*npb.EventSyncStudentPackage_StudentPackage) error {
	err := database.ExecInTx(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := multierr.Combine(
			svc.upsertJprepStudentPackage(ctx, data, tx),
		)
		return err
	})

	if err != nil {
		return err
	}
	return nil
}

func (svc *NotificationModifierService) upsertJprepStudentPackage(ctx context.Context, studentPackages []*npb.EventSyncStudentPackage_StudentPackage, tx pgx.Tx) error {
	for _, studentPackage := range studentPackages {
		switch studentPackage.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			notiStudentCourses, err := mappers.EventStudentPackageJPRPEPbToNotificationStudentCourseEnts(studentPackage)
			if err != nil {
				return fmt.Errorf("cannot convert jprep student package event payload to student course entity: %v", err)
			}

			softDeleteFilter := repositories.NewSoftDeleteNotificationStudentCourseFilter()
			err = multierr.Combine(
				softDeleteFilter.StudentIDs.Set([]string{studentPackage.StudentId}),
			)
			if err != nil {
				return err
			}

			err = svc.NotificationStudentCourseRepo.SoftDelete(ctx, tx, softDeleteFilter)
			if err != nil {
				return fmt.Errorf("svc.NotificationStudentCourseRepo.SoftDelete: %v", err)
			}

			err = svc.NotificationStudentCourseRepo.BulkCreate(ctx, tx, notiStudentCourses)
			if err != nil {
				return fmt.Errorf("svc.NotificationStudentCourseRepo.BulkCreate: %v", err)
			}

		case npb.ActionKind_ACTION_KIND_DELETED:
			courseIDs := make([]string, 0)
			for _, pkg := range studentPackage.Packages {
				courseIDs = append(courseIDs, pkg.CourseIds...)
			}
			softDeleteFilter := repositories.NewSoftDeleteNotificationStudentCourseFilter()
			err := multierr.Combine(
				softDeleteFilter.StudentIDs.Set([]string{studentPackage.StudentId}),
				softDeleteFilter.CourseIDs.Set(courseIDs),
			)
			if err != nil {
				return err
			}

			err = svc.NotificationStudentCourseRepo.SoftDelete(ctx, tx, softDeleteFilter)
			if err != nil {
				return fmt.Errorf("svc.NotificationStudentCourseRepo.SoftDelete: %v", err)
			}
		}
	}

	return nil
}
