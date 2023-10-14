package lessonrecording

type StartCall struct {
	SubscribeVideoUids    []string
	SubscribeAudioUids    []string
	FileNamePrefix        []string
	TranscodingConfigJSON string
}

type StopCall struct {
	UID     int    `json:"uid"`
	Channel string `json:"channel"`
	Rid     string `json:"rid"`
	Sid     string `json:"sid"`
}

type UserCredentials struct {
	Rtc string `json:"rtc"`
	UID int    `json:"uid"`
}

type CallStatus struct {
	Rid string `json:"rid"`
	Sid string `json:"sid"`
}

type Status struct {
	ResourceID     string         `json:"resourceId"`
	Sid            string         `json:"sid"`
	ServerResponse ServerResponse `json:"serverResponse"`
	Code           int            `json:"code"`
}

type ServerResponse struct {
	FileListMode    string     `json:"fileListMode"`
	FileList        []FileInfo `json:"fileList"`
	Status          int        `json:"status"`
	SliceStartTime  int64      `json:"sliceStartTime"`
	UploadingStatus string     `json:"uploadingStatus"`
}

type FileInfo struct {
	Filename       string `json:"filename"`
	TrackType      string `json:"trackType"`
	UID            string `json:"uid"`
	MixedAllUser   bool   `json:"mixedAllUser"`
	IsPlayAble     bool   `json:"isPlayable"`
	SliceStartTime int64  `json:"sliceStartTime"`
}
