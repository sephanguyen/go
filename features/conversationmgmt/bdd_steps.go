package conversationmgmt

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/manabie-com/backend/features/common"
	conv_common "github.com/manabie-com/backend/features/conversationmgmt/common"

	"github.com/cucumber/godog"
	"github.com/ettle/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func mapFeaturesToStepFuncs(parctx *godog.ScenarioContext, conf *common.Config) {
	parctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		uriSplit := strings.Split(sc.Uri, ":")
		uri := uriSplit[0]

		switch uri {
		case "conversationmgmt/test.feature",
			"conversationmgmt/create_agora_user.feature",
			"conversationmgmt/get_agora_app_info.feature",
			"conversationmgmt/create_conversation.feature",
			"conversationmgmt/add_conversation_members.feature",
			"conversationmgmt/update_conversation_info.feature",
			"conversationmgmt/conversation_members.feature",
			"conversationmgmt/delete_message.feature",
			"conversationmgmt/get_conversations_detail.feature":
			InitStepFuncDynamically(parctx, uri, conf)
			ctx = initConversationMgmtCommonState(ctx)
			return ctx, nil

		default:
			return ctx, fmt.Errorf("unknown mapping for files %s", uri)
		}
	})
}

type SuiteConstructor struct{}

type Dependency struct {
	convCommonSuite *conv_common.ConversationMgmtSuite
}

func (c *SuiteConstructor) InitScenarioStepMapping(ctx *godog.ScenarioContext, stepsMapping map[string]interface{}) {
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func InitStepFuncDynamically(parctx *godog.ScenarioContext, uri string, cfg *common.Config) {
	constructor := &SuiteConstructor{}
	parts := strings.Split(uri, "/")
	filename := parts[len(parts)-1]
	featureName := filename[:len(filename)-len(".feature")]
	featureCamelCase := strcase.ToCamel(featureName)
	caser := cases.Title(language.English, cases.NoLower)
	constructMethod := fmt.Sprintf("Init%s", caser.String(featureCamelCase))

	meth := reflect.ValueOf(constructor).MethodByName(constructMethod)
	if !meth.IsValid() {
		panic(fmt.Sprintf("feature %s has no construct method %s", featureName, constructMethod))
	}
	dep := reflect.ValueOf(&Dependency{convCommonSuite: newConversationMgmtCommonSuite(cfg)})
	parCtx := reflect.ValueOf(parctx)
	meth.Call([]reflect.Value{dep, parCtx})
}
