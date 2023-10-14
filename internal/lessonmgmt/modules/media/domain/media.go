package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MediaType string

const (
	MediaTypeNone           MediaType = "MEDIA_TYPE_NONE"
	MediaTypeVideo          MediaType = "MEDIA_TYPE_VIDEO"
	MediaTypeImage          MediaType = "MEDIA_TYPE_IMAGE"
	MediaTypePDF            MediaType = "MEDIA_TYPE_PDF"
	MediaTypeAudio          MediaType = "MEDIA_TYPE_AUDIO"
	MediaTypeRecordingVideo MediaType = "MEDIA_TYPE_RECORDING_VIDEO"
)

var (
	MediaTypeToProto = map[MediaType]lpb.MediaType{
		MediaTypeNone:           lpb.MediaType_MEDIA_TYPE_NONE,
		MediaTypeVideo:          lpb.MediaType_MEDIA_TYPE_VIDEO,
		MediaTypeImage:          lpb.MediaType_MEDIA_TYPE_IMAGE,
		MediaTypePDF:            lpb.MediaType_MEDIA_TYPE_PDF,
		MediaTypeAudio:          lpb.MediaType_MEDIA_TYPE_AUDIO,
		MediaTypeRecordingVideo: lpb.MediaType_MEDIA_TYPE_RECORDING_VIDEO,
	}

	ProtoToMediaType = map[lpb.MediaType]MediaType{
		lpb.MediaType_MEDIA_TYPE_NONE:            MediaTypeNone,
		lpb.MediaType_MEDIA_TYPE_VIDEO:           MediaTypeVideo,
		lpb.MediaType_MEDIA_TYPE_IMAGE:           MediaTypeImage,
		lpb.MediaType_MEDIA_TYPE_PDF:             MediaTypePDF,
		lpb.MediaType_MEDIA_TYPE_AUDIO:           MediaTypeAudio,
		lpb.MediaType_MEDIA_TYPE_RECORDING_VIDEO: MediaTypeRecordingVideo,
	}
)

type Media struct {
	ID              string
	Name            string
	Resource        string
	Type            MediaType
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Comments        []Comment
	ConvertedImages []ConvertedImage
	FileSizeBytes   int64
	Duration        time.Duration
}

func (m *Media) IsValid() error {
	if len(m.ID) == 0 {
		return fmt.Errorf("Media.ID could not be empty")
	}
	if len(m.Resource) == 0 {
		return fmt.Errorf("Media.Resource could not be empty")
	}
	if len(m.Type) == 0 {
		return fmt.Errorf("Media.Type could not be empty")
	}

	return nil
}

func (m *Media) PreCreate() {
	m.ID = idutil.ULIDNow()
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
}

type ConvertedImage struct {
	Width    int32
	Height   int32
	ImageURL string
}

type Comment struct {
	Comment  string
	Duration int64
}

type Medias []*Media

func (m Medias) GetMediaIDs() []string {
	mediaIDs := make([]string, 0, len(m))

	for _, media := range m {
		mediaIDs = append(mediaIDs, media.ID)
	}
	return mediaIDs
}

func (m Medias) ToRetrieveMediasByIDsResponse() *lpb.RetrieveMediasByIDsResponse {
	medias := make([]*lpb.Media, 0, len(m))

	for _, value := range m {
		var images []*lpb.ConvertedImage
		for _, image := range value.ConvertedImages {
			images = append(images, &lpb.ConvertedImage{
				ImageUrl: image.ImageURL,
				Width:    image.Width,
				Height:   image.Height,
			})
		}

		var comments []*lpb.Comment
		for _, comment := range value.Comments {
			comments = append(comments, &lpb.Comment{
				Comment: comment.Comment,
			})
		}

		media := &lpb.Media{
			MediaId:  value.ID,
			Name:     value.Name,
			Type:     MediaTypeToProto[value.Type],
			Resource: value.Resource,
			Comments: comments,
			Images:   images,
		}
		medias = append(medias, media)
	}
	return &lpb.RetrieveMediasByIDsResponse{Medias: medias}
}

func FromMediaProtoToMediaDomain(reqMedia *lpb.Media) (*Media, error) {
	media := &Media{
		ID:            reqMedia.MediaId,
		Name:          reqMedia.Name,
		Resource:      reqMedia.Resource,
		Type:          ProtoToMediaType[reqMedia.Type],
		FileSizeBytes: reqMedia.FileSizeBytes,
		Duration:      reqMedia.Duration.AsDuration(),
		CreatedAt:     reqMedia.CreatedAt.AsTime(),
		UpdatedAt:     reqMedia.UpdatedAt.AsTime(),
	}

	media.ConvertedImages = make([]ConvertedImage, 0, len(reqMedia.Images))
	media.Comments = make([]Comment, 0, len(reqMedia.Comments))

	for _, image := range reqMedia.GetImages() {
		media.ConvertedImages = append(media.ConvertedImages, ConvertedImage{
			ImageURL: image.GetImageUrl(),
			Width:    image.GetWidth(),
			Height:   image.GetHeight(),
		})
	}

	for _, comment := range reqMedia.GetComments() {
		media.Comments = append(media.Comments, Comment{
			Comment: comment.GetComment(),
		})
	}

	if err := media.IsValid(); err != nil {
		return nil, err
	}

	return media, nil
}

func FromMediaDomainToMediaProto(reqMedia *Media) *lpb.Media {
	media := &lpb.Media{
		MediaId:       reqMedia.ID,
		Name:          reqMedia.Name,
		Resource:      reqMedia.Resource,
		Type:          MediaTypeToProto[reqMedia.Type],
		FileSizeBytes: reqMedia.FileSizeBytes,
		Duration:      durationpb.New(reqMedia.Duration),
		CreatedAt:     timestamppb.New(reqMedia.CreatedAt),
		UpdatedAt:     timestamppb.New(reqMedia.UpdatedAt),
	}

	media.Images = make([]*lpb.ConvertedImage, 0, len(reqMedia.ConvertedImages))
	media.Comments = make([]*lpb.Comment, 0, len(reqMedia.Comments))

	for _, image := range reqMedia.ConvertedImages {
		media.Images = append(media.Images, &lpb.ConvertedImage{
			ImageUrl: image.ImageURL,
			Width:    image.Width,
			Height:   image.Height,
		})
	}

	for _, comment := range reqMedia.Comments {
		media.Comments = append(media.Comments, &lpb.Comment{
			Comment: comment.Comment,
		})
	}

	return media
}

func FromRetrieveMediasByIDsResponse(res *lpb.RetrieveMediasByIDsResponse) Medias {
	medias := Medias{}
	for _, value := range res.GetMedias() {
		var images []ConvertedImage
		for _, image := range value.GetImages() {
			images = append(images, ConvertedImage{
				ImageURL: image.GetImageUrl(),
				Width:    image.GetWidth(),
				Height:   image.GetHeight(),
			})
		}

		var comments []Comment
		for _, comment := range value.GetComments() {
			comments = append(comments, Comment{
				Comment: comment.GetComment(),
			})
		}

		medias = append(medias, &Media{
			ID:              value.GetMediaId(),
			Name:            value.GetName(),
			Type:            ProtoToMediaType[value.GetType()],
			Resource:        value.GetResource(),
			Comments:        comments,
			ConvertedImages: images,
		})
	}
	return medias
}
