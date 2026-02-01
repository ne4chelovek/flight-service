package flightRepo

import (
	"context"
	"errors"
	"flight-service/internal/model"
	"flight-service/internal/repository"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// Константы для таблицы flights
const (
	TableFlights          = "flights"
	ColumnFlightNumber    = "flight_number"
	ColumnDepartureDate   = "departure_date"
	ColumnAircraftType    = "aircraft_type"
	ColumnArrivalDate     = "arrival_date"
	ColumnPassengersCount = "passengers_count"
	ColumnUpdatedAt       = "updated_at"
)

type flightRepository struct {
	db repository.QueryRunner
	sq squirrel.StatementBuilderType
}

func NewFlightRepository(db *pgxpool.Pool) repository.FlightRepository {
	return &flightRepository{
		db: db,
		sq: squirrel.StatementBuilder,
	}
}

func (f *flightRepository) WithTx(tx pgx.Tx) repository.FlightRepository {
	return &flightRepository{
		db: tx,
		sq: squirrel.StatementBuilder,
	}
}

func (f *flightRepository) Upsert(ctx context.Context, flight *model.FlightData) error {
	query := f.sq.Insert(TableFlights).
		Columns(ColumnFlightNumber, ColumnDepartureDate, ColumnAircraftType, ColumnArrivalDate, ColumnPassengersCount, ColumnUpdatedAt).
		Values(flight.FlightNumber, flight.DepartureDate, flight.AircraftType, flight.ArrivalDate, flight.PassengersCount, flight.UpdatedAt).
		PlaceholderFormat(squirrel.Dollar).
		Suffix(fmt.Sprintf(`
			ON CONFLICT (%s, %s)
			DO UPDATE SET
				%s = EXCLUDED.%s,
				%s = EXCLUDED.%s,
				%s = EXCLUDED.%s,
				%s = EXCLUDED.%s
		`, ColumnFlightNumber, ColumnDepartureDate,
			ColumnAircraftType, ColumnAircraftType,
			ColumnArrivalDate, ColumnArrivalDate,
			ColumnPassengersCount, ColumnPassengersCount,
			ColumnUpdatedAt, ColumnUpdatedAt))

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = f.db.Exec(ctx, sql, args...)
	return err
}

func (f *flightRepository) Get(ctx context.Context, flightNumber string, departureDate time.Time) (*model.FlightData, error) {
	query := f.sq.Select(ColumnAircraftType, ColumnArrivalDate, ColumnPassengersCount, ColumnUpdatedAt).
		From(TableFlights).
		Where(squirrel.And{
			squirrel.Eq{ColumnFlightNumber: flightNumber},
			squirrel.Eq{ColumnDepartureDate: departureDate},
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("flight with number %s and departure date %s not found", flightNumber, departureDate.Format(time.RFC3339))
		}
		return nil, err
	}

	// Устанавливаем значения, которые мы знаем из параметров запроса
	flight.FlightNumber = flightNumber
	flight.DepartureDate = departureDate

	return flight, nil
}
