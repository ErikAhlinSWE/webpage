// main.go
package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type GameResult struct {
	ComputerSelection string `json:"computerSelection"`
	Winner            string `json:"winner"`
	YourSelection     string `json:"yourSelection"`
}

type Stats struct {
	TotalGames int `json:"totalGames"`
	Wins       int `json:"wins"`
}

func main() {
	// Initialize templates
	tmpl, err := template.ParseFiles("templates/game.html")
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	// Serve the main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Proxy endpoint for game
	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		selection := r.URL.Query().Get("yourSelection")

		resp, err := http.Get("http://golangsite1204.chickenkiller.com/api/play?yourSelection=" + selection)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})

	// Proxy endpoint for stats
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("http://golangsite1204.chickenkiller.com/api/stats")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
