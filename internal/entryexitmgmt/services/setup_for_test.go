package services

import (
	"context"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/mock"
)

type testSetup struct {
	MockRepos   *mockRepos
	MockStudent *mockStudent
	MockDB      *mock_database.Ext
	MockJSM     *mock_nats.JetStreamManagement
	MockIDs     []string
	User        *entities.User
}

func (s *testSetup) EmptySetup() func(context.Context) {
	return func(context.Context) {
		// no setup needed
	}
}

func (s *testSetup) InvalidStudentSetup(student *entities.Student) func(context.Context) {
	return func(ctx context.Context) {
		s.MockRepos.MockStudentRepo.On("FindByID", ctx, s.MockDB, mock.Anything).Once().Return(student, nil)
	}
}

func (s *testSetup) StudentNotExistSetup() func(context.Context) {
	return func(ctx context.Context) {
		s.MockRepos.MockStudentRepo.On("FindByID", ctx, s.MockDB, mock.Anything).Once().Return(s.MockStudent.MockValidStudent, puddle.ErrClosedPool)
	}
}

func (s *testSetup) UserSetup(returnArguments ...interface{}) func(context.Context) {
	return func(ctx context.Context) {
		s.MockRepos.MockStudentRepo.On("FindByID", ctx, s.MockDB, mock.Anything).Once().Return(s.MockStudent.MockValidStudent, nil)
		s.MockRepos.MockUserRepo.On("FindByID", ctx, s.MockDB, mock.Anything).Once().Return(returnArguments...)
	}
}

func (s *testSetup) DefaultCreateSetup(ctx context.Context, returnArguments ...interface{}) {
	s.MockRepos.MockStudentRepo.On("FindByID", ctx, s.MockDB, mock.Anything).Once().Return(s.MockStudent.MockValidStudent, nil)
	s.MockRepos.MockUserRepo.On("FindByID", ctx, s.MockDB, mock.Anything).Once().Return(s.User, nil)
	s.MockRepos.MockStudentEntryExitRecordsRepo.On("Create", ctx, s.MockDB, mock.Anything).Once().Return(returnArguments...)
}

func (s *testSetup) CreateSetup() func(context.Context) {
	return func(ctx context.Context) {
		s.DefaultCreateSetup(ctx, nil)
	}
}

func (s *testSetup) CreateWithNotifSetup() func(context.Context) {
	return func(ctx context.Context) {
		s.DefaultCreateSetup(ctx, nil)
		s.MockRepos.MockStudentParentRepo.On("GetParentIDsByStudentID", ctx, s.MockDB, mock.Anything).Once().Return(s.MockIDs, nil)
		s.MockJSM.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
	}
}

func (s *testSetup) CreateWithGetParentIDSetup(returnArguments ...interface{}) func(context.Context) {
	return func(ctx context.Context) {
		s.DefaultCreateSetup(ctx, nil)
		s.MockRepos.MockStudentParentRepo.On("GetParentIDsByStudentID", ctx, s.MockDB, mock.Anything).Once().Return(returnArguments...)
	}
}

func (s *testSetup) CreateWithFailedNotifSetup(returnArguments ...interface{}) func(context.Context) {
	return func(ctx context.Context) {
		s.CreateWithGetParentIDSetup(s.MockIDs, nil)(ctx)
		for i := 0; i < 4; i++ {
			s.MockJSM.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(returnArguments...)
		}
	}
}

func (s *testSetup) DefaultUpdateSetup(ctx context.Context, returnArguments ...interface{}) {
	s.MockRepos.MockStudentRepo.On("FindByID", ctx, s.MockDB, mock.Anything).Once().Return(s.MockStudent.MockValidStudent, nil)
	s.MockRepos.MockUserRepo.On("FindByID", ctx, s.MockDB, mock.Anything).Once().Return(s.User, nil)
	s.MockRepos.MockStudentEntryExitRecordsRepo.On("Update", ctx, s.MockDB, mock.Anything).Once().Return(returnArguments...)
}

func (s *testSetup) UpdateSetup() func(context.Context) {
	return func(ctx context.Context) {
		s.DefaultUpdateSetup(ctx, nil)
	}
}

func (s *testSetup) UpdateWithNotifSetup() func(context.Context) {
	return func(ctx context.Context) {
		s.DefaultUpdateSetup(ctx, nil)
		s.MockRepos.MockStudentParentRepo.On("GetParentIDsByStudentID", ctx, s.MockDB, mock.Anything).Once().Return(s.MockIDs, nil)
		s.MockJSM.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
	}
}

func (s *testSetup) UpdateWithGetParentIDSetup(returnArguments ...interface{}) func(context.Context) {
	return func(ctx context.Context) {
		s.DefaultUpdateSetup(ctx, nil)
		s.MockRepos.MockStudentParentRepo.On("GetParentIDsByStudentID", ctx, s.MockDB, mock.Anything).Once().Return(returnArguments...)
	}
}

func (s *testSetup) UpdateWithFailedNotifSetup(returnArguments ...interface{}) func(context.Context) {
	return func(ctx context.Context) {
		s.UpdateWithGetParentIDSetup(s.MockIDs, nil)(ctx)
		for i := 0; i < 4; i++ {
			s.MockJSM.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(returnArguments...)
		}
	}
}
