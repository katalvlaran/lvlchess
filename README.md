# lvlChess – Telegram-based Chess Bot & Web Frontend

This repository contains a hybrid project combining:
- A **Telegram bot** (written in Go) that allows playing turn-based chess between two real users.
- A **React** frontend (HTML5) that can be deployed as a separate web client or integrated with Telegram Games.
- A **PostgreSQL** database for persisting rooms, users, and tournament data.
- Optional **NATS/JetStream** for event streaming or microservices expansion.

---

## Table of Contents

1. [Overview](#overview)
2. [Project Structure](#project-structure)
3. [Key Features](#key-features)
4. [Installation](#installation)
5. [Configuration](#configuration)
6. [Usage](#usage)
7. [Docker & Deployment](#docker--deployment)
8. [How to Make Moves](#how-to-make-moves)
9. [Further Plans & TODO](#further-plans--todo)

---

## Overview

**lvlChess** aims to provide a convenient way to play classical chess through Telegram with minimal friction:
- Players either remain in a private chat with the bot or a group chat.
- Moves are selected via inline buttons (choose a piece, then choose a valid move).
- The board is rendered in ASCII style, with orientation logic (e.g., black side is flipped).
- PostgreSQL holds persistent data: user profiles, rooms, tournaments.
- React can serve as a web interface if you want to integrate a more graphical drag & drop board.

In the future, the project can be expanded to other messengers (e.g., Discord, WhatsApp) and offer a more advanced Web3 or NFT-based economy.

---

## Project Structure

The repository is roughly divided as follows:

```bash
lvlchess/
├── cmd/
│   └── bot.go                # Main entrypoint for the Go Telegram bot
├── config/
│   └── config.go             # Environment loading/validation
├── internal/
│   ├── db/
│   │   ├── models/           # Database models for rooms, users, tournaments
│   │   ├── repositories/     # CRUD logic for those models
│   │   └── pg.go             # pgxpool initialization + basic schema creation
│   ├── game/                 # Chess logic (ASCII rendering, utility)
│   └── telegram/             # Bot handlers (commands, callbacks, notifications)
│       ├── basic_handlers.go
│       ├── main_handlers.go
│       ├── move_handlers.go
│       ├── notification.go
│       ├── room_handlers.go
│       └── ...
├── frontend/
│   ├── public/               # Basic index.html
│   ├── src/                  # React source files (App.js, index.js)
│   └── Dockerfile            # Docker build for the React app
├── .env                      # Example environment variables
├── docker-compose.yml        # Docker services: Go bot, PostgreSQL, Redis, NATS, React
├── Dockerfile                # Dockerfile for building the Go bot
├── go.mod                    # Go modules
└── README.md                 # This file

```

---

## Key Features


1. **Telegram-based gameplay**:
    - `/start` command triggers inline menu: Create Room, List Rooms, etc.
    - Two-step move selection (pick the piece → pick the target).
    - ASCII board rendering (white perspective, black perspective, or horizontal).
2. **React Web Client**:
    - Minimal example included (Hello from lvlChess React).
    - Potential expansion into a fully interactive board (drag & drop).
3. **PostgreSQL**:
    - Storing user data, rooms, tournaments, etc.
    - On conflict user merges (CreateOrUpdateUser).
    - Basic migrations included (initSchema).
4. **NATS**:
    - Potential for microservices or event streaming (not mandatory in MVP).
5. **Tournament placeholders**:
    - Create/Join/Start a tournament logic (still a work in progress).
6. **Multi-architecture**:
    - Docker-based images for both Go bot and React.
    - Allows easy deployment to AWS EC2 (docker-compose) or Kubernetes (with some adjustments).

---

## Installation

1. **Clone the repo**:
   ```
   git clone https://github.com/your-username/lvlchess.git
   cd lvlchess
   ```
2. **Install Go and Node.js**
    - Optionally, ensure Docker is installed if you plan to containerize.
3. **Setup a Telegram Bot**via BotFather, get the BOT_TOKEN
4. **Configure** environment:
    - Duplicate `.example.env` to `.env`.
    - Fill in BOT_TOKEN, PG_HOST, PG_USER, etc.
5. **Initialize**:
    ```
   # Option A: Local run
    go mod tidy
    go run ./cmd/bot.go
    
    # Option B: Docker
    docker-compose build
    docker-compose up -d
    ```
---

## Configuration

The project reads .env using github.com/joho/godotenv and caarlos0/env. \
You can customize:
- **Команды** (в личке):
    - `BOT_TOKEN`: Telegram token,
    - `OWNER_ID`: (optional) your personal ID if you want to handle admin stuff,
    - `PG_USER`, `PG_PASS`, `PG_HOST`, `PG_DB_NAME`: PostgreSQL connection
    - `NATS`: If you integrate it, or skip if not needed. 
  
  For production, you can pass real environment variables or orchestrate them in your CI/CD pipeline.
---

## Usage

After the bot is running:
1. Open Telegram → your bot link → `/start`.
2. The bot greets you with an inline menu:
    - **Create Room**: sets up a new room in DB, sends an invite link (t.me/YourBot?start=room_<id>)
    - **Join**: if a user clicks that link, the bot merges them as the second player.
    - Then the game starts: White's turn or random assignment of colors.
3. **ASCII Board**: The bot sends a textual board. White sees the normal orientation, black sees reversed, or (in group chat) a horizontal layout.
---

## Docker & Deployment

1. **docker-compose.yml** includes:
    - **lvlchess_go** (the Go bot)
    - **lvlchess_front** (the React app)
    - **lvlchess_db** (Postgres)
    - **lvlchess_redis** (optional if you want caching)
    - **lvlchess_nats** (optional if you want microservices)
2. Run:
    ```
     docker-compose build
     docker-compose up -d
    ```
3. Access:
   - The Go bot doesn’t have an HTTP UI, but it listens on `:8080` for future expansions.
   - The React app runs on `:3000`.

   For production, set environment variables in `.env` or via your AWS EC2, then run `docker-compose up -d.`
---

## How to Make Moves
When two players are in the same room:
- If it’s your turn, you see an inline button to **Choose a figure** (like `choose_figure:b8`).
- The bot then lists possible moves (e.g., `move:b8-c6`, `move:b8-a6`, etc.).
- Click the move → the bot verifies with `notnil/chess`.
- If valid, the board is updated and it becomes the other player’s turn.

**ASCII** example:
```
□ | a  b  c  d  e  f  g  h | ■
--+------------------------+--
8 | ♜  ♞  ♝  ♛  ♚  ♝  ♞  ♜ | 8
7 | ♟  ♟  ♟  ♟  ♟  ♟  ♟  ♟ | 7
...
```

---

## Further Plans & TODO
 1. **Enhanced React board** with drag & drop.
 2. **Tournament** bracket UI and scheduling logic.
 3. **Optional AI** (integrate a chess engine so a user can play vs Bot).
 4. **Additional messaging platforms** (Discord, Slack) via microservices or NATS.
 5. **NFT/DAO** integration for advanced scenarios (still conceptual).
 6. **Monitoring** with Prometheus/Grafana (the docker-compose includes them optionally).
 7. **More sophisticated PG migrations** (Flyway, Goose, or manual SQL). 

Feel free to open PRs or issues with improvements, feedback, or new feature requests!
**Enjoy lvlChess and let’s bring more chess fans!**
---