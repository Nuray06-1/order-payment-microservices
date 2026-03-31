package usecase

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type HTTPPaymentClient struct {
	client  *http.Client
	baseURL string
}

func NewHTTPPaymentClient(baseURL string) *HTTPPaymentClient {
	return &HTTPPaymentClient{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		baseURL: baseURL,
	}
}

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponse struct {
	Status string `json:"status"`
}

func (c *HTTPPaymentClient) ProcessPayment(orderID string, amount int64) (string, error) {
	reqBody := paymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}

	jsonData, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(
		c.baseURL+"/payments",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res paymentResponse
	json.NewDecoder(resp.Body).Decode(&res)

	return res.Status, nil
}
