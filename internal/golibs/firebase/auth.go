package firebase

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"github.com/pkg/errors"
)

const MinimumPasswordLength = 6

var ErrUserNotExists = errors.New("firebase user is not exists")

type AuthClient interface {
	ImportUsers(ctx context.Context, users []*auth.UserToImport, opts ...auth.UserImportOption) (*auth.UserImportResult, error)
	GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error)
	GetUser(ctx context.Context, uid string) (*auth.UserRecord, error)
	UpdateUser(ctx context.Context, uid string, user *auth.UserToUpdate) (ur *auth.UserRecord, err error)
}

type authClient struct {
	client *auth.Client
}

func NewAuthFromApp(firebaseAuth *auth.Client) AuthClient {
	a := &authClient{
		client: firebaseAuth,
	}
	return a
}

func (a *authClient) ImportUsers(ctx context.Context, users []*auth.UserToImport, opts ...auth.UserImportOption) (*auth.UserImportResult, error) {
	return a.client.ImportUsers(ctx, users, opts...)
}

func (a *authClient) GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error) {
	return a.client.GetUserByEmail(ctx, email)
}

func (a *authClient) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	userRecord, err := a.client.GetUser(ctx, uid)
	if err != nil {
		if auth.IsUserNotFound(err) {
			err = ErrUserNotExists
		}
		return nil, err
	}

	return userRecord, nil
}

func (a *authClient) UpdateUser(ctx context.Context, uid string, user *auth.UserToUpdate) (ur *auth.UserRecord, err error) {
	return a.client.UpdateUser(ctx, uid, user)
}

type AuthUtils interface {
	IsUserNotFound(err error) bool
}

type authUtils struct{}

func (utils *authUtils) IsUserNotFound(err error) bool {
	return auth.IsUserNotFound(err)
}

func NewAuthUtils() AuthUtils {
	return new(authUtils)
}
