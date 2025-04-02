package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"frappuccino/internal/handlers"
	"frappuccino/internal/service"
)

type server struct {
	port   string
	db     *sql.DB
	logger *slog.Logger
}

func NewServer(port string, db *sql.DB, logger *slog.Logger) *server {
	return &server{
		port:   port,
		db:     db,
		logger: logger,
	}
}

func (s *server) RunServer() {
	app := handlers.NewApplication(s.logger,
		service.NewInventoryService(s.db, s.logger),
		service.NewMenuService(s.db, s.logger),
		service.NewOrderService(s.db, s.logger),
		service.NewReportService(s.db, s.logger),
	)

	srv := &http.Server{
		Addr:     s.port,
		Handler:  app.Routes(),
		ErrorLog: slog.NewLogLogger(s.logger.Handler(), slog.LevelError),
	}

	s.logger.Info("starting server", "addr", srv.Addr)

	err := srv.ListenAndServe()
	s.logger.Error(err.Error())
	os.Exit(1)
}
