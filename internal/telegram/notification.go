package telegram

import (
	"context"
	"fmt"
	"sort"

	"lvlchess/internal/db/models"
	"lvlchess/internal/game"
	"lvlchess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/notnil/chess"
	"go.uber.org/zap"
)

// MakeFinalTitle constructs a user-facing room title based on Player1 and Player2 usernames.
// Used to rename a group chat, for instance, to "tChess:@player1_⚔️_@player2".
func (h *Handler) MakeFinalTitle(ctx context.Context, r *models.Room) (title string) {
	// Default fallback title
	title = "tChess:????"

	if r != nil && r.Player1ID != 0 {
		// Look up the first player's user record for a username
		p1, err := h.UserRepo.GetUserByID(ctx, r.Player1ID)
		if err != nil {
			return
		}
		title = fmt.Sprintf("tChess:@%s_⚔️_??", p1.Username)

		// If we have a second player, embed them in the title
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

// tryRenameGroup attempts to change the group chat title to newTitle, using Telegram's SetChatTitle API call.
// If the bot lacks "Change group info" permission, we catch an error, log it, and propose a "Retry" button.
func (h *Handler) tryRenameGroup(bot *tgbotapi.BotAPI, chatID int64, newTitle string) {
	renameConfig := tgbotapi.SetChatTitleConfig{
		ChatID: chatID,
		Title:  newTitle,
	}

	_, err := bot.Request(renameConfig)
	if err != nil {
		utils.Logger.Error(
			fmt.Sprintf("Failed to rename group (chatID=%d): %v", chatID, err),
			zap.Error(err),
		)

		// Provide a button to retry after the user grants permissions
		retryBtn := tgbotapi.NewInlineKeyboardButtonData(
			"Повторить переименование",
			fmt.Sprintf("%s:%s", RetryRename, newTitle),
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

// notifyGameStarted is used once a room has two players and we want to announce "the game has started."
// It also sends the ASCII board and prompts the first player for their move.
func (h *Handler) notifyGameStarted(ctx context.Context, room *models.Room) {
	introMsg := "Игра началась!\n" + room.RoomTitle

	// 1) Announce the start in group or private chats
	h.sendMessageToRoomOrUsers(ctx, room, introMsg, tgbotapi.ModeHTML)

	// 2) Show the current board (ASCII-based)
	h.SendBoardToRoomOrUsers(ctx, room)

	// 3) We call prepareMoveButtons for the White side (since typically White starts).
	h.prepareMoveButtons(ctx, room, *room.WhiteID)
}

// sendMessageToRoom tries to post the message directly to the group's chatID.
// If for some reason chatID is nil, it returns an error.
func (h *Handler) sendMessageToRoom(ctx context.Context, room *models.Room, text string, mode string) error {
	if room.ChatID == nil {
		return fmt.Errorf("room.ChatID is nil, cannot send to a group")
	}

	msg := tgbotapi.NewMessage(*room.ChatID, text)

	// By default, we might parse as HTML or Markdown for styling.
	if mode != "" && (mode == tgbotapi.ModeMarkdownV2 || mode == tgbotapi.ModeHTML) {
		msg.ParseMode = mode
	} else {
		msg.ParseMode = tgbotapi.ModeMarkdown
	}

	_, err := h.Bot.Send(msg)
	return err
}

// sendMessageToUser sends a private message to a known user, if we have chatID in the DB.
// Usually, user.ChatID corresponds to private chat with the bot.
func (h *Handler) sendMessageToUser(ctx context.Context, userID int64, text string, mode string) {
	u1, err1 := h.UserRepo.GetUserByID(ctx, userID)
	if err1 == nil && u1.ChatID != 0 {
		m1 := tgbotapi.NewMessage(u1.ChatID, text)
		m1.ParseMode = mode
		h.Bot.Send(m1)
	}
}

// sendMessageToRoomOrUsers is a convenience method that checks if room.ChatID is set (meaning group chat)
// or not. If not, we push the message to both Player1 and Player2 in private chat (if Player2ID is set).
func (h *Handler) sendMessageToRoomOrUsers(ctx context.Context, room *models.Room, text string, mode string) {
	if room.ChatID != nil {
		err := h.sendMessageToRoom(ctx, room, text, mode)
		if err != nil {
			utils.Logger.Error("sendMessageToRoom error:"+err.Error(), zap.Error(err))
		}
	} else {
		// Fallback to private messages: Player1, and if present, Player2
		h.sendMessageToUser(ctx, room.Player1ID, text, mode)
		if room.Player2ID != nil {
			h.sendMessageToUser(ctx, *room.Player2ID, text, mode)
		}
	}
}

// SendBoardToRoomOrUsers dispatches an ASCII board representation. The orientation depends on whether
// a group is used (horizontal) or a private scenario (white sees "normal" board, black sees "flipped").
func (h *Handler) SendBoardToRoomOrUsers(ctx context.Context, r *models.Room) {
	var asciiBoard string
	var err error

	if r.ChatID != nil {
		// If a group chat is linked, we typically show "horizontal" style
		asciiBoard, err = game.RenderASCIIBoardHorizontal(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardHorizontal:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования горизонтальной доски"
		}
		h.sendMessageToRoomOrUsers(ctx, r, asciiBoard, tgbotapi.ModeMarkdownV2)
	} else {
		// In private games, show White's perspective to White, Black's perspective to Black
		asciiBoard, err = game.RenderASCIIBoardWhite(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardWhite:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования доски (white)."
		}
		if r.WhiteID != nil { // !!!
			h.sendMessageToUser(ctx, *r.WhiteID, asciiBoard, tgbotapi.ModeMarkdownV2)
		}

		asciiBoard, err = game.RenderASCIIBoardBlack(r.BoardState)
		if err != nil {
			utils.Logger.Error("game.RenderASCIIBoardBlack:"+err.Error(), zap.Error(err))
			asciiBoard = "Ошибка формирования доски (black)."
		}
		if r.BlackID != nil { // !!!
			h.sendMessageToUser(ctx, *r.BlackID, asciiBoard, tgbotapi.ModeMarkdownV2)
		}
	}
}

// keyboardSort is a helper that sorts squares in descending rank (8..1) and ascending file (a..h).
// It's used so that any list of squares is displayed in a predictable order.
func keyboardSort(slice []chess.Square) {
	sort.Slice(slice, func(i, j int) bool {
		rankI, fileI := slice[i].Rank(), slice[i].File()
		rankJ, fileJ := slice[j].Rank(), slice[j].File()

		// Compare ranks (descending)
		if rankI != rankJ {
			return rankI > rankJ
		}
		// If ranks are equal, compare files (ascending).
		return fileI < fileJ
	})
}
