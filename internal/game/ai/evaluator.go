package ai

import (
	"github.com/notnil/chess"
)

// PieceValues содержит стандартные значения фигур
var PieceValues = map[chess.PieceType]float64{
	chess.Pawn:   1.0,
	chess.Knight: 3.0,
	chess.Bishop: 3.0,
	chess.Rook:   5.0,
	chess.Queen:  9.0,
	chess.King:   0.0, // Король не оценивается по материалу
}

// PositionEvaluator оценивает шахматные позиции
type PositionEvaluator struct {
	pieceSquareTables map[chess.PieceType][64]float64
}

// NewPositionEvaluator создает новый оценщик позиций
func NewPositionEvaluator() *PositionEvaluator {
	return &PositionEvaluator{
		pieceSquareTables: initializePieceSquareTables(),
	}
}

// EvaluatePosition оценивает позицию комплексно
func (pe *PositionEvaluator) EvaluatePosition(pos *chess.Position) float64 {
	if pos == nil {
		return 0
	}

	switch pos.Status() {
	case chess.Checkmate:
		if pos.Turn() == chess.White {
			return -9999
		}
		return 9999
	case chess.Stalemate, chess.ThreefoldRepetition, chess.FiftyMoveRule, chess.InsufficientMaterial:
		return 0
	}

	score := pe.evaluateMaterial(pos)
	score += pe.evaluatePosition(pos)
	score += pe.evaluatePawnStructure(pos)
	score += pe.evaluateKingSafety(pos)
	score += pe.evaluateMobility(pos)

	if pos.Turn() == chess.Black {
		score = -score
	}

	return score
}

// evaluateMaterial оценивает материальное преимущество
func (pe *PositionEvaluator) evaluateMaterial(pos *chess.Position) float64 {
	var score float64
	board := pos.Board()

	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece.Type() == chess.NoPieceType {
			continue
		}

		value := PieceValues[piece.Type()]
		if piece.Color() == chess.White {
			score += value
		} else {
			score -= value
		}
	}

	return score
}

// evaluatePosition оценивает позиционное преимущество
func (pe *PositionEvaluator) evaluatePosition(pos *chess.Position) float64 {
	var score float64
	board := pos.Board()

	for sq := 0; sq < 64; sq++ {
		piece := board.Piece(chess.Square(sq))
		if piece.Type() == chess.NoPieceType {
			continue
		}

		psqValue := pe.pieceSquareTables[piece.Type()][sq]
		if piece.Color() == chess.White {
			score += psqValue
		} else {
			score -= psqValue
		}
	}

	return score
}

// evaluatePawnStructure оценивает структуру пешек
func (pe *PositionEvaluator) evaluatePawnStructure(pos *chess.Position) float64 {
	var score float64
	board := pos.Board()

	// Сдвоенные пешки
	for file := 0; file < 8; file++ {
		whitePawns := 0
		blackPawns := 0
		for rank := 0; rank < 8; rank++ {
			piece := board.Piece(chess.Square(rank*8 + file))
			if piece.Type() == chess.Pawn {
				if piece.Color() == chess.White {
					whitePawns++
				} else {
					blackPawns++
				}
			}
		}
		if whitePawns > 1 {
			score -= 0.5 * float64(whitePawns-1)
		}
		if blackPawns > 1 {
			score += 0.5 * float64(blackPawns-1)
		}
	}

	return score
}

// evaluateKingSafety оценивает безопасность короля
func (pe *PositionEvaluator) evaluateKingSafety(pos *chess.Position) float64 {
	var score float64
	board := pos.Board()

	whiteKingSquare := board.FindKing(chess.White)
	blackKingSquare := board.FindKing(chess.Black)

	whiteKingSafety := pe.evaluateKingSquareSafety(pos, whiteKingSquare, chess.White)
	blackKingSafety := pe.evaluateKingSquareSafety(pos, blackKingSquare, chess.Black)

	return whiteKingSafety - blackKingSafety
}

// evaluateKingSquareSafety оценивает безопасность короля на конкретном поле
func (pe *PositionEvaluator) evaluateKingSquareSafety(pos *chess.Position, square chess.Square, color chess.Color) float64 {
	var safety float64

	// Проверяем наличие пешек перед королем
	pawnShield := pe.evaluatePawnShield(pos, square, color)
	safety += pawnShield

	// Штраф за открытые линии около короля
	openFiles := pe.evaluateOpenFiles(pos, square)
	safety -= openFiles

	return safety
}

// evaluateMobility оценивает мобильность фигур
func (pe *PositionEvaluator) evaluateMobility(pos *chess.Position) float64 {
	whiteMoves := len(pos.ValidMoves())
	pos = pos.Update(nil) // Пропускаем ход
	blackMoves := len(pos.ValidMoves())

	return 0.1 * float64(whiteMoves-blackMoves)
}
