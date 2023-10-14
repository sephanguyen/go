package assignment

import (
	"github.com/manabie-com/backend/features/repository/syllabus/entity"
	"github.com/manabie-com/backend/internal/eureka/entities"
)

type StepState struct {
	DefaultSchoolID int32

	TopicIDs []string

	AssignmentDisplayOrder []int
	AssignmentIDs          []string
	Assignments            []*entities.Assignment
	ActualAssignments      []*entities.Assignment

	AssignmentOneQuery          entity.GraphqlAssignmentOneQuery
	AssignmentsByTopicIdQuery   entity.GraphqlAssignmentsByTopicIDQuery
	AssignmentsManyQuery        entity.GraphqlAssignmentsManyQuery
	AssignmentDisplayOrderQuery entity.GraphqlAssignmentDisplayOrder
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		`^user create a topic$`: s.userCreateATopic,

		`^a user insert some assignments with that topic id to database$`: s.aUserInsertSomeAssignmentsWithThatTopicIdToDatabase,

		`^user get assignments by call AssignmentDisplayOrder$`:     s.userGetAssignmentsByCallAssignmentDisplayOrder,
		`^our system must return AssignmentDisplayOrder correctly$`: s.ourSystemMustReturnAssignmentDisplayOrderCorrectly,

		`^user get assignments by call AssignmentOne$`:     s.userGetAssignmentsByCallAssignmentOne,
		`^our system must return AssignmentOne correctly$`: s.ourSystemMustReturnAssignmentOneCorrectly,

		`^user get assignments by call AssignmentsByTopicIds$`:     s.userGetAssignmentsByCallAssignmentsByTopicIds,
		`^our system must return AssignmentsByTopicIds correctly$`: s.ourSystemMustReturnAssignmentsByTopicIdsCorrectly,

		`^user get assignments by call AssignmentsMany$`:     s.userGetAssignmentsByCallAssignmentsMany,
		`^our system must return AssignmentsMany correctly$`: s.ourSystemMustReturnAssignmentsManyCorrectly,

		`^user get assignments by assignmentID$`:         s.userGetAssignmentsByAssignmentID,
		`^user get assignments by assignmentIDs$`:        s.userGetAssignmentsByAssignmentIDs,
		`^user get assignments by topicID$`:              s.userGetAssignmentsByTopicID,
		`^our system must return assignment correctly$`:  s.ourSystemMustReturnAssignmentCorrectly,
		`^our system must return assignments correctly$`: s.ourSystemMustReturnAssignmentsCorrectly,
	}

	return steps
}
