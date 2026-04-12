package grpc

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"order-service/internal/usecase"
	pb "order-service/pkg/order"
)

type OrderGRPCHandler struct {
	pb.UnimplementedOrderServiceServer
	orderUseCase *usecase.OrderUseCase
}

func NewOrderGRPCHandler(ou *usecase.OrderUseCase) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		orderUseCase: ou,
	}
}

func (h *OrderGRPCHandler) SubscribeToOrderUpdates(req *pb.OrderRequest, stream pb.OrderService_SubscribeToOrderUpdatesServer) error {
	lastStatus := ""

	for {
		order, err := h.orderUseCase.GetOrder(req.OrderId)
		if err != nil {
			return status.Errorf(codes.NotFound, "order not found: %v", err)
		}
		if order.Status != lastStatus {
			err := stream.Send(&pb.OrderStatusUpdate{
				Status:    order.Status,
				UpdatedAt: timestamppb.Now(),
			})
			if err != nil {
				return status.Errorf(codes.Internal, "failed to send stream update: %v", err)
			}
			lastStatus = order.Status
		}
		if order.Status == "Paid" || order.Status == "Cancelled" {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
}
