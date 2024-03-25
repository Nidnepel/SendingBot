package usecase

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"github.com/Nidnepel/SendingBot/internal/bot_client"
	"github.com/Nidnepel/SendingBot/internal/model"
	"github.com/Nidnepel/SendingBot/internal/repository"
)

var (
	ErrKeyExist = errors.New("key already exist")
)

type UC struct {
	repo repository.Repository
	bots map[string]bot_client.Bot
}

func New(repo repository.Repository) Usecase {
	bots := make(map[string]bot_client.Bot)
	return &UC{
		repo: repo,
		bots: bots,
	}
}

func (u *UC) AddBot(botName string, bot bot_client.Bot) {
	u.bots[botName] = bot
}

func (u *UC) KeyGen(ctx context.Context, chat *model.Chat) (string, error) {
	key, err := u.repo.GetChatKey(ctx, chat)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if key != "" {
		return "", ErrKeyExist
	}

	chat.Key = genKey()
	err = u.repo.SetKeyForChat(ctx, chat)
	if err != nil {
		return "", err
	}
	return chat.Key, nil
}

func genKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes)
}

func (u *UC) KeyIn(ctx context.Context, chat *model.Chat) error {
	currentKey, err := u.repo.GetChatKey(ctx, chat)
	if currentKey != "" {
		return ErrKeyExist
	}
	err = u.repo.SetKeyForChat(ctx, chat)
	if err != nil {
		return err
	}

	chats, err := u.repo.GetLinkedChats(ctx, chat.Key)
	if err != nil {
		return err
	}

	for botName, bot := range u.bots {
		for _, curchat := range chats {
			if curchat.Messenger == botName {
				bot.Send(model.Message{
					Data:   "Успешное соединение!",
					ChatTo: curchat,
				})
			}
		}
	}
	return nil
}

func (u *UC) Send(ctx context.Context, mes *model.Message) error {
	key, err := u.repo.GetChatKey(ctx, &mes.ChatFrom)
	if err != nil {
		return err
	}

	chats, err := u.repo.GetLinkedChats(ctx, key)
	if err != nil {
		return err
	}

	for botName, bot := range u.bots {
		if mes.ChatFrom.Messenger != botName {
			for _, chat := range chats {
				if chat.Messenger == botName {
					mes.ChatTo = chat
					bot.Send(*mes)
				}
			}
		}
	}
	return nil
}
