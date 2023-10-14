package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentEnrollmentStatusHistoryRepoWithSqlMock() (*StudentEnrollmentStatusHistoryRepo, *testutil.MockDB) {
	studentEnrollmentStatusHistoryRepo := &StudentEnrollmentStatusHistoryRepo{}
	return studentEnrollmentStatusHistoryRepo, testutil.NewMockDB()
}

func TestGetLatestStatusByStudentIDAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentID := "1"
	locationID := "1"
	studentEnrollmentStatusHistoryRepoWithSqlMock, mockDB := StudentEnrollmentStatusHistoryRepoWithSqlMock()
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentID, locationID)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)
		studentEnrollmentStatusHistory, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetLatestStatusByStudentIDAndLocationID(ctx, mockDB.DB, studentID, locationID)
		assert.Nil(t, err)
		assert.NotNil(t, studentEnrollmentStatusHistory)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentID, locationID)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(fmt.Errorf("something error"), fields, values)
		_, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetLatestStatusByStudentIDAndLocationID(ctx, mockDB.DB, studentID, locationID)
		assert.NotNil(t, err)
	})
}

func TestGetCurrentStatusByStudentIDAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentID := "1"
	locationID := "1"
	studentEnrollmentStatusHistoryRepoWithSqlMock, mockDB := StudentEnrollmentStatusHistoryRepoWithSqlMock()
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentID, locationID)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)
		studentEnrollmentStatusHistory, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetCurrentStatusByStudentIDAndLocationID(ctx, mockDB.DB, studentID, locationID)
		assert.Nil(t, err)
		assert.NotNil(t, studentEnrollmentStatusHistory)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentID, locationID)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(fmt.Errorf("something error"), fields, values)
		_, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetCurrentStatusByStudentIDAndLocationID(ctx, mockDB.DB, studentID, locationID)
		assert.NotNil(t, err)
	})
}

func TestGetListStudentEnrollmentStatusHistoryByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentID := "1"
	studentEnrollmentStatusHistoryRepoWithSqlMock, mockDB := StudentEnrollmentStatusHistoryRepoWithSqlMock()
	rows := mockDB.Rows
	t.Run("happy case", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		studentEnrollmentStatusHistory, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetListStudentEnrollmentStatusHistoryByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.NotNil(t, studentEnrollmentStatusHistory)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetListStudentEnrollmentStatusHistoryByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, err)
	})
}

func Test_CheckEnrollmentStatusOfStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentID := "1"
	studentEnrollmentStatusHistoryRepoWithSqlMock, mockDB := StudentEnrollmentStatusHistoryRepoWithSqlMock()
	rows := mockDB.Rows
	t.Run("happy case", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		studentEnrollmentStatusHistory, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetListEnrolledStudentEnrollmentStatusByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.NotNil(t, studentEnrollmentStatusHistory)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetListEnrolledStudentEnrollmentStatusByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, err)
	})
}

func TestStudentEnrollmentStatusHistoryRepo_GetListByStudentIDAndTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	time := time.Now()
	studentID := "1"
	studentEnrollmentStatusHistoryRepoWithSqlMock, mockDB := StudentEnrollmentStatusHistoryRepoWithSqlMock()
	rows := mockDB.Rows
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		studentEnrollmentStatusHistory, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetListEnrolledStatusByStudentIDAndTime(ctx, mockDB.DB, studentID, time)
		assert.Nil(t, err)
		assert.NotNil(t, studentEnrollmentStatusHistory)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetListEnrolledStatusByStudentIDAndTime(ctx, mockDB.DB, studentID, time)
		assert.NotNil(t, err)
	})
}

func TestStudentEnrollmentStatusHistoryRepo_GetLatestStatusEnrollmentByStudentIDAndLocationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	locationIDs := []string{}
	studentID := "1"
	studentEnrollmentStatusHistoryRepoWithSqlMock, mockDB := StudentEnrollmentStatusHistoryRepoWithSqlMock()
	rows := mockDB.Rows
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		studentEnrollmentStatusHistory, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetLatestStatusEnrollmentByStudentIDAndLocationIDs(ctx, mockDB.DB, studentID, locationIDs)
		assert.Nil(t, err)
		assert.NotNil(t, studentEnrollmentStatusHistory)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		e := &entities.StudentEnrollmentStatusHistory{}
		fields, _ := e.FieldMap()
		scanFields := database.GetScanFields(e, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := studentEnrollmentStatusHistoryRepoWithSqlMock.GetLatestStatusEnrollmentByStudentIDAndLocationIDs(ctx, mockDB.DB, studentID, locationIDs)
		assert.NotNil(t, err)
	})
}
