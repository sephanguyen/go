package payloads

type RetrieveRecordedVideosByLessonIDPayload struct {
	LessonID        string
	Limit           uint32
	RecordedVideoID string
}

type GetRecordingByIDPayload struct {
	RecordedVideoID string
}
