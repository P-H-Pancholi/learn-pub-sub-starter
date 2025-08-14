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

	rabbChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error while starting connection: %v\n", err)
	}

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.DurableQueueType)
	if err != nil {
		log.Fatalf("Error while declaring & binding queue: %v\n", err)
	}

	gamelogic.PrintServerHelp()
	for {
		s := gamelogic.GetInput()
		if len(s) == 0 {
			continue
		}
		switch s[0] {
		case "pause":
			fmt.Println("Sending pause message...")
			if err := pubsub.PublishJSON(rabbChan, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{
				IsPaused: true,
			}); err != nil {
				log.Fatalf("Error in publish json method: %v\n", err)
			}
		case "resume":
			fmt.Println("Sending resume message...")
			if err := pubsub.PublishJSON(rabbChan, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{
				IsPaused: false,
			}); err != nil {
				log.Fatalf("Error in publish json method: %v\n", err)
			}
		case "quit":
			fmt.Println("Exiting Bye....")
			return
		default:
			fmt.Println("Please enter valid command")
		}
	}
}
