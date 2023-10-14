package domain

import (
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type PortForwardClassDoRequest struct {
	ClassDoID string
	Body      string
}

func (r *PortForwardClassDoRequest) FromProto(proto *lpb.PortForwardClassDoRequest) {
	r.ClassDoID = proto.GetClassDoId()
	r.Body = proto.GetBody()
}

type PortForwardClassDoResponse struct {
	Response string
}

func (r *PortForwardClassDoResponse) ToProto() *lpb.PortForwardClassDoResponse {
	return &lpb.PortForwardClassDoResponse{
		Response: r.Response,
	}
}
