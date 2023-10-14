package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	entities "github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ToDoItem struct {
	ChapterID string
	TopicID   string
	Type      string
	ID        string

	ChapterDisplayOrder int
	TopicDisplayOrder   int
	LoDisplayOrdrer     int
}

func (s *suite) userAddLeaningObjectivesToBook(ctx context.Context, loType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if ctx, err := s.createABook(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var (
		topics      []*pb.Topic
		los         []*cpb.LearningObjective
		assignments []*pb.Assignment
		toDoItems   []*ToDoItem
		topicIDs    []string
	)
	n := rand.Int31() | 3
	for i := 0; i < 2; i++ {
		cBit := int32(1 << i)
		if (cBit & n) == 0 {
			continue
		}
		if ctx, err := s.createAChapter(ctx, i); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		m := rand.Int31() | 3
		for j := 0; j < 3; j++ {
			tBit := int32(1 << j)
			if (tBit & m) == 0 {
				continue
			}
			topicID := idutil.ULIDNow()
			topic := &pb.Topic{
				Id:           topicID,
				Name:         fmt.Sprintf("topic-name-%s", topicID),
				Country:      pb.Country_COUNTRY_VN,
				Subject:      pb.Subject_SUBJECT_BIOLOGY,
				DisplayOrder: int32(j),
				SchoolId:     constants.ManabieSchool,

				ChapterId: stepState.ChapterID,
			}
			topics = append(topics, topic)
			topicIDs = append(topicIDs, topic.Id)
			k := rand.Int31() | 3
			for g := 2; g < 10; g++ {
				loBit := int32(1 << g)
				if (loBit & k) == 0 {
					continue
				}
				loID := idutil.ULIDNow()
				if g%2 == 0 {
					lo := &cpb.LearningObjective{
						Info: &cpb.ContentBasicInfo{
							Id:           loID,
							Name:         fmt.Sprintf("lo-name-%s", loID),
							Country:      cpb.Country_COUNTRY_VN,
							Subject:      cpb.Subject_SUBJECT_BIOLOGY,
							DisplayOrder: int32(g),
							SchoolId:     constants.ManabieSchool,
						},
						TopicId: topicID,
						Type:    cpb.LearningObjectiveType(int32(g)),
					}
					los = append(los, lo)
					if loType != AssignmentType {
						toDoItems = append(toDoItems, &ToDoItem{
							ChapterID: stepState.ChapterID,
							TopicID:   topicID,
							Type:      pb.ToDoItemType_TO_DO_ITEM_TYPE_LO.String(),
							ID:        loID,

							ChapterDisplayOrder: i,
							TopicDisplayOrder:   j,
							LoDisplayOrdrer:     g + 1,
						})
					}
				} else {
					assignment := &pb.Assignment{}
					ctx, assignment = s.generateAssignment(ctx, idutil.ULIDNow(), true, true, true)
					assignment.Content.TopicId = topicID
					assignment.DisplayOrder = int32(g)
					assignments = append(assignments, assignment)
					if loType != LearningObjectiveType {
						toDoItems = append(toDoItems, &ToDoItem{
							ChapterID: stepState.ChapterID,
							TopicID:   topicID,
							Type:      pb.ToDoItemType_TO_DO_ITEM_TYPE_ASSIGNMENT.String(),
							ID:        assignment.AssignmentId,

							ChapterDisplayOrder: i,
							TopicDisplayOrder:   j,
							LoDisplayOrdrer:     g + 1,
						})
					}
				}
			}
		}
	}
	if ctx, err := s.aSignedIn(ctx, "school admin"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(contextWithToken(s, ctx), &pb.UpsertTopicsRequest{
		Topics: topics,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}
	if loType != AssignmentType {
		if _, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(contextWithToken(s, ctx), &pb.UpsertLOsRequest{
			LearningObjectives: los,
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create los: %w", err)
		}
	}

	if loType != LearningObjectiveType {
		if _, err := pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), &pb.UpsertAssignmentsRequest{
			Assignments: assignments,
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create assignments: %w", err)
		}
	}

	if ctx, err := s.addBookToCourse(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TopicIDs = topicIDs
	stepState.ToDoItemList = toDoItems
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetListTodoItemsByTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewTopicReaderServiceClient(s.Conn).ListToDoItemsByTopics(contextWithToken(s, ctx), &pb.ListToDoItemsByTopicsRequest{
		TopicIds:    stepState.TopicIDs,
		StudyPlanId: wrapperspb.String(stepState.StudyPlanID),
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetListTodoItemsByTopicsWithInvalidStudyPlanId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewTopicReaderServiceClient(s.Conn).ListToDoItemsByTopics(contextWithToken(s, ctx), &pb.ListToDoItemsByTopicsRequest{
		TopicIds:    stepState.TopicIDs,
		StudyPlanId: wrapperspb.String("invalid-study-plan-id"),
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetListTodoItemsByTopicsWithNullStudyPlanId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewTopicReaderServiceClient(s.Conn).ListToDoItemsByTopics(contextWithToken(s, ctx), &pb.ListToDoItemsByTopicsRequest{
		TopicIds: stepState.TopicIDs,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsTodoItemsHaveOrderCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.ListToDoItemsByTopicsResponse)
	sort.SliceStable(stepState.ToDoItemList, func(i, j int) bool {
		item1 := stepState.ToDoItemList[i]
		item2 := stepState.ToDoItemList[j]
		if item1.ChapterDisplayOrder != item2.ChapterDisplayOrder {
			return item1.ChapterDisplayOrder < item2.ChapterDisplayOrder
		}
		if item1.TopicDisplayOrder != item2.TopicDisplayOrder {
			return item1.TopicDisplayOrder < item2.TopicDisplayOrder
		}
		return item1.LoDisplayOrdrer < item2.LoDisplayOrdrer
	})
	getKey := func(loType, id string) string {
		return strings.Join([]string{loType, id}, "|")
	}
	doMap := make(map[string]int)
	doTopicMap := make(map[string]int)
	count := 0
	for i, item := range resp.Items {
		doTopicMap[item.TopicId] = i
		for _, todoItem := range item.TodoItems {
			key := getKey(todoItem.Type.String(), todoItem.ResourceId)
			doMap[key] = count
			count++
		}
	}

	var (
		topicDo int
		topicID string
	)
	for i, item := range stepState.ToDoItemList {
		do, ok := doMap[getKey(item.Type, item.ID)]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("list to do items have return StepStateToContext(ctx, stepState), wrong, missing type %s id %s", item.Type, item.ID)
		}
		if do != i {
			return StepStateToContext(ctx, stepState), fmt.Errorf("position of type %s id %s is wrong, expect %v but got %v", item.Type, item.ID, i, do)
		}
		if item.TopicID != topicID {
			topicID = item.TopicID
			tdo, ok := doTopicMap[topicID]
			if !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("missing topic %v", topicID)
			}
			if tdo != topicDo {
				return StepStateToContext(ctx, stepState), fmt.Errorf("topic %s order is wrong, expect %v but got %v", topicID, topicDo, tdo)
			}
			topicDo++
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userAddALeaningObjectiveAndAnAssignmentWithSameTopicToBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.createABook(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if ctx, err := s.createAChapter(ctx, 1); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if ctx, err := s.aSignedIn(ctx, "school admin"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TopicID = idutil.ULIDNow()
	if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(contextWithToken(s, ctx), &pb.UpsertTopicsRequest{
		Topics: []*pb.Topic{
			{
				Id:           stepState.TopicID,
				Name:         fmt.Sprintf("topic-name-%s", stepState.TopicID),
				Country:      pb.Country_COUNTRY_VN,
				Subject:      pb.Subject_SUBJECT_BIOLOGY,
				DisplayOrder: int32(1),
				SchoolId:     constants.ManabieSchool,

				ChapterId: stepState.ChapterID,
			},
		},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}
	stepState.LoID = idutil.ULIDNow()
	if _, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(contextWithToken(s, ctx), &pb.UpsertLOsRequest{
		LearningObjectives: []*cpb.LearningObjective{
			{
				Info: &cpb.ContentBasicInfo{
					Id:           stepState.LoID,
					Name:         fmt.Sprintf("lo-name-%s", stepState.LoID),
					Country:      cpb.Country_COUNTRY_VN,
					Subject:      cpb.Subject_SUBJECT_BIOLOGY,
					DisplayOrder: int32(1),
					SchoolId:     constants.ManabieSchool,
				},
				TopicId: stepState.TopicID,
				Type:    cpb.LearningObjectiveType(1),
			},
		},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create los: %w", err)
	}
	ctx, assignment := s.generateAssignment(ctx, idutil.ULIDNow(), true, true, true)
	assignment.Content.TopicId = stepState.TopicID
	assignment.DisplayOrder = int32(2)
	stepState.AssignmentID = assignment.AssignmentId
	toDoItems := []*ToDoItem{
		{
			ChapterID: stepState.ChapterID,
			TopicID:   stepState.TopicID,
			Type:      pb.ToDoItemType_TO_DO_ITEM_TYPE_LO.String(),
			ID:        stepState.LoID,

			ChapterDisplayOrder: 1,
			TopicDisplayOrder:   1,
			LoDisplayOrdrer:     1,
		},
		{
			ChapterID: stepState.ChapterID,
			TopicID:   stepState.TopicID,
			Type:      pb.ToDoItemType_TO_DO_ITEM_TYPE_ASSIGNMENT.String(),
			ID:        assignment.AssignmentId,

			ChapterDisplayOrder: 1,
			TopicDisplayOrder:   1,
			LoDisplayOrdrer:     2,
		},
	}
	if _, err := pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), &pb.UpsertAssignmentsRequest{
		Assignments: []*pb.Assignment{assignment},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create assignments: %w", err)
	}
	if ctx, err := s.addBookToCourse(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.ToDoItemList = toDoItems
	stepState.TopicIDs = []string{stepState.TopicID}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeleteAInBook(ctx context.Context, loType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch loType {
	case LearningObjectiveType:
		e := &entities.LearningObjective{}
		loID := stepState.LoID

		stmt := fmt.Sprintf(`
			UPDATE %s
			SET deleted_at = NOW()
			WHERE lo_id = $1
		`, e.TableName())
		cmd, err := s.DB.Exec(ctx, stmt, &loID)

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to delete learning objective: %w", err)
		}
		if cmd.RowsAffected() != 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to delete learning objective")
		}
		err = try.Do(func(attempt int) (retry bool, err error) {
			loRepo := &repositories.LearningObjectiveRepo{}
			los, err := loRepo.RetrieveByIDs(ctx, s.DB, database.TextArray([]string{loID}))
			if err != nil {
				time.Sleep(1 * time.Second)
				return attempt < 10, fmt.Errorf("error get learning objective id='%v' err='%v'", loID, err)
			}
			if len(los) == 0 || los[0].DeletedAt.Status != pgtype.Null {
				return false, nil
			}
			time.Sleep(1 * time.Second)
			return attempt < 10, fmt.Errorf("wait for too long to sync delete learing objective %v", loID)
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.ToDoItemList = stepState.ToDoItemList[1:]

	case AssignmentType:
		assignmentID := stepState.AssignmentID

		if ctx, err := s.aSignedIn(ctx, "school admin"); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err := pb.NewAssignmentModifierServiceClient(s.Conn).DeleteAssignments(s.signedCtx(ctx), &pb.DeleteAssignmentsRequest{
			AssignmentIds: []string{assignmentID},
		})

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to delete assignment %v %w", assignmentID, err)
		}
		stepState.ToDoItemList = stepState.ToDoItemList[:1]
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetListTodoItemsByTopicsWithAvailableDates(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewTopicReaderServiceClient(s.Conn).ListToDoItemsByTopics(contextWithToken(s, ctx), &pb.ListToDoItemsByTopicsRequest{
		TopicIds:            stepState.TopicIDs,
		StudyPlanId:         wrapperspb.String(stepState.StudyPlanID),
		StudyPlanItemFilter: pb.StudyPlanItemFilter_STUDY_PLAN_ITEM_FILTER_AVAILABLE,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateAvailableDatesForStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := `
		UPDATE study_plan_items
		SET start_date = NOW(), end_date = (select CURRENT_DATE + INTERVAL '1 day'), available_from = NOW(), available_to = (select CURRENT_DATE + INTERVAL '1 day')
		WHERE study_plan_id = $1
	`
	_, err := s.DB.Exec(ctx, query, stepState.StudyPlanID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update start_date/available_from study plan item: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
