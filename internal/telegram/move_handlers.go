package telegram

import (
	"context"
	"fmt"
	"strings"

	"lvlchess/internal/db/models"
	"lvlchess/internal/game"
	"lvlchess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/notnil/chess"
	"go.uber.org/zap"
)

// prepareMoveButtons is called whenever it's a player's turn and we want to list all possible moves.
// 1) We parse the board state (FEN), 2) filter which squares can move, 3) create inline buttons for each square.
func (h *Handler) prepareMoveButtons(ctx context.Context, room *models.Room, userID int64) {
	if room.BoardState == "" {
		h.sendMessageToRoomOrUsers(ctx, room, "Нет текущего состояния доски!", tgbotapi.ModeHTML)
		return
	}

	fenOption, err := chess.FEN(room.BoardState)
	if err != nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Не получилось проанализировать доску!", tgbotapi.ModeHTML)
		return
	}
	chGame := chess.NewGame(fenOption)
	if chGame == nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Ошибка восстановления FEN.", tgbotapi.ModeHTML)
		return
	}

	// Determine if it's White or Black to move, and confirm userID matches them.
	sideToMove := chGame.Position().Turn() // White or Black
	var mustMoveUserID int64
	if sideToMove == chess.White /*&& room.WhiteID != nil*/ {
		mustMoveUserID = *room.WhiteID
	} else /*if sideToMove == chess.Black && room.BlackID != nil*/ {
		mustMoveUserID = *room.BlackID
	}

	if mustMoveUserID != userID {
		h.sendMessageToRoomOrUsers(ctx, room, "Сейчас не ваш ход!", tgbotapi.ModeHTML)
		return
	}

	// Gather all valid moves from the library, then group them by the from-square.
	validMoves := chGame.ValidMoves()
	movesBySquare := make(map[chess.Square][]chess.Move)
	for _, mv := range validMoves {
		movesBySquare[mv.S1()] = append(movesBySquare[mv.S1()], *mv)
	}

	board := chGame.Position().Board()
	figureSquares := make([]chess.Square, 0)
	figureIcon := make(map[string]string)

	// Filter squares belonging to the current side (White or Black).
	for sq, moves := range movesBySquare {
		if len(moves) == 0 {
			continue
		}
		piece := board.Piece(sq)
		if piece.Color() == sideToMove {
			figureSquares = append(figureSquares, sq)
			figureIcon[sq.String()] = game.PieceToStr(piece)
		}
	}

	if len(figureSquares) == 0 {
		h.sendMessageToRoomOrUsers(ctx, room, "Нет доступных ходов!", tgbotapi.ModeHTML)
		return
	}

	// Build an inline keyboard: each from-square is a button leading to "choose_figure:<square>"
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	keyboardSort(figureSquares)
	for i, sq := range figureSquares {
		sqStr := sq.String()
		callbackData := fmt.Sprintf("%s:%s&%s:%s", ActionChooseFigure, sqStr, RoomID, room.RoomID)
		buttonText := fmt.Sprintf("%s %s", figureIcon[sqStr], sqStr)
		btn := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		row = append(row, btn)

		// For spacing, let's do two columns per row:
		if (i+1)%2 == 0 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	SendInlineKeyboard(h.Bot, room, "Выберите фигуру для хода:", keyboard)
}

// handleChooseFigureCallback is invoked when user picks a from-square, e.g. "choose_figure:b8" in the callback data.
// We'll parse out which squares can be moved to from that square, then build a new inline keyboard
// listing all possible moves.
func (h *Handler) handleChooseFigureCallback(ctx context.Context, query *tgbotapi.CallbackQuery) {
	action, figureSquare, roomID, err := parseCallbackData(query.Data)
	utils.Logger.Error("handleChooseFigureCallback debugging",
		zap.Any("action", action),
		zap.Any("param", figureSquare),
		zap.Any(RoomID, roomID),
		zap.Any("err", err),
	)
	if err != nil || (action != ActionMove && action != ActionChooseFigure) {
		utils.Logger.Error("handleChooseFigureCallback parse error", zap.Error(err))
		return
	}

	room, err := h.RoomRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		callback := tgbotapi.NewCallback(query.ID, "Комната не найдена.")
		utils.Logger.Error("Room not found: "+err.Error(), zap.Error(err))
		if _, err := h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
		}
		return
	}
	if room.BoardState == "" {
		h.sendMessageToRoomOrUsers(ctx, room, "Нет состояния доски!", tgbotapi.ModeHTML)
		return
	}

	// Parse the board to see valid moves from figureSquare
	fenOption, err := chess.FEN(room.BoardState)
	if err != nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Не получилось проанализировать доску!", tgbotapi.ModeHTML)
		return
	}
	chGame := chess.NewGame(fenOption)
	if chGame == nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Ошибка восстановления FEN.", tgbotapi.ModeHTML)
		return
	}

	validMoves := chGame.ValidMoves()
	fromSq, errParseFrom := game.StrToSquare(figureSquare)
	if errParseFrom != nil {
		callback := tgbotapi.NewCallback(query.ID, "Некорректный квадрат фигуры.")
		utils.Logger.Error("Bad square parse: "+errParseFrom.Error(), zap.Error(errParseFrom))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
		}
		return
	}

	var movesForThisSquare []chess.Move
	for _, mv := range validMoves {
		if mv.S1().String() == figureSquare {
			movesForThisSquare = append(movesForThisSquare, *mv)
		}
	}

	if len(movesForThisSquare) == 0 {
		h.sendMessageToRoomOrUsers(ctx, room, "У этой фигуры нет допустимых ходов.", tgbotapi.ModeHTML)
		return
	}

	// Build inline buttons for each possible "move" (like "move:b8-c6").
	board := chGame.Position().Board()
	piece := board.Piece(fromSq)
	var rows [][]tgbotapi.InlineKeyboardButton
	row := []tgbotapi.InlineKeyboardButton{}

	for i, mv := range movesForThisSquare {
		callbackData := fmt.Sprintf("move:%s-%s&%s:%s", mv.S1().String(), mv.S2().String(), RoomID, roomID)
		btnText := fmt.Sprintf("%s ", buildMoveButtonText(piece, mv))
		// пример: "♔↷🛡♖\n e1->g1" (short castling),
		// или "🪄♙💨✨♕✨\n d7->d8Q" (pawn transformation),
		// или "♞⤵\n f5->h6" (normal move).
		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, callbackData)
		row = append(row, btn)

		// For neatness, let's do up to 4 in a row:
		if (i+1)%4 == 0 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	SendInlineKeyboard(h.Bot, room, fmt.Sprintf("Ходы для фигуры %s:", figureSquare), kb)

	// Clear the callback spinner
	callback := tgbotapi.NewCallback(query.ID, "Пожалуйста, выберите ход.")
	if _, err = h.Bot.Request(callback); err != nil {
		utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
	}
}

// parseCallbackData splits something like "move:b8-c6&roomID:xxxx" into (action="move", param="b8-c6", roomID="xxxx").
func parseCallbackData(data string) (action, param, roomID string, err error) {
	mainParts := strings.Split(data, "&")
	if len(mainParts) != 2 {
		return "", "", "", fmt.Errorf("incorrect callback data format (missing &)")
	}
	left, right := mainParts[0], mainParts[1]

	leftParts := strings.Split(left, CommandDelimiter)
	if len(leftParts) != 2 {
		return "", "", "", fmt.Errorf("incorrect left part")
	}
	action, param = leftParts[0], leftParts[1]

	rightParts := strings.Split(right, CommandDelimiter)
	if len(rightParts) != 2 {
		return "", "", "", fmt.Errorf("incorrect right part")
	}
	if rightParts[0] != RoomID {
		return "", "", "", fmt.Errorf("expected 'roomID:' got: %s", rightParts[0])
	}
	roomID = rightParts[1]
	return action, param, roomID, nil
}

// SendInlineKeyboard decides where to post a message with inline keyboard
// (group chat if room.ChatID is set, otherwise each player's private chat).
func SendInlineKeyboard(bot *tgbotapi.BotAPI, room *models.Room, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	utils.Logger.Error(
		"SendInlineKeyboard debug info",
		zap.Any("room.ChatID", room.ChatID),
		zap.Any("room.Player2ID", room.Player2ID),
		zap.Any("(room.Player2ID != nil):", (room.Player2ID != nil)),
		zap.Any("room.WhiteID", *room.WhiteID),
		zap.Any("room.BlackID", *room.BlackID),
		zap.Any("room.IsWhiteTurn", room.IsWhiteTurn),
	)

	if room.ChatID != nil {
		msg := tgbotapi.NewMessage(*room.ChatID, text)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	} else if room.Player2ID != nil {
		// If there's no group chat, we might only have private players.
		// We guess who to show the move interface based on whose turn it is.
		if room.IsWhiteTurn && room.WhiteID != nil {
			msgWhite := tgbotapi.NewMessage(*room.WhiteID, text)
			msgWhite.ReplyMarkup = keyboard
			msgWhite.ParseMode = tgbotapi.ModeMarkdownV2
			bot.Send(msgWhite)
		} else if !room.IsWhiteTurn && room.BlackID != nil {
			msgBlack := tgbotapi.NewMessage(*room.BlackID, text)
			msgBlack.ReplyMarkup = keyboard
			msgBlack.ParseMode = tgbotapi.ModeMarkdownV2
			bot.Send(msgBlack)
		}
	} else {
		// Possibly no second player => just send to Player1 as fallback
		msgP1 := tgbotapi.NewMessage(room.Player1ID, text)
		msgP1.ReplyMarkup = keyboard
		msgP1.ParseMode = tgbotapi.ModeMarkdownV2
		bot.Send(msgP1)
	}
}

// handleMoveCallback processes an actual move command like "move:b8-c6&roomID:123".
// We parse the squares, check if it's the user's turn, attempt the move, update board state,
// then broadcast the updated position or finishing message if game ended.
func (h *Handler) handleMoveCallback(ctx context.Context, query *tgbotapi.CallbackQuery) {
	action, moveStr, roomID, err := parseCallbackData(query.Data)
	if err != nil || action != ActionMove {
		return
	}

	figureParts := strings.Split(moveStr, "-")
	if len(figureParts) != 2 {
		callback := tgbotapi.NewCallback(query.ID, "Некорректный формат хода.")
		if _, err := h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
		}
		return
	}
	fromSquare, toSquare := figureParts[0], figureParts[1]

	room, err := h.RoomRepo.GetRoomByID(ctx, roomID)
	if err != nil || room == nil {
		callback := tgbotapi.NewCallback(query.ID, "Комната не найдена.")
		utils.Logger.Error("Room not found: "+err.Error(), zap.Error(err))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
		}
		return
	}
	if room.BoardState == "" {
		h.sendMessageToRoomOrUsers(ctx, room, "Нет текущего состояния доски!", tgbotapi.ModeHTML)
		return
	}

	fenOption, err := chess.FEN(room.BoardState)
	if err != nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Не получилось проанализировать доску!", tgbotapi.ModeHTML)
		return
	}
	chGame := chess.NewGame(fenOption)
	if chGame == nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Ошибка восстановления FEN!", tgbotapi.ModeHTML)
		return
	}

	// Check if user is indeed the correct side to move.
	userID := query.From.ID
	sideToMove := chGame.Position().Turn()
	var mustMoveUserID int64
	if sideToMove == chess.White /* && room.WhiteID != nil*/ {
		mustMoveUserID = *room.WhiteID
	} else /* if sideToMove == chess.Black && room.BlackID != nil*/ {
		mustMoveUserID = *room.BlackID
	}
	if mustMoveUserID != userID {
		callback := tgbotapi.NewCallback(query.ID, "Сейчас не ваш ход!")
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
		}
		return
	}

	// Attempt to decode the move "b8c6" as UCINotation or fallback.
	mv, parseErr := chess.UCINotation{}.Decode(chGame.Position(), fromSquare+toSquare)
	if parseErr != nil {
		callback := tgbotapi.NewCallback(query.ID,
			fmt.Sprintf("Невозможно распарсить ход %s-%s: %v", fromSquare, toSquare, parseErr))
		utils.Logger.Error("Parse move error: "+parseErr.Error(), zap.Error(parseErr))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
		}
		return
	}

	// Try performing the move in the notnil/chess library.
	if errMove := chGame.Move(mv); errMove != nil {
		// If move is illegal, send an error.
		h.sendMessageToRoomOrUsers(ctx, room, "Невозможный ход!", tgbotapi.ModeHTML)
		callback := tgbotapi.NewCallback(query.ID, "")
		utils.Logger.Error("Illegal move: "+errMove.Error(), zap.Error(errMove))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
		}
		return
	}

	// If successful, store the new FEN.
	newFEN := chGame.FEN()
	room.BoardState = newFEN
	room.IsWhiteTurn = !room.IsWhiteTurn
	if err = h.RoomRepo.UpdateRoom(ctx, room); err != nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Ошибка при сохранении нового состояния доски!", tgbotapi.ModeHTML)
		callback := tgbotapi.NewCallback(query.ID, "")
		utils.Logger.Error("UpdateRoom error: "+err.Error(), zap.Error(err))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
		}
		return
	}

	// Check for game completion (checkmate, draw, etc.)
	outcome := chGame.Outcome()
	if outcome != chess.NoOutcome {
		switch outcome {
		case chess.WhiteWon:
			h.sendMessageToRoomOrUsers(ctx, room, "Игра завершена! Победили белые.", tgbotapi.ModeHTML)
		case chess.BlackWon:
			h.sendMessageToRoomOrUsers(ctx, room, "Игра завершена! Победили чёрные.", tgbotapi.ModeHTML)
		case chess.Draw:
			h.sendMessageToRoomOrUsers(ctx, room, "Игра завершена! Ничья.", tgbotapi.ModeHTML)
		}
		// Possibly mark room.Status="finished" here or do other final logic.
		callback := tgbotapi.NewCallback(query.ID, "Ход сделан! Игра окончена.")
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
			return
		}
		// Return because the game is done.
		return
	}

	// If the game continues, we announce the move to chat or private messages.
	moveMsg := "Ход сделан:\n"
	if mv.HasTag(chess.Capture) {
		moveMsg = fmt.Sprintf("```\n%s\n```", buildMoveButtonText(chGame.Position().Board().Piece(mv.S1()), *mv))
	} else {
		moveMsg = fmt.Sprintf("```%s-%s```", fromSquare, toSquare)
	}
	h.sendMessageToRoomOrUsers(ctx, room, moveMsg, tgbotapi.ModeMarkdownV2)

	// Send the updated ASCII board to relevant place(s).
	h.SendBoardToRoomOrUsers(ctx, room)

	// Then prepare next player's move.
	nextTurn := chGame.Position().Turn()
	var nextUserID int64
	if nextTurn == chess.White /* && room.WhiteID != nil*/ {
		nextUserID = *room.WhiteID
	} else /* if nextTurn == chess.Black && room.BlackID != nil*/ {
		nextUserID = *room.BlackID
	}
	h.prepareMoveButtons(ctx, room, nextUserID)

	// Confirm callback with "Move successful!"
	callback := tgbotapi.NewCallback(query.ID, "Ход успешен!")
	if _, err = h.Bot.Request(callback); err != nil {
		utils.Logger.Error("AnswerCallbackQuery error: "+err.Error(), zap.Error(err))
	}
}

// buildMoveButtonText returns a fancy Unicode string describing the move (e.g. castling, capture, promotion).
// It's purely for user-facing text on the inline buttons.
func buildMoveButtonText(p chess.Piece, mv chess.Move) string {
	// Check castling
	if mv.HasTag(chess.KingSideCastle) {
		return "♔↷🛡⟵♖\n " + mv.String()
	}
	if mv.HasTag(chess.QueenSideCastle) {
		return "♖⟶🛡↶♔\n " + mv.String()
	}
	// Check promotion(pawn transformation)
	if mv.Promo() != chess.NoPieceType {
		return fmt.Sprintf("🪄%s💨✨%s✨\n %s", p.String(),
			game.PieceToStr(chess.Piece(mv.Promo())), mv.String())
	}
	// Normal or capture
	text := ""
	if mv.HasTag(chess.Capture) {
		// Example: "(⤴)♘⚔️♜ (g6->h8)" or "♙⚔♞ (b2->b3)"
		text = fmt.Sprintf("%s⚔️ ", p.String())
	} else {
		// Если просто ход: "♙⬆️", "♞⤵", "♗↖️" etc.
		text = p.String()
	}
	fromSq, toSq := mv.S1(), mv.S2()
	color := p.Color() == chess.White
	arrow := game.ArrowForMove(fromSq, toSq, color) // Example: "↙️"
	text += fmt.Sprintf("%s (%s->%s)", arrow, fromSq.String(), toSq.String())

	return text
}
