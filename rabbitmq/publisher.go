package rabbitmq

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/ahmetb/go-linq/v3"
	"github.com/iancoleman/strcase"
	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	uuid "github.com/satori/go.uuid"

	jsoniterator "github.com/json-iterator/go"
)

type IPublisher interface {
	PublishMessage(msg interface{}) error
	IsPublished(msg interface{}) bool
}

var publishedMessages []string

type Publisher struct {
	cfg  *RabbitMQConfig
	conn *amqp.Connection
	ctx  context.Context
	//TODO: do log
}

func (p Publisher) PublishMessage(msg interface{}) error {
	data, err := jsoniterator.Marshal(msg)

	if err != nil {
		fmt.Println("Error in marshaling message to publish message")
		return err
	}

	typeName := reflect.TypeOf(msg).Elem().Name()
	snakeTypeName := strcase.ToSnake(typeName)

	//TODO: Add telemetry and headers

	channel, err := p.conn.Channel()
	if err != nil {
		fmt.Println("Error in opening channel to consume message")
		return err
	}

	defer channel.Close()

	err = channel.ExchangeDeclare(
		snakeTypeName,
		p.cfg.Kind,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		fmt.Println("Erorr in declaring exchange to publish message")
		return err
	}

	correlationId := ""

	if p.ctx.Value(echo.HeaderXCorrelationID) != nil {
		correlationId = p.ctx.Value(echo.HeaderXCorrelationID).(string)
	}

	publishingMsg := amqp.Publishing{
		Body:          data,
		ContentType:   "application/json",
		DeliveryMode:  amqp.Persistent,
		MessageId:     uuid.NewV4().String(),
		Timestamp:     time.Now(),
		CorrelationId: correlationId,
	}

	err = channel.Publish(snakeTypeName, snakeTypeName, false, false, publishingMsg)

	if err != nil {
		fmt.Println("Erorr in publishing message")
		return err
	}

	publishedMessages = append(publishedMessages, snakeTypeName)

	// h, err := jsoniter.Marshal(headers)

	// if err
	fmt.Println("Published message: %s", publishingMsg.Body)

	return nil
}

func (p Publisher) IsPublished(msg interface{}) bool {
	typeName := reflect.TypeOf(msg).Name()
	snakeTypeName := strcase.ToSnake(typeName)
	isPublished := linq.From(publishedMessages).Contains(snakeTypeName)

	return isPublished
}

func NewPublisher(ctx context.Context, cfg *RabbitMQConfig, conn *amqp.Connection) IPublisher {
	return &Publisher{ctx: ctx, cfg: cfg, conn: conn}
}
