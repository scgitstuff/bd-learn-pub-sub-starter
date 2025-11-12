package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(
	gs *gamelogic.GameState,
) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		// fmt.Printf("handlerPause : %v\n", ps.IsPaused)
		defer fmt.Print("> ")
		gs.HandlePause(ps)

		return pubsub.Ack
	}
}

func handlerMove(
	gs *gamelogic.GameState,
	channel *amqp.Channel,
) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		// fmt.Printf("handlerMove : %v\n", move.Player.Username)
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)

		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			fmt.Printf("\nWAR: A:%s - D:%s\n", move.Player.Username, gs.Player.Username)
			// return pubsub.Ack

			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.Player.Username,
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.Player,
				},
			)
			if err != nil {
				fmt.Printf("RecognitionOfWar message failed:\n%s\n", err)
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}

func handlerWar(
	gs *gamelogic.GameState,
) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		// fmt.Printf("handlerPause : %v\n", ps.IsPaused)
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		_ = winner
		_ = loser

		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}
