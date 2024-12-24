package ai

import (
	"fmt"
	"strings"

	"github.com/notnil/chess"
)

// PositionAnalyzer анализирует шахматные позиции
type PositionAnalyzer struct {
	evaluator *PositionEvaluator
}

// NewPositionAnalyzer создает новый анализатор позиций
func NewPositionAnalyzer() *PositionAnalyzer {
	return &PositionAnalyzer{
		evaluator: NewPositionEvaluator(),
	}
}

// AnalyzePosition анализирует позицию и возвращает подробный отчет
func (pa *PositionAnalyzer) AnalyzePosition(pos *chess.Position) string {
	if pos == nil {
		return "Invalid position"
	}

	var analysis strings.Builder

	// Проверяем специальные состояния
	switch pos.Status() {
	case chess.Checkmate:
		if pos.Turn() == chess.White {
			return "Мат. Черные победили."
		}
		return "Мат. Белые победили."
	case chess.Stalemate:
		return "Пат. Ничья."
	case chess.ThreefoldRepetition:
		return "Ничья по троекратному повторению позиции."
	case chess.FiftyMoveRule:
		return "Ничья по правилу 50 ходов."
	case chess.InsufficientMaterial:
		return "Ничья из-за недостаточного материала."
	}

	// Общая оценка позиции
	evaluation := pa.evaluator.EvaluatePosition(pos)
	analysis.WriteString(fmt.Sprintf("Оценка позиции: %.2f\n", evaluation))

	// Анализ материала
	material := pa.evaluator.evaluateMaterial(pos)
	analysis.WriteString(fmt.Sprintf("Материальное преимущество: %.2f\n", material))

	// Анализ контроля центра
	centerControl := pa.evaluator.evaluatePosition(pos)
	analysis.WriteString(fmt.Sprintf("Контроль центра: %.2f\n", centerControl))

	// Безопасность короля
	kingSafety := pa.evaluator.evaluateKingSafety(pos)
	analysis.WriteString(fmt.Sprintf("Безопасность короля: %.2f\n", kingSafety))

	// Рекомендации
	analysis.WriteString("\nРекомендации:\n")
	analysis.WriteString(pa.generateRecommendations(pos, evaluation))

	return analysis.String()
}

// generateRecommendations генерирует рекомендации на основе анализа позиции
func (pa *PositionAnalyzer) generateRecommendations(pos *chess.Position, evaluation float64) string {
	var recommendations strings.Builder

	if pos.Turn() == chess.White {
		if evaluation > 2.0 {
			recommendations.WriteString("- Используйте материальное преимущество\n")
			recommendations.WriteString("- Упрощайте позицию через размены\n")
		} else if evaluation < -2.0 {
			recommendations.WriteString("- Ищите тактические возможности\n")
			recommendations.WriteString("- Усильте давление на слабые пункты противника\n")
		} else {
			recommendations.WriteString("- Улучшайте позиции фигур\n")
			recommendations.WriteString("- Контролируйте центр\n")
		}
	} else {
		if evaluation < -2.0 {
			recommendations.WriteString("- Используйте материальное преимущество\n")
			recommendations.WriteString("- Упрощайте позицию через размены\n")
		} else if evaluation > 2.0 {
			recommendations.WriteString("- Ищите тактические возможности\n")
			recommendations.WriteString("- Усильте давление на слабые пункты противника\n")
		} else {
			recommendations.WriteString("- Улучшайте позиции фигур\n")
			recommendations.WriteString("- Контролируйте центр\n")
		}
	}

	return recommendations.String()
}
