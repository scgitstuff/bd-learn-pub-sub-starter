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

	fmt.Println("Starting Peril client...")

	con, err := amqp.Dial(url)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}
	defer con.Close()
	fmt.Printf("connected to: %s\n", url)

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}
	fmt.Printf("User: %s\n", user)

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, user)
	_, _, err = pubsub.DeclareAndBind(
		con,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}

	state := gamelogic.NewGameState(user)

	for {
		stuff := gamelogic.GetInput()
		if len(stuff) == 0 {
			continue
		}

		switch stuff[0] {
		case "spawn":
			err := state.CommandSpawn(stuff)
			if err != nil {
				fmt.Printf("Bad stuff happened:\n%s\n", err)
				return
			}
		case "move":
			_, err := state.CommandMove(stuff)
			if err != nil {
				fmt.Printf("Bad stuff happened:\n%s\n", err)
				return
			}
		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("bad message: %s\n", stuff[0])
		}
	}
}
