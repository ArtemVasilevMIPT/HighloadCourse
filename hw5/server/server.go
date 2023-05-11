package main

import (
	"container/heap"
	"context"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	pb "hw5/proto/messenger"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct{}
type userItem struct {
	msgQueue chan *pb.SecretMail
	mtx      *sync.Mutex
}

var (
	users           = make(map[string]userItem)
	secretMsgs      = make(PriorityQueue, 0)
	redisConnection *redis.Client
	secretMutex     = &sync.Mutex{}
)

func RemoveFromRedis(str string, from string, to string) {
	ctx := context.Background()
	if err := redisConnection.LRem(ctx, to, 1, str).Err(); err != nil {
		log.Printf("redis.LRem error: %v\n", err)
	}
	if err := redisConnection.LRem(ctx, from, 1, str).Err(); err != nil {
		log.Printf("redis.LRem error: %v\n", err)
	}
}

func MessageCleaner() {
	for {
		if len(secretMsgs) == 0 {
			time.Sleep(1 * time.Second)
		} else {
			secretMutex.Lock()
			mail := heap.Pop(&secretMsgs).(*Item)
			secretMutex.Unlock()
			str := MailToString(mail.Content)
			if mail.Priority < time.Now().Unix() {
				RemoveFromRedis(str, mail.Content.From, mail.Content.To)
			} else {
				duration := time.Duration(time.Now().Unix() - mail.Priority)
				time.Sleep(duration)
				RemoveFromRedis(str, mail.Content.From, mail.Content.To)
			}
		}
	}
}

func StringToMail(str string) *pb.Mail {
	split := strings.Split(str, "\t")
	t, _ := strconv.Atoi(split[0])
	mail := &pb.Mail{To: split[1], From: split[2], Msg: split[3], Time: int64(t)}
	return mail
}

func AddSecretMail(mail *pb.SecretMail) {
	item := &Item{Content: mail.Content, Priority: mail.Content.Time + mail.Duration}
	heap.Push(&secretMsgs, item)
}

func MailToString(mail *pb.Mail) string {
	return strings.Join([]string{strconv.FormatInt(mail.Time, 10), mail.To, mail.From, mail.Msg}, "\t")
}

func MailExpired(mail *pb.SecretMail) bool {
	if mail.Duration >= 0 {
		if time.Now().Unix() > mail.Content.Time+mail.Duration {
			return true
		}
	}
	return false
}

func DumpMessages(user string) error {
	users[user].mtx.Lock()
	defer users[user].mtx.Unlock()
	log.Println("Dumping messages for ", user)
	flag := true
	for flag {
		select {
		case mail := <-users[user].msgQueue:
			{
				if !MailExpired(mail) {
					msgString := MailToString(mail.Content)
					err := redisConnection.RPush(context.Background(), mail.Content.From, msgString).Err()
					if err != nil {
						log.Printf("LPush error: %v", err)
						return err
					}
					err = redisConnection.RPush(context.Background(), mail.Content.To, msgString).Err()
					if err != nil {
						log.Printf("LPush error: %v", err)
						return err
					}
				}
			}
		default:
			flag = false
		}
	}
	log.Println("Finished dumping")
	return nil
}

func CreateUserIfNotExists(user string) {
	if _, ok := users[user]; !ok {
		users[user] = userItem{msgQueue: make(chan *pb.SecretMail, 2), mtx: &sync.Mutex{}}
	}
}

func SendMailToUser(mail *pb.Mail, user string) error {
	CreateUserIfNotExists(user)
	packet := &pb.SecretMail{Content: mail, Duration: -1}
	users[user].mtx.Lock()
	defer users[user].mtx.Unlock()
	select {
	case users[user].msgQueue <- packet:
		{
			log.Printf("In Send From: %s To: %s", mail.From, mail.To)
		}
	default:
		{
			users[user].mtx.Unlock()
			err := DumpMessages(user)
			users[user].mtx.Lock()
			if err != nil {
				return err
			}
			users[user].msgQueue <- packet
			log.Printf("In Send From: %s To: %s", mail.From, mail.To)
		}
	}
	return nil
}

func SendSecretMailToUser(mail *pb.SecretMail, user string) error {
	CreateUserIfNotExists(user)
	AddSecretMail(mail)
	users[user].mtx.Lock()
	defer users[user].mtx.Unlock()
	select {
	case users[user].msgQueue <- mail:
		{
			log.Printf("In SendSecret From: %s To: %s", mail.Content.From, mail.Content.To)
		}
	default:
		{
			users[user].mtx.Unlock()
			err := DumpMessages(user)
			users[user].mtx.Lock()
			if err != nil {
				return err
			}
			users[user].msgQueue <- mail
			log.Printf("In SendSecret From: %s To: %s", mail.Content.From, mail.Content.To)
		}
	}
	return nil
}

func SendHistory(username string, stream pb.Messenger_EnterServer) error {
	ctx := context.Background()
	history, err := redisConnection.LRange(ctx, username, 0, -1).Result()
	if err != nil {
		return err
	}
	for _, msg := range history {
		mail := StringToMail(msg)
		stream.Send(mail)
	}
	return nil
}

func (s *server) Enter(user *pb.User, enterServer pb.Messenger_EnterServer) error {
	CreateUserIfNotExists(user.Username)
	log.Println(user.Username, "entered the room")
	err := SendHistory(user.Username, enterServer)
	return err
}

func (s *server) Leave(ctx context.Context, user *pb.User) (*pb.Ack, error) {
	err := DumpMessages(user.Username)
	return &pb.Ack{Done: true}, err
}

func (s *server) Send(sendServer pb.Messenger_SendServer) error {
	for {
		mail, err := sendServer.Recv()
		if err == io.EOF {
			return sendServer.SendAndClose(&pb.Ack{Done: true})
		}
		if err != nil {
			return err
		}
		err = SendMailToUser(mail, mail.To)
		if err != nil {
			return err
		}
	}
}

func (s *server) SendSecret(secretServer pb.Messenger_SendSecretServer) error {
	for {
		mail, err := secretServer.Recv()
		if err == io.EOF {
			return secretServer.SendAndClose(&pb.Ack{Done: true})
		}
		err = SendSecretMailToUser(mail, mail.Content.To)
		if err != nil {
			return err
		}
	}
}

func (s *server) Receive(user *pb.User, receiveServer pb.Messenger_ReceiveServer) error {
	for {
		mail := <-users[user.Username].msgQueue
		if !MailExpired(mail) {
			if err := receiveServer.Send(mail.Content); err != nil {
				return err
			}
		}
		msgString := MailToString(mail.Content)
		err := redisConnection.RPush(context.Background(), mail.Content.From, msgString).Err()
		if err != nil {
			return err
		}
		err = redisConnection.RPush(context.Background(), mail.Content.To, msgString).Err()
		if err != nil {
			return err
		}
	}
}

func (s *server) MustEmbedUnimplementedMessengerServer() {
	//TODO implement me
	panic("implement me")
}

func main() {
	redisConnection = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Port is not listening: %v", err)
	}
	defer lis.Close()

	log.Println("Server started at port 8080")

	s := grpc.NewServer()
	pb.RegisterMessengerServer(s, &server{})

	go MessageCleaner()

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("Server falied: %v", err)
	}
}
