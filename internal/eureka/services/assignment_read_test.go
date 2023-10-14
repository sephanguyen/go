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
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestListStudentAvailableContents(t *testing.T) {
	t.Parallel()
	t.Run("empty available contents", func(t *testing.T) {
		t.Parallel()
		loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
		assignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
		topicRepo := &mock_repositories.MockTopicRepo{}
		topicsLearningObjectivesRepo := &mock_repositories.MockTopicsLearningObjectivesRepo{}

		topicRepo.On("FindByBookIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.Topic{
			{
				ID: database.Text("topic-1"),
			},
		}, nil)
		topicsLearningObjectivesRepo.On("RetrieveByLoIDs", mock.Anything, mock.Anything, mock.Anything).Return([]*repositories.TopicLearningObjective{
			{
				Topic: &entities.Topic{
					ID: database.Text("topic-1"),
				},
				LearningObjective: &entities.LearningObjective{
					ID: database.Text("lo-1"),
				},
			},
		}, nil)

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListStudentAvailableContents", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo:         studentStudyPlanRepo,
			LoStudyPlanItemRepo:          loStudyPlanItemRepo,
			AssignmentStudyPlanItemRepo:  assignmentStudyPlanItemRepo,
			TopicRepo:                    topicRepo,
			TopicsLearningObjectivesRepo: topicsLearningObjectivesRepo,
		}

		resp, err := svc.ListStudentAvailableContents(context.Background(), &pb.ListStudentAvailableContentsRequest{})
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudentAvailableContentsResponse{}, resp)
	})

	t.Run("student has available contents", func(t *testing.T) {
		t.Parallel()
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}

		items := []*entities.StudyPlanItem{
			{
				ID:               database.Text("id1"),
				StudyPlanID:      database.Text("sid"),
				ContentStructure: database.JSONB([]byte(`{"book_id": "book-1", "topic_id": "topic-1", "assignment_id": "assignmentid"}`)),
				DisplayOrder:     database.Int4(1),
			},
			{
				ID:               database.Text("id2"),
				StudyPlanID:      database.Text("sid"),
				ContentStructure: database.JSONB([]byte(`{"book_id": "book-1", "topic_id": "topic-1", "lo_id": "loid"}`)),
				DisplayOrder:     database.Int4(2),
			},
		}
		studentStudyPlanRepo.On("ListStudentAvailableContents", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(items, nil)

		pgIDs := database.TextArray([]string{"id1", "id2"})

		loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
		loStudyPlanItems := []*entities.LoStudyPlanItem{
			{
				StudyPlanItemID: database.Text("id2"),
				LoID:            database.Text("loid"),
			},
		}
		loStudyPlanItemRepo.On("FindByStudyPlanItemIDs", mock.Anything, mock.Anything, pgIDs).Once().Return(loStudyPlanItems, nil)

		assignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
		assignmentPlanItems := []*entities.AssignmentStudyPlanItem{
			{
				StudyPlanItemID: database.Text("id1"),
				AssignmentID:    database.Text("assignmentid"),
			},
		}
		assignmentStudyPlanItemRepo.On("FindByStudyPlanItemIDs", mock.Anything, mock.Anything, pgIDs).Once().Return(assignmentPlanItems, nil)
		topicsAssignmentsRepo := &mock_repositories.MockTopicsAssignmentsRepo{}
		topicsAssignmentsRepo.On("RetrieveByAssignmentIDs", mock.Anything, mock.Anything, []string{"assignmentid"}).Once().
			Return([]*entities.TopicsAssignments{
				{
					TopicID:      database.Text("topic-1"),
					AssignmentID: database.Text("assignmentid"),
				},
			}, nil)

		topicRepo := &mock_repositories.MockTopicRepo{}
		topicRepo.On("FindByBookIDs", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.Topic{
			{
				ID: database.Text("topic-1"),
			},
		}, nil)
		topicsLearningObjectivesRepo := &mock_repositories.MockTopicsLearningObjectivesRepo{}
		topicsLearningObjectivesRepo.On("RetrieveByLoIDs", mock.Anything, mock.Anything, mock.Anything).Return([]*repositories.TopicLearningObjective{
			{
				Topic: &entities.Topic{
					ID: database.Text("topic-1"),
				},
				LearningObjective: &entities.LearningObjective{
					ID: database.Text("loid"),
				},
			},
		}, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo:         studentStudyPlanRepo,
			LoStudyPlanItemRepo:          loStudyPlanItemRepo,
			AssignmentStudyPlanItemRepo:  assignmentStudyPlanItemRepo,
			TopicRepo:                    topicRepo,
			TopicsAssignmentsRepo:        topicsAssignmentsRepo,
			TopicsLearningObjectivesRepo: topicsLearningObjectivesRepo,
		}

		ctx := metadata.NewIncomingContext(context.Background(),
			metadata.MD{
				"pkg":     []string{"pkg"},
				"version": []string{"version"},
				"token":   []string{"token"},
			})

		resp, err := svc.ListStudentAvailableContents(ctx, &pb.ListStudentAvailableContentsRequest{})
		assert.Nil(t, err)
		assert.Equal(t, len(items), len(resp.Contents))
		for i, item := range items {
			assert.Equal(t, item.ID.String, resp.Contents[i].StudyPlanItem.StudyPlanItemId)
			assert.Equal(t, item.StudyPlanID.String, resp.Contents[i].StudyPlanItem.StudyPlanId)

			resourceID := resp.Contents[i].ResourceId

			var inLOItems bool
			for _, loItem := range loStudyPlanItems {
				if loItem.StudyPlanItemID.String == item.ID.String && loItem.LoID.String == resourceID {
					inLOItems = true
					break
				}
			}

			var inAssignmentItems bool
			for _, it := range assignmentPlanItems {
				if it.StudyPlanItemID.String == item.ID.String && it.AssignmentID.String == resourceID {
					inAssignmentItems = true
					break
				}
			}

			if !inLOItems && !inAssignmentItems {
				t.Errorf("unexpected resource id: %q", resourceID)
			}
		}
	})
}

func TestListStudyPlans(t *testing.T) {
	t.Parallel()
	t.Run("invalid limit", func(t *testing.T) {
		t.Parallel()
		req := &pb.ListStudyPlansRequest{
			StudentId: "sid",
			CourseId:  "cid",
			SchoolId:  1,
			Paging: &cpb.Paging{
				Limit: 1010,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "offset",
				},
			},
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListStudyPlans", mock.Anything, mock.Anything, &repositories.ListStudyPlansArgs{
			StudentID: database.Text(req.StudentId),
			CourseID:  database.Text(req.CourseId),
			SchoolID:  database.Int4(req.SchoolId),
			Limit:     uint32(10),
			Offset:    database.Text(req.Paging.GetOffsetString()),
		}).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudyPlans(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudyPlansResponse{}, resp)
	})

	t.Run("empty offset", func(t *testing.T) {
		t.Parallel()
		req := &pb.ListStudyPlansRequest{
			StudentId: "sid",
			CourseId:  "cid",
			SchoolId:  1,
			Paging: &cpb.Paging{
				Limit: 7,
			},
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListStudyPlans", mock.Anything, mock.Anything, &repositories.ListStudyPlansArgs{
			StudentID: database.Text(req.StudentId),
			CourseID:  database.Text(req.CourseId),
			SchoolID:  database.Int4(req.SchoolId),
			Limit:     req.Paging.Limit,
			Offset:    pgtype.Text{Status: pgtype.Null},
		}).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudyPlans(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudyPlansResponse{}, resp)
	})

	t.Run("empty assigned study plans", func(t *testing.T) {
		t.Parallel()
		req := &pb.ListStudyPlansRequest{
			StudentId: "sid",
			CourseId:  "cid",
			SchoolId:  1,
			Paging: &cpb.Paging{
				Limit: 1,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "offset",
				},
			},
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListStudyPlans", mock.Anything, mock.Anything, &repositories.ListStudyPlansArgs{
			StudentID: database.Text(req.StudentId),
			CourseID:  database.Text(req.CourseId),
			SchoolID:  database.Int4(req.SchoolId),
			Limit:     req.Paging.Limit,
			Offset:    database.Text(req.Paging.GetOffsetString()),
		}).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudyPlans(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudyPlansResponse{}, resp)
		assert.Equal(t, &pb.ListStudyPlansResponse{}, resp)
	})

	t.Run("student has assigned study plans", func(t *testing.T) {
		t.Parallel()
		req := &pb.ListStudyPlansRequest{
			StudentId: "sid",
			CourseId:  "cid",
			SchoolId:  1,
			Paging: &cpb.Paging{
				Limit: 1,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "offset",
				},
			},
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}

		items := []*entities.StudyPlan{
			{
				ID:   database.Text("id1"),
				Name: database.Text("name1"),
			},
			{
				ID:   database.Text("id2"),
				Name: database.Text("name2"),
			},
		}
		studentStudyPlanRepo.On("ListStudyPlans", mock.Anything, mock.Anything, &repositories.ListStudyPlansArgs{
			StudentID: database.Text(req.StudentId),
			CourseID:  database.Text(req.CourseId),
			SchoolID:  database.Int4(req.SchoolId),
			Limit:     req.Paging.Limit,
			Offset:    database.Text(req.Paging.GetOffsetString()),
		}).Once().Return(items, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudyPlans(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, len(items), len(resp.Items))

		for i, item := range items {
			assert.Equal(t, item.ID.String, resp.Items[i].StudyPlanId)
			assert.Equal(t, item.Name.String, resp.Items[i].Name)
		}
	})
}

func TestListActiveStudyPlanItems(t *testing.T) {
	tNow := timeutil.Now
	defer func() {
		timeutil.Now = tNow
	}()

	now := time.Now().UTC()
	timeutil.Now = func() time.Time {
		return now
	}
	pgNow := database.Timestamptz(now)

	t.Run("limit out of range", func(t *testing.T) {
		req := &pb.ListStudentToDoItemsRequest{
			Status:    pb.ToDoStatus_TO_DO_STATUS_ACTIVE,
			StudentId: "sid",
			Paging: &cpb.Paging{
				Limit: 1010,
			},
		}

		args := &repositories.ListStudyPlanItemsArgs{
			StudentID:        database.Text(req.StudentId),
			Offset:           pgtype.Timestamptz{Status: pgtype.Null},
			Now:              pgNow,
			Limit:            10,
			CourseIDs:        pgtype.TextArray{Status: pgtype.Null},
			StudyPlanID:      pgtype.Text{Status: pgtype.Null},
			StudyPlanItemID:  pgtype.Text{Status: pgtype.Null},
			IncludeCompleted: false,
			DisplayOrder:     pgtype.Int4{Status: pgtype.Null},
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListActiveStudyPlanItems", mock.Anything, mock.Anything, args).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudentToDoItems(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudentToDoItemsResponse{}, resp)
	})

	t.Run("empty offset", func(t *testing.T) {
		req := &pb.ListStudentToDoItemsRequest{
			Status:    pb.ToDoStatus_TO_DO_STATUS_ACTIVE,
			StudentId: "sid",
			Paging: &cpb.Paging{
				Limit: 10,
			},
		}

		args := &repositories.ListStudyPlanItemsArgs{
			StudentID:        database.Text(req.StudentId),
			Offset:           pgtype.Timestamptz{Status: pgtype.Null},
			Now:              pgNow,
			Limit:            10,
			CourseIDs:        pgtype.TextArray{Status: pgtype.Null},
			StudyPlanID:      pgtype.Text{Status: pgtype.Null},
			StudyPlanItemID:  pgtype.Text{Status: pgtype.Null},
			IncludeCompleted: false,
			DisplayOrder:     pgtype.Int4{Status: pgtype.Null},
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListActiveStudyPlanItems", mock.Anything, mock.Anything, args).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudentToDoItems(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudentToDoItemsResponse{}, resp)
	})

	t.Run("empty assigned study plans", func(t *testing.T) {
		req := &pb.ListStudentToDoItemsRequest{
			Status:    pb.ToDoStatus_TO_DO_STATUS_ACTIVE,
			StudentId: "sid",
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetCombined{
					OffsetCombined: &cpb.Paging_Combined{
						OffsetString: "s",
						OffsetTime:   timestamppb.New(now.Add(time.Minute)),
					},
				},
			},
		}

		args := &repositories.ListStudyPlanItemsArgs{
			StudentID:        database.Text(req.StudentId),
			Offset:           pgtype.Timestamptz{Time: now.Add(time.Minute), Status: pgtype.Present},
			Now:              pgNow,
			Limit:            10,
			CourseIDs:        pgtype.TextArray{Status: pgtype.Null},
			StudyPlanID:      pgtype.Text{Status: pgtype.Null},
			StudyPlanItemID:  pgtype.Text{String: "s", Status: pgtype.Present},
			IncludeCompleted: false,
			DisplayOrder:     database.Int4(0),
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListActiveStudyPlanItems", mock.Anything, mock.Anything, args).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudentToDoItems(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudentToDoItemsResponse{}, resp)
	})
}

func TestListUpcomingStudyPlanItems(t *testing.T) {
	tNow := timeutil.Now
	defer func() {
		timeutil.Now = tNow
	}()

	now := time.Now().UTC()
	timeutil.Now = func() time.Time {
		return now
	}
	pgNow := database.Timestamptz(now)

	t.Run("limit out of range", func(t *testing.T) {
		req := &pb.ListStudentToDoItemsRequest{
			Status:    pb.ToDoStatus_TO_DO_STATUS_UPCOMING,
			StudentId: "sid",
			Paging: &cpb.Paging{
				Limit: 1010,
			},
		}

		args := &repositories.ListStudyPlanItemsArgs{
			StudentID:        database.Text(req.StudentId),
			Offset:           pgtype.Timestamptz{Time: now, Status: pgtype.Present},
			Now:              pgNow,
			Limit:            10,
			CourseIDs:        pgtype.TextArray{Status: pgtype.Null},
			StudyPlanID:      pgtype.Text{Status: pgtype.Null},
			StudyPlanItemID:  pgtype.Text{Status: pgtype.Null},
			IncludeCompleted: false,
			DisplayOrder:     pgtype.Int4{Status: pgtype.Null},
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListUpcomingStudyPlanItems", mock.Anything, mock.Anything, args).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudentToDoItems(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudentToDoItemsResponse{}, resp)
	})

	t.Run("empty offset", func(t *testing.T) {
		req := &pb.ListStudentToDoItemsRequest{
			Status:    pb.ToDoStatus_TO_DO_STATUS_UPCOMING,
			StudentId: "sid",
			Paging: &cpb.Paging{
				Limit: 10,
			},
		}

		args := &repositories.ListStudyPlanItemsArgs{
			StudentID:        database.Text(req.StudentId),
			Offset:           pgtype.Timestamptz{Time: now, Status: pgtype.Present},
			Now:              pgNow,
			Limit:            10,
			CourseIDs:        pgtype.TextArray{Status: pgtype.Null},
			StudyPlanID:      pgtype.Text{Status: pgtype.Null},
			StudyPlanItemID:  pgtype.Text{Status: pgtype.Null},
			IncludeCompleted: false,
			DisplayOrder:     pgtype.Int4{Status: pgtype.Null},
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListUpcomingStudyPlanItems", mock.Anything, mock.Anything, args).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudentToDoItems(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudentToDoItemsResponse{}, resp)
	})

	t.Run("empty assigned study plans", func(t *testing.T) {
		req := &pb.ListStudentToDoItemsRequest{
			Status:    pb.ToDoStatus_TO_DO_STATUS_UPCOMING,
			StudentId: "sid",
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetCombined{
					OffsetCombined: &cpb.Paging_Combined{
						OffsetString: "s",
						OffsetTime:   timestamppb.New(now.Add(time.Minute)),
					},
				},
			},
		}

		args := &repositories.ListStudyPlanItemsArgs{
			StudentID:        database.Text(req.StudentId),
			Offset:           pgtype.Timestamptz{Time: now.Add(time.Minute), Status: pgtype.Present},
			Now:              pgNow,
			Limit:            10,
			CourseIDs:        pgtype.TextArray{Status: pgtype.Null},
			StudyPlanID:      pgtype.Text{Status: pgtype.Null},
			StudyPlanItemID:  pgtype.Text{String: "s", Status: pgtype.Present},
			IncludeCompleted: false,
			DisplayOrder:     database.Int4(0),
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On("ListUpcomingStudyPlanItems", mock.Anything, mock.Anything, args).Once().Return(nil, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.ListStudentToDoItems(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudentToDoItemsResponse{}, resp)
	})
}

func TestRetrieveAssignments(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	mockDB := &mock_database.Ext{}

	s := &AssignmentReaderService{
		DB:             mockDB,
		AssignmentRepo: assignmentRepo,
	}

	assignmentIDs := []string{"assignment-1", "assignment-2"}
	validReq := &pb.RetrieveAssignmentsRequest{
		Ids: assignmentIDs,
	}
	e := &entities.Assignment{}
	database.AllNullEntity(e)
	_ = e.ID.Set("assignment-id-1")
	_ = e.Content.Set(entities.AssignmentContent{
		TopicID: "topic-id",
		LoIDs: []string{
			"lo-id", "lo-id2",
		},
	})
	_ = e.Settings.Set(entities.AssignmentSetting{
		AllowLateSubmission:       true,
		AllowResubmission:         false,
		RequireAssignmentNote:     true,
		RequireAttachment:         false,
		RequireCompleteDate:       false,
		RequireDuration:           true,
		RequireCorrectness:        false,
		RequireUnderstandingLevel: true,
	})
	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			req:         validReq,
			setup: func(ctx context.Context) {
				assignmentRepo.On("RetrieveAssignments", mock.Anything, s.DB, database.TextArray(assignmentIDs)).Once().Return(
					[]*entities.Assignment{
						e,
					}, nil,
				)
			},
		},
		{
			name:        "AssignmentRepo.RetrieveAssignments error",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, "s.AssignmentRepo.RetrieveAssignments: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				assignmentRepo.On("RetrieveAssignments", mock.Anything, s.DB, database.TextArray(assignmentIDs)).Once().Return(
					nil, pgx.ErrNoRows,
				)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.RetrieveAssignments(ctx, testCase.req.(*pb.RetrieveAssignmentsRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestRetrieveStudyPlanProgress(t *testing.T) {
	t.Parallel()
	t.Run("count completed study plan items error", func(t *testing.T) {
		t.Parallel()
		req := &pb.RetrieveStudyPlanProgressRequest{
			StudentId:   "sid",
			StudyPlanId: "spi",
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On(
			"CountStudentStudyPlanItems",
			mock.Anything,
			mock.Anything,
			database.Text(req.StudentId),
			database.Text(req.StudyPlanId),
			mock.Anything,
			database.Bool(true),
		).Once().Return(0, pgx.ErrTxClosed)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.RetrieveStudyPlanProgress(context.Background(), req)
		assert.Equal(t, pgx.ErrTxClosed, err)
		assert.Nil(t, resp)
	})

	t.Run("count total study plan items error", func(t *testing.T) {
		t.Parallel()
		req := &pb.RetrieveStudyPlanProgressRequest{
			StudentId:   "sid",
			StudyPlanId: "spi",
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On(
			"CountStudentStudyPlanItems",
			mock.Anything,
			mock.Anything,
			database.Text(req.StudentId),
			database.Text(req.StudyPlanId),
			mock.Anything,
			database.Bool(true),
		).Once().Return(1, nil)
		studentStudyPlanRepo.On(
			"CountStudentStudyPlanItems",
			mock.Anything,
			mock.Anything,
			database.Text(req.StudentId),
			database.Text(req.StudyPlanId),
			mock.Anything,
			database.Bool(false),
		).Once().Return(0, pgx.ErrTxClosed)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.RetrieveStudyPlanProgress(context.Background(), req)
		assert.Equal(t, pgx.ErrTxClosed, err)
		assert.Nil(t, resp)
	})

	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		req := &pb.RetrieveStudyPlanProgressRequest{
			StudentId:   "sid",
			StudyPlanId: "spi",
		}

		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studentStudyPlanRepo.On(
			"CountStudentStudyPlanItems",
			mock.Anything,
			mock.Anything,
			database.Text(req.StudentId),
			database.Text(req.StudyPlanId),
			mock.Anything,
			database.Bool(true),
		).Once().Return(1, nil)
		studentStudyPlanRepo.On(
			"CountStudentStudyPlanItems",
			mock.Anything,
			mock.Anything,
			database.Text(req.StudentId),
			database.Text(req.StudyPlanId),
			mock.Anything,
			database.Bool(false),
		).Once().Return(10, nil)

		svc := &AssignmentReaderService{
			StudentStudyPlanRepo: studentStudyPlanRepo,
		}

		resp, err := svc.RetrieveStudyPlanProgress(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, &pb.RetrieveStudyPlanProgressResponse{
			CompletedAssignments: 1,
			TotalAssignments:     10,
		}, resp)
	})
}

func TestListCourseTodo(t *testing.T) {
	t.Parallel()
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	assignmentStudyPlanItemRepo := &mock_repositories.MockAssignmentStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	svc := &AssignmentReaderService{
		StudyPlanItemRepo:           studyPlanItemRepo,
		LoStudyPlanItemRepo:         loStudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: assignmentStudyPlanItemRepo,
		DB:                          mockDB,
	}
	ctx := context.Background()
	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			expectedResp: &pb.ListCourseTodoResponse{
				StatisticItems: []*pb.StatisticTodoItem{
					{
						Item: &pb.ToDoItem{
							StudyPlanItem: &pb.StudyPlanItem{
								StudyPlanItemId: "study_plan_item_id_1",
							},
						},
						TotalAssignedStudent: 5,
						CompletedStudent:     4,
					},
					{
						Item: &pb.ToDoItem{
							StudyPlanItem: &pb.StudyPlanItem{
								StudyPlanItemId: "study_plan_item_id_2",
							},
						},
						TotalAssignedStudent: 7,
						CompletedStudent:     0,
					},
				},
			},
			req: &pb.ListCourseTodoRequest{
				StudyPlanId: "master_study_plan_id",
			},
			setup: func(ctx context.Context) {
				studyPlanItems := make([]*entities.StudyPlanItem, 0)
				studyPlanItems = append(studyPlanItems, &entities.StudyPlanItem{
					ID: database.Text("study_plan_item_id_1"),
				})
				studyPlanItems = append(studyPlanItems, &entities.StudyPlanItem{
					ID: database.Text("study_plan_item_id_2"),
				})

				studyPlanItemTotalStudentMap := make(map[string]int)
				studyPlanItemTotalStudentMap["study_plan_item_id_1"] = 5
				studyPlanItemTotalStudentMap["study_plan_item_id_2"] = 7
				sortedTotalStudyPlanItem := []string{"study_plan_item_id_1", "study_plan_item_id_2"}
				studyPlanItemRepo.On("CountStudentInStudyPlanItem", ctx, mockDB, mock.Anything, database.Bool(false)).Once().Return(studyPlanItemTotalStudentMap, sortedTotalStudyPlanItem, nil)

				studyPlanItemCompletedStudentMap := make(map[string]int)
				studyPlanItemCompletedStudentMap["study_plan_item_id_1"] = 4
				sortedCompletedStudyPlanItem := []string{"study_plan_item_id_1"}
				studyPlanItemRepo.On("CountStudentInStudyPlanItem", ctx, mockDB, mock.Anything, database.Bool(true)).Once().Return(studyPlanItemCompletedStudentMap, sortedCompletedStudyPlanItem, nil)

				studyPlanItemRepo.On("FindAndSortByIDs", ctx, mockDB, mock.Anything).Once().Return(studyPlanItems, nil)

				loItems := make([]*entities.LoStudyPlanItem, 0)
				loItems = append(loItems, &entities.LoStudyPlanItem{
					StudyPlanItemID: database.Text("study_plan_item_id_1"),
					LoID:            database.Text("lo_id_1"),
				})
				loStudyPlanItemRepo.On("FindByStudyPlanItemIDs", ctx, mockDB, mock.Anything).Once().Return(loItems, nil)

				assignmentItems := make([]*entities.AssignmentStudyPlanItem, 0)
				assignmentItems = append(assignmentItems, &entities.AssignmentStudyPlanItem{
					StudyPlanItemID: database.Text("study_plan_item_id_2"),
					AssignmentID:    database.Text("assignment_id_1"),
				})
				assignmentStudyPlanItemRepo.On("FindByStudyPlanItemIDs", ctx, mockDB, mock.Anything).Once().Return(assignmentItems, nil)
			},
		},
		{
			name:        "error no rows CountStudentInStudyPlanItem with total student",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.NotFound, fmt.Sprintf("ListCourseTodo.StudyPlanItemRepo.CountStudentInStudyPlanItem with total student err: %v", pgx.ErrNoRows)),
			req: &pb.ListCourseTodoRequest{
				StudyPlanId: "master_study_plan_id",
			},
			setup: func(ctx context.Context) {
				studyPlanItemTotalStudentMap := make(map[string]int)
				sortedStudyPlanItem := make([]string, 0)
				studyPlanItemRepo.On("CountStudentInStudyPlanItem", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(studyPlanItemTotalStudentMap, sortedStudyPlanItem, pgx.ErrNoRows)
			},
		},
		{
			name:         "error not found with empty master study plan",
			ctx:          interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr:  nil,
			expectedResp: &pb.ListCourseTodoResponse{},
			req: &pb.ListCourseTodoRequest{
				StudyPlanId: "",
			},
			setup: func(ctx context.Context) {
				studyPlanItemTotalStudentMap := make(map[string]int)
				sortedStudyPlanItem := make([]string, 0)
				studyPlanItemRepo.On("CountStudentInStudyPlanItem", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(studyPlanItemTotalStudentMap, sortedStudyPlanItem, nil)
			},
		},
	}

	for _, testcase := range testcases {
		testcase.setup(testcase.ctx)
		result, err := svc.ListCourseTodo(testcase.ctx, testcase.req.(*pb.ListCourseTodoRequest))
		assert.Equal(t, testcase.expectedErr, err)
		if testcase.expectedErr == nil {
			expectedResp := testcase.expectedResp.(*pb.ListCourseTodoResponse)
			assert.Equal(t, len(expectedResp.StatisticItems), len(result.StatisticItems))
			for i := range expectedResp.StatisticItems {
				assert.Equal(t, expectedResp.StatisticItems[i].TotalAssignedStudent, result.StatisticItems[i].TotalAssignedStudent)
				assert.Equal(t, expectedResp.StatisticItems[i].CompletedStudent, result.StatisticItems[i].CompletedStudent)
				assert.Equal(t, expectedResp.StatisticItems[i].Item.StudyPlanItem.StudyPlanItemId, result.StatisticItems[i].Item.StudyPlanItem.StudyPlanItemId)
			}
		}
	}
}

func TestGetChildStudyPlanItems(t *testing.T) {
	t.Parallel()
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	svc := &AssignmentReaderService{
		StudyPlanItemRepo: studyPlanItemRepo,
		DB:                mockDB,
	}
	studyPlanItemID := database.Text("copy-study-plan-item-id")
	ctx := context.Background()
	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			expectedResp: &pb.GetChildStudyPlanItemsResponse{
				Items: []*pb.GetChildStudyPlanItemsResponse_UserStudyPlanItem{
					{
						UserId: "user-id-1",
						StudyPlanItem: &pb.StudyPlanItem{
							StudyPlanItemId: "study-plan-item-id-1",
						},
					},
					{
						UserId: "user-id-2",
						StudyPlanItem: &pb.StudyPlanItem{
							StudyPlanItemId: "study-plan-item-id-2",
						},
					},
				},
			},
			req: &pb.GetChildStudyPlanItemsRequest{
				StudyPlanItemId: "copy-study-plan-item-id",
				UserIds:         []string{"user-id-1", "user-id-2"},
			},
			setup: func(ctx context.Context) {
				studyPlanItems := make([]*entities.StudyPlanItem, 0)
				studyPlanItems = append(studyPlanItems, &entities.StudyPlanItem{
					ID: database.Text("study-plan-item-id-1"),
				})
				studyPlanItems = append(studyPlanItems, &entities.StudyPlanItem{
					ID: database.Text("study-plan-item-id-2"),
				})

				userStudyPlanItem := make(map[string]*entities.StudyPlanItem)
				userStudyPlanItem["user-id-1"] = &entities.StudyPlanItem{
					ID: database.Text("study-plan-item-id-1"),
				}
				userStudyPlanItem["user-id-2"] = &entities.StudyPlanItem{
					ID: database.Text("study-plan-item-id-2"),
				}
				studyPlanItemRepo.On("RetrieveChildStudyPlanItem", ctx, mockDB, studyPlanItemID, mock.Anything).Once().Return(userStudyPlanItem, nil)
			},
		},
		{
			name:         "err no row",
			ctx:          interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr:  fmt.Errorf("StudyPlanItemRepo.RetrieveChildStudyPlanItem: %w", pgx.ErrNoRows),
			expectedResp: nil,
			req: &pb.GetChildStudyPlanItemsRequest{
				StudyPlanItemId: "copy-study-plan-item-id",
				UserIds:         []string{"user-id-1", "user-id-2"},
			},
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("RetrieveChildStudyPlanItem", ctx, mockDB, studyPlanItemID, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testcase := range testcases {
		testcase.setup(testcase.ctx)
		result, err := svc.GetChildStudyPlanItems(testcase.ctx, testcase.req.(*pb.GetChildStudyPlanItemsRequest))
		assert.Equal(t, testcase.expectedErr, err)
		count := 0
		if testcase.expectedErr == nil {
			expectedResp := testcase.expectedResp.(*pb.GetChildStudyPlanItemsResponse)
			assert.Equal(t, len(expectedResp.Items), len(result.Items))
			for i := 0; i < len(expectedResp.Items); i++ {
				for j := 0; j < len(result.Items); j++ {
					if expectedResp.Items[i].UserId == result.Items[j].UserId && expectedResp.Items[i].StudyPlanItem.StudyPlanItemId == result.Items[j].StudyPlanItem.StudyPlanItemId {
						count++
					}
				}
			}
			assert.Equal(t, len(expectedResp.Items), len(result.Items), fmt.Sprintf("unexpected number of items, expected: %d, actual: %d", len(expectedResp.Items), len(result.Items)))
		}
	}
	return
}

func TestRetrieveStatisticAssignmentClass(t *testing.T) {
	t.Parallel()
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	mockDB := &mock_database.Ext{}
	svc := &AssignmentReaderService{
		StudyPlanItemRepo: studyPlanItemRepo,
		DB:                mockDB,
	}

	ctx := context.Background()
	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: nil,
			expectedResp: &pb.RetrieveStatisticAssignmentClassResponse{
				StatisticItem: &pb.StatisticTodoItem{
					Item: &pb.ToDoItem{
						StudyPlanItem: &pb.StudyPlanItem{
							StudyPlanId:     "study-plan-id",
							StudyPlanItemId: "study-plan-item-id",
						},
					},
					CompletedStudent:     6,
					TotalAssignedStudent: 10,
				},
			},
			req: &pb.RetrieveStatisticAssignmentClassRequest{
				ClassId:         "class-id",
				StudyPlanItemId: "study-plan-item-id",
			},
			setup: func(ctx context.Context) {
				studyPlanItems := make([]*entities.StudyPlanItem, 0)
				studyPlanItems = append(studyPlanItems, &entities.StudyPlanItem{
					ID:          database.Text("study-plan-item-id"),
					StudyPlanID: database.Text("study-plan-id"),
				})

				studyPlanItemRepo.On("FindByIDs", ctx, mockDB, mock.Anything).Once().Return(studyPlanItems, nil)
				getTotalFilter := &repositories.CountStudentStudyPlanItemsInClassFilter{
					ClassID:         database.Text("class-id"),
					StudyPlanItemID: database.Text("study-plan-item-id"),
					IsCompleted:     database.Bool(false),
				}
				studyPlanItemRepo.On("CountStudentStudyPlanItemsInClass", ctx, mockDB, getTotalFilter).Once().Return(10, nil)
				getCompletedFilter := &repositories.CountStudentStudyPlanItemsInClassFilter{
					ClassID:         database.Text("class-id"),
					StudyPlanItemID: database.Text("study-plan-item-id"),
					IsCompleted:     database.Bool(true),
				}
				studyPlanItemRepo.On("CountStudentStudyPlanItemsInClass", ctx, mockDB, getCompletedFilter).Once().Return(6, nil)
			},
		},
		{
			name:        "error no row",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.Internal, fmt.Errorf("StudyPlanItemRepo.CountStudentStudyPlanItemsInClass: %w", pgx.ErrNoRows).Error()),
			expectedResp: &pb.RetrieveStatisticAssignmentClassResponse{
				StatisticItem: &pb.StatisticTodoItem{
					Item: &pb.ToDoItem{
						StudyPlanItem: &pb.StudyPlanItem{
							StudyPlanId:     "study-plan-id",
							StudyPlanItemId: "study-plan-item-id",
						},
					},
				},
			},
			req: &pb.RetrieveStatisticAssignmentClassRequest{
				ClassId:         "class-id",
				StudyPlanItemId: "study-plan-item-id",
			},
			setup: func(ctx context.Context) {
				studyPlanItems := make([]*entities.StudyPlanItem, 0)
				studyPlanItems = append(studyPlanItems, &entities.StudyPlanItem{
					ID:          database.Text("study-plan-item-id"),
					StudyPlanID: database.Text("study-plan-id"),
				})
				studyPlanItemRepo.On("FindByIDs", ctx, mockDB, mock.Anything).Once().Return(studyPlanItems, nil)
				getTotalFilter := &repositories.CountStudentStudyPlanItemsInClassFilter{
					ClassID:         database.Text("class-id"),
					StudyPlanItemID: database.Text("study-plan-item-id"),
					IsCompleted:     database.Bool(false),
				}
				studyPlanItemRepo.On("CountStudentStudyPlanItemsInClass", ctx, mockDB, getTotalFilter).Once().Return(0, pgx.ErrNoRows)
				getCompletedFilter := &repositories.CountStudentStudyPlanItemsInClassFilter{
					ClassID:         database.Text("class-id"),
					StudyPlanItemID: database.Text("study-plan-item-id"),
					IsCompleted:     database.Bool(true),
				}
				studyPlanItemRepo.On("CountStudentStudyPlanItemsInClass", ctx, mockDB, getCompletedFilter).Once().Return(2, nil)
			},
		},
	}

	for _, testcase := range testcases {
		testcase.setup(testcase.ctx)
		result, err := svc.RetrieveStatisticAssignmentClass(testcase.ctx, testcase.req.(*pb.RetrieveStatisticAssignmentClassRequest))
		assert.Equal(t, testcase.expectedErr, err)
		if testcase.expectedErr == nil {
			expectedResp := testcase.expectedResp.(*pb.RetrieveStatisticAssignmentClassResponse)
			assert.Equal(t, expectedResp.StatisticItem.Item.StudyPlanItem.StudyPlanItemId, result.StatisticItem.Item.StudyPlanItem.StudyPlanItemId)
			assert.Equal(t, expectedResp.StatisticItem.Item.StudyPlanItem.StudyPlanId, result.StatisticItem.Item.StudyPlanItem.StudyPlanId)
			assert.Equal(t, expectedResp.StatisticItem.CompletedStudent, result.StatisticItem.CompletedStudent)
			assert.Equal(t, expectedResp.StatisticItem.TotalAssignedStudent, result.StatisticItem.TotalAssignedStudent)
		}
	}
	return
}
