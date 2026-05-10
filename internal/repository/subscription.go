package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"subscription-service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionRepository struct{ db *pgxpool.Pool }

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s model.Subscription) (model.Subscription, error) {
	q := `INSERT INTO subscriptions(service_name, price, user_id, start_date, end_date)
		  VALUES ($1,$2,$3,$4,$5)
		  RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`
	return scanOne(r.db.QueryRow(ctx, q, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate))
}

func (r *SubscriptionRepository) Get(ctx context.Context, id int64) (model.Subscription, error) {
	q := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE id=$1`
	return scanOne(r.db.QueryRow(ctx, q, id))
}

func (r *SubscriptionRepository) List(ctx context.Context, f model.ListFilter) ([]model.Subscription, error) {
	args := []any{}
	where := []string{"deleted_at IS NULL"}
	if f.UserID != nil {
		args = append(args, *f.UserID)
		where = append(where, fmt.Sprintf("user_id=$%d", len(args)))
	}
	if f.ServiceName != nil {
		args = append(args, *f.ServiceName)
		where = append(where, fmt.Sprintf("service_name ILIKE $%d", len(args)))
	}
	args = append(args, f.Limit, f.Offset)
	q := fmt.Sprintf(`SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions
		WHERE %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, strings.Join(where, " AND "), len(args)-1, len(args))
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Subscription, 0)
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *SubscriptionRepository) Update(ctx context.Context, s model.Subscription) (model.Subscription, error) {
	q := `UPDATE subscriptions SET service_name=$2, price=$3, user_id=$4, start_date=$5, end_date=$6, updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`
	return scanOne(r.db.QueryRow(ctx, q, s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate))
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `UPDATE subscriptions SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepository) Sum(ctx context.Context, f model.SumFilter) (int, error) {
	q := `SELECT COALESCE(SUM(s.price),0)::int
		FROM subscriptions s
		JOIN generate_series($1::date, $2::date, interval '1 month') m(month_start)
		  ON s.start_date <= m.month_start AND (s.end_date IS NULL OR s.end_date >= m.month_start)
		WHERE s.deleted_at IS NULL
		  AND ($3::uuid IS NULL OR s.user_id = $3)
		  AND ($4::text IS NULL OR s.service_name ILIKE $4)`
	var userID any
	if f.UserID != nil {
		userID = *f.UserID
	}
	var service any
	if f.ServiceName != nil {
		service = *f.ServiceName
	}
	var total int
	if err := r.db.QueryRow(ctx, q, f.From, f.To, userID, service).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func scanOne(row pgx.Row) (model.Subscription, error) {
	var s model.Subscription
	err := row.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Subscription{}, ErrNotFound
	}
	return s, err
}

func ParseUUIDPtr(v string) (*uuid.UUID, error) {
	if v == "" {
		return nil, nil
	}
	u, err := uuid.Parse(v)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func MonthStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
