package metaRepo

import (
	"context"
	"flight-service/internal/model"
	"flight-service/internal/repository"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type metaRepository struct {
	db repository.QueryRunner
	sq squirrel.StatementBuilderType
}

func NewMetaRepository(db *pgxpool.Pool) repository.MetaRepository {
	return &metaRepository{
		db: db,
	}
}

func (r *metaRepository) WithTx(tx pgx.Tx) repository.MetaRepository {
	return &metaRepository{
		db: tx,
	}
}

func (r *metaRepository) Create(ctx context.Context, meta *model.FlightMeta) error {
	query := r.sq.Insert("flight_meta").
		Columns("flight_number", "departure_date", "status").
		Values(meta.FlightNumber, meta.DepartureDate, "pending").
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *metaRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := r.sq.Update("flight_meta").
		Set("status", status).
		Set("processed_at", squirrel.Expr("CURRENT_TIMESTAMP")).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

func (r *metaRepository) GetByFlightNumber(ctx context.Context, flightNumber string, status string, limit int, offset int) ([]*model.FlightMeta, int, error) {
	// Основной запрос на получение данных
	baseQuery := r.sq.Select("id", "flight_number", "departure_date", "status", "created_at", "processed_at").
		From("flight_meta").
		Where(squirrel.Eq{"flight_number": flightNumber}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar)

	// Добавляем условие по статусу только если он не пустой
	if status != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"status": status})
	}

	sql, args, err := baseQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var metas []*model.FlightMeta
	for rows.Next() {
		meta := &model.FlightMeta{}
		err := rows.Scan(&meta.ID, &meta.FlightNumber, &meta.DepartureDate, &meta.Status, &meta.CreatedAt, &meta.ProcessedAt)
		if err != nil {
			return nil, 0, err
		}
		metas = append(metas, meta)
	}

	// Запрос для получения общего количества
	countQuery := r.sq.Select("COUNT(*)").
		From("flight_meta").
		Where(squirrel.Eq{"flight_number": flightNumber}).
		PlaceholderFormat(squirrel.Dollar)

	if status != "" {
		countQuery = countQuery.Where(squirrel.Eq{"status": status})
	}

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int
	err = r.db.QueryRow(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return metas, total, nil
}
