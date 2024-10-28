package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	ExchangeName string
	Kind         string
}

func NewRabbitMQConn(cfg *RabbitMQConfig, ctx context.Context) (*amqp.Connection, error) {
	connAddr := fmt.Sprintf("amqp://guest:guest@localhost:5672/")
	// "amqp://guest:guest@localhost:%d/", cfg.Port)

	fmt.Println("test")
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second
	maxRetries := 5

	var conn *amqp.Connection
	var err error

	err = backoff.Retry(func() error {
		conn, err = amqp.Dial(connAddr)
		if err != nil {
			fmt.Printf("Error - %v, %s", err, connAddr)
			return err
		}
		return nil
	}, backoff.WithMaxRetries(bo, uint64(maxRetries-1)))

	fmt.Println("Connected to RabbitMQ")

	go func() {
		select {
		case <-ctx.Done():
			err := conn.Close()
			if err != nil {
				fmt.Printf("Failed to close RabbitMQ connection")
			}
			fmt.Printf("RabbitMQ connection is closed")
		}
	}()

	return conn, err
}
