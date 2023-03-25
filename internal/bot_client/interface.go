package bot_client

import "github.com/Nidnepel/SendingBot/internal/model"

type Bot interface {
	Run()
	Send(mes model.Message)
}

const (
	IncorrectInput = "Команда ввода ключа должна иметь формат: /key_in <key>"
	KeyNotExist = "Такого ключа не существует"
	KeyAlreadyExist = "Для этого чата уже установлен ключ: %s\n, вы можете его снять через команду /key_drop"
	KeyGenSuccess = "Ключ присоединения к этому чату: %s"
	KeyDropFailed = "Не удалось сбросить ключ для чата"
	KeyDropSuccess = "Ключ для чата сброшен"
)