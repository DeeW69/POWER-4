package main

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// store is a naive in-memory session store mapping session IDs to games.
// It is protected by a mutex because handlers may access it concurrently.
type store struct {
	mu    sync.Mutex
	games map[string]*Game
}

func newStore() *store {
	return &store{games: make(map[string]*Game)}
}

// getOrCreate returns the game for a session id if it exists.
// Otherwise, it creates a fresh game (default level is easy 6x7) and a new session id.
func (s *store) getOrCreate(id string) (string, *Game) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != "" {
		if g, ok := s.games[id]; ok {
			return id, g
		}
	}
	// create new session
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	sid := hex.EncodeToString(buf)
	g := newGame(6, 7)
	g.Level = "easy"
	s.games[sid] = g
	return sid, g
}

var (
	tpl       *template.Template
	gameStore = newStore()
)

func main() {
	// Parse templates (home.html, play.html)
	var err error
	tpl, err = template.ParseFiles("home.html", "play.html")
	if err != nil {
		log.Fatalf("erreur template: %v", err)
	}

	mux := http.NewServeMux()

	// Static assets (CSS)
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/play", handlePlay)

	addr := "127.0.0.1:8080"
	log.Printf("Serveur en cours d'exécution sur http://%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("erreur serveur: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	// Home is a simple GET that renders the index with players + level form.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	// Render home.html as a template (no dynamic data needed)
	if err := tpl.ExecuteTemplate(w, "home.html", nil); err != nil {
		http.Error(w, "Erreur de rendu", http.StatusInternalServerError)
	}
}

type playView struct {
	// Precomputed indices for template loops (rows/cols)
	RowsIdx []int
	ColsIdx []int
	// Board to render
	Board [][]string
	// Game state
	Current    string
	Winner     string
	ColumnFull []bool
	GameOver   bool
	// Player display names
	PlayerRed    string
	PlayerYellow string
	// Human-friendly names for current / winner / loser
	CurrentName string
	WinnerName  string
	LoserName   string
	// Board dimensions & level for CSS class binding
	Rows  int
	Cols  int
	Level string
	// Gravity state and move counters
	Gravity  string
	Moves    int
	NextFlip int
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
	// Get or create session
	var sid string
	if c, err := r.Cookie("session_id"); err == nil {
		sid = c.Value
	}
	sid, g := gameStore.getOrCreate(sid)
	// Ensure cookie set
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: sid, Path: "/"})

	switch r.Method {
	case http.MethodGet:
		// Capture optional player names and level from query string
		q := r.URL.Query()
		p1 := q.Get("player1")
		p2 := q.Get("player2")
		level := q.Get("level")
		if p1 != "" || p2 != "" {
			gameStore.mu.Lock()
			if p1 != "" {
				g.PlayerRed = p1
			}
			if p2 != "" {
				g.PlayerYellow = p2
			}
			gameStore.mu.Unlock()
		}
		if level != "" {
			rows, cols := g.Rows, g.Cols
			switch level {
			case "easy":
				rows, cols = 6, 7
			case "medium":
				rows, cols = 7, 8
			case "hard":
				rows, cols = 8, 9
			}
			// Rebuild game only if dimensions or level changed; preserve names
			if rows != g.Rows || cols != g.Cols || g.Level != level {
				gameStore.mu.Lock()
				red, yellow := g.PlayerRed, g.PlayerYellow
				g = newGame(rows, cols)
				g.PlayerRed, g.PlayerYellow = red, yellow
				g.Level = level
				gameStore.games[sid] = g
				gameStore.mu.Unlock()
			}
		}
	case http.MethodPost:
		if err := r.ParseForm(); err == nil {
			if r.Form.Get("reset") != "" {
				// Reset game: keep names and level, clear board and winner/current
				gameStore.mu.Lock()
				oldRed, oldYellow := g.PlayerRed, g.PlayerYellow
				rows, cols := g.Rows, g.Cols
				level := g.Level
				gameStore.games[sid] = newGame(rows, cols)
				g = gameStore.games[sid]
				g.PlayerRed, g.PlayerYellow = oldRed, oldYellow
				g.Level = level
				gameStore.mu.Unlock()
			} else {
				colStr := r.Form.Get("col")
				if colStr != "" {
					if c, err := strconv.Atoi(colStr); err == nil {
						// Apply a move in the clicked column
						g.drop(c)
					}
				}
			}
		}
	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Build view data for the template
	rowsIdx := make([]int, g.Rows)
	for i := 0; i < g.Rows; i++ {
		rowsIdx[i] = i
	}
	colsIdx := make([]int, g.Cols)
	for j := 0; j < g.Cols; j++ {
		colsIdx[j] = j
	}
	board := g.Board
	colFull := make([]bool, g.Cols)
	for j := 0; j < g.Cols; j++ {
		colFull[j] = g.columnFull(j)
	}

	// Helper closure to compute display names for colors
	nameFor := func(color string) string {
		switch color {
		case "red":
			if g.PlayerRed != "" {
				return g.PlayerRed
			}
			return "joueur rouge"
		case "yellow":
			if g.PlayerYellow != "" {
				return g.PlayerYellow
			}
			return "joueur jaune"
		default:
			return ""
		}
	}

	data := playView{
		RowsIdx:      rowsIdx,
		ColsIdx:      colsIdx,
		Board:        board,
		Current:      g.Current,
		Winner:       g.Winner,
		ColumnFull:   colFull,
		GameOver:     g.Winner != "",
		PlayerRed:    g.PlayerRed,
		PlayerYellow: g.PlayerYellow,
		Rows:         g.Rows,
		Cols:         g.Cols,
		Level:        g.Level,
		Gravity:      g.Gravity,
		Moves:        g.Moves,
	}

	// Fill friendly names for the template
	data.CurrentName = nameFor(g.Current)
	if g.Winner != "" {
		data.WinnerName = nameFor(g.Winner)
		loser := "red"
		if g.Winner == "red" {
			loser = "yellow"
		}
		data.LoserName = nameFor(loser)
	}

	// Compute turns until the next gravity flip (1..5)
	rem := 5 - (g.Moves % 5)
	if rem == 0 {
		rem = 5
	}
	data.NextFlip = rem

	if err := tpl.ExecuteTemplate(w, "play.html", data); err != nil {
		http.Error(w, "Erreur de rendu", http.StatusInternalServerError)
	}
}
