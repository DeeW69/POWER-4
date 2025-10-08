package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	registerGameRoutes(mux)

	// Serve static assets like CSS from the /static/ path.
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "home.html")
	})

	mux.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "play.html")
	})

	addr := "127.0.0.1:8080"
	log.Printf("Serveur en cours d'exécution sur http://%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("erreur serveur: %v", err)
	}
}
