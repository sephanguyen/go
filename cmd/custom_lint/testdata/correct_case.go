package testdata

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

func correctCase() {
	s := SimpleStruct{}
	dbPool, _ := pgxpool.Connect(context.Background(), "databaseUrl")
	s.DB = dbPool

	rows, _ := s.DB.Query(context.Background(), "query")

	defer rows.Close()
	for rows.Next() {

	}
}
