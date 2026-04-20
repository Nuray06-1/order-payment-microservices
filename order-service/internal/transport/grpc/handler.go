package grpc

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"order-service/internal/usecase"

	pb "github.com/Nuray06-1/proto-generated/order"
)

type OrderGRPCHandler struct {
	pb.UnimplementedOrderServiceServer
	orderUseCase *usecase.OrderUseCase
}

func NewOrderGRPCHandler(ou *usecase.OrderUseCase) *OrderGRPCHandler {
	return &OrderGRPCHandler{orderUseCase: ou}
}

func (h *OrderGRPCHandler) SubscribeToOrderUpdates(
	req *pb.OrderRequest,
	stream pb.OrderService_SubscribeToOrderUpdatesServer,
) error {

	lastStatus := ""

	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
		}

		order, err := h.orderUseCase.GetOrder(stream.Context(), req.OrderId)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to get order: %v", err)
		}
		if order == nil {
			return status.Error(codes.NotFound, "order not found")
		}

		if order.Status != lastStatus {
			err := stream.Send(&pb.OrderStatusUpdate{
				Status:    order.Status,
				UpdatedAt: time.Now().Format(time.RFC3339),
			})
			if err != nil {
				return status.Errorf(codes.Internal, "failed to send update: %v", err)
			}
			lastStatus = order.Status
		}

		if order.Status == "PAID" || order.Status == "CANCELLED" {
			return nil
		}

		time.Sleep(2 * time.Second)
	}
}
