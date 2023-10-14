package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultCountMetricValue = 0
)

type NotificationMetrics interface {
	RecordNotificationCreated(infoNotificationCount float64)
	RecordUserNotificationCreated(userNotificationCount float64)
	RecordPushNotificationErrors(status PushedNotificationStatus, numErr float64)
	GetCollectors() []prometheus.Collector

	InitCounterValue()

	RecordSystemNotificationCreated(count float64)
	RecordSystemNotificationError(count float64)
}

// nolint
func NewClientMetrics(appname string) *clientMetrics {
	c := &clientMetrics{
		notificationCreatedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "manabie_notification_created_counter",
			Help: "Whether notification is created successfully",
			ConstLabels: map[string]string{
				"app": appname, // this label is useful to route which alert goes to which channel (slack/opsgenie)
			},
		}),
		userNotificationCreatedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "manabie_user_notification_created_counter",
			Help: "Whether user notification is created successfully",
			ConstLabels: map[string]string{
				"app": appname, // this label is useful to route which alert goes to which channel (slack/opsgenie)
			},
		}),
		notificationPusherErrorCounter: *prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "manabie_notification_error_pushed_counter",
			Help: "Whether the notification has encounter errors when push notification",
			ConstLabels: map[string]string{
				"app": appname,
			},
		}, []string{"status"}),
		systemNotificationCreatedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "manabie_system_notification_created_counter",
			Help: "Whether system notification is created successfully",
			ConstLabels: map[string]string{
				"app": appname, // this label is useful to route which alert goes to which channel (slack/opsgenie)
			},
		}),
		systemNotificationErrorCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "manabie_system_notification_error_counter",
			Help: "Whether system notification encounterd error",
			ConstLabels: map[string]string{
				"app": appname, // this label is useful to route which alert goes to which channel (slack/opsgenie)
			},
		}),
	}

	return c
}

type clientMetrics struct {
	notificationCreatedCounter     prometheus.Counter
	userNotificationCreatedCounter prometheus.Counter

	notificationPusherErrorCounter prometheus.CounterVec

	systemNotificationCreatedCounter prometheus.Counter
	systemNotificationErrorCounter   prometheus.Counter
}

func (c *clientMetrics) RecordNotificationCreated(infoNotificationCount float64) {
	c.notificationCreatedCounter.Add(infoNotificationCount)
}

func (c *clientMetrics) RecordUserNotificationCreated(userNotificationCount float64) {
	c.userNotificationCreatedCounter.Add(userNotificationCount)
}

func (c *clientMetrics) RecordPushNotificationErrors(status PushedNotificationStatus, numErr float64) {
	c.notificationPusherErrorCounter.With(
		prometheus.Labels{"status": string(status)},
	).Add(numErr)
}

func (c *clientMetrics) RecordSystemNotificationCreated(count float64) {
	c.systemNotificationCreatedCounter.Add(count)
}

func (c *clientMetrics) RecordSystemNotificationError(count float64) {
	c.systemNotificationErrorCounter.Add(count)
}

func (c *clientMetrics) GetCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		c.notificationCreatedCounter,
		c.userNotificationCreatedCounter,
		c.notificationPusherErrorCounter,
		c.systemNotificationCreatedCounter,
		c.systemNotificationErrorCounter,
	}
}

func (c *clientMetrics) InitCounterValue() {
	c.RecordPushNotificationErrors(StatusOK, defaultCountMetricValue)
	c.RecordPushNotificationErrors(StatusFail, defaultCountMetricValue)
	c.RecordNotificationCreated(defaultCountMetricValue)
	c.RecordUserNotificationCreated(defaultCountMetricValue)
	c.RecordSystemNotificationCreated(defaultCountMetricValue)
	c.RecordSystemNotificationError(defaultCountMetricValue)
}
