package reaction

import (
	"context"
	"fmt"
	"log"
	"strings"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Onboarder struct {
	client *telegram.BotAPI
}

func NewOnboarder(client *telegram.BotAPI) *Onboarder {
	return &Onboarder{
		client: client,
	}
}

func (r *Onboarder) Name() string {
	return "onboarder"
}

func (r *Onboarder) ShouldReact(ctx context.Context, update telegram.Update) bool {
	if update.Message == nil {
		return false
	}

	text := strings.ToLower(strings.TrimSpace(update.Message.Text))
	if text == "" {
		return false
	}

	if !strings.Contains(text, "onboard") {
		return false
	}
	return true
}

func (r *Onboarder) React(ctx context.Context, update telegram.Update) error {
	msg := update.Message
	if msg == nil {
		return fmt.Errorf("message cannot be nil for onboard reaction")
	}

	reply := telegram.NewMessage(msg.Chat.ID, r.greeting(msg.From.String()))
	reply.ReplyToMessageID = msg.MessageID
	_, err := r.client.Send(reply)
	if err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) greeting(fromUserName string) string {
	return fmt.Sprintf(
		`Hi %s! To start onboarding respond to this message with the word "Yes"`,
		fromUserName,
	)
}
