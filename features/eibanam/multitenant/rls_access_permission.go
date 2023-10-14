package multitenant

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuo_pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bob_pbv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	com_pbv1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) initRLSDataInSchools(ctx context.Context, schoolIDs ...int32) error {
	for _, schoolID := range schoolIDs {
		// Create books
		bookID := idutil.ULIDNow()
		upsertBookReq := &epb.UpsertBooksRequest{
			Books: []*epb.UpsertBooksRequest_Book{
				{
					BookId: bookID,
					Name:   fmt.Sprintf("book-name+%s", bookID),
				},
			},
		}
		upsertBookResp, err := epb.NewBookModifierServiceClient(s.eurekaConn).UpsertBooks(ctx, upsertBookReq)
		if err != nil {
			return errors.Wrap(err, "UpsertBooks()")
		}
		s.RequestStack.Push(upsertBookReq)
		s.ResponseStack.Push(upsertBookResp)

		// Create chapters
		chapterID := idutil.ULIDNow()
		upsertChapterReq := &epb.UpsertChaptersRequest{
			Chapters: []*com_pbv1.Chapter{
				{
					Info: &com_pbv1.ContentBasicInfo{
						Id:           chapterID,
						Name:         fmt.Sprintf("chapter-name+%s", chapterID),
						Country:      com_pbv1.Country_COUNTRY_VN,
						Subject:      com_pbv1.Subject_SUBJECT_BIOLOGY,
						Grade:        cast.ToInt32(i18n.OutGradeMapV1[com_pbv1.Country_COUNTRY_VN][rand.Intn(12)]),
						DisplayOrder: 1,
						SchoolId:     schoolID,
					},
				},
			},
			BookId: bookID,
		}
		upsertChapterResp, err := epb.NewChapterModifierServiceClient(s.eurekaConn).UpsertChapters(ctx, upsertChapterReq)
		if err != nil {
			return errors.Wrap(err, "UpsertChapter()")
		}
		s.RequestStack.Push(upsertChapterReq)
		s.ResponseStack.Push(upsertChapterResp)

		// Create topics
		topicID := idutil.ULIDNow()
		upsertTopicsReq := &epb.UpsertTopicsRequest{
			Topics: []*epb.Topic{
				{
					Id:           topicID,
					Name:         fmt.Sprintf("topic-name+%s", topicID),
					Country:      epb.Country_COUNTRY_VN,
					Grade:        i18n.OutGradeMap[bob_pb.COUNTRY_VN][rand.Intn(12)],
					Subject:      epb.Subject_SUBJECT_BIOLOGY,
					Type:         epb.TopicType_TOPIC_TYPE_LEARNING,
					CreatedAt:    timestamppb.Now(),
					UpdatedAt:    timestamppb.Now(),
					DisplayOrder: int32(1),
					TotalLos:     1,
					ChapterId:    chapterID,
					SchoolId:     schoolID,
				},
			},
		}
		upsertTopicsResp, err := epb.NewTopicModifierServiceClient(s.eurekaConn).Upsert(ctx, upsertTopicsReq)
		s.RequestStack.Push(upsertTopicsReq)
		s.ResponseStack.Push(upsertTopicsResp)

		// Create courses
		courseID := idutil.ULIDNow()
		country := bob_pb.COUNTRY_VN
		grade, err := i18n.ConvertIntGradeToString(country, 1)
		if err != nil {
			return errors.Wrap(err, "ConvertIntGradeToString")
		}

		upsertCourseReq := &yasuo_pb.UpsertCoursesRequest{
			Courses: []*yasuo_pb.UpsertCoursesRequest_Course{
				{
					Id:           courseID,
					Name:         courseID,
					Country:      country,
					Subject:      bob_pb.SUBJECT_MATHS,
					Grade:        grade,
					DisplayOrder: 1,
					SchoolId:     schoolID,
					Icon:         "https://example-url.com",
					BookIds:      []string{bookID},
				},
			},
		}
		upsertCourseResp, err := yasuo_pb.NewCourseServiceClient(s.yasuoConn).UpsertCourses(ctx, upsertCourseReq)
		if err != nil {
			return errors.Wrap(err, "UpsertCourses()")
		}
		s.RequestStack.Push(upsertCourseReq)
		s.ResponseStack.Push(upsertCourseResp)
	}
	return nil
}

func (s *suite) superAdminLoginsOnCMS(ctx context.Context) (context.Context, error) {
	return ctx, s.aSignedInAdmin()
}

func (s *suite) superAdminSeesAllDataOfAllOrganizationOnCMS(ctx context.Context) (context.Context, error) {
	// Setup context
	nCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	s.UserGroupInContext = constant.UserGroupAdmin

	// Init test data if not exists
	if len(s.RequestStack.Requests) < 1 {
		if err := s.initRLSDataInSchools(contextWithTokenForGrpcCall(s, ctx), 1, 2); err != nil {
			return ctx, errors.Wrap(err, "initRLSDataInSchools")
		}
	}

	// Admin see courses in different schools
	nCtx = contextWithToken(s, nCtx)

	for _, req := range s.RequestStack.Requests {
		switch req := req.(type) {
		case *yasuo_pb.UpsertCoursesRequest:
			err := s.canSeeCourseOnBackOffice(nCtx, req.Courses[0].Id, true)
			if err != nil {
				return ctx, errors.Wrap(err, "canSeeCourseOnBackOffice")
			}
		case *epb.UpsertBooksRequest:
			err := canSeeBookOnBackOffice(nCtx, req.Books[0].BookId, true)
			if err != nil {
				return ctx, errors.Wrap(err, "canSeeBookOnBackOffice")
			}
		case *epb.UpsertTopicsRequest:
			err := canSeeTopicOnBackOffice(nCtx, req.Topics[0].Id, true)
			if err != nil {
				return ctx, errors.Wrap(err, "canSeeTopicOnBackOffice")
			}
		case *epb.UpsertChaptersRequest:
			err := canSeeChapterOnBackOffice(nCtx, req.Chapters[0].Info.Id, true)
			if err != nil {
				return ctx, errors.Wrap(err, "canSeeChapterOnBackOffice")
			}
		}
	}
	return ctx, nil
}

func (s *suite) loginsOnCMS(ctx context.Context, actor string) (context.Context, error) {
	texts := strings.Split(actor, " ")

	role := strings.Join(texts[:len(texts)-1], " ")
	ordinal, err := strconv.Atoi(texts[len(texts)-1])
	if err != nil {
		return ctx, errors.Wrap(err, "can't get actor ordinal")
	}

	s.CurrentSchoolID = int32(ordinal)

	return ctx, s.signedInAsAccount(role)
}

func (s *suite) logsOutOnCMS(actor string) error {
	return nil
}

func (s *suite) onlyInteractsWithContentFromOnCMS(ctx context.Context, actor, org string) (context.Context, error) {
	// Init data to test if not available
	if len(s.RequestStack.Requests) < 1 {
		// Setup context
		initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := s.aSignedInAdmin(); err != nil {
			return ctx, errors.Wrap(err, "aSignedInAdmin")
		}
		s.UserGroupInContext = constant.UserGroupAdmin
		ctx := contextWithTokenForGrpcCall(s, initCtx)
		if err := s.initRLSDataInSchools(ctx, 1, 2); err != nil {
			return ctx, errors.Wrap(err, "addCourseToSchool")
		}
	}

	// Get org id
	texts := strings.Split(org, " ")
	orgID, err := strconv.Atoi(texts[1])
	if err != nil {
		return ctx, errors.Wrap(err, "strconv.Atoi")
	}

	// See data in corresponding schools
	texts = strings.Split(actor, " ")
	actorRole := strings.Join(texts[:len(texts)-1], " ")

	switch actorRole {
	case "teacher":
		s.UserGroupInContext = constant.UserGroupTeacher
		if err := s.teacherSeeAllDataInOrg(contextWithTokenForGrpcCall(s, ctx), orgID); err != nil {
			return ctx, errors.Wrap(err, "teacherSeeAllDataInOrg")
		}
	case "school admin":
		s.UserGroupInContext = constant.UserGroupSchoolAdmin
		if err := s.schoolAdminSeeAllDataInOrg(contextWithToken(s, ctx), orgID); err != nil {
			return ctx, errors.Wrap(err, "schoolAdminSeeAllDataInOrg")
		}
	}
	return ctx, nil
}

func (s *suite) teacherSeeAllDataInOrg(ctx context.Context, orgID int) error {
	for _, req := range s.RequestStack.Requests {
		switch req := req.(type) {
		case *yasuo_pb.UpsertCoursesRequest:
			createdCourse := req.Courses[0]
			if int(createdCourse.SchoolId) == orgID {
				if err := s.canSeeCourseOnTeacherApp(ctx, createdCourse.Id); err != nil {
					return errors.Wrap(err, "canSeeCourseOnTeacherApp")
				}
			} else {
				err := s.cannotSeeCourseOnTeacherApp(ctx, createdCourse.Id)
				if err != nil {
					return errors.Wrap(err, "cannotSeeCourseOnTeacherApp")
				}
			}
		}
	}
	return nil
}

func (s *suite) schoolAdminSeeAllDataInOrg(ctx context.Context, orgID int) error {
	for _, req := range s.RequestStack.Requests {
		switch req := req.(type) {
		case *yasuo_pb.UpsertCoursesRequest:
			createdCourse := req.Courses[0]
			if int(createdCourse.SchoolId) == orgID {
				if err := s.canSeeCourseOnBackOffice(ctx, createdCourse.Id, false); err != nil {
					return errors.Wrap(err, "canSeeCourseOnBackOffice")
				}
			} else {
				if err := s.cannotSeeCourseOnBackOffice(ctx, createdCourse.Id, false); err != nil {
					return errors.Wrap(err, "cannotSeeCourseOnBackOffice")
				}
			}
		case *epb.UpsertBooksRequest:
			createdBook := req.Books[0]
			if err := canSeeBookOnBackOffice(ctx, createdBook.BookId, false); err != nil {
				return errors.Wrap(err, "canSeeBookOnBackOffice")
			}
		case *epb.UpsertTopicsRequest:
			createdTopic := req.Topics[0]
			if int(createdTopic.SchoolId) == orgID {
				if err := canSeeTopicOnBackOffice(ctx, createdTopic.Id, false); err != nil {
					return errors.Wrap(err, "canSeeBook")
				}
			} else {
				if err := cannotSeeTopicOnBackOffice(ctx, createdTopic.Id, false); err != nil {
					return errors.Wrap(err, "canSeeBook")
				}
			}
		case *epb.UpsertChaptersRequest:
			createdChapter := req.Chapters[0]
			if int(createdChapter.Info.SchoolId) == orgID {
				if err := canSeeChapterOnBackOffice(ctx, createdChapter.Info.Id, false); err != nil {
					return errors.Wrap(err, "canSeeChapterOnBackOffice")
				}
			} else {
				if err := cannotSeeChapterOnBackOffice(ctx, createdChapter.Info.Id, false); err != nil {
					return errors.Wrap(err, "cannotSeeChapterOnBackOffice")
				}
			}
		}
	}
	return nil
}

func (s *suite) canSeeCourseOnBackOffice(ctx context.Context, courseID string, queryWithAdminPermission bool) error {
	courses, err := queryCourses(ctx, courseID, queryWithAdminPermission)
	if err != nil {
		return errors.Wrap(err, "queryCourses")
	}

	if len(courses) < 1 {
		return fmt.Errorf("can't find course with id: %s", courseID)
	}

	queriedCourse := courses[0]
	if queriedCourse.CourseID != courseID {
		return fmt.Errorf(`expected course has id: "%s" but actual is "%s"`, courseID, queriedCourse.CourseID)
	}

	return nil
}

func (s *suite) cannotSeeCourseOnBackOffice(ctx context.Context, courseID string, queryWithAdminPermission bool) error {
	courses, err := queryCourses(ctx, courseID, queryWithAdminPermission)
	if err != nil {
		return errors.Wrap(err, "queryCourses")
	}

	if len(courses) > 0 {
		return fmt.Errorf("expect no courses")
	}

	return nil
}

func canSeeBookOnBackOffice(ctx context.Context, bookID string, queryWithAdminPermission bool) error {
	books, err := queryBooks(ctx, bookID, queryWithAdminPermission)
	if err != nil {
		return errors.Wrap(err, "queryBooks")
	}

	if len(books) < 1 {
		return fmt.Errorf("can't find book with id: %s", bookID)
	}

	queriedBook := books[0]
	if queriedBook.ID != bookID {
		return fmt.Errorf(`expected book has id: "%s" but actual is "%s"`, bookID, queriedBook.ID)
	}

	return nil
}

func cannotSeeBookOnBackOffice(ctx context.Context, bookID string, queryWithAdminPermission bool) error {
	books, err := queryBooks(ctx, bookID, queryWithAdminPermission)
	if err != nil {
		return errors.Wrap(err, "queryBooks")
	}

	if len(books) > 0 {
		return fmt.Errorf("expect no books")
	}

	return nil
}

func canSeeTopicOnBackOffice(ctx context.Context, topicID string, queryWithAdminPermission bool) error {
	topics, err := queryTopics(ctx, topicID, queryWithAdminPermission)
	if err != nil {
		return errors.Wrap(err, "queryTopics")
	}

	if len(topics) < 1 {
		return fmt.Errorf("can't find topic with id: %s", topicID)
	}

	queriedTopic := topics[0]

	if queriedTopic.ID != topicID {
		return fmt.Errorf(`expected topic has id: "%s" but actual is "%s"`, topicID, queriedTopic.ID)
	}

	return nil
}

func cannotSeeTopicOnBackOffice(ctx context.Context, topicID string, queryWithAdminPermission bool) error {
	topics, err := queryTopics(ctx, topicID, queryWithAdminPermission)
	if err != nil {
		return errors.Wrap(err, "queryTopics")
	}

	if len(topics) > 0 {
		return fmt.Errorf("expect no topics")
	}

	return nil
}

func canSeeChapterOnBackOffice(ctx context.Context, chapterID string, queryWithAdminPermission bool) error {
	chapters, err := queryChapter(ctx, chapterID, queryWithAdminPermission)
	if err != nil {
		return errors.Wrap(err, "queryChapter")
	}

	if len(chapters) < 1 {
		return fmt.Errorf("can't find chapter with id: %s", chapterID)
	}

	queriedChapter := chapters[0]

	if queriedChapter.ID != chapterID {
		return fmt.Errorf(`expected chapter has id: "%s" but actual is "%s"`, chapterID, queriedChapter.ID)
	}

	return nil
}

func cannotSeeChapterOnBackOffice(ctx context.Context, chapterID string, queryWithAdminPermission bool) error {
	chapters, err := queryChapter(ctx, chapterID, queryWithAdminPermission)
	if err != nil {
		return errors.Wrap(err, "queryChapter")
	}

	if len(chapters) > 0 {
		return fmt.Errorf("expect no chapters")
	}

	return nil
}

func (s *suite) canSeeCourseOnTeacherApp(ctx context.Context, courseID string) error {
	req := &bob_pbv1.ListCoursesRequest{
		Paging: &com_pbv1.Paging{
			Limit: 100,
		},
		Filter: &com_pbv1.CommonFilter{
			Ids: []string{courseID},
		},
	}
	resp, err := bob_pbv1.NewCourseReaderServiceClient(s.bobConn).ListCourses(ctx, req)
	if err != nil {
		return errors.Wrap(err, "ListCourses()")
	}

	if len(resp.Items) < 1 {
		return fmt.Errorf("can't find course with id: %s", courseID)
	}

	queriedCourse := resp.Items[0]

	if queriedCourse.Info.Id != courseID {
		return fmt.Errorf(`expected course has id: "%s" but actual is "%s"`, courseID, queriedCourse.Info.Id)
	}

	return nil
}

func (s *suite) cannotSeeCourseOnTeacherApp(ctx context.Context, courseID string) error {
	req := &bob_pbv1.ListCoursesRequest{
		Paging: &com_pbv1.Paging{
			Limit: 100,
		},
		Filter: &com_pbv1.CommonFilter{
			Ids: []string{courseID},
		},
	}
	resp, err := bob_pbv1.NewCourseReaderServiceClient(s.bobConn).ListCourses(ctx, req)
	if err != nil {
		return errors.Wrap(err, "ListCourses()")
	}

	if len(resp.Items) >= 1 {
		return errors.New("expected cannot see course, but actual can see")
	}

	return nil
}
