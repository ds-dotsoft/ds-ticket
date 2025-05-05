// commands/ticket.go
package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	typeCache     []TicketType
	typeCacheTime time.Time
)

type TicketType struct {
	Name        string
	Description string
}

func init() {
	Register(&Ticket{})
}

type Ticket struct{}

func (t *Ticket) Name() string        { return "ticket" }
func (t *Ticket) Description() string { return "Open a support ticket" }

func (t *Ticket) Options() []*discordgo.ApplicationCommandOption {
	types, err := getCachedTypes()
	if err != nil || len(types) == 0 {
		return []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Describe your issue",
				Required:    false,
			},
		}
	}

	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, tp := range types {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  tp.Name,
			Value: tp.Name,
		})
	}

	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "type",
			Description: "Which ticket type to open",
			Required:    true,
			Choices:     choices,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "Describe your issue",
			Required:    false,
		},
	}
}

func (t *Ticket) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.Member.User

	raw, _, err := supa.From("guild_settings").
		Select("ticket_category_id,support_role_id", "", false).
		Eq("guild_id", i.GuildID).
		Single().
		Execute()
	if err != nil {
		respondErr(s, i, "Please run `/ticketconfig configure` first.")
		return
	}
	var cfg map[string]interface{}
	_ = json.Unmarshal(raw, &cfg)
	openCat := cfg["ticket_category_id"].(string)
	roleID := cfg["support_role_id"].(string)

	rawOpen, _, _ := supa.From("tickets").
		Select("channel_id", "", false).
		Eq("user_id", user.ID).
		Eq("status", "open").
		Execute()
	var opens []map[string]interface{}
	_ = json.Unmarshal(rawOpen, &opens)
	if len(opens) > 0 {
		respond(s, i, fmt.Sprintf("⚠️ You already have an open ticket: <#%s>", opens[0]["channel_id"]))
		return
	}

	data := i.ApplicationCommandData().Options
	var choice, reason string
	if len(data) > 0 {
		choice = data[0].StringValue()
	}
	if len(data) > 1 && data[1].StringValue() != "" {
		reason = data[1].StringValue()
	} else {
		reason = "No reason provided"
	}

	slug := strings.ToLower(regexp.MustCompile(`[^a-z0-9_-]`).ReplaceAllString(user.Username, "-"))
	if len(slug) > 16 {
		slug = slug[:16]
	}
	name := "ticket-" + slug
	channels, _ := s.GuildChannels(i.GuildID)
	for idx := 1; ; idx++ {
		coll := false
		for _, c := range channels {
			if c.Name == name {
				coll = true
				break
			}
		}
		if !coll {
			break
		}
		name = fmt.Sprintf("ticket-%s-%d", slug, idx)
	}

	ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:     name,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: openCat,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{ID: i.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: discordgo.PermissionViewChannel},
			{ID: user.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages},
			{ID: roleID, Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages},
		},
	})
	if err != nil {
		respondErr(s, i, "Could not create ticket channel.")
		return
	}

	// log to database
	_, _, _ = supa.From("tickets").
		Insert(map[string]interface{}{
			"channel_id": ch.ID,
			"user_id":    user.ID,
			"type":       choice,
			"reason":     reason,
			"status":     "open",
		}, false, "", "", "").
		Execute()

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎫 %s — %s", choice, user.Username),
		Description: reason,
		Color:       0x7289DA,
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "Claim", Style: discordgo.PrimaryButton, CustomID: "ticket_claim"},
				discordgo.Button{Label: "Close", Style: discordgo.DangerButton, CustomID: "ticket_close"},
			},
		},
	}
	_, _ = s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
		Embed:      embed,
		Components: components,
	})

	respond(s, i, fmt.Sprintf("✅ Ticket created: <#%s>", ch.ID))
}

// HandleTicketSelect handles dropdown interactions rather than slash commands.
func HandleTicketSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.Member.User

	raw, _, err := supa.From("guild_settings").
		Select("ticket_category_id,support_role_id", "", false).
		Eq("guild_id", i.GuildID).
		Single().
		Execute()
	if err != nil {
		respondErr(s, i, "Bot not configured—run `/ticketconfig configure` first.")
		return
	}
	var cfg map[string]interface{}
	_ = json.Unmarshal(raw, &cfg)
	openCat := cfg["ticket_category_id"].(string)
	roleID := cfg["support_role_id"].(string)

	rawOpen, _, _ := supa.From("tickets").
		Select("channel_id", "", false).
		Eq("user_id", user.ID).
		Eq("status", "open").
		Execute()
	var opens []map[string]interface{}
	_ = json.Unmarshal(rawOpen, &opens)
	if len(opens) > 0 {
		respond(s, i, fmt.Sprintf("⚠️ You already have an open ticket: <#%s>", opens[0]["channel_id"]))
		return
	}

	choice := i.MessageComponentData().Values[0]
	reason := "Please describe your issue here."

	slug := strings.ToLower(regexp.MustCompile(`[^a-z0-9_-]`).ReplaceAllString(user.Username, "-"))
	if len(slug) > 16 {
		slug = slug[:16]
	}
	name := "ticket-" + slug
	channels, _ := s.GuildChannels(i.GuildID)
	for idx := 1; ; idx++ {
		coll := false
		for _, c := range channels {
			if c.Name == name {
				coll = true
				break
			}
		}
		if !coll {
			break
		}
		name = fmt.Sprintf("ticket-%s-%d", slug, idx)
	}

	ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:     name,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: openCat,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{ID: i.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: discordgo.PermissionViewChannel},
			{ID: user.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages},
			{ID: roleID, Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages},
		},
	})
	if err != nil {
		respondErr(s, i, "Could not create ticket channel.")
		return
	}

	_, _, _ = supa.From("tickets").
		Insert(map[string]interface{}{
			"channel_id": ch.ID,
			"user_id":    user.ID,
			"type":       choice,
			"status":     "open",
		}, false, "", "", "").
		Execute()

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎫 %s — %s", choice, user.Username),
		Description: reason,
		Color:       0x7289DA,
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "Claim", Style: discordgo.PrimaryButton, CustomID: "ticket_claim"},
				discordgo.Button{Label: "Close", Style: discordgo.DangerButton, CustomID: "ticket_close"},
			},
		},
	}
	_, _ = s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{Embed: embed, Components: components})

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Ticket created: <#%s>", ch.ID),
		},
	})
}

func HandleButtonClaim(s *discordgo.Session, i *discordgo.InteractionCreate) {
	raw, _, _ := supa.From("guild_settings").
		Select("support_role_id", "", false).
		Eq("guild_id", i.GuildID).
		Single().
		Execute()
	var cfg map[string]interface{}
	_ = json.Unmarshal(raw, &cfg)
	staffID := cfg["support_role_id"].(string)

	if !memberHasRole(i.Member, staffID) {
		respondErr(s, i, "🚫 Only staff can claim tickets.")
		return
	}

	user := i.Member.User
	_ = s.ChannelPermissionSet(
		i.ChannelID, user.ID, discordgo.PermissionOverwriteTypeMember,
		discordgo.PermissionViewChannel|discordgo.PermissionSendMessages, 0,
	)

	components := i.Message.Components
	for ri, row := range components {
		if ar, ok := row.(discordgo.ActionsRow); ok {
			for bi, comp := range ar.Components {
				if btn, ok := comp.(discordgo.Button); ok && btn.CustomID == "ticket_claim" {
					btn.Disabled = true
					ar.Components[bi] = btn
				}
			}
			components[ri] = ar
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf("🔖 Ticket claimed by %s", user.Username),
			Components: components,
		},
	})
}

func HandleButtonClose(s *discordgo.Session, i *discordgo.InteractionCreate) {
	chID := i.ChannelID

	_, _, _ = supa.From("tickets").
		Update(map[string]interface{}{"status": "closed"}, "", "").
		Eq("channel_id", chID).
		Execute()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "🔒 Closing ticket…"},
	})

	go func() {
		time.Sleep(2 * time.Second)
		raw, _, _ := supa.From("guild_settings").
			Select("closed_category_id", "", false).
			Eq("guild_id", i.GuildID).
			Single().
			Execute()
		var cfg map[string]interface{}
		_ = json.Unmarshal(raw, &cfg)
		closedCat := cfg["closed_category_id"].(string)

		_, err := s.ChannelEditComplex(chID, &discordgo.ChannelEdit{
			ParentID: closedCat,
		})
		if err != nil {
			s.ChannelDelete(chID)
		}
	}()
}

func getCachedTypes() ([]TicketType, error) {
	if time.Since(typeCacheTime) < 5*time.Minute {
		return typeCache, nil
	}
	raw, _, err := supa.From("ticket_types").
		Select("name,description", "", false).
		Execute()
	if err != nil {
		return nil, err
	}
	var rows []map[string]interface{}
	_ = json.Unmarshal(raw, &rows)

	out := make([]TicketType, 0, len(rows))
	for _, r := range rows {
		out = append(out, TicketType{
			Name:        r["name"].(string),
			Description: r["description"].(string),
		})
	}
	typeCache = out
	typeCacheTime = time.Now()
	return out, nil
}
