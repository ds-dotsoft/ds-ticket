// commands/ticketconfig.go
package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&TicketConfig{})
}

type TicketConfig struct{}

func (t *TicketConfig) Name() string {
	return "ticketconfig"
}

func (t *TicketConfig) Description() string {
	return "Manage ticket types, settings, and deploy the ticket prompt"
}

func (t *TicketConfig) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "add-type",
			Description: "Add a new ticket type",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Unique type name",
					Required:    true,
				},
				{
					Name:        "description",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Shown in the ticket menu",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "remove-type",
			Description: "Remove a ticket type (only if no tickets reference it)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Name of the type to remove",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list-types",
			Description: "List all configured ticket types",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "configure",
			Description: "Set open-category, support role & closed-category",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "category",
					Type:        discordgo.ApplicationCommandOptionChannel,
					Description: "Category for new tickets",
					Required:    true,
				},
				{
					Name:        "support_role",
					Type:        discordgo.ApplicationCommandOptionRole,
					Description: "Role that sees and can claim tickets",
					Required:    true,
				},
				{
					Name:        "closed_category",
					Type:        discordgo.ApplicationCommandOptionChannel,
					Description: "Category to archive closed tickets",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "prompt",
			Description: "Deploy or update the ticket prompt embed",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "channel",
					Type:        discordgo.ApplicationCommandOptionChannel,
					Description: "Channel to send the ticket prompt",
					Required:    true,
				},
				{
					Name:        "title",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Embed title",
					Required:    false,
				},
				{
					Name:        "description",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Embed description text",
					Required:    false,
				},
				{
					Name:        "footer",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Embed footer text",
					Required:    false,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "sync",
			Description: "Re-sync slash commands to Discord",
		},
	}
}

func (t *TicketConfig) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	sub := i.ApplicationCommandData().Options[0]
	guild := i.GuildID

	switch sub.Name {

	case "add-type":
		name := sub.Options[0].StringValue()
		desc := sub.Options[1].StringValue()
		_, _, err := supa.From("ticket_types").
			Insert(map[string]interface{}{
				"name":        name,
				"description": desc,
			}, false, "", "", "").
			Execute()
		if err != nil {
			respondErr(s, i, "Failed to add type: "+err.Error())
			return
		}
		respond(s, i, fmt.Sprintf("✅ Added ticket type **%s**", name))

	case "remove-type":
		name := sub.Options[0].StringValue()
		raw, _, err := supa.From("tickets").
			Select("id", "", false).
			Eq("type", name).
			Execute()
		if err != nil {
			respondErr(s, i, "Could not check tickets: "+err.Error())
			return
		}
		var tickets []map[string]interface{}
		_ = json.Unmarshal(raw, &tickets)
		if len(tickets) > 0 {
			respondErr(s, i, fmt.Sprintf(
				"Cannot remove type **%s**: %d tickets reference it. Close or reassign them first.",
				name, len(tickets),
			))
			return
		}
		_, _, err = supa.From("ticket_types").
			Delete("", "").
			Eq("name", name).
			Execute()
		if err != nil {
			respondErr(s, i, "Failed to remove type: "+err.Error())
			return
		}
		respond(s, i, fmt.Sprintf("🗑️ Removed ticket type **%s**", name))

	case "list-types":
		raw, _, err := supa.From("ticket_types").
			Select("name,description", "", false).
			Execute()
		if err != nil {
			respondErr(s, i, "Failed to list types: "+err.Error())
			return
		}
		var rows []map[string]interface{}
		_ = json.Unmarshal(raw, &rows)
		if len(rows) == 0 {
			respond(s, i, "No ticket types defined.")
			return
		}
		var lines []string
		for _, r := range rows {
			lines = append(lines, fmt.Sprintf("• **%s**: %s", r["name"], r["description"]))
		}
		respond(s, i, "📋 Ticket types:\n"+strings.Join(lines, "\n"))

	case "configure":
		openCat := sub.Options[0].ChannelValue(s).ID
		supportRole := sub.Options[1].RoleValue(s, guild).ID
		closedCat := sub.Options[2].ChannelValue(s).ID

		_, _, err := supa.From("guild_settings").
			Upsert(map[string]interface{}{
				"guild_id":           guild,
				"ticket_category_id": openCat,
				"support_role_id":    supportRole,
				"closed_category_id": closedCat,
			}, "guild_id", "", "").
			Execute()
		if err != nil {
			respondErr(s, i, "Failed to configure: "+err.Error())
			return
		}
		respond(s, i, fmt.Sprintf(
			"✅ Configured open-category <#%s>, support-role <@&%s>, closed-category <#%s>",
			openCat, supportRole, closedCat,
		))

	case "prompt":
		rawCfg, _, err := supa.From("guild_settings").
			Select("ticket_category_id,support_role_id,prompt_channel_id,prompt_message_id,prompt_title,prompt_description,prompt_footer", "", false).
			Eq("guild_id", guild).
			Single().
			Execute()
		if err != nil {
			respondErr(s, i, "Please run `/ticketconfig configure` first.")
			return
		}
		var cfg map[string]interface{}
		_ = json.Unmarshal(rawCfg, &cfg)

		openCat := cfg["ticket_category_id"].(string)
		roleID := cfg["support_role_id"].(string)
		prevCh, _ := cfg["prompt_channel_id"].(string)
		prevMsg, _ := cfg["prompt_message_id"].(string)
		storedTitle, _ := cfg["prompt_title"].(string)
		storedDesc, _ := cfg["prompt_description"].(string)
		storedFooter, _ := cfg["prompt_footer"].(string)

		rawTypes, _, err := supa.From("ticket_types").
			Select("name,description", "", false).
			Execute()
		if err != nil {
			respondErr(s, i, "Failed to load ticket types: "+err.Error())
			return
		}
		var types []map[string]interface{}
		_ = json.Unmarshal(rawTypes, &types)

		var menuOptions []discordgo.SelectMenuOption
		for _, tType := range types {
			menuOptions = append(menuOptions, discordgo.SelectMenuOption{
				Label:       tType["name"].(string),
				Value:       tType["name"].(string),
				Description: tType["description"].(string),
			})
		}

		channel := sub.Options[0].ChannelValue(s)
		title := storedTitle
		desc := storedDesc
		footer := storedFooter
		if title == "" {
			title = "Support Tickets"
		}
		if desc == "" {
			desc = "Select one of the options below to create a support ticket."
		}
		if footer == "" {
			footer = "TicketTool.xyz – Ticketing without clutter"
		}
		for _, opt := range sub.Options {
			switch opt.Name {
			case "title":
				if v := opt.StringValue(); v != "" {
					title = v
				}
			case "description":
				if v := opt.StringValue(); v != "" {
					desc = v
				}
			case "footer":
				if v := opt.StringValue(); v != "" {
					footer = v
				}
			}
		}

		embed := &discordgo.MessageEmbed{
			Title:       title,
			Description: desc,
			Color:       0x2F3136,
			Footer:      &discordgo.MessageEmbedFooter{Text: footer},
		}
		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "ticket_select",
						Placeholder: "How can we help you today?",
						Options:     menuOptions,
					},
				},
			},
		}

		shouldSend := true
		if prevCh != "" && prevMsg != "" {
			_, editErr := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel:    prevCh,
				ID:         prevMsg,
				Embeds:     &[]*discordgo.MessageEmbed{embed},
				Components: &components,
			})
			if editErr == nil {
				shouldSend = false
			} else if restErr, ok := editErr.(*discordgo.RESTError); ok && restErr.Message != nil && restErr.Message.Code == 10008 {
				prevCh, prevMsg = "", ""
			} else {
				respondErr(s, i, "Failed to update prompt: "+editErr.Error())
				return
			}
		}
		if shouldSend {
			msg, sendErr := s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
				Embed:      embed,
				Components: components,
			})
			if sendErr != nil {
				respondErr(s, i, "Failed to send prompt: "+sendErr.Error())
				return
			}
			prevCh = channel.ID
			prevMsg = msg.ID
		}

		_, _, err = supa.From("guild_settings").
			Upsert(map[string]interface{}{
				"guild_id":           guild,
				"ticket_category_id": openCat,
				"support_role_id":    roleID,
				"prompt_channel_id":  prevCh,
				"prompt_message_id":  prevMsg,
				"prompt_title":       title,
				"prompt_description": desc,
				"prompt_footer":      footer,
			}, "guild_id", "", "").
			Execute()
		if err != nil {
			respondErr(s, i, "Failed to save prompt record: "+err.Error())
			return
		}
		respond(s, i, "✅ Ticket prompt deployed/updated.")

	case "sync":
		_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, guild, All())
		if err != nil {
			respondErr(s, i, "Sync failed: "+err.Error())
			return
		}
		respond(s, i, "🔄 Commands re-synced")
	}
}
