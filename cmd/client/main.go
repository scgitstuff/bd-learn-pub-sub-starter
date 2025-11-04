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

	channel, err := conn.Channel()
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}
	fmt.Printf("User: %s\n", user)

	state := gamelogic.NewGameState(user)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+user,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(state),
	)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+user,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(state),
	)
	if err != nil {
		fmt.Printf("Bad stuff happened:\n%s\n", err)
		return
	}

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
				continue
			}
		case "move":
			move, err := state.CommandMove(stuff)
			if err != nil {
				fmt.Printf("Bad stuff happened:\n%s\n", err)
				continue
			}
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+user,
				move,
			)
			if err != nil {
				fmt.Printf("Bad stuff happened:\n%s\n", err)
				continue
			}
			fmt.Println("move was published successfully")
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		// fmt.Printf("handlerPause : %v\n", ps.IsPaused)
		defer fmt.Print("> ")
		gs.HandlePause(ps)

		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		// fmt.Printf("handlerMove : %v\n", move.Player.Username)
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)

		if outcome == gamelogic.MoveOutComeSafe ||
			outcome == gamelogic.MoveOutcomeMakeWar {
			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}
