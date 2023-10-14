package virtualclassroom

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type LiveLessonList struct {
	ListCfg     LessonInfoConfig
	LessonsList []LessonListItem
}

type LessonListItem struct {
	LessonID  string
	StudentID string
}

func (l *LiveLessonList) GetOneLessonWithStudentID(ctx context.Context, db database.Ext, schoolID string) (string, string, error) {
	var lessonID, studentID pgtype.Text

	lessonQuery := `SELECT l.lesson_id 
			FROM public.lessons l
			WHERE l.course_id = $1 
			AND l.center_id = $2 
			AND l.resource_path = $3
			AND l.deleted_at IS NULL
			LIMIT 1 `
	if err := db.QueryRow(ctx, lessonQuery, l.ListCfg.CourseID, l.ListCfg.LocationID, schoolID).Scan(&lessonID); err != nil {
		return "", "", fmt.Errorf("failed to get lesson ID using course_id %s and center_id %s: %w",
			l.ListCfg.CourseID,
			l.ListCfg.LocationID,
			err,
		)
	}

	lessonMemberQuery := `SELECT lm.user_id 
			FROM public.lesson_members lm
			WHERE lm.course_id = $1 
			AND lm.lesson_id = $2 
			AND lm.resource_path = $3
			AND lm.deleted_at IS NULL
			LIMIT 1 `
	if err := db.QueryRow(ctx, lessonMemberQuery, l.ListCfg.CourseID, lessonID.String, schoolID).Scan(&studentID); err != nil {
		return "", "", fmt.Errorf("failed to get lesson member ID using course_id %s and lesson_id %s: %w",
			l.ListCfg.CourseID,
			lessonID.String,
			err,
		)
	}

	return lessonID.String, studentID.String, nil
}

func (l *LiveLessonList) GetMultipleLessons(ctx context.Context, db database.Ext, schoolID string) error {
	var lessonID, studentID pgtype.Text
	errMsg := fmt.Sprintf(`failed to get lessons for stress test using course_id %s and center_id %s`, l.ListCfg.CourseID, l.ListCfg.LocationID)
	lessonQuery := `SELECT distinct lesson_id, user_id
					FROM (
						SELECT l.lesson_id, lm.user_id, row_number() OVER (PARTITION BY l.lesson_id ORDER BY l.lesson_id ASC) AS row_number
						FROM lessons l
						INNER JOIN lesson_members lm ON l.lesson_id = lm.lesson_id AND lm.deleted_at is null
						WHERE l.course_id = $1 
						AND l.center_id = $2 
						AND l.resource_path = $3
						AND l.deleted_at is null
					) temp WHERE row_number = 1 `

	rows, err := db.Query(ctx, lessonQuery, l.ListCfg.CourseID, l.ListCfg.LocationID, schoolID)
	if err != nil {
		return fmt.Errorf("%s, db.Query: %w", errMsg, err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&lessonID, &studentID); err != nil {
			return fmt.Errorf("%s, rows.Scan: %w", errMsg, err)
		}

		l.LessonsList = append(l.LessonsList, LessonListItem{
			LessonID:  lessonID.String,
			StudentID: studentID.String,
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%s, rows.Err: %w", errMsg, err)
	}

	return nil
}

func (l *LiveLessonList) GetOneFromLessonList() (string, string, error) {
	listCount := len(l.LessonsList)

	if listCount == 0 {
		return "", "", fmt.Errorf("no lessons are available")
	}
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(int64(listCount)))
	if err != nil {
		return "", "", fmt.Errorf("failed to get random number: %w", err)
	}
	item := l.LessonsList[randomNumber.Int64()]

	return item.LessonID, item.StudentID, nil
}
