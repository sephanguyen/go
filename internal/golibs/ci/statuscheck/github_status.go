package statuscheck

import "encoding/json"

type githubJobStatus int

const (
	statusNone githubJobStatus = iota
	statusSuccess
	statusFailure
	statusCancelled
	statusSkipped
)

func (js *githubJobStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*js = githubJobStatusFromString(s)
	return nil
}

func (js githubJobStatus) String() string {
	switch js {
	default:
		return "none"
	case statusSuccess:
		return "success"
	case statusFailure:
		return "failure"
	case statusCancelled:
		return "cancelled"
	case statusSkipped:
		return "skipped"
	}
}

func githubJobStatusFromString(s string) githubJobStatus {
	switch s {
	default:
		return statusNone
	case "success":
		return statusSuccess
	case "failure":
		return statusFailure
	case "cancelled":
		return statusCancelled
	case "skipped":
		return statusSkipped
	}
}
