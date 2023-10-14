package consumers

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_class_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStudentPackageHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	classMemberRepo := &mock_class_repo.MockClassMemberRepo{}
	db := &mock_database.Ext{}
	jsm := new(mock_nats.JetStreamManagement)
	tx := &mock_database.Tx{}
	tcs := []struct {
		name     string
		data     *npb.EventStudentPackageV2
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "success with active class member",
			data: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "user-id-1",
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-id-1",
						LocationId: "location-id-1",
						ClassId:    "class-1",
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.Now(),
					},
					IsActive: true,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				classMemberRepo.On("GetByUserAndCourse", mock.Anything, db, "user-id-1", "course-id-1").Once().Return(make(map[string]*domain.ClassMember), nil)
				classMemberRepo.On("UpsertClassMember", mock.Anything, db, mock.Anything).Return(nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectMasterMgmtClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "success with active class member and already exists class member",
			data: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "user-id-1",
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-id-1",
						LocationId: "location-id-1",
						ClassId:    "class-1",
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.Now(),
					},
					IsActive: true,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				classMemberRepo.On("GetByUserAndCourse", mock.Anything, db, "user-id-1", "course-id-1").Once().Return(map[string]*domain.ClassMember{
					"user-id-1": {
						ClassMemberID: idutil.ULIDNow(),
						ClassID:       "class-1",
						UserID:        "user-id-1",
					},
				}, nil)
				classMemberRepo.On("DeleteByUserIDAndClassID", mock.Anything, db, "user-id-1", "class-1").Once().Return(nil)
				classMemberRepo.On("UpsertClassMember", mock.Anything, db, mock.Anything).Return(nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectMasterMgmtClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "success with update class_id of class_member",
			data: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "user-id-1",
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-id-1",
						LocationId: "location-id-1",
						ClassId:    "class-2",
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.Now(),
					},
					IsActive: true,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				classMemberRepo.On("GetByUserAndCourse", mock.Anything, db, "user-id-1", "course-id-1").Once().Return(map[string]*domain.ClassMember{
					"user-id-1": {
						ClassMemberID: idutil.ULIDNow(),
						ClassID:       "class-1",
						UserID:        "user-id-1",
					},
				}, nil)
				classMemberRepo.On("DeleteByUserIDAndClassID", mock.Anything, db, "user-id-1", "class-1").Once().Return(nil)
				classMemberRepo.On("UpsertClassMember", mock.Anything, db, mock.Anything).Return(nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectMasterMgmtClassUpserted, mock.Anything).Twice().Return("", nil)
			},
		},
		{
			name: "success to delete class of course in course detail",
			data: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "user-id-1",
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-id-1",
						LocationId: "location-id-1",
						ClassId:    "",
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.Now(),
					},
					IsActive: true,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				classMemberRepo.On("GetByUserAndCourse", mock.Anything, db, "user-id-1", "course-id-1").Once().Return(map[string]*domain.ClassMember{
					"user-id-1": {
						ClassMemberID: idutil.ULIDNow(),
						ClassID:       "class-1",
						UserID:        "user-id-1",
					},
				}, nil)
				classMemberRepo.On("DeleteByUserIDAndClassID", mock.Anything, db, "user-id-1", "class-1").Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectMasterMgmtClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "success with inactive class member",
			data: &npb.EventStudentPackageV2{
				StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
					StudentId: "user-id-1",
					Package: &npb.EventStudentPackageV2_PackageV2{
						CourseId:   "course-id-1",
						LocationId: "location-id-1",
						ClassId:    "class-1",
						StartDate:  timestamppb.Now(),
						EndDate:    timestamppb.Now(),
					},
					IsActive: false,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				classMemberRepo.On("GetByUserAndCourse", mock.Anything, db, "user-id-1", "course-id-1").Once().Return(map[string]*domain.ClassMember{
					"user-id-1": {
						ClassMemberID: idutil.ULIDNow(),
						ClassID:       "class-1",
						UserID:        "user-id-1",
					},
				}, nil)
				classMemberRepo.On("DeleteByUserIDAndClassID", mock.Anything, db, "user-id-1", "class-1").Return(nil).Once()
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectMasterMgmtClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
	}
	for _, tc := range tcs {
		handler := StudentPackageHandler{
			Logger:          ctxzap.Extract(ctx),
			DB:              db,
			ClassMemberRepo: classMemberRepo,
			JSM:             jsm,
		}
		tc.setup(ctx)
		t.Run(tc.name, func(t *testing.T) {
			msg, _ := proto.Marshal(tc.data)
			res, err := handler.Handle(ctx, msg)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, res)
			}
			mock.AssertExpectationsForObjects(t, db, classMemberRepo, jsm)
		})
	}
}
