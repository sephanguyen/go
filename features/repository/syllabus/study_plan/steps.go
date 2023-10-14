package study_plan

import "github.com/manabie-com/backend/features/repository/syllabus/entity"

type StepState struct {
	StudyPlanID     string
	StudyPlanItemID string
	AssignmentID    string
	TopicID         string
	LoID            string

	StudyPlanOneQuery entity.GraphqlStudyPlanOneQuery
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^a valid StudyPlan in database$`:                          s.aUserInsertStudyPlanToDatabase,
		`^a user call StudyPlanOne$`:                               s.aUserCallStudyPlanOne,
		`^our System will return all information about StudyPlan$`: s.ourSystemWillReturnStudyPlan,
	}
	return steps
}
