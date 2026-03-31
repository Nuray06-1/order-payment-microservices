package usecase

import (
	"payment-service/internal/domain"

	"github.com/google/uuid"
)

type PaymentRepository interface {
	Create(payment *domain.Payment) error
	GetByOrderID(orderID string) (*domain.Payment, error)
}

type PaymentUseCase struct {
	repo PaymentRepository
}

func NewPaymentUseCase(r PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: r}
}

func (uc *PaymentUseCase) Process(orderID string, amount int64) (*domain.Payment, error) {
	if amount > 100000 {
		return &domain.Payment{
			ID:      uuid.New().String(),
			OrderID: orderID,
			Amount:  amount,
			Status:  "Declined",
		}, nil
	}

	payment := &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        "Authorized",
	}

	err := uc.repo.Create(payment)
	if err != nil {
		return nil, err
	}

	return payment, nil
}
func (uc *PaymentUseCase) GetPayment(orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(orderID)
}
