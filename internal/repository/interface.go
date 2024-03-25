package repository

import (
	"context"
	"github.com/Nidnepel/SendingBot/internal/model"
)

type Repository interface {
	SetKeyForChat(ctx context.Context, chat *model.Chat) error
	GetChatKey(ctx context.Context, chat *model.Chat) (string, error)
	ClearKeyForChat(ctx context.Context, chat *model.Chat) error
	GetLinkedChats(ctx context.Context, key string) ([]model.Chat, error)
}
