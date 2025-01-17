package game

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"telega_chess/internal/db"

	"github.com/notnil/chess"
)

const (
	modeKeyboard   = "inline_keyboard" // {btn1},{btn2},..
	modeMarkdown   = "Markdown"
	modeMarkdownV2 = "MarkdownV2" // code
	modeHTML       = "HTML"       // default text
)

var strToSquareMap = map[string]chess.Square{
	"a1": chess.A1, "a2": chess.A2, "a3": chess.A3, "a4": chess.A4, "a5": chess.A5, "a6": chess.A6, "a7": chess.A7, "a8": chess.A8,
	"b1": chess.B1, "b2": chess.B2, "b3": chess.B3, "b4": chess.B4, "b5": chess.B5, "b6": chess.B6, "b7": chess.B7, "b8": chess.B8,
	"c1": chess.C1, "c2": chess.C2, "c3": chess.C3, "c4": chess.C4, "c5": chess.C5, "c6": chess.C6, "c7": chess.C7, "c8": chess.C8,
	"d1": chess.D1, "d2": chess.D2, "d3": chess.D3, "d4": chess.D4, "d5": chess.D5, "d6": chess.D6, "d7": chess.D7, "d8": chess.D8,
	"e1": chess.E1, "e2": chess.E2, "e3": chess.E3, "e4": chess.E4, "e5": chess.E5, "e6": chess.E6, "e7": chess.E7, "e8": chess.E8,
	"f1": chess.F1, "f2": chess.F2, "f3": chess.F3, "f4": chess.F4, "f5": chess.F5, "f6": chess.F6, "f7": chess.F7, "f8": chess.F8,
	"g1": chess.G1, "g2": chess.G2, "g3": chess.G3, "g4": chess.G4, "g5": chess.G5, "g6": chess.G6, "g7": chess.G7, "g8": chess.G8,
	"h1": chess.H1, "h2": chess.H2, "h3": chess.H3, "h4": chess.H4, "h5": chess.H5, "h6": chess.H6, "h7": chess.H7, "h8": chess.H8,
}

func StrToSquare(p string) (chess.Square, error) {
	if _, ok := strToSquareMap[p]; !ok {
		return chess.NoSquare, errors.New(fmt.Sprintf("не существующая позиция:%s", p))
	}

	return strToSquareMap[p], nil
}

// AssignRandomColors задаёт WhiteID и BlackID, если они ещё не были назначены.
// Если already assigned (оба не nil), функция ничего не меняет.
func AssignRandomColors(r *db.Room) {
	// Если уже есть белые/чёрные перепроверка и выход.
	if r.WhiteID != nil && r.BlackID != nil {
		if r.WhiteID == nil {
			r.WhiteID = &r.Player2.ID
		} else if r.BlackID == nil {
			r.BlackID = &r.Player2.ID
		}

		return
	}

	// Убедимся, что у нас есть player1 и player2
	if r.Player2 == nil {
		// нет второго игрока — не будем назначать
		return
	}

	// Инициализируем seed (можно один раз в main(), но если не инициализировано, сделаем локально)
	rand.Seed(time.Now().UnixNano())

	if rand.Intn(2) == 0 {
		// player1 -> white, player2 -> black
		r.WhiteID = &r.Player1.ID
		r.BlackID = &r.Player2.ID
	} else {
		// player2 -> white, player1 -> black
		r.WhiteID = &r.Player2.ID
		black := &r.Player1.ID
		r.BlackID = black
	}
}

// ArrowForMove - определяет направление хода на стандартной доске.
// Параметр IsWhiteTurn указывает, играет ли сторона белыми фигурами.
func ArrowForMove(from, to chess.Square, IWT bool) string {
	fFile, fRank := from.File(), from.Rank()
	tFile, tRank := to.File(), to.Rank()

	dFile := int(tFile) - int(fFile) // >0 — вправо, <0 — влево
	dRank := int(tRank) - int(fRank) // >0 — вверх (для белых), <0 — вниз

	if !IWT {
		// Инвертируем направление для чёрных (поворот на 180 градусов)
		dRank = -dRank
		dFile = -dFile
	}

	// Вертикальные движения
	if dFile == 0 {
		if dRank > 0 {
			return "⬆️"
		} else {
			return "⬇️"
		}
	}

	// Горизонтальные движения
	if dRank == 0 {
		if dFile > 0 {
			return "➡️"
		} else {
			return "⬅️"
		}
	}

	// Диагональные движения
	if dFile > 0 && dRank > 0 {
		return "↗️"
	} else if dFile > 0 && dRank < 0 {
		return "↘️"
	} else if dFile < 0 && dRank > 0 {
		return "↖️"
	} else {
		return "↙️"
	}
}

// ArrowForMoveHorizontal - определяет направление хода для горизонтально расположенной доски.
// Параметр isWhiteTurn указывает, играет ли сторона белыми фигурами.
func ArrowForMoveHorizontal(from, to chess.Square, IWT bool) string {
	fFile, fRank := from.File(), from.Rank()
	tFile, tRank := to.File(), to.Rank()

	dFile := int(fRank) - int(tRank) // >0 — вверх, <0 — вниз (для горизонтальной доски)
	dRank := int(tFile) - int(fFile) // >0 — вправо, <0 — влево (для горизонтальной доски)

	if !IWT {
		// Инвертируем направление для чёрных (поворот на 180 градусов)
		dRank = -dRank
		dFile = -dFile
	}

	// Вертикальные движения (горизонтальная доска)
	if dRank == 0 {
		if dFile > 0 {
			return "⬆️"
		} else {
			return "⬇️"
		}
	}

	// Горизонтальные движения (горизонтальная доска)
	if dFile == 0 {
		if dRank > 0 {
			return "➡️"
		} else {
			return "⬅️"
		}
	}

	// Диагональные движения
	if dFile > 0 && dRank > 0 {
		return "↗️"
	} else if dFile > 0 && dRank < 0 {
		return "↘️"
	} else if dFile < 0 && dRank > 0 {
		return "↖️"
	} else {
		return "↙️"
	}
}
