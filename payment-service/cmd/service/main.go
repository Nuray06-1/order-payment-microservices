package main

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"

	"payment-service/internal/repository"
	"payment-service/internal/transport/http"
	"payment-service/internal/usecase"
)

func main() {
	r := gin.Default()

	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5435/paymentdb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	repo := repository.NewPostgresPaymentRepo(db)

	uc := usecase.NewPaymentUseCase(repo)

	handler := http.NewPaymentHandler(uc)

	r.GET("/payments/:order_id", handler.GetPayment)

	r.POST("/payments", handler.CreatePayment)

	r.Run(":8083")
}
