package services

import (
	"context"
	"io"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_protobuf "github.com/manabie-com/backend/mock/bob/protobuf"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"code.cloudfoundry.org/bytefmt"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTopicType2Enum(t *testing.T) {

}

func TestCountry2Enum(t *testing.T) {
	t.Parallel()
	countryString := []string{"MASTER", "VN", "NONE"}
	expectedTypeEnum := []pb.Country{pb.COUNTRY_MASTER, pb.COUNTRY_VN, pb.COUNTRY_NONE}
	s := &MasterDataService{}
	for i, country := range countryString {
		enumCountry, err := s.country2Enum(country)
		assert.Equal(t, enumCountry, expectedTypeEnum[i])
		if enumCountry == pb.COUNTRY_NONE {
			assert.Error(t, err)
		}
	}
}

func TestGetCountryFromTopicId(t *testing.T) {

}

func TestSubject2Enum(t *testing.T) {
	t.Parallel()
	testCase := map[pb.Subject][]string{
		pb.SUBJECT_PHYSICS:    {"Physics", "Vật Lý"},
		pb.SUBJECT_MATHS:      {"Math", "MA", "Toán"},
		pb.SUBJECT_BIOLOGY:    {"Biology", "Sinh Học", "Sinh"},
		pb.SUBJECT_CHEMISTRY:  {"Chemistry", "Hóa Học", "Hóa"},
		pb.SUBJECT_GEOGRAPHY:  {"Geography"},
		pb.SUBJECT_ENGLISH:    {"English", "Anh Văn", "Tiếng Anh"},
		pb.SUBJECT_ENGLISH_2:  {"English2"},
		pb.SUBJECT_LITERATURE: {"Literature", "Ngữ Văn", "Văn"},
		pb.SUBJECT_NONE:       {"wrongtopic", ""},
	}
	s := &MasterDataService{}
	for subject, subjectStrings := range testCase {
		for _, v := range subjectStrings {
			enumSubject, err := s.subject2Enum(v)
			assert.Equal(t, enumSubject, subject)
			if enumSubject == pb.SUBJECT_NONE {
				assert.Error(t, err)
			}
		}
	}
}

func TestNewLearningObjective(t *testing.T) {
}

func TestNewLearningObjectiveFromCsvFileCountryMaster(t *testing.T) {

}

func TestNewLearningObjectiveFromCsvFile(t *testing.T) {

}

func TestImportTableLO(t *testing.T) {

}

func TestImportTablePresetStudyPlan(t *testing.T) {
	t.Parallel()
	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)
	s := &MasterDataService{
		PresetStudyPlanRepo: presetStudyPlanRepo,
	}
	testcases := []TestCase{
		{
			name: "happy case",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,G12,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Learning T2,VN12-MA2,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Learning T3,VN12-MA3,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Practice P1,VN12-MA4,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("BulkImport", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err parse grade",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,InvalidGrade,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Practice P1,VN12-MA4,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: status.Error(codes.InvalidArgument, "invalid Grade at E1"),
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("BulkImport", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err parse country",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,InvalidCountry,G12,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Practice P1,VN12-MA4,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: status.Error(codes.InvalidArgument, "INVALIDCOUNTRY is not defined in enum"),
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("BulkImport", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err parse subject",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,G12,Invalid Subject,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Practice P1,VN12-MA4,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: status.Error(codes.InvalidArgument, "Invalid Subject is not defined in enum"),
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("BulkImport", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "empty content",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,G12,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46`),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("BulkImport", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "empty topic id",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,G12,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Learning T2,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Learning T3,VN12-MA3,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Practice P1,VN12-MA4,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				presetStudyPlanRepo.On("BulkImport", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "missing preset student plan id",
			req: []byte(`Math study plan,,Standard,VN,G12,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: status.Error(codes.InvalidArgument, "missing presetStudyPlanId at B1"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing preset student plan name",
			req: []byte(`Math study plan,S-VN-G12-MA,,VN,G12,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: status.Error(codes.InvalidArgument, "missing presetStudyPlanName at C1"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing country",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,,G12,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: status.Error(codes.InvalidArgument, "missing country at D1"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing grade",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: status.Error(codes.InvalidArgument, "missing Grade at E1"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "missing subject",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,G12,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: status.Error(codes.InvalidArgument, "missing subject at F1"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "wrong format csv",
			req: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,G12
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
			expectedErr: errors.New("record on line 2: wrong number of fields"),
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			req := testCase.req.([]byte)
			err := s.importTablePresetStudyPlan(ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestImportTableTopic(t *testing.T) {

}

func TestCheckRoleAdmin(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userRepo := new(mock_repositories.MockUserRepo)
	s := &MasterDataService{
		UserRepo: userRepo,
	}
	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         interceptors.ContextWithUserID(ctx, "admin"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
			},
		},
		{
			name:        "cant find user",
			ctx:         interceptors.ContextWithUserID(ctx, "invalidId"),
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return("", pgx.ErrNoRows)
			},
		},
		{
			name:        "student denied case",
			ctx:         interceptors.ContextWithUserID(ctx, "student"),
			expectedErr: status.Error(codes.PermissionDenied, codes.PermissionDenied.String()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupStudent, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.checkRoleAdmin(testCase.ctx)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}

}

func TestImportLO(t *testing.T) {
}

func TestImportTopic(t *testing.T) {
}

func TestImportPresetStudyPlan(t *testing.T) {
	t.Parallel()
	// configurations.Load()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cs, _ := bytefmt.ToBytes("500kb")
	chunksize := int64(cs)
	masterDataService_ImportPresetStudyPlanServer := new(mock_protobuf.MasterDataService_ImportPresetStudyPlanServer)
	userRepo := new(mock_repositories.MockUserRepo)
	presetStudyPlanRepo := new(mock_repositories.MockPresetStudyPlanRepo)
	s := &MasterDataService{
		Cfg: &configurations.Config{
			Upload: configs.UploadConfig{
				MaxChunkSize: chunksize,
				MaxFileSize:  chunksize * 2,
			},
		},
		UserRepo:            userRepo,
		PresetStudyPlanRepo: presetStudyPlanRepo,
	}

	testcases := []TestCase{
		{
			name:        "transmit error",
			ctx:         interceptors.ContextWithUserID(ctx, "transmit error"),
			expectedErr: errors.New("failed unexpectadely while reading chunks from stream"),
			setup: func(ctx context.Context) {
				masterDataService_ImportPresetStudyPlanServer.On("Context").Once().Return(ctx)
				masterDataService_ImportPresetStudyPlanServer.On("Recv").Once().Return(nil, errors.New("transmit error"))
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
			},
		},
		{
			name:        "oversize",
			ctx:         interceptors.ContextWithUserID(ctx, "oversize"),
			expectedErr: status.Error(codes.InvalidArgument, "chunk size over limited"),
			setup: func(ctx context.Context) {
				overMaxPayload := make([]byte, chunksize+1)
				rand.Read(overMaxPayload)
				overSizeRequest := &pb.ImportPresetStudyPlanRequest{
					Payload: overMaxPayload,
				}
				masterDataService_ImportPresetStudyPlanServer.On("Context").Once().Return(ctx)
				masterDataService_ImportPresetStudyPlanServer.On("Recv").Once().Return(overSizeRequest, nil)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
			},
		},
		{
			name:        "Happy Case",
			ctx:         interceptors.ContextWithUserID(ctx, "Happy Case"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				validRequest := &pb.ImportPresetStudyPlanRequest{
					Payload: []byte(`Math study plan,S-VN-G12-MA,Standard,VN,G12,MA,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
			Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
			Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Learning T2,VN12-MA2,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Learning T3,VN12-MA3,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
			Practice P1,VN12-MA4,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`),
				}
				masterDataService_ImportPresetStudyPlanServer.On("Context").Times(2).Return(ctx)
				masterDataService_ImportPresetStudyPlanServer.On("Recv").Times(1).Return(validRequest, nil)
				masterDataService_ImportPresetStudyPlanServer.On("Recv").Times(1).Return(nil, io.EOF)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
				presetStudyPlanRepo.On("BulkImport", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				masterDataService_ImportPresetStudyPlanServer.On("SendAndClose", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "Invalid file size",
			ctx:         interceptors.ContextWithUserID(ctx, "Invalid file size"),
			expectedErr: status.Error(codes.InvalidArgument, "file size over limited"),
			setup: func(ctx context.Context) {
				s.Cfg.Upload.MaxChunkSize = chunksize
				s.Cfg.Upload.MaxFileSize = chunksize - 1
				overMaxPayload := make([]byte, chunksize+1)
				rand.Read(overMaxPayload)
				req := &pb.ImportPresetStudyPlanRequest{
					Payload: overMaxPayload,
				}
				masterDataService_ImportPresetStudyPlanServer.On("Context").Once().Return(ctx)
				masterDataService_ImportPresetStudyPlanServer.On("Recv").Once().Return(req, nil)
				userRepo.On("UserGroup", ctx, mock.Anything, mock.Anything).Once().Return(entities.UserGroupAdmin, nil)
				presetStudyPlanRepo.On("BulkImport", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.ImportPresetStudyPlan(masterDataService_ImportPresetStudyPlanServer)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestParseSchoolsFromCSV(t *testing.T) {
	t.Parallel()
	s := &MasterDataService{}

	testCases := []struct {
		name    string
		payload string
		want    []*pb.School
	}{
		{
			name: "full sample",
			payload: "Name,Country,City,District,Latitude,Longitude\n" +
				"School 1,VN,Hồ Chí Minh,1,,\n" +
				"School 2,VN,Hồ Chí Minh,Bình Thạnh,10.2345,11.5678\n" +
				"School 1,VN,Hà Nội,1,,\n" +
				"School 2,VN,Đà Nẵng,1,1.234,2.345\n",
			want: []*pb.School{
				{
					Name:    "School 1",
					Country: pb.COUNTRY_VN,
					City: &pb.City{
						Name:    "Hồ Chí Minh",
						Country: pb.COUNTRY_VN,
					},
					District: &pb.District{
						Name:    "1",
						Country: pb.COUNTRY_VN,
						City: &pb.City{
							Name:    "Hồ Chí Minh",
							Country: pb.COUNTRY_VN,
						},
					},
					Point: nil,
				},
				{
					Name:    "School 2",
					Country: pb.COUNTRY_VN,
					City: &pb.City{
						Name:    "Hồ Chí Minh",
						Country: pb.COUNTRY_VN,
					},
					District: &pb.District{
						Name:    "Bình Thạnh",
						Country: pb.COUNTRY_VN,
						City: &pb.City{
							Name:    "Hồ Chí Minh",
							Country: pb.COUNTRY_VN,
						},
					},
					Point: &pb.Point{10.2345, 11.5678},
				},
				{
					Name:    "School 1",
					Country: pb.COUNTRY_VN,
					City: &pb.City{
						Name:    "Hà Nội",
						Country: pb.COUNTRY_VN,
					},
					District: &pb.District{
						Name:    "1",
						Country: pb.COUNTRY_VN,
						City: &pb.City{
							Name:    "Hà Nội",
							Country: pb.COUNTRY_VN,
						},
					},
					Point: nil,
				},
				{
					Name:    "School 2",
					Country: pb.COUNTRY_VN,
					City: &pb.City{
						Name:    "Đà Nẵng",
						Country: pb.COUNTRY_VN,
					},
					District: &pb.District{
						Name:    "1",
						Country: pb.COUNTRY_VN,
						City: &pb.City{
							Name:    "Đà Nẵng",
							Country: pb.COUNTRY_VN,
						},
					},
					Point: &pb.Point{1.234, 2.345},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			schools, err := s.parseSchools(strings.NewReader(tc.payload))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(schools) != len(tc.want) {
				t.Fatalf("expected total schools: %d, got: %d", len(tc.want), len(schools))
			}
			if !reflect.DeepEqual(schools, tc.want) {
				t.Fatalf("expected schools: \n%+v, \ngot: \n%+v", tc.want, schools)
			}
		})
	}
}
