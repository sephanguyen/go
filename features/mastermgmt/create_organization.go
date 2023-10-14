package mastermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/organization/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb_ms "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *suite) createNewOrganizationData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.OrganizationID = idutil.ULIDNow()
	stepState.TenantID = idutil.ULIDNow()

	organizationPB := &pb_ms.Organization{
		OrganizationId:   stepState.OrganizationID,
		TenantId:         stepState.TenantID,
		OrganizationName: "organization name test",
		DomainName:       strings.ToLower("domain-test" + idutil.ULIDNow()),
		LogoUrl:          "logo-url",
		CountryCode:      cpb.Country(pb.COUNTRY_VN),
	}
	stepState.Request = &pb_ms.CreateOrganizationRequest{
		Organization: organizationPB,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) locationTypeAndLocationDefaultCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	stepState := StepStateFromContext(ctx)

	ctx, err := s.createOrganizationSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createOrganizationSubscription: %w", err)
	}
	if ctx, err := s.createEventLocationTypeDefaultCreatedSubscription(StepStateToContext(ctx, stepState)); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if ctx, err := s.createEventLocationDefaultCreatedSubscription(StepStateToContext(ctx, stepState)); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*pb_ms.CreateOrganizationRequest)
	stepState.Response, stepState.ResponseErr = pb_ms.NewOrganizationServiceClient(s.MasterMgmtConn).CreateOrganization(contextWithToken(s, ctx), req)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	for {
		select {
		case data := <-stepState.FoundChanForJetStream:
			switch v := data.(type) {
			case *pb_ms.EvtOrganization_CreateOrganization_:
				if req.Organization.OrganizationName == v.CreateOrganization.OrganizationName {
					ctx, err = s.checkExistLocationTypeAndLocationDefault(ctx, req.Organization.OrganizationName)
					if err != nil {
						return StepStateToContext(ctx, stepState), err
					}
					return StepStateToContext(ctx, stepState), nil
				}

			case error:
				return StepStateToContext(ctx, stepState), fmt.Errorf("s.checkLocationTypeDefaultFromPublisher %w", v)
			default:
				continue
			}

		case <-ctx.Done():
			return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
		}
	}
}

func (s *suite) checkExistLocationTypeAndLocationDefault(ctx context.Context, orgName string) (context.Context, error) {
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	stepState := StepStateFromContext(ctx)
	// check location type existing

	queryLocationType := `
				SELECT location_type_id
				FROM location_types
				WHERE name = $1 and display_name = $2
					AND deleted_at IS NULL
				LIMIT 1`
	var locationTypeID string

	for {
		queryLocationTypeErr := s.BobDBTrace.DB.QueryRow(ctx, queryLocationType, domain.DefaultLocationType, orgName).Scan(&locationTypeID)
		if queryLocationTypeErr != nil && queryLocationTypeErr != pgx.ErrNoRows {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query location_types: %s", queryLocationTypeErr)
		}
		if locationTypeID != "" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	stepState.LocationTypeIDs = []string{locationTypeID}
	// check location type existing
	var locationID string
	queryLocation := `
			SELECT location_id
			FROM locations
			WHERE name = $1 and location_type = $2
				AND deleted_at IS NULL`
	for {
		queryLocationErr := s.BobDBTrace.QueryRow(ctx, queryLocation, orgName, locationTypeID).Scan(&locationID)
		if queryLocationErr != nil && queryLocationErr != pgx.ErrNoRows {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query locations: %s", queryLocationErr)
		}
		if locationID != "" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	stepState.LocationIDs = []string{locationID}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) checkEventPublisher(ctx context.Context, subject string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch subject {
	case constants.SubjectSyncLocationTypeUpserted:
		return s.checkLocationTypeDefaultFromPublisher(ctx)
	case constants.SubjectSyncLocationUpserted:
		return s.checkLocationDefaultFromPublisher(ctx)
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("type message is not available")
	}
}

func (s *suite) checkLocationTypeDefaultFromPublisher(ctx context.Context) (context.Context, error) {
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	stepState := StepStateFromContext(ctx)

	for {
		select {
		case data := <-stepState.FoundChanForJetStream:
			switch v := data.(type) {
			case *npb.EventSyncLocationType_LocationType:
				if v.LocationTypeId == stepState.LocationTypeIDs[0] {
					return StepStateToContext(ctx, stepState), nil
				}
			case error:
				return StepStateToContext(ctx, stepState), fmt.Errorf("s.checkLocationTypeDefaultFromPublisher %w", v)
			default:
				continue
			}
		case <-ctx.Done():
			return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
		}
	}
}
func (s *suite) checkLocationDefaultFromPublisher(ctx context.Context) (context.Context, error) {
	timer := time.NewTimer(time.Minute * 5)
	defer timer.Stop()
	stepState := StepStateFromContext(ctx)

	for {
		select {
		case data := <-stepState.FoundChanForJetStream:
			switch v := data.(type) {
			case *npb.EventSyncLocation_Location:
				if v.LocationId == stepState.LocationIDs[0] {
					return StepStateToContext(ctx, stepState), nil
				}

			case error:
				return StepStateToContext(ctx, stepState), fmt.Errorf("s.checkLocationDefaultFromPublisher %w", v)
			default:
				continue
			}

		case <-ctx.Done():
			return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
		}
	}
}

func (s *suite) createEventLocationTypeDefaultCreatedSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.AckWait(30 * time.Second),
			nats.MaxDeliver(10),
			nats.Bind(constants.StreamSyncLocationTypeUpserted, constants.DurableSyncLocationTypeUpsertedOrgCreation),
			nats.DeliverSubject(constants.DeliverSyncLocationTypeUpsertedOrgCreation),
		},
	}

	handleLocationTypeCreated := func(ctx context.Context, data []byte) (bool, error) {
		eventLocationType := &npb.EventSyncLocationType{}
		if err := proto.Unmarshal(data, eventLocationType); err != nil {
			stepState.FoundChanForJetStream <- fmt.Errorf("proto.Unmarshal(data, eventLocationType) %w", err)
			return false, err
		}

		if eventLocationType.LocationTypes[0].Name == domain.DefaultLocationType {
			stepState.FoundChanForJetStream <- eventLocationType.LocationTypes[0]
		}
		return true, nil
	}

	subs, err := s.JSM.QueueSubscribe(constants.SubjectSyncLocationTypeUpserted, constants.QueueSyncLocationTypeUpsertedOrgCreation, opts, handleLocationTypeCreated)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createEventLocationTypeDefaultCreatedSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createEventLocationDefaultCreatedSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{

			nats.ManualAck(),
			nats.AckWait(30 * time.Second),
			nats.MaxDeliver(10),
			nats.Bind(constants.StreamSyncLocationUpserted, constants.DurableSyncLocationUpsertedOrgCreation),
			nats.DeliverSubject(constants.DeliverSyncLocationUpsertedOrgCreation),
		},
	}

	var subs *nats.Subscription

	handleLocationCreated := func(ctx context.Context, data []byte) (bool, error) {
		eventLocation := &npb.EventSyncLocation{}
		if err := proto.Unmarshal(data, eventLocation); err != nil {
			stepState.FoundChanForJetStream <- fmt.Errorf("proto.Unmarshal(data, eventLocation) %w", err)
			return false, err
		}
		stepState.FoundChanForJetStream <- eventLocation.Locations[0]
		return true, nil
	}

	subs, err := s.JSM.QueueSubscribe(constants.SubjectSyncLocationUpserted, constants.QueueSyncLocationUpsertedOrgCreation, opts, handleLocationCreated)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createEventLocationDefaultCreatedSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

const (
	OrganizationID   = "organizationID"
	TenantID         = "tenantID"
	OrganizationName = "organization name"
	LogoUrl          = "logo url"
	DomainName       = "domain name"
	CountryCode      = "country code"
)

func (s *suite) newOrganizationData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.addOrganizationDataToCreateOrganizationReq(ctx)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addOrganizationDataToCreateOrganizationReq(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	stepState.OrganizationID = idutil.ULIDNow()
	stepState.TenantID = idutil.ULIDNow()

	organizationPB := &pb_ms.Organization{
		OrganizationId:   stepState.OrganizationID,
		TenantId:         stepState.TenantID,
		OrganizationName: "organization name test",
		DomainName:       strings.ToLower("domain-test" + idutil.ULIDNow()),
		LogoUrl:          "logo-url",
		CountryCode:      cpb.Country(pb.COUNTRY_VN),
	}

	stepState.Request = &pb_ms.CreateOrganizationRequest{
		Organization: organizationPB,
	}
	return StepStateToContext(ctx, stepState)
}

func (s *suite) createNewOrganization(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	ctx, err = s.createOrganizationSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createOrganizationSubscription: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb_ms.NewOrganizationServiceClient(s.MasterMgmtConn).CreateOrganization(contextWithToken(s, ctx), stepState.Request.(*pb_ms.CreateOrganizationRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) newOrganizationWereCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	ctx, err := s.validateCreateOrganizationResponse(ctx)
	if err != nil {
		return ctx, err
	}

	select {
	case <-stepState.FoundChanForJetStream:
		return s.validateOrganizationInfo(ctx)
	case <-ctx.Done():
		return ctx, fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) validateCreateOrganizationResponse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb_ms.CreateOrganizationRequest)
	resp := stepState.Response.(*pb_ms.CreateOrganizationResponse)
	switch {
	case req.Organization.OrganizationId != resp.Organization.OrganizationId:
		return ctx, fmt.Errorf(`validateCreateOrganizationResponse: expected response "organization_id": %v but actual is %v`, req.Organization.OrganizationId, resp.Organization.OrganizationId)
	case req.Organization.OrganizationName != resp.Organization.OrganizationName:
		return ctx, fmt.Errorf(`validateCreateOrganizationResponse: expected response "organization_name": %v but actual is %v`, req.Organization.OrganizationName, resp.Organization.OrganizationName)
	case req.Organization.TenantId != resp.Organization.TenantId:
		return ctx, fmt.Errorf(`validateCreateOrganizationResponse: expected response "tenant_id": %v but actual is %v`, req.Organization.TenantId, resp.Organization.TenantId)
	case req.Organization.CountryCode != resp.Organization.CountryCode:
		return ctx, fmt.Errorf(`validateCreateOrganizationResponse: expected response "country": %v but actual is %v`, req.Organization.CountryCode.String(), resp.Organization.CountryCode.String())
	case req.Organization.DomainName != resp.Organization.DomainName:
		return ctx, fmt.Errorf(`validateCreateOrganizationResponse: expected response "domain_name": %v but actual is %v`, req.Organization.DomainName, resp.Organization.DomainName)
	case req.Organization.LogoUrl != resp.Organization.LogoUrl:
		return ctx, fmt.Errorf(`validateCreateOrganizationResponse: expected response "logo_url": %v but actual is %v`, req.Organization.LogoUrl, resp.Organization.LogoUrl)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateOrganizationInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedCtx(ctx)
	req := stepState.Request.(*pb_ms.CreateOrganizationRequest)
	stmt :=
		`
		SELECT 
			organization_id,
			name,
			tenant_id,
			domain_name,
			logo_url,
			country
		FROM
			organizations
		WHERE 
			organization_id=$1
		`
	row := s.BobDBTrace.QueryRow(
		ctx,
		stmt,
		stepState.Response.(*pb_ms.CreateOrganizationResponse).Organization.OrganizationId,
	)

	organization := &entities.Organization{}
	if err := row.Scan(
		&organization.ID,
		&organization.Name,
		&organization.TenantID,
		&organization.DomainName,
		&organization.LogoURL,
		&organization.Country,
	); err != nil {
		return ctx, err
	}
	switch {
	case req.Organization.OrganizationId != organization.ID.String:
		return ctx, fmt.Errorf(`validateOrganizationInfo: expected inserted "organization_id": %v but actual is %v`, req.Organization.OrganizationId, organization.ID)
	case req.Organization.TenantId != organization.TenantID.String:
		return ctx, fmt.Errorf(`validateOrganizationInfo: expected inserted "tenant_id": %v but actual is %v`, req.Organization.TenantId, organization.TenantID)
	case req.Organization.OrganizationName != organization.Name.String:
		return ctx, fmt.Errorf(`validateOrganizationInfo: expected inserted "organization_name": %v but actual is %v`, req.Organization.OrganizationName, organization.Name)
	case req.Organization.CountryCode.String() != organization.Country.String:
		return ctx, fmt.Errorf(`validateOrganizationInfo: expected inserted "country": %v but actual is %v`, req.Organization.CountryCode.String(), organization.Country.String)
	case req.Organization.DomainName != organization.DomainName.String:
		return ctx, fmt.Errorf(`validateOrganizationInfo: expected inserted "domain_name": %v but actual is %v`, req.Organization.DomainName, organization.DomainName.String)
	case req.Organization.LogoUrl != organization.LogoURL.String:
		return ctx, fmt.Errorf(`validateOrganizationInfo: expected inserted "logo_url": %v but actual is %v`, req.Organization.LogoUrl, organization.LogoURL)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createOrganizationSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}

	handleCreateOrganization := func(ctx context.Context, data []byte) (bool, error) {
		evtOrganization := &pb_ms.EvtOrganization{}
		if err := proto.Unmarshal(data, evtOrganization); err != nil {
			return false, err
		}

		switch msg := evtOrganization.Message.(type) {
		case *pb_ms.EvtOrganization_CreateOrganization_:
			stepState.FoundChanForJetStream <- msg
		}

		return true, nil
	}

	subs, err := s.JSM.Subscribe(constants.SubjectOrganizationCreated, opts, handleCreateOrganization)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("createOrganizationSubscription: s.JSM.Subscribe: %w", err)
	}

	stepState.Subs = append(stepState.Subs, subs.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) organizationDataHasEmptyOrInvalid(ctx context.Context, fields string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.addOrganizationDataToCreateOrganizationReq(ctx)

	req := stepState.Request.(*pb_ms.CreateOrganizationRequest)
	switch fields {
	case OrganizationID:
		req.Organization.OrganizationId = ""
	case TenantID:
		req.Organization.TenantId = ""
	case OrganizationName:
		req.Organization.OrganizationName = ""
	case LogoUrl:
		req.Organization.LogoUrl = ""
	case DomainName:
		req.Organization.DomainName = ""
	case CountryCode:
		req.Organization.CountryCode = cpb.Country(pb.COUNTRY_VN)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) organizationDataHasInvalidDomainName(ctx context.Context, domain string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = s.addOrganizationDataToCreateOrganizationReq(ctx)

	req := stepState.Request.(*pb_ms.CreateOrganizationRequest)
	req.Organization.DomainName = domain

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCannotCreateOrganization(ctx context.Context, caller string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, caller)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx = s.addOrganizationDataToCreateOrganizationReq(ctx)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb_ms.NewOrganizationServiceClient(s.MasterMgmtConn).CreateOrganization(contextWithToken(s, ctx), stepState.Request.(*pb_ms.CreateOrganizationRequest))
	return StepStateToContext(ctx, stepState), nil
}
