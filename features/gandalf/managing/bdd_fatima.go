package managing

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/fatima"
)

type fatimaStepState struct{ UserId string }

func initStepForFatimaServiceFeature(s *suite) map[string]interface{} {
	steps := map[string]interface{}{`^user add a "([^"]*)" package for a student$`: s.fatimaSuite.UserAddAPackageForAStudentV2, `^some package data in db$`: s.fatimaSuite.SomePackageDataInDB, `^returns "([^"]*)" status code$`: s.fatimaSuite.ReturnsStatusCode, `^eureka must "([^"]*)" course student$`: s.eurekaMustCheckCourseStudent, `^eureka must "([^"]*)" student study plan$`: s.eurekaMustCheckStudentStudyPlan, `^eureka must "([^"]*)" study plan$`: s.eurekaMustCheckStudyPlan, `^user add a package by courses for a student$`: s.fatimaSuite.UserAddACourseForAStudent, `^server must store these courses for this student$`: s.fatimaSuite.ServerMustStoreTheseCoursesForThisStudent, `^server must store this package for this student$`: s.fatimaSuite.ServerMustStoreThisPackageForThisStudent, `^a valid "([^"]*)" token for eureka service$`: s.eurekaSuite.AValidToken, `^valid assignment in eureka db$`: s.eurekaSuite.ValidAssignmentInDB}
	return steps
}

func (s *suite) newFatimaSuite(fakeFirebase string) {
	s.fatimaSuite = &fatima.Suite{}
	s.fatimaSuite.BobConn = s.bobConn
	s.fatimaSuite.Conn = s.fatimaConn
	s.fatimaSuite.DB = s.fatimaDB
	s.fatimaSuite.StepState = &fatima.StepState{}
	s.fatimaSuite.EurekaDB = s.eurekaDB
	s.fatimaSuite.SetFirebaseAddr(fakeFirebase)
}

func (s *suite) eurekaMustCheckCourseStudent(ctx context.Context, arg string) (context.Context, error) {
	mainProcess := func() error {
		query := "select deleted_at " +
			"from course_students " +
			"where student_id = $1 " +
			"and course_id = ANY($2)"

		rows, err := s.eurekaDB.Query(ctx, query, s.fatimaSuite.StudentID, s.fatimaSuite.CourseIDs)
		if err != nil {
			return err

		}
		defer rows.Close()

		var deletedAt interface{}
		count := 0
		for rows.Next() {
			err = rows.Scan(&deletedAt)
			if err != nil {
				return err

			}
			if arg == "delete" {
				if deletedAt == nil {
					return fmt.Errorf("cannot delete course_students")
				}
			} else if arg == "store" {
				if deletedAt != nil {
					return fmt.Errorf("cannot store course_students")
				}
			}
			count++
		}

		if count == 0 {
			if arg == "not exist" {
				return nil
			}
			return fmt.Errorf("cannot find any course_students")
		}

		return nil
	}

	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}

func (s *suite) eurekaMustCheckStudentStudyPlan(ctx context.Context, arg string) (context.Context, error) {
	mainProcess := func() error {
		query := "select deleted_at " +
			"from student_study_plans " +
			"where student_id = $1"

		rows, err := s.fatimaSuite.EurekaDB.Query(ctx, query, s.fatimaSuite.StudentID)
		if err != nil {
			return err

		}
		defer rows.Close()

		var deletedAt interface{}
		count := 0
		for rows.Next() {
			err = rows.Scan(&deletedAt)
			if err != nil {
				return err

			}
			if arg == "delete" {
				if deletedAt == nil {
					return fmt.Errorf("cannot delete student_study_plans")
				}
			} else if arg == "store" {
				if deletedAt != nil {
					return fmt.Errorf("cannot store student_study_plans")
				}
			}

			count++
		}

		if count == 0 {
			if arg == "not exist" {
				return nil
			}
			return fmt.Errorf("cannot find any student_study_plans")
		}

		return nil
	}

	return ctx, s.ExecuteWithRetry(mainProcess, 3*time.Second, 30)
}

func (s *suite) eurekaMustCheckStudyPlan(ctx context.Context, arg string) (context.Context, error) {
	mainProcess := func() error {
		query := "select deleted_at " +
			"from study_plans " +
			"where master_study_plan_id = ANY($1)"
		rows, err := s.eurekaDB.Query(ctx, query, s.fatimaSuite.MasterStudyPlanIDs)
		if err != nil {
			return err

		}
		defer rows.Close()

		var deletedAt interface{}
		count := 0
		for rows.Next() {
			err = rows.Scan(&deletedAt)
			if err != nil {
				return err

			}

			if arg == "delete" {
				if deletedAt == nil {
					return fmt.Errorf("cannot delete study_plans")
				}
			} else if arg == "store" {
				if deletedAt != nil {
					return fmt.Errorf("cannot store study_plans")
				}
			}

			count++
		}
		if count == 0 {
			if arg == "not exist" {
				return nil
			}
			return fmt.Errorf("cannot find any study_plans")
		}

		return nil
	}

	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}
