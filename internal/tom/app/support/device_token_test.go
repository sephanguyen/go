package support

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeviceToken_HandleUserInfo(t *testing.T) {
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	conversationLocationRepo := new(mock_repositories.MockConversationLocationRepo)
	deviceTokenRepo := new(mock_repositories.MockUserDeviceTokenRepo)
	grantedPermissionRepo := new(mock_repositories.MockGrantedPermissionsRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)

	ctx := context.Background()

	jsm := &mock_nats.JetStreamManagement{}
	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	db.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Commit", mock.Anything).Return(nil)

	convIDs := []string{"student-conv", "parent-conv"}
	updatedLocations := []string{"current 2", "current 3"}
	removedLocations := []string{"current 1"}
	onlyParentLocations := map[string][]core.ConversationLocation{
		"parent-conv": {
			{
				ConversationID: dbText("parent-conv"),
				LocationID:     dbText("current 1"),
			},
			{
				ConversationID: dbText("parent-conv"),
				LocationID:     dbText("current 2"),
			},
		},
	}

	teacherConversationMembersMap := map[string][]string{
		"parent-conv": {"teacher-1", "teacher-2"},
	}

	grantedPermissionMap := map[string][]*core.GrantedPermission{
		"teacher-1": {
			{
				UserID:     database.Text("teacher-1"),
				LocationID: database.Text("current 1"),
			},
			{
				UserID:     database.Text("teacher-1"),
				LocationID: database.Text("current 2"),
			},
		},
		"teacher-2": {
			{
				UserID:     database.Text("teacher-2"),
				LocationID: database.Text("current 1"),
			},
			{
				UserID:     database.Text("teacher-2"),
				LocationID: database.Text("current 2"),
			},
		},
	}

	svc := &DeviceTokenModifier{
		DB:                       db,
		JSM:                      jsm,
		ConversationStudentRepo:  conversationStudentRepo,
		ConversationRepo:         conversationRepo,
		UserDeviceTokenRepo:      deviceTokenRepo,
		ConversationLocationRepo: conversationLocationRepo,
		ConversationMemberRepo:   conversationMemberRepo,
		GrantedPermissionRepo:    grantedPermissionRepo,
		Logger:                   zap.NewNop(),
	}
	userID := "user-id"
	name := "username"
	t.Parallel()
	cases := []TestCase{
		{
			name: "success with non-student user_id",
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			ctx: ctx,
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, db, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return([]string{"parent-conv"}, nil)

				// Location
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, db, database.TextArray([]string{"parent-conv"})).
					Once().Return(onlyParentLocations, nil)
				conversationLocationRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(news []core.ConversationLocation) bool {
					for _, item := range news {
						if !stringIn(item.LocationID.String, updatedLocations) {
							return false
						}
					}
					return true
				})).Once().Return(nil)
				conversationLocationRepo.On("RemoveLocationsForConversation", mock.Anything, tx, "parent-conv", removedLocations).
					Return(nil)

				conversationRepo.On("SetName", mock.Anything, db, database.TextArray([]string{"parent-conv"}), database.Text(name)).Once().Return(nil)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray([]string{"parent-conv"}), database.TextArray(constant.ConversationStaffRoles)).Once().Return(teacherConversationMembersMap, nil)
				grantedPermissionRepo.On("FindByUserIDAndPermissionName", mock.Anything, db, database.TextArray([]string{"teacher-1", "teacher-2"}), database.Text("master.location.read")).Once().Return(grantedPermissionMap, nil)

				deviceTokenRepo.On("Upsert", mock.Anything, db, mock.MatchedBy(func(u *core.UserDeviceToken) bool {
					return u.UserID.String == userID && u.UserName.String == name
				})).Once().Return(nil)

				jsm.On("PublishContext", mock.Anything, constants.SubjectChatUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)

			},
			expectedResp: maketuple(true, nil),
		},
		{
			name: "success with student user_id",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
					// return 2 convIDs (parent/student)
				).Once().Return(convIDs, nil)

				// location repo returns only parent locations
				// this means location for student does not exist, upserting anyway
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, db, database.TextArray(convIDs)).
					Once().Return(onlyParentLocations, nil)
				conversationLocationRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(news []core.ConversationLocation) bool {
					for _, item := range news {
						if !stringIn(item.LocationID.String, updatedLocations) {
							return false
						}
					}
					return true
				})).Twice().Return(nil)
				conversationLocationRepo.On("RemoveLocationsForConversation", mock.Anything, tx, "parent-conv", removedLocations).
					Return(nil)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray(convIDs), database.TextArray(constant.ConversationStaffRoles)).Once().Return(teacherConversationMembersMap, nil)
				conversationMemberRepo.On("SetStatusByConversationAndUserIDs", mock.Anything, tx, database.Text("parent-conv"), database.TextArray([]string{"teacher-2"}), database.Text(core.ConversationStatusActive)).Once().Return(nil)
				grantedPermissionRepo.On("FindByUserIDAndPermissionName", mock.Anything, db, database.TextArray([]string{"teacher-1", "teacher-2"}), database.Text("master.location.read")).Once().Return(grantedPermissionMap, nil)

				deviceTokenRepo.On("Upsert", mock.Anything, db, mock.MatchedBy(func(u *core.UserDeviceToken) bool {
					return u.UserID.String == userID && u.UserName.String == name
				})).Once().Return(nil)
				// update new name
				conversationRepo.On("SetName", mock.Anything, db, database.TextArray(convIDs), database.Text(name)).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectChatUpdated, mock.Anything).Twice().Return(&nats.PubAck{}, nil)
			},
			expectedResp: maketuple(true, nil),
		},
		{
			name: "error with FindByStudentIDs",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return(nil, pgx.ErrNoRows)

			},
			expectedErr: pgx.ErrNoRows,
		},
		{
			name: "error with FindByConversationIDsAndRoles",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return(convIDs, nil)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray(convIDs), database.TextArray(constant.ConversationStaffRoles)).
					Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: pgx.ErrNoRows,
		},
		{
			name: "error with FindByConversationIDs",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return(convIDs, nil)

				// location repo returns only parent locations
				// this means location for student does not exist, upserting anyway
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, db, database.TextArray(convIDs)).
					Once().Return(nil, pgx.ErrNoRows)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray(convIDs), database.TextArray(constant.ConversationStaffRoles)).
					Once().Return(teacherConversationMembersMap, nil)
			},
			expectedErr: pgx.ErrNoRows,
		},
		{
			name: "success with empty teacher members",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return(convIDs, nil)

				// location repo returns only parent locations
				// this means location for student does not exist, upserting anyway
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, db, database.TextArray(convIDs)).
					Once().Return(onlyParentLocations, nil)
				conversationLocationRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(news []core.ConversationLocation) bool {
					for _, item := range news {
						if !stringIn(item.LocationID.String, updatedLocations) {
							return false
						}
					}
					return true
				})).Twice().Return(nil)
				conversationLocationRepo.On("RemoveLocationsForConversation", mock.Anything, tx, "parent-conv", removedLocations).
					Return(nil)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray(convIDs), database.TextArray(constant.ConversationStaffRoles)).
					Once().Return(map[string][]string{}, nil)

				deviceTokenRepo.On("Upsert", mock.Anything, db, mock.MatchedBy(func(u *core.UserDeviceToken) bool {
					return u.UserID.String == userID && u.UserName.String == name
				})).Once().Return(nil)
				// update new name
				conversationRepo.On("SetName", mock.Anything, db, database.TextArray(convIDs), database.Text(name)).Once().Return(nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectChatUpdated, mock.Anything).Twice().Return(&nats.PubAck{}, nil)
			},
			expectedResp: maketuple(true, nil),
		},
		{
			name: "error with FindByUserIDAndPermissionName",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return(convIDs, nil)

				// location repo returns only parent locations
				// this means location for student does not exist, upserting anyway
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, db, database.TextArray(convIDs)).
					Once().Return(onlyParentLocations, nil)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray(convIDs), database.TextArray(constant.ConversationStaffRoles)).
					Once().Return(teacherConversationMembersMap, nil)
				grantedPermissionRepo.On("FindByUserIDAndPermissionName", mock.Anything, db, database.TextArray([]string{"teacher-1", "teacher-2"}), database.Text("master.location.read")).Once().Return(nil, pgx.ErrNoRows)

			},
			expectedErr: pgx.ErrNoRows,
		},
		{
			name: "error with BulkUpsert",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return(convIDs, nil)

				// location repo returns only parent locations
				// this means location for student does not exist, upserting anyway
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, db, database.TextArray(convIDs)).
					Once().Return(onlyParentLocations, nil)
				conversationLocationRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(news []core.ConversationLocation) bool {
					for _, item := range news {
						if !stringIn(item.LocationID.String, updatedLocations) {
							return false
						}
					}
					return true
				})).Once().Return(pgx.ErrTxClosed)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray(convIDs), database.TextArray(constant.ConversationStaffRoles)).
					Once().Return(teacherConversationMembersMap, nil)
				grantedPermissionRepo.On("FindByUserIDAndPermissionName", mock.Anything, db, database.TextArray([]string{"teacher-1", "teacher-2"}), database.Text("master.location.read")).Once().Return(grantedPermissionMap, nil)

				tx.On("Rollback", mock.Anything).Return(nil)
			},
			expectedErr: pgx.ErrTxClosed,
		},
		{
			name: "error with RemoveLocationsForConversation",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return(convIDs, nil)

				// location repo returns only parent locations
				// this means location for student does not exist, upserting anyway
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, db, database.TextArray(convIDs)).
					Once().Return(map[string][]core.ConversationLocation{
					"student-conv": {
						{
							ConversationID: dbText("student-conv"),
							LocationID:     dbText("current 1"),
						},
					},
				}, nil)
				conversationLocationRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(news []core.ConversationLocation) bool {
					for _, item := range news {
						if !stringIn(item.LocationID.String, updatedLocations) {
							return false
						}
					}
					return true
				})).Twice().Return(nil)
				conversationLocationRepo.On("RemoveLocationsForConversation", mock.Anything, tx, "student-conv", removedLocations).Once().
					Return(pgx.ErrTxClosed)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray(convIDs), database.TextArray(constant.ConversationStaffRoles)).
					Once().Return(teacherConversationMembersMap, nil)
				grantedPermissionRepo.On("FindByUserIDAndPermissionName", mock.Anything, db, database.TextArray([]string{"teacher-1", "teacher-2"}), database.Text("master.location.read")).Once().Return(grantedPermissionMap, nil)

				tx.On("Rollback", mock.Anything).Return(nil)
			},
			expectedErr: pgx.ErrTxClosed,
		},
		{
			name: "error with SetStatusByConversationAndUserIDs",
			ctx:  ctx,
			req: &upb.EvtUserInfo{
				UserId:      userID,
				Name:        name,
				LocationIds: updatedLocations,
			},
			setup: func(ctx context.Context) {
				// find conversation students given user_id
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{userID}),
					pgtype.Text{Status: pgtype.Null},
				).Once().Return([]string{"parent-conv"}, nil)

				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, db, database.TextArray([]string{"parent-conv"})).
					Once().Return(map[string][]core.ConversationLocation{
					"parent-conv": {
						{
							ConversationID: dbText("parent-conv"),
							LocationID:     dbText("current 4"),
						},
						{
							ConversationID: dbText("parent-conv"),
							LocationID:     dbText("current 5"),
						},
					},
				}, nil)
				conversationLocationRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(news []core.ConversationLocation) bool {
					for _, item := range news {
						if !stringIn(item.LocationID.String, updatedLocations) {
							return false
						}
					}
					return true
				})).Twice().Return(nil)
				conversationLocationRepo.On("RemoveLocationsForConversation", mock.Anything, tx, "parent-conv", []string{"current 4", "current 5"}).
					Return(nil)
				conversationMemberRepo.On("FindByConversationIDsAndRoles", mock.Anything, db, database.TextArray([]string{"parent-conv"}), database.TextArray(constant.ConversationStaffRoles)).
					Once().Return(teacherConversationMembersMap, nil)
				conversationMemberRepo.On("SetStatusByConversationAndUserIDs", mock.Anything, tx, database.TextArray([]string{"parent-conv"}), database.TextArray([]string{"teacher-1", "teacher-2"}), database.Text(core.ConversationStatusInActive)).Once().Return(pgx.ErrTxClosed)
				grantedPermissionRepo.On("FindByUserIDAndPermissionName", mock.Anything, db, database.TextArray([]string{"teacher-1", "teacher-2"}), database.Text("master.location.read")).Once().Return(map[string][]*core.GrantedPermission{
					"teacher-1": {
						{
							UserID:     database.Text("teacher-1"),
							LocationID: database.Text("current 4"),
						},
						{
							UserID:     database.Text("teacher-1"),
							LocationID: database.Text("current 5"),
						},
					},
					"teacher-2": {
						{
							UserID:     database.Text("teacher-2"),
							LocationID: database.Text("current 4"),
						},
						{
							UserID:     database.Text("teacher-2"),
							LocationID: database.Text("current 5"),
						},
					},
				}, nil)
				tx.On("Rollback", mock.Anything).Return(nil)

			},
			expectedErr: pgx.ErrTxClosed,
		},
	}
	for _, tcase := range cases {
		t.Run(tcase.name, func(t *testing.T) {
			tcase.setup(tcase.ctx)
			retry, err := svc.HandleEvtUserInfo(tcase.ctx, tcase.req.(*upb.EvtUserInfo), true)

			if tcase.expectedResp != nil {
				assert.Equal(t, maketuple(retry, err), tcase.expectedResp.(tuple))
			} else {
				assert.ErrorIs(t, err, tcase.expectedErr)
			}
			mock.AssertExpectationsForObjects(t, db, tx, jsm, conversationStudentRepo, conversationRepo, deviceTokenRepo)
		})
	}
}

type tuple = []interface{}

func maketuple(args ...interface{}) tuple {
	return args
}
