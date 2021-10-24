package listener

import (
	"context"
	"fmt"
	"log"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Listener struct {
	updates telegram.UpdatesChannel
}

func NewListener(client *telegram.BotAPI, timeout time.Duration) (l *Listener, err error) {
	updateConfig := telegram.NewUpdate(0)
	updateConfig.Timeout = int(timeout / time.Second)

	updates, err := client.GetUpdatesChan(updateConfig)
	if err != nil {
		return l, fmt.Errorf("failed to get updates channel: %w", err)
	}

	return &Listener{
		updates: updates,
	}, nil
}

func (l *Listener) Listen(ctx context.Context) {
	for {
		select {
		case update := <-l.updates:
			log.Printf("UPDATE %+v", update)
		case <-ctx.Done():
			log.Printf("listener closed (context canceled)")
			return
		}
	}
}
