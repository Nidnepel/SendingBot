package main

import (
	"fmt"
	"github.com/Nidnepel/SendingBot/internal/bot_client/tg"
	"github.com/Nidnepel/SendingBot/internal/bot_client/vk"
	"github.com/Nidnepel/SendingBot/internal/database"
	"github.com/Nidnepel/SendingBot/internal/repository"
	"github.com/Nidnepel/SendingBot/internal/usecase"
	"log"
)

func main() {
	db, err := database.New(
		fmt.Sprintf("postgres://test:test@localhost/test?sslmode=disable"),
	)
	if err != nil {
		log.Fatalf("невозможно подключиться к базе: %v", err)
	}

	repo := repository.New(database.NewPGX(db))
	uc := usecase.New(repo)
	tgBot := tg.New(uc)
	vkBot := vk.New(uc)
	uc.AddBot("vk", vkBot)
	uc.AddBot("tg", tgBot)
	vkBot.Run()
	tgBot.Run()
}
