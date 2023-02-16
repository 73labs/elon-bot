package handlers

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"seventythree/chatbot/commands"
	"seventythree/chatbot/sessions"
)

var sessionCommandMap = commands.Map[SessionCommands]{
	"session": {
		Command: &discordgo.ApplicationCommand{
			Name:        "session",
			Description: "Starts a new session with the bot",
		},
		Handler: SessionStartCommand,
	},
	"end": {
		Command: &discordgo.ApplicationCommand{
			Name:        "end",
			Description: "Ends an active session",
		},
		Handler: SessionEndCommand,
	},
}

// SessionEndCommand handles the command /end to and terminates a running session
func SessionEndCommand(h *commands.Context[SessionCommands], i *discordgo.InteractionCreate) {
	handler, ok := h.Handler.(*SessionCommands)
	if !ok {
		h.HandleError(i.Interaction)
		return
	}

	channel, err := h.Session.Channel(i.ChannelID)
	if err != nil {
		h.HandleError(i.Interaction)
		return
	}

	session := handler.sessions.GetSession(channel.ID)
	if session == nil {
		_ = h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No session is active in this channel.",
			},
		})

		return
	}

	handler.sessions.EndSession(channel.ID)
	_ = h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "The session was ended.",
		},
	})

}

// SessionManager interface describes the services that provides the API to create, cancel and save sessions with the character.
type SessionManager interface {
	PushNewSession(opts sessions.SessionOptions) error
	EndSession(channelID string)
	GetSession(channelID string) *sessions.Session
}

// SessionCommands holds objects that must be accessible to session-command-handlers like the SessionManager.
type SessionCommands struct {
	sessions SessionManager
}

func (h *SessionCommands) GetCommands() commands.Map[SessionCommands] {
	return sessionCommandMap
}

// RegisterNewSessionCommands registers the following commands:
//
//	/session - Starts a new session wit the bot.
func RegisterNewSessionCommands(d *discordgo.Session, s SessionManager, l *log.Logger) (*commands.Context[SessionCommands], error) {
	context := &commands.Context[SessionCommands]{
		Session: d,
		Logger:  l,
		Handler: &SessionCommands{
			sessions: s,
		},
	}

	if err := context.RegisterCommands(); err != nil {
		return nil, err
	}

	context.Logger.Print("session commands registered")
	return context, nil
}

// SessionStartCommand handles the /session command.
func SessionStartCommand(h *commands.Context[SessionCommands], i *discordgo.InteractionCreate) {
	handler, ok := h.Handler.(*SessionCommands)
	if !ok {
		h.HandleError(i.Interaction)
		return
	}

	channel, err := h.Session.Channel(i.ChannelID)
	if err != nil {
		h.HandleError(i.Interaction)
		return
	}

	opts := sessions.InChannel(channel)
	if i.GuildID == "" {
		opts.IsDMInteraction(i.User)
	} else {
		guild, err := h.Session.Guild(i.GuildID)
		if err != nil {
			h.HandleError(i.Interaction)
			return
		}
		opts.IsGuildInteraction(i.Member, guild)
	}

	err = handler.sessions.PushNewSession(*opts)
	if err != nil {
		switch err {
		case sessions.ErrSessionInProgress:
			_ = h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Session is already in progress",
				},
			})
			return
		default:
			h.HandleError(i.Interaction)
			return
		}
	}

	err = h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "A session was started. Use /end to close it. It will close automatically if you stop interacting.",
		},
	})

	if err != nil {
		handler.sessions.EndSession(i.ChannelID)
	}
}
