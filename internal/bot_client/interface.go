package bot_client

import "github.com/Nidnepel/SendingBot/internal/model"

type Bot interface {
	Run()
	Send(mes model.Message)
}
