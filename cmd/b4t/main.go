package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	redis "github.com/go-redis/redis/v8"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/laouji/b4t/pkg/command"
	"github.com/laouji/b4t/pkg/listener"
	"github.com/laouji/b4t/pkg/reaction"
)

var (
	version = "n/a"
)

type config struct {
	token     string
	redisAddr string
	redisDB   int
	dataDir   string

	groupID        int64
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
	log.Printf("connected as bot user %q, bot version %q", client.Self.UserName, version)

	chat, err := command.GetChat(conf.groupID, client)
	if err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: conf.redisAddr,
		DB:   conf.redisDB,
	})

	l, err := listener.NewListener(client, conf.pollingTimeout)
	if err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}

	if err := l.RegisterReacters(
		reaction.NewOnboarder(client, rdb, conf.dataDir, chat),
	); err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		fmt.Printf("ERROR %s", err)
		os.Exit(1)
	}

	log.Println("listener started")
	l.Listen(ctx)
}

func parseArgs() (conf config, err error) {
	token := os.Getenv("TOKEN")
	if token == "" {
		return conf, fmt.Errorf("no token")
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return conf, fmt.Errorf("REDIS_DB must be a valid integer: %w", err)
	}

	conf = config{
		token:     token,
		redisAddr: os.Getenv("REDIS_ADDR"),
		redisDB:   redisDB,
	}
	flag.DurationVar(&conf.pollingTimeout, "timeout", 60*time.Second, "polling timeout")
	flag.Int64Var(&conf.groupID, "group", 0, "ID of group")
	flag.StringVar(&conf.dataDir, "data", "./data", "data directory where config files can be found")

	flag.Parse()
	return conf, nil
}
