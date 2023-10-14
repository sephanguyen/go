package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_db "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestLOReader_RetrieveLOs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := &mock_db.Ext{}
	learningObjectiveRepo := new(mock_repositories.MockLearningObjectiveRepo)
	quizSetRepo := new(mock_repositories.MockQuizSetRepo)
	studentsLearningObjectivesCompletenessRepo := new(mock_repositories.MockStudentsLearningObjectivesCompletenessRepo)
	now := time.Now()

	expectedResp := &epb.RetrieveLOsResponse{
		LearningObjectives: []*cpb.LearningObjective{
			{
				Info: &cpb.ContentBasicInfo{
					Id:           "lo-1",
					Name:         "name-1",
					Country:      cpb.Country_COUNTRY_VN,
					Subject:      cpb.Subject_SUBJECT_ENGLISH,
					Grade:        1,
					SchoolId:     1,
					DisplayOrder: 1,
					MasterId:     "master-id-1",
					UpdatedAt:    timestamppb.New(now),
					CreatedAt:    timestamppb.New(now),
				},
				TopicId:        "topic-id-1",
				Video:          "video-1",
				StudyGuide:     "study-guide-1",
				Instruction:    "instruction-1",
				GradeToPass:    wrapperspb.Int32(1),
				ManualGrading:  true,
				TimeLimit:      wrapperspb.Int32(1),
				MaximumAttempt: wrapperspb.Int32(1),
				ApproveGrading: true,
				GradeCapping:   true,
				ReviewOption:   cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE,
			},
			{
				Info: &cpb.ContentBasicInfo{
					Id:           "lo-2",
					Name:         "name-2",
					Country:      cpb.Country_COUNTRY_VN,
					Subject:      cpb.Subject_SUBJECT_ENGLISH,
					Grade:        2,
					SchoolId:     2,
					DisplayOrder: 2,
					MasterId:     "master-id-2",
					UpdatedAt:    timestamppb.New(now),
					CreatedAt:    timestamppb.New(now),
				},
				TopicId:        "topic-id-2",
				Video:          "video-2",
				StudyGuide:     "study-guide-2",
				Instruction:    "instruction-2",
				GradeToPass:    &wrapperspb.Int32Value{},
				ManualGrading:  false,
				TimeLimit:      &wrapperspb.Int32Value{},
				ApproveGrading: false,
				GradeCapping:   false,
				ReviewOption:   cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY,
			},
		},
	}

	expectedLOs := []*entities.LearningObjective{
		{
			ID:             database.Text("lo-1"),
			Name:           database.Text("name-1"),
			Country:        database.Text(cpb.Country_COUNTRY_VN.String()),
			Grade:          database.Int2(1),
			Subject:        database.Text(cpb.Subject_SUBJECT_ENGLISH.String()),
			TopicID:        database.Text("topic-id-1"),
			MasterLoID:     database.Text("master-id-1"),
			DisplayOrder:   database.Int2(1),
			SchoolID:       database.Int4(1),
			Video:          database.Text("video-1"),
			StudyGuide:     database.Text("study-guide-1"),
			CreatedAt:      database.Timestamptz(now),
			UpdatedAt:      database.Timestamptz(now),
			Type:           database.Text(cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_NONE.String()),
			Instruction:    database.Text("instruction-1"),
			GradeToPass:    database.Int4(1),
			ManualGrading:  database.Bool(true),
			TimeLimit:      database.Int4(1),
			MaximumAttempt: database.Int4(1),
			ApproveGrading: database.Bool(true),
			GradeCapping:   database.Bool(true),
			ReviewOption:   database.Text(cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE.String()),
		},
		{
			ID:             database.Text("lo-2"),
			Name:           database.Text("name-2"),
			Country:        database.Text(cpb.Country_COUNTRY_VN.String()),
			Grade:          database.Int2(2),
			Subject:        database.Text(cpb.Subject_SUBJECT_ENGLISH.String()),
			SchoolID:       database.Int4(2),
			DisplayOrder:   database.Int2(2),
			MasterLoID:     database.Text("master-id-2"),
			TopicID:        database.Text("topic-id-2"),
			Video:          database.Text("video-2"),
			StudyGuide:     database.Text("study-guide-2"),
			CreatedAt:      database.Timestamptz(now),
			UpdatedAt:      database.Timestamptz(now),
			Type:           database.Text(cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_NONE.String()),
			Instruction:    database.Text("instruction-2"),
			ManualGrading:  database.Bool(false),
			ApproveGrading: database.Bool(false),
			GradeCapping:   database.Bool(false),
			ReviewOption:   database.Text(cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()),
		},
	}

	testCases := []TestCase{
		{
			name: "happy case with topicIDs",
			ctx:  ctx,
			req: &epb.RetrieveLOsRequest{
				TopicIds:         []string{"topic-1", "topic-2"},
				WithCompleteness: true,
				StudentId:        "studentID",
			},
			expectedErr:  nil,
			expectedResp: expectedResp,
			setup: func(ctx context.Context) {
				learningObjectiveRepo.On("RetrieveByTopicIDs", ctx, db, mock.Anything, mock.Anything).Once().Return(expectedLOs, nil)
				quizSetRepo.On("CountQuizOnLO", ctx, db, mock.Anything).Once().Return(map[string]int32{
					"quiz-1": 1,
					"quiz-2": 2,
				}, nil)
				studentsLearningObjectivesCompletenessRepo.On("Find", ctx, db, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness{}, nil)
			},
		},
		{
			name: "have no any topicIDs or LO IDs",
			ctx:  ctx,
			req: &epb.RetrieveLOsRequest{
				WithCompleteness: true,
				StudentId:        "studentID",
			},
			expectedErr:  nil,
			expectedResp: &epb.RetrieveLOsResponse{},
			setup:        func(ctx context.Context) {},
		},
		{
			name: "happy case with LoIds",
			ctx:  ctx,
			req: &epb.RetrieveLOsRequest{
				LoIds:            []string{"lo-1", "lo-2"},
				WithCompleteness: true,
				StudentId:        "studentID",
			},
			expectedErr:  nil,
			expectedResp: expectedResp,
			setup: func(ctx context.Context) {
				learningObjectiveRepo.On("RetrieveByIDs", ctx, db, mock.Anything, mock.Anything).Once().Return(expectedLOs, nil)
				quizSetRepo.On("CountQuizOnLO", ctx, db, mock.Anything).Once().Return(map[string]int32{
					"quiz-1": 1,
					"quiz-2": 2,
				}, nil)
				studentsLearningObjectivesCompletenessRepo.On("Find", ctx, db, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness{}, nil)
			},
		},
		{
			name: "repo err case",
			ctx:  ctx,
			req: &epb.RetrieveLOsRequest{
				LoIds:            []string{"lo-1", "lo-2"},
				WithCompleteness: true,
				StudentId:        "studentID",
			},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				learningObjectiveRepo.On("RetrieveByIDs", ctx, db, mock.Anything, mock.Anything).Once().Return(expectedLOs, nil)
				quizSetRepo.On("CountQuizOnLO", ctx, db, mock.Anything).Once().Return(map[string]int32{}, pgx.ErrNoRows)
			},
		},
		{
			name: "happy case with completeness true",
			ctx:  interceptors.ContextWithUserID(ctx, "happy complete true"),
			req: &epb.RetrieveLOsRequest{
				TopicIds: []string{
					"id",
				},
				WithCompleteness: true,
			},
			expectedErr:  nil,
			expectedResp: expectedResp,
			setup: func(ctx context.Context) {
				learningObjectiveRepo.On("RetrieveByTopicIDs", ctx, mock.Anything, mock.Anything).Once().Return(expectedLOs, nil)

				m := make(map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness)
				m[database.Text("lo-1")] = &entities.StudentsLearningObjectivesCompleteness{
					LoID: database.Text("lo-1"),
				}
				studentsLearningObjectivesCompletenessRepo.On("Find", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(m, nil)
				quizSetRepo.On("CountQuizOnLO", ctx, mock.Anything, mock.Anything).Once().Return(map[string]int32{
					"lo-1": 1,
					"lo-2": 2,
				}, nil)
			},
		},
	}

	s := &LoReaderService{
		DB:                    db,
		LearningObjectiveRepo: learningObjectiveRepo,
		QuizSetRepo:           quizSetRepo,
		StudentsLearningObjectivesCompletenessRepo: studentsLearningObjectivesCompletenessRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*epb.RetrieveLOsRequest)
			res, err := s.RetrieveLOs(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp != nil {
				expectedResp := testCase.expectedResp.(*epb.RetrieveLOsResponse)
				assert.Equal(t, len(expectedResp.LearningObjectives), len(res.LearningObjectives))
				for i, lo := range expectedResp.LearningObjectives {
					assert.Equal(t, lo, res.LearningObjectives[i])
				}
			}

		})
	}
}
