package main

// Game represents a Connect Four (Puissance 4) game state.
// - The board size is dynamic (Rows x Cols) for different difficulty levels
// - Board cells store "", "red" or "yellow"
// - Current indicates the player to move; Winner is set once a 4-in-a-row is found
// - PlayerRed/PlayerYellow store display names; Level stores the chosen difficulty
type Game struct {
	Rows, Cols   int
	Board        [][]string // "", "red", "yellow"
	Current      string     // current player color
	Winner       string     // "" if none, else winner color
	PlayerRed    string     // name for red player (player 1)
	PlayerYellow string     // name for yellow player (player 2)
	Level        string     // "easy", "medium", "hard"
	Moves        int        // total number of moves played
	Gravity      string     // "normal" or "inverse"
}

// newGame allocates a new empty game with the given dimensions.
func newGame(rows, cols int) *Game {
	g := &Game{Rows: rows, Cols: cols, Current: "red", Gravity: "normal", Moves: 0}
	g.Board = make([][]string, rows)
	for i := 0; i < rows; i++ {
		g.Board[i] = make([]string, cols)
	}
	return g
}

// columnFull reports whether a column has no available space (no empty cell).
// Note: With alternating gravity, tokens can stack from both ends.
//
//	A column is considered full only when ALL its cells are occupied.
func (g *Game) columnFull(c int) bool {
	if c < 0 || c >= g.Cols || g.Rows == 0 {
		return true
	}
	for r := 0; r < g.Rows; r++ {
		if g.Board[r][c] == "" {
			return false
		}
	}
	return true
}

// drop attempts to place the current player's piece into column c.
// Returns true if a piece was placed; false if invalid or the column is full.
func (g *Game) drop(c int) bool {
	if g.Winner != "" || c < 0 || c >= g.Cols || g.columnFull(c) {
		return false
	}
	var r int
	if g.Gravity == "inverse" {
		// Find the highest empty cell (from top to bottom)
		for r = 0; r < g.Rows; r++ {
			if g.Board[r][c] == "" {
				g.Board[r][c] = g.Current
				break
			}
		}
	} else {
		// Normal gravity: find the lowest empty cell (from bottom to top)
		for r = g.Rows - 1; r >= 0; r-- {
			if g.Board[r][c] == "" {
				g.Board[r][c] = g.Current
				break
			}
		}
	}
	g.checkWin(r, c)
	if g.Winner == "" {
		if g.Current == "red" {
			g.Current = "yellow"
		} else {
			g.Current = "red"
		}
	}
	// Count move and toggle gravity every 5 moves if no winner yet
	g.Moves++
	if g.Winner == "" && g.Moves%5 == 0 {
		if g.Gravity == "normal" {
			g.Gravity = "inverse"
		} else {
			g.Gravity = "normal"
		}
	}
	return true
}

// checkWin inspects four directions through (r,c) to find 4 aligned pieces.
// Directions checked: vertical, horizontal, diagonal ↘, diagonal ↗.
func (g *Game) checkWin(r, c int) {
	color := g.Board[r][c]
	if color == "" {
		return
	}
	dirs := [][2]int{{1, 0}, {0, 1}, {1, 1}, {1, -1}}
	for _, d := range dirs {
		count := 1
		// forward
		rr, cc := r+d[0], c+d[1]
		for rr >= 0 && rr < g.Rows && cc >= 0 && cc < g.Cols && g.Board[rr][cc] == color {
			count++
			rr += d[0]
			cc += d[1]
		}
		// backward
		rr, cc = r-d[0], c-d[1]
		for rr >= 0 && rr < g.Rows && cc >= 0 && cc < g.Cols && g.Board[rr][cc] == color {
			count++
			rr -= d[0]
			cc -= d[1]
		}
		if count >= 4 {
			g.Winner = color
			return
		}
	}
}
