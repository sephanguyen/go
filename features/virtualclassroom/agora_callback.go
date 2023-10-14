package virtualclassroom

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	recording "github.com/manabie-com/backend/internal/golibs/recording"
	lessonmgmt_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/middlewares"
)

func (s *suite) requestExistRecordingService(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	state, err := new(lessonmgmt_repo.LessonRoomStateRepo).GetLessonRoomStateByLessonID(ctx, s.LessonmgmtDB, database.Text(stepState.CurrentLessonID))
	if err != nil {
		return ctx, err
	}

	s.Request = domain.AgoraCallbackPayload{
		NoticeID:  "notice-id",
		ProductID: 34,
		EventType: domain.CloudRecordingServiceExited,
		NotifyMs:  32,
		Payload: domain.CloudRecordingPayload{
			ChannelName: stepState.CurrentLessonID,
			UID:         fmt.Sprintf(recording.UIDFormat, state.Recording.UID),
			SID:         state.Recording.SID,
			Sequence:    2,
			SendTS:      234235,
			ServiceType: 2,
		},
	}
	return ctx, nil
}

func (s *suite) AgoraCallback(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	url := fmt.Sprintf("%s/api/virtualclassroom/v1/agora-callback", s.VirtualClassroomHTTPSrvURL)
	bodyBytes, err := s.makeHTTPRequest(http.MethodPost, url)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if bodyBytes == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("body is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) makeHTTPRequest(method, url string) ([]byte, error) {
	bodyRequest, err := json.Marshal(s.Request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyRequest))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set(middlewares.AgoraHeaderKey, s.AgoraSignature)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: time.Duration(30) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}
	s.Response = resp
	return body, nil
}

func (s *suite) aValidAgoraSignatureInItsHeader(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	data, err := json.Marshal(s.Request)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	sig, err := s.generateSignature(s.Cfg.Agora.CallbackSignature, string(data))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.AgoraSignature = sig
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateSignature(key, message string) (string, error) {
	sig := hmac.New(sha256.New, []byte(key))
	if _, err := sig.Write([]byte(message)); err != nil {
		return "", err
	}
	return hex.EncodeToString(sig.Sum(nil)), nil
}
