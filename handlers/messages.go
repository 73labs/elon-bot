package handlers

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"seventythree/chatbot/sessions"
	"strings"
)

type Sessions interface {
	GetSession(channelID string) *sessions.Session
	EndSession(channelID string)
}

type BotClient interface {
	ReadMessagesAndRespond(s *sessions.Session) (sessions.Message, error)
}

type MessageHandler struct {
	logger   *log.Logger
	sessions Sessions
	bot      BotClient
}

func RegisterNewMessageHandler(s *discordgo.Session, m Sessions, b BotClient, l *log.Logger) *MessageHandler {
	handler := &MessageHandler{
		logger:   l,
		sessions: m,
		bot:      b,
	}
	s.AddHandler(handler.HandleMessage)
	return handler
}

func (h *MessageHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	session := h.sessions.GetSession(m.ChannelID)
	if session == nil {
		return
	}

	var inChat bool
	for _, user := range session.UsersInChat {
		if m.Author.ID == user.ID {
			inChat = true
			break
		}
	}

	if !inChat {
		session.UsersInChat = append(session.UsersInChat, sessions.DiscordEntity{Name: m.Author.Username, ID: m.Author.ID})
	}

	if len(session.UsersInChat) > 1 {
		var mentions bool
		for _, mention := range m.Mentions {
			if mention.ID == s.State.User.ID {
				mentions = true
				break
			}
		}

		if strings.Contains(m.Content, "elon") {
			mentions = true
		}

		if !mentions {
			return
		}
	}

	h.logger.Printf("user %s#%s said %s to bot", m.Author.Username, m.Author.Discriminator, m.Content)

	go func() {
		session.AddUserMessage(m.Author.Username, m.Content)
		response, err := h.bot.ReadMessagesAndRespond(session)
		if err != nil {
			h.logger.Println("error happened while bot was answering:", err)
			_, _ = s.ChannelMessageSend(m.ChannelID, "The bot had issues responding. The session is now closed.")
			h.sessions.EndSession(m.ChannelID)
			return
		}

		response.Content = strings.Replace(response.Content, m.Author.Username, m.Author.Username+" ", -1)
		_, err = s.ChannelMessageSend(m.ChannelID, response.Content)
		if err != nil {
			h.logger.Println("error happened while bot was trying to respond:", err)
			h.sessions.EndSession(m.ChannelID)
			return
		}
	}()
}
