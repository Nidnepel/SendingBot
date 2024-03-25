package model

type Message struct {
	Data     string
	UserFrom string
	ChatFrom Chat
	ChatTo   Chat
}
