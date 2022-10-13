package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	statusForbiddenResponse string
)

func TestMain(m *testing.M) {
	setupAppConfig()
	setupData()

	ctx, cancel := context.WithCancel(context.Background())
	go runTestWebServer(ctx)
	time.Sleep(2 * time.Millisecond)

	code := m.Run()

	cancel()
	os.Exit(code)
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

func TestGetAPIStatusReturnsStatusWhenCorrectAPIKeyIsGiven(t *testing.T) {
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

func TestGetAPIStatusReturnsStatusForbiddenWhenIncorrectAPIKeyIsGiven(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprint(SERVER_URL_BASE+"/status"), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("API_KEY", "incorrect API key")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, res.StatusCode, http.StatusForbidden)
	assert.Equal(t, string(body), statusForbiddenResponse)
}

func setupAppConfig() {
	appConfig := dtos.AppConfig{
		APIPassword: TESTING_API_KEY,
	}
	SetConfig(appConfig)
	util.SetConfig(appConfig)
}

func setupData() {
	statusForbiddenResponseStruct := struct {
		Error string `json:"error"`
	}{
		"You are not authorized to access this endpoint",
	}
	bytes, err := json.Marshal(&statusForbiddenResponseStruct)
	if err != nil {
		panic(err)
	}
	statusForbiddenResponse = string(bytes) + "\n"
}
