package game

import (
	"fmt"
	"strings"

	"github.com/katalvlaran/telega-shess/internal/utils"
	"github.com/notnil/chess"
)

// BoardTheme определяет цвета и стили доски
type BoardTheme struct {
	LastMoveHighlight string
	CheckHighlight    string
	Coordinates       bool
}

// DefaultTheme возвращает тему по умолчанию
func DefaultTheme() BoardTheme {
	return BoardTheme{
		LastMoveHighlight: "🔹",
		CheckHighlight:    "⚠️",
		Coordinates:       true,
	}
}

// EnhancedBoardRenderer улучшенный рендерер доски
type EnhancedBoardRenderer struct {
	theme    BoardTheme
	lastMove *chess.Move
}

// NewEnhancedBoardRenderer создает новый улучшенный рендерер
func NewEnhancedBoardRenderer(theme BoardTheme) *EnhancedBoardRenderer {
	return &EnhancedBoardRenderer{
		theme: theme,
	}
}

// RenderEnhancedBoard возвращает улучшенное представление доски
func (br *EnhancedBoardRenderer) RenderEnhancedBoard(game *chess.Game) string {
	var sb strings.Builder
	pos := game.Position()
	lastMove := game.LastMove()

	// Добавляем информацию о текущем ходе
	if pos.Turn() == chess.White {
		sb.WriteString("Ход белых\n")
	} else {
		sb.WriteString("Ход черных\n")
	}

	// Добавляем предупреждение о шахе
	if pos.InCheck() {
		sb.WriteString(br.theme.CheckHighlight + " Шах! " + br.theme.CheckHighlight + "\n")
	}

	// Рисуем доску
	sb.WriteString(br.renderBoard(pos, lastMove))

	// Добавляем историю ходов
	if len(game.Moves()) > 0 {
		sb.WriteString("\nПоследние ходы:\n")
		moves := game.Moves()
		start := len(moves)
		if start > 3 {
			start = len(moves) - 3
		}
		for i := start; i < len(moves); i++ {
			moveNum := i/2 + 1
			if i%2 == 0 {
				sb.WriteString(fmt.Sprintf("%d. %s", moveNum, moves[i]))
			} else {
				sb.WriteString(fmt.Sprintf(" %s\n", moves[i]))
			}
		}
	}

	return sb.String()
}

// renderBoard отрисовывает доску с подсветкой
func (br *EnhancedBoardRenderer) renderBoard(pos *chess.Position, lastMove *chess.Move) string {
	var sb strings.Builder

	if br.theme.Coordinates {
		sb.WriteString("   a b c d e f g h\n")
	}

	for rank := 7; rank >= 0; rank-- {
		if br.theme.Coordinates {
			sb.WriteString(fmt.Sprintf("%d  ", rank+1))
		}

		for file := 0; file < 8; file++ {
			square := chess.Square(rank*8 + file)
			piece := pos.Board().Piece(square)

			// Подсветка последнего хода
			if lastMove != nil && (square == lastMove.S1() || square == lastMove.S2()) {
				sb.WriteString(br.theme.LastMoveHighlight)
			}

			symbol := br.getPieceSymbol(piece)
			sb.WriteString(symbol + " ")
		}

		if br.theme.Coordinates {
			sb.WriteString(fmt.Sprintf(" %d", rank+1))
		}
		sb.WriteString("\n")
	}

	if br.theme.Coordinates {
		sb.WriteString("   a b c d e f g h")
	}

	return sb.String()
}

// getPieceSymbol возвращает символ фигуры с дополнительными эффектами
func (br *EnhancedBoardRenderer) getPieceSymbol(piece chess.Piece) string {
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
