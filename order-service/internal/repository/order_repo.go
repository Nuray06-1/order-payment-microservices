package repository

import (
	"database/sql"
	"order-service/internal/domain"
)

type PostgresOrderRepo struct {
	db *sql.DB
}

func NewPostgresOrderRepo(db *sql.DB) *PostgresOrderRepo {
	return &PostgresOrderRepo{db: db}
}

func (r *PostgresOrderRepo) Create(o *domain.Order) error {
	_, err := r.db.Exec(`
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, o.CreatedAt,
	)
	return err
}

func (r *PostgresOrderRepo) Update(o *domain.Order) error {
	_, err := r.db.Exec(`
		UPDATE orders SET status=$1 WHERE id=$2
	`, o.Status, o.ID)
	return err
}

func (r *PostgresOrderRepo) GetByID(id string) (*domain.Order, error) {
	row := r.db.QueryRow(`
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders WHERE id=$1
	`, id)

	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
	return &o, err
}
