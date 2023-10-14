package learning_objectives

type StepState struct {
	TopicID string
	LoID    string
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^insert a valid learning objectives with "([^"]*)"$`: s.insertAValidLOInDB,
		`^update a valid record learning objectives$`:         s.updateLOs,
		`^there is a "([^"]*)" in the exam lo table$`:         s.thereIsARowLOsInDB,
	}
	return steps
}
