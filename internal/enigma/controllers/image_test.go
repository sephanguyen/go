package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mock_controllers "github.com/manabie-com/backend/mock/enigma/controllers"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestImageController_GetUserAvatar(t *testing.T) {
	t.Parallel()
	userClient := &mock_controllers.UserServiceClient{}
	internalClient := &mock_controllers.InternalClient{}

	t.Run("err when calling bob", func(t *testing.T) {
		resp := httptest.NewRecorder()
		gin.SetMode(gin.TestMode)
		ctx, r := gin.CreateTestContext(resp)
		RegisterImageController(r.Group("image"), userClient, internalClient, zap.NewNop())

		userClient.On("GetBasicProfile", mock.Anything, &pb.GetBasicProfileRequest{
			UserIds: []string{"invalid-user-id"},
		}).Once().Return(nil, errors.New("connection close"))

		ctx.Request, _ = http.NewRequest(http.MethodGet, "/image/picture/invalid-user-id", nil)
		r.ServeHTTP(resp, ctx.Request)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("not found user", func(t *testing.T) {
		resp := httptest.NewRecorder()
		gin.SetMode(gin.TestMode)
		ctx, r := gin.CreateTestContext(resp)
		RegisterImageController(r.Group("image"), userClient, internalClient, zap.NewNop())

		userClient.On("GetBasicProfile", mock.Anything, &pb.GetBasicProfileRequest{
			UserIds: []string{"valid-user-id"},
		}).Once().Return(&pb.GetBasicProfileResponse{}, nil)

		ctx.Request, _ = http.NewRequest(http.MethodGet, "/image/picture/valid-user-id", nil)
		r.ServeHTTP(resp, ctx.Request)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("found user, but null avatar", func(t *testing.T) {
		resp := httptest.NewRecorder()
		gin.SetMode(gin.TestMode)
		ctx, r := gin.CreateTestContext(resp)
		RegisterImageController(r.Group("image"), userClient, internalClient, zap.NewNop())

		userClient.On("GetBasicProfile", mock.Anything, &pb.GetBasicProfileRequest{
			UserIds: []string{"valid-user-id"},
		}).Once().Return(&pb.GetBasicProfileResponse{
			Profiles: []*pb.BasicProfile{
				{Avatar: ""},
			},
		}, nil)

		ctx.Request, _ = http.NewRequest(http.MethodGet, "/image/picture/valid-user-id", nil)
		r.ServeHTTP(resp, ctx.Request)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Success", func(t *testing.T) {
		resp := httptest.NewRecorder()
		gin.SetMode(gin.TestMode)
		ctx, r := gin.CreateTestContext(resp)
		RegisterImageController(r.Group("image"), userClient, internalClient, zap.NewNop())

		userClient.On("GetBasicProfile", mock.Anything, &pb.GetBasicProfileRequest{
			UserIds: []string{"valid-user-id"},
		}).Once().Return(&pb.GetBasicProfileResponse{
			Profiles: []*pb.BasicProfile{
				{Avatar: "http://example.com"},
			},
		}, nil)

		ctx.Request, _ = http.NewRequest(http.MethodGet, "/image/picture/valid-user-id", nil)
		r.ServeHTTP(resp, ctx.Request)
		assert.Equal(t, "http://example.com", resp.Header().Get("Location"))
	})
}
