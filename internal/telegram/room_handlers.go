package telegram

import (
	"context"
	"fmt"
	"net/url"

	"lvlchess/internal/db/models"
	"lvlchess/internal/db/repositories"
	"lvlchess/internal/game"
	"lvlchess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// handleCreateRoomCommand is invoked when a user clicks the "Создать комнату" (Create Room) inline button.
// We create a new Room record, generate an invite link, and return it, plus optional inline options to
// create a group chat or delete the room.
func (h *Handler) handleCreateRoomCommand(ctx context.Context, query *tgbotapi.CallbackQuery) {
	// Prepare a new room for the user who clicked the button.
	room := models.PrepareNewRoom(query.From.ID, h.MakeFinalTitle(ctx, nil))

	if err := h.RoomRepo.CreateRoom(ctx, room); err != nil {
		// If there's a unique violation error, maybe a room with those two players already exists.
		if err.Error() == repositories.ErrUniqueViolation {
			// We could handle it, e.g. checkExistingRoom(...).
			return
		}

		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID,
			"Ошибка создания комнаты: "+err.Error()))
		return
	}

	// Generate a standard link like t.me/BOTUSERNAME?start=room_<roomID>
	inviteLink := fmt.Sprintf("https://t.me/%s?start=room_%s", h.Bot.Self.UserName, room.RoomID)
	text := fmt.Sprintf("Комната создана!\n\nRoomID: %s\nСсылка: %s", room.RoomID, inviteLink)

	// Provide an inline button to "Create and go to Chat"
	createChatButton := tgbotapi.NewInlineKeyboardButtonData(
		"Создать и перейти в Чат",
		fmt.Sprintf("%s%s", CreateChat, room.RoomID),
	)

	// A second button "Invite" that uses the telegram share/url scheme
	shareURL := fmt.Sprintf("https://t.me/share/url?url=%s&text=%s",
		url.QueryEscape(inviteLink),
		url.QueryEscape("Приглашаю сыграть в Telega-Chess!"),
	)
	inviteButton := tgbotapi.NewInlineKeyboardButtonURL("Пригласить", shareURL)

	// A third button "Удалить комнату"
	deleteButton := tgbotapi.NewInlineKeyboardButtonData("Удалить комнату", Delete+room.RoomID)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(inviteButton),
		tgbotapi.NewInlineKeyboardRow(createChatButton),
		tgbotapi.NewInlineKeyboardRow(deleteButton),
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	h.Bot.Send(msg)
}

// handleJoinRoom is triggered in two scenarios:
// 1) The user typed /start room_<id> in private chat.
// 2) Possibly a callback if we had "join_this_room:<roomID>" style callback.
// The user becomes Player2 if Player2 is nil, or we report "already full."
func (h *Handler) handleJoinRoom(ctx context.Context, update tgbotapi.Update, roomID string) {
	newPlayer := &models.User{
		ID:        update.Message.From.ID,
		Username:  update.Message.From.UserName,
		FirstName: update.Message.From.FirstName,
		ChatID:    update.Message.Chat.ID,
	}
	h.UserRepo.CreateOrUpdateUser(ctx, newPlayer)

	room, err := h.RoomRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Комната не найдена: "+err.Error())
		h.Bot.Send(msg)
		return
	}

	if room.Player1ID == newPlayer.ID {
		// We do not allow a user to join their own room as Player2
		h.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы не можете присоединиться к собственной комнате :)"))
		return
	}

	// Optionally check if there's an existing active room for the same pair of players
	h.checkExistingRoom(ctx, room.Player1ID, newPlayer.ID)

	if room.Player2ID != nil {
		// Room is already full
		h.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "В этой комнате уже есть второй игрок."))
		return
	}

	// If vacant, we assign the second player
	room.Player2ID = &newPlayer.ID
	room.Status = models.RoomStatusPlaying

	// Decide randomly who is White or Black if not set
	game.AssignRandomColors(room)

	// Update the title to reflect both participants
	room.RoomTitle = h.MakeFinalTitle(ctx, room)

	// Notify that the game has started
	h.notifyGameStarted(ctx, room)

	if err = h.RoomRepo.UpdateRoom(ctx, room); err != nil {
		h.Bot.Send(tgbotapi.NewMessage(newPlayer.ChatID, "Ошибка обновления комнаты: "+err.Error()))
		return
	}
}

// handleSetRoomCommand is used in a group chat to link that chat to a specific room via /setroom <roomID> command.
// Once linked, the bot can rename the group, manage invites, etc.
func (h *Handler) handleSetRoomCommand(ctx context.Context, update tgbotapi.Update) {
	args := update.Message.CommandArguments()
	if args == "" {
		h.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Пожалуйста, укажите room_id, например:\n/setroom 546e81dc-5aff-463a-9681-3e41627b8df2"))
		return
	}

	room, err := h.RoomRepo.GetRoomByID(ctx, args)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Комната не найдена. Проверьте идентификатор."))
		return
	}

	chatID := update.Message.Chat.ID
	room.ChatID = &chatID
	if err := h.RoomRepo.UpdateRoom(ctx, room); err != nil {
		h.Bot.Send(tgbotapi.NewMessage(chatID, "Не удалось сохранить chatID в БД: "+err.Error()))
		return
	}

	// Optionally rename the group to something referencing Player1 username
	p1, err := h.UserRepo.GetUserByID(ctx, room.Player1ID)
	if err != nil {
		h.Bot.Send(tgbotapi.NewMessage(chatID, "Не удалось определить пользователя: "+err.Error()))
		return
	}

	if p1.Username != "" {
		h.tryRenameGroup(h.Bot, chatID, fmt.Sprintf("tChess:@%s", p1.Username))
	}

	h.Bot.Send(tgbotapi.NewMessage(chatID,
		fmt.Sprintf("Группа успешно привязана к комнате %s!", room.RoomID)))

	// If there's no second player, create an invite link
	if room.Player2ID == nil {
		linkCfg := tgbotapi.ChatInviteLinkConfig{
			ChatConfig: tgbotapi.ChatConfig{ChatID: chatID},
		}
		inviteLink, err := h.Bot.GetInviteLink(linkCfg)
		if err != nil {
			h.Bot.Send(tgbotapi.NewMessage(chatID, "Ошибка создания ссылки-приглашения: "+err.Error()))
			return
		}
		text := fmt.Sprintf(
			"Сейчас в комнате только вы. Отправьте второму игроку эту ссылку:\n%s",
			inviteLink,
		)
		h.Bot.Send(tgbotapi.NewMessage(chatID, text))
	} else {
		// If 2nd player is already known, we rename the group using final title and start the game
		room.Status = models.RoomStatusPlaying
		newTitle := h.MakeFinalTitle(ctx, room)
		h.tryRenameGroup(h.Bot, chatID, newTitle)

		game.AssignRandomColors(room)
		room.RoomTitle = newTitle
		h.RoomRepo.UpdateRoom(ctx, room)

		h.notifyGameStarted(ctx, room)
	}
}

// handleSetupRoomWhiteChoice is used in the scenario "Кто будет за белых?" => "me" or "opponent".
// Then we create a new room, assign WhiteID or BlackID, and confirm creation.
func (h *Handler) handleSetupRoomWhiteChoice(ctx context.Context, query *tgbotapi.CallbackQuery, choice string) {
	userID := query.From.ID

	newRoom := models.PrepareNewRoom(userID, h.MakeFinalTitle(ctx, nil))
	if err := h.RoomRepo.CreateRoom(ctx, newRoom); err != nil {
		// If there's any DB error, handle it here
		return
	}

	// If user says "me," we explicitly set them as White; if "opponent," we set them as Black.
	if choice == "me" {
		newRoom.WhiteID = &userID
	} else {
		newRoom.BlackID = &userID
	}
	newRoom.IsWhiteTurn = true // default: white moves first

	err := h.RoomRepo.UpdateRoom(ctx, newRoom)
	if err != nil {
		return
	}

	roomCreatedMsg := fmt.Sprintf("Комната создана!\nRoomID: %s\nХод белых.\nWhiteID=%v, BlackID=%v",
		newRoom.RoomID, newRoom.WhiteID, newRoom.BlackID)
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, roomCreatedMsg))
}

// handleAskWhoIsWhite simply asks the user "Who will be white?" with two inline buttons: "me" or "opponent."
func (h *Handler) handleAskWhoIsWhite(ctx context.Context, query *tgbotapi.CallbackQuery) {
	btnMe := tgbotapi.NewInlineKeyboardButtonData("Я сам (создатель)", SetupRoomWhite+":me")
	btnOpponent := tgbotapi.NewInlineKeyboardButtonData("Соперник (второй игрок)", SetupRoomWhite+":opponent")

	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnMe, btnOpponent),
	)

	text := "Кто будет играть за белых?"
	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ReplyMarkup = kb
	h.Bot.Send(msg)
}

// handleRoomEntrance is a placeholder approach for user "entering" a room. Possibly sets user.CurrentRoom
// so that all subsequent commands in private chat apply to that room.
func (h *Handler) handleRoomEntrance(ctx context.Context, query *tgbotapi.CallbackQuery, roomID string) {
	userID := query.From.ID

	room, err := h.RoomRepo.GetRoomByID(ctx, roomID)
	if err != nil || room == nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Комната не найдена."))
		return
	}

	// E.g. we could check if user is either player1 or player2
	if room.Player2ID == nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Комната ещё не сформирована полностью."))
		return
	}
	if room.Player1ID != userID && (room.Player2ID == nil || *room.Player2ID != userID) {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Вы не являетесь участником этой комнаты."))
		return
	}

	u, _ := h.UserRepo.GetUserByID(ctx, userID)
	u.CurrentRoom = &models.Room{RoomID: roomID}
	h.UserRepo.CreateOrUpdateUser(ctx, u)

	text := fmt.Sprintf("Вы вошли в комнату %s (%s). В личке теперь используете её для ходов.",
		room.RoomID, room.RoomTitle)
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, text))

	// If it's the user's turn, we might call prepareMoveButtons right away
	if (room.IsWhiteTurn && room.WhiteID != nil && *room.WhiteID == userID) ||
		(!room.IsWhiteTurn && room.BlackID != nil && *room.BlackID == userID) {
		h.prepareMoveButtons(ctx, room, userID)
	}
}

func (h *Handler) handleChooseRoom(ctx context.Context, query *tgbotapi.CallbackQuery, roomID string) {
	room, err := h.RoomRepo.GetRoomByID(ctx, roomID)
	if err != nil || room == nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID,
			"Комната не найдена."))
		return
	}

	asciiBoard, err := game.RenderASCIIBoardBlack(room.BoardState)
	if err != nil {
		asciiBoard = "Ошибка формирования доски"
	}

	text := fmt.Sprintf("Войти в комнату_№%s (ход @...)?\n%s", room.RoomTitle)
	h.sendMessageToUser(ctx, query.Message.Chat.ID, text, tgbotapi.ModeHTML)

	callbackData := fmt.Sprintf("join_this_room:%s", room.RoomID)
	btn := tgbotapi.NewInlineKeyboardButtonData("Вход", callbackData)
	kb := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{btn})

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, asciiBoard)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = kb

	h.Bot.Send(msg)
}

func (h *Handler) handleJoinThisRoom(ctx context.Context, query *tgbotapi.CallbackQuery, roomID string) {
	userID := query.From.ID
	room, err := h.RoomRepo.GetRoomByID(ctx, roomID)
	if err != nil || room == nil {
		h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Комната не найдена."))
		return
	}

	u, _ := h.UserRepo.GetUserByID(ctx, userID)
	u.CurrentRoom = &models.Room{RoomID: roomID}
	h.UserRepo.CreateOrUpdateUser(ctx, u)

	text := fmt.Sprintf("Вы зашли в комнату %s. В личке теперь используете её для ходов.", room.RoomTitle)
	h.Bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, text))
	if (room.IsWhiteTurn && *room.WhiteID == userID) || (!room.IsWhiteTurn && *room.BlackID == userID) {
		h.prepareMoveButtons(ctx, room, userID)
	}
}

// checkExistingRoom is an optional method to see if p1ID and p2ID already have an active room.
// If so, we might display "You already have a room with this opponent: <existingRoom.Title>," and
// provide an inline button to re-enter that existing room.
func (h *Handler) checkExistingRoom(ctx context.Context, p1ID, p2ID int64) bool {
	existingRoom, err := h.RoomRepo.GetRoomByPlayerIDs(ctx, p1ID, p2ID)
	if err != nil {
		utils.Logger.Info("checkExistingRoom() found no existing or error", zap.Any("p1ID", p1ID), zap.Any("p2ID", p2ID))
		return false
	}
	if existingRoom != nil {
		text := fmt.Sprintf("У вас уже есть комната с этим соперником: %s\n", existingRoom.RoomTitle)
		callbackData := fmt.Sprintf("%s:%s", RoomEntrance, existingRoom.RoomID)
		btn := tgbotapi.NewInlineKeyboardButtonData("Войти в комнату", callbackData)
		kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))

		SendInlineKeyboard(h.Bot, existingRoom, text, kb)
		return true
	}
	return false
}
