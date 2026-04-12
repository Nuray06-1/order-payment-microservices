package grpc

import (
	"context"
	"payment-service/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	payment "payment-service/pkg/payment"
)

type PaymentGRPCHandler struct {
	payment.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentGRPCHandler(uc *usecase.PaymentUseCase) *PaymentGRPCHandler {
	return &PaymentGRPCHandler{uc: uc}
}

func (h *PaymentGRPCHandler) ProcessPayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	paymentResult, err := h.uc.Process(req.OrderId, int64(req.Amount))

	if err != nil {
		return nil, status.Errorf(codes.Internal, "payment failed: %v", err)
	}
	return &payment.PaymentResponse{
		Status:        paymentResult.Status,
		TransactionId: paymentResult.TransactionID,
	}, nil
}
