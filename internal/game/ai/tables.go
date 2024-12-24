package ai

import (
	"github.com/notnil/chess"
)

// initializePieceSquareTables инициализирует таблицы оценки позиций фигур
func initializePieceSquareTables() map[chess.PieceType][64]float64 {
	return map[chess.PieceType][64]float64{
		chess.Pawn:   pawnTable,
		chess.Knight: knightTable,
		chess.Bishop: bishopTable,
		chess.Rook:   rookTable,
		chess.Queen:  queenTable,
		chess.King:   kingMiddleGameTable,
	}
}

// Таблицы оценки позиций для каждой фигуры
var pawnTable = [64]float64{
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	5.0, 5.0, 5.0, 5.0, 5.0, 5.0, 5.0, 5.0,
	1.0, 1.0, 2.0, 3.0, 3.0, 2.0, 1.0, 1.0,
	0.5, 0.5, 1.0, 2.5, 2.5, 1.0, 0.5, 0.5,
	0.0, 0.0, 0.0, 2.0, 2.0, 0.0, 0.0, 0.0,
	0.5, -0.5, -1.0, 0.0, 0.0, -1.0, -0.5, 0.5,
	0.5, 1.0, 1.0, -2.0, -2.0, 1.0, 1.0, 0.5,
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
}

var knightTable = [64]float64{
	-5.0, -4.0, -3.0, -3.0, -3.0, -3.0, -4.0, -5.0,
	-4.0, -2.0, 0.0, 0.0, 0.0, 0.0, -2.0, -4.0,
	-3.0, 0.0, 1.0, 1.5, 1.5, 1.0, 0.0, -3.0,
	-3.0, 0.5, 1.5, 2.0, 2.0, 1.5, 0.5, -3.0,
	-3.0, 0.0, 1.5, 2.0, 2.0, 1.5, 0.0, -3.0,
	-3.0, 0.5, 1.0, 1.5, 1.5, 1.0, 0.5, -3.0,
	-4.0, -2.0, 0.0, 0.5, 0.5, 0.0, -2.0, -4.0,
	-5.0, -4.0, -3.0, -3.0, -3.0, -3.0, -4.0, -5.0,
}

var bishopTable = [64]float64{
	-2.0, -1.0, -1.0, -1.0, -1.0, -1.0, -1.0, -2.0,
	-1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -1.0,
	-1.0, 0.0, 0.5, 1.0, 1.0, 0.5, 0.0, -1.0,
	-1.0, 0.5, 0.5, 1.0, 1.0, 0.5, 0.5, -1.0,
	-1.0, 0.0, 1.0, 1.0, 1.0, 1.0, 0.0, -1.0,
	-1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, -1.0,
	-1.0, 0.5, 0.0, 0.0, 0.0, 0.0, 0.5, -1.0,
	-2.0, -1.0, -1.0, -1.0, -1.0, -1.0, -1.0, -2.0,
}

var rookTable = [64]float64{
	0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
	0.5, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.5,
	-0.5, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.5,
	-0.5, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.5,
	-0.5, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.5,
	-0.5, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.5,
	-0.5, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -0.5,
	0.0, 0.0, 0.0, 0.5, 0.5, 0.0, 0.0, 0.0,
}

var queenTable = [64]float64{
	-2.0, -1.0, -1.0, -0.5, -0.5, -1.0, -1.0, -2.0,
	-1.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, -1.0,
	-1.0, 0.0, 0.5, 0.5, 0.5, 0.5, 0.0, -1.0,
	-0.5, 0.0, 0.5, 0.5, 0.5, 0.5, 0.0, -0.5,
	0.0, 0.0, 0.5, 0.5, 0.5, 0.5, 0.0, -0.5,
	-1.0, 0.5, 0.5, 0.5, 0.5, 0.5, 0.0, -1.0,
	-1.0, 0.0, 0.5, 0.0, 0.0, 0.0, 0.0, -1.0,
	-2.0, -1.0, -1.0, -0.5, -0.5, -1.0, -1.0, -2.0,
}

var kingMiddleGameTable = [64]float64{
	-3.0, -4.0, -4.0, -5.0, -5.0, -4.0, -4.0, -3.0,
	-3.0, -4.0, -4.0, -5.0, -5.0, -4.0, -4.0, -3.0,
	-3.0, -4.0, -4.0, -5.0, -5.0, -4.0, -4.0, -3.0,
	-3.0, -4.0, -4.0, -5.0, -5.0, -4.0, -4.0, -3.0,
	-2.0, -3.0, -3.0, -4.0, -4.0, -3.0, -3.0, -2.0,
	-1.0, -2.0, -2.0, -2.0, -2.0, -2.0, -2.0, -1.0,
	2.0, 2.0, 0.0, 0.0, 0.0, 0.0, 2.0, 2.0,
	2.0, 3.0, 1.0, 0.0, 0.0, 1.0, 3.0, 2.0,
}

// Вспомогательные функции для оценки позиции
func (pe *PositionEvaluator) evaluatePawnShield(pos *chess.Position, kingSquare chess.Square, color chess.Color) float64 {
	var score float64
	board := pos.Board()
	rank := kingSquare.Rank()
	file := kingSquare.File()

	// Проверяем пешки перед королем
	for f := max(0, file-1); f <= min(7, file+1); f++ {
		if color == chess.White {
			// Для белого короля проверяем пешки на 6-й и 7-й горизонталях
			for r := max(0, rank-2); r <= rank-1; r++ {
				piece := board.Piece(chess.Square(r*8 + f))
				if piece.Type() == chess.Pawn && piece.Color() == chess.White {
					score += 0.5
				}
			}
		} else {
			// Для черного короля проверяем пешки на 1-й и 2-й горизонталях
			for r := rank + 1; r <= min(7, rank+2); r++ {
				piece := board.Piece(chess.Square(r*8 + f))
				if piece.Type() == chess.Pawn && piece.Color() == chess.Black {
					score += 0.5
				}
			}
		}
	}

	return score
}

// evaluateOpenFiles оценивает открытые линии около короля
func (pe *PositionEvaluator) evaluateOpenFiles(pos *chess.Position, kingSquare chess.Square) float64 {
	var penalty float64
	board := pos.Board()
	file := kingSquare.File()

	// Проверяем вертикали вокруг короля
	for f := max(0, file-1); f <= min(7, file+1); f++ {
		hasOwnPawn := false
		for rank := 0; rank < 8; rank++ {
			piece := board.Piece(chess.Square(rank*8 + f))
			if piece.Type() == chess.Pawn && piece.Color() == pos.Turn() {
				hasOwnPawn = true
				break
			}
		}
		if !hasOwnPawn {
			penalty += 0.5
		}
	}

	return penalty
}
