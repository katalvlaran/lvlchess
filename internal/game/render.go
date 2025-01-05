package game

import (
	"fmt"
	"strings"

	"github.com/notnil/chess"
)

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
--+------------------------+
8 | r  n  b  q  k  b  n  r |
7 | p  p  p  p  p  p  p  p |
6 | .  .  .  .  .  .  .  . |
5 | .  .  .  .  .  .  .  . |
4 | .  .  .  .  .  .  .  . |
3 | .  .  .  .  .  .  .  . |
2 | P  P  P  P  P  P  P  P |
1 | R  N  B  Q  K  B  N  R |
--+------------------------+
# | a  b  c  d  e  f  g  h |
*/
func RenderBoardCustom(fen string) (string, error) {
	FEN, err := chess.FEN(fen)
	if err != nil {
		return "", fmt.Errorf("FEN", err)
	}
	game := chess.NewGame(FEN)
	if game == nil {
		return "", fmt.Errorf("ошибка восстановления FEN")
	}
	board := game.Position().Board()

	// Слои ASCII
	var sb strings.Builder
	sb.WriteString("-+------------------------+\n")

	ranks := []chess.Rank{chess.Rank8, chess.Rank7, chess.Rank6, chess.Rank5,
		chess.Rank4, chess.Rank3, chess.Rank2, chess.Rank1}
	files := []chess.File{chess.FileA, chess.FileB, chess.FileC, chess.FileD,
		chess.FileE, chess.FileF, chess.FileG, chess.FileH}

	for _, rank := range ranks {
		// Отрисуем цифру строки
		sb.WriteString(fmt.Sprintf("%d | ", 8-int(rank))) // или rank.String()
		for _, file := range files {
			sq := chess.NewSquare(file, rank)
			piece := board.Piece(sq)
			if piece == chess.NoPiece {
				sb.WriteString(" ▫ ")
			} else {
				sb.WriteString(piece.String() + "  ")
			}
		}
		sb.WriteString("|\n")
	}
	sb.WriteString("--+------------------------+\n")
	sb.WriteString("# | a  b  c  d  e  f  g  h |\n")

	return sb.String(), nil
}
