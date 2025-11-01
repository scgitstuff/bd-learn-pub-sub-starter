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

	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}
	defer conn.Close()
	fmt.Printf("connected to: %s\n", url)

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}
	fmt.Printf("User: %s\n", user)

	state := gamelogic.NewGameState(user)

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, user)
	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(state),
	)

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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		// TODO: this is always false. Why?
		fmt.Printf("handlerPause : %v\n", ps.IsPaused)
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
