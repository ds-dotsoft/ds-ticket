# go-bot-template

Discord bot template in Go.  
Modular slash-command handler, drop-in commands under `commands/`, sample `/ping` ready to go.

---

## 🚀 Project Layout

```
.
├── go.mod
├── main.go
└── commands
    ├── command.go    # registry & helper
    └── ping.go       # example /ping command
```

---

## ⚙️ Setup

1. **Clone & enter**  
   ```bash
   git clone https://your.repo/url.git
   cd url
   ```

2. **Configure**  
   Create a `.env` (or export) with:
   ```env
   DISCORD_BOT_TOKEN=your_bot_token_here
   GUILD_ID=your_dev_guild_id_here   # for instant slash registration
   ```

3. **Install dependencies**  
   ```bash
   go mod tidy
   ```

---

## ▶️ Running

```bash
go run main.go
```

## 📋 Commands

### `/ping`
Replies with “🏓 Pong!”  

---

## 📢 That’s It

- **Modular**: No central switch statements.  
- **Compile-safe**: Interface catches mistakes at build time.  
- **Instant feedback**: Guild commands overwrite immediately—no waiting.  

Get in, drop your commands into `commands/`, and build. No excuses.