package game

import (
	"fmt"
	"strings"

	"github.com/notnil/chess"
)

const (
	// Board orientation constants.
	WhiteBoard      = "white"      // Standard white-oriented board (top ranks=8 down to 1).
	HorizontalBoard = "horizontal" // Rotated 90°, ranks become columns etc.
	BlackBoard      = "black"      // Inverted black-oriented board (top ranks=1 up to 8).
)

// ASCII placeholders for empty squares, used to visually render
// light/dark squares in a textual chessboard.
const (
	WhiteCell = "□"
	BlackCell = "■"
)

// RenderASCIIBoardWhite renders the board from White's perspective (ranks 8->1, files a->h).
// ```
// □ | a  b  c  d  e  f  g  h | ■
// --+------------------------+--
// 8 | ♜  ♞  ♝  ♛  ♚  ♝  ♞  ♜ | 8
// 7 | ♟  ♟  ♟  ♟  ♟  ♟  ♟  ♟ | 7
// 6 | □  ■  □  ■  □  ■  □  ■ | 6
// 5 | ■  □  ■  □  ■  □  ■  □ | 5
// 4 | □  ■  □  ■  □  ■  □  ■ | 4
// 3 | ■  □  ■  □  ■  □  ■  □ | 3
// 2 | ♙  ♙  ♙  ♙  ♙  ♙  ♙  ♙ | 2
// 1 | ♖  ♘  ♗  ♕  ♔  ♗  ♘  ♖ | 1
// --+------------------------+--
// ■ | a  b  c  d  e  f  g  h | □
// ```
// This is the classic orientation (white pieces at bottom, black at top).
// For example, if your FEN is the starting position, it prints lines from rank 8 down to rank 1.
func RenderASCIIBoardWhite(fen string) (string, error) {
	// The rank order from top to bottom (8..1) for White’s orientation.
	ranks := []chess.Rank{
		chess.Rank8, chess.Rank7, chess.Rank6, chess.Rank5,
		chess.Rank4, chess.Rank3, chess.Rank2, chess.Rank1,
	}
	// The file order (a..h).
	files := []chess.File{
		chess.FileA, chess.FileB, chess.FileC, chess.FileD,
		chess.FileE, chess.FileF, chess.FileG, chess.FileH,
	}
	return RenderASCIIBoard(fen, ranks, files, WhiteBoard)
}

// RenderASCIIBoardHorizontal creates a "horizontal" ASCII board, effectively a 90° rotation.
// ```
// ■ | 1  2  3  4  5  6  7  8 | □
// --+------------------------+--
// a | ♖  ♙  ■  □  ■  □  ♟  ♜ | a
// b | ♘  ♙  □  ■  □  ■  ♟  ♞ | b
// c | ♗  ♙  ■  □  ■  □  ♟  ♝ | c
// d | ♕  ♙  □  ■  □  ■  ♟  ♛ | d
// e | ♔  ♙  ■  □  ■  □  ♟  ♚ | e
// f | ♗  ♙  □  ■  □  ■  ♟  ♝ | f
// g | ♘  ♙  ■  □  ■  □  ♟  ♞ | g
// h | ♖  ♙  □  ■  □  ■  ♟  ♜ | h
// --+------------------------+--
// □ | 1  2  3  4  5  6  7  8 | ■
// ```
// Some prefer a sideways layout for demonstration. This is less common, but shows how you can
// choose an alternate rank/file ordering.
func RenderASCIIBoardHorizontal(fen string) (string, error) {
	// If "horizontal," you might treat rank 1..8 as left->right, etc.
	// Here, we define the rank order from top to bottom as rank1..rank8:
	ranks := []chess.Rank{
		chess.Rank1, chess.Rank2, chess.Rank3, chess.Rank4,
		chess.Rank5, chess.Rank6, chess.Rank7, chess.Rank8,
	}
	// The file order remains a..h, but we interpret them differently in rendering.
	files := []chess.File{
		chess.FileA, chess.FileB, chess.FileC, chess.FileD,
		chess.FileE, chess.FileF, chess.FileG, chess.FileH,
	}
	return RenderASCIIBoard(fen, ranks, files, HorizontalBoard)
}

// RenderASCIIBoardBlack renders the board from Black's perspective (ranks 1->8, files h->a).
// ```
// □ | h  g  f  e  d  c  b  a | ■
// --+------------------------+--
// 1 | ♖  ♘  ♗  ♕  ♔  ♗  ♘  ♖ | 1
// 2 | ♙  ♙  ♙  ♙  ♙  ♙  ♙  ♙ | 2
// 3 | □  ■  □  ■  □  ■  □  ■ | 3
// 4 | ■  □  ■  □  ■  □  ■  □ | 4
// 5 | □  ■  □  ■  □  ■  □  ■ | 5
// 6 | ■  □  ■  □  ■  □  ■  □ | 6
// 7 | ♟  ♟  ♟  ♟  ♟  ♟  ♟  ♟ | 7
// 8 | ♜  ♞  ♝  ♛  ♚  ♝  ♞  ♜ | 8
// --+------------------------+--
// ■ | h  g  f  e  d  c  b  a | □
// ```
// Essentially an inverted orientation where black is at the bottom and white at the top.
func RenderASCIIBoardBlack(fen string) (string, error) {
	// Rank order for the black-oriented board is rank1..rank8 from top to bottom.
	ranks := []chess.Rank{
		chess.Rank1, chess.Rank2, chess.Rank3, chess.Rank4,
		chess.Rank5, chess.Rank6, chess.Rank7, chess.Rank8,
	}
	// The file order is reversed: h..a, to create an inverted board.
	files := []chess.File{
		chess.FileH, chess.FileG, chess.FileF, chess.FileE,
		chess.FileD, chess.FileC, chess.FileB, chess.FileA,
	}
	return RenderASCIIBoard(fen, ranks, files, BlackBoard)
}

// RenderASCIIBoard is a helper method that implements the actual ASCII board logic.
// The orientation is determined by which rank/file slices you pass and the orientation constant.
func RenderASCIIBoard(fen string, ranks []chess.Rank, files []chess.File, orientation string) (string, error) {
	// Attempt to parse the provided FEN into a *chess.Game object.
	game, err := parseFEN(fen)
	if err != nil {
		return "", err
	}
	board := game.Position().Board()

	var sb strings.Builder
	header, footer := getHeaderFooter(orientation)

	// Start with a simple fence that we label as ~~~ to help markup.
	sb.WriteString("```\n")
	sb.WriteString(header)
	// This line is just a horizontal separator in the ASCII output.
	sb.WriteString("--+------------------------+--\n")

	// We then loop over each rank or file in the chosen orientation to produce lines of text.
	if orientation == HorizontalBoard {
		// "Horizontal" means we're enumerating files in the outer loop, ranks in the inner loop
		// (a typical 90° board).
		for i, file := range files {
			sb.WriteString(fmt.Sprintf("%s |", string('a'+i)))
			for j := range ranks {
				sq := chess.NewSquare(file, ranks[j])
				piece := board.Piece(sq)
				// Determine if the square is "light" or "dark," used for placeholders.
				sb.WriteString(formatSquare(piece, (i+j)%2 == 0))
			}
			sb.WriteString(fmt.Sprintf("| %s\n", string('a'+i)))
		}
	} else {
		// Standard (White/Black) board layout logic.
		for i, rank := range ranks {
			// "colorRank" is how we label the rank in the left margin for the user.
			// For WhiteBoard, we do 8 down to 1.
			// For BlackBoard, we do 1 up to 8 (depending on how we structured our rank slice).
			colorRank := i + 1
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

	// Bottom border.
	sb.WriteString("--+------------------------+--\n")
	sb.WriteString(footer)
	sb.WriteString("```")
	return sb.String(), nil
}

// getHeaderFooter returns a header line and footer line that label the files (a..h) or (1..8)
// depending on orientation. This helps users see which column is which file/rank in ASCII form.
func getHeaderFooter(orientation string) (string, string) {
	var header, footer, format string
	switch orientation {
	case WhiteBoard:
		// Standard top line: "□ | a b c d e f g h | ■"
		format = "%s | a  b  c  d  e  f  g  h | %s\n"
		header = fmt.Sprintf(format, WhiteCell, BlackCell)
		footer = fmt.Sprintf(format, BlackCell, WhiteCell)
	case HorizontalBoard:
		// For horizontal, we label columns as 1..8 instead of a..h, purely for demonstration.
		format = "%s | 1  2  3  4  5  6  7  8 | %s\n"
		header = fmt.Sprintf(format, BlackCell, WhiteCell)
		footer = fmt.Sprintf(format, WhiteCell, BlackCell)
	case BlackBoard:
		// For black orientation, files h..a in the header/ footer.
		format = "%s | h  g  f  e  d  c  b  a | %s\n"
		header = fmt.Sprintf(format, WhiteCell, BlackCell)
		footer = fmt.Sprintf(format, BlackCell, WhiteCell)
	}
	return header, footer
}

// parseFEN tries to parse a FEN string into a *chess.Game. If fen is empty, returns a new standard game.
func parseFEN(fen string) (*chess.Game, error) {
	if fen == "" {
		return chess.NewGame(), nil
	}
	fenOption, err := chess.FEN(fen)
	if err != nil {
		return nil, fmt.Errorf("invalid FEN: %w", err)
	}
	return chess.NewGame(fenOption), nil
}

// formatSquare prints either the piece symbol or an empty square placeholder (WhiteCell/BlackCell).
// The isWhite bool indicates whether it's a "light" or "dark" square, used for placeholder coloring.
func formatSquare(piece chess.Piece, isWhite bool) string {
	if piece == chess.NoPiece {
		if isWhite {
			return fmt.Sprintf(" %s ", WhiteCell)
		}
		return fmt.Sprintf(" %s ", BlackCell)
	}
	return fmt.Sprintf(" %s ", PieceToStr(piece))
}

// PieceToStr maps chess.Piece objects to Unicode characters (e.g. ♔, ♕).
// Helps produce a more visually recognizable board vs. plain letters (K, Q, R...).
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
		// Fallback: if we somehow get an unknown piece, use p.String().
		return p.String()
	}
}
