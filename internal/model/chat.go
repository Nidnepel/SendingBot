package model

type Chat struct {
	ID        int    `db:"id"`
	Key       string `db:"chat_key"`
	Messenger string `db:"messenger"`
}
