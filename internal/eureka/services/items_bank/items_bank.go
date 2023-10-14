package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	entities "github.com/manabie-com/backend/internal/eureka/entities"
	itemsbank_entities "github.com/manabie-com/backend/internal/eureka/entities/items_bank"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	learnosity_entity "github.com/manabie-com/backend/internal/golibs/learnosity/entity"
	"github.com/manabie-com/backend/internal/golibs/try"
	scpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2/common"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	gocsv "github.com/gocarina/gocsv"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	// Root folder for items bank content
	itemsBankContentFolder = "items_bank"
)

type ItemsBankService struct {
	sspb.UnimplementedItemsBankServiceServer
	DB                   database.Ext
	LearningMaterialRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningMaterial, error)
	}
	ContentBankMediaRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, media *entities.ContentBankMedia) (mediaID string, err error)
		FindByMediaNames(ctx context.Context, db database.QueryExecer, mediaNames []string) ([]*entities.ContentBankMedia, error)
		FindByID(ctx context.Context, db database.QueryExecer, mediaID string) (*entities.ContentBankMedia, error)
		DeleteByID(ctx context.Context, db database.QueryExecer, mediaID string) error
	}
	ItemsBankRepo interface {
		GetExistedIDs(ctx context.Context, itemIDs []string) (existedIDs []string, err error)
		UploadContentData(ctx context.Context, organizationID string, items map[string]*itemsbank_entities.ItemsBankItem, questions []*itemsbank_entities.ItemsBankQuestion) ([]string, error)
		MapItemsByActivity(ctx context.Context, organizationID string, itemIDsByLoID map[string][]string) error
		ArchiveItems(ctx context.Context, itemIDs []string, uploadedQuestionID string) error
		GetListItems(ctx context.Context, itemIDs []string, next *string, limit uint32) (res *learnosity.Result, err error)
		GetCurrentItemIDs(ctx context.Context, loIDs []string) (map[string][]string, error)
	}
	Cfg       *configs.StorageConfig
	FileStore interface {
		GenerateResumableObjectURL(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error)
		GeneratePublicObjectURL(objectName string) string
		MoveObject(ctx context.Context, srcObjectName, dstObjectName string) error
	}
}

func (ibs *ItemsBankService) ImportItems(ctx context.Context, req *sspb.ImportItemsRequest) (*sspb.ImportItemsResponse, error) {
	csvRows := []*itemsbank_entities.ItemsBankCsvRow{}
	err := gocsv.UnmarshalBytes(req.GetPayload(), &csvRows)
	if err != nil {
		return &sspb.ImportItemsResponse{
			Errors: []*sspb.ImportItemsResponseError{
				{
					RowNumber:        -1,
					ErrorCode:        sspb.ItemsBankErrorCode_ERR_UNKNOWN,
					ErrorDescription: err.Error(),
				},
			},
		}, nil
	}

	if len(csvRows) == 0 {
		return &sspb.ImportItemsResponse{
			Errors: []*sspb.ImportItemsResponseError{
				{
					RowNumber:        -1,
					ErrorCode:        sspb.ItemsBankErrorCode_ERR_UNKNOWN,
					ErrorDescription: "Empty payload",
				},
			},
		}, nil
	}

	csvRows, err = ibs.convertAllCSVImageNamesToPublicURLs(ctx, csvRows)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("convertAllCSVImageNamesToPublicURLs: %w", err).Error())
	}

	csvItems, csvQuestions, validationErrors := ibs.parseCsvRows(csvRows)

	// validate number of questions
	if len(csvQuestions) > 50 {
		return &sspb.ImportItemsResponse{
			Errors: []*sspb.ImportItemsResponseError{
				{
					RowNumber:        -1,
					ErrorCode:        sspb.ItemsBankErrorCode_ERR_QUESTION_LIMIT_EXCEEDED,
					ErrorDescription: "Exceeded maximum question limit. Max 50 questions.",
				},
			},
		}, nil
	}

	// validate csv content
	err = ibs.validateCsvContent(ctx, csvItems, csvQuestions, &validationErrors)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(validationErrors) > 0 {
		return &sspb.ImportItemsResponse{
			Errors: validationErrors,
		}, nil
	}

	organizationID, err := ibs.getOrganizationID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	itemIDsByLoID := map[string][]string{}
	for _, item := range csvItems {
		if strings.TrimSpace(item.LoID) != "" {
			itemIDsByLoID[item.LoID] = append(itemIDsByLoID[item.LoID], item.ItemID)
		}
	}

	uploadedQuestionIDs, err := ibs.ItemsBankRepo.UploadContentData(ctx, organizationID, csvItems, csvQuestions)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(itemIDsByLoID) > 0 {
		err = ibs.ItemsBankRepo.MapItemsByActivity(ctx, organizationID, itemIDsByLoID)
		if err != nil {
			itemIDs := make([]string, 0, len(csvItems))
			for itemID := range csvItems {
				itemIDs = append(itemIDs, itemID)
			}
			if len(uploadedQuestionIDs) > 0 {
				// set items api must have at least 1 widget in the definition, so we have to pass an uploaded widget id
				uploadedQuestion := uploadedQuestionIDs[0]
				archiveErr := ibs.ItemsBankRepo.ArchiveItems(ctx, itemIDs, uploadedQuestion)
				if archiveErr != nil {
					err = fmt.Errorf("%v, %v", err, archiveErr)
				}
			}

			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &sspb.ImportItemsResponse{
		Errors: nil,
	}, nil
}

func (ibs *ItemsBankService) convertAllCSVImageNamesToPublicURLs(ctx context.Context, csvRows []*itemsbank_entities.ItemsBankCsvRow) ([]*itemsbank_entities.ItemsBankCsvRow, error) {
	orgID, err := ibs.getOrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization id: %w", err)
	}

	for _, row := range csvRows {
		itemDescriptionImage := strings.TrimSpace(row.ItemDescriptionImage)
		if itemDescriptionImage != "" {
			row.ItemDescriptionImage = ibs.getImageURLFromImageName(orgID, itemDescriptionImage)
		}

		questionContentImage := strings.TrimSpace(row.QuestionContentImage)
		if questionContentImage != "" {
			row.QuestionContentImage = ibs.getImageURLFromImageName(orgID, questionContentImage)
		}

		explanationImage := strings.TrimSpace(row.ExplanationImage)
		if explanationImage != "" {
			row.ExplanationImage = ibs.getImageURLFromImageName(orgID, explanationImage)
		}

		optionImage := strings.TrimSpace(row.OptionImage)
		if optionImage != "" {
			row.OptionImage = ibs.getImageURLFromImageName(orgID, optionImage)
		}
	}

	return csvRows, nil
}

func (ibs *ItemsBankService) validateCsvContent(ctx context.Context, csvItems map[string]*itemsbank_entities.ItemsBankItem, csvQuestions []*itemsbank_entities.ItemsBankQuestion, validationErrors *[]*sspb.ImportItemsResponseError) error {
	for _, question := range csvQuestions {
		err := validateQuestionContent(question)
		if err != nil {
			*validationErrors = append(*validationErrors, err)
		}

		err = validateQuestionImages(question)
		if err != nil {
			*validationErrors = append(*validationErrors, err)
		}
	}

	itemIDs := make([]string, 0, len(csvItems))
	lineNumberByLoID := make(map[string]int)
	loIDs := make([]string, 0, len(csvItems))

	for itemID := range csvItems {
		item := csvItems[itemID]
		if !item.IsItemIDValid() {
			*validationErrors = append(*validationErrors, &sspb.ImportItemsResponseError{
				RowNumber:        int32(item.LineNumber),
				ErrorCode:        sspb.ItemsBankErrorCode_ERR_ITEM_ID_INVALID,
				ErrorDescription: "Item ID is invalid",
			},
			)
		}
		itemIDs = append(itemIDs, itemID)

		if strings.TrimSpace(item.LoID) != "" {
			lineNumberByLoID[item.LoID] = item.LineNumber
			loIDs = append(loIDs, item.LoID)
		}

		if item.ItemDescriptionImage != "" {
			err := validateImageURL(item.ItemDescriptionImage)
			if err != nil {
				*validationErrors = append(*validationErrors, &sspb.ImportItemsResponseError{
					RowNumber:        int32(item.LineNumber),
					ErrorCode:        sspb.ItemsBankErrorCode_ERR_IMAGE_INVALID,
					ErrorDescription: fmt.Sprintf("Error validating item image url: %v", err),
				},
				)
			}
		}

		continue
	}

	existedIds, err := ibs.ItemsBankRepo.GetExistedIDs(ctx, itemIDs)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if len(existedIds) > 0 {
		for _, existedID := range existedIds {
			*validationErrors = append(
				*validationErrors,
				&sspb.ImportItemsResponseError{
					RowNumber:        int32(csvItems[existedID].LineNumber),
					ErrorCode:        sspb.ItemsBankErrorCode_ERR_ITEM_ID_ALREADY_EXISTS,
					ErrorDescription: fmt.Sprintf("Item_Id [%s] already exists, please create another item_id", existedID),
				},
			)
		}
	}

	if len(loIDs) > 0 {
		lms, err := ibs.LearningMaterialRepo.FindByIDs(ctx, ibs.DB, database.TextArray(loIDs))
		if err != nil {
			return fmt.Errorf("s.LearningMaterialRepo.FindByIDs: %w", err)
		}

		responseMap := make(map[string]bool)
		for _, lm := range lms {
			if lm.VendorType.String == cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String() {
				responseMap[lm.ID.String] = true
			}
		}

		for _, loID := range loIDs {
			if _, ok := responseMap[loID]; !ok {
				*validationErrors = append(
					*validationErrors,
					&sspb.ImportItemsResponseError{
						RowNumber:        int32(lineNumberByLoID[loID]),
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_LO_ID_INVALID,
						ErrorDescription: fmt.Sprintf("Lo_Id [%s] not found", loID),
					},
				)
			}
		}
	}

	return nil
}

func (ibs *ItemsBankService) parseCsvRows(csvRows []*itemsbank_entities.ItemsBankCsvRow) (map[string]*itemsbank_entities.ItemsBankItem, []*itemsbank_entities.ItemsBankQuestion, []*sspb.ImportItemsResponseError) {
	validationErrors := []*sspb.ImportItemsResponseError{}
	items := map[string]*itemsbank_entities.ItemsBankItem{}
	questions := []*itemsbank_entities.ItemsBankQuestion{}
	for i, row := range csvRows {
		csvLineNumber := i + 2
		if row.IsQuestionRow() {
			question, err := itemsbank_entities.NewItemsBankQuestion(csvLineNumber, row)
			if err != nil {
				validationErrors = append(validationErrors, &sspb.ImportItemsResponseError{
					RowNumber:        int32(csvLineNumber),
					ErrorCode:        sspb.ItemsBankErrorCode_ERR_DATA_TYPE_INVALID,
					ErrorDescription: fmt.Sprintf("Error create question: %v", err),
				})
				continue
			}
			question.AddOption(row.OptionText, row.OptionImage)

			if len(row.CorrectOption) != 0 {
				parsedCorrectOption, err := strconv.ParseBool(row.CorrectOption)
				if err != nil {
					validationErrors = append(validationErrors, &sspb.ImportItemsResponseError{
						RowNumber:        int32(csvLineNumber),
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_DATA_TYPE_INVALID,
						ErrorDescription: fmt.Sprintf("Error adding option: %v", err),
					})
					continue
				}
				question.AddCorrectOption(parsedCorrectOption)
			}
			questions = append(questions, question)

			itemID := question.ItemID
			_, exists := items[itemID]
			if exists {
				if len(row.ItemName) != 0 {
					validationErrors = append(validationErrors, &sspb.ImportItemsResponseError{
						RowNumber:        int32(question.LineNumber),
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_ITEM_DESCRIPTION_INVALID,
						ErrorDescription: "Invalid item description. Only the first item_id should have item_name and item_description specified.",
					})
					continue
				}
			} else {
				items[itemID] = &itemsbank_entities.ItemsBankItem{
					LineNumber:           csvLineNumber,
					ItemID:               row.ItemID,
					ItemName:             row.ItemName,
					ItemDescriptionText:  row.ItemDescriptionText,
					ItemDescriptionImage: row.ItemDescriptionImage,
					LoID:                 row.LoID,
				}
			}
		} else {
			if len(questions) == 0 {
				continue
			}
			lastQuestion := questions[len(questions)-1]
			lastQuestion.AddOption(row.OptionText, row.OptionImage)
			if len(row.CorrectOption) != 0 {
				parsedCorrectOption, err := strconv.ParseBool(row.CorrectOption)
				if err != nil {
					validationErrors = append(validationErrors, &sspb.ImportItemsResponseError{
						RowNumber:        int32(csvLineNumber),
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_DATA_TYPE_INVALID,
						ErrorDescription: fmt.Sprintf("Error adding option: %v", err),
					})
					continue
				}
				lastQuestion.AddCorrectOption(parsedCorrectOption)
			}
		}
	}

	return items, questions, validationErrors
}

func checkFIBOption(q *itemsbank_entities.ItemsBankQuestion) error {
	for _, option := range q.Options {
		answers := strings.Split(option.OptionText, ";")
		if len(answers) != strings.Count(q.QuestionContentText, "{{response}}") {
			return fmt.Errorf("number of answers is invalid")
		}
	}
	return nil
}

func validateQuestionImages(q *itemsbank_entities.ItemsBankQuestion) *sspb.ImportItemsResponseError {
	if q.QuestionContentImage != "" {
		err := validateImageURL(q.QuestionContentImage)
		if err != nil {
			return &sspb.ImportItemsResponseError{
				RowNumber:        int32(q.LineNumber),
				ErrorCode:        sspb.ItemsBankErrorCode_ERR_IMAGE_INVALID,
				ErrorDescription: fmt.Sprintf("Error validating question image url: %v", err),
			}
		}
	}

	if q.ExplanationImage != "" {
		err := validateImageURL(q.ExplanationImage)
		if err != nil {
			return &sspb.ImportItemsResponseError{
				RowNumber:        int32(q.LineNumber),
				ErrorCode:        sspb.ItemsBankErrorCode_ERR_IMAGE_INVALID,
				ErrorDescription: fmt.Sprintf("Error validating explanation image url: %v", err),
			}
		}
	}

	for _, option := range q.Options {
		optionImage := option.OptionImage
		if optionImage != "" {
			err := validateImageURL(optionImage)
			if err != nil {
				return &sspb.ImportItemsResponseError{
					RowNumber:        int32(q.LineNumber),
					ErrorCode:        sspb.ItemsBankErrorCode_ERR_IMAGE_INVALID,
					ErrorDescription: fmt.Sprintf("Error validating option image url: %v", err),
				}
			}
		}
	}
	return nil
}

func validateQuestionContent(q *itemsbank_entities.ItemsBankQuestion) *sspb.ImportItemsResponseError {
	if err := q.ValidateRequiredFields(); err != nil {
		return &sspb.ImportItemsResponseError{
			RowNumber:        int32(q.LineNumber),
			ErrorCode:        sspb.ItemsBankErrorCode_ERR_REQUIRED_VALUE_MISSING,
			ErrorDescription: fmt.Sprintf("Error validating required fields: %v", err),
		}
	}
	if err := q.ValidateQuestionType(); err != nil {
		return &sspb.ImportItemsResponseError{
			RowNumber:        int32(q.LineNumber),
			ErrorCode:        sspb.ItemsBankErrorCode_ERR_QUESTION_TYPE_INVALID,
			ErrorDescription: fmt.Sprintf("Error validating question type: %v", err),
		}
	}
	if err := q.ValidateNumberOfOptions(); err != nil {
		var errCode sspb.ItemsBankErrorCode
		if q.QuestionType == itemsbank_entities.QuestionManualInput {
			errCode = sspb.ItemsBankErrorCode_ERR_VALUE_OUT_OF_RANGE
		} else {
			errCode = sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING
		}
		return &sspb.ImportItemsResponseError{
			RowNumber:        int32(q.LineNumber),
			ErrorCode:        errCode,
			ErrorDescription: fmt.Sprintf("Error check options: %v", err),
		}
	}

	if q.QuestionType == itemsbank_entities.QuestionTypeMultipleChoice || q.QuestionType == itemsbank_entities.QuestionTypeMultipleAnswers {
		if len(q.CorrectOptions) < len(q.Options) {
			return &sspb.ImportItemsResponseError{
				RowNumber:        int32(q.LineNumber),
				ErrorCode:        sspb.ItemsBankErrorCode_ERR_CORRECT_OPTION_MISSING,
				ErrorDescription: fmt.Sprintf("Error check correct options: %v", fmt.Errorf("missing correct option")),
			}
		}

		if !slices.Contains(q.CorrectOptions, true) {
			return &sspb.ImportItemsResponseError{
				RowNumber:        int32(q.LineNumber),
				ErrorCode:        sspb.ItemsBankErrorCode_ERR_NO_OPTION_SELECTED,
				ErrorDescription: fmt.Sprintf("Error check correct options: %v", fmt.Errorf("no option selected")),
			}
		}
	}

	if q.QuestionType == itemsbank_entities.QuestionTypeMultipleChoice {
		trueOptions := 0
		for _, option := range q.CorrectOptions {
			if option {
				trueOptions++
			}
		}
		if trueOptions > 1 {
			return &sspb.ImportItemsResponseError{
				RowNumber:        int32(q.LineNumber),
				ErrorCode:        sspb.ItemsBankErrorCode_ERR_MCQ_CORRECT_OPTIONS_INVALID,
				ErrorDescription: fmt.Sprintf("Error check correct options: %v", fmt.Errorf("more than 1 correct option")),
			}
		}
	}

	if q.QuestionType == itemsbank_entities.QuestionTypeFillInTheBlank {
		if err := checkFIBOption(q); err != nil {
			return &sspb.ImportItemsResponseError{
				RowNumber:        int32(q.LineNumber),
				ErrorCode:        sspb.ItemsBankErrorCode_ERR_FIB_OPTIONS_INVALID,
				ErrorDescription: fmt.Sprintf("Error check FIB value: %v", err),
			}
		}
	}

	if q.QuestionType == itemsbank_entities.QuestionManualInput && q.IsExplanationEmpty() {
		return &sspb.ImportItemsResponseError{
			RowNumber:        int32(q.LineNumber),
			ErrorCode:        sspb.ItemsBankErrorCode_ERR_MIQ_EXPLANATION_MISSING,
			ErrorDescription: "error check explanation: missing explanation",
		}
	}

	return nil
}

func validateImageURL(imageURL string) error {
	url, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("url.Parse error: %v - %s", err, url.Path)
	}
	if url.Scheme == "" || url.Host == "" {
		return fmt.Errorf("invalid url - %s", url.Path)
	}
	imageTypes := []string{".png", ".jpg", ".jpeg"}
	lowerExt := strings.ToLower(path.Ext(url.Path))
	if !slices.Contains(imageTypes, lowerExt) {
		return fmt.Errorf("invalid file extension - %s", url.Path)
	}
	return nil
}

func (ibs *ItemsBankService) getOrganizationID(ctx context.Context) (string, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}
	organizationID := organization.OrganizationID().String()

	return organizationID, nil
}

func (ibs *ItemsBankService) normalizeExpiry(expiry time.Duration) time.Duration {
	if expiry > ibs.Cfg.MaximumURLExpiryDuration || expiry < ibs.Cfg.MinimumURLExpiryDuration {
		expiry = ibs.Cfg.DefaultURLExpiryDuration
	}

	return expiry
}

func (ibs *ItemsBankService) generateFilePathBasedOnContext(ctx context.Context, fileName string) (string, error) {
	orgID, err := ibs.getOrganizationID(ctx)
	if err != nil {
		return "", fmt.Errorf("missing organization in context: %w", err)
	}

	fullPath := ibs.getFileStoreObjectName(orgID, fileName)
	return fullPath, nil
}

func (ibs *ItemsBankService) getImageURLFromImageName(orgID string, imageName string) string {
	objectName := ibs.getFileStoreObjectName(orgID, imageName)

	return ibs.FileStore.GeneratePublicObjectURL(objectName)
}

func (ibs *ItemsBankService) getFileStoreObjectName(orgID string, fileName string) string {
	return fmt.Sprintf("%s/%s/%s", itemsBankContentFolder, orgID, fileName)
}

func (ibs *ItemsBankService) GenerateItemBankResumableUploadURL(ctx context.Context, req *sspb.ItemBankResumableUploadURLRequest) (*sspb.ItemBankResumableUploadURLResponse, error) {
	fireStoreURL, err := ibs.generateResumableUploadURL(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sspb.ItemBankResumableUploadURLResponse{
		FileStoreUrl: &sspb.ItemBankResumableUploadURLResponse_FileStoreURL{
			ResumableUploadUrl: fireStoreURL.ResumableUploadUrl,
			PublicUrl:          fireStoreURL.PublicUrl,
		},
		Expiry: fireStoreURL.Expiry,
	}, nil
}

func (ibs *ItemsBankService) validateGenerateListItemBankResumableUploadURL(req *sspb.ListItemBankResumableUploadURLRequest) error {
	if len(req.FileNames) > 30 {
		return fmt.Errorf("exceeded maximum file limit, max 30 files")
	}

	fileNames := make(map[string]bool)
	for _, fileName := range req.FileNames {
		if _, ok := fileNames[fileName]; ok {
			return fmt.Errorf("duplicated file name: %s", fileName)
		}
		fileNames[fileName] = true
	}

	return nil
}

func (ibs *ItemsBankService) getExistedMediaNames(ctx context.Context, fileNames []string) (map[string]bool, error) {
	medias, err := ibs.ContentBankMediaRepo.FindByMediaNames(ctx, ibs.DB, fileNames)
	if err != nil {
		return nil, fmt.Errorf("error ibs.ContentBankMediaRepo.FindByMediaNames: %w", err)
	}
	existedMediaNames := make(map[string]bool)
	for _, media := range medias {
		existedMediaNames[media.Name.String] = true
	}

	return existedMediaNames, nil
}

func (ibs *ItemsBankService) GenerateListItemBankResumableUploadURL(ctx context.Context, req *sspb.ListItemBankResumableUploadURLRequest) (*sspb.ListItemBankResumableUploadURLResponse, error) {
	if err := ibs.validateGenerateListItemBankResumableUploadURL(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	existedMediaNames, err := ibs.getExistedMediaNames(ctx, req.FileNames)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	type FileStoreURLWithIndex struct {
		FileStoreURL *sspb.FileStoreURL
		Index        int
	}

	results := make([]*sspb.FileStoreURL, len(req.FileNames))
	resultChan := make(chan *FileStoreURLWithIndex, len(req.FileNames))

	var wg sync.WaitGroup

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	for i, fileName := range req.FileNames {
		if _, isExisted := existedMediaNames[fileName]; isExisted {
			fileStoreError := &sspb.FileStoreURLError{
				FileName:         fileName,
				ErrorCode:        sspb.FileStoreURLErrorCode_FIRE_STORE_URL_ERR_DUPLICATE_FILE_NAME,
				ErrorDescription: fmt.Sprintf("file name %s already existed", fileName),
			}

			results[i] = &sspb.FileStoreURL{
				ResumableUploadUrl: "",
				PublicUrl:          "",
				Error:              fileStoreError,
			}
		} else {
			wg.Add(1)
			go func(index int, fileName string, allowOrigin string) {
				defer wg.Done()

				fireStoreURL, err := ibs.generateResumableUploadURL(ctxWithTimeout, &sspb.ItemBankResumableUploadURLRequest{
					FileName:    fileName,
					AllowOrigin: allowOrigin,
					Expiry:      req.Expiry,
				})
				if err != nil {
					fileStoreError := &sspb.FileStoreURLError{
						FileName:         fileName,
						ErrorCode:        sspb.FileStoreURLErrorCode_FIRE_STORE_URL_ERR_UNKNOWN,
						ErrorDescription: err.Error(),
					}
					resultChan <- &FileStoreURLWithIndex{
						FileStoreURL: &sspb.FileStoreURL{
							ResumableUploadUrl: "",
							PublicUrl:          "",
							Error:              fileStoreError,
						},
						Index: index,
					}
					return
				}

				resultChan <- &FileStoreURLWithIndex{
					FileStoreURL: fireStoreURL,
					Index:        index,
				}
			}(i, fileName, req.AllowOrigin)
		}
	}

	go func() {
		wg.Wait() // Wait for all goroutines to finish
		close(resultChan)
	}()

	for v := range resultChan {
		results[v.Index] = v.FileStoreURL
	}

	return &sspb.ListItemBankResumableUploadURLResponse{
		FileStoreUrls: results,
	}, nil
}

func (ibs *ItemsBankService) generateResumableUploadURL(ctx context.Context, req *sspb.ItemBankResumableUploadURLRequest) (*sspb.FileStoreURL, error) {
	expiry := req.Expiry.AsDuration()
	expiry = ibs.normalizeExpiry(expiry)

	objectName, err := ibs.generateFilePathBasedOnContext(ctx, req.FileName)
	if err != nil {
		return nil, fmt.Errorf("error generating path: %v", err)
	}

	resumableUploadURL, err := ibs.FileStore.GenerateResumableObjectURL(ctx, objectName, expiry, req.AllowOrigin, "")
	if err != nil {
		return nil, fmt.Errorf("error generate resumable object url: %v", err)
	}

	return &sspb.FileStoreURL{
		ResumableUploadUrl: resumableUploadURL.String(),
		PublicUrl:          ibs.FileStore.GeneratePublicObjectURL(objectName),
		Expiry:             durationpb.New(expiry),
	}, nil
}

func (ibs *ItemsBankService) UpsertMedia(ctx context.Context, req *sspb.UpsertMediaRequest) (*sspb.UpsertMediaResponse, error) {
	err := ibs.validateMediaRequest(req)
	if err != nil {
		return nil, err
	}

	userID := interceptors.UserIDFromContext(ctx)
	media := new(entities.ContentBankMedia)
	now := time.Now()
	err = multierr.Combine(
		media.ID.Set(idutil.ULIDNow()),
		media.Name.Set(req.Media.Name),
		media.Type.Set(req.Media.Type),
		media.Resource.Set(req.Media.Resource),
		media.FileSizeBytes.Set(int64(req.Media.FileSize)),
		media.CreatedBy.Set(userID),
		media.CreatedAt.Set(now),
		media.UpdatedAt.Set(now),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("multierr.Combine: %v", err).Error())
	}

	mediaID, err := ibs.ContentBankMediaRepo.Upsert(ctx, ibs.DB, media)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ibs.ContentBankMediaRepo.Upsert: %v", err).Error())
	}
	resp := new(sspb.UpsertMediaResponse)
	resp.MediaId = mediaID

	return resp, nil
}

func (ibs *ItemsBankService) validateMediaRequest(req *sspb.UpsertMediaRequest) error {
	if req.Media.Name == "" {
		return status.Error(codes.InvalidArgument, "missing required fields: name")
	}
	if req.Media.Type.String() != sspb.MediaType_MEDIA_TYPE_IMAGE.String() {
		return status.Error(codes.InvalidArgument, "missing required fields: type")
	}
	if req.Media.Resource == "" {
		return status.Error(codes.InvalidArgument, "missing required fields: resource")
	}
	allowedExtensions := []string{".jpg", ".png", ".jpeg"}
	lowerExt := strings.ToLower(path.Ext(req.Media.Resource))
	if !slices.Contains(allowedExtensions, lowerExt) {
		return status.Error(codes.InvalidArgument, "invalid file extension")
	}
	if req.Media.FileSize == 0 {
		return status.Error(codes.InvalidArgument, "missing required fields: file_size")
	}

	if req.Media.FileSize < 0 {
		return status.Error(codes.InvalidArgument, "invalid file size")
	}

	return nil
}

func (ibs *ItemsBankService) GetItemsByLM(ctx context.Context, req *sspb.GetItemsByLMRequest) (*sspb.GetItemsByLMResponse, error) {
	loIDs := req.GetLearningMaterialId()
	offset := req.Paging.GetOffsetString()
	limit := req.Paging.GetLimit()
	var offsetToGetData *string
	if offset == "" {
		offsetToGetData = nil
	} else {
		offsetToGetData = &offset
	}

	currentItemIDs, err := ibs.ItemsBankRepo.GetCurrentItemIDs(ctx, loIDs)
	itemIDs := []string{}
	if err != nil {
		return nil, err
	}
	for _, v := range currentItemIDs {
		itemIDs = append(itemIDs, v...)
	}

	resultGetItems, err := ibs.ItemsBankRepo.GetListItems(ctx, itemIDs, offsetToGetData, limit)
	var offsetPagi *scpb.Paging_OffsetString
	var itemsLearnosityResponse []*learnosity_entity.Item
	var itemsRes []*sspb.GetItemsByLMResponse_Items

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	err = json.Unmarshal(resultGetItems.Data, &itemsLearnosityResponse)
	if err != nil {
		return nil, err
	}
	offsetString, ok := resultGetItems.Meta["next"]
	if ok {
		offsetPagi = &scpb.Paging_OffsetString{
			OffsetString: offsetString.(string),
		}
	}

	for _, ref := range itemsLearnosityResponse {
		temp := sspb.GetItemsByLMResponse_Items{
			Reference: ref.Reference,
		}
		itemsRes = append(itemsRes, &temp)
	}

	return &sspb.GetItemsByLMResponse{
		Items: itemsRes,
		NextPage: &scpb.Paging{
			Limit:  req.Paging.Limit,
			Offset: offsetPagi,
		},
	}, nil
}

func (ibs *ItemsBankService) DeleteMedia(ctx context.Context, req *sspb.DeleteMediaRequest) (*sspb.DeleteMediaResponse, error) {
	mediaID := req.GetMediaId()
	if mediaID == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields: media_id")
	}

	media, err := ibs.ContentBankMediaRepo.FindByID(ctx, ibs.DB, mediaID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ibs.ContentBankMediaRepo.FindByID: %v", err))
	}

	mediaName := media.Name.String
	orgID, err := ibs.getOrganizationID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ibs.getOrganizationID: %v", err))
	}
	sourceObjPath := ibs.getFileStoreObjectName(orgID, mediaName)
	binObjPath := ibs.getFileStoreObjectName(orgID, fmt.Sprintf("bin/%s", mediaName))

	// move object to bin folder
	fileStoreErr := ibs.FileStore.MoveObject(ctx,
		sourceObjPath,
		binObjPath,
	)

	if fileStoreErr != nil {
		if moveObjErr, ok := fileStoreErr.(*filestore.Error); ok {
			if moveObjErr.ErrorCode != filestore.FileNotFoundError {
				return nil, status.Error(codes.Internal, fmt.Sprintf("ibs.FileStore.MoveObject: %v", moveObjErr.Err))
			}
		}
	}

	// delete object in DB
	err = ibs.ContentBankMediaRepo.DeleteByID(ctx, ibs.DB, mediaID)
	if err != nil {
		retryErr := try.Do(func(attempt int) (bool, error) {
			// revert move object from bin folder
			fileStoreErr = ibs.FileStore.MoveObject(ctx,
				binObjPath,
				sourceObjPath,
			)
			if fileStoreErr == nil {
				return false, nil
			}

			retry := attempt < 5
			if retry {
				return true, fileStoreErr
			}

			return false, fileStoreErr
		})

		if retryErr != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("ibs.FileStore.MoveObject - revert failed: %v", err))
		}

		return nil, status.Error(codes.Internal, fmt.Sprintf("ibs.ContentBankMediaRepo.DeleteByID: %v", err))
	}

	return &sspb.DeleteMediaResponse{}, nil
}
