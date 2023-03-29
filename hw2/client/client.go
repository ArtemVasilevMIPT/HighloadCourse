package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "hw2/proto/messenger"
	"log"
)

func input(ctx context.Context, usr string, c pb.MessengerClient) {
	for {
		var (
			msg string
			to  string
		)
		fmt.Printf("Send Message To: ")
		fmt.Scanln(&to)
		fmt.Printf("Enter Message: ")
		fmt.Scanln(&msg)
		message := &pb.Text{Msg: &pb.TextMail{From: usr, To: to, Msg: msg}}
		c.Send(ctx, message)
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
	var usr string

	for {
		fmt.Println("Enter username: ")
		fmt.Scanln(&usr)

		str := &pb.Str{Noti: usr}
		r, err := c.Enter(ctx, str)
		if err != nil {
			log.Fatalf("Could not connect: %v", err)
			continue
		}
		fmt.Println("\t\t\t\t", r.Noti)
		break
	}
	go input(ctx, usr, c)

	for {
		txt, err := c.Receive(ctx, &pb.Str{Noti: usr})
		if err == nil && txt.Msg.From != "" {
			fmt.Println("\n\t\t\t\tFrom:", txt.Msg.From, "To:", txt.Msg.To, "Msg:", txt.Msg.Msg)
		}
	}
}
