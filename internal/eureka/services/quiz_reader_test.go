package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/s3"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	bpb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRetrieveQuizTests(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupTeacher)

	shuffledQuizSetRepo := new(mock_repositories.MockShuffledQuizSetRepo)
	studentEventLogRepo := new(mock_repositories.MockStudentEventLogRepo)
	learningTimeCalculator := &LearningTimeCalculator{}

	s := QuizReaderService{
		ShuffledQuizSetRepo:       shuffledQuizSetRepo,
		StudentEventLogRepo:       studentEventLogRepo,
		LearningTimeCalculatorSvc: learningTimeCalculator,
	}

	quizTests := entities.ShuffledQuizSets{}
	quizTests.Add()
	quizTests.Add()
	quizTests.Add()
	logs := make([]*entities.StudentEventLog, 0)

	studyPlanItemID := []string{"study plan item id"}
	testCases := []TestCase{
		{
			name: "happy case retrieve quiz tests",
			ctx:  ctx,
			req: &epb.RetrieveQuizTestsRequest{
				StudyPlanItemId: studyPlanItemID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetByStudyPlanItems", ctx, mock.Anything, mock.Anything).Once().Return(quizTests, nil)
				studentEventLogRepo.On("RetrieveStudentEventLogsByStudyPlanItemIDs", ctx, mock.Anything, mock.Anything).Return(logs, nil)
			},
		},
		{
			name: "missing study plan item id",
			ctx:  ctx,
			req: &epb.RetrieveQuizTestsRequest{
				StudyPlanItemId: []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have Study Plan Item Id").Error()),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)

		_, err := s.RetrieveQuizTests(testCase.ctx, testCase.req.(*epb.RetrieveQuizTestsRequest))

		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestRetrieveStudentsSubmissionHistory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	submittedAt := time.Now().UTC().Round(time.Second)
	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupStudent)
	db := &mock_database.Ext{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	questionGroupRepo := &mock_repositories.MockQuestionGroupRepo{}
	s := QuizReaderService{
		DB:                  db,
		ShuffledQuizSetRepo: shuffledQuizSetRepo,
		QuizRepo:            quizRepo,
		QuestionGroupRepo:   questionGroupRepo,
	}

	shuffledQuizSetID := "shuffled_quiz_set_id"
	submissionHistory := make(map[pgtype.Text]pgtype.JSONB)
	orderedQuizList := []pgtype.Text{database.Text("quiz_id_1"), database.Text("quiz_id_2"), database.Text("quiz_id_3"), database.Text("quiz_id_4"), database.Text("quiz_id_5")}
	submissionHistory[orderedQuizList[0]] = database.JSONB(fmt.Sprintf(`{"quiz_id": "quiz_id_1", "quiz_type": "QUIZ_TYPE_MCQ", "correctness": [false], "filled_text": null, "is_accepted": false, "correct_text": null, "submitted_at": "%s", "correct_index": [3], "selected_index": [1]}`, submittedAt.Format(time.RFC3339)))
	submissionHistory[orderedQuizList[1]] = database.JSONB(fmt.Sprintf(`{"quiz_id": "quiz_id_2", "quiz_type": "QUIZ_TYPE_FIB", "correctness": [true, false], "filled_text": ["abc","def"], "is_accepted": true, "correct_text": ["abc", "xyz"], "submitted_at": "%s", "correct_index": null, "selected_index": null}`, submittedAt.Format(time.RFC3339)))
	submissionHistory[orderedQuizList[2]] = pgtype.JSONB{Status: pgtype.Null}
	submissionHistory[orderedQuizList[3]] = database.JSONB(fmt.Sprintf(`{"quiz_id": "quiz_id_4", "quiz_type": "QUIZ_TYPE_ORD", "correctness": [true, true, true], "submitted_keys": ["1", "2", "3"], "is_accepted": true, "correct_keys": ["1", "2", "3"], "submitted_at": "%s", "correct_index": null, "selected_index": null}`, submittedAt.Format(time.RFC3339)))
	submissionHistory[orderedQuizList[4]] = pgtype.JSONB{Status: pgtype.Null}
	loID := database.Text("lo_id")
	q1 := generateQuiz()
	q1.ExternalID.Set("quiz_id_1")
	q2 := generateQuiz()
	q2.ExternalID.Set("quiz_id_2")
	q2.Kind = database.Text(cpb.QuizType_QUIZ_TYPE_FIB.String())
	q3 := generateQuiz()
	q3.ExternalID.Set("quiz_id_3")
	q3.Kind = database.Text(cpb.QuizType_QUIZ_TYPE_FIB.String())
	q4 := generateQuiz()
	q4.ExternalID = orderedQuizList[3]
	q4.Kind = database.Text(cpb.QuizType_QUIZ_TYPE_ORD.String())
	q5 := generateQuiz()
	q5.ExternalID = orderedQuizList[4]
	q5.Kind = database.Text(cpb.QuizType_QUIZ_TYPE_ORD.String())
	quizzesWrap := entities.Quizzes{&q1, &q2, &q3, &q4, &q5}
	seed := database.Text(strconv.FormatInt(time.Now().Unix(), 10))

	idx := []pgtype.Int4{database.Int4(1), database.Int4(2)}

	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")

	questionGroups := entities.QuestionGroups{
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("group-1"),
			LearningMaterialID: loID,
			Name:               database.Text("name 1"),
			Description:        database.Text("description 1"),
			RichDescription: database.JSONB(&entities.RichText{
				Raw:         "raw rich text",
				RenderedURL: url,
			}),
		},
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("group-2"),
			LearningMaterialID: loID,
			Name:               database.Text("name 2"),
			Description:        database.Text("description 2"),
			RichDescription: database.JSONB(&entities.RichText{
				Raw:         "raw rich text",
				RenderedURL: url,
			}),
		},
	}
	questionGroups[0].SetTotalChildrenAndPoints(2, 3)
	questionGroups[1].SetTotalChildrenAndPoints(5, 6)

	respQuestionGroups, _ := entities.QuestionGroupsToQuestionGroupProtoBufMess(questionGroups)
	emptyQuestionGroupsMessage, _ := entities.QuestionGroupsToQuestionGroupProtoBufMess(entities.QuestionGroups{})

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &epb.RetrieveSubmissionHistoryRequest{
				SetId: shuffledQuizSetID,
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 3,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.ListQuizzesOfLOResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:        orderedQuizList[0].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_MCQ,
						SelectedIndex: []uint32{1},
						CorrectIndex:  []uint32{3},
						FilledText:    nil,
						CorrectText:   nil,
						Correctness:   []bool{false},
						IsAccepted:    false,
						Core:          nil,
						SubmittedAt:   timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[1].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_FIB,
						SelectedIndex: nil,
						CorrectIndex:  nil,
						FilledText:    []string{"abc", "def"},
						CorrectText:   []string{"abc", "xyz"},
						Correctness:   []bool{true, false},
						IsAccepted:    true,
						Core:          nil,
						SubmittedAt:   timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[2].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_FIB,
						SelectedIndex: nil,
						CorrectIndex:  []uint32{},
						FilledText:    nil,
						CorrectText:   []string{"qwewqeqweqwe", "hello", "goodbye"},
						Correctness:   nil,
						IsAccepted:    false,
						Core:          nil,
						SubmittedAt:   nil,
					},
					{
						QuizId:        orderedQuizList[3].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_ORD,
						SelectedIndex: nil,
						CorrectIndex:  nil,
						FilledText:    nil,
						CorrectText:   nil,
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								SubmittedKeys: []string{"1", "2", "3"},
								CorrectKeys:   []string{"1", "2", "3"},
							},
						},
						Correctness: []bool{true, true, true},
						IsAccepted:  true,
						Core:        nil,
						SubmittedAt: timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[4].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_ORD,
						SelectedIndex: nil,
						CorrectIndex:  []uint32{},
						FilledText:    nil,
						CorrectText:   []string{},
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								CorrectKeys: []string{"1", "2", "3"},
							},
						},
						Correctness: nil,
						IsAccepted:  false,
						Core:        nil,
						SubmittedAt: nil,
					},
				},
				QuestionGroups: emptyQuestionGroupsMessage,
			},
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetLoID", ctx, s.DB, database.Text(shuffledQuizSetID)).
					Return(loID, nil).Once()
				shuffledQuizSetRepo.On("GetSubmissionHistory", ctx, s.DB, database.Text(shuffledQuizSetID), database.Int4(0), database.Int4(3)).
					Return(submissionHistory, orderedQuizList, nil).Once()
				quizIDs := make([]string, 0, len(orderedQuizList))
				for _, id := range orderedQuizList {
					quizIDs = append(quizIDs, id.String)
				}
				quizRepo.On("GetByExternalIDs", ctx, s.DB, database.TextArray(quizIDs), loID).Once().Return(quizzesWrap, nil)
				shuffledQuizSetRepo.On("GetSeed", ctx, s.DB, database.Text(shuffledQuizSetID)).Times(3).Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q1.ExternalID).Once().Return(idx[0], nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q4.ExternalID).Once().Return(database.Int4(3), nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q5.ExternalID).Once().Return(database.Int4(4), nil)
			},
		},
		{
			name: "retrieve submission history for quiz set having question groups",
			ctx:  ctx,
			req: &epb.RetrieveSubmissionHistoryRequest{
				SetId: shuffledQuizSetID,
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 3,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.ListQuizzesOfLOResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:        orderedQuizList[0].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_MCQ,
						SelectedIndex: []uint32{1},
						CorrectIndex:  []uint32{3},
						FilledText:    nil,
						CorrectText:   nil,
						Correctness:   []bool{false},
						IsAccepted:    false,
						Core:          nil,
						SubmittedAt:   timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[1].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_FIB,
						SelectedIndex: nil,
						CorrectIndex:  nil,
						FilledText:    []string{"abc", "def"},
						CorrectText:   []string{"abc", "xyz"},
						Correctness:   []bool{true, false},
						IsAccepted:    true,
						Core:          nil,
						SubmittedAt:   timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[2].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_FIB,
						SelectedIndex: nil,
						CorrectIndex:  []uint32{},
						FilledText:    nil,
						CorrectText:   []string{"qwewqeqweqwe", "hello", "goodbye"},
						Correctness:   nil,
						IsAccepted:    false,
						Core:          nil,
						SubmittedAt:   nil,
					},
					{
						QuizId:        orderedQuizList[3].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_ORD,
						SelectedIndex: nil,
						CorrectIndex:  nil,
						FilledText:    nil,
						CorrectText:   nil,
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								SubmittedKeys: []string{"1", "2", "3"},
								CorrectKeys:   []string{"1", "2", "3"},
							},
						},
						Correctness: []bool{true, true, true},
						IsAccepted:  true,
						Core:        nil,
						SubmittedAt: timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[4].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_ORD,
						SelectedIndex: nil,
						CorrectIndex:  []uint32{},
						FilledText:    nil,
						CorrectText:   []string{},
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								CorrectKeys: []string{"1", "2", "3"},
							},
						},
						Correctness: nil,
						IsAccepted:  false,
						Core:        nil,
						SubmittedAt: nil,
					},
				},
				QuestionGroups: respQuestionGroups,
			},
			setup: func(ctx context.Context) {
				q2.QuestionGroupID = database.Text("group-1")
				q3.QuestionGroupID = database.Text("group-2")
				shuffledQuizSetRepo.On("GetLoID", mock.Anything, s.DB, mock.Anything).Once().Return(loID, nil)
				shuffledQuizSetRepo.On("GetSubmissionHistory", mock.Anything, s.DB, mock.Anything, mock.Anything, mock.Anything).Once().Return(submissionHistory, orderedQuizList, nil)
				quizIDs := make([]string, 0, len(orderedQuizList))
				for _, id := range orderedQuizList {
					quizIDs = append(quizIDs, id.String)
				}
				quizRepo.On("GetByExternalIDs", mock.Anything, s.DB, database.TextArray(quizIDs), loID).Once().Return(quizzesWrap, nil)
				shuffledQuizSetRepo.On("GetSeed", ctx, s.DB, database.Text(shuffledQuizSetID)).Times(3).Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q1.ExternalID).Once().Return(idx[0], nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q4.ExternalID).Once().Return(database.Int4(3), nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q5.ExternalID).Once().Return(database.Int4(4), nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", mock.Anything, s.DB, mock.Anything).
					Once().
					Run(func(args mock.Arguments) {
						questionGroupIDs := args[2].([]string)
						assert.ElementsMatch(t, questionGroupIDs, []string{"group-1", "group-2"})
					}).
					Return(questionGroups, nil)
			},
		},
		{
			name: "retrieve submission history for quiz set having no question groups",
			ctx:  ctx,
			req: &epb.RetrieveSubmissionHistoryRequest{
				SetId: shuffledQuizSetID,
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 3,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.ListQuizzesOfLOResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:        orderedQuizList[0].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_MCQ,
						SelectedIndex: []uint32{1},
						CorrectIndex:  []uint32{3},
						FilledText:    nil,
						CorrectText:   nil,
						Correctness:   []bool{false},
						IsAccepted:    false,
						Core:          nil,
						SubmittedAt:   timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[1].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_FIB,
						SelectedIndex: nil,
						CorrectIndex:  nil,
						FilledText:    []string{"abc", "def"},
						CorrectText:   []string{"abc", "xyz"},
						Correctness:   []bool{true, false},
						IsAccepted:    true,
						Core:          nil,
						SubmittedAt:   timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[2].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_FIB,
						SelectedIndex: nil,
						CorrectIndex:  []uint32{},
						FilledText:    nil,
						CorrectText:   []string{"qwewqeqweqwe", "hello", "goodbye"},
						Correctness:   nil,
						IsAccepted:    false,
						Core:          nil,
						SubmittedAt:   nil,
					},
					{
						QuizId:        orderedQuizList[3].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_ORD,
						SelectedIndex: nil,
						CorrectIndex:  nil,
						FilledText:    nil,
						CorrectText:   nil,
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								SubmittedKeys: []string{"1", "2", "3"},
								CorrectKeys:   []string{"1", "2", "3"},
							},
						},
						Correctness: []bool{true, true, true},
						IsAccepted:  true,
						Core:        nil,
						SubmittedAt: timestamppb.New(submittedAt),
					},
					{
						QuizId:        orderedQuizList[4].String,
						QuizType:      cpb.QuizType_QUIZ_TYPE_ORD,
						SelectedIndex: nil,
						CorrectIndex:  []uint32{},
						FilledText:    nil,
						CorrectText:   []string{},
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								CorrectKeys: []string{"1", "2", "3"},
							},
						},
						Correctness: nil,
						IsAccepted:  false,
						Core:        nil,
						SubmittedAt: nil,
					},
				},
				QuestionGroups: emptyQuestionGroupsMessage,
			},
			setup: func(ctx context.Context) {
				q2.QuestionGroupID = database.Text("")
				q3.QuestionGroupID = database.Text("")
				shuffledQuizSetRepo.On("GetLoID", mock.Anything, s.DB, mock.Anything).Once().Return(loID, nil)
				shuffledQuizSetRepo.On("GetSubmissionHistory", mock.Anything, s.DB, mock.Anything, mock.Anything, mock.Anything).Once().Return(submissionHistory, orderedQuizList, nil)
				quizIDs := make([]string, 0, len(orderedQuizList))
				for _, id := range orderedQuizList {
					quizIDs = append(quizIDs, id.String)
				}
				quizRepo.On("GetByExternalIDs", mock.Anything, s.DB, database.TextArray(quizIDs), loID).Once().Return(quizzesWrap, nil)
				shuffledQuizSetRepo.On("GetSeed", ctx, s.DB, database.Text(shuffledQuizSetID)).Times(3).Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q1.ExternalID).Once().Return(idx[0], nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q4.ExternalID).Once().Return(database.Int4(3), nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q5.ExternalID).Once().Return(database.Int4(4), nil)
			},
		},
		{
			name: "GetQuestionGroupsByIDs failed",
			ctx:  ctx,
			req: &epb.RetrieveSubmissionHistoryRequest{
				SetId: shuffledQuizSetID,
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 3,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("QuestionGroupRepo.GetQuestionGroupsByIDs: no rows in result set").Error()),
			setup: func(ctx context.Context) {
				q2.QuestionGroupID = database.Text("group-1")
				q3.QuestionGroupID = database.Text("group-2")
				shuffledQuizSetRepo.On("GetLoID", mock.Anything, s.DB, mock.Anything).Once().Return(loID, nil)
				shuffledQuizSetRepo.On("GetSubmissionHistory", mock.Anything, s.DB, mock.Anything, mock.Anything, mock.Anything).Once().Return(submissionHistory, orderedQuizList, nil)
				quizIDs := make([]string, 0, len(orderedQuizList))
				for _, id := range orderedQuizList {
					quizIDs = append(quizIDs, id.String)
				}
				quizRepo.On("GetByExternalIDs", mock.Anything, s.DB, database.TextArray(quizIDs), loID).Once().Return(quizzesWrap, nil)
				shuffledQuizSetRepo.On("GetSeed", ctx, s.DB, database.Text(shuffledQuizSetID)).Times(3).Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q1.ExternalID).Once().Return(idx[0], nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q4.ExternalID).Once().Return(database.Int4(3), nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, s.DB, database.Text(shuffledQuizSetID), q5.ExternalID).Once().Return(database.Int4(4), nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, s.DB, mock.Anything).
					Run(func(args mock.Arguments) {
						questionGroupIDs := args[2].([]string)
						assert.ElementsMatch(t, questionGroupIDs, []string{"group-1", "group-2"})
					}).
					Once().
					Return(questionGroups, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			res, err := s.RetrieveSubmissionHistory(testCase.ctx, testCase.req.(*epb.RetrieveSubmissionHistoryRequest))
			assert.Equal(t, testCase.expectedErr, err)
			req := testCase.req.(*epb.RetrieveSubmissionHistoryRequest)
			if testCase.expectedErr == nil {
				expectedResp := testCase.expectedResp.(*epb.ListQuizzesOfLOResponse)
				require.NoError(t, err)
				assert.Equal(t, req.Paging.Limit, res.NextPage.Limit)
				assert.Equal(t, req.Paging.Offset, res.NextPage.Offset)
				assert.NotNil(t, res)
				assert.ElementsMatch(t, expectedResp.QuestionGroups, res.QuestionGroups)

				// check answer log
				assert.Len(t, res.Logs, len(expectedResp.Logs))

				for i, log := range expectedResp.Logs {
					assert.NotNil(t, res.Logs[i].Core)
					assert.Equal(t, log.SubmittedAt.AsTime(), res.Logs[i].SubmittedAt.AsTime())
					log.Core = res.Logs[i].Core
					assert.Equal(t, log, res.Logs[i])
				}
			}
			mock.AssertExpectationsForObjects(t, db, shuffledQuizSetRepo, quizRepo, questionGroupRepo)
		})
	}
}

func TestRetrieveTotalQuizLOs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupTeacher)

	quizSetRepo := new(mock_repositories.MockQuizSetRepo)

	s := QuizReaderService{
		QuizSetRepo: quizSetRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case retrieve quiz tests",
			ctx:  ctx,
			req: &epb.RetrieveTotalQuizLOsRequest{
				LoIds: []string{"lo-id"},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetTotalQuiz", ctx, mock.Anything, mock.Anything).Return(map[string]int32{"lo-id": 5}, nil)
			},
		},
		{
			name:        "missing los",
			ctx:         ctx,
			req:         &epb.RetrieveTotalQuizLOsRequest{},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have lo ids").Error()),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)

		_, err := s.RetrieveTotalQuizLOs(testCase.ctx, testCase.req.(*epb.RetrieveTotalQuizLOsRequest))

		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestQuizModifierService_ListQuizzesOfLO(t *testing.T) {
	t.Parallel()
	pgTimestamp := pgtype.Timestamptz{Status: pgtype.Null}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	quizSetRepo := &mock_repositories.MockQuizSetRepo{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	questionGroupRepo := &mock_repositories.MockQuestionGroupRepo{}

	quizzes := getQuizzes(
		8,
		cpb.QuizType_QUIZ_TYPE_MCQ.String(),
		cpb.QuizType_QUIZ_TYPE_FIB.String(),
		cpb.QuizType_QUIZ_TYPE_POW.String(),
		cpb.QuizType_QUIZ_TYPE_TAD.String(),
		cpb.QuizType_QUIZ_TYPE_MIQ.String(),
		cpb.QuizType_QUIZ_TYPE_MAQ.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
		cpb.QuizType_QUIZ_TYPE_ESQ.String(),
	)

	quizzesSameQuestionGroup := getQuizzes(
		8,
		cpb.QuizType_QUIZ_TYPE_MCQ.String(),
		cpb.QuizType_QUIZ_TYPE_FIB.String(),
		cpb.QuizType_QUIZ_TYPE_POW.String(),
		cpb.QuizType_QUIZ_TYPE_TAD.String(),
		cpb.QuizType_QUIZ_TYPE_MIQ.String(),
		cpb.QuizType_QUIZ_TYPE_MAQ.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
		cpb.QuizType_QUIZ_TYPE_ESQ.String(),
	)
	for i := range quizzesSameQuestionGroup {
		quizzesSameQuestionGroup[i].QuestionGroupID.Set("group-id")
	}

	quizzesDifferentQuestionGroup := getQuizzes(
		8,
		cpb.QuizType_QUIZ_TYPE_MCQ.String(),
		cpb.QuizType_QUIZ_TYPE_FIB.String(),
		cpb.QuizType_QUIZ_TYPE_POW.String(),
		cpb.QuizType_QUIZ_TYPE_TAD.String(),
		cpb.QuizType_QUIZ_TYPE_MIQ.String(),
		cpb.QuizType_QUIZ_TYPE_MAQ.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
		cpb.QuizType_QUIZ_TYPE_ESQ.String(),
	)
	for i := range quizzesDifferentQuestionGroup {
		quizzesDifferentQuestionGroup[i].QuestionGroupID.Set(fmt.Sprintf("group-id-%d", i))
	}
	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")

	s := &QuizReaderService{
		DB:                db,
		QuizSetRepo:       quizSetRepo,
		QuizRepo:          quizRepo,
		QuestionGroupRepo: questionGroupRepo,
	}
	testCases := []TestCase{
		{
			name: "happy case attach successfully",
			ctx:  ctx,
			req: &bpb.ListQuizzesOfLORequest{
				LoId: "lo_id_1",
				Paging: &cpb.Paging{
					Limit: 1,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.ListQuizzesOfLOResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:       quizzes[0].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzes[0].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzes[1].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzes[1].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzes[2].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzes[2].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzes[3].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzes[3].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzes[4].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzes[4].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzes[5].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzes[5].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzes[6].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzes[6].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{},
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								CorrectKeys: []string{"key-1", "key-2", "key-3"},
							},
						},
					},
					{
						QuizId:       quizzes[7].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzes[7].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{},
					},
				},
				NextPage:       nil,
				QuestionGroups: []*cpb.QuestionGroup{},
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizExternalIDs", ctx, s.DB, mock.AnythingOfType("pgtype.Text"), database.Int8(1), database.Int8(0)).
					Once().Return(quizzes.GetExternalIDs(), nil)
				quizRepo.On("GetByExternalIDs", ctx, s.DB, database.TextArray(quizzes.GetExternalIDs()), database.Text("lo_id_1")).
					Once().Return(quizzes, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{}).
					Once().
					Return(entities.QuestionGroups{}, nil)
			},
		},
		{
			name: "error no rows",
			ctx:  ctx,
			req: &bpb.ListQuizzesOfLORequest{
				LoId: "lo_id_1",
				Paging: &cpb.Paging{
					Limit: 1,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("ListQuizzesOfLO.QuizSetRepo.GetQuizSetByLoID: %v", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizExternalIDs", ctx, s.DB, mock.AnythingOfType("pgtype.Text"), mock.AnythingOfType("pgtype.Int8"), mock.AnythingOfType("pgtype.Int8")).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "happy case attach successfully when all quizzes having same group",
			ctx:  ctx,
			req: &bpb.ListQuizzesOfLORequest{
				LoId: "lo_id_1",
				Paging: &cpb.Paging{
					Limit: 1,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.ListQuizzesOfLOResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:       quizzesSameQuestionGroup[0].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesSameQuestionGroup[0].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzesSameQuestionGroup[1].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesSameQuestionGroup[1].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzesSameQuestionGroup[2].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesSameQuestionGroup[2].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzesSameQuestionGroup[3].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesSameQuestionGroup[3].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzesSameQuestionGroup[4].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesSameQuestionGroup[4].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzesSameQuestionGroup[5].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesSameQuestionGroup[5].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzesSameQuestionGroup[6].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesSameQuestionGroup[6].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{},
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								CorrectKeys: []string{"key-1", "key-2", "key-3"},
							},
						},
					},
					{
						QuizId:       quizzesSameQuestionGroup[7].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesSameQuestionGroup[7].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{},
					},
				},
				NextPage: nil,
				QuestionGroups: []*cpb.QuestionGroup{
					{
						QuestionGroupId:    "group-id",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						}},
				},
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizExternalIDs", ctx, s.DB, mock.AnythingOfType("pgtype.Text"), database.Int8(1), database.Int8(0)).
					Once().Return(quizzesSameQuestionGroup.GetExternalIDs(), nil)
				quizRepo.On("GetByExternalIDs", ctx, s.DB, database.TextArray(quizzesSameQuestionGroup.GetExternalIDs()), database.Text("lo_id_1")).
					Once().Return(quizzesSameQuestionGroup, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"group-id"}).
					Once().
					Return(entities.QuestionGroups{
						{
							QuestionGroupID:    database.Text("group-id"),
							LearningMaterialID: database.Text("lo_id_1"),
							RichDescription: database.JSONB(&entities.RichText{
								Raw:         "raw rich text",
								RenderedURL: url,
							}),
						},
					}, nil)
			},
		},
		{
			name: "happy case attach successfully when all quizzes having different groups",
			ctx:  ctx,
			req: &bpb.ListQuizzesOfLORequest{
				LoId: "lo_id_1",
				Paging: &cpb.Paging{
					Limit: 1,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.ListQuizzesOfLOResponse{
				Logs: []*cpb.AnswerLog{
					{
						QuizId:       quizzesDifferentQuestionGroup[0].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesDifferentQuestionGroup[0].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzesDifferentQuestionGroup[1].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesDifferentQuestionGroup[1].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzesDifferentQuestionGroup[2].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesDifferentQuestionGroup[2].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzesDifferentQuestionGroup[3].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesDifferentQuestionGroup[3].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{"3213213", "3213214", "3213215"},
					},
					{
						QuizId:       quizzesDifferentQuestionGroup[4].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesDifferentQuestionGroup[4].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzesDifferentQuestionGroup[5].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesDifferentQuestionGroup[5].Kind.String]),
						CorrectIndex: []uint32{0, 1},
						CorrectText:  []string{},
					},
					{
						QuizId:       quizzesDifferentQuestionGroup[6].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesDifferentQuestionGroup[6].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{},
						Result: &cpb.AnswerLog_OrderingResult{
							OrderingResult: &cpb.OrderingResult{
								CorrectKeys: []string{"key-1", "key-2", "key-3"},
							},
						},
					},
					{
						QuizId:       quizzesDifferentQuestionGroup[7].ExternalID.String,
						QuizType:     cpb.QuizType(cpb.QuizType_value[quizzesDifferentQuestionGroup[7].Kind.String]),
						CorrectIndex: []uint32{},
						CorrectText:  []string{},
					},
				},
				NextPage: nil,
				QuestionGroups: []*cpb.QuestionGroup{
					{
						QuestionGroupId:    "group-id-0",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						},
					},
					{
						QuestionGroupId:    "group-id-1",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						},
					},
					{
						QuestionGroupId:    "group-id-2",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						},
					},
					{
						QuestionGroupId:    "group-id-3",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						},
					},
					{
						QuestionGroupId:    "group-id-4",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						},
					},
					{
						QuestionGroupId:    "group-id-5",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						},
					},
					{
						QuestionGroupId:    "group-id-6",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						},
					},
					{
						QuestionGroupId:    "group-id-7",
						LearningMaterialId: "lo_id_1",
						CreatedAt:          timestamppb.New(pgTimestamp.Time),
						UpdatedAt:          timestamppb.New(pgTimestamp.Time),
						RichDescription: &cpb.RichText{
							Raw:      "raw rich text",
							Rendered: url,
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizExternalIDs", ctx, s.DB, mock.AnythingOfType("pgtype.Text"), database.Int8(1), database.Int8(0)).
					Once().Return(quizzesDifferentQuestionGroup.GetExternalIDs(), nil)
				quizRepo.On("GetByExternalIDs", ctx, s.DB, database.TextArray(quizzesDifferentQuestionGroup.GetExternalIDs()), database.Text("lo_id_1")).
					Once().Return(quizzesDifferentQuestionGroup, nil)

				grIDs := quizzesDifferentQuestionGroup.GetQuestionGroupIDs()
				questionGroups := make(entities.QuestionGroups, 0, len(grIDs))
				for _, id := range grIDs {
					questionGroups = append(questionGroups, &entities.QuestionGroup{
						QuestionGroupID:    database.Text(id),
						LearningMaterialID: database.Text("lo_id_1"),
						RichDescription: database.JSONB(&entities.RichText{
							Raw:         "raw rich text",
							RenderedURL: url,
						}),
					})
				}
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, grIDs).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			name: "err GetQuestionGroupsByIDs",
			ctx:  ctx,
			req: &bpb.ListQuizzesOfLORequest{
				LoId: "lo_id_1",
				Paging: &cpb.Paging{
					Limit: 1,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("QuestionGroupRepo.GetQuestionGroupsByIDs: %v", puddle.ErrClosedPool).Error()),
			setup: func(ctx context.Context) {
				quizExternalIDs := []string{"quiz_external_id_1", "quiz_external_id_2", "quiz_external_id_3"}
				quizSetRepo.On("GetQuizExternalIDs", ctx, s.DB, mock.AnythingOfType("pgtype.Text"), mock.AnythingOfType("pgtype.Int8"), mock.AnythingOfType("pgtype.Int8")).Once().Return(quizExternalIDs, nil)
				quizzes := entities.Quizzes{
					{
						ExternalID:      database.Text("quiz_external_id_1"),
						QuestionGroupID: database.Text("group-id"),
					},
					{
						ExternalID:      database.Text("quiz_external_id_2"),
						QuestionGroupID: database.Text("group-id-1"),
					},
					{
						ExternalID:      database.Text("quiz_external_id_3"),
						QuestionGroupID: database.Text("group-id"),
					},
				}

				quizRepo.On("GetByExternalIDs", ctx, s.DB, mock.AnythingOfType("pgtype.TextArray"), mock.AnythingOfType("pgtype.Text")).Once().Return(quizzes, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"group-id", "group-id-1"}).
					Once().
					Return(entities.QuestionGroups{}, puddle.ErrClosedPool)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			res, err := s.ListQuizzesOfLO(testCase.ctx, testCase.req.(*bpb.ListQuizzesOfLORequest))
			assert.Equal(t, testCase.expectedErr, err)

			if testCase.expectedErr == nil {
				assert.NotNil(t, res)
				expectedResp := testCase.expectedResp.(*epb.ListQuizzesOfLOResponse)
				assert.Equal(t, res.QuestionGroups, expectedResp.QuestionGroups)

				assert.Len(t, expectedResp.Logs, len(res.Logs))
				for i, actualLog := range res.Logs {
					assert.NotNil(t, actualLog.Core)
					expectedResp.Logs[i].Core = actualLog.Core
					assert.Equal(t, expectedResp.Logs[i], actualLog)
				}
			}

			mock.AssertExpectationsForObjects(t, db, quizSetRepo, quizRepo, questionGroupRepo)
		})
	}
}
