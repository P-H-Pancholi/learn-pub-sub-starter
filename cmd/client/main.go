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
	for {
		s := gamelogic.GetInput()
		switch s[0] {
		case "spawn":
			if err := state.CommandSpawn(s); err != nil {
				log.Fatalf("error while running spawn command: %v", err)
			}
		case "move":
			_, err := state.CommandMove(s)
			if err != nil {
				log.Fatalf("error while running spawn command: %v", err)
			}
			fmt.Println("Move success")
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
