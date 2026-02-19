# Southwest Check-in Bot

> [!WARNING]
> **Deprecated** — Southwest Airlines switched to assigned seating in 2026, making this tool obsolete. The open boarding system this bot was designed for no longer exists.


A command-line tool written in Go that automatically checks in to Southwest Airlines flights exactly 24 hours before departure.

## Why?

Southwest Airlines doesn't assign seats until check-in, which opens exactly 24 hours before your flight. Checking in early gets you a better boarding position (A1-A60 board first). This bot automates the check-in process with precision timing.

## Installation

### From Source

```bash
git clone https://github.com/nealwashere/Southwest-Check-in-Bot.git
cd Southwest-Check-in-Bot
go build -o southwest-bot .
```

### Requirements

- Go 1.21 or later

## Usage

```bash
./southwest-bot -c <confirmation> -f <first-name> -l <last-name> -d <departure-time>
```

### Options

| Flag | Long Flag | Description |
|------|-----------|-------------|
| `-c` | `--confirmation` | Southwest confirmation number (required) |
| `-f` | `--first` | Passenger first name (required) |
| `-l` | `--last` | Passenger last name (required) |
| `-d` | `--departure` | Flight departure time (required) |
| `-v` | `--verbose` | Enable verbose logging |
| `-h` | `--help` | Show help message |

### Departure Time Formats

The bot accepts several time formats:

- `2024-01-15 14:30` (recommended)
- `2024-01-15 14:30:05`
- `2024-01-15T14:30:00-06:00` (RFC3339 with timezone)
- `01/02/2024 14:30`
- `01/02/2024 3:04 PM`

### Examples

```bash
# Basic usage - schedule check-in for a flight
./southwest-bot -c ABC123 -f John -l Doe -d "2024-06-15 14:30"

# With verbose logging
./southwest-bot -c ABC123 -f John -l Doe -d "2024-06-15T14:30:00-06:00" -v

# If check-in window is already open, it checks in immediately
./southwest-bot -c ABC123 -f John -l Doe -d "2024-06-14 10:00"
```

## How It Works

1. **Schedule**: The bot calculates when check-in opens (24 hours before departure) and displays a countdown
2. **Wait**: It sleeps until the check-in window, then busy-waits in the final 100ms for precision
3. **Check-in**: Makes API requests to Southwest to complete the check-in
4. **Results**: Displays your boarding position (e.g., A24, B15, C40)

## Output Example

```
Southwest Check-in Bot
======================
Confirmation: ABC123
Passenger: John Doe
Departure: 2024-06-15 14:30:00 CST

Check-in opens: 2024-06-14 14:30:00 CST
Time until check-in: 23h 45m 12s

Waiting for check-in window...
[=============================>] 99% 5s remaining

Check-in window is NOW OPEN!

Attempting check-in...

Check-in successful!
====================

Flight 1234: SFO -> LAX
Departure: 2024-06-15 14:30
Passenger: John Doe
Boarding Position: A24
```

## Running in the Background

To keep the bot running even after closing your terminal:

```bash
# Using nohup
nohup ./southwest-bot -c ABC123 -f John -l Doe -d "2024-06-15 14:30" > checkin.log 2>&1 &

# Using screen
screen -S checkin
./southwest-bot -c ABC123 -f John -l Doe -d "2024-06-15 14:30"
# Press Ctrl+A, then D to detach

# Using tmux
tmux new -s checkin
./southwest-bot -c ABC123 -f John -l Doe -d "2024-06-15 14:30"
# Press Ctrl+B, then D to detach
```

## Troubleshooting

### 403 Error
Southwest may block requests if they detect automation. The bot uses mobile API headers to appear as a normal mobile browser request.

### Check-in Failed
- Verify your confirmation number is correct
- Ensure the passenger name matches exactly as it appears on the reservation
- Check that the departure time is in the future

## Disclaimer

This tool is for personal use only. Use at your own risk. Southwest Airlines may change their API at any time, which could break this tool.

## License

MIT
