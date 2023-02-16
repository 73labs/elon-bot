package sessions

import (
	"strings"
	"sync"
	"time"
)

type DiscordEntity struct {
	Name string
	ID   string
}

var (
	SessionDefaultDuration                  = time.Minute * 20
	SessionDefaultNoInteractionTimeout      = time.Minute * 5
	SessionDefaultMessageLimit         uint = 100
)

type Message struct {
	Content    string
	ByBot      bool
	AuthorName string
}

type Session struct {
	Creator         DiscordEntity
	Channel         DiscordEntity
	Guild           DiscordEntity
	InDM            bool
	LastStart       time.Time
	LastInteraction time.Time
	Covered         time.Duration
	SentCount       uint
	SessionLevel    uint8
	UsersInChat     []DiscordEntity
	timedOut        bool
	messages        []Message
	mutex           sync.Mutex
}

func NewSession(creator DiscordEntity, channel DiscordEntity, guild DiscordEntity, inDM bool) *Session {
	now := time.Now().UTC()

	return &Session{
		Creator:         creator,
		Channel:         channel,
		Guild:           guild,
		InDM:            inDM,
		LastStart:       now,
		LastInteraction: now,
		messages:        make([]Message, SessionDefaultMessageLimit),
		mutex:           sync.Mutex{},
		UsersInChat:     []DiscordEntity{creator},
	}
}

func (s *Session) HasTimedOut() bool {
	if s.timedOut {
		return true
	}

	now := time.Now().UTC()
	if s.SentCount >= SessionDefaultMessageLimit ||
		now.Sub(s.LastInteraction) >= SessionDefaultDuration ||
		now.Sub(s.LastInteraction) >= SessionDefaultNoInteractionTimeout {
		s.timedOut = true
		return true
	}

	return false
}

func (s *Session) ResetTimeout() {
	s.Covered = 0
	s.SentCount = 0
	s.timedOut = false
}

func (s *Session) AddUserMessage(username string, content string) {
	s.mutex.Lock()
	s.messages[s.SentCount] = Message{
		Content:    content,
		ByBot:      false,
		AuthorName: username,
	}
	s.SentCount++
	s.mutex.Unlock()
}

func (s *Session) AddBotMessage(content string) Message {
	s.mutex.Lock()
	message := Message{
		Content: content,
		ByBot:   true,
	}

	s.messages[s.SentCount] = message
	s.SentCount++

	s.mutex.Unlock()
	return message
}

func (s *Session) GetSessionProtocol() (string, error) {
	s.mutex.Lock()
	sb := strings.Builder{}
	for i := uint(0); i < s.SentCount; i++ {
		var message = s.messages[i]
		var user string
		if message.ByBot {
			user = "Elon Musk: "
		} else {
			user = message.AuthorName + ": "
		}

		_, err := sb.WriteString(user + message.Content + "\n")
		if err != nil {
			s.mutex.Unlock()
			return "", err
		}
	}

	s.mutex.Unlock()
	return sb.String(), nil
}
