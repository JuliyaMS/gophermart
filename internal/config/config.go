package config

import (
	"flag"
	"os"
)

var RunServerURL string
var DatabaseURI string
var AccrualURL string

func getEnvConfigServer() {
	if envRunSerAddr := os.Getenv("RUN_ADDRESS"); envRunSerAddr != "" {
		RunServerURL = envRunSerAddr
	}
	if BDAddr := os.Getenv("DATABASE_URI"); BDAddr != "" {
		DatabaseURI = BDAddr
	}
	if AccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); AccrualAddr != "" {
		AccrualURL = AccrualAddr
	}
}

func GetServerConfig() {

	flag.StringVar(&RunServerURL, "a", ":8080", "address and port to run server")
	flag.StringVar(&DatabaseURI, "d", "", "database address uri")
	flag.StringVar(&AccrualURL, "r", "", "accrual system address")

	flag.Parse()
	getEnvConfigServer()

}
