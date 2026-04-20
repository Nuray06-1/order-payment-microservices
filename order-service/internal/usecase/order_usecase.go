package usecase

import (
	"context"
	"errors"
	"time"

	"order-service/internal/domain"

	pb "github.com/Nuray06-1/proto-generated/payment"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	StatusPending   = "PENDING"
	StatusPaid      = "PAID"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELLED"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	Update(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
}

type PaymentClient interface {
	ProcessPayment(ctx context.Context, in *pb.PaymentRequest, opts ...grpc.CallOption) (*pb.PaymentResponse, error)
}

type OrderUseCase struct {
	repo    OrderRepository
	payment PaymentClient
}

func NewOrderUseCase(r OrderRepository, p PaymentClient) *OrderUseCase {
	return &OrderUseCase{
		repo:    r,
		payment: p,
	}
}

func (uc *OrderUseCase) CreateOrder(
	ctx context.Context,
	customerID string,
	itemName string,
	amount int64,
	key string,
) (*domain.Order, error) {

	if key == "" {
		return nil, errors.New("idempotency key required")
	}

	if customerID == "" || itemName == "" {
		return nil, errors.New("customer_id and item_name required")
	}

	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	existing, err := uc.repo.GetByIdempotencyKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         StatusPending,
		CreatedAt:      time.Now(),
		IdempotencyKey: key,
	}

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := uc.payment.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId: order.ID,
		Amount:  order.Amount,
	})

	if err != nil {
		order.Status = StatusFailed
		_ = uc.repo.Update(ctx, order)
		return order, err
	}

	if resp.Status == "AUTHORIZED" {
		order.Status = StatusPaid
	} else {
		order.Status = StatusFailed
	}

	if err := uc.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	if id == "" {
		return nil, errors.New("id required")
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id required")
	}

	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if order.Status == StatusPaid {
		return errors.New("cannot cancel paid order")
	}

	order.Status = StatusCancelled
	return uc.repo.Update(ctx, order)
}
