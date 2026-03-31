package repository

import (
	"database/sql"
	"payment-service/internal/domain"
)

type PostgresPaymentRepo struct {
	db *sql.DB
}

func NewPostgresPaymentRepo(db *sql.DB) *PostgresPaymentRepo {
	return &PostgresPaymentRepo{db: db}
}

func (r *PostgresPaymentRepo) Create(p *domain.Payment) error {
	_, err := r.db.Exec(`
		INSERT INTO payments (id, order_id, transaction_id, amount, status)
		VALUES ($1, $2, $3, $4, $5)
	`,
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status,
	)
	return err
}
func (r *PostgresPaymentRepo) GetByOrderID(orderID string) (*domain.Payment, error) {
	row := r.db.QueryRow(`
		SELECT id, order_id, transaction_id, amount, status
		FROM payments WHERE order_id=$1
	`, orderID)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	return &p, err
}
