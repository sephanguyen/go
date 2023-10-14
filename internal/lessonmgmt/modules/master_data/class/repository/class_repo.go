package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type ClassRepository struct{}

func (c *ClassRepository) GetByStudentCourse(ctx context.Context, db database.Ext, studentWithCourse []string) (map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepository.GetByStudentCourse")
	defer span.End()

	if len(studentWithCourse)%2 != 0 {
		return nil, fmt.Errorf("invalid student course input value,got %s", studentWithCourse)
	}
	studentCourse := make([]string, 0, len(studentWithCourse)/2) // will like ["($1, $2)", "($3, $4)", ...]
	args := make([]interface{}, 0, len(studentCourse))
	for i := 0; i < len(studentWithCourse); i += 2 {
		studentID := studentWithCourse[i]
		courseID := studentWithCourse[i+1]
		args = append(args, studentID, courseID)
		studentCourse = append(studentCourse, fmt.Sprintf("($%d, $%d)", i+1, i+2))
	}
	query := `select cm.user_id ,c.course_id,c.class_id
	from class_member cm join class c on c.class_id = cm.class_id where cm.deleted_at is null
	and (cm.user_id ,c.course_id) IN (:PlaceHolderVar)`
	placeHolderVar := strings.Join(studentCourse, ", ")
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	studentCourseWithClassMap := map[string]string{}
	for rows.Next() {
		var userID, courseID, ClassID pgtype.Text
		err := rows.Scan(&userID, &courseID, &ClassID)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		studentCourseWithClassMap[userID.String+"-"+courseID.String] = ClassID.String
	}
	return studentCourseWithClassMap, nil
}

func (c *ClassRepository) GetReserveClass(ctx context.Context, db database.Ext, studentWithCourse []string) (map[string]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepository.GetReserveClass")
	defer span.End()

	if len(studentWithCourse)%2 != 0 {
		return nil, fmt.Errorf("invalid student course input value,got %s", studentWithCourse)
	}
	studentCourse := make([]string, 0, len(studentWithCourse)/2) // will like ["($1, $2)", "($3, $4)", ...]
	args := make([]interface{}, 0, len(studentCourse))
	for i := 0; i < len(studentWithCourse); i += 2 {
		studentID := studentWithCourse[i]
		courseID := studentWithCourse[i+1]
		args = append(args, studentID, courseID)
		studentCourse = append(studentCourse, fmt.Sprintf("($%d, $%d)", i+1, i+2))
	}
	query := `select student_id,course_id,class_id from reserve_class where (student_id ,course_id) IN (:PlaceHolderVar) and deleted_at is null`
	placeHolderVar := strings.Join(studentCourse, ", ")
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	studentCourseWithClassMap := map[string]string{}
	for rows.Next() {
		var studentID, courseID, ClassID pgtype.Text
		if err = rows.Scan(&studentID, &courseID, &ClassID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		studentCourseWithClassMap[studentID.String+"-"+courseID.String] = ClassID.String
	}
	return studentCourseWithClassMap, nil
}
