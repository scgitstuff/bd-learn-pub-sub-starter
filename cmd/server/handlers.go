package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerLog() func(gl routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		// fmt.Printf("handlerLog : %v\n", gl.Message)
		defer fmt.Print("> ")
		gamelogic.WriteLog(gl)

		return pubsub.Ack
	}
}
