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

// UptimeResponse is the standard response
// for any service's /status endpoint
type UptimeResponse struct {
	Status string `json:"status"`
}

type uploadTransactionsDTO struct {
	Transactions      []filteredTransaction `json:"transactions"`
	ApplyRules        bool                  `json:"apply_rules"`
	SkipDuplicates    bool                  `json:"skip_duplicates"`
	CheckForRecurring bool                  `json:"check_for_recurring"`
}

type filteredTransaction struct {
	ID       string  `json:"id"`
	Date     string  `json:"date"`
	Payee    string  `json:"payee"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Category string  `json:"category"`
}

func getClient() *n26.Client {
	fmt.Println("waiting for 2FA confirmation in app")
	newClient, err := n26.NewClient(n26.Auth{
		UserName:    os.Getenv("USERNAME"),
		Password:    os.Getenv("PASSWORD"),
		DeviceToken: os.Getenv("DEVICE_TOKEN"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("auth complete")
	fmt.Println("auth woked")
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
	fmt.Fprint(w, string(jsonString))
}

func Status(w http.ResponseWriter, r *http.Request) {
	req := UptimeResponse{
		Status: "operational",
	}
	jsonObj, err := json.MarshalIndent(req, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%v %+v\n", time.Now().Format(time.RFC3339), r)
		next.ServeHTTP(w, r)
	})
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
	req.Close = true
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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := "2944"

	r := mux.NewRouter()
	r.HandleFunc("/status", Status).Methods("POST")
	r.HandleFunc("/anycans", Transactions).Methods("POST")
	r.Use(logMiddleware)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + fmt.Sprint(2944),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	fmt.Println("serving requests on :" + port)

	log.Fatal(srv.ListenAndServe())

}
