package openai

import (
	"context"
	gogpt "github.com/sashabaranov/go-gpt3"
	"log"
	"seventythree/chatbot/sessions"
	"strings"
)

type BotClient struct {
	client *gogpt.Client
	logger *log.Logger
}

func NewBotClient(client *gogpt.Client, logger *log.Logger) *BotClient {
	return &BotClient{
		client: client,
		logger: logger,
	}
}

func (c *BotClient) ReadMessagesAndRespond(s *sessions.Session) (sessions.Message, error) {
	sb := strings.Builder{}
	_, err := sb.WriteString("You are Elon Musk in a Discord conversation. \n" +
		"Server: " + s.Guild.Name + "\n" +
		"Session Owner: " + s.Creator.Name + "\n")
	if err != nil {
		return sessions.Message{}, err
	}

	protocol, err := s.GetSessionProtocol()
	if err != nil {
		return sessions.Message{}, err
	}

	sb.WriteString(protocol)
	sb.WriteString("Elon Musk: ")

	resp, err := c.client.CreateCompletion(context.Background(), gogpt.CompletionRequest{
		Model:     gogpt.GPT3TextDavinci003,
		MaxTokens: 256,
		Prompt:    sb.String(),
	})
	if err != nil {
		return sessions.Message{}, err
	}

	c.logger.Printf("davinci: %s", resp.Choices[0].Text)
	return s.AddBotMessage(resp.Choices[0].Text), nil
}
