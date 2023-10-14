package classes

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	golibs_constants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_master "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"
	mock_yasuo "github.com/manabie-com/backend/mock/yasuo/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClassService_CreateClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	configRepo := new(mock_repositories.MockConfigRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	schoolAdminRepo := new(mock_repositories.MockSchoolAdminRepo)
	classRepo := new(mock_repositories.MockClassRepo)
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	schoolConfigRepo := new(mock_repositories.MockSchoolConfigRepo)
	jsm := new(mock_nats.JetStreamManagement)
	masterClassRepo := new(mock_master.MockClassRepo)
	masterClassMemberRepo := new(mock_master.MockClassMemberRepo)

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	s := &ClassService{
		DB:                    mockDB,
		UserRepo:              userRepo,
		ClassRepo:             classRepo,
		ClassMemberRepo:       classMemberRepo,
		ConfigRepo:            configRepo,
		SchoolConfigRepo:      schoolConfigRepo,
		TeacherRepo:           teacherRepo,
		SchoolAdminRepo:       schoolAdminRepo,
		MasterClassRepo:       masterClassRepo,
		MasterClassMemberRepo: masterClassMemberRepo,
		JSM:                   jsm,
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	testCases := []TestCase{
		{
			name:         "missing class name",
			ctx:          ctx,
			req:          &pb.CreateClassRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "missing className"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "not found user",
			ctx:          ctx,
			req:          &pb.CreateClassRequest{ClassName: "class-name"},
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, "can't find current user"),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:         "admin create with ownerId empty",
			ctx:          ctx,
			req:          &pb.CreateClassRequest{ClassName: "class-name"},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "missing ownerId"),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(&entities.User{
					Group: database.Text(pb.USER_GROUP_ADMIN.String()),
				}, nil)
			},
		},
		{
			name:         "admin create with wrong ownerId",
			ctx:          ctx,
			req:          &pb.CreateClassRequest{ClassName: "class-name", OwnerIds: []string{"owner-id"}},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(&entities.User{
					Group: database.Text(pb.USER_GROUP_ADMIN.String()),
				}, nil)

				userRepo.On("Find", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "missing config avatar",
			ctx:  ctx,
			req: &pb.CreateClassRequest{
				SchoolId:  0,
				ClassName: "class-name",
				Grades:    []string{"G11", "G12"},
				Subjects:  []pb.Subject{pb.SUBJECT_ENGLISH},
				OwnerIds:  []string{"owner-id"},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, errors.Wrap(pgx.ErrNoRows, "rcv.ConfigRepo.Find").Error()),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(&entities.User{
					Group:   database.Text(pb.USER_GROUP_ADMIN.String()),
					Country: database.Text(pb.COUNTRY_VN.String()),
				}, nil)

				userRepo.On("Find", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.User{{
					Group:    database.Text(pb.USER_GROUP_TEACHER.String()),
					ID:       database.Text("owner-id"),
					LastName: database.Text("owner-name"),
				}}, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)

				configRepo.On("Find", mock.AnythingOfType("*context.valueCtx"), mockTxer, database.Text(pb.COUNTRY_MASTER.String()), database.Text(classAvatar)).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "err when create class",
			ctx:  ctx,
			req: &pb.CreateClassRequest{
				SchoolId:  0,
				ClassName: "class-name",
				Grades:    []string{"G11", "G12"},
				Subjects:  []pb.Subject{pb.SUBJECT_ENGLISH},
				OwnerIds:  []string{"owner-id"},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, errors.Wrap(pgx.ErrTxClosed, "rcv.ClassRepo.Create").Error()),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(&entities.User{
					Group:   database.Text(pb.USER_GROUP_ADMIN.String()),
					Country: database.Text(pb.COUNTRY_VN.String()),
				}, nil)

				userRepo.On("Find", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.User{{
					Group:    database.Text(pb.USER_GROUP_TEACHER.String()),
					ID:       database.Text("owner-id"),
					LastName: database.Text("owner-name"),
				}}, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)

				configRepo.On("Find", mock.AnythingOfType("*context.valueCtx"), mockTxer, database.Text(pb.COUNTRY_MASTER.String()), database.Text(classAvatar)).Once().Return([]*entities.Config{{}}, nil)
				schoolConfigRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("pgtype.Int4")).
					Once().Return(&entities.SchoolConfig{}, nil)

				classRepo.On("GetNextID", mock.AnythingOfType("*context.valueCtx"), mockTxer).
					Once().Return(&pgtype.Int4{}, nil)
				classRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.Class")).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "err create class member",
			ctx:  ctx,
			req: &pb.CreateClassRequest{
				SchoolId:  0,
				ClassName: "class-name",
				Grades:    []string{"G11", "G12"},
				Subjects:  []pb.Subject{pb.SUBJECT_ENGLISH},
				OwnerIds:  []string{"owner-id"},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, errors.Wrap(pgx.ErrTxClosed, "rcv.createClassMember: rcv.ClassMemberRepo.Create").Error()),
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(&entities.User{
					Group:   database.Text(pb.USER_GROUP_ADMIN.String()),
					Country: database.Text(pb.COUNTRY_VN.String()),
				}, nil)

				userRepo.On("Find", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.User{{
					Group:    database.Text(pb.USER_GROUP_TEACHER.String()),
					ID:       database.Text("owner-id"),
					LastName: database.Text("owner-name"),
				}}, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Rollback", mock.Anything).Return(nil)

				configRepo.On("Find", mock.AnythingOfType("*context.valueCtx"), mockTxer, database.Text(pb.COUNTRY_MASTER.String()), database.Text(classAvatar)).Once().Return([]*entities.Config{{}}, nil)

				schoolConfigRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("pgtype.Int4")).
					Once().Return(&entities.SchoolConfig{}, nil)

				classRepo.On("GetNextID", mock.AnythingOfType("*context.valueCtx"), mockTxer).
					Once().Return(&pgtype.Int4{}, nil)

				classRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.Class")).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)

				masterClassRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(&domain.Class{
					ClassID: "1",
				}, nil)
				masterClassMemberRepo.On("GetByClassIDAndUserIDs", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything, mock.Anything).Once().Return(nil, nil)
				masterClassMemberRepo.On("UpsertClassMembers", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(nil)
				classMemberRepo.On("FindByIDs", ctx, mockTxer, mock.AnythingOfType("pgtype.Int4"), database.TextArray([]string{"owner-id"}), database.Text(entities.ClassMemberStatusActive)).
					Once().Return(nil, pgx.ErrNoRows)

				classMemberRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.ClassMember")).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req: &pb.CreateClassRequest{
				SchoolId:  0,
				ClassName: "class-name",
				Grades:    []string{"G11", "G12"},
				Subjects:  []pb.Subject{pb.SUBJECT_ENGLISH},
				OwnerIds:  []string{"owner-id"},
			},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(&entities.User{
					Group:   database.Text(pb.USER_GROUP_ADMIN.String()),
					Country: database.Text(pb.COUNTRY_VN.String()),
				}, nil)

				userRepo.On("Find", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.User{{
					Group:    database.Text(pb.USER_GROUP_TEACHER.String()),
					ID:       database.Text("owner-id"),
					LastName: database.Text("owner-name"),
				}}, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				configRepo.On("Find", mock.AnythingOfType("*context.valueCtx"), mockTxer, database.Text(pb.COUNTRY_MASTER.String()), database.Text(classAvatar)).Once().Return([]*entities.Config{{}}, nil)

				schoolConfigRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("pgtype.Int4")).
					Once().Return(&entities.SchoolConfig{}, nil)

				classRepo.On("GetNextID", mock.AnythingOfType("*context.valueCtx"), mockTxer).
					Once().Return(&pgtype.Int4{}, nil)
				classRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.Class")).Once().Return(nil)

				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)
				masterClassRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(&domain.Class{
					ClassID: "1",
				}, nil)
				masterClassMemberRepo.On("GetByClassIDAndUserIDs", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything, mock.Anything).Once().Return(nil, nil)
				masterClassMemberRepo.On("UpsertClassMembers", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(nil)
				classMemberRepo.On("FindByIDs", ctx, mockTxer, mock.AnythingOfType("pgtype.Int4"), database.TextArray([]string{"owner-id"}), database.Text(entities.ClassMemberStatusActive)).
					Once().Return(nil, pgx.ErrNoRows)

				classMemberRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.ClassMember")).Once().Return(nil)

				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:         "teacher create class",
			ctx:          ctx,
			req:          &pb.CreateClassRequest{ClassName: "class-name", SchoolId: 1},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(&entities.User{
					ID:      database.Text(userID),
					Group:   database.Text(pb.USER_GROUP_TEACHER.String()),
					Country: database.Text(pb.COUNTRY_VN.String()),
				}, nil)

				teacherRepo.On("IsInSchool", ctx, mockDB, database.TextArray([]string{userID}), database.Int4(1)).Once().Return(true, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				configRepo.On("Find", mock.AnythingOfType("*context.valueCtx"), mockTxer, database.Text(pb.COUNTRY_MASTER.String()), database.Text(classAvatar)).Once().Return([]*entities.Config{{}}, nil)
				schoolConfigRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("pgtype.Int4")).
					Once().Return(&entities.SchoolConfig{}, nil)

				classRepo.On("GetNextID", mock.AnythingOfType("*context.valueCtx"), mockTxer).
					Once().Return(&pgtype.Int4{}, nil)
				classRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.Class")).Once().Return(nil)

				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)
				masterClassRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(&domain.Class{
					ClassID: "1",
				}, nil)
				masterClassMemberRepo.On("GetByClassIDAndUserIDs", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything, mock.Anything).Once().Return(nil, nil)
				masterClassMemberRepo.On("UpsertClassMembers", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(nil)
				classMemberRepo.On("FindByIDs", ctx, mockTxer, mock.AnythingOfType("pgtype.Int4"), mock.AnythingOfType("pgtype.TextArray"), database.Text(entities.ClassMemberStatusActive)).
					Once().Return(nil, pgx.ErrNoRows)

				classMemberRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.ClassMember")).Once().Return(nil)

				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:         "school admin create class",
			ctx:          ctx,
			req:          &pb.CreateClassRequest{ClassName: "class-name", SchoolId: 1, OwnerIds: []string{"owner-id"}},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("Get", ctx, mock.Anything, database.Text(userID)).Once().Return(&entities.User{
					ID:      database.Text(userID),
					Group:   database.Text(pb.USER_GROUP_SCHOOL_ADMIN.String()),
					Country: database.Text(pb.COUNTRY_VN.String()),
				}, nil)

				schoolAdminRepo.On("Get", ctx, mockDB, database.Text(userID)).Once().Return(&entities.SchoolAdmin{
					SchoolAdminID: database.Text(userID),
					SchoolID:      database.Int4(1),
				}, nil)

				userRepo.On("Find", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.User{{
					Group:    database.Text(pb.USER_GROUP_TEACHER.String()),
					ID:       database.Text("owner-id"),
					LastName: database.Text("owner-name"),
				}}, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				configRepo.On("Find", mock.AnythingOfType("*context.valueCtx"), mockTxer, database.Text(pb.COUNTRY_MASTER.String()), database.Text(classAvatar)).Once().Return([]*entities.Config{{}}, nil)
				schoolConfigRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("pgtype.Int4")).
					Once().Return(&entities.SchoolConfig{}, nil)

				classRepo.On("GetNextID", mock.AnythingOfType("*context.valueCtx"), mockTxer).
					Once().Return(&pgtype.Int4{}, nil)
				classRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.Class")).Once().Return(nil)

				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)
				masterClassRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(&domain.Class{
					ClassID: "1",
				}, nil)
				masterClassMemberRepo.On("GetByClassIDAndUserIDs", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything, mock.Anything).Once().Return(nil, nil)
				masterClassMemberRepo.On("UpsertClassMembers", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(nil)
				classMemberRepo.On("FindByIDs", ctx, mockTxer, mock.AnythingOfType("pgtype.Int4"), database.TextArray([]string{"owner-id"}), database.Text(entities.ClassMemberStatusActive)).
					Once().Return(nil, pgx.ErrNoRows)

				classMemberRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.ClassMember")).Once().Return(nil)

				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.CreateClass(testCase.ctx, testCase.req.(*pb.CreateClassRequest))
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}

			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestGroupSubmissions(t *testing.T) {
	t.Parallel()
	submissions := map[string][]*repositories.SubmissionResult{
		"lo1": {
			{
				QuestionID: "q1",
				Correct:    true,
				SessionID:  "s1",
			},
			{
				QuestionID: "q1",
				Correct:    false,
				SessionID:  "s2",
			},
			{
				QuestionID: "q2",
				Correct:    false,
				SessionID:  "s1",
			},
			{
				QuestionID: "q3",
				Correct:    true,
				SessionID:  "s1",
			},
			{
				QuestionID: "q2",
				Correct:    true,
				SessionID:  "s2",
			},
		},
		"lo2": {
			{
				QuestionID: "q1",
				Correct:    false,
				SessionID:  "s3",
			},
			{
				QuestionID: "q2",
				Correct:    true,
				SessionID:  "s3",
			},
			{
				QuestionID: "q2",
				Correct:    true,
				SessionID:  "s4",
			},
			{
				QuestionID: "q3",
				Correct:    true,
				SessionID:  "s3",
			},
		},
		"lo3": {
			{
				QuestionID: "q1",
				Correct:    true,
				SessionID:  "s5",
			},
		},
	}

	expected := map[string][][]*repositories.SubmissionResult{
		"lo1": {
			{
				{
					QuestionID: "q1",
					Correct:    true,
					SessionID:  "s1",
				},
				{
					QuestionID: "q2",
					Correct:    false,
					SessionID:  "s1",
				},
				{
					QuestionID: "q3",
					Correct:    true,
					SessionID:  "s1",
				},
			},
			{
				{
					QuestionID: "q1",
					Correct:    false,
					SessionID:  "s2",
				},
				{
					QuestionID: "q2",
					Correct:    true,
					SessionID:  "s2",
				},
			},
		},
		"lo2": {
			{
				{
					QuestionID: "q1",
					Correct:    false,
					SessionID:  "s3",
				},
				{
					QuestionID: "q2",
					Correct:    true,
					SessionID:  "s3",
				},
				{
					QuestionID: "q3",
					Correct:    true,
					SessionID:  "s3",
				},
			},
			{
				{
					QuestionID: "q2",
					Correct:    true,
					SessionID:  "s4",
				},
			},
		},
		"lo3": {
			{
				{
					QuestionID: "q1",
					Correct:    true,
					SessionID:  "s5",
				},
			},
		},
	}

	for k, v := range submissions {
		r := groupSubmissions(v)
		if diff := cmp.Diff(expected[k], r); diff != "" {
			t.Errorf("groupSubmissions() mismatch (-want, +got):\n%s", diff)
		}
	}
}

// deprecated services
func TestClassService_RetrieveActiveClassAssignment(t *testing.T) {
}

// deprecated services
func TestClassService_RetrievePastClassAssignment(t *testing.T) {
}

func TestClassService_SyncClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	configRepo := new(mock_repositories.MockConfigRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	classRepo := new(mock_repositories.MockClassRepo)
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	schoolConfigRepo := new(mock_repositories.MockSchoolConfigRepo)
	masterClassRepo := new(mock_master.MockClassRepo)
	jsm := new(mock_nats.JetStreamManagement)
	yasuoCourseClass := new(mock_yasuo.MockCourseClassRepo)

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	s := &ClassService{
		DB:                   mockDB,
		ClassRepo:            classRepo,
		ClassMemberRepo:      classMemberRepo,
		ConfigRepo:           configRepo,
		JSM:                  jsm,
		CourseRepo:           courseRepo,
		MasterClassRepo:      masterClassRepo,
		SchoolConfigRepo:     schoolConfigRepo,
		YasuoCourseClassRepo: yasuoCourseClass,
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	type TestCase struct {
		name         string
		ctx          context.Context
		classes      []*npb.EventMasterRegistration_Class
		req          interface{}
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	testCases := []TestCase{
		{
			name:         "sync create new class successfully",
			ctx:          ctx,
			expectedResp: nil,
			classes: []*npb.EventMasterRegistration_Class{
				{
					ClassId:    1,
					CourseId:   "course-1",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				classID := pgtype.Int4{Int: 1, Status: pgtype.Present}
				classRepo.On("FindByID", ctx, mockDB, classID).Once().Return(nil, nil)

				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)

				configRepo.On("Find", mock.AnythingOfType("*context.valueCtx"), mockTxer, database.Text(pb.COUNTRY_MASTER.String()), database.Text(classAvatar)).Once().Return([]*entities.Config{{}}, nil)
				schoolConfigRepo.On("FindByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("pgtype.Int4")).
					Once().Return(&entities.SchoolConfig{}, nil)

				classRepo.On("GetNextID", mock.AnythingOfType("*context.valueCtx"), mockTxer).
					Once().Return(&pgtype.Int4{}, nil)
				classRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.AnythingOfType("*entities.Class")).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Twice().Return("", nil)
				yasuoCourseClass.On("SoftDeleteClass", ctx, mockTxer, classID).Once().Return(nil)
				yasuoCourseClass.On("UpsertV2", ctx, mockTxer, mock.Anything).Once().Return(nil)

				courseRepo.On("FindByID", ctx, mockDB, database.Text("course-1")).Once().Return(&entities.Course{ID: database.Text("course-1")}, nil)
				masterClassRepo.On("UpsertClasses", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "sync update class successfully",
			ctx:          ctx,
			expectedResp: nil,
			classes: []*npb.EventMasterRegistration_Class{
				{
					ClassId:    1,
					CourseId:   "course-1",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				classID := pgtype.Int4{Int: 1, Status: pgtype.Present}
				classRepo.On("FindByID", ctx, mockDB, classID).Once().Return(&entities.Class{ID: classID, SchoolID: database.Int4(1)}, nil)
				classRepo.On("Update", mock.AnythingOfType("*context.valueCtx"), mockDB, mock.AnythingOfType("*entities.Class")).Once().Return(nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				yasuoCourseClass.On("SoftDeleteClass", ctx, mockTxer, classID).Once().Return(nil)
				yasuoCourseClass.On("UpsertV2", ctx, mockTxer, mock.Anything).Once().Return(nil)

				courseRepo.On("FindByID", ctx, mockDB, database.Text("course-1")).Once().Return(&entities.Course{ID: database.Text("course-1")}, nil)
				masterClassRepo.On("UpsertClasses", ctx, mockDB, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:         "sync update class fail",
			ctx:          ctx,
			expectedResp: nil,
			classes: []*npb.EventMasterRegistration_Class{
				{
					ClassId:    1,
					CourseId:   "course-1",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				},
			},
			expectedErr: errors.New("err findClass: can't get class"),
			setup: func(ctx context.Context) {
				classID := pgtype.Int4{Int: 1, Status: pgtype.Present}
				classRepo.On("FindByID", ctx, mockDB, classID).Once().Return(nil, errors.New("can't get class"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.SyncClass(testCase.ctx, testCase.classes)
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}
		})
	}
}

func TestClassService_SyncClassMember(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	configRepo := new(mock_repositories.MockConfigRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	classRepo := new(mock_repositories.MockClassRepo)
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	schoolConfigRepo := new(mock_repositories.MockSchoolConfigRepo)
	masterClassRepo := new(mock_master.MockClassRepo)
	masterClassMemberRepo := new(mock_master.MockClassMemberRepo)
	jsm := new(mock_nats.JetStreamManagement)
	yasuoCourseClass := new(mock_yasuo.MockCourseClassRepo)
	topicRepo := new(mock_repositories.MockTopicRepo)

	mockDB := &mock_database.Ext{}
	mockTxer := &mock_database.Tx{}

	s := &ClassService{
		DB:                    mockDB,
		ClassRepo:             classRepo,
		ClassMemberRepo:       classMemberRepo,
		ConfigRepo:            configRepo,
		JSM:                   jsm,
		CourseRepo:            courseRepo,
		UserRepo:              userRepo,
		MasterClassRepo:       masterClassRepo,
		SchoolConfigRepo:      schoolConfigRepo,
		MasterClassMemberRepo: masterClassMemberRepo,
		YasuoCourseClassRepo:  yasuoCourseClass,
		TopicRepo:             topicRepo,
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	type TestCase struct {
		name         string
		ctx          context.Context
		classMembers []*npb.EventUserRegistration_Student
		req          interface{}
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	testCases := []TestCase{
		{
			name:         "sync create new class member successfully",
			ctx:          ctx,
			expectedResp: nil,
			classMembers: []*npb.EventUserRegistration_Student{
				{
					StudentId:  "student-1",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					Packages: []*npb.EventUserRegistration_Student_Package{
						{
							ClassId: int64(1),
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				classID := pgtype.Int4{Int: 1, Status: pgtype.Present}
				classMemberRepo.On("ClassJoinNotIn", ctx, mockDB, database.Text("student-1"), database.Int4Array([]int32{1})).Once().Return([]int32{}, nil)
				classRepo.On("FindByID", ctx, mockDB, classID).Once().Return(&entities.Class{ID: classID, SchoolID: database.Int4(1), Code: database.Text("code-1")}, nil)
				classRepo.On("FindByCode", mock.AnythingOfType("*context.valueCtx"), mockDB, database.Text("code-1")).Once().Return(&entities.Class{ID: classID, SchoolID: database.Int4(1), Code: database.Text("code-1")}, nil)
				classMemberRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mockDB, pgtype.Int4{Int: 1, Status: pgtype.Present}, database.Text("student-1"), database.Text(entities.ClassMemberStatusActive)).Once().Return(nil, nil)
				userRepo.On("Get", mock.AnythingOfType("*context.valueCtx"), mockDB, database.Text("student-1")).Once().Return(&entities.User{ID: database.Text("student-1"), Group: database.Text(pb.USER_GROUP_STUDENT.String())}, nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				masterClassRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, "1").Once().Return(&domain.Class{
					ClassID: "1",
				}, nil)
				masterClassMemberRepo.On("GetByClassIDAndUserIDs", mock.AnythingOfType("*context.valueCtx"), mockTxer, "1", []string{"student-1"}).Once().Return(nil, nil)
				masterClassMemberRepo.On("UpsertClassMembers", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(nil)
				classMemberRepo.On("FindByIDs", mock.AnythingOfType("*context.valueCtx"), mockTxer, database.Int4(1), database.TextArray([]string{"student-1"}), database.Text(entities.ClassMemberStatusActive)).Once().Return(nil, nil)
				classMemberRepo.On("Create", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)
				classRepo.On("FindByID", ctx, mockDB, classID).Once().Return(&entities.Class{ID: classID, SchoolID: database.Int4(1), Code: database.Text("code-1")}, nil)
			},
		},
		{
			name:         "sync delete class member successfully",
			ctx:          ctx,
			expectedResp: nil,
			classMembers: []*npb.EventUserRegistration_Student{
				{
					StudentId:  "student-1",
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					Packages: []*npb.EventUserRegistration_Student_Package{
						{
							ClassId: int64(1),
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				classID := pgtype.Int4{Int: 1, Status: pgtype.Present}
				classMemberRepo.On("ClassJoinNotIn", ctx, mockDB, database.Text("student-1"), database.Int4Array([]int32{1})).Once().Return([]int32{}, nil)
				classRepo.On("FindByID", ctx, mockDB, classID).Once().Return(&entities.Class{ID: classID, SchoolID: database.Int4(1), Code: database.Text("code-1")}, nil)
				masterClassRepo.On("GetByID", mock.AnythingOfType("*context.valueCtx"), mockTxer, "1").Once().Return(&domain.Class{
					ClassID: "1",
				}, nil)
				masterClassMemberRepo.On("GetByClassIDAndUserIDs", mock.AnythingOfType("*context.valueCtx"), mockTxer, "1", []string{"student-1"}).Once().Return(nil, nil)
				masterClassMemberRepo.On("UpsertClassMembers", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything).Once().Return(nil)
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTxer, nil)
				mockTxer.On("Commit", mock.Anything).Return(nil)
				classMemberRepo.On("UpdateStatus", mock.AnythingOfType("*context.valueCtx"), mockTxer, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				jsm.On("PublishAsyncContext", mock.Anything, golibs_constants.SubjectClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:         "err create new class member successfully",
			ctx:          ctx,
			expectedResp: nil,
			classMembers: []*npb.EventUserRegistration_Student{
				{
					StudentId:  "student-1",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					Packages: []*npb.EventUserRegistration_Student_Package{
						{
							ClassId: int64(1),
						},
					},
				},
			},
			expectedErr: errors.New("err JoinClass: 1, studentID: student-1, err: rpc error: code = Unknown desc = can't get class"),
			setup: func(ctx context.Context) {
				classID := pgtype.Int4{Int: 1, Status: pgtype.Present}
				classMemberRepo.On("ClassJoinNotIn", ctx, mockDB, database.Text("student-1"), database.Int4Array([]int32{1})).Once().Return([]int32{}, nil)
				classRepo.On("FindByID", ctx, mockDB, classID).Once().Return(nil, nil)
				classRepo.On("FindByCode", mock.AnythingOfType("*context.valueCtx"), mockDB, database.Text("code-1")).Once().Return(nil, errors.New("can't get class"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.SyncClassMember(testCase.ctx, testCase.classMembers)
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}
		})
	}
}
