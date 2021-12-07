package reaction

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	redis "github.com/go-redis/redis/v8"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	onboarderConfigFileName = "onboarder.csv"

	// TODO make these configurable
	onboarderConvoGreeting = "Hi @%s! Would you like to join %s?"
	onboarderConvoEnd      = "OK, goodbye."
)

type (
	Onboarder struct {
		client *telegram.BotAPI
		rdb    *redis.Client

		// group where user will be onboarded to
		groupChat telegram.Chat
		expiry    time.Duration

		dataDir string

		hangingConvos map[string]string // messageID -> conversationkey
		questions     []string
	}
)

func NewOnboarder(
	client *telegram.BotAPI,
	rdb *redis.Client,
	dataDir string,
	groupChat telegram.Chat,
) *Onboarder {
	return &Onboarder{
		client:    client,
		rdb:       rdb,
		dataDir:   dataDir,
		groupChat: groupChat,
		expiry:    time.Hour,
	}
}

func (r *Onboarder) Name() string {
	return "onboarder"
}

func (r *Onboarder) Load() error {
	file, err := os.Open(filepath.Join(r.dataDir, onboarderConfigFileName))
	if err != nil {
		return err
	}
	defer file.Close()

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return err
	}

	if len(lines) == 0 {
		return fmt.Errorf("no onboarding questions found in config file")
	}

	r.questions = make([]string, 0, len(lines))
	for _, line := range lines {
		// csv is currently just one column
		r.questions = append(r.questions, line[0])
	}
	return nil
}

func (r *Onboarder) React(ctx context.Context, update telegram.Update) error {
	if err := r.handleCallback(ctx, update); err != nil {
		return err
	}

	msg := update.Message
	if msg == nil {
		return nil
	}

	if err := r.handleReply(ctx, msg); err != nil {
		return err
	}

	if !r.shouldReact(ctx, msg) {
		return nil
	}
	log.Printf("new user")

	reply := telegram.NewMessage(msg.Chat.ID, r.greeting(msg.From.String()))
	reply.ReplyToMessageID = msg.MessageID
	reply.ReplyMarkup = r.greetingKeyboard()
	sentMsg, err := r.client.Send(reply)
	if err != nil {
		return err
	}

	if err := r.addHangingMessage(ctx, sentMsg.MessageID, msg.From.UserName); err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) shouldReact(ctx context.Context, msg *telegram.Message) bool {
	if msg == nil {
		return false
	}

	text := strings.ToLower(strings.TrimSpace(msg.Text))
	if text == "" {
		return false
	}

	if !strings.Contains(text, "onboard") {
		return false
	}
	log.Printf("CHAT: %+v", msg.Chat)
	return true
}

func (r *Onboarder) handleCallback(ctx context.Context, update telegram.Update) error {
	callback := update.CallbackQuery
	if callback == nil {
		return nil
	}
	log.Printf("CALLBACK msg: %+v", callback.Message)
	log.Printf("CALLBACK reply: %+v", callback.Message.ReplyToMessage)

	if _, err := r.client.AnswerCallbackQuery(telegram.NewCallback(callback.ID, callback.Data)); err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	var reply telegram.MessageConfig
	if callback.Data == "n" {
		// end conversation
		reply = telegram.NewMessage(callback.Message.Chat.ID, onboarderConvoEnd)
	} else {
		// initiate questions
		reply = telegram.NewMessage(callback.Message.Chat.ID, r.questions[0])
		reply.ReplyToMessageID = callback.Message.ReplyToMessage.MessageID
		reply.ReplyMarkup = telegram.ForceReply{ForceReply: true, Selective: true}
	}
	sentMsg, err := r.client.Send(reply)
	if err != nil {
		return err
	}

	log.Printf("question chat id: %d", sentMsg.MessageID)
	if err := r.addHangingMessage(ctx, sentMsg.MessageID, callback.Message.ReplyToMessage.From.UserName); err != nil {
		return err
	}
	return nil
}

func (r *Onboarder) handleReply(ctx context.Context, msg *telegram.Message) error {
	if msg == nil {
		return fmt.Errorf("message should not be nil in reply check")
	}

	gotReply := msg.ReplyToMessage
	if gotReply == nil {
		return nil
	}

	log.Printf("GOT REPLY: %+v", gotReply)
	isHanging, err := r.isHangingMessage(ctx, gotReply.MessageID)
	if err != nil {
		return err
	}

	if isHanging {
		answers, err := r.getAnswers(ctx, msg.From.UserName)
		if err != nil {
			return err
		}
		log.Printf("ALL ANSWERS %s", answers)

		var shouldAddHanging bool
		var reply telegram.MessageConfig
		if len(answers) < len(r.questions) {
			// not all questions have been asked, proceed to next
			r.setAnswer(ctx, msg.From.UserName, msg.Text)
			answers = append(answers, msg.Text)

			log.Printf("LOGGED ANSWER: %s (from %s) to QUESTION: %s (from %s)", msg.Text, msg.From.UserName, gotReply.Text, gotReply.From.UserName)

			// check if this is the last question
			if len(answers) >= len(r.questions) {
				reply = telegram.NewMessage(msg.Chat.ID, onboarderConvoEnd)
				r.setMembershipPending(ctx, gotReply.From.UserName)
			} else {
				shouldAddHanging = true
				reply = telegram.NewMessage(msg.Chat.ID, r.questions[len(answers)])
				reply.ReplyToMessageID = msg.MessageID
				reply.ReplyMarkup = telegram.ForceReply{ForceReply: true, Selective: true}
			}

			r.removeHangingMessage(ctx, gotReply.MessageID)
		} else {
			// all questions have been asked. can maybe give a nice reply
		}
		sentMsg, err := r.client.Send(reply)
		if err != nil {
			return err
		}

		if shouldAddHanging {
			if err := r.addHangingMessage(ctx, sentMsg.MessageID, msg.From.UserName); err != nil {
				return err
			}
		}
		return nil
	}

	// ignore for now, but maybe the bot should respond with a confused message
	return nil
}

func (r *Onboarder) greeting(fromUserName string) string {
	return fmt.Sprintf(onboarderConvoGreeting, fromUserName, r.groupChat.Title)
}

func (r *Onboarder) greetingKeyboard() telegram.InlineKeyboardMarkup {
	return telegram.NewInlineKeyboardMarkup(
		telegram.NewInlineKeyboardRow(
			telegram.NewInlineKeyboardButtonData("Yes", "y"),
			telegram.NewInlineKeyboardButtonData("No", "n"),
		),
	)
}
