package services

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateAuth(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userId := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userId)

	userRepo := new(mock_repositories.MockUserRepo)

	testCases := []TestCase{
		{
			name:         "err db query user group",
			ctx:          ctx,
			req:          userRepo.UserGroup,
			expectedResp: "",
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return("", pgx.ErrNoRows)
			},
		},
		{
			name:         "err permission denied",
			ctx:          ctx,
			req:          userRepo.UserGroup,
			expectedResp: "",
			expectedErr:  status.Error(codes.PermissionDenied, "user group not allowed"),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name:         "success",
			ctx:          ctx,
			req:          userRepo.UserGroup,
			expectedResp: userId,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities_bob.UserGroupAdmin, nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(testCase.ctx)
		t.Run(testCase.name, func(t *testing.T) {
			resp, err := ValidateAuth(testCase.ctx, &mock_database.Ext{}, testCase.req.(func(context.Context, database.QueryExecer, pgtype.Text) (string, error)))
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestEncodeOrderIDNumber2String(t *testing.T) {
	t.Parallel()
	localRand := rand.New(rand.NewSource(time.Now().Unix()))

	// for number < 999999999 alway return string decoded with fixed len 10
	for i := 0; i < 5; i++ {
		n := localRand.Intn(999999999)
		s := EncodeOrderID2String(uint(n))
		decoded, err := DecodeString2OrderID(s)
		assert.Nil(t, err)
		assert.Equal(t, uint(n), decoded)
		assert.Equal(t, 10, len(s))
	}

	n := uint(1000000000)
	s := EncodeOrderID2String(n)
	decoded, err := DecodeString2OrderID(s)
	assert.Nil(t, err)
	assert.Equal(t, n, decoded)
	assert.Equal(t, 11, len(s))
}

func TestEncodeDecodeNumber2String(t *testing.T) {
	t.Parallel()
	localRand := rand.New(rand.NewSource(time.Now().Unix()))

	// for number < 999999999 alway return string decoded with fixed len 8
	for i := 0; i < 5; i++ {
		n := localRand.Intn(999999999)
		s := EncodeNumber2String(uint(n))
		decoded, err := DecodeString2Number(s)
		assert.Nil(t, err)
		assert.Equal(t, uint(n), decoded)
		assert.Equal(t, 8, len(s))
	}

	n := uint(1000000000)
	s := EncodeNumber2String(n)
	decoded, err := DecodeString2Number(s)
	assert.Nil(t, err)
	assert.Equal(t, n, decoded)
	assert.Equal(t, 9, len(s))
}

func TestFormatPromotionCodeExpiredDate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		c           pb.Country
		expiredDate time.Time
		expected    string
	}{
		{
			c:           pb.COUNTRY_VN,
			expiredDate: time.Date(2020, 12, 20, 0, 0, 0, 0, time.UTC),
			expected:    "20/12",
		},
		{
			c:           pb.COUNTRY_VN,
			expiredDate: time.Date(2020, 12, 20, 18, 0, 0, 0, time.UTC),
			expected:    "21/12",
		},
		{
			c:           pb.COUNTRY_MASTER,
			expiredDate: time.Date(2020, 12, 20, 0, 0, 0, 0, time.UTC),
			expected:    "December 20",
		},
		{
			c:           pb.COUNTRY_MASTER,
			expiredDate: time.Date(2020, 12, 20, 18, 0, 0, 0, time.UTC),
			expected:    "December 20",
		},
		{
			c:           pb.COUNTRY_NONE,
			expiredDate: time.Date(2020, 12, 20, 0, 0, 0, 0, time.UTC),
			expected:    "December 20",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s_%s", tc.c.String(), tc.expiredDate.Format(time.RFC3339)), func(t *testing.T) {
			t.Parallel()
			got := FormatPromotionCodeExpiredDate(tc.c, tc.expiredDate)
			if got != tc.expected {
				t.Errorf("formatPromotionCodeExpiredDate(%s, %v) = %q, want %q", tc.c.String(), tc.expiredDate, got, tc.expected)
			}
		})
	}
}
