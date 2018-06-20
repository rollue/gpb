package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

//Message is struct that have message for rabbitMQ
type Message struct {
	Text    string
	Channel int
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func (m *Message) getChannelName() string {
	return fmt.Sprintf("send_channel_%d", m.Channel)
}

//Send it's own message to RabbitMQ
func (m *Message) Send() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Faild to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		m.getChannelName(),
		false,
		false,
		false,
		false,
		nil,
	)

	failOnError(err, "Failed to declare a queue")

	body := m.Text

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	failOnError(err, "Failed to publish a message")
}
