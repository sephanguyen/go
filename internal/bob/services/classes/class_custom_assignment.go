package classes

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func isCustomAssignmentRequestValid(req *pb.CreateCustomAssignmentRequest) bool {
	switch {
	case req.CopiedTopicId != nil:
		return true
	case req.Name != nil && req.Grade != nil && req.Country != pb.COUNTRY_NONE && req.Subject != pb.SUBJECT_NONE && len(req.Attachments) > 0:
		return true
	default:
		return false
	}
}

// CreateCustomAssignment creates a custom topic used for creating an assignment.
func (s *ClassService) CreateCustomAssignment(ctx context.Context, req *pb.CreateCustomAssignmentRequest) (*pb.CreateCustomAssignmentResponse, error) {
	return &pb.CreateCustomAssignmentResponse{}, nil
}

// MarkTheSubmissions mock handler
// deprecated
func (s *ClassService) MarkTheSubmissions(ctx context.Context, req *pb.MarkTheSubmissionsRequest) (*pb.MarkTheSubmissionsResponse, error) {
	return nil, nil
}

// ListSubmissions returns list of submissions by student
func (s *ClassService) ListSubmissions(ctx context.Context, req *pb.ListSubmissionsRequest) (*pb.ListSubmissionsResponse, error) {
	return &pb.ListSubmissionsResponse{}, nil
}

func (s *ClassService) RetrieveScore(ctx context.Context, req *pb.RetrieveScoreRequest) (*pb.RetrieveScoreResponse, error) {
	return &pb.RetrieveScoreResponse{}, nil
}

func toComments(src pgtype.JSONB) ([]*pb.Comment, error) {
	var comments []entities.Comment
	err := src.AssignTo(&comments)
	if err != nil {
		return nil, err
	}
	dst := make([]*pb.Comment, 0, len(comments))
	for _, comment := range comments {
		duration := time.Duration(comment.Duration) * time.Second
		dst = append(dst, &pb.Comment{
			Comment:  comment.Comment,
			Duration: types.DurationProto(duration),
		})
	}
	return dst, nil
}

func toMediaPb(src *entities.Media) (*pb.Media, error) {
	createdAt, err := types.TimestampProto(src.CreatedAt.Time)
	if err != nil {
		return nil, err
	}

	updatedAt, err := types.TimestampProto(src.UpdatedAt.Time)
	if err != nil {
		return nil, err
	}

	comments, err := toComments(src.Comments)
	if err != nil {
		return nil, err
	}

	var convertedImages []*entities.ConvertedImage
	if err := src.ConvertedImages.AssignTo(&convertedImages); err != nil {
		return nil, err
	}

	var pbImages []*pb.ConvertedImage
	for _, c := range convertedImages {
		pbImages = append(pbImages, &pb.ConvertedImage{
			Width:    c.Width,
			Height:   c.Height,
			ImageUrl: c.ImageURL,
		})
	}

	return &pb.Media{
		MediaId:   src.MediaID.String,
		Name:      src.Name.String,
		Resource:  src.Resource.String,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Comments:  comments,
		Type:      pb.MediaType(pb.MediaType_value[src.Type.String]),
		Images:    pbImages,
	}, nil
}

func (s *ClassService) RetrieveMedia(ctx context.Context, req *pb.RetrieveMediaRequest) (*pb.RetrieveMediaResponse, error) {
	medias, err := s.MediaRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(req.MediaIds))
	if err != nil {
		return nil, services.ToStatusError(err)
	}
	result := make([]*pb.Media, 0, len(medias))
	for _, media := range medias {
		pbMedia, err := toMediaPb(media)
		if err != nil {
			return nil, fmt.Errorf("error convert: %w", err)
		}
		result = append(result, pbMedia)
	}
	return &pb.RetrieveMediaResponse{
		Media: result,
	}, nil
}

func commentToEn(comments []*pb.Comment) []*entities.Comment {
	result := make([]*entities.Comment, 0, len(comments))
	for _, comment := range comments {
		c := &entities.Comment{
			Comment:  comment.Comment,
			Duration: comment.Duration.GetSeconds(),
		}
		result = append(result, c)
	}
	return result
}

func toMediaEn(src *pb.Media) (*entities.Media, error) {
	dst := &entities.Media{}
	database.AllNullEntity(dst)
	if src.MediaId == "" {
		src.MediaId = idutil.ULIDNow()
	}
	comments := commentToEn(src.Comments)

	err := multierr.Combine(
		dst.MediaID.Set(src.MediaId),
		dst.Resource.Set(src.Resource),
		dst.Name.Set(src.Name),
		dst.Comments.Set(comments),
		dst.Type.Set(src.Type.String()),
	)
	if src.CreatedAt != nil {
		multierr.Append(err, dst.CreatedAt.Set(time.Unix(src.CreatedAt.Seconds, int64(src.CreatedAt.Nanos))))
	} else {
		multierr.Append(err, dst.CreatedAt.Set(time.Now()))
	}

	if src.UpdatedAt != nil {
		multierr.Append(err, dst.UpdatedAt.Set(time.Unix(src.UpdatedAt.Seconds, int64(src.UpdatedAt.Nanos))))
	} else {
		multierr.Append(err, dst.UpdatedAt.Set(time.Now()))
	}
	return dst, err
}

func (s *ClassService) UpsertMedia(ctx context.Context, req *pb.UpsertMediaRequest) (*pb.UpsertMediaResponse, error) {
	eMedia := make([]*entities.Media, 0, len(req.Media))
	for _, media := range req.Media {
		em, err := toMediaEn(media)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		eMedia = append(eMedia, em)
	}
	err := s.MediaRepo.UpsertMediaBatch(ctx, s.DB, eMedia)
	if err != nil {
		return nil, services.ToStatusError(err)
	}
	mediaIDs := make([]string, 0, len(req.Media))
	for _, m := range eMedia {
		mediaIDs = append(mediaIDs, m.MediaID.String)
	}
	return &pb.UpsertMediaResponse{
		MediaIds: mediaIDs,
	}, nil
}
