package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Game struct {
	Id                int
	CreatedAt         time.Time
	Winner            string `json:"winner"`
	YourSelection     string `json:"yourSelection"`
	ComputerSelection string `json:"computerSelection"`
}

type Customer struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	City                 string `json:"city"`
	TelephoneCountryCode string `json:"TelephoneCountryCode"`
	Telephone            string `json:"Telephone"`
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

	// Debug: Lista alla mallar som laddats
	for _, t := range tmpl.Templates() {
		log.Println("Loaded template:", t.Name())
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

		// Trimma bort eventuella prefix ("You", "Computer", "Tie")
		trimmedBody := strings.TrimPrefix(string(body), "You")
		trimmedBody = strings.TrimPrefix(trimmedBody, "Computer")
		trimmedBody = strings.TrimPrefix(trimmedBody, "Tie")

		var result Game
		err = json.Unmarshal([]byte(trimmedBody), &result)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			http.Error(w, "Error parsing response", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)
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
			// Fetch all customers from the specified URL and forward the response to the client
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
			// Read the request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}

			// Create a new reader from the body content
			bodyReader := bytes.NewBuffer(body)

			// Use bodyReader in the POST request
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
		customerID := filepath.Base(r.URL.Path)

		switch r.Method {
		case "DELETE":
			// Create DELETE request
			req, err := http.NewRequest("DELETE",
				"http://pythonsite0115.crabdance.com/api/customer/"+customerID,
				nil)
			if err != nil {
				log.Printf("Error creating request: %v", err)
				http.Error(w, "Error creating request", http.StatusInternalServerError)
				return
			}

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error: %v", err)
				http.Error(w, "Failed to reach customer server", http.StatusServiceUnavailable)
				return
			}
			defer resp.Body.Close()

			// Log response status
			log.Printf("Response Status: %d", resp.StatusCode)

			// Forward response status
			w.WriteHeader(resp.StatusCode)

			// Log response body
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body: %v", err)
				http.Error(w, "Error reading response", http.StatusInternalServerError)
				return
			}
			log.Printf("Response Body: %s", string(responseBody))
			w.Write(responseBody)

		case "PUT":
			// Read the request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}

			// Create a new reader from the body content
			bodyReader := bytes.NewBuffer(body)

			// Use bodyReader in the PUT request
			req, err := http.NewRequest("PUT",
				"http://pythonsite0115.crabdance.com/api/customer/"+customerID,
				bodyReader)
			if err != nil {
				log.Printf("Error creating request: %v", err)
				http.Error(w, "Error creating request", http.StatusInternalServerError)
				return
			}

			// Set the Content-Type header
			req.Header.Set("Content-Type", "application/json")

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error: %v", err)
				http.Error(w, "Failed to reach customer server", http.StatusServiceUnavailable)
				return
			}
			defer resp.Body.Close()

			// Log response status
			log.Printf("Response Status: %d", resp.StatusCode)

			// Forward response status
			w.WriteHeader(resp.StatusCode)

			// Log response body
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body: %v", err)
				http.Error(w, "Error reading response", http.StatusInternalServerError)
				return
			}
			log.Printf("Response Body: %s", string(responseBody))
			w.Write(responseBody)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
