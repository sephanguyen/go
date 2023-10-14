package transport

import (
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"

	"google.golang.org/grpc/codes"
)

var GrpcErrorMap = map[errors.ErrorKey]codes.Code{
	errors.ErrEntityNotFound:  codes.NotFound,
	errors.ErrConversion:      codes.Internal,
	errors.ErrInputValidation: codes.InvalidArgument,
}
