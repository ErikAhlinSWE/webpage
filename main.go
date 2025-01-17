// main.go
package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type GameResult struct {
	ComputerSelection string `json:"computerSelection"`
	Winner            string `json:"winner"`
	YourSelection     string `json:"yourSelection"`
}

func main() {
	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize templates
	tmpl, err := template.ParseFiles("templates/game.html")
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	// Serve the main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Proxy endpoints with additional error handling
	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		selection := r.URL.Query().Get("yourSelection")
		if selection == "" {
			http.Error(w, "Missing selection parameter", http.StatusBadRequest)
			return
		}

		resp, err := http.Get("http://golangsite1204.chickenkiller.com/api/play?yourSelection=" + selection)
		if err != nil {
			http.Error(w, "Failed to reach game server", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response", http.StatusInternalServerError)
			return
		}

		w.Write(body)
	})

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		resp, err := http.Get("http://golangsite1204.chickenkiller.com/api/stats")
		if err != nil {
			http.Error(w, "Failed to reach stats server", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading stats", http.StatusInternalServerError)
			return
		}

		w.Write(body)
	})

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
