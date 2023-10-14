package services

import (
	context "context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/filestore"
	entities "github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	ib_mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories/items_bank"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestItemsBankService_ImportItems(t *testing.T) {
	t.Parallel()
	mockItemsBankRepo := new(ib_mock_repositories.MockItemsBankRepo)
	mockLearningMaterialRepo := new(repositories.MockLearningMaterialRepo)
	storageURL := "https://storage.com"
	fileMock := &filestore.Mock{
		GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
			return url.Parse(storageURL + "/" + objectName + "?key=1234567" + "&contentType=image/png")
		},
		GeneratePublicObjectURLMock: func(objectName string) string {
			return storageURL + "/" + objectName
		},
	}
	service := &ItemsBankService{
		DB:                   nil,
		ItemsBankRepo:        mockItemsBankRepo,
		LearningMaterialRepo: mockLearningMaterialRepo,
		FileStore:            fileMock,
	}

	correctHeader := `lo_id,item_id,item_name,item_description_text,item_description_image,question_type,point,question_content_text,question_content_image,explanation_text,explanation_image,option_text,option_image,correct_option`
	validMcqQuestion :=
		`,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE
	,,,,,,,,,,,Blue,,FALSE
	,,,,,,,,,,,Green,,FALSE`
	validMAQQuestion :=
		`,00000002,,,,MAQ,1,Which of this has the smallest wavelength?,,explanation 2,,Red,,FALSE
	,,,,,,,,,,,Blue,,FALSE
	,,,,,,,,,,,Green,,TRUE
	,,,,,,,,,,,Pink,,TRUE`
	validOrderingQuestion :=
		`,00000003,,,,ORD,1,Ordering question content,,explanation 3,,Monday,,
	,,,,,,,,,,,Tuesday,,
	,,,,,,,,,,,Wednesday,,
	,,,,,,,,,,,Thursday,,
	,,,,,,,,,,,Friday,,`
	validFibQuestion :=
		`,00000004,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,japan;usa,,
	,,,,,,,,,,,japan;america,,`
	validStqQuestion :=
		`,00000005,,,,STQ,1,"Dad, mother, Taro, younger brother, (  ) ways",,explanation,,5,,`

	sixtyQuestions := fmt.Sprintf(`%s`, correctHeader)
	for i := 0; i < 60; i++ {
		sixtyQuestions += fmt.Sprintf(`
		%s`, validMcqQuestion)
	}

	ctx := context.WithValue(
		context.Background(),
		interceptors.JwtClaims(0),
		&interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: "1",
			},
		},
	)

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Setup            func(ctx context.Context)
		Request          any
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name: "error: found some existing items",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{"00000001", "00000002"}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s
					%s`,
						correctHeader,
						validMcqQuestion,
						validMAQQuestion,
						validOrderingQuestion,
						validFibQuestion,
						validStqQuestion,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: int32(2),
						ErrorCode: sspb.ItemsBankErrorCode_ERR_ITEM_ID_ALREADY_EXISTS,
					},
					{
						RowNumber: int32(5),
						ErrorCode: sspb.ItemsBankErrorCode_ERR_ITEM_ID_ALREADY_EXISTS,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: upload data error",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockItemsBankRepo.On("UploadContentData", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s
					%s`,
						correctHeader,
						validMcqQuestion,
						validMAQQuestion,
						validOrderingQuestion,
						validFibQuestion,
						validStqQuestion,
					),
				),
			},
			ExpectedError: status.Error(codes.Internal, "error"),
		},
		{
			Name: "error: GetExistedIDs error",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s
					%s`,
						correctHeader,
						validMcqQuestion,
						validMAQQuestion,
						validOrderingQuestion,
						validFibQuestion,
						validStqQuestion,
					),
				),
			},
			ExpectedError: status.Error(codes.Internal, "error"),
		},
		{
			Name: "happy case",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockItemsBankRepo.On("UploadContentData", mock.Anything, "1", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s
					%s`,
						correctHeader,
						validMcqQuestion,
						validMAQQuestion,
						validOrderingQuestion,
						validFibQuestion,
						validStqQuestion,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: nil,
			},
			ExpectedError: nil,
		},
		{
			Name: "happy case with lo id",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockItemsBankRepo.On("UploadContentData", mock.Anything, "1", mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockLearningMaterialRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*entities.LearningMaterial{
						{
							ID:         database.Text("LO_ID_1"),
							VendorType: database.Text(cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String()),
						},
					}, nil,
				)
				mockItemsBankRepo.On("MapItemsByActivity", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)

			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`LO_ID_1,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,FALSE`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: nil,
			},
			ExpectedError: nil,
		},
		{
			Name: "error: map items by activity error",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockItemsBankRepo.On("UploadContentData", mock.Anything, "1", mock.Anything, mock.Anything).Once().Return([]string{
					"question_01",
				}, nil)
				mockLearningMaterialRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*entities.LearningMaterial{
						{
							ID:         database.Text("LO_ID_1"),
							VendorType: database.Text(cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String()),
						},
					}, nil,
				)
				mockItemsBankRepo.On("MapItemsByActivity", mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("error"))
				mockItemsBankRepo.On("ArchiveItems", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`LO_ID_1,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,FALSE`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: nil,
			},
			ExpectedError: status.Error(codes.Internal, "error"),
		},
		{
			Name: "csv format error: empty csv",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(``),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber:        -1,
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_UNKNOWN,
						ErrorDescription: "Empty payload",
					},
				},
			},
		},
		{
			Name: "csv format error: missing column csv header",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s
					%s`,
						`item_id,item_name,item_description_text,item_description_image,question_type,point,question_content_text,,explanation_text,explanation_image,option,correct_option`,
						validMcqQuestion,
						validMAQQuestion,
						validOrderingQuestion,
						validFibQuestion,
						validStqQuestion,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber:        -1,
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_UNKNOWN,
						ErrorDescription: "record on line 2: wrong number of fields",
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "csv format error: wrong column name csv header",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s
					%s`,
						`leaning_objective_id,item_id,item_name,item_description_text,item_description_image,question_type,point,question_content_text,question_content_image,explanation_text,explanation_image,option,correct_option`,
						validMcqQuestion,
						validMAQQuestion,
						validOrderingQuestion,
						validFibQuestion,
						validStqQuestion,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber:        -1,
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_UNKNOWN,
						ErrorDescription: "record on line 2: wrong number of fields",
					},
				},
			},
			ExpectedError: nil,
		},
		// Todo: @hohieuu update this test case
		// {
		// 	Name: "csv format error: csv without header",
		// 	Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
		// 	Setup: func(ctx context.Context) {
		// 		mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
		// 	},
		// 	Request: &sspb.ImportItemsRequest{
		// 		Payload: []byte(
		// 			fmt.Sprintf(`%s
		// 			%s
		// 			%s
		// 			%s
		// 			%s`,
		// 				validMcqQuestion,
		// 				validMAQQuestion,
		// 				validOrderingQuestion,
		// 				validFibQuestion,
		// 				validStqQuestion,
		// 			),
		// 		),
		// 	},
		// 	ExpectedResponse: &sspb.ImportItemsResponse{
		// 		Errors: nil,
		// 	},
		// 	ExpectedError: status.Error(codes.InvalidArgument, "invalid csv"),
		// },
		{
			Name: "csv format error: more than 50 questions",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					sixtyQuestions,
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber:        -1,
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_QUESTION_LIMIT_EXCEEDED,
						ErrorDescription: "Exceeded maximum question limit. Max 50 questions.",
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "csv format error: invalid data type -point field and correct option field",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,cau_nay_sai`,
						`,00000005,,,,STQ,********five_points********,"Dad, mother, Taro, younger brother, (  ) ways",,explanation,,5,,`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber:        2,
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_DATA_TYPE_INVALID,
						ErrorDescription: "",
					},
					{
						RowNumber:        3,
						ErrorCode:        sspb.ItemsBankErrorCode_ERR_DATA_TYPE_INVALID,
						ErrorDescription: "",
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: check required fields - missing item_id,question_type,question_content_text",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s`,
						correctHeader,
						`,,,,,STQ,1,"Dad, mother, Taro, younger brother, (  ) ways",,explanation,,5,,`,
						`,00000001,,,,,1,"Dad, mother, Taro, younger brother, (  ) ways",,explanation,,5,,`,
						`,00000001,,,,STQ,1,,,explanation,,5,,`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_REQUIRED_VALUE_MISSING,
					},
					{
						RowNumber: 3,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_REQUIRED_VALUE_MISSING,
					},
					{
						RowNumber: 4,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_REQUIRED_VALUE_MISSING,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: check required fields - put item_id for option row",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE
						,00000002,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,FALSE`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 3,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_REQUIRED_VALUE_MISSING,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: Only first item_id have item_name and item_description ",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
						`,00000002,Group Name 2 ,Group description 2,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
						`,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
						`,00000001,,,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 4,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_ITEM_DESCRIPTION_INVALID,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: question_type in [MCQ, MAQ, FIB, ORD, STQ]",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s
					%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,,WRONGTYPE,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
						`,00000001,,,,MAQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
						`,00000001,,,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
						`,00000001,,,,WRONGTYPE2,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
						`,00000001,,,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`,
						`,00000001,,,,HELLOWORLD,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE`),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_QUESTION_TYPE_INVALID,
					},
					{
						RowNumber: 5,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_QUESTION_TYPE_INVALID,
					},
					{
						RowNumber: 7,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_QUESTION_TYPE_INVALID,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: number of options = 0 with all question types",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockItemsBankRepo.On("UploadContentData", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s
					%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,,,TRUE`,
						`,00000002,Group Name ,Group description,,MAQ,1,Which of this has the smallest wavelength?,,explanation 1,,,,TRUE`,
						`,00000003,,,,ORD,1,Ordering question content,,explanation 3,,,,`,
						`,00000004,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,,,`,
						`,00000005,,,,STQ,1,"Dad, mother, Taro, younger brother, (  ) ways",,explanation,,,,`),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
					{
						RowNumber: 3,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
					{
						RowNumber: 4,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
					{
						RowNumber: 5,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
					{
						RowNumber: 6,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: MCQ - insufficient options, missing correct_option and no correct option selected.",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,,,TRUE`,
						`,00000001,,,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,option_1,,FALSE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,`,
						`,00000001,,,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,option_1,,FALSE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,FALSE`),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
					{
						RowNumber: 3,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_CORRECT_OPTION_MISSING,
					},
					{
						RowNumber: 6,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_NO_OPTION_SELECTED,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: MCQ - invalid number of correct options",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`,00000001,,,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,option_1,,TRUE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,TRUE`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_MCQ_CORRECT_OPTIONS_INVALID,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: MAQ - insufficient options, missing correct_option and no correct option selected.",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,,MAQ,1,Which of this has the smallest wavelength?,,explanation 1,,,,TRUE`,
						`,00000001,,,,MAQ,1,Which of this has the smallest wavelength?,,explanation 1,,option_1,,FALSE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,`,
						`,00000002,,,,MAQ,1,Which of this has the smallest wavelength?,,explanation 1,,option_1,,FALSE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,FALSE`),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
					{
						RowNumber: 3,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_CORRECT_OPTION_MISSING,
					},
					{
						RowNumber: 6,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_NO_OPTION_SELECTED,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: ORD - insufficient options",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`,00000003,,,,ORD,1,Ordering question content,,explanation 3,,Monday,,`),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: FIB - insufficient options, answers is invalid",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockItemsBankRepo.On("UploadContentData", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s`,
						correctHeader,
						`,00000004,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,,,`,
						`,00000004,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,only_1_answer,,`),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
					{
						RowNumber: 3,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_FIB_OPTIONS_INVALID,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: STQ - invalid number of options",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`,00000005,,,,STQ,1,"Dad, mother, Taro, younger brother, (  ) ways",,explanation,,,,`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_OPTIONS_MISSING,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: MIQ - number of options > 0",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`,00000005,,,,MIQ,1,"Dad, mother, Taro, younger brother",,explanation,,option,,`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_VALUE_OUT_OF_RANGE,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: MIQ - missing explanation",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`,00000005,,,,MIQ,1,"Dad, mother, Taro, younger brother",,,,,,`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_MIQ_EXPLANATION_MISSING,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: invalid lo id",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockLearningMaterialRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*entities.LearningMaterial{
						{
							ID:         database.Text("LO_001"),
							VendorType: database.Text(cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE.String()),
						},
						{
							ID:         database.Text("LO_004"),
							VendorType: database.Text(cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String()),
						},
						{
							ID:         database.Text("LO_003"),
							VendorType: database.Text(cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String()),
						},
					}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`LO_001,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,FALSE
						LO_003,00000004,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,japan;usa,,
						,,,,,,,,,,,japan;america,,
						LO_004,00000005,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,japan;usa,,
						,,,,,,,,,,,japan;america,,`),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_LO_ID_INVALID,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error: get lo id failed",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockLearningMaterialRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					nil, errors.New("error"))
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`LO_001,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,FALSE
						LO_003,00000004,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,japan;usa,,
						,,,,,,,,,,,japan;america,,
						LO_004,00000005,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,japan;usa,,
						,,,,,,,,,,,japan;america,,`),
				),
			},
			ExpectedResponse: nil,
			ExpectedError:    errors.New("error"),
		},
		{
			Name: "error: set activities success",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
				mockLearningMaterialRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					nil, errors.New("error"))
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`LO_001,00000001,Group Name ,Group description,,MCQ,1,Which of this has the smallest wavelength?,,explanation 1,,Red,,TRUE
						,,,,,,,,,,,Blue,,FALSE
						,,,,,,,,,,,Green,,FALSE
						LO_003,00000004,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,japan;usa,,
						,,,,,,,,,,,japan;america,,
						LO_004,00000005,,,,FIB,1,"In the 2023 WBC, team {{response}} won 3-2 in victory over team {{response}}",,WBC = World Baseball Classic ,,japan;usa,,
						,,,,,,,,,,,japan;america,,`),
				),
			},
			ExpectedResponse: nil,
			ExpectedError:    errors.New("error"),
		},
		{
			Name: "happy case - with valid images",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,item_image.png,MCQ,1,Which of this has the smallest wavelength?,content_image.png,explanation 1,exp_image.png,Red,option_image.png,TRUE
						,,,,,,,,,,,Blue,option_2.jpeg,FALSE
						,,,,,,,,,,,Green,option_3.jpg,FALSE`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: nil,
			},
			ExpectedError: nil,
		},
		{
			Name: "error case - with invalid images: explanation image",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,item_image.png,MCQ,1,Which of this has the smallest wavelength?,content_image.png,explanation 1,exp_image_without_extension,Red,option_image.png,TRUE
						,,,,,,,,,,,Blue,option_2.jpeg,FALSE
						,,,,,,,,,,,Green,option_3.png,FALSE`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_IMAGE_INVALID,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error case - with invalid images: option image, item image",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s
					%s`,
						correctHeader,
						`,00000001,Group Name ,Group description,item_image.png,MCQ,1,Which of this has the smallest wavelength?,content_image.png,explanation 1,exp_iamge.png,Red,option_image.png,TRUE
						,,,,,,,,,,,Blue,option_2.jpeg,FALSE
						,,,,,,,,,,,Green,option_3.hello,FALSE`,
						`,00000002,Group Name 2 ,Group description 2,item_image.,MCQ,1,Which of this has the smallest wavelength?,content_image.png,explanation 1,exp_iamge.png,Red,option_image.png,TRUE
						,,,,,,,,,,,Blue,option_2.jpeg,FALSE
						,,,,,,,,,,,Green,option_3.png,FALSE`,
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_IMAGE_INVALID,
					},
					{
						RowNumber: 5,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_IMAGE_INVALID,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "error case - invalid item id",
			Ctx:  ctx,
			Setup: func(ctx context.Context) {
				mockItemsBankRepo.On("GetExistedIDs", mock.Anything, mock.Anything).Once().Return([]string{}, nil)
			},
			Request: &sspb.ImportItemsRequest{
				Payload: []byte(
					fmt.Sprintf(`%s
					%s`,
						correctHeader,
						",00000005€0000001,,,,MIQ,1,Dad,,explanation,,,,",
					),
				),
			},
			ExpectedResponse: &sspb.ImportItemsResponse{
				Errors: []*sspb.ImportItemsResponseError{
					{
						RowNumber: 2,
						ErrorCode: sspb.ItemsBankErrorCode_ERR_ITEM_ID_INVALID,
					},
				},
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx)
			request := tc.Request.(*sspb.ImportItemsRequest)
			request.Payload = []byte(strings.ReplaceAll(string(request.Payload), "\t", ""))
			response, err := service.ImportItems(tc.Ctx, request)
			if err != nil {
				assert.Contains(t, err.Error(), tc.ExpectedError.Error())
			}
			if response != nil {
				assert.Equal(t, compareImportResponseErr(tc.ExpectedResponse.(*sspb.ImportItemsResponse), response), true)
			}
		})
	}
}

func compareImportResponseErr(expectedResp *sspb.ImportItemsResponse, actualResp *sspb.ImportItemsResponse) bool {
	if len(expectedResp.Errors) != len(actualResp.Errors) {
		fmt.Printf("Errors length: expected %v but got %v\n", len(expectedResp.Errors), len(actualResp.Errors))
		fmt.Println(actualResp)
		return false
	}

	for i := 0; i < len(expectedResp.Errors); i++ {
		if expectedResp.Errors[i].RowNumber != actualResp.Errors[i].RowNumber {
			fmt.Printf("RowNumber: expected %v but got %v at line %v\n", expectedResp.Errors[i].RowNumber, actualResp.Errors[i].RowNumber, i+1)
			return false
		}

		if expectedResp.Errors[i].ErrorCode != actualResp.Errors[i].ErrorCode {
			fmt.Printf("Error: expected %v but got %v at line %v\n", expectedResp.Errors[i].ErrorCode, actualResp.Errors[i].ErrorCode, i+1)
			return false
		}
	}

	return true
}

func TestImportItems_getOrganizationID(t *testing.T) {
	t.Parallel()
	mockItemsBankRepo := new(ib_mock_repositories.MockItemsBankRepo)
	mockLearningMaterialRepo := new(repositories.MockLearningMaterialRepo)
	service := &ItemsBankService{
		DB:                   nil,
		ItemsBankRepo:        mockItemsBankRepo,
		LearningMaterialRepo: mockLearningMaterialRepo,
	}
	ctx := context.WithValue(
		context.Background(),
		interceptors.JwtClaims(0),
		&interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: "1",
			},
		},
	)

	ctxWithoutResourcePath := context.Background()

	testCases := []struct {
		Name             string
		Ctx              context.Context
		ExpectedResponse string
		ExpectedError    error
	}{
		{
			Name:             "success",
			Ctx:              ctx,
			ExpectedResponse: "1",
			ExpectedError:    nil,
		},
		{
			Name:             "error: resource path is empty",
			Ctx:              ctxWithoutResourcePath,
			ExpectedResponse: "",
			ExpectedError:    status.Error(codes.Internal, "failed to parse jwt"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := service.getOrganizationID(tc.Ctx)
			assert.Equal(t, tc.ExpectedResponse, resp)
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			}
		})
	}

}

func TestGenerateItemBankResumableUploadURL(t *testing.T) {
	t.Parallel()
	tx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	orgID := "-24244848"
	ctx := context.WithValue(
		tx,
		interceptors.JwtClaims(0),
		&interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: orgID,
			},
		},
	)

	fileName := "test.png"
	baseURL := "https://example.com"
	bucketName := "manabie"
	storageURL := baseURL + "/" + bucketName
	fileMock := &filestore.Mock{
		GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
			return url.Parse(storageURL + "/" + objectName + "?key=1234567" + "&contentType=image/png")
		},
		GeneratePublicObjectURLMock: func(objectName string) string {
			return storageURL + "/" + objectName
		},
	}

	expectedResumableUploadURL := "https://example.com/manabie/items_bank/-24244848/test.png?key=1234567&contentType=image/png"
	expectedPublicURL := "https://example.com/manabie/items_bank/-24244848/test.png"

	mockErr := fmt.Errorf("mock error")

	tcs := []struct {
		name string
		filestore.FileStore
		req          *sspb.ItemBankResumableUploadURLRequest
		expectedResp *sspb.ItemBankResumableUploadURLResponse
		err          error
	}{
		{
			"happy case",
			fileMock,

			&sspb.ItemBankResumableUploadURLRequest{
				FileName: fileName,
				Expiry:   durationpb.New(time.Second * 10),
			},
			&sspb.ItemBankResumableUploadURLResponse{
				FileStoreUrl: &sspb.ItemBankResumableUploadURLResponse_FileStoreURL{
					ResumableUploadUrl: expectedResumableUploadURL,
					PublicUrl:          expectedPublicURL,
				},
				Expiry: durationpb.New(time.Second * 10),
			},
			nil,
		},
		{
			"expiry < min",
			fileMock,

			&sspb.ItemBankResumableUploadURLRequest{
				FileName: fileName,
				Expiry:   durationpb.New(time.Second * -1),
			},
			&sspb.ItemBankResumableUploadURLResponse{
				FileStoreUrl: &sspb.ItemBankResumableUploadURLResponse_FileStoreURL{
					ResumableUploadUrl: expectedResumableUploadURL,
					PublicUrl:          expectedPublicURL,
				},
				Expiry: durationpb.New(time.Second * 5),
			},
			nil,
		},
		{
			"expiry > max",
			fileMock,

			&sspb.ItemBankResumableUploadURLRequest{
				FileName: fileName,
				Expiry:   durationpb.New(time.Second * 60),
			},
			&sspb.ItemBankResumableUploadURLResponse{
				FileStoreUrl: &sspb.ItemBankResumableUploadURLResponse_FileStoreURL{
					ResumableUploadUrl: expectedResumableUploadURL,
					PublicUrl:          expectedPublicURL,
				},
				Expiry: durationpb.New(time.Second * 5),
			},
			nil,
		},
		{
			"give a error",
			&filestore.Mock{
				GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
					return nil, mockErr
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			&sspb.ItemBankResumableUploadURLRequest{
				FileName: fileName,
				Expiry:   durationpb.New(time.Second * -1),
			},
			nil,
			status.Error(codes.Internal, fmt.Errorf("error generate resumable object url: %v", mockErr).Error()),
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			service := &ItemsBankService{
				DB:        nil,
				FileStore: tc.FileStore,
				Cfg: &configs.StorageConfig{
					Endpoint:                 baseURL,
					MaximumURLExpiryDuration: time.Second * 10,
					MinimumURLExpiryDuration: time.Second * 1,
					DefaultURLExpiryDuration: time.Second * 5,
				},
			}
			actual, err := service.GenerateItemBankResumableUploadURL(ctx, tc.req)
			if tc.err != nil {
				assert.Error(tt, err)
				assert.Equal(t, err, tc.err)
			} else {
				assert.Equal(t, tc.expectedResp, actual)
			}
		})
	}
}

func TestGenerateListItemBankResumableUploadURL(t *testing.T) {
	// t.Parallel()
	tx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	orgID := "-24244848"
	ctx := context.WithValue(
		tx,
		interceptors.JwtClaims(0),
		&interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: orgID,
			},
		},
	)

	fileNames := []string{
		"test1.png", "test2.png", "test3.png", "test4.png", "test5.png",
		"test6.png", "test7.png", "test8.png", "test9.png", "test10.png",
		"test11.png", "test12.png", "test13.png", "test14.png", "test15.png",
		"test16.png", "test17.png", "test18.png", "test19.png", "test20.png",
		"test21.png", "test22.png", "test23.png", "test24.png", "test25.png",
		"test26.png", "test27.png", "test28.png", "test29.png", "test30.png",
	}
	baseURL := "https://example.com"
	bucketName := "manabie"
	storageURL := baseURL + "/" + bucketName
	fileMock := &filestore.Mock{
		GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
			return url.Parse(storageURL + "/" + objectName + "?key=1234567" + "&contentType=image/png")
		},
		GeneratePublicObjectURLMock: func(objectName string) string {
			return storageURL + "/" + objectName
		},
	}

	getExpectedResumableUploadURL := func(fileName string) string {
		return fmt.Sprintf("https://example.com/manabie/items_bank/-24244848/%s?key=1234567&contentType=image/png", fileName)
	}

	getExpectedPublicURL := func(fileName string) string {
		return fmt.Sprintf("https://example.com/manabie/items_bank/-24244848/%s", fileName)
	}

	getExpectedFireStoreURL := func(fileNames []string, expiry *durationpb.Duration) []*sspb.FileStoreURL {
		rs := make([]*sspb.FileStoreURL, 0)
		for _, fileName := range fileNames {
			rs = append(rs, &sspb.FileStoreURL{
				ResumableUploadUrl: getExpectedResumableUploadURL(fileName),
				PublicUrl:          getExpectedPublicURL(fileName),
				Expiry:             expiry,
				Error:              nil,
			})
		}
		return rs
	}

	mockErr := fmt.Errorf("mock error")
	mockContentBankMediaRepo := new(repositories.MockContentBankMediaRepo)

	tcs := []struct {
		name string
		filestore.FileStore
		Setup        func(ctx context.Context)
		req          *sspb.ListItemBankResumableUploadURLRequest
		expectedResp *sspb.ListItemBankResumableUploadURLResponse
		err          error
	}{
		{
			"happy case",
			fileMock,
			func(ctx context.Context) {
				mockContentBankMediaRepo.On("FindByMediaNames", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.ContentBankMedia{}, nil)
			},
			&sspb.ListItemBankResumableUploadURLRequest{
				FileNames: fileNames,
				Expiry:    durationpb.New(time.Second * 10),
			},
			&sspb.ListItemBankResumableUploadURLResponse{
				FileStoreUrls: getExpectedFireStoreURL(fileNames, durationpb.New(time.Second*10)),
			},
			nil,
		},
		{
			"expiry < min",
			fileMock,
			func(ctx context.Context) {
				mockContentBankMediaRepo.On("FindByMediaNames", ctx, mock.Anything, fileNames).Once().Return([]*entities.ContentBankMedia{}, nil)
			},
			&sspb.ListItemBankResumableUploadURLRequest{
				FileNames: fileNames,
				Expiry:    durationpb.New(time.Second * -1),
			},
			&sspb.ListItemBankResumableUploadURLResponse{
				FileStoreUrls: getExpectedFireStoreURL(fileNames, durationpb.New(time.Second*5)),
			},
			nil,
		},
		{
			"expiry > max",
			fileMock,
			func(ctx context.Context) {
				mockContentBankMediaRepo.On("FindByMediaNames", ctx, mock.Anything, fileNames).Return([]*entities.ContentBankMedia{}, nil)
			},
			&sspb.ListItemBankResumableUploadURLRequest{
				FileNames: fileNames,
				Expiry:    durationpb.New(time.Second * 60),
			},
			&sspb.ListItemBankResumableUploadURLResponse{
				FileStoreUrls: getExpectedFireStoreURL(fileNames, durationpb.New(time.Second*5)),
			},
			nil,
		},
		{
			"duplicated file names",
			&filestore.Mock{
				GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
					return nil, mockErr
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			func(ctx context.Context) {
				mockContentBankMediaRepo.On("FindByMediaNames", ctx, mock.Anything, fileNames).Once().Return([]*entities.ContentBankMedia{}, nil)
			},
			&sspb.ListItemBankResumableUploadURLRequest{
				FileNames: []string{fileNames[0], fileNames[0]},
				Expiry:    durationpb.New(time.Second * 10),
			},
			nil,
			status.Error(codes.InvalidArgument, fmt.Errorf("duplicated file name: %s", fileNames[0]).Error()),
		},
		{
			"exceed max file names",
			&filestore.Mock{
				GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
					return nil, mockErr
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			func(ctx context.Context) {
				mockContentBankMediaRepo.On("FindByMediaNames", ctx, mock.Anything, fileNames).Return([]*entities.ContentBankMedia{}, nil)
			},
			&sspb.ListItemBankResumableUploadURLRequest{
				FileNames: append(fileNames, "test31.png"),
				Expiry:    durationpb.New(time.Second * 10),
			},
			nil,
			status.Error(codes.InvalidArgument, fmt.Errorf("exceeded maximum file limit, max 30 files").Error()),
		},
		{
			"1 failed & 2 succeed",
			&filestore.Mock{
				GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
					if strings.Contains(objectName, fileNames[0]) || strings.Contains(objectName, fileNames[2]) {
						return nil, mockErr
					}
					return url.Parse(storageURL + "/" + objectName + "?key=1234567" + "&contentType=image/png")
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			func(ctx context.Context) {
				mockContentBankMediaRepo.On("FindByMediaNames", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.ContentBankMedia{}, nil)
			},
			&sspb.ListItemBankResumableUploadURLRequest{
				FileNames: []string{fileNames[0], fileNames[1], fileNames[2], fileNames[3], fileNames[4]},
				Expiry:    durationpb.New(time.Second * 10),
			},
			&sspb.ListItemBankResumableUploadURLResponse{
				FileStoreUrls: []*sspb.FileStoreURL{
					{
						PublicUrl:          "",
						ResumableUploadUrl: "",
						Error: &sspb.FileStoreURLError{
							FileName:         fileNames[0],
							ErrorCode:        sspb.FileStoreURLErrorCode_FIRE_STORE_URL_ERR_UNKNOWN,
							ErrorDescription: "error generate resumable object url: mock error",
						},
					},
					{
						PublicUrl:          getExpectedPublicURL(fileNames[1]),
						ResumableUploadUrl: getExpectedResumableUploadURL(fileNames[1]),
						Expiry:             durationpb.New(time.Second * 10),
						Error:              nil,
					},
					{
						PublicUrl:          "",
						ResumableUploadUrl: "",
						Error: &sspb.FileStoreURLError{
							FileName:         fileNames[2],
							ErrorCode:        sspb.FileStoreURLErrorCode_FIRE_STORE_URL_ERR_UNKNOWN,
							ErrorDescription: "error generate resumable object url: mock error",
						},
					},
					{
						PublicUrl:          getExpectedPublicURL(fileNames[3]),
						ResumableUploadUrl: getExpectedResumableUploadURL(fileNames[3]),
						Expiry:             durationpb.New(time.Second * 10),
						Error:              nil,
					},
					{
						PublicUrl:          getExpectedPublicURL(fileNames[4]),
						ResumableUploadUrl: getExpectedResumableUploadURL(fileNames[4]),
						Expiry:             durationpb.New(time.Second * 10),
						Error:              nil,
					},
				},
			},
			nil,
		},
		{
			"2 duplicated name & 1 failed & 2 succeed",
			&filestore.Mock{
				GenerateResumableObjectURLMock: func(ctx context.Context, objectName string, expiry time.Duration, allowOrigin, contentType string) (*url.URL, error) {
					if strings.Contains(objectName, fileNames[0]) {
						return nil, mockErr
					}
					return url.Parse(storageURL + "/" + objectName + "?key=1234567" + "&contentType=image/png")
				},
				GeneratePublicObjectURLMock: func(objectName string) string {
					return "https://example.com/manabie/" + objectName
				},
			},
			func(ctx context.Context) {
				mockContentBankMediaRepo.On("FindByMediaNames", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.ContentBankMedia{
					{
						ID:   database.Text("Media_ID_1"),
						Name: database.Text(fileNames[1]),
					},
					{
						ID:   database.Text("Media_ID_1"),
						Name: database.Text(fileNames[2]),
					},
				}, nil)
			},
			&sspb.ListItemBankResumableUploadURLRequest{
				FileNames: fileNames[:5],
				Expiry:    durationpb.New(time.Second * 10),
			},
			&sspb.ListItemBankResumableUploadURLResponse{
				FileStoreUrls: []*sspb.FileStoreURL{
					{
						PublicUrl:          "",
						ResumableUploadUrl: "",
						Error: &sspb.FileStoreURLError{
							FileName:         fileNames[0],
							ErrorCode:        sspb.FileStoreURLErrorCode_FIRE_STORE_URL_ERR_UNKNOWN,
							ErrorDescription: "error generate resumable object url: mock error",
						},
					},
					{
						PublicUrl:          "",
						ResumableUploadUrl: "",
						Error: &sspb.FileStoreURLError{
							FileName:         fileNames[1],
							ErrorCode:        sspb.FileStoreURLErrorCode_FIRE_STORE_URL_ERR_DUPLICATE_FILE_NAME,
							ErrorDescription: fmt.Sprintf("file name %s already existed", fileNames[1]),
						},
					},
					{
						PublicUrl:          "",
						ResumableUploadUrl: "",
						Error: &sspb.FileStoreURLError{
							FileName:         fileNames[2],
							ErrorCode:        sspb.FileStoreURLErrorCode_FIRE_STORE_URL_ERR_DUPLICATE_FILE_NAME,
							ErrorDescription: fmt.Sprintf("file name %s already existed", fileNames[2]),
						},
					},
					{
						PublicUrl:          getExpectedPublicURL(fileNames[3]),
						ResumableUploadUrl: getExpectedResumableUploadURL(fileNames[3]),
						Expiry:             durationpb.New(time.Second * 10),
						Error:              nil,
					},
					{
						PublicUrl:          getExpectedPublicURL(fileNames[4]),
						ResumableUploadUrl: getExpectedResumableUploadURL(fileNames[4]),
						Expiry:             durationpb.New(time.Second * 10),
						Error:              nil,
					},
				},
			},
			nil,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tc.Setup(ctx)

			service := &ItemsBankService{
				DB:                   nil,
				ContentBankMediaRepo: mockContentBankMediaRepo,
				FileStore:            tc.FileStore,
				Cfg: &configs.StorageConfig{
					Endpoint:                 baseURL,
					MaximumURLExpiryDuration: time.Second * 10,
					MinimumURLExpiryDuration: time.Second * 1,
					DefaultURLExpiryDuration: time.Second * 5,
				},
			}
			actual, err := service.GenerateListItemBankResumableUploadURL(ctx, tc.req)
			if tc.err != nil {
				assert.Error(tt, err)
				assert.Equal(t, err, tc.err)
			} else {
				assert.Equal(t, tc.expectedResp, actual)
			}
		})
	}
}

func TestValidateImageURL(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name     string
		imageURL string
		err      error
	}{
		{
			"happy case - png",
			"https://example.com/manabie/items_bank/-24244848/test.png",
			nil,
		},
		{
			"happy case - jpg",
			"https://example.com/manabie/items_bank/-24244848/test.jpg",
			nil,
		},
		{
			"happy case - jpeg",
			"https://example.com/manabie/items_bank/-24244848/test.jpeg",
			nil,
		},
		{
			"happy case - with special characters",
			"https://example.com/manabie/items_bank/-24244848/hie%5E%5Euho%21%21%21__@@@.png",
			nil,
		},
		{
			"invalid extension",
			"https://example.com/manabie/items_bank/-24244848/test.pdf",
			fmt.Errorf("invalid file extension - /manabie/items_bank/-24244848/test.pdf"),
		},
		{
			"invalid url - without scheme",
			"example.com/manabie/items_bank/-24244848/test.png",
			fmt.Errorf("invalid url - example.com/manabie/items_bank/-24244848/test.png"),
		},
		{
			"invalid url - without host",
			"https:///manabie/items_bank/-24244848/test.png",
			fmt.Errorf("invalid url - /manabie/items_bank/-24244848/test.png"),
		},
		{
			"invalid url - without extension",
			"https://example.com/manabie/items_bank/-24244848/test",
			fmt.Errorf("invalid file extension - /manabie/items_bank/-24244848/test"),
		},
		{
			"invalid url - without extension 2",
			"https://example.com/manabie/items_bank/-24244848/test.",
			fmt.Errorf("invalid file extension - /manabie/items_bank/-24244848/test."),
		},
		{
			"valid uppercase file extension - PNG",
			"https://example.com/manabie/items_bank/-24244848/test.PNG",
			nil,
		},
		{
			"valid uppercase file extension - JPG",
			"https://example.com/manabie/items_bank/-24244848/test.JPG",
			nil,
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			err := validateImageURL(tc.imageURL)
			if tc.err != nil {
				assert.Error(tt, err)
				assert.Equal(t, tc.err, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestUpsertMedia(t *testing.T) {
	mockContentBankMediaRepo := new(repositories.MockContentBankMediaRepo)
	service := &ItemsBankService{
		ContentBankMediaRepo: mockContentBankMediaRepo,
	}

	testCases := []struct {
		name        string
		setup       func()
		req         *sspb.UpsertMediaRequest
		expectedRes *sspb.UpsertMediaResponse
		expectedErr error
	}{
		{
			name: "missing required fields: name",
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing required fields: name"),
		},
		{
			name: "invalid file extension",
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test.jpge",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "https://manabie.storage.com/test.jpge",
					FileSize: 1,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid file extension"),
		},
		{
			name: "uppercase file extension",
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test.PNG",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "https://manabie.storage.com/test.PNG",
					FileSize: 1,
				},
			},
			setup: func() {
				mockContentBankMediaRepo.On("Upsert", mock.Anything, mock.Anything,
					mock.MatchedBy(func(media *entities.ContentBankMedia) bool {
						return media.Name.String == "test.PNG" &&
							media.Type.String == "MEDIA_TYPE_IMAGE" &&
							media.Resource.String == "https://manabie.storage.com/test.PNG" &&
							media.FileSizeBytes.Int == 1
					}),
				).Once().Return("updated-id", nil)
			},
			expectedErr: nil,
			expectedRes: &sspb.UpsertMediaResponse{
				MediaId: "updated-id",
			},
		},
		{
			name: "missing required fields: resource",
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing required fields: resource"),
		},
		{
			name: "invalid media type",
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test",
					Type:     sspb.MediaType_MEDIA_TYPE_NONE,
					Resource: "test",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing required fields: type"),
		},
		{
			name: "missing required fields: file_size",
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "https://test.com/abc.png",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing required fields: file_size"),
		},
		{
			name: "invalid file size",
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "https://test.com/abc.png",
					FileSize: -1,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid file size"),
		},
		{
			name: "upsert media error",
			setup: func() {
				mockContentBankMediaRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).Once().Return("", errors.New("unknown error"))
			},
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "https://test.com/abc.png",
					FileSize: 1,
				},
			},
			expectedErr: status.Error(codes.Internal, "ibs.ContentBankMediaRepo.Upsert: unknown error"),
		},
		{
			name: "success upsert media",
			setup: func() {
				mockContentBankMediaRepo.On("Upsert", mock.Anything, mock.Anything,
					mock.MatchedBy(func(media *entities.ContentBankMedia) bool {
						return media.Name.String == "test" &&
							media.Type.String == "MEDIA_TYPE_IMAGE" &&
							media.Resource.String == "https://manabie.storage.com/test.jpg" &&
							media.FileSizeBytes.Int == 1
					}),
				).Once().Return("updated-id", nil)
			},
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "https://manabie.storage.com/test.jpg",
					FileSize: 1,
				},
			},
			expectedErr: nil,
			expectedRes: &sspb.UpsertMediaResponse{
				MediaId: "updated-id",
			},
		},
		{
			name: "success with float size",
			setup: func() {
				mockContentBankMediaRepo.On("FindByNames", mock.Anything, mock.Anything, mock.Anything).Once().Return(
					[]*entities.ContentBankMedia{},
					nil)
				mockContentBankMediaRepo.On("Upsert", mock.Anything, mock.Anything,
					mock.MatchedBy(func(media *entities.ContentBankMedia) bool {
						return media.Name.String == "test" &&
							media.Type.String == "MEDIA_TYPE_IMAGE" &&
							media.Resource.String == "https://manabie.storage.com/test.jpg" &&
							media.FileSizeBytes.Int == 1000
					}),
				).Once().Return("res-id", nil)
			},
			req: &sspb.UpsertMediaRequest{
				Media: &sspb.Media{
					Name:     "test",
					Type:     sspb.MediaType_MEDIA_TYPE_IMAGE,
					Resource: "https://manabie.storage.com/test.jpg",
					FileSize: 1000.9,
				},
			},
			expectedErr: nil,
			expectedRes: &sspb.UpsertMediaResponse{
				MediaId: "res-id",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		if tc.setup != nil {
			tc.setup()
		}
		t.Run(tc.name, func(tt *testing.T) {
			resp, err := service.UpsertMedia(context.Background(), tc.req)

			if err != nil {
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				assert.Equal(t, tc.expectedRes, resp)
			}
		})
	}

}

func TestDeleteMedia(t *testing.T) {
	mockContentBankMediaRepo := new(repositories.MockContentBankMediaRepo)
	fileStoreMock := &filestore.Mock{}

	service := &ItemsBankService{
		ContentBankMediaRepo: mockContentBankMediaRepo,
		FileStore:            fileStoreMock,
	}

	mediaID := "media-id"
	orgID := "-24244848"

	testCases := []struct {
		name        string
		setup       func()
		req         *sspb.DeleteMediaRequest
		expectedRes *sspb.DeleteMediaResponse
		expectedErr error
	}{
		{
			name: "missing required fields: media_id",
			req: &sspb.DeleteMediaRequest{
				MediaId: "",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing required fields: media_id"),
		},
		{
			name: "media not found",
			setup: func() {
				mockContentBankMediaRepo.
					On("FindByID", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(nil, errors.New("media not found"))
			},
			req: &sspb.DeleteMediaRequest{
				MediaId: mediaID,
			},
			expectedErr: status.Error(codes.Internal, "ibs.ContentBankMediaRepo.FindByID: media not found"),
		},
		{
			name: "move object error",
			setup: func() {
				mockContentBankMediaRepo.
					On("FindByID", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&entities.ContentBankMedia{
						ID: database.Text(mediaID),
					}, nil)

				fileStoreMock.MoveObjectMock = func(ctx context.Context, src, dst string) error {
					return &filestore.Error{
						ErrorCode: filestore.UnknownError,
						Err:       errors.New("Move object error"),
					}
				}
			},
			req: &sspb.DeleteMediaRequest{
				MediaId: mediaID,
			},
			expectedErr: status.Error(
				codes.Internal,
				fmt.Sprintf("ibs.FileStore.MoveObject: %v", errors.New("Move object error")),
			),
		},
		{
			name: "delete media error",
			setup: func() {
				mockContentBankMediaRepo.
					On("FindByID", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&entities.ContentBankMedia{
						ID: database.Text(mediaID),
					}, nil)
				fileStoreMock.MoveObjectMock = func(ctx context.Context, src, dst string) error {
					return nil
				}

				mockContentBankMediaRepo.
					On("DeleteByID", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(errors.New("unknown error"))

			},
			req: &sspb.DeleteMediaRequest{
				MediaId: mediaID,
			},
			expectedErr: status.Error(codes.Internal, "ibs.ContentBankMediaRepo.DeleteByID: unknown error"),
		},
		{
			name: "success",
			setup: func() {
				mockContentBankMediaRepo.
					On("FindByID", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&entities.ContentBankMedia{
						ID: database.Text(mediaID),
					}, nil)
				fileStoreMock.MoveObjectMock = func(ctx context.Context, src, dst string) error {
					return nil
				}
				mockContentBankMediaRepo.
					On("DeleteByID", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(nil)
			},
			req: &sspb.DeleteMediaRequest{
				MediaId: mediaID,
			},
			expectedErr: nil,
			expectedRes: &sspb.DeleteMediaResponse{},
		},
		{
			name: "object not found in file store - still can delete media in db",
			setup: func() {
				mockContentBankMediaRepo.
					On("FindByID", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(&entities.ContentBankMedia{
						ID: database.Text(mediaID),
					}, nil)

				fileStoreMock.MoveObjectMock = func(ctx context.Context, src, dst string) error {
					return &filestore.Error{
						ErrorCode: filestore.FileNotFoundError,
						Err:       errors.New("Object not found"),
					}
				}
				mockContentBankMediaRepo.
					On("DeleteByID", mock.Anything, mock.Anything, mock.Anything).
					Once().
					Return(nil)
			},
			expectedErr: nil,
			req: &sspb.DeleteMediaRequest{
				MediaId: mediaID,
			},
			expectedRes: &sspb.DeleteMediaResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		if tc.setup != nil {
			tc.setup()
		}
		t.Run(tc.name, func(tt *testing.T) {
			ctx := context.WithValue(
				context.Background(),
				interceptors.JwtClaims(0),
				&interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						ResourcePath: orgID,
					},
				},
			)

			resp, err := service.DeleteMedia(ctx, tc.req)

			if err != nil {
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				assert.Equal(t, tc.expectedRes, resp)
			}
		})
	}
}
