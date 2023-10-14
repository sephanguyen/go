package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestListGradeBookView(t *testing.T) {
	t.Parallel()
	ctx := interceptors.NewIncomingContext(context.Background())

	examLORepo := new(mock_repositories.MockExamLORepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	courseClient := new(mock_services.BobCourseClientServiceClient)

	s := StatisticService{
		ExamLORepo:   examLORepo,
		StudentRepo:  studentRepo,
		CourseClient: courseClient,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &sspb.GradeBookRequest{
				StudyPlanIds: []string{"study-plan-id-1", "study-plan-id-2"},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentRepo.On("FilterByGradeBookView", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*repositories.StudentInfo{}, nil)
				examLORepo.On("GetScores", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.ExamLoScore{
					{
						CourseID:      database.Text("course-id"),
						StudyPlanID:   database.Text("study-plan-id"),
						StudyPlanName: database.Text("study-plan-name"),
						StudentID:     database.Text("student-id"),
					},
				}, nil)
				courseClient.On("RetrieveCoursesByIDs", mock.Anything, mock.Anything).Return(&pb.RetrieveCoursesResponse{}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)

		_, err := s.ListGradeBook(testCase.ctx, testCase.req.(*sspb.GradeBookRequest))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestGetStudentProgress(t *testing.T) {
	t.Parallel()
	learningMaterialRepo := &mock_repositories.MockLearningMaterialRepo{}
	statisticsRepo := &mock_repositories.MockStatisticsRepo{}

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StatisticService{
		DB:                   mockDB,
		LearningMaterialRepo: learningMaterialRepo,
		StatisticsRepo:       statisticsRepo,
	}

	time := time.Now()

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.GetStudentProgressRequest{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId: "study_plan_id",
					StudentId:   wrapperspb.String("student_id"),
				},
				CourseId: "course_id",
			},
			expectedErr: nil,
			expectedResp: &sspb.GetStudentProgressResponse{
				StudentStudyPlanProgresses: []*sspb.GetStudentProgressResponse_StudentStudyPlanProgress{
					{
						StudyPlanId: "study_plan_id",
						TopicProgress: []*sspb.StudentTopicStudyProgress{
							{
								TopicId:                "topic_id",
								CompletedStudyPlanItem: wrapperspb.Int32(1),
								TotalStudyPlanItem:     wrapperspb.Int32(2),
								AverageScore:           wrapperspb.Int32(50),
								TopicName:              "topic_name",
							},
						},
						ChapterProgress: []*sspb.StudentChapterStudyProgress{
							{
								ChapterId:    "chapter_id",
								AverageScore: wrapperspb.Int32(50),
							},
						},
						LearningMaterialResults: []*sspb.LearningMaterialResult{
							{
								LearningMaterial: &sspb.LearningMaterialBase{
									LearningMaterialId: "lm_id_1",
									TopicId:            "topic_id",
									Name:               "name_1",
									Type:               "LEARNING_MATERIAL_FLASH_CARD",
									DisplayOrder:       wrapperspb.Int32(1),
								},
								IsCompleted: true,
								Crown:       0,
							},
							{
								LearningMaterial: &sspb.LearningMaterialBase{
									LearningMaterialId: "lm_id_2",
									TopicId:            "topic_id",
									Name:               "name_2",
									Type:               "LEARNING_MATERIAL_FLASH_CARD",
									DisplayOrder:       wrapperspb.Int32(2),
								},
								IsCompleted: false,
								Crown:       0,
							},
						},
						StudyPlanTrees: []*sspb.StudyPlanTree{
							{
								StudyPlanId: "study_plan_id",
								BookTree: &sspb.BookTree{
									BookId:              "book_id_1",
									ChapterId:           "chapter_id",
									ChapterDisplayOrder: 1,
									TopicId:             "topic_id",
									TopicDisplayOrder:   1,
									LearningMaterialId:  "lm_id_1",
									LmDisplayOrder:      1,
									LmType:              sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD,
								},
								AvailableFrom: timestamppb.New(time),
								AvailableTo:   timestamppb.New(time),
								StartDate:     timestamppb.New(time),
								EndDate:       timestamppb.New(time),
								CompletedAt:   timestamppb.New(time),
								Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
								SchoolDate:    timestamppb.New(time),
							},
							{
								StudyPlanId: "study_plan_id",
								BookTree: &sspb.BookTree{
									BookId:              "book_id_2",
									ChapterId:           "chapter_id",
									ChapterDisplayOrder: 2,
									TopicId:             "topic_id",
									TopicDisplayOrder:   2,
									LearningMaterialId:  "lm_id_2",
									LmDisplayOrder:      2,
									LmType:              sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD,
								},
								AvailableFrom: timestamppb.New(time),
								AvailableTo:   timestamppb.New(time),
								StartDate:     timestamppb.New(time),
								EndDate:       timestamppb.New(time),
								CompletedAt:   timestamppb.New(time),
								Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
								SchoolDate:    timestamppb.New(time),
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				statisticsRepo.On("GetStudentProgress", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.LearningMaterialProgress{
					{
						StudyPlanID:         database.Text("study_plan_id"),
						LearningMaterialID:  database.Text("lm_id_1"),
						Type:                database.Text("LEARNING_MATERIAL_FLASH_CARD"),
						Name:                database.Text("name_1"),
						BookID:              database.Text("book_id_1"),
						ChapterID:           database.Text("chapter_id"),
						ChapterDisplayOrder: database.Int2(1),
						TopicID:             database.Text("topic_id"),
						TopicDisplayOrder:   database.Int2(1),
						LmDisplayOrder:      database.Int2(1),
						IsCompleted:         database.Bool(true),
						HighestScore:        database.Int2(50),
						AvailableFrom:       database.Timestamptz(time),
						AvailableTo:         database.Timestamptz(time),
						StartDate:           database.Timestamptz(time),
						EndDate:             database.Timestamptz(time),
						CompletedAt:         database.Timestamptz(time),
						Status:              database.Text("STUDY_PLAN_ITEM_STATUS_ACTIVE"),
						SchoolDate:          database.Timestamptz(time),
					},
					{
						StudyPlanID:         database.Text("study_plan_id"),
						LearningMaterialID:  database.Text("lm_id_2"),
						Type:                database.Text("LEARNING_MATERIAL_FLASH_CARD"),
						Name:                database.Text("name_2"),
						BookID:              database.Text("book_id_2"),
						ChapterID:           database.Text("chapter_id"),
						ChapterDisplayOrder: database.Int2(2),
						TopicID:             database.Text("topic_id"),
						TopicDisplayOrder:   database.Int2(2),
						LmDisplayOrder:      database.Int2(2),
						IsCompleted:         database.Bool(false),
						HighestScore:        database.Int2(50),
						AvailableFrom:       database.Timestamptz(time),
						AvailableTo:         database.Timestamptz(time),
						StartDate:           database.Timestamptz(time),
						EndDate:             database.Timestamptz(time),
						CompletedAt:         database.Timestamptz(time),
						Status:              database.Text("STUDY_PLAN_ITEM_STATUS_ACTIVE"),
						SchoolDate:          database.Timestamptz(time),
					},
				}, []*repositories.StudentTopicProgress{
					{
						StudentID:        database.Text("student_id"),
						StudyPlanID:      database.Text("study_plan_id"),
						ChapterID:        database.Text("chapter_id"),
						TopicID:          database.Text("topic_id"),
						TopicName:        database.Text("topic_name"),
						CompletedSPItems: database.Int2(1),
						TotalSpItems:     database.Int2(2),
						AverageScore:     database.Int2(50),
					},
				}, []*repositories.StudentChapterProgress{
					{
						StudentID:    database.Text("student_id"),
						StudyPlanID:  database.Text("study_plan_id"),
						ChapterID:    database.Text("chapter_id"),
						AverageScore: database.Int2(50),
					},
				}, nil)
			},
		},
		{
			name: "error empty course id",
			req: &sspb.GetStudentProgressRequest{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudentId: wrapperspb.String("student_id"),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "course_id is required"),
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
				"token":   []string{"token"},
				"pkg":     []string{"package"},
				"version": []string{"version"},
			})
			testCase.setup(ctx)
			resp, err := svc.GetStudentProgress(ctx, testCase.req.(*sspb.GetStudentProgressRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})

	}
}

func TestCourseStatistic(t *testing.T) {
	t.Parallel()
	courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
	courseStudentAccessPathRepo := &mock_repositories.MockCourseStudentAccessPathRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StatisticService{
		DB:                          mockDB,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		CourseStudentRepo:           courseStudentRepo,
		CourseStudentAccessPathRepo: courseStudentAccessPathRepo,
		StudentRepo:                 studentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.CourseStatisticRequest{
				CourseId:    "course-1",
				StudyPlanId: "std-1",
				ClassId:     []string{},
			},

			expectedErr: nil,
			expectedResp: &sspb.CourseStatisticResponse{TopicStatistic: []*sspb.CourseStatisticResponse_TopicStatistic{
				{
					TopicId:              "tp-1",
					TotalAssignedStudent: int32(2),
					CompletedStudent:     int32(2),
					AverageScore:         int32(80),
					LearningMaterialStatistic: []*sspb.CourseStatisticResponse_TopicStatistic_LearningMaterialStatistic{
						{
							LearningMaterialId:   "lm-1",
							TotalAssignedStudent: int32(2),
							CompletedStudent:     int32(2),
							AverageScore:         int32(80),
						},
						{
							LearningMaterialId:   "lm-2",
							TotalAssignedStudent: int32(2),
							CompletedStudent:     int32(2),
							AverageScore:         int32(80),
						},
					},
				},
			}},
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentByCourseID", mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, nil)

				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				studentRepo.On("FilterOutDeletedStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				courseStudyPlanRepo.On("ListCourseStatisticV3", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*repositories.TopicStatistic{{
						TopicID:            "tp-1",
						TotalAssignStudent: database.Int4(2),
						CompletedStudent:   database.Int4(2),
						AverageScore:       database.Int4(80),
					}},
					[]*repositories.LearningMaterialStatistic{
						{
							TopicID:            "tp-1",
							LearningMaterialID: "lm-1",
							TotalAssignStudent: database.Int4(2),
							CompletedStudent:   database.Int4(2),
							AverageScore:       database.Int4(80),
							AverageScoreRaw:    database.Int4(80),
						},
						{
							TopicID:            "tp-1",
							LearningMaterialID: "lm-2",
							TotalAssignStudent: database.Int4(2),
							CompletedStudent:   database.Int4(2),
							AverageScore:       database.Int4(80),
							AverageScoreRaw:    database.Int4(80),
						},
					}, nil)
			},
		},
		{
			name: "unhappy case",
			req: &sspb.CourseStatisticRequest{
				CourseId:    "course-1",
				StudyPlanId: "std-1",
				ClassId:     []string{},
			},

			expectedErr:  status.Errorf(codes.Internal, "Topic not exist in LearningMaterialStatistic"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentByCourseID", mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, nil)

				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				studentRepo.On("FilterOutDeletedStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				courseStudyPlanRepo.On("ListCourseStatisticV3", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*repositories.TopicStatistic{{
						TopicID:            "tp-1",
						TotalAssignStudent: database.Int4(2),
						CompletedStudent:   database.Int4(2),
						AverageScore:       database.Int4(80),
					}},
					[]*repositories.LearningMaterialStatistic{
						{
							TopicID:            "tp-2",
							LearningMaterialID: "lm-1",
							TotalAssignStudent: database.Int4(2),
							CompletedStudent:   database.Int4(2),
							AverageScore:       database.Int4(80),
							AverageScoreRaw:    database.Int4(80),
						},
						{
							TopicID:            "tp-3",
							LearningMaterialID: "lm-2",
							TotalAssignStudent: database.Int4(2),
							CompletedStudent:   database.Int4(2),
							AverageScore:       database.Int4(80),
							AverageScoreRaw:    database.Int4(80),
						},
					}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
				"token":   []string{"token"},
				"pkg":     []string{"package"},
				"version": []string{"version"},
			})
			testCase.setup(ctx)
			resp, err := svc.RetrieveCourseStatistic(ctx, testCase.req.(*sspb.CourseStatisticRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})

	}
}
func TestCourseStatisticValidateReq(t *testing.T) {
	t.Parallel()

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.CourseStatisticRequest{
				CourseId:    "",
				StudyPlanId: "std-1",
				ClassId:     []string{},
			},
			expectedErr: errors.New("Missing course"),
		},
		{
			name: "unhappy case",
			req: &sspb.CourseStatisticRequest{
				CourseId:    "course-1",
				StudyPlanId: "",
				ClassId:     []string{},
			},
			expectedErr: errors.New("Missing study plan"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateCourseStatisticRequest(testCase.req.(*sspb.CourseStatisticRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})

	}
}

func TestStatisticService_ListSubmissions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockDB := &mock_database.Ext{}
	mockStudentSubmissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	s := &StatisticService{
		DB:                    mockDB,
		StudentSubmissionRepo: mockStudentSubmissionRepo,
	}

	now := time.Now()
	later := now.AddDate(0, 0, 10)

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.ListSubmissionsRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "sp-id-1",
						LearningMaterialId: "lm-id-1",
						StudentId:          wrapperspb.String("student-id-1"),
					},
				},
			},
			setup: func(ctx context.Context) {
				studyPlanItemIdentities := []*repositories.StudyPlanItemIdentity{
					{
						StudyPlanID:        database.Text("sp-id-1"),
						LearningMaterialID: database.Text("lm-id-1"),
						StudentID:          database.Text("student-id-1"),
					},
				}
				studentSubmissions := []*repositories.StudentSubmissionInfo{
					{
						StudentSubmission: entities.StudentSubmission{
							ID:                 database.Text("student-submission-id-1"),
							StudyPlanID:        database.Text("sp-id-1"),
							LearningMaterialID: database.Text("lm-id-1"),
							StudentID:          database.Text("student-id-1"),
							BaseEntity: entities.BaseEntity{
								CreatedAt: database.Timestamptz(now),
								UpdatedAt: database.Timestamptz(now),
							},
						},
						CourseID:  database.Text("course-id-1"),
						StartDate: database.Timestamptz(now),
						EndDate:   database.Timestamptz(later),
					},
				}
				mockStudentSubmissionRepo.
					On("RetrieveByStudyPlanIdentities", mock.Anything, mock.Anything, studyPlanItemIdentities).
					Once().
					Return(studentSubmissions, nil)
			},
			expectedResp: &sspb.ListSubmissionsResponse{
				Submissions: []*sspb.Submission{{
					SubmissionId: "student-submission-id-1",
					StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
						StudyPlanId:        "sp-id-1",
						LearningMaterialId: "lm-id-1",
						StudentId:          wrapperspb.String("student-id-1"),
					},
					CourseId:  "course-id-1",
					StartDate: timestamppb.New(now),
					EndDate:   timestamppb.New(later),
					CreatedAt: timestamppb.New(now),
					UpdatedAt: timestamppb.New(now),
				}},
			},
		},
		{
			name: "invalid argument",
			req: &sspb.ListSubmissionsRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId: "",
					},
				},
			},
			setup:       func(ctx context.Context) {},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateListSubmissionsReq: StudyPlanItemIdentities[0]: StudyPlanId must not empty").Error()),
		},
		{
			name: "error query submissions",
			req: &sspb.ListSubmissionsRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "sp-id-1",
						LearningMaterialId: "lm-id-1",
						StudentId:          wrapperspb.String("student-id-1"),
					},
				},
			},
			setup: func(ctx context.Context) {
				studyPlanItemIdentities := []*repositories.StudyPlanItemIdentity{
					{
						StudyPlanID:        database.Text("sp-id-1"),
						LearningMaterialID: database.Text("lm-id-1"),
						StudentID:          database.Text("student-id-1"),
					},
				}
				mockStudentSubmissionRepo.
					On("RetrieveByStudyPlanIdentities", mock.Anything, mock.Anything, studyPlanItemIdentities).
					Once().
					Return(nil, pgx.ErrTxClosed)
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.StudentSubmissionRepo.RetrieveByStudyPlanIdentities: %w", pgx.ErrTxClosed).Error()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := s.ListSubmissions(ctx, testCase.req.(*sspb.ListSubmissionsRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})

	}
}

func TestStatisticService_validateListSubmissionsReq(t *testing.T) {
	t.Parallel()

	s := &StatisticService{}

	testCases := []TestCase{
		{
			name: "missing study plan id",
			req: &sspb.ListSubmissionsRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId: "",
					},
				},
			},
			expectedErr: fmt.Errorf("StudyPlanItemIdentities[0]: StudyPlanId must not empty"),
		},
		{
			name: "missing learning material id",
			req: &sspb.ListSubmissionsRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "sp-id-1",
						LearningMaterialId: "",
					},
				},
			},
			expectedErr: fmt.Errorf("StudyPlanItemIdentities[0]: LearningMaterialId must not empty"),
		},
		{
			name: "missing student id",
			req: &sspb.ListSubmissionsRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "sp-id-1",
						LearningMaterialId: "lm-id-1",
						StudentId:          wrapperspb.String(""),
					},
				},
			},
			expectedErr: fmt.Errorf("StudyPlanItemIdentities[0]: StudentId must not empty"),
		},
	}

	for _, testCase := range testCases {
		err := s.validateListSubmissionsReq(testCase.req.(*sspb.ListSubmissionsRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStatisticService_RetrieveSchoolHistoryByStudentInCourse(t *testing.T) {
	t.Parallel()
	ctx := interceptors.NewIncomingContext(context.Background())

	mockDB := &mock_database.Ext{}
	mockCourseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
	mockStudentReaderClient := &mock_services.BobStudentReaderServiceClient{}
	s := &StatisticService{
		DB:                  mockDB,
		CourseStudentRepo:   mockCourseStudentRepo,
		StudentReaderClient: mockStudentReaderClient,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &sspb.RetrieveSchoolHistoryByStudentInCourseRequest{
				CourseId: "course_id",
			},
			expectedResp: &sspb.RetrieveSchoolHistoryByStudentInCourseResponse{
				Schools: map[string]*sspb.RetrieveSchoolHistoryByStudentInCourseResponse_School{
					"school_01": {SchoolId: "school_01", SchoolName: "01"},
					"school_02": {SchoolId: "school_02", SchoolName: "02"},
				},
			},
			setup: func(ctx context.Context) {
				studentIDs := []string{"student_01", "student_02"}
				mockCourseStudentRepo.On("FindStudentByCourseID", mock.Anything, mockDB, mock.Anything).Once().Return(studentIDs, nil)
				mockStudentReaderClient.On(
					"RetrieveStudentSchoolHistory", mock.Anything, &bpb.RetrieveStudentSchoolHistoryRequest{
						StudentIds: studentIDs,
					},
				).Once().Return(&bpb.RetrieveStudentSchoolHistoryResponse{
					Schools: map[string]*bpb.RetrieveStudentSchoolHistoryResponse_School{
						"school_01": {SchoolId: "school_01", SchoolName: "01"},
						"school_02": {SchoolId: "school_02", SchoolName: "02"},
					},
				}, nil)
			},
		},
		{
			name: "don't have student in course",
			ctx:  ctx,
			req: &sspb.RetrieveSchoolHistoryByStudentInCourseRequest{
				CourseId: "course_id",
			},
			expectedResp: &sspb.RetrieveSchoolHistoryByStudentInCourseResponse{},
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("FindStudentByCourseID", mock.Anything, mockDB, mock.Anything).Once().Return([]string{}, nil)
			},
		},
		{
			name: "error case",
			ctx:  ctx,
			req: &sspb.RetrieveSchoolHistoryByStudentInCourseRequest{
				CourseId: "course_id",
			},
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("FindStudentByCourseID", mock.Anything, mockDB, mock.Anything).Once().Return([]string{}, pgx.ErrTxClosed)
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.CourseStudentRepo.FindStudentByCourseID %w", pgx.ErrTxClosed).Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := s.RetrieveSchoolHistoryByStudentInCourse(ctx, testCase.req.(*sspb.RetrieveSchoolHistoryByStudentInCourseRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestStatisticService_ListTagByStudentInCourse(t *testing.T) {
	mockDB := &mock_database.Ext{}
	ctx := interceptors.NewIncomingContext(context.Background())
	courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
	svc := &StatisticService{
		DB:                mockDB,
		CourseStudentRepo: courseStudentRepo,
	}

	testCases := []TestCase{
		{
			name: "cannot empty course_id",
			ctx:  ctx,
			req: &sspb.ListTagByStudentInCourseRequest{
				CourseId: "",
			},
			setup:        func(ctx context.Context) {},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Errorf("cannot empty course_id").Error()),
		},
		{
			name: "empty student tag",
			ctx:  ctx,
			req: &sspb.ListTagByStudentInCourseRequest{
				CourseId: "course_id",
			},
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentTagByCourseID", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudentTag{}, nil)
			},
			expectedResp: &sspb.ListTagByStudentInCourseResponse{},
			expectedErr:  nil,
		},
		{
			name: "happy case",
			ctx:  ctx,
			req: &sspb.ListTagByStudentInCourseRequest{
				CourseId: "course_id",
			},
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentTagByCourseID", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*entities.StudentTag{
						{ID: database.Text("tag_id_1"), Name: database.Text("tag_name_1")},
						{ID: database.Text("tag_id_2"), Name: database.Text("tag_name_2")},
					},
					nil)
			},
			expectedResp: &sspb.ListTagByStudentInCourseResponse{
				StudentTags: []*sspb.ListTagByStudentInCourseResponse_StudentTag{
					{
						TagId:   "tag_id_1",
						TagName: "tag_name_1",
					},
					{
						TagId:   "tag_id_2",
						TagName: "tag_name_2",
					},
				},
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		testCase.setup(testCase.ctx)
		_, err := svc.ListTagByStudentInCourse(testCase.ctx, testCase.req.(*sspb.ListTagByStudentInCourseRequest))
		if testCase.expectedErr != nil {
			fmt.Println(err.Error())
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

// write unit test for RetrieveCourseStatisticV2 function

func TestStatisticService_RetrieveCourseStatisticV2(t *testing.T) {
	t.Parallel()
	courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StatisticService{
		DB:                  mockDB,
		CourseStudyPlanRepo: courseStudyPlanRepo,
		CourseStudentRepo:   courseStudentRepo,
		StudentRepo:         studentRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.CourseStatisticRequest{
				CourseId:    "course-1",
				StudyPlanId: "std-1",
				ClassId:     []string{},
			},

			expectedErr: nil,
			expectedResp: &sspb.CourseStatisticResponse{TopicStatistic: []*sspb.CourseStatisticResponse_TopicStatistic{
				{
					TopicId:              "tp-1",
					TotalAssignedStudent: int32(2),
					CompletedStudent:     int32(2),
					AverageScore:         int32(80),
					LearningMaterialStatistic: []*sspb.CourseStatisticResponse_TopicStatistic_LearningMaterialStatistic{
						{
							LearningMaterialId:   "lm-1",
							TotalAssignedStudent: int32(2),
							CompletedStudent:     int32(2),
							AverageScore:         int32(80),
						},
						{
							LearningMaterialId:   "lm-2",
							TotalAssignedStudent: int32(2),
							CompletedStudent:     int32(2),
							AverageScore:         int32(80),
						},
					},
				},
			}},
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentByCourseID", mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				studentRepo.On("FilterOutDeletedStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, nil)

				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				courseStudyPlanRepo.On("ListCourseStatisticV4", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*repositories.TopicStatistic{{
						TopicID:            "tp-1",
						TotalAssignStudent: database.Int4(2),
						CompletedStudent:   database.Int4(2),
						AverageScore:       database.Int4(80),
					}},
					[]*repositories.LearningMaterialStatistic{
						{
							TopicID:            "tp-1",
							LearningMaterialID: "lm-1",
							TotalAssignStudent: database.Int4(2),
							CompletedStudent:   database.Int4(2),
							AverageScore:       database.Int4(80),
							AverageScoreRaw:    database.Int4(80),
						},
						{
							TopicID:            "tp-1",
							LearningMaterialID: "lm-2",
							TotalAssignStudent: database.Int4(2),
							CompletedStudent:   database.Int4(2),
							AverageScore:       database.Int4(80),
							AverageScoreRaw:    database.Int4(80),
						},
					}, nil)
			},
		},
		{
			name: "unhappy case",
			req: &sspb.CourseStatisticRequest{
				CourseId:    "course-1",
				StudyPlanId: "std-1",
				ClassId:     []string{},
			},

			expectedErr:  status.Errorf(codes.Internal, "Topic not exist in LearningMaterialStatistic"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentByCourseID", mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				studentRepo.On("FilterOutDeletedStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, nil)

				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				courseStudyPlanRepo.On("ListCourseStatisticV4", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*repositories.TopicStatistic{{
						TopicID:            "tp-1",
						TotalAssignStudent: database.Int4(2),
						CompletedStudent:   database.Int4(2),
						AverageScore:       database.Int4(80),
					}},
					[]*repositories.LearningMaterialStatistic{
						{
							TopicID:            "tp-2",
							LearningMaterialID: "lm-1",
							TotalAssignStudent: database.Int4(2),
							CompletedStudent:   database.Int4(2),
							AverageScore:       database.Int4(80),
							AverageScoreRaw:    database.Int4(80),
						},
						{
							TopicID:            "tp-3",
							LearningMaterialID: "lm-2",
							TotalAssignStudent: database.Int4(2),
							CompletedStudent:   database.Int4(2),
							AverageScore:       database.Int4(80),
							AverageScoreRaw:    database.Int4(80),
						},
					}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
				"token":   []string{"token"},
				"pkg":     []string{"package"},
				"version": []string{"version"},
			})
			testCase.setup(ctx)
			resp, err := svc.RetrieveCourseStatisticV2(ctx, testCase.req.(*sspb.CourseStatisticRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})

	}
}
