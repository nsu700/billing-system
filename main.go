package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	var err error

	// Get the path of the SQLite database file from the environment variable
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./mydb.db"
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS mytable (id INTEGER PRIMARY KEY AUTOINCREMENT, date TEXT, amount INTEGER, type TEXT, purpose TEXT)")
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	http.HandleFunc("/submit", submitHandler)
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
	purposes := r.Form["purpose[]"]

	// Loop through the arrays of values for each input field
	for i := 0; i < len(dates); i++ {
		// Insert the values into the database
		stmt, err := db.Prepare("INSERT INTO mytable (date, amount, type, purpose) VALUES (?, ?, ?, ?)")
		if err != nil {
			log.Printf("Failed to prepare statement: %v", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(dates[i], amounts[i], types[i], purposes[i])
		if err != nil {
			log.Panicf("Failed to execute statement: %v", err)
			return
		}
	}

	// Respond with a success message
	fmt.Fprintln(w, "Form submission received and written to database!")
	log.Println("Form submission received and written to database.")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
