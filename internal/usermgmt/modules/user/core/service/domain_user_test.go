package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	entity_mock "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"

	"gotest.tools/assert"
)

func Test_ValidateUserEmailsExistedInSystem(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockUserRepo := &mock_repositories.MockDomainUserRepo{}
	mockDb := &mock_database.Ext{}

	userReq := entity.Users{
		entity_mock.User{
			RandomUser: entity_mock.RandomUser{
				Email: field.NewString("user1@manabie.com"),
			},
		},
		entity_mock.User{
			RandomUser: entity_mock.RandomUser{
				UserID: field.NewString("user2-id"),
				Email:  field.NewString("user2@manabie.com"),
			},
		},
	}

	testCases := []TestCase{
		{
			name:        "user emails do not exist in system, checking by user id",
			ctx:         ctx,
			req:         userReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByEmailsInsensitiveCase", ctx, mockDb, []string{"user1@manabie.com", "user2@manabie.com"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "user emails do not exist in system, checking by external user id",
			ctx:  ctx,
			req: entity.Users{
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						Email:          field.NewString("user1@manabie.com"),
						ExternalUserID: field.NewString("external-user-id-1"),
					},
				},
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						UserID: field.NewString("user2-id"),
						Email:  field.NewString("user2@manabie.com"),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByEmailsInsensitiveCase", ctx, mockDb, []string{"user1@manabie.com", "user2@manabie.com"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:         field.NewString("user-id-1"),
							Email:          field.NewString("user1@manabie.com"),
							ExternalUserID: field.NewString("external-user-id-1"),
						},
					},
				}, nil)
			},
		},
		{
			name: "user emails exist in system when checking with user id",
			ctx:  ctx,
			req:  userReq,
			expectedErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldEmail),
				EntityName: entity.UserEntity,
				Index:      0,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByEmailsInsensitiveCase", ctx, mockDb, []string{"user1@manabie.com", "user2@manabie.com"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID: field.NewString("user-id"),
							Email:  field.NewString("user1@manabie.com"),
						},
					},
				}, nil)
			},
		},
		{
			name: "user emails exist in system, checking with user external id",
			ctx:  ctx,
			req: entity.Users{
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						Email:          field.NewString("user1@manabie.com"),
						ExternalUserID: field.NewString("external-user-id-1"),
					},
				},
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						UserID: field.NewString("user2-id"),
						Email:  field.NewString("user2@manabie.com"),
					},
				},
			},
			expectedErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldEmail),
				EntityName: entity.UserEntity,
				Index:      0,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByEmailsInsensitiveCase", ctx, mockDb, []string{"user1@manabie.com", "user2@manabie.com"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:         field.NewString("user-id-3"),
							Email:          field.NewString("user1@manabie.com"),
							ExternalUserID: field.NewString("external-user-id-3"),
						},
					},
				}, nil)
			},
		},
		// {
		// 	name: "user emails exist in system (case-insensitive)",
		// 	ctx:  ctx,
		// 	req:  userReq,
		// 	expectedErr: errcode.Error{
		// 		Code:      errcode.DataExist,
		// 		FieldName: "users[0].email",
		// 	},
		// 	setup: func(ctx context.Context) {
		// 		mockUserRepo.On("GetByEmailsInsensitiveCase", ctx, mockDb, []string{"user1@manabie.com", "user2@manabie.com"}).Once().Return(entity.DomainUsers{
		// 			entity_mock.User{
		// 				RandomUser: entity_mock.RandomUser{
		// 					UserID: field.NewString("user-id"),
		// 					Email:  field.NewString("USER1@manabie.com"),
		// 				},
		// 			},
		// 		}, nil)
		// 	},
		// },
		{
			name:        "user emails exist in system but same userID",
			ctx:         ctx,
			req:         userReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByEmailsInsensitiveCase", ctx, mockDb, []string{"user1@manabie.com", "user2@manabie.com"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID: field.NewString("user2-id"),
							Email:  field.NewString("user2@manabie.com"),
						},
					},
				}, nil)
			},
		},
		// {
		// 	name: "user emails exist in system but same userID (case-insensitive)",
		// 	ctx:  ctx,
		// 	req:  userReq,
		// 	expectedErr: errcode.Error{
		// 		Code:      errcode.DataExist,
		// 		FieldName: "users[1].email",
		// 	},
		// 	setup: func(ctx context.Context) {
		// 		mockUserRepo.On("GetByEmailsInsensitiveCase", ctx, mockDb, []string{"user1@manabie.com", "user2@manabie.com"}).Once().Return(entity.DomainUsers{
		// 			entity_mock.User{
		// 				RandomUser: entity_mock.RandomUser{
		// 					UserID: field.NewString("user2-id"),
		// 					Email:  field.NewString("USER2@manabie.com"),
		// 				},
		// 			},
		// 		}, nil)
		// 	},
		// },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			testCase.setup(testCase.ctx)
			err := ValidateUserEmailsExistedInSystem(ctx, mockUserRepo, mockDb, testCase.req.(entity.Users))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestValidateUserNamesExistedInSystem(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockUserRepo := &mock_repositories.MockDomainUserRepo{}
	mockDb := &mock_database.Ext{}

	userReq := entity.Users{
		entity_mock.User{
			RandomUser: entity_mock.RandomUser{
				UserName: field.NewString("username1"),
			},
		},
		entity_mock.User{
			RandomUser: entity_mock.RandomUser{
				UserID:   field.NewString("user2-id"),
				UserName: field.NewString("username2"),
			},
		},
	}

	testCases := []TestCase{
		{
			name:        "user usernames do not exist in system, checking by user id",
			ctx:         ctx,
			req:         userReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByUserNames", ctx, mockDb, []string{"username1", "username2"}).Once().Return(entity.Users{}, nil)
			},
		},
		{
			name: "user usernames exist in system when checking with user id",
			ctx:  ctx,
			req:  userReq,
			expectedErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByUserNames", ctx, mockDb, []string{"username1", "username2"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:   field.NewString("user-id"),
							UserName: field.NewString("username1"),
						},
					},
				}, nil)
			},
		},
		{
			name: "user usernames exist in system with lowercase, checking with user external id",
			ctx:  ctx,
			req: entity.Users{
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						UserName:       field.NewString("USERNAME1"),
						ExternalUserID: field.NewString("external-user-id-1"),
					},
				},
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						UserID:   field.NewString("user2-id"),
						UserName: field.NewString("username2"),
					},
				},
			},
			expectedErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByUserNames", ctx, mockDb, []string{"username1", "username2"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:         field.NewString("user-id-3"),
							UserName:       field.NewString("username1"),
							ExternalUserID: field.NewString("external-user-id-3"),
						},
					},
				}, nil)
			},
		},
		{
			name: "user usernames exist in system, checking with user external id",
			ctx:  ctx,
			req: entity.Users{
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						UserName:       field.NewString("username1"),
						ExternalUserID: field.NewString("external-user-id-1"),
					},
				},
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						UserID:   field.NewString("user2-id"),
						UserName: field.NewString("username2"),
					},
				},
			},
			expectedErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByUserNames", ctx, mockDb, []string{"username1", "username2"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:         field.NewString("user-id-3"),
							UserName:       field.NewString("USERNAME1"),
							ExternalUserID: field.NewString("external-user-id-3"),
						},
					},
				}, nil)
			},
		},
		{
			name: "user usernames exist in system, checking with user external id",
			ctx:  ctx,
			req: entity.Users{
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						UserName:       field.NewString("username1"),
						ExternalUserID: field.NewString("external-user-id-1"),
					},
				},
				entity_mock.User{
					RandomUser: entity_mock.RandomUser{
						UserID:   field.NewString("user2-id"),
						UserName: field.NewString("username2"),
					},
				},
			},
			expectedErr: entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.UserEntity,
				Index:      0,
			},
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByUserNames", ctx, mockDb, []string{"username1", "username2"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:         field.NewString("user-id-3"),
							UserName:       field.NewString("username1"),
							ExternalUserID: field.NewString("external-user-id-3"),
						},
					},
				}, nil)
			},
		},
		{
			name:        "user usernames exist in system but same userID",
			ctx:         ctx,
			req:         userReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUserRepo.On("GetByUserNames", ctx, mockDb, []string{"username1", "username2"}).Once().Return(entity.Users{
					entity_mock.User{
						RandomUser: entity_mock.RandomUser{
							UserID:   field.NewString("user2-id"),
							UserName: field.NewString("username2"),
						},
					},
				}, nil)
			},
		},
	}

	t.Parallel()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)
			testCase.setup(testCase.ctx)
			err := ValidateUserNamesExistedInSystem(ctx, mockUserRepo, mockDb, testCase.req.(entity.Users))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
