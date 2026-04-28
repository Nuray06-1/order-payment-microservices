package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	CustomerEmail string `json:"customer_email"`
}

var processedEvents = make(map[string]bool)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			break
		}
		log.Println("Waiting for RabbitMQ...", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		"payment.dlx", "direct", true, false, false, false, nil,
	); err != nil {
		log.Fatal(err)
	}

	if _, err := ch.QueueDeclare(
		"payment.failed", true, false, false, false, nil,
	); err != nil {
		log.Fatal(err)
	}

	if err := ch.QueueBind(
		"payment.failed", "payment.failed", "payment.dlx", false, nil,
	); err != nil {
		log.Fatal(err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    "payment.dlx",
		"x-dead-letter-routing-key": "payment.failed",
	}

	q, err := ch.QueueDeclare(
		"payment.completed", true, false, false, false, args,
	)
	if err != nil {
		log.Fatal(err)
	}
	consumerTag := "notification-consumer"

	msgs, err := ch.Consume(
		q.Name,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Notification service is waiting for messages...")

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping consumer loop...")
				return

			case msg, ok := <-msgs:
				if !ok {
					log.Println("Message channel closed")
					return
				}

				var event PaymentEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					log.Println("Invalid message format:", err)
					msg.Nack(false, false)
					continue
				}

				if event.EventID == "" {
					log.Println("Invalid event: missing EventID")
					msg.Nack(false, false)
					continue
				}

				if processedEvents[event.EventID] {
					log.Println("Duplicate event ignored:", event.EventID)
					msg.Ack(false)
					continue
				}
				processedEvents[event.EventID] = true

				if event.CustomerEmail == "fail@example.com" {
					log.Println("Permanent error. Sending to DLQ:", event.EventID)
					msg.Nack(false, false)
					continue
				}

				log.Printf(
					"[Notification] Sent email to %s for Order #%s. Amount: $%.2f",
					event.CustomerEmail,
					event.OrderID,
					float64(event.Amount),
				)

				msg.Ack(false)
				log.Println("Message processed and ACKed:", event.EventID)
			}
		}
	}()

	<-sigChan
	log.Println("Shutdown signal received...")
	cancel()
	if err := ch.Cancel(consumerTag, false); err != nil {
		log.Println("Consumer cancel error:", err)
	}
	<-done

	log.Println("Graceful shutdown completed")
}
