package services

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
)

// SubmitAssignment mock handler
// / deprecated
func (s *StudentService) SubmitAssignment(ctx context.Context, req *pb.SubmitAssignmentRequest) (*pb.SubmitAssignmentResponse, error) {
	return nil, nil
}

func submissionsToAssignmentSubmissionProto(subs []*entities.StudentSubmission) []*pb.AssignmentSubmission {
	results := make([]*pb.AssignmentSubmission, 0, len(subs))
	for _, s := range subs {
		results = append(results, submissionEntToAssignmentSubmissionProto(s))
	}

	return results
}

func submissionEntToAssignmentSubmissionProto(e *entities.StudentSubmission) *pb.AssignmentSubmission {
	p := &pb.AssignmentSubmission{
		SubmissionId: e.ID.String,
		TopicId:      e.TopicID.String,
		StudentId:    e.StudentID.String,
		Content:      e.Content.String,
	}
	createdAt, _ := types.TimestampProto(e.CreatedAt.Time)
	p.CreatedAt = createdAt
	p.Attachments = make([]*pb.Attachment, 0, len(e.AttachmentNames.Elements))
	for i, name := range e.AttachmentNames.Elements {
		p.Attachments = append(p.Attachments, &pb.Attachment{
			Name: name.String,
			Url:  e.AttachmentURLs.Elements[i].String,
		})
	}
	return p
}
