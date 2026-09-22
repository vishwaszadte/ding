# 🔔 ding

> **Modern, cross-platform scheduled desktop notifications from your terminal.**

`ding` is a fast, zero-dependency CLI written in Go that delivers rich desktop notifications across **Windows, macOS, and Linux**. It features natural language scheduling, an auto-starting background daemon, pure-Go SQLite persistence, and modern Fluent/WinRT desktop alerts.

---

## ⚡ Quick Install

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/vishwaszadte/ding/main/install.ps1 | iex
```

### macOS & Linux (curl)
```bash
curl -fsSL https://raw.githubusercontent.com/vishwaszadte/ding/main/install.sh | bash
```

### Go Developers (`go install`)
```bash
go install github.com/vishwaszadte/ding@latest
```

---

## 🚀 Usage

### Scheduling Notifications
```bash
# Relative time
ding in 15m "Take a break from the screen"
ding in 2h "Review PR comments"
ding in 30s "Boiling water ready"

# Natural calendar & clock times
ding today 4pm --message "Team sync meeting"
ding tomorrow 6am --message "Morning run"
ding today 17:30 "Submit daily progress"

# Recurring reminders
ding every hour --message "Get up and stretch"
ding every 30m "Hydration check" --tag health
ding every weekday 9am -m "Team daily standup" -t "Work Sync"
ding every Monday at 10am -m "Sprint Planning"

# Flags & Customizations
ding in 20m "Deploy to staging" --urgency critical --sound
ding in 45m "Check issue #42" --open "https://github.com/issues" --tag dev
```

---

## 📋 Managing Reminders

```bash
# List all scheduled and recurring reminders
ding list
ding ls

# Filter reminders by status
ding list --status active
ding list --status paused

# Cancel reminders
ding cancel <id>
ding cancel --all

# Pause and resume recurring reminders
ding pause <id>
ding resume <id>

# Notification history & audit logs
ding history
ding history -n 50

# Diagnostics & verification
ding test       # Triggers an immediate toast to check notification display & audio
ding doctor     # Health check (database, background daemon, notification subsystem)
```

---

## ⚙️ Daemon Management

You never need to start the daemon manually: **`ding` auto-spawns it in the background the first time you schedule a reminder**.

However, you can also control it directly:
```bash
ding daemon status   # Check if background daemon is active and show PID
ding daemon stop     # Stop the background process
ding daemon start    # Start the daemon in the background
ding daemon logs     # Tail background scheduler logs
```

---

## 🧠 How It Works Under the Hood

---

## 🗑️ Uninstallation

If you ever wish to completely remove `ding` and its data from your system:

### Windows (PowerShell)
```powershell
# 1. Stop background daemon
ding daemon stop

# 2. Remove binary executable
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\Programs\ding"

# 3. (Optional) Remove data & database
Remove-Item -Recurse -Force "$HOME\.ding"
```

### macOS & Linux
```bash
# 1. Stop background daemon
ding daemon stop

# 2. Remove binary executable
rm -f "$HOME/.local/bin/ding"

# 3. (Optional) Remove data & database
rm -rf "$HOME/.ding"
```

---

## 🧠 How It Works Under the Hood

```
   ┌────────────────────────────────────────────────────────┐
   │                       ding CLI                         │
   │  (add / list / cancel / pause / history / doctor)      │
   └──────────────────────────┬─────────────────────────────┘
                              │
              Writes & reads  │  Spawns if not active
                              ▼
   ┌───────────────────────────────────┐    ┌─────────────────────────┐
   │ SQLite DB (~/.ding/ding.db)       │    │ Background Daemon       │
   │ • Pure Go (modernc.org/sqlite)    │◄───┤ • Auto-starts detached  │
   │ • WAL mode for fast concurrency   │    │ • Ticker loop (1s)      │
   │ • No CGO / No C compiler needed   │    │ • Sleep/wake catch-up   │
   └───────────────────────────────────┘    └────────────┬────────────┘
                                                         │
                                   Dispatches via OS API │
                                                         ▼
                              ┌────────────────────────────────────────┐
                              │ Native Modern Desktop Notifications    │
                              │ • Windows: WinRT Fluent Toast XML      │
                              │ • macOS: Notification Center           │
                              │ • Linux: D-Bus / libnotify             │
                              │ • Native OS chime & 🔔 bell emoji      │
                              └────────────────────────────────────────┘
```

- **Catch-up on Wake**: If your laptop goes to sleep when a reminder was scheduled to trigger, `ding` catches up immediately when you open your laptop, marking it as `missed at <time>`.
- **Zero CGO**: Pure Go from top to bottom. It compiles anywhere without GCC or MinGW.

---

## 🛠️ Contributing & Local Development

Prerequisites: Go 1.23+

```bash
# Clone the repository
git clone https://github.com/vishwaszadte/ding.git
cd ding

# Run all unit tests
go test ./... -v

# Build the binary locally
go build -o ding.exe .     # on Windows
go build -o ding .         # on macOS / Linux

# Run a test notification
./ding.exe test
```

---

## 📄 License

MIT © [vishwaszadte](https://github.com/vishwaszadte)
