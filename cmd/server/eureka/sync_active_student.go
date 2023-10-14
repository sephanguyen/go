package eureka

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var zapLogger *zap.Logger

func init() {
	bootstrap.RegisterJob("eureka_sync_active_student", RunSyncActiveStudent)
}

// RunSyncActiveStudent sync active student.
func RunSyncActiveStudent(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()

	eurekaDBConn := rsc.DB().DB.(*pgxpool.Pool)
	fatimaConn := rsc.GRPCDial("fatima")

	zapLogger.Info("Starting sync active student")
	err := syncStudentPackages(ctx, eurekaDBConn, fatimaConn)
	if err != nil {
		zapLogger.Fatal("unable to sync student package", zap.Error(err))
	}
	zapLogger.Info("Complete sync active student")
	return nil
}

func syncStudentPackages(ctx context.Context, eurekaDB *pgxpool.Pool, fatimaConn grpc.ClientConnInterface) error {
	batchSize := 2000
	totalStudentPackage := 0
	totalStudent, err := countTotalStudent(ctx, eurekaDB)
	if err != nil {
		return err
	}
	zapLogger.Info("total student need to progress", zap.Int("total student", totalStudent), zap.Int("batch size", batchSize))
	limit := database.Int8(int64(batchSize))
	offset := pgtype.Text{Status: pgtype.Null}

	expectedCourseStudent := make([]*entities.CourseStudent, 0)
	completedCount := 0
	for {
		zapLogger.Info("current batch", zap.Int("complete", completedCount), zap.Int("total", totalStudent))
		studentIDs, err := getStudentListPaging(ctx, eurekaDB, limit, offset)
		if err != nil {
			return err
		}
		if len(studentIDs) == 0 {
			break
		}

		studentPackages, err := getStudentPackages(ctx, fatimaConn, studentIDs)
		if err != nil {
			return err
		}

		courseStudentEnts := extractCourseStudentFromStudentPackage(studentPackages)

		err = updateCourseStudent(ctx, eurekaDB, courseStudentEnts)
		if err != nil {
			return err
		}

		offset = database.Text(studentIDs[len(studentIDs)-1])
		totalStudentPackage += len(studentPackages)
		completedCount += len(studentIDs)
		expectedCourseStudent = append(expectedCourseStudent, courseStudentEnts...)
	}

	currentCourseStudent, err := getCurrentCourseStudentHaveStartTimeAndEndTime(ctx, eurekaDB)
	if err != nil {
		return err
	}

	err = compare(ctx, eurekaDB, expectedCourseStudent, currentCourseStudent)
	if err != nil {
		return err
	}

	zapLogger.Info("progressed total number of student package", zap.Int("total student package", totalStudentPackage))
	return nil
}

// some student packages in some way is not progressed, it leads to missing data in student course
// we will return these noise data, and ignore theme
func removeWrongRecordOfStudentPackages(ctx context.Context, eurekaDB *pgxpool.Pool, expectedStudentIDs, currentStudentIDs []string) ([]string, error) {
	st := make(map[string]struct{})

	wrongStudentIDs := make([]string, 0)

	for _, id := range currentStudentIDs {
		st[id] = struct{}{}
	}

	// these expectedStudentIDs is extracted from student packages
	for _, id := range expectedStudentIDs {
		if _, ok := st[id]; !ok {
			wrongStudentIDs = append(wrongStudentIDs, id)
		}
	}

	query := `SELECT count(*) FROM course_students WHERE student_id = ANY($1) AND deleted_at is NULL`
	count := database.Int8(0)
	err := eurekaDB.QueryRow(ctx, query, database.TextArray(wrongStudentIDs)).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count.Int > 0 {
		return nil, fmt.Errorf("expect these student ids %v not exists in table course_students, count %d", wrongStudentIDs, count.Int)
	}

	return wrongStudentIDs, nil
}

func removeWrongRecordOfCurrentCourseStudent(expectedStudentIDs, currentStudentIDs []string) []string {
	st := make(map[string]struct{})

	wrongStudentIDs := make([]string, 0)

	for _, id := range expectedStudentIDs {
		st[id] = struct{}{}
	}

	for _, id := range currentStudentIDs {
		if _, ok := st[id]; !ok {
			wrongStudentIDs = append(wrongStudentIDs, id)
		}
	}

	return wrongStudentIDs
}

func compare(ctx context.Context, eurekaDB *pgxpool.Pool, expectedCourseStudent, currentCourseStudent []*entities.CourseStudent) error {
	expectedStudentMap := make(map[string][]string)
	expectedStudentIDs := make([]string, 0)
	currentStudentMap := make(map[string][]string)
	currentStudentIDs := make([]string, 0)
	for _, e := range expectedCourseStudent {
		expectedStudentMap[e.StudentID.String] = append(expectedStudentMap[e.StudentID.String], e.CourseID.String)
	}

	for _, e := range currentCourseStudent {
		currentStudentMap[e.StudentID.String] = append(currentStudentMap[e.StudentID.String], e.CourseID.String)
	}

	for id := range expectedStudentMap {
		expectedStudentIDs = append(expectedStudentIDs, id)
	}
	for id := range currentStudentMap {
		currentStudentIDs = append(currentStudentIDs, id)
	}

	wrongExpectedStudentIDs, err := removeWrongRecordOfStudentPackages(ctx, eurekaDB, expectedStudentIDs, currentStudentIDs)
	if err != nil {
		return err
	}

	wrongCurrentStudentIDs := removeWrongRecordOfCurrentCourseStudent(expectedStudentIDs, currentStudentIDs)

	// ignore wrong data
	if len(expectedStudentMap)-len(wrongExpectedStudentIDs) != len(currentStudentMap)-len(wrongCurrentStudentIDs) {
		return fmt.Errorf("expect number of student is %d but got %d", len(expectedStudentMap)-len(wrongExpectedStudentIDs), len(currentStudentMap)-len(wrongCurrentStudentIDs))
	}

	for studentID := range expectedStudentMap {
		// ignore wrong data
		if golibs.InArrayString(studentID, wrongExpectedStudentIDs) || golibs.InArrayString(studentID, wrongCurrentStudentIDs) {
			continue
		}
		expectedCourseID := expectedStudentMap[studentID]
		currentCourseID, ok := currentStudentMap[studentID]
		if !ok {
			return fmt.Errorf("cannot find student id %s", studentID)
		}
		expectedCourseID = golibs.Uniq(expectedCourseID)
		currentCourseID = golibs.Uniq(currentCourseID)
		if len(expectedCourseID) != len(currentCourseID) {
			return fmt.Errorf("expect number of course_student of student %s is %d but got %d", studentID, len(expectedCourseID), len(currentCourseID))
		}

		sort.Slice(expectedCourseID, func(i, j int) bool {
			return expectedCourseID[i] < expectedCourseID[j]
		})

		sort.Slice(currentCourseID, func(i, j int) bool {
			return currentCourseID[i] < currentCourseID[j]
		})
		for i := 0; i < len(expectedCourseID); i++ {
			if expectedCourseID[i] != currentCourseID[i] {
				return fmt.Errorf("expect student id %s have course ids %v but got %v", studentID, expectedCourseID, currentCourseID)
			}
		}
	}
	return nil
}

func countTotalStudent(ctx context.Context, eurekaDB *pgxpool.Pool) (int, error) {
	query := `SELECT count(*) FROM (SELECT student_id FROM course_students WHERE deleted_at IS NULL GROUP BY student_id) AS students`
	var totalStudent pgtype.Int8
	err := eurekaDB.QueryRow(ctx, query).Scan(&totalStudent)
	if err != nil {
		return 0, err
	}
	return int(totalStudent.Int), nil
}

func getStudentListPaging(ctx context.Context, eurekaDB *pgxpool.Pool, limit pgtype.Int8, offset pgtype.Text) ([]string, error) {
	query := `
		SELECT student_id 
		FROM course_students 
		WHERE ($1::TEXT IS NULL OR student_id > $1) AND deleted_at IS NULL
		GROUP BY student_id
		ORDER BY student_id	
		LIMIT $2 
	`
	studentIDs := make([]string, 0)
	rows, err := eurekaDB.Query(ctx, query, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		id := pgtype.Text{Status: pgtype.Null}
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, id.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return studentIDs, nil
}

func getStudentPackages(ctx context.Context, fatimaConn grpc.ClientConnInterface, studentIDs []string) ([]*fpb.StudentPackage, error) {
	studentPackages := make([]*fpb.StudentPackage, 0)
	stream, err := fpb.NewSubscriptionModifierServiceClient(fatimaConn).ListStudentPackageV2(ctx, &fpb.ListStudentPackageV2Request{
		StudentIds: studentIDs,
	})
	if err != nil {
		return nil, err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if resp.StudentPackage != nil {
			studentPackages = append(studentPackages, resp.StudentPackage)
		}
	}

	return studentPackages, nil
}

func getCourseIDsFromPackage(p *fpb.StudentPackage) []string {
	courseIDs := make([]string, 0)
	courseIDs = append(courseIDs, p.Properties.CanDoQuiz...)
	courseIDs = append(courseIDs, p.Properties.CanViewStudyGuide...)
	courseIDs = append(courseIDs, p.Properties.CanWatchVideo...)
	courseIDs = golibs.Uniq(courseIDs)
	return courseIDs
}

func extractCourseStudentFromStudentPackage(studentPackages []*fpb.StudentPackage) []*entities.CourseStudent {
	courseStudents := make([]*entities.CourseStudent, 0)
	for _, p := range studentPackages {
		if !p.IsActive {
			inActiveCourseIDs := getCourseIDsFromPackage(p)
			// remove inactive course
			for _, cid := range inActiveCourseIDs {
				for i := 0; i < len(courseStudents); i++ {
					if p.StudentId == courseStudents[i].StudentID.String && cid == courseStudents[i].CourseID.String {
						courseStudents = append(courseStudents[:i], courseStudents[i+1:]...)
					}
				}
			}
			continue
		}
		activeCourse := getCourseIDsFromPackage(p)

		for _, courseID := range activeCourse {
			courseStudents = append(courseStudents, &entities.CourseStudent{
				CourseID:  database.Text(courseID),
				StudentID: database.Text(p.StudentId),
				StartAt:   database.TimestamptzFromPb(p.StartAt),
				EndAt:     database.TimestamptzFromPb(p.EndAt),
			})
		}
	}
	return courseStudents
}

const updateCourseStudentStmt = `
UPDATE course_students SET 
	start_at = $3,
	end_at = $4,
	updated_at = now()
WHERE course_id = $1 AND student_id = $2 AND start_at IS NULL AND end_at IS NULL AND deleted_at IS NULL`

func queueUpdateCourseStudent(b *pgx.Batch, item *entities.CourseStudent) {
	query := updateCourseStudentStmt
	scanFields := database.GetScanFields(item, []string{"course_id", "student_id", "start_at", "end_at"})
	b.Queue(query, scanFields...)
}

func updateCourseStudent(ctx context.Context, eurekaDB *pgxpool.Pool, courseStudents []*entities.CourseStudent) error {
	b := &pgx.Batch{}

	for _, item := range courseStudents {
		queueUpdateCourseStudent(b, item)
	}

	result := eurekaDB.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func getCurrentCourseStudentHaveStartTimeAndEndTime(ctx context.Context, eurekaDB *pgxpool.Pool) ([]*entities.CourseStudent, error) {
	e := &entities.CourseStudent{}
	query := fmt.Sprintf(`SELECT %s FROM course_students WHERE start_at IS NOT NULL AND end_at IS NOT NULL AND deleted_at IS NULL`, strings.Join(database.GetFieldNames(e), ","))

	rows, err := eurekaDB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*entities.CourseStudent, 0)

	for rows.Next() {
		e := &entities.CourseStudent{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
