// commands/ping.go
package commands

import "github.com/bwmarrin/discordgo"

func init() {
	Register(&Ping{})
}

type Ping struct{}

func (p *Ping) Name() string {
	return "ping"
}

func (p *Ping) Description() string {
	return "Replies with Pong!"
}

func (p *Ping) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (p *Ping) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🏓 Pong!",
		},
	})
}
