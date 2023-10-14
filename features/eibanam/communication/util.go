package communication

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	bentities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

var (
	insertSchoolAdmin      = `INSERT INTO users (%s) VALUES (%s);`
	insertSchoolAdminGroup = `INSERT INTO users_groups (%s) VALUES (%s);`
)

func (s *suite) insertNewAdmin(ctx context.Context, id string, email string) error {
	u := &bentities.User{
		ID:        database.Text(id),
		Group:     database.Text(constant.UserGroupAdmin),
		Email:     database.Text(email),
		Country:   database.Text(constant.CountryVN),
		LastName:  database.Text(id),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
	}

	fields := []string{"user_id", "email", "user_group", "country", "name", "updated_at", "created_at"}
	placeHolder := database.GeneratePlaceholders(len(fields))

	insertStatement := fmt.Sprintf(insertSchoolAdmin, strings.Join(fields, ","), placeHolder)
	_, err := s.bobDB.Exec(ctx, insertStatement, database.GetScanFields(u, fields)...)
	if err != nil {
		return err
	}
	ug := &bentities.UserGroup{
		UserID:    database.Text(id),
		GroupID:   database.Text(constant.UserGroupAdmin),
		Status:    database.Text(bentities.UserGroupStatusActive),
		IsOrigin:  database.Bool(true),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
	}

	fields = []string{"user_id", "group_id", "status", "is_origin", "updated_at", "created_at"}
	placeHolder = database.GeneratePlaceholders(len(fields))
	insertStatement = fmt.Sprintf(insertSchoolAdminGroup, strings.Join(fields, ","), placeHolder)
	_, err = s.bobDB.Exec(ctx, insertStatement, database.GetScanFields(ug, fields)...)
	return err
}

var (
	oneOfRegex         = regexp.MustCompile(`^1 of \[([^\]]*)\]`)
	parentOfRegex      = regexp.MustCompile(`(.*)'s parent`)
	userWithEmailRegex = regexp.MustCompile(`(.*)'s email`)
	studentNameRegex   = regexp.MustCompile(`$(.*) name`)
	userWithGradeRegex = regexp.MustCompile(`(.*)'s grade`)
)

func parseOneOf(oneOfString string) []string {
	matches := oneOfRegex.FindStringSubmatch(oneOfString)
	if matches == nil {
		return nil
	}
	return strings.Split(matches[1], ",")
}

func selectOneOf(oneOfString string) string {
	choices := parseOneOf(oneOfString)
	randIdx := rand.Intn(len(choices))
	return strings.TrimSpace(choices[randIdx])
}

func randomNameFromSamples() string {
	return sampleNames[rand.Intn(len(sampleNames))]
}

// key-value is fullname-partialname
var sampleNames []string

func init() {
	seedLanguagesData()
}

func readIntoSlice(f io.Reader) ([]string, error) {
	sl := []string{}
	reader := bufio.NewReader(f)
	line := 0
	for {
		bs, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return sl, nil
			}
			return nil, fmt.Errorf("failed to read file at line %d: %s", line, err)
		}
		if len(bs) == 0 {
			return nil, fmt.Errorf("unexpected blank line in seed data file at line %d", line)
		}
		sl = append(sl, string(bs))
		line++
	}
}

func seedLanguagesData() {
	workDir, err := os.Getwd()
	zapLogger := logger.NewZapLogger("info", true) // log level doesn't matter since only Panic is called here
	if err != nil {
		zapLogger.Panic("failed to get current working directory")
	}
	filenames := []string{"hiragana", "kanji", "katakana", "english"}
	for _, languageform := range filenames {
		file, err := os.Open(filepath.Join(workDir, "eibanam", "communication", "samples", languageform+".txt"))
		if err != nil {
			zapLogger.Panic(fmt.Sprintf("failed to open %s: %s", languageform, err))
		}
		defer file.Close()
		slc, err := readIntoSlice(file)
		if err != nil {
			zapLogger.Panic(fmt.Sprintf("cannot read file %s data: %s", languageform, err))
		}
		sampleNames = append(sampleNames, slc...)
	}
}

func (s *stepState) getID(target string) string {
	return s.getProfile(target).id
}

func (s *stepState) getProfile(target string) profile {
	prof := profile{}
	switch target {
	case student:
		prof = s.profile.defaultStudent
	case "newly created student":
		prof = s.profile.newlyCreatedStudent
	case parent:
		prof = s.profile.defaultParent
	case teacher:
		prof = s.profile.defaultTeacher
	case schoolAdmin:
		prof = s.profile.schoolAdmin
	case "admin":
		prof = s.profile.admin
	case newParent:
		prof = s.profile.newParent
	case existingParent:
		prof = s.profile.anExistingParent
	case "parent P1":
		prof = s.profile.defaultParent
	default:
		panic(fmt.Sprintf("unsupported target %s", target))
	}
	return prof
}

func (s *stepState) getUserGroup(target string) string {
	if strings.Contains(target, "student") {
		return cpb.UserGroup_USER_GROUP_STUDENT.String()
	}
	if strings.Contains(target, "parent") {
		return cpb.UserGroup_USER_GROUP_PARENT.String()
	}

	panic(fmt.Sprintf("cannot get user group for target %s", target))
}

func (s *stepState) getMail(target string) string {
	prof := s.getProfile(target)
	if prof.email == "" {
		panic(fmt.Sprintf("missing logic assigning email to target %s in previous steps, check again", target))
	}
	return prof.email
}

func (s *stepState) getToken(target string) string {
	cred := s.credentials[s.getMail(target)]
	return cred.token
}

func (s *stepState) getPassword(target string) string {
	cred := s.credentials[s.getMail(target)]
	return cred.password
}

func (s *stepState) setPassword(target string, password string) {
	mail := s.getMail(target)
	s.credentials[mail] = credential{password: password}
}

func (s *stepState) updateToken(target string, token string) {
	mail := s.getMail(target)
	cred := s.credentials[mail]
	cred.token = token
	s.credentials[mail] = cred
}

func (s *stepState) getParentMails() []string {
	ret := []string{}
	for _, m := range s.profile.multipleParents {
		ret = append(ret, m.email)
	}
	return ret
}

func (s *stepState) getParentTokens() map[string]string {
	ret := map[string]string{}
	for _, m := range s.profile.multipleParents {
		ret[m.email] = s.credentials[m.email].token
	}
	return ret
}

func contextWithResourcePath(ctx context.Context, rp string) context.Context {
	claim := interceptors.JWTClaimsFromContext(ctx)
	if claim == nil {
		claim = &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{},
		}
	}
	claim.Manabie.ResourcePath = rp
	return interceptors.ContextWithJWTClaims(ctx, claim)
}
