package repository

import (
	"context"
	"database/sql"
	"payment-service/internal/domain"
)

type PostgresPaymentRepo struct {
	db *sql.DB
}

func NewPostgresPaymentRepo(db *sql.DB) *PostgresPaymentRepo {
	return &PostgresPaymentRepo{db: db}
}

func (r *PostgresPaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO payments (id, order_id, transaction_id, amount, status)
		VALUES ($1, $2, $3, $4, $5)
	`,
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status,
	)
	return err
}
func (r *PostgresPaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, order_id, transaction_id, amount, status
		FROM payments WHERE order_id=$1
	`, orderID)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}
