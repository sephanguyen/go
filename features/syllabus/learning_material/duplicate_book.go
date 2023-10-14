package learning_material

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
)

func (s *Suite) userSendDuplicateBookRequest(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).DuplicateBook(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.DuplicateBookRequest{
		BookId:   stepState.BookID,
		BookName: "BookName",
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnCopiedBookCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.DuplicateBookResponse)

	orgBook, orgChapters, orgTopics, orgTopicIDs, err := s.getAllBookContentByBookID(ctx, stepState.BookID) // original book
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.getAllBookContentByBookID: failed to get original book content: %w", err)
	}
	newBook, newChapters, newTopics, newTopicIDs, err := s.getAllBookContentByBookID(ctx, resp.NewBookID) // new book
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.getAllBookContentByBookID: failed to get new book content: %w", err)
	}

	ctx, err = s.getAllLearningMaterialTypeByTopicID(ctx, orgTopicIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.getAllLearningMaterialTypeByTopicID: failed to get all original lm type: %w", err)
	}
	orgExamLOs := stepState.ExamLOs
	orgFlashcards := stepState.Flashcards
	orgLos := stepState.LearningObjectiveV2s
	orgAssignments := stepState.Assignments
	orgTaskAssignments := stepState.TaskAssignments
	ctx, err = s.getAllLearningMaterialTypeByTopicID(ctx, newTopicIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.getAllLearningMaterialTypeByTopicID: failed to get all new lm type: %w", err)
	}
	if err := multierr.Combine(
		s.compareBook(orgBook, newBook),
		s.compareChapters(orgChapters, newChapters),
		s.compareTopics(orgTopics, newTopics),

		s.compareLMTypes(orgExamLOs, stepState.ExamLOs, len(orgExamLOs), len(stepState.ExamLOs), sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
		s.compareLMTypes(orgFlashcards, stepState.Flashcards, len(orgFlashcards), len(stepState.Flashcards), sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
		s.compareLMTypes(orgLos, stepState.LearningObjectiveV2s, len(orgLos), len(stepState.LearningObjectiveV2s), sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String()),
		s.compareLMTypes(orgAssignments, stepState.Assignments, len(orgAssignments), len(stepState.Assignments), sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String()),
		s.compareLMTypes(orgTaskAssignments, stepState.TaskAssignments, len(orgTaskAssignments), len(stepState.TaskAssignments), sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("our system return incorrect, err: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// get all learning material type include examlo, flashcard, learning objectivev2, task assignment, assignmentv2
func (s *Suite) getAllLearningMaterialTypeByTopicID(ctx context.Context, topicIDs []string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	examLOs, err := s.getExamLOsByTopicIDs(ctx, topicIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error when get exam los by topic id: %w", err)
	}
	flashcards, err := s.getFlashCardsByTopicIDs(ctx, topicIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error when get flashcards by topic id: %w", err)
	}
	learningObjectives, err := s.getLearningObjectivesByTopicIDs(ctx, topicIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error when get los by topic id: %w", err)
	}
	assignments, err := s.getAssignmentsByTopicIDs(ctx, topicIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error when get assignments by topic id: %w", err)
	}
	taskAssignments, err := s.getTaskAssignmentsByTopicIDs(ctx, topicIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error when get task assignments by topic id: %w", err)
	}
	stepState.ExamLOs = examLOs
	stepState.Flashcards = flashcards
	stepState.Assignments = assignments
	stepState.LearningObjectiveV2s = learningObjectives
	stepState.TaskAssignments = taskAssignments
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getExamLOsByTopicIDs(ctx context.Context, topicIDs []string) ([]entities.ExamLO, error) {
	var examLO entities.ExamLO
	var examLOs []entities.ExamLO
	queryExamLOs := fmt.Sprintf(`SELECT learning_material_id, topic_id, name FROM %s WHERE topic_id = ANY($1::_TEXT)`, examLO.TableName())
	rows, err := s.EurekaDB.Query(ctx, queryExamLOs, topicIDs)
	if err != nil {
		return examLOs, fmt.Errorf("rows error: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("row.Err: %w", err)
		}
		database.AllNullEntity(&examLO)
		if err := rows.Scan(&examLO.ID, &examLO.TopicID, &examLO.Name); err != nil {
			return examLOs, fmt.Errorf("rows error: %w", err)
		}
		examLOs = append(examLOs, examLO)
	}
	return examLOs, nil
}

func (s *Suite) getFlashCardsByTopicIDs(ctx context.Context, topicIDs []string) ([]entities.Flashcard, error) {
	var flashcard entities.Flashcard
	var flashcards []entities.Flashcard
	queryFlashcards := fmt.Sprintf(`SELECT learning_material_id, topic_id, name FROM %s WHERE topic_id = ANY($1::_TEXT)`, flashcard.TableName())
	rows, err := s.EurekaDB.Query(ctx, queryFlashcards, topicIDs)
	if err != nil {
		return flashcards, fmt.Errorf("rows error: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("row.Err: %w", err)
		}
		database.AllNullEntity(&flashcard)
		if err := rows.Scan(&flashcard.ID, &flashcard.TopicID, &flashcard.Name); err != nil {
			return flashcards, fmt.Errorf("rows error: %w", err)
		}
		flashcards = append(flashcards, flashcard)
	}
	return flashcards, nil
}

func (s *Suite) getLearningObjectivesByTopicIDs(ctx context.Context, topicIDs []string) ([]entities.LearningObjectiveV2, error) {
	var learningObjective entities.LearningObjectiveV2
	var learningObjectives []entities.LearningObjectiveV2
	queryLearningObjectives := fmt.Sprintf(`SELECT learning_material_id, topic_id, name FROM %s WHERE topic_id = ANY($1::_TEXT)`, learningObjective.TableName())
	rows, err := s.EurekaDB.Query(ctx, queryLearningObjectives, topicIDs)
	if err != nil {
		return learningObjectives, fmt.Errorf("rows error: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("row.Err: %w", err)
		}
		database.AllNullEntity(&learningObjective)
		if err := rows.Scan(&learningObjective.ID, &learningObjective.TopicID, &learningObjective.Name); err != nil {
			return learningObjectives, fmt.Errorf("rows error: %w", err)
		}
		learningObjectives = append(learningObjectives, learningObjective)
	}
	return learningObjectives, nil
}

func (s *Suite) getAssignmentsByTopicIDs(ctx context.Context, topicIDs []string) ([]entities.GeneralAssignment, error) {
	var generalAssignment entities.GeneralAssignment
	var generalAssignments []entities.GeneralAssignment
	queryGeneralAssignment := fmt.Sprintf(`SELECT learning_material_id, topic_id, name FROM %s WHERE topic_id = ANY($1::_TEXT)`, generalAssignment.TableName())
	rows, err := s.EurekaDB.Query(ctx, queryGeneralAssignment, topicIDs)
	if err != nil {
		return generalAssignments, fmt.Errorf("rows error: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("row.Err: %w", err)
		}
		database.AllNullEntity(&generalAssignment)
		if err := rows.Scan(&generalAssignment.ID, &generalAssignment.TopicID, &generalAssignment.Name); err != nil {
			return generalAssignments, fmt.Errorf("rows error: %w", err)
		}
		generalAssignments = append(generalAssignments, generalAssignment)
	}
	return generalAssignments, nil
}

func (s *Suite) getTaskAssignmentsByTopicIDs(ctx context.Context, topicIDs []string) ([]entities.TaskAssignment, error) {
	var taskAssignment entities.TaskAssignment
	var taskAssignments []entities.TaskAssignment
	queryTaskAssignment := fmt.Sprintf(`SELECT learning_material_id, topic_id, name FROM %s WHERE topic_id = ANY($1::_TEXT)`, taskAssignment.TableName())
	rows, err := s.EurekaDB.Query(ctx, queryTaskAssignment, topicIDs)
	if err != nil {
		return taskAssignments, fmt.Errorf("rows error: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if rows.Err() != nil {
			return nil, fmt.Errorf("row.Err: %w", err)
		}
		database.AllNullEntity(&taskAssignment)
		if err := rows.Scan(&taskAssignment.ID, &taskAssignment.TopicID, &taskAssignment.Name); err != nil {
			return taskAssignments, fmt.Errorf("rows error: %w", err)
		}
		taskAssignments = append(taskAssignments, taskAssignment)
	}
	return taskAssignments, nil
}

// get all book content include book, chapter, topic
func (s *Suite) getAllBookContentByBookID(ctx context.Context, bookID string) (entities.Book, []entities.Chapter, []entities.Topic, []string, error) {
	book, err := s.getBookByID(ctx, bookID)
	if err != nil {
		return book, nil, nil, nil, fmt.Errorf("error when get book by id: %w", err)
	}

	chapters, chapterIDs, err := s.getChaptersByBookID(ctx, bookID)
	if err != nil {
		return book, nil, nil, nil, fmt.Errorf("error when get chapters by book id: %w", err)
	}

	topics, topicIDs, err := s.getTopicsByChapterIDs(ctx, chapterIDs)
	if err != nil {
		return book, nil, nil, nil, fmt.Errorf("error when get topics by chapter ids: %w", err)
	}

	return book, chapters, topics, topicIDs, nil
}

func (s *Suite) getBookByID(ctx context.Context, bookID string) (entities.Book, error) {
	var book entities.Book
	queryBook := fmt.Sprintf(`SELECT book_id, name FROM %s WHERE book_id = $1 LIMIT 1`, book.TableName())
	if err := s.EurekaDB.QueryRow(ctx, queryBook, &bookID).Scan(&book.ID, &book.Name); err != nil {
		return book, fmt.Errorf("rows error: %w", err)
	}
	return book, nil
}

func (s *Suite) getChaptersByBookID(ctx context.Context, bookID string) ([]entities.Chapter, []string, error) {
	var chapterIDs []string
	var chapters []entities.Chapter
	var chapter entities.Chapter

	// query chapters
	queryChapters := fmt.Sprintf(`SELECT chapter_id, name FROM %s WHERE book_id = $1`, chapter.TableName())
	rowChapters, err := s.EurekaDB.Query(ctx, queryChapters, &bookID)
	if err != nil {
		return chapters, chapterIDs, fmt.Errorf("rows error: %w", err)
	}
	defer rowChapters.Close()
	for rowChapters.Next() {
		if err := rowChapters.Scan(&chapter.ID, &chapter.Name); err != nil {
			return chapters, chapterIDs, fmt.Errorf("rows error: %w", err)
		}
		chapters = append(chapters, chapter)
		chapterIDs = append(chapterIDs, chapter.ID.String)
	}
	return chapters, chapterIDs, nil
}

func (s *Suite) getTopicsByChapterIDs(ctx context.Context, chapterIDs []string) ([]entities.Topic, []string, error) {
	var topic entities.Topic
	var topics []entities.Topic
	var topicIDs []string
	queryTopics := fmt.Sprintf(`SELECT topic_id, name, chapter_id, total_los FROM %s WHERE chapter_id = ANY($1::_TEXT)`, topic.TableName())
	rowTopics, err := s.EurekaDB.Query(ctx, queryTopics, &chapterIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("rows error: %w", err)
	}
	defer rowTopics.Close()
	for rowTopics.Next() {
		if err := rowTopics.Scan(&topic.ID, &topic.Name, &topic.ChapterID, &topic.TotalLOs); err != nil {
			return nil, nil, fmt.Errorf("rows error: %w", err)
		}
		topics = append(topics, topic)
		topicIDs = append(topicIDs, topic.ID.String)
	}
	return topics, topicIDs, nil
}

// compare function
func (s *Suite) compareBook(book1 entities.Book, book2 entities.Book) error {
	if book1.ID == book2.ID {
		return fmt.Errorf("wrong book id return")
	}
	return nil
}

func (s *Suite) compareChapters(oldChapters []entities.Chapter, newChapters []entities.Chapter) error {
	if len(oldChapters) != len(newChapters) {
		return fmt.Errorf("length chapters mismatch, expected equal, got: %d and %d", len(oldChapters), len(newChapters))
	}
	var chapters1ID = make([]string, 0, len(oldChapters))
	var chapters2ID = make([]string, 0, len(oldChapters))
	var chapters1Name = make([]string, 0, len(oldChapters))
	var chapters2Name = make([]string, 0, len(oldChapters))

	for i := range oldChapters {
		chapters1ID = append(chapters1ID, oldChapters[i].ID.String)
		chapters2ID = append(chapters2ID, newChapters[i].ID.String)
		chapters1Name = append(chapters1Name, oldChapters[i].Name.String)
		chapters2Name = append(chapters2Name, newChapters[i].Name.String)
	}

	for i := range oldChapters {
		if golibs.InArrayString(chapters1ID[i], chapters2ID) || !golibs.InArrayString(chapters1Name[i], chapters2Name) {
			return fmt.Errorf("wrong id or name chapter return")
		}
	}
	return nil
}

func (s *Suite) compareTopics(topics1 []entities.Topic, topics2 []entities.Topic) error {
	if len(topics1) != len(topics2) {
		return fmt.Errorf("length topics mismatch, expected equal, got: %d and %d", len(topics1), len(topics2))
	}
	var topics1ID = make([]string, 0, len(topics1))
	var topics2ID = make([]string, 0, len(topics1))
	var topics1Name = make([]string, 0, len(topics1))
	var topics2Name = make([]string, 0, len(topics1))
	var topics1ChapterID = make([]string, 0, len(topics1))
	var topics2ChapterID = make([]string, 0, len(topics1))

	for i := range topics1 {
		topics1ID = append(topics1ID, topics1[i].ID.String)
		topics2ID = append(topics2ID, topics2[i].ID.String)
		topics1ChapterID = append(topics1ChapterID, topics1[i].ChapterID.String)
		topics2ChapterID = append(topics2ChapterID, topics2[i].ChapterID.String)
		topics1Name = append(topics1Name, topics1[i].Name.String)
		topics2Name = append(topics2Name, topics2[i].Name.String)
	}

	for i := range topics1 {
		if golibs.InArrayString(topics1ID[i], topics2ID) || golibs.InArrayString(topics1ChapterID[i], topics2ChapterID) || !golibs.InArrayString(topics1Name[i], topics2Name) {
			return fmt.Errorf("wrong topic field return")
		}
	}
	return nil
}

func (s *Suite) compareLMTypes(orgLMs interface{}, newLMs interface{}, lenOrgLM int, lenNewLM int, types string) error {
	if lenOrgLM != lenNewLM {
		return fmt.Errorf("length learning material types mismatch, expected equal, got: %d and %d", lenOrgLM, lenNewLM)
	}
	var lm1ID = make([]string, 0, lenOrgLM)
	var lm2ID = make([]string, 0, lenOrgLM)
	var lm1Name = make([]string, 0, lenOrgLM)
	var lm2Name = make([]string, 0, lenOrgLM)
	var lm1TopicIDs = make([]string, 0, lenOrgLM)
	var lm2TopicIDs = make([]string, 0, lenOrgLM)
	switch types {
	case sspb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT.String():
		orgAssignments, okOrg := orgLMs.([]entities.GeneralAssignment)
		newAssignments, okNew := newLMs.([]entities.GeneralAssignment)
		if !okOrg || !okNew {
			return fmt.Errorf("cannot convert to general assignment")
		}

		for i := 0; i < lenOrgLM; i++ {
			lm1ID = append(lm1ID, orgAssignments[i].ID.String)
			lm2ID = append(lm2ID, newAssignments[i].ID.String)
			lm1Name = append(lm1Name, orgAssignments[i].Name.String)
			lm2Name = append(lm2Name, newAssignments[i].Name.String)
			lm1TopicIDs = append(lm1TopicIDs, orgAssignments[i].TopicID.String)
			lm2TopicIDs = append(lm2TopicIDs, newAssignments[i].TopicID.String)
		}
	case sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String():
		orgExamLO, okOrg := orgLMs.([]entities.ExamLO)
		newExamLO, okNew := newLMs.([]entities.ExamLO)
		if !okOrg || !okNew {
			return fmt.Errorf("cannot convert to exam lo")
		}

		for i := 0; i < lenOrgLM; i++ {
			lm1ID = append(lm1ID, orgExamLO[i].ID.String)
			lm2ID = append(lm2ID, newExamLO[i].ID.String)
			lm1Name = append(lm1Name, orgExamLO[i].Name.String)
			lm2Name = append(lm2Name, newExamLO[i].Name.String)
			lm1TopicIDs = append(lm1TopicIDs, orgExamLO[i].TopicID.String)
			lm2TopicIDs = append(lm2TopicIDs, newExamLO[i].TopicID.String)
		}
	case sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String():
		orgFlashcards, okOrg := orgLMs.([]entities.Flashcard)
		newFlashcards, okNew := newLMs.([]entities.Flashcard)
		if !okOrg || !okNew {
			return fmt.Errorf("cannot convert to flash card")
		}

		for i := 0; i < lenOrgLM; i++ {
			lm1ID = append(lm1ID, orgFlashcards[i].ID.String)
			lm2ID = append(lm2ID, newFlashcards[i].ID.String)
			lm1Name = append(lm1Name, orgFlashcards[i].Name.String)
			lm2Name = append(lm2Name, newFlashcards[i].Name.String)
			lm1TopicIDs = append(lm1TopicIDs, orgFlashcards[i].TopicID.String)
			lm2TopicIDs = append(lm2TopicIDs, newFlashcards[i].TopicID.String)
		}
	case sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE.String():
		orgLOs, okOrg := orgLMs.([]entities.LearningObjectiveV2)
		newLOs, okNew := newLMs.([]entities.LearningObjectiveV2)
		if !okOrg || !okNew {
			return fmt.Errorf("cannot convert to learning objective")
		}

		for i := 0; i < lenOrgLM; i++ {
			lm1ID = append(lm1ID, orgLOs[i].ID.String)
			lm2ID = append(lm2ID, newLOs[i].ID.String)
			lm1Name = append(lm1Name, orgLOs[i].Name.String)
			lm2Name = append(lm2Name, newLOs[i].Name.String)
			lm1TopicIDs = append(lm1TopicIDs, orgLOs[i].TopicID.String)
			lm2TopicIDs = append(lm2TopicIDs, newLOs[i].TopicID.String)
		}
	case sspb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT.String():
		orgTaskAssignments, okOrg := orgLMs.([]entities.TaskAssignment)
		newTaskAssignments, okNew := newLMs.([]entities.TaskAssignment)
		if !okOrg || !okNew {
			return fmt.Errorf("cannot convert to task assignment")
		}

		for i := 0; i < lenOrgLM; i++ {
			lm1ID = append(lm1ID, orgTaskAssignments[i].ID.String)
			lm2ID = append(lm2ID, newTaskAssignments[i].ID.String)
			lm1Name = append(lm1Name, orgTaskAssignments[i].Name.String)
			lm2Name = append(lm2Name, newTaskAssignments[i].Name.String)
			lm1TopicIDs = append(lm1TopicIDs, orgTaskAssignments[i].TopicID.String)
			lm2TopicIDs = append(lm2TopicIDs, newTaskAssignments[i].TopicID.String)
		}
	}

	for i := 0; i < len(lm1ID); i++ {
		if golibs.InArrayString(lm1ID[i], lm2ID) || golibs.InArrayString(lm1TopicIDs[i], lm2TopicIDs) || !golibs.InArrayString(lm1Name[i], lm2Name) {
			return fmt.Errorf("wrong id, topic id or name learning material type return")
		}
	}
	return nil
}
