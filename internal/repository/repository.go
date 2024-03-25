package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/Nidnepel/SendingBot/internal/database"
	"github.com/Nidnepel/SendingBot/internal/model"
)

type Repo struct {
	db database.Queryable
}

func New(db database.Queryable) Repository {
	return &Repo{db: db}
}

func (r *Repo) SetKeyForChat(ctx context.Context, chat *model.Chat) error {
	err := r.getChat(ctx, chat)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		err = r.createChat(ctx, chat)
		if err != nil {
			return err
		}
	}

	err = r.setKeyForChat(ctx, chat)
	return err
}

func (r *Repo) setKeyForChat(ctx context.Context, chat *model.Chat) error {
	query := database.PSQL.Update(database.TableChat).
		Set("chat_key", chat.Key).
		Where(squirrel.And{
			squirrel.Eq{
				"id": chat.ID,
			},
			squirrel.Eq{
				"messenger": chat.Messenger,
			},
		})

	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *Repo) getChat(ctx context.Context, chat *model.Chat) error {
	query := database.PSQL.Select(
		"*",
	).
		From(database.TableChat).
		Where(squirrel.And{
			squirrel.Eq{
				"id": chat.ID,
			},
			squirrel.Eq{
				"messenger": chat.Messenger,
			},
		})

	var taken model.Chat
	err := r.db.Get(ctx, &taken, query)
	return err
}

func (r *Repo) createChat(ctx context.Context, chat *model.Chat) error {
	query := database.PSQL.Insert(database.TableChat).Columns(
		"id",
		"chat_key",
		"messenger",
	).Values(
		chat.ID,
		chat.Key,
		chat.Messenger,
	)

	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *Repo) ClearKeyForChat(ctx context.Context, chat *model.Chat) error {
	query := database.PSQL.Update(database.TableChat).
		Set("chat_key", "").
		Where(squirrel.And{
			squirrel.Eq{
				"id": chat.ID,
			},
			squirrel.Eq{
				"messenger": chat.Messenger,
			},
		})

	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *Repo) GetChatKey(ctx context.Context, chat *model.Chat) (string, error) {
	query := database.PSQL.Select(
		"chat_key",
	).
		From(database.TableChat).
		Where(squirrel.And{
			squirrel.Eq{
				"id": chat.ID,
			},
			squirrel.Eq{
				"messenger": chat.Messenger,
			},
		})

	var key string
	err := r.db.Get(ctx, &key, query)
	if err != nil {
		return "", fmt.Errorf("получение ключа: %w", err)
	}
	return key, nil
}

func (r *Repo) GetLinkedChats(ctx context.Context, key string) ([]model.Chat, error) {
	query := database.PSQL.
		Select(
			"id",
			"chat_key",
			"messenger",
		).
		From(database.TableChat).
		Where(
			squirrel.Eq{
				"chat_key": key,
			})

	var chats []model.Chat
	err := r.db.Select(ctx, &chats, query)
	if err != nil {
		return nil, fmt.Errorf("получение чатов по ключу: %w", err)
	}
	return chats, nil
}
