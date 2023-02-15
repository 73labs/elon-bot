package handlers

import (
	"fmt"
	"log"
	"runtime"
	"seventythree/chatbot/commands"

	"github.com/bwmarrin/discordgo"
)

type StatCommands struct {
}

func (h *StatCommands) GetCommands() commands.Map[StatCommands] {
	return statCommandMap
}

func RegisterNewStatCommands(s *discordgo.Session, l *log.Logger) (*commands.Context[StatCommands], error) {
	h := &commands.Context[StatCommands]{
		Session: s,
		Logger:  l,
		Handler: &StatCommands{},
	}

	if err := h.RegisterCommands(); err != nil {
		return nil, err
	}

	h.Logger.Print("stat commands registered")
	return h, nil
}

func GetStats(h *commands.Context[StatCommands], i *discordgo.InteractionCreate) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Memory usage: %d/%d MiB", m.Alloc/1000000, m.TotalAlloc/1000000),
		},
	})

}

var (
	statCommandMap = commands.Map[StatCommands]{
		"stats": {
			Command: &discordgo.ApplicationCommand{
				Name:        "stats",
				Description: "Gives you infos about the current instance of the bot",
			},
			Handler: GetStats,
		},
	}
)
