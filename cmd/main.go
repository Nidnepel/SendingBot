package main

import (
	"log"
	"os"

	"github.com/Nidnepel/SendingBot/internal/bot_client/tg"
	"github.com/Nidnepel/SendingBot/internal/bot_client/vk"
	"github.com/Nidnepel/SendingBot/internal/database"
	"github.com/Nidnepel/SendingBot/internal/repository"
	"github.com/Nidnepel/SendingBot/internal/usecase"
	"github.com/joho/godotenv"
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}


func main() {
	db, err := database.New("postgres://test:test@localhost/test?sslmode=disable")
	if err != nil {
		log.Fatalf("невозможно подключиться к базе: %v", err)
	}

	repo := repository.New(database.NewPGX(db))
	uc := usecase.New(repo)
	tgToken := os.Getenv("TG_TOKEN")
	vkToken := os.Getenv("VK_TOKEN")
	tgBot := tg.New(uc, tgToken)
	vkBot := vk.New(uc, vkToken)
	uc.AddBot("vk", vkBot)
	uc.AddBot("tg", tgBot)
	vkBot.Run()
	tgBot.Run()
}
