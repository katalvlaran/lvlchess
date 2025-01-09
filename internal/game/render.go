package game

import (
	"fmt"
	"strings"

	"github.com/notnil/chess"
)

const WhiteCell = "□"
const BlackCell = "■"

// MakeNewChessGame возвращает FEN начальной позиции
func MakeNewChessGame() string {
	game := chess.NewGame()
	return game.FEN() // начальная FEN: "rn...kq..." etc.
}

// RenderBoardFromFEN восстанавливает Game из FEN и возвращает ASCII-рисунок
func RenderBoardFromFEN(fen string) (string, error) {
	FEN, err := chess.FEN(fen)
	if err != nil {
		return "", fmt.Errorf("FEN", err)
	}
	game := chess.NewGame(FEN)
	if game == nil {
		return "", fmt.Errorf("ошибка восстановления игры из FEN")
	}

	// У notnil/chess уже есть ASCII-отрисовка через Board()
	ascii := fmt.Sprint(game.Position().Board())
	return ascii, nil
}

/*
Ожидаемый результат генерации
```
□ | a  b  c  d  e  f  g  h | ■
8 | ♖  ♘  ♗  ♕  ♔  ♗  ♘  ♖ | 8
7 | ♙  ♙  ♙  ♙  ♙  ♙  ♙  ♙ | 7
3 | □  ■  □  ■  □  ■  □  ■ | 3
4 | ■  □  ■  □  ■  □  ■  □ | 4
5 | □  ■  □  ■  □  ■  □  ■ | 5
6 | ■  □  ■  □  ■  □  ■  □ | 6
2 | ♟  ♟  ♟  ♟  ♟  ♟  ♟  ♟ | 2
1 | ♜  ♞  ♝  ♛  ♚  ♝  ♞  ♜ | 1
--+------------------------+--
■ | a  b  c  d  e  f  g  h | □
```
*/
func RenderBoardCustom(fen string) (string, error) {
	var game *chess.Game
	if fen != "" {
		// Распарсим FEN через notnil/chess
		fenOption, err := chess.FEN(fen)
		if err != nil {
			return "", fmt.Errorf("ошибка FEN: %w", err)
		}
		game = chess.NewGame(fenOption)
		if game == nil {
			return "", fmt.Errorf("ошибка восстановления FEN (nil game)")
		}
	} else {
		game = chess.NewGame()
		if game == nil {
			return "", fmt.Errorf("ошибка восстановления FEN (nil game)")
		}
	}
	board := game.Position().Board()

	var sb strings.Builder
	// Верхняя "шапка"
	sb.WriteString("```\n") // Начало форматирования (например, Markdown code block)
	sb.WriteString(fmt.Sprintf("%s | a  b  c  d  e  f  g  h | %s\n", WhiteCell, BlackCell))
	sb.WriteString("--+------------------------+--\n")

	// Отрисуем строки с 8 по 1
	ranks := []chess.Rank{
		chess.Rank8, chess.Rank7, chess.Rank6, chess.Rank5,
		chess.Rank4, chess.Rank3, chess.Rank2, chess.Rank1,
	}
	for i, rank := range ranks {
		// Печатаем номер строки слева
		// (i=0 => rank8, i=1 => rank7, ...)
		rowNumber := 8 - i // для визуального совпадения слева
		sb.WriteString(fmt.Sprintf("%d |", rowNumber))

		// Отрисуем клетки a..h
		files := []chess.File{
			chess.FileA, chess.FileB, chess.FileC, chess.FileD,
			chess.FileE, chess.FileF, chess.FileG, chess.FileH,
		}
		for j, file := range files {
			sq := chess.NewSquare(file, rank)
			piece := board.Piece(sq)
			if piece == chess.NoPiece {
				// Выбираем "□" или "■" в зависимости от (i+j), чтобы клетка чередовалась
				if (i+j)%2 == 0 {
					sb.WriteString(fmt.Sprintf(" %s ", WhiteCell))
				} else {
					sb.WriteString(fmt.Sprintf(" %s ", BlackCell))
				}
			} else {
				sb.WriteString(pieceToStr(piece))
			}
		}

		sb.WriteString(fmt.Sprintf("| %d\n", rowNumber)) // номер строки справа
	}

	// "разделитель"
	sb.WriteString("--+------------------------+--\n")

	// Нижняя "шапка"
	sb.WriteString(fmt.Sprintf("%s | a  b  c  d  e  f  g  h | %s\n", BlackCell, WhiteCell))
	sb.WriteString("```") // конец форматирования
	return sb.String(), nil
}

func pieceToStr(p chess.Piece) string {
	// Если хотите эмодзи-фигуры, можно подменять
	// Пока оставим то, что отдаёт p.String() (R/N/B/Q/K/P в верхнем/нижнем регистре),
	// либо используем Юникод-символы:
	switch p {
	case chess.WhitePawn:
		return " ♙ "
	case chess.WhiteRook:
		return " ♖ "
	case chess.WhiteKnight:
		return " ♘ "
	case chess.WhiteBishop:
		return " ♗ "
	case chess.WhiteQueen:
		return " ♕ "
	case chess.WhiteKing:
		return " ♔ "
	case chess.BlackPawn:
		return " ♟ "
	case chess.BlackRook:
		return " ♜ "
	case chess.BlackKnight:
		return " ♞ "
	case chess.BlackBishop:
		return " ♝ "
	case chess.BlackQueen:
		return " ♛ "
	case chess.BlackKing:
		return " ♚ "
	default:
		// fallback (если что-то неизвестное):
		return p.String() + "  "
	}
}
