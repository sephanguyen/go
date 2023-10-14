package course_statistical

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	ys_pb_v1 "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userCreateABook(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = epb.NewBookModifierServiceClient(s.EurekaConn).UpsertBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertBooksRequest{
		Books: []*epb.UpsertBooksRequest_Book{
			{
				BookId: idutil.ULIDNow(),
				Name:   fmt.Sprintf("book-name+%s", stepState.BookID),
			},
		},
	})
	return utils.StepStateToContext(ctx, stepState), nil
}

// Create mock student login
func (s *Suite) studentLogin(ctx context.Context, numStudent int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for i := 0; i < numStudent; i++ {
		studentID, StudentToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, "student")
		stepState.StudentIDs = append(stepState.StudentIDs, studentID)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		stepState.Students = append(
			stepState.Students,
			entity.Student{
				ID:    studentID,
				Token: StudentToken,
			},
		)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// Mock teacher login
func (s *Suite) teacherLogin(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	_, token, err := s.AuthHelper.AUserSignedInAsRole(ctx, "teacher")
	if err != nil {
		return nil, fmt.Errorf("error teacher login failed: %w", err)
	}
	stepState.TeacherToken = token

	return utils.StepStateToContext(ctx, stepState), nil
}

// Mock school admin login
func (s *Suite) schoolAdminLogin(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	_, token, err := s.AuthHelper.AUserSignedInAsRole(ctx, "school admin")
	if err != nil {
		return nil, fmt.Errorf("error school admin login failed: %w", err)
	}
	stepState.SchoolAdminToken = token

	return utils.StepStateToContext(ctx, stepState), nil
}

// This step create mock data on the course
// base on number of provided by args
func (s *Suite) hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzesV2(ctx context.Context, role string, numLos, numAsses, numTopics, numChapters, numQuizzes int) (context.Context, error) {
	var err error
	stepState := utils.StepStateFromContext[StepState](ctx)
	// add school_admin token to context
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	// Create 1 book
	bookGenerated, err := utils.GenerateBooksV2(ctx, 1, nil, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("school admin unable to create a book: %v", err)
	}
	stepState.BookID = bookGenerated.BookIDs[0]

	// Create content book
	if ctx, err = s.hasCreateContentBookForStudentProgress(ctx, numLos, numAsses, numTopics, numChapters, numQuizzes); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create a content book: %v", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) schoolAdminCreateACourseWithABook(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Create course
	// Assume Course has create successfully in yasuo service
	// and return courseID
	stepState.CourseID = idutil.ULIDNow()

	// Add book to couse
	if _, err := epb.NewCourseModifierServiceClient(s.EurekaConn).AddBooks(ctx, &epb.AddBooksRequest{
		CourseId: stepState.CourseID,
		BookIds:  []string{stepState.BookID},
	}); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to add book to course: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) hasCreateContentBookForStudentProgress(ctx context.Context, numLos, numAsses, numTopics, numChapters, numQuizzes int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	// Create Book chapters
	// Use school admin token for auth
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	chaptersGenerated, err := utils.GenerateChaptersV2(ctx, stepState.BookID, numChapters, nil, s.EurekaConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	for _, chapterID := range chaptersGenerated.ChapterIDs {
		// Create topic
		topicsGenerated, err := utils.GenerateTopicsV2(ctx, chapterID, numTopics, nil, s.EurekaConn)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		for _, topicID := range topicsGenerated.TopicIDs {
			// Create learning objective
			losGenerated, err := utils.GenerateLearningObjectivesV2(ctx, topicID, numLos, cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_LEARNING, nil, s.EurekaConn)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
			stepState.LoIDs = losGenerated.LoIDs

			for _, loID := range losGenerated.LoIDs {
				// Create quiz
				// Use school admin token for auth
				stepState.Token = stepState.SchoolAdminToken
				ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

				if err := utils.GenerateQuizzes(ctx, loID, numQuizzes, nil, s.EurekaConn); err != nil {
					return utils.StepStateToContext(ctx, stepState), err
				}
			}

			// Create assignment
			assignmentGenerated, err := utils.GenerateAssignment(ctx, topicID, numAsses, losGenerated.LoIDs, s.EurekaConn, nil)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
			stepState.AssIDs = assignmentGenerated.AssignmentIDs
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsCorrectTopicStatistic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := stepState.Request.(*epb.RetrieveCourseStatisticRequestV2)
	statisticResp := stepState.Response.(*epb.RetrieveCourseStatisticResponseV2)
	classID := pgtype.TextArray{Status: pgtype.Null}
	if len(req.ClassId) != 0 {
		classID = database.TextArray(req.ClassId)
	}

	items, err := (&repositories.CourseStudyPlanRepo{}).ListCourseStatisticItemsV2(ctx, s.EurekaDB, &repositories.ListCourseStatisticItemsArgsV2{
		CourseID:    database.Text(req.CourseId),
		StudyPlanID: database.Text(req.StudyPlanId),
		ClassID:     classID,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error repo list topic statistic items %w", err)
	}

	// order of topic, study plan is correct
	topicIDs := []string{}
	studyPlanItemIDs := []string{}
	for _, item := range items {
		topicIDs = append(topicIDs, item.ContentStructure.TopicID)
		studyPlanItemIDs = append(studyPlanItemIDs, item.RootStudyPlanItemID)
	}
	topicIDs = golibs.GetUniqueElementStringArray(topicIDs)
	studyPlanItemIDs = golibs.GetUniqueElementStringArray(studyPlanItemIDs)
	stepState.TopicIDs = topicIDs

	if len(topicIDs) != len(statisticResp.GetTopicStatistic()) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expect %v topics got %v, courseID: %s, masterStudyPlamitemID: %s", len(topicIDs), len(statisticResp.GetTopicStatistic()), stepState.CourseID, stepState.StudyPlanID)
	}

	sort.Strings(stepState.ArchivedStudyPlanItemIDs)

	spiIndex := 0
	for i, statItem := range statisticResp.GetTopicStatistic() {
		if statItem.TopicId != topicIDs[i] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong topic order")
		}

		for _, studyPlanItem := range statItem.GetLearningMaterialStatistic() {
			if spiIndex >= len(studyPlanItemIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("more study plan items return than expected")
			}
			if sort.SearchStrings(stepState.ArchivedStudyPlanItemIDs, studyPlanItem.GetStudyPlanItemId()) != len(stepState.ArchivedStudyPlanItemIDs) {
				// archived item
				if studyPlanItem.GetTotalAssignedStudent() != 0 {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("total assigned student mustn't count archived item")
				}
				if studyPlanItem.GetCompletedStudent() != 0 {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("completed student mustn't count archived item")
				}
				if studyPlanItem.GetAverageScore() != 0 {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("average score mustn't include archived item")
				}
			}
			spiIndex++
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateCouseDurationForStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	for _, student := range stepState.Students {
		query := `SELECT email FROM users WHERE user_id = $1`
		var studentEmail string
		err := s.BobDB.QueryRow(ctx, query, student.ID).Scan(&studentEmail)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}

		_, err = ys_pb_v1.NewUserModifierServiceClient(s.YasuoConn).UpdateStudent(
			ctx,
			&ys_pb_v1.UpdateStudentRequest{
				StudentProfile: &ys_pb_v1.UpdateStudentRequest_StudentProfile{
					Id:               student.ID,
					Name:             "test-name",
					Grade:            5,
					EnrollmentStatus: ys_pb_v1.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
					Email:            studentEmail,
				},
				SchoolId: stepState.SchoolIDInt,
			},
		)
		_, err = ys_pb_v1.NewUserModifierServiceClient(s.UserMgmtConn).UpsertStudentCoursePackage(ctx, &ys_pb_v1.UpsertStudentCoursePackageRequest{
			StudentPackageProfiles: []*ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile{{
				Id: &ys_pb_v1.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
					CourseId: stepState.CourseID,
				},
				StartTime: timestamppb.New(time.Now().Add(time.Hour * -20)),
				EndTime:   timestamppb.New(time.Now().Add(time.Hour * 24 * 10)),
			}},
			StudentId: student.ID,
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to update student course duration: %w", err)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// Create mock studyplan data
func (s *Suite) hasCreatedAStudyplanForStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use School Admin token
	stepState.Token = stepState.SchoolAdminToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	// This studyPlanID is also a masterStudyPlanID
	generatedStudyPlan, err := utils.GenerateStudyPlanV2(ctx, s.EurekaConn, stepState.CourseID, stepState.BookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanID = generatedStudyPlan.StudyPlanID

	// Find master study plan Items
	masterStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	// Find child study plan item
	for _, masterStudyPlanItem := range masterStudyPlanItems {
		childStudyPlanItems, err := (&repositories.StudyPlanItemRepo{}).RetrieveChildStudyPlanItem(ctx, s.EurekaDB, masterStudyPlanItem.ID, database.TextArray(stepState.StudentIDs))
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error retrieve child study plan item")
		}
		stepState.StudyPlanItems = append(stepState.StudyPlanItems, masterStudyPlanItem)
		stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, masterStudyPlanItem.ID.String)
		for _, childStudyPlanItem := range childStudyPlanItems {
			stepState.StudyPlanItems = append(stepState.StudyPlanItems, childStudyPlanItem)
			stepState.StudyPlanItemsIDs = append(stepState.StudyPlanItemsIDs, childStudyPlanItem.ID.String)
		}
	}

	upsertSpiReq := &epb.UpsertStudyPlanItemV2Request{}
	for i := 0; i < len(stepState.LoIDs); i++ {
		for _, item := range stepState.StudyPlanItems {
			cse := &entities.ContentStructure{}
			err := item.ContentStructure.AssignTo(cse)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
			}

			cs := &epb.ContentStructure{}
			err = item.ContentStructure.AssignTo(cs)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error unmarshal ContentStructure: %w", err)
			}

			if len(cse.LoID) != 0 {
				cs.ItemId = &epb.ContentStructure_LoId{LoId: wrapperspb.String(cse.LoID)}
			} else if len(cse.AssignmentID) != 0 {
				stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, cse.AssignmentID)
				cs.ItemId = &epb.ContentStructure_AssignmentId{AssignmentId: wrapperspb.String(cse.AssignmentID)}
			}

			upsertSpiReq.StudyPlanItems = append(upsertSpiReq.StudyPlanItems, &epb.StudyPlanItem{
				StudyPlanId:             item.StudyPlanID.String,
				StudyPlanItemId:         item.ID.String,
				AvailableFrom:           timestamppb.New(time.Now().Add(-24 * time.Hour)),
				AvailableTo:             timestamppb.New(time.Now().AddDate(0, 0, 10)),
				StartDate:               timestamppb.New(time.Now().Add(-23 * time.Hour)),
				EndDate:                 timestamppb.New(time.Now().AddDate(0, 0, 1)),
				ContentStructure:        cs,
				ContentStructureFlatten: item.ContentStructureFlatten.String,
				Status:                  epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
			})
			_, err = epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlanItemV2(ctx, upsertSpiReq)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan item: %w", err)
			}
			upsertSpiReq = &epb.UpsertStudyPlanItemV2Request{}
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
func (s *Suite) someStudentDoTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(ctx context.Context, numStudent int, arg1 int, arg2 int, arg3 int, arg4 int, arg5 int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if len(stepState.Students) < numStudent {
		stepState.Students = stepState.Students[0:numStudent]
	}
	for i := 0; i < numStudent; i++ {
		stepState.StudentID = stepState.Students[i].ID
		stepState.StudentToken = stepState.Students[i].Token
		if ctx, err := s.doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(utils.StepStateToContext(ctx, stepState), arg1, arg2, arg3, arg4, arg5); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) someStudentDoTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopicsV2(ctx context.Context, numStudent int, arg1 int, arg2 int, arg3 int, arg4 int, arg5 int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if len(stepState.Students) < numStudent {
		stepState.Students = stepState.Students[0:numStudent]
	}
	for i := 0; i < numStudent; i++ {
		stepState.StudentID = stepState.Students[i].ID
		stepState.StudentToken = stepState.Students[i].Token
		if ctx, err := s.doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopicsV2(utils.StepStateToContext(ctx, stepState), arg1, arg2, arg3, arg4, arg5); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopicsV2(ctx context.Context, numWorkDoneLO int, numCorrectQuiz int, numWorkDoneAss int, assPoint int, skippedTopic int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if ctx, err := s.waitForImportStudentStudyPlanCompleted(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch student study plan: %w", err)
	}

	studyPlanItemRepo := repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FetchByStudyProgressRequest(ctx, s.EurekaDB, database.Text(stepState.CourseID), database.Text(stepState.BookID), database.Text(stepState.StudentID))
	stepState.debug = len(studyPlanItems)

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch study plan items: %w", err)
	}

	chapters := make(map[string][]string)

	for _, each := range studyPlanItems {
		cs := new(entities.ContentStructure)
		if err := each.ContentStructure.AssignTo(cs); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error assignto ContentStructure for chapter id")
		}
		chapters[cs.ChapterID] = append(chapters[cs.ChapterID], cs.TopicID)
	}

	for _, each := range chapters {
		for i := 0; i < skippedTopic; i++ {
			stepState.SkippedTopics = append(stepState.SkippedTopics, each[i])
		}
	}

	for i := 0; i < numWorkDoneLO; i++ {
		loID := stepState.LoIDs[i]
		for _, each := range studyPlanItems {
			cs := new(entities.ContentStructure)
			if err := each.ContentStructure.AssignTo(cs); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error assignto ContentStructure in studyplan")
			}
			if utils.ContainsStr(stepState.SkippedTopics, cs.TopicID) {
				continue
			}

			if cs.LoID == loID {
				// Use student token for auth
				stepState.Token = stepState.StudentToken
				ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
				resp, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CreateQuizTest(ctx, &epb.CreateQuizTestRequest{
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
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create quiz test: %w", err)
				}
				stepState.ShuffledQuizSetID = resp.QuizzesId

				lNumCorrectQuiz := numCorrectQuiz
				for _, quiz := range resp.Items {
					if lNumCorrectQuiz > 0 {
						if _, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
							SetId:  resp.QuizzesId,
							QuizId: quiz.Core.ExternalId,
							Answer: []*epb.Answer{
								{Format: &epb.Answer_FilledText{FilledText: "A"}},
							},
						}); err != nil {
							return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
						}
					} else {
						if _, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
							SetId:  resp.QuizzesId,
							QuizId: quiz.Core.ExternalId,
							Answer: []*epb.Answer{
								{Format: &epb.Answer_FilledText{FilledText: "B"}},
							},
						}); err != nil {
							return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
						}
						stepState.WrongQuizExternalIDs = append(stepState.WrongQuizExternalIDs, quiz.Core.ExternalId)
					}
					lNumCorrectQuiz--
				}

				// For new refactor
				studentLog := []*epb.StudentEventLog{
					{
						EventId:   idutil.ULIDNow(),
						EventType: "quiz_answer_selected",
						Payload: &epb.StudentEventLogPayload{
							StudyPlanItemId: each.ID.String,
							LoId:            loID,
							Event:           "completed",
						},
						CreatedAt: timestamppb.Now(),
					},
				}
				if err := utils.GenerateStudentEventLogs(ctx, studentLog, s.EurekaConn); err != nil {
					return utils.StepStateToContext(ctx, stepState), err
				}
			}
		}
		for i := 0; i < numWorkDoneAss; i++ {
			assID := stepState.AssIDs[i]
			for _, each := range studyPlanItems {
				cs := new(entities.ContentStructure)
				if err := each.ContentStructure.AssignTo(cs); err != nil {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error assignto ContentStructure in studyplan")
				}
				if utils.ContainsStr(stepState.SkippedTopics, cs.TopicID) {
					continue
				}

				if cs.AssignmentID == assID {
					// Use student token for auth
					stepState.Token = stepState.StudentToken
					ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
					submission := &epb.StudentSubmission{
						AssignmentId:    cs.AssignmentID,
						StudyPlanItemId: each.ID.String,
						StudentId:       stepState.StudentID,
						CourseId:        stepState.CourseID,
						Note:            "submit",
					}
					assignment, err := (&repositories.AssignmentRepo{}).RetrieveAssignments(ctx, s.EurekaDB, database.TextArray([]string{cs.AssignmentID}))

					if err != nil || len(assignment) == 0 {
						return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve assignment: %w", err)
					}

					// if assignment[0].Type.String == epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String() {
					submission.CorrectScore = cast.ToFloat32(assPoint)
					submission.TotalScore = 10
					submission.CompleteDate = timestamppb.Now()

					submitAssRes, err := epb.NewStudentAssignmentWriteServiceClient(s.EurekaConn).SubmitAssignment(ctx, &epb.SubmitAssignmentRequest{
						Submission: submission,
					})
					if err != nil {
						return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
					}

					// Use teacher token
					stepState.Token = stepState.TeacherToken
					ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

					if _, err := epb.NewStudentAssignmentWriteServiceClient(s.EurekaConn).GradeStudentSubmission(ctx, &epb.GradeStudentSubmissionRequest{
						Grade: &epb.SubmissionGrade{
							Note:         "good job",
							SubmissionId: submitAssRes.SubmissionId,
							Grade:        cast.ToFloat64(assPoint),
						},
						Status: epb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
					}); err != nil {
						return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to grade assignment: %w", err)
					}

				}
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(ctx context.Context, numWorkDoneLO int, numCorrectQuiz int, numWorkDoneAss int, assPoint int, skippedTopic int) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if ctx, err := s.waitForImportStudentStudyPlanCompleted(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch student study plan: %w", err)
	}

	studyPlanItemRepo := repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FetchByStudyProgressRequest(ctx, s.EurekaDB, database.Text(stepState.CourseID), database.Text(stepState.BookID), database.Text(stepState.StudentID))
	stepState.debug = len(studyPlanItems)

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch study plan items: %w", err)
	}

	chapters := make(map[string][]string)

	for _, each := range studyPlanItems {
		cs := new(entities.ContentStructure)
		if err := each.ContentStructure.AssignTo(cs); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error assignto ContentStructure for chapter id")
		}
		chapters[cs.ChapterID] = append(chapters[cs.ChapterID], cs.TopicID)
	}

	for _, each := range chapters {
		for i := 0; i < skippedTopic; i++ {
			stepState.SkippedTopics = append(stepState.SkippedTopics, each[i])
		}
	}

	for i := 0; i < numWorkDoneLO; i++ {
		loID := stepState.LoIDs[i]
		for _, each := range studyPlanItems {
			cs := new(entities.ContentStructure)
			if err := each.ContentStructure.AssignTo(cs); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error assignto ContentStructure in studyplan")
			}
			if utils.ContainsStr(stepState.SkippedTopics, cs.TopicID) {
				continue
			}

			if cs.LoID == loID {
				// Use student token for auth
				stepState.Token = stepState.StudentToken
				ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
				resp, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CreateQuizTest(ctx, &epb.CreateQuizTestRequest{
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
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create quiz test: %w", err)
				}
				stepState.ShuffledQuizSetID = resp.QuizzesId

				lNumCorrectQuiz := numCorrectQuiz
				for _, quiz := range resp.Items {
					if lNumCorrectQuiz > 0 {
						if _, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
							SetId:  resp.QuizzesId,
							QuizId: quiz.Core.ExternalId,
							Answer: []*epb.Answer{
								{Format: &epb.Answer_FilledText{FilledText: "A"}},
							},
						}); err != nil {
							return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
						}
					} else {
						if _, err := epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(ctx, &epb.CheckQuizCorrectnessRequest{
							SetId:  resp.QuizzesId,
							QuizId: quiz.Core.ExternalId,
							Answer: []*epb.Answer{
								{Format: &epb.Answer_FilledText{FilledText: "B"}},
							},
						}); err != nil {
							return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to check quiz correctness: %w", err)
						}
						stepState.WrongQuizExternalIDs = append(stepState.WrongQuizExternalIDs, quiz.Core.ExternalId)
					}
					lNumCorrectQuiz--
				}

				query := `UPDATE study_plan_items set completed_at = now() WHERE study_plan_item_id = $1::TEXT`
				if _, err := s.EurekaDB.Exec(ctx, query, each.ID.String); err != nil {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable update complete study plan items: %v", err)
				}
			}
		}
		for i := 0; i < numWorkDoneAss; i++ {
			assID := stepState.AssIDs[i]
			for _, each := range studyPlanItems {
				cs := new(entities.ContentStructure)
				if err := each.ContentStructure.AssignTo(cs); err != nil {
					return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error assignto ContentStructure in studyplan")
				}
				if utils.ContainsStr(stepState.SkippedTopics, cs.TopicID) {
					continue
				}

				if cs.AssignmentID == assID {
					// Use student token for auth
					stepState.Token = stepState.StudentToken
					ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)
					submission := &epb.StudentSubmission{
						AssignmentId:    cs.AssignmentID,
						StudyPlanItemId: each.ID.String,
						StudentId:       stepState.StudentID,
						CourseId:        stepState.CourseID,
						Note:            "submit",
					}
					assignment, err := (&repositories.AssignmentRepo{}).RetrieveAssignments(ctx, s.EurekaDB, database.TextArray([]string{cs.AssignmentID}))

					if err != nil || len(assignment) == 0 {
						return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve assignment: %w", err)
					}

					if assignment[0].Type.String == epb.AssignmentType_ASSIGNMENT_TYPE_TASK.String() {
						submission.CorrectScore = cast.ToFloat32(assPoint)
						submission.TotalScore = 10
					}

					submitAssRes, err := epb.NewStudentAssignmentWriteServiceClient(s.EurekaConn).SubmitAssignment(ctx, &epb.SubmitAssignmentRequest{
						Submission: submission,
					})
					if err != nil {
						return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
					}

					// Use teacher token
					stepState.Token = stepState.TeacherToken
					ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

					if _, err := epb.NewStudentAssignmentWriteServiceClient(s.EurekaConn).GradeStudentSubmission(ctx, &epb.GradeStudentSubmissionRequest{
						Grade: &epb.SubmissionGrade{
							Note:         "good job",
							SubmissionId: submitAssRes.SubmissionId,
							Grade:        cast.ToFloat64(assPoint),
						},
						Status: epb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
					}); err != nil {
						return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to grade assignment: %w", err)
					}
				}
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) waitForImportStudentStudyPlanCompleted(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studentStudyPlanRepo := &repositories.StudentStudyPlanRepo{}

	err := try.Do(func(attempt int) (bool, error) {
		findByStudentIDsResp, err := studentStudyPlanRepo.FindByStudentIDs(ctx, s.EurekaDB, database.TextArray([]string{stepState.StudentID}))
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
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("waitForImportStudentStudyPlanCompleted error: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

const differentClass = "different"
const sameClass = "same"

// assign student to the same class or differentclass
func (s *Suite) AssignStudentsToClass(ctx context.Context, numStudent int, option string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.ClassID = idutil.ULIDNow()

	if option == sameClass {
		stepState.ClassIDs = append(stepState.ClassIDs, stepState.ClassID)
	}
	for i := 0; i < numStudent; i++ {
		if option == differentClass {
			stepState.ClassID = idutil.ULIDNow()
			stepState.ClassIDs = append(stepState.ClassIDs, stepState.ClassID)
		}
		studentID := stepState.StudentIDs[i]
		err := (&repositories.CourseClassRepo{}).BulkUpsert(ctx, s.EurekaDB, []*entities.CourseClass{
			{
				BaseEntity: entities.BaseEntity{
					CreatedAt: database.Timestamptz(time.Now()),
					UpdatedAt: database.Timestamptz(time.Now()),
					DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
				},
				ID:       database.Text(idutil.ULIDNow()),
				CourseID: database.Text(stepState.CourseID),
				ClassID:  database.Text(stepState.ClassID),
			},
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("CourseClassRepo.BulkUpsert: %w", err)
		}

		err = (&repositories.ClassStudentRepo{}).Upsert(ctx, s.EurekaDB, &entities.ClassStudent{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(time.Now()),
				UpdatedAt: database.Timestamptz(time.Now()),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			StudentID: database.Text(studentID),
			ClassID:   database.Text(stepState.ClassID),
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("ClassStudentRepo.Upsert: %w", err)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// Mock teacher retrieve topic statistical with no class filter
func (s *Suite) retrieveCourseStatisticWithNoClassFilter(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use teacher token for auth
	stepState.Token = stepState.TeacherToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	req := &epb.RetrieveCourseStatisticRequestV2{
		CourseId:    stepState.CourseID,
		StudyPlanId: stepState.StudyPlanID,
		ClassId:     []string{},
	}

	if len(stepState.ClassIDs) != 0 {
		req.ClassId = stepState.ClassIDs
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = epb.NewCourseReaderServiceClient(s.EurekaConn).RetrieveCourseStatisticV2(ctx, req)

	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error RetrieveCourseStatisticV2: %w", stepState.ResponseErr)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// Mock teacher retrieve topic statistical with class filter
func (s *Suite) retrieveCourseStatisticWithClassFilter(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use teacher token for auth
	stepState.Token = stepState.TeacherToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	req := &epb.RetrieveCourseStatisticRequestV2{
		CourseId:    stepState.CourseID,
		StudyPlanId: stepState.StudyPlanID,
		ClassId:     stepState.ClassIDs,
	}

	if len(stepState.ClassIDs) != 0 {
		req.ClassId = stepState.ClassIDs
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = epb.NewCourseReaderServiceClient(s.EurekaConn).RetrieveCourseStatisticV2(ctx, req)

	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error RetrieveCourseStatisticV2: %w", stepState.ResponseErr)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIs(ctx context.Context, topicAssigned, topicCompleted, topicAverageScore int32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	statisticResp := stepState.Response.(*epb.RetrieveCourseStatisticResponseV2)
	if len(statisticResp.TopicStatistic) == 0 {
		if topicAssigned != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong total_assigned_student expected = 0 got = %v", topicAssigned)
		}
		if topicCompleted != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_student expected = 0 got = %v", topicCompleted)
		}

		if topicAverageScore != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong average_score expected = 0 got = %v", topicAverageScore)
		}
		return utils.StepStateToContext(ctx, stepState), nil
	}

	if topicAssigned != statisticResp.TopicStatistic[0].TotalAssignedStudent {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong total_assigned_student expected = %v got = %v", topicAssigned, statisticResp.TopicStatistic[0].TotalAssignedStudent)
	}

	if topicCompleted != statisticResp.TopicStatistic[0].CompletedStudent {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_student expected = %v got = %v", topicCompleted, statisticResp.TopicStatistic[0].CompletedStudent)
	}

	if topicAverageScore != statisticResp.TopicStatistic[0].AverageScore {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong average_score expected = 0 got = %v", topicAverageScore)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) AssignClassToCourse(ctx context.Context, createFlag string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if createFlag == "true" {
		cs := []*entities.CourseClass{{
			ID:       database.Text(stepState.CourseID + stepState.ClassIDs[0]),
			CourseID: database.Text(stepState.CourseID),
			ClassID:  database.Text(stepState.ClassIDs[0])}}
		repo := repositories.CourseClassRepo{}
		err := repo.BulkUpsert(ctx, s.EurekaDB, cs)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// Mock teacher retrieve topic statistical with no class filter
func (s *Suite) retrieveCourseV3StatisticWithNoClassFilter(ctx context.Context, school, tag string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// Use teacher token for auth
	stepState.Token = stepState.TeacherToken
	ctx = s.AuthHelper.SignedCtx(ctx, stepState.Token)

	req := &spb.CourseStatisticRequest{
		CourseId:    stepState.CourseID,
		StudyPlanId: stepState.StudyPlanID,
		ClassId:     []string{},
		School: &spb.CourseStatisticRequest_AllSchool{
			AllSchool: true,
		},
	}

	if len(stepState.ClassIDs) != 0 {
		req.ClassId = stepState.ClassIDs
	}

	switch school {
	case "school_id":
		req.School = &spb.CourseStatisticRequest_SchoolId{
			SchoolId: stepState.SchoolIDs[0],
		}
	case "unassigned":
		req.School = &spb.CourseStatisticRequest_Unassigned{
			Unassigned: true,
		}
	}
	if tag == "tag_id" {
		req.StudentTagIds = stepState.TagIDs
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = spb.NewStatisticsClient(s.EurekaConn).RetrieveCourseStatisticV2(ctx, req)

	if stepState.ResponseErr != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error RetrieveCourseStatisticV3: %w", stepState.ResponseErr)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIsV3(ctx context.Context, topicAssigned, topicCompleted, topicAverageScore int32) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	statisticResp := stepState.Response.(*spb.CourseStatisticResponse)

	if len(statisticResp.TopicStatistic) == 0 {
		if topicAssigned != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong total_assigned_student expected = 0 got = %v", topicAssigned)
		}
		if topicCompleted != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_student expected = 0 got = %v", topicCompleted)
		}

		if topicAverageScore != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong average_score expected = 0 got = %v", topicAverageScore)
		}
		return utils.StepStateToContext(ctx, stepState), nil
	}

	if topicAssigned != statisticResp.TopicStatistic[0].TotalAssignedStudent {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong total_assigned_student expected = %v got = %v", topicAssigned, statisticResp.TopicStatistic[0].TotalAssignedStudent)
	}

	if topicCompleted != statisticResp.TopicStatistic[0].CompletedStudent {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_student expected = %v got = %v\n, courseID: %s, studypaln: %s, studyplanitem: %v", topicCompleted, statisticResp.TopicStatistic[0].CompletedStudent, stepState.CourseID, stepState.StudyPlanID, stepState.debug)
	}

	if topicAverageScore != statisticResp.TopicStatistic[0].AverageScore {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong average_score expected = 0 got = %v", topicAverageScore)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) tagUsersValid(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	for _, student := range stepState.Students {
		tagID := idutil.ULIDNow()

		stmtTag := `INSERT INTO public.user_tag (user_tag_id, user_tag_name, user_tag_type, is_archived, user_tag_partner_id, resource_path, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, now(), now()) ON CONFLICT DO NOTHING`
		if _, err := s.EurekaDB.Exec(ctx, stmtTag, tagID, tagID, tagID, false, idutil.ULIDNow(), fmt.Sprintf("%d", constant.ManabieSchool)); err != nil {
			return nil, err
		}

		stmtTU := `INSERT INTO public.tagged_user (user_id, tag_id, resource_path, created_at, updated_at)
					VALUES ($1, $2, $3, now(), now()) ON CONFLICT DO NOTHING`
		if _, err := s.EurekaDB.Exec(ctx, stmtTU, student.ID, tagID, fmt.Sprintf("%d", constant.ManabieSchool)); err != nil {
			return nil, err
		}

		stepState.TagIDs = append(stepState.TagIDs, tagID)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
