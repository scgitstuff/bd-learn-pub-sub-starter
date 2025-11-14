package main

import (
	"fmt"
	"strconv"
	"time"

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
		fmt.Printf("amqp.Dial() failed:\n%s\n", err)
		return
	}
	defer conn.Close()
	fmt.Printf("connected to: %s\n", url)

	channel, err := conn.Channel()
	if err != nil {
		fmt.Printf("con.Channel() failed:\n%s\n", err)
		return
	}

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("ClientWelcome() failed:\n%s\n", err)
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
		fmt.Printf("Subscribe to pause failed:\n%s\n", err)
		return
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+user,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(state, channel),
	)
	if err != nil {
		fmt.Printf("Subscribe to move failed:\n%s\n", err)
		return
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(state, channel),
	)
	if err != nil {
		fmt.Printf("Subscribe to war failed:\n%s\n", err)
		return
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err := state.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Bad stuff happened:\n%s\n", err)
				continue
			}
		case "move":
			move, err := state.CommandMove(words)
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
			// fmt.Println("Spamming not allowed yet!")
			if len(words) < 2 {
				fmt.Println("spam requires an argument")
				continue
			}
			count, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println("spam argument must be an int")
				continue
			}
			for ; count > 0; count-- {
				// fmt.Printf("spam: %v\n", count)
				msg := gamelogic.GetMaliciousLog()
				err = pubsub.PublishGob(
					channel,
					routing.ExchangePerilTopic,
					routing.GameLogSlug+"."+user,
					routing.GameLog{
						CurrentTime: time.Now(),
						Username:    user,
						Message:     msg,
					},
				)
				if err != nil {
					fmt.Printf("Bad stuff happened:\n%s\n", err)
					continue
				}
			}

		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("bad message: %s\n", words[0])
		}
	}
}
