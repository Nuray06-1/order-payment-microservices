package grpc

import (
	"context"

	"payment-service/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Nuray06-1/proto-generated/payment"
)

type PaymentGRPCHandler struct {
	pb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentGRPCHandler(uc *usecase.PaymentUseCase) *PaymentGRPCHandler {
	return &PaymentGRPCHandler{uc: uc}
}

func (h *PaymentGRPCHandler) ProcessPayment(
	ctx context.Context,
	req *pb.PaymentRequest,
) (*pb.PaymentResponse, error) {

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	paymentResult, err := h.uc.Process(ctx, req.OrderId, int64(req.Amount))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "payment failed: %v", err)
	}

	return &pb.PaymentResponse{
		Status: paymentResult.Status,
	}, nil
}
