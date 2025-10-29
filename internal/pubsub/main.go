package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func PublishJSON[T any](
	ch *amqp.Channel,
	exchange, key string,
	val T) error {

	data, err := json.Marshal(val)
	if err != nil {
		fmt.Println("Marshal() failed")
		return err
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange, key,
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	if err != nil {
		fmt.Println("PublishWithContext() failed")
		return err
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {

	channel, err := conn.Channel()
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return nil, amqp.Queue{}, err
	}
	fmt.Println("channel open")

	tranQ, err := channel.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return nil, amqp.Queue{}, err
	}
	_ = tranQ

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return nil, amqp.Queue{}, err
	}

	return channel, tranQ, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {

	return nil
}
