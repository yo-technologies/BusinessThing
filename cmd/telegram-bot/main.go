package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	webAppURL := os.Getenv("WEBAPP_URL")

	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}
	if webAppURL == "" {
		log.Fatal("WEBAPP_URL environment variable is required")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Debug = false
	log.Printf("Authorized as @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down bot...")
		cancel()
	}()

	log.Println("Bot started. Waiting for updates...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot stopped")
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				handleCommand(bot, update.Message, webAppURL)
			}
		}
	}
}

func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, webAppURL string) {
	switch msg.Command() {
	case "start":
		text := fmt.Sprintf(
			"Привет, %s! 👋\n\n"+
				"Добро пожаловать в BusinessThing — твой личный бизнес-ассистент.\n\n"+
				"Нажми на кнопку ниже, чтобы открыть приложение:",
			msg.From.FirstName,
		)

		button := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonWebApp("🚀 Открыть приложение", tgbotapi.WebAppInfo{URL: webAppURL}),
			),
		)

		reply := tgbotapi.NewMessage(msg.Chat.ID, text)
		reply.ReplyMarkup = button

		if _, err := bot.Send(reply); err != nil {
			log.Printf("Failed to send message: %v", err)
		}

	case "help":
		text := "Команды бота:\n\n" +
			"/start - Открыть приложение\n" +
			"/help - Показать эту справку"

		reply := tgbotapi.NewMessage(msg.Chat.ID, text)
		if _, err := bot.Send(reply); err != nil {
			log.Printf("Failed to send message: %v", err)
		}

	default:
		text := "Используй /start, чтобы открыть приложение."
		reply := tgbotapi.NewMessage(msg.Chat.ID, text)
		if _, err := bot.Send(reply); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}
