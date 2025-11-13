package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {

	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType)
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

	go func() {
		for msg := range stuff {
			x, err := unmarshaller(msg.Body)
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
	}()

	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {

	f := func(body []byte) (T, error) {
		var x T
		err := json.Unmarshal(body, &x)

		return x, err
	}

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, f)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {

	f := func(body []byte) (T, error) {
		var x T
		var buf bytes.Buffer

		buf.Write(body)
		dec := gob.NewDecoder(&buf)
		err := dec.Decode(&x)

		return x, err
	}

	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, f)
}
