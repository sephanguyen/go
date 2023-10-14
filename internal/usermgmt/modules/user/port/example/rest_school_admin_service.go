// Package example demonstrates how the User-side will be implemented
package example

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

// SchoolAdminHTTPService represents an Application Service
type SchoolAdminHTTPService struct {
	SchoolAdminService SchoolAdminDomainService
}

type SchoolAdminDomainService interface {
	CreateSchoolAdmin(ctx context.Context, schoolAdminToCreate entity.DomainSchoolAdminProfile) error
}

func (service *SchoolAdminHTTPService) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		data, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		createSchoolAdminRequest := &RESTCreateSchoolAdminRequest{}
		if err := json.Unmarshal(data, createSchoolAdminRequest); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = service.SchoolAdminService.CreateSchoolAdmin(request.Context(), createSchoolAdminRequest)
		if err != nil {
			return
		}
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
