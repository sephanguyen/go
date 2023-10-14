package controllers

import (
	"context"
	"net/http"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ImageController struct {
	userServiceClient UserServiceClient
	internalClient    InternalClient
	Logger            *zap.Logger
}

type UserServiceClient interface {
	GetBasicProfile(ctx context.Context, req *pb.GetBasicProfileRequest, opts ...grpc.CallOption) (*pb.GetBasicProfileResponse, error)
}

type InternalClient interface {
}

// RegisterImageController register controller
func RegisterImageController(r *gin.RouterGroup, userServiceClient UserServiceClient, internalClient InternalClient, zapLogger *zap.Logger) {
	c := &ImageController{}
	c.userServiceClient = userServiceClient
	c.internalClient = internalClient
	c.Logger = zapLogger
	r.GET("/picture/:user-id", c.GetUserAvatar)
}

func (rcv *ImageController) GetUserAvatar(c *gin.Context) {
	userID := c.Param("user-id")
	ctx := c.Request.Context()
	resp, err := rcv.userServiceClient.GetBasicProfile(ctx, &pb.GetBasicProfileRequest{
		UserIds: []string{userID},
	})
	if err != nil {
		rcv.Logger.Error("error when calling bob", zap.Error(err))
		c.Status(http.StatusNotFound)
		return
	}

	if len(resp.Profiles) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	avatar := resp.Profiles[0].Avatar
	if avatar == "" {
		c.Status(http.StatusNotFound)
		return
	}

	c.Redirect(http.StatusFound, avatar)
}
