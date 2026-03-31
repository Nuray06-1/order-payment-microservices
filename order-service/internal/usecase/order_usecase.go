package usecase

import (
	"errors"
	"time"

	"order-service/internal/domain"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(order *domain.Order) error
	Update(order *domain.Order) error
	GetByID(id string) (*domain.Order, error)
	GetByIdempotencyKey(key string) (*domain.Order, error)
}

type PaymentClient interface {
	ProcessPayment(orderID string, amount int64) (string, error)
}

type OrderUseCase struct {
	repo    OrderRepository
	payment PaymentClient
}

func NewOrderUseCase(r OrderRepository, p PaymentClient) *OrderUseCase {
	return &OrderUseCase{repo: r, payment: p}
}

func (uc *OrderUseCase) CreateOrder(customerID, itemName string, amount int64, key string) (*domain.Order, error) {
	existing, err := uc.repo.GetByIdempotencyKey(key)
	if err == nil && existing.ID != "" {
		return existing, nil
	}

	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         "Pending",
		CreatedAt:      time.Now(),
		IdempotencyKey: key,
	}

	err = uc.repo.Create(order)
	if err != nil {
		return nil, err
	}

	status, err := uc.payment.ProcessPayment(order.ID, order.Amount)
	if err != nil {
		order.Status = "Failed"
		uc.repo.Update(order)
		return order, err
	}

	if status == "Authorized" {
		order.Status = "Paid"
	} else {
		order.Status = "Failed"
	}

	uc.repo.Update(order)

	return order, nil
}
func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return uc.repo.GetByID(id)
}
func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}

	if order.Status == "Paid" {
		return errors.New("cannot cancel paid order")
	}

	order.Status = "Cancelled"
	return uc.repo.Update(order)
}
