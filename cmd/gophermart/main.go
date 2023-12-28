package main

import (
	"github.com/JuliyaMS/gophermart/internal/accrual"
	"github.com/JuliyaMS/gophermart/internal/config"
	"github.com/JuliyaMS/gophermart/internal/logger"
	"github.com/JuliyaMS/gophermart/internal/server"
	"github.com/JuliyaMS/gophermart/internal/storage"
)

func main() {
	lg := logger.NewLogger()

	lg.Infow("Get server config")
	config.GetServerConfig()

	lg.Infow("Get connection to database")
	storageDB := storage.NewConnectionDB(lg)

	lg.Infow("Create new handlers")
	handlers := server.NewHandlers(lg, storageDB)

	lg.Infow("Create new router")
	router := server.NewRouter(handlers)

	lg.Infow("Create new accrual system")
	acc := accrual.NewSystemAccrual(lg, 3)
	go acc.Start()

	defer router.Close()
	defer acc.Close()

	lg.Infow("Create new server")
	s := server.NewServer(lg, router)
	s.Start()
}
