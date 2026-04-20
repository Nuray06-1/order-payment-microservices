package repository

import (
	"context"
	"database/sql"
	"errors"
	"order-service/internal/domain"
)

type PostgresOrderRepo struct {
	db *sql.DB
}

func NewPostgresOrderRepo(db *sql.DB) *PostgresOrderRepo {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) Create(ctx context.Context, o *domain.Order) error {
	query := `INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		o.ID,
		o.CustomerID,
		o.ItemName,
		o.Amount,
		o.Status,
		o.CreatedAt,
		o.IdempotencyKey,
	)
	return err
}

func (r *PostgresOrderRepo) Update(ctx context.Context, o *domain.Order) error {
	res, err := r.db.ExecContext(ctx, `
    UPDATE orders SET status=$1 WHERE id=$2
    `, o.Status, o.ID)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("order not found")
	}

	return nil
}

func (r *PostgresOrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key
        FROM orders WHERE id=$1
    `, id)

	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt, &o.IdempotencyKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &o, nil
}

func (r *PostgresOrderRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT "id", "customer_id", "item_name", "amount", "status", "created_at", "idempotency_key"
        FROM "orders" WHERE "idempotency_key"=$1
    `, key)

	var o domain.Order
	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.ItemName,
		&o.Amount,
		&o.Status,
		&o.CreatedAt,
		&o.IdempotencyKey,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}
