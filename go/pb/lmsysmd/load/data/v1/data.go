package data

import (
	"context"
	"sync"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DataService struct {
	dbOnce sync.Once
	db     *pgxpool.Pool
}

func (ds *DataService) BatchCreateData(
	ctx context.Context,
	req *connect.Request[datav1.BatchCreateDataRequest],
) *pgxpool.Pool {

	// Iterate through google sheet
	// convert to proto message for field validation
	// if anything fails - roll back the entire thing
	return ds.db
}
