package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_firebase "github.com/manabie-com/backend/mock/golibs/firebase"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveParentsAndFamilyRelationship(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := new(mock_database.Tx)
	studentRepo := new(mock_repositories.MockStudentRepo)
	parentRepo := new(mock_repositories.MockParentRepo)
	studentParentRepo := new(mock_repositories.MockStudentParentRepo)
	jsm := new(mock_nats.JetStreamManagement)
	firebaseAuth := new(mock_firebase.AuthClient)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	s := UserModifierService{
		DB:                db,
		StudentRepo:       studentRepo,
		ParentRepo:        parentRepo,
		StudentParentRepo: studentParentRepo,
		FirebaseClient:    firebaseAuth,
		JSM:               jsm,
		UnleashClient:     mockUnleashClient,
	}

	existingParentUser := &entity.LegacyUser{
		ID:          database.Text("id"),
		Email:       database.Text("existing-parent-email@example.com"),
		PhoneNumber: database.Text("existing-parent-phone-number"),
	}
	existingParent := &entity.Parent{
		ID:         existingParentUser.ID,
		SchoolID:   database.Int4(1),
		LegacyUser: *existingParentUser,
		ParentAdditionalInfo: &entity.ParentAdditionalInfo{
			Relationship: pb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER.String(),
		},
	}
	existingStudentUser := &entity.LegacyUser{
		ID:          database.Text(idutil.ULIDNow()),
		Email:       database.Text("existing-student-email@example.com"),
		PhoneNumber: database.Text("existing-student-phone-number"),
	}
	existingStudent := &entity.LegacyStudent{
		ID:         existingStudentUser.ID,
		SchoolID:   database.Int4(1),
		LegacyUser: *existingStudentUser,
	}

	testCases := []TestCase{
		{
			name: "remove parent and family relationship successfully with feature flag on",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "some-parent-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"some-parent-id"})).Once().Return(entity.Parents{existingParent}, nil)
				studentParentRepo.On("FindStudentParentsByParentID", ctx, tx, "some-parent-id").Once().Return([]*entity.StudentParent{
					{
						StudentID: database.Text("StudentID_01"),
						ParentID:  database.Text("ParentID_01"),
					},
					{
						StudentID: database.Text("StudentID_02"),
						ParentID:  database.Text("ParentID_02"),
					},
				}, nil)
				studentParentRepo.On("RemoveParentFromStudent", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				// bus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByID", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "remove parent and family relationship successfully with feature flag off",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "some-parent-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"some-parent-id"})).Once().Return(entity.Parents{existingParent}, nil)
				studentParentRepo.On("RemoveParentFromStudent", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				// bus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				studentParentRepo.On("UpsertParentAccessPathByID", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "remove parent and family relationship fail by empty studentID",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "",
				ParentId:  "some-parent-id",
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.InvalidArgument, "student id cannot be empty")),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
		},
		{
			name: "remove parent and family relationship fail by empty ParentID",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "",
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.InvalidArgument, "parent id cannot be empty")),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
		},
		{
			name: "remove parent and family relationship fail by query student repo have error",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "some-parent-id",
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.Internal, "something wrong")),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return(nil, fmt.Errorf("something wrong"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "remove parent and family relationship fail by student id un-exist",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "some-parent-id",
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.InvalidArgument, "cannot remove parents associated with un-existing student in system: some-student-id")),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "remove parent and family relationship fail by query parent repo have error",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "some-parent-id",
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.Internal, "something wrong")),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"some-parent-id"})).Once().Return(entity.Parents{}, fmt.Errorf("something wrong"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "remove parent and family relationship fail by parent id un-exist",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "some-parent-id",
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.InvalidArgument, "cannot remove un-existing parents in system: some-parent-id")),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"some-parent-id"})).Once().Return(entity.Parents{}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "remove parent and family relationship fail by remove student_parent repo have error",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "some-parent-id",
			},
			expectedErr: fmt.Errorf("something wrong"),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"some-parent-id"})).Once().Return(entity.Parents{existingParent}, nil)
				studentParentRepo.On("RemoveParentFromStudent", ctx, tx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("something wrong"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "cannot remove parent when parent have only one student",
			ctx:  ctx,
			req: &pb.RemoveParentFromStudentRequest{
				StudentId: "some-student-id",
				ParentId:  "some-parent-id",
			},
			expectedErr: errorx.ToStatusError(status.Error(codes.InvalidArgument, fmt.Sprint(constant.InvalidRemoveParent))),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Once().Return(tx, nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				studentRepo.On("FindStudentProfilesByIDs", ctx, tx, database.TextArray([]string{"some-student-id"})).Once().Return([]*entity.LegacyStudent{existingStudent}, nil)
				parentRepo.On("GetByIds", ctx, tx, database.TextArray([]string{"some-parent-id"})).Once().Return(entity.Parents{existingParent}, nil)
				studentParentRepo.On("FindStudentParentsByParentID", ctx, tx, "some-parent-id").Once().Return([]*entity.StudentParent{
					{
						StudentID: database.Text("StudentID_01"),
						ParentID:  database.Text("ParentID_01"),
					},
				}, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)

		_, err := s.RemoveParentFromStudent(testCase.ctx, testCase.req.(*pb.RemoveParentFromStudentRequest))
		if err != nil {
			fmt.Println(err)
		}
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
