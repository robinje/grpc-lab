package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	chat "path/to/your/chat.pb.go"
)

func main() {
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	client := chat.NewChatClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	stream, err := client.Connect(ctx, &chat.ConnectRequest{Name: "Client"})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			msg, err := stream.Recv()
			if err != nil {
				return
			}
			fmt.Printf("<client%d> %s\n", msg.Id, msg.Text)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()

		if strings.TrimSpace(msg) == "" {
			continue
		}

		if _, err := client.SendMessage(ctx, &chat.Message{Text: msg}); err != nil {
			log.Printf("failed to send message: %v", err)
		}
	}

	wg.Wait()
}