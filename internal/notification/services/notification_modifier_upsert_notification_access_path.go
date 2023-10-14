package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

func (svc *NotificationModifierService) upsertNotificationAccessPath(ctx context.Context, tx pgx.Tx, notificationID string, selectedLocationIDs []string, userID string) ([]string, error) {
	locationIDs := []string{}
	notificationAccessPathEnts := []*entities.InfoNotificationAccessPath{}
	var err error
	if len(selectedLocationIDs) > 0 {
		mapLocationAccessPath, err := svc.LocationRepo.GetLocationAccessPathsByIDs(ctx, tx, selectedLocationIDs)
		if err != nil {
			return nil, fmt.Errorf("svc.LocationRepo.GetLocationAccessPathsByIDs: %v", err)
		}

		if len(mapLocationAccessPath) != len(selectedLocationIDs) {
			return nil, fmt.Errorf("some location not found access path")
		}

		for _, selectedLocationID := range selectedLocationIDs {
			notificationAccessPathEnt := &entities.InfoNotificationAccessPath{}
			err = multierr.Combine(
				notificationAccessPathEnt.NotificationID.Set(notificationID),
				notificationAccessPathEnt.LocationID.Set(selectedLocationID),
				notificationAccessPathEnt.AccessPath.Set(mapLocationAccessPath[selectedLocationID]),
				notificationAccessPathEnt.CreatedUserID.Set(userID),
			)
			if err != nil {
				return nil, fmt.Errorf("upsertNotificationAccessPath.multierr.Combine: %v", err)
			}

			notificationAccessPathEnts = append(notificationAccessPathEnts, notificationAccessPathEnt)
		}
		locationIDs = selectedLocationIDs
	} else {
		notificationPermissions := []string{
			consts.NotificationWritePermission,
			consts.NotificationOwnerPermission,
		}
		grantedLocationIDs, mapLocationAccessPath, err := svc.LocationRepo.GetGrantedLocationsByUserIDAndPermissions(ctx, tx, userID, notificationPermissions)
		if err != nil {
			return nil, fmt.Errorf("svc.LocationRepo.GetGrantedLocationsByUserIDAndPermissions: %v", err)
		}

		for _, grantedLocationID := range grantedLocationIDs {
			notificationAccessPathEnt := &entities.InfoNotificationAccessPath{}
			err = multierr.Combine(
				notificationAccessPathEnt.NotificationID.Set(notificationID),
				notificationAccessPathEnt.LocationID.Set(grantedLocationID),
				notificationAccessPathEnt.AccessPath.Set(mapLocationAccessPath[grantedLocationID]),
				notificationAccessPathEnt.CreatedUserID.Set(userID),
			)
			if err != nil {
				return nil, fmt.Errorf("upsertNotificationAccessPath.multierr.Combine: %v", err)
			}

			notificationAccessPathEnts = append(notificationAccessPathEnts, notificationAccessPathEnt)
		}
		locationIDs = grantedLocationIDs
	}

	notificationAccessPathsDelete, err := svc.InfoNotificationAccessPathRepo.GetByNotificationIDAndNotInLocationIDs(ctx, tx, notificationID, locationIDs)
	if err != nil {
		return nil, fmt.Errorf("svc.InfoNotificationAccessPathRepo.GetByNotificationIDAndNotInLocationIDs: %v", err)
	}
	if len(notificationAccessPathsDelete) > 0 {
		deleteLocationIDs := []string{}
		for _, ent := range notificationAccessPathsDelete {
			deleteLocationIDs = append(deleteLocationIDs, ent.LocationID.String)
		}
		softDeleteNotificationAccessPathFilter := repositories.NewSoftDeleteNotificationAccessPathFilter()
		_ = softDeleteNotificationAccessPathFilter.NotificationIDs.Set([]string{notificationID})
		_ = softDeleteNotificationAccessPathFilter.LocationIDs.Set(deleteLocationIDs)
		err = svc.InfoNotificationAccessPathRepo.SoftDelete(ctx, tx, softDeleteNotificationAccessPathFilter)
		if err != nil {
			return nil, fmt.Errorf("svc.InfoNotificationAccessPathRepo.SoftDelete: %v", err)
		}
	}

	err = svc.InfoNotificationAccessPathRepo.BulkUpsert(ctx, tx, notificationAccessPathEnts)
	if err != nil {
		return nil, fmt.Errorf("svc.InfoNotificationAccessPathRepo.BulkUpsert: %v", err)
	}

	return locationIDs, nil
}
