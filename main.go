package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type GameResult struct {
	ComputerSelection string `json:"computerSelection"`
	Winner            string `json:"winner"`
	YourSelection     string `json:"yourSelection"`
}

type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func main() {
	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize templates with glob pattern to include all template files
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	// Add security headers middleware
	addSecurityHeaders := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			next(w, r)
		}
	}

	// Serve static files if needed
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route handlers
	http.HandleFunc("/", addSecurityHeaders(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		err := tmpl.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	http.HandleFunc("/game", addSecurityHeaders(func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, "game.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	http.HandleFunc("/customer", addSecurityHeaders(func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, "customer.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	// Game API endpoints
	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		selection := r.URL.Query().Get("yourSelection")
		if selection == "" {
			http.Error(w, "Missing selection parameter", http.StatusBadRequest)
			return
		}

		apiURL := "http://golangsite1204.chickenkiller.com/api/play?yourSelection=" + selection
		log.Printf("Making request to: %s", apiURL)

		resp, err := http.Get(apiURL)
		if err != nil {
			log.Printf("Error making request: %v", err)
			http.Error(w, "Failed to reach game server", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "Error reading response", http.StatusInternalServerError)
			return
		}

		log.Printf("API Response: %s", string(body))

		// Skicka svaret direkt till klienten utan att försöka parsa det
		w.WriteHeader(resp.StatusCode)
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

	// Customer API endpoints
	http.HandleFunc("/api/customers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			resp, err := http.Get("http://pythonsite0115.crabdance.com/api/customer")
			if err != nil {
				http.Error(w, "Failed to reach customer server", http.StatusServiceUnavailable)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Error reading customers", http.StatusInternalServerError)
				return
			}

			w.Write(body)

		case "POST":
			// Läs request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}

			// Skapa en ny reader från body-innehållet
			bodyReader := bytes.NewBuffer(body)

			// Använd bodyReader i POST-anropet
			resp, err := http.Post("http://pythonsite0115.crabdance.com/api/customer",
				"application/json",
				bodyReader)
			if err != nil {
				http.Error(w, "Failed to reach customer server", http.StatusServiceUnavailable)
				return
			}
			defer resp.Body.Close()

			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Error reading response", http.StatusInternalServerError)
				return
			}

			w.Write(responseBody)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/customers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get customer ID from URL
		customerID := filepath.Base(r.URL.Path)

		// Create DELETE request
		req, err := http.NewRequest("DELETE",
			"http://pythonsite0115.crabdance.com/api/customer/"+customerID,
			nil)
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to reach customer server", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Forward response status
		w.WriteHeader(resp.StatusCode)
	})

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
