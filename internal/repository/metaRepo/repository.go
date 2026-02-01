package metaRepo

import (
	"context"
	"flight-service/internal/model"
	"flight-service/internal/repository"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Константы для таблицы flight_meta
const (
	TableFlightMeta     = "flight_meta"
	ColumnID            = "id"
	ColumnFlightNumber  = "flight_number"
	ColumnDepartureDate = "departure_date"
	ColumnStatus        = "status"
	ColumnCreatedAt     = "created_at"
	ColumnProcessedAt   = "processed_at"
)

type metaRepository struct {
	db repository.QueryRunner
	sq squirrel.StatementBuilderType
}

func NewMetaRepository(db *pgxpool.Pool) repository.MetaRepository {
	return &metaRepository{
		db: db,
		sq: squirrel.StatementBuilder,
	}
}

func (r *metaRepository) WithTx(tx pgx.Tx) repository.MetaRepository {
	return &metaRepository{
		db: tx,
		sq: squirrel.StatementBuilder,
	}
}

func (r *metaRepository) Create(ctx context.Context, meta *model.FlightMeta) (int, error) {
	query := r.sq.Insert(TableFlightMeta).
		Columns(ColumnFlightNumber, ColumnDepartureDate, ColumnStatus).
		Values(meta.FlightNumber, meta.DepartureDate, "pending").
		Suffix("RETURNING " + ColumnID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var id int
	err = r.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *metaRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := r.sq.Update(TableFlightMeta).
		Set(ColumnStatus, status).
		Set(ColumnProcessedAt, squirrel.Expr("CURRENT_TIMESTAMP")).
		Where(squirrel.Eq{ColumnID: id}).
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
	baseQuery := r.sq.Select(ColumnID, ColumnFlightNumber, ColumnDepartureDate, ColumnStatus, ColumnCreatedAt, ColumnProcessedAt).
		From(TableFlightMeta).
		Where(squirrel.Eq{ColumnFlightNumber: flightNumber}).
		OrderBy(ColumnCreatedAt + " DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar)

	// Добавляем условие по статусу только если он не пустой
	if status != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{ColumnStatus: status})
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
		var processedAt pgtype.Timestamp

		err = rows.Scan(&meta.ID, &meta.FlightNumber, &meta.DepartureDate, &meta.Status, &meta.CreatedAt, &processedAt)
		if err != nil {
			return nil, 0, err
		}
		if processedAt.Valid {
			meta.ProcessedAt = &processedAt.Time
		} else {
			meta.ProcessedAt = nil
		}

		metas = append(metas, meta)
	}

	// Запрос для получения общего количества
	countQuery := r.sq.Select("COUNT(*)").
		From(TableFlightMeta).
		Where(squirrel.Eq{ColumnFlightNumber: flightNumber}).
		PlaceholderFormat(squirrel.Dollar)

	if status != "" {
		countQuery = countQuery.Where(squirrel.Eq{ColumnStatus: status})
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

// GetStatusCounts возвращает количество записей по каждому статусу
func (r *metaRepository) GetStatusCounts(ctx context.Context) (map[string]int, error) {
	query := r.sq.Select(ColumnStatus, "COUNT(*) as count").
		From(TableFlightMeta).
		GroupBy(ColumnStatus).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statusCounts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, err
		}
		statusCounts[status] = count
	}

	return statusCounts, nil
}
