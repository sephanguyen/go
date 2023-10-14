package eureka

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	ys "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// nolint
var (
	COURSE_ID  = "course_id"
	STUDENT_ID = "student_id"
	BOOK_ID    = "book_id"
)

func (s *suite) prepareAssignment(ctx context.Context, topicID string, numberOfAssignments int) []*epb.Assignment {
	stepState := StepStateFromContext(ctx)
	pbAssignments := make([]*epb.Assignment, 0, numberOfAssignments)
	for i := 0; i < numberOfAssignments; i++ {
		assignmentID := idutil.ULIDNow()
		pbAssignments = append(pbAssignments, &epb.Assignment{
			AssignmentId: assignmentID,
			Name:         fmt.Sprintf("assignment-%s", assignmentID),
			Content: &epb.AssignmentContent{
				TopicId: topicID,
				LoId:    stepState.LoIDs,
			},
			CheckList: &epb.CheckList{
				Items: []*epb.CheckListItem{
					{
						Content:   "Complete all learning objectives",
						IsChecked: true,
					},
					{
						Content:   "Submitted required videos",
						IsChecked: false,
					},
				},
			},
			Instruction:    "teacher's instruction",
			MaxGrade:       10,
			Attachments:    []string{"media-id-1", "media-id-2"},
			AssignmentType: epb.AssignmentType_ASSIGNMENT_TYPE_LEARNING_OBJECTIVE,
			Setting: &epb.AssignmentSetting{
				AllowLateSubmission: true,
				AllowResubmission:   true,
			},
			RequiredGrade: true,
			DisplayOrder:  0,
		})
	}

	return pbAssignments
}

func (s *suite) prepareQuizzes(ctx context.Context, loID string, numberOfLOs int) []*epb.UpsertQuizRequest {
	stepState := StepStateFromContext(ctx)
	yQuizzes := make([]*epb.UpsertQuizRequest, 0)
	for i := 0; i < numberOfLOs; i++ {
		yQuizzes = append(yQuizzes, &epb.UpsertQuizRequest{
			Quiz: &epb.QuizCore{
				ExternalId: idutil.ULIDNow(),
				Kind:       cpb.QuizType_QUIZ_TYPE_FIB,
				SchoolId:   stepState.SchoolIDInt,
				Country:    cpb.Country_COUNTRY_VN,
				Question: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				Explanation: &cpb.RichText{
					Raw:      "raw",
					Rendered: "rendered " + idutil.ULIDNow(),
				},
				TaggedLos:       []string{"123", "abc"},
				DifficultyLevel: 2,
				Options: []*cpb.QuizOption{
					{
						Content: &cpb.RichText{
							Raw: `
							{
								"blocks": [
									{
										"key": "2lnf5",
										"text": "A",
										"type": "unstyled",
										"depth": 0,
										"inlineStyleRanges": [],
										"entityRanges": [],
										"data": {}
									}
								],
								"entityMap": {}
							}
						`,
							Rendered: "rendered " + idutil.ULIDNow(),
						},
						Correctness: true,
						Label:       "(1)",
						Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
						Key:         idutil.ULIDNow(),
					},
				},
			},
			LoId: loID,
		})
	}

	return yQuizzes
}

func (s *suite) prepareQuizzesEureka(ctx context.Context, loID string, numberOfLOs int) []*sspb.UpsertFlashcardContentRequest {
	stepState := StepStateFromContext(ctx)
	yQuizzes := make([]*sspb.UpsertFlashcardContentRequest, 0)
	for i := 0; i < numberOfLOs; i++ {
		yQuizzes = append(yQuizzes, &sspb.UpsertFlashcardContentRequest{
			Quizzes: []*cpb.QuizCore{
				{

					ExternalId: idutil.ULIDNow(),
					Kind:       cpb.QuizType_QUIZ_TYPE_FIB,
					Info: &cpb.ContentBasicInfo{
						SchoolId: stepState.SchoolIDInt,
						Country:  cpb.Country_COUNTRY_VN,
					},
					Question: &cpb.RichText{
						Raw:      "raw",
						Rendered: "rendered " + idutil.ULIDNow(),
					},
					Explanation: &cpb.RichText{
						Raw:      "raw",
						Rendered: "rendered " + idutil.ULIDNow(),
					},
					TaggedLos:       []string{"123", "abc"},
					DifficultyLevel: 2,
					Options: []*cpb.QuizOption{
						{
							Content: &cpb.RichText{
								Raw: `
									{
										"blocks": [
											{
												"key": "2lnf5",
												"text": "A",
												"type": "unstyled",
												"depth": 0,
												"inlineStyleRanges": [],
												"entityRanges": [],
												"data": {}
											}
										],
										"entityMap": {}
									}
								`,
								Rendered: "rendered " + idutil.ULIDNow(),
							},
							Attribute: &cpb.QuizItemAttribute{
								Configs: []cpb.QuizItemAttributeConfig{
									cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
								},
							},
							Correctness: true,
							Label:       "(1)",
							Configs:     []cpb.QuizOptionConfig{cpb.QuizOptionConfig_QUIZ_OPTION_CONFIG_CASE_SENSITIVE},
							Key:         idutil.ULIDNow(),
						},
					},
					Attribute: &cpb.QuizItemAttribute{
						Configs: []cpb.QuizItemAttributeConfig{
							cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG,
						},
					},
				},
			},

			FlashcardId: loID,
		})
	}

	return yQuizzes
}

func (s *suite) schoolAdminCreateContentBookForStudentProgress(ctx context.Context, numLos, numAsses, numTopics, numChapters, numQuzzes int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	_, pbChapters := s.prepareChapterV1(ctx, numChapters)
	if ctx, err := s.createChapterV1(ctx, stepState.BookID, pbChapters); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	for _, chapter := range pbChapters {
		_, pbTopics := s.prepareTopicInfoV1(ctx, chapter.Info.Id, numTopics)
		if _, err := s.createTopicV1(ctx, pbTopics); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, topic := range pbTopics {
			_, pbLOs := s.prepareLOV1(ctx, topic.Id, numLos, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_LEARNING)
			if _, err := s.createLOEureka(ctx, pbLOs); err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			for _, lo := range pbLOs {
				quizzes := s.prepareQuizzesEureka(ctx, lo.Info.Id, numQuzzes)
				for _, quiz := range quizzes {
					if _, err := sspb.NewQuizClient(s.Conn).UpsertFlashcardContent(ctx, quiz); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("unable create a quiz: %v", err)
					}
					stepState.QuizIDs = append(stepState.QuizIDs, quiz.Quizzes[0].ExternalId)
					time.Sleep(200 * time.Millisecond) // Work around no healthy upstream error
				}
			}

			pbAssignments := s.prepareAssignment(ctx, topic.Id, numAsses)
			if _, err := epb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(ctx, &epb.UpsertAssignmentsRequest{
				Assignments: pbAssignments,
			}); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable create a assignment: %v", err)
			}
			stepState.Assignments = pbAssignments
			time.Sleep(200 * time.Millisecond) // Work around no healthy upstream error
		}
		time.Sleep(200 * time.Millisecond) // Work around no healthy upstream error
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someOfCreatedAssignmentAreTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	assignments := stepState.Assignments
	if len(assignments) > 0 {
		assignments[0].AssignmentType = epb.AssignmentType_ASSIGNMENT_TYPE_TASK
		if _, err := epb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(s.signedCtx(ctx), &epb.UpsertAssignmentsRequest{
			Assignments: []*epb.Assignment{
				assignments[0],
			},
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("UpsertAssignments err %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentRetryDoQuizzes(ctx context.Context, numCorrectQuizzes int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SessionID = idutil.ULIDNow()

	request := &epb.CreateRetryQuizTestRequest{
		LoId:            stepState.LoID,
		StudentId:       stepState.StudentID,
		StudyPlanItemId: stepState.StudyPlanItemIDs[0],
		SetId:           wrapperspb.String(stepState.ShuffledQuizSetID),
		SessionId:       stepState.SessionID,
		Paging: &cpb.Paging{
			Limit: uint32(numCorrectQuizzes),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(stepState.NumQuizzes) - int64(len(stepState.WrongQuizExternalIDs)) + 1,
			},
		},
	}

	var resp *epb.CreateRetryQuizTestResponse
	resp, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).CreateRetryQuizTest(s.signedCtx(ctx), request)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	for i := 0; i < numCorrectQuizzes; i++ {
		if _, err := epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
			SetId:  resp.QuizzesId,
			QuizId: stepState.WrongQuizExternalIDs[i],
			Answer: []*epb.Answer{
				{Format: &epb.Answer_FilledText{FilledText: "A"}},
			},
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzes(ctx context.Context, role string, numLos, numAsses, numTopics, numChapters, numQuizzes int) (context.Context, error) {
	var err error
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	stepState.NumTopics = numTopics * numChapters
	stepState.NumChapter = numChapters
	stepState.NumQuizzes = numQuizzes

	stepState.BookID = idutil.ULIDNow()
	if _, err := epb.NewBookModifierServiceClient(s.Conn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: []*epb.UpsertBooksRequest_Book{
			{
				BookId: stepState.BookID,
				Name:   fmt.Sprintf("book-name+%s", stepState.BookID),
			},
		},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to create a book: %v", err)
	}

	if ctx, err = s.schoolAdminCreateContentBookForStudentProgress(ctx, numLos, numAsses, numTopics, numChapters, numQuizzes); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable create a content book: %v", err)
	}

	stepState.CourseID = idutil.ULIDNow()
	stepState.CourseIDs = []string{stepState.CourseID}

	if _, err := ys.NewCourseServiceClient(s.YasuoConn).UpsertCourses(ctx, &ys.UpsertCoursesRequest{
		Courses: []*ys.UpsertCoursesRequest_Course{
			{
				Id:           stepState.CourseID,
				Name:         fmt.Sprintf("course-name+%s", stepState.CourseID),
				Country:      bob_pb.COUNTRY_VN,
				Subject:      bob_pb.SUBJECT_MATHS,
				Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][int(stepState.Grade)],
				SchoolId:     constant.ManabieSchool,
				BookIds:      []string{stepState.BookID},
				DisplayOrder: 1,
			},
		},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create course: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) waitForImportStudentStudyPlanCompleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}

	err := try.Do(func(attempt int) (bool, error) {
		findByStudentIDsResp, err := studentStudyPlanRepo.FindByStudentIDs(ctx, s.DB, database.TextArray([]string{stepState.StudentID}))
		if err != nil {
			if err == pgx.ErrNoRows {
				time.Sleep(time.Second)
				return attempt < 10, fmt.Errorf("no row found")
			}
			return false, err
		}

		if len(findByStudentIDsResp) != 0 {
			return false, nil
		}

		time.Sleep(time.Second)
		return attempt < 10, fmt.Errorf("timeout sync import student study plan")
	})

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) calculateStudentProgress(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error

	stepState.AuthToken = stepState.StudentToken
	ctx = contextWithToken(s, ctx)

	stepState.Response, err = epb.NewStudyPlanReaderServiceClient(s.Conn).StudentBookStudyProgress(ctx, &epb.StudentBookStudyProgressRequest{
		CourseId:  stepState.CourseID,
		BookId:    stepState.BookID,
		StudentId: stepState.StudentID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch student book study progress: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(ctx context.Context, arg1 string, numWorkonLO, numCorrectQuiz, numWorkonAss, assPoint, skippedTopic string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lnumWorkonLO, _ := strconv.Atoi((numWorkonLO))
	lnumCorrectQuizInput, _ := strconv.Atoi((numCorrectQuiz))
	lnumWorkonAss, _ := strconv.Atoi((numWorkonAss))
	lSkippedTopic, _ := strconv.Atoi((skippedTopic))

	if ctx, err := s.waitForImportStudentStudyPlanCompleted(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch student study plan: %w", err)
	}
	studyPlanItemRepo := repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FetchByStudyProgressRequest(ctx, s.DB, database.Text(stepState.CourseID), database.Text(stepState.BookID), database.Text(stepState.StudentID))

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch study pan items: %w", err)
	}

	topicLOs := make(map[string]int)
	topicAss := make(map[string]int)
	chapters := make(map[string][]string)

	for _, each := range studyPlanItems {
		cs := new(entities.ContentStructure)
		each.ContentStructure.AssignTo(cs)
		chapters[cs.ChapterID] = append(chapters[cs.ChapterID], cs.TopicID)
	}
	stepState.SkippedTopics = nil
	for _, each := range chapters {
		for i := 0; i < lSkippedTopic; i++ {
			stepState.SkippedTopics = append(stepState.SkippedTopics, each[i])
		}
	}
	for _, each := range studyPlanItems {
		cs := new(entities.ContentStructure)
		each.ContentStructure.AssignTo(cs)
		stepState.StudyPlanItemIDs = append(stepState.StudyPlanItemIDs, each.ID.String)
		if containsStr(stepState.SkippedTopics, cs.TopicID) {
			continue
		}

		if cs.LoID != "" {
			if topicLOs[cs.TopicID] == lnumWorkonLO {
				continue
			}
			stepState.AuthToken = stepState.StudentToken
			ctx = contextWithToken(s, ctx)
			resp, err := epb.NewQuizModifierServiceClient(s.Conn).CreateQuizTest(s.signedCtx(ctx), &epb.CreateQuizTestRequest{
				LoId:            cs.LoID,
				StudentId:       stepState.StudentID,
				StudyPlanItemId: each.ID.String,
				KeepOrder:       false,
				Paging: &cpb.Paging{
					Limit: uint32(100),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(1),
					},
				},
			})
			stepState.ShuffledQuizSetID = resp.QuizzesId
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create quiz test: %w", err)
			}

			lNumCorrectQuiz := lnumCorrectQuizInput
			for _, quiz := range resp.Items {
				if lNumCorrectQuiz > 0 {
					if _, err := epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
						SetId:  resp.QuizzesId,
						QuizId: quiz.Core.ExternalId,
						Answer: []*epb.Answer{
							{Format: &epb.Answer_FilledText{FilledText: "A"}},
						},
					}); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
					}
				} else {
					if _, err := epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
						SetId:  resp.QuizzesId,
						QuizId: quiz.Core.ExternalId,
						Answer: []*epb.Answer{
							{Format: &epb.Answer_FilledText{FilledText: "B"}},
						},
					}); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
					}
					stepState.WrongQuizExternalIDs = append(stepState.WrongQuizExternalIDs, quiz.Core.ExternalId)
				}
				lNumCorrectQuiz--
			}
			topicLOs[cs.TopicID]++

			query := `
			UPDATE study_plan_items set completed_at = now() WHERE study_plan_item_id = $1::TEXT
		`
			if _, err := s.DB.Exec(ctx, query, each.ID.String); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable update study plan items: %v", err)
			}
		}

		if cs.AssignmentID != "" {
			if topicAss[cs.TopicID] == lnumWorkonAss {
				continue
			}
			stepState.AuthToken = stepState.StudentToken
			ctx = contextWithToken(s, ctx)
			submission := &epb.StudentSubmission{
				AssignmentId:    cs.AssignmentID,
				StudyPlanItemId: each.ID.String,
				StudentId:       stepState.StudentID,
				CourseId:        stepState.CourseID,
				Note:            "submit",
			}
			assignment, err := (&repositories.AssignmentRepo{}).RetrieveAssignments(ctx, s.DB, database.TextArray([]string{cs.AssignmentID}))

			if err != nil || len(assignment) == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve assignment: %w", err)
			}

			if assignment[0].Type.String == epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String() {
				submission.CorrectScore = cast.ToFloat32(assPoint)
				submission.TotalScore = 10
			}

			submitAssRes, err := epb.NewStudentAssignmentWriteServiceClient(s.Conn).SubmitAssignment(ctx, &epb.SubmitAssignmentRequest{
				Submission: submission,
			})
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
			}

			stepState.AuthToken = stepState.TeacherToken
			ctx = contextWithToken(s, ctx)
			if _, err := epb.NewStudentAssignmentWriteServiceClient(s.Conn).GradeStudentSubmission(ctx, &epb.GradeStudentSubmissionRequest{
				Grade: &epb.SubmissionGrade{
					Note:         "good job",
					SubmissionId: submitAssRes.SubmissionId,
					Grade:        cast.ToFloat64(assPoint),
				},
				Status: epb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
			}); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to grade assignment: %w", err)
			}
			topicAss[cs.TopicID]++
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointInFourTopics(ctx context.Context, arg1 string, numWorkonLO1, numCorrectQuiz1, numWorkonAss1, assPoint1, numWorkonLO2, numCorrectQuiz2, numWorkonAss2, assPoint2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if ctx, err := s.waitForImportStudentStudyPlanCompleted(ctx); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch student study plan: %w", err)
	}

	studyPlanItemRepo := repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FetchByStudyProgressRequest(ctx, s.DB, database.Text(stepState.CourseID), database.Text(stepState.BookID), database.Text(stepState.StudentID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch study pan items: %w", err)
	}

	topicLOs := make(map[string]int)
	topicAss := make(map[string]int)
	chapters := make(map[string][]string)
	lnumWorkonLO := 0
	lnumCorrectQuizInput := 0
	lnumWorkonAss := 0
	lassPoint := 0
	firstTopicPair := []string{stepState.TopicIDs[0], stepState.TopicIDs[1], stepState.TopicIDs[4], stepState.TopicIDs[5], stepState.TopicIDs[8], stepState.TopicIDs[9]}

	for _, each := range studyPlanItems {
		cs := new(entities.ContentStructure)
		each.ContentStructure.AssignTo(cs)
		chapters[cs.ChapterID] = append(chapters[cs.ChapterID], cs.TopicID)
	}

	for _, each := range studyPlanItems {
		cs := new(entities.ContentStructure)
		each.ContentStructure.AssignTo(cs)

		if containsStr(firstTopicPair, cs.TopicID) {
			lnumWorkonLO, _ = strconv.Atoi((numWorkonLO1))
			lnumCorrectQuizInput, _ = strconv.Atoi((numCorrectQuiz1))
			lnumWorkonAss, _ = strconv.Atoi((numWorkonAss1))
			lassPoint, _ = strconv.Atoi((assPoint1))
		} else {
			lnumWorkonLO, _ = strconv.Atoi((numWorkonLO2))
			lnumCorrectQuizInput, _ = strconv.Atoi((numCorrectQuiz2))
			lnumWorkonAss, _ = strconv.Atoi((numWorkonAss2))
			lassPoint, _ = strconv.Atoi((assPoint2))
		}

		if cs.LoID != "" {
			if topicLOs[cs.TopicID] == lnumWorkonLO {
				continue
			}
			stepState.AuthToken = stepState.StudentToken
			ctx = contextWithToken(s, ctx)
			resp, err := epb.NewQuizModifierServiceClient(s.Conn).CreateQuizTest(s.signedCtx(ctx), &epb.CreateQuizTestRequest{
				LoId:            cs.LoID,
				StudentId:       stepState.StudentID,
				StudyPlanItemId: each.ID.String,
				KeepOrder:       false,
				Paging: &cpb.Paging{
					Limit: uint32(100),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(1),
					},
				},
			})
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create quiz test: %w", err)
			}

			lNumCorrectQuiz := lnumCorrectQuizInput
			for _, quiz := range resp.Items {
				if lNumCorrectQuiz > 0 {
					if _, err := epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
						SetId:  resp.QuizzesId,
						QuizId: quiz.Core.ExternalId,
						Answer: []*epb.Answer{
							{Format: &epb.Answer_FilledText{FilledText: "A"}},
						},
					}); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
					}
				} else {
					if _, err := epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
						SetId:  resp.QuizzesId,
						QuizId: quiz.Core.ExternalId,
						Answer: []*epb.Answer{
							{Format: &epb.Answer_FilledText{FilledText: "B"}},
						},
					}); err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
					}
				}
				lNumCorrectQuiz--
			}
			topicLOs[cs.TopicID]++

			query := `
			UPDATE study_plan_items set completed_at = now() WHERE study_plan_item_id = $1::TEXT
		`
			if _, err := s.DB.Exec(ctx, query, each.ID.String); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable update study plan items: %v", err)
			}
		}

		if cs.AssignmentID != "" {
			if topicAss[cs.TopicID] == lnumWorkonAss {
				continue
			}
			stepState.AuthToken = stepState.StudentToken
			ctx = contextWithToken(s, ctx)
			submitAssRes, err := epb.NewStudentAssignmentWriteServiceClient(s.Conn).SubmitAssignment(ctx, &epb.SubmitAssignmentRequest{
				Submission: &epb.StudentSubmission{
					AssignmentId:    cs.AssignmentID,
					StudyPlanItemId: each.ID.String,
					StudentId:       stepState.StudentID,
					CourseId:        stepState.CourseID,
					Note:            "submit",
				},
			})
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
			}

			stepState.AuthToken = stepState.TeacherToken
			ctx = contextWithToken(s, ctx)
			if _, err := epb.NewStudentAssignmentWriteServiceClient(s.Conn).GradeStudentSubmission(ctx, &epb.GradeStudentSubmissionRequest{
				Grade: &epb.SubmissionGrade{
					Note:         "good job",
					SubmissionId: submitAssRes.SubmissionId,
					Grade:        float64(lassPoint),
				},
				Status: epb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
			}); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to grade assignment: %w", err)
			}
			topicAss[cs.TopicID]++
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) topicScoreIsAndChapterScoreIs(ctx context.Context, topicScore, chapterScore int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*epb.StudentBookStudyProgressResponse)

	if len(res.TopicProgress) != stepState.NumTopics || len(res.ChapterProgress) != stepState.NumChapter {
		return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for number of chapters and topics")
	}

	for _, each := range res.TopicProgress {
		if !containsStr(stepState.SkippedTopics, each.TopicId) && each.AverageScore != nil && each.AverageScore.Value != int32(topicScore) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for topic average got = %v, want = %v", each.AverageScore.Value, topicScore)
		}

		if !containsStr(stepState.SkippedTopics, each.TopicId) && each.AverageScore == nil && int32(topicScore) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for topic average got = %v, want = %v", 0, topicScore)
		}
	}

	for _, each := range res.ChapterProgress {
		if each.AverageScore != nil && each.AverageScore.Value != int32(chapterScore) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for chapter average got = %v, want = %v", each.AverageScore.Value, chapterScore)
		}

		if each.AverageScore == nil && int32(chapterScore) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for chapter average got = %v, want = %v", 0, chapterScore)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) firstPairTopicScoreIsAndSecondPairTopicScoreIsAndChapterScoreIs(ctx context.Context, firstPairTopicScore, secondPairTopicScore, chapterScore int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*epb.StudentBookStudyProgressResponse)
	firstPairTopicIds := []string{stepState.TopicIDs[0], stepState.TopicIDs[1], stepState.TopicIDs[4], stepState.TopicIDs[5], stepState.TopicIDs[8], stepState.TopicIDs[9]}
	secondPairTopicIds := []string{stepState.TopicIDs[2], stepState.TopicIDs[3], stepState.TopicIDs[6], stepState.TopicIDs[7], stepState.TopicIDs[10], stepState.TopicIDs[11]}

	if len(res.TopicProgress) != stepState.NumTopics || len(res.ChapterProgress) != stepState.NumChapter {
		return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for number of chapters and topics")
	}

	for _, each := range res.TopicProgress {
		if containsStr(firstPairTopicIds, each.TopicId) && each.AverageScore != nil && each.AverageScore.Value != int32(firstPairTopicScore) {
			fmt.Println("first topic score ", each.AverageScore.Value)
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for first topic pair average")
		}
		if containsStr(secondPairTopicIds, each.TopicId) && each.AverageScore != nil && each.AverageScore.Value != int32(secondPairTopicScore) {
			fmt.Println("second topic score ", each.AverageScore.Value)
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for second topic pair average")
		}
	}

	for _, each := range res.ChapterProgress {
		if each.AverageScore != nil && each.AverageScore.Value != int32(chapterScore) {
			fmt.Println("chapter score ", each.AverageScore.Value)
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect result for chapter average")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) calculateStudentProgressWithMissing(ctx context.Context, missingParam string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	req := &epb.StudentBookStudyProgressRequest{}

	switch missingParam {
	case COURSE_ID:
		req.BookId = BOOK_ID
		req.StudentId = STUDENT_ID
	case BOOK_ID:
		req.CourseId = COURSE_ID
		req.StudentId = STUDENT_ID
	case STUDENT_ID:
		req.BookId = BOOK_ID
		req.CourseId = COURSE_ID
	}

	_, stepState.ResponseErr = epb.NewStudyPlanReaderServiceClient(s.Conn).StudentBookStudyProgress(ctx, req)
	return StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *suite) correctLoCompletedWithAnd(ctx context.Context, doneLOs, doneAssignments string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lDoneLo, _ := strconv.Atoi((doneLOs))
	lDoneAssignments, _ := strconv.Atoi((doneAssignments))

	resp := stepState.Response.(*epb.StudentBookStudyProgressResponse)

	for _, each := range resp.TopicProgress {
		if each.CompletedStudyPlanItem.Value != (int32(lDoneAssignments) + int32(lDoneLo)) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("mismatch completed study plan items")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func containsStr(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}

func structEqual(lhs interface{}, rhs interface{}) bool {
	lhsRaw, _ := json.Marshal(lhs)
	rhsRaw, _ := json.Marshal(rhs)
	var lhsMap, rhsMap interface{}
	_ = json.Unmarshal(lhsRaw, &lhsMap)
	_ = json.Unmarshal(rhsRaw, &rhsMap)
	return reflect.DeepEqual(lhsMap, rhsMap)
}
