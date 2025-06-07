package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/wiktoz/sentry/db"
	"github.com/wiktoz/sentry/routes"
	"github.com/wiktoz/sentry/scripts"

	_ "modernc.org/sqlite"
)

func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // change to specific origin in production
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func basicAuthMiddleware(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers here so they are included on every response
		enableCORS(w, r)

		// Allow OPTIONS requests without auth for CORS preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		auth := r.Header.Get("Authorization")
		const prefix = "Basic "
		if !strings.HasPrefix(auth, prefix) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || pair[0] != username || pair[1] != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Credentials for basic auth (hardcoded here, replace with env/config for production)
	const (
		authUsername = "admin"
		authPassword = "secret"
	)

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

	// Wrap handlers with CORS and BasicAuth
	apiMux := http.NewServeMux()

	apiMux.Handle("/api/scan/run", withCORS(http.HandlerFunc(routes.RunScan)))
	apiMux.Handle("/api/scan/", withCORS(http.HandlerFunc(routes.GetScanById)))
	apiMux.Handle("/api/scans", withCORS(http.HandlerFunc(routes.GetScans)))

	apiMux.Handle("/api/config", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			routes.GetConfig(w, r)
		case http.MethodPut:
			routes.UpdateConfig(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Static files
	fs := http.FileServer(http.Dir("./web/dist"))

	mainMux := http.NewServeMux()
	mainMux.Handle("/api/", apiMux) // apiMux has no withCORS wrapping now
	mainMux.Handle("/", fs)

	// Wrap with CORS and basic auth in order
	protectedMux := withCORS(basicAuthMiddleware(mainMux, authUsername, authPassword))

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           protectedMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server started on :8080")
	log.Fatal(srv.ListenAndServe())
}
