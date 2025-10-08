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

const (
    rows = 6
    cols = 7
)

type Game struct {
    Board   [rows][cols]string // "", "red", "yellow"
    Current string             // current player color
    Winner  string             // "" if none, else winner color
}

func newGame() *Game {
    return &Game{Current: "red"}
}

func (g *Game) columnFull(c int) bool {
    return g.Board[0][c] != ""
}

func (g *Game) drop(c int) bool {
    if g.Winner != "" || c < 0 || c >= cols || g.columnFull(c) {
        return false
    }
    var r int
    for r = rows - 1; r >= 0; r-- {
        if g.Board[r][c] == "" {
            g.Board[r][c] = g.Current
            break
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
    return true
}

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
        for rr >= 0 && rr < rows && cc >= 0 && cc < cols && g.Board[rr][cc] == color {
            count++
            rr += d[0]
            cc += d[1]
        }
        // backward
        rr, cc = r-d[0], c-d[1]
        for rr >= 0 && rr < rows && cc >= 0 && cc < cols && g.Board[rr][cc] == color {
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

// session store (in-memory)
type store struct {
    mu    sync.Mutex
    games map[string]*Game
}

func newStore() *store {
    return &store{games: make(map[string]*Game)}
}

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
    g := newGame()
    s.games[sid] = g
    return sid, g
}

var (
    tpl      *template.Template
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
    RowsIdx     []int
    ColsIdx     []int
    Board       [][]string
    Current     string
    Winner      string
    ColumnFull  []bool
    GameOver    bool
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
    // Get or create session
    var sid string
    if c, err := r.Cookie("session_id"); err == nil {
        sid = c.Value
    }
    sid, g := gameStore.getOrCreate(sid)
    // ensure cookie set
    http.SetCookie(w, &http.Cookie{Name: "session_id", Value: sid, Path: "/"})

    switch r.Method {
    case http.MethodGet:
        // no-op: just render
    case http.MethodPost:
        if err := r.ParseForm(); err == nil {
            if r.Form.Get("reset") != "" {
                // reset game
                gameStore.mu.Lock()
                gameStore.games[sid] = newGame()
                g = gameStore.games[sid]
                gameStore.mu.Unlock()
            } else {
                colStr := r.Form.Get("col")
                if colStr != "" {
                    if c, err := strconv.Atoi(colStr); err == nil {
                        g.drop(c)
                    }
                }
            }
        }
    default:
        http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
        return
    }

    // Build view data
    rowsIdx := make([]int, rows)
    for i := 0; i < rows; i++ { rowsIdx[i] = i }
    colsIdx := make([]int, cols)
    for j := 0; j < cols; j++ { colsIdx[j] = j }
    board := make([][]string, rows)
    for i := 0; i < rows; i++ {
        board[i] = make([]string, cols)
        for j := 0; j < cols; j++ {
            board[i][j] = g.Board[i][j]
        }
    }
    colFull := make([]bool, cols)
    for j := 0; j < cols; j++ { colFull[j] = g.columnFull(j) }

    data := playView{
        RowsIdx: rowsIdx,
        ColsIdx: colsIdx,
        Board: board,
        Current: g.Current,
        Winner: g.Winner,
        ColumnFull: colFull,
        GameOver: g.Winner != "",
    }

    if err := tpl.ExecuteTemplate(w, "play.html", data); err != nil {
        http.Error(w, "Erreur de rendu", http.StatusInternalServerError)
    }
}
