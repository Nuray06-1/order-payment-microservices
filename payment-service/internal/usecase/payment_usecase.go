package usecase

import (
	"context"
	"errors"

	"payment-service/internal/domain"

	"github.com/google/uuid"
)

const (
	StatusAuthorized = "AUTHORIZED"
	StatusDeclined   = "DECLINED"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
}

type PaymentUseCase struct {
	repo PaymentRepository
}

func NewPaymentUseCase(r PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: r}
}

func (uc *PaymentUseCase) Process(
	ctx context.Context,
	orderID string,
	amount int64,
) (*domain.Payment, error) {

	if orderID == "" {
		return nil, errors.New("orderID is required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	existing, err := uc.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	if amount > 100000 {
		payment := &domain.Payment{
			ID:      uuid.New().String(),
			OrderID: orderID,
			Amount:  amount,
			Status:  StatusDeclined,
		}

		if err := uc.repo.Create(ctx, payment); err != nil {
			return nil, err
		}

		return payment, nil
	}

	payment := &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        StatusAuthorized,
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetPayment(ctx context.Context, orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}
