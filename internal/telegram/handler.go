package telegram

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"telega_chess/internal/db"
	"telega_chess/internal/game"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

var Rooms = make(map[string]int64) // roomID -> player1ID
/*
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
}*/

func handleStartCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	// 1) Сохраним пользователя
	p1 := db.User{
		ID:        update.Message.From.ID,
		Username:  update.Message.From.UserName,
		FirstName: update.Message.From.FirstName,
		ChatID:    update.Message.Chat.ID, // Личная переписка
	}
	db.CreateOrUpdateUser(&p1)

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
	// Сохраним/обновим user
	np := &db.User{
		ID:        update.Message.From.ID,
		Username:  update.Message.From.UserName,
		FirstName: update.Message.From.FirstName,
		ChatID:    update.Message.Chat.ID, // личка
	}
	db.CreateOrUpdateUser(np)

	room, err := db.GetRoomByID(roomID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Комната не найдена: "+err.Error())
		bot.Send(msg)
		return
	}

	if room.Player1.ID == np.ID {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не можете присоединиться к собственной комнате :)")
		bot.Send(msg)
		return
	}
	if room.Player2 != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "В этой комнате уже есть второй игрок.")
		bot.Send(msg)
		return
	}

	// Присвоим второго игрока
	room.Player2 = np
	room.Status = "playing"
	game.AssignRandomColors(room)
	if err := db.UpdateRoom(room); err != nil {
		bot.Send(tgbotapi.NewMessage(np.ChatID, "Ошибка обновления комнаты: "+err.Error()))
		return
	}

	notifyGameStarted(bot, room)
	// Готовимся к началу игры:
	/*// 1. Случайное назначение цветов (player1 белыми или player2 белыми).
	rand.Seed(time.Now().UnixNano()) // или инициализировать в main()
	if rand.Intn(2) == 0 {
		// player1 - белые
		room.WhiteID = &room.Player1.ID
		room.BlackID = &player2ID
	} else {
		// player2 - белые
		room.WhiteID = &player2ID
		room.BlackID = &room.Player1.ID
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
	//asciiBoard, _ := game.RenderBoardCustom(initialFen)
	asciiBoard, _ := game.RenderBoardCustom(*room.BoardState)

	whiteUser := "???"
	blackUser := "???"

	if room.WhiteID != nil && *room.WhiteID == room.Player1.ID && room.Player1.Username != "" {
		whiteUser = "@" + room.Player1.Username + " (♙)"
	} else if room.WhiteID != nil && *room.WhiteID == player2ID {
		whiteUser = "@" + secondUsername + " (♙)"
	}
	if room.BlackID != nil && *room.BlackID == room.Player1.ID && room.Player1.Username != "" {
		blackUser = "@" + room.Player1.Username + " (♟)"
	} else if room.BlackID != nil && *room.BlackID == player2ID {
		blackUser = "@" + secondUsername + " (♟)"
	}

	text := fmt.Sprintf("Игра началась!\n%s vs %s\n\nНачальная позиция:\n%s",
		whiteUser, blackUser, asciiBoard)
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))*/

	// По желанию, оповещаем и первого игрока (через player1_id).
	// Можно хранить chat_id каждого игрока, если нужно отдельно рассылать.
	// Или просто писать в той же комнате, если игроки используют общий групповой чат.

	// Для простоты предположим, что это личная переписка с каждым игроком,
	// и мы знаем chatID (update.Message.Chat.ID). Но на практике —
	// лучше иметь chatID каждого игрока в БД.
}

func handleCreateRoomCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	/*player1ID := update.Message.From.ID
	username := update.Message.From.UserName
	firstName := update.Message.From.FirstName
	chatID := update.Message.Chat.ID // Личная переписка

		// 1) Сохраним пользователя
		p1 := db.User{
			ID:        player1ID,
			Username:  username,
			FirstName: firstName,
			ChatID:    chatID,
		}
		db.CreateOrUpdateUser(&p1)
	*/
	// Создаём запись в БД (без username, т.к. CreateRoom ещё не знает поля)
	room, err := db.CreateRoom(update.Message.From.ID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Ошибка создания комнаты: "+err.Error()))
		return
	}

	/*	// status = "waiting" — уже есть, но можно перестраховаться
		room.Status = "waiting"
		if err := db.UpdateRoom(room); err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"Ошибка при обновлении комнаты (username): "+err.Error()))
			return
		}*/

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
	if args == "" {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Пожалуйста, укажите room_id, например:\n/setroom 546e81dc-5aff-463a-9681-3e41627b8df2"))
		return
	}

	room, err := db.GetRoomByID(args)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Комната не найдена. Проверьте идентификатор."))
		return
	}

	// Привяжем chat.ID
	chatID := update.Message.Chat.ID
	room.ChatID = &chatID
	if err := db.UpdateRoom(room); err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось сохранить chatID в БД: "+err.Error()))
		return
	}

	// Переименуем на основе player1Username, если хотим
	if room.Player1.Username != "" {
		tryRenameGroup(bot, chatID, fmt.Sprintf("tChess:@%s", room.Player1.Username))
	}

	// Сообщим "Готово!"
	bot.Send(tgbotapi.NewMessage(chatID,
		fmt.Sprintf("Группа успешно привязана к комнате %s!", room.RoomID)))

	// Теперь проверим, есть ли player2ID
	if room.Player2 == nil {
		// Предлагаем пригласить второго
		// Создадим invite-link
		linkCfg := tgbotapi.ChatInviteLinkConfig{
			ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
			// Можно ExpireDate, MemberLimit...
		}
		inviteLink, err := bot.GetInviteLink(linkCfg)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка создания ссылки-приглашения: "+err.Error()))
			return
		}

		// Отправляем
		text := fmt.Sprintf(
			"Сейчас в комнате только вы. Отправьте второму игроку эту ссылку:\n%s",
			inviteLink,
		)
		bot.Send(tgbotapi.NewMessage(chatID, text))
	} else {
		// Второй игрок есть => "Игра началась!"
		room.Status = "playing"
		if room.Player2.Username != "" {
			newTitle := fmt.Sprintf("tChess:@%s_⚔️_@%s",
				room.Player1.Username, room.Player2.Username)
			tryRenameGroup(bot, chatID, newTitle)
		}
		game.AssignRandomColors(room)
		db.UpdateRoom(room)
		notifyGameStarted(bot, room)
	}
}

func handleGameListCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ваши активные игры (заглушка).\n1. Комната 12345.\n2. Комната 67890.")
	bot.Send(msg)
}

func handlePlayWithBotCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Игра с ботом в разработке.")
	bot.Send(msg)
}

/*func handleUnknownCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Команда не распознана. Используйте /start для списка команд.")
	bot.Send(msg)
}*/

func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	data := query.Data

	switch {
	case data == "manage_room":
		handleManageRoomMenu(bot, query)
	case data == "continue_setup":
		handleContinueSetup(bot, query)
	case strings.HasPrefix(data, "retry_rename:"):
		newTitle := data[len("retry_rename:"):]
		handleRetryRename(bot, query, newTitle)
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
	*пожалуйсиа, постарайтесь создать простую группу(где будите только Вы)
2) Добавьте меня (@TelegaChessBot) в группу 
3) Перейдите в настройки группы и назначьте меня администратором (минимум с правами «Change group info» и «Invite users»)
4) Готово! Я автоматически переименую группу и приглашу второго игрока.
`
	// Подставим имя бота
	formattedText := fmt.Sprintf(instructionText, bot.Self.UserName)

	/*
		// Можно отправить alert (короткий всплывающий) или полноценное сообщение
		// Alert обычно ограничен по длине, лучше отправить отдельное сообщение
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, formattedText)
		bot.Send(msg)*/
	bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, formattedText))

	hint := tgbotapi.NewMessage(query.Message.Chat.ID, formattedText)
	hint.ParseMode = tgbotapi.ModeMarkdownV2
	bot.Send(hint)
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
	chatID := query.Message.Chat.ID

	// Проверим, есть ли уже room, привязанная к этому chatID
	room, err := db.GetRoomByChatID(chatID)
	if err != nil {
		// Нет привязки
		text := `
Пока к этой группе не привязана никакая комната.
Введите команду /setroom <room_id> для привязки:
Например: /setroom 546e81dc-5aff-463a-9681-3e41627b8df2
`
		bot.Send(tgbotapi.NewMessage(chatID, text))
		return
	}

	// Если есть, проверим, есть ли второй игрок
	if room.Player2 == nil {
		// Предлагаем сгенерировать invite-link
		linkCfg := tgbotapi.ChatInviteLinkConfig{
			ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
		}
		link, err := bot.GetInviteLink(linkCfg)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Ошибка при создании invite-link: "+err.Error()))
			return
		}

		text := fmt.Sprintf("Комната уже привязана к room_id=%s, но пока нет второго игрока.\n"+
			"Пригласите его ссылкой:\n%s", room.RoomID, link)
		bot.Send(tgbotapi.NewMessage(chatID, text))
	} else {
		// Есть 2 игрока => "Игра началась!" (или уже идёт)
		newTitle := makeFinalTitle(room)
		tryRenameGroup(bot, chatID, newTitle)

		notifyGameStarted(bot, room)
	}
}

func handleRetryRename(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, newTitle string) {
	// Просто заново вызываем tryRenameGroup
	// chatID = query.Message.Chat.ID
	tryRenameGroup(bot, query.Message.Chat.ID, newTitle)
}

// ------------------------------
// notifyGameStarted - общая логика вывода "Игра началась!"
func notifyGameStarted(bot *tgbotapi.BotAPI, room *db.Room) {
	// 1) Обновим room (уже сделали), тут просто ещё раз получим актуальные данные
	r, _ := db.GetRoomByID(room.RoomID)

	// 2) Генерим сообщение
	msgText := game.MakeGameStartedMessage(r)

	// 3) ASCII-доска
	fen := ""
	if room.BoardState != nil {
		fen = *room.BoardState
	}
	asciiBoard, err := game.RenderBoardCustom(fen) // или RenderBoardFromFEN
	//utils.Logger.Info("RenderBoardCustom() -> ", zap.String("board", asciiBoard))
	if err != nil {
		//bot.Send(tgbotapi.NewMessage(*r.ChatID, "Ошибка генерации доски RenderBoardCustom: "+err.Error()))
		utils.Logger.Error("RenderBoardCustom() -> ", zap.Error(err))
		return
	}

	// 4) Куда отправить?
	if r.ChatID != nil {
		// Это групповой чат
		bot.Send(tgbotapi.NewMessage(*r.ChatID, msgText))
		// Отправим ASCII-доску
		chattableBoard := tgbotapi.NewMessage(*r.ChatID, asciiBoard)
		chattableBoard.ParseMode = tgbotapi.ModeMarkdownV2
		bot.Send(chattableBoard)
	} else {
		// 1:1 игра
		// Шлём обоим игрокам
		u1, _ := db.GetUserByID(r.Player1.ID)
		if r.Player2 != nil {
			u2, _ := db.GetUserByID(r.Player2.ID)
			// User1
			bot.Send(tgbotapi.NewMessage(u1.ChatID, msgText))
			chattableBoard := tgbotapi.NewMessage(u1.ChatID, asciiBoard)
			chattableBoard.ParseMode = tgbotapi.ModeMarkdownV2
			bot.Send(chattableBoard)

			// User2
			bot.Send(tgbotapi.NewMessage(u2.ChatID, msgText))
			chattableBoard = tgbotapi.NewMessage(u2.ChatID, asciiBoard)
			chattableBoard.ParseMode = tgbotapi.ModeMarkdownV2
			bot.Send(chattableBoard)
		}
	}
}
