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
	// Создаём запись в БД (без username, т.к. CreateRoom ещё не знает поля)
	utils.Logger.Info("handleCreateRoomCommand()", zap.Any("query", query))
	room, err := db.CreateRoom(query.From.ID)
	if err != nil {
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

func handleSetupRoomWhiteChoice(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, choice string) {
	userID := query.From.ID

	// Создаём новую комнату
	newRoom, err := db.CreateRoom(userID) // базовый метод
	if err != nil {
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
