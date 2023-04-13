package email

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"hw3/config"
	"log"
	"net/smtp"
	"strings"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
func main() {

	viper.SetConfigName("emailConf")
	viper.AddConfigPath("../../config")
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}
	var configuration config.EmailConfigurations
	err := viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Printf("Unable to decode config into struct, %v", err)
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"emails", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	var forever chan struct{}
	go func() {
		for d := range msgs {
			msg := string(d.Body)
			splits := strings.SplitN(msg, ":", 2)
			addr := splits[0]
			text := []byte(splits[1])

			auth := smtp.PlainAuth("", configuration.SMTP_USERNAME, configuration.SMTP_PASSWORD, configuration.SMTP_HOSTNAME)
			er := smtp.SendMail(configuration.SMTP_HOSTNAME+":"+configuration.SMTP_PORT, auth,
				configuration.SMTP_USERNAME, []string{addr}, text)
			if er != nil {
				fmt.Printf("%v\n", err)
				panic(err)
			}
		}
	}()

	fmt.Printf("Waiting for messages\n")
	<-forever
}
