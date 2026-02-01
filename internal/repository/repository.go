package repository

import (
	"context"
	"flight-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type QueryRunner interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type MetaRepository interface {
	WithTx(tx pgx.Tx) MetaRepository
	Create(ctx context.Context, meta *model.FlightMeta) (int, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	GetStatusCounts(ctx context.Context) (map[string]int, error)
	GetByFlightNumber(ctx context.Context, flightNumber string, status string, limit int, offset int) ([]*model.FlightMeta, int, error)
}

type FlightRepository interface {
	WithTx(tx pgx.Tx) FlightRepository
	Upsert(ctx context.Context, flight *model.FlightData) error
	Get(ctx context.Context, flightNumber string, departureDate time.Time) (*model.FlightData, error)
}
