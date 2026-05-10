package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"notification-service/provider"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	CustomerEmail string `json:"customer_email"`
}

func main() {

	godotenv.Load()

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	var sender provider.EmailSender

	mode := os.Getenv(
		"PROVIDER_MODE",
	)

	switch mode {

	case "SIMULATED":

		sender = &provider.MockProvider{}

	default:

		sender = &provider.MockProvider{}
	}

	redisClient := redis.NewClient(
		&redis.Options{
			Addr: os.Getenv(
				"REDIS_ADDR",
			),
		},
	)

	maxRetries, _ := strconv.Atoi(
		os.Getenv(
			"MAX_RETRIES",
		),
	)

	backoffSeconds, _ := strconv.Atoi(
		os.Getenv(
			"INITIAL_BACKOFF",
		),
	)

	rabbitURL := os.Getenv(
		"RABBITMQ_URL",
	)

	sigChan := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		sigChan,
		os.Interrupt,
		syscall.SIGTERM,
	)

	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {

		conn, err = amqp.Dial(
			rabbitURL,
		)

		if err == nil {
			break
		}

		log.Println(
			"Waiting for RabbitMQ...",
			err,
		)

		time.Sleep(
			3 * time.Second,
		)
	}

	if err != nil {

		log.Fatal(
			"Failed to connect to RabbitMQ",
		)
	}

	defer conn.Close()

	ch, err := conn.Channel()

	if err != nil {
		log.Fatal(err)
	}

	defer ch.Close()

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
		log.Fatal(err)
	}

	_, err = ch.QueueDeclare(
		"payment.failed",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		"payment.failed",
		"payment.failed",
		"payment.dlx",
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"notification-consumer",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(
		"Notification service is waiting for messages...",
	)

	done := make(
		chan struct{},
	)

	go func() {

		defer close(done)

		for {

			select {

			case <-ctx.Done():

				log.Println(
					"Stopping consumer...",
				)

				return

			case msg, ok := <-msgs:

				if !ok {
					return
				}

				var event PaymentEvent

				err := json.Unmarshal(
					msg.Body,
					&event,
				)

				if err != nil {

					msg.Nack(
						false,
						false,
					)

					continue
				}

				exists, _ := redisClient.Exists(
					ctx,
					"event:"+event.EventID,
				).Result()

				if exists > 0 {

					log.Println(
						"Duplicate ignored:",
						event.EventID,
					)

					msg.Ack(false)

					continue
				}

				backoff := time.Duration(
					backoffSeconds,
				) * time.Second

				var sendErr error

				for i := 0; i < maxRetries; i++ {

					sendErr = sender.Send(
						event.CustomerEmail,
						event.OrderID,
						event.Amount,
					)

					if sendErr == nil {
						break
					}

					log.Println(
						"Retry:",
						i+1,
						sendErr,
					)

					time.Sleep(
						backoff,
					)

					backoff *= 2
				}

				if sendErr != nil {

					msg.Nack(
						false,
						false,
					)

					continue
				}

				redisClient.Set(
					ctx,
					"event:"+event.EventID,
					"done",
					24*time.Hour,
				)

				msg.Ack(false)

				log.Println(
					"Notification sent:",
					event.EventID,
				)
			}
		}
	}()

	<-sigChan

	log.Println(
		"Shutdown signal received...",
	)

	cancel()

	ch.Cancel(
		"notification-consumer",
		false,
	)

	<-done

	log.Println(
		"Graceful shutdown completed",
	)
}