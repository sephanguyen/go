package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgx/v4"
)

func (svc *NotificationModifierService) upsertNotificationFilterInfo(ctx context.Context, tx pgx.Tx, targetGroup *entities.InfoNotificationTarget, notificationID string) error {

	// delete old location filter
	err := svc.NotificationLocationFilterRepo.SoftDeleteByNotificationID(ctx, tx, notificationID)
	if err != nil {
		return fmt.Errorf("cannot delete old location filter: %v", err)
	}
	// upsert new location filter
	if targetGroup.LocationFilter.Type == consts.TargetGroupSelectTypeList.String() {
		locationFiltersEnts := make(entities.NotificationLocationFilters, 0)
		for _, locationID := range targetGroup.LocationFilter.LocationIDs {
			locationFiltersEnts = append(locationFiltersEnts, &entities.NotificationLocationFilter{
				NotificationID: database.Text(notificationID),
				LocationID:     database.Text(locationID),
			})
		}

		err = svc.NotificationLocationFilterRepo.BulkUpsert(ctx, tx, locationFiltersEnts)
		if err != nil {
			return fmt.Errorf("cannot upsert location filter: %v", err)
		}
	}

	// delete old course filter
	err = svc.NotificationCourseFilterRepo.SoftDeleteByNotificationID(ctx, tx, notificationID)
	if err != nil {
		return fmt.Errorf("cannot delete old course filter: %v", err)
	}
	// upsert new course filter
	if targetGroup.CourseFilter.Type == consts.TargetGroupSelectTypeList.String() {
		courseFiltersEnts := make(entities.NotificationCourseFilters, 0)
		for _, courseID := range targetGroup.CourseFilter.CourseIDs {
			courseFiltersEnts = append(courseFiltersEnts, &entities.NotificationCourseFilter{
				NotificationID: database.Text(notificationID),
				CourseID:       database.Text(courseID),
			})
		}

		err = svc.NotificationCourseFilterRepo.BulkUpsert(ctx, tx, courseFiltersEnts)
		if err != nil {
			return fmt.Errorf("cannot upsert course filter: %v", err)
		}
	}

	// delete old class filter
	err = svc.NotificationClassFilterRepo.SoftDeleteByNotificationID(ctx, tx, notificationID)
	if err != nil {
		return fmt.Errorf("cannot delete old class filter: %v", err)
	}
	// upsert new class filter
	if targetGroup.ClassFilter.Type == consts.TargetGroupSelectTypeList.String() {
		classFiltersEnts := make(entities.NotificationClassFilters, 0)
		for _, classID := range targetGroup.ClassFilter.ClassIDs {
			classFiltersEnts = append(classFiltersEnts, &entities.NotificationClassFilter{
				NotificationID: database.Text(notificationID),
				ClassID:        database.Text(classID),
			})
		}

		err = svc.NotificationClassFilterRepo.BulkUpsert(ctx, tx, classFiltersEnts)
		if err != nil {
			return fmt.Errorf("cannot upsert class filter: %v", err)
		}
	}

	return nil
}
