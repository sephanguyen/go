package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) retrieveCourseStatistic(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aUserSignedInTeacher(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &pb.RetrieveCourseStatisticRequest{
		CourseId:    stepState.CourseID,
		StudyPlanId: stepState.StudyPlanID,
	}
	if len(stepState.ClassID) != 0 {
		req.ClassId = stepState.ClassID
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseReaderServiceClient(s.Conn).RetrieveCourseStatistic(s.signedCtx(ctx), stepState.Request.(*pb.RetrieveCourseStatisticRequest))

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error call RetrieveCourseStatistic %w", stepState.ResponseErr)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveCourseStatisticV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aUserSignedInTeacher(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &pb.RetrieveCourseStatisticRequestV2{
		CourseId:    stepState.CourseID,
		StudyPlanId: stepState.StudyPlanID,
		ClassId:     []string{},
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseReaderServiceClient(s.Conn).RetrieveCourseStatisticV2(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasCreatedAStudyplanExactMatchWithTheBookContentForAllLoginStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, student := range stepState.Students {
		stepState.StudentID = student.StudentID
		stepState.StudentToken = student.StudentToken
		var err error
		if ctx, err = s.hasCreatedAStudyplanExactMatchWithTheBookContentForStudent(StepStateToContext(ctx, stepState), stepState.StudentID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentDoTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(ctx context.Context, numStudent int, arg1, arg2, arg3, arg4, arg5 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.Students) < numStudent {
		stepState.Students = stepState.Students[0:numStudent]
	}
	for _, student := range stepState.Students {
		stepState.StudentID = student.StudentID
		stepState.StudentToken = student.StudentToken
		if ctx, err := s.doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics(StepStateToContext(ctx, stepState), "", arg1, arg2, arg3, arg4, arg5); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsLogins(ctx context.Context, studentCount int, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentIDs := make([]string, 0, studentCount)
	students := make([]*Student, 0, studentCount)
	for i := 0; i < studentCount; i++ {
		if ctx, err := s.logins(ctx, studentRawText); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentIDs = append(studentIDs, stepState.StudentID)
		students = append(students, &Student{
			StudentID:    stepState.StudentID,
			StudentToken: stepState.StudentToken,
		})
	}

	stepState.StudentIDs = studentIDs
	stepState.Students = students
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsCorrectCourseStatisticItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.RetrieveCourseStatisticRequest)
	statisticResp := stepState.Response.(*pb.RetrieveCourseStatisticResponse)
	classID := pgtype.Text{Status: pgtype.Null}
	if len(req.ClassId) != 0 {
		classID = database.Text(req.ClassId)
	}

	items, err := (&repositories.CourseStudyPlanRepo{}).ListCourseStatisticItems(ctx, s.DB, &repositories.ListCourseStatisticItemsArgs{
		CourseID:    database.Text(req.CourseId),
		StudyPlanID: database.Text(req.StudyPlanId),
		ClassID:     classID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error repo list course statistic items %w", err)
	}

	// order of topic, study plan is correct
	topicIDs := []string{}
	studyPlanItemIDs := []string{}
	for _, item := range items {
		topicIDs = append(topicIDs, item.ContentStructure.TopicID)
		studyPlanItemIDs = append(studyPlanItemIDs, item.RootStudyPlanItemID)
	}
	topicIDs = golibs.GetUniqueElementStringArray(topicIDs)
	studyPlanItemIDs = golibs.GetUniqueElementStringArray(studyPlanItemIDs)

	if len(topicIDs) != len(statisticResp.GetCourseStatisticItems()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v topics got %v", len(topicIDs), len(statisticResp.GetCourseStatisticItems()))
	}

	sort.Strings(stepState.ArchivedStudyPlanItemIDs)

	spiIndex := 0
	for i, statItem := range statisticResp.GetCourseStatisticItems() {
		if statItem.TopicId != topicIDs[i] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong topic order")
		}
		for _, studyPlanItem := range statItem.GetStudyPlanItemStatisticItems() {
			if spiIndex >= len(studyPlanItemIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("more study plan items return than expected")
			}
			if studyPlanItem.GetStudyPlanItemId() != studyPlanItemIDs[spiIndex] {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong study plan item order")
			}
			if sort.SearchStrings(stepState.ArchivedStudyPlanItemIDs, studyPlanItem.GetStudyPlanItemId()) != len(stepState.ArchivedStudyPlanItemIDs) {
				// archived item
				if studyPlanItem.GetTotalAssignedStudent() != 0 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("total assigned student mustn't count archived item")
				}
				if studyPlanItem.GetCompletedStudent() != 0 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("completed student mustn't count archived item")
				}
				if studyPlanItem.GetAverageScore() != 0 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("average score mustn't include archived item")
				}
			}
			spiIndex++
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someOfCreatedStudyPlanItemAreArchived(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedIn(ctx, "teacher")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(stepState.StudyPlanIDs) != 0 {
		archived := stepState.StudyPlanIDs[rand.Intn(len(stepState.StudyPlanIDs))]
		rows, err := s.DB.Query(ctx,
			`
UPDATE study_plan_items
SET status = 'STUDY_PLAN_ITEM_STATUS_ARCHIVED'
FROM study_plans
WHERE study_plan_items.study_plan_id = study_plans.study_plan_id
  AND (study_plans.master_study_plan_id = $1
       OR study_plans.study_plan_id = $1)
  AND study_plan_items.copy_study_plan_item_id IS NOT NULL
RETURNING study_plan_items.copy_study_plan_item_id
`,
			database.Text(archived))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		defer rows.Close()
		for rows.Next() {
			var studyPlanItemID string
			if err := rows.Scan(&studyPlanItemID); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
			stepState.ArchivedStudyPlanItemIDs = append(stepState.ArchivedStudyPlanItemIDs, studyPlanItemID)
		}
		if err := rows.Err(); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) someStudentsAreMembersOfSomeClasses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	classID := idutil.ULIDNow()
	for i, studentID := range stepState.StudentIDs {
		if rand.Int()&1 == 0 {
			classID = idutil.ULIDNow()
		}
		if i > 0 && rand.Int()&1 == 0 {
			break
		}

		err := (&repositories.CourseClassRepo{}).BulkUpsert(ctx, s.DB, []*entities.CourseClass{
			{
				BaseEntity: entities.BaseEntity{
					CreatedAt: database.Timestamptz(time.Now()),
					UpdatedAt: database.Timestamptz(time.Now()),
					DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
				},
				ID:       database.Text(idutil.ULIDNow()),
				CourseID: database.Text(stepState.CourseID),
				ClassID:  database.Text(classID),
			},
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		err = (&repositories.ClassStudentRepo{}).Upsert(ctx, s.DB, &entities.ClassStudent{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(time.Now()),
				UpdatedAt: database.Timestamptz(time.Now()),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			StudentID: database.Text(studentID),
			ClassID:   database.Text(classID),
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.ClassID = classID
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIs(ctx context.Context, topicAssigned, topicCompleted, topicAverageScore int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	statisticResp := stepState.Response.(*pb.RetrieveCourseStatisticResponse)

	if topicAssigned != statisticResp.CourseStatisticItems[0].TotalAssignedStudent {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong total_assigned_student expected = %v got = %v", topicAssigned, statisticResp.CourseStatisticItems[0].TotalAssignedStudent)
	}

	if topicCompleted != statisticResp.CourseStatisticItems[0].CompletedStudent {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_student expected = %v got = %v", topicCompleted, statisticResp.CourseStatisticItems[0].CompletedStudent)
	}

	if topicAverageScore != statisticResp.CourseStatisticItems[0].AverageScore {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong average_score expected = %v got = %v", topicAverageScore, statisticResp.CourseStatisticItems[0].AverageScore)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIsV2(ctx context.Context, topicAssigned, topicCompleted, topicAverageScore int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	statisticResp := stepState.Response.(*pb.RetrieveCourseStatisticResponseV2)

	if topicAssigned != statisticResp.TopicStatistic[0].TotalAssignedStudent {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong total_assigned_student expected = %v got = %v", topicAssigned, statisticResp.TopicStatistic[0].TotalAssignedStudent)
	}

	if topicCompleted != statisticResp.TopicStatistic[0].CompletedStudent {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong completed_student expected = %v got = %v", topicCompleted, statisticResp.TopicStatistic[0].CompletedStudent)
	}

	if topicAverageScore != statisticResp.TopicStatistic[0].AverageScore {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong average_score expected = %v got = %v", topicAverageScore, statisticResp.TopicStatistic[0].AverageScore)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsCorrectCourseStatisticItemsV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.RetrieveCourseStatisticRequestV2)
	statisticResp := stepState.Response.(*pb.RetrieveCourseStatisticResponseV2)

	classID := pgtype.TextArray{Status: pgtype.Null}
	if len(req.ClassId) != 0 {
		classID = database.TextArray(req.ClassId)
	}

	items, err := (&repositories.CourseStudyPlanRepo{}).ListCourseStatisticItemsV2(ctx, s.DB, &repositories.ListCourseStatisticItemsArgsV2{
		CourseID:    database.Text(req.CourseId),
		StudyPlanID: database.Text(req.StudyPlanId),
		ClassID:     classID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error repo list topic statistic items %w", err)
	}

	// order of topic, study plan is correct
	topicIDs := []string{}
	studyPlanItemIDs := []string{}
	for _, item := range items {
		topicIDs = append(topicIDs, item.ContentStructure.TopicID)
		studyPlanItemIDs = append(studyPlanItemIDs, item.RootStudyPlanItemID)
	}
	topicIDs = golibs.GetUniqueElementStringArray(topicIDs)
	studyPlanItemIDs = golibs.GetUniqueElementStringArray(studyPlanItemIDs)

	if len(topicIDs) != len(statisticResp.GetTopicStatistic()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v topics got %v", len(topicIDs), len(statisticResp.GetTopicStatistic()))
	}

	sort.Strings(stepState.ArchivedStudyPlanItemIDs)

	spiIndex := 0
	for i, statItem := range statisticResp.GetTopicStatistic() {
		fmt.Println("====================")
		fmt.Printf("TopicID: %v\n", statItem.GetTopicId())
		fmt.Printf("Total Ass: %v\n", statItem.GetTotalAssignedStudent())
		fmt.Printf("Complete: %v\n", statItem.GetCompletedStudent())
		fmt.Printf("Avg: %v\n", statItem.GetAverageScore())
		fmt.Println("====================")
		if statItem.TopicId != topicIDs[i] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("wrong topic order")
		}

		for _, studyPlanItem := range statItem.GetLearningMaterialStatistic() {
			if spiIndex >= len(studyPlanItemIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("more study plan items return than expected")
			}
			if studyPlanItem.GetStudyPlanItemId() == studyPlanItemIDs[spiIndex] {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong study plan item order")
			}
			if sort.SearchStrings(stepState.ArchivedStudyPlanItemIDs, studyPlanItem.GetStudyPlanItemId()) != len(stepState.ArchivedStudyPlanItemIDs) {
				// archived item
				if studyPlanItem.GetTotalAssignedStudent() != 0 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("total assigned student mustn't count archived item")
				}
				if studyPlanItem.GetCompletedStudent() != 0 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("completed student mustn't count archived item")
				}
				if studyPlanItem.GetAverageScore() != 0 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("average score mustn't include archived item")
				}
			}
			spiIndex++
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
