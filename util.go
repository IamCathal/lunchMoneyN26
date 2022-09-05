package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func Status(w http.ResponseWriter, r *http.Request) {
	req := uptimeResponse{
		Status: "operational",
	}
	jsonObj, err := json.MarshalIndent(req, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(string(jsonObj))
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%v %+v\n", time.Now().Format(time.RFC3339), r)
		next.ServeHTTP(w, r)
	})
}

func ensureAllEnvVarsAreSet() {
	requiredEnvVars := []string{
		"N26_USERNAME",
		"N26_PASSWORD",
		"N26_DEVICE_TOKEN",
		"LUNCHMONEY_TOKEN",
	}

	for _, envVar := range requiredEnvVars {
		if isSet := envVarIsSet(envVar); !isSet {
			panic(fmt.Sprint("env var " + envVar + " is not set"))
		}
	}
}

func getEnvVarOrDefault(varName, defaultVal string) string {
	if isSet := envVarIsSet(varName); isSet {
		return os.Getenv(varName)
	}
	return defaultVal
}

func envVarIsSet(varName string) bool {
	if _, exists := os.LookupEnv(varName); exists {
		return true
	}
	return false
}
