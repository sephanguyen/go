package media

import (
	"bytes"
	"context"
	"fmt"
	httpPkg "net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/bob/services/uploads"
	"github.com/manabie-com/backend/internal/golibs/curl"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/speeches"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

// nolint
type MediaModifierService struct {
	http               curl.IHTTP
	text2SpeechBuilder interface {
		NewClient() error
		GetClient() speeches.IText2Speech
	}
	uploadReaderService interface {
		GenerateResumableUploadURL(ctx context.Context, req *bpb.ResumableUploadURLRequest) (*bpb.ResumableUploadURLResponse, error)
	}
}

func NewMediaModifierService(svc *uploads.UploadReaderService) bpb.MediaModifierServiceServer {
	return &MediaModifierService{
		uploadReaderService: svc,
		http: &curl.HTTP{
			InsecureSkipVerify: svc.Cfg.InsecureSkipVerify,
		},
		text2SpeechBuilder: &speeches.Text2SpeechBuilder{},
	}
}

func (svc *MediaModifierService) GenerateAudioFile(ctx context.Context, req *bpb.GenerateAudioFileRequest) (*bpb.GenerateAudioFileResponse, error) {
	ctx, span := interceptors.StartSpan(ctx, "SpeechesWriterService.GenerateAudioFile")
	defer span.End()

	if len(req.Options) == 0 {
		return nil, status.Error(codes.InvalidArgument, "options cannot be empty")
	}

	chanErrs := make(chan error, len(req.Options))
	chanResp := make(chan *bpb.AudioOptionResponse, len(req.Options))

	if err := svc.text2SpeechBuilder.NewClient(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	client := svc.text2SpeechBuilder.GetClient()
	defer client.Close()
	var wg sync.WaitGroup

	for _, option := range req.Options {
		wg.Add(1)
		go svc.generateFile(ctx, client, option, &wg, chanErrs, chanResp)
	}

	go func() {
		wg.Wait()
		close(chanErrs)
		close(chanResp)
	}()

	resps := make([]*bpb.AudioOptionResponse, 0, len(req.Options))
	for i := 0; i < len(req.Options); i++ {
		select {
		case res, ok := <-chanResp:
			if ok {
				resps = append(resps, res)
			}
		case err, ok := <-chanErrs:
			if ok {
				return nil, status.Error(codes.Internal, fmt.Sprintf("SpeechesWriterService.GenerateAudioFile: %s", err.Error()))
			}
		}
	}

	return &bpb.GenerateAudioFileResponse{
		Options: resps,
	}, nil
}

func (svc *MediaModifierService) generateFile(
	ctx context.Context,
	client speeches.IText2Speech,
	option *bpb.AudioOptionRequest,
	wg *sync.WaitGroup,
	chanErrs chan error,
	chanResp chan *bpb.AudioOptionResponse,
) {
	defer wg.Done()
	trimText := strings.Trim(option.Text, "")

	fileReq := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: trimText},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: option.Configs.Language,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
			Name:         "en-US-Standard-C",
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_LINEAR16,
			SpeakingRate:  1,
			Pitch:         0,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &fileReq)
	if err != nil {
		chanErrs <- err
		return
	}

	contentyType := httpPkg.DetectContentType(resp.AudioContent)

	url, err := svc.uploadReaderService.GenerateResumableUploadURL(ctx, &bpb.ResumableUploadURLRequest{
		Expiry:        durationpb.New(time.Second * time.Duration(3600)),
		PrefixName:    option.Text,
		FileExtension: "wav",
		ContentType:   contentyType,
	})
	if err != nil {
		chanErrs <- err
		return
	}

	if err := svc.http.Request(
		curl.PUT,
		url.ResumableUploadUrl,
		map[string]string{
			"Content-Length": strconv.Itoa(len(resp.AudioContent)),
		},
		bytes.NewReader(resp.AudioContent),
		nil,
	); err != nil {
		chanErrs <- err
		return
	}

	//nolint
	r, _ := regexp.Compile("^(https?|ftp|file)://(www.)?(.*?)([^(%2)]*.(wav))")
	chanResp <- &bpb.AudioOptionResponse{
		Link:    r.FindString(url.ResumableUploadUrl),
		QuizId:  option.QuizId,
		Type:    option.Type,
		Text:    option.Text,
		Configs: option.Configs,
	}
}

func (svc *MediaModifierService) UploadAsset(bpb.MediaModifierService_UploadAssetServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadAsset not implemented")
}

// nolint
func (svc *MediaModifierService) CreateBrightCoveUploadUrl(context.Context, *bpb.CreateBrightCoveUploadUrlRequest) (*bpb.CreateBrightCoveUploadUrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateBrightCoveUploadUrl not implemented")
}

func (svc *MediaModifierService) FinishUploadBrightCove(context.Context, *bpb.FinishUploadBrightCoveRequest) (*bpb.FinishUploadBrightCoveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinishUploadBrightCove not implemented")
}

func (svc *MediaModifierService) UpsertMedia(context.Context, *bpb.UpsertMediaRequest) (*bpb.UpsertMediaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertMedia not implemented")
}
