# DS-Ticket Bot

A modular Discord ticketing bot built with Go + DisGo and Supabase.

---

## 🚀 Features

- Slash commands:
  - `/ticket ‹type› [reason]` – create a new ticket.
  - `/ticketconfig add-type ‹name› ‹description›` – add a ticket type.
  - `/ticketconfig remove-type ‹name›` – remove a ticket type.
  - `/ticketconfig list-types` – list all ticket types.
  - `/ticketconfig configure ‹category› ‹support_role› ‹closed_category›` – set up your categories & support role.
  - `/ticketconfig prompt [channel] [title] [description] [footer]` – deploy/update the ticket prompt embed.
  - `/ticketconfig sync` – re-sync slash commands to Discord.

- Dropdown menu ticket selection  
- “Claim” & “Close” buttons on new ticket channels  
- Prevents duplicate open tickets per user  
- Sanitized, unique channel names  
- Only staff can “Claim” tickets  
- Caches ticket-types in memory (5 min TTL)  
- Archives closed tickets to your configured category  

---

## 🛠️ Prerequisites

1. Go 1.20+  
2. A Supabase project  
3. A Discord bot token (with **applications.commands** & **bot** scopes)

---

## 📦 Setup

### 1. Clone & build  
```bash
git clone https://github.com/ds-dotsoft/ds-ticket.git
cd ds-ticket
go build -o ds-ticket
```

### 2. Environment  
Create a `.env` file in the repo root:
```env
TOKEN=your_discord_bot_token
GUILD_ID=your_dev_guild_id      # for command registration
SUPABASE_URL=https://<project>.supabase.co
SUPABASE_KEY=your_anon_or_service_role_key
```

### 3. Database schema  
In Supabase → SQL editor, run:
```sql
-- store guild settings
create table if not exists guild_settings (
  guild_id            text primary key,
  ticket_category_id  text,
  support_role_id     text,
  closed_category_id  text,
  prompt_channel_id   text,
  prompt_message_id   text,
  prompt_title        text,
  prompt_description  text,
  prompt_footer       text
);

-- store types & tickets
create table if not exists ticket_types (
  name        text primary key,
  description text
);

create table if not exists tickets (
  channel_id text primary key,
  user_id    text not null,
  type       text not null references ticket_types(name),
  reason     text,
  status     text not null default 'open'
);
```

### 4. Run  
```bash
./ds-ticket
```

---

## 📝 Usage

1. **Configure**  
   ```
   /ticketconfig configure
     category: #open-tickets
     support_role: @Staff
     closed_category: #closed-tickets
   ```

2. **Add Ticket Types**  
   ```
   /ticketconfig add-type
     name: billing
     description: Billing & payments
   ```

3. **Deploy Prompt**  
   ```
   /ticketconfig prompt
     channel: #support
     title: Support Center
     description: Pick your issue category
     footer: We’re here to help!
   ```

4. **Open a Ticket**  
   In `#support`, select a type from the menu (or use `/ticket …`) and a new private ticket channel is created.

5. **Claim & Close**  
   - Click **Claim** (staff only) to lock in responsibility.  
   - Click **Close** to archive the channel to your “closed” category.
