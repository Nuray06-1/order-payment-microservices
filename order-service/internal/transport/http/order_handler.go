package http

import (
	"net/http"

	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	usecase *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{usecase: uc}
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name" binding:"required"`
	Amount     int64  `json:"amount" binding:"required,gt=0"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := c.GetHeader("Idempotency-Key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key required"})
		return
	}

	order, err := h.usecase.CreateOrder(
		c.Request.Context(),
		req.CustomerID,
		req.ItemName,
		req.Amount,
		key,
	)

	if err != nil {
		status := "FAILED"
		if order != nil {
			status = order.Status
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"status": status,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, order)
}
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.usecase.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	err := h.usecase.CancelOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "order cancelled"})
}
