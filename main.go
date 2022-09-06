package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/guitmz/n26"
	"github.com/joho/godotenv"
)

var (
	config appConfig
)

func getClient() *n26.Client {
	fmt.Println("waiting for 2FA confirmation in app")
	newClient, err := n26.NewClient(n26.Auth{
		UserName:    config.n26Username,
		Password:    config.n26Password,
		DeviceToken: config.n26DeviceToken,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("auth complete")
	return newClient
}

func Transactions(w http.ResponseWriter, r *http.Request) {
	fullDaysToLookupString := r.URL.Query().Get("days")
	daysToLookup, err := strconv.Atoi(fullDaysToLookupString)
	if err != nil {
		panic(err)
	}

	transactions := getAndFilterTransactions(daysToLookup)
	uploadTransactions(uploadTransactionsDTO{transactions, true, true, true})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

func uploadTransactions(transactions uploadTransactionsDTO) {
	jsonObj, err := json.Marshal(transactions)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "https://dev.lunchmoney.app/v1/transactions", bytes.NewBuffer(jsonObj))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("LUNCHMONEY_TOKEN")))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("=====\n%v\n======\n", string(body))
}

func initConfig(args []string) appConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ensureAllEnvVarsAreSet()

	appConfig := appConfig{
		offlineMode:     false,
		port:            2944,
		n26Username:     os.Getenv("N26_USERNAME"),
		n26Password:     os.Getenv("N26_PASSWORD"),
		n26DeviceToken:  os.Getenv("N26_DEVICE_TOKEN"),
		lunchMoneyToken: os.Getenv("LUNCHMONEY_TOKEN"),

		days: 0,
	}
	if os.Getenv("API_PORT") != "" {
		intAPIPort, err := strconv.Atoi(os.Getenv("API_PORT"))
		if err != nil {
			panic(err)
		}
		appConfig.port = intAPIPort
	}

	offlineMode := flag.Bool("x", false, "use the application not as a webserver")
	days := flag.Int("d", 1, "how many days of transactions should be queried")
	flag.Parse()

	if *offlineMode {
		config.offlineMode = true
		appConfig.days = *days
	}

	return appConfig
}

func runWebServer(config appConfig) {
	r := mux.NewRouter()
	r.HandleFunc("/status", Status).Methods("POST")
	r.HandleFunc("/anycans", Transactions).Methods("POST")
	r.Use(logMiddleware)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + fmt.Sprint(config.port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	fmt.Printf("Serving requests on :%d\n", config.port)
	log.Fatal(srv.ListenAndServe())
}

func main() {
	config = initConfig(os.Args[1:])

	if !config.offlineMode {
		runWebServer(config)
		return
	}

	transactions := getAndFilterTransactions(config.days)
	uploadTransactions(uploadTransactionsDTO{transactions, true, true, true})
}
