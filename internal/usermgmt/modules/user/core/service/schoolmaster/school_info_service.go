package schoolmaster

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SchoolInfoService struct {
	pb.UnimplementedSchoolInfoServiceServer
	DB             database.Ext
	JSM            nats.JetStreamManagement
	SchoolInfoRepo interface {
		BulkImport(context.Context, database.QueryExecer, []*entity.SchoolInfo) []*repository.ImportError
	}
}

func NewSchoolInfoService(
	db database.Ext,
	schoolInfoRepo *repository.SchoolInfoRepo,
	jsm nats.JetStreamManagement,
) *SchoolInfoService {
	return &SchoolInfoService{
		DB:             db,
		SchoolInfoRepo: schoolInfoRepo,
		JSM:            jsm,
	}
}

func (s *SchoolInfoService) ImportSchoolInfo(ctx context.Context, req *pb.ImportSchoolInfoRequest) (res *pb.ImportSchoolInfoResponse, err error) {
	var (
		errorCSVs []*pb.ImportSchoolInfoResponse_ImportSchoolInfoError
		lines     [][]string
	)

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	lines, err = readAndValidatePayload(req.Payload)
	if err != nil {
		return
	}
	allSchoolInfo := []*entity.SchoolInfo{}
	for i, line := range lines[1:] {
		var (
			errLineRes pb.ImportSchoolInfoResponse_ImportSchoolInfoError
			school     entity.SchoolInfo
		)
		school, errLineRes = convertLineCSVToSchool(line, i, resourcePath)
		if errLineRes.Error != "" {
			errorCSVs = append(errorCSVs, &errLineRes)
			continue
		}
		allSchoolInfo = append(allSchoolInfo, &school)
	}
	errors := s.SchoolInfoRepo.BulkImport(ctx, s.DB, allSchoolInfo)
	errorCSVs = append(errorCSVs, convertImportErrorToErrRes(errors)...)

	res = &pb.ImportSchoolInfoResponse{
		Errors: errorCSVs,
	}
	return
}

func readAndValidatePayload(payload []byte) (lines [][]string, err error) {
	r := csv.NewReader(bytes.NewReader(payload))
	lines, err = r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, "no data in csv file")
	}

	if len(lines[0]) != 6 {
		return nil, status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 6")
	}

	if strings.ToLower(lines[0][0]) != "school_id" {
		return nil, status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'school_id'")
	}
	if strings.ToLower(lines[0][1]) != "school_name" {
		return nil, status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'school_name'")
	}
	if strings.ToLower(lines[0][2]) != "school_name_phonetic" {
		return nil, status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'school_name_phonetic'")
	}
	if strings.ToLower(lines[0][3]) != "school_level_id" {
		return nil, status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'school_level_id'")
	}
	if strings.ToLower(lines[0][4]) != "address" {
		return nil, status.Error(codes.InvalidArgument, "csv file invalid format - fifth column (toLowerCase) should be 'address'")
	}
	if strings.ToLower(lines[0][5]) != "is_archived" {
		return nil, status.Error(codes.InvalidArgument, "csv file invalid format - sixth column (toLowerCase) should be 'is_archived'")
	}
	return
}

func convertLineCSVToSchool(line []string, order int, resourcePath string) (school entity.SchoolInfo, errLine pb.ImportSchoolInfoResponse_ImportSchoolInfoError) {
	database.AllNullEntity(&school)
	const (
		ID = iota
		Name
		NamePhonetic
		LevelID
		Address
		IsArchived
	)
	var (
		id         string
		err        error
		isArchived bool
	)
	mandatory := []int{Name, IsArchived}
	if !checkMandatoryColumn(line, mandatory) {
		return school, convertErrToErrResForEachLineCSV(fmt.Errorf("missing mandatory column"), order)
	}

	isArchived, err = strconv.ParseBool(line[IsArchived])
	if err != nil {
		return school, convertErrToErrResForEachLineCSV(fmt.Errorf("error parsing IsArchived: %w", err), order)
	}

	if strings.TrimSpace(line[ID]) == "" {
		if err = school.ID.Set(idutil.ULIDNow()); err != nil {
			return school, convertErrToErrResForEachLineCSV(err, order)
		}
	} else {
		id = strings.TrimSpace(line[ID])
		if err != nil {
			return school, convertErrToErrResForEachLineCSV(err, order)
		}
		if err = school.ID.Set(id); err != nil {
			return school, convertErrToErrResForEachLineCSV(err, order)
		}
	}

	now := time.Now()
	if err = multierr.Combine(
		school.Name.Set(strings.TrimSpace(line[Name])),
		school.NamePhonetic.Set(strings.TrimSpace(line[NamePhonetic])),
		school.LevelID.Set(strings.TrimSpace(line[LevelID])),
		school.Address.Set(strings.TrimSpace(line[Address])),
		school.IsArchived.Set(isArchived),
		school.ResourcePath.Set(resourcePath),

		// in update case, it will automatically ignore the CreatedAt and DeletedAt field
		school.UpdatedAt.Set(now),
		school.CreatedAt.Set(now),
		school.DeletedAt.Set(nil),
	); err != nil {
		return school, convertErrToErrResForEachLineCSV(err, order)
	}

	return
}

func checkMandatoryColumn(column []string, positions []int) bool {
	for _, position := range positions {
		if column[position] == "" {
			return false
		}
	}
	return true
}

func convertImportErrorToErrRes(errors []*repository.ImportError) (errRes []*pb.ImportSchoolInfoResponse_ImportSchoolInfoError) {
	for _, v := range errors {
		errRes = append(errRes, &pb.ImportSchoolInfoResponse_ImportSchoolInfoError{
			RowNumber: v.RowNumber, // i = 0 <=> line number 2 in csv file
			Error:     v.Error,
		})
	}
	return errRes
}

func convertErrToErrResForEachLineCSV(err error, i int) pb.ImportSchoolInfoResponse_ImportSchoolInfoError {
	return pb.ImportSchoolInfoResponse_ImportSchoolInfoError{
		RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
		Error:     fmt.Sprintf("unable to parse school_info item: %s", err),
	}
}
