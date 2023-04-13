package main

import (
	"crypto/tls"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"hw3/config"
	"log"
	"net/mail"
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
	viper.AddConfigPath("config")
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
			fmt.Printf("Received message\n")
			msg := string(d.Body)
			splits := strings.SplitN(msg, ":", 2)
			addr := splits[0]
			text := splits[1]
			/*
				auth := smtp.PlainAuth("", configuration.SMTP_USERNAME, configuration.SMTP_PASSWORD, configuration.SMTP_HOSTNAME)
				er := smtp.SendMail(configuration.SMTP_HOSTNAME+":"+configuration.SMTP_PORT, auth,
					configuration.SMTP_USERNAME, []string{addr}, text)
				if er != nil {
					fmt.Printf("%v\n", er)
					panic(err)
				}
			*/
			servername := configuration.SMTP_HOSTNAME + ":" + configuration.SMTP_PORT
			message := fmt.Sprintf("From: %s\r\nTo: %s\r\n\r\n%s\n", configuration.SMTP_USERNAME, addr, text)
			//fmt.Printf("Message:\n%s\n", message)
			auth := smtp.PlainAuth("", configuration.SMTP_USERNAME, configuration.SMTP_PASSWORD, configuration.SMTP_HOSTNAME)
			tlsconfig := &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         configuration.SMTP_HOSTNAME,
			}
			con, er := tls.Dial("tcp", servername, tlsconfig)
			if er != nil {
				log.Panic(er)
			}
			c, er := smtp.NewClient(con, configuration.SMTP_HOSTNAME)
			if er != nil {
				log.Panic(er)
			}
			if er = c.Auth(auth); er != nil {
				log.Panic(er)
			}
			from := mail.Address{Address: configuration.SMTP_USERNAME}
			to := mail.Address{Address: addr}
			if er = c.Mail(from.Address); er != nil {
				log.Panic(er)
			}
			if er = c.Rcpt(to.Address); er != nil {
				log.Panic(er)
			}
			w, er := c.Data()
			if er != nil {
				log.Panic(er)
			}
			_, er = w.Write([]byte(message))
			if er != nil {
				log.Panic(er)
			}
			er = w.Close()
			if er != nil {
				log.Panic(er)
			}
			c.Quit()
		}
	}()

	fmt.Printf("Waiting for messages\n")
	<-forever
}
