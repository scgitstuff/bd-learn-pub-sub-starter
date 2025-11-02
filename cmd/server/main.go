package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	url := "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril server...")

	con, err := amqp.Dial(url)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}
	defer con.Close()
	fmt.Printf("connected to: %s\n", url)

	channel, err := con.Channel()
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}
	fmt.Println("channel open")

	key := routing.GameLogSlug + ".*"
	_, queue, err := pubsub.DeclareAndBind(
		con,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		key,
		pubsub.Durable)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	// TODO: no constants defined, just hard coded shit
	gamelogic.PrintServerHelp()

	for {
		stuff := gamelogic.GetInput()
		if len(stuff) == 0 {
			continue
		}

		switch stuff[0] {
		case routing.PauseKey:
			fmt.Println("sending a pause message")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				fmt.Printf("Bad stuff happened:\n%s\n", err)
			}
		case "resume":
			fmt.Println("sending a resume message")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				fmt.Printf("Bad stuff happened:\n%s\n", err)
			}
		case "quit":
			fmt.Println("program is shutting down")
			return
		default:
			fmt.Printf("bad message: %s\n", stuff[0])
		}
	}
}
