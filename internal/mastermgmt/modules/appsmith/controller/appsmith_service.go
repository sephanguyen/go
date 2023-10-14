package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/infrastructure"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppsmithService struct {
	Alert                 alert.SlackFactory
	Env                   string
	Org                   string
	AppsmithsQueryHandler queries.AppsmithQueryHandler
}

func NewAppsmithService(
	mongoDB *mongo.Database,
	newPageRepo infrastructure.NewPageRepo,
	env string,
	org string,
	alert alert.SlackFactory,
) *AppsmithService {
	return &AppsmithService{
		Env: env,
		Org: org,
		AppsmithsQueryHandler: queries.AppsmithQueryHandler{
			DB:          mongoDB,
			NewPageRepo: newPageRepo,
		},
		Alert: alert,
	}
}

func (app *AppsmithService) GetPageInfoBySlug(ctx context.Context, req *mpb.GetPageInfoBySlugRequest) (res *mpb.GetPageInfoBySlugResponse, err error) {
	page, err := app.AppsmithsQueryHandler.GetPageInfoBySlug(ctx, req.Slug, req.ApplicationId, req.BranchName)
	if err != nil {
		errAlert := app.Alert.Send(alert.Payload{
			Text: fmt.Sprintf("*ALERT DETAILS:*\n*ENV-ORG:* `%s-%s` \n*Payload:* \n*- Slug:* `%s`\n*- ApplicationId:* `%s`\n*- BranchName:* `%s` \n*Error:* %s \n", app.Env, app.Org, req.Slug, req.ApplicationId, req.BranchName, err.Error()),
		})
		if errAlert != nil {
			return &mpb.GetPageInfoBySlugResponse{}, status.Error(codes.Internal, errAlert.Error())
		}
		return &mpb.GetPageInfoBySlugResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &mpb.GetPageInfoBySlugResponse{
		Id:            page.ID,
		ApplicationId: page.ApplicationID,
	}, nil
}

func (app *AppsmithService) GetSchemaByWorkspaceID(_ context.Context, req *mpb.GetSchemaNameByWorkspaceIDRequest) (res *mpb.GetSchemaNameByWorkspaceIDResponse, err error) {
	configs := map[string]map[string]string{
		"local": {
			"635f5299b3ce396b06d52db8": "architecture",
		},
		"stag": {
			"635f5299b3ce396b06d52db8": "architecture",
		},
	}

	return &mpb.GetSchemaNameByWorkspaceIDResponse{
		Schema: configs[app.Env][req.GetWorkspaceId()],
	}, nil
}
