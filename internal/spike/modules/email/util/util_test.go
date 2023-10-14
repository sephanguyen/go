package util

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/spike/modules/email/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	"github.com/stretchr/testify/assert"
)

func Test_GetEventFromSGEvent(t *testing.T) {
	t.Parallel()

	t.Run("should handle delivery events", func(t *testing.T) {
		inputEvents := []dto.SGEmailEvent{
			{
				Event: "bounce",
				Type:  "bounce",
			},
			{
				Event: "deferred",
			},
			{
				Event: "bounce",
				Type:  "blocked",
			},
			{
				Event: "processed",
			},
			{
				Event: "delivered",
			},
			{
				Event: "dropped",
			},
		}

		outputEvents := []constants.EmailEvent{
			constants.EmailEventBounce,
			constants.EmailEventDeferred,
			constants.EmailEventBlocked,
			constants.EmailEventProcessed,
			constants.EmailEventDelivered,
			constants.EmailEventDropped,
		}

		s := []constants.EmailEvent{}
		for _, ev := range inputEvents {
			e := GetEventFromSGEvent(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputEvents, s)
	})

	t.Run("should handle engagement events", func(t *testing.T) {
		inputEvents := []dto.SGEmailEvent{
			{
				Event: "click",
			},
			{
				Event: "open",
			},
			{
				Event: "spamreport",
			},
			{
				Event: "unsubscribe",
			},
			{
				Event: "group_unsubscribe",
			},
			{
				Event: "group_resubscribe",
			},
		}

		outputEvents := []constants.EmailEvent{
			constants.EmailEventClick,
			constants.EmailEventOpen,
			constants.EmailEventSpamReport,
			constants.EmailEventUnsubscribe,
			constants.EmailEventGroupUnsubscribe,
			constants.EmailEventGroupResubscribe,
		}

		s := []constants.EmailEvent{}
		for _, ev := range inputEvents {
			e := GetEventFromSGEvent(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputEvents, s)
	})
}

func Test_GetEventTypeFromSGEvent(t *testing.T) {
	t.Parallel()

	t.Run("should handle delivery events", func(t *testing.T) {
		inputEvents := []constants.SGEmailEvent{
			"bounce",
			"deferred",
			"bounce",
			"processed",
			"delivered",
			"dropped",
		}

		outputEvents := []constants.EmailEventType{
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
		}

		s := []constants.EmailEventType{}
		for _, ev := range inputEvents {
			e := GetEventTypeFromSGEvent(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputEvents, s)
	})

	t.Run("should handle engagement events", func(t *testing.T) {
		inputEvents := []constants.SGEmailEvent{
			"click",
			"open",
			"spamreport",
			"unsubscribe",
			"group_unsubscribe",
			"group_resubscribe",
		}

		outputEvents := []constants.EmailEventType{
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
		}

		s := []constants.EmailEventType{}
		for _, ev := range inputEvents {
			e := GetEventTypeFromSGEvent(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputEvents, s)
	})
}

func Test_GetEventTypeFromEvent(t *testing.T) {
	t.Parallel()

	t.Run("should handle delivery events", func(t *testing.T) {
		inputEvents := []constants.EmailEvent{
			constants.EmailEventBounce,
			constants.EmailEventDeferred,
			constants.EmailEventBlocked,
			constants.EmailEventProcessed,
			constants.EmailEventDelivered,
			constants.EmailEventDropped,
		}

		outputEvents := []constants.EmailEventType{
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
			constants.EventTypeDelivery,
		}

		s := []constants.EmailEventType{}
		for _, ev := range inputEvents {
			e := GetEventTypeFromEvent(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputEvents, s)
	})

	t.Run("should handle engagement events", func(t *testing.T) {
		inputEvents := []constants.EmailEvent{
			constants.EmailEventClick,
			constants.EmailEventOpen,
			constants.EmailEventSpamReport,
			constants.EmailEventUnsubscribe,
			constants.EmailEventGroupUnsubscribe,
			constants.EmailEventGroupResubscribe,
		}

		outputEvents := []constants.EmailEventType{
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
			constants.EventTypeEngagement,
		}

		s := []constants.EmailEventType{}
		for _, ev := range inputEvents {
			e := GetEventTypeFromEvent(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputEvents, s)
	})
}

func Test_GetSGEventFromEvent(t *testing.T) {
	t.Parallel()

	t.Run("should handle delivery events", func(t *testing.T) {
		inputEvents := []constants.EmailEvent{
			constants.EmailEventBounce,
			constants.EmailEventDeferred,
			constants.EmailEventBlocked,
			constants.EmailEventProcessed,
			constants.EmailEventDelivered,
			constants.EmailEventDropped,
		}

		outputEvents := []constants.SGEmailEvent{
			"bounce",
			"deferred",
			"bounce",
			"processed",
			"delivered",
			"dropped",
		}

		s := []constants.SGEmailEvent{}
		for _, ev := range inputEvents {
			e := GetSGEventFromEvent(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputEvents, s)
	})

	t.Run("should handle engagement events", func(t *testing.T) {
		inputEvents := []constants.EmailEvent{
			constants.EmailEventClick,
			constants.EmailEventOpen,
			constants.EmailEventSpamReport,
			constants.EmailEventUnsubscribe,
			constants.EmailEventGroupUnsubscribe,
			constants.EmailEventGroupResubscribe,
		}

		outputEvents := []constants.SGEmailEvent{
			"click",
			"open",
			"spamreport",
			"unsubscribe",
			"group_unsubscribe",
			"group_resubscribe",
		}

		s := []constants.SGEmailEvent{}
		for _, ev := range inputEvents {
			e := GetSGEventFromEvent(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputEvents, s)
	})
}

func Test_GetEventIdentifyInfo(t *testing.T) {
	t.Parallel()
	emailRecipientID := idutil.ULIDNow()

	t.Run("happy case delivery events", func(t *testing.T) {
		inputEvents := []dto.SGEmailEvent{
			{
				EmailRecipientID: emailRecipientID,
				Event:            "bounce",
				Type:             "bounce",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "deferred",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "bounce",
				Type:             "blocked",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "processed",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "delivered",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "dropped",
			},
		}

		outputs := []string{
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventBounce),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventDeferred),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventBlocked),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventProcessed),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventDelivered),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventDropped),
		}

		s := []string{}
		for _, ev := range inputEvents {
			e := GetEventIdentifyInfo(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputs, s)
	})

	t.Run("happy case engagement events", func(t *testing.T) {
		inputEvents := []dto.SGEmailEvent{
			{
				EmailRecipientID: emailRecipientID,
				Event:            "click",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "open",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "spamreport",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "unsubscribe",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "group_unsubscribe",
			},
			{
				EmailRecipientID: emailRecipientID,
				Event:            "group_resubscribe",
			},
		}

		outputs := []string{
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventClick),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventOpen),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventSpamReport),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventUnsubscribe),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventGroupUnsubscribe),
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventGroupResubscribe),
		}

		s := []string{}
		for _, ev := range inputEvents {
			e := GetEventIdentifyInfo(ev)
			s = append(s, e)
		}
		assert.Equal(t, outputs, s)
	})
}

func Test_FromEventIdentifyInfo(t *testing.T) {
	t.Parallel()
	emailRecipientID := idutil.ULIDNow()
	t.Run("happy case delivery events", func(t *testing.T) {
		inputs := map[string]constants.EmailEvent{
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventBounce):    constants.EmailEventBounce,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventDeferred):  constants.EmailEventDeferred,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventBlocked):   constants.EmailEventBlocked,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventProcessed): constants.EmailEventProcessed,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventDelivered): constants.EmailEventDelivered,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventDropped):   constants.EmailEventDropped,
		}

		for key, eventExpected := range inputs {
			emailRecipientIDRet, event := FromEventIdentifyInfo(key)
			assert.Equal(t, emailRecipientID, emailRecipientIDRet)
			assert.Equal(t, event, string(eventExpected))
		}
	})

	t.Run("happy case engagement events", func(t *testing.T) {
		inputs := map[string]constants.EmailEvent{
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventClick):            constants.EmailEventClick,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventOpen):             constants.EmailEventOpen,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventSpamReport):       constants.EmailEventSpamReport,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventUnsubscribe):      constants.EmailEventUnsubscribe,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventGroupUnsubscribe): constants.EmailEventGroupUnsubscribe,
			fmt.Sprintf("%s-%s", emailRecipientID, constants.EmailEventGroupResubscribe): constants.EmailEventGroupResubscribe,
		}

		for key, eventExpected := range inputs {
			emailRecipientIDRet, event := FromEventIdentifyInfo(key)
			assert.Equal(t, emailRecipientID, emailRecipientIDRet)
			assert.Equal(t, event, string(eventExpected))
		}
	})
}
