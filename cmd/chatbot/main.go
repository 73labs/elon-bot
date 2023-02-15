package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"seventythree/chatbot/handlers"
	"seventythree/chatbot/openai"
	"seventythree/chatbot/sessions"
	"syscall"

	gogpt "github.com/sashabaranov/go-gpt3"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	BotTokenEnv     = "DISCORD_BOT_TOKEN"
	OpenAiAPIKeyEnv = "OPENAI_API_KEY"
)

func handleError(err error, serviceName string) {
	if err != nil {
		log.Panicf("%s returned an error on startup: %v", serviceName, err)
	}
}

func setupHandlers(session *discordgo.Session) {
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("logged in as: %v#%s", s.State.User.Username, s.State.User.Discriminator)
	})
}

func main() {
	logger := log.Default()

	err := godotenv.Load()
	handleError(err, "loading environment variables")

	botToken := os.Getenv(BotTokenEnv)
	if botToken == "" {
		handleError(fmt.Errorf("%s was not set", BotTokenEnv), "getting environment variable")
	}

	openAiApiKey := os.Getenv(OpenAiAPIKeyEnv)
	if openAiApiKey == "" {
		handleError(fmt.Errorf("%s was not set", OpenAiAPIKeyEnv), "getting environment variable")
	}

	discordSession, err := discordgo.New(fmt.Sprintf("Bot %s", botToken))
	handleError(err, "creating discord client")
	setupHandlers(discordSession)

	err = discordSession.Open()
	handleError(err, "opening gateway socket connection")

	sessionManager := sessions.NewSessionManager(logger)

	sessionHandler, err := handlers.RegisterNewSessionCommands(discordSession, sessionManager, logger)
	handleError(err, "registering session commands")
	statHandler, err := handlers.RegisterNewStatCommands(discordSession, logger)
	handleError(err, "registering stat commands")

	goGpt := gogpt.NewClient(openAiApiKey)
	bot := openai.NewBotClient(goGpt, logger)
	_ = handlers.RegisterNewMessageHandler(discordSession, sessionManager, bot, logger)

	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, syscall.SIGINT, syscall.SIGABRT)
	log.Println("press ^C to exit")

	<-cancel

	fmt.Println()
	log.Println("shutting down ...")
	sessionHandler.DeleteCommands()
	statHandler.DeleteCommands()
}
