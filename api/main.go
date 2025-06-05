package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/wiktoz/sentry/db"
	"github.com/wiktoz/sentry/routes"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var err error
	db.DB, err = sql.Open("sqlite3", "./results.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	schema, err := os.ReadFile("./sql/db_init.sql")
	if err != nil {
		log.Fatalf("failed to read schema file: %v", err)
	}
	if _, err = db.DB.Exec(string(schema)); err != nil {
		log.Fatalf("failed to exec schema: %v", err)
	}

	// API routes
	http.HandleFunc("/api/scan/run", routes.RunScan)
	http.HandleFunc("/api/scan/", routes.GetScanById)
	http.HandleFunc("/api/scans", routes.GetScans)

	// React static files from ./web/dist
	fs := http.FileServer(http.Dir("./web/dist"))
	http.Handle("/", fs)

	srv := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server started on :8080")
	log.Fatal(srv.ListenAndServe())
}
