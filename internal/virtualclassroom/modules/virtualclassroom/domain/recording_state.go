package domain

import "fmt"

type RecordingState struct {
	IsRecording bool
	Creator     *string
}

func (r *RecordingState) isValid() error {
	if r.IsRecording {
		if r.Creator == nil {
			return fmt.Errorf("creator cannot be empty when is recording video")
		}
	} else if r.Creator != nil {
		return fmt.Errorf("creator cannot existed when is not recording video")
	}

	return nil
}

type CompositeRecordingState struct {
	ResourceID  string `json:"resourceId"`
	SID         string `json:"s_id"`
	UID         int    `json:"u_id"`
	IsRecording bool   `json:"is_recording"`
	Creator     string `json:"creator"`
}
