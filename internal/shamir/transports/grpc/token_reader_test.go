package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/shamir/services"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_features "github.com/manabie-com/backend/mock/usermgmt/config"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/square/go-jose/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockVerifier struct {
	exchangeVerifiedToken func(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error)
	exchangeTokenFn       func(ctx context.Context, originalToken string, newTokenInfo *interceptors.TokenInfo) (string, error)
	verifyFn              func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error)
}

func (m *mockVerifier) ExchangeVerifiedToken(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error) {
	return m.exchangeVerifiedToken(claims, newTokenInfo)
}

func (m *mockVerifier) ExchangeToken(ctx context.Context, originalToken string, newTokenInfo *interceptors.TokenInfo) (string, error) {
	return m.exchangeTokenFn(ctx, originalToken, newTokenInfo)
}

func (m *mockVerifier) Verify(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
	return m.verifyFn(ctx, originalToken)
}

type mockUserRepoV2 struct {
	getFn         func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error)
	getByUsername func(ctx context.Context, db database.QueryExecer, username string, organizationID string) (*entity.AuthUser, error)
	getByEmail    func(ctx context.Context, db database.QueryExecer, username string, organizationID string) (*entity.AuthUser, error)
}

func (m *mockUserRepoV2) GetByAuthInfo(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
	return m.getFn(ctx, db, defaultOrganizationAuthValues, userID, projectID, tenantID)
}

func (m *mockUserRepoV2) GetByAuthInfoV2(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.AuthUser, error) {
	user, err := m.getFn(ctx, db, defaultOrganizationAuthValues, userID, projectID, tenantID)
	if err != nil {
		return nil, err
	}

	authUser := &entity.AuthUser{
		UserID:       user.UserID,
		UserGroup:    user.Group,
		ResourcePath: user.ResourcePath,
	}
	return authUser, nil
}

func (m *mockUserRepoV2) GetByUsername(ctx context.Context, db database.QueryExecer, username string, organizationID string) (*entity.AuthUser, error) {
	return m.getByUsername(ctx, db, username, organizationID)
}

func (m *mockUserRepoV2) GetByEmail(ctx context.Context, db database.QueryExecer, username string, organizationID string) (*entity.AuthUser, error) {
	return m.getByEmail(ctx, db, username, organizationID)
}

func TestService_ExchangeToken(t *testing.T) {
	t.Parallel()

	t.Run("verify token error", func(tt *testing.T) {
		tt.Parallel()

		db := new(mock_database.Ext)

		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				exchangeVerifiedToken: func(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error) {
					return "", services.ErrWrongDivision
				},
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, fmt.Errorf("verify token error")
				},
			},
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupTeacher),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
			DB: db,
		}

		resp, err := s.ExchangeToken(context.Background(), &spb.ExchangeTokenRequest{
			OriginalToken: "OriginalToken",
			NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
				Applicant: "manabie-local",
				UserId:    "user-id",
			},
		})
		assert.Error(tt, err)
		assert.Equal(tt, codes.Unknown, status.Code(err))
		assert.Nil(tt, resp)
	})

	t.Run("exchange verified token error", func(tt *testing.T) {
		tt.Parallel()
		mErr := fmt.Errorf("new err")

		db := new(mock_database.Ext)

		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				exchangeVerifiedToken: func(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error) {
					return "", mErr
				},
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB: db,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupTeacher),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.ExchangeToken(context.Background(), &spb.ExchangeTokenRequest{
			OriginalToken: "OriginalToken",
			NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
				Applicant: "manabie-local",
				UserId:    "user-id",
			},
		})
		assert.Error(tt, err)
		assert.Equal(tt, codes.Unknown, status.Code(err))
		assert.Nil(tt, resp)
	})

	t.Run("user was deactivated", func(tt *testing.T) {
		tt.Parallel()
		req := &spb.ExchangeTokenRequest{
			OriginalToken: "OriginalToken",
			NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
				Applicant: "manabie-local",
				UserId:    "user-id",
			},
		}

		db := new(mock_database.Ext)

		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				exchangeVerifiedToken: func(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error) {
					return "exchanged-token", nil
				},
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB: db,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					return &entity.LegacyUser{
						ID:            database.Text("user-id"),
						Group:         database.Text(entities.UserGroupTeacher),
						ResourcePath:  database.Text(fmt.Sprint(constants.ManabieSchool)),
						DeactivatedAt: database.Timestamptz(time.Now()),
					}, nil
				},
			},
		}

		_, err := s.ExchangeToken(context.Background(), req)
		assert.Equal(t, err, errorx.ErrDeactivatedUser)
	})

	t.Run("teacher exchange token success", func(tt *testing.T) {
		tt.Parallel()
		req := &spb.ExchangeTokenRequest{
			OriginalToken: "OriginalToken",
			NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
				Applicant: "manabie-local",
				UserId:    "user-id",
			},
		}

		db := new(mock_database.Ext)

		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				exchangeVerifiedToken: func(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error) {
					return "exchanged-token", nil
				},
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB: db,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupTeacher),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.ExchangeToken(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "exchanged-token", resp.NewToken)
	})

	t.Run("school admin exchange token success", func(tt *testing.T) {
		tt.Parallel()
		req := &spb.ExchangeTokenRequest{
			OriginalToken: "OriginalToken",
			NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
				Applicant: "manabie-local",
				UserId:    "user-id",
			},
		}

		bobDB := new(mock_database.Ext)
		authDB := new(mock_database.Ext)

		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				exchangeVerifiedToken: func(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error) {
					return "exchanged-token", nil
				},
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB:     bobDB,
			AuthDB: authDB,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					if db != bobDB {
						return nil, fmt.Errorf("invalid bob db connection")
					}
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupSchoolAdmin),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.ExchangeToken(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "exchanged-token", resp.NewToken)
	})

	t.Run("user exchange token success with auth DB", func(tt *testing.T) {
		tt.Parallel()
		req := &spb.ExchangeTokenRequest{
			OriginalToken: "OriginalToken",
			NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
				Applicant: "manabie-local",
				UserId:    "user-id",
			},
		}

		db := new(mock_database.Ext)
		authDB := new(mock_database.Ext)

		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(true, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				exchangeVerifiedToken: func(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error) {
					return "exchanged-token", nil
				},
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB:     db,
			AuthDB: authDB,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					if db != authDB {
						return nil, fmt.Errorf("invalid auth db connection")
					}
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupSchoolAdmin),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.ExchangeToken(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "exchanged-token", resp.NewToken)
	})
}

func TestService_Verify(t *testing.T) {
	t.Parallel()
	t.Run("verify token error", func(tt *testing.T) {
		tt.Parallel()
		mErr := fmt.Errorf("new err")
		req := &spb.VerifyTokenRequest{
			OriginalToken: "OriginalToken-xxx",
		}
		m := &mockVerifier{
			verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
				if originalToken != req.OriginalToken {
					tt.Error("unexpected token input", originalToken)
				}

				return nil, mErr
			},
		}

		s := &Service{
			Verifier: m,
		}

		resp, err := s.VerifyToken(context.Background(), req)
		assert.Error(tt, err)
		assert.Nil(tt, resp)
	})

	t.Run("verify token success", func(tt *testing.T) {
		tt.Parallel()
		req := &spb.VerifyTokenRequest{
			OriginalToken: "OriginalToken-xxx",
		}
		m := &mockVerifier{
			verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
				if originalToken != req.OriginalToken {
					tt.Error("unexpected token input", originalToken)
				}

				return &interceptors.CustomClaims{
					Claims: jwt.Claims{
						Subject: "expected-subject",
					},
				}, nil
			},
		}

		s := &Service{
			Verifier: m,
		}

		resp, err := s.VerifyToken(context.Background(), req)
		assert.NoError(tt, err)
		assert.Equal(tt, "expected-subject", resp.UserId)
	})
}

func TestService_VerifyTokenV2(t *testing.T) {
	t.Parallel()

	t.Run("user verify token success with auth db", func(tt *testing.T) {
		db := new(mock_database.Ext)
		authDB := new(mock_database.Ext)

		req := &spb.VerifyTokenRequest{
			OriginalToken: "OriginalToken-xxx",
		}
		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(true, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB:     db,
			AuthDB: authDB,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					if db != authDB {
						return nil, fmt.Errorf("invalid auth db connection")
					}
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupStudent),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.VerifyTokenV2(context.Background(), req)
		assert.NoError(tt, err)
		assert.Equal(tt, "user-id", resp.UserId)
	})

	t.Run("student verify token success", func(tt *testing.T) {
		bobDB := new(mock_database.Ext)
		authDB := new(mock_database.Ext)

		req := &spb.VerifyTokenRequest{
			OriginalToken: "OriginalToken-xxx",
		}
		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB:     bobDB,
			AuthDB: authDB,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					if db != bobDB {
						return nil, fmt.Errorf("invalid bob db connection")
					}
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupStudent),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.VerifyTokenV2(context.Background(), req)
		assert.NoError(tt, err)
		assert.Equal(tt, "user-id", resp.UserId)
	})

	t.Run("parent verify token success", func(tt *testing.T) {
		db := new(mock_database.Ext)

		req := &spb.VerifyTokenRequest{
			OriginalToken: "OriginalToken-xxx",
		}
		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB: db,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupParent),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.VerifyTokenV2(context.Background(), req)
		assert.NoError(tt, err)
		assert.Equal(tt, "user-id", resp.UserId)
	})

	t.Run("staff exchange custom token failed", func(tt *testing.T) {
		db := new(mock_database.Ext)

		req := &spb.VerifyTokenRequest{
			OriginalToken: "OriginalToken-xxx",
		}
		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, nil
				},
			},
			DB: db,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupSchoolAdmin),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.VerifyTokenV2(context.Background(), req)
		assert.Error(tt, err)
		assert.Equal(tt, codes.Unknown, status.Code(err))
		assert.Nil(tt, resp)
	})

	t.Run("verify token failed", func(tt *testing.T) {
		db := new(mock_database.Ext)

		req := &spb.VerifyTokenRequest{
			OriginalToken: "OriginalToken-xxx",
		}
		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabledOnOrganization", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient: unleashClient,
			Verifier: &mockVerifier{
				verifyFn: func(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
					return &interceptors.CustomClaims{
						Claims: jwt.Claims{
							Subject: "user-id",
						},
					}, fmt.Errorf("verify token failed")
				},
			},
			DB: db,
			UserRepoV2: &mockUserRepoV2{
				getFn: func(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
					return &entity.LegacyUser{
						ID:           database.Text("user-id"),
						Group:        database.Text(entities.UserGroupSchoolAdmin),
						ResourcePath: database.Text(fmt.Sprint(constants.ManabieSchool)),
					}, nil
				},
			},
		}

		resp, err := s.VerifyTokenV2(context.Background(), req)
		assert.Error(tt, err)
		assert.Equal(tt, codes.Unknown, status.Code(err))
		assert.Nil(tt, resp)
	})
}

func TestService_VerifySignature(t *testing.T) {
	t.Parallel()

	t.Run("verify signature error", func(tt *testing.T) {
		tt.Parallel()
		req := &spb.VerifySignatureRequest{
			Signature: "invalid-signature",
		}

		domainAPIKeypairRepo := new(mock_repositories.MockDomainAPIKeypairRepo)
		domainAPIKeypairRepo.On("GetByPublicKey", mock.Anything, mock.Anything, req.PublicKey).Once().Return(&repository.APIKeyPair{}, nil)
		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabled", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything).Return(false, nil)

		s := &Service{
			UnleashClient:        unleashClient,
			DomainAPIKeypairRepo: domainAPIKeypairRepo,
		}

		resp, err := s.VerifySignature(context.Background(), req)
		assert.Error(tt, err)
		assert.Nil(tt, resp)
	})
	t.Run("happy case", func(tt *testing.T) {
		tt.Parallel()
		body := "custom-body"
		apiKeypair := &repository.APIKeyPair{}

		mac := hmac.New(sha256.New, []byte(apiKeypair.PrivateKey().String()))
		_, _ = mac.Write([]byte(body))
		signature := hex.EncodeToString(mac.Sum(nil))
		req := &spb.VerifySignatureRequest{
			Signature: signature,
			Body:      []byte(body),
		}

		db := new(mock_database.Ext)
		authDB := new(mock_database.Ext)

		domainAPIKeypairRepo := new(mock_repositories.MockDomainAPIKeypairRepo)
		domainAPIKeypairRepo.On("GetByPublicKey", mock.Anything, mock.MatchedBy(func(i interface{}) bool {
			return i == db
		}), req.PublicKey).Once().Return(&repository.APIKeyPair{}, nil)
		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabled", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything).Return(false, nil)

		s := &Service{
			DB:                   db,
			AuthDB:               authDB,
			UnleashClient:        unleashClient,
			DomainAPIKeypairRepo: domainAPIKeypairRepo,
		}

		_, err := s.VerifySignature(context.Background(), req)
		assert.Nil(tt, err)
	})

	t.Run("happy case with auth db", func(tt *testing.T) {
		tt.Parallel()
		body := "custom-body"
		apiKeypair := &repository.APIKeyPair{}

		mac := hmac.New(sha256.New, []byte(apiKeypair.PrivateKey().String()))
		_, _ = mac.Write([]byte(body))
		signature := hex.EncodeToString(mac.Sum(nil))
		req := &spb.VerifySignatureRequest{
			Signature: signature,
			Body:      []byte(body),
		}

		db := new(mock_database.Ext)
		authDB := new(mock_database.Ext)

		domainAPIKeypairRepo := new(mock_repositories.MockDomainAPIKeypairRepo)
		domainAPIKeypairRepo.On("GetByPublicKey", mock.Anything, mock.MatchedBy(func(i interface{}) bool {
			return i == authDB
		}), req.PublicKey).Once().Return(&repository.APIKeyPair{}, nil)
		unleashClient := new(mock_unleash_client.UnleashClientInstance)
		unleashClient.On("IsFeatureEnabled", unleash.FeatureDecouplingUserAndAuthDB, mock.Anything).Return(true, nil)

		s := &Service{
			DB:                   db,
			AuthDB:               authDB,
			UnleashClient:        unleashClient,
			DomainAPIKeypairRepo: domainAPIKeypairRepo,
		}

		_, err := s.VerifySignature(context.Background(), req)
		assert.Nil(tt, err)
	})
}

func Test_fakeVerifiedClaimAndNewTokenInfo(t *testing.T) {
	t.Parallel()
	type suite struct {
		env         string
		tenantID    string
		applicantID string
		projectID   string
		schoolID    string
	}
	suites := []suite{
		{
			env:         "local",
			tenantID:    "end-to-end-dopvo",
			applicantID: "manabie-local",
			projectID:   "dev-manabie-online",
		},
		{
			env:         "stag",
			tenantID:    "end-to-end-school-5xn27",
			applicantID: "manabie-stag",
			projectID:   "staging-manabie-online",
		},
		{
			env:         "uat",
			tenantID:    "end-to-end-school-5mqoc",
			applicantID: "manabie-stag",
			projectID:   "uat-manabie",
		},
		{
			env:         "prod",
			tenantID:    "prod-end-to-end-og7nh",
			applicantID: "prod-tokyo",
			projectID:   "student-coach-e1e95",
			schoolID:    fmt.Sprint(constants.E2ETokyo),
		},
		{
			env:         "prod",
			tenantID:    "prod-end-to-end-og7nh",
			applicantID: "prod-aic",
			projectID:   "student-coach-e1e95",
			schoolID:    fmt.Sprint(constants.AICSchool),
		},
		{
			env:         "prod",
			tenantID:    "prod-end-to-end-og7nh",
			applicantID: "prod-ga",
			projectID:   "student-coach-e1e95",
			schoolID:    fmt.Sprint(constants.GASchool),
		},
		{
			env:         "prod",
			tenantID:    "prod-end-to-end-og7nh",
			applicantID: "prod-renseikai",
			projectID:   "student-coach-e1e95",
			schoolID:    fmt.Sprint(constants.RenseikaiSchool),
		},
		{
			env:         "prod",
			tenantID:    "prod-end-to-end-og7nh",
			applicantID: "prod-synersia",
			projectID:   "student-coach-e1e95",
			schoolID:    fmt.Sprint(constants.SynersiaSchool),
		},
	}
	for idx := range suites {
		item := suites[idx]
		t.Run(item.env, func(t *testing.T) {
			claim, newInfo, err := fakeVerifiedClaimAndNewTokenInfo(item.env, &spb.GenerateFakeTokenRequest{
				UserId:    "user_id",
				SchoolId:  item.schoolID,
				ProjectId: item.projectID,
				TenantId:  item.tenantID,
			})
			assert.NoError(t, err)
			assert.Equal(t, claim.FirebaseClaims.Identity.Tenant, item.tenantID)
			assert.Equal(t, claim.Issuer, fmt.Sprintf("https://securetoken.google.com/%s", item.projectID))
			assert.Equal(t, claim.Subject, "user_id")
			assert.Equal(t, newInfo.Applicant, item.applicantID)
			assert.Equal(t, newInfo.UserID, "user_id")
		})
	}
}

func TestService_GetAuthInfo(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx     context.Context
		request *spb.GetAuthInfoRequest
	}
	tests := []struct {
		name     string
		args     args
		setup    func() *Service
		response *spb.GetAuthInfoResponse
		wantErr  error
	}{
		{
			name: "happy case: get auth info by username and domain name",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				}, nil)

				userRepo := &mockUserRepoV2{
					getByUsername: func(context.Context, database.QueryExecer, string, string) (*entity.AuthUser, error) {
						return &entity.AuthUser{Email: database.Text("Email"), LoginEmail: database.Text("LoginEmail")}, nil
					},
				}

				featureManager := new(mock_features.MockFeatureManager)
				featureManager.On("IsEnableDecouplingUserAndAuthDB", mock.Anything).Return(true)
				featureManager.On("IsEnableUsernameStudentParentStaff", mock.Anything, mock.Anything).Return(true)

				return &Service{
					OrganizationRepo: organizationRepo,
					UserRepoV2:       userRepo,
					FeatureManager:   featureManager,
				}
			},
			response: &spb.GetAuthInfoResponse{
				LoginEmail:     "LoginEmail",
				TenantId:       "TenantID",
				Email:          "Email",
				OrganizationId: "48",
			},
			wantErr: nil,
		},
		{
			name: "happy case: get auth info by username and domain name with disable username feature",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				}, nil)

				userRepo := &mockUserRepoV2{
					getByEmail: func(context.Context, database.QueryExecer, string, string) (*entity.AuthUser, error) {
						return &entity.AuthUser{Email: database.Text("Email"), LoginEmail: database.Text("LoginEmail")}, nil
					},
				}

				featureManager := new(mock_features.MockFeatureManager)
				featureManager.On("IsEnableDecouplingUserAndAuthDB", mock.Anything).Return(true)
				featureManager.On("IsEnableUsernameStudentParentStaff", mock.Anything, mock.Anything).Return(false)

				return &Service{
					OrganizationRepo: organizationRepo,
					UserRepoV2:       userRepo,
					FeatureManager:   featureManager,
				}
			},
			response: &spb.GetAuthInfoResponse{
				LoginEmail:     "",
				TenantId:       "TenantID",
				Email:          "Email",
				OrganizationId: "48",
			},
			wantErr: nil,
		},
		{
			name: "bad case: domain name not found",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)

				return &Service{OrganizationRepo: organizationRepo}
			},
			response: nil,
			wantErr:  status.Error(codes.NotFound, errorx.ErrOrganizationNotFound.Error()),
		},
		{
			name: "bad case: username not found",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				}, nil)

				userRepo := &mockUserRepoV2{
					getByUsername: func(context.Context, database.QueryExecer, string, string) (*entity.AuthUser, error) {
						return nil, pgx.ErrNoRows
					},
				}

				featureManager := new(mock_features.MockFeatureManager)
				featureManager.On("IsEnableDecouplingUserAndAuthDB", mock.Anything).Return(true)
				featureManager.On("IsEnableUsernameStudentParentStaff", mock.Anything, mock.Anything).Return(true)

				return &Service{
					OrganizationRepo: organizationRepo,
					UserRepoV2:       userRepo,
					FeatureManager:   featureManager,
				}
			},
			response: nil,
			wantErr:  status.Error(codes.NotFound, errorx.ErrUsernameNotFound.Error()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			response, err := s.GetAuthInfo(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.response, response, "s.GetAuthInfo(%v, %v)", tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err, "s.GetAuthInfo(%v, %v)", tt.args.ctx, tt.args.request)
		})
	}
}

func TestService_getAuthInfoByUsernameDomainName(t *testing.T) {
	t.Parallel()

	type (
		bobDB  struct{ *mock_database.Ext }
		authDB struct{ *mock_database.Ext }
	)

	bob := bobDB{}
	auth := authDB{}

	type args struct {
		ctx     context.Context
		request *spb.GetAuthInfoRequest
	}
	type result struct {
		org  *entity.Organization
		user *entity.AuthUser
	}
	tests := []struct {
		name    string
		args    args
		setup   func() *Service
		result  result
		wantErr error
	}{
		{
			name: "happy case: get auth info by username and domain name",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				}, nil)

				userRepo := &mockUserRepoV2{
					getByUsername: func(ctx context.Context, db database.QueryExecer, _ string, _ string) (*entity.AuthUser, error) {
						if db != auth {
							return nil, errors.New("db is not auth db")
						}
						return &entity.AuthUser{Email: database.Text("Email"), LoginEmail: database.Text("LoginEmail")}, nil
					},
				}

				featureManager := new(mock_features.MockFeatureManager)
				featureManager.On("IsEnableDecouplingUserAndAuthDB", mock.Anything).Return(true)
				featureManager.On("IsEnableUsernameStudentParentStaff", mock.Anything, mock.Anything).Return(true)

				return &Service{
					OrganizationRepo: organizationRepo,
					FeatureManager:   featureManager,
					UserRepoV2:       userRepo,
					DB:               bob,
					AuthDB:           auth,
				}
			},
			result: result{
				org: &entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				},
				user: &entity.AuthUser{
					Email:      database.Text("Email"),
					LoginEmail: database.Text("LoginEmail"),
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: get auth info by username and domain name with disable IsEnableUsernameStudentParentStaff",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				}, nil)

				userRepo := &mockUserRepoV2{
					getByEmail: func(ctx context.Context, db database.QueryExecer, _ string, _ string) (*entity.AuthUser, error) {
						if db != auth {
							return nil, errors.New("db is not auth db")
						}
						return &entity.AuthUser{Email: database.Text("Email"), LoginEmail: database.Text("LoginEmail")}, nil
					},
				}

				featureManager := new(mock_features.MockFeatureManager)
				featureManager.On("IsEnableDecouplingUserAndAuthDB", mock.Anything).Return(true)
				featureManager.On("IsEnableUsernameStudentParentStaff", mock.Anything, mock.Anything).Return(false)

				return &Service{
					OrganizationRepo: organizationRepo,
					FeatureManager:   featureManager,
					UserRepoV2:       userRepo,
					DB:               bob,
					AuthDB:           auth,
				}
			},
			result: result{
				org: &entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				},
				user: &entity.AuthUser{
					Email:      database.Text("Email"),
					LoginEmail: database.Text(""),
				},
			},
			wantErr: nil,
		},
		{
			name: "happy case: get auth info by username and domain name with disable IsEnableDecouplingUserAndAuthDB",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				}, nil)

				userRepo := &mockUserRepoV2{
					getByUsername: func(ctx context.Context, db database.QueryExecer, _ string, _ string) (*entity.AuthUser, error) {
						if db != bob {
							return nil, errors.New("db is not auth db")
						}
						return &entity.AuthUser{Email: database.Text("Email"), LoginEmail: database.Text("LoginEmail")}, nil
					},
				}

				featureManager := new(mock_features.MockFeatureManager)
				featureManager.On("IsEnableDecouplingUserAndAuthDB", mock.Anything).Return(false)
				featureManager.On("IsEnableUsernameStudentParentStaff", mock.Anything, mock.Anything).Return(true)

				return &Service{
					OrganizationRepo: organizationRepo,
					FeatureManager:   featureManager,
					UserRepoV2:       userRepo,
					DB:               bob,
					AuthDB:           auth,
				}
			},
			result: result{
				org: &entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				},
				user: &entity.AuthUser{
					Email:      database.Text("Email"),
					LoginEmail: database.Text("LoginEmail"),
				},
			},
			wantErr: nil,
		},
		{
			name: "bad case: domain name not found",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(nil, pgx.ErrNoRows)

				return &Service{OrganizationRepo: organizationRepo}
			},
			result: result{
				org:  nil,
				user: nil,
			},
			wantErr: status.Error(codes.NotFound, errorx.ErrOrganizationNotFound.Error()),
		},
		{
			name: "bad case: username not found",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				}, nil)

				userRepo := mockUserRepoV2{
					getByUsername: func(ctx context.Context, db database.QueryExecer, _ string, _ string) (*entity.AuthUser, error) {
						if db != auth {
							return nil, errors.New("db is not auth db")
						}
						return nil, pgx.ErrNoRows
					},
				}

				featureManager := new(mock_features.MockFeatureManager)
				featureManager.On("IsEnableDecouplingUserAndAuthDB", mock.Anything).Return(true)
				featureManager.On("IsEnableUsernameStudentParentStaff", mock.Anything, mock.Anything).Return(true)

				return &Service{
					OrganizationRepo: organizationRepo,
					FeatureManager:   featureManager,
					UserRepoV2:       &userRepo,
					DB:               bob,
					AuthDB:           auth,
				}
			},
			result: result{
				org:  nil,
				user: nil,
			},
			wantErr: status.Error(codes.NotFound, errorx.ErrUsernameNotFound.Error()),
		},
		{
			name: "bad case: username not found with bob db",
			args: args{
				ctx: context.Background(),
				request: &spb.GetAuthInfoRequest{
					Username:   "Username",
					DomainName: "DomainName",
				},
			},
			setup: func() *Service {
				organizationRepo := new(mock_repositories.MockOrganizationRepo)
				organizationRepo.On("GetByDomainName", mock.Anything, mock.Anything, mock.Anything).Return(&entity.Organization{
					OrganizationID: database.Text("48"),
					TenantID:       database.Text("TenantID"),
					Name:           database.Text("Name"),
				}, nil)

				userRepo := mockUserRepoV2{
					getByUsername: func(ctx context.Context, db database.QueryExecer, _ string, _ string) (*entity.AuthUser, error) {
						if db != bob {
							return nil, errors.New("db is not auth db")
						}
						return nil, pgx.ErrNoRows
					},
				}

				featureManager := new(mock_features.MockFeatureManager)
				featureManager.On("IsEnableDecouplingUserAndAuthDB", mock.Anything).Return(false)
				featureManager.On("IsEnableUsernameStudentParentStaff", mock.Anything, mock.Anything).Return(true)

				return &Service{
					OrganizationRepo: organizationRepo,
					FeatureManager:   featureManager,
					UserRepoV2:       &userRepo,
					DB:               bob,
					AuthDB:           auth,
				}
			},
			result: result{
				org:  nil,
				user: nil,
			},
			wantErr: status.Error(codes.NotFound, errorx.ErrUsernameNotFound.Error()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			org, user, err := s.getAuthInfoByUsernameDomainName(tt.args.ctx, tt.args.request.GetUsername(), tt.args.request.GetDomainName())
			assert.Equalf(t, tt.result, result{org: org, user: user}, "s.GetAuthInfo(%v, %v)", tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err, "s.GetAuthInfo(%v, %v)", tt.args.ctx, tt.args.request)
		})
	}
}
