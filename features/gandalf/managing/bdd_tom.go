package managing

import (
	"github.com/manabie-com/backend/features/tom"
)

func initStepForTomServiceFeature(s *suite) map[string]interface{} {
	steps := make(map[string]interface{})

	// TODO: add step define for tom service feature
	return steps
}

type TomStepState struct{}

func (s *suite) newTomSuite() {
	s.tomSuite = &tom.Suite{}
	s.tomSuite.DB = s.tomDB
	s.tomSuite.SetBobDBTrace(s.bobDBTrace)
	s.tomSuite.Conn = s.tomConn
	s.tomSuite.LessonChatState = &tom.LessonChatState{
		LessonConversationMap: make(map[string]string),
	}
	s.tomSuite.ZapLogger = zapLogger
	s.tomSuite.JSM = s.jsm
}
