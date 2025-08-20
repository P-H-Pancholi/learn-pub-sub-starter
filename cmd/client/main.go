package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	CONN_STRING = "amqp://guest:guest@localhost:5672/"
)

func main() {
	conn, err := amqp.Dial(CONN_STRING)
	if err != nil {
		log.Fatalf("Error while starting connection: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connection was successful")
	s, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error while starting connection: %v", err)
	}
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+s, routing.PauseKey, pubsub.TransientQueueType)
	if err != nil {
		log.Fatalf("Error in Declare and bind function: %v", err)
	}

	state := gamelogic.NewGameState(s)
	if err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		"pause."+state.GetUsername(),
		routing.PauseKey,
		pubsub.TransientQueueType,
		handlerPause(state),
	); err != nil {
		fmt.Printf("error in subscribeJson func: %v\n", err)
		return
	}

	moveChannel, moveQueue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+state.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.TransientQueueType,
	)
	if err != nil {
		log.Fatalf("Error in Declare and bind function: %v", err)
	}

	if err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		moveQueue.Name,
		routing.ArmyMovesPrefix+".*",
		pubsub.TransientQueueType,
		handlerMove(state),
	); err != nil {
		fmt.Printf("error in subscribeJson func: %v\n", err)
		return
	}

	for {
		s := gamelogic.GetInput()
		switch s[0] {
		case "spawn":
			if err := state.CommandSpawn(s); err != nil {
				fmt.Printf("error while running spawn command: %v\n", err)
			}
		case "move":
			m, err := state.CommandMove(s)
			if err != nil {
				fmt.Printf("error while running move command: %v\n", err)
			} else {
				pubsub.PublishJSON(moveChannel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+state.GetUsername(), m)
				fmt.Println("Move was published successfully")
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
			fmt.Println("Please enter valid command")
		}

	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(m gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		mv := gs.HandleMove(m)
		switch mv {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}
