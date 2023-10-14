package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_enigma_services "github.com/manabie-com/backend/mock/enigma/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_locationRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestUserRegistrationService_SyncStudentHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)

	userRepo := new(mock_repositories.MockUserRepo)
	studentRepo := new(mock_repositories.MockStudentRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	userGroupMemberRepo := new(mock_repositories.MockUserGroupsMemberRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	studentEnrollmentStatusHistoryRepo := new(mock_repositories.MockStudentEnrollmentStatusHistoryRepo)

	partnerLogService := new(mock_enigma_services.MockPartnerSyncDataLogService)
	zapLogger := ctxzap.Extract(ctx)

	service := UserRegistrationService{
		Logger:                             zapLogger,
		DB:                                 db,
		PartnerSyncDataLogService:          partnerLogService,
		UserRepo:                           userRepo,
		StudentRepo:                        studentRepo,
		UserGroupV2Repo:                    userGroupV2Repo,
		UserGroupMemberRepo:                userGroupMemberRepo,
		LocationRepo:                       locationRepo,
		UserAccessPathRepo:                 userAccessPathRepo,
		StudentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case: create new student",
			ctx:  ctx,
			req: &npb.EventUserRegistration{
				Students: []*npb.EventUserRegistration_Student{
					{
						ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
						StudentId:   idutil.ULIDNow(),
						StudentDivs: []int64{2},
					},
				},
				Signature: idutil.ULIDNow(),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", mock.Anything, db, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				studentRepo.On("Find", mock.Anything, db, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationOrg", mock.Anything, tx, mock.Anything).Once().Return(&domain.Location{}, nil)
				userAccessPathRepo.On("Upsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
				studentEnrollmentStatusHistoryRepo.On("Upsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userGroupMemberRepo.On("UpsertBatch", mock.Anything, tx, mock.Anything).Return(nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: update existing student",
			ctx:  ctx,
			req: &npb.EventUserRegistration{
				Students: []*npb.EventUserRegistration_Student{
					{
						ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
						StudentId:   idutil.ULIDNow(),
						StudentDivs: []int64{2},
					},
				},
				Signature: idutil.ULIDNow(),
			},
			setup: func(ctx context.Context) {
				id := idutil.ULIDNow()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", mock.Anything, db, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				studentRepo.On("Find", mock.Anything, db, mock.Anything).Once().Return(&entity.LegacyStudent{ID: database.Text(id)}, nil)
				userRepo.On("FindByIDUnscope", mock.Anything, tx, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(id)}, nil)
				studentRepo.On("Update", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userGroupMemberRepo.On("UpsertBatch", mock.Anything, tx, mock.Anything).Return(nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: create 1 student and update 1 existing student",
			ctx:  ctx,
			req: &npb.EventUserRegistration{
				Students: []*npb.EventUserRegistration_Student{
					{
						ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
						StudentId:   idutil.ULIDNow(),
						StudentDivs: []int64{2},
					},
					{
						ActionKind:  npb.ActionKind_ACTION_KIND_UPSERTED,
						StudentId:   idutil.ULIDNow(),
						StudentDivs: []int64{2},
					},
				},
				Signature: idutil.ULIDNow(),
			},
			setup: func(ctx context.Context) {
				id := idutil.ULIDNow()
				db.On("Begin", mock.Anything).Twice().Return(tx, nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", mock.Anything, db, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				studentRepo.On("Find", mock.Anything, db, mock.Anything).Once().Return(&entity.LegacyStudent{ID: database.Text(id)}, nil)
				studentRepo.On("Find", mock.Anything, db, mock.Anything).Once().Return(nil, nil)
				studentRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(nil)
				studentEnrollmentStatusHistoryRepo.On("Upsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationOrg", mock.Anything, tx, mock.Anything).Once().Return(&domain.Location{}, nil)
				userAccessPathRepo.On("Upsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userRepo.On("FindByIDUnscope", mock.Anything, tx, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(id)}, nil)
				studentRepo.On("Update", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userGroupMemberRepo.On("UpsertBatch", mock.Anything, tx, mock.Anything).Return(nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Twice().Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: soft delete student",
			ctx:  ctx,
			req: &npb.EventUserRegistration{
				Students: []*npb.EventUserRegistration_Student{
					{
						ActionKind:  npb.ActionKind_ACTION_KIND_DELETED,
						StudentId:   idutil.ULIDNow(),
						StudentDivs: []int64{2},
					},
				},
				Signature: idutil.ULIDNow(),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", mock.Anything, db, constant.RoleStudent).Once().Return(&entity.UserGroupV2{}, nil)
				studentRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
				studentEnrollmentStatusHistoryRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userGroupMemberRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Return(nil)
				userAccessPathRepo.On("Delete", mock.Anything, tx, mock.Anything).Return(nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.JPREPSchool),
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.setup(testCase.ctx)

			req, err := proto.Marshal(testCase.req.(*npb.EventUserRegistration))
			assert.Nil(t, err)

			_, err = service.SyncStudentHandler(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)

			mock.AssertExpectationsForObjects(t, db, tx, studentRepo, userRepo, userGroupV2Repo, userGroupMemberRepo, partnerLogService)
		})
	}
}

func TestUserRegistrationService_SyncStaffHandler(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)

	userRepo := new(mock_repositories.MockUserRepo)
	userGroupV2Repo := new(mock_repositories.MockUserGroupV2Repo)
	userGroupMemberRepo := new(mock_repositories.MockUserGroupsMemberRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	staffRepo := new(mock_repositories.MockStaffRepo)
	locationRepo := new(mock_locationRepo.MockLocationRepo)
	userAccessPathRepo := new(mock_repositories.MockUserAccessPathRepo)
	partnerLogService := new(mock_enigma_services.MockPartnerSyncDataLogService)
	zapLogger := ctxzap.Extract(ctx)

	service := UserRegistrationService{
		Logger:                    zapLogger,
		DB:                        db,
		PartnerSyncDataLogService: partnerLogService,
		UserRepo:                  userRepo,
		TeacherRepo:               teacherRepo,
		StaffRepo:                 staffRepo,
		UserGroupV2Repo:           userGroupV2Repo,
		UserGroupMemberRepo:       userGroupMemberRepo,
		LocationRepo:              locationRepo,
		UserAccessPathRepo:        userAccessPathRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case: create new staff",
			ctx:  ctx,
			req: &npb.EventUserRegistration{
				Staffs: []*npb.EventUserRegistration_Staff{
					{
						ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
						StaffId:    idutil.ULIDNow(),
						Name:       idutil.ULIDNow(),
					},
				},
				Signature: idutil.ULIDNow(),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", mock.Anything, db, constant.RoleTeacher).Once().Return(&entity.UserGroupV2{}, nil)
				staffRepo.On("Find", mock.Anything, db, mock.Anything).Once().Return(nil, nil)
				staffRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationOrg", mock.Anything, tx, mock.Anything).Once().Return(&domain.Location{}, nil)
				userAccessPathRepo.On("Upsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userGroupMemberRepo.On("UpsertBatch", mock.Anything, tx, mock.Anything).Return(nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: update existing staff",
			ctx:  ctx,
			req: &npb.EventUserRegistration{
				Staffs: []*npb.EventUserRegistration_Staff{
					{
						ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
						StaffId:    idutil.ULIDNow(),
						Name:       idutil.ULIDNow(),
					},
				},
				Signature: idutil.ULIDNow(),
			},
			setup: func(ctx context.Context) {
				id := idutil.ULIDNow()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", mock.Anything, db, constant.RoleTeacher).Once().Return(&entity.UserGroupV2{}, nil)
				staffRepo.On("Find", mock.Anything, db, mock.Anything).Once().Return(&entity.Staff{ID: database.Text(id)}, nil)
				userRepo.On("FindByIDUnscope", mock.Anything, tx, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(id)}, nil)
				staffRepo.On("Update", mock.Anything, tx, mock.Anything).Once().Return(&entity.Staff{ID: database.Text(id)}, nil)
				userGroupMemberRepo.On("UpsertBatch", mock.Anything, tx, mock.Anything).Return(nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: create 1 new staff and update 1 existing staff",
			ctx:  ctx,
			req: &npb.EventUserRegistration{
				Staffs: []*npb.EventUserRegistration_Staff{
					{
						ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
						StaffId:    idutil.ULIDNow(),
						Name:       idutil.ULIDNow(),
					},
					{
						ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
						StaffId:    idutil.ULIDNow(),
						Name:       idutil.ULIDNow(),
					},
				},
				Signature: idutil.ULIDNow(),
			},
			setup: func(ctx context.Context) {
				id := idutil.ULIDNow()
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", mock.Anything, db, constant.RoleTeacher).Once().Return(&entity.UserGroupV2{}, nil)
				staffRepo.On("Find", mock.Anything, db, mock.Anything).Once().Return(&entity.Staff{ID: database.Text(id)}, nil)
				userRepo.On("FindByIDUnscope", mock.Anything, tx, mock.Anything).Once().Return(&entity.LegacyUser{ID: database.Text(id)}, nil)
				staffRepo.On("Update", mock.Anything, tx, mock.Anything).Once().Return(&entity.Staff{ID: database.Text(id)}, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				staffRepo.On("Find", mock.Anything, db, mock.Anything).Once().Return(nil, nil)
				staffRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationOrg", mock.Anything, tx, mock.Anything).Once().Return(&domain.Location{}, nil)
				userAccessPathRepo.On("Upsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userGroupMemberRepo.On("UpsertBatch", mock.Anything, tx, mock.Anything).Return(nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case: soft delete staff",
			ctx:  ctx,
			req: &npb.EventUserRegistration{
				Staffs: []*npb.EventUserRegistration_Staff{
					{
						ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
						StaffId:    idutil.ULIDNow(),
						Name:       idutil.ULIDNow(),
					},
				},
				Signature: idutil.ULIDNow(),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				userGroupV2Repo.On("FindUserGroupByRoleName", mock.Anything, db, constant.RoleTeacher).Once().Return(&entity.UserGroupV2{}, nil)
				staffRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
				teacherRepo.On("SoftDeleteMultiple", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
				userGroupMemberRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Return(nil)
				partnerLogService.On("UpdateLogStatus", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.JPREPSchool),
				},
			}

			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			testCase.setup(testCase.ctx)

			req, err := proto.Marshal(testCase.req.(*npb.EventUserRegistration))
			assert.Nil(t, err)

			_, err = service.SyncStaffHandler(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)

			mock.AssertExpectationsForObjects(t, db, tx, staffRepo, teacherRepo, userRepo, userGroupV2Repo, userGroupMemberRepo, partnerLogService)
		})
	}
}
