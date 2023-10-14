package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/draft/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	dpb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"

	"github.com/jackc/pgtype"
)

type BDDSuite struct {
	DB   database.Ext
	Repo interface {
		AddInstance(ctx context.Context, db database.QueryExecer, b *entities.BDDInstance) error
		MarkInstanceEnded(ctx context.Context, db database.QueryExecer, e *entities.BDDInstance) error

		AddFeature(ctx context.Context, db database.QueryExecer, b *entities.BDDFeature) (pgtype.Text, error)
		MarkFeatureEnded(ctx context.Context, db database.QueryExecer, e *entities.BDDFeature) error
		SetFeatureStatus(ctx context.Context, db database.QueryExecer, e *entities.BDDFeature) error

		AddScenario(ctx context.Context, db database.QueryExecer, b *entities.BDDScenario) error
		MarkScenarioEnded(ctx context.Context, db database.QueryExecer, e *entities.BDDScenario) error

		AddStep(ctx context.Context, db database.QueryExecer, e *entities.BDDStep) error
		MarkStepEnded(ctx context.Context, db database.QueryExecer, e *entities.BDDStep) error

		RetrieveSkippedBDDTestsByRepository(ctx context.Context, db database.QueryExecer, repo pgtype.Varchar) ([]*entities.SkippedBDDTest, error)
	}
}

var _ dpb.BDDSuiteServiceServer = (*BDDSuite)(nil)

func (b *BDDSuite) AddInstance(ctx context.Context, req *dpb.AddInstanceRequest) (*dpb.AddInstanceResponse, error) {
	e := toBDDInstanceEntity(req)
	if err := b.Repo.AddInstance(ctx, b.DB, e); err != nil {
		return nil, fmt.Errorf("AddInstance: %v", err)
	}
	return &dpb.AddInstanceResponse{Id: e.ID.String}, nil
}

func (b *BDDSuite) MarkInstanceEnded(ctx context.Context, req *dpb.MarkInstanceEndedRequest) (*dpb.MarkInstanceEndedResponse, error) {
	e := new(entities.BDDInstance)
	e.ID.Set(req.Id)
	e.Status.Set(req.Status)
	e.StatusStatistics.Set(req.Stats)
	e.EndedAt.Set(time.Now().UTC())
	if err := b.Repo.MarkInstanceEnded(ctx, b.DB, e); err != nil {
		return nil, fmt.Errorf("MarkInstanceEnded: %v", err)
	}
	return &dpb.MarkInstanceEndedResponse{}, nil
}

func (b *BDDSuite) AddFeature(ctx context.Context, req *dpb.AddFeatureRequest) (*dpb.AddFeatureResponse, error) {
	e := toBDDFeatureEntity(req)
	featureID, err := b.Repo.AddFeature(ctx, b.DB, e)
	if err != nil {
		return nil, fmt.Errorf("AddFeature: %v", err)
	}
	return &dpb.AddFeatureResponse{Id: featureID.String}, nil
}

func (b *BDDSuite) MarkFeatureEnded(ctx context.Context, req *dpb.MarkFeatureEndedRequest) (*dpb.MarkFeatureEndedResponse, error) {
	e := new(entities.BDDFeature)
	e.ID.Set(req.Id)
	e.Status.Set(req.Status)
	e.EndedAt.Set(time.Now().UTC())
	if err := b.Repo.MarkFeatureEnded(ctx, b.DB, e); err != nil {
		return nil, fmt.Errorf("MarkFeatureEnded: %v", err)
	}
	return &dpb.MarkFeatureEndedResponse{}, nil
}

func (b *BDDSuite) SetFeatureStatus(ctx context.Context, req *dpb.SetFeatureStatusRequest) (*dpb.SetFeatureStatusResponse, error) {
	e := new(entities.BDDFeature)
	e.ID.Set(req.Id)
	e.Status.Set(req.Status)
	if err := b.Repo.SetFeatureStatus(ctx, b.DB, e); err != nil {
		return nil, fmt.Errorf("MarkFeatureEnded: %v", err)
	}
	return &dpb.SetFeatureStatusResponse{}, nil
}

func (b *BDDSuite) AddScenario(ctx context.Context, req *dpb.AddScenarioRequest) (*dpb.AddScenarioResponse, error) {
	e := toBDDScenarioEntity(req)
	if err := b.Repo.AddScenario(ctx, b.DB, e); err != nil {
		return nil, fmt.Errorf("AddScenario: %v", err)
	}
	return &dpb.AddScenarioResponse{Id: e.ID.String}, nil
}

func (b *BDDSuite) MarkScenarioEnded(ctx context.Context, req *dpb.MarkScenarioEndedRequest) (*dpb.MarkScenarioEndedResponse, error) {
	e := new(entities.BDDScenario)
	e.ID.Set(req.Id)
	e.Status.Set(req.Status)
	e.EndedAt.Set(time.Now().UTC())
	if err := b.Repo.MarkScenarioEnded(ctx, b.DB, e); err != nil {
		return nil, fmt.Errorf("MarkScenarioEnded: %v", err)
	}
	return &dpb.MarkScenarioEndedResponse{}, nil
}

func (b *BDDSuite) AddStep(ctx context.Context, req *dpb.AddStepRequest) (*dpb.AddStepResponse, error) {
	e := toBDDStepEntity(req)
	if err := b.Repo.AddStep(ctx, b.DB, e); err != nil {
		return nil, fmt.Errorf("AddScenario: %v", err)
	}
	return &dpb.AddStepResponse{Id: e.ID.String}, nil
}

func (b *BDDSuite) MarkStepEnded(ctx context.Context, req *dpb.MarkStepEndedRequest) (*dpb.MarkStepEndedResponse, error) {
	e := new(entities.BDDStep)
	e.ID.Set(req.Id)
	e.Status.Set(req.Status)
	if req.Message == "" {
		e.Message.Set(nil)
	} else {
		e.Message.Set(req.Message)
	}
	e.EndedAt.Set(time.Now().UTC())
	if err := b.Repo.MarkStepEnded(ctx, b.DB, e); err != nil {
		return nil, fmt.Errorf("MarkStepEnded: %v", err)
	}
	return &dpb.MarkStepEndedResponse{}, nil
}

//nolint:errcheck
func toBDDInstanceEntity(r *dpb.AddInstanceRequest) *entities.BDDInstance {
	e := new(entities.BDDInstance)
	database.AllNullEntity(e)
	e.ID.Set(idutil.ULIDNow())
	e.Name.Set(r.Name)
	e.StatusStatistics.Set(r.Stats)
	e.Flavor.Set(r.Flavor)
	e.Tags.Set(r.Tags)
	e.StartedAt.Set(time.Now().UTC())
	return e
}

//nolint:errcheck
func toBDDFeatureEntity(r *dpb.AddFeatureRequest) *entities.BDDFeature {
	e := new(entities.BDDFeature)
	database.AllNullEntity(e)
	e.ID.Set(idutil.ULIDNow())
	e.InstanceID.Set(r.InstanceId)
	e.URI.Set(r.Uri)
	e.Keyword.Set(r.Keyword)
	e.Name.Set(r.Name)
	e.Tags.Set(r.Tags)
	e.StartedAt.Set(time.Now().UTC())
	return e
}

//nolint:errcheck
func toBDDScenarioEntity(r *dpb.AddScenarioRequest) *entities.BDDScenario {
	e := new(entities.BDDScenario)
	database.AllNullEntity(e)
	e.ID.Set(idutil.ULIDNow())
	e.FeatureID.Set(r.FeatureId)
	e.Keyword.Set(r.Keyword)
	e.Name.Set(r.Name)
	e.Steps.Set(r.Steps)
	e.Tags.Set(r.Tags)
	e.StartedAt.Set(time.Now().UTC())
	return e
}

//nolint:errcheck
func toBDDStepEntity(r *dpb.AddStepRequest) *entities.BDDStep {
	e := new(entities.BDDStep)
	database.AllNullEntity(e)
	e.ID.Set(idutil.ULIDNow())
	e.ScenarioID.Set(r.ScenarioId)
	e.Name.Set(r.Name)
	e.URI.Set(r.Uri)
	e.StartedAt.Set(time.Now().UTC())
	return e
}

func (b *BDDSuite) RetrieveSkippedBDDTests(ctx context.Context, req *dpb.RetrieveSkippedBDDTestsRequest) (*dpb.RetrieveSkippedBDDTestsResponse, error) {
	repo := pgtype.Varchar{}
	repo.Set(req.Repository)

	data, err := b.Repo.RetrieveSkippedBDDTestsByRepository(ctx, b.DB, repo)
	if err != nil {
		return nil, fmt.Errorf("b.Repo.RetrieveSkippedBDDTestsByRepository: %v", err)
	}

	resp := make([]*dpb.SkippedBDDTest, 0, len(data))
	for _, e := range data {
		resp = append(resp, &dpb.SkippedBDDTest{
			FeaturePath:  e.FeaturePath.String,
			ScenarioName: e.ScenarioName.String,
			CreatedBy:    e.CreatedBy.String,
		})
	}
	return &dpb.RetrieveSkippedBDDTestsResponse{SkippedBddTests: resp}, nil
}
