package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	SERVER_URL_BASE = "http://localhost:9049"
	API_KEY         = "expectedAPIPassword"
)

func TestMain(m *testing.M) {
	setupProps()
	code := m.Run()
	os.Exit(code)
}

func setupProps() {
	os.Setenv("API_KEY", API_KEY)
}

func runTestWebServer(ctx context.Context) {
	r := setupRouter()
	srv := &http.Server{
		Handler:      r,
		Addr:         ":9049",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func TestGetAPIStatus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runTestWebServer(ctx)
	time.Sleep(2 * time.Millisecond)

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprint(SERVER_URL_BASE+"/status"), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("API_KEY", API_KEY)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, res.StatusCode, 200)
}
