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

func GetHandler(name string) (Command, bool) {
	cmd, ok := registry[name]
	return cmd, ok
}

func All() []*discordgo.ApplicationCommand {
	out := make([]*discordgo.ApplicationCommand, 0, len(registry))
	for name, cmd := range registry {
		ac := &discordgo.ApplicationCommand{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Options:     cmd.Options(),
		}
		if name == "ticketconfig" {
			perm := int64(discordgo.PermissionAdministrator)
			ac.DefaultMemberPermissions = &perm
			dm := false
			ac.DMPermission = &dm
		}
		out = append(out, ac)
	}
	return out
}
