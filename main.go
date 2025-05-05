package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/ds-dotsoft/go-bot-template/commands"
	"github.com/joho/godotenv"
)

func main() {
	repoName := "go-bot-template"
	err := godotenv.Load()
	if err != nil {
		log.Panicf("[%s]: Error while loading env file.", repoName)

	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatalf("[%s]: DISCORD_BOT_TOKEN not set", repoName)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("[%s]: Failed to create session: %v", repoName, err)
	}
	dg.Identify.Intents = discordgo.IntentsGuilds

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		name := i.ApplicationCommandData().Name
		if cmd, ok := commands.GetHandler(name); ok {
			cmd.Execute(s, i)
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Unknown command",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	})

	if err = dg.Open(); err != nil {
		log.Fatalf("[%s]: Error opening connection: %v", repoName, err)
	}
	defer dg.Close()

	appID := dg.State.User.ID
	guildID := os.Getenv("GUILD_ID")
	if _, err := dg.ApplicationCommandBulkOverwrite(appID, guildID, commands.All()); err != nil {
		log.Fatalf("[%s]: Cannot register commands: %v", repoName, err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
