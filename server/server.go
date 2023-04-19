package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	pb "github.com/robinje/grpc-lab/chat"
)

type server struct {
	pb.UnimplementedChatServer
	clients map[int32]chan<- *pb.Message
	mu      sync.Mutex
}

func (s *server) Connect(req *pb.ConnectRequest, stream pb.Chat_ConnectServer) error {
	s.mu.Lock()
	id := int32(len(s.clients) + 1)
	msgChan := make(chan *pb.Message, 10)
	s.clients[id] = msgChan
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, id)
		s.mu.Unlock()
	}()

	go func() {
		for msg := range msgChan {
			if err := stream.Send(msg); err != nil {
				return
			}
		}
	}()

	return nil
}

func (s *server) SendMessage(ctx context.Context, msg *pb.Message) (*pb.MessageAck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ch := range s.clients {
		select {
		case ch <- msg:
		default:
		}
	}

	return &pb.MessageAck{Success: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterChatServer(s, &server{clients: make(map[int32]chan<- *pb.Message)})
	fmt.Println("Server listening on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}