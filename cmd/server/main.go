package main

import (
	"log/slog"
	"os"

	"github.com/davidcarrington/eagle-bank/internal/api"
	"github.com/davidcarrington/eagle-bank/internal/config"
	"github.com/davidcarrington/eagle-bank/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	s, err := store.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open store", "error", err)
		os.Exit(1)
	}
	defer s.DB.Close()

	r := api.NewRouter(api.Deps{Store: s, Config: cfg})

	slog.Info("starting server", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
