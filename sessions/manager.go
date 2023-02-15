package sessions

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrSessionInProgress = errors.New("a session is already in progress in given channel")
)

// SessionManager handles creation, deletion and persistence of sessions.
type SessionManager struct {
	logger   *log.Logger
	sessions map[string]*Session
}

func NewSessionManager(l *log.Logger) *SessionManager {
	return &SessionManager{
		logger:   l,
		sessions: map[string]*Session{},
	}
}

type SessionOptions struct {
	channel DiscordEntity
	creator DiscordEntity
	guild   DiscordEntity
	inDM    bool
}

// InChannel creates a *SessionOptions with the channel set.
func InChannel(channel *discordgo.Channel) *SessionOptions {
	return &SessionOptions{
		channel: DiscordEntity{
			ID:   channel.ID,
			Name: channel.Name,
		},
	}
}

func (o *SessionOptions) IsDMInteraction(user *discordgo.User) *SessionOptions {
	o.creator = DiscordEntity{
		Name: user.Username,
		ID:   user.ID,
	}
	o.inDM = true
	return o
}

func (o *SessionOptions) IsGuildInteraction(member *discordgo.Member, guild *discordgo.Guild) *SessionOptions {
	o.guild = DiscordEntity{
		Name: guild.Name,
		ID:   guild.ID,
	}
	o.creator = DiscordEntity{
		Name: member.User.Username,
		ID:   member.User.ID,
	}
	return o
}

// PushNewSession creates a session, if it is possible for given options and sets it to the session cache.
func (m *SessionManager) PushNewSession(opts SessionOptions) error {
	if session, has := m.sessions[opts.channel.ID]; has {
		if !session.HasTimedOut() {
			return ErrSessionInProgress
		}

		delete(m.sessions, opts.channel.ID)
	}

	s := NewSession(opts.creator, opts.channel, opts.guild, opts.inDM)

	m.sessions[opts.channel.ID] = s
	return nil
}

func (m *SessionManager) Store(s *Session) error {
	return nil
}

// GetSession returns a *Session or nil if no session has been started in given channel.
func (m *SessionManager) GetSession(channelID string) *Session {
	s, has := m.sessions[channelID]
	if !has {
		return nil
	}

	if s.HasTimedOut() {
		m.logger.Printf("session %s has timed out", s.Channel.ID)
		delete(m.sessions, channelID)
		go func() {
			err := m.Store(s)
			if err != nil {
				m.logger.Fatal("error happened when attempting to store session:", err)
			}
		}()

		return nil
	}

	return s
}

func (m *SessionManager) EndSession(channelID string) {
	if _, has := m.sessions[channelID]; !has {
		return
	}

	delete(m.sessions, channelID)
}
