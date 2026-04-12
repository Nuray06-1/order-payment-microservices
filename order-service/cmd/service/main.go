package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"order-service/internal/repository"
	"order-service/internal/transport/http"
	"order-service/internal/usecase"

	grpcHandler "order-service/internal/transport/grpc"
	pb "order-service/pkg/order"
	payment "order-service/pkg/payment"
)

func main() {
	godotenv.Load()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := grpc.Dial(os.Getenv("PAYMENT_SERVICE_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to payment service: %v", err)
	}
	defer conn.Close()
	paymentGRPCClient := payment.NewPaymentServiceClient(conn)

	repo := repository.NewPostgresOrderRepo(db)
	uc := usecase.NewOrderUseCase(repo, paymentGRPCClient)
	r := gin.Default()
	handler := http.NewOrderHandler(uc)

	r.GET("/orders/:id", handler.GetOrder)
	r.POST("/orders", handler.CreateOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)

	go func() {
		grpcPort := os.Getenv("ORDER_GRPC_PORT")
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen for gRPC: %v", err)
		}
		grpcServer := grpc.NewServer()
		orderGRPCHandler := grpcHandler.NewOrderGRPCHandler(uc)
		pb.RegisterOrderServiceServer(grpcServer, orderGRPCHandler)
		reflection.Register(grpcServer)

		log.Println("Order gRPC Streaming Server is REALY running on :50052")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	log.Printf("Order REST API started on port %s", os.Getenv("REST_PORT"))
	r.Run(":" + os.Getenv("REST_PORT"))
}
