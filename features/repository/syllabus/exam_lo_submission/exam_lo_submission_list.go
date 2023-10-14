package exam_lo_submission

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
)

func (s *Suite) userCreateASetOfExamLOSubmissionsToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.ExtendedExamLOSubmissionMap = make(map[string]*repositories.ExtendedExamLOSubmission)
	insertUser := func(dbTrace *database.DBTrace, userID string) error {
		num := rand.Int()
		var now pgtype.Timestamptz
		now.Set(time.Now())
		u := bob_entities.User{}
		database.AllNullEntity(&u)
		u.ID = database.Text(userID)
		u.LastName.Set(fmt.Sprintf("valid-user-%d", num))
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num))
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num))
		u.Country.Set(cpb.Country_COUNTRY_VN.String())
		u.Group.Set(constants.RoleStudent)
		u.DeviceToken.Set(nil)
		u.AllowNotification.Set(true)
		u.CreatedAt = now
		u.UpdatedAt = now
		u.IsTester.Set(nil)
		u.FacebookID.Set(nil)
		u.ResourcePath.Set(fmt.Sprint(constant.ManabieSchool))

		_, err := database.InsertOnConflictDoNothing(ctx, &u, dbTrace.Exec)
		if err != nil {
			return fmt.Errorf("insert student error: %v", err)
		}
		return nil
	}
	insertMasterStudyPlan := func(studyPlanID, learningMaterialID string, now, startDate, endDate time.Time) (context.Context, error) {
		masterStudyPlan := &entities.MasterStudyPlan{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			StudyPlanID:        database.Text(studyPlanID),
			SchoolDate:         database.Timestamptz(now),
			LearningMaterialID: database.Text(learningMaterialID),
			Status:             database.Text(epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE.String()),
			StartDate:          database.Timestamptz(startDate),
			EndDate:            database.Timestamptz(endDate),
			AvailableFrom:      database.Timestamptz(startDate),
			AvailableTo:        database.Timestamptz(endDate),
		}
		if _, err := database.Insert(ctx, masterStudyPlan, s.DB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert master_study_plan: %w", err)
		}

		return utils.StepStateToContext(ctx, stepState), nil
	}

	insertIndividualStudyPlan := func(studyPlanID, studentID, learningMaterialID string, now, startDate, endDate time.Time) (context.Context, error) {
		individualStudyPlan := &entities.IndividualStudyPlan{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			ID:                 database.Text(studyPlanID),
			StudentID:          database.Text(studentID),
			SchoolDate:         database.Timestamptz(now),
			LearningMaterialID: database.Text(learningMaterialID),
			Status:             database.Text(epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE.String()),
			StartDate:          database.Timestamptz(startDate),
			EndDate:            database.Timestamptz(endDate),
			AvailableFrom:      database.Timestamptz(startDate),
			AvailableTo:        database.Timestamptz(endDate),
		}
		if _, err := database.Insert(ctx, individualStudyPlan, s.DB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert individual_study_plan: %w", err)
		}

		return utils.StepStateToContext(ctx, stepState), nil
	}

	now := time.Now()
	studentID := idutil.ULIDNow()
	startDateArg := now.Add(-100 * time.Hour)
	endDateArg := now.Add(100 * time.Hour)
	insertUser(s.DBTrace, studentID)

	for i := 0; i < 20; i++ {
		courseID := idutil.ULIDNow()
		submissionID := idutil.ULIDNow()
		studyPlanID := idutil.ULIDNow()
		bookID := idutil.ULIDNow()
		topicID := idutil.ULIDNow()
		learningMaterialID := idutil.ULIDNow()
		shuffledQuizSetID := idutil.ULIDNow()
		stepState.SubmissionIDs = append(stepState.SubmissionIDs, submissionID)
		topic := &entities.Topic{
			ID:                    database.Text(topicID),
			Name:                  database.Text("topic-name"),
			Country:               database.Text(cpb.Country_COUNTRY_JP.String()),
			Grade:                 database.Int2(10),
			Subject:               database.Text(cpb.Subject_SUBJECT_BIOLOGY.String()),
			TopicType:             database.Text(cpb.TopicStatus_TOPIC_STATUS_NONE.String()),
			Status:                database.Text(cpb.TopicStatus_TOPIC_STATUS_NONE.String()),
			ChapterID:             database.Text(idutil.ULIDNow()),
			DisplayOrder:          database.Int2(1),
			IconURL:               database.Text("icon-url"),
			SchoolID:              database.Int4(constant.ManabieSchool),
			TotalLOs:              database.Int4(1),
			CreatedAt:             database.Timestamptz(now),
			UpdatedAt:             database.Timestamptz(now),
			PublishedAt:           pgtype.Timestamptz{Status: pgtype.Null},
			AttachmentNames:       pgtype.TextArray{Status: pgtype.Null},
			AttachmentURLs:        pgtype.TextArray{Status: pgtype.Null},
			Instruction:           database.Text("instruction"),
			CopiedTopicID:         pgtype.Text{Status: pgtype.Null},
			EssayRequired:         database.Bool(true),
			DeletedAt:             pgtype.Timestamptz{Status: pgtype.Null},
			LODisplayOrderCounter: database.Int4(0),
		}
		if _, err := database.Insert(ctx, topic, s.DBTrace.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert topic: %w", err)
		}

		examLO := &entities.ExamLO{
			LearningMaterial: entities.LearningMaterial{
				BaseEntity: entities.BaseEntity{
					CreatedAt: database.Timestamptz(now),
					UpdatedAt: database.Timestamptz(now),
					DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
				},
				ID:           database.Text(learningMaterialID),
				TopicID:      database.Text(topicID),
				Name:         database.Text("name"),
				Type:         database.Text(sspb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO.String()),
				DisplayOrder: database.Int2(0),
			},
			Instruction:    database.Text("instruction"),
			GradeToPass:    database.Int4(0),
			ManualGrading:  database.Bool(true),
			TimeLimit:      database.Int4(0),
			MaximumAttempt: database.Int4(10),
			ApproveGrading: database.Bool(false),
			GradeCapping:   database.Bool(true),
			ReviewOption:   database.Text(sspb.ExamLOReviewOption_EXAM_LO_REVIEW_OPTION_IMMEDIATELY.String()),
		}
		if _, err := database.Insert(ctx, examLO, s.DBTrace.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert exam_lo: %w", err)
		}

		examLOSubmission := &entities.ExamLOSubmission{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			SubmissionID:       database.Text(submissionID),
			StudentID:          database.Text(studentID),
			StudyPlanID:        database.Text(studyPlanID),
			LearningMaterialID: database.Text(learningMaterialID),
			ShuffledQuizSetID:  database.Text(shuffledQuizSetID),
			Status:             database.Text(sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String()),
			Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String()),
			TeacherFeedback:    database.Text("teacher-feedback"),
			TeacherID:          database.Text(idutil.ULIDNow()),
			MarkedAt:           database.Timestamptz(now),
			RemovedAt:          database.Timestamptz(now),
			TotalPoint:         database.Int4(0),
			LastAction:         database.Text(sspb.ApproveGradingAction_APPROVE_ACTION_NONE.String()),
			LastActionAt:       database.Timestamptz(now),
			LastActionBy:       database.Text("someone"),
		}
		if _, err := database.Insert(ctx, examLOSubmission, s.DB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert exam_lo_submission: %w", err)
		}

		studyPlan := &entities.StudyPlan{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			ID:                  database.Text(studyPlanID),
			MasterStudyPlan:     pgtype.Text{Status: pgtype.Null},
			Name:                database.Text("study-plan-name"),
			StudyPlanType:       database.Text(epb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String()),
			Status:              database.Text(epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE.String()),
			SchoolID:            database.Int4(constant.ManabieSchool),
			CourseID:            database.Text(courseID),
			BookID:              database.Text(bookID),
			TrackSchoolProgress: database.Bool(true),
			Grades:              database.Int4Array([]int32{1, 2}),
		}
		if _, err := database.Insert(ctx, studyPlan, s.DB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert study_plan: %w", err)
		}

		courseStudentID := idutil.ULIDNow()
		courseStudent := &entities.CourseStudent{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			ID:        database.Text(courseStudentID),
			CourseID:  database.Text(courseID),
			StudentID: database.Text(studentID),
			StartAt:   database.Timestamptz(now.Add(-720 * time.Hour)),
			EndAt:     database.Timestamptz(now.Add(720 * time.Hour)),
		}
		if _, err := database.Insert(ctx, courseStudent, s.DB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert course_student: %w", err)
		}

		courseStudentsAccessPath := &entities.CourseStudentsAccessPath{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			CourseID:        database.Text(courseID),
			StudentID:       database.Text(studentID),
			CourseStudentID: database.Text(courseStudentID),
			LocationID:      database.Text(idutil.ULIDNow()),
			AccessPath:      database.Text(idutil.ULIDNow()),
		}
		if _, err := database.Insert(ctx, courseStudentsAccessPath, s.DB.Exec); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to insert course_students_access_path: %w", err)
		}

		num := rand.Int31() % 3
		duration := rand.Int31n(200)
		switch num {
		case 0: // only master
			if ctx, err := insertMasterStudyPlan(studyPlanID, learningMaterialID, now, now.Add(-time.Duration(duration)*time.Hour), now.Add(time.Duration(duration)*time.Hour)); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
		case 1: // only individual
			if ctx, err := insertIndividualStudyPlan(studyPlanID, studentID, learningMaterialID, now, now.Add(-time.Duration(duration)*time.Hour), now.Add(time.Duration(duration)*time.Hour)); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
		default: // both
			if ctx, err := insertMasterStudyPlan(studyPlanID, learningMaterialID, now, now.Add(-time.Duration(2*duration)*time.Hour), now.Add(time.Duration(2*duration)*time.Hour)); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
			if ctx, err := insertIndividualStudyPlan(studyPlanID, studentID, learningMaterialID, now, now.Add(-time.Duration(duration)*time.Hour), now.Add(time.Duration(duration)*time.Hour)); err != nil {
				return utils.StepStateToContext(ctx, stepState), err
			}
		}

		if (startDateArg.Equal(now.Add(-time.Duration(duration)*time.Hour)) || startDateArg.Before(now.Add(-time.Duration(duration)*time.Hour))) &&
			(endDateArg.Equal(now.Add(time.Duration(duration)*time.Hour)) || endDateArg.After(now.Add(time.Duration(duration)*time.Hour))) {
			stepState.ExtendedExamLOSubmissionMap[submissionID] = &repositories.ExtendedExamLOSubmission{
				ExamLOSubmission: entities.ExamLOSubmission{
					BaseEntity: entities.BaseEntity{
						CreatedAt: database.Timestamptz(now),
						UpdatedAt: database.Timestamptz(now),
						DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
					},
					SubmissionID:       database.Text(submissionID),
					StudentID:          database.Text(studentID),
					StudyPlanID:        database.Text(studyPlanID),
					LearningMaterialID: database.Text(learningMaterialID),
					ShuffledQuizSetID:  database.Text(shuffledQuizSetID),
					Status:             database.Text(sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS.String()),
					Result:             database.Text(sspb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED.String()),
					TeacherFeedback:    database.Text("teacher-feedback"),
					TeacherID:          database.Text(idutil.ULIDNow()),
					MarkedAt:           database.Timestamptz(now),
					RemovedAt:          database.Timestamptz(now),
					TotalPoint:         database.Int4(0),
				},
				CourseID:  database.Text(courseID),
				StartDate: database.Timestamptz(now.Add(-time.Duration(duration) * time.Hour)),
				EndDate:   database.Timestamptz(now.Add(time.Duration(duration) * time.Hour)),
			}
		}
	}

	stepState.Request = &repositories.ExamLOSubmissionFilter{
		StudentIDs:         database.TextArray([]string{studentID}),
		StartDate:          database.Timestamptz(startDateArg),
		EndDate:            database.Timestamptz(endDateArg),
		Limit:              uint(30),
		OffsetID:           pgtype.Text{Status: pgtype.Null},
		Statuses:           pgtype.TextArray{Status: pgtype.Null},
		CreatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
		CourseID:           pgtype.Text{Status: pgtype.Null},
		ClassIDs:           pgtype.TextArray{Status: pgtype.Null},
		LocationIDs:        pgtype.TextArray{Status: pgtype.Null},
		StudentName:        pgtype.Text{Status: pgtype.Null},
		ExamName:           pgtype.Text{Status: pgtype.Null},
		CorrectorID:        pgtype.Text{Status: pgtype.Null},
		SubmittedStartDate: database.Timestamptz(startDateArg),
		SubmittedEndDate:   database.Timestamptz(endDateArg),
		UpdatedStartDate:   database.Timestamptz(startDateArg),
		UpdatedEndDate:     database.Timestamptz(endDateArg),
		SubmissionID:       database.Text(stepState.SubmissionIDs[0]),
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallFunctionExamLOSubmissionList(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	examLOSubmissionRepo := &repositories.ExamLOSubmissionRepo{}

	stepState.Response, stepState.ResponseError = examLOSubmissionRepo.List(ctx, s.DBTrace, stepState.Request.(*repositories.ExamLOSubmissionFilter))
	if stepState.ResponseError != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("examLOSubmissionRepo.List: %w", stepState.ResponseError)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) systemReturnsCorrectListExamLOSubmissions(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp := stepState.Response.([]*repositories.ExtendedExamLOSubmission)
	for _, examLOSubmissionWithDates := range resp {
		examLOSubmissionWithDatesEnt := stepState.ExtendedExamLOSubmissionMap[examLOSubmissionWithDates.SubmissionID.String]
		if examLOSubmissionWithDates.StartDate.Time.Unix() != examLOSubmissionWithDatesEnt.StartDate.Time.Unix() {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected start_date %v, got %v", examLOSubmissionWithDatesEnt.StartDate.Time.String(), examLOSubmissionWithDates.StartDate.Time.String())
		}

		if examLOSubmissionWithDates.EndDate.Time.Unix() != examLOSubmissionWithDatesEnt.EndDate.Time.Unix() {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected end_date %v, got %v", examLOSubmissionWithDatesEnt.EndDate.Time.String(), examLOSubmissionWithDates.EndDate.Time.String())
		}
		if examLOSubmissionWithDates.CourseID.String != examLOSubmissionWithDatesEnt.CourseID.String {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected course_id %v, got %v", examLOSubmissionWithDates.CourseID.String, examLOSubmissionWithDatesEnt.CourseID.String)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
