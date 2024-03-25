package vk

import (
	"context"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/object"
	"log"
	"strings"

	"github.com/Nidnepel/SendingBot/internal/bot_client"
	"github.com/Nidnepel/SendingBot/internal/model"
	"github.com/Nidnepel/SendingBot/internal/usecase"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	longpoll "github.com/SevereCloud/vksdk/v2/longpoll-bot"
)

type Bot struct {
	api *api.VK
	uc  usecase.Usecase
}

func New(uc usecase.Usecase) bot_client.Bot {
	bot := api.NewVK("token")

	return &Bot{api: bot, uc: uc}
}

func (b *Bot) Run() {
	// Получение информации о группе
	group, err := b.api.GroupsGetByID(nil)
	if err != nil {
		log.Fatalf("Ошибка при получении информации о группе: %v", err)
	}

	// Создание объекта Long Poll
	lp, err := longpoll.NewLongPoll(b.api, group[0].ID)
	if err != nil {
		log.Fatalf("Ошибка при создании Long Poll: %v", err)
	}

	// Обработчик новых сообщений
	lp.MessageNew(b.routeMessage)

	// Запуск Long Poll
	log.Println("Запуск Long Poll сервера.")
	go func() {
		if err := lp.Run(); err != nil {
			log.Fatalf("Ошибка Long Poll: %v", err)
		}
	}()
}

func (b *Bot) routeMessage(ctx context.Context, obj events.MessageNewObject) {
	mes := obj.Message
	args := strings.Split(mes.Text, " ")
	switch args[0] {
	case "/key_in":
		if len(args) != 2 {
			return
		}
		chat := &model.Chat{
			Messenger: "vk",
			Key:       args[1],
			ID:        mes.PeerID,
		}
		err := b.uc.KeyIn(ctx, chat)
		if err != nil {
			// todo
			return
		}
	case "/key_gen":
		key, err := b.uc.KeyGen(ctx, &model.Chat{
			Messenger: "vk",
			ID:        mes.PeerID,
		})
		if err != nil {
			//todo
			return
		}
		b.Send(model.Message{
			Data: fmt.Sprintf("your link key: %s", key),
			ChatTo: model.Chat{
				ID: mes.PeerID,
			},
		})
	default:
		userFrom := b.getUserName(mes.FromID)
		mess := transformFromMessageToModel(mes)
		mess.UserFrom = userFrom
		err := b.uc.Send(ctx, mess)
		if err != nil {
			return
		}
	}
}

func transformFromMessageToModel(message object.MessagesMessage) *model.Message {
	return &model.Message{
		Data: message.Text,
		ChatFrom: model.Chat{
			Messenger: "vk",
			ID:        message.PeerID,
		},
	}
}

func (b *Bot) Send(mes model.Message) {
	_, err := b.api.MessagesSend(api.Params{
		"peer_id":   mes.ChatTo.ID,
		"message":   mes.UserFrom + mes.Data,
		"random_id": 0,
	})
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func (b *Bot) getUserName(userID int) string {
	users, err := b.api.UsersGet(api.Params{
		"user_ids": userID,
	})

	if err != nil {
		log.Fatalf("Ошибка при получении данных пользователя: %v", err)
	}

	if len(users) > 0 {
		user := users[0]
		return fmt.Sprintf("%s %s:\n", user.FirstName, user.LastName)
	} else {
		return ""
	}
}
