package metrics

type EmailEventMetricType string

const (
	EmailQueued    EmailEventMetricType = "queued"
	EmailProcessed EmailEventMetricType = "processed"
	EmailBounce    EmailEventMetricType = "bounce"
	EmailDropped   EmailEventMetricType = "dropped"
)
