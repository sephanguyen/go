package book

import (
	"github.com/manabie-com/backend/features/repository/syllabus/utils"
)

type Suite utils.Suite[StepState]

// type Suite struct {
// 	*StepState
// 	DB             database.Ext
// 	ZapLogger      *zap.Logger
// 	HasuraAdminURL string
// 	HasuraPassword string
// }

// func NewSuite(stepState *StepState, db database.Ext, zapLogger *zap.Logger, hasuraURL, hasuraPwd string) *Suite {
// 	return &Suite{
// 		StepState:      stepState,
// 		DB:             db,
// 		ZapLogger:      zapLogger,
// 		HasuraAdminURL: hasuraURL,
// 		HasuraPassword: hasuraPwd,
// 	}
// }
