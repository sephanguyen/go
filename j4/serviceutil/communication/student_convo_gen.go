package serviceutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	"github.com/manabie-com/backend/j4/serviceutil/usermgmt"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

// TODO: make this pool more generic
type StudentConvoPool struct {
	genPerBatch    int
	j4Cfg          *infras.ManabieJ4Config
	bobDB          database.Ext
	tomDB          database.Ext
	tokenGenerator *serviceutil.TokenGenerator
	logger         *zap.SugaredLogger
	userSvc        usermgmt.GrpClient

	studentPoolMu *sync.Mutex
	students      []StudentConvo
}

func InitStudentConvoPool(ctx context.Context, c *infras.ManabieJ4Config, conns *infras.Connections) *StudentConvoPool {
	ctx = golibs.ResourcePathToCtx(ctx, c.SchoolID)

	tokenGenerator := serviceutil.NewTokenGenerator(c, conns)

	s := &StudentConvoPool{
		genPerBatch:    50,
		j4Cfg:          c,
		tokenGenerator: tokenGenerator,
		logger:         logger.NewZapLogger("debug", false).Sugar(),
		tomDB:          conns.DBConnPools["tom"],
		bobDB:          conns.DBConnPools["bob"],
		studentPoolMu:  &sync.Mutex{},
		students:       []StudentConvo{},
	}

	commonGrpcConn, err := conns.PoolToGateWay.Get(ctx)
	if err != nil {
		panic(err)
	}

	// dedicated connection to do some supporting stuff
	// s.bobSvc = bpb.NewUserModifierServiceClient(commonGrpcConn)
	s.userSvc = upb.NewStudentServiceClient(commonGrpcConn)
	// s.userSvc = upb.NewUserModifierServiceClient(commonGrpcConn)

	go s.checkStudentPool(ctx)
	return s
}

type StudentConvo struct {
	ConvID string
	UserID string
}

func (s *StudentConvoPool) checkStudentPool(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	// check available students in db
	var currOffset string
stillAvailableInDB:
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.studentPoolMu.Lock()
			if len(s.students) < 10 {
				s.studentPoolMu.Unlock()
				newStu, newOffset, stillAvailable := s.findAvailableStudentConvos(ctx, currOffset, s.genPerBatch)
				s.studentPoolMu.Lock()
				s.students = append(s.students, newStu...)
				currOffset = newOffset
				fmt.Printf("==============foundStudent: %d\n", len(newStu))
				// some workers are sleeping to wait for this slice to be filled up
				s.studentPoolMu.Unlock()
				if !stillAvailable {
					break stillAvailableInDB
				}
				continue
			}
			s.studentPoolMu.Unlock()
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.studentPoolMu.Lock()
			if len(s.students) < 10 {
				s.studentPoolMu.Unlock()
				newStu := s.genStudentConvos(ctx, s.genPerBatch)
				s.studentPoolMu.Lock()
				s.students = append(s.students, newStu...)
				fmt.Printf("==============genstudent: %d\n", len(newStu))
				// some workers are sleeping to wait for this slice to be filled up
				s.studentPoolMu.Unlock()
				continue
			}
			s.studentPoolMu.Unlock()
		}
	}
}

func (s *StudentConvoPool) findAvailableStudentConvos(ctx context.Context, offset string, total int) ([]StudentConvo, string, bool) {
	offsetText := pgtype.Text{Status: pgtype.Null}
	if offset != "" {
		offsetText.Set(offset)
	}
	rows, err := s.tomDB.Query(ctx, `select student_id, conversation_id from
	conversation_students where ($1::text is null or student_id > $1) and deleted_at is null 
	and conversation_type='CONVERSATION_STUDENT'
	order by student_id asc limit $2`, offsetText, total)
	if err != nil {
		s.logger.Fatalf("failed to find available student", err)
	}
	defer rows.Close()
	students := []StudentConvo{}
	stuIDs := []string{}
	for rows.Next() {
		var id, convID string
		err = rows.Scan(&id, &convID)
		if err != nil {
			panic(err)
		}
		students = append(students, StudentConvo{ConvID: convID, UserID: id})
		stuIDs = append(stuIDs, id)
	}
	if len(stuIDs) == 0 {
		return nil, "", false
	}
	newOffset := stuIDs[len(stuIDs)-1]

	userMap := map[string]struct{}{}

	rows2, err := s.bobDB.Query(ctx, "select user_id from users where user_id=any($1)",
		database.TextArray(stuIDs))
	if err != nil {
		s.logger.Fatalf("failed to find available student", err)
	}

	defer rows2.Close()
	for rows2.Next() {
		var id string
		err := rows2.Scan(&id)
		if err != nil {
			panic(err)
		}

		userMap[id] = struct{}{}
	}

	// we only return found entities
	acceptedStu := []StudentConvo{}
	for idx, stu := range students {
		if _, exist := userMap[stu.UserID]; exist {
			acceptedStu = append(acceptedStu, students[idx])
		}
	}

	return acceptedStu, newOffset, true
}

func (s *StudentConvoPool) genStudentConvos(ctx context.Context, total int) []StudentConvo {
	tok, err := s.tokenGenerator.GetTokenFromShamir(ctx, s.j4Cfg.AdminID, s.j4Cfg.SchoolID)
	if err != nil {
		panic(fmt.Sprintf("todo %s", err))
	}
	stus, err := s.CreateStudentConvos(contextWithToken(ctx, tok), total, s.j4Cfg.SchoolID)
	if err != nil {
		panic(fmt.Sprintf("todo %s", err))
	}
	stuIDs := sliceutils.Map(stus, func(s StudentConvo) string { return s.UserID })

	userConvMap := map[string]string{}

	try.DoBackOff(func(_ int) (bool, error) {
		rows, err := s.tomDB.Query(ctx, "select student_id, conversation_id from conversation_students where student_id=any($1)",
			database.TextArray(stuIDs))
		if err != nil {
			return true, fmt.Errorf("querying conversations for users %s", err)
		}

		defer rows.Close()
		total := 0
		for rows.Next() {
			total++
			var stuID, convID string
			err := rows.Scan(&stuID, &convID)
			if err != nil {
				panic(err)
			}
			userConvMap[stuID] = convID
		}
		if total != len(stuIDs) {
			return true, fmt.Errorf("not enough conv created, want %d has %d", total, len(stuIDs))
		}
		return false, nil
	}, 6*time.Second)

	// we only return found entities
	acceptedStu := []StudentConvo{}
	for idx, stu := range stus {
		if convID, exist := userConvMap[stu.UserID]; exist {
			stus[idx].ConvID = convID
			acceptedStu = append(acceptedStu, stus[idx])
		}
	}
	return acceptedStu
}

// CreateStudentConvos ctx must already have token, check TokenGenerator
func (s *StudentConvoPool) CreateStudentConvos(ctx context.Context, num int, schoolID string) ([]StudentConvo, error) {
	students, err := usermgmt.CreateStudents(ctx, num, schoolID, s.bobDB, s.userSvc)
	if err != nil {
		return nil, err
	}
	convos := make([]StudentConvo, 0, len(students))
	for _, stu := range students {
		convos = append(convos, StudentConvo{UserID: stu.UserID})
	}

	return convos, nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}

// GetOne this function may block for a while if students are not available
func (s *StudentConvoPool) GetOne(ctx context.Context) *StudentConvo {
try:
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		s.studentPoolMu.Lock()
		if len(s.students) == 0 {
			s.studentPoolMu.Unlock()
			time.Sleep(3 * time.Second)
			continue try
		}

		stu := s.students[len(s.students)-1]
		s.students = s.students[:len(s.students)-1]
		s.studentPoolMu.Unlock()
		return &stu
	}
}
