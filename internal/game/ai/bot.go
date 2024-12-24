package ai

import (
	"math/rand"
	"time"

	"github.com/notnil/chess"
)

// Difficulty определяет уровень сложности бота
type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

// ChessBot представляет AI для игры
type ChessBot struct {
	difficulty Difficulty
	evaluator  *PositionEvaluator
}

// NewChessBot создает нового бота
func NewChessBot(difficulty Difficulty) *ChessBot {
	return &ChessBot{
		difficulty: difficulty,
		evaluator:  NewPositionEvaluator(),
	}
}

// GetMove возвращает лучший ход для текущей позиции
func (cb *ChessBot) GetMove(position *chess.Position) *chess.Move {
	validMoves := position.ValidMoves()
	if len(validMoves) == 0 {
		return nil
	}

	switch cb.difficulty {
	case Easy:
		return cb.getRandomMove(validMoves)
	case Medium:
		return cb.getMediumMove(position, validMoves)
	case Hard:
		return cb.getHardMove(position, validMoves, 3) // глубина 3
	default:
		return cb.getRandomMove(validMoves)
	}
}

// getRandomMove возвращает случайный ход
func (cb *ChessBot) getRandomMove(moves []*chess.Move) *chess.Move {
	rand.Seed(time.Now().UnixNano())
	return moves[rand.Intn(len(moves))]
}

// getMediumMove использует простую эвристику для выбора хода
func (cb *ChessBot) getMediumMove(pos *chess.Position, moves []*chess.Move) *chess.Move {
	var bestMove *chess.Move
	bestScore := float64(-9999)

	for _, move := range moves {
		newPos := pos.Update(move)
		score := cb.evaluator.EvaluatePosition(newPos)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	return bestMove
}

// getHardMove использует минимакс с альфа-бета отсечением
func (cb *ChessBot) getHardMove(pos *chess.Position, moves []*chess.Move, depth int) *chess.Move {
	var bestMove *chess.Move
	bestScore := float64(-9999)
	alpha := float64(-9999)
	beta := float64(9999)

	for _, move := range moves {
		newPos := pos.Update(move)
		score := -cb.alphaBeta(newPos, depth-1, -beta, -alpha, false)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		alpha = max(alpha, score)
	}

	return bestMove
}

// alphaBeta реализует алгоритм альфа-бе��а отсечения
func (cb *ChessBot) alphaBeta(pos *chess.Position, depth int, alpha, beta float64, maximizing bool) float64 {
	if depth == 0 {
		return cb.evaluator.EvaluatePosition(pos)
	}

	moves := pos.ValidMoves()
	if len(moves) == 0 {
		if pos.InCheck() {
			return -9999 // Мат
		}
		return 0 // Пат
	}

	if maximizing {
		value := float64(-9999)
		for _, move := range moves {
			newPos := pos.Update(move)
			value = max(value, cb.alphaBeta(newPos, depth-1, alpha, beta, false))
			alpha = max(alpha, value)
			if alpha >= beta {
				break
			}
		}
		return value
	} else {
		value := float64(9999)
		for _, move := range moves {
			newPos := pos.Update(move)
			value = min(value, cb.alphaBeta(newPos, depth-1, alpha, beta, true))
			beta = min(beta, value)
			if alpha >= beta {
				break
			}
		}
		return value
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
