package tom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tom_cmd "github.com/manabie-com/backend/cmd/server/tom"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) migrateConversationLocations(ctx context.Context) (context.Context, error) {
	err := tom_cmd.MigrateConversationLocations(ctx, tomConfig, s.CommonSuite.TomDBTrace, nil)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *suite) createStudentConversationWithNoLocation(ctx context.Context) (context.Context, error) {
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32(constants.ManabieSchool))
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	stu, err := s.CommonSuite.CreateStudent(ctx, []string{}, nil)
	if err != nil {
		return ctx, err
	}
	s.studentID = stu.UserProfile.UserId

	token, err := s.genStudentToken(s.studentID)
	if err != nil {
		return ctx, err
	}

	s.studentToken = token
	s.chatName = stu.UserProfile.Name
	schoolText := strconv.Itoa(constants.ManabieSchool)

	var (
		convID, stuID string
	)

	err = doRetry(func() (bool, error) {
		ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		err := s.DB.QueryRow(ctx2, `SELECT cs.conversation_id,cs.student_id FROM conversation_students cs LEFT JOIN conversations c ON cs.conversation_id = c.conversation_id
		WHERE cs.student_id = $1 AND owner = $2 AND c.status= 'CONVERSATION_STATUS_NONE' AND cs.conversation_type = 'CONVERSATION_STUDENT'`, database.Text(s.studentID), schoolText).Scan(&convID, &stuID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return true, err
			}
			return false, err
		}

		checkLocs := "select count(*) from conversation_locations cl where cl.conversation_id=$1"
		var count pgtype.Int8
		ctx3, cancel2 := context.WithTimeout(ctx, 2*time.Second)
		defer cancel2()
		if err := s.DB.QueryRow(ctx3, checkLocs, convID).Scan(&count); err != nil {
			return false, err
		}
		if count.Int != 0 {
			return false, fmt.Errorf("expected conversation %s has 0 locations but got %v locations", convID, count.Int)
		}

		return false, nil
	})

	if err != nil {
		return ctx, err
	}

	s.conversationID = convID

	return ctx, nil
}

func (s *suite) insertStudentAccessPaths(ctx context.Context) (context.Context, error) {
	cmd, err := s.DB.Exec(ctx, `INSERT INTO user_access_paths (
		user_id,
		location_id,
		created_at,
		updated_at,
		resource_path
	)
	VALUES ($1, $2, now(), now(), $3)`,
		s.studentID,
		constants.ManabieOrgLocation,
		strconv.Itoa(constants.ManabieSchool),
	)

	if err != nil {
		return ctx, err
	}

	if cmd.RowsAffected() == 0 {
		return ctx, fmt.Errorf("no rows affected")
	}

	return ctx, nil
}

func (s *suite) conversationLocationInserted(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var count pgtype.Int8
	if err := s.DB.QueryRow(ctx2, `SELECT COUNT(*) FROM conversation_locations WHERE conversation_id=$1 AND location_id=$2 AND access_path=$2`, database.Text(s.conversationID), constants.ManabieOrgLocation).Scan(&count); err != nil {
		return ctx, err
	}

	if count.Int != 1 {
		return ctx, fmt.Errorf("expect conversation has 1 location but got %d", count.Int)
	}
	return ctx, nil
}
