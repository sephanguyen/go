package domain

type UnpublishStatus string

const (
	UnpublishStatsNone              UnpublishStatus = "UNPUBLISH_STATUS_UNPUBLISHED_NONE"
	UnpublishStatsUnpublishedBefore UnpublishStatus = "UNPUBLISH_STATUS_UNPUBLISHED_BEFORE"
)
