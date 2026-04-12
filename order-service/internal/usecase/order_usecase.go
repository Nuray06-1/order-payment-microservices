package usecase

import (
	"context"
	"errors"
	"time"

	"order-service/internal/domain"
	payment "order-service/pkg/payment"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type OrderRepository interface {
	Create(order *domain.Order) error
	Update(order *domain.Order) error
	GetByID(id string) (*domain.Order, error)
	GetByIdempotencyKey(key string) (*domain.Order, error)
}
type PaymentClient interface {
	ProcessPayment(ctx context.Context, in *payment.PaymentRequest, opts ...grpc.CallOption) (*payment.PaymentResponse, error)
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payResp, err := uc.payment.ProcessPayment(ctx, &payment.PaymentRequest{
		OrderId: order.ID,
		Amount:  float64(order.Amount),
	})

	if err != nil {
		order.Status = "Failed"
		_ = uc.repo.Update(order)
		return order, err
	}

	if payResp.Status == "Authorized" || payResp.Status == "Success" {
		order.Status = "Paid"
	} else {
		order.Status = "Failed"
	}

	_ = uc.repo.Update(order)

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
