package main

import (
	"fmt"
	"math/rand"
	"time"
)

func balancer(messages chan string) {
	//Receive Message From Slack RTM API
	chans := []chan string{
		make(chan string),
		make(chan string),
		make(chan string),
		make(chan string),
	}

	for i := 0; i < 4; i++ {
		go messenger(i, chans[i])
	}

	current := 0

	for text := range messages {
		fmt.Printf("[Balancer] Receive Message '%s'\n", text)

		current = (current + 1) % 4

		chans[current] <- text
	}
}

func messenger(id int, value chan string) {
	fmt.Printf("[Channel %d] Ready for listen\n", id)
	for text := range value {
		message := &Message{
			Text:    text,
			Channel: id,
		}

		message.Send()

		time.Sleep(time.Duration(rand.Int()%3000+2000) * time.Millisecond)
	}
}
