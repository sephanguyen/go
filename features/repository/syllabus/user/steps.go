package user

type StepState struct {
	UserIDs          []string
	StudentID        string
	CurrentStudentID string
	RoleIDs          []string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^create a list of user on bob$`:      s.createAListOfUserOnBob,
		`^users created correctly in eureka$`: s.usersCreatedCorrectlyInEureka,
	}
	return steps
}
