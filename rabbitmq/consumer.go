package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/iancoleman/strcase"
	amqp "github.com/rabbitmq/amqp091-go"
)

type IConsumer[T any] interface {
	ConsumeMessage(msg interface{}, dependencies T) error
	//TODO: IsConsumed
}

var consumedMessages []string

type Consumer[T any] struct {
	cfg  *RabbitMQConfig
	conn *amqp.Connection
	ctx  context.Context
	//TODO:
	//  log          logger.ILogger
	// handler      func(queue string, msg amqp.Delivery, dependencies T) error
	// jaegerTracer trace.Tracer
}

func (c Consumer[T]) ConsumeMessage(msg interface{}, dependencies T) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	typeName := reflect.TypeOf(msg).Name()
	snakeTypeName := strcase.ToSnake(typeName)

	err = ch.ExchangeDeclare(
		snakeTypeName,
		c.cfg.Kind,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		fmt.Sprintf("%s_%s", snakeTypeName, "queue"),
		false,
		false,
		true,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	err = ch.QueueBind(
		q.Name,
		snakeTypeName,
		snakeTypeName,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	deliveries, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	go func() {
		select {
		case <-c.ctx.Done():
			defer func(ch *amqp.Channel) {
				err := ch.Close()
				if err != nil {
					fmt.Println(err)
				}
			}(ch)
			return
		case delivery, ok := <-deliveries:
			{
				if !ok {
					return
				}
				consumedMessages = append(consumedMessages, snakeTypeName)

				h, err := json.Marshal(delivery.Headers)
				if err != nil {
					fmt.Printf("Error: %v", h)
					return
				}
				time.Sleep(1 * time.Millisecond)

				err = delivery.Ack(false)
				if err != nil {
					return
				}
			}
		}
	}()

	return nil
}

func NewConsumer[T any](ctx context.Context, cfg *RabbitMQConfig, conn *amqp.Connection) IConsumer[T] {
	return &Consumer[T]{ctx: ctx, cfg: cfg, conn: conn}
}
