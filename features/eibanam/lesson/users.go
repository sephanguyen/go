package lesson

import (
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/eibanam"
)

func (s *suite) loginsCMS(role string) error {
	// new school
	school, err := s.helper.CreateSchool()
	if err != nil {
		return err
	}
	s.CurrentSchoolID = school.ID.Int

	userCredential, err := s.helper.SignedInAsAccount(s.CurrentSchoolID, eibanam.Role(role))
	if err != nil {
		return err
	}
	s.AddUserCredential(userCredential)

	return nil
}

func (s *suite) loginsTeacherApp(name string) error {
	userCredential, err := s.helper.SignedInAsAccount(s.CurrentSchoolID, eibanam.RoleTeacher)
	if err != nil {
		return err
	}
	s.AddUserCredential(userCredential)
	s.AddUserCredentialByName(userCredential, name)
	s.AddTeacherIDs(userCredential.UserID)

	return nil
}

func (s *suite) createTeacher() error {
	id, _, err := s.helper.CreateUser(s.CurrentSchoolID, eibanam.RoleTeacher)
	if err != nil {
		return fmt.Errorf("could not create new teacher %v", err)
	}
	s.AddTeacherIDs(id)

	return nil
}

func (s *suite) loginsLearnerApp(name string) error {
	// get role from name
	var role string
	res := strings.SplitN(name, " ", 2)
	role = res[0]

	if role != string(eibanam.RoleParent) && role != string(eibanam.RoleStudent) {
		return fmt.Errorf("could not login learner app with role %s", role)
	}

	userCredential, err := s.helper.SignedInAsAccount(s.CurrentSchoolID, eibanam.Role(role))
	if err != nil {
		return err
	}
	s.AddUserCredential(userCredential)
	s.AddUserCredentialByName(userCredential, name)
	s.AddStudentIDs(userCredential.UserID)

	return nil
}

func (s *suite) createStudent() error {
	id, _, err := s.helper.CreateUser(s.CurrentSchoolID, eibanam.RoleStudent)
	if err != nil {
		return fmt.Errorf("could not create new student %v", err)
	}
	s.AddStudentIDs(id)

	return nil
}
