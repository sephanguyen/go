package services

import (
	"context"
	"fmt"
	"strings"
	"testing"

	entities "github.com/manabie-com/backend/internal/eureka/entities/learning_history_data_sync"
	repositories "github.com/manabie-com/backend/internal/eureka/repositories/learning_history_data_sync"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories/learning_history_data_sync"
	mock_services "github.com/manabie-com/backend/mock/eureka/services/learning_history_data_sync"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	ctx          context.Context
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestDownloadMappingFile(t *testing.T) {
	t.Parallel()

	lhdsRepo := new(mock_repositories.MockLearningHistoryDataSyncRepo)
	yasuoClient := new(mock_services.YasuoUploadModifierService)
	ctx := interceptors.NewIncomingContext(context.Background())

	s := LearningHistoryDataSyncService{
		LearningHistoryDataSyncRepo: lhdsRepo,
		YasuoUploadModifierService:  yasuoClient,
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         &sspb.DownloadMappingFileRequest{},
			expectedErr: nil,
			expectedResp: &sspb.DownloadMappingFileResponse{
				MappingCourseIdUrl:           "https://storage.googleapis.com/learning-history-data-sync/mapping_course_id.csv",
				MappingExamLoIdUrl:           "https://storage.googleapis.com/learning-history-data-sync/mapping_exam_lo_id.csv",
				MappingQuestionTagUrl:        "https://storage.googleapis.com/learning-history-data-sync/mapping_question_tag.csv",
				FailedSyncEmailRecipientsUrl: "https://storage.googleapis.com/learning-history-data-sync/failed_sync_email_recipient.csv",
			},
			setup: func(ctx context.Context) {
				lhdsRepo.On("RetrieveMappingCourseID", mock.Anything, mock.Anything).Return([]*entities.MappingCourseID{}, nil)
				lhdsRepo.On("RetrieveMappingExamLoID", mock.Anything, mock.Anything).Return([]*entities.MappingExamLoID{}, nil)
				lhdsRepo.On("RetrieveMappingQuestionTag", mock.Anything, mock.Anything).Return([]*entities.MappingQuestionTag{}, nil)
				lhdsRepo.On("RetrieveFailedSyncEmailRecipient", mock.Anything, mock.Anything).Return([]*entities.FailedSyncEmailRecipient{}, nil)
				yasuoClient.On("BulkUploadFile", mock.Anything, mock.Anything).Return(&ypb.BulkUploadFileResponse{
					Files: []*ypb.BulkUploadFileResponse_File{
						{
							FileName: "mapping_course_id.csv",
							Url:      "https://storage.googleapis.com/learning-history-data-sync/mapping_course_id.csv",
						},
						{
							FileName: "mapping_exam_lo_id.csv",
							Url:      "https://storage.googleapis.com/learning-history-data-sync/mapping_exam_lo_id.csv",
						},
						{
							FileName: "mapping_question_tag.csv",
							Url:      "https://storage.googleapis.com/learning-history-data-sync/mapping_question_tag.csv",
						},
						{
							FileName: "failed_sync_email_recipient.csv",
							Url:      "https://storage.googleapis.com/learning-history-data-sync/failed_sync_email_recipient.csv",
						},
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := s.DownloadMappingFile(ctx, testCase.req.(*sspb.DownloadMappingFileRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			}
		})

	}
}

func TestUploadMappingFile(t *testing.T) {
	t.Parallel()

	lhdsRepo := new(mock_repositories.MockLearningHistoryDataSyncRepo)
	ctx := interceptors.NewIncomingContext(context.Background())

	s := LearningHistoryDataSyncService{
		LearningHistoryDataSyncRepo: lhdsRepo,
	}

	headerMappingCourseID := "manabie_course_id,withus_course_id,is_archived"
	row1MappingCourseID := "test_manabie_course_id_1,test_withus_course_id_1,false"
	row2MappingCourseID := "test_manabie_course_id_2,test_withus_course_id_2,false"

	headerMappingExamLoID := "exam_lo_id,material_code,is_archived"
	row1MappingExamLoID := "test_exam_lo_id_1,test_material_code_1,false"
	row2MappingExamLoID := "test_exam_lo_id_2,test_material_code_2,false"

	headerMappingQuestionTag := "manabie_tag_id,manabie_tag_name,withus_tag_name,is_archived"
	row1MappingQuestionTag := "test_manabie_tag_id_1,test_manabie_tag_name_1,withus_tag_name_1,false"
	row2MappingQuestionTag := "test_manabie_tag_id_2,test_manabie_tag_name_2,withus_tag_name_2,false"

	headerFailedSync := "recipient_id,email_address,is_archived"
	row1FailedSync := "test_recipient_id_1,test_email_address_1,false"
	row2FailedSync := "test_recipient_id_2,test_email_address_2,false"

	req := &sspb.UploadMappingFileRequest{
		MappingCourseId: []byte(fmt.Sprintf(`%s
		%s
		%s`, headerMappingCourseID, row1MappingCourseID, row2MappingCourseID)),
		MappingExamLoId: []byte(fmt.Sprintf(`%s
		%s
		%s`, headerMappingExamLoID, row1MappingExamLoID, row2MappingExamLoID)),
		MappingQuestionTag: []byte(fmt.Sprintf(`%s
		%s
		%s`, headerMappingQuestionTag, row1MappingQuestionTag, row2MappingQuestionTag)),
		FailedSyncEmailRecipients: []byte(fmt.Sprintf(`%s
		%s
		%s`, headerFailedSync, row1FailedSync, row2FailedSync)),
	}

	testCases := []TestCase{
		{
			name:         "happy case",
			req:          req,
			expectedErr:  nil,
			expectedResp: &sspb.UploadMappingFileResponse{},
			setup: func(ctx context.Context) {
				lhdsRepo.On("BulkUpsertMappingCourseID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				lhdsRepo.On("BulkUpsertMappingExamLoID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				lhdsRepo.On("BulkUpsertMappingQuestionTag", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				lhdsRepo.On("BulkUpsertFailedSyncEmailRecipient", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			resp, err := s.UploadMappingFile(ctx, testCase.req.(*sspb.UploadMappingFileRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
			}
		})

	}
}

func TestExportLearningHistoryData(t *testing.T) {
	t.Parallel()

	lhdsRepo := new(mock_repositories.MockLearningHistoryDataSyncRepo)
	ctx := interceptors.NewIncomingContext(context.Background())

	s := LearningHistoryDataSyncService{
		LearningHistoryDataSyncRepo: lhdsRepo,
	}

	withusRows := []*repositories.WithusDataRow{
		{
			CustomerNumber:    database.Text("2210017"),
			StudentNumber:     database.Text("2210017"),
			MaterialCode:      database.Text("2322013100"),
			PaperCount:        database.Text("2"),
			Score:             database.Int4(999),
			DateSubmitted:     database.Text(""),
			ApproverID:        pgtype.Text{},
			PaperApprovalDate: pgtype.Text{},
			PerspectiveScore:  database.JSONB([]*PerspectiveScore{}),
			IsResubmission:    database.Bool(false),
		},
		{
			CustomerNumber:    database.Text("2210018"),
			StudentNumber:     database.Text("2210018"),
			MaterialCode:      database.Text("2322013100"),
			PaperCount:        database.Text("5"),
			Score:             database.Int4(100),
			DateSubmitted:     database.Text("2023/3/22"),
			ApproverID:        database.Text("01GPMT30ZQGVMZWES75KEAVC65"),
			PaperApprovalDate: database.Text("2023/3/25"),
			PerspectiveScore: database.JSONB([]*PerspectiveScore{
				{TagID: "01GNEKW73ZVJ0KNA2ZKB9J02X1", Score: ":20/20"},
				{TagID: "01GNEKW73ZVJ0KNA2ZKC335BS1", Score: ":20/20"},
				{TagID: "01GNEKW73ZVJ0KNA2ZKF5PZ12C", Score: ":60/60"},
			}),
			IsResubmission: database.Bool(false),
		},
		{
			CustomerNumber:    database.Text("2210019"),
			StudentNumber:     database.Text("2210019"),
			MaterialCode:      database.Text("2322013100"),
			PaperCount:        database.Text("8"),
			Score:             database.Int4(30),
			DateSubmitted:     database.Text("2023/3/22"),
			ApproverID:        database.Text("01GPMT30ZQGVMZWES75KEAVC65"),
			PaperApprovalDate: database.Text("2023/3/25"),
			PerspectiveScore: database.JSONB([]*PerspectiveScore{
				{TagID: "01GNEKW73ZVJ0KNA2ZKB9J02X1", Score: ":/20"},
				{TagID: "01GNEKW73ZVJ0KNA2ZKC335BS1", Score: ":/50"},
			}),
			IsResubmission: database.Bool(true),
		},
	}
	questionTags := []*entities.MappingQuestionTag{
		{ManabieTagID: database.Text("01GNEKW73ZVJ0KNA2ZKB9J02X1"), WithusTagName: database.Text("知")},
		{ManabieTagID: database.Text("01GNEKW73ZVJ0KNA2ZKC335BS1"), WithusTagName: database.Text("考")},
		{ManabieTagID: database.Text("01GNEKW73ZVJ0KNA2ZKF5PZ12C"), WithusTagName: database.Text("主")},
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				lhdsRepo.On("RetrieveWithusData", mock.Anything, mock.Anything).Return(withusRows, nil)
				lhdsRepo.On("RetrieveMappingQuestionTag", mock.Anything, mock.Anything).Return(questionTags, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			filename1, filename2, data, err := s.ExportLearningHistoryData(ctx)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, data)
				assert.NotEmpty(t, filename1)
				assert.NotEmpty(t, filename2)
			}
		})
	}
}

func Test_formatPerspectiveScores(t *testing.T) {

	questionTags := []*entities.MappingQuestionTag{
		{ManabieTagID: database.Text("01GNEKW73ZVJ0KNA2ZKB9J02X1"), WithusTagName: database.Text("知")},
		{ManabieTagID: database.Text("01GNEKW73ZVJ0KNA2ZKC335BS1"), WithusTagName: database.Text("考")},
		{ManabieTagID: database.Text("01GNEKW73ZVJ0KNA2ZKF5PZ12C"), WithusTagName: database.Text("主")},
	}

	{
		// Full case
		perspectiveScore := []*PerspectiveScore{
			{TagID: "01GNEKW73ZVJ0KNA2ZKB9J02X1", Score: ":20/20"},
			{TagID: "01GNEKW73ZVJ0KNA2ZKC335BS1", Score: ":20/20"},
			{TagID: "01GNEKW73ZVJ0KNA2ZKF5PZ12C", Score: ":60/60"},
		}

		out := formatPerspectiveScores(questionTags, perspectiveScore, false)

		assert.Equal(t, strings.Join(out, "$"), "知:20/20$考:20/20$主:60/60")
	}

	{
		// Missing case
		perspectiveScore := []*PerspectiveScore{
			{TagID: "01GNEKW73ZVJ0KNA2ZKB9J02X1", Score: ":20/20"},
			{TagID: "01GNEKW73ZVJ0KNA2ZKF5PZ12C", Score: ":60/60"},
		}

		out := formatPerspectiveScores(questionTags, perspectiveScore, false)

		assert.Equal(t, strings.Join(out, "$"), "知:20/20$考:0/0$主:60/60")
	}
}
