package mappers

import (
	"math"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
)

func Test_QNQuestionsToPb(t *testing.T) {
	t.Parallel()
	questionnaireId := idutil.ULIDNow()

	checkQNQuestions := func(t *testing.T, qnQuestionsEnts entities.QuestionnaireQuestions, qnQuestionsPb []*cpb.Question) {
		for idx, question := range qnQuestionsPb {
			questionEnt := qnQuestionsEnts[idx]

			assert.Equal(t, question.Type.String(), questionEnt.Type.String)
			assert.Equal(t, question.Choices, database.FromTextArray(questionEnt.Choices))
			assert.Equal(t, question.OrderIndex, int64(questionEnt.OrderIndex.Int))
			assert.Equal(t, question.Required, questionEnt.IsRequired.Bool)
			assert.Equal(t, question.QuestionnaireQuestionId, questionEnt.QuestionnaireQuestionID.String)
			assert.Equal(t, question.Title, questionEnt.Title.String)
		}
	}

	t.Run("happy case", func(t *testing.T) {
		questionnaireQuestion := utils.GenQNQuestions(questionnaireId)

		qnQuestionPb := QNQuestionsToPb(questionnaireQuestion)
		checkQNQuestions(t, questionnaireQuestion, qnQuestionPb)
	})
}

func Test_QNUserAnswersToPb(t *testing.T) {
	t.Parallel()
	timeSubmitted := time.Now()
	responders := []*repositories.QuestionnaireResponder{
		{
			UserID:      database.Text(idutil.ULIDNow()),
			TargetID:    database.Text(idutil.ULIDNow()),
			Name:        database.Text(idutil.ULIDNow()),
			TargetName:  database.Text(idutil.ULIDNow()),
			SubmittedAt: database.Timestamptz(timeSubmitted),
		},
		{
			UserID:      database.Text(idutil.ULIDNow()),
			TargetID:    database.Text(idutil.ULIDNow()),
			Name:        database.Text(idutil.ULIDNow()),
			TargetName:  database.Text(idutil.ULIDNow()),
			SubmittedAt: database.Timestamptz(timeSubmitted),
		},
	}

	questionnaireUserAnswers := entities.QuestionnaireUserAnswers{}
	questionnaireQuestions := utils.GenQNQuestions(idutil.ULIDNow())

	checkResponders := func(t *testing.T, responders []*repositories.QuestionnaireResponder, userAnswersProto []*npb.GetAnswersByFilterResponse_UserAnswer) {
		for idx, userAnswerProto := range userAnswersProto {
			responder := responders[idx]

			assert.Equal(t, userAnswerProto.UserId, responder.UserID.String)
			assert.Equal(t, userAnswerProto.ResponderName, responder.Name.String)
			assert.Equal(t, userAnswerProto.TargetId, responder.TargetID.String)
			assert.Equal(t, userAnswerProto.TargetName, responder.TargetName.String)
			assert.Equal(t, userAnswerProto.SubmittedAt.AsTime(), responder.SubmittedAt.Time.UTC())
			assert.Equal(t, userAnswerProto.UserNotificationId, responder.UserNotificationID.String)
		}
	}

	checkQuesionnaireUserAnswers := func(t *testing.T, questionnaireUserAnswers entities.QuestionnaireUserAnswers, userAnswersProto []*npb.GetAnswersByFilterResponse_UserAnswer, questionnaireQuestions entities.QuestionnaireQuestions) {
		// Collect all responders answers to check valid data
		answers := []*cpb.Answer{}
		for _, userAnswerProto := range userAnswersProto {
			answers = append(answers, userAnswerProto.Answers...)
		}

		countCorrect := 0
		for _, questionnaireUserAnswer := range questionnaireUserAnswers {
			for _, answer := range answers {
				if answer.Answer == questionnaireUserAnswer.Answer.String &&
					answer.QuestionnaireQuestionId == questionnaireUserAnswer.QuestionnaireQuestionID.String {
					countCorrect++
				}
			}
		}
		assert.Equal(t, len(answers), countCorrect)

		// Check order answer for each responder
		mapQuestionIdAndOrderIndex := make(map[string]int64)
		for _, question := range questionnaireQuestions {
			mapQuestionIdAndOrderIndex[question.QuestionnaireQuestionID.String] = int64(question.OrderIndex.Int)
		}

		for _, userAnswerProto := range userAnswersProto {
			countInvalidOrder := 0
			for i := 0; i < len(userAnswerProto.Answers)-1; i++ {
				if mapQuestionIdAndOrderIndex[userAnswerProto.Answers[i].QuestionnaireQuestionId] > mapQuestionIdAndOrderIndex[userAnswerProto.Answers[i+1].QuestionnaireQuestionId] {
					countInvalidOrder++
				}
			}
			assert.Equal(t, 0, countInvalidOrder)
		}
	}

	t.Run("happy case", func(t *testing.T) {
		for _, responder := range responders {
			for _, questionnaireQuestion := range questionnaireQuestions {
				answer := utils.GenQNUserAnswer(responder.UserID.String, responder.TargetID.String, questionnaireQuestion.QuestionnaireQuestionID.String)
				questionnaireUserAnswers = append(questionnaireUserAnswers, &answer)
			}
		}

		userAnswersPb := QNUserAnswersToPb(responders, questionnaireUserAnswers, questionnaireQuestions)
		checkResponders(t, responders, userAnswersPb)
		checkQuesionnaireUserAnswers(t, questionnaireUserAnswers, userAnswersPb, questionnaireQuestions)
	})

	t.Run("empty some answer", func(t *testing.T) {
		// Don't add answer for some question
		for _, responder := range responders {
			for index, questionnaireQuestion := range questionnaireQuestions {
				if math.Mod(float64(index), 2) == 1.0 {
					continue
				}
				answer := utils.GenQNUserAnswer(responder.UserID.String, responder.TargetID.String, questionnaireQuestion.QuestionnaireQuestionID.String)
				questionnaireUserAnswers = append(questionnaireUserAnswers, &answer)
			}
		}

		userAnswersPb := QNUserAnswersToPb(responders, questionnaireUserAnswers, questionnaireQuestions)

		// Add empty answer for missed question to check
		for _, responder := range responders {
			for index, questionnaireQuestion := range questionnaireQuestions {
				if math.Mod(float64(index), 2) == 1.0 {
					answer := entities.QuestionnaireUserAnswer{
						QuestionnaireQuestionID: database.Text(questionnaireQuestion.QuestionnaireQuestionID.String),
						Answer:                  database.Text(""),
						UserID:                  responder.UserID,
					}
					questionnaireUserAnswers = append(questionnaireUserAnswers, &answer)
				}
			}
		}

		checkResponders(t, responders, userAnswersPb)
		checkQuesionnaireUserAnswers(t, questionnaireUserAnswers, userAnswersPb, questionnaireQuestions)
	})
}

func Test_ToQNUserAnswerProtobuf(t *testing.T) {
	t.Parallel()
	userID, targetID := "parent1", "student_1"
	qnnaire := &entities.Questionnaire{}
	database.AllRandomEntity(qnnaire)
	qn := utils.GenQuestionaire()
	qnqs := utils.GenQNQuestions(qn.QuestionnaireID.String)
	answers := utils.GenQNUserAnswers(userID, targetID)
	protobuf := QNUserAnswerToPb(&qn, qnqs, answers, true)
	protoQn := protobuf.Questionnaire
	protoQus := protobuf.Questionnaire.Questions
	protoAnswers := protobuf.Answers
	assert.True(t, qn.ExpirationDate.Time.Equal(protoQn.ExpirationDate.AsTime()))
	assert.Equal(t, qn.QuestionnaireID.String, protoQn.QuestionnaireId)
	assert.Equal(t, qn.ResubmitAllowed.Bool, protoQn.ResubmitAllowed)
	assert.Equal(t, len(qnqs), len(protoQus))
	for i := 0; i < len(qnqs); i++ {
		assert.Equal(t, database.FromTextArray(qnqs[i].Choices), protoQus[i].Choices)
		assert.Equal(t, qnqs[i].Title.String, protoQus[i].Title)
		assert.Equal(t, qnqs[i].QuestionnaireQuestionID.String, protoQus[i].QuestionnaireQuestionId)
		assert.Equal(t, qnqs[i].Type.String, protoQus[i].Type.String())
		assert.Equal(t, int64(qnqs[i].OrderIndex.Int), protoQus[i].OrderIndex)
		assert.Equal(t, qnqs[i].IsRequired.Bool, protoQus[i].Required)
	}

	assert.Equal(t, len(protoAnswers), len(answers))
	for i := 0; i < len(answers); i++ {
		assert.Equal(t, answers[i].Answer.String, protoAnswers[i].Answer)
		assert.Equal(t, answers[i].QuestionnaireQuestionID.String, protoAnswers[i].QuestionnaireQuestionId)
	}
}

func Test_ToNotificationPb(t *testing.T) {
	t.Parallel()
	checkNotiMsg := func(t *testing.T, notiMsgEnt *entities.InfoNotificationMsg, notiPb *cpb.Notification) {
		content, _ := notiMsgEnt.GetContent()

		assert.Equal(t, notiMsgEnt.Title.String, notiPb.Message.Title)
		assert.Equal(t, content.Raw, notiPb.Message.Content.Raw)
		assert.Equal(t, content.RenderedURL, notiPb.Message.Content.Rendered)
		assert.Equal(t, database.FromTextArray(notiMsgEnt.MediaIDs), notiPb.Message.MediaIds)
	}

	checkNoti := func(t *testing.T, notiEnt *entities.InfoNotification, notiPb *cpb.Notification) {
		targetGroup, _ := notiEnt.GetTargetGroup()
		userGroups := make([]cpb.UserGroup, 0)
		for _, gr := range targetGroup.UserGroupFilter.UserGroups {
			userGroups = append(userGroups, cpb.UserGroup(cpb.UserGroup_value[gr]))
		}
		assert.Equal(t, notiEnt.NotificationID.String, notiPb.NotificationId)
		assert.Equal(t, notiEnt.NotificationMsgID.String, notiPb.Message.NotificationMsgId)
		assert.Equal(t, notiEnt.Type.String, notiPb.Type.String())
		assert.Equal(t, string(notiEnt.Data.Bytes), notiPb.Data)
		assert.Equal(t, notiEnt.CreatedUserID.String, notiPb.CreatedUserId)
		assert.Equal(t, notiEnt.EditorID.String, notiPb.EditorId)
		assert.Equal(t, targetGroup.CourseFilter.CourseIDs, notiPb.TargetGroup.CourseFilter.CourseIds)
		assert.Equal(t, targetGroup.CourseFilter.Type, notiPb.TargetGroup.CourseFilter.Type.String())
		assert.Equal(t, targetGroup.GradeFilter.GradeIDs, notiPb.TargetGroup.GradeFilter.GradeIds)
		assert.Equal(t, targetGroup.GradeFilter.Type, notiPb.TargetGroup.GradeFilter.Type.String())
		assert.Equal(t, userGroups, notiPb.TargetGroup.UserGroupFilter.UserGroups)
		assert.Equal(t, database.FromTextArray(notiEnt.ReceiverIDs), notiPb.ReceiverIds)
		assert.Equal(t, notiEnt.Event.String, notiPb.Event.String())
		assert.Equal(t, notiEnt.Status.String, notiPb.Status.String())
		assert.True(t, notiEnt.ScheduledAt.Time.Round(time.Second).Equal(notiPb.ScheduledAt.AsTime().Round(time.Second)))
		assert.True(t, notiEnt.CreatedAt.Time.Round(time.Second).Equal(notiPb.CreatedAt.AsTime().Round(time.Second)))
	}

	t.Run("happy case", func(t *testing.T) {
		notiEnt, notiMsgEnt := utils.GenSampleNotificationWithMsg()

		notiPb := NotificationToPb(notiEnt, notiMsgEnt)
		checkNotiMsg(t, notiMsgEnt, notiPb)
		checkNoti(t, notiEnt, notiPb)
	})

	t.Run("non empty data", func(t *testing.T) {
		notiEnt, notiMsgEnt := utils.GenSampleNotificationWithMsg()
		notiEnt.Data.Set(`{"key": "value"}`)

		notiPb := NotificationToPb(notiEnt, notiMsgEnt)

		checkNotiMsg(t, notiMsgEnt, notiPb)
		checkNoti(t, notiEnt, notiPb)
	})
}

func Test_ToUserNotificationPb(t *testing.T) {
	t.Parallel()
	checkUserNoti := func(t *testing.T, notiEnt *entities.UserInfoNotification, notiPb *cpb.UserNotification) {
		assert.Equal(t, notiEnt.UserID.String, notiPb.UserId)
		assert.Equal(t, notiEnt.NotificationID.String, notiPb.NotificationId)
		assert.Equal(t, database.FromTextArray(notiEnt.Courses), notiPb.CourseId)
		assert.Equal(t, notiEnt.Status.String, notiPb.Status.String())
		assert.True(t, notiEnt.CreatedAt.Time.Round(time.Second).Equal(notiPb.CreatedAt.AsTime().Round(time.Second)))
		assert.True(t, notiEnt.UpdatedAt.Time.Round(time.Second).Equal(notiPb.UpdatedAt.AsTime().Round(time.Second)))
	}

	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		var userNotiEnt entities.UserInfoNotification
		database.AllRandomEntity(&userNotiEnt)
		userNotiEnt.Status.Set(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String())
		userNotiPb := ToUserNotificationPb(&userNotiEnt)
		checkUserNoti(t, &userNotiEnt, userNotiPb)
	})

	t.Run("happy case status read", func(t *testing.T) {
		t.Parallel()
		var userNotiEnt entities.UserInfoNotification
		database.AllRandomEntity(&userNotiEnt)
		userNotiEnt.Status = database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ.String())
		userNotiPb := ToUserNotificationPb(&userNotiEnt)
		checkUserNoti(t, &userNotiEnt, userNotiPb)
	})
}

func Test_NotificationTargetToPb(t *testing.T) {
	t.Parallel()
	checkTargetGroup := func(t *testing.T, expectPb *entities.InfoNotificationTarget, currentEnt *cpb.NotificationTargetGroup) {
		assert.Equal(t, expectPb.CourseFilter.Type, currentEnt.CourseFilter.Type.String())
		assert.Equal(t, expectPb.CourseFilter.CourseIDs, currentEnt.CourseFilter.CourseIds)
		assert.Equal(t, expectPb.GradeFilter.Type, currentEnt.GradeFilter.Type.String())
		assert.Equal(t, expectPb.GradeFilter.GradeIDs, currentEnt.GradeFilter.GradeIds)
		assert.Equal(t, expectPb.ClassFilter.Type, currentEnt.ClassFilter.Type.String())
		assert.Equal(t, expectPb.ClassFilter.ClassIDs, currentEnt.ClassFilter.ClassIds)
		assert.Equal(t, expectPb.LocationFilter.Type, currentEnt.LocationFilter.Type.String())
		assert.Equal(t, expectPb.LocationFilter.LocationIDs, currentEnt.LocationFilter.LocationIds)
	}

	t.Run("happy case all course", func(t *testing.T) {
		t.Parallel()
		notiEnt, _ := utils.GenSampleNotificationWithMsg()
		targetGroupEnt, err := notiEnt.GetTargetGroup()
		targetGroupEnt.CourseFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL.String()
		targetGroupEnt.CourseFilter.CourseIDs = nil
		assert.Nil(t, err)
		targetGroupPb := NotificationTargetToPb(targetGroupEnt)
		checkTargetGroup(t, targetGroupEnt, targetGroupPb)
	})

	t.Run("happy case none course", func(t *testing.T) {
		t.Parallel()
		notiEnt, _ := utils.GenSampleNotificationWithMsg()
		targetGroupEnt, err := notiEnt.GetTargetGroup()
		targetGroupEnt.CourseFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String()
		targetGroupEnt.CourseFilter.CourseIDs = make([]string, 0)
		assert.Nil(t, err)
		targetGroupPb := NotificationTargetToPb(targetGroupEnt)
		checkTargetGroup(t, targetGroupEnt, targetGroupPb)
	})

	t.Run("happy case all grade", func(t *testing.T) {
		t.Parallel()
		notiEnt, _ := utils.GenSampleNotificationWithMsg()
		targetGroupEnt, err := notiEnt.GetTargetGroup()
		targetGroupEnt.GradeFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL.String()
		targetGroupEnt.GradeFilter.GradeIDs = nil
		assert.Nil(t, err)
		targetGroupPb := NotificationTargetToPb(targetGroupEnt)
		checkTargetGroup(t, targetGroupEnt, targetGroupPb)
	})

	t.Run("happy case none grade", func(t *testing.T) {
		t.Parallel()
		notiEnt, _ := utils.GenSampleNotificationWithMsg()
		targetGroupEnt, err := notiEnt.GetTargetGroup()
		targetGroupEnt.GradeFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String()
		targetGroupEnt.GradeFilter.GradeIDs = make([]string, 0)
		assert.Nil(t, err)
		targetGroupPb := NotificationTargetToPb(targetGroupEnt)
		checkTargetGroup(t, targetGroupEnt, targetGroupPb)
	})

	t.Run("happy case all class", func(t *testing.T) {
		t.Parallel()
		notiEnt, _ := utils.GenSampleNotificationWithMsg()
		targetGroupEnt, err := notiEnt.GetTargetGroup()
		targetGroupEnt.ClassFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL.String()
		targetGroupEnt.ClassFilter.ClassIDs = nil
		assert.Nil(t, err)
		targetGroupPb := NotificationTargetToPb(targetGroupEnt)
		checkTargetGroup(t, targetGroupEnt, targetGroupPb)
	})

	t.Run("happy case none class", func(t *testing.T) {
		t.Parallel()
		notiEnt, _ := utils.GenSampleNotificationWithMsg()
		targetGroupEnt, err := notiEnt.GetTargetGroup()
		targetGroupEnt.ClassFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String()
		targetGroupEnt.ClassFilter.ClassIDs = make([]string, 0)
		assert.Nil(t, err)
		targetGroupPb := NotificationTargetToPb(targetGroupEnt)
		checkTargetGroup(t, targetGroupEnt, targetGroupPb)
	})

	t.Run("happy case none location", func(t *testing.T) {
		t.Parallel()
		notiEnt, _ := utils.GenSampleNotificationWithMsg()
		targetGroupEnt, err := notiEnt.GetTargetGroup()
		targetGroupEnt.LocationFilter.Type = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String()
		targetGroupEnt.LocationFilter.LocationIDs = make([]string, 0)
		assert.Nil(t, err)
		targetGroupPb := NotificationTargetToPb(targetGroupEnt)
		checkTargetGroup(t, targetGroupEnt, targetGroupPb)
	})
}

func Test_NotificationsFilteredToPb(t *testing.T) {
	checkNoti := func(t *testing.T, notifications []*entities.InfoNotification, notiMsgMap map[string]*entities.InfoNotificationMsg, noificationsTags map[string]entities.InfoNotificationsTags, res []*npb.GetNotificationsByFilterResponse_Notification) {
		for idx, noti := range notifications {
			notiMsg := notiMsgMap[noti.NotificationID.String]
			targetGroup, _ := noti.GetTargetGroup()
			userGroups := make([]cpb.UserGroup, 0)

			for _, gr := range targetGroup.UserGroupFilter.UserGroups {
				userGroups = append(userGroups, cpb.UserGroup(cpb.UserGroup_value[gr]))
			}

			tagIDs := []string{}
			notiTags, ok := noificationsTags[noti.NotificationID.String]
			if ok {
				for _, notiTag := range notiTags {
					tagIDs = append(tagIDs, notiTag.TagID.String)
				}
			}

			assert.Equal(t, noti.NotificationID.String, res[idx].NotificationId)
			assert.Equal(t, noti.NotificationMsgID.String, res[idx].NotificationMgsId)
			assert.Equal(t, noti.Status.String, res[idx].Status.String())
			assert.Equal(t, notiMsg.Title.String, res[idx].Title)
			assert.Equal(t, noti.EditorID.String, res[idx].ComposerId)
			assert.Equal(t, userGroups, res[idx].UserGroupFilter.UserGroups)
			assert.True(t, noti.SentAt.Time.Round(time.Second).Equal(res[idx].SentAt.AsTime().Round(time.Second)))
			assert.True(t, noti.UpdatedAt.Time.Round(time.Second).Equal(res[idx].UpdatedAt.AsTime().Round(time.Second)))
			assert.Equal(t, tagIDs, res[idx].TagIds)

			expectTargetGroup := &entities.InfoNotificationTarget{}
			err := noti.TargetGroups.AssignTo(expectTargetGroup)
			assert.Nil(t, err)
			assert.Equal(t, NotificationTargetToPb(expectTargetGroup), res[idx].TargetGroup)
		}
	}

	t.Run("happy case without tags", func(t *testing.T) {
		notiEnts := []*entities.InfoNotification{}
		notiMsgMap := make(map[string]*entities.InfoNotificationMsg, 0)
		for i := 0; i < 10; i++ {
			notiEnt, notiMsgEnt := utils.GenSampleNotificationWithMsg()
			notiEnts = append(notiEnts, notiEnt)
			notiMsgMap[notiEnt.NotificationID.String] = notiMsgEnt
		}

		notiPbs, err := NotificationsFilteredToPb(notiEnts, notiMsgMap, nil)
		assert.Nil(t, err)
		checkNoti(t, notiEnts, notiMsgMap, nil, notiPbs)
	})

	t.Run("happy case with tags", func(t *testing.T) {
		notiEnts := []*entities.InfoNotification{}
		notiMsgMap := make(map[string]*entities.InfoNotificationMsg, 0)
		notiTagsMap := make(map[string]entities.InfoNotificationsTags, 0)
		for i := 0; i < 10; i++ {
			notiEnt, notiMsgEnt := utils.GenSampleNotificationWithMsg()
			notiEnts = append(notiEnts, notiEnt)
			notiMsgMap[notiEnt.NotificationID.String] = notiMsgEnt

			for j := 0; j < 3; j++ {
				notiTagEnt := &entities.InfoNotificationTag{}
				database.AllRandomEntity(notiTagEnt)
				notiTagEnt.NotificationID = notiEnt.NotificationID
				notiTagsMap[notiEnt.NotificationID.String] = append(notiTagsMap[notiEnt.NotificationID.String], notiTagEnt)
			}
		}

		notiPbs, err := NotificationsFilteredToPb(notiEnts, notiMsgMap, nil)
		assert.Nil(t, err)
		checkNoti(t, notiEnts, notiMsgMap, nil, notiPbs)
	})
}

func Test_NotificationTargetCoursesToPb(t *testing.T) {
	t.Parallel()
	courseIDs := []string{"course-id-1", "course-id-2"}
	courseNames := []string{"course-name-1", "course-name-2"}

	courses := make([]entities.InfoNotificationTarget_CourseFilter_Course, 0)
	courses = append(courses, entities.InfoNotificationTarget_CourseFilter_Course{
		CourseID:   courseIDs[0],
		CourseName: courseNames[0],
	})
	courses = append(courses, entities.InfoNotificationTarget_CourseFilter_Course{
		CourseID:   courseIDs[1],
		CourseName: courseNames[1],
	})

	t.Run("happy case", func(t *testing.T) {
		coursesRes := NotificationTargetCoursesToPb(courses)

		assert.Equal(t, 2, len(coursesRes))
		assert.Equal(t, courseIDs[0], coursesRes[0].CourseId)
		assert.Equal(t, courseIDs[1], coursesRes[1].CourseId)
		assert.Equal(t, courseNames[0], coursesRes[0].CourseName)
		assert.Equal(t, courseNames[1], coursesRes[1].CourseName)
	})
}

func Test_NotificationTargetLocationsToPb(t *testing.T) {
	t.Parallel()
	locationIDs := []string{"location-id-1", "location-id-2"}
	locationNames := []string{"location-name-1", "location-name-2"}

	locations := make([]entities.InfoNotificationTarget_LocationFilter_Location, 0)
	locations = append(locations, entities.InfoNotificationTarget_LocationFilter_Location{
		LocationID:   locationIDs[0],
		LocationName: locationNames[0],
	})
	locations = append(locations, entities.InfoNotificationTarget_LocationFilter_Location{
		LocationID:   locationIDs[1],
		LocationName: locationNames[1],
	})

	t.Run("happy case", func(t *testing.T) {
		locationsRes := NotificationTargetLocationsToPb(locations)

		assert.Equal(t, 2, len(locationsRes))
		assert.Equal(t, locationIDs[0], locationsRes[0].LocationId)
		assert.Equal(t, locationIDs[1], locationsRes[1].LocationId)
		assert.Equal(t, locationNames[0], locationsRes[0].LocationName)
		assert.Equal(t, locationNames[1], locationsRes[1].LocationName)
	})
}

func Test_NotificationTargetClassesToPb(t *testing.T) {
	t.Parallel()
	classIDs := []string{"class-id-1", "class-id-2"}
	classNames := []string{"class-name-1", "class-name-2"}

	classes := make([]entities.InfoNotificationTarget_ClassFilter_Class, 0)
	classes = append(classes, entities.InfoNotificationTarget_ClassFilter_Class{
		ClassID:   classIDs[0],
		ClassName: classNames[0],
	})
	classes = append(classes, entities.InfoNotificationTarget_ClassFilter_Class{
		ClassID:   classIDs[1],
		ClassName: classNames[1],
	})

	t.Run("happy case", func(t *testing.T) {
		classesRes := NotificationTargetClassesToPb(classes)

		assert.Equal(t, 2, len(classesRes))
		assert.Equal(t, classIDs[0], classesRes[0].ClassId)
		assert.Equal(t, classIDs[1], classesRes[1].ClassId)
		assert.Equal(t, classNames[0], classesRes[0].ClassName)
		assert.Equal(t, classNames[1], classesRes[1].ClassName)
	})
}
