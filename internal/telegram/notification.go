package telegram

import (
	"fmt"

	"telega_chess/internal/db"
	"telega_chess/internal/game"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func makeFinalTitle(r *db.Room) string {
	if r.Player1.Username == "" {
		return "tChess:????"
	}
	if r.Player2.Username == "" {
		return fmt.Sprintf("tChess:@%s_⚔️_??", r.Player1.Username)
	}
	return fmt.Sprintf("tChess:@%s_⚔️_@%s", r.Player1.Username, r.Player2.Username)
}

func tryRenameGroup(bot *tgbotapi.BotAPI, chatID int64, newTitle string) {
	renameConfig := tgbotapi.SetChatTitleConfig{
		ChatID: chatID,
		Title:  newTitle,
	}
	_, err := bot.Request(renameConfig)
	if err != nil {

		utils.Logger.Error(
			fmt.Sprintf("😖 Не удалось переименовать группу (chatID=%d): %v 🤕", chatID),
			zap.Error(err))

		// Сообщим пользователю, что нужны права
		retryBtn := tgbotapi.NewInlineKeyboardButtonData(
			"Повторить переименование",
			fmt.Sprintf("retry_rename:%s", newTitle),
		)
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(retryBtn),
		)
		msg := tgbotapi.NewMessage(chatID,
			"У меня нет прав на изменение названия группы. Дайте права 'Change group info' и нажмите [Повторить переименование].")
		msg.ReplyMarkup = kb
		bot.Send(msg)
	}
}

func notifyGameStarted(bot *tgbotapi.BotAPI, room *db.Room) {
	// 1) Сформируем текст "Игра началась!"
	introMsg := game.MakeGameStartedMessage(room)

	// 2) Отправим интро (в группу или в личку)
	sendMessageToRoomOrUsers(bot, room, introMsg)

	// 3) ASCII-доска
	fen := ""
	if room.BoardState != nil {
		fen = *room.BoardState
	}
	asciiBoard, err := game.RenderBoardCustom(fen) // или RenderBoardFromFEN
	//utils.Logger.Info("RenderBoardCustom() -> ", zap.String("board", asciiBoard))
	if err != nil {
		utils.Logger.Error("RenderBoardCustom() -> ", zap.Error(err))
		asciiBoard = "Ошибка генерации доски"
		return
	}

	// 4) Отправим ASCII доску
	sendMessageToRoomOrUsers(bot, room, asciiBoard)
	/*~~~~~~~~~~~~~~~~~~~~~~~*/

	////1) Обновим room (уже сделали), тут просто ещё раз получим актуальные данные
	//r, _ := db.GetRoomByID(room.RoomID)
	//
	//// 2) Генерим сообщение
	//msgText := game.MakeGameStartedMessage(r)
	//
	//// 3) ASCII-доска
	//fen := ""
	//if room.BoardState != nil {
	//	fen = *room.BoardState
	//}
	//asciiBoard, err := game.RenderBoardCustom(fen) // или RenderBoardFromFEN
	////utils.Logger.Info("RenderBoardCustom() -> ", zap.String("board", asciiBoard))
	//if err != nil {
	//	//bot.Send(tgbotapi.NewMessage(*r.ChatID, "Ошибка генерации доски RenderBoardCustom: "+err.Error()))
	//	utils.Logger.Error("RenderBoardCustom() -> ", zap.Error(err))
	//	return
	//}
	// 4) Куда отправить?
	//if r.ChatID != nil {
	//	// Это групповой чат
	//	bot.Send(tgbotapi.NewMessage(*r.ChatID, msgText))
	//	// Отправим ASCII-доску
	//	chattableBoard := tgbotapi.NewMessage(*r.ChatID, asciiBoard)
	//	chattableBoard.ParseMode = tgbotapi.ModeMarkdownV2
	//	bot.Send(chattableBoard)
	//} else {
	//	// 1:1 игра
	//	// Шлём обоим игрокам
	//	u1, _ := db.GetUserByID(r.Player1.ID)
	//	if r.Player2 != nil {
	//		u2, _ := db.GetUserByID(r.Player2.ID)
	//		// User1
	//		bot.Send(tgbotapi.NewMessage(u1.ChatID, msgText))
	//		chattableBoard := tgbotapi.NewMessage(u1.ChatID, asciiBoard)
	//		chattableBoard.ParseMode = tgbotapi.ModeMarkdownV2
	//		bot.Send(chattableBoard)
	//
	//		// User2
	//		bot.Send(tgbotapi.NewMessage(u2.ChatID, msgText))
	//		chattableBoard = tgbotapi.NewMessage(u2.ChatID, asciiBoard)
	//		chattableBoard.ParseMode = tgbotapi.ModeMarkdownV2
	//		bot.Send(chattableBoard)
	//	}
	//}
}

func sendMessageToRoom(bot *tgbotapi.BotAPI, room *db.Room, text string) error {
	// Если ChatID не задан, ничего не делаем
	if room.ChatID == nil {
		return fmt.Errorf("room.ChatID is nil, cannot send to group")
	}

	msg := tgbotapi.NewMessage(*room.ChatID, text)
	// опционально msg.ParseMode = "Markdown" or "HTML"
	_, err := bot.Send(msg)
	return err
}

func sendMessageToUsers(bot *tgbotapi.BotAPI, room *db.Room, text string) {
	// Выгружаем player1
	u1, err1 := db.GetUserByID(room.Player1.ID)
	if err1 == nil && u1.ChatID != 0 {
		m1 := tgbotapi.NewMessage(u1.ChatID, text)
		bot.Send(m1)
	}

	// Выгружаем player2, если есть
	if room.Player2 != nil {
		u2, err2 := db.GetUserByID(room.Player2.ID)
		if err2 == nil && u2.ChatID != 0 {
			m2 := tgbotapi.NewMessage(u2.ChatID, text)
			bot.Send(m2)
		}
	}
}

func sendMessageToRoomOrUsers(bot *tgbotapi.BotAPI, room *db.Room, text string) {
	// Если group chatID задан, шлём туда
	if room.ChatID != nil {
		err := sendMessageToRoom(bot, room, text)
		if err != nil {
			utils.Logger.Error("sendMessageToRoom error:", zap.Error(err))
		}
	} else {
		// Иначе шлём обоим
		sendMessageToUsers(bot, room, text)
	}
}
