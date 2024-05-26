package vk

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"strings"

	"errors"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/SevereCloud/vksdk/v2/callback"

	"github.com/Nidnepel/SendingBot/internal/bot_client"
	"github.com/Nidnepel/SendingBot/internal/model"
	"github.com/Nidnepel/SendingBot/internal/usecase"
)

type Bot struct {
	api *api.VK
	uc  usecase.Usecase
	callbackSecret string
}

func New(uc usecase.Usecase, token string, callbackSecret string) bot_client.Bot {
	bot := api.NewVK(token)

	return &Bot{api: bot, uc: uc}
}

func (b *Bot) Run() {
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		var callbackEvent events.GroupEvent
		if err := json.NewDecoder(r.Body).Decode(&callbackEvent); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if callbackEvent.Secret != b.callbackSecret {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch callbackEvent.Type {
		case events.EventMessageNew:
			var obj events.MessageNewObject
			err := json.Unmarshal(callbackEvent.Object, &obj)
			if err != nil {
				log.Printf("Error unmarshaling MessageNew object: %v", err)
				return
			}
			go b.routeMessage(context.Background(), obj) 
			fmt.Fprint(w, "ok")
		case events.EventConfirmation:
			fmt.Fprint(w, "54kmlkBf7vf98Nfdvk4FKML43")
		}
	})

	log.Fatal(http.ListenAndServe(":8443", nil)) 
}

func (b *Bot) routeMessage(ctx context.Context, obj events.MessageNewObject) {
	mes := transformFromVKMessageToModel(obj.Message)
	ansMes := model.Message{
		Data: model.Data{},
		ChatTo: model.Chat{
			ID: mes.ChatFrom.ID,
		},
	}
	args := strings.Split(mes.Data.Text, " ")
	switch args[0] {
	case "/key_in":
		if len(args) != 2 {
			ansMes.Data.Text = bot_client.IncorrectInput
			b.Send(ansMes)
			return
		}

		mes.ChatFrom.Key = args[1]
		key, err := b.uc.KeyIn(ctx, &mes.ChatFrom)
		if err != nil {
			if errors.Is(err, usecase.ErrKeyNotExist) {
				ansMes.Data.Text = bot_client.KeyNotExist
				b.Send(ansMes)
				return
			}
			ansMes.Data.Text = fmt.Sprintf(bot_client.KeyAlreadyExist, key)
			b.Send(ansMes)
		}
	case "/key_gen":
		key, err := b.uc.KeyGen(ctx, &mes.ChatFrom)
		if err != nil {
			ansMes.Data.Text = fmt.Sprintf(bot_client.KeyAlreadyExist, key)
			b.Send(ansMes)
			return
		}

		ansMes.Data.Text = fmt.Sprintf(bot_client.KeyGenSuccess, key)
		b.Send(ansMes)
	case "/key_drop":
		err := b.uc.KeyDrop(ctx, &mes.ChatFrom)
		if err != nil {
			ansMes.Data.Text = bot_client.KeyDropFailed
			b.Send(ansMes)
		} else {
			ansMes.Data.Text = bot_client.KeyDropSuccess
			b.Send(ansMes)
		}
	default:
		userFrom := b.getUserName(obj.Message.FromID)
		mes.UserFrom = userFrom
		err := b.uc.Send(ctx, mes)
		if err != nil {
			log.Printf("failed to send message: %s", err)
			return
		}
	}
}

func (b *Bot) Send(mes model.Message) {
	_, err := b.api.MessagesSend(api.Params{
		"peer_id":    mes.ChatTo.ID,
		"message":    mes.UserFrom + mes.Data.Text,
		"attachment": createAttachment(mes),
		"random_id":  0,
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

func transformFromVKMessageToModel(message object.MessagesMessage) *model.Message {
	mes := model.Message{
		ChatFrom: model.Chat{
			Messenger: "vk",
			ID:        message.PeerID,
		},
	}

	mes.Data = model.Data{
		Text: message.Text,
	}

	for _, attachment := range message.Attachments {
		switch attachment.Type {
		case "photo":
			mes.Data.AddPhoto(getPhotoUrl(attachment.Photo))
		case "doc":
			if attachment.Doc.Ext == "gif" {
				mes.Data.AddGif(attachment.Doc.URL)
			} else {
				mes.Data.AddFile(attachment.Doc.URL)
			}
		}
	}

	return &mes
}

func getPhotoUrl(photo object.PhotosPhoto) string {
	photoSizes := photo.Sizes
	if len(photoSizes) == 0 {
		return ""
	}
	photoUrl := photoSizes[len(photoSizes)-1].URL
	return photoUrl
}

func createAttachment(mes model.Message) string {
	attachment := ""
	for _, photoUrl := range mes.Data.Photos {
		attachment += fmt.Sprintf("%s,", photoUrl)
	}
	for _, gifUrl := range mes.Data.Gif {
		attachment += fmt.Sprintf("%s,", gifUrl)
	}
	for _, docUrl := range mes.Data.Doc {
		attachment += fmt.Sprintf("%s,", docUrl)
	}

	if len(attachment) != 0 {
		return attachment[:len(attachment)-1]
	}

	return ""
}
