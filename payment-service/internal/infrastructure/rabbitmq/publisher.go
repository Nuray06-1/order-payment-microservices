package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"payment-service/internal/domain"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

type PaymentEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	CustomerEmail string `json:"customer_email"`
}

func NewPublisher(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		"payment.dlx",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    "payment.dlx",
		"x-dead-letter-routing-key": "payment.failed",
	}

	q, err := ch.QueueDeclare(
		"payment.completed",
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

func (p *Publisher) PublishPaymentCompleted(ctx context.Context, payment *domain.Payment) error {

	event := PaymentEvent{
		EventID: uuid.New().String(),
		OrderID: payment.OrderID,
		Amount:  payment.Amount,
		Status:  payment.Status,
		//временно захардкодить:
		CustomerEmail: "test@example.com",
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		"",
		p.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)

	if err != nil {
		return err
	}

	log.Println("[RabbitMQ] Event sent:", string(body))
	return nil
}
