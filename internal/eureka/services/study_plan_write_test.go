package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_StudyPlanWriteInternalFunc(t *testing.T) {
	i := copyStudyPlanItem(&pb.StudyPlanItem{})
	assert.NotNil(t, i)

	i2 := newItem("spi-id", &pb.ContentStructure{})
	assert.NotNil(t, i2)

	i3 := newLOItem("lo-id", "item")
	assert.NotNil(t, i3)

	i4 := newAssignmentItem("assinment-id", "item")
	assert.NotNil(t, i4)

	i5 := toContentStructuresPbV2([]entities.ContentStructure{
		{
			CourseID:  "course-id",
			BookID:    "book-id",
			ChapterID: "chapter-id",
			TopicID:   "topic-id",
			LoID:      "lo-id",
		},
		{
			CourseID:     "course-id-2",
			BookID:       "book-id-2",
			ChapterID:    "chapter-id",
			TopicID:      "topic-id",
			AssignmentID: "assignment-id",
		},
	})
	assert.NotNil(t, i5)
}

func Test_ImportStudyPlan(t *testing.T) {
	db := new(mock_database.Ext)
	jsm := new(mock_nats.JetStreamManagement)
	studyPlanRepo := new(mock_repositories.MockStudyPlanRepo)
	studyPlanItemRepo := new(mock_repositories.MockStudyPlanItemRepo)
	assignmentStudyPlanItemRepo := new(mock_repositories.MockAssignmentStudyPlanItemRepo)
	loStudyPlanItemRepo := new(mock_repositories.MockLoStudyPlanItemRepo)
	studentStudyPlanRepo := new(mock_repositories.MockStudentStudyPlanRepo)
	bookChapterRepo := new(mock_repositories.MockBookChapterRepo)
	assignmentStudyPlanTaskRepo := new(mock_repositories.MockAssignStudyPlanTaskRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	courseStudyPlanRepo := new(mock_repositories.MockCourseStudyPlanRepo)
	s := &ImportService{
		DB:                          db,
		StudyPlanRepo:               studyPlanRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		StudentStudyPlan:            studentStudyPlanRepo,
		AssignStudyPlanTaskRepo:     assignmentStudyPlanTaskRepo,
		StudentRepo:                 studentRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		BookChapterRepo:             bookChapterRepo,
		JSM:                         jsm,
	}

	ctx := context.Background()
	_, err := s.ImportStudyPlan(ctx, &pb.ImportStudyPlanRequest{})
	assert.NotNil(t, err)
}

func Test_SyncStudyPlanItemsOnLOsCreated(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	jsm := new(mock_nats.JetStreamManagement)
	studyPlanRepo := new(mock_repositories.MockStudyPlanRepo)
	studyPlanItemRepo := new(mock_repositories.MockStudyPlanItemRepo)
	assignmentStudyPlanItemRepo := new(mock_repositories.MockAssignmentStudyPlanItemRepo)
	loStudyPlanItemRepo := new(mock_repositories.MockLoStudyPlanItemRepo)
	studentStudyPlanRepo := new(mock_repositories.MockStudentStudyPlanRepo)
	bookChapterRepo := new(mock_repositories.MockBookChapterRepo)
	assignmentStudyPlanTaskRepo := new(mock_repositories.MockAssignStudyPlanTaskRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	courseStudyPlanRepo := new(mock_repositories.MockCourseStudyPlanRepo)
	s := &ImportService{
		DB:                          db,
		StudyPlanRepo:               studyPlanRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		StudentStudyPlan:            studentStudyPlanRepo,
		AssignStudyPlanTaskRepo:     assignmentStudyPlanTaskRepo,
		StudentRepo:                 studentRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		BookChapterRepo:             bookChapterRepo,
		JSM:                         jsm,
	}

	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &npb.EventLearningObjectivesCreated{
				LearningObjectives: []*cpb.LearningObjective{
					&cpb.LearningObjective{
						Info: &cpb.ContentBasicInfo{
							Id:   "id",
							Name: "name",
						},
						TopicId: "topic-id",
					},
				},
				LoContentStructures: map[string]*npb.ContentStructures{
					"id": &npb.ContentStructures{
						ContentStructures: []*pb.ContentStructure{
							&pb.ContentStructure{CourseId: "course-id", BookId: "book-id", TopicId: "topic-id", ItemId: &pb.ContentStructure_LoId{LoId: wrapperspb.String("id")}},
						},
					},
				},
			},

			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				bookChapterRepo.On("RetrieveContentStructuresByLOs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.ContentStructure{
					"lo-id": entities.ContentStructure{CourseID: "course-id"},
				}, nil)
				studyPlanRepo.On("RetrieveStudyPlanItemInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.StudyPlanItemInfo{{
					StudyPlanID: database.Text("some-id"),
					BookID:      database.Text("book-id"),
					CourseID:    database.Text("course-id"),
				}}, nil)
				studyPlanItemRepo.On("BulkSync", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				loStudyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkUpsertByStudyPlanItem", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "err RetrieveStudyPlanItemInfo",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: fmt.Errorf("StudyPlanRepo.RetrieveStudyPlanItemInfo: %v", ErrSomethingWentWrong),
			req: &npb.EventLearningObjectivesCreated{
				LoContentStructures: map[string]*npb.ContentStructures{},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				bookChapterRepo.On("RetrieveContentStructuresByLOs", mock.Anything, mock.Anything, mock.Anything).Once().Return(map[string][]entities.ContentStructure{}, nil)
				studyPlanRepo.On("RetrieveStudyPlanItemInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "err studyPlanItemRepo.BulkSync",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: fmt.Errorf("database.ExecInTx: s.StudyPlanItemRepo.BulkSync: %w", ErrSomethingWentWrong),
			req: &npb.EventLearningObjectivesCreated{
				LoContentStructures: map[string]*npb.ContentStructures{},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				bookChapterRepo.On("RetrieveContentStructuresByLOs", mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.ContentStructure{}, nil)
				studyPlanRepo.On("RetrieveStudyPlanItemInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.StudyPlanItemInfo{{
					StudyPlanID: database.Text("some-id"),
				}}, nil)
				studyPlanItemRepo.On("BulkSync", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.SyncStudyPlanItemsOnLOsCreated(ctx, testCase.req.(*npb.EventLearningObjectivesCreated))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}

func Test_ImportStudyPlanItems(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	jsm := new(mock_nats.JetStreamManagement)
	masterStudyPlanRepo := new(mock_repositories.MockMasterStudyPlanRepo)
	individualStudyPlanRepo := new(mock_repositories.MockIndividualStudyPlan)
	studyPlanItemRepo := new(mock_repositories.MockStudyPlanItemRepo)
	importStudyPlanTaskRepo := new(mock_repositories.MockImportStudyPlanTaskRepo)
	s := &ImportService{
		DB:                      db,
		StudyPlanItemRepo:       studyPlanItemRepo,
		ImportStudyPlanTaskRepo: importStudyPlanTaskRepo,
		MasterStudyPlanRepo:     masterStudyPlanRepo,
		IndividualStudyPlanRepo: individualStudyPlanRepo,
		JSM:                     jsm,
	}

	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &npb.EventImportStudyPlan{
				TaskId: "task-id",
				StudyPlanItems: []*sspb.StudyPlanItemImport{
					&sspb.StudyPlanItemImport{
						StudyPlanId:        "study-plan-id",
						LearningMaterialId: "lm-id",
					},
				},
			},

			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				importStudyPlanTaskRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				masterStudyPlanRepo.On("BulkUpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				individualStudyPlanRepo.On("BulkUpdateTime", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("FindByStudyPlanID", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudyPlanItem{}, nil)
				studyPlanItemRepo.On("BulkInsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("UpdateWithCopiedFromItem", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				importStudyPlanTaskRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.ImportStudyPlanItems(ctx, testCase.req.(*npb.EventImportStudyPlan))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}

func Test_SyncStudyPlanItemsOnAssignmentsCreatedV2(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	jsm := new(mock_nats.JetStreamManagement)
	studyPlanRepo := new(mock_repositories.MockStudyPlanRepo)
	studyPlanItemRepo := new(mock_repositories.MockStudyPlanItemRepo)
	assignmentStudyPlanItemRepo := new(mock_repositories.MockAssignmentStudyPlanItemRepo)
	loStudyPlanItemRepo := new(mock_repositories.MockLoStudyPlanItemRepo)
	studentStudyPlanRepo := new(mock_repositories.MockStudentStudyPlanRepo)
	assignmentStudyPlanTaskRepo := new(mock_repositories.MockAssignStudyPlanTaskRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	courseStudyPlanRepo := new(mock_repositories.MockCourseStudyPlanRepo)
	bookChapterRepo := new(mock_repositories.MockBookChapterRepo)
	s := &ImportService{
		DB:                          db,
		StudyPlanRepo:               studyPlanRepo,
		StudyPlanItemRepo:           studyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		StudentStudyPlan:            studentStudyPlanRepo,
		AssignStudyPlanTaskRepo:     assignmentStudyPlanTaskRepo,
		StudentRepo:                 studentRepo,
		CourseStudyPlanRepo:         courseStudyPlanRepo,
		BookChapterRepo:             bookChapterRepo,
		JSM:                         jsm,
	}

	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req: &npb.EventAssignmentsCreated{
				Assignments: []*pb.Assignment{
					&pb.Assignment{
						AssignmentId: "id",
						Content: &pb.AssignmentContent{
							TopicId: "topic-id",
							LoId:    []string{"id"},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				resp := map[string][]entities.ContentStructure{
					"topic-id": []entities.ContentStructure{
						entities.ContentStructure{
							CourseID:     "course-id",
							BookID:       "book-id",
							TopicID:      "topic-id",
							AssignmentID: "id",
						},
					},
				}
				bookChapterRepo.On("RetrieveContentStructuresByTopics", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
				studyPlanRepo.On("RetrieveStudyPlanItemInfo", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.StudyPlanItemInfo{{
					StudyPlanID: database.Text("some-id"),
					BookID:      database.Text("book-id"),
					CourseID:    database.Text("course-id"),
				}}, nil)
				studyPlanItemRepo.On("BulkSync", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				assignmentStudyPlanItemRepo.On("BulkUpsertByStudyPlanItem", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.SyncStudyPlanItemsOnAssignmentsCreated(ctx, testCase.req.(*npb.EventAssignmentsCreated))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}
