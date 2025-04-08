package telegram

import (
	"context"
	"fmt"
	"sort"

	"telega_chess/internal/db/models"
	"telega_chess/internal/game"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/notnil/chess"
	"go.uber.org/zap"
)

func (h *Handler) MakeFinalTitle(ctx context.Context, r *models.Room) (title string) {
	title = "tChess:????"
	if r != nil && r.Player1ID != 0 {
		p1, err := h.UserRepo.GetUserByID(ctx, r.Player1ID)
		if err != nil {
			return
		}
		title = fmt.Sprintf("tChess:@%s_⚔️_??", p1.Username)
		if r.Player2ID != nil {
			p2, err := h.UserRepo.GetUserByID(ctx, *r.Player2ID)
			if err != nil {
				return
			}
			title = fmt.Sprintf("@%s_⚔️_@%s", p1.Username, p2.Username)
		}
	}

	return
}

func (h *Handler) tryRenameGroup(bot *tgbotapi.BotAPI, chatID int64, newTitle string) {
	//func tryRenameGroup(bot *tgbotapi.BotAPI, room *Room, newTitle string) {
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

func (h *Handler) notifyGameStarted(ctx context.Context, room *models.Room) {
	// 1) Сформируем текст "Игра началась!"
	introMsg := "Игра началась!\n" + room.RoomTitle
	// 2) Отправим интро (в группу или в личку)
	h.sendMessageToRoomOrUsers(ctx, room, introMsg, tgbotapi.ModeHTML)

	// 2) Отправим ASCII доску
	h.SendBoardToRoomOrUsers(ctx, room)
	// 3) Подготавливаем и отправыялем кнопки
	h.prepareMoveButtons(ctx, room, *room.WhiteID)
}

func (h *Handler) sendMessageToRoom(ctx context.Context, room *models.Room, text string, mode string) error {
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
	_, err := h.Bot.Send(msg)
	return err
}

func (h *Handler) sendMessageToUser(ctx context.Context, userID int64, text string, mode string) {
	// Выгружаем user
	u1, err1 := h.UserRepo.GetUserByID(ctx, userID)
	if err1 == nil && u1.ChatID != 0 {
		m1 := tgbotapi.NewMessage(u1.ChatID, text)
		m1.ParseMode = mode
		h.Bot.Send(m1)
	}
}

func (h *Handler) sendMessageToRoomOrUsers(ctx context.Context, room *models.Room, text string, mode string) {
	// Если group chatID задан, шлём туда
	if room.ChatID != nil {
		err := h.sendMessageToRoom(ctx, room, text, mode)
		if err != nil {
			utils.Logger.Error("sendMessageToRoom error:"+err.Error(), zap.Error(err))
		}
	} else {
		// Иначе шлём обоим
		h.sendMessageToUser(ctx, room.Player1ID, text, mode)
		h.sendMessageToUser(ctx, *room.Player2ID, text, mode)
	}
}

func (h *Handler) SendBoardToRoomOrUsers(ctx context.Context, r *models.Room) {
	var asciiBoard string
	var err error
	if r.ChatID != nil {
		// for chat(both)
		asciiBoard, err = game.RenderASCIIBoardHorizontal(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardWhite:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования горизонтальной доски"
		}
		h.sendMessageToRoomOrUsers(ctx, r, asciiBoard, tgbotapi.ModeMarkdownV2)
	} else {
		// for White
		asciiBoard, err = game.RenderASCIIBoardWhite(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardWhite:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования горизонтальной доски"
		}
		h.sendMessageToUser(ctx, *r.WhiteID, asciiBoard, tgbotapi.ModeMarkdownV2)
		// for Black
		asciiBoard, err = game.RenderASCIIBoardBlack(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardWhite:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования горизонтальной доски"
		}
		h.sendMessageToUser(ctx, *r.BlackID, asciiBoard, tgbotapi.ModeMarkdownV2)
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
