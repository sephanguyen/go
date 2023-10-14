package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	entities "github.com/manabie-com/backend/internal/eureka/entities/learning_history_data_sync"
	repo "github.com/manabie-com/backend/internal/eureka/repositories/learning_history_data_sync"
	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/scanner"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileType int

// nolint
const (
	FileType_MAPPING_COURSE_ID           = "mapping_course_id"
	FileType_MAPPING_EXAM_LO_ID          = "mapping_exam_lo_id"
	FileType_MAPPING_QUESTION_TAG        = "mapping_question_tag"
	FileType_FAILED_SYNC_EMAIL_RECIPIENT = "failed_sync_email_recipient"
	TimeFormat                           = "2006/01/02 15:04:05"
	ARCHIVED_TEXT                        = "is_archived"
)

type WithusFileURL struct {
	MappingCourseID          string
	MappingExamLoID          string
	MappingQuestionTag       string
	FailedSyncEmailRecipient string
}

type PerspectiveScore struct {
	TagID string `json:"tag_id"`
	Score string `json:"score"`
}

type YasuoUploadModifierService interface {
	BulkUploadFile(ctx context.Context, req *ypb.BulkUploadFileRequest, opts ...grpc.CallOption) (*ypb.BulkUploadFileResponse, error)
}

type LearningHistoryDataSyncService struct {
	DBTrace database.Ext

	Alert alert.SlackFactory

	YasuoUploadModifierService YasuoUploadModifierService

	LearningHistoryDataSyncRepo interface {
		RetrieveMappingCourseID(ctx context.Context, db database.QueryExecer) ([]*entities.MappingCourseID, error)
		RetrieveMappingExamLoID(ctx context.Context, db database.QueryExecer) ([]*entities.MappingExamLoID, error)
		RetrieveMappingQuestionTag(ctx context.Context, db database.QueryExecer) ([]*entities.MappingQuestionTag, error)
		RetrieveFailedSyncEmailRecipient(ctx context.Context, db database.QueryExecer) ([]*entities.FailedSyncEmailRecipient, error)
		BulkUpsertMappingCourseID(ctx context.Context, db database.QueryExecer, items []*entities.MappingCourseID) error
		BulkUpsertMappingExamLoID(ctx context.Context, db database.QueryExecer, items []*entities.MappingExamLoID) error
		BulkUpsertMappingQuestionTag(ctx context.Context, db database.QueryExecer, items []*entities.MappingQuestionTag) error
		BulkUpsertFailedSyncEmailRecipient(ctx context.Context, db database.QueryExecer, items []*entities.FailedSyncEmailRecipient) error
		RetrieveWithusData(ctx context.Context, db database.QueryExecer) ([]*repo.WithusDataRow, error)
	}
}

func NewLearningHistoryDataSyncService(db database.Ext, yasuoSvc ypb.UploadModifierServiceClient) sspb.LearningHistoryDataSyncServiceServer {
	return &LearningHistoryDataSyncService{
		DBTrace:                     db,
		LearningHistoryDataSyncRepo: &repo.LearningHistoryDataSyncRepo{},
		YasuoUploadModifierService:  yasuoSvc,
	}
}

func (s *LearningHistoryDataSyncService) ExportLearningHistoryData(ctx context.Context) (filename1 string, filename2 string, payload []byte, err error) {
	learningHistoryData, err := s.LearningHistoryDataSyncRepo.RetrieveWithusData(ctx, s.DBTrace)
	if err != nil {
		return "", "", nil, status.Errorf(codes.Internal, "failed to retrieve learning history data: %v", err)
	}

	tags, err := s.LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag(ctx, s.DBTrace)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil, status.Errorf(codes.Internal, "failed to RetrieveMappingQuestionTag: %v", err)
	}

	csvData := make([][]string, 0, len(learningHistoryData))
	csvData = append(csvData, []string{
		`"顧客番号"`,
		`"生徒番号"`,
		`"教材コード"`,
		`"レポート回"`,
		`"得点"`,
		`"提出日"`,
		`"承認者のログインID"`,
		`"レポート承認日"`,
		`"観点別得点"`,
	})

	for _, row := range learningHistoryData {
		perspectiveScores := make([]*PerspectiveScore, 0)

		// Parse PerspectiveScore
		if ps := row.PerspectiveScore; ps.Status == pgtype.Present {
			if err := ps.AssignTo(&perspectiveScores); err != nil {
				return "", "", nil, err
			}
		}

		perspectives := formatPerspectiveScores(tags, perspectiveScores, row.IsResubmission.Bool)

		csvData = append(csvData, []string{
			formatDoubleQuote(row.CustomerNumber.String),
			formatDoubleQuote(row.StudentNumber.String),
			formatDoubleQuote(row.MaterialCode.String),
			formatDoubleQuote(row.PaperCount.String),
			formatDoubleQuote(strconv.Itoa(int(row.Score.Int))),
			formatDoubleQuote(row.DateSubmitted.String),
			formatDoubleQuote(row.ApproverID.String),
			formatDoubleQuote(row.PaperApprovalDate.String),
			formatDoubleQuote(strings.Join(perspectives, "$")),
		})
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(transform.NewWriter(&buffer, japanese.ShiftJIS.NewEncoder()))
	err = writer.WriteAll(csvData)
	if err != nil {
		return "", "", nil, status.Errorf(codes.Internal, "failed to write csv: %v", err)
	}

	writer.Flush()

	if writer.Error() != nil {
		return "", "", nil, status.Errorf(codes.Internal, "failed to write csv: %v", err)
	}

	tokyoLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", "", nil, status.Errorf(codes.Internal, "failed LoadLocation time: %v", err)
	}

	now := time.Now().In(tokyoLocation)

	filename1 = fmt.Sprintf("REPORTS_L6_%s.csv", now.Format("20060102"))
	filename2 = fmt.Sprintf("REPORTS_M1_%s.csv", now.Format("20060102"))

	raws := bytes.ReplaceAll(buffer.Bytes(), []byte(`"""`), []byte(`"`))

	return filename1, filename2, raws, nil
}

func formatDoubleQuote(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}

func formatPerspectiveScores(tags []*entities.MappingQuestionTag, perspectiveScores []*PerspectiveScore, isResubmission bool) (perspectives []string) {
	mapPerspectiveScores := make(map[string]string, 0)

	for _, ps := range perspectiveScores {
		mapPerspectiveScores[ps.TagID] = ps.Score
	}

	// No question tag
	if len(mapPerspectiveScores) == 0 {
		return perspectives
	}

	for _, tg := range tags {
		if tg.IsArchived.Bool || tg.WithusTagName.String == "" {
			continue
		}

		score := ":0/0"

		if isResubmission {
			score = ":/0"
		}
		if sc, ok := mapPerspectiveScores[tg.ManabieTagID.String]; ok {
			score = sc
		}

		perspectives = append(perspectives, fmt.Sprintf("%s%s", tg.WithusTagName.String, score))
	}

	return perspectives
}

func (s *LearningHistoryDataSyncService) DownloadMappingFile(ctx context.Context, req *sspb.DownloadMappingFileRequest) (*sspb.DownloadMappingFileResponse, error) {
	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}
	mappingCourseIDFile, err := s.retrieveMappingCourseIDToFile(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "err retrieveMappingCourseIDToFile: %v", err.Error())
	}
	mappingExamLOIDFile, err := s.retrieveMappingExamLoIDToFile(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "err retrieveMappingExamLoIDToFile: %v", err.Error())
	}
	mappingQuestionTagFile, err := s.retrieveMappingQuestionTagToFile(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "err retrieveMappingQuestionTagToFile: %v", err.Error())
	}
	failedSyncEmailRecipientFile, err := s.retrieveFailedSyncEmailRecipientToFile(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "err retrieveFailedSyncEmailRecipientToFile: %v", err.Error())
	}
	bulkUploadFileReq := &ypb.BulkUploadFileRequest{
		Files: []*ypb.BulkUploadFileRequest_File{
			mappingCourseIDFile,
			mappingExamLOIDFile,
			mappingQuestionTagFile,
			failedSyncEmailRecipientFile,
		},
	}
	bulkUploadFileResp, err := s.YasuoUploadModifierService.BulkUploadFile(mdCtx, bulkUploadFileReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "err BulkUploadFile: %v", err.Error())
	}
	res := &sspb.DownloadMappingFileResponse{}
	for _, v := range bulkUploadFileResp.Files {
		if v.FileName == mappingCourseIDFile.FileName {
			res.MappingCourseIdUrl = v.Url
		}
		if v.FileName == mappingExamLOIDFile.FileName {
			res.MappingExamLoIdUrl = v.Url
		}
		if v.FileName == mappingQuestionTagFile.FileName {
			res.MappingQuestionTagUrl = v.Url
		}
		if v.FileName == failedSyncEmailRecipientFile.FileName {
			res.FailedSyncEmailRecipientsUrl = v.Url
		}
	}
	return res, nil
}

func (s *LearningHistoryDataSyncService) UploadMappingFile(ctx context.Context, req *sspb.UploadMappingFileRequest) (*sspb.UploadMappingFileResponse, error) {
	err := s.importMappingCourseID(ctx, req.MappingCourseId)
	if err != nil {
		return nil, err
	}
	err = s.importMappingExamLoID(ctx, req.MappingExamLoId)
	if err != nil {
		return nil, err
	}
	err = s.importMappingQuestionTag(ctx, req.MappingQuestionTag)
	if err != nil {
		return nil, err
	}
	err = s.importFailedSyncEmailRecipient(ctx, req.FailedSyncEmailRecipients)
	if err != nil {
		return nil, err
	}
	return &sspb.UploadMappingFileResponse{}, nil
}

func (s *LearningHistoryDataSyncService) retrieveMappingCourseIDToFile(ctx context.Context) (*ypb.BulkUploadFileRequest_File, error) {
	mappingCourseID, err := s.LearningHistoryDataSyncRepo.RetrieveMappingCourseID(ctx, s.DBTrace)
	if err != nil {
		return nil, fmt.Errorf("err RetrieveMappingCourseID: %w", err)
	}
	mappingCourseIDDatas, err := generateCSVDataFromMappingCourseID(mappingCourseID)
	if err != nil {
		return nil, fmt.Errorf("err generateCSVDataFromMappingCourseID: %w", err)
	}

	files := &ypb.BulkUploadFileRequest_File{
		FileName:    fmt.Sprintf("%s_%s.csv", FileType_MAPPING_COURSE_ID, time.Now().Format("20060102150405")),
		Payload:     mappingCourseIDDatas,
		ContentType: "text/csv",
	}
	return files, nil
}

func (s *LearningHistoryDataSyncService) retrieveMappingExamLoIDToFile(ctx context.Context) (*ypb.BulkUploadFileRequest_File, error) {
	mappingExamLoID, err := s.LearningHistoryDataSyncRepo.RetrieveMappingExamLoID(ctx, s.DBTrace)
	if err != nil {
		return nil, fmt.Errorf("err RetrieveMappingExamLoID: %w", err)
	}
	mappingExamLoIDDatas, err := generateCSVDataFromMappingExamLoID(mappingExamLoID)
	if err != nil {
		return nil, fmt.Errorf("err generateCSVDataFromMappingExamLoID: %w", err)
	}

	files := &ypb.BulkUploadFileRequest_File{
		FileName:    fmt.Sprintf("%s_%s.csv", FileType_MAPPING_EXAM_LO_ID, time.Now().Format("20060102150405")),
		Payload:     mappingExamLoIDDatas,
		ContentType: "text/csv",
	}
	return files, nil
}

func (s *LearningHistoryDataSyncService) retrieveMappingQuestionTagToFile(ctx context.Context) (*ypb.BulkUploadFileRequest_File, error) {
	mappingQuestionTag, err := s.LearningHistoryDataSyncRepo.RetrieveMappingQuestionTag(ctx, s.DBTrace)
	if err != nil {
		return nil, fmt.Errorf("err RetrieveMappingQuestionTag: %w", err)
	}
	mappingQuestionTagDatas, err := generateCSVDataFromMappingQuestionTag(mappingQuestionTag)
	if err != nil {
		return nil, fmt.Errorf("err generateCSVDataFromMappingQuestionTag: %w", err)
	}

	files := &ypb.BulkUploadFileRequest_File{
		FileName:    fmt.Sprintf("%s_%s.csv", FileType_MAPPING_QUESTION_TAG, time.Now().Format("20060102150405")),
		Payload:     mappingQuestionTagDatas,
		ContentType: "text/csv",
	}
	return files, nil
}

func (s *LearningHistoryDataSyncService) retrieveFailedSyncEmailRecipientToFile(ctx context.Context) (*ypb.BulkUploadFileRequest_File, error) {
	failedSyncEmailRecipient, err := s.LearningHistoryDataSyncRepo.RetrieveFailedSyncEmailRecipient(ctx, s.DBTrace)
	if err != nil {
		return nil, fmt.Errorf("err RetrieveFailedSyncEmailRecipient: %w", err)
	}
	failedSyncEmailRecipientDatas, err := generateCSVDataFromFailedSyncEmailRecipient(failedSyncEmailRecipient)
	if err != nil {
		return nil, fmt.Errorf("err generateCSVDataFromFailedSyncEmailRecipient: %w", err)
	}

	files := &ypb.BulkUploadFileRequest_File{
		FileName:    fmt.Sprintf("%s_%s.csv", FileType_FAILED_SYNC_EMAIL_RECIPIENT, time.Now().Format("20060102150405")),
		Payload:     failedSyncEmailRecipientDatas,
		ContentType: "text/csv",
	}
	return files, nil
}

func generateCSVDataFromMappingCourseID(mappingCourseID []*entities.MappingCourseID) ([]byte, error) {
	csvData := make([][]string, 0, len(mappingCourseID))
	// header
	csvData = append(csvData, []string{
		"manabie_course_id",
		"withus_course_id",
		"last_updated_date",
		"last_updated_by",
		"is_archived",
	})
	for _, v := range mappingCourseID {
		csvData = append(csvData, []string{
			v.ManabieCourseID.String,
			v.WithusCourseID.String,
			v.LastUpdatedDate.Time.Format(TimeFormat),
			v.LastUpdatedBy.String,
			strconv.FormatBool(v.IsArchived.Bool),
		})
	}
	return genCSVBytesFromData(csvData)
}

func generateCSVDataFromMappingExamLoID(mappingExamLoID []*entities.MappingExamLoID) ([]byte, error) {
	csvData := make([][]string, 0, len(mappingExamLoID))
	// header
	csvData = append(csvData, []string{
		"exam_lo_id",
		"material_code",
		"last_updated_date",
		"last_updated_by",
		"is_archived",
	})
	for _, v := range mappingExamLoID {
		csvData = append(csvData, []string{
			v.ExamLoID.String,
			v.MaterialCode.String,
			v.LastUpdatedDate.Time.Format(TimeFormat),
			v.LastUpdatedBy.String,
			strconv.FormatBool(v.IsArchived.Bool),
		})
	}

	return genCSVBytesFromData(csvData)
}

func generateCSVDataFromMappingQuestionTag(mappingQuestionTag []*entities.MappingQuestionTag) ([]byte, error) {
	csvData := make([][]string, 0, len(mappingQuestionTag))
	// header
	csvData = append(csvData, []string{
		"manabie_tag_id",
		"manabie_tag_name",
		"withus_tag_name",
		"last_updated_date",
		"last_updated_by",
		"is_archived",
	})
	for _, v := range mappingQuestionTag {
		csvData = append(csvData, []string{
			v.ManabieTagID.String,
			v.ManabieTagName.String,
			v.WithusTagName.String,
			v.LastUpdatedDate.Time.Format(TimeFormat),
			v.LastUpdatedBy.String,
			strconv.FormatBool(v.IsArchived.Bool),
		})
	}

	return genCSVBytesFromData(csvData)
}

func generateCSVDataFromFailedSyncEmailRecipient(failedSyncEmailRecipient []*entities.FailedSyncEmailRecipient) ([]byte, error) {
	csvData := make([][]string, 0, len(failedSyncEmailRecipient))
	// header
	csvData = append(csvData, []string{
		"recipient_id",
		"email_address",
		"last_updated_date",
		"last_updated_by",
		"is_archived",
	})
	for _, v := range failedSyncEmailRecipient {
		csvData = append(csvData, []string{
			v.RecipientID.String,
			v.EmailAddress.String,
			v.LastUpdatedDate.Time.Format(TimeFormat),
			v.LastUpdatedBy.String,
			strconv.FormatBool(v.IsArchived.Bool),
		})
	}
	return genCSVBytesFromData(csvData)
}

func genCSVBytesFromData(csvData [][]string) ([]byte, error) {
	// write the CSV data to buffer
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	err := writer.WriteAll(csvData)
	if err != nil {
		return nil, fmt.Errorf("writer.WriteAll err: %w", writer.Error())
	}

	writer.Flush()

	if writer.Error() != nil {
		return nil, fmt.Errorf("error on writing CSV err: %w", writer.Error())
	}

	return buffer.Bytes(), nil
}

func checkDuplicateData(data []string) error {
	// check duplicate data
	seen := make(map[string]bool)
	for _, v := range data {
		if seen[v] {
			return fmt.Errorf("duplicate data: %s", v)
		}
		seen[v] = true
	}
	return nil
}

func (s *LearningHistoryDataSyncService) importMappingCourseID(ctx context.Context, payload []byte) error {
	sc := scanner.NewCSVScanner(bytes.NewReader(payload))
	// no columns
	if len(sc.GetRow()) == 0 {
		return nil
	}
	err := validateCSVMappingCourseIDFormat(sc)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "validateCSVMappingCourseIDFormat: %v", err)
	}

	mappingCourseIDs := make([]*entities.MappingCourseID, 0)
	for sc.Scan() {
		manabieCourseID := sc.Text("manabie_course_id")
		withusCourseID := sc.Text("withus_course_id")
		lastUpdatedDate := time.Now()
		lastUpdatedBy := interceptors.UserIDFromContext(ctx)
		isArchived := sc.Text("is_archived")

		row, err := newMappingCourseIDRow(manabieCourseID, withusCourseID, lastUpdatedDate, lastUpdatedBy, isArchived)
		if err != nil {
			line := sc.GetCurRow()
			return status.Errorf(codes.InvalidArgument, "newMappingCourseIDRow: %v, line: %d", err, line)
		}
		mappingCourseID, err := row.toEntity()
		if err != nil {
			line := sc.GetCurRow()
			return status.Errorf(codes.InvalidArgument, "toEntity: %v, line: %d", err, line)
		}
		mappingCourseIDs = append(mappingCourseIDs, mappingCourseID)
	}

	// if no rows
	if len(mappingCourseIDs) == 0 {
		return status.Errorf(codes.InvalidArgument, "no data in mapping course id csv file")
	}
	manabieCourseIDTexts := make([]string, 0, len(mappingCourseIDs))
	for _, v := range mappingCourseIDs {
		manabieCourseIDTexts = append(manabieCourseIDTexts, v.ManabieCourseID.String)
	}
	err = checkDuplicateData(manabieCourseIDTexts)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	err = s.LearningHistoryDataSyncRepo.BulkUpsertMappingCourseID(ctx, s.DBTrace, mappingCourseIDs)
	if err != nil {
		return status.Errorf(codes.Internal, "BulkUpsertMappingCourseID: %v", err)
	}
	return nil
}

func (s *LearningHistoryDataSyncService) importMappingExamLoID(ctx context.Context, payload []byte) error {
	sc := scanner.NewCSVScanner(bytes.NewReader(payload))
	// no columns
	if len(sc.GetRow()) == 0 {
		return nil
	}
	err := validateCSVMappingExamLoIDFormat(sc)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "validateCSVMappingExamLoIDFormat: %v", err)
	}

	mappingExamLoIDs := make([]*entities.MappingExamLoID, 0)
	for sc.Scan() {
		manabieExamLoID := sc.Text("exam_lo_id")
		withusExamLoID := sc.Text("material_code")
		lastUpdatedDate := time.Now()
		lastUpdatedBy := interceptors.UserIDFromContext(ctx)
		isArchived := sc.Text("is_archived")

		row, err := newMappingExamLoIDRow(manabieExamLoID, withusExamLoID, lastUpdatedDate, lastUpdatedBy, isArchived)
		if err != nil {
			line := sc.GetCurRow()
			return status.Errorf(codes.InvalidArgument, "newMappingExamLoIDRow: %v, line: %d", err, line)
		}
		mappingExamLoID, err := row.toEntity()
		if err != nil {
			line := sc.GetCurRow()
			return status.Errorf(codes.InvalidArgument, "toEntity: %v, line: %d", err, line)
		}
		mappingExamLoIDs = append(mappingExamLoIDs, mappingExamLoID)
	}

	// if no rows
	if len(mappingExamLoIDs) == 0 {
		return status.Errorf(codes.InvalidArgument, "no data in mapping exam lo csv file")
	}
	manabieExamLoIDTexts := make([]string, 0, len(mappingExamLoIDs))
	for _, v := range mappingExamLoIDs {
		manabieExamLoIDTexts = append(manabieExamLoIDTexts, v.ExamLoID.String)
	}
	err = checkDuplicateData(manabieExamLoIDTexts)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	err = s.LearningHistoryDataSyncRepo.BulkUpsertMappingExamLoID(ctx, s.DBTrace, mappingExamLoIDs)
	if err != nil {
		return status.Errorf(codes.Internal, "BulkUpsertMappingExamLoID: %v", err)
	}
	return nil
}

func (s *LearningHistoryDataSyncService) importFailedSyncEmailRecipient(ctx context.Context, payload []byte) error {
	sc := scanner.NewCSVScanner(bytes.NewReader(payload))
	// no columns
	if len(sc.GetRow()) == 0 {
		return nil
	}
	err := validateCSVFailedSyncEmailRecipientFormat(sc)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "validateCSVFailedSyncEmailRecipientFormat: %v", err)
	}

	failedSyncEmailRecipients := make([]*entities.FailedSyncEmailRecipient, 0)
	for sc.Scan() {
		recipientID := sc.Text("recipient_id")
		emailAddress := sc.Text("email_address")
		lastUpdatedDate := time.Now()
		lastUpdatedBy := interceptors.UserIDFromContext(ctx)
		isArchived := sc.Text("is_archived")

		row, err := newFailedSyncEmailRecipientRow(recipientID, emailAddress, lastUpdatedDate, lastUpdatedBy, isArchived)
		if err != nil {
			line := sc.GetCurRow()
			return status.Errorf(codes.InvalidArgument, "newFailedSyncEmailRecipientRow: %v, line: %d", err, line)
		}
		failedSyncEmailRecipient, err := row.toEntity()
		if err != nil {
			line := sc.GetCurRow()
			return status.Errorf(codes.InvalidArgument, "toEntity: %v, line: %d", err, line)
		}
		failedSyncEmailRecipients = append(failedSyncEmailRecipients, failedSyncEmailRecipient)
	}

	// if no rows
	if len(failedSyncEmailRecipients) == 0 {
		return nil
	}
	err = s.LearningHistoryDataSyncRepo.BulkUpsertFailedSyncEmailRecipient(ctx, s.DBTrace, failedSyncEmailRecipients)
	if err != nil {
		return status.Errorf(codes.Internal, "BulkUpsertFailedSyncEmailRecipient: %v", err)
	}
	return nil
}

func (s *LearningHistoryDataSyncService) importMappingQuestionTag(ctx context.Context, payload []byte) error {
	sc := scanner.NewCSVScanner(bytes.NewReader(payload))
	// no columns
	if len(sc.GetRow()) == 0 {
		return nil
	}
	err := validateCSVMappingQuestionTagFormat(sc)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "validateCSVMappingQuestionTagFormat: %v", err)
	}

	mappingQuestionTags := make([]*entities.MappingQuestionTag, 0)
	for sc.Scan() {
		manabieTagID := sc.Text("manabie_tag_id")
		manabieTagName := sc.Text("manabie_tag_name")
		withusTagName := sc.Text("withus_tag_name")
		lastUpdatedDate := time.Now()
		lastUpdatedBy := interceptors.UserIDFromContext(ctx)
		isArchived := sc.Text("is_archived")

		row, err := newMappingQuestionTagRow(manabieTagID, manabieTagName, withusTagName, lastUpdatedDate, lastUpdatedBy, isArchived)
		if err != nil {
			line := sc.GetCurRow()
			return status.Errorf(codes.InvalidArgument, "newMappingQuestionTagRow: %v, line: %d", err, line)
		}
		mappingQuestionTag, err := row.toEntity()
		if err != nil {
			line := sc.GetCurRow()
			return status.Errorf(codes.InvalidArgument, "toEntity: %v, line: %d", err, line)
		}
		mappingQuestionTags = append(mappingQuestionTags, mappingQuestionTag)
	}

	// if no rows
	if len(mappingQuestionTags) == 0 {
		return status.Errorf(codes.InvalidArgument, "no data in mapping question tag csv file")
	}
	manabieTagIDTexts := make([]string, 0, len(mappingQuestionTags))
	for _, v := range mappingQuestionTags {
		manabieTagIDTexts = append(manabieTagIDTexts, v.ManabieTagID.String)
	}
	err = checkDuplicateData(manabieTagIDTexts)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
	err = s.LearningHistoryDataSyncRepo.BulkUpsertMappingQuestionTag(ctx, s.DBTrace, mappingQuestionTags)
	if err != nil {
		return status.Errorf(codes.Internal, "BulkUpsertMappingQuestionTag: %v", err)
	}
	return nil
}

func validateCSVMappingCourseIDFormat(sc scanner.CSVScanner) error {
	// length columns
	if len(sc.GetRow()) != 3 {
		return fmt.Errorf("csv file has invalid format - number of column should be 5")
	}
	if sc.GetRow()[0] != "manabie_course_id" {
		return fmt.Errorf("csv file has invalid format - first column (toLowerCase) should be 'manabie_course_id'")
	}
	if sc.GetRow()[1] != "withus_course_id" {
		return fmt.Errorf("csv file has invalid format - second column (toLowerCase) should be 'withus_course_id'")
	}
	if sc.GetRow()[2] != ARCHIVED_TEXT {
		return fmt.Errorf("csv file has invalid format - column %s (toLowerCase) should be 'is_archived'", sc.GetRow()[2])
	}
	return nil
}

func validateCSVMappingExamLoIDFormat(sc scanner.CSVScanner) error {
	// length columns
	if len(sc.GetRow()) != 3 {
		return fmt.Errorf("csv file has invalid format - number of column should be 5")
	}
	if sc.GetRow()[0] != "exam_lo_id" {
		return fmt.Errorf("csv file has invalid format - first column (toLowerCase) should be 'exam_lo_id'")
	}
	if sc.GetRow()[1] != "material_code" {
		return fmt.Errorf("csv file has invalid format - second column (toLowerCase) should be 'material_code'")
	}
	if sc.GetRow()[2] != ARCHIVED_TEXT {
		return fmt.Errorf("csv file has invalid format - column %s (toLowerCase) should be 'is_archived'", sc.GetRow()[2])
	}
	return nil
}

func validateCSVFailedSyncEmailRecipientFormat(sc scanner.CSVScanner) error {
	// length columns
	if len(sc.GetRow()) != 3 {
		return fmt.Errorf("csv file has invalid format - number of column should be 5")
	}
	if sc.GetRow()[0] != "recipient_id" {
		return fmt.Errorf("csv file has invalid format - first column (toLowerCase) should be 'recipient_id'")
	}
	if sc.GetRow()[1] != "email_address" {
		return fmt.Errorf("csv file has invalid format - second column (toLowerCase) should be 'email_address'")
	}
	if sc.GetRow()[2] != ARCHIVED_TEXT {
		return fmt.Errorf("csv file has invalid format - column %s (toLowerCase) should be 'is_archived'", sc.GetRow()[2])
	}
	return nil
}

func validateCSVMappingQuestionTagFormat(sc scanner.CSVScanner) error {
	// length columns
	if len(sc.GetRow()) != 4 {
		return fmt.Errorf("csv file has invalid format - number of column should be 6")
	}
	if sc.GetRow()[0] != "manabie_tag_id" {
		return fmt.Errorf("csv file has invalid format - first column (toLowerCase) should be 'manabie_tag_id'")
	}
	if sc.GetRow()[1] != "manabie_tag_name" {
		return fmt.Errorf("csv file has invalid format - second column (toLowerCase) should be 'manabie_tag_name'")
	}
	if sc.GetRow()[2] != "withus_tag_name" {
		return fmt.Errorf("csv file has invalid format - third column (toLowerCase) should be 'withus_tag_name'")
	}
	if sc.GetRow()[3] != ARCHIVED_TEXT {
		return fmt.Errorf("csv file has invalid format - column %s (toLowerCase) should be 'is_archived'", sc.GetRow()[3])
	}
	return nil
}
