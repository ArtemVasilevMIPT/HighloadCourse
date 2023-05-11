package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "hw5/proto/messenger"
	"io"
	"log"
	"time"
)

func printMail(mail *pb.Mail) {
	fmt.Printf("\n\t\t\t\tFrom: %s To: %s Msg: %s\n", mail.From, mail.To, mail.Msg)
}

func printMyMail(mail *pb.Mail) {
	fmt.Printf("\nFrom: %s To: %s Msg: %s\n", mail.From, mail.To, mail.Msg)
}

func input(ctx context.Context, usr string, c pb.MessengerClient) {
	stream, err := c.Send(ctx)
	defer func() {
		_, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		}
	}()
	if err != nil {
		log.Fatalf("%v.Send(_) = _, %v", c, err)
	}
	for {
		var (
			msg string
			to  string
		)
		fmt.Printf("Send Message To: ")
		if _, err := fmt.Scanln(&to); err == io.EOF {
			break
		}
		fmt.Printf("Enter Message: ")
		if _, err := fmt.Scanln(&msg); err == io.EOF {
			break
		}
		mail := &pb.Mail{From: usr, To: to, Msg: msg, Time: time.Now().Unix()}
		err = stream.Send(mail)
		if err != nil {
			log.Fatalf("%v.Send(_) = _, %v", stream, err)
		}
	}
}

func inputSecret(ctx context.Context, usr string, duration int64, c pb.MessengerClient) {
	stream, err := c.SendSecret(ctx)
	defer func() {
		_, err := stream.CloseAndRecv()
		if err != nil {
			log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		}
	}()
	if err != nil {
		log.Fatalf("%v.SendSecret(_) = _, %v", c, err)
	}
	for {
		var (
			msg string
			to  string
		)
		fmt.Printf("Send Message To: ")
		if _, err := fmt.Scanln(&to); err == io.EOF {
			break
		}
		fmt.Printf("Enter Message: ")
		if _, err := fmt.Scanln(&msg); err == io.EOF {
			break
		}
		mail := &pb.Mail{From: usr, To: to, Msg: msg, Time: time.Now().Unix()}
		secretMail := &pb.SecretMail{Content: mail, Duration: duration}
		err = stream.Send(secretMail)
		if err != nil {
			log.Fatalf("%v.SendSecret(_) = _, %v", stream, err)
		}
	}
}

func main() {
	conn, err := grpc.Dial("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Connection is not established: %v", err)
	}
	defer conn.Close()

	c := pb.NewMessengerClient(conn)
	ctx := context.Background()
	var (
		usr  string
		user *pb.User
	)

	for {
		fmt.Println("Enter username: ")
		fmt.Scanln(&usr)

		user = &pb.User{Username: usr}
		r, err := c.Enter(ctx, user)
		if err != nil {
			log.Fatalf("Could not connect: %v", err)
			continue
		}
		for {
			mail, err := r.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("%v.Enter(_), %v", c, err)
			}
			if mail.From == user.Username {
				printMyMail(mail)
			} else {
				printMail(mail)
			}
		}
		break
	}
	defer c.Leave(ctx, user)
	go func() {
		stream, err := c.Receive(ctx, user)
		if err != nil {
			log.Fatalf("%v.Receive(_), %v", c, err)
		}
		for {
			mail, err := stream.Recv()
			if err == nil && mail.From != "" {
				printMail(mail)
			}
		}
	}()
	fmt.Println("Activate secret chat? [Y/n]")
	isSecret := ""
	fmt.Scanln(&isSecret)
	if isSecret == "Y" {
		fmt.Println("Enter duration in seconds (max 3600)")
		var duration int64
		fmt.Scanln(&duration)
		inputSecret(ctx, usr, duration, c)
	} else {
		input(ctx, usr, c)
	}
}
