package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Game struct {
	Board    *Board
	Current  Cell
	Winner   Cell
	GameOver bool
}

var (
	game *Game
	mtx  sync.Mutex
	tmpl *template.Template
)

func main() {
	funcMap := template.FuncMap{
		"seq": func(n int) []int {
			out := make([]int, n)
			for i := 0; i < n; i++ {
				out[i] = i
			}
			return out
		},
	}

	tmpl = template.Must(template.New("").Funcs(funcMap).ParseFiles("templates/game.html"))

	game = &Game{Board: NewBoard(6, 7), Current: P1}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", handleGame)
	http.HandleFunc("/play", handlePlay)
	http.HandleFunc("/reset", handleReset)

	log.Println("Serveur lancé sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleGame(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	defer mtx.Unlock()
	tmpl.ExecuteTemplate(w, "game.html", game)
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	colStr := r.FormValue("col")
	col, err := strconv.Atoi(colStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	mtx.Lock()
	defer mtx.Unlock()

	if game.GameOver {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if col < 0 || col >= game.Board.Cols {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if ok := game.Board.Drop(col, game.Current); !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	winner := game.Board.CheckWin()
	if winner != Empty {
		game.Winner = winner
		game.GameOver = true
	} else if game.Board.IsFull() {
		game.GameOver = true
	} else {
		if game.Current == P1 {
			game.Current = P2
		} else {
			game.Current = P1
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	game = &Game{Board: NewBoard(6, 7), Current: P1}
	mtx.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
