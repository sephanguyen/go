package services

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCourseModel(t *testing.T) {
}

func TestToTopicPb(t *testing.T) {
	t.Parallel()
	e1 := generateEnTopic()
	topic1 := ToTopicPb(e1)
	require.True(t, isEqualTopicEnAndPb(e1, topic1))
}

func generateTopic() *pb.Topic {
	rand.Seed(time.Now().UnixNano())
	num := rand.Int()
	return &pb.Topic{
		Id:           fmt.Sprintf("%d", num),
		Name:         "Random name",
		Country:      pb.COUNTRY_VN,
		Grade:        "12",
		Subject:      pb.SUBJECT_MATHS,
		Type:         pb.TOPIC_TYPE_LEARNING,
		CreatedAt:    &types.Timestamp{Seconds: time.Now().Unix()},
		UpdatedAt:    &types.Timestamp{Seconds: time.Now().Unix()},
		Status:       pb.TOPIC_STATUS_DRAFT,
		ChapterId:    "mock-chapter-id",
		DisplayOrder: 1,
	}
}

func generateEnTopic() *entities_bob.Topic {
	e := new(entities_bob.Topic)
	e.ID.Set("1")
	e.Name.Set("Some topic name")
	e.Country.Set("COUNTRY_VN")
	e.Grade.Set("2")
	e.Subject.Set("SUBJECT_MATHS")
	e.TopicType.Set("TOPIC_TYPE_LEARNING")
	e.CreatedAt.Set(time.Now())
	e.UpdatedAt.Set(time.Now())
	e.DeletedAt.Set(nil)
	e.Status.Set("TOPIC_STATUS_DRAFT")
	e.PublishedAt.Set(nil)
	return e
}

func isEqualTopicEnAndPb(e *entities_bob.Topic, topic *pb.Topic) bool {
	if topic.UpdatedAt == nil {
		topic.UpdatedAt = &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	}
	if topic.CreatedAt == nil {
		topic.CreatedAt = &types.Timestamp{Seconds: e.CreatedAt.Time.Unix()}
	}
	updatedAt := &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	createdAt := &types.Timestamp{Seconds: e.CreatedAt.Time.Unix()}

	// grade, _ := convertIntGradeToString(topic.Country, int(e.Grade.Int))
	return (e.ID.String == topic.Id) &&
		(e.Name.String == topic.Name) &&
		(pb.Country(pb.Country_value[e.Country.String]) == topic.Country) &&
		// (grade == topic.Grade) &&
		(pb.Subject(pb.Subject_value[e.Subject.String]) == topic.Subject) &&
		(pb.TopicType(pb.TopicType_value[e.TopicType.String]) == topic.Type) &&
		updatedAt.Equal(topic.UpdatedAt) &&
		createdAt.Equal(topic.CreatedAt)
}

func TestValidateAnswer(t *testing.T) {
	t.Parallel()
	s := &CourseService{}
	correctAnswerList := []string{"answer 1", "answer 2", "answer 3", "answer 4"}
	wrongAnswerList := []string{"answer 1", "", "answer 2", ""}
	err1 := s.validateAnswer(correctAnswerList)
	err2 := s.validateAnswer(wrongAnswerList)
	require.Nil(t, err1)
	require.Error(t, err2)
}

func TestValidateQuestionCorrect(t *testing.T) {
	t.Parallel()
	q1 := generateValidQuestion()
	q2 := generateValidQuestion()
	q3 := generateValidQuestion()
	s := &CourseService{}
	questionList := []*pb.Question{
		&q1, &q2, &q3,
	}
	err := s.validateQuestion(context.Background(), questionList)
	require.Nil(t, err)
}

func TestValidateQuestionEmptyAnswer(t *testing.T) {
	t.Parallel()
	q1 := generateValidQuestion()
	q2 := generateValidQuestion()
	q3 := generateValidQuestion()
	q3.Answers = nil

	s := &CourseService{}
	questionList := []*pb.Question{
		&q1, &q2, &q3,
	}
	err := s.validateQuestion(context.Background(), questionList)
	require.Error(t, err)
}

func TestValidateQuestionEmptyQuestion(t *testing.T) {
	t.Parallel()
	q1 := generateValidQuestion()
	q2 := generateValidQuestion()
	q3 := generateValidQuestion()
	q3.Question = ""

	s := &CourseService{}
	questionList := []*pb.Question{
		&q1, &q2, &q3,
	}
	err := s.validateQuestion(context.Background(), questionList)
	require.Error(t, err)
}

func TestValidateQuestionInvalidDiffLv(t *testing.T) {
	t.Parallel()
	q1 := generateValidQuestion()
	q2 := generateValidQuestion()
	q3 := generateValidQuestion()
	q3.DifficultyLevel = -1

	s := &CourseService{}
	questionList := []*pb.Question{
		&q1, &q2, &q3,
	}
	err := s.validateQuestion(context.Background(), questionList)
	require.Error(t, err)
}

func TestValidateQuestionInvalidMasterID(t *testing.T) {
	t.Parallel()
	q1 := generateValidQuestion()
	q2 := generateValidQuestion()
	q3 := generateValidQuestion()
	q3.MasterQuestionId = "1"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	questionRepo := new(mock_repositories.MockQuestionRepo)
	s := &CourseService{
		QuestionRepo: questionRepo,
	}
	questionList := []*pb.Question{
		&q1, &q2, &q3,
	}
	questionRepo.On("ExistMasterQuestion", ctx, mock.Anything, "1").Return(false, nil)
	err := s.validateQuestion(ctx, questionList)
	require.Error(t, err)
}

func TestRulesForMasterQuestionID(t *testing.T) {
	t.Parallel()
	q1 := generateValidQuestion()
	q1.MasterQuestionId = "1"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	questionRepo := new(mock_repositories.MockQuestionRepo)
	questionRepo.On("ExistMasterQuestion", ctx, mock.Anything, mock.Anything).Once().Return(true, nil)
	s := &CourseService{
		QuestionRepo: questionRepo,
	}
	err := s.rulesForMasterQuestionID(ctx, &q1)
	require.Nil(t, err)
}

func TestQuestionTagLoMapFromReq(t *testing.T) {
	t.Parallel()
}

func generateValidQuestion() pb.Question {
	rand.Seed(time.Now().UnixNano())
	num := rand.Int()
	return pb.Question{
		Id:               fmt.Sprintf("%d", num),
		MasterQuestionId: "",
		Country:          pb.COUNTRY_VN,
		Question:         fmt.Sprintf("valid question %d", num),
		Answers: []string{
			fmt.Sprintf("valid answer %d", 0),
			fmt.Sprintf("valid answer %d", 1),
			fmt.Sprintf("valid answer %d", 2),
			fmt.Sprintf("valid answer %d", 3),
		},
		Explanation:     "Some explanation for question",
		DifficultyLevel: 2,
		UpdatedAt:       nil,
		CreatedAt:       nil,
		QuestionsTagLo: []string{
			"", "", "",
		},
	}
}

func TestToLoEntity(t *testing.T) {
	t.Parallel()
	lo1 := generateValidLearningObjective()
	e1, _ := toLoEntity(lo1)
	lo2 := generateValidLearningObjective()
	lo2.MasterLo = ""
	lo2.CreatedAt = nil
	lo2.UpdatedAt = nil
	lo2.Id = ""
	e2, _ := toLoEntity(lo2)
	require.True(t, isEqualLearningObjectiveEnAndPb(e1, lo1))
	require.True(t, isEqualLearningObjectiveEnAndPb(e2, lo2))
}

func isEqualLearningObjectiveEnAndPb(e *entities_bob.LearningObjective, lo *pb.LearningObjective) bool {
	if lo.Id == "" {
		lo.Id = e.ID.String
	}
	if lo.MasterLo == "" {
		lo.MasterLo = e.MasterLoID.String
	}
	if lo.UpdatedAt == nil {
		lo.UpdatedAt = &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	}
	if lo.CreatedAt == nil {
		lo.CreatedAt = &types.Timestamp{Seconds: e.CreatedAt.Time.Unix()}
	}
	updatedAt := &types.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	createdAt := &types.Timestamp{Seconds: e.CreatedAt.Time.Unix()}

	var prerequisites pgtype.TextArray
	prerequisites.Set(lo.Prerequisites)

	return (e.Name.String == lo.Name) &&
		(e.TopicID.String == lo.TopicId) &&
		(e.VideoScript.String == lo.VideoScript) &&
		(e.DisplayOrder.Int == int16(lo.DisplayOrder)) &&
		reflect.DeepEqual(e.Prerequisites, prerequisites) &&
		(e.Video.String == lo.Video) &&
		(e.StudyGuide.String == lo.StudyGuide) &&
		(updatedAt.Equal(lo.UpdatedAt)) &&
		(createdAt.Equal(lo.CreatedAt))
}

func generateValidLearningObjective() *pb.LearningObjective {
	rand.Seed(time.Now().UnixNano())
	num := rand.Int()
	return &pb.LearningObjective{
		Id:           fmt.Sprintf("%d", num),
		Name:         "learning",
		Country:      pb.COUNTRY_VN,
		Grade:        "G12",
		Subject:      pb.SUBJECT_MATHS,
		TopicId:      "VN12-MA1",
		MasterLo:     "1",
		DisplayOrder: 1,
		VideoScript:  "script",
		Prerequisites: []string{
			"AL-PH3.1", "AL-PH3.2",
		},
		StudyGuide: "https://guides/1/master",
		Video:      "https://videos/1/master",
		SchoolId:   constants.ManabieSchool,
		CreatedAt:  &types.Timestamp{Seconds: time.Now().Unix()},
		UpdatedAt:  &types.Timestamp{Seconds: time.Now().Unix()},
	}
}

func TestListTopic(t *testing.T) {
	t.Parallel()
	topicRepo := new(mock_repositories.MockTopicRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	testCases := []struct {
		name        string
		ctx         context.Context
		req         *pb.ListTopicRequest
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name: "happy case",
			ctx:  context.Background(),
			req: &pb.ListTopicRequest{
				Country:   pb.COUNTRY_VN,
				Subject:   pb.SUBJECT_MATHS,
				TopicType: pb.TOPIC_TYPE_LEARNING,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupStudent, nil)
				topicRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Topic{
					generateEnTopic(),
				}, nil)
			},
		},
		{
			name: "can't find usesr",
			ctx:  context.Background(),
			req: &pb.ListTopicRequest{
				Country:   pb.COUNTRY_VN,
				Subject:   pb.SUBJECT_MATHS,
				TopicType: pb.TOPIC_TYPE_LEARNING,
			},
			expectedErr: status.Error(codes.Unauthenticated, "wrong token"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupStudent, pgx.ErrNoRows)
			},
		},
		{
			name:        "empty req case",
			ctx:         context.Background(),
			req:         &pb.ListTopicRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupStudent, nil)
				topicRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Topic{
					generateEnTopic(),
				}, nil)
			},
		},
		{
			name:        "fail repo retrieve case",
			ctx:         context.Background(),
			req:         &pb.ListTopicRequest{},
			expectedErr: errors.Wrap(pgx.ErrNoRows, "c.TopicRepo.Retrieve"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupStudent, nil)
				topicRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "err parse grade",
			ctx:         context.Background(),
			req:         &pb.ListTopicRequest{Grade: "G12", Country: pb.COUNTRY_NONE},
			expectedErr: errors.Wrap(status.Error(codes.InvalidArgument, "cannot find country grade map"), "c.TopicRepo.Retrieve"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupStudent, nil)
			},
		},
	}
	s := &CourseService{
		TopicRepo: topicRepo,
		UserRepo:  userRepo,
	}
	for _, testCase := range testCases {
		testCase.setup(testCase.ctx)
		_, err := s.ListTopic(testCase.ctx, testCase.req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestRetrieveSubmissionHistoryV2(t *testing.T) {
}

func TestUpsertQuizSets(t *testing.T) {

}

func TestTakeQuizTest(t *testing.T) {
}

func TestUpsertTopics(t *testing.T) {
}

func TestUpsertLOs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	learningObjectiveRepo := new(mock_repositories.MockLearningObjectiveRepo)
	topicsLearningObjectivesRepo := new(mock_repositories.MockTopicsLearningObjectivesRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)
	activityLogRepo := new(mock_repositories.MockActivityLogRepo)

	lo1 := generateValidLearningObjective()
	lo2 := generateValidLearningObjective()

	topics := []*entities.Topic{
		{
			ID:                    database.Text(lo1.TopicId),
			LODisplayOrderCounter: database.Int4(0),
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &pb.UpsertLOsRequest{
				LearningObjectives: []*pb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				topicsLearningObjectivesRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Times(1).Return(nil)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupAdmin, nil)
				learningObjectiveRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateTotalLOs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				activityLogRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Times(1).Return(nil)
			},
		},
		{
			name: "Fail Create",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &pb.UpsertLOsRequest{
				LearningObjectives: []*pb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to bulk import learning objective: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				topicsLearningObjectivesRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Times(1).Return(nil)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupAdmin, nil)
				learningObjectiveRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
				topicRepo.On("UpdateTotalLOs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				activityLogRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Times(0).Return(nil)
			},
		},
	}

	s := &CourseService{
		EurekaDBTrace:              db,
		DB:                         db,
		UserRepo:                   userRepo,
		LearningObjectiveRepo:      learningObjectiveRepo,
		TopicLearningObjectiveRepo: topicsLearningObjectivesRepo,
		TopicRepo:                  topicRepo,
		ActivityLogRepo:            activityLogRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpsertLOsRequest)
			_, err := s.UpsertLOs(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestUpsertPresetStudyPlans(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)
	userRepo := new(mock_repositories.MockUserRepo)

	s := &CourseService{
		UserRepo:            userRepo,
		PresetStudyPlanRepo: presetStudyPlanRepo,
	}

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)
	studyPlan1 := generatePresetStudyPlan()
	testCases := []TestCase{
		{
			name: "error query",
			ctx:  ctx,
			req: &pb.UpsertPresetStudyPlansRequest{
				PresetStudyPlans: []*pb.PresetStudyPlan{studyPlan1},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupAdmin, nil)
				presetStudyPlanRepo.On("CreatePresetStudyPlan", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req: &pb.UpsertPresetStudyPlansRequest{
				PresetStudyPlans: []*pb.PresetStudyPlan{studyPlan1},
			},
			expectedResp: &pb.UpsertPresetStudyPlansResponse{
				PresetStudyPlanIds: []string{studyPlan1.Id},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupAdmin, nil)
				presetStudyPlanRepo.On("CreatePresetStudyPlan", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.UpsertPresetStudyPlans(testCase.ctx, testCase.req.(*pb.UpsertPresetStudyPlansRequest))
			assert.Equal(t, testCase.expectedErr, err)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func generatePresetStudyPlan() *pb.PresetStudyPlan {
	rand.Seed(time.Now().UnixNano())
	num := rand.Int()
	return &pb.PresetStudyPlan{
		Id:        fmt.Sprintf("%d", num),
		Name:      "Random name",
		Country:   pb.COUNTRY_VN,
		Grade:     "G12",
		Subject:   pb.SUBJECT_MATHS,
		CreatedAt: &types.Timestamp{Seconds: time.Now().Unix()},
		UpdatedAt: &types.Timestamp{Seconds: time.Now().Unix()},
	}
}

func TestUpsertPresetStudyPlanWeeklies(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	pw1 := generatePresetStudyPlanWeekly()
	pw1.Id = ""
	pw2 := generatePresetStudyPlanWeekly()
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &pb.UpsertPresetStudyPlanWeekliesRequest{
				PresetStudyPlanWeeklies: []*pb.PresetStudyPlanWeekly{
					pw1, pw2,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupAdmin, nil)
				presetStudyPlanRepo.On("CreatePresetStudyPlanWeekly", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error query",
			ctx:  interceptors.ContextWithUserID(ctx, "error query"),
			req: &pb.UpsertPresetStudyPlanWeekliesRequest{
				PresetStudyPlanWeeklies: []*pb.PresetStudyPlanWeekly{
					pw1, pw2,
				},
			},
			expectedErr: status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupAdmin, nil)
				presetStudyPlanRepo.On("CreatePresetStudyPlanWeekly", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	s := &CourseService{
		UserRepo:            userRepo,
		PresetStudyPlanRepo: presetStudyPlanRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpsertPresetStudyPlanWeekliesRequest)
			_, err := s.UpsertPresetStudyPlanWeeklies(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestRetrieveMap(t *testing.T) {
	t.Parallel()
	s := &CourseService{}
	rsp, err := s.RetrieveGradeMap(context.Background(), &pb.RetrieveGradeMapRequest{})
	assert.Nil(t, err)
	gradeMap := rsp.GradeMap
	testCases := []struct {
		country pb.Country
		in      []string
		out     []int
	}{
		{
			pb.COUNTRY_VN,
			[]string{"Lớp 1", "Lớp 2", "Lớp 3", "Lớp 4", "Lớp 5", "Lớp 6", "Lớp 7", "Lớp 8", "Lớp 9", "Lớp 10", "Lớp 11", "Lớp 12"},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			pb.COUNTRY_MASTER,
			[]string{"Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5", "Grade 6", "Grade 7", "Grade 8", "Grade 9", "Grade 10", "Grade 11", "Grade 12"},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.country.String(), func(t *testing.T) {
			for i, grade := range tt.in {
				localGrade := gradeMap[tt.country.String()]
				localGradeMap := localGrade.LocalGrade
				cGrade := localGradeMap[grade]
				assert.Equal(t, cGrade, int32(tt.out[i]))
			}
		})
	}
}

func generatePresetStudyPlanWeekly() *pb.PresetStudyPlanWeekly {
	rand.Seed(time.Now().UnixNano())
	num := rand.Int()
	week := rand.Intn(30)
	kid := ksuid.New().String()
	return &pb.PresetStudyPlanWeekly{
		Id:                fmt.Sprintf(kid),
		TopicId:           fmt.Sprintf("%d", num),
		PresetStudyPlanId: fmt.Sprintf("presetID_%d", num),
		Week:              int32(week),
	}
}

func TestCourseService_GetHistoryQuizDetail(t *testing.T) {
}

func TestCourseService_SuggestLO(t *testing.T) {
}

func TestCourseService_PublishTopics(t *testing.T) {
}

func TestCourseService_PublishTopics_Abac(t *testing.T) {
}

func TestCourseService_RetrieveCourses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	classRepo := new(mock_repositories.MockClassRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	chapterRepo := new(mock_repositories.MockChapterRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)
	bookRepo := new(mock_repositories.MockBookRepo)
	courseBookRepo := new(mock_repositories.MockCourseBookRepo)
	bookChapterRepo := new(mock_repositories.MockBookChapterRepo)

	s := &CourseService{
		UserRepo:        userRepo,
		ClassRepo:       classRepo,
		CourseRepo:      courseRepo,
		CourseClassRepo: courseClassRepo,
		ChapterRepo:     chapterRepo,
		LessonRepo:      lessonRepo,
		BookRepo:        bookRepo,
		CourseBookRepo:  courseBookRepo,
		TopicRepo:       topicRepo,
		BookChapterRepo: bookChapterRepo,
	}
	class := &entities_bob.Class{}
	_ = class.ID.Set(1)
	_ = class.SchoolID.Set(2)

	classes := []*entities_bob.Class{
		class,
	}

	users := []*entities_bob.User{}

	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	_ = course.TeacherIDs.Set([]string{"teacherIDs"})
	_ = course.ID.Set("1")
	_ = course.Country.Set(pb.COUNTRY_VN.String())
	_ = course.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = course.Grade.Set(8)
	_ = course.SchoolID.Set(2)

	chapter := &entities_bob.Chapter{}
	database.AllNullEntity(chapter)
	_ = chapter.ID.Set("1")
	_ = chapter.Country.Set(pb.COUNTRY_VN.String())
	_ = chapter.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = chapter.Grade.Set(8)
	_ = chapter.SchoolID.Set(2)

	lesson := &entities_bob.Lesson{}
	database.AllNullEntity(lesson)
	_ = lesson.CourseID.Set("1")
	_ = lesson.LessonID.Set("1")

	courses := entities_bob.Courses{course}
	courseClass := make(map[pgtype.Text]pgtype.Int4Array)

	enChapters := []*entities_bob.Chapter{}
	enChapters = append(enChapters, chapter)

	enLessons := []*entities_bob.Lesson{}
	enLessons = append(enLessons, lesson)

	book := &entities_bob.Book{}
	database.AllNullEntity(book)
	_ = book.ID.Set("1")
	_ = book.Country.Set(pb.COUNTRY_VN.String())
	_ = book.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = book.Grade.Set(8)
	_ = book.SchoolID.Set(2)

	enCourseBook := map[string][]string{}
	enCourseBook[course.ID.String] = []string{book.ID.String}

	enBookChapters := map[string][]*entities_bob.BookChapter{}
	bookChapter := &entities_bob.BookChapter{}
	database.AllNullEntity(bookChapter)
	bookChapter.BookID.Set(book.ID.String)
	bookChapter.ChapterID.Set(chapter.ID.String)

	enBookChapters[book.ID.String] = []*entities_bob.BookChapter{bookChapter}

	testcases := map[string]TestCase{
		"retrieve content course have a book": {
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			req:         &pb.RetrieveCoursesRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				lesson1 := &entities_bob.Lesson{}
				database.AllNullEntity(lesson1)
				_ = lesson1.CourseID.Set("1")
				_ = lesson1.LessonID.Set("1")

				course1 := &entities_bob.Course{}
				database.AllNullEntity(course1)
				_ = course1.TeacherIDs.Set([]string{"teacherIDs"})
				_ = course1.ID.Set("1")
				_ = course1.CourseType.Set(pb.COURSE_TYPE_CONTENT.String())

				courses1 := entities_bob.Courses{course1}
				courseClass1 := make(map[pgtype.Text]pgtype.Int4Array)

				enChapters1 := map[string]*entities_bob.Chapter{}
				enChapters1[chapter.ID.String] = chapter

				enLessons1 := []*entities_bob.Lesson{}
				enLessons1 = append(enLessons, lesson)

				book1 := &entities_bob.Book{}
				database.AllNullEntity(book1)
				_ = book1.ID.Set("1")
				_ = book1.Country.Set(pb.COUNTRY_VN.String())
				_ = book1.Subject.Set(pb.SUBJECT_MATHS.String())
				_ = book1.Grade.Set(8)
				_ = book1.SchoolID.Set(2)

				enCourseBook := map[string][]string{}
				enCourseBook[course1.ID.String] = []string{book1.ID.String}

				enBooks := map[string]*entities_bob.Book{}
				enBooks[book1.ID.String] = book1

				chapter2 := &entities_bob.Chapter{}
				database.AllNullEntity(chapter2)
				_ = chapter2.ID.Set("2")
				_ = chapter2.Country.Set(pb.COUNTRY_VN.String())
				_ = chapter2.Subject.Set(pb.SUBJECT_MATHS.String())
				_ = chapter2.Grade.Set(8)
				_ = chapter2.SchoolID.Set(2)

				enTopics := []*entities_bob.Topic{}
				topic := &entities_bob.Topic{}
				database.AllNullEntity(topic)
				_ = topic.ID.Set("1")
				_ = topic.ChapterID.Set(chapter2.ID.String)
				_ = topic.Country.Set(pb.COUNTRY_VN.String())
				_ = topic.Subject.Set(pb.SUBJECT_MATHS.String())
				_ = topic.Grade.Set(8)
				_ = topic.SchoolID.Set(2)
				_ = topic.DisplayOrder.Set(1)
				_ = topic.CreatedAt.Set(time.Now())
				_ = topic.UpdatedAt.Set(time.Now())

				enTopics = append(enTopics, topic)

				enBookChapters1 := map[string][]*entities_bob.BookChapter{}
				bookChapter := &entities_bob.BookChapter{}
				database.AllNullEntity(bookChapter)
				bookChapter.BookID.Set(book1.ID.String)
				bookChapter.ChapterID.Set(chapter.ID.String)

				bookChapter2 := &entities_bob.BookChapter{}
				database.AllNullEntity(bookChapter2)
				bookChapter2.BookID.Set(book1)
				bookChapter2.ChapterID.Set(chapter2.ID.String)

				enBookChapters1[book1.ID.String] = []*entities_bob.BookChapter{bookChapter, bookChapter2}

				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)

				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{constants.ManabieSchool, 2},
					ClassIDs:  []int{},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
				}

				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses1, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass1, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(enCourseBook, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters1, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{"1"}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons1, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Chapter{chapter2}, 1, nil)

				topicRepo.On("FindByChapterIds", ctx, mock.Anything, mock.Anything).Once().Return(enTopics, nil)
			},
		},
		"retrieve content course have many book": {
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			req:         &pb.RetrieveCoursesRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				lesson1 := &entities_bob.Lesson{}
				database.AllNullEntity(lesson1)
				_ = lesson1.CourseID.Set("1")
				_ = lesson1.LessonID.Set("1")

				course1 := &entities_bob.Course{}
				database.AllNullEntity(course1)
				_ = course1.TeacherIDs.Set([]string{"teacherIDs"})
				_ = course1.ID.Set("1")
				_ = course1.CourseType.Set(pb.COURSE_TYPE_CONTENT.String())

				courses1 := entities_bob.Courses{course1}
				courseClass1 := make(map[pgtype.Text]pgtype.Int4Array)

				enChapters1 := map[string]*entities_bob.Chapter{}
				enChapters1[chapter.ID.String] = chapter

				enLessons1 := []*entities_bob.Lesson{}
				enLessons1 = append(enLessons, lesson)

				book1 := &entities_bob.Book{}
				database.AllNullEntity(book1)
				_ = book1.ID.Set("1")
				_ = book1.Country.Set(pb.COUNTRY_VN.String())
				_ = book1.Subject.Set(pb.SUBJECT_MATHS.String())
				_ = book1.Grade.Set(8)
				_ = book1.SchoolID.Set(2)

				enCourseBook := map[string][]string{}
				enCourseBook[course1.ID.String] = []string{book1.ID.String, book.ID.String}

				enBooks := map[string]*entities_bob.Book{}
				enBooks[book1.ID.String] = book1
				enBooks[book.ID.String] = book

				chapter2 := &entities_bob.Chapter{}
				database.AllNullEntity(chapter2)
				_ = chapter2.ID.Set("2")
				_ = chapter2.Country.Set(pb.COUNTRY_VN.String())
				_ = chapter2.Subject.Set(pb.SUBJECT_MATHS.String())
				_ = chapter2.Grade.Set(8)
				_ = chapter2.SchoolID.Set(2)

				enTopics := []*entities_bob.Topic{}
				topic := &entities_bob.Topic{}
				database.AllNullEntity(topic)
				_ = topic.ID.Set("1")
				_ = topic.ChapterID.Set(chapter2.ID.String)
				_ = topic.Country.Set(pb.COUNTRY_VN.String())
				_ = topic.Subject.Set(pb.SUBJECT_MATHS.String())
				_ = topic.Grade.Set(8)
				_ = topic.SchoolID.Set(2)
				_ = topic.DisplayOrder.Set(1)
				_ = topic.CreatedAt.Set(time.Now())
				_ = topic.UpdatedAt.Set(time.Now())

				enTopics = append(enTopics, topic)

				enBookChapters1 := map[string][]*entities_bob.BookChapter{}
				bookChapter := &entities_bob.BookChapter{}
				database.AllNullEntity(bookChapter)
				bookChapter.BookID.Set(book1.ID.String)
				bookChapter.ChapterID.Set(chapter.ID.String)

				bookChapter2 := &entities_bob.BookChapter{}
				database.AllNullEntity(bookChapter2)
				bookChapter2.BookID.Set(book1)
				bookChapter2.ChapterID.Set(chapter2.ID.String)

				enBookChapters1[book1.ID.String] = []*entities_bob.BookChapter{bookChapter, bookChapter2}

				bookChapter3 := &entities_bob.BookChapter{}
				database.AllNullEntity(bookChapter3)
				bookChapter3.BookID.Set(book.ID.String)
				bookChapter3.ChapterID.Set(chapter.ID.String)

				bookChapter4 := &entities_bob.BookChapter{}
				database.AllNullEntity(bookChapter4)
				bookChapter4.BookID.Set(book)
				bookChapter4.ChapterID.Set(chapter2.ID.String)

				enBookChapters1[book.ID.String] = []*entities_bob.BookChapter{bookChapter3, bookChapter4}

				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)

				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{constants.ManabieSchool, 2},
					ClassIDs:  []int{},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
				}

				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses1, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass1, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(enCourseBook, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters1, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{"1"}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons1, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Chapter{}, 0, nil)

				topicRepo.On("FindByChapterIds", ctx, mock.Anything, mock.Anything).Once().Return(enTopics, nil)
			},
		},
		"retrieve content course have a invalid chapter id": {
			ctx:         interceptors.ContextWithUserID(ctx, "retrieve content course have a invalid chapter id"),
			req:         &pb.RetrieveCoursesRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				lesson1 := &entities_bob.Lesson{}
				database.AllNullEntity(lesson1)
				_ = lesson1.CourseID.Set("1")
				_ = lesson1.LessonID.Set("1")

				course1 := &entities_bob.Course{}
				database.AllNullEntity(course1)
				_ = course1.TeacherIDs.Set([]string{"teacherIDs"})
				_ = course1.ID.Set("1")
				_ = course1.CourseType.Set(pb.COURSE_TYPE_CONTENT.String())

				courses1 := entities_bob.Courses{course1}
				courseClass1 := make(map[pgtype.Text]pgtype.Int4Array)

				enChapters1 := map[string]*entities_bob.Chapter{}
				enChapters1[chapter.ID.String] = chapter

				enLessons1 := []*entities_bob.Lesson{}
				enLessons1 = append(enLessons, lesson)

				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)

				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{constants.ManabieSchool, 2},
					ClassIDs:  []int{},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
				}
				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses1, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass1, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{}, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{"1"}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons1, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(enChapters, 1, nil)
			},
		},
		"student retrieve live course, is assign equal true": {
			ctx: interceptors.ContextWithUserID(ctx, "student retrieve live course, is assign equal true"),
			req: &pb.RetrieveCoursesRequest{
				IsAssigned: true,
				CourseType: pb.COURSE_TYPE_LIVE,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				lesson1 := &entities_bob.Lesson{}
				database.AllNullEntity(lesson1)
				_ = lesson1.CourseID.Set("1")
				_ = lesson1.LessonID.Set("1")

				course1 := &entities_bob.Course{}
				database.AllNullEntity(course1)
				_ = course1.TeacherIDs.Set([]string{"teacherIDs"})
				_ = course1.ID.Set("1")
				_ = course1.CourseType.Set(pb.COURSE_TYPE_LIVE.String())
				_ = course1.Country.Set(2)

				courses1 := entities_bob.Courses{course1}
				courseClass1 := make(map[pgtype.Text]pgtype.Int4Array)
				courseClass1[course1.ID] = database.Int4Array([]int32{1})

				enChapters1 := map[string]*entities_bob.Chapter{}
				enChapters1[chapter.ID.String] = chapter

				enLessons1 := []*entities_bob.Lesson{}
				enLessons1 = append(enLessons, lesson)

				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)

				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{},
					ClassIDs:  []int{1},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_LIVE.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
				}
				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses1, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass1, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{}, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{"1"}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons1, nil)
				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(enChapters, 1, nil)
			},
		},
		"err query FindJoined": {
			ctx:         interceptors.ContextWithUserID(ctx, "err query FindJoined"),
			req:         &pb.RetrieveCoursesRequest{},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err query count course": {
			ctx:         interceptors.ContextWithUserID(ctx, "err query count course"),
			req:         &pb.RetrieveCoursesRequest{},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)
				courseRepo.On("CountCourses", ctx, mock.Anything, mock.Anything).Once().Return(0, pgx.ErrTxClosed)
			},
		},
		"err query retrieve course": {
			ctx:         interceptors.ContextWithUserID(ctx, "err query retrieve course"),
			req:         &pb.RetrieveCoursesRequest{},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)

				courseRepo.On("CountCourses", ctx, mock.Anything, mock.Anything).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.Courses{}, pgx.ErrTxClosed)
			},
		},
		"err retrieve course class": {
			ctx:         interceptors.ContextWithUserID(ctx, "err retrieve course class"),
			req:         &pb.RetrieveCoursesRequest{},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)

				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{constants.ManabieSchool, 2},
					ClassIDs:  []int{},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
				}
				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)
				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass, pgx.ErrNoRows)
			},
		},
		"happy case retrieve all": {
			ctx:         interceptors.ContextWithUserID(ctx, "happy case retrieve all"),
			req:         &pb.RetrieveCoursesRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)

				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{constants.ManabieSchool, 2},
					ClassIDs:  []int{},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
				}
				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{}, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{course.ID.String}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(enChapters, 1, nil)
			},
		},
		"valid query assigned course": {
			ctx: interceptors.ContextWithUserID(ctx, "valid query assigned course"),
			req: &pb.RetrieveCoursesRequest{
				IsAssigned: true,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)

				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{constants.ManabieSchool},
					ClassIDs:  []int{1},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
				}
				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)
				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{}, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{course.ID.String}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(enChapters, 1, nil)
			},
		},
		"valid query course by class id": {
			ctx: interceptors.ContextWithUserID(ctx, "valid query course by class id"),
			req: &pb.RetrieveCoursesRequest{
				ClassId: 100,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)
				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{},
					ClassIDs:  []int{100},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
				}

				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{}, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{course.ID.String}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(enChapters, 1, nil)
			},
		},
		"retrive course with status": {
			ctx: interceptors.ContextWithUserID(ctx, "retrive course with status"),
			req: &pb.RetrieveCoursesRequest{
				ClassId:      100,
				CourseStatus: pb.COURSE_STATUS_ACTIVE,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)
				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{},
					ClassIDs:  []int{100},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_ACTIVE.String(),
				}

				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[string][]string{}, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{course.ID.String}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(enChapters, 1, nil)
			},
		},
	}
	for name, testCase := range testcases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.RetrieveCoursesRequest)
			_, err := s.RetrieveCourses(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestCourseService_HandleRetrieveCourses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	classRepo := new(mock_repositories.MockClassRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	chapterRepo := new(mock_repositories.MockChapterRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)

	bookRepo := new(mock_repositories.MockBookRepo)
	bookChapterRepo := new(mock_repositories.MockBookChapterRepo)
	courseBookRepo := new(mock_repositories.MockCourseBookRepo)

	s := &CourseService{
		UserRepo:        userRepo,
		ClassRepo:       classRepo,
		CourseRepo:      courseRepo,
		CourseClassRepo: courseClassRepo,
		ChapterRepo:     chapterRepo,
		LessonRepo:      lessonRepo,
		BookRepo:        bookRepo,
		CourseBookRepo:  courseBookRepo,
		TopicRepo:       topicRepo,
		BookChapterRepo: bookChapterRepo,
	}
	class := &entities_bob.Class{}
	_ = class.ID.Set(1)
	_ = class.SchoolID.Set(constants.ManabieSchool)

	users := []*entities_bob.User{}

	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	_ = course.TeacherIDs.Set([]string{"teacherIDs"})
	_ = course.ID.Set("1")

	chapter := &entities_bob.Chapter{}
	database.AllNullEntity(chapter)
	_ = chapter.ID.Set("1")
	_ = chapter.Country.Set(pb.COUNTRY_VN.String())
	_ = chapter.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = chapter.Grade.Set(8)
	_ = chapter.SchoolID.Set(constants.ManabieSchool)

	lesson := &entities_bob.Lesson{}
	database.AllNullEntity(lesson)
	_ = lesson.CourseID.Set("1")
	_ = lesson.LessonID.Set("1")

	courses := entities_bob.Courses{course}
	courseClass := make(map[pgtype.Text]pgtype.Int4Array)

	enChapters := []*entities_bob.Chapter{}
	enChapters = append(enChapters, chapter)

	enLessons := []*entities_bob.Lesson{}
	enLessons = append(enLessons, lesson)

	book := &entities_bob.Book{}
	database.AllNullEntity(book)
	_ = book.ID.Set("1")
	_ = book.Country.Set(pb.COUNTRY_VN.String())
	_ = book.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = book.Grade.Set(8)
	_ = book.SchoolID.Set(constants.ManabieSchool)

	enCourseBook := map[string][]string{}
	enCourseBook[course.ID.String] = []string{book.ID.String}

	enBooks := map[string]*entities_bob.Book{}
	enBooks[book.ID.String] = book

	enTopics := []*entities_bob.Topic{}
	topic := &entities_bob.Topic{}
	database.AllNullEntity(topic)
	_ = topic.ID.Set("1")
	_ = topic.ChapterID.Set(chapter.ID.String)
	_ = topic.Country.Set(pb.COUNTRY_VN.String())
	_ = topic.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = topic.Grade.Set(8)
	_ = topic.SchoolID.Set(constants.ManabieSchool)
	_ = topic.DisplayOrder.Set(1)
	_ = topic.CreatedAt.Set(time.Now())
	_ = topic.UpdatedAt.Set(time.Now())

	enTopics = append(enTopics, topic)

	enBookChapters := map[string][]*entities_bob.BookChapter{}
	bookChapter := &entities_bob.BookChapter{}
	database.AllNullEntity(bookChapter)
	bookChapter.BookID.Set(book.ID.String)
	bookChapter.ChapterID.Set(chapter.ID.String)

	enBookChapters[book.ID.String] = []*entities_bob.BookChapter{bookChapter}

	testcases := map[string]TestCase{
		"retrieve courses empty": {
			ctx: interceptors.ContextWithUserID(ctx, "retrieve courses empty"),
			req: &repositories.CourseQuery{
				Name:      "",
				Subject:   pb.SUBJECT_NONE.String(),
				Grade:     0,
				SchoolIDs: []int{constants.ManabieSchool, 1},
				ClassIDs:  []int{},
				Limit:     10,
				Offset:    -10,
				Type:      pb.COURSE_TYPE_CONTENT.String(),
				Status:    "",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     0,
					SchoolIDs: []int{constants.ManabieSchool, 1},
					ClassIDs:  []int{},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    "",
				}

				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(0, nil)
			},
		},
		"retrieve content course have a book": {
			ctx: interceptors.ContextWithUserID(ctx, "retrieve content course have a book"),
			req: &repositories.CourseQuery{
				Name:      "",
				Subject:   pb.SUBJECT_NONE.String(),
				Grade:     1,
				SchoolIDs: []int{constants.ManabieSchool, 1},
				ClassIDs:  []int{},
				Limit:     10,
				Offset:    -10,
				Type:      pb.COURSE_TYPE_CONTENT.String(),
				Status:    pb.COURSE_STATUS_NONE.String(),
				Countries: []string{pb.COUNTRY_VN.String()},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				query := &repositories.CourseQuery{
					Name:      "",
					Subject:   pb.SUBJECT_NONE.String(),
					Grade:     1,
					SchoolIDs: []int{constants.ManabieSchool, 1},
					ClassIDs:  []int{},
					Limit:     10,
					Offset:    -10,
					Type:      pb.COURSE_TYPE_CONTENT.String(),
					Status:    pb.COURSE_STATUS_NONE.String(),
					Countries: []string{pb.COUNTRY_VN.String()},
				}

				courseRepo.On("CountCourses", ctx, mock.Anything, query).Once().Return(1, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, query).Once().Return(courses, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(users, nil)

				courseClassRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseClass, nil)
				courseBookRepo.On("FindByCourseIDs", ctx, mock.Anything, mock.Anything).Once().Return(enCourseBook, nil)

				bookChapterRepo.On("FindByBookIDs", ctx, mock.Anything, mock.Anything).Once().Return(enBookChapters, nil)

				lessonRepo.On("Find", ctx, mock.Anything, &repositories.LessonFilter{
					CourseID:  database.TextArray([]string{course.ID.String}),
					TeacherID: database.TextArray(nil),
					LessonID:  database.TextArray(nil),
				}).Once().Return(enLessons, nil)

				chapterRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(enChapters, 1, nil)

				topicRepo.On("FindByChapterIds", ctx, mock.Anything, mock.Anything).Once().Return(enTopics, nil)
			},
		},
	}
	for name, testCase := range testcases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*repositories.CourseQuery)
			_, err := s.handleRetrieveCourses(testCase.ctx, req)

			if testCase.expectedErr != nil {
				assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestCourseService_IsJoinCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	classRepo := new(mock_repositories.MockClassRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	chapterRepo := new(mock_repositories.MockChapterRepo)
	bookRepo := new(mock_repositories.MockBookRepo)
	courseBookRepo := new(mock_repositories.MockCourseBookRepo)

	s := &CourseService{
		ClassRepo:       classRepo,
		CourseRepo:      courseRepo,
		CourseClassRepo: courseClassRepo,
		ChapterRepo:     chapterRepo,
		BookRepo:        bookRepo,
		CourseBookRepo:  courseBookRepo,
	}
	class := &entities_bob.Class{}
	_ = class.ID.Set(1)
	_ = class.SchoolID.Set(2)

	classes := []*entities_bob.Class{
		class,
	}

	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	_ = course.TeacherIDs.Set([]string{"teacherIDs"})
	_ = course.ID.Set("1")
	courses := entities_bob.Courses{course}

	type CourseUserID struct {
		CourseID string
		UserID   string
	}
	userID := "user-id"
	testcases := map[string]TestCase{
		"err query": {
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			req:         &CourseUserID{CourseID: "1", UserID: userID},
			expectedErr: fmt.Errorf("CourseRepo.RetrieveByIDs: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.Courses{}, pgx.ErrTxClosed)
			},
		},
		"course not found": {
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			req:         &CourseUserID{CourseID: "not-found", UserID: userID},
			expectedErr: status.Error(codes.NotFound, "cannot find course"),
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.Courses{}, nil)
			},
		},
		"not found class": {
			ctx:         interceptors.ContextWithUserID(ctx, "userID-11"),
			req:         &CourseUserID{CourseID: "not-found", UserID: userID},
			expectedErr: status.Error(codes.NotFound, "user do not join this course"),
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Class{}, nil)
			},
		},
		"not found course class": {
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			req:         &CourseUserID{CourseID: "1", UserID: userID},
			expectedErr: status.Error(codes.NotFound, "user do not join classes course"),
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text]pgtype.Int4Array{}, nil)
			},
		},
		"user dont join classes course": {
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			req:         &CourseUserID{CourseID: "1", UserID: userID},
			expectedErr: status.Error(codes.NotFound, "user do not join this course"),
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text]pgtype.Int4Array{
					course.ID: database.Int4Array([]int32{99}),
				}, nil)
			},
		},
		"user join this course": {
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			req:         &CourseUserID{CourseID: "1", UserID: userID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(classes, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text]pgtype.Int4Array{
					course.ID: database.Int4Array([]int32{class.ID.Int}),
				}, nil)
			},
		},
		"teacher join this course": {
			ctx:         interceptors.ContextWithUserID(ctx, userID),
			req:         &CourseUserID{CourseID: "1", UserID: userID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseTeacher := &entities_bob.Course{}
				database.AllNullEntity(courseTeacher)
				_ = courseTeacher.TeacherIDs.Set([]string{userID})
				_ = courseTeacher.ID.Set("1")
				courseTeachers := entities_bob.Courses{courseTeacher}
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courseTeachers, nil)
			},
		},
	}
	for name, testCase := range testcases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*CourseUserID)
			_, err := s.IsJoinCourse(testCase.ctx, req.CourseID, req.UserID)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestCourseService_RetrieveBooks(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classRepo := new(mock_repositories.MockClassRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	chapterRepo := new(mock_repositories.MockChapterRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)
	bookRepo := new(mock_repositories.MockBookRepo)
	courseBookRepo := new(mock_repositories.MockCourseBookRepo)

	s := &CourseService{
		UserRepo:        userRepo,
		ClassRepo:       classRepo,
		CourseRepo:      courseRepo,
		CourseClassRepo: courseClassRepo,
		ChapterRepo:     chapterRepo,
		LessonRepo:      lessonRepo,
		BookRepo:        bookRepo,
		CourseBookRepo:  courseBookRepo,
		TopicRepo:       topicRepo,
	}
	userID := "teacherIDs"

	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	_ = course.TeacherIDs.Set([]string{userID})
	_ = course.ID.Set("1")
	courses := entities_bob.Courses{course}

	book := &entities_bob.Book{}
	books := []*entities_bob.Book{}
	database.AllNullEntity(book)
	_ = book.Name.Set("book-name")
	_ = book.ID.Set("1")
	_ = book.Grade.Set(1)
	books = append(books, book)

	chapter := &entities_bob.Chapter{}
	database.AllNullEntity(chapter)
	_ = chapter.ID.Set("1")
	_ = chapter.Name.Set("chapter-name")
	_ = chapter.Grade.Set(1)
	_ = chapter.SchoolID.Set(constants.ManabieSchool)
	_ = chapter.Subject.Set(pb.SUBJECT_BIOLOGY.String())
	_ = chapter.Country.Set(pb.COUNTRY_VN.String())

	enTopics := []*entities_bob.Topic{}
	topic := &entities_bob.Topic{}
	database.AllNullEntity(topic)
	_ = topic.ID.Set("1")
	_ = topic.ChapterID.Set(chapter.ID.String)
	_ = topic.Country.Set(pb.COUNTRY_VN.String())
	_ = topic.Subject.Set(pb.SUBJECT_MATHS.String())
	_ = topic.Grade.Set(8)
	_ = topic.SchoolID.Set(constants.ManabieSchool)
	_ = topic.DisplayOrder.Set(1)
	_ = topic.CreatedAt.Set(time.Now())
	_ = topic.UpdatedAt.Set(time.Now())

	enTopics = append(enTopics, topic)

	testcases := map[string]TestCase{
		"missing course id": {
			ctx: interceptors.ContextWithUserID(ctx, userID),
			req: &pb.RetrieveBooksRequest{
				CourseId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing course id"),
			setup: func(ctx context.Context) {
			},
		},
		"invalid grade map": {
			ctx: interceptors.ContextWithUserID(ctx, userID),
			req: &pb.RetrieveBooksRequest{
				CourseId: course.ID.String,
			},
			expectedErr: status.Error(codes.InvalidArgument, "cannot find country grade map"),
			setup: func(ctx context.Context) {
				// IsJoinCourse
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)

				book := &entities_bob.Book{}
				books := []*entities_bob.Book{}
				database.AllNullEntity(book)
				_ = book.Name.Set("book-name")
				_ = book.ID.Set("1")
				_ = book.Grade.Set(1)
				_ = book.SchoolID.Set(constants.ManabieSchool)
				_ = book.Subject.Set(pb.SUBJECT_BIOLOGY.String())
				books = append(books, book)

				chapter := &entities_bob.Chapter{}
				database.AllNullEntity(chapter)
				_ = chapter.ID.Set("1")
				_ = chapter.Name.Set("chapter-name")
				chapter.Grade.Set(1)
				chapter.SchoolID.Set(constants.ManabieSchool)
				chapter.Subject.Set(pb.SUBJECT_BIOLOGY.String())
				chapter.Country.Set(pb.COUNTRY_VN.String())

				bookRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Book{book}, 1, nil)
				chapterRepo.On("FindByBookID", ctx, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Chapter{chapter}, nil)

				topicRepo.On("FindByChapterIds", ctx, mock.Anything, mock.Anything).Once().Return(enTopics, nil)
			},
		},
		"missing grade": {
			ctx: interceptors.ContextWithUserID(ctx, userID),
			req: &pb.RetrieveBooksRequest{
				CourseId: course.ID.String,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				// IsJoinCourse
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)

				book := &entities_bob.Book{}
				books := []*entities_bob.Book{}
				database.AllNullEntity(book)
				_ = book.Name.Set("book-name")
				_ = book.ID.Set("1")
				_ = book.SchoolID.Set(constants.ManabieSchool)
				_ = book.Subject.Set(pb.SUBJECT_BIOLOGY.String())
				_ = book.Country.Set(pb.COUNTRY_VN.String())

				books = append(books, book)

				chapter := &entities_bob.Chapter{}
				database.AllNullEntity(chapter)
				_ = chapter.ID.Set("1")
				_ = chapter.Name.Set("chapter-name")
				chapter.Grade.Set(1)
				chapter.SchoolID.Set(constants.ManabieSchool)
				chapter.Subject.Set(pb.SUBJECT_BIOLOGY.String())
				chapter.Country.Set(pb.COUNTRY_VN.String())

				bookRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Book{book}, 1, nil)
				chapterRepo.On("FindByBookID", ctx, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Chapter{chapter}, nil)

				topicRepo.On("FindByChapterIds", ctx, mock.Anything, mock.Anything).Once().Return(enTopics, nil)
			},
		},
		"missing subject": {
			ctx: interceptors.ContextWithUserID(ctx, userID),
			req: &pb.RetrieveBooksRequest{
				CourseId: course.ID.String,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				// IsJoinCourse
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)

				book := &entities_bob.Book{}
				books := []*entities_bob.Book{}
				database.AllNullEntity(book)
				_ = book.Name.Set("book-name")
				_ = book.ID.Set("1")
				_ = book.SchoolID.Set(constants.ManabieSchool)
				_ = book.Grade.Set(1)
				_ = book.Country.Set(pb.COUNTRY_VN.String())

				books = append(books, book)

				chapter := &entities_bob.Chapter{}
				database.AllNullEntity(chapter)
				_ = chapter.ID.Set("1")
				_ = chapter.Name.Set("chapter-name")
				chapter.Grade.Set(1)
				chapter.SchoolID.Set(constants.ManabieSchool)
				chapter.Subject.Set(pb.SUBJECT_BIOLOGY.String())
				chapter.Country.Set(pb.COUNTRY_VN.String())

				bookRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Book{book}, 1, nil)
				chapterRepo.On("FindByBookID", ctx, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Chapter{chapter}, nil)

				topicRepo.On("FindByChapterIds", ctx, mock.Anything, mock.Anything).Once().Return(enTopics, nil)
			},
		},
		"happy case retrieve all": {
			ctx: interceptors.ContextWithUserID(ctx, userID),
			req: &pb.RetrieveBooksRequest{
				CourseId: course.ID.String,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				// IsJoinCourse
				courseRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)

				book.Grade.Set(1)
				book.SchoolID.Set(constants.ManabieSchool)
				book.Subject.Set(pb.SUBJECT_BIOLOGY.String())
				book.Country.Set(pb.COUNTRY_VN.String())

				bookRepo.On("FindWithFilter", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(books, 1, nil)
				chapterRepo.On("FindByBookID", ctx, mock.Anything, mock.Anything).Once().Return([]*entities_bob.Chapter{chapter}, nil)

				topicRepo.On("FindByChapterIds", ctx, mock.Anything, mock.Anything).Once().Return(enTopics, nil)
			},
		},
	}
	for name, testCase := range testcases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.RetrieveBooksRequest)
			_, err := s.RetrieveBooks(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
