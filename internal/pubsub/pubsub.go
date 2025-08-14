package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string

const (
	DurableQueueType   simpleQueueType = "durable"
	TransientQueueType simpleQueueType = "transient"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("error while marshaling data: %v", err)
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})

	if err != nil {
		return fmt.Errorf("error while publishing data: %v", err)
	}

	return nil

}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType simpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	ampqChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error while creating channel: %v", err)
	}
	queue, err := ampqChan.QueueDeclare(queueName, queueType == DurableQueueType, queueType == TransientQueueType, queueType == TransientQueueType, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error while creating queue: %v", err)
	}
	if err := ampqChan.QueueBind(queueName, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error while binding queue: %v", err)
	}
	return ampqChan, queue, nil
}
