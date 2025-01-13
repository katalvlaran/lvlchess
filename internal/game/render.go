package game

import (
	"fmt"
	"strings"

	"github.com/notnil/chess"
)

const (
	//типы отображений доски
	WhiteBoard      = "white"
	HorizontalBoard = "horizontal"
	BlackBoard      = "black"

	// иконка клетки соответствующего цвета
	WhiteCell = "□"
	BlackCell = "■"
)

// RenderASCIIBoardWhite — Стандартное(для белых) отображение доски.
/* Ожидаемый результат генерации:
```
□ | a  b  c  d  e  f  g  h | ■
--+------------------------+--
8 | ♜  ♞  ♝  ♛  ♚  ♝  ♞  ♜ | 8
7 | ♟  ♟  ♟  ♟  ♟  ♟  ♟  ♟ | 7
6 | □  ■  □  ■  □  ■  □  ■ | 6
5 | ■  □  ■  □  ■  □  ■  □ | 5
4 | □  ■  □  ■  □  ■  □  ■ | 4
3 | ■  □  ■  □  ■  □  ■  □ | 3
2 | ♙  ♙  ♙  ♙  ♙  ♙  ♙  ♙ | 2
1 | ♖  ♘  ♗  ♕  ♔  ♗  ♘  ♖ | 1
--+------------------------+--
■ | a  b  c  d  e  f  g  h | □
```
*/
func RenderASCIIBoardWhite(fen string) (string, error) {
	// порядок chess.Rank для WhiteBoard положения
	ranks := []chess.Rank{
		chess.Rank8, chess.Rank7, chess.Rank6, chess.Rank5,
		chess.Rank4, chess.Rank3, chess.Rank2, chess.Rank1,
	}
	// порядок chess.File для WhiteBoard положения
	files := []chess.File{
		chess.FileA, chess.FileB, chess.FileC, chess.FileD,
		chess.FileE, chess.FileF, chess.FileG, chess.FileH,
	}

	return RenderASCIIBoard(fen, ranks, files, WhiteBoard)
}

// RenderASCIIBoardHorizontal — Горизонтальное отображение доски(повёрнуто в право, на угол 90).
/* Ожидаемый результат генерации:
```
■ | 1  2  3  4  5  6  7  8 | □
--+------------------------+--
a | ♖  ♙  ■  □  ■  □  ♟  ♜ | a
b | ♘  ♙  □  ■  □  ■  ♟  ♞ | b
c | ♗  ♙  ■  □  ■  □  ♟  ♝ | c
d | ♕  ♙  □  ■  □  ■  ♟  ♛ | d
e | ♔  ♙  ■  □  ■  □  ♟  ♚ | e
f | ♗  ♙  □  ■  □  ■  ♟  ♝ | f
g | ♘  ♙  ■  □  ■  □  ♟  ♞ | g
h | ♖  ♙  □  ■  □  ■  ♟  ♜ | h
--+------------------------+--
□ | 1  2  3  4  5  6  7  8 | ■
```
*/
func RenderASCIIBoardHorizontal(fen string) (string, error) {
	// порядок chess.Rank для HorizontalBoard положения
	ranks := []chess.Rank{
		chess.Rank1, chess.Rank2, chess.Rank3, chess.Rank4,
		chess.Rank5, chess.Rank6, chess.Rank7, chess.Rank8,
	}
	// порядок chess.File для HorizontalBoard положения
	files := []chess.File{
		chess.FileA, chess.FileB, chess.FileC, chess.FileD,
		chess.FileE, chess.FileF, chess.FileG, chess.FileH,
	}

	return RenderASCIIBoard(fen, ranks, files, HorizontalBoard)
}

// RenderASCIIBoardBlack — перевёрнутое отображение доски().
/* Ожидаемый результат генерации:
```
□ | h  g  f  e  d  c  b  a | ■
--+------------------------+--
1 | ♖  ♘  ♗  ♕  ♔  ♗  ♘  ♖ | 1
2 | ♙  ♙  ♙  ♙  ♙  ♙  ♙  ♙ | 2
3 | □  ■  □  ■  □  ■  □  ■ | 3
4 | ■  □  ■  □  ■  □  ■  □ | 4
5 | □  ■  □  ■  □  ■  □  ■ | 5
6 | ■  □  ■  □  ■  □  ■  □ | 6
7 | ♟  ♟  ♟  ♟  ♟  ♟  ♟  ♟ | 7
8 | ♜  ♞  ♝  ♛  ♚  ♝  ♞  ♜ | 8
--+------------------------+--
■ | h  g  f  e  d  c  b  a | □
```
*/
func RenderASCIIBoardBlack(fen string) (string, error) {
	// порядок chess.Rank для BlackBoard положения
	ranks := []chess.Rank{
		chess.Rank1, chess.Rank2, chess.Rank3, chess.Rank4,
		chess.Rank5, chess.Rank6, chess.Rank7, chess.Rank8,
	}
	// порядок chess.File для BlackBoard положения
	files := []chess.File{
		chess.FileH, chess.FileG, chess.FileF, chess.FileE,
		chess.FileD, chess.FileC, chess.FileB, chess.FileA,
	}

	return RenderASCIIBoard(fen, ranks, files, BlackBoard)
}

// RenderASCIIBoard — общая логика рендера доски.
func RenderASCIIBoard(fen string, ranks []chess.Rank, files []chess.File, orientation string) (string, error) {
	game, err := parseFEN(fen)
	if err != nil {
		return "", err
	}
	board := game.Position().Board()

	var sb strings.Builder
	header, footer := getHeaderFooter(orientation)
	// Верхняя шапка
	sb.WriteString("```\n")
	sb.WriteString(header)
	sb.WriteString("--+------------------------+--\n")

	// Отрисовка строк
	if orientation == HorizontalBoard { // логика рендера для горизонтального(HorizontalBoard) отображения.
		for i, file := range files {
			sb.WriteString(fmt.Sprintf("%s |", string('a'+i)))
			for j := range ranks {
				sq := chess.NewSquare(file, ranks[j])
				piece := board.Piece(sq)
				sb.WriteString(formatSquare(piece, (i+j)%2 == 0))
			}
			sb.WriteString(fmt.Sprintf("| %s\n", string('a'+i)))
		}
	} else { // рендерим доску для WhiteBoard(белых) или BlackBoard(чёрных)
		for i, rank := range ranks {
			colorRank := i + 1 // for black side
			if orientation == WhiteBoard {
				colorRank = 8 - i
			}
			sb.WriteString(fmt.Sprintf("%d |", colorRank))
			for j, file := range files {
				sq := chess.NewSquare(file, rank)
				piece := board.Piece(sq)
				sb.WriteString(formatSquare(piece, (i+j)%2 == 0))
			}
			sb.WriteString(fmt.Sprintf("| %d\n", colorRank))
		}
	}

	// Нижняя шапка
	sb.WriteString("--+------------------------+--\n")
	sb.WriteString(footer)
	sb.WriteString("```")
	return sb.String(), nil
}

// getHeaderFooter — шапка и нижняя граница доски.
func getHeaderFooter(orientation string) (string, string) {
	var header, footer, format string
	switch orientation {
	case WhiteBoard:
		format = "%s | a  b  c  d  e  f  g  h | %s\n"
		header, footer = fmt.Sprintf(format, WhiteCell, BlackCell), fmt.Sprintf(format, BlackCell, WhiteCell)
	case HorizontalBoard:
		format = "%s | 1  2  3  4  5  6  7  8 | %s\n"
		header, footer = fmt.Sprintf(format, BlackCell, WhiteCell), fmt.Sprintf(format, WhiteCell, BlackCell)
	case BlackBoard:
		format = "%s | h  g  f  e  d  c  b  a | %s\n"
		header, footer = fmt.Sprintf(format, WhiteCell, BlackCell), fmt.Sprintf(format, BlackCell, WhiteCell)
	}

	return header, footer
}

// parseFEN — парсинг FEN.
func parseFEN(fen string) (*chess.Game, error) {
	if fen == "" {
		return chess.NewGame(), nil
	}
	fenOption, err := chess.FEN(fen)
	if err != nil {
		return nil, fmt.Errorf("ошибка FEN: %w", err)
	}

	return chess.NewGame(fenOption), nil
}

// formatSquare — форматирование клетки доски.
func formatSquare(piece chess.Piece, isWhite bool) string {
	if piece == chess.NoPiece {
		if isWhite {
			return fmt.Sprintf(" %s ", WhiteCell)
		}
		return fmt.Sprintf(" %s ", BlackCell)
	}
	return fmt.Sprintf(" %s ", PieceToStr(piece))
}

// PieceToStr — представление фигур.
func PieceToStr(p chess.Piece) string {
	switch p {
	case chess.WhitePawn:
		return "♙"
	case chess.WhiteRook:
		return "♖"
	case chess.WhiteKnight:
		return "♘"
	case chess.WhiteBishop:
		return "♗"
	case chess.WhiteQueen:
		return "♕"
	case chess.WhiteKing:
		return "♔"
	case chess.BlackPawn:
		return "♟"
	case chess.BlackRook:
		return "♜"
	case chess.BlackKnight:
		return "♞"
	case chess.BlackBishop:
		return "♝"
	case chess.BlackQueen:
		return "♛"
	case chess.BlackKing:
		return "♚"
	default:
		return p.String()
	}
}
