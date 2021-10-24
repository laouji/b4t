package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/laouji/b4t/pkg/listener"
)

var (
	version = "n/a"
)

type config struct {
	token          string
	pollingTimeout time.Duration
}

func main() {
	conf, err := parseArgs()
	if err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}

	client, err := telegram.NewBotAPI(conf.token)
	if err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}
	log.Printf("connected as bot user %q ver %s", client.Self.UserName, version)

	l, err := listener.NewListener(client, conf.pollingTimeout)
	if err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("listener started")
	l.Listen(ctx)
}

func parseArgs() (conf config, err error) {
	token := os.Getenv("TOKEN")
	if token == "" {
		return conf, fmt.Errorf("no token")
	}

	var pollingTimeout time.Duration
	flag.DurationVar(&pollingTimeout, "timeout", 60*time.Second, "polling timeout")

	flag.Parse()

	return config{
		token:          token,
		pollingTimeout: pollingTimeout,
	}, nil
}
