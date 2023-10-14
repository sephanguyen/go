package services

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/grpc/metadata"
)

func signCtx(ctx context.Context) context.Context {
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}
	return metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token)
}

func toBobSchoolPb(school *upb.School) *pb.School {
	if school == nil {
		return nil
	}

	var (
		point    *pb.Point
		district *pb.District
	)
	if school.Point != nil {
		point = &pb.Point{
			Lat:  school.Point.Lat,
			Long: school.Point.Long,
		}
	}
	if school.District != nil {
		district = &pb.District{
			Id:      school.District.Id,
			Name:    school.District.Name,
			Country: pb.Country(school.District.Country),
			City:    toBobCityPb(school.District.City),
		}
	}

	return &pb.School{
		Id:       school.Id,
		Name:     school.Name,
		Country:  pb.Country(school.Country),
		City:     toBobCityPb(school.City),
		District: district,
		Point:    point,
	}
}

func toBobCityPb(city *upb.City) *pb.City {
	if city == nil {
		return nil
	}

	return &pb.City{
		Id:      city.Id,
		Name:    city.Name,
		Country: pb.Country(city.Country),
	}
}
