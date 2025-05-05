package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/ds-dotsoft/ds-ticket/commands"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment")
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN not set")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("failed to create session: %v", err)
	}
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			name := i.ApplicationCommandData().Name
			if cmd, ok := commands.GetHandler(name); ok {
				cmd.Execute(s, i)
			}
		case discordgo.InteractionMessageComponent:
			id := i.MessageComponentData().CustomID
			switch id {
			case "ticket_select":
				commands.HandleTicketSelect(s, i)
			case "ticket_claim":
				commands.HandleButtonClaim(s, i)
			case "ticket_close":
				commands.HandleButtonClose(s, i)
			}
		}
	})

	if err = dg.Open(); err != nil {
		log.Fatalf("error opening gateway: %v", err)
	}
	defer dg.Close()

	guild := os.Getenv("GUILD_ID")
	if _, err := dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, guild, commands.All()); err != nil {
		log.Fatalf("cannot register commands: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
