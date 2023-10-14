package domain

type Details interface{}

type EventType int

const (
	CloudRecordingServiceError        EventType = 1
	CloudRecordingServiceWarning      EventType = 2
	CloudRecordingServiceStatusUpdate EventType = 3
	CloudRecordingServiceFileInfo     EventType = 4
	CloudRecordingServiceExited       EventType = 11
	CloudRecordingServiceFailover     EventType = 12

	UploadingStarts     EventType = 30
	UploadingDone       EventType = 31
	UploadingBackupDone EventType = 32
	UploadingProgress   EventType = 33

	RecordingStarts                  EventType = 40
	RecordingExits                   EventType = 41
	RecordingSliceStart              EventType = 42
	RecordingAudioStreamStateChanged EventType = 43
	RecordingVideoStreamStateChanged EventType = 44
	RecordingSnapshotFile            EventType = 45

	VODStarted   EventType = 60
	VODTriggered EventType = 61

	WebRecorderStarted         EventType = 70
	WebRecorderStopped         EventType = 71
	WebRecorderCapabilityLimit EventType = 72
	WebRecorderReload          EventType = 73

	TranscoderStarted   EventType = 80
	TranscoderCompleted EventType = 81

	DownloadFailed               EventType = 90
	RTMPPublishStatus            EventType = 100
	PostponeTranscodeFinalResult EventType = 1001
)

type AgoraCallbackPayload struct {
	NoticeID  string                `json:"noticeId"`
	ProductID int                   `json:"productId"`
	EventType EventType             `json:"eventType"`
	NotifyMs  int                   `json:"notifyMs"`
	Payload   CloudRecordingPayload `json:"payload"`
}

type CloudRecordingPayload struct {
	ChannelName string  `json:"cname"`
	UID         string  `json:"uid"`
	SID         string  `json:"sid"`
	Sequence    uint32  `json:"sequence"`
	SendTS      int64   `json:"sendts"`
	ServiceType uint8   `json:"serviceType"`
	Details     Details `json:"details"`
}
type SessionExitDetail struct {
	MsgName    string `json:"msgName"`
	ExitStatus uint8  `json:"exitStatus"`
}
type RecorderLeaveDetail struct {
	MsgName   string    `json:"msgName"`
	LeaveCode LeaveCode `json:"leaveCode"`
}

type CloudRecordingErrorDetail struct {
	MsgName    string `json:"msgName"`
	Module     uint8  `json:"module"`
	ErrorLevel uint8  `json:"errorLevel"`
	ErrorCode  uint8  `json:"errorCode"`
	Stat       uint8  `json:"stat"`
	ErrorMsg   string `json:"errorMsg"`
}

type LeaveCode uint8

const (
	LeaveCodeInit       LeaveCode = 0
	LeaveCodeSig        LeaveCode = 2
	LeaveCodeNoUsers    LeaveCode = 4
	LeaveCodeTimerCatch LeaveCode = 8

	ExitStatusNormal uint8 = 0
)
