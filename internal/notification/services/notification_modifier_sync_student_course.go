package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

func (svc *NotificationModifierService) SyncStudentPackageV2(ctx context.Context, data *npb.EventStudentPackageV2) error {
	err := database.ExecInTx(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := multierr.Combine(
			svc.upsertStudentCourse(ctx, data, tx),
			svc.upsertClassMember(ctx, data, tx),
		)
		return err
	})

	if err != nil {
		return err
	}
	return nil
}

func (svc *NotificationModifierService) upsertStudentCourse(ctx context.Context, data *npb.EventStudentPackageV2, tx pgx.Tx) error {
	filter := repositories.NewFindNotificationStudentCourseFilter()
	err := multierr.Combine(
		filter.StudentID.Set(data.StudentPackage.StudentId),
		filter.CourseID.Set(data.StudentPackage.Package.CourseId),
	)
	if err != nil {
		return err
	}

	existStudentCourses, err := svc.NotificationStudentCourseRepo.Find(ctx, tx, filter)
	if err != nil {
		return fmt.Errorf("svc.NotificationStudentCourseRepo.Find: %v", err)
	}

	studentCourseEnt, err := mappers.EventStudentPackageV2PbToNotificationStudentCourseEnt(data)
	if err != nil {
		return fmt.Errorf("cannot convert student package event payload to student course entity: %v", err)
	}

	if len(existStudentCourses) > 0 {
		studentCourseEnt.StudentCourseID = existStudentCourses[0].StudentCourseID

		// for other partners (this flow exclude JPREP) delete the duplicated data of student course
		if len(existStudentCourses) > 1 {
			softDeleteFilter := repositories.NewSoftDeleteNotificationStudentCourseFilter()
			err := multierr.Combine(
				softDeleteFilter.StudentIDs.Set([]string{data.StudentPackage.Package.CourseId}),
				softDeleteFilter.CourseIDs.Set([]string{data.StudentPackage.StudentId}),
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

	if data.StudentPackage.IsActive {
		err = multierr.Append(err, studentCourseEnt.DeletedAt.Set(nil))
	} else {
		now := time.Now()
		err = multierr.Append(err, studentCourseEnt.DeletedAt.Set(now))
	}
	if err != nil {
		return fmt.Errorf("multierr when assign to student course entity: %v", err)
	}

	err = svc.NotificationStudentCourseRepo.Upsert(ctx, tx, studentCourseEnt)
	if err != nil {
		return fmt.Errorf("error when upsert notification student course: %v", err)
	}

	return nil
}

func (svc *NotificationModifierService) upsertClassMember(ctx context.Context, data *npb.EventStudentPackageV2, tx pgx.Tx) error {
	// soft delete when student no longer belong to class or student package is un-active
	if data.StudentPackage.Package.ClassId == "" || !data.StudentPackage.IsActive {
		softDeletedFilter := repositories.NewNotificationClassMemberFilter()
		err := multierr.Combine(
			softDeletedFilter.StudentIDs.Set([]string{data.StudentPackage.StudentId}),
			softDeletedFilter.CourseIDs.Set([]string{data.StudentPackage.Package.CourseId}),
		)
		if err != nil {
			return fmt.Errorf("failed Combine: %v", err)
		}
		err = svc.NotificationClassMemberRepo.SoftDeleteByFilter(ctx, tx, softDeletedFilter)
		if err != nil {
			return err
		}
		return nil
	}

	// upsert when existing class_id
	if data.StudentPackage.Package.ClassId != "" {
		// delete old class member, in case change class
		softDeletedFilter := repositories.NewNotificationClassMemberFilter()
		err := multierr.Combine(
			softDeletedFilter.StudentIDs.Set([]string{data.StudentPackage.StudentId}),
			softDeletedFilter.CourseIDs.Set([]string{data.StudentPackage.Package.CourseId}),
		)
		if err != nil {
			return fmt.Errorf("failed Combine: %v", err)
		}
		err = svc.NotificationClassMemberRepo.SoftDeleteByFilter(ctx, tx, softDeletedFilter)
		if err != nil {
			return err
		}

		// upsert by student_id, course_id, location_id
		notiClassMember, err := mappers.EventStudentPackageV2PbToNotificationClassMemberEnt(data)
		if err != nil {
			return fmt.Errorf("cannot convert student package event payload to class member entity: %v", err)
		}
		if err := svc.NotificationClassMemberRepo.Upsert(ctx, tx, notiClassMember); err != nil {
			return fmt.Errorf("ss.NotificationClassMemberRepo.Upsert: %v", err)
		}
	}
	return nil
}
