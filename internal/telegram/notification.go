package telegram

import (
	"fmt"
	"sort"

	"telega_chess/internal/db"
	"telega_chess/internal/game"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/notnil/chess"
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
	sendMessageToRoomOrUsers(bot, room, introMsg, tgbotapi.ModeHTML)

	// 2) Отправим ASCII доску
	SendBoardToRoomOrUsers(bot, room)
	// 3) Подготавливаем и отправыялем кнопки
	prepareMoveButtons(bot, room, *room.WhiteID)
}

func sendMessageToRoom(bot *tgbotapi.BotAPI, room *db.Room, text string, mode string) error {
	// Если ChatID не задан, ничего не делаем
	if room.ChatID == nil {
		return fmt.Errorf("room.ChatID is nil, cannot send to group")
	}

	msg := tgbotapi.NewMessage(*room.ChatID, text)
	if mode != "" && (mode == tgbotapi.ModeMarkdownV2 || mode == tgbotapi.ModeHTML) {
		msg.ParseMode = mode
	} else {
		msg.ParseMode = tgbotapi.ModeMarkdown
	}
	_, err := bot.Send(msg)
	return err
}

func sendMessageToUser(bot *tgbotapi.BotAPI, userID int64, text string, mode string) {
	// Выгружаем user
	u1, err1 := db.GetUserByID(userID)
	if err1 == nil && u1.ChatID != 0 {
		m1 := tgbotapi.NewMessage(u1.ChatID, text)
		m1.ParseMode = mode
		bot.Send(m1)
	}
}

func sendMessageToRoomOrUsers(bot *tgbotapi.BotAPI, room *db.Room, text string, mode string) {
	// Если group chatID задан, шлём туда
	if room.ChatID != nil {
		err := sendMessageToRoom(bot, room, text, mode)
		if err != nil {
			utils.Logger.Error("sendMessageToRoom error:"+err.Error(), zap.Error(err))
		}
	} else {
		// Иначе шлём обоим
		sendMessageToUser(bot, room.Player1.ID, text, mode)
		sendMessageToUser(bot, room.Player2.ID, text, mode)
	}
}
func SendBoardToRoomOrUsers(bot *tgbotapi.BotAPI, r *db.Room) {
	var asciiBoard string
	var err error
	if r.ChatID != nil {
		// for chat(both)
		asciiBoard, err = game.RenderASCIIBoardHorizontal(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardWhite:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования горизонтальной доски"
		}
		sendMessageToRoomOrUsers(bot, r, asciiBoard, tgbotapi.ModeMarkdownV2)
	} else {
		// for White
		asciiBoard, err = game.RenderASCIIBoardWhite(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardWhite:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования горизонтальной доски"
		}
		sendMessageToUser(bot, *r.WhiteID, asciiBoard, tgbotapi.ModeMarkdownV2)
		// for Black
		asciiBoard, err = game.RenderASCIIBoardBlack(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardWhite:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования горизонтальной доски"
		}
		sendMessageToUser(bot, *r.BlackID, asciiBoard, tgbotapi.ModeMarkdownV2)
	}
}

func keyboardSort(slice []chess.Square) {
	sort.Slice(slice, func(i, j int) bool {
		// Получаем ранги (цифры) и файлы (буквы)
		rankI, fileI := slice[i].Rank(), slice[i].File()
		rankJ, fileJ := slice[j].Rank(), slice[j].File()

		// Сначала сравниваем ранги (по убыванию)
		if rankI != rankJ {
			return rankI > rankJ
		}

		// Если ранги равны, сравниваем файлы (по возрастанию)
		return fileI < fileJ
	})
}
