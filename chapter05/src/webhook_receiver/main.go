package main

import (
	"fmt"

	"github.com/nlopes/slack"
)

func main() {
	api := slack.New("xoxb-297980043237-2tFWdlp6KrJvB0qzZxO15BCV")

	messages := make(chan string, 10)

	rtm := api.NewRTM()
	go rtm.ManageConnection()
	go balancer(messages)

	for msg := range rtm.IncomingEvents {

		switch msg.Data.(type) {
		case *slack.MessageEvent:
			obj := msg.Data.(*slack.MessageEvent)
			messages <- obj.Text

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:

		}
	}

}
