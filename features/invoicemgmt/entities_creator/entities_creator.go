package entitiescreator

import (
	"time"

	"github.com/manabie-com/backend/features/common"
)

type (
	InsertEntityFunction func(stepState *common.StepState) error
)

const (
	MaxRetry = 1000
)

type EntitiesCreator struct {
}

func NewEntitiesCreator() *EntitiesCreator {
	return &EntitiesCreator{}
}

// WaitForKafkaSync should be used to add delay to wait for kafka sync
func (c *EntitiesCreator) WaitForKafkaSync(d time.Duration) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		time.Sleep(d)
		return nil
	}
}
