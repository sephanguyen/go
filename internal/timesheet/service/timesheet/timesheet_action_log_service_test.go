package timesheet

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestActionLogService_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		actionLogRepo = new(mock_repositories.MockTimesheetActionLogRepoImpl)
		db            = new(mock_database.Ext)
		staffId       = idutil.ULIDNow()
	)

	s := &ActionLogServiceImpl{
		DB:            db,
		ActionLogRepo: actionLogRepo,
	}
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserRoles(ctx, []string{constant.RoleSchoolAdmin}),
			req: &tpb.TimesheetActionLogRequest{
				TimesheetId: idutil.ULIDNow(),
				ExecutedBy:  staffId,
				Action:      tpb.TimesheetAction_EDITED,
				ExecutedAt:  timestamppb.Now(),
				IsSystem:    false,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				actionLogRepo.On("Create", ctx, db, mock.Anything).
					Return(nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*tpb.TimesheetActionLogRequest)
			err := s.Create(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
