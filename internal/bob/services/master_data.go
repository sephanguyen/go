package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterDataService struct {
	Cfg      *configurations.Config
	DB       database.QueryExecer
	UserRepo interface {
		UserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error)
	}
	PresetStudyPlanRepo interface {
		BulkImport(ctx context.Context, db database.QueryExecer, presetStudyPlans []*entities.PresetStudyPlan, presetStudyPlanWeeklies []*entities.PresetStudyPlanWeekly) error
	}
	CourseService interface {
		UpsertLOs(ctx context.Context, req *pb.UpsertLOsRequest) (*pb.UpsertLOsResponse, error)
	}
	TopicRepo interface {
		BulkImport(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error
	}
	versions        map[string]string
	ghnProvinceData map[string]map[int32]*pb.LocationEntity
	provinces       []string
}

func (rcv *MasterDataService) GetClientVersion(ctx context.Context, req *pb.GetClientVersionRequest) (*pb.GetClientVersionResponse, error) {
	if rcv.versions == nil {
		rcv.versions = make(map[string]string)
		versions := strings.Split("com.manabie.student_manabie_app:1.1.0,com.manabie.studentManabieApp:1.1.0,com.manabie.liz:1.0.0", ",")
		for _, ver := range versions {
			parts := strings.Split(ver, ":")
			if len(parts) != 2 {
				return nil, errors.New("invalid version, must match patter <pkg_name>:<required_version>")
			}
			rcv.versions[parts[0]] = parts[1]
		}
	}

	return &pb.GetClientVersionResponse{
		Versions: rcv.versions,
	}, nil
}

func (rcv *MasterDataService) ImportPresetStudyPlan(sv pb.MasterDataService_ImportPresetStudyPlanServer) error {
	if err := rcv.checkRoleAdmin(sv.Context()); err != nil {
		return err
	}

	var buffer bytes.Buffer
	for {
		req, err := sv.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrapf(err, "failed unexpectedly while reading chunks from stream")
		}

		if int64(len(req.Payload)) > rcv.Cfg.Upload.MaxChunkSize {
			return status.Error(codes.InvalidArgument, "chunk size over limited")
		}

		buffer.Write(req.Payload)
		if int64(buffer.Len()) > rcv.Cfg.Upload.MaxFileSize {
			return status.Error(codes.InvalidArgument, "file size over limited")
		}
	}
	if err := rcv.importTablePresetStudyPlan(sv.Context(), buffer.Bytes()); err != nil {
		return err
	}

	return sv.SendAndClose(&pb.ImportPresetStudyPlanResponse{})
}

func (rcv *MasterDataService) ImportLO(sv pb.MasterDataService_ImportLOServer) error {
	return nil
}

func (rcv *MasterDataService) ImportTopic(sv pb.MasterDataService_ImportTopicServer) error {
	return nil
}

func (rcv *MasterDataService) importTablePresetStudyPlan(ctx context.Context, data []byte) error {
	r := csv.NewReader(bytes.NewReader(data))
	lines, err := r.ReadAll()
	if err != nil {
		return err
	}

	if len(lines) <= 3 {
		return status.Error(codes.InvalidArgument, "no data in csv file")
	}

	var (
		mapPresetStudyPlanWeeklies = make(map[string]*entities.PresetStudyPlanWeekly)
		presetStudyPlan            = new(entities.PresetStudyPlan)
		presetStudyPlanWeeklies    []*entities.PresetStudyPlanWeekly
		eCountry                   pb.Country
		eSubject                   pb.Subject
	)

	presetStudyPlanID := lines[0][1]
	if presetStudyPlanID == "" {
		return status.Error(codes.InvalidArgument, "missing presetStudyPlanId at B1")
	}

	presetStudyPlanName := lines[0][2]
	if presetStudyPlanName == "" {
		return status.Error(codes.InvalidArgument, "missing presetStudyPlanName at C1")
	}

	country := lines[0][3]
	if country == "" {
		return status.Error(codes.InvalidArgument, "missing country at D1")
	}

	grade := lines[0][4]
	if grade == "" {
		return status.Error(codes.InvalidArgument, "missing Grade at E1")
	}

	subject := lines[0][5]
	if subject == "" {
		return status.Error(codes.InvalidArgument, "missing subject at F1")
	}

	startDate := lines[0][6]

	presetStudyPlan.ID = database.Text(presetStudyPlanID)
	presetStudyPlan.Name = database.Text(presetStudyPlanName)
	eCountry, err = rcv.country2Enum(country)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	_ = presetStudyPlan.Country.Set(eCountry.String())

	intGrade, err := i18n.ConvertStringGradeToInt(eCountry, grade)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid Grade at E1")
	}
	_ = presetStudyPlan.Grade.Set(intGrade)

	eSubject, err = rcv.subject2Enum(subject)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	_ = presetStudyPlan.Subject.Set(eSubject.String())

	_ = presetStudyPlan.StartDate.Set(timeutil.PSPStartDate(eCountry, startDate))

	// remove 3 lines header title
	for _, line := range lines[3:] {
		if len(line) < 3 {
			continue
		}
		topicId := line[1]
		if topicId == "" {
			continue
		}

		var week int
		for i, item := range line[2:] {
			if item == "1" {
				week = i + 1
				break
			}
		}
		if week == 0 {
			continue
		}

		presetStudyPlanWeeklyId := strings.Join([]string{presetStudyPlan.ID.String, topicId}, "-")
		if _, ok := mapPresetStudyPlanWeeklies[presetStudyPlanWeeklyId]; !ok {
			presetStudyPlanWeekly := new(entities.PresetStudyPlanWeekly)
			_ = presetStudyPlanWeekly.ID.Set(presetStudyPlanWeeklyId)
			presetStudyPlanWeekly.PresetStudyPlanID = presetStudyPlan.ID
			_ = presetStudyPlanWeekly.TopicID.Set(topicId)
			_ = presetStudyPlanWeekly.Week.Set(week)
			presetStudyPlanWeeklies = append(presetStudyPlanWeeklies, presetStudyPlanWeekly)
			mapPresetStudyPlanWeeklies[presetStudyPlanWeeklyId] = presetStudyPlanWeekly
		}
	}

	err = rcv.PresetStudyPlanRepo.BulkImport(ctx, rcv.DB, []*entities.PresetStudyPlan{presetStudyPlan}, presetStudyPlanWeeklies)
	if err != nil {
		return toStatusError(err)
	}
	return err
}

func (rcv *MasterDataService) subject2Enum(masterId string) (pb.Subject, error) {
	switch masterId {
	case "Physics", "Vật Lý":
		return pb.SUBJECT_PHYSICS, nil
	case "Math", "MA", "Toán":
		return pb.SUBJECT_MATHS, nil
	case "Biology", "Sinh Học", "Sinh":
		return pb.SUBJECT_BIOLOGY, nil
	case "Chemistry", "Hóa Học", "Hóa":
		return pb.SUBJECT_CHEMISTRY, nil
	case "Geography":
		return pb.SUBJECT_GEOGRAPHY, nil
	case "English", "Anh Văn", "Tiếng Anh":
		return pb.SUBJECT_ENGLISH, nil
	case "English2":
		return pb.SUBJECT_ENGLISH_2, nil
	case "Literature", "Ngữ Văn", "Văn":
		return pb.SUBJECT_LITERATURE, nil
	default:
		return pb.SUBJECT_NONE, errors.New(masterId + " is not defined in enum")
	}
}

func (rcv *MasterDataService) country2Enum(masterId string) (pb.Country, error) {
	masterId = strings.ToUpper(masterId)
	switch masterId {
	case "MASTER":
		return pb.COUNTRY_MASTER, nil
	case "VN":
		return pb.COUNTRY_VN, nil
	default:
		return pb.COUNTRY_NONE, errors.New(masterId + " is not defined in enum")
	}
}

func (rcv *MasterDataService) checkRoleAdmin(ctx context.Context) error {
	currentUserID := interceptors.UserIDFromContext(ctx)
	uGroup, err := rcv.UserRepo.UserGroup(ctx, rcv.DB, database.Text(currentUserID))
	if err != nil {
		return errors.Wrapf(err, "s.UserRepo.UserGroup: userID: %q", currentUserID)
	}
	if uGroup != entities.UserGroupAdmin {
		return status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return nil
}

func (rcv *MasterDataService) parseSchools(b io.Reader) ([]*pb.School, error) {
	r := csv.NewReader(b)
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	// ignore header
	lines = lines[1:]
	schools := make([]*pb.School, 0, len(lines))
	for _, l := range lines {
		country, err := rcv.country2Enum(l[1])
		if err != nil {
			return nil, err
		}

		city := &pb.City{
			Name:    l[2],
			Country: country,
		}
		district := &pb.District{
			Name:    l[3],
			Country: country,
			City:    city,
		}
		s := &pb.School{
			Name:     l[0],
			Country:  country,
			City:     city,
			District: district,
		}
		if lat, long := l[4], l[5]; lat != "" && long != "" {
			latitude, err := strconv.ParseFloat(lat, 64)
			if err != nil {
				return nil, err
			}
			longitude, err := strconv.ParseFloat(long, 64)
			if err != nil {
				return nil, err
			}
			s.Point = &pb.Point{Lat: latitude, Long: longitude}
		}
		schools = append(schools, s)
	}
	return schools, nil
}
