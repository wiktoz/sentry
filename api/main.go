package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/wiktoz/sentry/db"
	"github.com/wiktoz/sentry/routes"
	"github.com/wiktoz/sentry/scripts"

	_ "modernc.org/sqlite"
)

func main() {
	// DB setup
	var err error
	db.DB, err = sql.Open("sqlite", "./results.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	if _, err = db.DB.Exec(db.Schema); err != nil {
		log.Fatalf("failed to exec schema: %v", err)
	}

	// Start auto scan
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scripts.StartAutoScan(ctx)

	// HTTP API
	http.HandleFunc("/api/scan/run", routes.RunScan)
	http.HandleFunc("/api/scan/", routes.GetScanById)
	http.HandleFunc("/api/scans", routes.GetScans)

	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			routes.GetConfig(w, r)
		case http.MethodPut:
			routes.UpdateConfig(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Static files
	fs := http.FileServer(http.Dir("./web/dist"))
	http.Handle("/", fs)

	srv := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server started on :8080")
	log.Fatal(srv.ListenAndServe())
}
