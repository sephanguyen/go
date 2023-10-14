package services

import (
	"container/ring"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/s3"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCreateQuizTest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	quizRepo := new(mock_repositories.MockQuizRepo)
	shuffledQuizSetRepo := new(mock_repositories.MockShuffledQuizSetRepo)
	db := new(mock_database.Ext)
	questionGroupRepo := &mock_repositories.MockQuestionGroupRepo{}
	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")

	s := QuizModifierService{
		DB:                  db,
		QuizRepo:            quizRepo,
		QuizSetRepo:         quizSetRepo,
		ShuffledQuizSetRepo: shuffledQuizSetRepo,
		QuestionGroupRepo:   questionGroupRepo,
	}

	loID := database.Text("VN10-CH-01-L-001.1")
	studentID := database.Text("STUDENT_C1")
	studyPlanItemID := database.Text("STUDY_PLAN_ITEM_ID_C1")
	start := database.Timestamptz(timeutil.Now())

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
	quizExternalIDs := make([]string, 0, len(quizzes))
	for _, quiz := range quizzes {
		quizExternalIDs = append(quizExternalIDs, quiz.ExternalID.String)
	}
	quizzes[0].QuestionGroupID = database.Text("question-gr-id-1")
	quizzes[1].QuestionGroupID = database.Text("question-gr-id-2")

	shuffledQuizExternalIDs := quizExternalIDs
	seed := time.Now().UTC().UnixNano()

	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(shuffledQuizExternalIDs), func(i, j int) {
		shuffledQuizExternalIDs[i], shuffledQuizExternalIDs[j] = shuffledQuizExternalIDs[j], shuffledQuizExternalIDs[i]
	})

	quizSet := &entities.QuizSet{
		ID:              database.Text(idutil.ULIDNow()),
		LoID:            loID,
		QuizExternalIDs: database.TextArray(quizExternalIDs),
		Status:          database.Text("QUIZSET_STATUS_APPROVED"),
		UpdatedAt:       start,
		CreatedAt:       start,
	}

	shuffledQuizSet := &entities.ShuffledQuizSet{
		ID:                database.Text(idutil.ULIDNow()),
		OriginalQuizSetID: quizSet.ID,
		QuizExternalIDs:   database.TextArray(shuffledQuizExternalIDs),
		Status:            quizSet.Status,
		RandomSeed:        database.Text(strconv.FormatInt(seed, 10)),
		UpdatedAt:         start,
		CreatedAt:         start,
	}

	var questionHierarchy pgtype.JSONBArray
	err := questionHierarchy.Set([]interface{}{"a", "b"})
	require.NoError(t, err)

	questionGroups := entities.QuestionGroups{
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("question-gr-id-1"),
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
			QuestionGroupID:    database.Text("question-gr-id-2"),
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

	type expectedRespType struct {
		questionGroup         []*cpb.QuestionGroup
		questionGroupIDofQuiz []string
	}
	respQuestionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(questionGroups)
	respNilQuestionGroups, err := entities.QuestionGroupsToQuestionGroupProtoBufMess(nil)
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 9,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: nil,
			expectedResp: expectedRespType{
				questionGroup:         respQuestionGroups,
				questionGroupIDofQuiz: []string{"question-gr-id-1", "question-gr-id-2"},
			},
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						assert.ElementsMatch(t, database.FromTextArray(ids), quizzes.GetExternalIDs())
					}).
					Return(quizzes, nil).Twice()
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizSet, nil)
				shuffledQuizSetGenerated := &entities.ShuffledQuizSet{}
				shuffledQuizSetRepo.On("GetBySessionID", ctx, mock.Anything, mock.Anything).Return(&entities.ShuffledQuizSet{}, nil)
				shuffledQuizSetRepo.On("Create", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						sh := args[2].(*entities.ShuffledQuizSet)
						assert.NotEmpty(t, sh.RandomSeed.String)
						assert.ElementsMatch(t, quizzes.GetExternalIDs(), database.FromTextArray(sh.QuizExternalIDs))
						shuffledQuizSetGenerated.ID.Set(sh.ID)
						shuffledQuizSetGenerated.OriginalQuizSetID.Set(sh.OriginalQuizSetID)
						shuffledQuizSetGenerated.QuizExternalIDs.Set(sh.QuizExternalIDs.Elements)
						shuffledQuizSetGenerated.RandomSeed.Set(sh.RandomSeed)
						shuffledQuizSetGenerated.QuestionHierarchy.Set(sh.QuestionHierarchy)
						shuffledQuizSetGenerated.Status.Set(sh.Status)
					}).
					Once().
					Return(shuffledQuizSet.ID, nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSetGenerated, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			name: "missing loID",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            "",
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have learning objective id").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing paging field",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have paging field").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "limit equals to zero",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("limit must be positive").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "offset equals to zero",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("offset must be positive").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "offset is negative",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: -1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("offset must be positive").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "offset larger than number of quizzes",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 3,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1000,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: nil,
			expectedResp: expectedRespType{
				questionGroup:         respQuestionGroups,
				questionGroupIDofQuiz: []string{"question-gr-id-1", "question-gr-id-2"},
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizSet, nil)
				shuffledQuizSetGenerated := &entities.ShuffledQuizSet{}
				shuffledQuizSetRepo.On("Create", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						sh := args[2].(*entities.ShuffledQuizSet)
						assert.NotEmpty(t, sh.RandomSeed.String)
						assert.ElementsMatch(t, quizzes.GetExternalIDs(), database.FromTextArray(sh.QuizExternalIDs))
						shuffledQuizSetGenerated.ID.Set(sh.ID)
						shuffledQuizSetGenerated.OriginalQuizSetID.Set(sh.OriginalQuizSetID)
						shuffledQuizSetGenerated.QuizExternalIDs.Set(sh.QuizExternalIDs.Elements)
						shuffledQuizSetGenerated.RandomSeed.Set(sh.RandomSeed)
						shuffledQuizSetGenerated.QuestionHierarchy.Set(sh.QuestionHierarchy)
						shuffledQuizSetGenerated.Status.Set(sh.Status)
					}).
					Once().
					Return(shuffledQuizSet.ID, nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSetGenerated, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						assert.ElementsMatch(t, database.FromTextArray(ids), quizzes.GetExternalIDs())
					}).Return(quizzes, nil).Twice()
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			ctx:  ctx,
			name: "get ShuffledQuizSet when set id empty and there are existing question groups",
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: nil,
			expectedResp: expectedRespType{
				questionGroup:         respQuestionGroups,
				questionGroupIDofQuiz: []string{"question-gr-id-1", "question-gr-id-2"},
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, loID).Once().Return(&entities.QuizSet{
					ID:                database.Text("quiz-set-id"),
					LoID:              loID,
					QuizExternalIDs:   database.TextArray(quizExternalIDs),
					Status:            database.Text("QUIZSET_STATUS_APPROVED"),
					UpdatedAt:         start,
					CreatedAt:         start,
					QuestionHierarchy: questionHierarchy,
				}, nil)
				shuffledQuizSetRepo.On("Create", ctx, db, mock.Anything).Run(func(args mock.Arguments) {
					newShQuizSet := args[2].(*entities.ShuffledQuizSet)
					assert.NotEmpty(t, newShQuizSet.RandomSeed.String)
					assert.NotEmpty(t, newShQuizSet.ID.String)
					assert.Equal(t, "quiz-set-id", newShQuizSet.OriginalQuizSetID.String)
					assert.ElementsMatch(t, quizExternalIDs, database.FromTextArray(newShQuizSet.QuizExternalIDs))
					assert.EqualValues(t, studentID, newShQuizSet.StudentID)
					assert.Equal(t, "QUIZSET_STATUS_APPROVED", newShQuizSet.Status.String)
					assert.Equal(t, studyPlanItemID, newShQuizSet.StudyPlanItemID)
					assert.NotEmpty(t, newShQuizSet.SessionID.String)
					assert.EqualValues(t, questionHierarchy, newShQuizSet.QuestionHierarchy)
				},
				).Once().Return(shuffledQuizSet.ID, nil)
				shuffledQuizSetRepo.On("Get", ctx, db, shuffledQuizSet.ID, database.Int8(1), database.Int8(2)).Once().Return(shuffledQuizSet, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, db, shuffledQuizSet.QuizExternalIDs, loID).Twice().Return(quizzes, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			name: "quiz have question group id not exist",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedResp: expectedRespType{
				questionGroup:         respNilQuestionGroups,
				questionGroupIDofQuiz: []string{"question-gr-id-1", "question-gr-id-2"},
			},
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						assert.ElementsMatch(t, database.FromTextArray(ids), quizzes.GetExternalIDs())
					}).
					Return(quizzes, nil).Twice()
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizSet, nil)
				shuffledQuizSetGenerated := &entities.ShuffledQuizSet{}
				shuffledQuizSetRepo.On("Create", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						sh := args[2].(*entities.ShuffledQuizSet)
						assert.NotEmpty(t, sh.RandomSeed.String)
						assert.ElementsMatch(t, quizzes.GetExternalIDs(), database.FromTextArray(sh.QuizExternalIDs))
						shuffledQuizSetGenerated.ID.Set(sh.ID)
						shuffledQuizSetGenerated.OriginalQuizSetID.Set(sh.OriginalQuizSetID)
						shuffledQuizSetGenerated.QuizExternalIDs.Set(sh.QuizExternalIDs.Elements)
						shuffledQuizSetGenerated.RandomSeed.Set(sh.RandomSeed)
						shuffledQuizSetGenerated.QuestionHierarchy.Set(sh.QuestionHierarchy)
						shuffledQuizSetGenerated.Status.Set(sh.Status)
					}).
					Once().
					Return(shuffledQuizSet.ID, nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSetGenerated, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(entities.QuestionGroups{}, nil)
			},
		},
		{
			name: "GetQuestionGroupsByIDs failed",
			ctx:  ctx,
			req: &epb.CreateQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("QuestionGroupRepo.GetQuestionGroupsByIDs: error").Error()),
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						assert.ElementsMatch(t, database.FromTextArray(ids), quizzes.GetExternalIDs())
					}).
					Return(quizzes, nil).Twice()
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizSet, nil)
				shuffledQuizSetGenerated := &entities.ShuffledQuizSet{}
				shuffledQuizSetRepo.On("Create", ctx, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						sh := args[2].(*entities.ShuffledQuizSet)
						assert.NotEmpty(t, sh.RandomSeed.String)
						assert.ElementsMatch(t, quizzes.GetExternalIDs(), database.FromTextArray(sh.QuizExternalIDs))
						shuffledQuizSetGenerated.ID.Set(sh.ID)
						shuffledQuizSetGenerated.OriginalQuizSetID.Set(sh.OriginalQuizSetID)
						shuffledQuizSetGenerated.QuizExternalIDs.Set(sh.QuizExternalIDs.Elements)
						shuffledQuizSetGenerated.RandomSeed.Set(sh.RandomSeed)
						shuffledQuizSetGenerated.QuestionHierarchy.Set(sh.QuestionHierarchy)
						shuffledQuizSetGenerated.Status.Set(sh.Status)
					}).
					Once().
					Return(shuffledQuizSet.ID, nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(entities.QuestionGroups{}, fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			res, err := s.CreateQuizTest(testCase.ctx, testCase.req.(*epb.CreateQuizTestRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedErr == nil {
				assert.NotNil(t, res)
				expectedResp := testCase.expectedResp.(expectedRespType)
				assert.ElementsMatch(t, expectedResp.questionGroup, res.QuestionGroups)
				for i, v := range expectedResp.questionGroupIDofQuiz {
					assert.Equal(t, v, res.Items[i].Core.QuestionGroupId.GetValue())
				}
				assert.Len(t, res.Items, len(quizzes))

				// check order options
				for _, quiz := range res.Items {
					if quiz.Core.Kind.String() != cpb.QuizType_QUIZ_TYPE_MCQ.String() &&
						quiz.Core.Kind.String() != cpb.QuizType_QUIZ_TYPE_MAQ.String() &&
						quiz.Core.Kind.String() != cpb.QuizType_QUIZ_TYPE_ORD.String() {
						// compare label
						for i, opt := range quiz.Core.Options {
							assert.Equal(t, expectedOrderingQuestionKeys[i], opt.Key)
						}

						if quiz.Core.Kind.String() == cpb.QuizType_QUIZ_TYPE_ESQ.String() {
							assert.Equal(t, quiz.Core.AnswerConfig, &cpb.QuizCore_Essay{
								Essay: &cpb.EssayConfig{
									LimitEnabled: true,
									LimitType:    cpb.EssayLimitType_ESSAY_LIMIT_TYPE_WORD,
									Limit:        10,
								},
							})
						}
					} else {
						actualKey := make([]string, 0, 3)
						count := 0
						for i, opt := range quiz.Core.Options {
							actualKey = append(actualKey, opt.Key)
							// have at least 2 items was shuffled
							if expectedOrderingQuestionKeys[i] != opt.Key {
								count++
							}
						}
						// TODO: Still have risk to list option was not shuffled,
						// plz recheck later
						//assert.LessOrEqual(t, 2, count, fmt.Errorf("must have at least 2 items was shuffled"))
						assert.ElementsMatch(t, actualKey, expectedOrderingQuestionKeys)
					}
				}
			}
			mock.AssertExpectationsForObjects(t, db, quizSetRepo, shuffledQuizSetRepo, quizRepo, questionGroupRepo)
		})
	}
}

func TestCheckQuizCorrectness(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupStudent)

	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	quizRepo := new(mock_repositories.MockQuizRepo)
	shuffledQuizSetRepo := new(mock_repositories.MockShuffledQuizSetRepo)
	SLOCRepo := new(mock_repositories.MockStudentsLearningObjectivesCompletenessRepo)

	// this is the list of quizzes returned from Create quiz test service
	// suppose of having a response when call service create quiz test
	quizzes := getQuizzes(1)
	quizzesbp, _ := toListQuizpb(idutil.ULIDNow(), quizzes)
	createQuizTestResponse := &epb.CreateQuizTestResponse{
		NextPage: &cpb.Paging{
			Limit: 1,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 2,
			},
		},
		Items:     quizzesbp,
		QuizzesId: idutil.ULIDNow(),
	}

	var optionRepoResp []*entities.QuizOption
	for _, quiz := range quizzes {
		err := json.Unmarshal(quiz.Options.Bytes, &optionRepoResp)
		require.NoError(t, err)
	}

	s := QuizModifierService{
		DB:                  db,
		QuizRepo:            quizRepo,
		ShuffledQuizSetRepo: shuffledQuizSetRepo,
		StudentsLearningObjectivesCompletenessRepo: SLOCRepo,
	}

	seed := database.Text(strconv.FormatInt(time.Now().UnixNano(), 10))

	idx := database.Int4(1)
	loID := database.Text("loid")
	testCases := []TestCase{
		// {
		// 	name: "happy case multiple choice",
		// 	ctx:  ctx,
		// 	req: &epb.CheckQuizCorrectnessRequest{
		// 		SetId:  createQuizTestResponse.QuizzesId,
		// 		QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
		// 		Answer: []*epb.Answer{
		// 			{
		// 				Format: &epb.Answer_SelectedIndex{SelectedIndex: 1},
		// 			},
		// 			{
		// 				Format: &epb.Answer_SelectedIndex{SelectedIndex: 2},
		// 			},
		// 		},
		// 	},
		// 	expectedErr: nil,
		// 	setup: func(ctx context.Context) {
		// 		db.On("Begin", ctx).Return(tx, nil)
		// 		tx.On("Commit", mock.Anything).Return(nil)
		// 		shuffledQuizSetRepo.On("GetLoID", ctx, mock.Anything, mock.Anything).Once().Return(loID, nil)
		// 		quizRepo.On("GetByExternalIDs", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzes, nil)
		// 		shuffledQuizSetRepo.On("Get", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities.ShuffledQuizSet{}, nil)
		// 		shuffledQuizSetRepo.On("GetSeed", ctx, mock.Anything, mock.Anything).Once().Return(seed, nil)
		// 		shuffledQuizSetRepo.On("GetQuizIdx", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(idx, nil)
		// 		shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		// 		shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, mock.Anything, mock.Anything).Once().Return(nil)
		// 		shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, mock.Anything, mock.Anything).Once().Return(database.Bool(false), nil)
		// 	},
		// },
		{
			name: "check correctness first time for Fill In The Blank Quiz successfully",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_FilledText{FilledText: "hello"},
					},
					{
						Format: &epb.Answer_FilledText{FilledText: "goodbye"},
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.CheckQuizCorrectnessResponse{
				Correctness:  []bool{false, false},
				IsCorrectAll: false,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil)
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).Once().Return(loID, nil)
				FIBQuizzes := getQuizzes(1)
				for _, quiz := range FIBQuizzes {
					_ = quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_FIB.String())
				}

				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(FIBQuizzes, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var quizAns *entities.QuizAnswer
						err := data.AssignTo(&quizAns)
						require.NoError(t, err)
						assert.NotZero(t, quizAns.SubmittedAt)

						expected := &entities.QuizAnswer{
							QuizID:       FIBQuizzes[0].ExternalID.String,
							QuizType:     cpb.QuizType_QUIZ_TYPE_FIB.String(),
							FilledText:   []string{"hello", "goodbye"},
							CorrectText:  []string{"3213213", "3213214", "3213215"},
							Correctness:  []bool{false, false},
							IsAccepted:   false,
							IsAllCorrect: false,
							SubmittedAt:  quizAns.SubmittedAt,
						}
						assert.Equal(t, expected, quizAns)
					}).
					Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Bool(false), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{}, nil)
			},
		},
		{
			name: "retry correctness for Fill In The Blank Quiz successfully (incorrect)",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_FilledText{FilledText: "hello"},
					},
					{
						Format: &epb.Answer_FilledText{FilledText: "goodbye"},
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.CheckQuizCorrectnessResponse{
				Correctness:  []bool{false, false},
				IsCorrectAll: false,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil)
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(loID, nil)
				FIBQuizzes := getQuizzes(1)
				for _, quiz := range FIBQuizzes {
					_ = quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_FIB.String())
				}

				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(FIBQuizzes, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var quizAns *entities.QuizAnswer
						err := data.AssignTo(&quizAns)
						require.NoError(t, err)
						assert.NotZero(t, quizAns.SubmittedAt)

						expected := &entities.QuizAnswer{
							QuizID:       FIBQuizzes[0].ExternalID.String,
							QuizType:     cpb.QuizType_QUIZ_TYPE_FIB.String(),
							FilledText:   []string{"hello", "goodbye"},
							CorrectText:  []string{"3213213", "3213214", "3213215"},
							Correctness:  []bool{false, false},
							IsAccepted:   false,
							IsAllCorrect: false,
							SubmittedAt:  quizAns.SubmittedAt,
						}
						assert.Equal(t, expected, quizAns)
					}).
					Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Bool(true), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{
					QuizExternalIDs:          database.TextArray([]string{"id-1", "id-2", FIBQuizzes[0].ExternalID.String}),
					OriginalShuffleQuizSetID: database.Text("OriginalShuffleQuizSetID"),
				}, nil)
				shuffledQuizSetRepo.On("GetStudentID", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Text("student-id"), nil)
				shuffledQuizSetRepo.On("GetScore", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Int4(2), database.Int4(3), nil)
				shuffledQuizSetRepo.On("GetExternalIDsFromSubmissionHistory", ctx, tx, database.Text("OriginalShuffleQuizSetID"), false).
					Once().Return(database.TextArray([]string{FIBQuizzes[0].ExternalID.String}), nil)
				score := float32(math.Floor(float64(2) / float64(3) * 100))
				SLOCRepo.On("UpsertFirstQuizCompleteness", ctx, tx, loID, database.Text("student-id"), database.Float4(score)).
					Once().Return(nil)
				SLOCRepo.On("UpsertHighestQuizScore", ctx, tx, loID, database.Text("student-id"), database.Float4(score)).
					Once().Return(nil)
			},
		},
		{
			name: "retry correctness for Fill In The Blank Quiz successfully (correct)",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_FilledText{FilledText: "3213213"},
					},
					{
						Format: &epb.Answer_FilledText{FilledText: "3213214"},
					},
					{
						Format: &epb.Answer_FilledText{FilledText: "3213215"},
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.CheckQuizCorrectnessResponse{
				Correctness:  []bool{true, true, true},
				IsCorrectAll: true,
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil)
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(loID, nil)
				FIBQuizzes := getQuizzes(1)
				for _, quiz := range FIBQuizzes {
					_ = quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_FIB.String())
				}

				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(FIBQuizzes, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var quizAns *entities.QuizAnswer
						err := data.AssignTo(&quizAns)
						require.NoError(t, err)
						assert.NotZero(t, quizAns.SubmittedAt)

						expected := &entities.QuizAnswer{
							QuizID:       FIBQuizzes[0].ExternalID.String,
							QuizType:     cpb.QuizType_QUIZ_TYPE_FIB.String(),
							FilledText:   []string{"3213213", "3213214", "3213215"},
							CorrectText:  []string{"3213213", "3213214", "3213215"},
							Correctness:  []bool{true, true, true},
							IsAccepted:   true,
							IsAllCorrect: true,
							SubmittedAt:  quizAns.SubmittedAt,
							Point:        uint32(FIBQuizzes[0].Point.Int),
						}
						assert.Equal(t, expected, quizAns)
					}).
					Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Bool(true), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{
					QuizExternalIDs:          database.TextArray([]string{"id-1", "id-2", FIBQuizzes[0].ExternalID.String}),
					OriginalShuffleQuizSetID: database.Text("OriginalShuffleQuizSetID"),
				}, nil)
				shuffledQuizSetRepo.On("GetStudentID", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Text("student-id"), nil)
				shuffledQuizSetRepo.On("GetScore", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Int4(3), database.Int4(3), nil)
				shuffledQuizSetRepo.On("GetExternalIDsFromSubmissionHistory", ctx, tx, database.Text("OriginalShuffleQuizSetID"), false).
					Once().Return(database.TextArray([]string{FIBQuizzes[0].ExternalID.String}), nil)
				SLOCRepo.On("UpsertFirstQuizCompleteness", ctx, tx, loID, database.Text("student-id"), database.Float4(100)).
					Once().Return(nil)
				SLOCRepo.On("UpsertHighestQuizScore", ctx, tx, loID, database.Text("student-id"), database.Float4(100)).
					Once().Return(nil)
			},
		},
		{
			name: "missing quizid",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: "",
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SelectedIndex{SelectedIndex: 1},
					},
					{
						Format: &epb.Answer_SelectedIndex{SelectedIndex: 2},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have QuizId").Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "selected index is out of range",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SelectedIndex{SelectedIndex: 1},
					},
					{
						Format: &epb.Answer_SelectedIndex{SelectedIndex: 200},
					},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, errors.New("CheckQuizCorrectness.MultipleChoice.CheckCorrectness: selected index out of range").Error()),
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(loID, nil)
				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(quizzes, nil)
				shuffledQuizSetRepo.On("GetSeed", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, db, database.Text(createQuizTestResponse.QuizzesId), quizzes[0].ExternalID).
					Once().Return(idx, nil)
			},
		},
		{
			name: "missing setid",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  "",
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SelectedIndex{SelectedIndex: 1},
					},
					{
						Format: &epb.Answer_SelectedIndex{SelectedIndex: 2},
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have SetId").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "answer is both multiple choice and fill in the blank",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SelectedIndex{SelectedIndex: 1},
					},
					{
						Format: &epb.Answer_FilledText{FilledText: "hello"},
					},
				},
			},
			expectedErr: status.Error(codes.FailedPrecondition, errors.New("CheckQuizCorrectness.MultipleChoice.CheckCorrectness: your answer is not the multiple choice type").Error()),
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(loID, nil)
				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(quizzes, nil)
				shuffledQuizSetRepo.On("GetSeed", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, db, database.Text(createQuizTestResponse.QuizzesId), quizzes[0].ExternalID).
					Once().Return(idx, nil)
			},
		},
		{
			name: "empty answer",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have answer").Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "check correctness first time for Ordering Quiz successfully (correct)",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-1"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-2"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-3"},
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.CheckQuizCorrectnessResponse{
				Correctness:  []bool{true, true, true},
				IsCorrectAll: true,
				Result: &epb.CheckQuizCorrectnessResponse_OrderingResult{
					OrderingResult: &cpb.OrderingResult{
						CorrectKeys:   []string{"key-1", "key-2", "key-3"},
						SubmittedKeys: []string{"key-1", "key-2", "key-3"},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil)
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).Once().Return(loID, nil)
				orderingQuiz := getQuizzes(1, cpb.QuizType_QUIZ_TYPE_ORD.String())
				for _, quiz := range orderingQuiz {
					//_ = quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
					quiz.ExternalID = database.Text(createQuizTestResponse.Items[0].Core.ExternalId)
				}

				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(orderingQuiz, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var quizAns *entities.QuizAnswer
						err := data.AssignTo(&quizAns)
						require.NoError(t, err)
						assert.NotZero(t, quizAns.SubmittedAt)

						expected := &entities.QuizAnswer{
							QuizID:        orderingQuiz[0].ExternalID.String,
							QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
							SubmittedKeys: []string{"key-1", "key-2", "key-3"},
							CorrectKeys:   []string{"key-1", "key-2", "key-3"},
							Correctness:   []bool{true, true, true},
							IsAccepted:    true,
							IsAllCorrect:  true,
							SubmittedAt:   quizAns.SubmittedAt,
							Point:         uint32(orderingQuiz[0].Point.Int),
						}
						assert.Equal(t, expected, quizAns)
					}).
					Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Bool(false), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{}, nil)
			},
		},
		{
			name: "check correctness first time for Ordering Quiz successfully (incorrect)",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-1"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-3"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-2"},
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.CheckQuizCorrectnessResponse{
				Correctness:  []bool{true, false, false},
				IsCorrectAll: false,
				Result: &epb.CheckQuizCorrectnessResponse_OrderingResult{
					OrderingResult: &cpb.OrderingResult{
						CorrectKeys:   []string{"key-1", "key-2", "key-3"},
						SubmittedKeys: []string{"key-1", "key-3", "key-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil)
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).Once().Return(loID, nil)
				orderingQuiz := getQuizzes(1)
				for _, quiz := range orderingQuiz {
					_ = quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
					quiz.ExternalID = database.Text(createQuizTestResponse.Items[0].Core.ExternalId)
				}

				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(orderingQuiz, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var quizAns *entities.QuizAnswer
						err := data.AssignTo(&quizAns)
						require.NoError(t, err)
						assert.NotZero(t, quizAns.SubmittedAt)

						expected := &entities.QuizAnswer{
							QuizID:        orderingQuiz[0].ExternalID.String,
							QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
							SubmittedKeys: []string{"key-1", "key-3", "key-2"},
							CorrectKeys:   []string{"key-1", "key-2", "key-3"},
							Correctness:   []bool{true, false, false},
							IsAccepted:    false,
							IsAllCorrect:  false,
							SubmittedAt:   quizAns.SubmittedAt,
						}
						assert.Equal(t, expected, quizAns)
					}).
					Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Bool(false), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{}, nil)
			},
		},
		{
			name: "retry correctness for Ordering Quiz successfully (correct)",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-1"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-2"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-3"},
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.CheckQuizCorrectnessResponse{
				Correctness:  []bool{true, true, true},
				IsCorrectAll: true,
				Result: &epb.CheckQuizCorrectnessResponse_OrderingResult{
					OrderingResult: &cpb.OrderingResult{
						CorrectKeys:   []string{"key-1", "key-2", "key-3"},
						SubmittedKeys: []string{"key-1", "key-2", "key-3"},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil)
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(loID, nil)
				orderingQuiz := getQuizzes(1)
				for _, quiz := range orderingQuiz {
					_ = quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
					quiz.ExternalID = database.Text(createQuizTestResponse.Items[0].Core.ExternalId)
				}

				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(orderingQuiz, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var quizAns *entities.QuizAnswer
						err := data.AssignTo(&quizAns)
						require.NoError(t, err)
						assert.NotZero(t, quizAns.SubmittedAt)

						expected := &entities.QuizAnswer{
							QuizID:        orderingQuiz[0].ExternalID.String,
							QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
							SubmittedKeys: []string{"key-1", "key-2", "key-3"},
							CorrectKeys:   []string{"key-1", "key-2", "key-3"},
							Correctness:   []bool{true, true, true},
							IsAccepted:    true,
							IsAllCorrect:  true,
							SubmittedAt:   quizAns.SubmittedAt,
							Point:         uint32(orderingQuiz[0].Point.Int),
						}
						assert.Equal(t, expected, quizAns)
					}).
					Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Bool(true), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{
					QuizExternalIDs:          database.TextArray([]string{"id-1", "id-2", orderingQuiz[0].ExternalID.String}),
					OriginalShuffleQuizSetID: database.Text("OriginalShuffleQuizSetID"),
				}, nil)
				shuffledQuizSetRepo.On("GetStudentID", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Text("student-id"), nil)
				shuffledQuizSetRepo.On("GetScore", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Int4(3), database.Int4(3), nil)
				shuffledQuizSetRepo.On("GetExternalIDsFromSubmissionHistory", ctx, tx, database.Text("OriginalShuffleQuizSetID"), false).
					Once().Return(database.TextArray([]string{orderingQuiz[0].ExternalID.String}), nil)
				SLOCRepo.On("UpsertFirstQuizCompleteness", ctx, tx, loID, database.Text("student-id"), database.Float4(100)).
					Once().Return(nil)
				SLOCRepo.On("UpsertHighestQuizScore", ctx, tx, loID, database.Text("student-id"), database.Float4(100)).
					Once().Return(nil)
			},
		},
		{
			name: "retry correctness for Ordering Quiz successfully (incorrect)",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-1"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-3"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-2"},
					},
				},
			},
			expectedErr: nil,
			expectedResp: &epb.CheckQuizCorrectnessResponse{
				Correctness:  []bool{true, false, false},
				IsCorrectAll: false,
				Result: &epb.CheckQuizCorrectnessResponse_OrderingResult{
					OrderingResult: &cpb.OrderingResult{
						CorrectKeys:   []string{"key-1", "key-2", "key-3"},
						SubmittedKeys: []string{"key-1", "key-3", "key-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", ctx).Return(nil)
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(loID, nil)
				orderingQuiz := getQuizzes(1)
				for _, quiz := range orderingQuiz {
					_ = quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
					quiz.ExternalID = database.Text(createQuizTestResponse.Items[0].Core.ExternalId)
				}

				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(orderingQuiz, nil)
				shuffledQuizSetRepo.On("UpdateSubmissionHistory", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[3].(pgtype.JSONB)
						var quizAns *entities.QuizAnswer
						err := data.AssignTo(&quizAns)
						require.NoError(t, err)
						assert.NotZero(t, quizAns.SubmittedAt)

						expected := &entities.QuizAnswer{
							QuizID:        orderingQuiz[0].ExternalID.String,
							QuizType:      cpb.QuizType_QUIZ_TYPE_ORD.String(),
							SubmittedKeys: []string{"key-1", "key-3", "key-2"},
							CorrectKeys:   []string{"key-1", "key-2", "key-3"},
							Correctness:   []bool{true, false, false},
							IsAccepted:    false,
							IsAllCorrect:  false,
							SubmittedAt:   quizAns.SubmittedAt,
						}
						assert.Equal(t, expected, quizAns)
					}).
					Once().Return(nil)
				shuffledQuizSetRepo.On("UpdateTotalCorrectness", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(nil)
				shuffledQuizSetRepo.On("IsFinishedQuizTest", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Bool(true), nil)
				shuffledQuizSetRepo.On("Get", ctx, tx, database.Text(createQuizTestResponse.QuizzesId), database.Int8(1), database.Int8(1)).
					Once().Return(&entities.ShuffledQuizSet{
					QuizExternalIDs:          database.TextArray([]string{"id-1", "id-2", orderingQuiz[0].ExternalID.String}),
					OriginalShuffleQuizSetID: database.Text("OriginalShuffleQuizSetID"),
				}, nil)
				shuffledQuizSetRepo.On("GetStudentID", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Text("student-id"), nil)
				shuffledQuizSetRepo.On("GetScore", ctx, tx, database.Text(createQuizTestResponse.QuizzesId)).
					Once().Return(database.Int4(2), database.Int4(3), nil)
				shuffledQuizSetRepo.On("GetExternalIDsFromSubmissionHistory", ctx, tx, database.Text("OriginalShuffleQuizSetID"), false).
					Once().Return(database.TextArray([]string{orderingQuiz[0].ExternalID.String}), nil)
				score := float32(math.Floor(float64(2) / float64(3) * 100))
				SLOCRepo.On("UpsertFirstQuizCompleteness", ctx, tx, loID, database.Text("student-id"), database.Float4(score)).
					Once().Return(nil)
				SLOCRepo.On("UpsertHighestQuizScore", ctx, tx, loID, database.Text("student-id"), database.Float4(score)).
					Once().Return(nil)
			},
		},
		{
			name: "check correctness first time for Ordering Quiz with wrong answer format",
			ctx:  ctx,
			req: &epb.CheckQuizCorrectnessRequest{
				SetId:  createQuizTestResponse.QuizzesId,
				QuizId: createQuizTestResponse.Items[0].Core.ExternalId,
				Answer: []*epb.Answer{
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-1"},
					},
					{
						Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-2"},
					},
					{
						Format: &epb.Answer_FilledText{FilledText: "key-3"},
					},
				},
			},
			expectedErr: status.Error(
				codes.FailedPrecondition,
				fmt.Sprintf("questionSrv.CheckQuestionsCorrectness: run opt: WithSubmitQuizAnswersRequest.executor.GetUserAnswerFromSubmitQuizAnswersRequest: your answer is not the ordering type, question %s (external_id), %s (quiz_id)", createQuizTestResponse.Items[0].Core.ExternalId, "quiz-id-1"),
			),
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetLoID", ctx, db, database.Text(createQuizTestResponse.QuizzesId)).Once().Return(loID, nil)
				orderingQuiz := getQuizzes(1)
				for _, quiz := range orderingQuiz {
					_ = quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
					quiz.ID = database.Text("quiz-id-1")
					quiz.ExternalID = database.Text(createQuizTestResponse.Items[0].Core.ExternalId)
				}

				quizRepo.On("GetByExternalIDs", ctx, db, database.TextArray([]string{createQuizTestResponse.Items[0].Core.ExternalId}), loID).
					Once().Return(orderingQuiz, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			res, err := s.CheckQuizCorrectness(testCase.ctx, testCase.req.(*epb.CheckQuizCorrectnessRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedResp, res)
			}

			mock.AssertExpectationsForObjects(t, db, shuffledQuizSetRepo, quizRepo)
		})
	}
}

func getQuizzes(numOfQuizzes int, kind ...string) entities.Quizzes {
	start := database.Timestamptz(timeutil.Now())
	quizzes := entities.Quizzes{}
	r := ring.New(len(kind))
	for i := 0; i < len(kind); i++ {
		r.Value = kind[i]
		r = r.Next()
	}
	ra := rand.New(rand.NewSource(99))
	for i := 0; i < numOfQuizzes; i++ {
		quiz := &entities.Quiz{
			ID:          database.Text(idutil.ULIDNow()),
			ExternalID:  database.Text(idutil.ULIDNow()),
			Country:     database.Text("COUNTRY_VN"),
			SchoolID:    database.Int4(-2147483648),
			Kind:        database.Text("QUIZ_TYPE_MCQ"),
			Question:    database.JSONB(`{"raw": "{\"blocks\":[{\"key\":\"eq20k\",\"text\":\"3213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html"}`),
			Explanation: database.JSONB(`{"raw": "{\"blocks\":[{\"key\":\"4rpf3\",\"text\":\"213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html"}`),
			Options: database.JSONB(`[
				{"key":"key-1" , "label": "1", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true},
				{"key":"key-2" , "label": "2", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so84\",\"text\":\"3213214\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true},
				{"key":"key-3" , "label": "3", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so85\",\"text\":\"3213215\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": false}
				]`),
			TaggedLOs:      database.TextArray([]string{"VN10-CH-01-L-001.1"}),
			DifficultLevel: database.Int4(1),
			CreatedBy:      database.Text("QC6KC30TZWc97APf99ydhonPRct1"),
			ApprovedBy:     database.Text("QC6KC30TZWc97APf99ydhonPRct1"),
			Status:         database.Text("QUIZ_STATUS_APPROVED"),
			UpdatedAt:      start,
			CreatedAt:      start,
			DeletedAt:      pgtype.Timestamptz{},
			Point:          database.Int4(ra.Int31()),
		}
		if len(kind) != 0 {
			quiz.Kind = database.Text(fmt.Sprintf("%v", r.Value))
			if fmt.Sprintf("%v", r.Value) == cpb.QuizType_QUIZ_TYPE_ESQ.String() {
				quiz.Options = database.JSONB(`
				[
					{"key":"key-1" , "label": "1", "configs": [], "content": {"raw": "{\"blocks\":[{\"key\":\"2so83\",\"text\":\"3213213\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html"}, "correctness": true, "answer_config": {"essay": {"limit": 10, "limit_type": "ESSAY_LIMIT_TYPE_WORD", "limit_enabled": true}}}
				]
				`)
			}
			r = r.Next()
		}
		quizzes = append(quizzes, quiz)
	}
	return quizzes
}
func TestCourseModifier_DeleteQuiz(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	quizRepo := new(mock_repositories.MockQuizRepo)
	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	db.On("Begin").Once().Return(tx, nil)
	s := &QuizModifierService{
		DB:          db,
		QuizRepo:    quizRepo,
		QuizSetRepo: quizSetRepo,
	}
	quizID := idutil.ULIDNow()
	quizSet := entities.QuizSet{QuizExternalIDs: database.TextArray([]string{"another quiz id", quizID, "and other quiz id"})}
	quizSets := entities.QuizSets{&quizSet}

	testCases := []TestCase{
		{
			name: "happy case delete successfully",
			ctx:  ctx,
			req: &epb.DeleteQuizRequest{
				QuizId: quizID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				quizRepo.On("DeleteByExternalID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				quizSetRepo.On("GetQuizSetsContainQuiz", mock.Anything, mock.Anything, mock.Anything).Once().Return(quizSets, nil)
				quizSetRepo.On("Delete", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				quizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "missing quiz id in the request",
			ctx:  ctx,
			req: &epb.DeleteQuizRequest{
				QuizId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have quiz id").Error()),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.DeleteQuiz(testCase.ctx, testCase.req.(*epb.DeleteQuizRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestCourseService_UpsertQuiz(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	quizSetRepo := &mock_repositories.MockQuizSetRepo{}
	yasuoUploadModifierService := &mock_services.YasuoUploadModifierServiceClient{}
	yasuoUploadReaderService := &mock_services.YasuoUploadReaderServiceClient{}

	s := &QuizModifierService{
		DB:                  db,
		QuizRepo:            quizRepo,
		QuizSetRepo:         quizSetRepo,
		YasuoUploadModifier: yasuoUploadModifierService,
		YasuoUploadReader:   yasuoUploadReaderService,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	question := "rendered " + idutil.ULIDNow()
	explanation := "rendered " + idutil.ULIDNow()
	option := "rendered " + idutil.ULIDNow()
	url1, _ := generateUploadURL("", "", question)
	url2, _ := generateUploadURL("", "", explanation)
	url3, _ := generateUploadURL("", "", option)
	m := make(map[string]string)
	m[url1] = question
	m[url2] = explanation
	m[url3] = option

	validReq := &epb.UpsertQuizRequest{
		Quiz: &epb.QuizCore{
			ExternalId:  "externalID",
			SchoolId:    constant.ManabieSchool,
			Question:    &cpb.RichText{Raw: question, Rendered: question},
			Explanation: &cpb.RichText{Raw: explanation, Rendered: explanation},
			Options: []*cpb.QuizOption{
				{
					Content:     &cpb.RichText{Raw: option, Rendered: option},
					Correctness: true,
					Label:       "(1)",
				},
			},
		},
		LoId: "lo-id",
	}
	validResp := &epb.UpsertQuizResponse{
		Id: "quiz-id",
	}

	testCases := []TestCase{
		{
			ctx:  ctx,
			name: "err check quiz by externalID",
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrTxClosed)
				ctx, _ = interceptors.GetOutgoingContext(ctx)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
			},
			req:         validReq,
			expectedErr: status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
		},
		{
			ctx:  ctx,
			name: "err get quizset by loID",
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, pgx.ErrTxClosed)
			},
			req:         validReq,
			expectedErr: status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
		},
		{
			ctx:  ctx,
			name: "err delete by externalID",
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)

				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)

				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(pgx.ErrTxClosed)
			},
			req:         validReq,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.QuizRepo.DeleteByExternalID: %w", pgx.ErrTxClosed).Error()),
		},
		{
			ctx:  ctx,
			name: "err create quiz",
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Once().Return(pgx.ErrTxClosed)
			},
			req:         validReq,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.QuizRepo.Create: %w", pgx.ErrTxClosed).Error()),
		},
		{
			ctx:  ctx,
			name: "err upload html content",
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for _, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(nil, fmt.Errorf("err connect s3 host"))
				}
			},
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, errors.New("s.YasuoUploadModifier.UploadHtmlContent: err connect s3 host").Error()),
		},
		{
			ctx:  ctx,
			name: "success without found quizSet",
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{QuizExternalIDs: database.TextArray([]string{"externalID"})}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
			},
			req:          validReq,
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "err delete quiz set",
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{ID: database.Text("quiz-set-id"), QuizExternalIDs: database.TextArray([]string{"another externalID"})}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("quiz-set-id")).
					Once().Return(pgx.ErrTxClosed)
			},
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, "quizSetRepo.Delete: tx is closed"),
		},
		{
			ctx:  ctx,
			name: "succeed with new quizset",
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
					}).
					Return(nil)
			},
			req:          validReq,
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "succeed with upsert quizset",
			setup: func(ctx context.Context) {
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
					}).
					Return(nil)
			},
			req:          validReq,
			expectedResp: validResp,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Test case: " + testCase.name)
			mdCtx := interceptors.NewIncomingContext(testCase.ctx)
			testCase.setup(mdCtx)

			resp, err := s.UpsertQuiz(mdCtx, testCase.req.(*epb.UpsertQuizRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedResp.(*epb.UpsertQuizResponse), resp)
			}
		})
	}
}

func TestCourseService_UpsertSingleQuiz(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	quizSetRepo := &mock_repositories.MockQuizSetRepo{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	questionGroupRepo := &mock_repositories.MockQuestionGroupRepo{}
	yasuoUploadModifierService := &mock_services.YasuoUploadModifierServiceClient{}
	yasuoUploadReaderService := &mock_services.YasuoUploadReaderServiceClient{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	examLOSubmissionAnswerRepo := &mock_repositories.MockExamLOSubmissionAnswerRepo{}
	examLOSubmissionRepo := &mock_repositories.MockExamLOSubmissionRepo{}
	examLORepo := &mock_repositories.MockExamLORepo{}

	s := &QuizModifierService{
		DB:                         db,
		UnleashClient:              mockUnleashClient,
		Env:                        "local",
		QuizRepo:                   quizRepo,
		QuizSetRepo:                quizSetRepo,
		ShuffledQuizSetRepo:        shuffledQuizSetRepo,
		YasuoUploadModifier:        yasuoUploadModifierService,
		YasuoUploadReader:          yasuoUploadReaderService,
		QuestionGroupRepo:          questionGroupRepo,
		ExamLOSubmissionAnswerRepo: examLOSubmissionAnswerRepo,
		ExamLOSubmissionRepo:       examLOSubmissionRepo,
		ExamLORepo:                 examLORepo,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	question := "rendered " + idutil.ULIDNow()
	explanation := "rendered " + idutil.ULIDNow()
	option := "rendered " + idutil.ULIDNow()
	url1, _ := generateUploadURL("", "", question)
	url2, _ := generateUploadURL("", "", explanation)
	url3, _ := generateUploadURL("", "", option)
	m := make(map[string]string)
	m[url1] = question
	m[url2] = explanation
	m[url3] = option

	const validReq, validQuestionGroupReq = "validRequest", "validQuestionGroupRequest"
	getQuizLO := func(typ string, kind ...cpb.QuizType) *epb.UpsertSingleQuizRequest {
		base := &epb.QuizLO{
			Quiz: &cpb.QuizCore{
				Info: &cpb.ContentBasicInfo{
					SchoolId: constant.ManabieSchool,
				},
				ExternalId:  "externalID",
				Question:    &cpb.RichText{Raw: question, Rendered: question},
				Explanation: &cpb.RichText{Raw: explanation, Rendered: explanation},
				Options: []*cpb.QuizOption{
					{
						Content:     &cpb.RichText{Raw: option, Rendered: option},
						Correctness: true,
						Label:       "(1)",
						Attribute: &cpb.QuizItemAttribute{
							Configs: []cpb.QuizItemAttributeConfig{cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_JP},
						},
					},
				},
				Attribute: &cpb.QuizItemAttribute{
					Configs: []cpb.QuizItemAttributeConfig{cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_JP},
				},
				QuestionTagIds:  []string{"id-1", "id-2"},
				QuestionGroupId: &wrapperspb.StringValue{Value: "question group id"},
				LabelType:       cpb.QuizLabelType_QUIZ_LABEL_TYPE_NUMBER,
			},
			LoId: "lo-id",
		}
		if typ != validQuestionGroupReq {
			base.Quiz.QuestionGroupId = nil
		}
		if len(kind) != 0 {
			base.Quiz.Kind = kind[0]
			switch kind[0] {
			case cpb.QuizType_QUIZ_TYPE_ORD:
			case cpb.QuizType_QUIZ_TYPE_ESQ:
				base.Quiz.AnswerConfig = &cpb.QuizCore_Essay{
					Essay: &cpb.EssayConfig{
						LimitEnabled: true,
						LimitType:    cpb.EssayLimitType_ESSAY_LIMIT_TYPE_CHAR,
						Limit:        5000,
					},
				}
			}
		}

		return &epb.UpsertSingleQuizRequest{
			QuizLo: base,
		}
	}

	validResp := &epb.UpsertSingleQuizResponse{
		Id: "quiz-id",
	}

	testCases := []TestCase{
		{
			ctx:  ctx,
			name: "err check quiz by externalID",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrTxClosed)
				ctx, _ = interceptors.GetOutgoingContext(ctx)
			},
			req:         getQuizLO(validReq),
			expectedErr: status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
		},
		{
			ctx:  ctx,
			name: "err get quizset by loID",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, pgx.ErrTxClosed)
			},
			req:         getQuizLO(validReq),
			expectedErr: status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
		},
		{
			ctx:  ctx,
			name: "err delete by externalID",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)

				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)

				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(pgx.ErrTxClosed)
			},
			req:         getQuizLO(validReq),
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.QuizRepo.DeleteByExternalID: %w", pgx.ErrTxClosed).Error()),
		},
		{
			ctx:  ctx,
			name: "err create quiz",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Once().Return(pgx.ErrTxClosed)
			},
			req:         getQuizLO(validReq),
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.QuizRepo.Create: %w", pgx.ErrTxClosed).Error()),
		},
		{
			ctx:  ctx,
			name: "err upload html content",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for _, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(nil, fmt.Errorf("err connect s3 host"))
				}
			},
			req:         getQuizLO(validReq),
			expectedErr: status.Errorf(codes.Internal, errors.New("s.YasuoUploadModifier.UploadHtmlContent: err connect s3 host").Error()),
		},
		{
			ctx:  ctx,
			name: "success without found quizSet",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{QuizExternalIDs: database.TextArray([]string{"externalID"})}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
			},
			req:          getQuizLO(validReq),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "err delete quiz set",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{ID: database.Text("quiz-set-id"), QuizExternalIDs: database.TextArray([]string{"another externalID"})}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("quiz-set-id")).
					Once().Return(pgx.ErrTxClosed)
			},
			req:         getQuizLO(validReq),
			expectedErr: status.Errorf(codes.Internal, "quizSetRepo.Delete: tx is closed"),
		},
		{
			ctx:  ctx,
			name: "succeed with new quizset",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, len(quizSet.QuestionHierarchy.Elements), 1)

					}).
					Return(nil)
			},
			req:          getQuizLO(validReq),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "succeed with upsert quizset",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, len(quizSet.QuestionHierarchy.Elements), 1)

					}).
					Return(nil)
			},
			req:          getQuizLO(validReq),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "succeed with upsert quizset with question group",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{
					{ID: database.Text("quiz-set-id"), LoID: database.Text("lo-id"), Status: database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()), QuizExternalIDs: database.TextArray([]string{}), QuestionHierarchy: database.JSONBArray([]interface{}{`{"id":"question group id","type":"QUESTION_GROUP"}`})},
				}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, len(quizSet.QuestionHierarchy.Elements), 1)

						var questionHierarchy entities.QuestionHierarchy
						questionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy)
						assert.Equal(t, len(questionHierarchy[0].ChildrenIDs), 1)
					}).
					Return(nil)
				questionGroup := &entities.QuestionGroup{}
				questionGroupRepo.On("GetByQuestionGroupIDAndLoID", ctx, db, database.Text("question group id"), database.Text("lo-id")).Once().Return(questionGroup, nil)
			},
			req:          getQuizLO(validQuestionGroupReq),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "success add lv1 quiz to quizset",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("quiz-set-id"),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuizExternalIDs:   database.TextArray([]string{"another externalID"}),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"another externalID\",\"type\":\"QUESTION\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, len(quizSet.QuestionHierarchy.Elements), 2)
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "success add lv2 quiz to quizset",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{
					{ID: database.Text("quiz-set-id"), LoID: database.Text("lo-id"), Status: database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()), QuizExternalIDs: database.TextArray([]string{"another externalID"}), QuestionHierarchy: database.JSONBArray([]interface{}{
						"{\"id\":\"another externalID\",\"type\":\"QUESTION\"}",
						"{\"id\": \"question group id\",\"type\":\"QUESTION_GROUP\"}",
					})},
				}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Equal(t, len(quizSet.QuestionHierarchy.Elements), 2)

						var questionHierarchy entities.QuestionHierarchy
						questionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy)
						assert.Equal(t, len(questionHierarchy[1].ChildrenIDs), 1)
					}).
					Return(nil)
				questionGroup := &entities.QuestionGroup{}
				questionGroupRepo.On("GetByQuestionGroupIDAndLoID", ctx, db, database.Text("question group id"), database.Text("lo-id")).Once().Return(questionGroup, nil)
			},
			req:          getQuizLO(validQuestionGroupReq),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "err get question group",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				questionGroupRepo.On("GetByQuestionGroupIDAndLoID", ctx, db, database.Text("question group id"), database.Text("lo-id")).Once().Return(nil, pgx.ErrTxClosed)
			},
			req:         getQuizLO(validQuestionGroupReq),
			expectedErr: status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
		},
		{
			ctx:  ctx,
			name: "insert ordering question successfully without existing quizSet",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), quiz.Kind.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 1)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ORD),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "insert ordering question successfully with existing quizSet",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"current-externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"current-externalID\",\"type\":\"QUESTION\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), quiz.Kind.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("current-quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 2)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"current-externalID", "externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ORD),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "insert ordering question which belong to question group successfully",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"current-externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"question group id\",\"type\":\"QUESTION_GROUP\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				questionGroupRepo.On("GetByQuestionGroupIDAndLoID", ctx, db, database.Text("question group id"), database.Text("lo-id")).
					Once().Return(&entities.QuestionGroup{
					QuestionGroupID:    database.Text("question group id"),
					LearningMaterialID: database.Text("lo-id"),
				}, nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), quiz.Kind.String)
						assert.Equal(t, database.Text("question group id"), quiz.QuestionGroupID)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("current-quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 1)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validQuestionGroupReq, cpb.QuizType_QUIZ_TYPE_ORD),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "update ordering question successfully without existing quizSet",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), quiz.Kind.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 1)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ORD),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "update ordering question successfully with existing quizSet",
			setup: func(ctx context.Context) {
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"current-externalID", "externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"current-externalID\",\"type\":\"QUESTION\"}", "{\"id\":\"externalID\",\"type\":\"QUESTION\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), quiz.Kind.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ORD),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "update ordering question successfully with existing quizSet but dont contain this question id",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"current-externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"current-externalID\",\"type\":\"QUESTION\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), quiz.Kind.String)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("current-quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 2)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"current-externalID", "externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ORD),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "update ordering question which belong to question group successfully",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set("question group id")
				quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"question group id\",\"type\":\"QUESTION_GROUP\", \"children_ids\": [\"externalID\"]}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				questionGroupRepo.On("GetByQuestionGroupIDAndLoID", ctx, db, database.Text("question group id"), database.Text("lo-id")).
					Once().Return(&entities.QuestionGroup{
					QuestionGroupID:    database.Text("question group id"),
					LearningMaterialID: database.Text("lo-id"),
				}, nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ORD.String(), quiz.Kind.String)
						assert.Equal(t, database.Text("question group id"), quiz.QuestionGroupID)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
			},
			req:          getQuizLO(validQuestionGroupReq, cpb.QuizType_QUIZ_TYPE_ORD),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "insert essay question successfully with existing quizSet",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"current-externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"current-externalID\",\"type\":\"QUESTION\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ESQ.String(), quiz.Kind.String)
						var optionsEnt []*entities.QuizOption
						json.Unmarshal(quiz.Options.Bytes, &optionsEnt)
						assert.Equal(t, true, optionsEnt[0].AnswerConfig.Essay.LimitEnabled)
						assert.Equal(t, entities.EssayLimitTypeCharacter, optionsEnt[0].AnswerConfig.Essay.LimitType)
						assert.Equal(t, uint32(5000), optionsEnt[0].AnswerConfig.Essay.Limit)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("current-quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 2)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"current-externalID", "externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ESQ),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "update essay question successfully with existing quizSet",
			setup: func(ctx context.Context) {
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
						AnswerConfig: entities.AnswerConfig{
							Essay: entities.EssayConfig{
								LimitEnabled: false,
								LimitType:    entities.EssayLimitTypeCharacter,
								Limit:        0,
							},
						},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ESQ.String())
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"current-externalID", "externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"current-externalID\",\"type\":\"QUESTION\"}", "{\"id\":\"externalID\",\"type\":\"QUESTION\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ESQ.String(), quiz.Kind.String)
						var optionsEnt []*entities.QuizOption
						json.Unmarshal(quiz.Options.Bytes, &optionsEnt)
						assert.Equal(t, true, optionsEnt[0].AnswerConfig.Essay.LimitEnabled)
						assert.Equal(t, entities.EssayLimitTypeCharacter, optionsEnt[0].AnswerConfig.Essay.LimitType)
						assert.Equal(t, uint32(5000), optionsEnt[0].AnswerConfig.Essay.Limit)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}

			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ESQ),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "insert essay question successfully without existing quizSet",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ESQ.String(), quiz.Kind.String)
						var optionsEnt []*entities.QuizOption
						json.Unmarshal(quiz.Options.Bytes, &optionsEnt)
						assert.Equal(t, true, optionsEnt[0].AnswerConfig.Essay.LimitEnabled)
						assert.Equal(t, entities.EssayLimitTypeCharacter, optionsEnt[0].AnswerConfig.Essay.LimitType)
						assert.Equal(t, uint32(5000), optionsEnt[0].AnswerConfig.Essay.Limit)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 1)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ESQ),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "update essay question successfully without existing quizSet",
			setup: func(ctx context.Context) {
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ESQ.String())
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ESQ.String(), quiz.Kind.String)
						var optionsEnt []*entities.QuizOption
						json.Unmarshal(quiz.Options.Bytes, &optionsEnt)
						assert.Equal(t, true, optionsEnt[0].AnswerConfig.Essay.LimitEnabled)
						assert.Equal(t, entities.EssayLimitTypeCharacter, optionsEnt[0].AnswerConfig.Essay.LimitType)
						assert.Equal(t, uint32(5000), optionsEnt[0].AnswerConfig.Essay.Limit)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 1)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ESQ),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "update essay question successfully with existing quizSet but dont contain this question id",
			setup: func(ctx context.Context) {
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set(nil)
				quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ORD.String())
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"current-externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"current-externalID\",\"type\":\"QUESTION\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ESQ.String(), quiz.Kind.String)
						var optionsEnt []*entities.QuizOption
						json.Unmarshal(quiz.Options.Bytes, &optionsEnt)
						assert.Equal(t, true, optionsEnt[0].AnswerConfig.Essay.LimitEnabled)
						assert.Equal(t, entities.EssayLimitTypeCharacter, optionsEnt[0].AnswerConfig.Essay.LimitType)
						assert.Equal(t, uint32(5000), optionsEnt[0].AnswerConfig.Essay.Limit)

					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("current-quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 2)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"current-externalID", "externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validReq, cpb.QuizType_QUIZ_TYPE_ESQ),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "insert essay question which belong to question group successfully",
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(nil, pgx.ErrNoRows)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"current-externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"question group id\",\"type\":\"QUESTION_GROUP\"}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				questionGroupRepo.On("GetByQuestionGroupIDAndLoID", ctx, db, database.Text("question group id"), database.Text("lo-id")).
					Once().Return(&entities.QuestionGroup{
					QuestionGroupID:    database.Text("question group id"),
					LearningMaterialID: database.Text("lo-id"),
				}, nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ESQ.String(), quiz.Kind.String)
						assert.Equal(t, database.Text("question group id"), quiz.QuestionGroupID)
						var optionsEnt []*entities.QuizOption
						json.Unmarshal(quiz.Options.Bytes, &optionsEnt)
						assert.Equal(t, true, optionsEnt[0].AnswerConfig.Essay.LimitEnabled)
						assert.Equal(t, entities.EssayLimitTypeCharacter, optionsEnt[0].AnswerConfig.Essay.LimitType)
						assert.Equal(t, uint32(5000), optionsEnt[0].AnswerConfig.Essay.Limit)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
				quizSetRepo.On("Delete", ctx, tx, database.Text("current-quiz-set-id")).
					Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.QuizSet")).
					Once().
					Run(func(args mock.Arguments) {
						quizSet := args[2].(*entities.QuizSet)
						assert.Equal(t, epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String(), quizSet.Status.String)
						assert.Equal(t, "lo-id", quizSet.LoID.String)
						assert.Len(t, quizSet.QuestionHierarchy.Elements, 1)
						assert.Equal(t, database.FromTextArray(quizSet.QuizExternalIDs), []string{"externalID"})
					}).
					Return(nil)
			},
			req:          getQuizLO(validQuestionGroupReq, cpb.QuizType_QUIZ_TYPE_ESQ),
			expectedResp: validResp,
		},
		{
			ctx:  ctx,
			name: "update essay question which belong to question group successfully",
			setup: func(ctx context.Context) {
				quiz := &entities.Quiz{}
				quiz.Options.Set([]*entities.QuizOption{
					{
						Content: entities.RichText{RenderedURL: "old url"},
					},
				})
				quiz.LoIDs.Set([]string{"lo-id"})
				quiz.QuestionGroupID.Set("question group id")
				quiz.Kind.Set(cpb.QuizType_QUIZ_TYPE_ESQ.String())
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
				quizRepo.On("GetByExternalID", ctx, db, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().
					Return(quiz, nil)
				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				quizSetRepo.On("Search", ctx, tx, repositories.QuizSetFilter{
					ObjectiveIDs: database.TextArray([]string{"lo-id"}),
					Status:       database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					Limit:        1,
				}).Once().Return(entities.QuizSets{{
					ID:                database.Text("current-quiz-set-id"),
					QuizExternalIDs:   database.TextArray([]string{"externalID"}),
					LoID:              database.Text("lo-id"),
					Status:            database.Text(epb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()),
					QuestionHierarchy: database.JSONBArray([]interface{}{"{\"id\":\"question group id\",\"type\":\"QUESTION_GROUP\", \"children_ids\": [\"externalID\"]}"}),
				}}, nil)
				quizRepo.On("DeleteByExternalID", ctx, tx, database.Text("externalID"), database.Int4(constant.ManabieSchool)).
					Once().Return(nil)
				questionGroupRepo.On("GetByQuestionGroupIDAndLoID", ctx, db, database.Text("question group id"), database.Text("lo-id")).
					Once().Return(&entities.QuestionGroup{
					QuestionGroupID:    database.Text("question group id"),
					LearningMaterialID: database.Text("lo-id"),
				}, nil)
				quizRepo.On("Create", ctx, tx, mock.AnythingOfType("*entities.Quiz")).
					Run(func(args mock.Arguments) {
						quiz := args[2].(*entities.Quiz)
						quiz.ID = database.Text("quiz-id")
						assert.Equal(t, cpb.QuizStatus_QUIZ_STATUS_APPROVED.String(), quiz.Status.String)
						assert.Equal(t, cpb.QuizType_QUIZ_TYPE_ESQ.String(), quiz.Kind.String)
						assert.Equal(t, database.Text("question group id"), quiz.QuestionGroupID)
						var optionsEnt []*entities.QuizOption
						json.Unmarshal(quiz.Options.Bytes, &optionsEnt)
						assert.Equal(t, true, optionsEnt[0].AnswerConfig.Essay.LimitEnabled)
						assert.Equal(t, entities.EssayLimitTypeCharacter, optionsEnt[0].AnswerConfig.Essay.LimitType)
						assert.Equal(t, uint32(5000), optionsEnt[0].AnswerConfig.Essay.Limit)
					}).
					Once().Return(nil)
				for url, content := range m {
					yasuoUploadModifierService.On("UploadHtmlContent", mock.Anything, &ypb.UploadHtmlContentRequest{
						Content: content,
					}, mock.Anything).Once().Return(&ypb.UploadHtmlContentResponse{
						Url: url,
					}, nil)
				}
			},
			req:          getQuizLO(validQuestionGroupReq, cpb.QuizType_QUIZ_TYPE_ESQ),
			expectedResp: validResp,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Test case: " + testCase.name)
			mdCtx := interceptors.NewIncomingContext(testCase.ctx)
			testCase.setup(mdCtx)

			resp, err := s.UpsertSingleQuiz(mdCtx, testCase.req.(*epb.UpsertSingleQuizRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedResp.(*epb.UpsertSingleQuizResponse), resp)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				tx,
				quizRepo,
				questionGroupRepo,
				yasuoUploadModifierService,
				yasuoUploadReaderService,
				quizSetRepo,
			)
		})
	}
}

func TestQuizModifierService_CreateRetryQuizTest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	quizRepo := new(mock_repositories.MockQuizRepo)
	shuffledQuizSetRepo := new(mock_repositories.MockShuffledQuizSetRepo)
	questionGroupRepo := &mock_repositories.MockQuestionGroupRepo{}

	s := QuizModifierService{
		DB:                  db,
		QuizRepo:            quizRepo,
		QuizSetRepo:         quizSetRepo,
		ShuffledQuizSetRepo: shuffledQuizSetRepo,
		QuestionGroupRepo:   questionGroupRepo,
	}

	loID := database.Text("mock-LO-id")
	studentID := database.Text("mock-student")
	studyPlanItemID := database.Text("mock-study-plan-item-id")

	pgQuizExternalIDs := database.TextArray([]string{"id-1", "id-2", "id-3, id-4", "id-5", "id-6, id-7"})

	shuffledQuestionHierarchy := make([]interface{}, 0)
	for _, extID := range pgQuizExternalIDs.Elements {
		shuffledQuestionHierarchy = append(shuffledQuestionHierarchy, &entities.QuestionHierarchyObj{
			ID:   extID.String,
			Type: entities.QuestionHierarchyQuestion,
		})
	}

	start := database.Timestamptz(timeutil.Now())

	shuffledQuizzes := []*entities.ShuffledQuizSet{
		{
			ID:                       database.Text("mock-shuffle-id"),
			Status:                   database.Text("mock-status"),
			RandomSeed:               database.Text("1631588078904363955"),
			StudyPlanItemID:          database.Text("mock-study-plan-item-id"),
			QuizExternalIDs:          database.TextArray([]string{"1", "2"}),
			TotalCorrectness:         database.Int4(2),
			StudentID:                database.Text("student-id"),
			SubmissionHistory:        database.JSONB("{}"),
			OriginalQuizSetID:        database.Text("origin-quiz-set-id"),
			OriginalShuffleQuizSetID: database.Text("origin-shuffle-id"),
		},
	}
	correctExternalIDs := database.TextArray([]string{"id-1", "id-2"})
	quizzes := getQuizzes(
		7,
		cpb.QuizType_QUIZ_TYPE_MCQ.String(),
		cpb.QuizType_QUIZ_TYPE_FIB.String(),
		cpb.QuizType_QUIZ_TYPE_POW.String(),
		cpb.QuizType_QUIZ_TYPE_TAD.String(),
		cpb.QuizType_QUIZ_TYPE_MIQ.String(),
		cpb.QuizType_QUIZ_TYPE_MAQ.String(),
		cpb.QuizType_QUIZ_TYPE_ORD.String(),
	)
	quizExternalIDs := []string{}
	for idx, quiz := range quizzes {
		quiz.ExternalID = database.Text(fmt.Sprintf("id-%d", idx+1))
		quizExternalIDs = append(quizExternalIDs, quiz.ExternalID.String)
	}

	quizzes[0].QuestionGroupID = database.Text("question-gr-id-1")
	quizzes[1].QuestionGroupID = database.Text("question-gr-id-2")

	shuffleID := idutil.ULIDNow()

	shuffledQuizExternalIDs := quizExternalIDs
	seed := time.Now().UTC().UnixNano()

	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(shuffledQuizExternalIDs), func(i, j int) {
		shuffledQuizExternalIDs[i], shuffledQuizExternalIDs[j] = shuffledQuizExternalIDs[j], shuffledQuizExternalIDs[i]
	})

	questionHierarchy := make([]interface{}, 0)
	for _, extID := range pgQuizExternalIDs.Elements {
		questionHierarchy = append(questionHierarchy, &entities.QuestionHierarchyObj{
			ID:   extID.String,
			Type: entities.QuestionHierarchyQuestion,
		})
	}

	quizzSet := &entities.QuizSet{
		ID:                database.Text(idutil.ULIDNow()),
		LoID:              loID,
		QuizExternalIDs:   database.TextArray(quizExternalIDs),
		Status:            database.Text("QUIZSET_STATUS_APPROVED"),
		UpdatedAt:         start,
		CreatedAt:         start,
		QuestionHierarchy: database.JSONBArray(questionHierarchy),
	}

	shuffledQuizSet := &entities.ShuffledQuizSet{
		ID:                database.Text(idutil.ULIDNow()),
		OriginalQuizSetID: quizzSet.ID,
		QuizExternalIDs:   database.TextArray(shuffledQuizExternalIDs),
		Status:            quizzSet.Status,
		RandomSeed:        database.Text(strconv.FormatInt(seed, 10)),
		UpdatedAt:         start,
		CreatedAt:         start,
	}

	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")

	questionGroups := entities.QuestionGroups{
		{
			BaseEntity:         entities.BaseEntity{},
			QuestionGroupID:    database.Text("question-gr-id-1"),
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
			QuestionGroupID:    database.Text("question-gr-id-2"),
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

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &epb.CreateRetryQuizTestRequest{
				LoId:            loID.String,
				StudentId:       studentID.String,
				SetId:           wrapperspb.String(shuffleID),
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						retry := getRetryQuizIDs(database.TextArray(quizzes.GetExternalIDs()), correctExternalIDs)
						assert.ElementsMatch(t, database.FromTextArray(ids), retry)
					}).
					Return(quizzes, nil).Once()
				shuffledQuizSetRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizzes, nil)
				shuffledQuizSetRepo.On("GetExternalIDsFromSubmissionHistory", mock.Anything, mock.Anything, mock.Anything, true).Once().Return(correctExternalIDs, nil)
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzSet, nil)
				shuffledQuizSetRepo.On("GetBySessionID", ctx, mock.Anything, mock.Anything).Return(&entities.ShuffledQuizSet{}, nil)
				shuffledQuizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet.ID, nil).Run(func(args mock.Arguments) {
					shuffledQuizSet := args.Get(2).(*entities.ShuffledQuizSet)

					var questionHierarchy entities.QuestionHierarchy
					questionHierarchy.UnmarshalJSONBArray(shuffledQuizSet.QuestionHierarchy)

					var expectedQuestionHierarchy entities.QuestionHierarchy
					expectedQuestionHierarchy.UnmarshalJSONBArray(database.JSONBArray(shuffledQuestionHierarchy))
					expectedQuestionHierarchy = expectedQuestionHierarchy.ExcludeQuestionIDs(database.FromTextArray(correctExternalIDs))

					assert.Equal(t, questionHierarchy, expectedQuestionHierarchy)
				})
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet, nil)
				quizRepo.On("GetByExternalIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzes, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			name: "missing loID",
			ctx:  ctx,
			req: &epb.CreateRetryQuizTestRequest{
				LoId:            "",
				StudentId:       studentID.String,
				SetId:           wrapperspb.String(shuffleID),
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have lo id").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing paging field",
			ctx:  ctx,
			req: &epb.CreateRetryQuizTestRequest{
				LoId:            loID.String,
				SetId:           wrapperspb.String(shuffleID),
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				SessionId:       idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have paging field").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "limit equals to zero",
			ctx:  ctx,
			req: &epb.CreateRetryQuizTestRequest{
				LoId:            loID.String,
				SetId:           wrapperspb.String(shuffleID),
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				SessionId:       idutil.ULIDNow(),
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						retry := getRetryQuizIDs(database.TextArray(quizzes.GetExternalIDs()), correctExternalIDs)
						assert.ElementsMatch(t, database.FromTextArray(ids), retry)
					}).
					Return(quizzes, nil).Once()
				shuffledQuizSetRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizzes, nil)
				shuffledQuizSetRepo.On("GetExternalIDsFromSubmissionHistory", mock.Anything, mock.Anything, mock.Anything, true).Once().Return(correctExternalIDs, nil)
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzSet, nil)
				shuffledQuizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet.ID, nil).Run(func(args mock.Arguments) {
					shuffledQuizSet := args.Get(2).(*entities.ShuffledQuizSet)

					var questionHierarchy entities.QuestionHierarchy
					questionHierarchy.UnmarshalJSONBArray(shuffledQuizSet.QuestionHierarchy)

					var expectedQuestionHierarchy entities.QuestionHierarchy
					expectedQuestionHierarchy.UnmarshalJSONBArray(database.JSONBArray(shuffledQuestionHierarchy))
					expectedQuestionHierarchy = expectedQuestionHierarchy.ExcludeQuestionIDs(database.FromTextArray(correctExternalIDs))

					assert.Equal(t, questionHierarchy, expectedQuestionHierarchy)
				})
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet, nil)
				quizRepo.On("GetByExternalIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzes, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			name: "offset equals to zero",
			ctx:  ctx,
			req: &epb.CreateRetryQuizTestRequest{
				LoId:            loID.String,
				SetId:           wrapperspb.String(shuffleID),
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("offset must be positive").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "create retry quizset successfully for quizset having questions and question groups",
			ctx:  ctx,
			req: &epb.CreateRetryQuizTestRequest{
				LoId:            loID.String,
				SetId:           wrapperspb.String(shuffleID),
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				SessionId:       idutil.ULIDNow(),
				Paging: &cpb.Paging{
					Limit: 0,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						retry := getRetryQuizIDs(database.TextArray(quizzes.GetExternalIDs()), correctExternalIDs)
						assert.ElementsMatch(t, database.FromTextArray(ids), retry)
					}).
					Return(quizzes, nil).Once()
				shuffledQuizzes[0].QuestionHierarchy.Set(entities.QuestionHierarchy{
					{
						ID:   "id-1",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:          "gr-2",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{"id-2", "id-3"},
					},
					{
						ID:   "id-4",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:   "id-5",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:   "id-6",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:          "gr-3",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{"id-7"},
					},
				})
				shuffledQuizSet.QuestionHierarchy.Set(entities.QuestionHierarchy{
					{
						ID:   "id-1",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:          "gr-2",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{"id-2", "id-3"},
					},
					{
						ID:   "id-4",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:   "id-5",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:   "id-6",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:          "gr-3",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{"id-7"},
					},
				})

				shuffledQuizSetRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizzes, nil)
				shuffledQuizSetRepo.On("GetExternalIDsFromSubmissionHistory", mock.Anything, mock.Anything, mock.Anything, true).Once().Return(correctExternalIDs, nil)
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzSet, nil)
				shuffledQuizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet.ID, nil).Run(func(args mock.Arguments) {
					shuffledQuizSet := args.Get(2).(*entities.ShuffledQuizSet)

					var questionHierarchy entities.QuestionHierarchy
					questionHierarchy.UnmarshalJSONBArray(shuffledQuizSet.QuestionHierarchy)

					var expectedQuestionHierarchy entities.QuestionHierarchy
					expectedQuestionHierarchy.UnmarshalJSONBArray(database.JSONBArray(shuffledQuestionHierarchy))
					expectedQuestionHierarchy = expectedQuestionHierarchy.ExcludeQuestionIDs(database.FromTextArray(correctExternalIDs))

					assert.Equal(t, questionHierarchy, expectedQuestionHierarchy)
				})
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet, nil)
				quizRepo.On("GetByExternalIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzes, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			name: "GetQuestionGroupsByIDs failed",
			ctx:  ctx,
			req: &epb.CreateRetryQuizTestRequest{
				LoId:            loID.String,
				SetId:           wrapperspb.String(shuffleID),
				StudentId:       studentID.String,
				StudyPlanItemId: studyPlanItemID.String,
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("getQuestionGroupByQuiz: QuestionGroupRepo.GetQuestionGroupsByIDs: error").Error()),
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						retry := getRetryQuizIDs(database.TextArray(quizzes.GetExternalIDs()), correctExternalIDs)
						assert.ElementsMatch(t, database.FromTextArray(ids), retry)
					}).
					Return(quizzes, nil).Once()
				shuffledQuizzes[0].QuestionHierarchy.Set(entities.QuestionHierarchy{
					{
						ID:   "id-1",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:          "gr-2",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{"id-2", "id-3"},
					},
					{
						ID:   "id-4",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:   "id-5",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:   "id-6",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:          "gr-3",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{"id-7"},
					},
				})
				shuffledQuizSet.QuestionHierarchy.Set(entities.QuestionHierarchy{
					{
						ID:   "id-1",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:          "gr-2",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{"id-2", "id-3"},
					},
					{
						ID:   "id-4",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:   "id-5",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:   "id-6",
						Type: entities.QuestionHierarchyQuestion,
					},
					{
						ID:          "gr-3",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{"id-7"},
					},
				})

				shuffledQuizSetRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizzes, nil)
				shuffledQuizSetRepo.On("GetExternalIDsFromSubmissionHistory", mock.Anything, mock.Anything, mock.Anything, true).Once().Return(correctExternalIDs, nil)
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzSet, nil)
				shuffledQuizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet.ID, nil).Run(func(args mock.Arguments) {
					shuffledQuizSet := args.Get(2).(*entities.ShuffledQuizSet)

					var questionHierarchy entities.QuestionHierarchy
					questionHierarchy.UnmarshalJSONBArray(shuffledQuizSet.QuestionHierarchy)

					var expectedQuestionHierarchy entities.QuestionHierarchy
					expectedQuestionHierarchy.UnmarshalJSONBArray(database.JSONBArray(shuffledQuestionHierarchy))
					expectedQuestionHierarchy = expectedQuestionHierarchy.ExcludeQuestionIDs(database.FromTextArray(correctExternalIDs))

					assert.Equal(t, questionHierarchy, expectedQuestionHierarchy)
				})
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet, nil)
				quizRepo.On("GetByExternalIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzes, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, db, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(entities.QuestionGroups{}, fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			res, err := s.CreateRetryQuizTest(testCase.ctx, testCase.req.(*epb.CreateRetryQuizTestRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedErr == nil {
				assert.NotNil(t, res)
				assert.Len(t, res.Items, len(quizzes))
				// check order options
				for _, quiz := range res.Items {
					if quiz.Core.Kind.String() != cpb.QuizType_QUIZ_TYPE_MCQ.String() &&
						quiz.Core.Kind.String() != cpb.QuizType_QUIZ_TYPE_MAQ.String() &&
						quiz.Core.Kind.String() != cpb.QuizType_QUIZ_TYPE_ORD.String() {
						// compare label
						for i, opt := range quiz.Core.Options {
							assert.Equal(t, expectedOrderingQuestionKeys[i], opt.Key)
						}
					} else {
						actualKey := make([]string, 0, 3)
						count := 0
						for i, opt := range quiz.Core.Options {
							actualKey = append(actualKey, opt.Key)
							// have at least 2 items was shuffled
							if expectedOrderingQuestionKeys[i] != opt.Key {
								count++
							}
						}
						// TODO: Still have risk to list option was not shuffled,
						// plz recheck later
						//assert.LessOrEqual(t, 2, count, fmt.Errorf("must have at least 2 items was shuffled"))
						assert.ElementsMatch(t, actualKey, expectedOrderingQuestionKeys)
					}
				}
			}
			mock.AssertExpectationsForObjects(t, quizSetRepo, shuffledQuizSetRepo, quizRepo)
		})

	}
}

func TestCourseModifier_RemoveQuizFromLO(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	quizRepo := new(mock_repositories.MockQuizRepo)
	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	shuffledQuizSetRepo := new(mock_repositories.MockShuffledQuizSetRepo)
	studyPlanRepo := new(mock_repositories.MockStudyPlanRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	db.On("Begin").Once().Return(tx, nil)
	s := &QuizModifierService{
		DB:                  db,
		QuizRepo:            quizRepo,
		QuizSetRepo:         quizSetRepo,
		ShuffledQuizSetRepo: shuffledQuizSetRepo,
		StudyPlanRepo:       studyPlanRepo,
	}

	loID := "lo that need to remove quiz"
	quizID := idutil.ULIDNow()
	quizSetBelongToLOID1 := entities.QuizSet{
		QuizExternalIDs: database.TextArray([]string{"another quiz id", "and other quiz id", quizID}),
		QuestionHierarchy: database.JSONBArray([]interface{}{
			&entities.QuestionHierarchyObj{
				ID:   "another quiz id",
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   "and other quiz id",
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   quizID,
				Type: entities.QuestionHierarchyQuestion,
			},
		}),
	}
	quizSetBelongToLOID2 := entities.QuizSet{
		QuizExternalIDs: database.TextArray([]string{"other quiz id", quizID, "other quiz again"}),
		QuestionHierarchy: database.JSONBArray([]interface{}{
			&entities.QuestionHierarchyObj{
				ID:   "other quiz id",
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   quizID,
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   "other quiz again",
				Type: entities.QuestionHierarchyQuestion,
			},
		}),
	}

	quizSetBelongToLOID3 := entities.QuizSet{
		QuizExternalIDs: database.TextArray([]string{"other quiz id", quizID, "quiz-belong-to-group-1", "quiz-belong-to-group-2"}),
		QuestionHierarchy: database.JSONBArray([]interface{}{
			&entities.QuestionHierarchyObj{
				ID:   "other quiz id",
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   quizID,
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:          "group-id",
				Type:        entities.QuestionHierarchyQuestionGroup,
				ChildrenIDs: []string{"quiz-belong-to-group-1", "quiz-belong-to-group-2"},
			},
		}),
	}
	quizSets := entities.QuizSets{&quizSetBelongToLOID1, &quizSetBelongToLOID2, &quizSetBelongToLOID3}

	testCases := []TestCase{
		{
			name: "happy case remove quiz from lo successfully",
			ctx:  ctx,
			req: &epb.RemoveQuizFromLORequest{
				QuizId: quizID,
				LoId:   loID,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				studyPlanRepo.On("RetrieveStudyPlanItemInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.StudyPlanItemInfo{
					{},
				}, nil)
				shuffledQuizSetRepo.On("GetByStudyPlanItems", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.ShuffledQuizSets{}, nil)
				quizSetRepo.On("GetQuizSetsOfLOContainQuiz", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizSets, nil)
				quizSetRepo.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				quizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "missing quiz id in the request",
			ctx:  ctx,
			req: &epb.RemoveQuizFromLORequest{
				LoId: loID,
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have quiz id").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing lo id in the request",
			ctx:  ctx,
			req: &epb.RemoveQuizFromLORequest{
				QuizId: quizID,
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have LO id").Error()),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.RemoveQuizFromLO(testCase.ctx, testCase.req.(*epb.RemoveQuizFromLORequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestQuizModifierService_CreateFlashCardStudy(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}

	req := &epb.CreateFlashCardStudyRequest{
		StudyPlanItemId: "study-plan_item_id",
		LoId:            "lo-id",
		StudentId:       "student-id",
		KeepOrder:       false,
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	}
	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	quizRepo := new(mock_repositories.MockQuizRepo)
	flashcardProgressionRepo := new(mock_repositories.MockFlashcardProgressionRepo)
	s := QuizModifierService{
		DB:                       db,
		QuizRepo:                 quizRepo,
		QuizSetRepo:              quizSetRepo,
		FlashcardProgressionRepo: flashcardProgressionRepo,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         ctx,
			req:         req,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(&entities.QuizSet{ID: database.Text("quiz-set-id")}, nil)
				flashcardProgressionRepo.On("Create", ctx, db, mock.Anything).Once().Return(database.Text("study-set-id"), nil)
				flashcardProgressionRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entities.FlashcardProgression{}, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, db, mock.Anything, mock.Anything).Once().Return(entities.Quizzes{}, nil)
			},
		},
		{
			name:        "err quizSetRepo.GetQuizSetByLoID",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("CreateFlashCardStudy.FlashcardProgressionRepo.GetQuizSetByLoID: %v", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "err flashcardProgressionRepo.Create",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("CreateFlashCardStudy.FlashcardProgressionRepo.Create: %v", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(&entities.QuizSet{ID: database.Text("quiz-set-id")}, nil)
				flashcardProgressionRepo.On("Create", ctx, db, mock.Anything).Once().Return(database.Text("study-set-id"), ErrSomethingWentWrong)
			},
		},
		{
			name:        "err flashcardProgressionRepo.Get",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.FlashcardProgressionRepo.Get: %v", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(&entities.QuizSet{ID: database.Text("quiz-set-id")}, nil)
				flashcardProgressionRepo.On("Create", ctx, db, mock.Anything).Once().Return(database.Text("study-set-id"), nil)
				flashcardProgressionRepo.On("Get", ctx, db, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "err quizRepo.GetByExternalIDsAndLmID",
			ctx:         ctx,
			req:         req,
			expectedErr: status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.QuizRepo.GetByExternalIDsAndLmID: %v", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", ctx, db, mock.Anything, mock.Anything).Once().Return(&entities.QuizSet{ID: database.Text("quiz-set-id")}, nil)
				flashcardProgressionRepo.On("Create", ctx, db, mock.Anything).Once().Return(database.Text("study-set-id"), nil)
				flashcardProgressionRepo.On("Get", ctx, db, mock.Anything).Once().Return(&entities.FlashcardProgression{}, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, db, mock.Anything, mock.Anything).Once().Return(entities.Quizzes{}, ErrSomethingWentWrong)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.CreateFlashCardStudy(testCase.ctx, testCase.req.(*epb.CreateFlashCardStudyRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})

	}
}

func TestQuizModifierService_validateRemoveQuizFromLORequest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &epb.RemoveQuizFromLORequest{
				LoId:   "lo-id",
				QuizId: "quiz-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "req must have LO id",
			ctx:  ctx,
			req: &epb.RemoveQuizFromLORequest{
				QuizId: "quiz-id",
			},
			expectedErr: fmt.Errorf("req must have LO id"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "req must have quiz id",
			ctx:  ctx,
			req: &epb.RemoveQuizFromLORequest{
				LoId: "lo-id",
			},
			expectedErr: fmt.Errorf("req must have quiz id"),
			setup: func(ctx context.Context) {
			},
		},
	}
	s := QuizModifierService{}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.validateRemoveQuizFromLORequest(testCase.req.(*epb.RemoveQuizFromLORequest))
			assert.Equal(t, testCase.expectedErr, err)
		})

	}
}

func TestQuizModifierService_removeQuizFromQuizSets(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tx := &mock_database.Tx{}
	quizSetRepo := new(mock_repositories.MockQuizSetRepo)

	quizID1 := "quiz-id"
	quizID2 := "quiz-id-2"

	quizSets := []*entities.QuizSet{{
		ID:              database.Text("id"),
		QuizExternalIDs: database.TextArray([]string{quizID1, quizID2}),
		QuestionHierarchy: database.JSONBArray([]interface{}{
			&entities.QuestionHierarchyObj{
				ID:   quizID1,
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   quizID2,
				Type: entities.QuestionHierarchyQuestion,
			},
		}),
	}}

	expectedQuizSet := &entities.QuizSet{
		ID:              database.Text("id"),
		QuizExternalIDs: database.TextArray([]string{quizID2}),
		QuestionHierarchy: database.JSONBArray([]interface{}{
			&entities.QuestionHierarchyObj{
				ID:   quizID2,
				Type: entities.QuestionHierarchyQuestion,
			},
		}),
	}

	expectedQuestionGroupContainingQuizSet := &entities.QuizSet{
		ID:              database.Text("id"),
		QuizExternalIDs: database.TextArray([]string{"quiz-id-3", quizID2, "quiz-id-4"}),
		QuestionHierarchy: database.JSONBArray([]interface{}{
			&entities.QuestionHierarchyObj{
				ID:   "quiz-id-3",
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:   quizID2,
				Type: entities.QuestionHierarchyQuestion,
			},
			&entities.QuestionHierarchyObj{
				ID:          "group-1",
				Type:        entities.QuestionHierarchyQuestionGroup,
				ChildrenIDs: []string{"quiz-id-4"},
			},
		}),
	}
	s := QuizModifierService{
		QuizSetRepo: quizSetRepo,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			expectedResp: entities.QuizSets{
				{
					ID:              database.Text("id"),
					QuizExternalIDs: database.TextArray([]string{quizID2}),
					QuestionHierarchy: database.JSONBArray([]interface{}{
						&entities.QuestionHierarchyObj{
							ID:   quizID2,
							Type: entities.QuestionHierarchyQuestion,
						},
					}),
				},
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("Delete", ctx, tx, database.Text("id")).Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, expectedQuizSet).Once().Return(nil)
			},
		},
		{
			name:        "err quizSetRepo.Delete",
			ctx:         ctx,
			expectedErr: fmt.Errorf("QuizSetRepo.Delete:%w", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				quizSetRepo.On("Delete", ctx, tx, database.Text("id")).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "err quizSetRepo.Create",
			ctx:         ctx,
			expectedErr: fmt.Errorf("QuizSetRepo.Create:%w", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				quizSetRepo.On("Delete", ctx, tx, database.Text("id")).Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, expectedQuizSet).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:         "remove successfully when quizsets having question groups",
			ctx:          ctx,
			expectedErr:  nil,
			expectedResp: entities.QuizSets{expectedQuestionGroupContainingQuizSet},
			setup: func(ctx context.Context) {
				quizSets[0].QuizExternalIDs = database.TextArray([]string{"quiz-id-3", quizID2, quizID1, "quiz-id-4"})
				quizSets[0].QuestionHierarchy = database.JSONBArray([]interface{}{
					&entities.QuestionHierarchyObj{
						ID:   "quiz-id-3",
						Type: entities.QuestionHierarchyQuestion,
					},
					&entities.QuestionHierarchyObj{
						ID:   quizID2,
						Type: entities.QuestionHierarchyQuestion,
					},
					&entities.QuestionHierarchyObj{
						ID:          "group-1",
						Type:        entities.QuestionHierarchyQuestionGroup,
						ChildrenIDs: []string{quizID1, "quiz-id-4"},
					},
				})
				quizSetRepo.On("Delete", ctx, tx, database.Text("id")).Once().Return(nil)
				quizSetRepo.On("Create", ctx, tx, expectedQuestionGroupContainingQuizSet).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.removeQuizFromQuizSets(testCase.ctx, tx, quizID1, quizSets)

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
			mock.AssertExpectationsForObjects(t, tx, quizSetRepo)
		})

	}
}

func TestCourseModifierService_UpdateDisplayOrderOfQuizSet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	quizSetRepo := &mock_repositories.MockQuizSetRepo{}

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	s := &QuizModifierService{
		DB:          mockDB,
		QuizSetRepo: quizSetRepo,
	}

	loID := "lo_id_1"
	quizExternalIDs := []string{"quiz_external_id_1", "quiz_external_id_2", "quiz_external_id_3", "quiz_external_id_4", "quiz_external_id_5"}
	quizSet := &entities.QuizSet{
		ID:              database.Text(idutil.ULIDNow()),
		LoID:            database.Text(loID),
		QuizExternalIDs: database.TextArray(quizExternalIDs),
	}
	pairs := []*epb.UpdateDisplayOrderOfQuizSetRequest_QuizExternalIDPair{
		{First: "quiz_external_id_1", Second: "quiz_external_id_2"},
		{First: "quiz_external_id_2", Second: "quiz_external_id_4"},
	}
	notExistPair := []*epb.UpdateDisplayOrderOfQuizSetRequest_QuizExternalIDPair{
		{First: "not_exist_quiz_external_id_1", Second: "not_exist_quiz_external_id_2"},
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &epb.UpdateDisplayOrderOfQuizSetRequest{
				LoId:  loID,
				Pairs: pairs,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSet.ID).Once().Return(nil)
				quizSetRepo.On("Create", mock.Anything, mockTxer, quizSet).Once().Return(nil)
				mockTxer.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "Err when quiz_external_id not exist",
			ctx:  ctx,
			req: &epb.UpdateDisplayOrderOfQuizSetRequest{
				LoId:  loID,
				Pairs: notExistPair,
			},
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, s.DB, database.Text(loID)).Once().Return(quizSet, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				quizSetRepo.On("Delete", mock.Anything, mockTxer, quizSet.ID).Once().Return(nil)
				quizSetRepo.On("Create", mock.Anything, mockTxer, quizSet).Once().Return(nil)
				mockTxer.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "UpdateDisplayOrderOfQuizSet QuizExternalID %v is not exist in quiz set", notExistPair[0].First),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*epb.UpdateDisplayOrderOfQuizSetRequest)
			_, err := s.UpdateDisplayOrderOfQuizSet(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
