package main

type Cell int

const (
	Empty Cell = 0
	P1    Cell = 1
	P2    Cell = 2
)

type Board struct {
	Rows int
	Cols int
	Grid [][]Cell
}

func NewBoard(rows, cols int) *Board {
	g := make([][]Cell, rows)
	for i := range g {
		g[i] = make([]Cell, cols)
	}
	return &Board{Rows: rows, Cols: cols, Grid: g}
}

func (b *Board) Drop(col int, player Cell) bool {
	for r := b.Rows - 1; r >= 0; r-- {
		if b.Grid[r][col] == Empty {
			b.Grid[r][col] = player
			return true
		}
	}
	return false
}

func (b *Board) CheckWin() Cell {
	for r := 0; r < b.Rows; r++ {
		for c := 0; c < b.Cols; c++ {
			player := b.Grid[r][c]
			if player == Empty {
				continue
			}
			// 4 directions : droite, bas, diag bas droite, diag haut droite
			dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {-1, 1}}
			for _, d := range dirs {
				count := 1
				for i := 1; i < 4; i++ {
					nr := r + d[0]*i
					nc := c + d[1]*i
					if nr < 0 || nr >= b.Rows || nc < 0 || nc >= b.Cols {
						break
					}
					if b.Grid[nr][nc] == player {
						count++
					} else {
						break
					}
				}
				if count >= 4 {
					return player
				}
			}
		}
	}
	return Empty
}

func (b *Board) IsFull() bool {
	for _, row := range b.Grid {
		for _, cell := range row {
			if cell == Empty {
				return false
			}
		}
	}
	return true
}
