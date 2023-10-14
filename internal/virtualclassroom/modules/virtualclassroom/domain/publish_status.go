package domain

type PrepareToPublishStatus string

const (
	PublishStatusNone           PrepareToPublishStatus = "PREPARE_TO_PUBLISH_STATUS_NONE"
	PublishStatusMaxLimit       PrepareToPublishStatus = "PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT"
	PublishStatusPreparedBefore PrepareToPublishStatus = "PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE"
)
