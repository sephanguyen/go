package services

import (
	"context"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type SchoolService struct {
	DB       database.Ext
	UserRepo interface {
		UserGroup(ctx context.Context, db database.QueryExecer, id pgtype.Text) (string, error)
	}
	SchoolRepo interface {
		Import(context.Context, database.QueryExecer, []*entities_bob.School) error
		RetrieveCities(ctx context.Context, db database.QueryExecer, country string) ([]*entities_bob.City, error)
		RetrieveDistricts(ctx context.Context, db database.QueryExecer, country string, cityID int32) ([]*entities_bob.District, error)
		RetrieveSchools(ctx context.Context, db database.QueryExecer, country string, cityID, districtID int32, includeSystemSchools bool, schoolIDs []int32) ([]*entities_bob.School, error)
	}
	SchoolConfigRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (map[pgtype.Int4]*entities_bob.SchoolConfig, error)
	}
	ClassRepo interface {
		FindJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entities_bob.Class, error)
	}
	ConfigRepo interface {
		Retrieve(ctx context.Context, db database.QueryExecer, country pgtype.Text, group pgtype.Text, keys pgtype.TextArray) ([]*entities_bob.Config, error)
	}
	CourseClassRepo interface {
		Find(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (mapCourseIDsByClassID map[pgtype.Int4]pgtype.TextArray, err error)
	}
}

func toCityEntity(cp *pb.City) *entities_bob.City {
	if cp == nil {
		return nil
	}
	c := new(entities_bob.City)
	c.Name.Set(cp.Name)
	c.Country.Set(cp.Country.String())
	return c
}

func toDistrictEntity(dp *pb.District) *entities_bob.District {
	if dp == nil {
		return nil
	}
	d := new(entities_bob.District)
	d.Name.Set(dp.Name)
	d.Country.Set(dp.Country.String())
	d.City = toCityEntity(dp.City)
	return d
}

func toSchoolEntity(sp *pb.School) *entities_bob.School {
	s := new(entities_bob.School)
	database.AllNullEntity(s)
	s.Name.Set(sp.Name)
	s.Country.Set(sp.Country.String())

	// sp.City.Id == 0 when:
	//   - admin imports schools from csv
	// sp.City.Id != 0 when:
	//   - student selects existed school when register
	//   - student inputs new school when register, but select existed city
	if sp.City != nil {
		if sp.City.Id == 0 {
			s.City = toCityEntity(sp.City)
		} else {
			s.CityID.Set(sp.City.Id)
		}
	}

	// sp.District.Id == 0 when:
	//   - admin imports schools from csv
	// sp.District.Id != 0 when:
	//   - student selects existed school when register
	//   - student inputs new school when register, but select existed district
	if sp.District != nil {
		if sp.District.Id == 0 {
			s.District = toDistrictEntity(sp.District)
		} else {
			s.DistrictID.Set(sp.District.Id)
		}
	}
	p := pgtype.Point{}
	if sp.Point == nil {
		p.Status = pgtype.Null
	} else {
		p.Status = pgtype.Present
		p.P = pgtype.Vec2{
			X: sp.Point.Lat,
			Y: sp.Point.Long,
		}
	}
	s.Point = p

	return s
}

func toCityPb(ce *entities_bob.City) *pb.City {
	if ce == nil {
		return nil
	}
	c := &pb.City{
		Id:      ce.ID.Int,
		Name:    ce.Name.String,
		Country: pb.Country(pb.Country_value[ce.Country.String]),
	}
	return c
}

func (s *SchoolService) RetrieveCities(ctx context.Context, req *pb.RetrieveCitiesRequest) (*pb.RetrieveCitiesResponse, error) {
	cities, err := s.SchoolRepo.RetrieveCities(ctx, s.DB, req.Country.String())
	if err != nil {
		return nil, errors.Wrapf(err, "s.SchoolRepo.RetrieveCities: country: %q", req.Country.String())
	}
	ret := make([]*pb.City, 0, len(cities))
	for _, c := range cities {
		ret = append(ret, toCityPb(c))
	}
	return &pb.RetrieveCitiesResponse{Cities: ret}, nil
}

func toDistrictPb(de *entities_bob.District) *pb.District {
	if de == nil {
		return nil
	}
	d := &pb.District{
		Id:      de.ID.Int,
		Name:    de.Name.String,
		Country: pb.Country(pb.Country_value[de.Country.String]),
	}
	if de.City == nil {
		d.City = &pb.City{
			Id: de.CityID.Int,
		}
	} else {
		d.City = toCityPb(de.City)
	}
	return d
}

func toSchoolPb(se *entities_bob.School) *pb.School {
	s := &pb.School{
		Id:      se.ID.Int,
		Name:    se.Name.String,
		Country: pb.Country(pb.Country_value[se.Country.String]),
	}
	if se.City == nil {
		s.City = &pb.City{
			Id: se.CityID.Int,
		}
	} else {
		s.City = toCityPb(se.City)
	}
	if se.District == nil {
		s.District = &pb.District{
			Id: se.DistrictID.Int,
		}
	} else {
		s.District = toDistrictPb(se.District)
	}
	if se.Point.Status == pgtype.Present {
		s.Point = &pb.Point{
			Lat:  se.Point.P.X,
			Long: se.Point.P.Y,
		}
	}
	return s
}
