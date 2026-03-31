package http

import (
	"net/http"

	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	usecase *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{usecase: uc}
}

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req paymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.usecase.Process(req.OrderID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": payment.Status,
	})
}
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("order_id")

	payment, err := h.usecase.GetPayment(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "payment not found"})
		return
	}

	c.JSON(200, payment)
}
