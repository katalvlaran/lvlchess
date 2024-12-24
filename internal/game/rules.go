package game

import (
	"fmt"

	"github.com/katalvlaran/telega-shess/internal/utils"
	"github.com/notnil/chess"
)

// RuleValidator проверяет правила игры
type RuleValidator struct {
	log *utils.Logger
}

// NewRuleValidator создает новый валидатор правил
func NewRuleValidator() *RuleValidator {
	return &RuleValidator{
		log: utils.Logger(),
	}
}

// ValidateMove проверяет корректность хода
func (rv *RuleValidator) ValidateMove(game *chess.Game, move *chess.Move) error {
	pos := game.Position()

	// Проверка шаха
	if pos.InCheck() {
		validMoves := pos.ValidMoves()
		moveIsValid := false
		for _, validMove := range validMoves {
			if validMove.String() == move.String() {
				moveIsValid = true
				break
			}
		}
		if !moveIsValid {
			return fmt.Errorf("move does not resolve check")
		}
	}

	// Проверка рокировки
	if isCastling(move) {
		if err := rv.validateCastling(pos, move); err != nil {
			return err
		}
	}

	// Проверка взятия на п��оходе
	if isEnPassant(pos, move) {
		if err := rv.validateEnPassant(pos, move); err != nil {
			return err
		}
	}

	// Проверка превращения пешки
	if isPawnPromotion(pos, move) && move.Promotion == chess.NoPieceType {
		return fmt.Errorf("pawn promotion piece type required")
	}

	return nil
}

// validateCastling проверяет возможность рокировки
func (rv *RuleValidator) validateCastling(pos *chess.Position, move *chess.Move) error {
	// Проверка, не ходил ли король
	if pos.CastleRights().Has(chess.WhiteKingSide) || pos.CastleRights().Has(chess.WhiteQueenSide) ||
		pos.CastleRights().Has(chess.BlackKingSide) || pos.CastleRights().Has(chess.BlackQueenSide) {
		return nil
	}
	return fmt.Errorf("castling not allowed: king or rook has moved")
}

// validateEnPassant проверяет возможность взятия на проходе
func (rv *RuleValidator) validateEnPassant(pos *chess.Position, move *chess.Move) error {
	if pos.EnPassantSquare() == nil {
		return fmt.Errorf("en passant not possible: no valid target")
	}
	return nil
}

// Вспомогательные функции
func isCastling(move *chess.Move) bool {
	from := move.S1()
	to := move.S2()
	return (from == chess.E1 && (to == chess.G1 || to == chess.C1)) ||
		(from == chess.E8 && (to == chess.G8 || to == chess.C8))
}

func isEnPassant(pos *chess.Position, move *chess.Move) bool {
	if pos.EnPassantSquare() == nil {
		return false
	}
	piece := pos.Board().Piece(move.S1())
	return piece.Type() == chess.Pawn && move.S2() == *pos.EnPassantSquare()
}

func isPawnPromotion(pos *chess.Position, move *chess.Move) bool {
	piece := pos.Board().Piece(move.S1())
	return piece.Type() == chess.Pawn && (move.S2().Rank() == 7 || move.S2().Rank() == 0)
}
