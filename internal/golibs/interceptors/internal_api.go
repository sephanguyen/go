package interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

type FakeJwtContext struct {
	methods       map[string]struct{}
	simulatedRole string
}

func NewFakeJwtContext(methods map[string]struct{}, simulateRole string) *FakeJwtContext {
	return &FakeJwtContext{
		methods:       methods,
		simulatedRole: simulateRole,
	}
}

// nolint: revive
type OrgIDGetter interface {
	GetOrganizationId() string
}

// nolint: revive
type CurrentUserIDGetter interface {
	GetCurrentUserId() string
}

func (a *FakeJwtContext) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if _, exist := a.methods[info.FullMethod]; !exist {
		return handler(ctx, req)
	}
	resourcePathGetter, ok := req.(OrgIDGetter)
	if !ok {
		return nil, fmt.Errorf("type %T does not impl OrgIDGetter", req)
	}
	fakeClaim := &CustomClaims{
		Manabie: &ManabieClaims{
			ResourcePath: resourcePathGetter.GetOrganizationId(),
			UserGroup:    a.simulatedRole,
			SchoolIDs:    []string{resourcePathGetter.GetOrganizationId()},
		},
	}

	// If the request implements the CurrentUserIDGetter interface, add the userID to claims
	if userIDGetter, ok := req.(CurrentUserIDGetter); ok {
		fakeClaim.Manabie.UserID = userIDGetter.GetCurrentUserId()
	}

	ctx = ContextWithJWTClaims(ctx, fakeClaim)

	return handler(ctx, req)
}
