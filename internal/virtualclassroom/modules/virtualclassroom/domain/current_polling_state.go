package domain

import (
	"fmt"
	"time"
)

type CurrentPollingStatus string

const (
	CurrentPollingStatusStarted CurrentPollingStatus = "POLLING_STATE_STARTED"
	CurrentPollingStatusStopped CurrentPollingStatus = "POLLING_STATE_STOPPED"
	CurrentPollingStatusEnded   CurrentPollingStatus = "POLLING_STATE_ENDED"
)

type CurrentPolling struct {
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	StoppedAt *time.Time            `json:"stopped_at,omitempty"`
	EndedAt   *time.Time            `json:"end_at,omitempty"`
	Options   CurrentPollingOptions `json:"options"`
	Status    CurrentPollingStatus  `json:"status"`
	IsShared  bool                  `json:"is_shared"`
	Question  string                `json:"question"`
}

func (c *CurrentPolling) isValid() error {
	if len(c.Options) == 0 {
		return fmt.Errorf("options cannot be empty")
	}

	if c.StoppedAt != nil && c.StoppedAt.Before(c.CreatedAt) {
		return fmt.Errorf("stopped at cannot before created at")
	}
	if c.EndedAt != nil && c.EndedAt.Before(c.CreatedAt) {
		return fmt.Errorf("ended at cannot before created at")
	}

	haveRightAnswer := false
	answerMap := make(map[string]bool)
	for _, o := range c.Options {
		if v, ok := answerMap[o.Answer]; ok {
			if v != o.IsCorrect {
				return fmt.Errorf("answer %s cannot have 2 different value", o.Answer)
			}
		} else {
			answerMap[o.Answer] = o.IsCorrect
		}
		if o.IsCorrect {
			haveRightAnswer = true
		}
	}
	if !haveRightAnswer {
		return fmt.Errorf("there are no right answer for this polling")
	}

	if len(c.Status) == 0 {
		return fmt.Errorf("status cannot be empty")
	}
	switch c.Status {
	case CurrentPollingStatusStopped:
		if c.StoppedAt == nil {
			return fmt.Errorf("stopped at cannot be empty")
		}
	case CurrentPollingStatusEnded:
		if c.EndedAt == nil {
			return fmt.Errorf("ended at cannot be empty")
		}
	}

	return nil
}

func (c *CurrentPolling) GetAnswerMap() map[string]bool {
	res := make(map[string]bool)
	for _, answer := range c.Options {
		res[answer.Answer] = true
	}

	return res
}

type CurrentPollingOption struct {
	Answer    string `json:"answer"`
	IsCorrect bool   `json:"is_correct"`
	Content   string `json:"content"`
}

type CurrentPollingOptions []*CurrentPollingOption

func (pos CurrentPollingOptions) ValidatePollingOptions(answers []string) error {
	if len(pos) < 2 {
		return fmt.Errorf("option must be larger than 1")
	}
	if len(pos) > 10 {
		return fmt.Errorf("option can not be larger than 10")
	}
	allOptions := make(map[string]*CurrentPollingOption)
	for _, o := range pos {
		allOptions[o.Answer] = o
	}
	for _, answer := range answers {
		if _, ok := allOptions[answer]; !ok {
			return fmt.Errorf("the answer %s doesn't belong to options", answer)
		}
	}

	return nil
}
