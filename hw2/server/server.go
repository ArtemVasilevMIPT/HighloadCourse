package main

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	pb "hw2/proto/messenger"
	"log"
	"net"
)

type server struct{}

// Users on server
var user = make(map[string]bool)
var msgQueue = make([]pb.Text, 0)

func (s *server) Enter(ctx context.Context, u *pb.Str) (*pb.Str, error) {
	if _, ok := user[u.Noti]; ok {
		return &pb.Str{Noti: "User already exists"}, errors.New("user already exists")
	}
	user[u.Noti] = true
	log.Println(u.Noti, "Entered the room")
	return &pb.Str{Noti: "You have entered the room"}, nil
}

func (s *server) Send(ctx context.Context, t *pb.Text) (*pb.Ack, error) {
	msgQueue = append(msgQueue, *t)
	log.Println("In Send: From:", t.Msg.From, "To:", t.Msg.To, "Msg:", t.Msg.Msg)
	return &pb.Ack{Done: true}, nil
}
func (s *server) Receive(ctx context.Context, usr *pb.Str) (*pb.Text, error) {
	for i, txt := range msgQueue {
		if txt.Msg.To == usr.Noti {
			log.Println("From:", txt.Msg.From, "To:", txt.Msg.To, "Mess:", txt.Msg.Msg)
			msgQueue[i] = msgQueue[len(msgQueue)-1]
			msgQueue = msgQueue[:len(msgQueue)-1]
			return &txt, nil
		}
	}
	return &pb.Text{Msg: &pb.TextMail{From: "", To: "", Msg: ""}}, nil
}

func (s *server) MustEmbedUnimplementedMessengerServer() {
	//TODO implement me
	panic("implement me")
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Port is not listening: %v", err)
	}
	defer lis.Close()

	log.Println("Server started at port 8080")

	s := grpc.NewServer()
	pb.RegisterMessengerServer(s, &server{})

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("Server falied: %v", err)
	}
}
