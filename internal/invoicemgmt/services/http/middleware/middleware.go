package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	golibs_interceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	invoicemgmt_http "github.com/manabie-com/backend/internal/invoicemgmt/services/http"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/http/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	SignatureHeader = "manabie-signature"
	PublicKeyHeader = "manabie-public-key"
)

var rbacDecider = map[string][]string{
	constant.HealthCheckStatusEndpoint: nil,
	constant.StudentBankInfoEndpoint:   {constant.RoleOpenAPI},
}

func NewGroupDecider(db database.QueryExecer) *interceptors.GroupDecider {
	return &interceptors.GroupDecider{
		GroupFetcher: func(ctx context.Context, userID string) ([]string, error) {
			userRepo := &repository.UserRepo{}
			return interceptors.RetrieveUserRoles(ctx, userRepo, db)
		},
		AllowedGroups: rbacDecider,
	}
}

func VerifySignature(logger *zap.Logger, groupDecider *interceptors.GroupDecider, client spb.TokenReaderServiceClient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		buf, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logger.Sugar().Errorf("io.ReadAll err: %v", err)
			invoicemgmt_http.ResponseError(ctx, err)
			return
		}
		resp, err := client.VerifySignature(ctx, &spb.VerifySignatureRequest{
			PublicKey: ctx.GetHeader(PublicKeyHeader),
			Signature: ctx.GetHeader(SignatureHeader),
			Body:      buf,
		})
		if err != nil {
			logger.Sugar().Errorf("client.VerifySignature err: %v", err)
			s := status.Convert(err)
			returnError := errcode.Error{
				Err:  err,
				Code: errcode.InternalError,
			}

			if s.Code() == codes.PermissionDenied {
				switch s.Message() {
				case errorx.ErrShamirInvalidPublicKey.Error():
					returnError.Code = errcode.InvalidPublicKey
				case errorx.ErrShamirInvalidSignature.Error():
					returnError.Code = errcode.InvalidSignature
				default:
					returnError.Code = errcode.PermissionDenied
				}
			}

			invoicemgmt_http.ResponseError(ctx, returnError)
			return
		}

		claims := utils.ManabieUserCustomClaims(
			cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
			resp.UserId,
			resp.OrganizationId,
		)

		requestCtx := golibs_interceptors.ContextWithJWTClaims(ctx.Request.Context(), claims)
		requestCtx = golibs_interceptors.ContextWithUserID(requestCtx, resp.UserId)

		_, err = groupDecider.Check(requestCtx, resp.UserId, ctx.FullPath())
		if err != nil {
			logger.Sugar().Errorf("groupDecider.Check err: %v", err)
			ctx.AbortWithStatusJSON(http.StatusForbidden, invoicemgmt_http.Response{
				Message: err.Error(),
			})
			return
		}

		ctx.Request = ctx.Request.WithContext(requestCtx)
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(buf))

		ctx.Next()
	}
}
