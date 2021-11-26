package reaction

import (
	"context"
	"fmt"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Reacter interface {
	Name() string
	Load() error
	React(ctx context.Context, update telegram.Update) error
}

func conversationKey(msg *telegram.Message) string {
	return fmt.Sprintf("%d:%s", msg.Chat.ID, msg.From.UserName)
}
