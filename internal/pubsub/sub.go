package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange, queueName, key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {

	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType)
	if err != nil {
		return err
	}

	stuff, err := ch.Consume(
		queue.Name,
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

	go doStuff(stuff, handler)

	return nil
}

func doStuff[T any](stuff <-chan amqp.Delivery, handler func(T) AckType) {
	for msg := range stuff {
		var x T

		err := json.Unmarshal(msg.Body, &x)
		if err != nil {
			fmt.Printf("could not unmarshal message: %v\n", err)
			continue
		}

		ack := handler(x)
		switch ack {
		case Ack:
			msg.Ack(false)
			fmt.Println("Ack")
		case NackRequeue:
			msg.Nack(false, true)
			fmt.Println("NackRequeue")
		case NackDiscard:
			msg.Nack(false, false)
			fmt.Println("NackDiscard")
		default:
			fmt.Println("This should never happen")
		}
	}
}
