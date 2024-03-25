package usecase

import (
	"context"
	"github.com/Nidnepel/SendingBot/internal/bot_client"
	"github.com/Nidnepel/SendingBot/internal/model"
)

type Usecase interface {
	Send(ctx context.Context, mes *model.Message) error
	KeyIn(ctx context.Context, chat *model.Chat) error
	KeyGen(ctx context.Context, chat *model.Chat) (string, error)
	AddBot(botName string, bot bot_client.Bot)
}
