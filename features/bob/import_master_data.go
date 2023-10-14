package bob

import (
	"bytes"
	"context"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"
)

func (s *suite) adminUploadPreset_study_planFile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var file bytes.Buffer
	file.Write([]byte(`Math study plan,S-VN-G12-MA,Standard,VN,G11,MA,August 09,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T2,VN12-MA2,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T3,VN12-MA3,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Practice P1,VN12-MA4,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T4,VN12-MA5,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T5,VN12-MA6,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T6,VN12-MA7,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Practice P2,VN12-MA8,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T7,VN12-MA9,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T8,VN12-MA10,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T9,VN12-MA11,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Practice P3,VN12-MA12,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T10,VN12-MA13,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T11,VN12-MA14,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T12,VN12-MA15,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Practice P4,VN12-MA16,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T13,VN12-MA17,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T14,VN12-MA18,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T15,VN12-MA19,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Practice P5,VN12-MA20,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T16,VN12-MA21,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T17,VN12-MA22,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T18,VN12-MA23,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Practice P6,VN12-MA24,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T19,VN12-MA25,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T20,VN12-MA26,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T21,VN12-MA27,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,
Practice P7,VN12-MA28,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,
Learning T22,VN12-MA29,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,
Learning T23,VN12-MA30,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,,
Learning T24,VN12-MA31,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,,
Practice P8,VN12-MA32,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,
Learning T25,VN12-MA33,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,,
Learning T26,VN12-MA34,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,,
Learning T27,VN12-MA35,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,,
Practice P9,VN12-MA36,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,
Learning T28,VN12-MA37,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,,
Learning T29,VN12-MA38,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,,
Learning T30,VN12-MA39,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,,
Practice P10,VN12-MA40,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,
Learning T31,VN12-MA41,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,,
Learning T32,VN12-MA42,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,,
Learning T33,VN12-MA43,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,,
Practice P11,VN12-MA44,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,
Learning T34,VN12-MA45,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,,
Learning T35,VN12-MA46,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,,
Learning T36,VN12-MA47,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,,
Practice P12,VN12-MA48,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,,
Practice P13,VN12-MA49,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,,
Practice P14,VN12-MA50,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,,
Practice P15,VN12-MA51,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,,
Practice P16,VN12-MA52,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,,
Practice P17,VN12-MA53,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,,
Practice P18,VN12-MA54,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,,
Practice P19,VN12-MA55,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,,
Practice P20,VN12-MA56,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,1,`))
	stream, _ := pb.NewMasterDataServiceClient(s.Conn).ImportPresetStudyPlan(s.signedCtx(ctx))

	chunksize, _ := bytefmt.ToBytes("500kb")
	buf := make([]byte, chunksize)

	for {
		n, err := file.Read(buf)
		if err != nil {
			break
		}
		_ = stream.Send(&pb.ImportPresetStudyPlanRequest{
			Payload: buf[:n],
		})
	}
	stepState.Response, stepState.ResponseErr = stream.CloseAndRecv()
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) uploadInvalidPreset_study_planFile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var file bytes.Buffer
	file.Write([]byte(`Math study plan,S-VN-G12-MA,Standard,VN,G11,MA,August 09,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
,,August,,September,,,,October,,,,November,,,,December,,,,January,,,,February,,,,March,,,,April,,,,May,,,,June,,,,July,,,
Topic name,Topic ID,W1,W2,W3,W4,W5,W6,W7,W8,W9,W10,W11,W12,W13,W14,W15,W16,W17,W18,W19,W20,W21,W22,W23,W24,W25,W26,W27,W28,W29,W30,W31,W32,W33,W34,W35,W36,W37,W38,W39,W40,W41,W42,W43,W44,W45,W46
Learning T1,VN12-MA1,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T2,VN12-MA2,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T3,VN12-MA3,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Practice P1,INVALID-1,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T4,INVALID-2,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,
Learning T5,VN12-MA6,,,,,1,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`))
	stream, _ := pb.NewMasterDataServiceClient(s.Conn).ImportPresetStudyPlan(s.signedCtx(ctx))

	chunksize, _ := bytefmt.ToBytes("500kb")
	buf := make([]byte, chunksize)

	for {
		n, err := file.Read(buf)
		if err != nil {
			break
		}
		_ = stream.Send(&pb.ImportPresetStudyPlanRequest{
			Payload: buf[:n],
		})
	}
	stepState.Response, stepState.ResponseErr = stream.CloseAndRecv()
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustStoreAllDataInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var startDate time.Time
	_ = s.DB.QueryRow(ctx, "SELECT start_date FROM preset_study_plans WHERE preset_study_plan_id='S-VN-G12-MA';").Scan(&startDate)

	expected := time.Date(time.Now().Year(), time.August, 9, 0, 0, 0, 0, time.UTC)
	if !startDate.Equal(expected) {
		return StepStateToContext(ctx, stepState), errors.New("fail to import preset_study_plans")
	}

	var count int
	_ = s.DB.QueryRow(ctx, "SELECT COUNT(*) FROM preset_study_plans_weekly").Scan(&count)
	if count == 0 {
		return StepStateToContext(ctx, stepState), errors.New("fail to import preset_study_plans_weeklies")
	}
	return StepStateToContext(ctx, stepState), nil
}
