package game

import (
	"strings"

	"github.com/katalvlaran/telega-shess/internal/utils"
	"github.com/notnil/chess"
)

// BoardRenderer отвечает за визуализацию шахматной доски
type BoardRenderer struct {
	lastMove *chess.Move // Для подсветки последнего хода
}

// NewBoardRenderer создает новый рендерер доски
func NewBoardRenderer() *BoardRenderer {
	return &BoardRenderer{}
}

// RenderBoard возвращает строковое представление доски
func (br *BoardRenderer) RenderBoard(position *chess.Position, lastMove *chess.Move) string {
	var sb strings.Builder

	// Заголовок доски
	sb.WriteString("   a b c d e f g h\n")

	// Отрисовка каждой строки
	for rank := 7; rank >= 0; rank-- {
		// Номер строки
		sb.WriteString(string('1' + rune(rank)))
		sb.WriteString("  ")

		// Отрисовка клеток
		for file := 0; file < 8; file++ {
			square := chess.Square(rank*8 + file)
			piece := position.Board().Piece(square)

			// Получение символа фигуры
			symbol := br.getPieceSymbol(piece)

			// Добавление пробела после символа для выравнивания
			sb.WriteString(symbol + " ")
		}

		// Номер строки в конце
		sb.WriteString(" ")
		sb.WriteString(string('1' + rune(rank)))
		sb.WriteString("\n")
	}

	// Нижняя часть доски
	sb.WriteString("   a b c d e f g h")

	return sb.String()
}

// getPieceSymbol возвращает Unicode символ для фигуры
func (br *BoardRenderer) getPieceSymbol(piece chess.Piece) string {
	if piece == chess.NoPiece {
		return utils.EmptySquare
	}

	symbols := map[chess.Piece]string{
		chess.WhiteKing:   utils.KingWhite,
		chess.WhiteQueen:  utils.QueenWhite,
		chess.WhiteRook:   utils.RookWhite,
		chess.WhiteBishop: utils.BishopWhite,
		chess.WhiteKnight: utils.KnightWhite,
		chess.WhitePawn:   utils.PawnWhite,
		chess.BlackKing:   utils.KingBlack,
		chess.BlackQueen:  utils.QueenBlack,
		chess.BlackRook:   utils.RookBlack,
		chess.BlackBishop: utils.BishopBlack,
		chess.BlackKnight: utils.KnightBlack,
		chess.BlackPawn:   utils.PawnBlack,
	}

	return symbols[piece]
}
