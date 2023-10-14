package services

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	natsJS "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestLearningObjectiveModifierService_UpsertLOs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	learningObjectiveRepo := new(mock_repositories.MockLearningObjectiveRepo)
	topicsLearningObjectivesRepo := new(mock_repositories.MockTopicsLearningObjectivesRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)
	bookChapterRepo := new(mock_repositories.MockBookChapterRepo)
	jsm := new(mock_nats.JetStreamManagement)

	newLo1 := &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Name:         "name-1",
			Country:      cpb.Country_COUNTRY_VN,
			Subject:      cpb.Subject_SUBJECT_ENGLISH,
			Grade:        1,
			SchoolId:     1,
			DisplayOrder: 1,
			MasterId:     "master-id-1",
			IconUrl:      "icon-url-1",
			UpdatedAt:    timestamppb.New(time.Now()),
			CreatedAt:    timestamppb.New(time.Now()),
		},
		TopicId:        "topic-id-1",
		Video:          "video-1",
		StudyGuide:     "study-guide-1",
		GradeToPass:    wrapperspb.Int32(1),
		ManualGrading:  true,
		TimeLimit:      wrapperspb.Int32(1),
		ApproveGrading: false,
		GradeCapping:   false,
		VendorType:     cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_NONE,
	}

	newLo2 := &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Name:         "name-1",
			Country:      cpb.Country_COUNTRY_VN,
			Subject:      cpb.Subject_SUBJECT_ENGLISH,
			Grade:        1,
			SchoolId:     1,
			DisplayOrder: 1,
			MasterId:     "master-id-1",
			IconUrl:      "icon-url-1",
			UpdatedAt:    timestamppb.New(time.Now()),
			CreatedAt:    timestamppb.New(time.Now()),
		},
		TopicId:        "topic-id-1",
		Video:          "video-1",
		StudyGuide:     "study-guide-1",
		GradeToPass:    wrapperspb.Int32(1),
		ManualGrading:  true,
		TimeLimit:      wrapperspb.Int32(1),
		ApproveGrading: false,
		GradeCapping:   false,
		VendorType:     cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_NONE,
	}

	lo1 := &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:           idutil.ULIDNow(),
			Name:         "name-1",
			Country:      cpb.Country_COUNTRY_VN,
			Subject:      cpb.Subject_SUBJECT_ENGLISH,
			Grade:        1,
			SchoolId:     1,
			DisplayOrder: 1,
			MasterId:     "master-id-1",
			IconUrl:      "icon-url-1",
			UpdatedAt:    timestamppb.New(time.Now()),
			CreatedAt:    timestamppb.New(time.Now()),
		},
		TopicId:        "topic-id-1",
		Video:          "video-1",
		StudyGuide:     "study-guide-1",
		GradeToPass:    wrapperspb.Int32(1),
		ManualGrading:  true,
		TimeLimit:      wrapperspb.Int32(1),
		ApproveGrading: false,
		GradeCapping:   false,
		VendorType:     cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_NONE,
	}
	lo2 := &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:           idutil.ULIDNow(),
			Name:         "name-2",
			Country:      cpb.Country_COUNTRY_VN,
			Subject:      cpb.Subject_SUBJECT_ENGLISH,
			Grade:        1,
			SchoolId:     1,
			DisplayOrder: 1,
			MasterId:     "master-id-2",
			IconUrl:      "icon-url-2",
			UpdatedAt:    timestamppb.New(time.Now()),
			CreatedAt:    timestamppb.New(time.Now()),
		},
		TopicId:        "topic-id-1",
		Video:          "video-2",
		StudyGuide:     "study-guide-2",
		GradeToPass:    wrapperspb.Int32(1),
		ManualGrading:  true,
		TimeLimit:      wrapperspb.Int32(1),
		MaximumAttempt: wrapperspb.Int32(10),
		ApproveGrading: true,
		GradeCapping:   true,
		ReviewOption:   cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE,
		VendorType:     cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE,
	}

	lo3 := &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:           idutil.ULIDNow(),
			Name:         "name-2",
			Country:      cpb.Country_COUNTRY_VN,
			Subject:      cpb.Subject_SUBJECT_ENGLISH,
			Grade:        1,
			SchoolId:     1,
			DisplayOrder: 1,
			MasterId:     "master-id-2",
			IconUrl:      "icon-url-2",
			UpdatedAt:    timestamppb.New(time.Now()),
			CreatedAt:    timestamppb.New(time.Now()),
		},
		TopicId:        "topic-id-1",
		Video:          "video-2",
		StudyGuide:     "study-guide-2",
		GradeToPass:    wrapperspb.Int32(1),
		ManualGrading:  true,
		TimeLimit:      wrapperspb.Int32(1),
		MaximumAttempt: wrapperspb.Int32(10),
		ApproveGrading: true,
		GradeCapping:   true,
		ReviewOption:   cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE,
		VendorType:     cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY,
	}

	topics := []*entities.Topic{
		{
			ID:                    database.Text("topic-id-1"),
			LODisplayOrderCounter: database.Int4(0),
		},
	}

	testCases := []TestCase{
		{
			name: "err PublishContext",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					newLo1,
				},
			},
			expectedErr: fmt.Errorf("s.JSM.PublishContext: subject: %q, %v", constants.SubjectLearningObjectivesCreated, ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicsLearningObjectivesRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateTotalLOs", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectLearningObjectivesCreated, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
				bookChapterRepo.On("RetrieveContentStructuresByLOs", ctx, mock.Anything, mock.Anything).Once().Return(map[string]entities.ContentStructure{}, nil)
			},
		},
		{
			name: "err bookChapterRepo.RetrieveContentStructuresByLOs",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					newLo2,
				},
			},
			expectedErr: status.Errorf(codes.Internal, "cm.BookChapterRepo.RetrieveContentStructuresByLOs: %v", ErrSomethingWentWrong),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicsLearningObjectivesRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateTotalLOs", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				bookChapterRepo.On("RetrieveContentStructuresByLOs", ctx, mock.Anything, mock.Anything).Once().Return(map[string]entities.ContentStructure{}, ErrSomethingWentWrong)
			},
		},
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicsLearningObjectivesRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateTotalLOs", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectLearningObjectivesCreated, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
				bookChapterRepo.On("RetrieveContentStructuresByLOs", ctx, mock.Anything, mock.Anything).Once().Return(map[string]entities.ContentStructure{}, nil)
			},
		}, {
			name: "Happy case create lo with default vendor type is MANABIE",
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("BulkImport", ctx, tx, mock.MatchedBy(func(los []*entities.LearningObjective) bool {

					inputVendorTypes := []string{
						los[0].VendorType.String, los[1].VendorType.String, los[2].VendorType.String,
					}

					expectVendorTypes := []string{
						cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String(),
						cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String(),
						cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String(),
					}

					if assert.Equal(t, expectVendorTypes, inputVendorTypes) {
						return true
					}
					return false
				})).Return(nil)

				topicsLearningObjectivesRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateTotalLOs", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectLearningObjectivesCreated, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
				bookChapterRepo.On("RetrieveContentStructuresByLOs", ctx, mock.Anything, mock.Anything).Once().Return(map[string]entities.ContentStructure{}, nil)
			},
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2, lo3,
				},
			},
			expectedErr: nil,
		},
		{
			name: "error TopicRepo.RetrieveByIDs",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to retrieve topics by ids: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "error TopicRepo.RetrieveByID",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to retrieve topic by id: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "error TopicRepo.isTopicsExisted",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "some topics does not exists"),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Topic{
					{
						ID:                    database.Text("topic-id-2"),
						LODisplayOrderCounter: database.Int4(0),
					},
				}, nil)
			},
		},
		{
			name: "error LearningObjectiveRepo.BulkImport",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to bulk import learning objective: %w", pgx.ErrTxCommitRollback).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				learningObjectiveRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxCommitRollback)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
			},
		},
		{
			name: "error TopicsLearningObjectivesRepo.BulkImport",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to bulk import topic learning objective: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				learningObjectiveRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicsLearningObjectivesRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
			},
		},
		{
			name: "error cm.UpdateLODisplayOrderCounter",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to update lo display order counter: %w", pgx.ErrTxCommitRollback).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				learningObjectiveRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicsLearningObjectivesRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxCommitRollback)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
			},
		},
		{
			name: "error cm.updateTotalLOs",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1, lo2,
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to update total learing objectives: %w", pgx.ErrTxCommitRollback).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				learningObjectiveRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicsLearningObjectivesRepo.On("BulkImport", ctx, tx, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateLODisplayOrderCounter", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByID", ctx, tx, mock.Anything, mock.Anything).Once().Return(topics[0], nil)
				topicRepo.On("UpdateTotalLOs", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxCommitRollback)
			},
		},
		{
			name: "error validate",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &epb.UpsertLOsRequest{
				LearningObjectives: []*cpb.LearningObjective{
					lo1,
					{
						Info: &cpb.ContentBasicInfo{
							Id:           idutil.ULIDNow(),
							Name:         "name-2",
							Country:      cpb.Country_COUNTRY_VN,
							Subject:      cpb.Subject_SUBJECT_ENGLISH,
							Grade:        1,
							SchoolId:     1,
							DisplayOrder: 1,
							MasterId:     "master-id-2",
							IconUrl:      "icon-url-2",
							UpdatedAt:    timestamppb.New(time.Now()),
							CreatedAt:    timestamppb.New(time.Now()),
						},
						MaximumAttempt: wrapperspb.Int32(100),
						ApproveGrading: true,
						GradeCapping:   true,
						ReviewOption:   cpb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE,
					},
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "learningObjectives[1].maximum_attempt must be Null or between 1 to 99"),
			setup:       func(ctx context.Context) {},
		},
	}

	s := &LearningObjectiveModifierService{
		DB:                           db,
		JSM:                          jsm,
		TopicRepo:                    topicRepo,
		LearningObjectiveRepo:        learningObjectiveRepo,
		TopicsLearningObjectivesRepo: topicsLearningObjectivesRepo,
		BookChapterRepo:              bookChapterRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*epb.UpsertLOsRequest)
			_, err := s.UpsertLOs(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestLearningObjectiveModifierService_DeleteLos(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	learningObjectiveRepo := new(mock_repositories.MockLearningObjectiveRepo)
	topicLearningObjectiveRepo := new(mock_repositories.MockTopicsLearningObjectivesRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)
	bookChapterRepo := new(mock_repositories.MockBookChapterRepo)
	loStudyPlanItemRepo := new(mock_repositories.MockLoStudyPlanItemRepo)
	jsm := new(mock_nats.JetStreamManagement)

	s := &LearningObjectiveModifierService{
		DB:                           db,
		JSM:                          jsm,
		TopicRepo:                    topicRepo,
		LearningObjectiveRepo:        learningObjectiveRepo,
		TopicsLearningObjectivesRepo: topicLearningObjectiveRepo,
		BookChapterRepo:              bookChapterRepo,
		LoStudyPlanItemRepo:          loStudyPlanItemRepo,
	}

	req := &epb.DeleteLosRequest{
		LoIds: []string{"lo-id-1", "lo-id-2"},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						ID: database.Text("lo-id-1"),
					},
					{
						ID: database.Text("lo-id-2"),
					},
				}, nil)
				learningObjectiveRepo.On("SoftDeleteWithLoIDs", ctx, tx, mock.Anything).Once().Return(int64(0), nil)
				loStudyPlanItemRepo.On("DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs", ctx, tx, mock.Anything).Once().Return(nil)
				topicLearningObjectiveRepo.On("SoftDeleteByLoIDs", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "err LearningObjectiveRepo.RetrieveByIDs",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Errorf(codes.NotFound, fmt.Errorf("LearningObjectiveRepo.RetrieveByIDs: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return([]*entities.LearningObjective{}, ErrSomethingWentWrong)
			},
		},
		{
			name:        "err LoRepo.SoftDeleteWithLoIDs",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("s.LoRepo.SoftDeleteWithLoIDs: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						ID: database.Text("lo-id-1"),
					},
					{
						ID: database.Text("lo-id-2"),
					},
				}, nil)
				learningObjectiveRepo.On("SoftDeleteWithLoIDs", ctx, tx, mock.Anything).Once().Return(int64(0), ErrSomethingWentWrong)
			},
		},
		{
			name:        "err loStudyPlanItemRepo.DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("LoStudyPlanItemRepo.DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Commit", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						ID: database.Text("lo-id-1"),
					},
					{
						ID: database.Text("lo-id-2"),
					},
				}, nil)
				learningObjectiveRepo.On("SoftDeleteWithLoIDs", ctx, tx, mock.Anything).Once().Return(int64(0), nil)
				loStudyPlanItemRepo.On("DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
		{
			name:        "err topicLearningObjectiveRepo.SoftDeleteByLoIDs",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("TopicLearningObjectiveRepo.SoftDeleteByLoIDs: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Times(2).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Times(2).Return(nil)

				learningObjectiveRepo.On("RetrieveByIDs", ctx, db, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						ID: database.Text("lo-id-1"),
					},
					{
						ID: database.Text("lo-id-2"),
					},
				}, nil)
				learningObjectiveRepo.On("SoftDeleteWithLoIDs", ctx, tx, mock.Anything).Once().Return(int64(0), nil)
				loStudyPlanItemRepo.On("DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs", ctx, tx, mock.Anything).Once().Return(nil)
				topicLearningObjectiveRepo.On("SoftDeleteByLoIDs", ctx, tx, mock.Anything).Once().Return(ErrSomethingWentWrong)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*epb.DeleteLosRequest)
			_, err := s.DeleteLos(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func compareTopicsLOs(src, target *entities.TopicsLearningObjectives) bool {
	return src.LoID.String == target.LoID.String && src.TopicID.String == target.TopicID.String && src.DisplayOrder.Int == target.DisplayOrder.Int
}

func matchedByTopicsLOsArray(src []*entities.TopicsLearningObjectives) interface{} {
	return mock.MatchedBy(func(target []*entities.TopicsLearningObjectives) bool {
		if len(src) != len(target) {
			return false
		}
		for i := range target {
			if !compareTopicsLOs(src[i], target[i]) {
				return false
			}
		}
		return true
	})
}

func TestLearningObjectiveModifierService_UpdateDisplayOrdersOfLOs(t *testing.T) {
	ctx := context.Background()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	topicRepo := new(mock_repositories.MockTopicRepo)
	topicLearningObjectiveRepo := new(mock_repositories.MockTopicsLearningObjectivesRepo)
	learningObjectiveRepo := new(mock_repositories.MockLearningObjectiveRepo)

	s := &LearningObjectiveModifierService{
		DB:                           db,
		TopicRepo:                    topicRepo,
		LearningObjectiveRepo:        learningObjectiveRepo,
		TopicsLearningObjectivesRepo: topicLearningObjectiveRepo,
	}

	testCases := []TestCase{
		{
			name: "missing some fields",
			ctx:  ctx,
			req: []*epb.TopicLODisplayOrder{
				{LoId: "lo-1", TopicId: "topic-1", DisplayOrder: 1},
				{LoId: "lo-2", TopicId: "topic-1", DisplayOrder: 2},
				{LoId: "", TopicId: "", DisplayOrder: 0},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				topicRepo.On("RetrieveByID", ctx, tx, database.Text("topic-1"), mock.Anything).Return(&entities.Topic{}, nil)

				topicLearningObjectiveRepo.On("BulkUpdateDisplayOrder", ctx, tx, matchedByTopicsLOsArray([]*entities.TopicsLearningObjectives{
					{LoID: database.Text("lo-1"), TopicID: database.Text("topic-1"), DisplayOrder: database.Int2(1)},
					{LoID: database.Text("lo-2"), TopicID: database.Text("topic-1"), DisplayOrder: database.Int2(2)},
				})).Return(nil).Once()

				learningObjectiveRepo.On("UpdateDisplayOrders", ctx, tx, map[pgtype.Text]pgtype.Int2{
					database.Text("lo-1"): database.Int2(1),
					database.Text("lo-2"): database.Int2(2),
				}).Return(nil).Once()

				tx.On("Commit", ctx).Return(nil).Once()
			},
			expectedResp: []*epb.TopicLO{
				{LoId: "lo-1", TopicId: "topic-1"},
				{LoId: "lo-2", TopicId: "topic-1"},
			},
		},
		{
			name: "happy case",
			ctx:  ctx,
			req: []*epb.TopicLODisplayOrder{
				{LoId: "lo-1", TopicId: "topic-1", DisplayOrder: 1},
				{LoId: "lo-2", TopicId: "topic-1", DisplayOrder: 2},
				{LoId: "lo-3", TopicId: "topic-2", DisplayOrder: 1},
			},
			setup: func(ctx context.Context) {
				// Topic-1
				db.On("Begin", ctx).Return(tx, nil).Once()
				topicRepo.On("RetrieveByID", ctx, tx, database.Text("topic-1"), mock.Anything).Return(&entities.Topic{}, nil)
				topicLearningObjectiveRepo.On("BulkUpdateDisplayOrder", ctx, tx, matchedByTopicsLOsArray([]*entities.TopicsLearningObjectives{
					{LoID: database.Text("lo-1"), TopicID: database.Text("topic-1"), DisplayOrder: database.Int2(1)},
					{LoID: database.Text("lo-2"), TopicID: database.Text("topic-1"), DisplayOrder: database.Int2(2)},
				})).Return(nil).Once()
				learningObjectiveRepo.On("UpdateDisplayOrders", ctx, tx, map[pgtype.Text]pgtype.Int2{
					database.Text("lo-1"): database.Int2(1),
					database.Text("lo-2"): database.Int2(2),
				}).Return(nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				// Topic-2
				db.On("Begin", ctx).Return(tx, nil).Once()
				topicRepo.On("RetrieveByID", ctx, tx, database.Text("topic-2"), mock.Anything).Return(&entities.Topic{}, nil)
				topicLearningObjectiveRepo.On("BulkUpdateDisplayOrder", ctx, tx, matchedByTopicsLOsArray([]*entities.TopicsLearningObjectives{
					{LoID: database.Text("lo-3"), TopicID: database.Text("topic-2"), DisplayOrder: database.Int2(1)},
				})).Return(nil).Once()
				learningObjectiveRepo.On("UpdateDisplayOrders", ctx, tx, map[pgtype.Text]pgtype.Int2{
					database.Text("lo-3"): database.Int2(1),
				}).Return(nil).Once()

				tx.On("Commit", ctx).Return(nil).Once()
			},
			expectedResp: []*epb.TopicLO{
				{LoId: "lo-1", TopicId: "topic-1"},
				{LoId: "lo-2", TopicId: "topic-1"},
				{LoId: "lo-3", TopicId: "topic-2"},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.UpdateDisplayOrdersOfLOs(testCase.ctx, testCase.req.([]*epb.TopicLODisplayOrder))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedResp.([]*epb.TopicLO), resp)
			}
		})
	}
}

func Test_validateUpsertLO(t *testing.T) {
	ctx := context.Background()

	rand.Seed(time.Now().Unix())
	maxRandom := rand.Intn(99-1) + 1

	testCases := []TestCase{
		{
			ctx:  ctx,
			name: "happy case",
			req: &cpb.LearningObjective{
				MaximumAttempt: wrapperspb.Int32(int32(maxRandom)),
			},
			expectedErr: nil,
		},
		{
			ctx:  ctx,
			name: "err case < 1",
			req: &cpb.LearningObjective{
				MaximumAttempt: wrapperspb.Int32(0),
			},
			expectedErr: fmt.Errorf("maximum_attempt must be Null or between 1 to 99"),
		},
		{
			ctx:  ctx,
			name: "err case < 99",
			req: &cpb.LearningObjective{
				MaximumAttempt: wrapperspb.Int32(100),
			},
			expectedErr: fmt.Errorf("maximum_attempt must be Null or between 1 to 99"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUpsertLO(tc.req.(*cpb.LearningObjective))
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLearningObjectiveModifierService_UpdateLearningObjectiveName(t *testing.T) {
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	mockLearningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepo{}
	s := &LearningObjectiveModifierService{
		DB:                    mockDB,
		LearningObjectiveRepo: mockLearningObjectiveRepo,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockLearningObjectiveRepo.On("UpdateName", mock.Anything, mockDB, database.Text("id"), database.Text("name")).Return(int64(1), nil)
			},
			req: &epb.UpdateLearningObjectiveNameRequest{
				LoId:                     "id",
				NewLearningObjectiveName: "name",
			},
		},
		{
			name: "no row",
			setup: func(ctx context.Context) {
				mockLearningObjectiveRepo.On("UpdateName", mock.Anything, mockDB, database.Text("id-1"), database.Text("name")).Return(int64(0), nil)
			},
			req: &epb.UpdateLearningObjectiveNameRequest{
				LoId:                     "id-1",
				NewLearningObjectiveName: "name",
			},
			expectedErr: status.Error(codes.NotFound, fmt.Errorf("s.LearningObjectiveRepo.UpdateName not found any learning objective to update name: %w", pgx.ErrNoRows).Error()),
		},
		{
			name: "internal error",
			setup: func(ctx context.Context) {
				mockLearningObjectiveRepo.On("UpdateName", mock.Anything, mockDB, database.Text("id-2"), database.Text("name")).Return(int64(0), pgx.ErrTxClosed)
			},
			req: &epb.UpdateLearningObjectiveNameRequest{
				LoId:                     "id-2",
				NewLearningObjectiveName: "name",
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("s.LearningObjectiveRepo.UpdateName: %w", pgx.ErrTxClosed).Error()),
		},
		{
			name: "missing lo id",
			setup: func(ctx context.Context) {
			},
			req: &epb.UpdateLearningObjectiveNameRequest{
				NewLearningObjectiveName: "randomName",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateUpdateLearningObjectiveNameRequest: missing field LoId"),
		},
		{
			name: "missing name",
			setup: func(ctx context.Context) {
			},
			req: &epb.UpdateLearningObjectiveNameRequest{
				LoId: "randomID",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateUpdateLearningObjectiveNameRequest: missing field NewLearningObjectiveName"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			_, err := s.UpdateLearningObjectiveName(ctx, testCase.req.(*epb.UpdateLearningObjectiveNameRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
