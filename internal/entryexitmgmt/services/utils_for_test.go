package services

import (
	"context"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/entryexitmgmt/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

// nolint:unused,structcheck
type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

const (
	MockStudentEntryExitRecordsRepo = "MockStudentEntryExitRecordsRepo"
	MockStudentRepo                 = "MockStudentRepo"
	MockStudentParentRepo           = "MockStudentParentRepo"
	MockUserRepo                    = "MockUserRepo"
)

type mockRepos struct {
	MockStudentEntryExitRecordsRepo *mock_repositories.MockStudentEntryExitRecordsRepo
	MockStudentRepo                 *mock_repositories.MockStudentRepo
	MockStudentParentRepo           *mock_repositories.MockStudentParentRepo
	MockUserRepo                    *mock_repositories.MockUserRepo
}

func newMockRepos(repos ...string) *mockRepos {
	m := &mockRepos{}

	for _, repo := range repos {
		m.initMockRepo(repo)
	}

	return m
}

func (m *mockRepos) initMockRepo(repo string) {
	switch repo {
	case MockStudentEntryExitRecordsRepo:
		m.MockStudentEntryExitRecordsRepo = new(mock_repositories.MockStudentEntryExitRecordsRepo)
	case MockStudentRepo:
		m.MockStudentRepo = new(mock_repositories.MockStudentRepo)
	case MockStudentParentRepo:
		m.MockStudentParentRepo = new(mock_repositories.MockStudentParentRepo)
	case MockUserRepo:
		m.MockUserRepo = new(mock_repositories.MockUserRepo)
	}
}

type mockStudent struct {
	MockValidStudent                  *entities.Student
	MockInvalidStudentWithdrawnStatus *entities.Student
	MockInvalidStudentGraduatedStatus *entities.Student
	MockInvalidStudentLOAStatus       *entities.Student
}

func newMockStudent() *mockStudent {
	m := &mockStudent{}

	m.MockValidStudent = m.generateStudent(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED.String())
	m.MockInvalidStudentWithdrawnStatus = m.generateStudent(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String())
	m.MockInvalidStudentGraduatedStatus = m.generateStudent(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String())
	m.MockInvalidStudentLOAStatus = m.generateStudent(upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_LOA.String())

	return m
}

func (m *mockStudent) generateStudent(status string) *entities.Student {
	return &entities.Student{ID: database.Text("test")}
}

func initTestSetupAndServiceForCreateUpdate() (EntryExitModifierService, testSetup) {
	mockDB := new(mock_database.Ext)
	mockRepos := newMockRepos(
		MockStudentEntryExitRecordsRepo,
		MockStudentRepo,
		MockStudentParentRepo,
		MockUserRepo,
	)
	mockStudents := newMockStudent()
	mockIds := []string{"1", "2"}
	mockJsm := new(mock_nats.JetStreamManagement)
	user := &entities.User{
		FullName: database.Text("TestUser1"),
	}

	s := EntryExitModifierService{
		DB:                          mockDB,
		StudentEntryExitRecordsRepo: mockRepos.MockStudentEntryExitRecordsRepo,
		StudentRepo:                 mockRepos.MockStudentRepo,
		StudentParentRepo:           mockRepos.MockStudentParentRepo,
		JSM:                         mockJsm,
		UserRepo:                    mockRepos.MockUserRepo,
	}
	testSetup := testSetup{
		MockRepos:   mockRepos,
		MockStudent: mockStudents,
		MockDB:      mockDB,
		MockJSM:     mockJsm,
		MockIDs:     mockIds,
		User:        user,
	}

	return s, testSetup
}
