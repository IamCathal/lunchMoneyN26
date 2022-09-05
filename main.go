package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	client := getClient()

	fullDaysToLookupString := r.URL.Query().Get("days")
	daysToLookup, err := strconv.Atoi(fullDaysToLookupString)
	if err != nil {
		panic(err)
	}

	endTime := n26.TimeStamp{Time: time.Now()}
	startTime := n26.TimeStamp{Time: endTime.Time.Add((-time.Hour * 24) * time.Duration(daysToLookup))}

	transactions, err := client.GetTransactions(startTime, endTime, fmt.Sprint(daysToLookup))
	if err != nil {
		panic(err)
	}

	filteredTransactions := []filteredTransaction{}
	for _, transaction := range *transactions {
		currTransaction := filteredTransaction{
			ID:       transaction.ID,
			Date:     transaction.VisibleTS.Time.Format(time.RFC3339),
			Amount:   transaction.Amount,
			Currency: strings.ToLower(transaction.OriginalCurrency),
		}

		// If the transaction was from a friend
		if transaction.PartnerIban != "" {
			currTransaction.Payee = transaction.PartnerName
			currTransaction.Category = "friends"
		} else {
			// or from a business
			currTransaction.Payee = transaction.MerchantName
			currTransaction.Category = transaction.Category
		}
		filteredTransactions = append(filteredTransactions, currTransaction)
	}

	jsonString, err := json.MarshalIndent(filteredTransactions, "", "\t")
	if err != nil {
		panic(err)
	}

	uploadTransactions(uploadTransactionsDTO{filteredTransactions, true, true, true})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(string(jsonString))
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

func initConfig() appConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ensureAllEnvVarsAreSet()

	appConfig := appConfig{
		port:            2944,
		n26Username:     os.Getenv("N26_USERNAME"),
		n26Password:     os.Getenv("N26_PASSWORD"),
		n26DeviceToken:  os.Getenv("N26_DEVICE_TOKEN"),
		lunchMoneyToken: os.Getenv("LUNCHMONEY_TOKEN"),
	}
	if os.Getenv("API_PORT") != "" {
		intAPIPort, err := strconv.Atoi(os.Getenv("API_PORT"))
		if err != nil {
			panic(err)
		}
		appConfig.port = intAPIPort
	}
	return appConfig
}

func main() {
	config = initConfig()

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
