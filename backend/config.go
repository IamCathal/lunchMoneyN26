package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func initConfig(args []string) {
	ApplicationStartUpTime = time.Now()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ensureAllEnvVarsAreSet()

	appConfig := appConfig{
		webServer:       false,
		port:            2944,
		n26Username:     os.Getenv("N26_USERNAME"),
		n26Password:     os.Getenv("N26_PASSWORD"),
		n26DeviceToken:  os.Getenv("N26_DEVICE_TOKEN"),
		lunchMoneyToken: os.Getenv("LUNCHMONEY_TOKEN"),
		APIPassword:     os.Getenv("API_KEY"),

		days: 0,
	}
	if os.Getenv("API_PORT") != "" {
		intAPIPort, err := strconv.Atoi(os.Getenv("API_PORT"))
		if err != nil {
			panic(err)
		}
		appConfig.port = intAPIPort
	}

	webServer := flag.Bool("webserver", false, "Run the application as a webserver")
	days := flag.Int("d", 1, "Search for the last n days for any transactions (if not in webserver mode)")
	flag.Parse()

	if *webServer {
		appConfig.webServer = true
	} else {
		appConfig.days = *days
	}
	config = appConfig
}
