package commands

import "github.com/bwmarrin/discordgo"

type Command interface {
	Name() string
	Description() string
	Options() []*discordgo.ApplicationCommandOption
	Execute(s *discordgo.Session, i *discordgo.InteractionCreate)
}

var registry = map[string]Command{}

func Register(cmd Command) {
	registry[cmd.Name()] = cmd
}

func All() []*discordgo.ApplicationCommand {
	list := make([]*discordgo.ApplicationCommand, 0, len(registry))
	for _, cmd := range registry {
		list = append(list, &discordgo.ApplicationCommand{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Options:     cmd.Options(),
		})
	}
	return list
}

func GetHandler(name string) (Command, bool) {
	cmd, ok := registry[name]
	return cmd, ok
}
