package tg

import (
	"context"
	"fmt"
	"github.com/Nidnepel/SendingBot/internal/bot_client"
	"github.com/Nidnepel/SendingBot/internal/model"
	"github.com/Nidnepel/SendingBot/internal/usecase"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

type Bot struct {
	api *tgbotapi.BotAPI
	uc  usecase.Usecase
}

func New(uc usecase.Usecase) bot_client.Bot {
	bot, err := tgbotapi.NewBotAPI("token") // todo: getenv
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	return &Bot{api: bot, uc: uc}
}

func (tb *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.api.GetUpdatesChan(u)

	for update := range updates {
		if mes := update.Message; mes != nil { // If we got a message
			tb.routeMessage(mes)
		}
	}
}

func (tb *Bot) routeMessage(message *tgbotapi.Message) {
	ctx := context.Background()
	args := strings.Split(message.Text, " ")
	switch args[0] {
	case "/key_in":
		if len(args) != 2 {
			return
		}
		chat := &model.Chat{
			Messenger: "tg",
			Key:       args[1],
			ID:        int(message.Chat.ID),
		}
		err := tb.uc.KeyIn(ctx, chat)
		if err != nil {
			return
		}
	case "/key_gen":
		key, err := tb.uc.KeyGen(ctx, &model.Chat{
			Messenger: "tg",
			ID:        int(message.Chat.ID),
		})
		if err != nil {
			return
		}
		tb.Send(model.Message{
			Data: fmt.Sprintf("Ключ присоединения к этому чату: %s", key),
			ChatTo: model.Chat{
				ID: int(message.Chat.ID),
			},
		})
	case "/key_drop":
		tb.Send(model.Message{
			Data: fmt.Sprintf("Ключ для чата сброшен"),
			ChatTo: model.Chat{
				ID: int(message.Chat.ID),
			},
		})
	default:
		err := tb.uc.Send(ctx, transformFromMessageToModel(message))
		if err != nil {
			return
		}
	}
}

func transformFromMessageToModel(message *tgbotapi.Message) *model.Message {
	return &model.Message{
		Data: message.Text,
		ChatFrom: model.Chat{
			Messenger: "tg",
			ID:        int(message.Chat.ID),
		},
		UserFrom: fmt.Sprintf("%s %s:\n", message.From.FirstName, message.From.LastName),
	}
}

func (tb *Bot) Send(mes model.Message) {
	msg := tgbotapi.NewMessage(int64(mes.ChatTo.ID), mes.UserFrom+mes.Data)
	_, err := tb.api.Send(msg)
	if err != nil {
		return
	}
}
