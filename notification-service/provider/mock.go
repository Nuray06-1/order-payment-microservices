package provider

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

type MockProvider struct{}

func (m *MockProvider) Send(
	email string,
	orderID string,
	amount int64,
) error {

	time.Sleep(
		2*time.Second,
	)

	if rand.Intn(3) == 0 {

		return errors.New(
			"temporary network error",
		)
	}

	log.Printf(
		"[Mock Email] Sent to %s for order %s",
		email,
		orderID,
	)

	return nil
}