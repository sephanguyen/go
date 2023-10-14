package users

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"firebase.google.com/go/v4/auth"
	"github.com/jackc/pgtype"
	"google.golang.org/grpc"
)

type UserMgmtModifierSvc interface {
	UpdateUserProfile(context.Context, *upb.UpdateUserProfileRequest, ...grpc.CallOption) (*upb.UpdateUserProfileResponse, error)
}

type UserMgmtAuthSvc interface {
	ExchangeCustomToken(context.Context, *upb.ExchangeCustomTokenRequest, ...grpc.CallOption) (*upb.ExchangeCustomTokenResponse, error)
}

// UserModifierService implements core business logic
type UserModifierService struct {
	bpb.UnimplementedUserModifierServiceServer
	JSM                 nats.JetStreamManagement
	FirebaseClient      *auth.Client
	TenantManager       multitenant.TenantManager
	ApplicantID         string
	DB                  database.Ext
	UserRepo            repositories.UserRepository
	UserMgmtModifierSvc UserMgmtModifierSvc
	UserMgmtAuthSvc     UserMgmtAuthSvc
	UserGroupRepo       interface {
		Find(context.Context, database.QueryExecer, pgtype.Text) ([]*entities.UserGroup, error)
	}

	SchoolAdminRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error)
	}

	ShamirClient spb.TokenReaderServiceClient

	TeacherRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Teacher, error)
	}
}
