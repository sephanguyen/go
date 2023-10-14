package producers

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LocationProducer struct {
	JSM nats.JetStreamManagement
}

func (l *LocationProducer) PublishLocationTypeEvent(ctx context.Context, msg []*domain.LocationType) error {
	locationTypes := make([]*npb.EventSyncLocationType_LocationType, 0, len(msg))
	for i := 0; i < len(msg); i++ {
		locationType := &npb.EventSyncLocationType_LocationType{
			LocationTypeId:       msg[i].LocationTypeID,
			Name:                 msg[i].Name,
			DisplayName:          msg[i].DisplayName,
			ParentName:           msg[i].ParentName,
			ParentLocationTypeId: msg[i].ParentLocationTypeID,
			IsArchived:           msg[i].IsArchived,
			CreatedAt:            timestamppb.New(msg[i].CreatedAt),
			UpdatedAt:            timestamppb.New(msg[i].UpdatedAt),
		}
		if msg[i].DeletedAt != nil {
			locationType.DeletedAt = timestamppb.New(*msg[i].DeletedAt)
		}
		locationTypes = append(locationTypes, locationType)
	}
	err := nats.ChunkHandler(len(locationTypes), 100, func(start, end int) error {
		msg, err := proto.Marshal(&npb.EventSyncLocationType{
			LocationTypes: locationTypes[start:end],
		})
		if err != nil {
			return fmt.Errorf("unable to marshal data: %w", err)
		}
		_, err = l.JSM.PublishAsyncContext(ctx, constants.SubjectSyncLocationTypeUpserted, msg)
		return err
	})
	if err != nil {
		return fmt.Errorf("PublishLocationTypeEvent err: %w", err)
	}

	return nil
}

func (l *LocationProducer) PublishLocationEvent(ctx context.Context, msg []*domain.Location) error {
	locations := make([]*npb.EventSyncLocation_Location, 0, len(msg))
	for i := 0; i < len(msg); i++ {
		location := &npb.EventSyncLocation_Location{
			LocationId:              msg[i].LocationID,
			Name:                    msg[i].Name,
			LocationType:            msg[i].LocationType,
			ParentLocationId:        msg[i].ParentLocationID,
			PartnerInternalId:       msg[i].PartnerInternalID,
			PartnerInternalParentId: msg[i].PartnerInternalParentID,
			IsArchived:              msg[i].IsArchived,
			AccessPath:              msg[i].AccessPath,
			CreatedAt:               timestamppb.New(msg[i].CreatedAt),
			UpdatedAt:               timestamppb.New(msg[i].UpdatedAt),
		}
		if msg[i].DeletedAt != nil {
			location.DeletedAt = timestamppb.New(*msg[i].DeletedAt)
		}
		locations = append(locations, location)
	}
	err := nats.ChunkHandler(len(locations), 100, func(start, end int) error {
		msg, err := proto.Marshal(&npb.EventSyncLocation{
			Locations: locations[start:end],
		})
		if err != nil {
			return fmt.Errorf("unable to marshal data: %w", err)
		}
		_, err = l.JSM.PublishAsyncContext(ctx, constants.SubjectSyncLocationUpserted, msg)
		return err
	})
	if err != nil {
		return fmt.Errorf("PublishLocationEvent err: %w", err)
	}

	return nil
}
