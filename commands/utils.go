package commands

import (
	"github.com/bwmarrin/discordgo"
)

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func respondErr(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	respond(s, i, "❌ "+msg)
}

func memberHasRole(m *discordgo.Member, roleID string) bool {
	for _, r := range m.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}
