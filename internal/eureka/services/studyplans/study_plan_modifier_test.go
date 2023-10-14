package studyplans

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type TestCase struct {
	ctx          context.Context
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestDeleteStudyPlanBelongsToACourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)

	courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	s := &StudyPlanModifierService{
		db:                   db,
		courseStudyPlanRepo:  courseStudyPlanRepo,
		studyPlanRepo:        studyPlanRepo,
		studentStudyPlanRepo: studentStudyPlanRepo,
		studyPlanItemRepo:    studyPlanItemRepo,
	}

	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         context.Background(),
			expectedErr: nil,
			req: &epb.DeleteStudyPlanBelongsToACourseRequest{
				CourseId:    "true",
				StudyPlanId: "true",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything, mock.Anything).Return(nil)
				courseStudyPlanRepo.On("FindByCourseIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.CourseStudyPlan{
					{
						CourseID:    pgtype.Text{String: "true"},
						StudyPlanID: pgtype.Text{String: "true"},
					},
				}, nil)
				courseStudyPlanRepo.On("DeleteCourseStudyPlanBy", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studyPlanRepo.On("RecursiveSoftDeleteStudyPlanByStudyPlanIDInCourse", mock.Anything, mock.Anything, mock.Anything).Return([]string{"true"}, nil)
				studyPlanItemRepo.On("DeleteStudyPlanItemsByStudyPlans", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentStudyPlanRepo.On("DeleteStudentStudyPlans", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.DeleteStudyPlanBelongsToACourse(ctx, testCase.req.(*epb.DeleteStudyPlanBelongsToACourseRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestUpdateSchoolDateStudyPlanItem(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := new(mock_database.Ext)

	courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	s := &StudyPlanModifierService{
		db:                   db,
		courseStudyPlanRepo:  courseStudyPlanRepo,
		studyPlanRepo:        studyPlanRepo,
		studentStudyPlanRepo: studentStudyPlanRepo,
		studyPlanItemRepo:    studyPlanItemRepo,
	}

	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         context.Background(),
			expectedErr: nil,
			req: &epb.UpdateStudyPlanItemsSchoolDateRequest{
				StudyPlanItemIds: []string{"id-1"},
				StudentId:        "student-1",
				SchoolDate:       timestamppb.New(time.Now()),
			},
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("UpdateSchoolDate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "missing student id",
			ctx:         context.Background(),
			expectedErr: fmt.Errorf("student id required"),
			req: &epb.UpdateStudyPlanItemsSchoolDateRequest{
				StudyPlanItemIds: []string{"id-1"},
				SchoolDate:       timestamppb.New(time.Now()),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "missing study plan item ids",
			ctx:         context.Background(),
			expectedErr: fmt.Errorf("empty study plan item ids"),
			req: &epb.UpdateStudyPlanItemsSchoolDateRequest{
				SchoolDate: timestamppb.New(time.Now()),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "error",
			ctx:         context.Background(),
			expectedErr: fmt.Errorf("studyPlanItemRepo.UpdateSchoolDate: %w", fmt.Errorf("error")),
			req: &epb.UpdateStudyPlanItemsSchoolDateRequest{
				StudyPlanItemIds: []string{"id-1"},
				StudentId:        "student-1",
				SchoolDate:       timestamppb.New(time.Now()),
			},
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("UpdateSchoolDate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.UpdateStudyPlanItemsSchoolDate(ctx, testCase.req.(*epb.UpdateStudyPlanItemsSchoolDateRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestUpdateStatusStudyPlanItem(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := new(mock_database.Ext)

	courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	s := &StudyPlanModifierService{
		db:                   db,
		courseStudyPlanRepo:  courseStudyPlanRepo,
		studyPlanRepo:        studyPlanRepo,
		studentStudyPlanRepo: studentStudyPlanRepo,
		studyPlanItemRepo:    studyPlanItemRepo,
	}

	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         context.Background(),
			expectedErr: nil,
			req: &epb.UpdateStudyPlanItemsStatusRequest{
				StudyPlanItemIds:    []string{"id-1"},
				StudentId:           "student-1",
				StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED,
			},
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "invalid status",
			ctx:         context.Background(),
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("status is invalid").Error()),
			req: &epb.UpdateStudyPlanItemsStatusRequest{
				StudyPlanItemIds:    []string{"id-1"},
				StudentId:           "student-1",
				StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "missing student id",
			ctx:         context.Background(),
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("student id required").Error()),
			req: &epb.UpdateStudyPlanItemsStatusRequest{
				StudyPlanItemIds:    []string{"id-1"},
				StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "missing study plan item ids",
			ctx:         context.Background(),
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("empty study plan item ids").Error()),
			req: &epb.UpdateStudyPlanItemsStatusRequest{
				StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "error",
			ctx:         context.Background(),
			expectedErr: status.Error(codes.Internal, fmt.Errorf("studyPlanItemRepo.UpdateStatus: %w", fmt.Errorf("error")).Error()),
			req: &epb.UpdateStudyPlanItemsStatusRequest{
				StudyPlanItemIds:    []string{"id-1"},
				StudentId:           "student-1",
				StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED,
			},
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.UpdateStudyPlanItemsStatus(ctx, testCase.req.(*epb.UpdateStudyPlanItemsStatusRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

var ErrSomethingWentWrong = fmt.Errorf("something went wrong")

func TestUpsertStudyPlan(t *testing.T) {
	t.Parallel()
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
	bookRepo := &mock_repositories.MockBookRepo{}
	courseBookRepo := &mock_repositories.MockCourseBookRepo{}
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	learningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepo{}
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	assignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
	loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	StudyPlanModifierService := &StudyPlanModifierService{
		db:                          db,
		studyPlanRepo:               studyPlanRepo,
		courseStudyPlanRepo:         courseStudyPlanRepo,
		studentRepo:                 studentRepo,
		studentStudyPlanRepo:        studentStudyPlanRepo,
		assignmentRepo:              assignmentRepo,
		learningObjectiveRepo:       learningObjectiveRepo,
		studyPlanItemRepo:           studyPlanItemRepo,
		assignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		loStudyPlanItemRepo:         loStudyPlanItemRepo,
		bookRepo:                    bookRepo,
		courseBookRepo:              courseBookRepo,
		internalModifierService: &services.InternalModifierService{
			AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
			LoStudyPlanItemRepo:         loStudyPlanItemRepo,
			StudyPlanItemRepo:           studyPlanItemRepo,
			StudyPlanRepo:               studyPlanRepo,
			AssignmentRepo:              assignmentRepo,
			LearningObjectiveRepo:       learningObjectiveRepo,
			BookRepo:                    bookRepo,
		},
	}
	updateReq := &epb.UpsertStudyPlanRequest{
		StudyPlanId:         wrapperspb.String("id"),
		Name:                "New study plan",
		TrackSchoolProgress: true,
		Grades:              []int32{1},
	}
	createReq := &epb.UpsertStudyPlanRequest{
		Name:                "New study plan",
		SchoolId:            12,
		Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		TrackSchoolProgress: true,
		Grades:              []int32{1},
		BookId:              "book-id-1",
		CourseId:            "course-id-1",
	}
	studyPlan := &entities.StudyPlan{
		StudyPlanType: database.Text(epb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String()),
		CourseID:      database.Text(createReq.CourseId),
	}

	bookTreeResp := []*repositories.BookTreeInfo{
		{
			LoID:      database.Text("lo-id-1"),
			TopicID:   database.Text("topic-1"),
			ChapterID: database.Text("chapter-1"),
		},
		{
			LoID:      database.Text("lo-id-2"),
			TopicID:   database.Text("topic-2"),
			ChapterID: database.Text("chapter-2"),
		},
	}

	assignments := []*entities.Assignment{
		{
			ID: database.Text("assingment-id-1"),
			Content: database.JSONB([]byte(`{
				"topic_id": "topic-1"
			}`)),
		},
		{
			ID: database.Text("assingment-id-2"),
			Content: database.JSONB([]byte(`{
				"topic_id": "topic-2"
			}`)),
		},
	}
	learningObjectives := []*entities.LearningObjective{
		{
			ID:      database.Text("assingment-id-1"),
			TopicID: database.Text("topic-1"),
		},
		{
			ID:      database.Text("assingment-id-2"),
			TopicID: database.Text("topic-2"),
		},
	}

	studentIDs := database.TextArray([]string{"1", "2"})

	testCases := []TestCase{
		{
			name:        "update happy case",
			req:         updateReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(studyPlan, nil)
				studyPlanRepo.On("BulkUpdateByMaster", ctx, db, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "update err studyPlanRepo.FindByID NotFound",
			req:         updateReq,
			expectedErr: status.Errorf(codes.NotFound, "study plan id %v does not exists", updateReq.StudyPlanId.Value),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "update err studyPlanRepo.FindByID",
			req:         updateReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve study plan: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "update err studyPlanRepo.BulkUpdateByMaster",
			req:         updateReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("unable to update study plan by master: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(studyPlan, nil)
				studyPlanRepo.On("BulkUpdateByMaster", ctx, db, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "create happy case",
			req:         createReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeResp, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(learningObjectives, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return(assignments, nil)
				studyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "missing book id",
			req: &epb.UpsertStudyPlanRequest{
				Name:                "New study plan",
				SchoolId:            12,
				Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
				TrackSchoolProgress: true,
				Grades:              []int32{1},
				CourseId:            "course-id-1",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, ErrMustHaveBookID.Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing course id",
			req: &epb.UpsertStudyPlanRequest{
				Name:                "New study plan",
				SchoolId:            12,
				Status:              epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
				TrackSchoolProgress: true,
				Grades:              []int32{1},
				BookId:              "book-id-1",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, ErrMustHaveCourseID.Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "create err courseBookRepo.FindByCourseIDAndBookID",
			req:         createReq,
			expectedErr: status.Errorf(codes.InvalidArgument, status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve course book by course id and book id: %w", ErrSomethingWentWrong).Error()).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},

		{
			name:        "create err studentRepo.FindStudentsByCourseID NotFound",
			req:         createReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeResp, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return(assignments, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(learningObjectives, nil)
				studyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "create err studentRepo.FindStudentsByCourseID",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("studentRepo.FindStudentsByCourseID: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},

		{
			name:        "create err studyPlanRepo.BulkUpsert",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("studyPlanRepo.BulkUpsert: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "create err courseStudyPlanRepo.BulkUpsert",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("courseStudyPlanRepo.BulkUpsert: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "create err studentStudyPlanRepo.BulkUpsert",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("studentStudyPlan.BulkUpsert: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "create err bookRepo.RetrieveBookTreeByBookID",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, ErrSomethingWentWrong.Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "create err bookRepo.RetrieveBookTreeByBookID no rows",
			req:         createReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.BookTreeInfo{}, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return(assignments, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(learningObjectives, nil)
				studyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "create err learningObjectiveRepo.RetrieveLearningObjectivesByTopicIDs",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("learningObjectiveRepo.RetrieveLearningObjectivesByTopicIDs: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeResp, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "create err assignmentRepo.RetrieveAssignmentsByTopicIDs",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("assignmentRepo.RetrieveAssignmentsByTopicIDs: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeResp, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(learningObjectives, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "create err assignmentRepo.RetrieveAssignmentsByTopicIDs no rows",
			req:         createReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeResp, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(learningObjectives, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
				studyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "create err studyPlanItemRepo.BulkInsert",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("studyPlanItemRepo.BulkInsert: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeResp, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(learningObjectives, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return(assignments, nil)
				studyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "create err assignmentStudyPlanItemRepo.BulkInsert",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("loStudyPlanItemRepo.BulkInsert: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeResp, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(learningObjectives, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return(assignments, nil)
				studyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
				assignmentStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "create err loStudyPlanItemRepo.BulkInsert",
			req:         createReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("loStudyPlanItemRepo.BulkInsert: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				courseBookRepo.On("FindByCourseIDAndBookID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				courseStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studentRepo.On("FindStudentsByCourseID", ctx, tx, mock.Anything).Once().Return(&studentIDs, nil)
				studentStudyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				studyPlanRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
				bookRepo.On("RetrieveBookTreeByBookID", mock.Anything, mock.Anything, mock.Anything).Once().Return(bookTreeResp, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, tx, mock.Anything).Once().Return(learningObjectives, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, tx, mock.Anything).Once().Return(assignments, nil)
				studyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				assignmentStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(nil)
				loStudyPlanItemRepo.On("BulkInsert", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := interceptors.NewIncomingContext(context.Background())
			testCase.setup(ctx)
			rsp, err := StudyPlanModifierService.UpsertStudyPlan(ctx, testCase.req.(*epb.UpsertStudyPlanRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, rsp)
			} else {
				assert.Equal(t, testCase.expectedResp.(*epb.UpsertStudyPlanResponse), rsp)
			}
		})
	}
}
