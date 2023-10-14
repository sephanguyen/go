package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
	"github.com/manabie-com/backend/internal/spike/modules/email/infrastructure"
	"github.com/manabie-com/backend/internal/spike/modules/email/metrics"
	"github.com/manabie-com/backend/internal/spike/modules/email/util"

	"go.uber.org/multierr"
	"k8s.io/utils/strings/slices"
)

type UpsertEmailEventHandler struct {
	DB           database.Ext
	EmailMetrics metrics.EmailMetrics

	infrastructure.EmailRecipientEventRepo
}

func CheckAllowedTenantID(allowedIDs []string, orgID string) bool {
	if len(allowedIDs) == 0 {
		return false
	}
	return slices.Contains(allowedIDs, orgID)
}

func (h *UpsertEmailEventHandler) UpsertEmailEvent(ctx context.Context, payload UpsertEmailEventPayload) error {
	mapOrgIDAndEvents := make(map[string][]dto.SGEmailEvent)
	for _, event := range payload.EmailEvents {
		if event.OrganizationID != "" && CheckAllowedTenantID(payload.AllowedOrgIDs, event.OrganizationID) {
			mapOrgIDAndEvents[event.OrganizationID] = append(mapOrgIDAndEvents[event.OrganizationID], event)
		}
	}

	var err error
	for orgID, events := range mapOrgIDAndEvents {
		// Fake context to go with resource path.
		orgCtx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: orgID,
			},
		})

		err = multierr.Append(err, h.upsertEmailEventForOrg(orgCtx, events))
	}

	if err != nil {
		return err
	}

	return nil
}

func (h *UpsertEmailEventHandler) upsertEmailEventForOrg(ctx context.Context, orgEvents []dto.SGEmailEvent) error {
	eventsStr := make([]string, 0)
	emailRecipientIDs := make([]string, 0)
	for _, event := range orgEvents {
		eventsStr = append(eventsStr, string(util.GetEventFromSGEvent(event)))
		emailRecipientIDs = append(emailRecipientIDs, event.EmailRecipientID)
	}

	mapEventEntsByEventAndEmailRecipientID, err := h.EmailRecipientEventRepo.GetMapEventsByEventsAndEmailRecipientIDs(ctx, h.DB, eventsStr, emailRecipientIDs)
	if err != nil {
		return fmt.Errorf("failed EmailRecipientEventRepo.GetMapEventsByEventsAndEmailRecipientIDs: %+v", err)
	}

	mapSGEventsByEventAndEmailRecipientID := h.getMapEventsByEventAndEmailRecipientID(orgEvents)
	emailRecipientEventEntsWillUpsert := make([]*model.EmailRecipientEvent, 0)
	processedMetricRecording, droppedMetricRecording, bounceMetricRecording := 0, 0, 0
	for key, events := range mapSGEventsByEventAndEmailRecipientID {
		emailRecipientID, eventStr := util.FromEventIdentifyInfo(key)
		if emailRecipientID == "" || eventStr == "" {
			continue
		}

		var emailRecipientEventEnt *model.EmailRecipientEvent
		eventDescription := &model.EventDescription{}
		isExistEventOnDB := false
		if emailRecipientEventEnt, isExistEventOnDB = mapEventEntsByEventAndEmailRecipientID[key]; !isExistEventOnDB {
			emailRecipientEventEnt = &model.EmailRecipientEvent{}
			database.AllNullEntity(emailRecipientEventEnt)
			err := multierr.Combine(
				emailRecipientEventEnt.EmailRecipientEventID.Set(idutil.ULIDNow()),
				emailRecipientEventEnt.EmailRecipientID.Set(emailRecipientID),
				emailRecipientEventEnt.Type.Set(util.GetEventTypeFromEvent(constants.EmailEvent(eventStr))),
				emailRecipientEventEnt.Event.Set(eventStr),
				emailRecipientEventEnt.Description.Set(nil),
			)
			if err != nil {
				return fmt.Errorf("failed multierr.Combine for emailRecipientEventEnt: %+v", err)
			}

			eventDescription.Event = string(util.GetSGEventFromEvent(constants.EmailEvent(eventStr)))

			// Temporary record [processed/bounce/dropped] email event (in case this event doesn't existing on DB, if yes, already recorded)
			switch constants.EmailEvent(eventStr) {
			case constants.EmailEventProcessed:
				processedMetricRecording++
			case constants.EmailEventBounce, constants.EmailEventBlocked:
				bounceMetricRecording++
			case constants.EmailEventDropped:
				// Tested and see that for the [dropped] event, never receive the [processed] event before -> Consider it as [processed] (metric only)
				processedMetricRecording++
				droppedMetricRecording++
			}
		} else {
			err = emailRecipientEventEnt.Description.AssignTo(eventDescription)
			if err != nil {
				return fmt.Errorf("failed ent.Description.AssignTo: %+v", err)
			}
		}

		for _, event := range events {
			eventDescription.Details = append(eventDescription.Details, model.EventDescriptionDetail{
				SGEventID:            event.SGEventID,
				SGMessageID:          event.SGMessageID,
				Type:                 event.Type,
				Status:               event.Status,
				BounceClassification: event.BounceClassification,
				Reason:               event.Reason,
				Response:             event.Response,
				Attempt:              event.Attempt,
				Timestamp:            event.Timestamp,
			})
		}

		err = emailRecipientEventEnt.Description.Set(eventDescription)
		if err != nil {
			return fmt.Errorf("failed ent.Description.Set: %+v", err)
		}

		emailRecipientEventEntsWillUpsert = append(emailRecipientEventEntsWillUpsert, emailRecipientEventEnt)
	}

	err = h.EmailRecipientEventRepo.BulkInsertEmailRecipientEventRepo(ctx, h.DB, emailRecipientEventEntsWillUpsert)
	if err != nil {
		return fmt.Errorf("failed BulkInsertEmailRecipientEventRepo: %+v", err)
	}

	// Record metrics if don't have any error
	h.EmailMetrics.RecordEmailEvents(metrics.EmailProcessed, float64(processedMetricRecording))
	h.EmailMetrics.RecordEmailEvents(metrics.EmailDropped, float64(droppedMetricRecording))
	h.EmailMetrics.RecordEmailEvents(metrics.EmailBounce, float64(bounceMetricRecording))

	return nil
}

func (h *UpsertEmailEventHandler) getMapEventsByEventAndEmailRecipientID(events []dto.SGEmailEvent) map[string][]dto.SGEmailEvent {
	mapEventsByEventAndEmailRecipientID := make(map[string][]dto.SGEmailEvent)
	for _, event := range events {
		key := util.GetEventIdentifyInfo(event)
		mapEventsByEventAndEmailRecipientID[key] = append(mapEventsByEventAndEmailRecipientID[key], event)
	}

	return mapEventsByEventAndEmailRecipientID
}
