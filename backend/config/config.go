package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/iamcathal/lunchMoneyN26/dtos"
)

func InitConfig(args []string) dtos.AppConfig {
	ensureAllEnvVarsAreSet()

	appConfig := dtos.AppConfig{
		WebServer:       false,
		Port:            2944,
		N26Username:     os.Getenv("N26_USERNAME"),
		N26Password:     os.Getenv("N26_PASSWORD"),
		N26DeviceToken:  os.Getenv("N26_DEVICE_TOKEN"),
		LunchMoneyToken: os.Getenv("LUNCHMONEY_TOKEN"),
		APIPassword:     os.Getenv("API_KEY"),

		Days: 0,

		ApplicationStartUpTime: time.Now(),
	}
	if os.Getenv("API_PORT") != "" {
		intAPIPort, err := strconv.Atoi(os.Getenv("API_PORT"))
		if err != nil {
			panic(err)
		}
		appConfig.Port = intAPIPort
	}

	webServer := flag.Bool("webserver", false, "Run the application as a webserver")
	days := flag.Int("d", 1, "Search for the last n days for any transactions (if not in webserver mode)")
	flag.Parse()

	if *webServer {
		appConfig.WebServer = true
	} else {
		appConfig.Days = *days
	}
	return appConfig
}

func ensureAllEnvVarsAreSet() {
	requiredEnvVars := []string{
		"N26_USERNAME",
		"N26_PASSWORD",
		"N26_DEVICE_TOKEN",
		"LUNCHMONEY_TOKEN",
		"API_KEY",
	}

	for _, envVar := range requiredEnvVars {
		if isSet := envVarIsSet(envVar); !isSet {
			panic(fmt.Sprint("env var " + envVar + " is not set"))
		}
	}
}

func envVarIsSet(varName string) bool {
	if _, exists := os.LookupEnv(varName); exists {
		return true
	}
	return false
}
