package game

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/notnil/chess"

	"lvlchess/internal/db/models"
)

// These constants define modes for rendering messages or injecting markup.
// They’re used in Telegram-based logic (some or all might be replaced with Telegram’s parse modes).
const (
	modeKeyboard   = "inline_keyboard" // {btn1},{btn2},..
	modeMarkdown   = "Markdown"
	modeMarkdownV2 = "MarkdownV2" // code
	modeHTML       = "HTML"       // default text
)

// strToSquareMap helps us quickly convert a string like "e4" into a chess.Square constant.
// This is used when we parse user actions (like "move:e2-e4").
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

// StrToSquare attempts to convert something like "e4" into a chess.Square. Returns an error if invalid.
func StrToSquare(p string) (chess.Square, error) {
	if _, ok := strToSquareMap[p]; !ok {
		return chess.NoSquare, errors.New(fmt.Sprintf("invalid square: %s", p))
	}

	return strToSquareMap[p], nil
}

// AssignRandomColors sets WhiteID or BlackID in the given room if not already assigned.
// If the second player has joined (Player2ID != nil) and WhiteID/BlackID are nil, we do a random coin flip
// to decide who is white vs black. If already assigned, does nothing.
func AssignRandomColors(r *models.Room) {
	// If both WhiteID and BlackID are already set, do nothing.
	if r.WhiteID != nil && r.BlackID != nil {
		if r.WhiteID == nil {
			r.WhiteID = r.Player2ID
		} else if r.BlackID == nil {
			r.BlackID = r.Player2ID
		}
		return
	}

	// If we don't even have a second player, we can't assign.
	if r.Player2ID == nil {
		return
	}

	// Initialize the random seed once. If we haven't seeded at startup, do a local seed here.
	rand.Seed(time.Now().UnixNano())

	// 50/50 chance for player1 to be white or black.
	if rand.Intn(2) == 0 {
		r.WhiteID = &r.Player1ID
		r.BlackID = r.Player2ID
	} else {
		r.WhiteID = r.Player2ID
		black := &r.Player1ID
		r.BlackID = black
	}
}

// The arrow icons below are an optional way to indicate direction for a move in ASCII text.
// E.g., if a piece moves from b2->b4, White might see "⬆️", while Black might see "⬇️", etc.

// ArrowForMove calculates an arrow symbol (⬆️, ↘️, etc.) for normal board orientation
// depending on the difference in ranks/files, flipping if it's black's turn.
func ArrowForMove(from, to chess.Square, isWhiteTurn bool) string {
	fFile, fRank := from.File(), from.Rank()
	tFile, tRank := to.File(), to.Rank()

	dFile := int(tFile) - int(fFile) // >0 => right, <0 => left
	dRank := int(tRank) - int(fRank) // >0 => up for white, <0 => down

	if !isWhiteTurn {
		// If black is moving, we invert the direction to match black's perspective
		dRank = -dRank
		dFile = -dFile
	}

	// Vertical
	if dFile == 0 {
		if dRank > 0 {
			return "⬆️"
		}
		return "⬇️"
	}
	// Horizontal
	if dRank == 0 {
		if dFile > 0 {
			return "➡️"
		}
		return "⬅️"
	}
	// Diagonal
	switch {
	case dFile > 0 && dRank > 0:
		return "↗️"
	case dFile > 0 && dRank < 0:
		return "↘️"
	case dFile < 0 && dRank > 0:
		return "↖️"
	default:
		return "↙️"
	}
}

// ArrowForMoveHorizontal is similar but for a "horizontal" orientation.
// Instead of rank/file meaning up/down or left/right, we swap them for a 90° layout.
func ArrowForMoveHorizontal(from, to chess.Square, isWhiteTurn bool) string {
	fFile, fRank := from.File(), from.Rank()
	tFile, tRank := to.File(), to.Rank()

	// For horizontal, let's treat rank differences as "vertical," file differences as "horizontal," etc.
	dFile := int(fRank) - int(tRank) // >0 => up, <0 => down
	dRank := int(tFile) - int(fFile) // >0 => right, <0 => left

	if !isWhiteTurn {
		dRank = -dRank
		dFile = -dFile
	}

	// Then the logic is analogous:
	if dRank == 0 {
		if dFile > 0 {
			return "⬆️"
		}
		return "⬇️"
	}
	if dFile == 0 {
		if dRank > 0 {
			return "➡️"
		}
		return "⬅️"
	}
	switch {
	case dFile > 0 && dRank > 0:
		return "↗️"
	case dFile > 0 && dRank < 0:
		return "↘️"
	case dFile < 0 && dRank > 0:
		return "↖️"
	default:
		return "↙️"
	}
}
