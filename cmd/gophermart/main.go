package main

import (
	"context"
	"github.com/JuliyaMS/gophermart/internal/accrual"
	"github.com/JuliyaMS/gophermart/internal/config"
	"github.com/JuliyaMS/gophermart/internal/logger"
	"github.com/JuliyaMS/gophermart/internal/server"
	"github.com/JuliyaMS/gophermart/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	lg := logger.NewLogger()

	lg.Infow("Get server config")
	config.GetServerConfig()

	lg.Infow("Create connection to database")
	conn, err := pgxpool.New(context.Background(), config.DatabaseURI)
	if err != nil {
		lg.Error("Get error while create connection to database: ", err)
	}

	defer conn.Close()

	lg.Infow("Get connection to database")
	storageDB := storage.NewConnectionDB(conn, lg)

	lg.Infow("Create new handlers")
	handlers := server.NewHandlers(lg, storageDB)

	lg.Infow("Create new router")
	router := server.NewRouter(handlers)

	lg.Infow("Create new accrual system")
	acc := accrual.NewSystemAccrual(lg, 3)
	go acc.Start()

	lg.Infow("Create new server")
	s := server.NewServer(lg, router)
	s.Start()
}
