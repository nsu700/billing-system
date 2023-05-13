package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"text/template"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type Spending struct {
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
}

func main() {
	var err error

	// Get the path of the SQLite database file from the environment variable
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./billing-system.db"
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS mytable (id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT, amount INTEGER, type TEXT, description TEXT)")
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/add", addDataHandler)
	http.HandleFunc("/", indexHandler)

	log.Println("Server listening on port 8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Failed to parse form data: %v", err)
		return
	}

	// Get the values from the form
	dates := r.Form["date[]"]
	amounts := r.Form["amount[]"]
	types := r.Form["type[]"]
	descriptions := r.Form["description[]"]

	// Loop through the arrays of values for each input field
	for i := 0; i < len(dates); i++ {
		// Insert the values into the database
		stmt, err := db.Prepare("INSERT INTO mytable (date, amount, type, description) VALUES (?, ?, ?, ?)")
		if err != nil {
			log.Printf("Failed to prepare statement: %v", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(dates[i], amounts[i], types[i], descriptions[i])
		if err != nil {
			log.Panicf("Failed to execute statement: %v", err)
			return
		}
	}

	// Respond with a success message
	http.Redirect(w, r, "/", http.StatusFound)
	log.Println("Form submission received and written to database.")
}

func addDataHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "submit-form.html")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Check if there is any spending data in the database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM mytable").Scan(&count)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		http.Error(w, "Failed to fetch spending data", http.StatusInternalServerError)
		return
	}

	if count == 0 {
		// If there is no spending data, display the submit page
		http.ServeFile(w, r, "submit-form.html")
		return
	}

	// Fetch the spending data from the database
	rows, err := db.Query("SELECT date, amount, type, description FROM mytable")
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		http.Error(w, "Failed to fetch spending data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var spendings []Spending
	for rows.Next() {
		var spending Spending
		err := rows.Scan(&spending.Date, &spending.Amount, &spending.Type, &spending.Description)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		spendings = append(spendings, spending)
	}
	err = rows.Err()
	if err != nil {
		log.Printf("Error while iterating rows: %v", err)
		http.Error(w, "Failed to fetch spending data", http.StatusInternalServerError)
		return
	}

	// Parse the HTML template
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Printf("Failed to parse HTML template: %v", err)
		http.Error(w, "Failed to parse HTML template", http.StatusInternalServerError)
		return
	}

	// Execute the HTML template with the spending data
	err = tmpl.Execute(w, spendings)
	if err != nil {
		log.Printf("Failed to execute HTML template: %v", err)
		http.Error(w, "Failed to execute HTML template", http.StatusInternalServerError)
		return
	}
}
