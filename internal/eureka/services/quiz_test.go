package services

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/s3"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	yasuo_entities "github.com/manabie-com/backend/internal/yasuo/entities"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var expectedOrderingQuestionKeys = []string{"key-1", "key-2", "key-3"}

func TestCreateQuizTestV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}
	questionGroupRepo := &mock_repositories.MockQuestionGroupRepo{}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupSchoolAdmin)

	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	quizRepo := new(mock_repositories.MockQuizRepo)
	shuffledQuizSetRepo := new(mock_repositories.MockShuffledQuizSetRepo)
	learningTimeCalculator := &LearningTimeCalculator{}

	s := QuizService{
		DB:                        mockDB,
		QuizRepo:                  quizRepo,
		QuizSetRepo:               quizSetRepo,
		ShuffledQuizSetRepo:       shuffledQuizSetRepo,
		QuestionGroupRepo:         questionGroupRepo,
		LearningTimeCalculatorSvc: learningTimeCalculator,
	}

	loID := database.Text("VN10-CH-01-L-001.1")
	studentID := database.Text("STUDENT_C1")
	studyPlanID := database.Text("STUDY_PLAN_ID_C1")
	start := database.Timestamptz(timeutil.Now())
	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")

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

	quizzSet := &entities.QuizSet{
		ID:              database.Text(idutil.ULIDNow()),
		LoID:            loID,
		QuizExternalIDs: database.TextArray(quizExternalIDs),
		Status:          database.Text("QUIZSET_STATUS_APPROVED"),
		UpdatedAt:       start,
		CreatedAt:       start,
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

	respQuestionGroups, _ := entities.QuestionGroupsToQuestionGroupProtoBufMess(questionGroups)
	respNilQuestionGroups, _ := entities.QuestionGroupsToQuestionGroupProtoBufMess(nil)
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &sspb.CreateQuizTestV2Request{

				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr:  nil,
			expectedResp: respQuestionGroups,
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzSet, nil)
				shuffledQuizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet.ID, nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet, nil)
				quizRepo.On("GetByExternalIDsAndLmID", mock.Anything, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						assert.ElementsMatch(t, database.FromTextArray(ids), quizzes.GetExternalIDs())
					}).Twice().Return(quizzes, nil)
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, mockDB, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			name: "missing loID",
			ctx:  ctx,
			req: &sspb.CreateQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: "",
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have learning material id").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing paging field",
			ctx:  ctx,
			req: &sspb.CreateQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have paging field").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "limit equals to zero",
			ctx:  ctx,
			req: &sspb.CreateQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
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
			req: &sspb.CreateQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
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
			req: &sspb.CreateQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
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
			req: &sspb.CreateQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				Paging: &cpb.Paging{
					Limit: 3,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1000,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr:  nil,
			expectedResp: respQuestionGroups,
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzSet, nil)
				shuffledQuizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet.ID, nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						assert.ElementsMatch(t, database.FromTextArray(ids), quizzes.GetExternalIDs())
					}).
					Return(quizzes, nil).Twice()
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, mockDB, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(questionGroups, nil)
			},
		},
		{
			name: "quiz have question group id not exist",
			ctx:  ctx,
			req: &sspb.CreateQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				Paging: &cpb.Paging{
					Limit: 3,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1000,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedResp: respNilQuestionGroups,
			setup: func(ctx context.Context) {
				quizSetRepo.On("GetQuizSetByLoID", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(quizzSet, nil)
				shuffledQuizSetRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet.ID, nil)
				shuffledQuizSetRepo.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(shuffledQuizSet, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, mock.Anything, mock.Anything, loID).
					Run(func(args mock.Arguments) {
						ids := args[2].(pgtype.TextArray)
						assert.ElementsMatch(t, database.FromTextArray(ids), quizzes.GetExternalIDs())
					}).
					Return(quizzes, nil).Twice()
				questionGroupRepo.
					On("GetQuestionGroupsByIDs", ctx, mockDB, []string{"question-gr-id-1", "question-gr-id-2"}).
					Once().
					Return(entities.QuestionGroups{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			res, err := s.CreateQuizTestV2(testCase.ctx, testCase.req.(*sspb.CreateQuizTestV2Request))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedErr == nil {
				assert.NotNil(t, res)
				assert.ElementsMatch(t, testCase.expectedResp, res.QuestionGroups)
				assert.Len(t, res.Quizzes, len(quizzes))

				// check order options
				for _, quiz := range res.Quizzes {
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
			mock.AssertExpectationsForObjects(t, mockDB, quizSetRepo, shuffledQuizSetRepo, quizRepo, questionGroupRepo)
		})
	}
}

func TestQuizService_CreateRetryQuizTestV2(t *testing.T) {
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
	learningTimeCalculator := &LearningTimeCalculator{}

	s := QuizService{
		QuizRepo:                  quizRepo,
		QuizSetRepo:               quizSetRepo,
		ShuffledQuizSetRepo:       shuffledQuizSetRepo,
		LearningTimeCalculatorSvc: learningTimeCalculator,
		QuestionGroupRepo:         questionGroupRepo,
		DB:                        db,
	}

	loID := database.Text("mock-LO-id")
	studentID := database.Text("mock-student")
	studyPlanID := database.Text("mock-study-plan-id")
	pgQuizExternalIDs := database.TextArray([]string{"id-1", "id-2", "id-3"})

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
			QuizExternalIDs:          pgQuizExternalIDs,
			TotalCorrectness:         database.Int4(2),
			StudentID:                database.Text("student-id"),
			SubmissionHistory:        database.JSONB("{}"),
			OriginalQuizSetID:        database.Text("origin-quiz-set-id"),
			OriginalShuffleQuizSetID: database.Text("origin-shuffle-id"),
			QuestionHierarchy:        database.JSONBArray(shuffledQuestionHierarchy),
		},
	}
	correctExternalIDs := database.TextArray([]string{"id-1", "id-2"})
	quizzes := getQuizzes(3)
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
		QuestionHierarchy: database.JSONBArray(questionHierarchy),
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
			req: &sspb.CreateRetryQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				ShuffleQuizSetId: wrapperspb.String(shuffleID),
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
			req: &sspb.CreateRetryQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: "",
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				ShuffleQuizSetId: wrapperspb.String(shuffleID),
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have learning material id").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing paging field",
			ctx:  ctx,
			req: &sspb.CreateRetryQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				ShuffleQuizSetId: wrapperspb.String(shuffleID),
				SessionId:        idutil.ULIDNow(),
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have paging field").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "limit equals to zero",
			ctx:  ctx,
			req: &sspb.CreateRetryQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				ShuffleQuizSetId: wrapperspb.String(shuffleID),
				SessionId:        idutil.ULIDNow(),
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
			req: &sspb.CreateRetryQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				ShuffleQuizSetId: wrapperspb.String(shuffleID),
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
			req: &sspb.CreateRetryQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				ShuffleQuizSetId: wrapperspb.String(shuffleID),
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
			req: &sspb.CreateRetryQuizTestV2Request{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: loID.String,
					StudentId:          wrapperspb.String(studentID.String),
					StudyPlanId:        studyPlanID.String,
				},
				ShuffleQuizSetId: wrapperspb.String(shuffleID),
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
				SessionId: idutil.ULIDNow(),
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("getQuestionGroupByQuiz: QuestionGroupRepo.GetQuestionGroupsByIDs: error").Error()),
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
			res, err := s.CreateRetryQuizTestV2(testCase.ctx, testCase.req.(*sspb.CreateRetryQuizTestV2Request))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedErr == nil {
				assert.NotNil(t, res)
				assert.Len(t, res.Quizzes, len(quizzes))

				// check order options
				for _, quiz := range res.Quizzes {
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
		})

	}
}

func TestRetrieveQuizTestsV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, entities.UserGroupTeacher)

	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	quizRepo := new(mock_repositories.MockQuizRepo)
	shuffledQuizSetRepo := new(mock_repositories.MockShuffledQuizSetRepo)
	studentEventLogRepo := new(mock_repositories.MockStudentEventLogRepo)
	learningTimeCalculator := &LearningTimeCalculator{}

	s := QuizService{
		QuizRepo:                  quizRepo,
		QuizSetRepo:               quizSetRepo,
		ShuffledQuizSetRepo:       shuffledQuizSetRepo,
		LearningTimeCalculatorSvc: learningTimeCalculator,
		StudentEventLogRepo:       studentEventLogRepo,
	}

	quizTests := entities.ShuffledQuizSets{}
	quizTests.Add()
	quizTests.Add()
	quizTests.Add()
	logs := make([]*entities.StudentEventLog, 0)

	studyPlanItemIdentities := []*sspb.StudyPlanItemIdentity{
		{
			StudentId:          wrapperspb.String("student_id"),
			StudyPlanId:        "study_plan_id",
			LearningMaterialId: "learning_material_id",
		},
	}
	testCases := []TestCase{
		{
			name: "happy case retrieve quiz tests",
			ctx:  ctx,
			req: &sspb.RetrieveQuizTestV2Request{
				StudyPlanItemIdentities: studyPlanItemIdentities,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.On("GetByStudyPlanItemIdentities", ctx, mock.Anything, mock.Anything).Once().Return(quizTests, nil)
				studentEventLogRepo.On("RetrieveStudentEventLogsByStudyPlanIdentities", ctx, mock.Anything, mock.Anything).Return(logs, nil)
			},
		},
		{
			name: "missing study plan item Identity",
			ctx:  ctx,
			req: &sspb.RetrieveQuizTestV2Request{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("req must have Study Plan Item Identity").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing study plan item Identity.StudentId",
			ctx:  ctx,
			req: &sspb.RetrieveQuizTestV2Request{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("StudyPlanItemIdentities[0] is error req must have student id").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing study plan item Identity.LearningMaterialId",
			ctx:  ctx,
			req: &sspb.RetrieveQuizTestV2Request{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{StudentId: wrapperspb.String("student_id")},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("StudyPlanItemIdentities[0] is error req must have learning material id").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing study plan item Identity.StudyPlanId",
			ctx:  ctx,
			req: &sspb.RetrieveQuizTestV2Request{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{StudentId: wrapperspb.String("student_id"), LearningMaterialId: "learning_material_id"},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, errors.New("StudyPlanItemIdentities[0] is error req must have study plan id").Error()),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)

		_, err := s.RetrieveQuizTestsV2(testCase.ctx, testCase.req.(*sspb.RetrieveQuizTestV2Request))

		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestQuizService_UpsertFlashcardContent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	quizRepo := &mock_repositories.MockQuizRepo{}
	quizSetRepo := &mock_repositories.MockQuizSetRepo{}
	speechesRepo := &mock_repositories.MockSpeechesRepository{}
	mediaModifierService := &mock_services.BobMediaModifierServiceClient{}
	yasuoUploadReaderService := &mock_services.YasuoUploadReaderServiceClient{}
	yasuoUploadModifierService := &mock_services.YasuoUploadModifierServiceClient{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	svc := QuizService{
		QuizRepo:    quizRepo,
		QuizSetRepo: quizSetRepo,

		SpeechesRepo:        speechesRepo,
		BobMediaModifier:    mediaModifierService,
		YasuoUploadReader:   yasuoUploadReaderService,
		YasuoUploadModifier: yasuoUploadModifierService,
		DB:                  db,
	}

	question := "rendered " + idutil.ULIDNow()
	explanation := "rendered " + idutil.ULIDNow()
	option := "rendered " + idutil.ULIDNow()
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "thanh_dinh"),
			req: &sspb.UpsertFlashcardContentRequest{
				Quizzes: []*cpb.QuizCore{
					{
						ExternalId: "external_id",
						Kind:       cpb.QuizType_QUIZ_TYPE_POW,
						Info: &cpb.ContentBasicInfo{
							SchoolId: 1,
							Country:  cpb.Country_COUNTRY_VN,
						},
						Question: &cpb.RichText{
							Raw:      "raw",
							Rendered: question,
						},
						Attribute: &cpb.QuizItemAttribute{
							AudioLink: "link",
							Configs:   []cpb.QuizItemAttributeConfig{cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG},
						},
						Explanation: &cpb.RichText{
							Raw:      "raw",
							Rendered: explanation,
						},
						TaggedLos:       []string{"123", "abc"},
						DifficultyLevel: 2,
						Options: []*cpb.QuizOption{
							{
								Content: &cpb.RichText{
									Raw:      "raw",
									Rendered: option,
								},
								Correctness: true,
								Label:       "(1)",
								Key:         idutil.ULIDNow(),
								Attribute: &cpb.QuizItemAttribute{
									Configs: []cpb.QuizItemAttributeConfig{cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_JP},
								},
							},
						},
					},
				},
				FlashcardId: "flashcard_id",
				Kind:        cpb.QuizType_QUIZ_TYPE_POW,
			},
			setup: func(ctx context.Context) {
				quizRepo.On("GetByExternalIDsAndLmID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.Quizzes{}, nil)

				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.AnythingOfType("*context.valueCtx")).Return(nil)
				tx.On("Rollback", mock.AnythingOfType("*context.valueCtx")).Return(nil)

				quizSetRepo.On("RetrieveByLoIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.QuizSets{}, nil)

				yasuoUploadReaderService.On("RetrieveUploadInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.RetrieveUploadInfoResponse{}, nil)
				url1, _ := generateUploadURL("", "", question)
				url2, _ := generateUploadURL("", "", explanation)
				url3, _ := generateUploadURL("", "", option)
				m := make(map[string]string)
				m[url1] = question
				m[url2] = explanation
				m[url3] = option
				urls := []string{url1, url2, url3}

				yasuoUploadModifierService.On("BulkUploadHtmlContent", mock.Anything, mock.Anything, mock.Anything).Once().Return(&ypb.BulkUploadHtmlContentResponse{
					Urls: urls,
				}, nil)

				quizSetRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(entities.QuizSets{}, nil)
				quizRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.Quiz{
					{
						ID:          database.Text("quiz_id"),
						ExternalID:  database.Text("this is test"),
						Country:     database.Text("this is test"),
						SchoolID:    database.Int4(10),
						LoIDs:       database.TextArray([]string{"this is test"}),
						Kind:        database.Text("this is test"),
						Question:    database.JSONB("{}"),
						Explanation: database.JSONB("{}"),
						Options: database.JSONB([]*entities.QuizOption{
							{
								Content: entities.RichText{
									Raw:         "raw",
									RenderedURL: "render",
								},
								Correctness: true,
								Label:       "lable",
								Configs:     []string{},
								Key:         "key",
								Attribute: entities.QuizItemAttribute{
									AudioLink: "link",
									ImgLink:   "link",
									Configs:   []string{},
								},
							},
						}),
						TaggedLOs:      database.TextArray([]string{"this is test"}),
						DifficultLevel: database.Int4(10),
						CreatedBy:      database.Text("this is test"),
						ApprovedBy:     database.Text("this is test"),
						Status:         database.Text("this is test"),
						UpdatedAt:      database.Timestamptz(time.Now()),
						CreatedAt:      database.Timestamptz(time.Now()),
					},
				}, nil)

				mediaModifierService.On("GenerateAudioFile", mock.Anything, mock.Anything).Once().Return(&bpb.GenerateAudioFileResponse{
					Options: []*bpb.AudioOptionResponse{
						{
							Link:   "link",
							QuizId: "quiz_id",
							Text:   "text",
							Type:   bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_TERM,
						},
					},
				}, nil)

				speechesRepo.On("UpsertSpeeches", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*yasuo_entities.Speeches{
					{
						Speeches: bob_entities.Speeches{
							SpeechID: database.Text("this is test 3"),
							Link:     database.Text("this is test 3"),
							Type:     database.Text(bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_TERM.String()),
							QuizID:   database.Text("quiz_id"),
							Sentence: database.Text("this is test 3"),
							Settings: database.JSONB("{}"),
						},
					},
				}, nil)
				quizRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.Quiz{
					{
						ID:             database.Text("quiz_id"),
						ExternalID:     database.Text("this is test"),
						Country:        database.Text("this is test"),
						SchoolID:       database.Int4(10),
						LoIDs:          database.TextArray([]string{"this is test"}),
						Kind:           database.Text("this is test"),
						Question:       database.JSONB("{}"),
						Explanation:    database.JSONB("{}"),
						Options:        database.JSONB("{}"),
						TaggedLOs:      database.TextArray([]string{"this is test"}),
						DifficultLevel: database.Int4(10),
						CreatedBy:      database.Text("this is test"),
						ApprovedBy:     database.Text("this is test"),
						Status:         database.Text("this is test"),
						UpdatedAt:      database.Timestamptz(time.Now()),
						CreatedAt:      database.Timestamptz(time.Now()),
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Test case: " + testCase.name)
			mdCtx := interceptors.NewIncomingContext(testCase.ctx)
			testCase.setup(mdCtx)

			_, err := svc.UpsertFlashcardContent(mdCtx, testCase.req.(*sspb.UpsertFlashcardContentRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_validateCheckQuizCorrectnessRequest(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Empty ShuffledQuizSetId",
			req: &sspb.CheckQuizCorrectnessRequest{
				ShuffledQuizSetId: "",
				QuizId:            "quiz_id",
			},
			expectedErr: fmt.Errorf("req must have ShuffledQuizSetId"),
		},
		{
			name: "Empty ShuffledQuizSetId",
			req: &sspb.CheckQuizCorrectnessRequest{
				ShuffledQuizSetId: "shuffled_quiz_id",
				QuizId:            "",
			},
			expectedErr: fmt.Errorf("req must have QuizId"),
		},
		{
			name: "Empty Answer",
			req: &sspb.CheckQuizCorrectnessRequest{
				ShuffledQuizSetId: "shuffled_quiz_id",
				QuizId:            "quiz_id",
			},
			expectedErr: fmt.Errorf("req must have Answer"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateCheckQuizCorrectnessRequest(testCase.req.(*sspb.CheckQuizCorrectnessRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func Test_CheckQuizCorrectness(t *testing.T) {
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	loSubmissionAnswerRepo := &mock_repositories.MockLOSubmissionAnswerRepo{}
	flashCardSubmissionAnswerRepo := &mock_repositories.MockFlashCardSubmissionAnswerRepo{}
	studentsLearningObjectivesCompletenessRepo := &mock_repositories.MockStudentsLearningObjectivesCompletenessRepo{}

	svc := &QuizService{
		DB:                            mockDB,
		QuizRepo:                      quizRepo,
		ShuffledQuizSetRepo:           shuffledQuizSetRepo,
		LOSubmissionAnswerRepo:        loSubmissionAnswerRepo,
		FlashCardSubmissionAnswerRepo: flashCardSubmissionAnswerRepo,
		StudentsLearningObjectivesCompletenessRepo: studentsLearningObjectivesCompletenessRepo,
	}

	testCases := []TestCase{
		{
			name: "quiz type is not supported",
			req: &sspb.CheckQuizCorrectnessRequest{
				ShuffledQuizSetId: "shuffled_quiz_set_id",
				QuizId:            "quiz_id",
				Answer: []*sspb.Answer{
					{
						Format: &sspb.Answer_SelectedIndex{
							SelectedIndex: 1,
						},
					},
				},
				LmType: sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE,
			},
			setup: func(ctx context.Context) {
				quizRepo.On("GetQuizByExternalID", ctx, mockDB, database.Text("quiz_id")).Once().Return(&entities.Quiz{
					ID:   database.Text("quiz_id"),
					Kind: database.Text(cpb.QuizType_QUIZ_TYPE_ESQ.String()),
				}, nil)
				shuffledQuizSetRepo.On("GetCorrectnessInfo", ctx, mockDB, database.Text("shuffled_quiz_set_id"), database.Text("quiz_id")).Once().Return(&entities.CorrectnessInfo{}, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("CorrectnessQuiz.Check: quiz type is not supported").Error()),
		},
		{
			name: "your answer is not the selected index type",
			req: &sspb.CheckQuizCorrectnessRequest{
				ShuffledQuizSetId: "shuffled_quiz_set_id",
				QuizId:            "quiz_id",
				Answer: []*sspb.Answer{
					{
						Format: &sspb.Answer_FilledText{
							FilledText: "filled_text",
						},
					},
				},
				LmType: sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE,
			},
			setup: func(ctx context.Context) {
				quizRepo.On("GetQuizByExternalID", ctx, mockDB, database.Text("quiz_id")).Once().Return(&entities.Quiz{
					ID:   database.Text("quiz_id"),
					Kind: database.Text(cpb.QuizType_QUIZ_TYPE_MCQ.String()),
				}, nil)
				shuffledQuizSetRepo.On("GetCorrectnessInfo", ctx, mockDB, database.Text("shuffled_quiz_set_id"), database.Text("quiz_id")).Once().Return(&entities.CorrectnessInfo{}, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("CorrectnessQuiz.Check: your answer is not the selected index type").Error()),
		},
		{
			name: "happy case - LO - QUIZ_TYPE_MIQ",
			req: &sspb.CheckQuizCorrectnessRequest{
				ShuffledQuizSetId: "shuffled_quiz_set_id",
				QuizId:            "quiz_id",
				Answer: []*sspb.Answer{
					{
						Format: &sspb.Answer_SelectedIndex{
							SelectedIndex: 1,
						},
					},
				},
				LmType: sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE,
			},
			setup: func(ctx context.Context) {
				quizRepo.On("GetQuizByExternalID", ctx, mockDB, database.Text("quiz_id")).Once().Return(&entities.Quiz{
					ID:   database.Text("quiz_id"),
					Kind: database.Text(cpb.QuizType_QUIZ_TYPE_MIQ.String()),
					Options: database.JSONB(
						[]*cpb.QuizOption{
							{
								Key:     idutil.ULIDNow(),
								Label:   "",
								Configs: []cpb.QuizOptionConfig{},
								Content: &cpb.RichText{
									Raw: "A",
								},
								Correctness: true,
								Attribute:   &cpb.QuizItemAttribute{},
							},
							{
								Key:     idutil.ULIDNow(),
								Label:   "",
								Configs: []cpb.QuizOptionConfig{},
								Content: &cpb.RichText{
									Raw: "B",
								},
								Correctness: false,
								Attribute:   &cpb.QuizItemAttribute{},
							},
						},
					),
				}, nil)
				shuffledQuizSetRepo.On("GetCorrectnessInfo", ctx, mockDB, database.Text("shuffled_quiz_set_id"), database.Text("quiz_id")).Once().Return(&entities.CorrectnessInfo{
					RandomSeed: database.Text("1111"),
				}, nil)

				mockDB.On("Begin", ctx).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				shuffledQuizSetRepo.On("UpsertLOSubmission", ctx, mockTx, database.Text("shuffled_quiz_set_id")).Once().Return(
					&entities.LOSubmissionAnswerKey{
						StudentID:          database.Text("student_id"),
						SubmissionID:       database.Text("submission_id"),
						StudyPlanID:        database.Text("study_plan_id"),
						LearningMaterialID: database.Text("lm_id"),
						ShuffledQuizSetID:  database.Text("shuffled_quiz_set_id"),
					},
					nil,
				)
				loSubmissionAnswerRepo.On("List", ctx, mockTx, mock.Anything).Once().Return([]*entities.LOSubmissionAnswer{}, nil)
				loSubmissionAnswerRepo.On("BulkUpsert", mock.Anything, mockTx, mock.Anything).Once().Return(nil)
				loSubmissionAnswerRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				shuffledQuizSetRepo.On("UpdateTotalCorrectnessAndSubmissionHistory", ctx, mockTx, mock.Anything).Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertFirstQuizCompleteness", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertHighestQuizScore", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			expectedResp: &sspb.CheckQuizCorrectnessResponse{},
			expectedErr:  nil,
		},
		{
			name: "happy case - QUIZ_TYPE_MCQ",
			req: &sspb.CheckQuizCorrectnessRequest{
				ShuffledQuizSetId: "shuffled_quiz_set_id",
				QuizId:            "quiz_id",
				Answer: []*sspb.Answer{
					{
						Format: &sspb.Answer_SelectedIndex{
							SelectedIndex: 1,
						},
					},
				},
				LmType: sspb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE,
			},
			setup: func(ctx context.Context) {
				quizRepo.On("GetQuizByExternalID", ctx, mockDB, database.Text("quiz_id")).Once().Return(&entities.Quiz{
					ID:   database.Text("quiz_id"),
					Kind: database.Text(cpb.QuizType_QUIZ_TYPE_MCQ.String()),
					Options: database.JSONB(
						[]*cpb.QuizOption{
							{
								Key:     idutil.ULIDNow(),
								Label:   "",
								Configs: []cpb.QuizOptionConfig{},
								Content: &cpb.RichText{
									Raw: "A",
								},
								Correctness: true,
								Attribute:   &cpb.QuizItemAttribute{},
							},
							{
								Key:     idutil.ULIDNow(),
								Label:   "",
								Configs: []cpb.QuizOptionConfig{},
								Content: &cpb.RichText{
									Raw: "B",
								},
								Correctness: false,
								Attribute:   &cpb.QuizItemAttribute{},
							},
						},
					),
				}, nil)
				shuffledQuizSetRepo.On("GetCorrectnessInfo", ctx, mockDB, database.Text("shuffled_quiz_set_id"), database.Text("quiz_id")).Once().Return(&entities.CorrectnessInfo{
					RandomSeed: database.Text("1111"),
				}, nil)

				mockDB.On("Begin", ctx).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				shuffledQuizSetRepo.On("UpsertLOSubmission", ctx, mockTx, database.Text("shuffled_quiz_set_id")).Once().Return(
					&entities.LOSubmissionAnswerKey{
						StudentID:          database.Text("student_id"),
						SubmissionID:       database.Text("submission_id"),
						StudyPlanID:        database.Text("study_plan_id"),
						LearningMaterialID: database.Text("lm_id"),
						ShuffledQuizSetID:  database.Text("shuffled_quiz_set_id"),
					},
					nil,
				)
				loSubmissionAnswerRepo.On("List", ctx, mockTx, mock.Anything).Once().Return([]*entities.LOSubmissionAnswer{}, nil)
				loSubmissionAnswerRepo.On("BulkUpsert", mock.Anything, mockTx, mock.Anything).Once().Return(nil)
				loSubmissionAnswerRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)

				shuffledQuizSetRepo.On("UpdateTotalCorrectnessAndSubmissionHistory", ctx, mockTx, mock.Anything).Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertFirstQuizCompleteness", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertHighestQuizScore", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			expectedResp: &sspb.CheckQuizCorrectnessResponse{},
			expectedErr:  nil,
		},
		{
			name: "happy case - FC - QUIZ_TYPE_FIB",
			req: &sspb.CheckQuizCorrectnessRequest{
				ShuffledQuizSetId: "shuffled_quiz_set_id",
				QuizId:            "quiz_id",
				Answer: []*sspb.Answer{
					{
						Format: &sspb.Answer_FilledText{
							FilledText: "filled_text",
						},
					},
				},
				LmType: sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD,
			},
			setup: func(ctx context.Context) {
				quizRepo.On("GetQuizByExternalID", ctx, mockDB, database.Text("quiz_id")).Once().Return(&entities.Quiz{
					ID:   database.Text("quiz_id"),
					Kind: database.Text(cpb.QuizType_QUIZ_TYPE_FIB.String()),
					Options: database.JSONB(
						[]*cpb.QuizOption{
							{
								Key:     idutil.ULIDNow(),
								Label:   "",
								Configs: []cpb.QuizOptionConfig{},
								Content: &cpb.RichText{
									Raw: "A",
								},
								Correctness: true,
								Attribute:   &cpb.QuizItemAttribute{},
							},
							{
								Key:     idutil.ULIDNow(),
								Label:   "",
								Configs: []cpb.QuizOptionConfig{},
								Content: &cpb.RichText{
									Raw: "B",
								},
								Correctness: false,
								Attribute:   &cpb.QuizItemAttribute{},
							},
						},
					),
				}, nil)
				shuffledQuizSetRepo.On("GetCorrectnessInfo", ctx, mockDB, database.Text("shuffled_quiz_set_id"), database.Text("quiz_id")).Once().Return(&entities.CorrectnessInfo{
					RandomSeed: database.Text("1111"),
				}, nil)

				mockDB.On("Begin", ctx).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)

				shuffledQuizSetRepo.On("UpdateTotalCorrectnessAndSubmissionHistory", ctx, mockTx, mock.Anything).Once().Return(nil)
				shuffledQuizSetRepo.On("UpsertFlashCardSubmission", ctx, mockTx, database.Text("shuffled_quiz_set_id")).Once().Return(
					&entities.FlashCardSubmissionAnswerKey{
						StudentID:          database.Text("student_id"),
						SubmissionID:       database.Text("submission_id"),
						StudyPlanID:        database.Text("study_plan_id"),
						LearningMaterialID: database.Text("lm_id"),
						ShuffledQuizSetID:  database.Text("shuffled_quiz_set_id"),
					},
					nil,
				)
				flashCardSubmissionAnswerRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertFirstQuizCompleteness", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentsLearningObjectivesCompletenessRepo.On("UpsertHighestQuizScore", ctx, mockTx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			expectedResp: &sspb.CheckQuizCorrectnessResponse{},
			expectedErr:  nil,
		},
	}

	for _, testCase := range testCases {
		testCase.setup(testCase.ctx)
		_, err := svc.CheckQuizCorrectness(testCase.ctx, testCase.req.(*sspb.CheckQuizCorrectnessRequest))
		if testCase.expectedErr != nil {
			fmt.Println(err.Error())
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
