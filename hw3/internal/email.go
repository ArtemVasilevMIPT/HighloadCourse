package internal

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
		panic(err)
	}
}

func SendRegistrationEmail(addr string, token string) {
	URL := "localhost:8080/confirm-register?jwt=" + token
	mail := "To confirm your registration, please, follow this link:\n" + URL
	QueueToSend(addr, mail)
}

func SendResetEmail(addr string, token string) {
	URL := "localhost:8080/confirm-reset?jwt=" + token
	mail := "To reset your password, please, follow this link:\n" + URL
	QueueToSend(addr, mail)
}

func QueueToSend(addr string, text string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare("emails", false, false, false, false, nil)
	failOnError(err, "Failed to declare a queue")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(addr + ":" + text),
		})
	failOnError(err, "Failed to publish a message")
	fmt.Printf("Queued email for %s\n", addr)
}
