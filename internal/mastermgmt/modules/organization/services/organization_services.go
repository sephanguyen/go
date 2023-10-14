package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/organization/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type OrganizationService struct {
	pb.UnimplementedOrganizationServiceServer
	DB  database.Ext
	JSM nats.JetStreamManagement

	OrganizationRepo interface {
		Create(context.Context, database.QueryExecer, *entities.Organization) error
	}
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.CreateOrganizationResponse, error) {
	//zapLogger := ctxzap.Extract(ctx).Sugar()

	if err := validCreateRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resourcePath, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}
	var organizationPB *pb.Organization

	// Insert new organization
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		organization, err := organizationPbToOrganizationEntity(int32(resourcePath), req)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err := s.OrganizationRepo.Create(ctx, tx, organization); err != nil {
			return errorx.ToStatusError(err)
		}
		organizationPB = organizationToOrganizationPBInCreateOrganizationResponse(organization)

		organizationEvents := newCreateOrganizationEvents(organization)

		if err = s.publishOrganizationEvent(ctx, organizationEvents...); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	response := &pb.CreateOrganizationResponse{
		Organization: organizationPB,
	}
	return response, nil
}

func organizationPbToOrganizationEntity(resourcePath int32, req *pb.CreateOrganizationRequest) (*entities.Organization, error) {
	organizationEnt := &entities.Organization{}
	database.AllNullEntity(organizationEnt)
	if err := multierr.Combine(
		organizationEnt.ID.Set(req.Organization.OrganizationId),
		organizationEnt.TenantID.Set(req.Organization.TenantId),
		organizationEnt.Name.Set(req.Organization.OrganizationName),
		organizationEnt.ResourcePath.Set(fmt.Sprint(resourcePath)),
		organizationEnt.DomainName.Set(req.Organization.DomainName),
		organizationEnt.LogoURL.Set(req.Organization.LogoUrl),
		organizationEnt.Country.Set(req.Organization.CountryCode),
	); err != nil {
		return nil, err
	}
	return organizationEnt, nil
}

func organizationToOrganizationPBInCreateOrganizationResponse(organization *entities.Organization) *pb.Organization {
	organizationPB := &pb.Organization{
		OrganizationId:   organization.ID.String,
		TenantId:         organization.TenantID.String,
		OrganizationName: organization.Name.String,
		DomainName:       organization.DomainName.String,
		LogoUrl:          organization.LogoURL.String,
		CountryCode:      cpb.Country(cpb.Country_value[organization.Country.String]),
	}
	return organizationPB
}

func newCreateOrganizationEvents(organizations ...*entities.Organization) []*pb.EvtOrganization {
	createOrganizationEvents := make([]*pb.EvtOrganization, 0, len(organizations))

	for _, organization := range organizations {
		createOrganizationEvent := &pb.EvtOrganization{
			Message: &pb.EvtOrganization_CreateOrganization_{
				CreateOrganization: &pb.EvtOrganization_CreateOrganization{
					OrganizationId:   organization.ID.String,
					TenantId:         organization.TenantID.String,
					OrganizationName: organization.Name.String,
					DomainName:       organization.DomainName.String,
				},
			},
		}
		createOrganizationEvents = append(createOrganizationEvents, createOrganizationEvent)
	}
	return createOrganizationEvents
}

func (s *OrganizationService) publishOrganizationEvent(ctx context.Context, organizationEvents ...*pb.EvtOrganization) error {
	for _, event := range organizationEvents {
		data, err := proto.Marshal(event)
		if err != nil {
			return err
		}
		_, err = s.JSM.TracedPublish(ctx, "publishOrganizationEvent", constants.SubjectOrganizationCreated, data)
		if err != nil {
			return fmt.Errorf("publishOrganizationEvent: s.JSM.Publish failed")
		}
	}

	return nil
}

func validCreateRequest(req *pb.CreateOrganizationRequest) error {
	switch {
	case req.Organization == nil:
		return errors.New("organization invalid params")
	case req.Organization.OrganizationId == "":
		return errors.New("organization id cannot be empty")
	case req.Organization.TenantId == "":
		return errors.New("tenant id cannot be empty")
	case req.Organization.OrganizationName == "":
		return errors.New("organization name cannot be empty")
	}

	if _, found := cpb.Country_name[int32(req.Organization.CountryCode)]; !found {
		return errors.New("organization invalid country code")
	}

	return nil
}
