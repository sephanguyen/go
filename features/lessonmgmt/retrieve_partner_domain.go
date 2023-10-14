package lessonmgmt

import (
	"context"
	"fmt"
	"net/url"

	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) userGetPartnerDomain(ctx context.Context, domainType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	domainTypes := map[string]lpb.DomainType{
		"Bo":      lpb.DomainType_DOMAIN_TYPE_BO,
		"Teacher": lpb.DomainType_DOMAIN_TYPE_TEACHER,
		"Learner": lpb.DomainType_DOMAIN_TYPE_LEARNER,
	}
	if _, ok := domainTypes[domainType]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("domain type %s is not valid", domainType)
	}

	req := &lpb.GetPartnerDomainRequest{
		Type: domainTypes[domainType],
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewLessonReportReaderServiceClient(s.LessonMgmtConn).RetrievePartnerDomain(contextWithToken(s, ctx), req)

	resp := stepState.Response.(*lpb.GetPartnerDomainResponse)
	val, err := url.Parse(resp.Domain)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot parse domain %s", resp.Domain)
	}
	if val.Scheme != "https" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("domain %s need ssl", resp.Domain)
	}
	return StepStateToContext(ctx, stepState), nil
}
