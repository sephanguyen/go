package media

import (
	"context"
	"fmt"
	"testing"
	"time"

	mock_uploads "github.com/manabie-com/backend/mock/bob/services/uploads"
	mock_curl "github.com/manabie-com/backend/mock/golibs/curl"
	mock_speeches "github.com/manabie-com/backend/mock/golibs/speeches"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name           string
	ctx            context.Context
	req            interface{}
	expectedResp   interface{}
	expectedErr    error
	setup          func(ctx context.Context)
	expectedErrMsg string
	expectedCode   codes.Code
}

func TestGenerateAudioFile(t *testing.T) {
	t.Parallel()

	uploadServiceMock := new(mock_uploads.MockUploadReaderService)
	golibHTTPMock := new(mock_curl.IHTTP)
	t2sBuilder := new(mock_speeches.MockText2SpeechBuilder)
	t2sClient := new(mock_speeches.MockText2SpeechClient)

	s := &MediaModifierService{
		uploadReaderService: uploadServiceMock,
		http:                golibHTTPMock,
		text2SpeechBuilder:  t2sBuilder,
	}

	testCases := []TestCase{
		{
			name: "missing options in request",
			req: &bpb.GenerateAudioFileRequest{
				Options: []*bpb.AudioOptionRequest{},
			},
			expectedErr:  status.Error(codes.InvalidArgument, "options cannot be empty"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "generate file error",
			req: &bpb.GenerateAudioFileRequest{
				Options: []*bpb.AudioOptionRequest{
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("SpeechesWriterService.GenerateAudioFile: %s", fmt.Errorf("error SynthesizeSpeech").Error())),
			setup: func(ctx context.Context) {
				t2sBuilder.On("NewClient").Once().Return(nil)
				t2sBuilder.On("GetClient").Once().Return(t2sClient)
				t2sClient.On("Close").Once().Return(nil)
				t2sClient.On("SynthesizeSpeech", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error SynthesizeSpeech"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Test case: " + testCase.name)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			testCase.setup(ctx)

			_, err := s.GenerateAudioFile(ctx, testCase.req.(*bpb.GenerateAudioFileRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestGenerateAudioFile_AnotherCase(t *testing.T) {
	t.Parallel()

	uploadServiceMock := new(mock_uploads.MockUploadReaderService)
	golibHTTPMock := new(mock_curl.IHTTP)
	t2sBuilder := new(mock_speeches.MockText2SpeechBuilder)
	t2sClient := new(mock_speeches.MockText2SpeechClient)

	s := &MediaModifierService{
		uploadReaderService: uploadServiceMock,
		http:                golibHTTPMock,
		text2SpeechBuilder:  t2sBuilder,
	}

	testCases := []TestCase{
		{
			name: "generate resumable upload url error",
			req: &bpb.GenerateAudioFileRequest{
				Options: []*bpb.AudioOptionRequest{
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("SpeechesWriterService.GenerateAudioFile: %s", fmt.Errorf("error GenerateResumableUploadURL").Error())),
			setup: func(ctx context.Context) {
				t2sBuilder.On("NewClient").Once().Return(nil)
				t2sBuilder.On("GetClient").Once().Return(t2sClient)
				t2sClient.On("Close").Once().Return(nil)
				audioContent := "this is audio content"
				t2sClient.On("SynthesizeSpeech", mock.Anything, mock.Anything, mock.Anything).Return(&texttospeechpb.SynthesizeSpeechResponse{
					AudioContent: []byte(audioContent),
				}, nil)
				uploadServiceMock.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error GenerateResumableUploadURL"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Test case: " + testCase.name)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			testCase.setup(ctx)

			_, err := s.GenerateAudioFile(ctx, testCase.req.(*bpb.GenerateAudioFileRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestGenerateAudioFile_AnotherCase1(t *testing.T) {
	t.Parallel()

	uploadServiceMock := new(mock_uploads.MockUploadReaderService)
	golibHTTPMock := new(mock_curl.IHTTP)
	t2sBuilder := new(mock_speeches.MockText2SpeechBuilder)
	t2sClient := new(mock_speeches.MockText2SpeechClient)

	s := &MediaModifierService{
		uploadReaderService: uploadServiceMock,
		http:                golibHTTPMock,
		text2SpeechBuilder:  t2sBuilder,
	}

	testCases := []TestCase{
		{
			name: "upload file error",
			req: &bpb.GenerateAudioFileRequest{
				Options: []*bpb.AudioOptionRequest{
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("SpeechesWriterService.GenerateAudioFile: %s", fmt.Errorf("error Request").Error())),
			setup: func(ctx context.Context) {
				t2sBuilder.On("NewClient").Once().Return(nil)
				t2sBuilder.On("GetClient").Once().Return(t2sClient)
				t2sClient.On("Close").Once().Return(nil)
				audioContent := "this is audio content"
				t2sClient.On("SynthesizeSpeech", mock.Anything, mock.Anything, mock.Anything).Return(&texttospeechpb.SynthesizeSpeechResponse{
					AudioContent: []byte(audioContent),
				}, nil)
				uploadServiceMock.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Return(&bpb.ResumableUploadURLResponse{
					ResumableUploadUrl: "this is string upload",
				}, nil)
				golibHTTPMock.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error Request"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Test case: " + testCase.name)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			testCase.setup(ctx)

			_, err := s.GenerateAudioFile(ctx, testCase.req.(*bpb.GenerateAudioFileRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestGenerateAudioFile_AnotherCase2(t *testing.T) {
	t.Parallel()

	uploadServiceMock := new(mock_uploads.MockUploadReaderService)
	golibHTTPMock := new(mock_curl.IHTTP)
	t2sBuilder := new(mock_speeches.MockText2SpeechBuilder)
	t2sClient := new(mock_speeches.MockText2SpeechClient)

	s := &MediaModifierService{
		uploadReaderService: uploadServiceMock,
		http:                golibHTTPMock,
		text2SpeechBuilder:  t2sBuilder,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &bpb.GenerateAudioFileRequest{
				Options: []*bpb.AudioOptionRequest{
					{
						Configs: &bpb.AudioConfig{
							Language: "en-US",
						},
						Text:   "this is test",
						QuizId: "quiz_id",
					},
				},
			},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				t2sBuilder.On("NewClient").Once().Return(nil)
				t2sBuilder.On("GetClient").Once().Return(t2sClient)
				t2sClient.On("Close").Once().Return(nil)
				audioContent := "this is audio content"
				t2sClient.On("SynthesizeSpeech", mock.Anything, mock.Anything, mock.Anything).Return(&texttospeechpb.SynthesizeSpeechResponse{
					AudioContent: []byte(audioContent),
				}, nil)
				uploadServiceMock.On("GenerateResumableUploadURL", mock.Anything, mock.Anything).Return(&bpb.ResumableUploadURLResponse{
					ResumableUploadUrl: "this is string upload",
				}, nil)
				golibHTTPMock.On("Request", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log("Test case: " + testCase.name)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			testCase.setup(ctx)

			_, err := s.GenerateAudioFile(ctx, testCase.req.(*bpb.GenerateAudioFileRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
