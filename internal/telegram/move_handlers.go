package telegram

import (
	"context"
	"fmt"
	"strings"

	"telega_chess/internal/db/models"
	"telega_chess/internal/game"
	"telega_chess/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/notnil/chess"
	"go.uber.org/zap"
)

// prepareMoveButtons формирует список фигур, которые могут ходить,
// и отправляет кнопки "choose_figure:<square>" для тех фигур, у которых есть хотя бы 1 valid move.
//
// 1. Загружаем FEN из room.BoardState и создаём объект игры.
// 2. Определяем, чей ход — White или Black.
// 3. Проверяем, совпадает ли userID c room.WhiteID/room.BlackID (в зависимости от sideToMove).
// 4. Проходим по всем ValidMoves, группируем их по "from"-square.
// 5. Для каждой фигуры "from" (которая имеет ходы) делаем одну inline-кнопку "choose_figure:fromSquare".
// 6. Отправляем результат через sendMessageToRoomOrUsers или SendMessageToRoom.
func (h *Handler) prepareMoveButtons(ctx context.Context, room *models.Room, userID int64) {
	// 0. Проверим, есть ли вообще boardState
	if room.BoardState == "" {
		h.sendMessageToRoomOrUsers(ctx, room, "Нет текущего состояния доски!", tgbotapi.ModeHTML)
		return
	}

	// 1. Загружаем FEN в notnil/chess
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

	// 2. Определяем, кто ходит
	sideToMove := chGame.Position().Turn() // chess.White или chess.Black
	var mustMoveUserID int64
	if sideToMove == chess.White {
		mustMoveUserID = *room.WhiteID // условно, у вас может быть room.WhiteID (int64)
	} else {
		mustMoveUserID = *room.BlackID
	}

	// Сверим, тот ли это userID
	if mustMoveUserID != userID {
		h.sendMessageToRoomOrUsers(ctx, room, "Сейчас не ваш ход!", tgbotapi.ModeHTML)
		return
	}

	// 3. Составляем карту "fromSquare => []Move"
	validMoves := chGame.ValidMoves()
	movesBySquare := make(map[chess.Square][]chess.Move)
	for _, mv := range validMoves {
		movesBySquare[mv.S1()] = append(movesBySquare[mv.S1()], *mv)
	}

	// 4. Фильтруем "fromSquare" так, чтобы фигура принадлежала текущему sideToMove
	// (notnil/chess обычно корректно даёт validMoves только для текущей стороны,
	//  но мы можем дополнительно убедиться, что piece.Color() == sideToMove)
	board := chGame.Position().Board()

	// Собираем все squares, у которых есть хотя бы 1 move
	figureSquares := make([]chess.Square, 0)
	figureIcone := make(map[string]string)
	for sq, mvs := range movesBySquare {
		if len(mvs) == 0 {
			continue
		}
		// Проверка: фигура на sq действительно sideToMove?
		piece := board.Piece(sq)
		if piece.Color() == sideToMove {
			figureSquares = append(figureSquares, sq)
			figureIcone[sq.String()] = game.PieceToStr(piece)
		}
	}

	if len(figureSquares) == 0 {
		h.sendMessageToRoomOrUsers(ctx, room, "Нет доступных ходов!", tgbotapi.ModeHTML)
		return
	}

	// 5. Формируем InlineKeyboard
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	keyboardSort(figureSquares)
	for i, sq := range figureSquares {
		// Пример: "choose_figure:b8"
		sqStr := sq.String()
		callbackData := fmt.Sprintf("choose_figure:%s&roomID:%s", sqStr, room.RoomID)
		buttonText := fmt.Sprintf("%s %s", figureIcone[sqStr], sqStr)
		btn := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)

		row = append(row, btn)
		// Пусть по 2 кнопки в ряд (позже выстроим по sq.Rank )
		// TODO to const + dynamyc
		if (i+1)%2 == 0 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	// 6. Отправляем сообщение
	SendInlineKeyboard(h.Bot, room, "Выберите фигуру для хода:", keyboard)
}

// handleChooseFigureCallback обрабатывает callback вроде "choose_figure:b8".
// 1. Парсим b8.
// 2. Загружаем room, boardState.
// 3. Выбираем valid moves, где from == b8.
// 4. Создаём Inline-кнопки вида "move:b8-c6", "move:b8-a6" и отправляем.
func (h *Handler) handleChooseFigureCallback(ctx context.Context, query *tgbotapi.CallbackQuery) {
	action, figureSquare, roomID, err := parseCallbackData(query.Data)
	utils.Logger.Error("😖 handleChooseFigureCallback  👾",
		zap.Any("action", action),
		zap.Any("param", figureSquare),
		zap.Any("roomID", roomID),
		zap.Any("err", err),
		//zap.Any("err.Error()", err.Error()),
	)
	if err != nil || (action != ActionMove && action != ActionChooseFigure) {
		// error handling
		utils.Logger.Error("😖 handleChooseFigureCallback  👾", zap.Error(err))
		return
	}

	room, err := h.RoomRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		callback := tgbotapi.NewCallback(query.ID, "Комната не найдена.")
		utils.Logger.Error("😖 Комната не найдена 👾"+err.Error(), zap.Error(err))
		if _, err := h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
		}

		return
	}

	if room.BoardState == "" {
		h.sendMessageToRoomOrUsers(ctx, room, "Нет состояния доски!", tgbotapi.ModeHTML)
		return
	}

	// Распарсим FEN через notnil/chess
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
	// 2. Фильтруем validMoves, где from == figureSquare
	validMoves := chGame.ValidMoves()
	var movesForThisSquare []chess.Move
	fromSq, errParseFrom := game.StrToSquare(figureSquare)
	if errParseFrom != nil {
		callback := tgbotapi.NewCallback(query.ID, "Некорректный квадрат фигуры.")
		utils.Logger.Error("😖 Некорректный квадрат фигуры. 👾"+errParseFrom.Error(), zap.Error(errParseFrom))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
		}
		return
	}
	for _, mv := range validMoves {
		if mv.S1().String() == figureSquare {
			movesForThisSquare = append(movesForThisSquare, *mv)
		}
	}

	if len(movesForThisSquare) == 0 {
		h.sendMessageToRoomOrUsers(ctx, room, "У этой фигуры нет допустимых ходов.", tgbotapi.ModeHTML)
		return
	}

	// 3. Формируем кнопки "move:b8-c6" etc.
	var rows [][]tgbotapi.InlineKeyboardButton
	row := []tgbotapi.InlineKeyboardButton{}

	// Определим саму фигуру (для эмодзи)
	board := chGame.Position().Board()
	piece := board.Piece(fromSq)

	for i, mv := range movesForThisSquare {
		callbackData := fmt.Sprintf("move:%s-%s&roomID:%s", mv.S1().String(), mv.S2().String(), roomID)

		// --- формируем "текст" кнопки c эмодзи ---
		btnText := fmt.Sprintf("%s ", buildMoveButtonText(piece, mv))
		// пример: "♔↷🛡♖\n e1->g1" (рокировка короткая),
		// или "🪄♙💨✨♕✨\n d7->d8Q" (превращение),
		// или "♞⤵\n f5->h6" (обычный ход).

		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, callbackData)
		row = append(row, btn)

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

	// Подтверждаем callback, чтобы Telegram не показывал "крутилку"
	callback := tgbotapi.NewCallback(query.ID, "Пожалуйста, выберите ход.")
	if _, err = h.Bot.Request(callback); err != nil {
		utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
	}
}

// parseCallbackData разбирает data вида "move:b8-c6&roomID:xxxx-..."
// возвращает (action="move", param="b8-c6", roomID="xxxx-...") или ошибку.
func parseCallbackData(data string) (action, param, roomID string, err error) {
	mainParts := strings.Split(data, "&")
	if len(mainParts) != 2 {
		return "", "", "", fmt.Errorf("incorrect callback data format (no &)")
	}
	left, right := mainParts[0], mainParts[1]
	// left = "move:b8-c6", right="roomID:xxxx..."

	leftParts := strings.Split(left, ":")
	if len(leftParts) != 2 {
		return "", "", "", fmt.Errorf("incorrect left part format")
	}
	action, param = leftParts[0], leftParts[1]

	rightParts := strings.Split(right, ":")
	if len(rightParts) != 2 {
		return "", "", "", fmt.Errorf("incorrect right part format")
	}
	if rightParts[0] != "roomID" {
		return "", "", "", fmt.Errorf("expected 'roomID:', got %s", rightParts[0])
	}
	roomID = rightParts[1]

	return action, param, roomID, nil
}

// Отправляет сообщение с inline-клавиатурой в личку ИЛИ в группу, в зависимости от room.ChatID
func SendInlineKeyboard(bot *tgbotapi.BotAPI, room *models.Room, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	// Предположим, у вас есть логика:
	//  - если room.ChatID != nil => шлём туда
	//  - иначе шлём обоим игрокам (или только userID?), смотря как вы устроили проект

	//sendMessageToRoomOrUsers(bot, room, text, modeKeyboard)
	utils.Logger.Error(
		"😖 SendInlineKeyboard  👾",
		zap.Any("room.ChatID", room.ChatID),
		zap.Any("room.Player2ID", room.Player2ID),
		zap.Any("(room.Player2ID != nil):", (room.Player2ID != nil)),
		zap.Any("room.room.WhiteID", *room.WhiteID),
		zap.Any("room.BlackID", *room.BlackID),
		zap.Any("room.IsWhiteTurn", room.IsWhiteTurn))
	if room.ChatID != nil {
		msg := tgbotapi.NewMessage(*room.ChatID, text)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	} else if room.Player2ID != nil {
		// Личная игра => шлём пользователю, который должен ходить?
		// Или сразу обоим? up to you.
		// Для иллюстрации отправим WhiteID:
		if room.IsWhiteTurn {
			msgWhite := tgbotapi.NewMessage(*room.WhiteID, text)
			msgWhite.ReplyMarkup = keyboard
			msgWhite.ParseMode = tgbotapi.ModeMarkdownV2
			bot.Send(msgWhite)
		} else {
			msgBlack := tgbotapi.NewMessage(*room.BlackID, text)
			msgBlack.ReplyMarkup = keyboard
			msgBlack.ParseMode = tgbotapi.ModeMarkdownV2
			bot.Send(msgBlack)
		}
	} else { // send to Player1
		msgP1 := tgbotapi.NewMessage(room.Player1ID, text)
		msgP1.ReplyMarkup = keyboard
		msgP1.ParseMode = tgbotapi.ModeMarkdownV2
		bot.Send(msgP1)
	}
}

func (h *Handler) handleMoveCallback(ctx context.Context, query *tgbotapi.CallbackQuery) {
	action, moveStr, roomID, err := parseCallbackData(query.Data)
	if err != nil || action != "move" {
		// error handling
		return
	}
	// Распарсим "b8-c6" в from->to
	figureParts := strings.Split(moveStr, "-")
	if len(figureParts) != 2 {
		callback := tgbotapi.NewCallback(query.ID, "Некорректный формат хода.")
		if _, err := h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
		}

		return
	}
	fromSquare, toSquare := figureParts[0], figureParts[1] // "b8", "c6"
	room, err := h.RoomRepo.GetRoomByID(ctx, roomID)
	if err != nil || room == nil {
		callback := tgbotapi.NewCallback(query.ID, "Комната не найдена.")
		utils.Logger.Error("😖 Комната не найдена 👾"+err.Error(), zap.Error(err))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
		}

		return
	}

	if room.BoardState == "" {
		h.sendMessageToRoomOrUsers(ctx, room, "Нет текущего состояния доски!", tgbotapi.ModeHTML)
		return
	}

	// Распарсим FEN через notnil/chess
	fenOption, err := chess.FEN(room.BoardState)
	if err != nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Не получилось проанализировать доску!", tgbotapi.ModeHTML)
		return
	}
	// --- Загружаем FEN в notnil/chess ---
	chGame := chess.NewGame(fenOption)
	if chGame == nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Ошибка восстановления FEN!", tgbotapi.ModeHTML)
		return
	}

	// --- Проверим, что ход действительно этого пользователя ---
	// (Аналогично, как в prepareMoveButtons).
	// Получим userID из query, если нужно.
	userID := query.From.ID
	sideToMove := chGame.Position().Turn() // White/Black
	var mustMoveUserID int64
	if sideToMove == chess.White {
		mustMoveUserID = *room.WhiteID
	} else {
		mustMoveUserID = *room.BlackID
	}
	if mustMoveUserID != userID {
		callback := tgbotapi.NewCallback(query.ID, "Сейчас не ваш ход!")
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
		}
		return
	}

	// --- Формируем объект "move" (b8 -> c6) для notnil/chess ---
	// notnil/chess позволяет делать AlgebraicNotation или SAN. Мы можем сделать:
	mv, errParse := chess.UCINotation{}.Decode(chGame.Position(), fromSquare+toSquare)
	if errParse != nil {
		// Либо fallback:
		// mv, errParse = chess.AlgebraicNotation{}.Decode(...)
		callback := tgbotapi.NewCallback(
			query.ID,
			fmt.Sprintf("Невозможно распарсить ход %s-%s: %v", fromSquare, toSquare, errParse))
		utils.Logger.Error("😖 Комната не найдена 👾"+errParse.Error(), zap.Error(errParse))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
		}
		return
	}

	// --- Пытаемся сделать ход ---
	errMove := chGame.Move(mv)
	if errMove != nil {
		// Ход невозможен
		h.sendMessageToRoomOrUsers(ctx, room, "Невозможный ход!", tgbotapi.ModeHTML)
		callback := tgbotapi.NewCallback(query.ID, "")
		utils.Logger.Error("😖 Комната не найдена 👾"+errMove.Error(), zap.Error(errMove))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
		}

		return
	}

	// --- Ход успешно совершен ---
	// Обновим room.BoardState -> текущее FEN
	newFEN := chGame.FEN()
	room.BoardState = newFEN
	room.IsWhiteTurn = !room.IsWhiteTurn
	if err = h.RoomRepo.UpdateRoom(ctx, room); err != nil {
		h.sendMessageToRoomOrUsers(ctx, room, "Ошибка при сохранении нового состояния доски!", tgbotapi.ModeHTML)
		callback := tgbotapi.NewCallback(query.ID, "")
		utils.Logger.Error("😖 Ошибка при сохранении нового состояния доски 👾"+err.Error(), zap.Error(err))
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
		}

		return
	}

	// --- Проверим, не закончилась ли партия (мат, пат, ничья) ---
	outcome := chGame.Outcome()
	if outcome != chess.NoOutcome {
		// Игра завершена
		switch outcome {
		case chess.WhiteWon:
			h.sendMessageToRoomOrUsers(ctx, room, "Игра завершена! Победили белые.", tgbotapi.ModeHTML)
		case chess.BlackWon:
			h.sendMessageToRoomOrUsers(ctx, room, "Игра завершена! Победили чёрные.", tgbotapi.ModeHTML)
		case chess.Draw:
			h.sendMessageToRoomOrUsers(ctx, room, "Игра завершена! Ничья.", tgbotapi.ModeHTML)
		}
		// Можно поставить room.Status="finished", убрать кнопки и т.д.
		callback := tgbotapi.NewCallback(query.ID, "Ход сделан! Игра окончена.")
		if _, err = h.Bot.Request(callback); err != nil {
			utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
			return
		}
	}

	// --- Иначе игра продолжается ---
	// 1) Сообщаем: "Ход сделан: b8-c6"
	moveMsg := "Ход сделан:\n"
	if mv.HasTag(chess.Capture) {
		moveMsg = fmt.Sprintf("```\n%s\n```", buildMoveButtonText(chGame.Position().Board().Piece(mv.S1()), *mv))
	} else {
		moveMsg = fmt.Sprintf("```%s-%s```", fromSquare, toSquare)
	}
	h.sendMessageToRoomOrUsers(ctx, room, moveMsg, tgbotapi.ModeMarkdownV2)

	// 2) Отправляем обновленную доску
	h.SendBoardToRoomOrUsers(ctx, room)

	// 3) Передаем ход второму игроку:
	nextTurn := chGame.Position().Turn() // White/Black
	var nextUserID int64
	if nextTurn == chess.White {
		nextUserID = *room.WhiteID
	} else {
		nextUserID = *room.BlackID
	}

	// 4) Подготовим кнопки для следующего игрока
	h.prepareMoveButtons(ctx, room, nextUserID)

	// --- Подтверждаем callback без alert
	callback := tgbotapi.NewCallback(query.ID, "Ход успешен!")
	if _, err = h.Bot.Request(callback); err != nil {
		utils.Logger.Error("😖 AnswerCallbackQuery error 👾"+err.Error(), zap.Error(err))
	}
}

// buildMoveButtonText - формирует красивый текст кнопки,
// учитывая рокировку, превращение, взятие и т.д.
func buildMoveButtonText(p chess.Piece, mv chess.Move) string {
	// shortCastle / longCastle?
	if mv.HasTag(chess.KingSideCastle) {
		// Белый: e1-g1, Чёрный: e8-g8
		return "♔↷🛡⟵♖\n " + mv.String() // условно: "♔↷🛡♖\n e1g1"
	}
	if mv.HasTag(chess.QueenSideCastle) {
		return "♖⟶🛡↶♔\n " + mv.String() // условно "♖⟶🛡↶♔\n e1c1"
	}

	// Превращение? (Promotion)
	if mv.Promo() != chess.NoPieceType {
		// обычно last char = Q/R/N/B
		mv.Promo()
		//promoChar := string(mv.String()[len(mv.String())-1]) // "Q","R","N","B"
		return fmt.Sprintf("🪄%s💨✨%s✨\n %s", p.String(), game.PieceToStr(chess.Piece(mv.Promo())), mv.String())
	}

	// Обычный ход. Проверим, есть ли захват:
	text := ""
	if mv.HasTag(chess.Capture) {
		// Пример: "(⤴)♘⚔️♜ (g6->h8)"
		// Либо "♙⚔♞ (b2->b3)"
		text = fmt.Sprintf("%s⚔️ ", p.String())
	} else {
		// Если просто ход: "♙⬆️", "♞⤵", "♗↖️" etc.
		text = p.String()
	}

	fromSq, toSq := mv.S1(), mv.S2()
	color := p.Color() == chess.White
	arrow := game.ArrowForMove(fromSq, toSq, color) // напр. "↙️"
	text += fmt.Sprintf("%s (%s->%s)", arrow, fromSq.String(), toSq.String())

	return text
}
