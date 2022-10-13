package endpoints

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/iamcathal/lunchMoneyN26/dtos"
	"github.com/iamcathal/lunchMoneyN26/util"
	"gotest.tools/assert"
)

var (
	SERVER_URL_BASE = "http://localhost:9049"
	TESTING_API_KEY = "expectedAPIPassword"
)

func TestMain(m *testing.M) {
	setupAppConfig()
	code := m.Run()
	os.Exit(code)
}

func setupAppConfig() {
	appConfig := dtos.AppConfig{
		APIPassword: TESTING_API_KEY,
	}
	SetConfig(appConfig)
	util.SetConfig(appConfig)
}

func runTestWebServer(ctx context.Context) {
	r := SetupRouter()
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
	req.Header.Set("API_KEY", TESTING_API_KEY)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, res.StatusCode, 200)
}
