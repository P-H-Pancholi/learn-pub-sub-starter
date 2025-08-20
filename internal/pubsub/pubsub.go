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

type Acktype int64

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
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

func SubscribeJSON[T any](conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	deliChan, err := ch.Consume(q.Name, "", false, false, false, false, amqp.Table{})
	if err != nil {
		return err
	}
	go func() {
		for d := range deliChan {
			var msg T
			if err := json.Unmarshal(d.Body, &msg); err == nil {
				ack := handler(msg)
				switch ack {
				case Ack:
					fmt.Println("Message processed successfully")
					d.Ack(false)
				case NackRequeue:
					fmt.Println("Message not processed successfully, but should be requeued on the same queue to be processed again (retry).")
					d.Nack(false, true)
				case NackDiscard:
					fmt.Println("Message not processed successfully, and should be discarded (to a dead-letter queue if configured or just deleted entirely).")
					d.Nack(false, false)
				}
				d.Ack(false)
			} else {
				d.Nack(false, false)
			}
		}
	}()

	return nil
}
