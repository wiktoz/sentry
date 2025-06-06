package routes

import (
	"net/http"

	"github.com/wiktoz/sentry/db"
	"github.com/wiktoz/sentry/helpers"
	"github.com/wiktoz/sentry/models"
)

func GetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := db.GetConfig(db.DB)
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}
	helpers.WriteJSON(w, config)
}

func UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg models.Config
	if err := helpers.ReadJSON(r.Body, &cfg); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := db.SaveConfig(db.DB, cfg); err != nil {
		http.Error(w, "failed to update config", http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, map[string]string{"status": "updated"})
}
