package telegram

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"telega_chess/internal/db"
	"telega_chess/internal/game"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Rooms = make(map[string]int64) // roomID -> player1ID

func HandleCommands(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		handleStartCommand(bot, update)
	case "create_room":
		handleCreateRoomCommand(bot, update)
	case "game_list":
		handleGameListCommand(bot, update)
	case "play_with_bot":
		handlePlayWithBotCommand(bot, update)
	case "setroom":
		handleSetRoomCommand(bot, update)
	default:
		handleUnknownCommand(bot, update)
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	args := update.Message.CommandArguments() // то, что идёт после /start
	if len(args) > 5 && args[:5] == "room_" {
		roomID := args[5:]
		handleJoinRoom(bot, update, roomID)
		return
	}

	// Стандартное приветствие, если нет room_
	messageText := "Добро пожаловать в Telega-Chess!\n" +
		"Команды:\n" +
		"- /create_room — создать новую игровую комнату.\n" +
		"- /game_list — вернуться к текущей игре.\n" +
		"- /play_with_bot — играть против AI (заглушка)."
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
	bot.Send(msg)
}

func handleJoinRoom(bot *tgbotapi.BotAPI, update tgbotapi.Update, roomID string) {
	player2ID := update.Message.From.ID

	// username or fallback
	secondUsername := update.Message.From.UserName
	if secondUsername == "" {
		secondUsername = update.Message.From.FirstName
	}

	room, err := db.GetRoomByID(roomID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Комната не найдена: "+err.Error())
		bot.Send(msg)
		return
	}

	if room.Player1ID == int64(player2ID) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не можете присоединиться к собственной комнате :)")
		bot.Send(msg)
		return
	}
	if room.Player2ID != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "В этой комнате уже есть второй игрок.")
		bot.Send(msg)
		return
	}

	// Присвоим второго игрока
	p2 := int64(player2ID)
	room.Player2ID = &p2

	// Готовимся к началу игры:
	// 1. Случайное назначение цветов (player1 белыми или player2 белыми).
	rand.Seed(time.Now().UnixNano()) // или инициализировать в main()
	if rand.Intn(2) == 0 {
		// player1 - белые
		room.WhiteID = &room.Player1ID
		room.BlackID = &p2
	} else {
		// player2 - белые
		room.WhiteID = &p2
		room.BlackID = &room.Player1ID
	}

	// 2. Создаём начальную позицию в FEN
	initialFen := game.MakeNewChessGame() // функция MakeNewChessGame
	room.BoardState = &initialFen
	// 3. Меняем статус на "playing"
	room.Status = "playing"

	err = db.UpdateRoom(room)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка обновления комнаты: "+err.Error())
		bot.Send(msg)
		return
	}

	// Теперь формируем ASCII-доску и сообщаем игрокам, кто за какой цвет
	//asciiBoard, _ := game.RenderBoardFromFEN(initialFen)
	asciiBoard, _ := game.RenderBoardCustom(*room.BoardState)

	whiteUser := "???"
	blackUser := "???"

	if room.WhiteID != nil && *room.WhiteID == room.Player1ID && room.Player1Username != nil {
		whiteUser = "@" + *room.Player1Username + " (♙)"
	} else if room.WhiteID != nil && *room.WhiteID == player2ID {
		whiteUser = "@" + secondUsername + " (♙)"
	}
	if room.BlackID != nil && *room.BlackID == room.Player1ID && room.Player1Username != nil {
		blackUser = "@" + *room.Player1Username + " (♟)"
	} else if room.BlackID != nil && *room.BlackID == player2ID {
		blackUser = "@" + secondUsername + " (♟)"
	}

	text := fmt.Sprintf("Игра началась!\n%s vs %s\n\nНачальная позиция:\n%s",
		whiteUser, blackUser, asciiBoard)
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))

	text = fmt.Sprintf("Начальная позиция:\n%s", asciiBoard)
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))
	// По желанию, оповещаем и первого игрока (через player1_id).
	// Можно хранить chat_id каждого игрока, если нужно отдельно рассылать.
	// Или просто писать в той же комнате, если игроки используют общий групповой чат.

	// Для простоты предположим, что это личная переписка с каждым игроком,
	// и мы знаем chatID (update.Message.Chat.ID). Но на практике —
	// лучше иметь chatID каждого игрока в БД.
}

func handleCreateRoomCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	player1ID := update.Message.From.ID

	// Достаём юзернейм (fallback -> FirstName)
	firstUsername := update.Message.From.UserName
	if firstUsername == "" {
		firstUsername = update.Message.From.FirstName
	}

	// Создаём запись в БД (без username, т.к. CreateRoom ещё не знает поля)
	room, err := db.CreateRoom(player1ID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Ошибка создания комнаты: "+err.Error()))
		return
	}

	// Теперь room создан. Присвоим username в нашей структуре и сделаем UpdateRoom
	room.Player1Username = &firstUsername
	// status = "waiting" — уже есть, но можно перестраховаться
	room.Status = "waiting"
	if err := db.UpdateRoom(room); err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Ошибка при обновлении комнаты (username): "+err.Error()))
		return
	}

	// Формируем ссылку-приглашение (как прежде)
	inviteLink := fmt.Sprintf("https://t.me/%s?start=room_%s", bot.Self.UserName, room.RoomID)
	text := fmt.Sprintf(
		"Комната создана!\n\nRoomID: %s\nСсылка: %s",
		room.RoomID, inviteLink,
	)

	// Добавим inline-кнопку «Создать и перейти в Чат»
	// Она не создаёт чат мгновенно, а просто показывает инструкцию (или ссылку)
	createChatButton := tgbotapi.NewInlineKeyboardButtonData(
		"Создать и перейти в Чат",
		fmt.Sprintf("create_chat_%s", room.RoomID),
	)
	// При нажатии на неё, пользователь получит инструкцию

	// Другие кнопки (Пригласить, Удалить комнату) как прежде

	// Кнопка «Пригласить» (Telegram Share)
	shareURL := fmt.Sprintf("https://t.me/share/url?url=%s&text=%s",
		url.QueryEscape(inviteLink),
		url.QueryEscape("Приглашаю сыграть в Telega-Chess!"),
	)
	inviteButton := tgbotapi.NewInlineKeyboardButtonURL("Пригласить", shareURL)

	// Кнопка «Удалить комнату»
	deleteButton := tgbotapi.NewInlineKeyboardButtonData("Удалить комнату", "delete_"+room.RoomID)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(inviteButton),
		tgbotapi.NewInlineKeyboardRow(createChatButton),
		tgbotapi.NewInlineKeyboardRow(deleteButton),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func handleSetRoomCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	args := update.Message.CommandArguments()
	// args = "546e81dc-5aff-463a-9681-3e41627b8df2"

	// Смотрим, есть ли такая roomID в БД
	room, err := db.GetRoomByID(args)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Комната не найдена. Убедитесь, что ввели корректный room_id."))
		return
	}

	// Сохраняем chat.ID
	chatID := update.Message.Chat.ID
	room.ChatID = &chatID
	err = db.UpdateRoom(room)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Не удалось обновить комнату: "+err.Error()))
		return
	}

	// Переименуем группу окончательно
	if room.Player1Username != nil {
		newTitle := fmt.Sprintf("tChess:@%s", *room.Player1Username)
		renameConfig := tgbotapi.SetChatTitleConfig{
			ChatID: chatID,
			Title:  newTitle,
		}
		if _, err := bot.Request(renameConfig); err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"Не смог переименовать группу. Дайте права 'Change group info'!"))
		}
	}

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("Теперь эта группа связана с комнатой %s!", room.RoomID)))
}

func handleGameListCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваши активные игры (заглушка).\n1. Комната 12345.\n2. Комната 67890.")
	bot.Send(msg)
}

func handlePlayWithBotCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Игра с ботом в разработке.")
	bot.Send(msg)
}

func handleUnknownCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Команда не распознана. Используйте /start для списка команд.")
	bot.Send(msg)
}

func HandleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	data := query.Data

	switch {
	case data == "manage_room":
		handleManageRoomMenu(bot, query)

	case data == "continue_setup":
		handleContinueSetup(bot, query)
	case strings.HasPrefix(data, "create_chat_"):
		// пользователь нажал "Создать и перейти в Чат"
		roomID := data[len("create_chat_"):]
		handleCreateChatInstruction(bot, query, roomID)

	case strings.HasPrefix(data, "delete_"):
		roomID := data[7:]
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Комната "+roomID+" будет удалена (заглушка).")
		bot.Send(msg)
	default:
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Неизвестный callback: "+data)
		bot.Send(msg)
	}

	// Подтверждаем callback
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("AnswerCallbackQuery error: %v", err)
	}
}

func handleCreateChatInstruction(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, roomID string) {
	// Здесь мы не создаём группу автоматически (Telegram API не даёт).
	// Просто даём инструкцию.
	instructionText := `
Чтобы создать новый групповой чат:
1) Выйдите в главное меню Telegram → «Новая группа»
2) Добавьте меня (@TelegaChessBot) в группу 
3) Назначьте меня администратором с правами «Change group info» и «Invite users»
4) Готово! Я автоматически переименую группу и приглашу второго игрока.
`
	// Подставим имя бота
	formattedText := fmt.Sprintf(instructionText, bot.Self.UserName)

	// Можно отправить alert (короткий всплывающий) или полноценное сообщение
	// Alert обычно ограничен по длине, лучше отправить отдельное сообщение
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, formattedText)
	bot.Send(msg)
}

func handleManageRoomMenu(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	// Показываем 2-3 кнопки:
	// 1) "Продолжить настройку"
	// 2) "Отмена" (или "Назад")

	continueBtn := tgbotapi.NewInlineKeyboardButtonData("Продолжить настройку", "continue_setup")
	cancelBtn := tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel_setup")

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(continueBtn),
		tgbotapi.NewInlineKeyboardRow(cancelBtn),
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "Выберите действие:")
	msg.ReplyMarkup = kb
	bot.Send(msg)
}

func handleContinueSetup(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	// Говорим пользователю: "Введите /setroom <room_id> в этом чате,
	// чтобы связать комнату с группой."

	text := `Чтобы завершить настройку, введите команду:
/setroom <room_id>
(например, /setroom 546e81dc-5aff-463a-9681-3e41627b8df2)
Это свяжет текущий групповой чат с вашей комнатой.`
	bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, text))
}
