package commands

import "github.com/bwmarrin/discordgo"

func init() {
	Register(&Ping{})
}

type Ping struct{}

// name returns the command name
func (p *Ping) Name() string {
	return "ping"
}

// description returns the command’s description
func (p *Ping) Description() string {
	return "Replies with Pong!"
}

// options returns nil since /ping has no parameters
func (p *Ping) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

// execute is called when /ping is invoked.
func (p *Ping) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🏓 Pong!",
		},
	}
	s.InteractionRespond(i.Interaction, resp)
}
