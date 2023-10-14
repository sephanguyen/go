package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	"github.com/nats-io/nats.go"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestStudyPlanService_UpsertIndividualStudyPlan(t *testing.T) {
	t.Parallel()
	individualStudyPlanRepo := &mock_repositories.MockIndividualStudyPlan{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudyPlanService{
		DB:                          mockDB,
		IndividualStudyPlanItemRepo: individualStudyPlanRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.UpsertIndividualInfoRequest{
				IndividualItems: []*sspb.StudyPlanItem{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							StudyPlanId:        "study-plan-id-1",
							LearningMaterialId: "learning-material-id-1",
							StudentId: &wrapperspb.StringValue{
								Value: "student-id-1",
							},
						},
						AvailableFrom: timestamppb.Now(),
						AvailableTo:   timestamppb.Now(),
						StartDate:     timestamppb.Now(),
						EndDate:       timestamppb.Now(),
						Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
						SchoolDate:    timestamppb.Now(),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				individualStudyPlanRepo.On("BulkSync", ctx, tx, mock.Anything).Once().Return([]*entities.IndividualStudyPlan{}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error student id is empty",
			req: &sspb.UpsertIndividualInfoRequest{
				IndividualItems: []*sspb.StudyPlanItem{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							LearningMaterialId: "learning-material-id-1",
							StudyPlanId:        "study-plan-id-1",
						},
						AvailableFrom: timestamppb.Now(),
						AvailableTo:   timestamppb.Now(),
						StartDate:     timestamppb.Now(),
						EndDate:       timestamppb.Now(),
						Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
						SchoolDate:    timestamppb.Now(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertIndividualRequest: %w", fmt.Errorf("Student id must not be empty")).Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "error learning material id is empty",
			req: &sspb.UpsertIndividualInfoRequest{
				IndividualItems: []*sspb.StudyPlanItem{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							StudyPlanId: "study-plan-id-1",
							StudentId: &wrapperspb.StringValue{
								Value: "student-id-1",
							},
						},
						AvailableFrom: timestamppb.Now(),
						AvailableTo:   timestamppb.Now(),
						StartDate:     timestamppb.Now(),
						EndDate:       timestamppb.Now(),
						Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
						SchoolDate:    timestamppb.Now(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertIndividualRequest: %w", fmt.Errorf("Learning material id must not be empty")).Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "error study plan id is empty",
			req: &sspb.UpsertIndividualInfoRequest{
				IndividualItems: []*sspb.StudyPlanItem{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							LearningMaterialId: "learning-material-id-1",
							StudentId: &wrapperspb.StringValue{
								Value: "student-id-1",
							},
						},
						AvailableFrom: timestamppb.Now(),
						AvailableTo:   timestamppb.Now(),
						StartDate:     timestamppb.Now(),
						EndDate:       timestamppb.Now(),
						Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
						SchoolDate:    timestamppb.Now(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertIndividualRequest: %w", fmt.Errorf("Study plan id must not be empty")).Error()),
			setup:       func(ctx context.Context) {},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.UpsertIndividual(ctx, testCase.req.(*sspb.UpsertIndividualInfoRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestStudyPlanService_ImportStudyPlan(t *testing.T) {
	t.Parallel()
	masterStudyPlanRepo := &mock_repositories.MockMasterStudyPlanRepo{}
	importStudyPlanTaskRepo := &mock_repositories.MockImportStudyPlanTaskRepo{}
	jsm := &mock_nats.JetStreamManagement{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudyPlanService{
		DB:                      mockDB,
		MasterStudyPlanRepo:     masterStudyPlanRepo,
		ImportStudyPlanTaskRepo: importStudyPlanTaskRepo,
		JSM:                     jsm,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.ImportStudyPlanRequest{
				StudyPlanItems: []*sspb.StudyPlanItemImport{
					{
						StudyPlanId:        "study-plan-id",
						LearningMaterialId: "lm-id",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				masterStudyPlanRepo.On("FindByID", ctx, mockDB, mock.Anything).Once().Return([]*entities.MasterStudyPlan{
					{
						StudyPlanID:        database.Text("study-plan-id"),
						LearningMaterialID: database.Text("lm-id"),
					},
				}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				importStudyPlanTaskRepo.On("Insert", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStudyPlanItemsImported, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.ImportStudyPlan(ctx, testCase.req.(*sspb.ImportStudyPlanRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestStudyPlanService_UpsertSchoolDate(t *testing.T) {
	t.Parallel()
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudyPlanService{
		DB:                db,
		StudyPlanItemRepo: studyPlanItemRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.UpsertSchoolDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
						StudentId:          wrapperspb.String("student_id_1"),
					},
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
						StudentId:          wrapperspb.String("student_id_2"),
					},
				},
				SchoolDate: timestamppb.Now(),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				studyPlanItemRepo.On("BulkUpdateSchoolDate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studyPlanItemRepo.On("ListSPItemByIdentity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{"1"}, nil)
			},
		},
		{
			name: "error learning material id is empty",
			req: &sspb.UpsertSchoolDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId: "study_plan_id",
						StudentId:   wrapperspb.String("student_id_1"),
					},
					{
						StudyPlanId: "study_plan_id",
						StudentId:   wrapperspb.String("student_id_2"),
					},
				},
				SchoolDate: timestamppb.Now(),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertSchoolDateRequest: validateSPItemIdentities: %w", fmt.Errorf("learning material id must not be empty")).Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "error study plan id is empty",
			req: &sspb.UpsertSchoolDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						LearningMaterialId: "learning_material_id",
						StudentId:          wrapperspb.String("student_id_1"),
					},
					{
						LearningMaterialId: "learning_material_id",
						StudentId:          wrapperspb.String("student_id_2"),
					},
				},
				SchoolDate: timestamppb.Now(),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertSchoolDateRequest: validateSPItemIdentities: %w", fmt.Errorf("study plan id must not be empty")).Error()),
			setup:       func(ctx context.Context) {},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.UpsertSchoolDate(ctx, testCase.req.(*sspb.UpsertSchoolDateRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestStudyPlanService_UpsertMasterInfo(t *testing.T) {
	t.Parallel()
	masterStudyPlanRepo := &mock_repositories.MockMasterStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudyPlanService{
		DB:                  mockDB,
		MasterStudyPlanRepo: masterStudyPlanRepo,
	}

	validReq := &sspb.UpsertMasterInfoRequest{
		MasterItems: []*sspb.MasterStudyPlan{
			{
				MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
					StudyPlanId:        "study-plan-id-1",
					LearningMaterialId: "learning-material-id-1",
				},
				AvailableFrom: timestamppb.Now(),
				AvailableTo:   timestamppb.Now(),
				StartDate:     timestamppb.Now(),
				EndDate:       timestamppb.Now(),
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
				SchoolDate:    timestamppb.Now(),
			},
			{
				MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
					StudyPlanId:        "study-plan-id-2",
					LearningMaterialId: "learning-material-id-2",
				},
				AvailableFrom: timestamppb.Now(),
				AvailableTo:   timestamppb.Now(),
				StartDate:     timestamppb.Now(),
				EndDate:       timestamppb.Now(),
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
				SchoolDate:    timestamppb.Now(),
			},
			{
				MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
					StudyPlanId:        "study-plan-id-3",
					LearningMaterialId: "learning-material-id-3",
				},
				AvailableFrom: timestamppb.Now(),
				AvailableTo:   timestamppb.Now(),
				StartDate:     timestamppb.Now(),
				EndDate:       timestamppb.Now(),
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
				SchoolDate:    timestamppb.Now(),
			},
		},
	}

	testCases := []TestCase{
		{
			name: "error study plan id is empty",
			req: &sspb.UpsertMasterInfoRequest{
				MasterItems: []*sspb.MasterStudyPlan{
					{
						MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
							StudyPlanId:        "",
							LearningMaterialId: "learning-material-id-1",
						},
						AvailableFrom: timestamppb.Now(),
						AvailableTo:   timestamppb.Now(),
						StartDate:     timestamppb.Now(),
						EndDate:       timestamppb.Now(),
						Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
						SchoolDate:    timestamppb.Now(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertMasterInfoRequest: %w", fmt.Errorf("StudyPlanId is empty")).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				masterStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error learning material id is empty",
			req: &sspb.UpsertMasterInfoRequest{
				MasterItems: []*sspb.MasterStudyPlan{
					{
						MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
							StudyPlanId:        "study-plan-id-1",
							LearningMaterialId: "",
						},
						AvailableFrom: timestamppb.Now(),
						AvailableTo:   timestamppb.Now(),
						StartDate:     timestamppb.Now(),
						EndDate:       timestamppb.Now(),
						Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
						SchoolDate:    timestamppb.Now(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertMasterInfoRequest: %w", fmt.Errorf("LearningMaterialId is empty")).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				masterStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error end date before start date",
			req: &sspb.UpsertMasterInfoRequest{
				MasterItems: []*sspb.MasterStudyPlan{
					{
						MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
							StudyPlanId:        "study-plan-id-1",
							LearningMaterialId: "learning-material-id-1",
						},
						AvailableFrom: timestamppb.Now(),
						AvailableTo:   timestamppb.Now(),
						StartDate:     timestamppb.New(time.Now().Add(time.Hour)),
						EndDate:       timestamppb.Now(),
						Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
						SchoolDate:    timestamppb.Now(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertMasterInfoRequest: %w", fmt.Errorf("StudyPlanId: %s, LearningMaterialId: %s, end_date before start_date", "study-plan-id-1", "learning-material-id-1")).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				masterStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "error available to before available from",
			req: &sspb.UpsertMasterInfoRequest{
				MasterItems: []*sspb.MasterStudyPlan{
					{
						MasterStudyPlanIdentify: &sspb.MasterStudyPlanIdentify{
							StudyPlanId:        "study-plan-id-1",
							LearningMaterialId: "learning-material-id-1",
						},
						AvailableFrom: timestamppb.New(time.Now().Add(time.Hour)),
						AvailableTo:   timestamppb.Now(),
						StartDate:     timestamppb.Now(),
						EndDate:       timestamppb.Now(),
						Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
						SchoolDate:    timestamppb.Now(),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertMasterInfoRequest: %w", fmt.Errorf("StudyPlanId: %s, LearningMaterialId: %s, available_to before available_from", "study-plan-id-1", "learning-material-id-1")).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				masterStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				masterStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.UpsertMasterInfo(ctx, testCase.req.(*sspb.UpsertMasterInfoRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestStudyPlanService_ListAllocateTeacher(t *testing.T) {
	t.Parallel()
	allocateMarkerRepo := &mock_repositories.MockAllocateMarkerRepo{}
	mockDB := &mock_database.Ext{}
	svc := &StudyPlanService{
		DB:                 mockDB,
		AllocateMarkerRepo: allocateMarkerRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.ListAllocateTeacherRequest{
				LocationIds: []string{"location_id"},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				allocateMarkerRepo.On("GetAllocateTeacherByCourseAccess", mock.Anything, mockDB, mock.Anything).Once().Return([]*entities.AllocateTeacherItem{}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.ListAllocateTeacher(ctx, testCase.req.(*sspb.ListAllocateTeacherRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestStudyPlanService_UpsertAllocateMarker(t *testing.T) {
	t.Parallel()
	allocateMarkerRepo := &mock_repositories.MockAllocateMarkerRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudyPlanService{
		DB:                 tx,
		AllocateMarkerRepo: allocateMarkerRepo,
	}

	testCases := []TestCase{
		{
			name: "submission is empty",
			req: &sspb.UpsertAllocateMarkerRequest{
				Submissions: []*sspb.UpsertAllocateMarkerRequest_SubmissionItem{
					{
						SubmissionId: "submission-id",
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertAllocateMarkerRequest: %w", fmt.Errorf("submission must be not empty")).Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "number of allocated submission is zero",
			req: &sspb.UpsertAllocateMarkerRequest{
				Submissions: []*sspb.UpsertAllocateMarkerRequest_SubmissionItem{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{},
					},
				},
				AllocateMarkers: []*sspb.UpsertAllocateMarkerRequest_AllocateMarkerItem{
					{
						NumberAllocated: 0,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertAllocateMarkerRequest: %w", fmt.Errorf("number of allocated submission must be not less or equal than zero")).Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "total allocated submission does not equal the total of submission selected",
			req: &sspb.UpsertAllocateMarkerRequest{
				Submissions: []*sspb.UpsertAllocateMarkerRequest_SubmissionItem{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{},
					},
				},
				AllocateMarkers: []*sspb.UpsertAllocateMarkerRequest_AllocateMarkerItem{
					{
						NumberAllocated: 2,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpsertAllocateMarkerRequest: %w", fmt.Errorf("total allocated submission does not equal the total of submission selected")).Error()),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "happy case",
			req: &sspb.UpsertAllocateMarkerRequest{
				Submissions: []*sspb.UpsertAllocateMarkerRequest_SubmissionItem{
					{
						StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
							StudyPlanId:        "study-plan-id",
							LearningMaterialId: "lm-id",
							StudentId:          wrapperspb.String("student-id"),
						},
					},
				},
				AllocateMarkers: []*sspb.UpsertAllocateMarkerRequest_AllocateMarkerItem{
					{
						NumberAllocated: 1,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
				allocateMarkerRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.UpsertAllocateMarker(ctx, testCase.req.(*sspb.UpsertAllocateMarkerRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestStudyPlanService_UpdateStudyPlanItemsStartEndDate(t *testing.T) {
	t.Parallel()

	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	studyPlanService := &StudyPlanService{
		DB:                db,
		StudyPlanItemRepo: studyPlanItemRepo,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: timestamppb.New(time.Now()),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				studyPlanItemRepo.On("BulkUpdateStartEndDate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(int64(1), nil)
				studyPlanItemRepo.On("ListSPItemByIdentity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{"1"}, nil)
			},
		},
		{
			name: "missing study plan id",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: timestamppb.New(time.Now()),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: validateSPItemIdentities: %s", fmt.Errorf("study plan id must not be empty")).Error()),
		},
		{
			name: "missing learning material id",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId: "study_plan_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: timestamppb.New(time.Now()),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: validateSPItemIdentities: %s", fmt.Errorf("learning material id must not be empty")).Error()),
		},
		{
			name: "student ids is empty",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
						StudentId:          &wrapperspb.StringValue{},
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: timestamppb.New(time.Now()),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: validateSPItemIdentities: %s", fmt.Errorf("student id must be nil or have value")).Error()),
		},
		{
			name: "startdate is null when update field is start",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_START_DATE,
				StartDate: nil,
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: %s", fmt.Errorf("startdate have to not null")).Error()),
		},
		{
			name: "enddate is null when update field is end",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_END_DATE,
				StartDate: timestamppb.New(time.Now()),
				EndDate:   nil,
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: %s", fmt.Errorf("enddate have to not null")).Error()),
		},
		{
			name: "startdate is null when update field is all",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: nil,
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: %s", fmt.Errorf("startdate and enddate have to not null")).Error()),
		},
		{
			name: "enddate is null when update field is all",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: timestamppb.New(time.Now()),
				EndDate:   nil,
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: %s", fmt.Errorf("startdate and enddate have to not null")).Error()),
		},
		{
			name: "start date after end date",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: timestamppb.New(time.Now().Add(time.Hour)),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: %s", fmt.Errorf("startdate after enddate")).Error()),
		},
		{
			name: "invalid update type",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    -1,
				StartDate: timestamppb.New(time.Now().Add(time.Hour)),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateStudyPlanItemsStartEndDateRequest: %s", fmt.Errorf("invalid fields need to update")).Error()),
		},
		{
			name: "StudyPlanItemRepo.BulkUpdateStartEndDate return error",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: timestamppb.New(time.Now()),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: StudyPlanItemRepo.BulkUpdateStartEndDate: %s", fmt.Errorf("error")).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				studyPlanItemRepo.On("BulkUpdateStartEndDate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(int64(1), fmt.Errorf("error"))
				studyPlanItemRepo.On("ListSPItemByIdentity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{"1"}, nil)
			},
		},
		{
			name: "StudyPlanItemRepo.ListSPItemByIdentity return error",
			req: &sspb.UpdateStudyPlanItemsStartEndDateRequest{
				StudyPlanItemIdentities: []*sspb.StudyPlanItemIdentity{
					{
						StudyPlanId:        "study_plan_id",
						LearningMaterialId: "learning_material_id",
					},
				},
				Fields:    sspb.UpdateStudyPlanItemsStartEndDateFields_ALL,
				StartDate: timestamppb.New(time.Now()),
				EndDate:   timestamppb.New(time.Now()),
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: StudyPlanItemRepo.ListSPItemByIdentity: %s", fmt.Errorf("error")).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil)
				studyPlanItemRepo.On("BulkUpdateStartEndDate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(int64(10), nil)
				studyPlanItemRepo.On("ListSPItemByIdentity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, fmt.Errorf("error"))
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			if testCase.setup != nil {
				testCase.setup(ctx)
			}
			_, err := studyPlanService.UpdateStudyPlanItemsStartEndDate(ctx, testCase.req.(*sspb.UpdateStudyPlanItemsStartEndDateRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRetrieveStudyPlanIdentity(t *testing.T) {
	t.Parallel()
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}

	svc := &StudyPlanService{
		StudyPlanRepo: studyPlanRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.RetrieveStudyPlanIdentityRequest{
				StudyPlanItemIds: []string{"study-plan-item-id-1", "study-plan-item-id-2"},
			},
			expectedResp: &sspb.RetrieveStudyPlanIdentityResponse{
				StudyPlanIdentities: []*sspb.StudyPlanIdentity{
					{
						StudyPlanId:        "study-plan-id-1",
						StudentId:          "student-id-1",
						LearningMaterialId: "learning-material-id-1",
						StudyPlanItemId:    "study-plan-item-id-1",
					},
					{
						StudyPlanId:        "study-plan-id-2",
						StudentId:          "student-id-2",
						LearningMaterialId: "learning-material-id-2",
						StudyPlanItemId:    "study-plan-item-id-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				studyPlanRepo.On("RetrieveStudyPlanIdentity", mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]*repositories.RetrieveStudyPlanIdentityResponse{
					{
						StudyPlanID:        database.Text("study-plan-id-1"),
						StudentID:          database.Text("student-id-1"),
						LearningMaterialID: database.Text("learning-material-id-1"),
						StudyPlanItemID:    database.Text("study-plan-item-id-1"),
					},
					{
						StudyPlanID:        database.Text("study-plan-id-2"),
						StudentID:          database.Text("student-id-2"),
						LearningMaterialID: database.Text("learning-material-id-2"),
						StudyPlanItemID:    database.Text("study-plan-item-id-2"),
					},
				}, nil)
			},
		},
		{
			name: "empty study plan item ids",
			req: &sspb.RetrieveStudyPlanIdentityRequest{
				StudyPlanItemIds: []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing study plan item ids"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "not found",
			req: &sspb.RetrieveStudyPlanIdentityRequest{
				StudyPlanItemIds: []string{"study-plan-item-id-1", "study-plan-item-id-2"},
			},
			expectedErr: status.Errorf(codes.NotFound, "StudyPlanRepo.RetrieveStudyPlanIdentity: %s", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				studyPlanRepo.On("RetrieveStudyPlanIdentity", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			resp, err := svc.RetrieveStudyPlanIdentity(ctx, testCase.req.(*sspb.RetrieveStudyPlanIdentityRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestBulkUpdateStudyPlanItemStatus(t *testing.T) {
	//TODO: do later
}

func Test_ListToDoItems(t *testing.T) {
	now := time.Now().UTC()
	t.Parallel()
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	svc := &StudyPlanService{
		DB:            mockDB,
		StudyPlanRepo: studyPlanRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.ListToDoItemRequest{
				Status:    sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_ACTIVE,
				StudentId: "student-id",
				Page: &cpb.Paging{
					Limit: 10,
				},
				CourseIds: []string{"course-id-1", "course-id-2"},
			},
			expectedErr: nil,
			expectedResp: &sspb.ListToDoItemResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString: "learning-material-id",
							OffsetTime:   timestamppb.New(now),
						},
					},
				},
				TodoItems: []*sspb.StudyPlanToDoItem{
					{
						Status: sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_ACTIVE,
						Crown:  0,
						IndividualStudyPlanItem: &sspb.StudyPlanItem{
							StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
								StudyPlanId:        "study-plan-id",
								LearningMaterialId: "learning-material-id",
								StudentId:          &wrapperspb.StringValue{Value: "student-id"},
							},
							AvailableFrom: timestamppb.New(now),
							AvailableTo:   timestamppb.New(now),
							StartDate:     timestamppb.New(now),
							EndDate:       timestamppb.New(now),
							Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
							SchoolDate:    timestamppb.New(now),
						},
						LearningMaterialType: sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				studyPlanRepo.On("ListIndividualStudyPlanItems", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.IndividualStudyPlanItem{
					{
						StudyPlanID:        database.Text("study-plan-id"),
						LearningMaterialID: database.Text("learning-material-id"),
						StudentID:          database.Text("student-id"),
						AvailableFrom:      database.Timestamptz(now),
						AvailableTo:        database.Timestamptz(now),
						StartDate:          database.Timestamptz(now),
						EndDate:            database.Timestamptz(now),
						Status:             database.Text("STUDY_PLAN_ITEM_STATUS_ACTIVE"),
						SchoolDate:         database.Timestamptz(now),
						Score:              database.Int2(0),
						Type:               database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String()),
					},
				}, nil)
			},
		},
		{
			name:        "error empty student id",
			req:         &sspb.ListToDoItemRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "student id is required"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "error empty course ids",
			req: &sspb.ListToDoItemRequest{
				StudentId: "student-id",
			},
			expectedErr: status.Error(codes.InvalidArgument, "a list of course ids is required"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "error empty status",
			req: &sspb.ListToDoItemRequest{
				StudentId: "student-id",
				CourseIds: []string{"course-id-1", "course-id-2"},
			},
			expectedErr: fmt.Errorf("unknown todo status: STUDY_PLAN_ITEM_TO_DO_STATUS_NONE"),
			setup:       func(ctx context.Context) {},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			resp, err := svc.ListToDoItem(ctx, testCase.req.(*sspb.ListToDoItemRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})

	}
}

func Test_ListToDoItemStructuredBookTree(t *testing.T) {
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	now := time.Now().UTC()
	svc := &StudyPlanService{
		DB:            mockDB,
		StudyPlanRepo: studyPlanRepo,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &sspb.ListToDoItemStructuredBookTreeRequest{
				StudyPlanIdentity: &sspb.StudyPlanIdt{
					StudyPlanId: "study-plan-id",
					StudentId:   wrapperspb.String("student-id"),
				},
				Page: &cpb.Paging{
					Limit: 10,
				},
			},
			expectedResp: &sspb.ListToDoItemStructuredBookTreeResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString: "topic-id",
						},
					},
				},
				TodoItems: []*sspb.StudentStudyPlanItem{
					{
						LearningMaterial: &sspb.LearningMaterialBase{
							LearningMaterialId: "learning-material-id",
							TopicId:            "topic-id",
							DisplayOrder:       &wrapperspb.Int32Value{Value: 0},
						},
						StartDate:           timestamppb.New(now),
						EndDate:             timestamppb.New(now),
						CompletedAt:         timestamppb.New(now),
						StudyPlanItemStatus: sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
						SchoolDate:          timestamppb.New(now),
					},
				},
				TopicProgresses: []*sspb.StudentTopicStudyProgress{
					{
						TopicId:                "topic-id",
						CompletedStudyPlanItem: &wrapperspb.Int32Value{Value: 1},
						TotalStudyPlanItem:     &wrapperspb.Int32Value{Value: 1},
						AverageScore:           &wrapperspb.Int32Value{Value: 1},
						TopicName:              "topic-name",
					},
				},
			},
			setup: func(ctx context.Context) {
				studyPlanRepo.On("ListStudentToDoItem", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.StudentStudyPlanItem{
					{
						LearningMaterialID:  database.Text("learning-material-id"),
						TopicID:             database.Text("topic-id"),
						StartDate:           database.Timestamptz(now),
						EndDate:             database.Timestamptz(now),
						CompletedAt:         database.Timestamptz(now),
						StudyPlanItemStatus: database.Text("STUDY_PLAN_ITEM_STATUS_ACTIVE"),
						SchoolDate:          database.Timestamptz(now),
					},
				}, []*repositories.TopicProgress{
					{
						TopicID:         database.Text("topic-id"),
						CompletedSPItem: database.Int2(1),
						TotalSPItem:     database.Int2(1),
						AverageScore:    database.Int2(1),
						Name:            database.Text("topic-name"),
					},
				}, nil)
			},
		},
		{
			name: "error empty student id",
			req: &sspb.ListToDoItemStructuredBookTreeRequest{
				StudyPlanIdentity: &sspb.StudyPlanIdt{
					StudyPlanId: "study-plan-id",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "student id is required"),
			setup:       func(ctx context.Context) {},
		},

		{
			name: "error empty study plan id",
			req: &sspb.ListToDoItemStructuredBookTreeRequest{
				StudyPlanIdentity: &sspb.StudyPlanIdt{
					StudentId: wrapperspb.String("student-id"),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "study plan id is required"),
			setup:       func(ctx context.Context) {},
		},

		{
			name: "not found",
			req: &sspb.ListToDoItemStructuredBookTreeRequest{
				StudyPlanIdentity: &sspb.StudyPlanIdt{
					StudentId:   wrapperspb.String("id"),
					StudyPlanId: "id",
				},
			},
			expectedErr: fmt.Errorf("StudyPlanRepo.List: no rows in result set"),
			setup: func(ctx context.Context) {
				studyPlanRepo.On("ListStudentToDoItem", mock.Anything, mock.Anything, mock.Anything).
					Once().Return(nil, []*repositories.TopicProgress{}, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			resp, err := svc.ListToDoItemStructuredBookTree(ctx, testCase.req.(*sspb.ListToDoItemStructuredBookTreeRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})

	}
}

func TestStudyPlanService_RetrieveAllocateMarker(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	allocateMarkerRepo := &mock_repositories.MockAllocateMarkerRepo{}
	svc := &StudyPlanService{
		DB:                 mockDB,
		AllocateMarkerRepo: allocateMarkerRepo,
	}

	studyPlanItemIdentity := &sspb.StudyPlanItemIdentity{
		LearningMaterialId: "LM-ID",
		StudyPlanId:        "SP__ID",
		StudentId:          wrapperspb.String("STUDENT-ID"),
	}

	testCases := []TestCase{
		{
			name: "Happy case return teacherID from GetTeacherID",
			setup: func(ctx context.Context) {
				args := append([]interface{}{
					mock.Anything,
					mock.Anything,
					&repositories.StudyPlanItemIdentity{
						LearningMaterialID: database.Text(studyPlanItemIdentity.LearningMaterialId),
						StudyPlanID:        database.Text(studyPlanItemIdentity.StudyPlanId),
						StudentID:          database.Text(studyPlanItemIdentity.StudentId.GetValue()),
					},
				})
				allocateMarkerRepo.On("GetTeacherID", args...).Once().Return(database.Text("Teacher-ID"), nil)
			},
			req: &sspb.RetrieveAllocateMarkerRequest{
				StudyPlanItemIdentity: studyPlanItemIdentity,
			},
			expectedResp: &sspb.RetrieveAllocateMarkerResponse{
				MarkerId: "Teacher-ID",
			},
		},
		{
			name: "Error when GetTeacherID err",
			setup: func(ctx context.Context) {
				allocateMarkerRepo.On("GetTeacherID",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(database.Text(""), pgx.ErrNoRows)
			},
			req: &sspb.RetrieveAllocateMarkerRequest{
				StudyPlanItemIdentity: studyPlanItemIdentity,
			},
			expectedResp: &sspb.RetrieveAllocateMarkerResponse{
				MarkerId: "",
			},
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()

		tc.setup(ctx)

		req := tc.req.(*sspb.RetrieveAllocateMarkerRequest)

		resp, err := svc.RetrieveAllocateMarker(ctx, req)

		if tc.expectedErr != nil {
			assert.Nil(t, resp)
			assert.Error(t, err)
			assert.Equal(t, err.Error(), tc.expectedErr.Error())
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, resp, tc.expectedResp)
	}
}

func TestStudyPlanService_RetrieveAllocateMarkerInvalidRequest(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	allocateMarkerRepo := &mock_repositories.MockAllocateMarkerRepo{}
	svc := &StudyPlanService{
		DB:                 mockDB,
		AllocateMarkerRepo: allocateMarkerRepo,
	}

	testCases := []TestCase{
		{
			name:  "Missing LearningMaterialId",
			setup: func(ctx context.Context) {},
			req: &sspb.RetrieveAllocateMarkerRequest{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId: "SP__ID",
					StudentId:   wrapperspb.String("STUDENT-ID"),
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateRetrieveAllocateMarkerRequest: %s", "learning_material_id must not be empty"),
		},
		{
			name:  "Missing StudyPlanId",
			setup: func(ctx context.Context) {},
			req: &sspb.RetrieveAllocateMarkerRequest{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					LearningMaterialId: "LM-ID",
					StudentId:          wrapperspb.String("STUDENT-ID"),
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateRetrieveAllocateMarkerRequest: %s", "study_plan_id must not be empty"),
		},
		{
			name:  "Missing StudentId",
			setup: func(ctx context.Context) {},
			req: &sspb.RetrieveAllocateMarkerRequest{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        "SP__ID",
					LearningMaterialId: "LM-ID",
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "validateRetrieveAllocateMarkerRequest: %s", "student_id must not be empty"),
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()

		tc.setup(ctx)

		req := tc.req.(*sspb.RetrieveAllocateMarkerRequest)

		resp, err := svc.RetrieveAllocateMarker(ctx, req)

		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), tc.expectedErr.Error())
		allocateMarkerRepo.AssertNotCalled(t, "GetTeacherID")
	}
}
