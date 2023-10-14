package jprep

import (
	"context"
	"time"

	"github.com/manabie-com/backend/features/yasuo"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"

	"go.uber.org/multierr"
)

func initStepForYasuoServiceFeature(s *suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^jprep a valid admin$`:         s.yasuoSuite.ASignedInAdmin,
		`^jprep a valid academic year$`: s.aValidAcademicYear,
	}

	return steps
}

type YasuoStepState struct {
	Random                string
	CurrentCourseID       int
	CurrentTeacherID      string
	CurrentUserGroup      string
	CurrentLessonIDs      []string
	CurrentLessonNames    []string
	CurrentConversationID []string
	CurrentAcademicYearID int
	ConversationID        string
	CurrentClassID        int
	CurrentCourseIDs      []int
	CurrentLessonID       int
	CurrentLessonGroup    string
	CurrentTopicID        string

	CurrentPresetStudyPlanID string
}

func (s *suite) newYasuoSuite(fakeFirebaseAddr string) {
	s.yasuoSuite = &yasuo.Suite{}
	s.yasuoSuite.Conn = s.yasuoConn
	s.yasuoSuite.BobConn = s.bobConn
	s.yasuoSuite.DBTrace = s.bobDBTrace
	s.yasuoSuite.ZapLogger = s.ZapLogger
	s.yasuoSuite.JSM = s.jsm
	s.yasuoSuite.ApplicantID = s.ApplicantID
	s.yasuoSuite.ShamirConn = s.shamirConn

	yasuo.SetFirebaseAddr(fakeFirebaseAddr)
}

func (s *suite) aValidAcademicYear() error {
	e := &entities.AcademicYear{}
	s.CurrentAcademicYearID = 2021

	err := multierr.Combine(
		e.ID.Set(toJprepAcedemicYearID(s.CurrentAcademicYearID)),
		e.SchoolID.Set(constants.JPREPSchool),
		e.Name.Set(toJprepAcedemicYearID(s.CurrentAcademicYearID)),
		e.StartYearDate.Set(time.Now()),
		e.EndYearDate.Set(time.Now().Add(200*24*time.Hour)),
		e.Status.Set(entities.AcademicYearStatusActive),
	)
	if err != nil {
		return err
	}

	aRepo := &repositories.AcademicYearRepo{}
	err = aRepo.Create(context.Background(), s.bobDBTrace, e)
	if err != nil {
		return err
	}

	return nil
}
