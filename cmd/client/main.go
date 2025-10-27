package main

import (
	"fmt"
	"os"
	"os/signal"

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

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("\nprogram is shutting down")
}
