package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/speeches"
	entities_yasuo "github.com/manabie-com/backend/internal/yasuo/entities"
	y_repositories "github.com/manabie-com/backend/internal/yasuo/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AssignLosToQuiz for old version quizzes which dont have column lo ids, we need to migrate assign lo ids to old quizzes
func (s *CourseService) AssignLosToQuizV1(ctx context.Context, quiz *entities_bob.Quiz) error {
	quizSets, err := s.QuizSetRepo.GetQuizSetsContainQuiz(ctx, s.EurekaDBTrace, quiz.ExternalID)
	if err != nil {
		return err
	}
	loID := make([]string, 0, len(quizSets))
	duplicateLOMap := make(map[string]bool)
	for _, set := range quizSets {
		if _, ok := duplicateLOMap[set.LoID.String]; !ok {
			loID = append(loID, set.LoID.String)
		}
		duplicateLOMap[set.LoID.String] = true
	}

	err = quiz.LoIDs.Set(loID)
	if err != nil {
		return err
	}
	return nil
}

func inTextArray(ta pgtype.TextArray, s pgtype.Text) bool {
	for _, t := range ta.Elements {
		if t.Status == pgtype.Present && s.Status == pgtype.Present && t.String == s.String {
			return true
		}
	}
	return false
}

func getQuizSet(ctx context.Context, s *CourseService, db database.QueryExecer, loID string) (*entities_bob.QuizSet, error) {
	set, err := s.QuizSetRepo.Search(ctx, db, repositories.QuizSetFilter{
		ObjectiveIDs: database.TextArray([]string{loID}),
		Status:       database.Text(pb.QUIZSET_STATUS_APPROVED.String()),
		Limit:        1,
	})
	if err != nil {
		return nil, err
	}

	if len(set) == 0 {
		return nil, fmt.Errorf("no quiz_sets found: %w", pgx.ErrNoRows)
	}

	return set[0], nil
}

// nolint
// UploadHtmlContent upload html content
func (s *CourseService) UploadHtmlContent(ctx context.Context, content string) (string, error) {
	url, fileName := generateUploadURL(s.Config.Storage.Endpoint, s.Config.Storage.Bucket, content)
	if s.Config.Common.Environment != "local" {
		client, err := storage.NewClient(ctx)
		if err != nil {
			return "", fmt.Errorf("err storage.NewClient: %w", err)
		}
		wc := client.Bucket(s.Config.Storage.Bucket).Object(fileName[1:]).NewWriter(ctx)
		err = uploadToCloudStorage(wc, content, "text/html; charset=utf-8")
		if err != nil {
			return "", fmt.Errorf("err uploadToCloudStorage: %w", err)
		}
	} else {
		err := uploadToS3(ctx, s.Uploader, content, s.Config.Storage.Bucket, fileName, "text/html; charset=UTF-8")
		if err != nil {
			return "", fmt.Errorf("err uploadToS3: %w", err)
		}
	}

	return url, nil
}

// nolint
func generateUploadURL(endpoint, bucket, content string) (url, fileName string) {
	h := md5.New()
	io.WriteString(h, content)
	fileName = "/content/" + fmt.Sprintf("%x.html", h.Sum(nil))

	return endpoint + "/" + bucket + fileName, fileName
}

func foundExternalID(a []pgtype.Text, s string) bool {
	var found bool
	for _, e := range a {
		if e.String == s {
			found = true
			break
		}
	}

	return found
}

func quizCfg2ArrayString(cfg interface{}) []string {
	switch cfgs := cfg.(type) {
	case []pb.QuizOptionConfig:
		result := make([]string, 0, len(cfgs))
		for _, c := range cfgs {
			result = append(result, c.String())
		}

		return result
	case []cpb.QuizOptionConfig:
		result := make([]string, 0, len(cfgs))
		for _, c := range cfgs {
			result = append(result, c.String())
		}

		return result
	default:
		return nil
	}
}

func (s *CourseService) generateAudioFilesHandler(ctx context.Context, tx pgx.Tx, createdQuizzes []*entities_bob.Quiz) error {
	userID := interceptors.UserIDFromContext(ctx)

	speechesReq := new(bpb.GenerateAudioFileRequest)
	for i, quiz := range createdQuizzes {
		q, err := quiz.GetQuestionV2()
		if err != nil {
			return err
		}

		if lang := speeches.GetLanguage(q.Attribute.Configs); lang != "" && q.GetText() != "" {
			if golibs.InArrayString(lang, speeches.WhiteList) {
				if ok, existed := s.SpeechesRepo.CheckExistedSpeech(ctx, tx, &y_repositories.CheckExistedSpeechReq{
					Text: database.Text(q.GetText()),
					Config: database.JSONB(&y_repositories.SpeechConfig{
						Language: lang,
					}),
				}); !ok {
					speechesReq.Options = append(speechesReq.Options, &bpb.AudioOptionRequest{
						Text: q.GetText(),
						Configs: &bpb.AudioConfig{
							Language: lang,
						},
						QuizId: quiz.ID.String,
						Type:   bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_TERM,
					})
				} else {
					q.Attribute.AudioLink = existed.Link.String
					createdQuizzes[i].Question = database.JSONB(q)
				}
			} else {
				q.Attribute.AudioLink = ""
				createdQuizzes[i].Question = database.JSONB(q)
			}
		}

		o, err := quiz.GetOptions()
		if err != nil {
			return err
		}
		for j, each := range o {
			if lang := speeches.GetLanguage(each.Attribute.Configs); lang != "" && each.GetText() != "" {
				if golibs.InArrayString(lang, speeches.WhiteList) {
					if ok, existed := s.SpeechesRepo.CheckExistedSpeech(ctx, tx, &y_repositories.CheckExistedSpeechReq{
						Text: database.Text(each.GetText()),
						Config: database.JSONB(&y_repositories.SpeechConfig{
							Language: lang,
						}),
					}); !ok {
						speechesReq.Options = append(speechesReq.Options, &bpb.AudioOptionRequest{
							Text: each.GetText(),
							Configs: &bpb.AudioConfig{
								Language: lang,
							},
							QuizId: quiz.ID.String,
							Type:   bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_DEFINITION,
						})
					} else {
						o[j].Attribute.AudioLink = existed.Link.String
					}
				} else {
					o[j].Attribute.AudioLink = ""
				}
			}
		}

		createdQuizzes[i].Options = database.JSONB(o)
	}

	if len(speechesReq.Options) > 0 {
		mdctx, err := interceptors.GetOutgoingContext(ctx)
		if err != nil {
			return status.Error(codes.Unauthenticated, fmt.Errorf("CourseService.generateAudioFilesHandler.GetOutgoingContext: %w", err).Error())
		}
		resp, err := s.MediaModifierService.GenerateAudioFile(mdctx, speechesReq)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("MediaModifierService.GenerateAudioFile: %w", err).Error())
		}

		upsertSpeechesReq := make([]*entities_yasuo.Speeches, 0)
		for _, each := range createdQuizzes {
			for _, option := range resp.Options {
				if each.ID.String == option.QuizId {
					entity := new(entities_yasuo.Speeches)
					database.AllNullEntity(entity)
					if err := multierr.Combine(
						entity.SpeechID.Set(idutil.ULIDNow()),
						entity.Sentence.Set(option.Text),
						entity.Link.Set(option.Link),
						entity.Settings.Set(&y_repositories.SpeechConfig{
							Language: option.Configs.Language,
						}),
						entity.CreatedBy.Set(userID),
						entity.UpdatedBy.Set(userID),
						entity.Type.Set(option.Type.String()),
						entity.QuizID.Set(option.QuizId),
					); err != nil {
						return err
					}

					upsertSpeechesReq = append(upsertSpeechesReq, entity)
				}
			}
		}

		createdSpeeches, err := s.SpeechesRepo.UpsertSpeeches(ctx, tx, upsertSpeechesReq)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("SpeechesRepo.UpsertSpeeches: %w", err).Error())
		}

		for i, each := range createdQuizzes {
			r, err := each.GetQuestionV2()
			if err != nil {
				return err
			}

			o, err := each.GetOptions()
			if err != nil {
				return err
			}

			changeOptionIdx := []int{}
			for i, option := range o {
				if !golibs.InArrayString(cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_NONE.String(), option.Attribute.Configs) {
					changeOptionIdx = append(changeOptionIdx, i)
				}
			}

			for _, speech := range createdSpeeches {
				if each.ID == speech.QuizID {
					switch speech.Type.String {
					case bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_TERM.String():
						r.Attribute = entities_bob.QuizItemAttribute{
							AudioLink: speech.Link.String,
							ImgLink:   r.Attribute.ImgLink,
							Configs:   r.Attribute.Configs,
						}
						if err := createdQuizzes[i].Question.Set(r); err != nil {
							return fmt.Errorf("Question.Set %w", err)
						}
					case bpb.AudioOptionType_FLASHCARD_AUDIO_TYPE_DEFINITION.String():
						for z, idx := range changeOptionIdx {
							if o[idx].Content.GetText() == speech.Sentence.String {
								o[idx].Attribute = entities_bob.QuizItemAttribute{
									AudioLink: speech.Link.String,
									Configs:   o[idx].Attribute.Configs,
								}
								changeOptionIdx = removeIndex(changeOptionIdx, z)
							}
						}
					}
				}
			}

			if err := createdQuizzes[i].Options.Set(o); err != nil {
				return err
			}
		}
	}

	if _, err := s.QuizRepo.Upsert(ctx, tx, createdQuizzes); err != nil {
		return status.Error(codes.Internal, fmt.Errorf("QuizRepo.Upsert: %w", err).Error())
	}

	return nil
}

func removeIndex(s []int, index int) []int {
	return append(s[:index], s[index+1:]...)
}
