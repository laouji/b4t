package listener

import (
	"context"
	"fmt"
	"log"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/laouji/b4t/pkg/reaction"
)

type Listener struct {
	updates telegram.UpdatesChannel

	reacters []reaction.Reacter
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

func (l *Listener) RegisterReacters(reacters ...reaction.Reacter) {
	for _, r := range reacters {
		log.Printf("registered reactor %q", r.Name())
		l.reacters = append(l.reacters, r)
	}
}

func (l *Listener) Listen(ctx context.Context) {
	for {
		select {
		case update := <-l.updates:
			if err := l.handleUpdate(ctx, update); err != nil {
				log.Printf("failed to handle update: %s", err)
				return
			}
		case <-ctx.Done():
			log.Printf("listener closed (context canceled)")
			return
		}
	}
}

func (l *Listener) handleUpdate(ctx context.Context, update telegram.Update) error {
	for _, r := range l.reacters {
		if !r.ShouldReact(ctx, update) {
			continue
		}

		if err := r.React(ctx, update); err != nil {
			return fmt.Errorf("reacter %q failed to react to update with err: %w", r.Name(), err)
		}
	}
	return nil
}
