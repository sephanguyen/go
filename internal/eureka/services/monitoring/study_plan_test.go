package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_monitors_repositories "github.com/manabie-com/backend/mock/eureka/repositories/monitors"
	mock_alert "github.com/manabie-com/backend/mock/golibs/alert"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type TestCase struct {
	ctx          context.Context
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func Test_genTimeLCLTimeUCL(t *testing.T) {
	t.Parallel()
	type testcase struct {
		name       string
		input      int
		LCLTimeOup pgtype.Text
		ULCTimeOup pgtype.Text
	}
	tests := []testcase{
		{
			name:       "odd number",
			input:      9,
			LCLTimeOup: database.Text("10 mins"),
			ULCTimeOup: database.Text("1 mins"),
		},
		{
			name:       "even number",
			input:      10,
			LCLTimeOup: database.Text("11 mins"),
			ULCTimeOup: database.Text("1 mins"),
		},
	}
	for _, c := range tests {
		c := c
		t.Run(c.name, func(t *testing.T) {
			LCLTimeOup, ULCTimeOup := genTimeLCLTimeUCL(c.input)
			if LCLTimeOup.String != c.LCLTimeOup.String {
				t.Errorf("LCLTimeOup: expected %v, got %v", c.LCLTimeOup, LCLTimeOup)
			}
			if ULCTimeOup.String != c.ULCTimeOup.String {
				t.Errorf("ULCTimeOup: expected %v, got %v", c.ULCTimeOup, ULCTimeOup)
			}
		})
	}
}

func Test_ReVerifyMissingStudentStudyPlan(t *testing.T) {
	t.Parallel()

	mockErr := fmt.Errorf("mock-error")
	db := &mock_database.Ext{}
	StudentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	CourseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
	studyplanRepo := &mock_repositories.MockStudyPlanRepo{}
	StudyPlanMonitorRepo := &mock_monitors_repositories.MockStudyPlanMonitorRepo{}
	loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	assStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}

	monitorService := &StudyPlanMonitorService{
		DB:                          db,
		StudentStudyPlanRepo:        StudentStudyPlanRepo,
		CourseStudentRepo:           CourseStudentRepo,
		StudyPlanRepo:               studyplanRepo,
		StudyPlanMonitorRepo:        StudyPlanMonitorRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assStudyPlanItemRepo,
	}

	testcases := []TestCase{
		{
			name:         "err when soft delete",
			expectedResp: nil,
			expectedErr:  fmt.Errorf("StudyPlanMonitorService.ReVerifyMissingStudentStudyPlan.StudyPlanMonitorRepo.SoftDeleteTypeStudyPlan: %w", mockErr),
			setup: func(ctx context.Context) {
				StudyPlanMonitorRepo.On("SoftDeleteTypeStudyPlan", ctx, db, mock.Anything).Once().Return(mockErr)
			},
		},
		{
			name:         "happy case",
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				StudyPlanMonitorRepo.On("SoftDeleteTypeStudyPlan", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			err := monitorService.ReVerifyMissingStudentStudyPlan(ctx, 15)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else if tc.expectedErr == nil && err != nil {
				t.Errorf("ReVerifyMissingStudentStudyPlan: expected %v, got %v", nil, err)
			}
		})
	}
}

func Test_CollectMissingStudentStudyPlan(t *testing.T) {
	t.Parallel()

	mockErr := fmt.Errorf("mock-error")
	db := &mock_database.Ext{}
	StudentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	CourseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
	StudyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	StudyPlanMonitorRepo := &mock_monitors_repositories.MockStudyPlanMonitorRepo{}
	mockAlert := &mock_alert.SlackFactory{}

	monitorService := &StudyPlanMonitorService{
		DB:                   db,
		Alert:                mockAlert,
		Cfg:                  &configurations.Config{},
		StudentStudyPlanRepo: StudentStudyPlanRepo,
		CourseStudentRepo:    CourseStudentRepo,
		StudyPlanRepo:        StudyPlanRepo,
		StudyPlanMonitorRepo: StudyPlanMonitorRepo,
	}
	studentStudyPlans := []*entities.StudentStudyPlan{
		{
			StudentID:         database.Text("mock-student-id-1"),
			StudyPlanID:       database.Text("study-plan-id-1.1"),
			MasterStudyPlanID: database.Text("master-study-plan-id-1"),
		},
	}
	courseStudents := []*entities.CourseStudent{
		{
			CourseID:  database.Text("course-1"),
			StudentID: database.Text("student-1"),
		},
		{
			CourseID:  database.Text("course-2"),
			StudentID: database.Text("student-2"),
		},
	}
	masterStudyPlans := []*entities.StudyPlan{
		{
			ID:       database.Text("master-study-plan-id-1"),
			CourseID: database.Text("course-1"),
		},
		{
			ID:       database.Text("master-study-plan-id-2"),
			CourseID: database.Text("course-2"),
		},
	}
	testcases := []TestCase{
		{
			name:         "fail when upsert",
			expectedResp: nil,
			expectedErr:  fmt.Errorf("StudyPlanMonitorService.StudyPlanMonitorRepo.BulkUpsert: %w", mockErr),
			setup: func(ctx context.Context) {
				CourseStudentRepo.On("RetrieveByIntervalTime", ctx, db, mock.Anything).Once().Return(courseStudents, nil)
				StudentStudyPlanRepo.On("RetrieveByStudentCourse", ctx, db, mock.Anything, mock.Anything).Once().Return(studentStudyPlans, nil)
				StudyPlanRepo.On("RetrieveMasterByCourseIDs", ctx, db, mock.Anything, mock.Anything).Once().Return(masterStudyPlans, nil)
				StudyPlanMonitorRepo.On("BulkUpsert", ctx, db, mock.Anything).Once().Return(mockErr)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "fail when RetrieveMasterByCourseIDs",
			expectedResp: nil,
			expectedErr:  fmt.Errorf("StudyPlanMonitorService.StudyPlanRepo.RetrieveMasterByCourseIDs: %w", mockErr),
			setup: func(ctx context.Context) {
				CourseStudentRepo.On("RetrieveByIntervalTime", ctx, db, mock.Anything).Once().Return(courseStudents, nil)
				StudentStudyPlanRepo.On("RetrieveByStudentCourse", ctx, db, mock.Anything, mock.Anything).Once().Return(studentStudyPlans, nil)
				StudyPlanRepo.On("RetrieveMasterByCourseIDs", ctx, db, mock.Anything, mock.Anything).Once().Return(masterStudyPlans, mockErr)
				StudyPlanMonitorRepo.On("BulkUpsert", ctx, db, mock.Anything).Once().Return(nil)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "fail when RetrieveByStudentCourse",
			expectedResp: nil,
			expectedErr:  fmt.Errorf("StudyPlanMonitorService.StudentStudyPlanRepo.RetrieveByStudentCourse: %w", mockErr),
			setup: func(ctx context.Context) {
				CourseStudentRepo.On("RetrieveByIntervalTime", ctx, db, mock.Anything).Once().Return(courseStudents, nil)
				StudentStudyPlanRepo.On("RetrieveByStudentCourse", ctx, db, mock.Anything, mock.Anything).Once().Return(studentStudyPlans, mockErr)
				StudyPlanRepo.On("RetrieveMasterByCourseIDs", ctx, db, mock.Anything, mock.Anything).Once().Return(masterStudyPlans, nil)
				StudyPlanMonitorRepo.On("BulkUpsert", ctx, db, mock.Anything).Once().Return(nil)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "happy case",
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				CourseStudentRepo.On("RetrieveByIntervalTime", ctx, db, mock.Anything).Once().Return(courseStudents, nil)
				StudentStudyPlanRepo.On("RetrieveByStudentCourse", ctx, db, mock.Anything, mock.Anything).Once().Return(studentStudyPlans, nil)
				StudyPlanRepo.On("RetrieveMasterByCourseIDs", ctx, db, mock.Anything, mock.Anything).Once().Return(masterStudyPlans, nil)
				StudyPlanMonitorRepo.On("BulkUpsert", ctx, db, mock.Anything).Once().Return(nil)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			err := monitorService.CollectMissingStudentStudyPlan(ctx, 15)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else if tc.expectedErr == nil && err != nil {
				t.Errorf("CollectMissingStudentStudyPlan: expected %v, got %v", nil, err)
			}
		})
	}
}

func Test_convert2BookAssignmentIDs(t *testing.T) {
	t.Parallel()
	type testcase struct {
		name              string
		input1            []*entities.BookAssignment
		mapBookAssignment map[string][]string
	}
	mapBookAssignmentHappyCase := make(map[string][]string)
	mapBookAssignmentHappyCase["mock-book-1"] = []string{"mock-assignment-1", "mock-assignment-2"}
	tests := []testcase{
		{
			name: "happy case",
			input1: []*entities.BookAssignment{
				{
					Assignment: entities.Assignment{
						ID: database.Text("mock-assignment-1"),
						Content: database.JSONB(entities.AssignmentContent{
							TopicID: "mock-topic-1",
						}),
					},
					BookID: database.Text("mock-book-1"),
				},
				{
					Assignment: entities.Assignment{
						ID: database.Text("mock-assignment-2"),
						Content: database.JSONB(entities.AssignmentContent{
							TopicID: "mock-topic-1",
						}),
					},
					BookID: database.Text("mock-book-1"),
				},
			},
			mapBookAssignment: mapBookAssignmentHappyCase,
		},
	}
	for _, c := range tests {
		c := c
		t.Run(c.name, func(t *testing.T) {
			mapBookAssignment, _, _, _, _ := retrieveInfoBookAssignmentIDs(c.input1)
			assert.Equal(t, mapBookAssignmentHappyCase["mock-book-1"], mapBookAssignment["mock-book-1"])
		})
	}
}

func Test_CollectMissingLearningItems(t *testing.T) {
	t.Parallel()

	mockErr := fmt.Errorf("mock-error")
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	StudentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	CourseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
	StudyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	StudyPlanMonitorRepo := &mock_monitors_repositories.MockStudyPlanMonitorRepo{}
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	learningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepo{}
	loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	assStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
	mockAlert := &mock_alert.SlackFactory{}

	intervalTime := database.Text("16 mins")
	monitorService := &StudyPlanMonitorService{
		DB:                          db,
		Cfg:                         &configurations.Config{},
		Logger:                      *zap.NewNop(),
		Alert:                       mockAlert,
		StudentStudyPlanRepo:        StudentStudyPlanRepo,
		CourseStudentRepo:           CourseStudentRepo,
		StudyPlanRepo:               StudyPlanRepo,
		StudyPlanMonitorRepo:        StudyPlanMonitorRepo,
		AssignmentRepo:              assignmentRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		LearningObjectiveRepo:       learningObjectiveRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assStudyPlanItemRepo,
	}
	bookAssignments := []*entities.BookAssignment{
		{
			Assignment: entities.Assignment{
				ID: database.Text("assignment-1"),
			},
			BookID: database.Text("book-1"),
		},
		{
			Assignment: entities.Assignment{
				ID: database.Text("assignment-2"),
			},
			BookID: database.Text("book-1"),
		},
	}
	bookLOs := []*entities.BookLearningObjective{
		{
			LearningObjective: entities.LearningObjective{
				ID: database.Text("lo-1"),
			},
			BookID: database.Text("book-1"),
		},
		{
			LearningObjective: entities.LearningObjective{
				ID: database.Text("lo-1"),
			},
			BookID: database.Text("book-1"),
		},
	}
	// combine studentid
	studyPlans := []*entities.StudyPlanCombineStudentID{
		{
			StudentID: database.Text("student-1"),
			StudyPlan: entities.StudyPlan{
				ID:              database.Text("study-plan-1"),
				BookID:          database.Text("book-1"),
				MasterStudyPlan: database.Text("master-study-plan"),
			},
		},
		{
			StudentID: database.Text("student-2"),
			StudyPlan: entities.StudyPlan{
				ID:              database.Text("study-plan-2"),
				BookID:          database.Text("book-1"),
				MasterStudyPlan: database.Text("master-study-plan"),
			},
		},
		{
			StudentID: pgtype.Text{Status: pgtype.Null},
			StudyPlan: entities.StudyPlan{
				ID:     database.Text("study-plan-3"),
				BookID: database.Text("book-1"),
			},
		},
	}

	studyPlanItems := []*entities.StudyPlanItem{
		{
			ID:          database.Text("study-plan-item-1"),
			StudyPlanID: database.Text("study-plan-1"),
			ContentStructure: database.JSONB(entities.ContentStructure{
				LoID: "lo-1",
			}),
		},
		{
			ID:          database.Text("study-plan-item-2"),
			StudyPlanID: database.Text("study-plan-2"),
			ContentStructure: database.JSONB(entities.ContentStructure{
				LoID: "assignment-1",
			}),
		},
	}

	testcases := []TestCase{
		{
			name:         "happy case",
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				assignmentRepo.On("RetrieveBookAssignmentByIntervalTime", ctx, tx, intervalTime).Once().Return(bookAssignments, nil)
				learningObjectiveRepo.On("RetrieveBookLoByIntervalTime", ctx, tx, intervalTime).Once().Return(bookLOs, nil)
				StudyPlanRepo.On("RetrieveCombineStudent", ctx, tx, mock.Anything).Once().Return(studyPlans, nil)
				studyPlanItemRepo.On("RetrieveByBookContent", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(studyPlanItems, nil)
				StudyPlanMonitorRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("FindWithFilterV2", ctx, tx, mock.Anything).Return([]*entities.StudyPlanItem{}, nil)
				studyPlanItemRepo.On("BulkSync", ctx, tx, mock.Anything).Once().Return([]*entities.StudyPlanItem{}, nil)
				loStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				assStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				StudyPlanMonitorRepo.On("MarkItemsAutoUpserted", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("BulkSync", ctx, tx, mock.Anything).Once().Return([]*entities.StudyPlanItem{}, nil)
			},
		},
		{
			name:         "fail when upsert",
			expectedResp: nil,
			expectedErr:  fmt.Errorf("StudyPlanMonitorService.CollectMissingLearningItems.StudyPlanMonitorRepo.BulkUpsert: %w", mockErr),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				assignmentRepo.On("RetrieveBookAssignmentByIntervalTime", ctx, tx, intervalTime).Once().Return(bookAssignments, nil)
				learningObjectiveRepo.On("RetrieveBookLoByIntervalTime", ctx, tx, intervalTime).Once().Return(bookLOs, nil)
				StudyPlanRepo.On("RetrieveCombineStudent", ctx, tx, mock.Anything).Once().Return(studyPlans, nil)
				studyPlanItemRepo.On("RetrieveByBookContent", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(studyPlanItems, nil)
				StudyPlanMonitorRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(mockErr)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "fail when RetrieveBookLoByIntervalTime",
			expectedResp: nil,
			expectedErr:  fmt.Errorf("StudyPlanMonitorService.CollectMissingLearningItems.LearningObjectiveRepo.RetrieveBookLoByIntervalTime: %w", mockErr),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				assignmentRepo.On("RetrieveBookAssignmentByIntervalTime", ctx, tx, intervalTime).Once().Return(bookAssignments, nil)
				learningObjectiveRepo.On("RetrieveBookLoByIntervalTime", ctx, tx, intervalTime).Once().Return(nil, mockErr)
				StudyPlanRepo.On("RetrieveCombineStudent", ctx, tx, mock.Anything).Once().Return(studyPlans, nil)
				studyPlanItemRepo.On("RetrieveByBookContent", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(studyPlanItems, nil)
				StudyPlanMonitorRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
			},
		},

		{
			name:         "fail when RetrieveBookAssignmentByIntervalTime",
			expectedResp: nil,
			expectedErr:  fmt.Errorf("StudyPlanMonitorService.CollectMissingLearningItems.AssignmentRepo.RetrieveBookAssignmentByIntervalTime: %w", mockErr),
			setup: func(ctx context.Context) {
				tx.On("Begin", ctx).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				assignmentRepo.On("RetrieveBookAssignmentByIntervalTime", ctx, tx, intervalTime).Once().Return(nil, mockErr)
				learningObjectiveRepo.On("RetrieveBookLoByIntervalTime", ctx, tx, intervalTime).Once().Return(bookLOs, nil)
				StudyPlanRepo.On("RetrieveCombineStudent", ctx, tx, mock.Anything).Once().Return(studyPlans, nil)
				studyPlanItemRepo.On("RetrieveByBookContent", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Once().Return(studyPlanItems, nil)
				StudyPlanMonitorRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				mockAlert.On("Send", mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			ctx := context.Background()
			tc.setup(ctx)
			err := monitorService.CollectMissingLearningItems(ctx, 15, "")
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else if tc.expectedErr == nil && err != nil {
				t.Errorf("CollectMissingLearningItems: expected %v, got %v", nil, err)
			}
		})
	}
}
