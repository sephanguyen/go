package eibanam

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/internal/bob/entities"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	yasuo_pb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Role string

const (
	RoleAdmin           Role = "admin"
	RoleSchoolAdmin     Role = "school admin"
	RoleParent          Role = "parent"
	RoleTeacher         Role = "teacher"
	RoleStudent         Role = "student"
	RoleUnauthenticated Role = "unauthenticated"
)

type Helper struct {
	connections
	config
	*gandalf.Config
}

type config struct {
	EnigmaSrvURL string
	JPREPKey     string
	// JPREPSignature           string
	// HasClassJPREPInActive    bool
	FirebaseAddr             string
	ApplicantID              string
	HasuraAdminUrl           string
	GoogleIdentityToolkitUrl string
}

//nolint:structcheck
type connections struct {
	BobConn       *grpc.ClientConn
	TomConn       *grpc.ClientConn
	YasuoConn     *grpc.ClientConn
	EurekaConn    *grpc.ClientConn
	FatimaConn    *grpc.ClientConn
	ShamirConn    *grpc.ClientConn
	UsermgmtConn  *grpc.ClientConn
	EntryExitConn *grpc.ClientConn
	bobDB         *pgxpool.Pool
	tomDB         *pgxpool.Pool
	eurekaDB      *pgxpool.Pool
	fatimaDB      *pgxpool.Pool
	zeusDB        *pgxpool.Pool
	bobDBTrace    *database.DBTrace
}

func NewHelper(ctx context.Context, c *gandalf.Config, appID, fakeFirebaseAddr, hasuraAdminUrl, googleIdentityToolkitUrl string) (*Helper, error) {
	h := Helper{
		config: config{
			EnigmaSrvURL:             "http://" + c.EnigmaSrvAddr,
			JPREPKey:                 c.JPREPSignatureSecret,
			FirebaseAddr:             fakeFirebaseAddr,
			ApplicantID:              appID,
			HasuraAdminUrl:           hasuraAdminUrl,
			GoogleIdentityToolkitUrl: googleIdentityToolkitUrl,
		},
		Config: c,
	}

	if err := c.ConnectGRPCInsecure(
		ctx,
		&h.BobConn,
		&h.TomConn,
		&h.YasuoConn,
		&h.EurekaConn,
		&h.FatimaConn,
		&h.ShamirConn,
		&h.UsermgmtConn,
		&h.EntryExitConn,
	); err != nil {
		return nil, err
	}
	c.ConnectDB(ctx, &h.bobDB, &h.tomDB, &h.eurekaDB, &h.fatimaDB, &h.zeusDB)
	db, dbcancel, err := database.NewPool(context.Background(), zap.NewNop(), c.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zap.NewNop().Error("dbcancel() failed", zap.Error(err))
		}
	}()
	h.bobDBTrace = &database.DBTrace{
		DB: db,
	}
	bob.SetFirebaseAddr(fakeFirebaseAddr)

	return &h, nil
}

func (h *Helper) Destructor() error {
	h.bobDB.Close()
	h.tomDB.Close()
	h.eurekaDB.Close()
	h.fatimaDB.Close()

	if err := multierr.Combine(
		h.BobConn.Close(),
		h.TomConn.Close(),
		h.YasuoConn.Close(),
		h.EurekaConn.Close(),
		h.FatimaConn.Close(),
		h.ShamirConn.Close(),
	); err != nil {
		return fmt.Errorf("could not close connection of helper: %v", err)
	}

	return nil
}

// SignedInAsAccount create a new user with user_group <=> role in db
// and generate token for that new user using firebase emulator with
// predefined template and shamir ExchangeToken api
func (h *Helper) SignedInAsAccount(schoolID int32, role Role) (*UserCredential, error) {
	id, userGroup, err := h.CreateUser(schoolID, role)
	if err != nil {
		return nil, err
	}

	authToken, err := GenerateExchangeToken(h.FirebaseAddr, id, userGroup, h.ApplicantID, schoolID, h.ShamirConn)
	if err != nil {
		return nil, err
	}

	return &UserCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: userGroup,
	}, nil
}

func (h *Helper) CreateUser(schoolID int32, role Role) (id, userGroup string, err error) {
	if schoolID == 0 {
		schoolID = 1
	}

	if role == RoleUnauthenticated || len(role) == 0 {
		return "", "", fmt.Errorf("could create use with role %s", role)
	}

	switch role {
	case RoleAdmin:
		userGroup = constant.UserGroupAdmin
	case RoleSchoolAdmin:
		userGroup = constant.UserGroupSchoolAdmin
	case RoleParent:
		userGroup = constant.UserGroupParent
	case RoleTeacher:
		userGroup = constant.UserGroupTeacher
	case RoleStudent:
		userGroup = constant.UserGroupStudent
	}

	id = idutil.ULIDNow()
	if err = h.newUser(schoolID, withID(id), withRole(userGroup)); err != nil {
		return "", "", err
	}

	return id, userGroup, nil
}

type userOption func(u *bob_entities.User)

func withID(id string) userOption {
	return func(u *bob_entities.User) {
		u.ID = database.Text(id)
	}
}

func withRole(group string) userOption {
	return func(u *bob_entities.User) {
		u.Group = database.Text(group)
	}
}

func (h *Helper) newUser(schoolID int32, opts ...userOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))

	num := rand.Int()
	u := &bob_entities.User{}
	database.AllNullEntity(u)
	if err := multierr.Combine(
		u.LastName.Set(fmt.Sprintf("valid-user-%d", num)),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
		u.Country.Set(bob_pb.COUNTRY_VN.String()),
		u.Group.Set(constant.UserGroupAdmin),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)),
	); err != nil {
		return err
	}

	for _, opt := range opts {
		opt(u)
	}

	err := h.createUserInDB(ctx, u, schoolID)
	if err != nil {
		return err
	}

	// create user group for new user
	if err = h.createUserGroupInDB(ctx, u.ID.String, u.Group.String); err != nil {
		return err
	}

	return nil
}

func (h *Helper) createUserInDB(ctx context.Context, user *bob_entities.User, schoolID int32) error {
	err := database.ExecInTx(ctx, h.bobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := bob_repo.UserRepo{}
		err := userRepo.Create(ctx, tx, user)
		if err != nil {
			return err
		}
		switch user.Group.String {
		case constant.UserGroupStudent:
			studentRepo := &bob_repo.StudentRepo{}
			now := time.Now()
			student := &bob_entities.Student{}
			database.AllNullEntity(student)
			if err = multierr.Combine(
				student.ID.Set(user.ID),
				student.CurrentGrade.Set(12),
				student.OnTrial.Set(true),
				student.TotalQuestionLimit.Set(10),
				student.SchoolID.Set(schoolID),
				student.CreatedAt.Set(now),
				student.UpdatedAt.Set(now),
				student.BillingDate.Set(now),
				student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
			); err != nil {
				return err
			}

			if err = studentRepo.CreateEn(ctx, h.bobDBTrace, student); err != nil {
				return err
			}
		case constant.UserGroupTeacher:
			teacherRepo := bob_repo.TeacherRepo{}
			t := &bob_entities.Teacher{}
			database.AllNullEntity(t)
			t.ID = user.ID
			if err = t.SchoolIDs.Set([]int32{schoolID}); err != nil {
				return err
			}

			if err = teacherRepo.CreateMultiple(ctx, tx, []*entities.Teacher{t}); err != nil {
				return err
			}
		case constant.UserGroupSchoolAdmin:
			schoolAdminRepo := bob_repo.SchoolAdminRepo{}
			schoolAdminAccount := &bob_entities.SchoolAdmin{}
			database.AllNullEntity(schoolAdminAccount)
			if err = multierr.Combine(
				schoolAdminAccount.SchoolAdminID.Set(user.ID.String),
				schoolAdminAccount.SchoolID.Set(schoolID),
			); err != nil {
				return err
			}

			if err = schoolAdminRepo.CreateMultiple(ctx, tx, []*entities.SchoolAdmin{schoolAdminAccount}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *Helper) createUserGroupInDB(ctx context.Context, userId, userGR string) error {
	uGroup := &bob_entities.UserGroup{}
	database.AllNullEntity(uGroup)

	err := multierr.Combine(
		uGroup.GroupID.Set(userGR),
		uGroup.UserID.Set(userId),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
	)
	if err != nil {
		return err
	}

	userGroupRepo := &bob_repo.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, h.bobDBTrace, uGroup)
	if err != nil {
		return fmt.Errorf("userGroupRepo.Upsert: %w %s", err, userGR)
	}

	return nil
}

type UserCredential struct {
	UserID    string
	AuthToken string
	UserGroup string
}

func (h *Helper) CreateACourseViaGRPC(authToken string, schoolID int32) (*yasuo_pb.UpsertCoursesRequest, error) {
	id := idutil.ULIDNow()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = ContextWithTokenForGrpcCall(ctx, authToken)

	req := &yasuo_pb.UpsertCoursesRequest{
		Courses: []*yasuo_pb.UpsertCoursesRequest_Course{},
	}
	course := h.newUpsertCourseReq(id, "course name for id: "+id, schoolID)
	req.Courses = append(req.Courses, course)
	_, err := yasuo_pb.NewCourseServiceClient(h.YasuoConn).UpsertCourses(ContextWithToken(ctx, authToken), req)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (h *Helper) newUpsertCourseReq(ID, name string, schoolID int32) *yasuo_pb.UpsertCoursesRequest_Course {
	r := &yasuo_pb.UpsertCoursesRequest_Course{
		Id:           ID,
		Name:         name,
		Country:      bob_pb.COUNTRY_VN,
		Subject:      bob_pb.SUBJECT_ENGLISH,
		Grade:        "G12",
		DisplayOrder: 1,
		ChapterIds:   nil,
		SchoolId:     schoolID,
		BookIds:      nil,
		Icon:         "link-icon",
	}
	return r
}

func (h *Helper) CreateSchool() (*bob_entities.School, error) {
	random := idutil.ULIDNow()
	sch := &bob_entities.School{
		Name:           database.Text(random),
		Country:        database.Text(constant.CountryVN),
		IsSystemSchool: database.Bool(false),
		CreatedAt:      database.Timestamptz(time.Now()),
		UpdatedAt:      database.Timestamptz(time.Now()),
		Point: pgtype.Point{
			P:      pgtype.Vec2{X: 0, Y: 0},
			Status: 2,
		},
	}

	city := &bob_entities.City{
		Name:         database.Text(random),
		Country:      database.Text(constant.CountryVN),
		CreatedAt:    database.Timestamptz(time.Now()),
		UpdatedAt:    database.Timestamptz(time.Now()),
		DisplayOrder: database.Int2(0),
	}

	district := &bob_entities.District{
		Name:    database.Text(random),
		Country: database.Text(constant.CountryVN),
		City:    city,
	}
	sch.City = city
	sch.District = district
	repo := &bob_repo.SchoolRepo{}
	err := repo.Import(context.Background(), h.bobDB, []*bob_entities.School{sch})
	if err != nil {
		return nil, err
	}

	// fake org id = school id
	orgID := database.Text(strconv.Itoa(int(sch.ID.Int)))
	resourcePath := orgID
	schoolText := database.Text(strconv.Itoa(int(sch.ID.Int)))

	// multi-tenant needs this
	_, err = h.bobDB.Exec(context.Background(), `INSERT INTO organizations(
	organization_id, tenant_id, name, resource_path)
	VALUES ($1, $2, $3, $4)`, orgID, schoolText, sch.Name, resourcePath)
	if err != nil {
		return nil, err
	}

	// Init auth info
	stmt := `
		INSERT INTO organization_auths
			(organization_id, auth_project_id, auth_tenant_id)
		VALUES
			($1, 'fake_aud', ''),
			($2, 'dev-manabie-online', ''),
			($2, 'dev-manabie-online', 'integration-test-1-909wx')
		ON CONFLICT 
			DO NOTHING
		;
		`
	_, err = h.bobDB.Exec(context.Background(), stmt, sch.ID.Int, sch.ID.Int)
	if err != nil {
		return nil, fmt.Errorf("cannot init auth info: %v", err)
	}

	return sch, nil
}

func (h *Helper) CreateCenterInDB(name string) (string, error) {
	repo := lesson_repo.MasterDataRepo{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id := idutil.ULIDNow()
	if _, err := repo.InsertCenter(ctx, h.bobDB, &domain.Location{
		LocationID: id,
		Name:       name,
	}); err != nil {
		return "", fmt.Errorf("could not CreateCenterInDB: %w", err)
	}

	return id, nil
}

func (h *Helper) generateMedia() *bob_pb.Media {
	return &bob_pb.Media{MediaId: "", Name: fmt.Sprintf("random-name-%s", idutil.ULIDNow()), Resource: idutil.ULIDNow(), CreatedAt: types.TimestampNow(), UpdatedAt: types.TimestampNow(), Comments: []*bob_pb.Comment{{Comment: "Comment-1", Duration: types.DurationProto(10 * time.Second)}, {Comment: "Comment-2", Duration: types.DurationProto(20 * time.Second)}}, Type: bob_pb.MEDIA_TYPE_VIDEO}
}

func (h *Helper) CreateMediaViaGRPC(authToken string, numberMedia int) ([]string, error) {
	mediaList := []*bob_pb.Media{}
	for i := 0; i < numberMedia; i++ {
		mediaList = append(mediaList, h.generateMedia())
	}
	req := &bob_pb.UpsertMediaRequest{
		Media: mediaList,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = ContextWithTokenForGrpcCall(ctx, authToken)
	res, err := bob_pb.NewClassClient(h.BobConn).UpsertMedia(ContextWithToken(ctx, authToken), req)
	if err != nil {
		return nil, err
	}

	return res.MediaIds, nil
}

func (h *Helper) GetLessonByID(ctx context.Context, lessonID string) (*lesson_repo.Lesson, error) {
	lesson := &lesson_repo.Lesson{}
	fields, values := lesson.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM lessons
		WHERE lesson_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	err := h.bobDBTrace.QueryRow(ctx, query, &lessonID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return lesson, nil
}

func (h *Helper) InsertStudentSubscription(studentIDWithCourseID ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queueFn := func(b *pgx.Batch, studentID, courseID string) {
		id := idutil.ULIDNow()
		query := `INSERT INTO lesson_student_subscriptions (student_subscription_id, subscription_id, student_id, course_id) VALUES ($1, $2, $3, $4)`
		b.Queue(query, id, id, studentID, courseID)
	}

	b := &pgx.Batch{}
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		queueFn(b, studentIDWithCourseID[i], studentIDWithCourseID[i+1])
	}
	result := h.bobDBTrace.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return nil
}
