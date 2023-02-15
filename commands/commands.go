package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// HandlerFunc describes the signature of a command handler.
type HandlerFunc[HandlerType interface{}] func(h *Context[HandlerType], i *discordgo.InteractionCreate)

// CommandHandler interface is passed to the handler functions. It's implementation can contain any objects
// that should be exposed to the handler functions.
type CommandHandler[HandlerType interface{}] interface {
	GetCommands() Map[HandlerType]
}

// Command describes a command with its signature that is used to register it on Discord and its handler function.
type Command[HandlerType any] struct {
	Command *discordgo.ApplicationCommand
	Handler HandlerFunc[HandlerType]
}

// Map is a map that has the command names as keys and points to their corresponding Commands.
type Map[HandlerType any] map[string]Command[HandlerType]

// Context is responsible for:
//
// - Routing commands to their corresponding handler functions.
//
// - Creating commands on Discord and holding references to them until they are deleted.
type Context[HandlerType interface{}] struct {
	Handler            CommandHandler[HandlerType]
	Session            *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
	Logger             *log.Logger
}

func (h *Context[T]) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command, has := h.Handler.GetCommands()[i.ApplicationCommandData().Name]
	if !has {
		return
	}

	var username, discriminator string
	if i.GuildID == "" {
		username, discriminator = i.User.Username, i.User.Discriminator
	} else {
		username, discriminator = i.Member.User.Username, i.Member.User.Discriminator
	}

	h.Logger.Printf("command /%s was executed by user %s#%s", command.Command.Name, username, discriminator)
	command.Handler(h, i)
}

// RegisterCommands uses the CommandMap that is retrieved from the SubCommandHandler and registers the commands on Discords. It stores
// the ID's that are assigned to them by Discord until DeleteCommands is called.
func (h *Context[T]) RegisterCommands() error {
	commands := h.Handler.GetCommands()
	h.registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	i := 0
	for _, cmd := range commands {
		cmdRes, err := h.Session.ApplicationCommandCreate(h.Session.State.Application.ID, "", cmd.Command)
		if err != nil {
			return err
		}

		h.registeredCommands[i] = cmdRes
		i++
	}

	h.Session.AddHandler(h.handleCommand)
	return nil
}

// DeleteCommand deletes the commands that have been created when RegisterCommands was called.
func (h *Context[HandlerType]) DeleteCommands() {
	for _, command := range h.registeredCommands {
		err := h.Session.ApplicationCommandDelete(h.Session.State.Application.ID, "", command.ID)
		if err != nil {
			h.Logger.Fatalf("Could not delete command %s, %s", command.Name, command.ID)
		}
	}
}

// HandleError responds to an interaction with a generic error message.
func (h *Context[HandlerType]) HandleError(i *discordgo.Interaction) {
	h.Session.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "An error occurred. I cannot serve you at this time. Sorry.",
		},
	})
}
