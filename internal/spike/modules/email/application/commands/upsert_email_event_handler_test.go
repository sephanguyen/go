package commands

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/model"
	"github.com/manabie-com/backend/internal/spike/modules/email/metrics"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo "github.com/manabie-com/backend/mock/spike/modules/email/infrastructure/repositories"
	mock_metrics "github.com/manabie-com/backend/mock/spike/modules/email/metrics"
	"k8s.io/utils/strings/slices"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_UpsertEmailEvent(t *testing.T) {
	t.Parallel()

	mockDB := new(mock_database.Ext)
	mockEmailMetrics := mock_metrics.EmailMetrics{}
	mockEmailRecipientEventRepo := mock_repo.MockEmailRecipientEventRepo{}
	handler := &UpsertEmailEventHandler{
		DB:                      mockDB,
		EmailMetrics:            &mockEmailMetrics,
		EmailRecipientEventRepo: &mockEmailRecipientEventRepo,
	}

	t.Run("happy case, one event", func(t *testing.T) {
		webhookEvent := []dto.SGEmailEvent{
			{
				OrganizationID:   "org-id",
				EmailID:          "email-id",
				EmailRecipientID: "email_recipient_id",
				Email:            "email",
				Event:            "processed",
				SGEventID:        "sg-event-id",
			},
		}

		orgCtx := interceptors.ContextWithJWTClaims(context.Background(), &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: "org-id",
			},
		})

		mockEmailRecipientEventRepo.On("GetMapEventsByEventsAndEmailRecipientIDs", orgCtx, mockDB, []string{"EMAIL_EVENT_PROCESSED"}, []string{"email_recipient_id"}).Once().Return(make(map[string]*model.EmailRecipientEvent), nil)
		mockEmailRecipientEventRepo.On("BulkInsertEmailRecipientEventRepo", orgCtx, mockDB, mock.Anything).Once().Return(nil)
		mockEmailMetrics.On("RecordEmailEvents", metrics.EmailProcessed, float64(1)).Once().Return(nil)
		mockEmailMetrics.On("RecordEmailEvents", metrics.EmailBounce, float64(0)).Once().Return(nil)
		mockEmailMetrics.On("RecordEmailEvents", metrics.EmailDropped, float64(0)).Once().Return(nil)

		err := handler.UpsertEmailEvent(context.Background(), UpsertEmailEventPayload{EmailEvents: webhookEvent})
		assert.Nil(t, err)
	})

	t.Run("happy case with some allowed tenants", func(t *testing.T) {
		webhookEvent := []dto.SGEmailEvent{
			{
				OrganizationID:   "org-1",
				EmailID:          "email-id",
				EmailRecipientID: "email_recipient_id",
				Email:            "email",
				Event:            "processed",
				SGEventID:        "sg-event-id",
			},
			{
				OrganizationID:   "org-2",
				EmailID:          "email-id",
				EmailRecipientID: "email_recipient_id",
				Email:            "email",
				Event:            "processed",
				SGEventID:        "sg-event-id",
			},
			{
				OrganizationID:   "org-3",
				EmailID:          "email-id",
				EmailRecipientID: "email_recipient_id",
				Email:            "email",
				Event:            "processed",
				SGEventID:        "sg-event-id",
			},
		}

		allowedOrg := []string{"org-1", "org-2"}

		mockEmailRecipientEventRepo.On("GetMapEventsByEventsAndEmailRecipientIDs", mock.MatchedBy(func(in context.Context) bool {
			rp, err := interceptors.ResourcePathFromContext(in)
			if err != nil {
				return false
			}

			if !slices.Contains(allowedOrg, rp) {
				return false
			}
			return true
		}), mockDB, []string{"EMAIL_EVENT_PROCESSED"}, []string{"email_recipient_id"}).Return(make(map[string]*model.EmailRecipientEvent), nil)
		mockEmailRecipientEventRepo.On("BulkInsertEmailRecipientEventRepo", mock.MatchedBy(func(in context.Context) bool {
			rp, err := interceptors.ResourcePathFromContext(in)
			if err != nil {
				return false
			}

			if !slices.Contains(allowedOrg, rp) {
				return false
			}
			return true
		}), mockDB, mock.Anything).Return(nil)
		mockEmailMetrics.On("RecordEmailEvents", metrics.EmailProcessed, float64(1)).Return(nil)
		mockEmailMetrics.On("RecordEmailEvents", metrics.EmailBounce, float64(0)).Return(nil)
		mockEmailMetrics.On("RecordEmailEvents", metrics.EmailDropped, float64(0)).Return(nil)

		err := handler.UpsertEmailEvent(context.Background(), UpsertEmailEventPayload{
			EmailEvents:   webhookEvent,
			AllowedOrgIDs: allowedOrg,
		})
		assert.Nil(t, err)
	})

	t.Run("happy case with NONE allowed tenants", func(t *testing.T) {
		webhookEvent := []dto.SGEmailEvent{
			{
				OrganizationID:   "org-1",
				EmailID:          "email-id",
				EmailRecipientID: "email_recipient_id",
				Email:            "email",
				Event:            "processed",
				SGEventID:        "sg-event-id",
			},
			{
				OrganizationID:   "org-2",
				EmailID:          "email-id",
				EmailRecipientID: "email_recipient_id",
				Email:            "email",
				Event:            "processed",
				SGEventID:        "sg-event-id",
			},
			{
				OrganizationID:   "org-3",
				EmailID:          "email-id",
				EmailRecipientID: "email_recipient_id",
				Email:            "email",
				Event:            "processed",
				SGEventID:        "sg-event-id",
			},
		}

		// should none process any event
		allowedOrg := []string{}

		err := handler.UpsertEmailEvent(context.Background(), UpsertEmailEventPayload{
			EmailEvents:   webhookEvent,
			AllowedOrgIDs: allowedOrg,
		})
		assert.Nil(t, err)
	})
}
