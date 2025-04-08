# Telega-Chess

**Telega-Chess** — это Telegram-бот для пошаговой игры в шахматы между двумя (реальными) игроками. Поддерживается:

- **Создание комнат** (с опциональной настройкой, кто играет за белых),
- **Выбор фигур и ходов** через inline-кнопки,
- **Альтернативные способы отображения доски** (ASCII в личках, «горизонтальный» ASCII/или SVG в группах),
- **Синхронизация** с библиотекой [notnil/chess](https://github.com/notnil/chess) для валидации ходов (включая рокировку, превращение пешки).

## Содержание

1. [Основные особенности](#основные-особенности)
2. [Установка и запуск](#установка-и-запуск)
3. [Переменные окружения](#переменные-окружения)
4. [Архитектура](#архитектура)
5. [Основные файлы и функции](#основные-файлы-и-функции)
6. [Сценарии использования](#сценарии-использования)
7. [Команды и inline-кнопки](#команды-и-inline-кнопки)
8. [Как делать ходы](#как-делать-ходы)
9. [Тестирование](#тестирование)
10. [Дальнейшие планы](#дальнейшие-планы)

---

## 1. Основные особенности

- **Чистая механика шахмат** (на базе `github.com/notnil/chess`): все правила (включая рокировку, взятие на проходе, превращение пешки).
- **Двухшаговый выбор хода** через inline-кнопки: сначала выбираем фигуру, затем доступный ход.
- **ASCII-доска** в личном чате:
    - Для белых — классический вид,
    - Для чёрных — перевёрнутый.
- **Горизонтальная** (или SVG) доска в групповом чате.
- **Сохранение** состояния в PostgreSQL (комнаты, пользователи).
- **Разделение** на файлы по логическим областям (handlers, db, game, utils).

---

## 2. Установка и запуск

1. **Склонировать** репозиторий:
   ```bash
   git clone https://github.com/your-username/telega-chess.git
   cd telega-chess
   ```
2. **Сконфигурировать** переменные окружения (см. [.env](#переменные-окружения)). Убедитесь, что указали `BOT_TOKEN`, `PG_USER`, `PG_PASS`, и т.д.

3. **Собрать и запустить**:
   ```bash
   go build -o telega-chess ./cmd/bot.go
   ./telega-chess
   ```
   Или просто `go run ./cmd/bot.go`.

4. **Проверить**, что бот отвечает: откройте `t.me/<ВашБот>` в Telegram, введите `/start`.

---

## 3. Переменные окружения

В файле **`.env`** должны находиться ключевые переменные:

- `BOT_TOKEN` — токен Telegram-бота (от BotFather).
- `OWNER_ID` — ваш ID (если нужно).
- `PG_USER`, `PG_PASS`, `PG_HOST`, `PG_PORT`, `PG_DB_NAME` — для подключения к PostgreSQL.
- Пример:
  ```env
  BOT_TOKEN="123456:ABC-..."
  OWNER_ID=467103090
  PG_USER="postgres"
  PG_PASS="secret"
  PG_HOST="localhost"
  PG_PORT="5432"
  PG_DB_NAME="telega_chess"
  ```

---

## 4. Архитектура

```
/telega-chess
├── /cmd
│   └── bot.go                              # Точка входа: запуск бота (Telegram), LoadConfig(),InitLogger(),InitDB(),NewBotAPI(botToken),NewHandler(bot),HandleUpdate(context.Background(), update) ...
├── /config
│   └── config.go                           # Чтение/валидация переменных среды (env)
├── /internal
│   ├── /db
│   │   ├── /models                         # Структуры описывающие таблицы 
│   │   │   ├── rooms.go                    # Структура по таблице rooms +валидацией
│   │   │   ├── tournaments.go              # Структуры по таблицам tournaments и tournament_settings +валидацией
│   │   │   └── users.go                    # Структура по таблице users +валидацией
│   │   ├── /repositories                   # CRUD-операции с таблицами
│   │   │   ├── rooms_repo.go               # CRUD-операции по таблице rooms
│   │   │   ├── tournament_settings_repo.go # CRUD-операции по таблице tournament_settings
│   │   │   ├── tournaments_repo.go         # CRUD-операции по таблице tournaments
│   │   │   └── users_repo.go               # CRUD-операции по таблице users
│   │   └── pg.go                           # Инициализация PostgreSQL (pgxpool)
│   └── /game
│   │   ├── render.go                       # ASCII-отрисовка шахматной доски (RenderASCIIBoardWhite(),RenderASCIIBoardHorizontal(),RenderASCIIBoardBlack())
│   │   └── utils.go                        # Логика AssignRandomColors, parseSquare и др. вспомогательные методы
│   └── /telegram
│   │   ├── basic_handlers.go               # /start, /game_list, /play_with_bot
│   │   ├── chat_handlers.go                # Настройка группы/чата (create_chat, setroom)
│   │   ├── main_handlers.go                # HandleUpdate, handleCallback, handleNewChatMembers
│   │   ├── move_handlers.go                # Ходы (prepareMoveButtons, handleMoveCallback), parseCallbackData
│   │   ├── notification.go                 # notifyGameStarted, SendBoardToRoomOrUsers, MakeFinalTitle и т.д.
│   │   ├── room_handlers.go                # handleCreateRoom, handleJoinRoom, ...
│   │   └── tournament_handlers.go          # handleCreateRoom, handleJoinRoom, ...
│   └── /utils
│       └── logger.go                       # Глобальный zap-логгер
├── .env                                    # Переменные окружения (BOT_TOKEN, PG_HOST, PG_PORT, ...)
├── docker-compose.yml
├── Dockerfile
├── go.mod                                  # Go-модуль, список зависимостей (go-telegram-bot-api, jackc/pgx, etc.)
└── README.md                               # Документация (вы здесь!)
```

---

## 5. Основные файлы и функции

1. **`cmd/bot.go`**:
    - `main()` — запускает всё: Config, DB, Bot API.
2. **`config/config.go`**:
    - `ReadConfig()`, `Validate()`.
3. **`internal/db/rooms.go`**:
    - `CreateRoom()`, `GetRoomByID()`, `UpdateRoom()`, поля `Room{
      RoomID, RoomTitle, Player1, Player2, Status, BoardState, IsWhiteTurn, WhiteID, BlackID, ChatID, CreatedAt, UpdatedAt}`.
4. **`internal/db/users.go`**:
    - `CreateOrUpdateUser()`, `GetUserByID()`, поля `User{
      ID, Username, FirstName, ChatID, CurrentRoom, Rating, Wins, TotalGames}`.
5. **`internal/game/render.go`**:
    - `RenderASCIIBoardWhite()`, `RenderASCIIBoardHorizontal()`, `RenderASCIIBoardBlack()`.
6. **`internal/telegram/move_handlers.go`**:
    - `prepareMoveButtons()`, `handleChooseFigureCallback()`, `handleMoveCallback()`.
7. **`internal/telegram/notification.go`**:
    - `NotifyGameStarted()`, `SendMessageToRoomOrUsers()`.
8. **`internal/telegram/basic_handlers.go`**:
    - `handleStartCommand()`, `handlePlayWithBotCommand()`, `handleGameListCommand()`.
9. **`internal/telegram/main_handlers.go`**:
    - `HandleUpdate()`, `handleMessage()`, `handleCallback()`.
10. **`internal/telegram/room_handlers.go`**:
    - `handleCreateRoomCommand()`, `handleJoinRoom()`.
11. **`internal/utils/logger.go`**:
    - `InitLogger()`.

---

## 6. Сценарии использования

1. **Запуск**:
    1. `go run ./cmd/bot.go`
    2. Бот читает `.env`, подключается к PostgreSQL, слушает Telegram Updates.
2. **Пользователь** вводит `/start`:
    - Бот выводит приветствие и **4 кнопки**: «Создать комнату», «Мои игры», «Играть с ботом», «Создать и настроить комнату».
3. **Создать комнату**:
    - При нажатии «🆕 Создать комнату», вызывается `handleCreateRoomCommand`, создаёт `rooms.CreateRoom`, `IsWhiteTurn=true`, возвращает «Комната создана!».
4. **Создать и настроить**:
    - Спрашивает «Кто за белых? (Я сам / Соперник)».
    - В зависимости от выбора → `WhiteID`=creator или `nil`, `BlackID`=`nil` или `creator`.
5. **Второй игрок** (пришёл по `t.me/bot?start=room_<id>`) → `handleJoinRoom`: если `WhiteID == nil`, он становится белым; если нет, он становится чёрным.
6. **Ходы**:
    - Для лички: `prepareMoveButtons(...)` → «choose_figure: e2» → «move:e2-e4».
    - Для группового чата: `RenderBoardHorizontal(...)` при каждом ходе.
    - `room.IsWhiteTurn` = !`room.IsWhiteTurn` после хода.
7. **Окончание**:
    - Если `chess.Outcome()` = WhiteWon / BlackWon / Draw → «Игра завершена!».

---

## 7. Команды и inline-кнопки

- **Команды** (в личке):
    - `/start` — приветствие, 4 кнопки,
    - `/create_room` — аналог «Создать комнату»,
    - `/play_with_bot`, `/game_list`, …
- **Inline-кнопки**:
    - `"create_room"`, `"my_games"`, `"play_bot"`, `"setup_room"`,
    - `"choose_figure:b8"`, `"move:b8-c6"`,
    - `"setup_room_white:me"`, `"setup_room_white:opponent"`.

---

## 8. Как делать ходы

1. **Фигуры**:
    - При нажатии на кнопку «choose_figure:b8», бот смотрит valid moves для `b8` (в `move_handlers.go`).
    - Генерирует кнопки «move:b8-c6», «move:b8-a6» и т.д.
2. **Ход**:
    - При «move:b8-c6», мы делаем `game.Move(...)`. Если ок, `room.IsWhiteTurn = !room.IsWhiteTurn`.
    - В личку каждому игроку уходит ASCII-доска (белому — нормальная, чёрному — перевёрнутая).
    - В групповой чат (если есть) идёт «горизонтальная» ASCII (или SVG).

---

## 9. Тестирование

- **Unit-тесты**: использовать Go-пакет `testing`.
    - Мокать BotAPI (`mockBot.Send(...)`).
    - Проверять, что при `/start` (команда) вызывается `handleStartCommand` и создаёт 4 кнопки.
    - Тестить «/start room_{id}» → `handleJoinRoom`.
    - Тестить «move:b2-b4» → `room.IsWhiteTurn` переключается.
- **Интеграционные тесты**: поднять бота локально, реально походить двумя аккаунтами.

---

## 10. Дальнейшие планы

- **Полноценная поддержка SVG** в группах (вместо ASCII).
- **Логирование** ходов в PGN, /moves_history.
- **Система рейтинга** (wins, totalGames) в таблице `users`.
- **Турнирная система** (уже есть заготовки tournaments_repo.go).
- **HTML5 Web-интерфейс** (React) интегрировать с Telegram Games API.
- Расширение на **WhatsApp/Discord** (через микросервисы, NATS).
- Дополнить **AI/ChessEngine**-бот (против компьютера) через notnil/chess.Engine API или сторонний движок.
---