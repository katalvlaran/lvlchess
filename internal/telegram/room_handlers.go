package telegram

import (
	"fmt"
	"net/url"

	"telega_chess/internal/db"
	"telega_chess/internal/game"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// func handleCreateRoomCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
func handleCreateRoomCommand(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	// Подготавливаем Player1 (user = Room.Player1)
	userID := query.From.ID
	p1, err := db.GetUserByID(userID)
	// Создаём новую комнату
	room := db.PrepareNewRoom(p1, MakeFinalTitle(nil))
	if err = db.CreateRoom(room); err != nil {
		if err.Error() == db.ErrUniqueViolation {
			// Ищем уже существующую комнату
			//checkExistingRoom(bot, p1.ID, 0, query.Message.Chat.ID)

			return
		}

		// Иначе обрабатываем как прежде
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID,
			"Ошибка создания комнаты: "+err.Error()))
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

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func handleJoinRoom(bot *tgbotapi.BotAPI, update tgbotapi.Update, roomID string) {
	// Сохраним/обновим user
	newPlayer := &db.User{
		ID:        update.Message.From.ID,
		Username:  update.Message.From.UserName,
		FirstName: update.Message.From.FirstName,
		ChatID:    update.Message.Chat.ID, // личка
	}
	db.CreateOrUpdateUser(newPlayer)

	room, err := db.GetRoomByID(roomID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Комната не найдена: "+err.Error())
		bot.Send(msg)
		return
	}

	if room.Player1.ID == newPlayer.ID {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не можете присоединиться к собственной комнате :)")
		bot.Send(msg)
		return
	}

	// Проверяем нет ли уже существующей комнаты с room.P
	checkExistingRoom(bot, room.Player1.ID, newPlayer.ID)
	/*	if existingRoom, _ := db.GetRoomByPlayerIDs(room.Player1.ID, newPlayer.ID); existingRoom != nil {
		// Далее: "У вас уже есть комната: room.Title"
		text := fmt.Sprintf(
			"У вас уже есть комната с этим соперником: %s\n",
			existingRoom.RoomTitle)
		// + добавляем кнопку «Войти в комнату»
		callbackData := fmt.Sprintf("room_entrance:%s", existingRoom.RoomID)
		btn := tgbotapi.NewInlineKeyboardButtonData("Войти в комнату", callbackData)
		kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
		SendInlineKeyboard(bot, room, text, kb)
		SendInlineKeyboard(bot, existingRoom, text, kb)

		return
	}*/

	if room.Player2 != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "В этой комнате уже есть второй игрок.")
		bot.Send(msg)
		return
	}

	// Присвоим второго игрока
	room.Player2 = newPlayer
	room.Status = db.RoomStatusPlaying
	game.AssignRandomColors(room)
	room.RoomTitle = MakeFinalTitle(room)
	notifyGameStarted(bot, room)
	if err = db.UpdateRoom(room); err != nil {
		bot.Send(tgbotapi.NewMessage(newPlayer.ChatID, "Ошибка обновления комнаты: "+err.Error()))
		return
	}
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
		room.Status = db.RoomStatusPlaying
		newTitle := MakeFinalTitle(room)
		tryRenameGroup(bot, chatID, newTitle)
		game.AssignRandomColors(room)
		room.RoomTitle = newTitle
		db.UpdateRoom(room)
		notifyGameStarted(bot, room)
	}
}

func handleSetupRoomWhiteChoice(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, choice string) {
	userID := query.From.ID

	p1, err := db.GetUserByID(userID)
	// Создаём новую комнату
	newRoom := db.PrepareNewRoom(p1, MakeFinalTitle(nil))
	if err = db.CreateRoom(newRoom); err != nil {
		// handle err
		return
	}
	// newRoom.IsWhiteTurn = true (по умолчанию)

	// Если "me" => newRoom.WhiteID = userID
	// Если "opponent" => newRoom.WhiteID = nil, newRoom.BlackID = userID
	if choice == "me" {
		newRoom.WhiteID = &userID
		// BlackID = nil
	} else {
		// "opponent"
		newRoom.WhiteID = nil
		newRoom.BlackID = &userID
	}
	newRoom.IsWhiteTurn = true

	// update DB
	err = db.UpdateRoom(newRoom)
	if err != nil {
		// handle err
		return
	}

	// Отправляем ответ, например "Комната создана! RoomID = ... Напишите /start room_XXX"
	roomCreatedMsg := fmt.Sprintf("Комната создана!\nRoomID: %s\nХод белых.\nWhiteID=%v, BlackID=%v",
		newRoom.RoomID, newRoom.WhiteID, newRoom.BlackID)
	bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, roomCreatedMsg))
}

func askWhoIsWhite(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	// отправим 2 кнопки
	btnMe := tgbotapi.NewInlineKeyboardButtonData("Я сам (создатель)", "setup_room_white:me")
	btnOpponent := tgbotapi.NewInlineKeyboardButtonData("Соперник (второй игрок)", "setup_room_white:opponent")

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnMe, btnOpponent),
	)

	text := "Кто будет играть за белых?"
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ReplyMarkup = kb
	bot.Send(msg)
}

func handleRoomEntrance(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, roomID string) {
	userID := query.From.ID
	// 1) Найдём комнату
	room, err := db.GetRoomByID(roomID)
	if err != nil || room == nil {
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Комната не найдена."))
		return
	}

	// 2) Проверим, имеет ли пользователь отношение к этой комнате
	//    (или разрешаем любому входить?)
	if room.Player1 == nil || room.Player2 == nil {
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Комната ещё не сформирована полностью."))
		return
	}
	if room.Player1.ID != userID && (room.Player2 == nil || room.Player2.ID != userID) {
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Вы не являетесь участником этой комнаты."))
		return
	}

	// 3) user.CurrentRoomID = roomID
	user, _ := db.GetUserByID(userID)
	//user.CurrentRoomID = roomID
	user.CurrentRoom = &db.Room{RoomID: roomID}
	db.CreateOrUpdateUser(user)

	// 4) Сообщим: "Теперь вы вошли в комнату %s"
	text := fmt.Sprintf(
		"Вы вошли в комнату %s (%s). Теперь в личке все действия идут в контексте этой комнаты.",
		room.RoomID,
		room.RoomTitle)
	bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, text))

	// Можно сразу вызвать prepareMoveButtons(bot, room, userID),
	// если userID == текущий ход.
}

func handleChooseRoom(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, roomID string) {
	// 1) Найдём room
	room, err := db.GetRoomByID(roomID)
	if err != nil || room == nil {
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID,
			"Комната не найдена."))
		return
	}

	// 2) Генерируем ASCII-доску (пусть будет BlackBoard)
	//    Или WhiteBoard, или HorizontalBoard, на ваш выбор
	asciiBoard, err := game.RenderASCIIBoardBlack(room.BoardState)
	if err != nil {
		asciiBoard = "Ошибка формирования доски"
	}

	// 3) "Войти в комнату?"
	// text = ...
	text := fmt.Sprintf("Войти в комнату_№%s (ход @...)?\n%s", room.RoomTitle)
	sendMessageToUser(bot, query.Message.Chat.ID, text, tgbotapi.ModeHTML)
	// 4) Создаём кнопку "Вход"
	callbackData := fmt.Sprintf("join_this_room:%s", room.RoomID)
	btn := tgbotapi.NewInlineKeyboardButtonData("Вход", callbackData)
	kb := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, asciiBoard)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = kb
	bot.Send(msg)
}

func handleJoinThisRoom(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, roomID string) {
	userID := query.From.ID
	room, err := db.GetRoomByID(roomID)
	if err != nil || room == nil {
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Комната не найдена."))
		return
	}
	// возможно, check user belongs to that room
	// затем user.CurrentRoomID = roomID
	u, _ := db.GetUserByID(userID)
	u.CurrentRoom = &db.Room{RoomID: roomID}
	db.CreateOrUpdateUser(u)

	text := fmt.Sprintf("Вы зашли в комнату %s. В личке теперь используете её для ходов.", room.RoomTitle)
	bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, text))
	if (room.IsWhiteTurn && *room.WhiteID == userID) || (!room.IsWhiteTurn && *room.BlackID == userID) {
		prepareMoveButtons(bot, room, userID)
	}
}

func checkExistingRoom(bot *tgbotapi.BotAPI, p1ID, p2ID int64 /*, chatID int64*/) bool {
	// true => есть уже
	existingRoom, err := db.GetRoomByPlayerIDs(p1ID, p2ID)
	if err != nil {
		utils.Logger.Info("FindRoomByPlayerIDs()", zap.Any("p1ID:", p1ID), zap.Any("p2ID:", p2ID))
		return false
	}
	// Проверяем нет ли уже существующей комнаты с room.P
	if existingRoom != nil {
		// Далее: "У вас уже есть комната: room.Title"
		text := fmt.Sprintf(
			"У вас уже есть комната с этим соперником: %s\n",
			existingRoom.RoomTitle)
		// + добавляем кнопку «Войти в комнату»
		callbackData := fmt.Sprintf("room_entrance:%s", existingRoom.RoomID)
		btn := tgbotapi.NewInlineKeyboardButtonData("Войти в комнату", callbackData)
		kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
		//SendInlineKeyboard(bot, room, text, kb)
		SendInlineKeyboard(bot, existingRoom, text, kb)

		return true
	}
	return false

}
