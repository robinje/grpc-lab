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
	chat "github.com/robinje/grpc-lab/chat"
)

func main() {
	conn, err := grpc.Dial(":50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	client := chat.NewChatClient(conn)

	stream, err := client.Connect(context.Background(), &chat.ConnectRequest{Name: "Client"})
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

		// Use a separate context for SendMessage to avoid cancellation due to timeout.
		sendMessageCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		if _, err := client.SendMessage(sendMessageCtx, &chat.Message{Text: msg}); err != nil {
			log.Printf("failed to send message: %v", err)
		}
		cancel()
	}

	wg.Wait()
}