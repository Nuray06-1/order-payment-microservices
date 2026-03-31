package main

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"

	"order-service/internal/repository"
	"order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func main() {
	r := gin.Default()

	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5433/orderdb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	paymentClient := usecase.NewHTTPPaymentClient("http://localhost:8083")

	repo := repository.NewPostgresOrderRepo(db)

	uc := usecase.NewOrderUseCase(repo, paymentClient)

	handler := http.NewOrderHandler(uc)

	r.GET("/orders/:id", handler.GetOrder)

	r.POST("/orders", handler.CreateOrder)

	r.PATCH("/orders/:id/cancel", handler.CancelOrder)

	r.Run(":8081")
}
