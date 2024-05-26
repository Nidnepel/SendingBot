package main

import (
	"log"
	"os"
	"fmt"

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
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	tgToken := os.Getenv("TG_TOKEN")
	vkToken := os.Getenv("VK_TOKEN")
	callbackSecret := os.Getenv("CALLBACK_VK_SECRET")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := database.New(connStr)
	if err != nil {
		log.Fatalf("невозможно подключиться к базе: %v", err)
	}

	repo := repository.New(database.NewPGX(db))
	uc := usecase.New(repo)
	tgBot := tg.New(uc, tgToken)
	vkBot := vk.New(uc, vkToken, callbackSecret)
	uc.AddBot("vk", vkBot)
	uc.AddBot("tg", tgBot)
	vkBot.Run()
	tgBot.Run()
}
