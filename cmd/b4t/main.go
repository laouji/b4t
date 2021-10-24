package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/laouji/b4t/pkg/listener"
)

var (
	version = "n/a"
)

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		fmt.Printf("ERROR no token")
		os.Exit(1)
	}

	client, err := telegram.NewBotAPI(token)
	if err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}
	log.Printf("connected as bot user %q ver %s", client.Self.UserName, version)

	pollingTimeout := 60 * time.Second
	l, err := listener.NewListener(client, pollingTimeout)
	if err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}

	ctx := context.Background()
	log.Println("listener started")
	l.Listen(ctx)
}
