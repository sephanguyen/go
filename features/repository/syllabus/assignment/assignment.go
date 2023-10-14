package assignment

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
)

func (s *Suite) aUserInsertSomeAssignmentsWithThatTopicIdToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var assignments []*entities.Assignment
	for i := 0; i < rand.Intn(5)+3; i++ {
		assignment, err := GenerateAssignment(stepState.TopicIDs[0], i)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create an assignment, err: " + err.Error())
		}
		stepState.AssignmentIDs = append(stepState.AssignmentIDs, assignment.ID.String)
		stepState.AssignmentDisplayOrder = append(stepState.AssignmentDisplayOrder, int(assignment.DisplayOrder.Int))
		assignments = append(assignments, &assignment)
	}
	assignmentRepo := repositories.AssignmentRepo{}
	if err := assignmentRepo.BulkUpsert(ctx, s.DB, assignments); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert an assignment, err: " + err.Error())
	}
	stepState.Assignments = assignments
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetAssignmentsByAssignmentID(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	assignmentRepo := repositories.AssignmentRepo{}
	assignments, err := assignmentRepo.RetrieveAssignments(ctx, s.DB, database.TextArray([]string{stepState.AssignmentIDs[0]}))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve assignments by id array, err: " + err.Error())
	}
	stepState.ActualAssignments = assignments
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetAssignmentsByAssignmentIDs(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	assignmentRepo := repositories.AssignmentRepo{}
	assignments, err := assignmentRepo.RetrieveAssignments(ctx, s.DB, database.TextArray(stepState.AssignmentIDs))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve assignments by id, err: " + err.Error())
	}
	stepState.ActualAssignments = assignments
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetAssignmentsByTopicID(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	assignmentRepo := repositories.AssignmentRepo{}
	assignments, err := assignmentRepo.RetrieveAssignmentsByTopicIDs(ctx, s.DB, database.TextArray(stepState.TopicIDs))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to get assignments by topic id, err: " + err.Error())
	}
	stepState.ActualAssignments = assignments
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnAssignmentsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(stepState.Assignments) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("system return assignment incorrect, not found any")
	}
	for i := range stepState.ActualAssignments {
		if stepState.Assignments[i].ID.String != stepState.AssignmentIDs[i] ||
			stepState.Assignments[i].Name.String != stepState.ActualAssignments[i].Name.String {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("system return assignment incorrect, expected: %s, get: %s", stepState.AssignmentIDs[i], stepState.Assignments[i].ID.String)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnAssignmentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(stepState.Assignments) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("system return assignment incorrect, not found any")
	}
	if stepState.Assignments[0].ID.String != stepState.AssignmentIDs[0] {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("system return assignment incorrect, expected: %s, get: %s", stepState.AssignmentIDs[0], stepState.Assignments[0].ID.String)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateATopic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	topic, err := UserCreateAValidTopic(ctx, s.DB, idutil.ULIDNow(), int(stepState.DefaultSchoolID))
	stepState.TopicIDs = append(stepState.TopicIDs, topic.ID.String)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a topic, err: " + err.Error())
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func UserCreateAValidTopic(ctx context.Context, db database.Ext, chapterID string, defaultSchoolID int) (entities.Topic, error) {
	topic := &entities.Topic{}

	database.AllNullEntity(topic)
	id := idutil.ULIDNow()
	now := time.Now()

	if err := multierr.Combine(
		topic.SchoolID.Set(defaultSchoolID),
		topic.ID.Set(id),
		topic.ChapterID.Set(chapterID),
		topic.Name.Set("topic 1"),
		topic.Grade.Set(rand.Intn(11)+1),
		topic.Subject.Set(epb.Subject_SUBJECT_BIOLOGY),
		topic.Status.Set(epb.TopicStatus_TOPIC_STATUS_NONE),
		topic.CreatedAt.Set(now),
		topic.UpdatedAt.Set(now),
		topic.TotalLOs.Set(1),
		topic.TopicType.Set(epb.TopicType_TOPIC_TYPE_LEARNING),
		topic.EssayRequired.Set(true),
	); err != nil {
		return entities.Topic{}, err
	}

	topicRepo := repositories.TopicRepo{}
	if err := topicRepo.BulkUpsertWithoutDisplayOrder(ctx, db, []*entities.Topic{topic}); err != nil {
		return entities.Topic{}, err
	}

	return *topic, nil
}

func GenerateAssignment(topicID string, displayOrder int) (entities.Assignment, error) {
	id := idutil.ULIDNow()

	assignment := entities.Assignment{}
	if err := multierr.Combine(
		assignment.CreatedAt.Set(time.Now()),
		assignment.UpdatedAt.Set(time.Now()),
		assignment.DeletedAt.Set(nil),

		assignment.ID.Set(id),
		assignment.Content.Set(epb.AssignmentContent{
			TopicId: topicID,
			LoId:    []string{"lo-id-1", "lo-id-2"},
		}),
		assignment.Attachment.Set([]string{"media-id-1", "media-id-2"}),
		assignment.Settings.Set(&epb.AssignmentSetting{
			AllowLateSubmission: false,
			AllowResubmission:   false,
			RequireAttachment:   false,
		}),
		assignment.CheckList.Set(&epb.CheckList{
			Items: []*epb.CheckListItem{{Content: "Complete all learning objectives", IsChecked: true}, {Content: "Submitted required videos", IsChecked: false}},
		}),
		assignment.Type.Set(epb.AssignmentType_ASSIGNMENT_TYPE_LEARNING_OBJECTIVE),
		assignment.Status.Set(epb.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE),
		assignment.MaxGrade.Set(100),
		assignment.Instruction.Set("teacher's instruction"),
		assignment.Name.Set(fmt.Sprintf("assignment-%s", idutil.ULIDNow())),
		assignment.IsRequiredGrade.Set(false),
		assignment.DisplayOrder.Set(displayOrder),
		assignment.OriginalTopic.Set(topicID),
		assignment.TopicID.Set(topicID),
	); err != nil {
		return assignment, err
	}
	return assignment, nil
}
