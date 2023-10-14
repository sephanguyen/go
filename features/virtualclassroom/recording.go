package virtualclassroom

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	lessonmgmt_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	lesson_media_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	liveroom_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *suite) userStartRecording(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)
	req := &vpb.StartRecordingRequest{
		LessonId:           stepState.CurrentLessonID,
		SubscribeVideoUids: []string{"#allstream#"},
		SubscribeAudioUids: []string{"#allstream#"},
		FileNamePrefix:     []string{stepState.CurrentLessonID, currentTime},
		TranscodingConfig: &vpb.StartRecordingRequest_TranscodingConfig{
			Height:           720,
			Width:            1280,
			Bitrate:          2000,
			Fps:              30,
			MixedVideoLayout: 0,
			BackgroundColor:  "#FF0000",
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).StartRecording(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userStartsRecordingInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)

	req := &vpb.StartRecordingRequest{
		LessonId:           stepState.CurrentLessonID,
		SubscribeVideoUids: []string{"#allstream#"},
		SubscribeAudioUids: []string{"#allstream#"},
		FileNamePrefix:     []string{stepState.CurrentLessonID, currentTime},
		TranscodingConfig: &vpb.StartRecordingRequest_TranscodingConfig{
			Height:           720,
			Width:            1280,
			Bitrate:          2000,
			Fps:              30,
			MixedVideoLayout: 0,
			BackgroundColor:  "#FF0000",
		},
		ChannelId: stepState.CurrentChannelID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).
		StartRecording(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userStartsRecordingInTheLiveRoomOnly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	currentTime := strconv.FormatInt(time.Now().Unix(), 10)

	req := &vpb.StartRecordingRequest{
		SubscribeVideoUids: []string{"#allstream#"},
		SubscribeAudioUids: []string{"#allstream#"},
		FileNamePrefix:     []string{stepState.CurrentLessonID, currentTime},
		TranscodingConfig: &vpb.StartRecordingRequest_TranscodingConfig{
			Height:           720,
			Width:            1280,
			Bitrate:          2000,
			Fps:              30,
			MixedVideoLayout: 0,
			BackgroundColor:  "#FF0000",
		},
		ChannelId: stepState.CurrentChannelID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).
		StartRecording(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) startRecordingStateIsUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	state, err := new(lessonmgmt_repo.LessonRoomStateRepo).GetLessonRoomStateByLessonID(ctx, s.LessonmgmtDBTrace, database.Text(stepState.CurrentLessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in fetching lesson room state %s: %w", stepState.CurrentLessonID, err)
	}

	if state.Recording.Creator != stepState.CurrentUserID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("creator expect %s but got %s", stepState.CurrentUserID, state.Recording.Creator)
	}
	if !state.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("isRecording expect true but got false")
	}
	if state.Recording.ResourceID == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("resourceId expect not empty")
	}
	if state.Recording.SID == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("sID expect not empty")
	}
	if state.Recording.UID == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("uID expect not 0")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsTheLiveRoomStateRecordingHas(ctx context.Context, state string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := new(liveroom_repo.LiveRoomStateRepo).GetLiveRoomStateByChannelID(ctx, s.LessonmgmtDBTrace, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if res.Recording == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected recording state to be not empty")
	}
	recordingState := res.Recording

	switch state {
	case Started:
		if recordingState.Creator != stepState.CurrentUserID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("creator expect %s but got %s", stepState.CurrentUserID, recordingState.Creator)
		}
		if !recordingState.IsRecording {
			return StepStateToContext(ctx, stepState), fmt.Errorf("isRecording expect true but got false")
		}
		if recordingState.ResourceID == "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("resourceId expect not empty")
		}
		if recordingState.SID == "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("sID expect not empty")
		}
		if recordingState.UID == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("uID expect not 0")
		}
	case Stopped:
		if recordingState.Creator != "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("creator expect empty but got %s", recordingState.Creator)
		}
		if recordingState.IsRecording {
			return StepStateToContext(ctx, stepState), fmt.Errorf("isRecording expect false but got true")
		}
		if recordingState.ResourceID != "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("resourceId expect empty")
		}
		if recordingState.SID != "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("sID expect empty")
		}
		if recordingState.UID != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("uID expect 0")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetStartRecordingState(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	state, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if state.Recording.Creator != stepState.CurrentUserID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("creator expect %s but got %s", stepState.CurrentUserID, state.Recording.Creator)
	}

	if !state.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("isRecording expect true but got false")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userStopRecording(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.StopRecordingRequest{
		LessonId: stepState.CurrentLessonID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).StopRecording(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userStopRecordingInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.StopRecordingRequest{
		LessonId:  stepState.CurrentLessonID,
		ChannelId: stepState.CurrentChannelID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).
		StopRecording(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userStopRecordingInTheLiveRoomOnly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.StopRecordingRequest{
		ChannelId: stepState.CurrentChannelID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).
		StopRecording(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) stopRecordingStateIsUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	state, err := new(lessonmgmt_repo.LessonRoomStateRepo).GetLessonRoomStateByLessonID(ctx, s.LessonmgmtDBTrace, database.Text(stepState.CurrentLessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in fetching lesson room state %s: %w", stepState.CurrentLessonID, err)
	}

	if state.Recording.Creator != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("creator expect empty but got %s", state.Recording.Creator)
	}
	if state.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("isRecording expect false but got true")
	}
	if state.Recording.ResourceID != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("resourceId expect empty")
	}
	if state.Recording.SID != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("sID expect empty")
	}
	if state.Recording.UID != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("uID expect 0")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetStopRecordingState(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	state, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if state.Recording.Creator != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("creator expect empty but got %s", state.Recording.Creator)
	}

	if state.Recording.IsRecording {
		return StepStateToContext(ctx, stepState), fmt.Errorf("isRecording expect false but got true")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) recordedVideosAreSaved(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonID := stepState.CurrentLessonID
	records, err := (&repo.RecordedVideoRepo{}).ListRecordingByLessonIDs(ctx, s.LessonmgmtDB, []string{lessonID})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("RecordedVideoRepo.ListByIDs: %v", err)
	}

	if len(records) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not find recorded video with lessonID %s", lessonID)
	}

	// get media to fill video id field
	medias, err := new(lesson_media_infrastructure.MediaRepo).RetrieveMediasByIDs(ctx, s.BobDB, records.GetMediaIDs())
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("MediaRepo.RetrieveMediasByIDs: %v", err)
	}
	if err = records.WithMedias(medias); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("records.WithMedias: %s %w", lessonID, err)
	}
	stepState.RecordedVideos = records

	return StepStateToContext(ctx, stepState), checkIsValidRecordedVideos(records, lessonID, "lesson_id")
}

func (s *suite) recordedVideosAreSavedInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	channelID := stepState.CurrentChannelID

	records, err := new(liveroom_repo.LiveRoomRecordedVideosRepo).GetLiveRoomRecordingsByChannelIDs(ctx, s.LessonmgmtDBTrace, []string{channelID})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("RecordedVideoRepo.ListByIDs: %v", err)
	}
	if len(records) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not find recorded video with channel ID %s", channelID)
	}

	// get media to fill video id field
	medias, err := new(lesson_media_infrastructure.MediaRepo).RetrieveMediasByIDs(ctx, s.BobDB, records.GetMediaIDs())
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("MediaRepo.RetrieveMediasByIDs: %v", err)
	}
	if err = records.WithMedias(medias); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("records.WithMedias: %s %w", channelID, err)
	}
	stepState.RecordedVideos = records

	return StepStateToContext(ctx, stepState), checkIsValidRecordedVideos(records, channelID, "channel_id")
}

func checkIsValidRecordedVideos(rs domain.RecordedVideos, referenceID string, reference string) error {
	for _, v := range rs {
		if v.RecordingChannelID != referenceID {
			return fmt.Errorf("expected %s %s but got %s", reference, referenceID, v.RecordingChannelID)
		}
		if v.Creator == "" {
			return fmt.Errorf("expected creator not empty")
		}
		if v.DateTimeRecorded.IsZero() {
			return fmt.Errorf("expected DateTimeRecorded not zero but got %s", v.DateTimeRecorded)
		}
		if v.Media.Type != media_domain.MediaTypeRecordingVideo {
			return fmt.Errorf("expected Media type is %s but got %s", domain.MediaTypeRecordingVideo, v.Media.Type)
		}
		if v.Media.Duration == 0 {
			return fmt.Errorf("expected Duration not zero but got %s", v.Media.Duration)
		}
		if v.Media.FileSizeBytes == 0 {
			return fmt.Errorf("expected FileSizeBytes not zero but got %d", v.Media.FileSizeBytes)
		}
		if v.Media.Resource == "" {
			return fmt.Errorf("expected Resource not empty but got %s", v.Media.Resource)
		}
	}
	return nil
}

func (s *suite) userGetRecordedVideos(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentLessonID == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stepState.CurrentLessonID must be not empty")
	}

	req := &vpb.GetRecordingByLessonIDRequest{
		LessonId: stepState.CurrentLessonID,
		Paging: &cpb.Paging{
			Limit: 2,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: "",
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).GetRecordingByLessonID(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) mustReturnAListRecordedVideo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*vpb.GetRecordingByLessonIDResponse)
	if res.TotalItems > 0 {
		for _, v := range res.GetItems() {
			if v.Id == "" {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected recorded video id not empty")
			}
			if v.StartTime.AsTime().IsZero() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected DateTimeRecorded not zero but got %s", v.StartTime.AsTime())
			}
			if v.Duration.AsDuration() == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected Duration not zero but got %s", v.Duration.AsDuration())
			}
			if v.GetFileSize() == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected FileSizeBytes not zero but got %f", v.GetFileSize())
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDownloadEachRecordedVideo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.RecordedVideos) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stepState.RecordedVideoIDs must be have element")
	}
	prefixURL := "http://minio-infras.emulator.svc.cluster.local:9000/manabie/"
	fileNames := stepState.RecordedVideos.GetResources()
	for i, recordID := range stepState.RecordedVideos.GetRecordIDs() {
		req := &vpb.GetRecordingDownloadLinkByIDRequest{
			RecordingId: recordID,
			Expiry:      durationpb.New(time.Minute * 3),
			FileName:    "es300_20220824_150034.mp4",
		}

		res, err := vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).GetRecordingDownloadLinkByID(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

		if err != nil {
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}

		p := prefixURL + fileNames[i]
		if !strings.HasPrefix(res.GetUrl(), p) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect download url has prefix string: %s, got: %s", p, res.GetUrl())
		}
		if !strings.Contains(res.GetUrl(), req.FileName) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect download url don't contain file name: %s, got: %s", p, res.GetUrl())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
