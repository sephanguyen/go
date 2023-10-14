package services

import (
	"context"
	"encoding/json"

	"math"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/fatima/repositories"
	mocks_protobuf "github.com/manabie-com/backend/mock/fatima/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSubscriptionService_CreatePackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	packageRepo := &mock_repositories.MockPackageRepo{}

	s := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:          db,
			PackageRepo: packageRepo,
		},
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("empty name", func(t *testing.T) {
		resp, err := s.CreatePackage(ctx, &pb.CreatePackageRequest{Name: ""})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "name is required"), err)
	})

	t.Run("empty start_at, endAt and duration", func(t *testing.T) {
		resp, err := s.CreatePackage(ctx, &pb.CreatePackageRequest{Name: "packageName"})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "should have duration or (start_at, end_at)"), err)
	})

	t.Run("empty properties", func(t *testing.T) {
		resp, err := s.CreatePackage(ctx, &pb.CreatePackageRequest{
			Name:     "packageName",
			Duration: 7,
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "properties is required"), err)
	})

	t.Run("empty courseID in properties", func(t *testing.T) {
		resp, err := s.CreatePackage(ctx, &pb.CreatePackageRequest{
			Name:     "packageName",
			Duration: 7,
			Properties: &pb.PackageProperties{
				CanWatchVideo:      nil,
				CanViewStudyGuide:  nil,
				CanDoQuiz:          nil,
				LimitOnlineLession: 0,
			},
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "should have one of properties: can_do_quiz, can_watch_video, can_view_study_guide"), err)
	})

	t.Run("invalid start and end date", func(t *testing.T) {
		startAt := timestamppb.Now()
		endAt := timestamppb.Now()
		endAt.Seconds -= 86400

		resp, err := s.CreatePackage(ctx, &pb.CreatePackageRequest{
			Name:     "packageName",
			Duration: 7,
			Properties: &pb.PackageProperties{
				CanWatchVideo:      []string{"course_1", "course_2"},
				CanViewStudyGuide:  []string{"course_1", "course_2"},
				CanDoQuiz:          nil,
				LimitOnlineLession: 0,
			},
			StartAt: startAt,
			EndAt:   endAt,
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "startAt must be before endAt"), err)
	})

	t.Run("invalid EndAt", func(t *testing.T) {
		startAt := timestamppb.Now()
		endAt := timestamppb.Now()

		startAt.Seconds -= 86400 * 2
		endAt.Seconds -= 86400

		resp, err := s.CreatePackage(ctx, &pb.CreatePackageRequest{
			Name:     "packageName",
			Duration: 7,
			Properties: &pb.PackageProperties{
				CanWatchVideo:      []string{"course_1", "course_2"},
				CanViewStudyGuide:  []string{"course_1", "course_2"},
				CanDoQuiz:          nil,
				LimitOnlineLession: 0,
			},
			StartAt: startAt,
			EndAt:   endAt,
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "endAt must be in the future"), err)
	})

	t.Run("err upsert", func(t *testing.T) {
		startAt := timestamppb.Now()
		endAt := timestamppb.Now()

		endAt.Seconds += 86400

		packageRepo.On("Upsert", ctx, db, mock.AnythingOfType("*entities.Package")).Once().
			Return(puddle.ErrNotAvailable)

		resp, err := s.CreatePackage(ctx, &pb.CreatePackageRequest{
			Name:     "packageName",
			Duration: 7,
			Properties: &pb.PackageProperties{
				CanWatchVideo:      []string{"course_1", "course_2"},
				CanViewStudyGuide:  []string{"course_1", "course_2"},
				CanDoQuiz:          nil,
				LimitOnlineLession: 0,
				AskTutor:           &pb.PackageProperties_AskTutorCfg{},
			},
			StartAt: startAt,
			EndAt:   endAt,
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.Internal, puddle.ErrNotAvailable.Error()), err)
	})

	t.Run("success", func(t *testing.T) {
		startAt := timestamppb.Now()
		endAt := timestamppb.Now()

		endAt.Seconds += 86400

		packageRepo.On("Upsert", ctx, db, mock.AnythingOfType("*entities.Package")).Once().
			Run(func(args mock.Arguments) {
				p := args[2].(*entities.Package)

				assert.Equal(t, bob_pb.COUNTRY_VN.String(), p.Country.String)
				assert.Equal(t, "packageName", p.Name.String)
				assert.Equal(t, int32(7), p.Duration.Int)

				tStartAt := startAt.AsTime()
				assert.Equal(t, tStartAt, p.StartAt.Time)
				tEndAt := endAt.AsTime()
				assert.Equal(t, tEndAt, p.EndAt.Time)

				props, _ := p.GetProperties()
				assert.Equal(t, []string{"course_1", "course_2"}, props.CanWatchVideo)
				assert.Equal(t, []string{"course_1", "course_2"}, props.CanViewStudyGuide)
				assert.Equal(t, []string(nil), props.CanDoQuiz)
				assert.Equal(t, 1, props.LimitOnlineLesson)
				assert.Equal(t, 5, props.AskTutor.TotalQuestionLimit)
				assert.Equal(t, "ASK_DURATION_WEEK", props.AskTutor.LimitDuration)
			}).
			Return(nil)

		resp, err := s.CreatePackage(ctx, &pb.CreatePackageRequest{
			Name:     "packageName",
			Duration: 7,
			Country:  cpb.Country_COUNTRY_VN,
			Properties: &pb.PackageProperties{
				CanWatchVideo:      []string{"course_1", "course_2"},
				CanViewStudyGuide:  []string{"course_1", "course_2"},
				CanDoQuiz:          nil,
				LimitOnlineLession: 1,
				AskTutor: &pb.PackageProperties_AskTutorCfg{
					TotalQuestionLimit: 5,
					LimitDuration:      bpb.AskDuration_ASK_DURATION_WEEK,
				},
			},
			StartAt: startAt,
			EndAt:   endAt,
		})
		assert.Nil(t, err)
		assert.NotEmpty(t, resp.PackageId)
	})
}

func TestSubscriptionService_ToggleActivePackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	packageRepo := &mock_repositories.MockPackageRepo{}

	s := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:          db,
			PackageRepo: packageRepo,
		},
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("empty packageId", func(t *testing.T) {
		resp, err := s.ToggleActivePackage(ctx, &pb.ToggleActivePackageRequest{PackageId: ""})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "missing packageId"), err)
	})

	t.Run("not found", func(t *testing.T) {
		packageRepo.On("Get", ctx, db, database.Text("packageId")).Once().
			Return(nil, pgx.ErrNoRows)

		resp, err := s.ToggleActivePackage(ctx, &pb.ToggleActivePackageRequest{PackageId: "packageId"})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.NotFound, pgx.ErrNoRows.Error()), err)
	})

	t.Run("err upsert", func(t *testing.T) {
		p := &entities.Package{}
		packageRepo.On("Get", ctx, db, database.Text("packageId")).Once().
			Return(p, nil)

		p.IsActive.Bool = !p.IsActive.Bool
		packageRepo.On("Upsert", ctx, db, p).Once().
			Return(puddle.ErrClosedPool)

		resp, err := s.ToggleActivePackage(ctx, &pb.ToggleActivePackageRequest{PackageId: "packageId"})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.Internal, puddle.ErrClosedPool.Error()), err)
	})

	t.Run("success", func(t *testing.T) {
		p := &entities.Package{}
		packageRepo.On("Get", ctx, db, database.Text("packageId")).Once().
			Return(p, nil)

		p.IsActive.Bool = !p.IsActive.Bool
		packageRepo.On("Upsert", ctx, db, p).Once().
			Return(nil)

		resp, err := s.ToggleActivePackage(ctx, &pb.ToggleActivePackageRequest{PackageId: "packageId"})
		assert.NotNil(t, resp)
		assert.Nil(t, err)
	})
}

func TestSubscriptionService_AddStudentPackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	packageRepo := &mock_repositories.MockPackageRepo{}
	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
	jsm := new(mock_nats.JetStreamManagement)

	s := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:                 db,
			PackageRepo:        packageRepo,
			StudentPackageRepo: studentPackageRepo,
			JSM:                jsm,
		},
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("missing packageID", func(t *testing.T) {
		resp, err := s.AddStudentPackage(ctx, &pb.AddStudentPackageRequest{PackageId: ""})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "missing packageId"), err)
	})

	t.Run("missing studentId", func(t *testing.T) {
		resp, err := s.AddStudentPackage(ctx, &pb.AddStudentPackageRequest{PackageId: "packageId", StudentId: ""})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "missing studentId"), err)
	})

	t.Run("not found packageId", func(t *testing.T) {
		packageRepo.On("Get", ctx, db, database.Text("packageId")).Once().Return(nil, pgx.ErrNoRows)

		resp, err := s.AddStudentPackage(ctx, &pb.AddStudentPackageRequest{PackageId: "packageId", StudentId: "studentId"})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.NotFound, pgx.ErrNoRows.Error()), err)
	})

	t.Run("package is not active", func(t *testing.T) {
		p := &entities.Package{
			ID:       database.Text("packageId"),
			IsActive: pgtype.Bool{Bool: false, Status: pgtype.Present},
		}

		packageRepo.On("Get", ctx, db, database.Text("packageId")).Once().Return(p, nil)

		resp, err := s.AddStudentPackage(ctx, &pb.AddStudentPackageRequest{PackageId: "packageId", StudentId: "studentId"})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "package is disabled"), err)
	})

	t.Run("err insert", func(t *testing.T) {
		startAt := time.Now()
		endAt := startAt.Add(7 * 24 * time.Hour)

		props := &entities.PackageProperties{
			CanWatchVideo: []string{
				"course_1",
				"course_2",
			},
			CanViewStudyGuide: []string{
				"course_1",
				"course_2",
			},
			CanDoQuiz: []string{
				"course_1",
				"course_2",
			},
			LimitOnlineLesson: 0,
			AskTutor: &entities.AskTutorCfg{
				TotalQuestionLimit: 5,
				LimitDuration:      "ASK_DURATION_WEEK",
			},
		}

		p := &entities.Package{
			ID:       database.Text("packageId"),
			StartAt:  pgtype.Timestamptz{Time: startAt, Status: pgtype.Present},
			EndAt:    pgtype.Timestamptz{Time: endAt, Status: pgtype.Present},
			IsActive: pgtype.Bool{Bool: true, Status: pgtype.Present},
		}
		_ = p.Properties.Set(props)

		packageRepo.On("Get", ctx, db, database.Text("packageId")).Once().Return(p, nil)

		studentPackageRepo.On("Insert", ctx, db, mock.AnythingOfType("*entities.StudentPackage")).Once().
			Run(func(args mock.Arguments) {
				sp := args[2].(*entities.StudentPackage)
				assert.NotEmpty(t, sp.ID.String)
				assert.Equal(t, "studentId", sp.StudentID.String)
				assert.Equal(t, "packageId", sp.PackageID.String)
				assert.Equal(t, startAt, sp.StartAt.Time)
				assert.Equal(t, endAt, sp.EndAt.Time)
				assert.True(t, sp.IsActive.Bool)

				spProps, _ := sp.GetProperties()
				assert.Equal(t, props.CanWatchVideo, spProps.CanWatchVideo)
				assert.Equal(t, props.CanViewStudyGuide, spProps.CanViewStudyGuide)
				assert.Equal(t, props.CanDoQuiz, spProps.CanDoQuiz)
				assert.Equal(t, props.AskTutor, spProps.AskTutor)
			}).Return(puddle.ErrClosedPool)

		resp, err := s.AddStudentPackage(ctx, &pb.AddStudentPackageRequest{PackageId: "packageId", StudentId: "studentId"})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.Internal, puddle.ErrClosedPool.Error()), err)
	})

	t.Run("success", func(t *testing.T) {
		props := &entities.PackageProperties{
			CanWatchVideo: []string{
				"course_1",
				"course_2",
			},
			CanViewStudyGuide: []string{
				"course_1",
				"course_2",
			},
			CanDoQuiz: []string{
				"course_1",
				"course_2",
			},
			LimitOnlineLesson: 0,
			AskTutor: &entities.AskTutorCfg{
				TotalQuestionLimit: 5,
				LimitDuration:      "ASK_DURATION_WEEK",
			},
		}

		p := &entities.Package{
			ID:       database.Text("packageId"),
			Duration: database.Int4(7),
			IsActive: pgtype.Bool{Bool: true, Status: pgtype.Present},
		}
		_ = p.Properties.Set(props)

		packageRepo.On("Get", ctx, db, database.Text("packageId")).Once().Return(p, nil)

		studentPackageRepo.On("Insert", ctx, db, mock.AnythingOfType("*entities.StudentPackage")).Once().
			Run(func(args mock.Arguments) {
				now := time.Now()
				endAt := now.Add(time.Duration(p.Duration.Int) * 24 * time.Hour)

				sp := args[2].(*entities.StudentPackage)
				assert.NotEmpty(t, sp.ID.String)
				assert.Equal(t, "studentId", sp.StudentID.String)
				assert.Equal(t, "packageId", sp.PackageID.String)
				assert.Equal(t, float64(0), math.Ceil(sp.StartAt.Time.Sub(now).Seconds()))
				assert.Equal(t, float64(0), math.Ceil(sp.EndAt.Time.Sub(endAt).Seconds()))
				assert.True(t, sp.IsActive.Bool)

				spProps, _ := sp.GetProperties()
				assert.Equal(t, props.CanWatchVideo, spProps.CanWatchVideo)
				assert.Equal(t, props.CanViewStudyGuide, spProps.CanViewStudyGuide)
				assert.Equal(t, props.CanDoQuiz, spProps.CanDoQuiz)
				assert.Equal(t, props.AskTutor, spProps.AskTutor)
			}).Return(nil)
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)
		resp, err := s.AddStudentPackage(ctx, &pb.AddStudentPackageRequest{PackageId: "packageId", StudentId: "studentId"})
		assert.Nil(t, err)
		assert.NotEmpty(t, resp.StudentPackageId)
	})
}

func TestSubscriptionService_ToggleActiveStudentPackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
	jsm := new(mock_nats.JetStreamManagement)
	s := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:                 db,
			StudentPackageRepo: studentPackageRepo,
			JSM:                jsm,
		},
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("empty studentPackageId", func(t *testing.T) {
		resp, err := s.ToggleActiveStudentPackage(ctx, &pb.ToggleActiveStudentPackageRequest{StudentPackageId: ""})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "missing student packageId"), err)
	})

	t.Run("not found", func(t *testing.T) {
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(nil, pgx.ErrNoRows)

		resp, err := s.ToggleActiveStudentPackage(ctx, &pb.ToggleActiveStudentPackageRequest{StudentPackageId: "studentPackageId"})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.NotFound, pgx.ErrNoRows.Error()), err)
	})

	t.Run("err update", func(t *testing.T) {
		p := &entities.StudentPackage{}
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(p, nil)

		p.IsActive.Bool = !p.IsActive.Bool
		studentPackageRepo.On("Update", ctx, db, p).Once().
			Return(puddle.ErrClosedPool)

		resp, err := s.ToggleActiveStudentPackage(ctx, &pb.ToggleActiveStudentPackageRequest{StudentPackageId: "studentPackageId"})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.Internal, puddle.ErrClosedPool.Error()), err)
	})

	t.Run("success", func(t *testing.T) {
		p := &entities.StudentPackage{}
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(p, nil)

		p.IsActive.Bool = !p.IsActive.Bool
		studentPackageRepo.On("Update", ctx, db, p).Once().
			Return(nil)
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)

		resp, err := s.ToggleActiveStudentPackage(ctx, &pb.ToggleActiveStudentPackageRequest{StudentPackageId: "studentPackageId"})
		assert.NotNil(t, resp)
		assert.Nil(t, err)
	})
}

func TestStudentPackageService_CreateStudentPackageCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
	studentPackageAccessPathRepo := &mock_repositories.MockStudentPackageAccessPathRepo{}
	studentPackageClassRepo := &mock_repositories.MockStudentPackageClassRepo{}

	jsm := new(mock_nats.JetStreamManagement)
	s := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:                           db,
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			StudentPackageClassRepo:      studentPackageClassRepo,
			JSM:                          jsm,
		},
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	now := time.Now()
	startAt := timestamppb.Now()
	endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))

	t.Run("missing courseIDs", func(t *testing.T) {
		resp, err := s.AddStudentPackageCourse(ctx, &pb.AddStudentPackageCourseRequest{
			StudentId:   "123",
			CourseIds:   nil,
			StartAt:     timestamppb.Now(),
			EndAt:       timestamppb.Now(),
			LocationIds: []string{constants.ManabieOrgLocation},
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "missing courseId"), err)
	})

	t.Run("missing courseIDs", func(t *testing.T) {
		resp, err := s.AddStudentPackageCourse(ctx, &pb.AddStudentPackageCourseRequest{
			StudentId:   "123",
			CourseIds:   []string{},
			StartAt:     timestamppb.Now(),
			EndAt:       timestamppb.Now(),
			LocationIds: []string{constants.ManabieOrgLocation},
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "missing courseId"), err)
	})

	t.Run("missing studentId", func(t *testing.T) {
		resp, err := s.AddStudentPackageCourse(ctx, &pb.AddStudentPackageCourseRequest{
			StudentId:   "",
			CourseIds:   []string{"course_1", "course_2"},
			StartAt:     timestamppb.Now(),
			EndAt:       timestamppb.Now(),
			LocationIds: []string{constants.ManabieOrgLocation},
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "missing studentId"), err)
	})

	t.Run("invalid time", func(t *testing.T) {
		resp, err := s.AddStudentPackageCourse(ctx, &pb.AddStudentPackageCourseRequest{
			StudentId:   "123",
			CourseIds:   []string{"course_1", "course_2"},
			StartAt:     endAt,
			EndAt:       startAt,
			LocationIds: []string{constants.ManabieOrgLocation},
		})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "startAt must be before endAt"), err)
	})

	t.Run("success", func(t *testing.T) {
		props := &entities.PackageProperties{
			CanWatchVideo: []string{
				"course_1",
				"course_2",
			},
			CanViewStudyGuide: []string{
				"course_1",
				"course_2",
			},
			CanDoQuiz: []string{
				"course_1",
				"course_2",
			},
		}
		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)

		studentPackageRepo.On("Insert", ctx, tx, mock.AnythingOfType("*entities.StudentPackage")).Once().
			Run(func(args mock.Arguments) {
				sp := args[2].(*entities.StudentPackage)
				assert.NotEmpty(t, sp.ID.String)
				assert.Equal(t, "studentId", sp.StudentID.String)
				assert.Equal(t, "", sp.PackageID.String)
				assert.Equal(t, startAt.AsTime().UTC(), sp.StartAt.Time.UTC())
				assert.Equal(t, endAt.AsTime().UTC(), sp.EndAt.Time.UTC())
				assert.True(t, sp.IsActive.Bool)

				spProps, _ := sp.GetProperties()
				assert.Equal(t, props.CanWatchVideo, spProps.CanWatchVideo)
				assert.Equal(t, props.CanViewStudyGuide, spProps.CanViewStudyGuide)
				assert.Equal(t, props.CanDoQuiz, spProps.CanDoQuiz)
			}).Return(nil)

		studentPackageAccessPathRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)

		resp, err := s.AddStudentPackageCourse(ctx, &pb.AddStudentPackageCourseRequest{
			StudentId:   "studentId",
			CourseIds:   []string{"course_1", "course_2"},
			StartAt:     startAt,
			EndAt:       endAt,
			LocationIds: []string{constants.ManabieOrgLocation},
		})
		assert.Nil(t, err)
		assert.NotEmpty(t, resp.StudentPackageId)
	})

	t.Run("success with student package", func(t *testing.T) {
		props := &entities.PackageProperties{
			CanWatchVideo: []string{
				"course_1",
			},
			CanViewStudyGuide: []string{
				"course_1",
			},
			CanDoQuiz: []string{
				"course_1",
			},
		}
		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)

		studentPackageRepo.On("Insert", ctx, tx, mock.AnythingOfType("*entities.StudentPackage")).Once().
			Run(func(args mock.Arguments) {
				sp := args[2].(*entities.StudentPackage)
				assert.NotEmpty(t, sp.ID.String)
				assert.Equal(t, "studentId", sp.StudentID.String)
				assert.Equal(t, "", sp.PackageID.String)
				assert.Equal(t, startAt.AsTime().UTC(), sp.StartAt.Time.UTC())
				assert.Equal(t, endAt.AsTime().UTC(), sp.EndAt.Time.UTC())
				assert.True(t, sp.IsActive.Bool)

				spProps, _ := sp.GetProperties()
				assert.Equal(t, props.CanWatchVideo, spProps.CanWatchVideo)
				assert.Equal(t, props.CanViewStudyGuide, spProps.CanViewStudyGuide)
				assert.Equal(t, props.CanDoQuiz, spProps.CanDoQuiz)
			}).Return(nil)

		studentPackageClassRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageAccessPathRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageV2EventNats, mock.Anything).Return("", nil)

		resp, err := s.AddStudentPackageCourse(ctx, &pb.AddStudentPackageCourseRequest{
			StudentId: "studentId",
			CourseIds: []string{"course_1"},
			StartAt:   startAt,
			EndAt:     endAt,
			StudentPackageExtra: []*pb.AddStudentPackageCourseRequest_AddStudentPackageExtra{
				{
					CourseId:   "course_1",
					LocationId: "location_1",
					ClassId:    "class_1",
				},
			},
		})
		assert.Nil(t, err)
		assert.NotEmpty(t, resp.StudentPackageId)
	})
	t.Run("error bulk insert student package class", func(t *testing.T) {
		props := &entities.PackageProperties{
			CanWatchVideo: []string{
				"course_1",
			},
			CanViewStudyGuide: []string{
				"course_1",
			},
			CanDoQuiz: []string{
				"course_1",
			},
		}
		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		tx.On("Rollback", mock.Anything).Return(nil)

		studentPackageRepo.On("Insert", ctx, tx, mock.AnythingOfType("*entities.StudentPackage")).Once().
			Run(func(args mock.Arguments) {
				sp := args[2].(*entities.StudentPackage)
				assert.NotEmpty(t, sp.ID.String)
				assert.Equal(t, "studentId", sp.StudentID.String)
				assert.Equal(t, "", sp.PackageID.String)
				assert.Equal(t, startAt.AsTime().UTC(), sp.StartAt.Time.UTC())
				assert.Equal(t, endAt.AsTime().UTC(), sp.EndAt.Time.UTC())
				assert.True(t, sp.IsActive.Bool)

				spProps, _ := sp.GetProperties()
				assert.Equal(t, props.CanWatchVideo, spProps.CanWatchVideo)
				assert.Equal(t, props.CanViewStudyGuide, spProps.CanViewStudyGuide)
				assert.Equal(t, props.CanDoQuiz, spProps.CanDoQuiz)
			}).Return(nil)

		studentPackageClassRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
		studentPackageAccessPathRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)

		_, err := s.AddStudentPackageCourse(ctx, &pb.AddStudentPackageCourseRequest{
			StudentId: "studentId",
			StartAt:   startAt,
			EndAt:     endAt,
			StudentPackageExtra: []*pb.AddStudentPackageCourseRequest_AddStudentPackageExtra{
				{
					CourseId:   "course_1",
					LocationId: "location_1",
					ClassId:    "class_1",
				},
			},
		})
		assert.NotNil(t, err)
	})
}

func TestSubscriptionService_EditTimeStudentPackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
	studentPackageAccessPathRepo := &mock_repositories.MockStudentPackageAccessPathRepo{}
	studentPackageClassRepo := &mock_repositories.MockStudentPackageClassRepo{}

	jsm := new(mock_nats.JetStreamManagement)
	s := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:                           db,
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			StudentPackageClassRepo:      studentPackageClassRepo,
			JSM:                          jsm,
		},
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	now := time.Now()
	startAt := timestamppb.Now()
	endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))

	t.Run("empty studentPackageId", func(t *testing.T) {
		resp, err := s.EditTimeStudentPackage(ctx, &pb.EditTimeStudentPackageRequest{StudentPackageId: "", StartAt: startAt, EndAt: endAt})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "missing student packageId"), err)
	})

	t.Run("not found", func(t *testing.T) {
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(nil, pgx.ErrNoRows)

		resp, err := s.EditTimeStudentPackage(ctx, &pb.EditTimeStudentPackageRequest{StudentPackageId: "studentPackageId", StartAt: startAt, EndAt: endAt})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.NotFound, pgx.ErrNoRows.Error()), err)
	})

	t.Run("invalid time", func(t *testing.T) {
		resp, err := s.EditTimeStudentPackage(ctx, &pb.EditTimeStudentPackageRequest{StudentPackageId: "studentPackageId", StartAt: endAt, EndAt: startAt})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.InvalidArgument, "startAt must be before endAt"), err)
	})

	t.Run("err update", func(t *testing.T) {
		p := &entities.StudentPackage{}
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(p, nil)

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)

		studentPackageRepo.On("Update", ctx, tx, p).Once().
			Return(puddle.ErrClosedPool)

		resp, err := s.EditTimeStudentPackage(ctx, &pb.EditTimeStudentPackageRequest{StudentPackageId: "studentPackageId", StartAt: startAt, EndAt: endAt})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.Internal, puddle.ErrClosedPool.Error()), err)
	})

	t.Run("success", func(t *testing.T) {
		p := &entities.StudentPackage{}
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(p, nil)

		p.StartAt.Set(startAt)
		p.EndAt.Set(endAt)
		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)

		studentPackageRepo.On("Update", ctx, tx, p).Once().
			Return(nil)

		studentPackageAccessPathRepo.On("DeleteByStudentPackageIDs", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageAccessPathRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)

		resp, err := s.EditTimeStudentPackage(ctx, &pb.EditTimeStudentPackageRequest{StudentPackageId: "studentPackageId", StartAt: startAt, EndAt: endAt})
		assert.NotNil(t, resp)
		assert.Nil(t, err)
	})

	t.Run("success with student package", func(t *testing.T) {
		p := &entities.StudentPackage{}
		props := &entities.PackageProperties{
			CanWatchVideo: []string{
				"course_1",
			},
			CanViewStudyGuide: []string{
				"course_1",
			},
			CanDoQuiz: []string{
				"course_1",
			},
		}
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(p, nil)

		p.StudentID.Set("student_id")
		p.StartAt.Set(startAt)
		p.EndAt.Set(endAt)
		p.ID.Set("student_package_id")
		propsJson, err := json.Marshal(props)
		assert.Nil(t, err)
		p.Properties.Set(propsJson)
		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		studentPackageRepo.On("Update", ctx, tx, p).Once().
			Return(nil)

		studentPackageAccessPathRepo.On("DeleteByStudentPackageIDs", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageAccessPathRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageClassRepo.On("DeleteByStudentPackageIDs", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageClassRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageV2EventNats, mock.Anything).Return("", nil)

		resp, err := s.EditTimeStudentPackage(ctx, &pb.EditTimeStudentPackageRequest{
			StudentPackageId: "studentPackageId",
			StartAt:          startAt,
			EndAt:            endAt,
			StudentPackageExtra: []*pb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra{
				{
					LocationId: "location_1",
					ClassId:    "class_1",
				},
			},
		})
		assert.NotNil(t, resp)
		assert.Nil(t, err)
	})

	t.Run("error bulk insert student package class", func(t *testing.T) {
		p := &entities.StudentPackage{}
		props := &entities.PackageProperties{
			CanWatchVideo: []string{
				"course_1",
			},
			CanViewStudyGuide: []string{
				"course_1",
			},
			CanDoQuiz: []string{
				"course_1",
			},
		}
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(p, nil)

		p.StudentID.Set("student_id")
		p.StartAt.Set(startAt)
		p.EndAt.Set(endAt)
		p.ID.Set("student_package_id")
		propsJson, err := json.Marshal(props)
		assert.Nil(t, err)
		p.Properties.Set(propsJson)
		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		studentPackageRepo.On("Update", ctx, tx, p).Once().
			Return(nil)
		tx.On("Rollback", mock.Anything).Return(nil)

		studentPackageAccessPathRepo.On("DeleteByStudentPackageIDs", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageAccessPathRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageClassRepo.On("DeleteByStudentPackageIDs", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageClassRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)

		resp, err := s.EditTimeStudentPackage(ctx, &pb.EditTimeStudentPackageRequest{
			StudentPackageId: "studentPackageId",
			StartAt:          startAt,
			EndAt:            endAt,
			StudentPackageExtra: []*pb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra{
				{
					LocationId: "location_1",
					ClassId:    "class_1",
				},
			},
		})
		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error delete student package classes", func(t *testing.T) {
		p := &entities.StudentPackage{}
		props := &entities.PackageProperties{
			CanWatchVideo: []string{
				"course_1",
			},
			CanViewStudyGuide: []string{
				"course_1",
			},
			CanDoQuiz: []string{
				"course_1",
			},
		}
		studentPackageRepo.On("Get", ctx, db, database.Text("studentPackageId")).Once().
			Return(p, nil)

		p.StudentID.Set("student_id")
		p.StartAt.Set(startAt)
		p.EndAt.Set(endAt)
		p.ID.Set("student_package_id")
		propsJson, err := json.Marshal(props)
		assert.Nil(t, err)
		p.Properties.Set(propsJson)
		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		studentPackageRepo.On("Update", ctx, tx, p).Once().
			Return(nil)
		tx.On("Rollback", mock.Anything).Return(nil)

		studentPackageAccessPathRepo.On("DeleteByStudentPackageIDs", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageAccessPathRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)
		studentPackageClassRepo.On("DeleteByStudentPackageIDs", ctx, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
		studentPackageClassRepo.On("BulkUpsert", ctx, tx, mock.Anything).Once().Return(nil)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Return("", nil)

		resp, err := s.EditTimeStudentPackage(ctx, &pb.EditTimeStudentPackageRequest{
			StudentPackageId: "studentPackageId",
			StartAt:          startAt,
			EndAt:            endAt,
			StudentPackageExtra: []*pb.EditTimeStudentPackageRequest_EditTimeStudentPackageExtra{
				{
					LocationId: "location_1",
					ClassId:    "class_1",
				},
			},
		})
		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})
}

func TestSubscriptionModifyService_ListStudentPackage(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
	s := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:                 db,
			StudentPackageRepo: studentPackageRepo,
		},
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	now := time.Now()
	startAt := now
	endAt := now.Add(7 * 24 * time.Hour)

	testCases := []struct {
		name        string
		req         *pb.ListStudentPackageRequest
		expectedErr error
		setup       func(context.Context)
	}{
		{
			name: "happy case",
			req:  &pb.ListStudentPackageRequest{StudentIds: []string{"student_id_1", "student_id_2"}},
			setup: func(ctx context.Context) {
				prop := entities.PackageProperties{
					CanWatchVideo:     []string{"course_id_1"},
					CanViewStudyGuide: []string{"course_id_1"},
					CanDoQuiz:         []string{"course_id_1"},
					AskTutor: &entities.AskTutorCfg{
						TotalQuestionLimit: 1,
						LimitDuration:      bpb.AskDuration_ASK_DURATION_DAY.String(),
					},
				}
				studentPackagesResult := []*entities.StudentPackage{
					{
						ID:         database.Text(idutil.ULIDNow()),
						StudentID:  database.Text("student_id_1"),
						PackageID:  database.Text(idutil.ULIDNow()),
						StartAt:    database.Timestamptz(startAt),
						EndAt:      database.Timestamptz(endAt),
						Properties: database.JSONB(prop),
						IsActive:   database.Bool(true),
					},
					{
						ID:         database.Text(idutil.ULIDNow()),
						StudentID:  database.Text("student_id_2"),
						PackageID:  database.Text(idutil.ULIDNow()),
						StartAt:    database.Timestamptz(startAt),
						EndAt:      database.Timestamptz(endAt),
						Properties: database.JSONB(prop),
						IsActive:   database.Bool(true),
					},
				}
				studentPackageRepo.On("GetByStudentIDs", ctx, s.DB, database.TextArray([]string{"student_id_1", "student_id_2"})).Once().Return(studentPackagesResult, nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(ctx)
		_, err := s.ListStudentPackage(ctx, testCase.req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestSubscriptionModifyService_ListStudentPackageV2(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}

	s := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:                 db,
			StudentPackageRepo: studentPackageRepo,
		},
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	now := time.Now()
	startAt := now
	endAt := now.Add(7 * 24 * time.Hour)

	stream := &mocks_protobuf.SubscriptionModifierService_ListStudentPackageV2Server{}
	testCases := []struct {
		name        string
		req         *pb.ListStudentPackageV2Request
		expectedErr error
		setup       func(context.Context)
	}{
		{
			name: "happy case",
			req:  &pb.ListStudentPackageV2Request{StudentIds: []string{"student_id_1", "student_id_2"}},
			setup: func(ctx context.Context) {
				prop := entities.PackageProperties{
					CanWatchVideo:     []string{"course_id_1"},
					CanViewStudyGuide: []string{"course_id_1"},
					CanDoQuiz:         []string{"course_id_1"},
					AskTutor: &entities.AskTutorCfg{
						TotalQuestionLimit: 1,
						LimitDuration:      bpb.AskDuration_ASK_DURATION_DAY.String(),
					},
				}
				studentPackagesResult := []*entities.StudentPackage{
					{
						ID:         database.Text(idutil.ULIDNow()),
						StudentID:  database.Text("student_id_1"),
						PackageID:  database.Text(idutil.ULIDNow()),
						StartAt:    database.Timestamptz(startAt),
						EndAt:      database.Timestamptz(endAt),
						Properties: database.JSONB(prop),
						IsActive:   database.Bool(true),
					},
					{
						ID:         database.Text(idutil.ULIDNow()),
						StudentID:  database.Text("student_id_2"),
						PackageID:  database.Text(idutil.ULIDNow()),
						StartAt:    database.Timestamptz(startAt),
						EndAt:      database.Timestamptz(endAt),
						Properties: database.JSONB(prop),
						IsActive:   database.Bool(true),
					},
				}
				studentPackageRepo.On("GetByStudentIDs", mock.Anything, s.DB, database.TextArray([]string{"student_id_1", "student_id_2"})).Once().Return(studentPackagesResult, nil)

				stream.On("Send", mock.Anything).Once().Return(nil)
				stream.On("Send", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(ctx)
		err := s.ListStudentPackageV2(testCase.req, stream)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestSubscriptionModifyService_RegisterStudentClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	courseID := []string{"course_1"}

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	studentPackageClassRepo := &mock_repositories.MockStudentPackageClassRepo{}
	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
	jsm := new(mock_nats.JetStreamManagement)
	service := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:                      db,
			StudentPackageClassRepo: studentPackageClassRepo,
			StudentPackageRepo:      studentPackageRepo,
			JSM:                     jsm,
		},
	}

	studentPackage := &entities.StudentPackage{
		ID:          database.Text("student-package-id"),
		StudentID:   database.Text("student-id"),
		LocationIDs: database.TextArray([]string{"123456"}),
	}
	props := &entities.PackageProperties{
		CanWatchVideo:     courseID,
		CanViewStudyGuide: courseID,
		CanDoQuiz:         courseID,
	}
	propsJson, _ := json.Marshal(props)
	studentPackage.Properties.Set(propsJson)

	testCases := []TestCase{
		{
			name: "happy case",
			req: &pb.RegisterStudentClassRequest{
				ClassesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						ClassId:          "class-1",
						StartTime:        timestamppb.New(time.Now()),
						EndTime:          timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentPackageRepo.On("Get", ctx, db, database.Text("student-package-id")).Once().
					Return(studentPackage, nil)
				studentPackageClassRepo.On("DeleteByStudentPackageIDAndCourseID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studentPackageClassRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageV2EventNats, mock.Anything).Return("", nil)
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name: "error Get student package",
			req: &pb.RegisterStudentClassRequest{
				ClassesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						ClassId:          "class-1",
						StartTime:        timestamppb.New(time.Now()),
						EndTime:          timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
					},
				},
			},
			expectedErr: status.Error(codes.NotFound, "errors"),
			setup: func(ctx context.Context) {
				studentPackageRepo.On("Get", ctx, db, database.Text("student-package-id")).Once().
					Return(studentPackage, errors.New("errors"))
			},
		},
		{
			name: "error DeleteByStudentPackageIDs",
			req: &pb.RegisterStudentClassRequest{
				ClassesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						ClassId:          "class-1",
						StartTime:        timestamppb.New(time.Now()),
						EndTime:          timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "errors"),
			setup: func(ctx context.Context) {
				studentPackageRepo.On("Get", ctx, db, database.Text("student-package-id")).Once().
					Return(studentPackage, nil)
				studentPackageClassRepo.On("DeleteByStudentPackageIDAndCourseID", ctx, tx, mock.Anything, mock.Anything).Once().Return(errors.New("errors"))
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			name: "error BulkUpsert",
			req: &pb.RegisterStudentClassRequest{
				ClassesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						ClassId:          "class-1",
						StartTime:        timestamppb.New(time.Now()),
						EndTime:          timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "errors"),
			setup: func(ctx context.Context) {
				studentPackageRepo.On("Get", ctx, db, database.Text("student-package-id")).Once().
					Return(studentPackage, nil)
				studentPackageClassRepo.On("DeleteByStudentPackageIDAndCourseID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studentPackageClassRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("errors"))
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(ctx)
		_, err := service.RegisterStudentClass(ctx, testCase.req.(*pb.RegisterStudentClassRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func Test_validateRegisterStudentClassRequest(t *testing.T) {
	t.Parallel()
	testCases := []TestCase{
		{
			name: "happy case",
			req: &pb.RegisterStudentClassRequest{
				ClassesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						ClassId:          "class-1",
						StartTime:        timestamppb.New(time.Now()),
						EndTime:          timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "missing student_id",
			req: &pb.RegisterStudentClassRequest{
				ClassesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentPackageId: "student-package-id",
						ClassId:          "class-1",
						StartTime:        timestamppb.New(time.Now()),
						EndTime:          timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "student_id cannot be empty"),
		},
		{
			name: "missing student_package_id",
			req: &pb.RegisterStudentClassRequest{
				ClassesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentId: "student-id",
						ClassId:   "class-1",
						StartTime: timestamppb.New(time.Now()),
						EndTime:   timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "student_package_id cannot be empty"),
		},
		{
			name: "missing class_id",
			req: &pb.RegisterStudentClassRequest{
				ClassesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						StartTime:        timestamppb.New(time.Now()),
						EndTime:          timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "class_id cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := validateRegisterStudentClassRequest(testCase.req.(*pb.RegisterStudentClassRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func Test_generateListStudentPackageClassesFromStudentPackageAndClassInformation(t *testing.T) {
	t.Parallel()
	courseID := []string{"course_1"}
	studentPackage := &entities.StudentPackage{
		ID:          database.Text("student-package-id"),
		StudentID:   database.Text("student-id"),
		LocationIDs: database.TextArray([]string{"123456"}),
	}
	props := &entities.PackageProperties{
		CanWatchVideo:     courseID,
		CanViewStudyGuide: courseID,
		CanDoQuiz:         courseID,
	}
	propsJson, _ := json.Marshal(props)
	studentPackage.Properties.Set(propsJson)
	type Input struct {
		studentPackage     *entities.StudentPackage
		classesInformation []*pb.RegisterStudentClassRequest_ClassInformation
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: Input{
				studentPackage: studentPackage,
				classesInformation: []*pb.RegisterStudentClassRequest_ClassInformation{
					{
						StudentId:        studentPackage.StudentID.String,
						StudentPackageId: studentPackage.ID.String,
						ClassId:          "class-id",
					},
				},
			},
			expectedErr: nil,
		},
	}
	for _, testCase := range testCases {
		_, err := generateListStudentPackageClassesFromStudentPackageAndClassInformation(testCase.req.(Input).studentPackage, testCase.req.(Input).classesInformation)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestSubscriptionModifyService_WrapperRegisterStudentClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	courseID := []string{"course_1"}
	now := time.Now()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	studentPackageClassRepo := &mock_repositories.MockStudentPackageClassRepo{}
	studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
	jsm := new(mock_nats.JetStreamManagement)
	service := &SubscriptionServiceABAC{
		SubscriptionModifyService: &SubscriptionModifyService{
			DB:                      db,
			StudentPackageClassRepo: studentPackageClassRepo,
			StudentPackageRepo:      studentPackageRepo,
			JSM:                     jsm,
		},
	}

	studentPackage := &entities.StudentPackage{
		ID:          database.Text("student-package-id"),
		StudentID:   database.Text("student-id"),
		LocationIDs: database.TextArray([]string{"123456"}),
		StartAt:     database.Timestamptz(now.Add(-10 * 24 * time.Hour)),
		EndAt:       database.Timestamptz(now.Add(100 * 24 * time.Hour)),
	}
	props := &entities.PackageProperties{
		CanWatchVideo:     courseID,
		CanViewStudyGuide: courseID,
		CanDoQuiz:         courseID,
	}
	propsJson, _ := json.Marshal(props)
	studentPackage.Properties.Set(propsJson)

	testCases := []TestCase{
		{
			name: "happy case",
			req: &pb.WrapperRegisterStudentClassRequest{
				ReserveClassesInformation: []*pb.WrapperRegisterStudentClassRequest_ReserveClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						CourseId:         courseID[0],
						ClassId:          "class-1",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentPackageRepo.On("GetByStudentPackageIDAndStudentIDAndCourseID", ctx, db, database.Text("student-package-id"), database.Text("student-id"), database.Text(courseID[0])).
					Once().Return(studentPackage, nil)
				studentPackageRepo.On("Get", ctx, db, database.Text("student-package-id")).Once().
					Return(studentPackage, nil)
				studentPackageClassRepo.On("DeleteByStudentPackageIDAndCourseID", ctx, tx, mock.Anything, mock.Anything).Once().Return(nil)
				studentPackageClassRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageV2EventNats, mock.Anything).Return("", nil)
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			name: "GetByStudentPackageIDAndStudentIDAndCourseID fail",
			req: &pb.WrapperRegisterStudentClassRequest{
				ReserveClassesInformation: []*pb.WrapperRegisterStudentClassRequest_ReserveClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						CourseId:         courseID[0],
						ClassId:          "class-1",
					},
				},
			},
			expectedErr: status.Error(codes.NotFound, "GetByStudentPackageIDAndStudentIDAndCourseID not found: errors"),
			setup: func(ctx context.Context) {
				studentPackageRepo.On("GetByStudentPackageIDAndStudentIDAndCourseID", ctx, db, database.Text("student-package-id"), database.Text("student-id"), database.Text(courseID[0])).
					Once().Return(nil, errors.New("errors"))
			},
		},
		{
			name: "RegisterStudentClass fail",
			req: &pb.WrapperRegisterStudentClassRequest{
				ReserveClassesInformation: []*pb.WrapperRegisterStudentClassRequest_ReserveClassInformation{
					{
						StudentId:        "student-id",
						StudentPackageId: "student-package-id",
						CourseId:         courseID[0],
						ClassId:          "class-1",
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "errors"),
			setup: func(ctx context.Context) {
				studentPackageRepo.On("GetByStudentPackageIDAndStudentIDAndCourseID", ctx, db, database.Text("student-package-id"), database.Text("student-id"), database.Text(courseID[0])).
					Once().Return(studentPackage, nil)
				studentPackageRepo.On("Get", ctx, db, database.Text("student-package-id")).Once().
					Return(studentPackage, nil)
				studentPackageClassRepo.On("DeleteByStudentPackageIDAndCourseID", ctx, tx, mock.Anything, mock.Anything).Once().Return(errors.New("errors"))
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(ctx)
		_, err := service.WrapperRegisterStudentClass(ctx, testCase.req.(*pb.WrapperRegisterStudentClassRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
