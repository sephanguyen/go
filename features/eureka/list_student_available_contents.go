package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

const (
	one  string = "one"
	many string = "many"
)

func (s *suite) dataForListStudentAvailableContentsWithBooks(ctx context.Context, numberOfBooksStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	timeNow := database.Timestamptz(now)
	timeNil := pgtype.Timestamptz{Status: pgtype.Null}
	textNil := pgtype.Text{Status: pgtype.Null}
	jsonbNil := pgtype.JSONB{Status: pgtype.Null}
	numNil := pgtype.Int4{Status: pgtype.Null}

	var numberOfBooks int32
	numberOfTopics := 5
	numberOfStudyPlanItems := 20

	switch numberOfBooksStr {
	case one:
		numberOfBooks = 1
		stepState.BookIDs = append(stepState.BookIDs, idutil.ULIDNow())
	case many:
		numberOfBooks = 2 + rand.Int31n(3)

		for i := 0; i < int(numberOfBooks); i++ {
			stepState.BookIDs = append(stepState.BookIDs, idutil.ULIDNow())
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid number_of_books arguments")
	}

	numberOfCourses := 2 + rand.Int31n(3)
	for courseIndex := 0; courseIndex < int(numberOfCourses); courseIndex++ {
		stepState.CourseIDs = append(stepState.CourseIDs, idutil.ULIDNow())
	}

	for bookIndex := 0; bookIndex < int(numberOfBooks); bookIndex++ {
		bookID := stepState.BookIDs[bookIndex]
		book := &entities.Book{
			ID:                         database.Text(bookID),
			Name:                       database.Text(fmt.Sprintf("Book-%s", stepState.BookIDs[bookIndex])),
			Country:                    database.Text("Viet Nam"),
			Subject:                    database.Text("Math"),
			Grade:                      database.Int2(10),
			SchoolID:                   database.Int4(constants.ManabieSchool),
			CreatedAt:                  timeNow,
			UpdatedAt:                  timeNow,
			DeletedAt:                  timeNil,
			CopiedFrom:                 textNil,
			CurrentChapterDisplayOrder: database.Int4(10),
			BookType:                   database.Text(cpb.BookType_BOOK_TYPE_GENERAL.String()),
		}
		if _, err := database.Insert(ctx, book, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a book: %v", err)
		}

		chapterID := idutil.ULIDNow()
		chapter := &entities.Chapter{
			ID:                       database.Text(chapterID),
			Name:                     database.Text(fmt.Sprintf("chapter-%s", chapterID)),
			Country:                  database.Text("Viet Nam"),
			Subject:                  database.Text("Math"),
			Grade:                    database.Int2(10),
			DisplayOrder:             database.Int2(int16(rand.Int31())),
			SchoolID:                 database.Int4(constants.ManabieSchool),
			CreatedAt:                timeNow,
			UpdatedAt:                timeNow,
			DeletedAt:                timeNil,
			CopiedFrom:               textNil,
			CurrentTopicDisplayOrder: database.Int4(10),
			BookID:                   database.Text(bookID),
		}

		if _, err := database.Insert(ctx, chapter, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a chapter: %v", err)
		}

		bookChapter := &entities.BookChapter{
			BookID:    database.Text(bookID),
			ChapterID: database.Text(chapterID),
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
			DeletedAt: timeNil,
		}
		if _, err := database.Insert(ctx, bookChapter, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a book_chapter")
		}

		studyPlanID := idutil.ULIDNow()
		courseID := idutil.ULIDNow()
		studyPlan := &entities.StudyPlan{
			BaseEntity: entities.BaseEntity{
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
				DeletedAt: timeNil,
			},
			ID:                  database.Text(studyPlanID),
			MasterStudyPlan:     textNil,
			Name:                database.Text(fmt.Sprintf("StudyPlan-%s", stepState.StudyPlanID)),
			StudyPlanType:       database.Text(epb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String()),
			SchoolID:            database.Int4(constants.ManabieSchool),
			CourseID:            database.Text(courseID),
			BookID:              database.Text(bookID),
			Status:              database.Text(epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE.String()),
			TrackSchoolProgress: database.Bool(true),
			Grades:              database.Int4Array([]int32{1, 2, 3}),
		}
		if _, err := database.Insert(ctx, studyPlan, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a study_plan: %v", err)
		}

		studentStudyPlan := &entities.StudentStudyPlan{
			StudentID: database.Text(stepState.StudentID),
			BaseEntity: entities.BaseEntity{
				CreatedAt: timeNow,
				UpdatedAt: timeNow,
				DeletedAt: timeNil,
			},
			StudyPlanID:       database.Text(studyPlanID),
			MasterStudyPlanID: textNil,
		}
		if _, err := database.Insert(ctx, studentStudyPlan, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a studnet_study_plan: %v", err)
		}

		for topicIndex := 0; topicIndex < numberOfTopics; topicIndex++ {
			topicID := idutil.ULIDNow()

			topic := &entities.Topic{
				ID:                    database.Text(topicID),
				Name:                  database.Text(fmt.Sprintf("name-%s", topicID)),
				Country:               database.Text("Viet Nam"),
				Subject:               database.Text("Math"),
				Grade:                 database.Int2(int16(rand.Int31n(100000))),
				DisplayOrder:          database.Int2(int16(rand.Int31n(100000))),
				Status:                database.Text("status"),
				TopicType:             database.Text("topic-type"),
				ChapterID:             database.Text(chapterID),
				SchoolID:              database.Int4(constant.ManabieSchool),
				IconURL:               database.Text("icon-url"),
				TotalLOs:              database.Int4(1),
				PublishedAt:           timeNow,
				AttachmentNames:       database.TextArray([]string{}),
				AttachmentURLs:        database.TextArray([]string{}),
				Instruction:           database.Text("instruction"),
				CopiedTopicID:         database.Text("copiedTopicID"),
				EssayRequired:         database.Bool(true),
				LODisplayOrderCounter: database.Int4(1),
				CreatedAt:             timeNow,
				UpdatedAt:             timeNow,
				DeletedAt:             timeNil,
			}
			if _, err := database.Insert(ctx, topic, s.DB.Exec); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a topic: %v", err)
			}

			for studyPlanItemIdx := 0; studyPlanItemIdx < numberOfStudyPlanItems; studyPlanItemIdx++ {
				var (
					contentStructureStr string
					assignmentID        string
					loID                string
				)
				if studyPlanItemIdx%2 == 0 {
					assignmentID = idutil.ULIDNow()
					assignment := &entities.Assignment{
						BaseEntity: entities.BaseEntity{
							CreatedAt: timeNow,
							UpdatedAt: timeNow,
							DeletedAt: timeNil,
						},
						ID: database.Text(assignmentID),
						Content: func() pgtype.JSONB {
							jsonb := pgtype.JSONB{}
							_ = jsonb.Set(map[string]string{
								"topic_id": topicID,
							})
							return jsonb
						}(),
						Attachment:      database.TextArray([]string{"a", "b"}),
						Settings:        jsonbNil,
						CheckList:       jsonbNil,
						Type:            database.Text(epb.AssignmentType_ASSIGNMENT_TYPE_LEARNING_OBJECTIVE.String()),
						Status:          database.Text(epb.AssignmentStatus_ASSIGNMENT_STATUS_ACTIVE.String()),
						MaxGrade:        database.Int4(10),
						Instruction:     textNil,
						Name:            database.Text(fmt.Sprintf("assignment-%s", assignmentID)),
						IsRequiredGrade: database.Bool(true),
						DisplayOrder:    database.Int4(5),
						OriginalTopic:   database.Text(topicID),
						TopicID:         database.Text(topicID),
					}
					if _, err := database.Insert(ctx, assignment, s.DB.Exec); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a assignment: %v", err)
					}

					topicAssignment := &entities.TopicsAssignments{
						TopicID:      database.Text(topicID),
						AssignmentID: database.Text(assignmentID),
						DisplayOrder: database.Int2(int16(rand.Int31n(100000))),
						CreatedAt:    timeNow,
						UpdatedAt:    timeNow,
						DeletedAt:    timeNil,
					}
					if _, err := database.Insert(ctx, topicAssignment, s.DB.Exec); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a topic_assignment: %v", err)
					}
				} else {
					loID = idutil.ULIDNow()
					learningObjective := &entities.LearningObjective{
						ID:             database.Text(loID),
						Name:           database.Text(fmt.Sprintf("lo-%s", loID)),
						Country:        database.Text("Viet Nam"),
						Grade:          database.Int2(10),
						Subject:        database.Text("Chemistry"),
						TopicID:        database.Text(topicID),
						MasterLoID:     textNil,
						DisplayOrder:   database.Int2(15),
						VideoScript:    textNil,
						Prerequisites:  database.TextArray([]string{"a", "b", "c"}),
						Video:          textNil,
						StudyGuide:     textNil,
						SchoolID:       database.Int4(constants.ManabieSchool),
						CreatedAt:      timeNow,
						UpdatedAt:      timeNow,
						DeletedAt:      timeNil,
						Type:           textNil,
						Instruction:    textNil,
						GradeToPass:    numNil,
						ManualGrading:  database.Bool(false),
						TimeLimit:      numNil,
						MaximumAttempt: numNil,
						ApproveGrading: database.Bool(false),
						GradeCapping:   database.Bool(false),
						ReviewOption:   database.Text(cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()),
						VendorType:     database.Text(cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String()),
					}
					if _, err := database.Insert(ctx, learningObjective, s.DB.Exec); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a learning_objective: %v", err)
					}

					topicLO := &entities.TopicsLearningObjectives{
						TopicID:      database.Text(topicID),
						LoID:         database.Text(loID),
						DisplayOrder: database.Int2(int16(rand.Int31n(100000))),
						CreatedAt:    timeNow,
						UpdatedAt:    timeNow,
						DeletedAt:    timeNil,
					}
					if _, err := database.Insert(ctx, topicLO, s.DB.Exec); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a topic_learning_objective: %v", err)
					}
				}
				for courseIndex := 0; courseIndex < int(numberOfCourses); courseIndex++ {
					courseID := stepState.CourseIDs[courseIndex]
					if studyPlanItemIdx%2 == 0 {
						contentStructureStr = fmt.Sprintf(`{"book_id": "%s", "course_id": "%s", "topic_id": "%s", "chapter_id": "%s", "assignment_id": "%s"}`, bookID, courseID, topicID, chapterID, assignmentID)
					} else {
						contentStructureStr = fmt.Sprintf(`{"book_id": "%s", "course_id": "%s", "topic_id": "%s", "chapter_id": "%s", "lo_id": "%s"}`, bookID, courseID, topicID, chapterID, loID)
					}

					studyPlanItemID := idutil.ULIDNow()
					studyPlanItem := &entities.StudyPlanItem{
						ID:               database.Text(studyPlanItemID),
						StudyPlanID:      database.Text(studyPlanID),
						ContentStructure: database.JSONB(string(contentStructureStr)),
						BaseEntity: entities.BaseEntity{
							CreatedAt: timeNow,
							UpdatedAt: timeNow,
							DeletedAt: timeNil,
						},
						Status:                  database.Text(epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String()),
						ContentStructureFlatten: database.Text(fmt.Sprintf("content_structure_flatter-%s", idutil.ULIDNow())),
						AvailableFrom:           pgtype.Timestamptz{Time: now.Add(-10 * 24 * time.Hour), Status: pgtype.Present},
						AvailableTo:             pgtype.Timestamptz{Time: now.Add(10 * 24 * time.Hour), Status: pgtype.Present},
						StartDate:               timeNil,
						EndDate:                 timeNil,
						CompletedAt:             timeNil,
						DisplayOrder:            database.Int4(rand.Int31n(100)),
						CopyStudyPlanItemID:     textNil,
						SchoolDate:              timeNil,
					}

					if _, err := database.Insert(ctx, studyPlanItem, s.DB.Exec); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a study_plan_item: %v", err)
					}

					if studyPlanItemIdx%2 == 0 {
						assignmentStudyPlanItem := &entities.AssignmentStudyPlanItem{
							BaseEntity: entities.BaseEntity{
								CreatedAt: timeNow,
								UpdatedAt: timeNow,
								DeletedAt: timeNil,
							},
							StudyPlanItemID: database.Text(studyPlanItemID),
							AssignmentID:    database.Text(assignmentID),
						}
						if _, err := database.Insert(ctx, assignmentStudyPlanItem, s.DB.Exec); err != nil {
							return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a assignment_study_plan_item: %v", err)
						}
					} else {
						loStudyPlanItem := &entities.LoStudyPlanItem{
							BaseEntity: entities.BaseEntity{
								CreatedAt: timeNow,
								UpdatedAt: timeNow,
								DeletedAt: timeNil,
							},
							StudyPlanItemID: database.Text(studyPlanItemID),
							LoID:            database.Text(loID),
						}
						if _, err := database.Insert(ctx, loStudyPlanItem, s.DB.Exec); err != nil {
							return StepStateToContext(ctx, stepState), fmt.Errorf("failed when creating a lo_study_plan_item: %v", err)
						}
					}
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) listStudentAvailableContents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = epb.NewAssignmentReaderServiceClient(s.Conn).
		ListStudentAvailableContents(s.signedCtx(ctx), &epb.ListStudentAvailableContentsRequest{
			StudyPlanId: stepState.StudyPlanIDs,
		})
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) verifyListContentsAfterListStudentAvailableContentsWithBooks(ctx context.Context, numberOfBooksStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.ListStudentAvailableContentsResponse)

	// map topicID and StudyPlanItems
	m := make(map[string][]*epb.StudyPlanItem)
	for _, content := range resp.Contents {
		studyPlanItem := content.StudyPlanItem
		m[studyPlanItem.ContentStructure.TopicId] = append(m[studyPlanItem.ContentStructure.TopicId], studyPlanItem)
	}

	for topicID, studyPlanItems := range m {
		for i := 1; i < len(studyPlanItems); i++ {
			if studyPlanItems[i-1].DisplayOrder > studyPlanItems[i].DisplayOrder {
				return StepStateToContext(ctx, stepState), fmt.Errorf("response is out of order (display_order of study_plan_item) with topic_id (%s)", topicID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
