package flightRepo

import (
	"context"
	"flight-service/internal/model"
	"flight-service/internal/repository"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type flightRepository struct {
	db repository.QueryRunner
	sq squirrel.StatementBuilderType
}

func NewFlightRepository(db *pgxpool.Pool) repository.FlightRepository {
	return &flightRepository{
		db: db,
	}
}

func (f *flightRepository) WithTx(tx pgx.Tx) repository.FlightRepository {
	return &flightRepository{
		db: tx,
	}
}

func (f *flightRepository) Upsert(ctx context.Context, flight *model.FlightData) error {
	query := f.sq.Insert("flights").
		Columns("flight_number", "departure_date", "aircraft_type", "arrival_date", "passengers_count", "updated_at").
		Values(flight.FlightNumber, flight.DepartureDate, flight.AircraftType, flight.ArrivalDate, flight.PassengersCount, flight.UpdatedAt).
		PlaceholderFormat(squirrel.Dollar).
		Suffix(`
			ON CONFLICT (flight_number, departure_date)
			DO UPDATE SET
				aircraft_type = EXCLUDED.aircraft_type,
				arrival_date = EXCLUDED.arrival_date,
				passengers_count = EXCLUDED.passengers_count,
				updated_at = EXCLUDED.updated_at
		`)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = f.db.Exec(ctx, sql, args...)
	return err
}

func (f *flightRepository) Get(ctx context.Context, flightNumber string, departureDate time.Time) (*model.FlightData, error) {
	query := f.sq.Select("aircraft_type", "arrival_date", "passengers_count", "updated_at").
		From("flights").
		Where(squirrel.And{
			squirrel.Eq{"flight_number": flightNumber},
			squirrel.Eq{"departure_date": departureDate},
		}).PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	flight := &model.FlightData{}
	err = f.db.QueryRow(ctx, sql, args...).Scan(
		&flight.AircraftType,
		&flight.ArrivalDate,
		&flight.PassengersCount,
		&flight.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("flight with number %s and departure date %s not found", flightNumber, departureDate.Format(time.RFC3339))
		}
		return nil, err
	}

	// Устанавливаем значения, которые мы знаем из параметров запроса
	flight.FlightNumber = flightNumber
	flight.DepartureDate = departureDate

	return flight, nil
}
