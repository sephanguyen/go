package types

import (
	"database/sql/driver"
	"fmt"
)

// Point represents an x,y coordinate in EPSG:4326 for PostGIS.
type Point struct {
	Data   [2]float64
	IsNull bool
}

func (p *Point) String() string {
	if p.IsNull {
		return "NULL"
	}
	return fmt.Sprintf("(%v, %v)", p.Data[0], p.Data[1])
}

// Value impl.
func (p Point) Value() (driver.Value, error) {
	value := p.String()
	if value == "NULL" {
		return nil, nil
	}
	return p.String(), nil
}
