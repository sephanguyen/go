package http

import (
	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/service"

	"github.com/gin-gonic/gin"
)

type UserService struct {
	UserDomainService service.User
}

func (service *UserService) CreateUser(c *gin.Context) {
	responseHandler(service.upsertUser)(c)
}

func (service *UserService) GetUser(c *gin.Context) {
	responseHandler(service.getUser)(c)
}
