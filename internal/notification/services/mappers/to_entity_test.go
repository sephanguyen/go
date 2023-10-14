package mappers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	natspb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_PbToInfoNotificationMsgEnt(t *testing.T) {
	t.Parallel()
	var msg = &cpb.NotificationMessage{}
	assert.NoError(t, faker.FakeData(msg))
	msgEnt, err := PbToInfoNotificationMsgEnt(msg)
	assert.NoError(t, err)
	assert.Equal(t, msg.NotificationMsgId, msgEnt.NotificationMsgID.String)
	assert.Equal(t, msg.Title, msgEnt.Title.String)
	json.Marshal(msg.Content)
	contentJson := fmt.Sprintf(`{"raw":"%s","rendered_url":"%s"}`, msg.Content.Raw, msg.Content.Rendered)
	assert.NoError(t, err)
	assert.JSONEq(t, string(contentJson), string(msgEnt.Content.Bytes))
	assert.Equal(t, msg.MediaIds, database.FromTextArray(msgEnt.MediaIDs))
}

func Test_PbToInfoNotificationTarget(t *testing.T) {
	t.Parallel()

	checkTargetGroup := func(targetGroupPb *cpb.NotificationTargetGroup, targetGroupEnt *entities.InfoNotificationTarget) {
		stringUserGroups := []string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}

		assert.Equal(t, targetGroupPb.CourseFilter.Type.String(), targetGroupEnt.CourseFilter.Type)
		assert.Equal(t, targetGroupPb.GradeFilter.Type.String(), targetGroupEnt.GradeFilter.Type)
		assert.Equal(t, targetGroupPb.LocationFilter.Type.String(), targetGroupEnt.LocationFilter.Type)
		assert.Equal(t, targetGroupPb.ClassFilter.Type.String(), targetGroupEnt.ClassFilter.Type)
		assert.Equal(t, targetGroupPb.SchoolFilter.Type.String(), targetGroupEnt.SchoolFilter.Type)

		assert.Equal(t, len(targetGroupPb.CourseFilter.CourseIds), len(targetGroupEnt.CourseFilter.CourseIDs))
		assert.Equal(t, len(targetGroupPb.GradeFilter.GradeIds), len(targetGroupEnt.GradeFilter.GradeIDs))
		assert.Equal(t, len(targetGroupPb.LocationFilter.LocationIds), len(targetGroupEnt.LocationFilter.LocationIDs))
		assert.Equal(t, len(targetGroupPb.ClassFilter.ClassIds), len(targetGroupEnt.ClassFilter.ClassIDs))
		assert.Equal(t, len(targetGroupPb.SchoolFilter.SchoolIds), len(targetGroupEnt.SchoolFilter.SchoolIDs))

		if targetGroupPb.CourseFilter.CourseIds != nil {
			assert.Equal(t, targetGroupPb.CourseFilter.CourseIds, targetGroupEnt.CourseFilter.CourseIDs)
		}
		if targetGroupPb.GradeFilter.GradeIds != nil {
			assert.Equal(t, targetGroupPb.GradeFilter.GradeIds, targetGroupEnt.GradeFilter.GradeIDs)
		}
		if targetGroupPb.LocationFilter.LocationIds != nil {
			assert.Equal(t, targetGroupPb.LocationFilter.LocationIds, targetGroupEnt.LocationFilter.LocationIDs)
		}
		if targetGroupPb.ClassFilter.ClassIds != nil {
			assert.Equal(t, targetGroupPb.ClassFilter.ClassIds, targetGroupEnt.ClassFilter.ClassIDs)
		}
		if targetGroupPb.SchoolFilter.SchoolIds != nil {
			assert.Equal(t, targetGroupPb.SchoolFilter.SchoolIds, targetGroupEnt.SchoolFilter.SchoolIDs)
		}
		assert.Equal(t, stringUserGroups, targetGroupEnt.UserGroupFilter.UserGroups)
	}

	var target = &cpb.NotificationTargetGroup{}

	nilTypes := []cpb.NotificationTargetGroupSelect{
		cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL,
	}
	fullUserGroups := []cpb.UserGroup{cpb.UserGroup_USER_GROUP_STUDENT, cpb.UserGroup_USER_GROUP_PARENT}
	for _, nilType := range nilTypes {
		t.Run(fmt.Sprintf("convert nil-value filter type: %s", nilType.String()), func(t *testing.T) {
			target.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
				Type: nilType,
			}
			target.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
				Type: nilType,
			}
			target.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
				Type: nilType,
			}
			target.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
				Type: nilType,
			}
			target.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
				Type: nilType,
			}
			target.UserGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{UserGroups: fullUserGroups}
			targetEnt := PbToNotificationTargetEnt(target)
			checkTargetGroup(target, targetEnt)
		})
	}

	for _, nilType := range nilTypes {
		for caseIndex := 0; caseIndex < 4; caseIndex++ {
			t.Run(fmt.Sprintf("convert nil-value filter type and list filter type: %s", nilType.String()), func(t *testing.T) {
				target.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
					Type: nilType,
				}
				target.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
					Type: nilType,
				}
				target.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
					Type: nilType,
				}
				target.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
					Type: nilType,
				}
				target.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
					Type: nilType,
				}

				switch caseIndex {
				case 0:
					target.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
						Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
						CourseIds: []string{"course_1", "course_2"},
					}
				case 1:
					target.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
						Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
						GradeIds: []string{"grade_1", "grade_2"},
					}
				case 2:
					target.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
						Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
						LocationIds: []string{"location_1", "location_2"},
					}
				case 3:
					target.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
						Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
						ClassIds: []string{"class_1", "class_2"},
					}
				case 4:
					target.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
						Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
						SchoolIds: []string{"school_1", "school_2"},
					}
				}

				target.UserGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{UserGroups: fullUserGroups}
				targetEnt := PbToNotificationTargetEnt(target)
				checkTargetGroup(target, targetEnt)
			})
		}
	}

	t.Run("convert specific ids", func(t *testing.T) {
		target.CourseFilter = &cpb.NotificationTargetGroup_CourseFilter{
			Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			CourseIds: []string{"course_1", "course_2"},
		}
		target.GradeFilter = &cpb.NotificationTargetGroup_GradeFilter{
			Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			GradeIds: []string{"grade_1", "grade_2"},
		}
		target.LocationFilter = &cpb.NotificationTargetGroup_LocationFilter{
			Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			LocationIds: []string{"location_1", "location_2"},
		}
		target.ClassFilter = &cpb.NotificationTargetGroup_ClassFilter{
			Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			ClassIds: []string{"class_1", "class_2"},
		}
		target.SchoolFilter = &cpb.NotificationTargetGroup_SchoolFilter{
			Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
			SchoolIds: []string{"school_1", "school_2"},
		}
		target.UserGroupFilter = &cpb.NotificationTargetGroup_UserGroupFilter{UserGroups: fullUserGroups}

		targetEnt := PbToNotificationTargetEnt(target)
		assert.Equal(t, target.CourseFilter.Type.String(), targetEnt.CourseFilter.Type)
		assert.Equal(t, target.GradeFilter.Type.String(), targetEnt.GradeFilter.Type)
		assert.Equal(t, target.CourseFilter.CourseIds, targetEnt.CourseFilter.CourseIDs)
		assert.Equal(t, target.GradeFilter.GradeIds, targetEnt.GradeFilter.GradeIDs)
		assert.Equal(t, target.LocationFilter.LocationIds, targetEnt.LocationFilter.LocationIDs)
		assert.Equal(t, target.ClassFilter.ClassIds, targetEnt.ClassFilter.ClassIDs)
		assert.Equal(t, target.SchoolFilter.SchoolIds, targetEnt.SchoolFilter.SchoolIDs)
	})
}

func Test_PbToInfoNotificationEnt(t *testing.T) {
	t.Parallel()

	var noti = &cpb.Notification{}
	assert.NoError(t, faker.FakeData(noti))
	// can't fake enum type for now
	noti.Type = cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED
	noti.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD

	ent, err := PbToInfoNotificationEnt(noti)
	assert.NoError(t, err)
	assert.Equal(t, noti.NotificationId, ent.NotificationID.String)
	assert.Equal(t, noti.Type.String(), ent.Type.String)
	assert.Equal(t, noti.Data, string(ent.Data.Bytes))
	assert.Equal(t, noti.EditorId, ent.EditorID.String)
	assert.Equal(t, noti.CreatedUserId, ent.CreatedUserID.String)
	assert.Equal(t, noti.Status.String(), ent.Status.String)
	assert.Equal(t, noti.SchoolId, ent.Owner.Int)
	assert.Equal(t, noti.ScheduledAt.AsTime(), ent.ScheduledAt.Time)
	assert.Equal(t, noti.IsImportant, ent.IsImportant.Bool)
	entGenericReceiverIDs := []string{}
	_ = ent.GenericReceiverIDs.AssignTo(&entGenericReceiverIDs)
	assert.Equal(t, noti.GenericReceiverIds, entGenericReceiverIDs)
	entExcludedGenericReceiverIDs := []string{}
	_ = ent.ExcludedGenericReceiverIDs.AssignTo(&entExcludedGenericReceiverIDs)
	assert.Equal(t, noti.ExcludedGenericReceiverIds, entExcludedGenericReceiverIDs)
}

func Test_ToParentNotificationEnts(t *testing.T) {
	studentParents := []*bobEntities.StudentParent{}

	for i := 0; i < 10; i++ {
		var par bobEntities.StudentParent
		database.AllRandomEntity(&par)
		studentParents = append(studentParents, &par)
	}
	notiID := idutil.ULIDNow()
	userNotis, err := ToParentNotificationEnts(studentParents, notiID)
	assert.NoError(t, err)
	for idx, userNoti := range userNotis {
		studentParent := studentParents[idx]
		assert.Equal(t, userNoti.UserGroup.String, cpb.UserGroup_USER_GROUP_PARENT.String())
		assert.Equal(t, userNoti.StudentID.String, studentParent.StudentID.String)
		assert.Equal(t, userNoti.UserID.String, studentParent.ParentID.String)
		assert.Equal(t, userNoti.NotificationID.String, notiID)
		assert.Equal(t, userNoti.QuestionnaireStatus.String, cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED.String())
	}
}
func Test_ToStudentNotificationEnts(t *testing.T) {
	students := []*entities.Audience{}
	studentGradeMap := map[string]int{}
	studentCoursesMap := map[string][]string{}
	studentIDs := []string{}

	for i := 0; i < 10; i++ {
		var stu entities.Audience
		studentIDs = append(studentIDs, stu.StudentID.String)
		students = append(students, &stu)
	}
	seedRandomMap(studentIDs, studentGradeMap)
	seedRandomMap(studentIDs, studentCoursesMap)
	notiID := idutil.ULIDNow()
	userNotis, err := ToStudentNotificationEnts(students, notiID, studentGradeMap, studentCoursesMap)
	assert.NoError(t, err)
	for idx, userNoti := range userNotis {
		student := students[idx]
		assert.Equal(t, userNoti.UserGroup.String, cpb.UserGroup_USER_GROUP_STUDENT.String())
		assert.Equal(t, userNoti.StudentID.String, student.StudentID.String)
		assert.Equal(t, userNoti.UserID.String, student.StudentID.String)
		assert.Equal(t, userNoti.NotificationID.String, notiID)
		assert.Equal(t, database.FromTextArray(userNoti.Courses), studentCoursesMap[student.StudentID.String])
		assert.Equal(t, userNoti.CurrentGrade.Int, int16(studentGradeMap[student.StudentID.String]))
		assert.Equal(t, userNoti.QuestionnaireStatus.String, cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED.String())
	}
}

func Test_PbToQuestionnaireUserAnswerEnts(t *testing.T) {
	t.Parallel()

	questionnaireSampleProto := utils.GenSampleQuestionnaire()
	answersQuestionnaireProto := utils.GenSampleAnswersForQuestionnaire(questionnaireSampleProto)
	userId := idutil.ULIDNow()
	targetId := idutil.ULIDNow()

	reqSubmit := &npb.SubmitQuestionnaireRequest{
		UserInfoNotificationId: idutil.ULIDNow(),
		QuestionnaireId:        idutil.ULIDNow(),
		Answers:                answersQuestionnaireProto,
	}

	checkCorrectlyMapping := func(reqSubmit *npb.SubmitQuestionnaireRequest, qnUserAnswerEnts entities.QuestionnaireUserAnswers) {
		for idx, userAnswer := range reqSubmit.Answers {
			qnUserAnswerEnt := qnUserAnswerEnts[idx]

			assert.Equal(t, qnUserAnswerEnt.Answer.String, userAnswer.Answer)
			assert.Equal(t, qnUserAnswerEnt.QuestionnaireQuestionID.String, userAnswer.QuestionnaireQuestionId)
			assert.Equal(t, qnUserAnswerEnt.UserNotificationID.String, reqSubmit.UserInfoNotificationId)
			assert.Equal(t, qnUserAnswerEnt.TargetID.String, targetId)
			assert.Equal(t, qnUserAnswerEnt.UserID.String, userId)
		}
	}

	t.Run("happy case", func(t *testing.T) {
		qnUserAnswerEnts, err := PbToQuestionnaireUserAnswerEnts(userId, targetId, reqSubmit)
		assert.NoError(t, err)
		checkCorrectlyMapping(reqSubmit, qnUserAnswerEnts)
	})

	t.Run("with empty answer", func(t *testing.T) {
		reqSubmit.Answers = append(reqSubmit.Answers, &cpb.Answer{
			QuestionnaireQuestionId: idutil.ULIDNow(),
			Answer:                  "",
		})

		answerWithoutEmpty := make([]*cpb.Answer, 0)
		for _, userAnswer := range reqSubmit.Answers {
			if userAnswer.Answer != "" {
				answerWithoutEmpty = append(answerWithoutEmpty, &cpb.Answer{
					QuestionnaireQuestionId: userAnswer.QuestionnaireQuestionId,
					Answer:                  userAnswer.Answer,
				})
			}
		}
		qnUserAnswerEnts, err := PbToQuestionnaireUserAnswerEnts(userId, targetId, reqSubmit)
		assert.NoError(t, err)

		reqSubmit.Answers = answerWithoutEmpty

		checkCorrectlyMapping(reqSubmit, qnUserAnswerEnts)
	})
}

func Test_PbToQuestionnaireEnt(t *testing.T) {
	questionnaireSampleProto := utils.GenSampleQuestionnaire()
	questionnaireEnt, err := PbToQuestionnaireEnt(questionnaireSampleProto)
	assert.NoError(t, err)
	assert.Equal(t, questionnaireSampleProto.ExpirationDate.AsTime(), questionnaireEnt.ExpirationDate.Time)
	assert.Equal(t, questionnaireSampleProto.ResubmitAllowed, questionnaireEnt.ResubmitAllowed.Bool)
	assert.Equal(t, questionnaireSampleProto.QuestionnaireId, questionnaireEnt.QuestionnaireID.String)
	assert.Equal(t, questionnaireSampleProto.QuestionnaireTemplateId, questionnaireEnt.QuestionnaireTemplateID.String)
}

func Test_PbToQuestionnaireQuestionEnts(t *testing.T) {
	questionnaireSampleProto := utils.GenSampleQuestionnaire()
	questionnaireQuestionEnts, err := PbToQuestionnaireQuestionEnts(questionnaireSampleProto)
	assert.NoError(t, err)

	for idx, question := range questionnaireSampleProto.Questions {
		questionEnt := questionnaireQuestionEnts[idx]

		assert.Equal(t, question.Type.String(), questionEnt.Type.String)
		assert.Equal(t, question.Choices, database.FromTextArray(questionEnt.Choices))
		assert.Equal(t, question.OrderIndex, int64(questionEnt.OrderIndex.Int))
		assert.Equal(t, question.Required, questionEnt.IsRequired.Bool)
		assert.Equal(t, question.QuestionnaireQuestionId, questionEnt.QuestionnaireQuestionID.String)
		assert.Equal(t, question.Title, questionEnt.Title.String)
	}
}

func seedRandomMap[T int | []string](ids []string, m map[string]T) {
	for _, id := range ids {
		if rand.Intn(2) == 0 {
			var sample T
			switch p := any(&sample).(type) {
			case *int:
				*p = rand.Intn(100)
			case *[]string:
				*p = []string{idutil.ULIDNow()}
			}
			m[id] = sample
		}
	}
}

func Test_AudienceToUserNotificationEnt(t *testing.T) {
	t.Parallel()
	parentID := "parent-id"
	studentID := "student-id"
	courseIDs := []string{"course-1", "course-2"}
	currentGrade := 1
	notiID := "noti-1"

	makeData := func(notiID, userGroup string, currentGrade int, courseIDs []string) (*entities.UserInfoNotification, *entities.Audience) {
		audience := &entities.Audience{}
		audience.UserID.Set(nil)
		audience.CurrentGrade.Set(currentGrade)
		audience.UserGroup.Set(userGroup)
		audience.ParentID.Set(nil)
		audience.StudentID.Set(nil)

		userNoti := &entities.UserInfoNotification{}
		userNoti.UserID.Set(nil)
		userNoti.NotificationID.Set(notiID)
		userNoti.CurrentGrade.Set(currentGrade)
		userNoti.UserGroup.Set(userGroup)
		userNoti.Courses.Set(nil)
		userNoti.ParentID.Set(nil)
		userNoti.StudentID.Set(nil)
		userNoti.IsIndividual.Set(false)
		userNoti.Status.Set(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String())
		userNoti.QuestionnaireStatus.Set(cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED)
		userNoti.UserNotificationID.Set(nil)
		userNoti.CreatedAt.Set(nil)
		userNoti.UpdatedAt.Set(nil)
		userNoti.QuestionnaireSubmittedAt.Set(nil)
		userNoti.GradeID.Set(nil)
		userNoti.DeletedAt.Set(nil)
		userNoti.StudentName.Set(nil)
		userNoti.ParentName.Set(nil)
		return userNoti, audience
	}

	t.Run("case parent", func(t *testing.T) {
		expectedNoti, audience := makeData(notiID, cpb.UserGroup_USER_GROUP_PARENT.String(), currentGrade, courseIDs)
		audience.UserID = database.Text(parentID)
		audience.ParentID = database.Text(parentID)
		audience.StudentID = database.Text(studentID)
		res, err := AudienceToUserNotificationEnt(notiID, audience)

		expectedNoti.UserID = database.Text(parentID)
		expectedNoti.ParentID = database.Text(parentID)
		expectedNoti.StudentID = database.Text(studentID)
		assert.Nil(t, err)
		assert.Equal(t, expectedNoti, res)
	})

	t.Run("case student", func(t *testing.T) {
		expectedNoti, audience := makeData(notiID, cpb.UserGroup_USER_GROUP_STUDENT.String(), currentGrade, courseIDs)
		audience.UserID = database.Text(studentID)
		audience.StudentID = database.Text(studentID)
		res, err := AudienceToUserNotificationEnt(notiID, audience)

		expectedNoti.UserID = database.Text(studentID)
		expectedNoti.StudentID = database.Text(studentID)
		assert.Nil(t, err)
		assert.Equal(t, expectedNoti, res)
	})
}

func Test_PbToTargetGroupFilterCoursesEnt(t *testing.T) {
	t.Parallel()
	courseIDs := []string{"course-id-1", "course-id-2"}
	courseNames := []string{"course-name-1", "course-name-2"}
	courses := make([]*cpb.NotificationTargetGroup_CourseFilter_Course, 0)
	courses = append(courses, &cpb.NotificationTargetGroup_CourseFilter_Course{
		CourseId:   courseIDs[0],
		CourseName: courseNames[0],
	})
	courses = append(courses, &cpb.NotificationTargetGroup_CourseFilter_Course{
		CourseId:   courseIDs[1],
		CourseName: courseNames[1],
	})

	t.Run("happy case", func(t *testing.T) {
		coursesRes := PbToTargetGroupFilterCoursesEnt(courses)

		assert.Equal(t, 2, len(coursesRes))
		assert.Equal(t, courseIDs[0], coursesRes[0].CourseID)
		assert.Equal(t, courseIDs[1], coursesRes[1].CourseID)
		assert.Equal(t, courseNames[0], coursesRes[0].CourseName)
		assert.Equal(t, courseNames[1], coursesRes[1].CourseName)
	})
}

func Test_PbToTargetGroupFilterLocationsEnt(t *testing.T) {
	t.Parallel()
	locationIDs := []string{"location-id-1", "location-id-2"}
	locationNames := []string{"location-name-1", "location-name-2"}
	locations := make([]*cpb.NotificationTargetGroup_LocationFilter_Location, 0)
	locations = append(locations, &cpb.NotificationTargetGroup_LocationFilter_Location{
		LocationId:   locationIDs[0],
		LocationName: locationNames[0],
	})
	locations = append(locations, &cpb.NotificationTargetGroup_LocationFilter_Location{
		LocationId:   locationIDs[1],
		LocationName: locationNames[1],
	})

	t.Run("happy case", func(t *testing.T) {
		locationsRes := PbToTargetGroupFilterLocationsEnt(locations)

		assert.Equal(t, 2, len(locationsRes))
		assert.Equal(t, locationIDs[0], locationsRes[0].LocationID)
		assert.Equal(t, locationIDs[1], locationsRes[1].LocationID)
		assert.Equal(t, locationNames[0], locationsRes[0].LocationName)
		assert.Equal(t, locationNames[1], locationsRes[1].LocationName)
	})
}

func Test_PbToTargetGroupFilterClassesEnt(t *testing.T) {
	t.Parallel()
	classIDs := []string{"class-id-1", "class-id-2"}
	classNames := []string{"class-name-1", "class-name-2"}
	classes := make([]*cpb.NotificationTargetGroup_ClassFilter_Class, 0)
	classes = append(classes, &cpb.NotificationTargetGroup_ClassFilter_Class{
		ClassId:   classIDs[0],
		ClassName: classNames[0],
	})
	classes = append(classes, &cpb.NotificationTargetGroup_ClassFilter_Class{
		ClassId:   classIDs[1],
		ClassName: classNames[1],
	})

	t.Run("happy case", func(t *testing.T) {
		classesRes := PbToTargetGroupFilterClassesEnt(classes)

		assert.Equal(t, 2, len(classesRes))
		assert.Equal(t, classIDs[0], classesRes[0].ClassID)
		assert.Equal(t, classIDs[1], classesRes[1].ClassID)
		assert.Equal(t, classNames[0], classesRes[0].ClassName)
		assert.Equal(t, classNames[1], classesRes[1].ClassName)
	})
}

func Test_EventStudentPackageV2PbToNotificationStudentCourseEnt(t *testing.T) {
	now := time.Now()
	studentPackage := &natspb.EventStudentPackageV2{
		StudentPackage: &natspb.EventStudentPackageV2_StudentPackageV2{
			StudentId: "student-id",
			Package: &natspb.EventStudentPackageV2_PackageV2{
				CourseId:   "course-id",
				LocationId: "location-id",
				ClassId:    "class-id",
				StartDate:  timestamppb.New(now),
				EndDate:    timestamppb.New(now),
			},
			IsActive: true,
		},
	}

	t.Run("happy case", func(t *testing.T) {
		notiStudentCourse, err := EventStudentPackageV2PbToNotificationStudentCourseEnt(studentPackage)

		assert.NoError(t, err)
		assert.Equal(t, notiStudentCourse.StudentID.String, studentPackage.StudentPackage.StudentId)
		assert.Equal(t, notiStudentCourse.CourseID.String, studentPackage.StudentPackage.Package.CourseId)
		assert.Equal(t, notiStudentCourse.LocationID.String, studentPackage.StudentPackage.Package.LocationId)
		assert.Equal(t, notiStudentCourse.StartAt.Time, studentPackage.StudentPackage.Package.StartDate.AsTime())
		assert.Equal(t, notiStudentCourse.EndAt.Time, studentPackage.StudentPackage.Package.EndDate.AsTime())
	})
}

func Test_EventStudentPackageJPRPEPbToNotificationStudentCourseEnts(t *testing.T) {
	now := time.Now()
	studentPackage := &natspb.EventSyncStudentPackage_StudentPackage{
		ActionKind: natspb.ActionKind_ACTION_KIND_NONE,
		StudentId:  "student-id",
		Packages: []*natspb.EventSyncStudentPackage_Package{
			{
				CourseIds: []string{"course-1"},
				StartDate: timestamppb.New(now),
				EndDate:   timestamppb.New(now),
			},
			{
				CourseIds: []string{"course-2", "course-3"},
				StartDate: timestamppb.New(now),
				EndDate:   timestamppb.New(now),
			},
		},
	}

	t.Run("happy case", func(t *testing.T) {
		notiStudentCourses, err := EventStudentPackageJPRPEPbToNotificationStudentCourseEnts(studentPackage)
		assert.NoError(t, err)
		studentID := studentPackage.StudentId

		idxStudentCourses := 0
		for _, pkg := range studentPackage.Packages {
			for _, course := range pkg.CourseIds {
				assert.Equal(t, notiStudentCourses[idxStudentCourses].StudentID.String, studentID)
				assert.Equal(t, notiStudentCourses[idxStudentCourses].CourseID.String, course)
				assert.Equal(t, notiStudentCourses[idxStudentCourses].LocationID.String, constants.JPREPOrgLocation)
				assert.Equal(t, notiStudentCourses[idxStudentCourses].StartAt.Time, pkg.StartDate.AsTime())
				assert.Equal(t, notiStudentCourses[idxStudentCourses].EndAt.Time, pkg.EndDate.AsTime())
				idxStudentCourses++
			}
		}
	})
}

func Test_EventStudentPackageV2PbToNotificationClassMemberEnt(t *testing.T) {
	now := time.Now()
	studentPackage := &natspb.EventStudentPackageV2{
		StudentPackage: &natspb.EventStudentPackageV2_StudentPackageV2{
			StudentId: "student-id",
			Package: &natspb.EventStudentPackageV2_PackageV2{
				CourseId:   "course-id",
				LocationId: "location-id",
				ClassId:    "class-id",
				StartDate:  timestamppb.New(now),
				EndDate:    timestamppb.New(now),
			},
			IsActive: true,
		},
	}

	t.Run("happy case", func(t *testing.T) {
		notiClassMember, err := EventStudentPackageV2PbToNotificationClassMemberEnt(studentPackage)

		assert.NoError(t, err)
		assert.Equal(t, notiClassMember.StudentID.String, studentPackage.StudentPackage.StudentId)
		assert.Equal(t, notiClassMember.CourseID.String, studentPackage.StudentPackage.Package.CourseId)
		assert.Equal(t, notiClassMember.ClassID.String, studentPackage.StudentPackage.Package.ClassId)
		assert.Equal(t, notiClassMember.LocationID.String, studentPackage.StudentPackage.Package.LocationId)
		assert.Equal(t, notiClassMember.StartAt.Time, studentPackage.StudentPackage.Package.StartDate.AsTime())
		assert.Equal(t, notiClassMember.EndAt.Time, studentPackage.StudentPackage.Package.EndDate.AsTime())
	})
}

var makeNotiClassMember = func(idx int, courseID, locationID string, deletedAt interface{}) *entities.NotificationClassMember {
	ent := &entities.NotificationClassMember{}
	database.AllNullEntity(ent)

	strIdx := strconv.Itoa(idx)
	strClassID := strconv.Itoa(1)
	ent.StudentID.Set("student-id-" + strIdx)
	ent.ClassID.Set(strClassID)
	ent.CourseID.Set(courseID)
	ent.LocationID.Set(locationID)
	ent.DeletedAt.Set(deletedAt)
	return ent
}

func Test_EventLeaveClassRoomToNotificationClassMemberEnts(t *testing.T) {
	leaveClassMsg := &pb.EvtClassRoom{
		Message: &pb.EvtClassRoom_LeaveClass_{
			LeaveClass: &pb.EvtClassRoom_LeaveClass{
				ClassId: 1,
				UserIds: []string{"student-id-1", "student-id-2"},
			},
		},
	}

	t.Run("happy case", func(t *testing.T) {
		courseID := "course-id"
		locationID := "location-id"
		expectedClassMembers := []*entities.NotificationClassMember{
			makeNotiClassMember(1, courseID, locationID, nil),
			makeNotiClassMember(2, courseID, locationID, nil),
		}
		cm, err := EventLeaveClassRoomToNotificationClassMemberEnts(leaveClassMsg.GetLeaveClass(), courseID, locationID, nil)
		assert.Nil(t, err)
		assert.Equal(t, expectedClassMembers, cm)
	})

	t.Run("happy case with delete", func(t *testing.T) {
		courseID := "course-id"
		locationID := "location-id"
		deletedAt := time.Now()
		expectedClassMembers := []*entities.NotificationClassMember{
			makeNotiClassMember(1, courseID, locationID, deletedAt),
			makeNotiClassMember(2, courseID, locationID, deletedAt),
		}
		cm, err := EventLeaveClassRoomToNotificationClassMemberEnts(leaveClassMsg.GetLeaveClass(), courseID, locationID, deletedAt)
		assert.Nil(t, err)
		assert.Equal(t, expectedClassMembers, cm)
	})
}

func Test_EventJoinClassRoomToNotificationClassMemberEnt(t *testing.T) {
	joinClassMsg := &pb.EvtClassRoom{
		Message: &pb.EvtClassRoom_JoinClass_{
			JoinClass: &pb.EvtClassRoom_JoinClass{
				ClassId: 1,
				UserId:  "student-id-1",
			},
		},
	}

	t.Run("happy case", func(t *testing.T) {
		courseID := "course-id"
		locationID := "location-id"
		expectedClassMember := makeNotiClassMember(1, courseID, locationID, nil)
		cm, err := EventJoinClassRoomToNotificationClassMemberEnt(joinClassMsg.GetJoinClass(), courseID, locationID)
		assert.Nil(t, err)
		assert.Equal(t, expectedClassMember, cm)
	})
}

func Test_PbToQuestionnaireTemplateEnt(t *testing.T) {
	questionnaireTemplateSample := utils.GetSampleQuestionnaireTemplate()
	questionnaireTemplateEnt, err := PbToQuestionnaireTemplateEnt(questionnaireTemplateSample)
	assert.NoError(t, err)
	assert.Equal(t, questionnaireTemplateSample.Name, questionnaireTemplateEnt.Name.String)
	assert.Equal(t, questionnaireTemplateSample.ExpirationDate.AsTime(), questionnaireTemplateEnt.ExpirationDate.Time)
	assert.Equal(t, questionnaireTemplateSample.ResubmitAllowed, questionnaireTemplateEnt.ResubmitAllowed.Bool)
	assert.Equal(t, questionnaireTemplateSample.QuestionnaireTemplateId, questionnaireTemplateEnt.QuestionnaireTemplateID.String)
	assert.Equal(t, npb.QuestionnaireTemplateType_QUESTION_TEMPLATE_TYPE_DEFAULT.String(), questionnaireTemplateEnt.Type.String)
}

func Test_PbToQuestionnaireTemplateQuestionEnts(t *testing.T) {
	questionnaireTemplateSample := utils.GetSampleQuestionnaireTemplate()
	questionnaireTemplateQuestionEnts, err := PbToQuestionnaireTemplateQuestionEnts(questionnaireTemplateSample)
	assert.NoError(t, err)

	for idx, question := range questionnaireTemplateSample.Questions {
		questionEnt := questionnaireTemplateQuestionEnts[idx]

		assert.Equal(t, question.Type.String(), questionEnt.Type.String)
		assert.Equal(t, question.Choices, database.FromTextArray(questionEnt.Choices))
		assert.Equal(t, question.OrderIndex, int64(questionEnt.OrderIndex.Int))
		assert.Equal(t, question.Required, questionEnt.IsRequired.Bool)
		assert.Equal(t, question.QuestionnaireTemplateQuestionId, questionEnt.QuestionnaireTemplateQuestionID.String)
		assert.Equal(t, question.Title, questionEnt.Title.String)
	}
}
