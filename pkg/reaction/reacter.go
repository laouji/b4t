package reaction

import (
	"context"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Reacter interface {
	Name() string
	ShouldReact(ctx context.Context, update telegram.Update) bool
	React(ctx context.Context, update telegram.Update) error
}
