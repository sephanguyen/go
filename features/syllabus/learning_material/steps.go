package learning_material

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
)

type StepState struct {
	Token               string
	Response            interface{}
	Request             interface{}
	ResponseErr         error
	BookID              string
	TopicIDs            []string
	ChapterIDs          []string
	SchoolAdmin         entity.SchoolAdmin
	Student             entity.Student
	LearningMaterialID  string
	LearningMaterialIDs []string
	MapLearningMaterial map[string]*entity.LearningMaterialPb

	FlashcardIDs         []string
	ExamLOIDs            []string
	LearningObjectiveIDs []string
	GeneralAssignmentIDs []string
	TaskAssignmentIDs    []string

	ExamLOs              []entities.ExamLO
	Flashcards           []entities.Flashcard
	LearningObjectiveV2s []entities.LearningObjectiveV2
	Assignments          []entities.GeneralAssignment
	TaskAssignments      []entities.TaskAssignment

	LMName    string
	CurrentLM *entity.LearningMaterialPb
}

func InitStep(s *Suite) map[string]interface{} {
	steps := map[string]interface{}{
		// BEGIN====common=====BEGIN
		`^<learning_material>a signed in "([^"]*)"$`:         s.aSignedIn,
		`^<learning_material>a valid book content$`:          s.aValidBookContent,
		`^<learning_material>returns "([^"]*)" status code$`: s.returnsStatusCode,
		// END==== common ===END
		`^some existing learning materials in an arbitrary topic of the book$`: s.someExistingLMInTopic,
		// delete learning material
		`^user deletes an arbitrary learning material$`:   s.userDeletesAnArbitraryTheLearningMaterial,
		`^user deletes the "([^"]*)"$`:                    s.userDeletesTheLearningMaterial,
		`^our system must delete the "([^"]*)" correctly`: s.ourSystemMustDeleteTheLearningMaterialCorrectly,
		`^user deletes the "([^"]*)" with wrong ID$`:      s.userDeletesTheLearningMaterialWithWrongID,
		// swap learning material
		`^user swap LM display order$`:                                         s.swapLMDisplayOrder,
		`^our system must swap display orders of learning material correctly$`: s.displayOrdersOfFlashcardsSwapped,
		// list learning material
		`^user send list arbitrary learning material request$`:           s.userSendListArbitraryLearningMaterialRequest,
		`^our system must return learning material "([^"]*)" correctly$`: s.ourSystemMustReturnLearningMaterialCorrectly,
		`^user send list learning material "([^"]*)" request$`:           s.userSendListLearningMaterialRequest,
		// duplicate book
		`^a valid book in db$`:                           s.aValidBookInDB,
		`^our system must return copied book correctly$`: s.ourSystemMustReturnCopiedBookCorrectly,
		`^user send duplicate book request$`:             s.userSendDuplicateBookRequest,
		// update learning material name
		`^user update LM name$`: s.userUpdateLMName,
		`^our system must update learning material name correctly$`: s.ourSystemMustUpdateLearningMaterialNameCorrectly,
	}
	return steps
}
func (s *Suite) aSignedIn(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// reset token
	stepState.Token = ""
	_, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, arg)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.Token = authToken
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsStatusCode(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	return utils.StepStateToContext(ctx, stepState), utils.ValidateStatusCode(stepState.ResponseErr, arg)
}

// someExistingLMInTopic for simple this func only ensure each type existed in the topic.
func (s *Suite) someExistingLMInTopic(ctx context.Context) (context.Context, error) {
	var err error
	stepState := utils.StepStateFromContext[StepState](ctx)
	// init the map
	stepState.MapLearningMaterial = make(map[string]*entity.LearningMaterialPb)

	// not insert parellel to ensure our expectation: I want to get the display_order without retrieve the database.
	for lmType, i := range sspb.LearningMaterialType_value {
		switch lmType {
		case sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String():
			if ctx, err = s.simpleInsertLO(ctx, i); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
		case sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String():
			if ctx, err = s.simpleInsertAssignment(ctx, i); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
		case sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String():
			if ctx, err = s.simpleInsertFlashcard(ctx, i); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
		case sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String():
			if ctx, err = s.simpleInsertTaskAssignment(ctx, i); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
		case sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String():
			if ctx, err = s.simpleInsertExamLO(ctx, i); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) simpleInsertTaskAssignment(ctx context.Context, val int32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.InsertTaskAssignmentRequest{
		TaskAssignment: &sspb.TaskAssignmentBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    fmt.Sprintf("task_assignment_%d", val),
			}}}

	resp, err := sspb.NewTaskAssignmentClient(s.EurekaConn).InsertTaskAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert the task assignment: %w", err)
	}
	taskAssignment := &entity.LearningMaterialPb{
		LearningMaterialBase: req.TaskAssignment.Base,
	}
	taskAssignment.LearningMaterialBase.LearningMaterialId = resp.GetLearningMaterialId()
	stepState.MapLearningMaterial["task_assignment"] = taskAssignment
	stepState.TaskAssignmentIDs = append(stepState.TaskAssignmentIDs, taskAssignment.LearningMaterialId)
	stepState.LearningMaterialIDs = append(stepState.LearningObjectiveIDs, taskAssignment.LearningMaterialId)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) simpleInsertAssignment(ctx context.Context, val int32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.InsertAssignmentRequest{
		Assignment: &sspb.AssignmentBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    fmt.Sprintf("assignment_%d", val),
			}}}

	resp, err := sspb.NewAssignmentClient(s.EurekaConn).InsertAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert the assignment: %w", err)
	}
	assignment := &entity.LearningMaterialPb{
		LearningMaterialBase: req.Assignment.Base,
	}
	assignment.LearningMaterialBase.LearningMaterialId = resp.GetLearningMaterialId()
	stepState.MapLearningMaterial["assignment"] = assignment
	stepState.GeneralAssignmentIDs = append(stepState.GeneralAssignmentIDs, assignment.LearningMaterialId)
	stepState.LearningMaterialIDs = append(stepState.LearningObjectiveIDs, assignment.LearningMaterialId)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) simpleInsertLO(ctx context.Context, val int32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.InsertLearningObjectiveRequest{
		LearningObjective: &sspb.LearningObjectiveBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    fmt.Sprintf("learning_objective_%d", val),
			},
		},
	}
	resp, err := sspb.NewLearningObjectiveClient(s.EurekaConn).InsertLearningObjective(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert the learning_objective: %w", err)
	}
	temp := &entity.LearningMaterialPb{
		LearningMaterialBase: req.LearningObjective.Base,
	}
	temp.LearningMaterialBase.LearningMaterialId = resp.GetLearningMaterialId()
	stepState.MapLearningMaterial["learning_objective"] = temp
	stepState.LearningObjectiveIDs = append(stepState.LearningObjectiveIDs, temp.LearningMaterialId)
	stepState.LearningMaterialIDs = append(stepState.LearningObjectiveIDs, temp.LearningMaterialId)
	return utils.StepStateToContext(ctx, stepState), nil
}
func (s *Suite) simpleInsertFlashcard(ctx context.Context, val int32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &sspb.InsertFlashcardRequest{
		Flashcard: &sspb.FlashcardBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    fmt.Sprintf("flashcard_%d", val),
			},
		},
	}
	resp, err := sspb.NewFlashcardClient(s.EurekaConn).InsertFlashcard(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert the flashcard: %w", err)
	}
	temp := &entity.LearningMaterialPb{
		LearningMaterialBase: req.Flashcard.Base,
	}
	temp.LearningMaterialBase.LearningMaterialId = resp.GetLearningMaterialId()
	stepState.MapLearningMaterial["flashcard"] = temp
	stepState.FlashcardIDs = append(stepState.FlashcardIDs, temp.LearningMaterialId)
	stepState.LearningMaterialIDs = append(stepState.LearningObjectiveIDs, temp.LearningMaterialId)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) simpleInsertExamLO(ctx context.Context, val int32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &sspb.InsertExamLORequest{
		ExamLo: &sspb.ExamLOBase{
			Base: &sspb.LearningMaterialBase{
				Name:    fmt.Sprintf("exam_lo_%d", val),
				TopicId: stepState.TopicIDs[0],
			},
		},
	}
	resp, err := sspb.NewExamLOClient(s.EurekaConn).InsertExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert the examLO: %w", err)
	}
	temp := &entity.LearningMaterialPb{
		LearningMaterialBase: req.ExamLo.Base,
	}
	temp.LearningMaterialBase.LearningMaterialId = resp.GetLearningMaterialId()
	stepState.MapLearningMaterial["exam_lo"] = temp
	stepState.ExamLOIDs = append(stepState.ExamLOIDs, temp.LearningMaterialId)
	stepState.LearningMaterialIDs = append(stepState.LearningObjectiveIDs, temp.LearningMaterialId)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aValidBookInDB(ctx context.Context) (context.Context, error) {
	// upsert a random book
	stepState := utils.StepStateFromContext[StepState](ctx)
	bookResp, err := epb.NewBookModifierServiceClient(s.EurekaConn).UpsertBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertBooksRequest{
		Books: utils.GenerateBooks(1, nil),
	})
	if err != nil {
		err = fmt.Errorf("NewBookModifierService.UpsertBooks: %w", err)
		return utils.StepStateToContext(ctx, stepState), err
	}
	bookID := bookResp.BookIds[0]

	// upsert 1 chapter to book id
	numChapters := rand.Intn(3) + 2
	chapterResp, err := epb.NewChapterModifierServiceClient(s.EurekaConn).UpsertChapters(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertChaptersRequest{
		Chapters: utils.GenerateChapters(bookID, numChapters, nil),
		BookId:   bookID,
	})
	if err != nil {
		err = fmt.Errorf("NewChapterModifierService.UpsertChapters: %w", err)

		return utils.StepStateToContext(ctx, stepState), err
	}
	chapterIDs := chapterResp.GetChapterIds()

	// upsert random number of topics to chapter
	numTopics := rand.Intn(3) + 1
	var topicIDs = make([]string, 0, numTopics)
	topicResp, err := epb.NewTopicModifierServiceClient(s.EurekaConn).Upsert(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertTopicsRequest{
		Topics: utils.GenerateTopics(chapterIDs[0], numTopics, nil),
	})
	if err != nil {
		err = fmt.Errorf("NewTopicModifierService.Upsert: %w", err)
		return utils.StepStateToContext(ctx, stepState), err
	}
	topicIDs = append(topicIDs, topicResp.GetTopicIds()...)

	// insert each type of lm to each topic
	n := len(topicIDs)
	wg := &sync.WaitGroup{}
	cErrs := make(chan error, n)
	defer func() {
		close(cErrs)
	}()

	genAndInsert := func(ctx context.Context, i int, wg *sync.WaitGroup) {
		defer wg.Done()
		// insert examLO
		_, err = sspb.NewExamLOClient(s.EurekaConn).InsertExamLO(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.InsertExamLORequest{
			ExamLo: &sspb.ExamLOBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: topicIDs[i],
					Name:    fmt.Sprintf("Exam-LO-name+%s", idutil.ULIDNow()),
				},
				Instruction: "Exam-Lo-instruction",
			},
		})
		if err != nil {
			err = fmt.Errorf("NewExamLOClient.InsertExamLO: %w", err)
		}

		// insert flashcard
		_, err = sspb.NewFlashcardClient(s.EurekaConn).InsertFlashcard(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.InsertFlashcardRequest{
			Flashcard: &sspb.FlashcardBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: topicIDs[i],
					Name:    fmt.Sprintf("flashcard-name-%s", idutil.ULIDNow()),
				},
			},
		})
		if err != nil {
			err = fmt.Errorf("NewFlashcardClient.InsertFlashcard: %w", err)
		}

		// insert learning objective
		_, err = sspb.NewLearningObjectiveClient(s.EurekaConn).InsertLearningObjective(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.InsertLearningObjectiveRequest{
			LearningObjective: &sspb.LearningObjectiveBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: topicIDs[i],
					Name:    fmt.Sprintf("LO-name+%s", idutil.ULIDNow()),
				},
			},
		})
		if err != nil {
			err = fmt.Errorf("NewLearningObjectiveClient.InsertLearningObjective: %w", err)
		}

		// insert assignment
		insertAssignmentReq := &sspb.InsertAssignmentRequest{
			Assignment: &sspb.AssignmentBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: topicIDs[i],
					Name:    "assignment-name",
				},
				Attachments:            []string{"attachment-1", "attachment-2"},
				Instruction:            "instruction",
				MaxGrade:               10,
				IsRequiredGrade:        true,
				AllowResubmission:      false,
				RequireAttachment:      true,
				AllowLateSubmission:    false,
				RequireAssignmentNote:  true,
				RequireVideoSubmission: false,
			},
		}
		_, err := sspb.NewAssignmentClient(s.EurekaConn).InsertAssignment(s.AuthHelper.SignedCtx((ctx), stepState.Token), insertAssignmentReq)
		if err != nil {
			err = fmt.Errorf("NewAssignmentClient.InsertAssignment: %w", err)
		}

		// insert task assignment
		_, err = sspb.NewTaskAssignmentClient(s.EurekaConn).InsertTaskAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.InsertTaskAssignmentRequest{
			TaskAssignment: &sspb.TaskAssignmentBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: topicIDs[i],
					Name:    fmt.Sprintf("task-assignment-name-%s", idutil.ULIDNow()),
				},
				Attachments: []string{"attachment-1", "attachment-2"},
				Instruction: "instruction",
			},
		})
		if err != nil {
			err = fmt.Errorf("NewTaskAssignmentClient.InsertTaskAssignment: %w", err)
		}
		cErrs <- err
	}
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go genAndInsert(ctx, i, wg)
	}
	go func() {
		wg.Wait()
	}()
	for i := 0; i < n; i++ {
		errTemp := <-cErrs
		err = multierr.Combine(err, errTemp)
	}
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	stepState.BookID = bookID
	stepState.ChapterIDs = chapterIDs
	stepState.TopicIDs = topicIDs
	return utils.StepStateToContext(ctx, stepState), nil
}
