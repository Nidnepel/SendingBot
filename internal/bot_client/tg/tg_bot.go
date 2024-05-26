package tg

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"strings"

	"github.com/Nidnepel/SendingBot/internal/bot_client"
	"github.com/Nidnepel/SendingBot/internal/model"
	"github.com/Nidnepel/SendingBot/internal/usecase"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api   *tgbotapi.BotAPI
	uc    usecase.Usecase
	token string
}

func New(uc usecase.Usecase, token string) bot_client.Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	return &Bot{api: bot, uc: uc, token: token}
}

func (tb *Bot) Run() {
	whURL := "https://fdvmk23f.ngrok.io/sending_bot"
	wh, _ := tgbotapi.NewWebhookWithCert(whURL, nil)
	_, err := tb.api.SetWebhook(wh)
	if err != nil {
		log.Fatalf("Failed to set webhook: %v", err)
	}


	info, err := tb.api.GetWebhookInfo()
	if err != nil {
		log.Fatalf("Failed to get webhook info: %s", err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram webhook failed, last error: %s", info.LastErrorMessage)
	}

	http.HandleFunc("/sending_bot", func(w http.ResponseWriter, r *http.Request) {
		var update tgbotapi.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if mes := update.Message; mes != nil { 
			go tb.routeMessage(mes)
		}
	})

	log.Fatal(http.ListenAndServe(":8444", nil)) 
}

func (tb *Bot) routeMessage(message *tgbotapi.Message) {
	ctx := context.Background()
	mes := tb.transformFromMessageToModel(message)
	ansMes := model.Message{
		Data: model.Data{},
		ChatTo: model.Chat{
			ID: mes.ChatFrom.ID,
		},
	}
	args := strings.Split(message.Text, " ")
	switch args[0] {
	case "/key_in":
		if len(args) != 2 {
			if len(args) != 2 {
				ansMes.Data.Text = bot_client.IncorrectInput
				tb.Send(ansMes)
				return
			}
			return
		}
		mes.ChatFrom.Key = args[1]
		key, err := tb.uc.KeyIn(ctx, &mes.ChatFrom)
		if err != nil {
			if errors.Is(err, usecase.ErrKeyNotExist) {
				ansMes.Data.Text = bot_client.KeyNotExist
				tb.Send(ansMes)
				return
			}
			ansMes.Data.Text = fmt.Sprintf(bot_client.KeyAlreadyExist, key)
			tb.Send(ansMes)
		}
	case "/key_gen":
		key, err := tb.uc.KeyGen(ctx, &mes.ChatFrom)
		if err != nil {
			ansMes.Data.Text = fmt.Sprintf(bot_client.KeyAlreadyExist, key)
			tb.Send(ansMes)
			return
		}

		ansMes.Data.Text = fmt.Sprintf(bot_client.KeyGenSuccess, key)
		tb.Send(ansMes)
	case "/key_drop":
		err := tb.uc.KeyDrop(ctx, &mes.ChatFrom)
		if err != nil {
			ansMes.Data.Text = bot_client.KeyDropFailed
			tb.Send(ansMes)
		} else {
			ansMes.Data.Text = bot_client.KeyDropSuccess
			tb.Send(ansMes)
		}
	default:
		err := tb.uc.Send(ctx, mes)
		if err != nil {
			log.Printf("failed to send message: %s", err)
			return
		}
	}
}

func (tb *Bot) transformFromMessageToModel(message *tgbotapi.Message) *model.Message {
	mes := model.Message{
		ChatFrom: model.Chat{
			Messenger: "tg",
			ID:        int(message.Chat.ID),
		},
		UserFrom: fmt.Sprintf("%s %s:\n", message.From.FirstName, message.From.LastName),
	}

	mes.Data = model.Data{
		Text: message.Text,
	}

	if len(message.Photo) > 0 {
		photo := message.Photo[len(message.Photo)-1]
		fileID := photo.FileID
		mes.Data.AddPhoto(tb.GetUrl(fileID))
	}

	if message.Document != nil {
		fileID := message.Document.FileID
		if message.Document.MimeType == "image/gif" {
			mes.Data.AddGif(tb.GetUrl(fileID))
		} else {
			mes.Data.AddFile(tb.GetUrl(fileID))
		}
	}

	return &mes
}

func (tb *Bot) GetUrl(fieldID string) string {
	fileConfig := tgbotapi.FileConfig{FileID: fieldID}
	file, err := tb.api.GetFile(fileConfig)
	if err != nil {
		log.Printf("failed to get url: %s", err)
		return ""
	}

	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s",
		tb.token, file.FilePath)
}

func (tb *Bot) Send(mes model.Message) {
	for _, tgMsg := range ConvertModelToTg(mes) {
		_, err := tb.api.Send(tgMsg)
		if err != nil {
			log.Print(err)
		}
	}
}

func ConvertModelToTg(mes model.Message) []tgbotapi.Chattable {
	messages := make([]tgbotapi.Chattable, 0)
	chatID := int64(mes.ChatTo.ID)

	messages = append(messages, tgbotapi.NewMessage(chatID, mes.UserFrom+mes.Data.Text))

	for _, photoUrl := range mes.Data.Photos {
		photoFile := tgbotapi.FileURL(photoUrl)
		messages = append(messages, tgbotapi.NewPhoto(chatID, photoFile))
	}
	for _, gifUrl := range mes.Data.Gif {
		gifFile := tgbotapi.FileURL(gifUrl)
		messages = append(messages, tgbotapi.NewAnimation(chatID, gifFile))
	}
	for _, docUrl := range mes.Data.Doc {
		docFile := tgbotapi.FileURL(docUrl)
		messages = append(messages, tgbotapi.NewAnimation(chatID, docFile))
	}

	return messages
}
