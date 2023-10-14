package constants

type EmailEvent string
type EmailEventType string

type SGEmailEvent string

const (
	EventTypeNone       EmailEventType = "EMAIL_EVENT_TYPE_NONE"
	EventTypeDelivery   EmailEventType = "EMAIL_EVENT_TYPE_DELIVERY"
	EventTypeEngagement EmailEventType = "EMAIL_EVENT_TYPE_ENGAGEMENT"

	EmailEventNone             EmailEvent = "EMAIL_EVENT_NONE"
	EmailEventProcessed        EmailEvent = "EMAIL_EVENT_PROCESSED"
	EmailEventDropped          EmailEvent = "EMAIL_EVENT_DROPPED"
	EmailEventDelivered        EmailEvent = "EMAIL_EVENT_DELIVERED"
	EmailEventDeferred         EmailEvent = "EMAIL_EVENT_DEFERRED"
	EmailEventBounce           EmailEvent = "EMAIL_EVENT_BOUNCE"  // actually this is an Event Bounce Type, not Event
	EmailEventBlocked          EmailEvent = "EMAIL_EVENT_BLOCKED" // actually this is an Event Bounce Type, not Event
	EmailEventOpen             EmailEvent = "EMAIL_EVENT_OPEN"
	EmailEventClick            EmailEvent = "EMAIL_EVENT_CLICK"
	EmailEventSpamReport       EmailEvent = "EMAIL_EVENT_SPAM_REPORT"
	EmailEventUnsubscribe      EmailEvent = "EMAIL_EVENT_UNSUBSCRIBE"
	EmailEventGroupUnsubscribe EmailEvent = "EMAIL_EVENT_GROUP_UNSUBSCRIBE"
	EmailEventGroupResubscribe EmailEvent = "EMAIL_EVENT_GROUP_RESUBSCRIBE"
)

// map SendGrid event type -> Manabie event type

var EventEventTypeBySGEvent = map[SGEmailEvent]EmailEventType{
	// delivery
	"processed": EventTypeDelivery,
	"dropped":   EventTypeDelivery,
	"delivered": EventTypeDelivery,
	"deferred":  EventTypeDelivery,
	"bounce":    EventTypeDelivery,

	// engagement
	"open":              EventTypeEngagement,
	"click":             EventTypeEngagement,
	"spamreport":        EventTypeEngagement,
	"unsubscribe":       EventTypeEngagement,
	"group_unsubscribe": EventTypeEngagement,
	"group_resubscribe": EventTypeEngagement,
}

// map SendGrid event -> Manabie event

var EmailEventBySGEvent = map[SGEmailEvent]EmailEvent{
	// delivery
	"processed": EmailEventProcessed,
	"dropped":   EmailEventDropped,
	"delivered": EmailEventDelivered,
	"deferred":  EmailEventDeferred,
	"bounce":    EmailEventBounce,
	"blocked":   EmailEventBlocked,

	// engagement
	"open":              EmailEventOpen,
	"click":             EmailEventClick,
	"spamreport":        EmailEventSpamReport,
	"unsubscribe":       EmailEventUnsubscribe,
	"group_unsubscribe": EmailEventGroupUnsubscribe,
	"group_resubscribe": EmailEventGroupResubscribe,
}

// map Manabie event -> SendGrid event

var SGEventByEmailEvent = map[EmailEvent]SGEmailEvent{
	// delivery
	EmailEventProcessed: "processed",
	EmailEventDropped:   "dropped",
	EmailEventDelivered: "delivered",
	EmailEventDeferred:  "deferred",
	EmailEventBounce:    "bounce",
	EmailEventBlocked:   "bounce",

	// engagement
	EmailEventOpen:             "open",
	EmailEventClick:            "click",
	EmailEventSpamReport:       "spamreport",
	EmailEventUnsubscribe:      "unsubscribe",
	EmailEventGroupUnsubscribe: "group_unsubscribe",
	EmailEventGroupResubscribe: "group_resubscribe",
}

// map Manabie event -> Manabie event type

var EmailEventTypeByEvent = map[EmailEvent]EmailEventType{
	// delivery
	EmailEventProcessed: EventTypeDelivery,
	EmailEventDropped:   EventTypeDelivery,
	EmailEventDelivered: EventTypeDelivery,
	EmailEventDeferred:  EventTypeDelivery,
	EmailEventBounce:    EventTypeDelivery,
	EmailEventBlocked:   EventTypeDelivery,

	// engagement
	EmailEventOpen:             EventTypeEngagement,
	EmailEventClick:            EventTypeEngagement,
	EmailEventSpamReport:       EventTypeEngagement,
	EmailEventUnsubscribe:      EventTypeEngagement,
	EmailEventGroupUnsubscribe: EventTypeEngagement,
	EmailEventGroupResubscribe: EventTypeEngagement,
}
