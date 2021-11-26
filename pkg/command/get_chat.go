package command

import (
	"fmt"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GetChat(chatID int64, client *telegram.BotAPI) (chat telegram.Chat, err error) {
	chat, err = client.GetChat(telegram.ChatConfig{
		ChatID: chatID,
	})
	if err != nil {
		return chat, fmt.Errorf("failed to get chat for id %d: %w", chatID, err)
	}
	return chat, nil
}

func GetChatAdministrators(chatID int64, client *telegram.BotAPI) error {
	members, err := client.GetChatAdministrators(telegram.ChatConfig{
		ChatID: chatID,
	})
	if err != nil || len(members) == 0 {
		return fmt.Errorf("failed to get chat admins %w", err)
	}

	return nil
}
