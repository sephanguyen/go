package managing

import (
	"github.com/manabie-com/backend/features/eureka"
)

type EurekaStepState struct {
	CourseID    string
	Request     interface{}
	Response    interface{}
	ResponseErr error
}

func (s *suite) newEurekaSuite(fakeFirebase string) {
	s.eurekaSuite = &eureka.Suite{}

	s.eurekaSuite.BobConn = s.bobConn
	s.eurekaSuite.BobDB = s.bobDB
	s.eurekaSuite.Conn = s.eurekaConn
	s.eurekaSuite.DB = s.eurekaDB
	s.eurekaSuite.StepState = &eureka.StepState{}
	s.eurekaSuite.ShamirConn = s.shamirConn
	s.eurekaSuite.ApplicantID = s.ApplicantID

	s.eurekaSuite.SetFirebaseAddr(fakeFirebase)
}

func initStepForEurekaServiceFeature(s *suite) map[string]interface{} {
	steps := map[string]interface{}{}

	return steps
}
