package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/guitmz/n26"
)

var (
	ApplicationStartUpTime time.Time
	upgrader               = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	config appConfig
)

func getAndFilterTransactions(client *n26.Client, daysToLookup int) []filteredTransaction {
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
	return filteredTransactions
}

func uploadTransactions(transactions uploadTransactionsDTO) lunchMoneyInsertTransactionResponse {
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
	transactionIDs := lunchMoneyInsertTransactionResponse{}
	err = json.Unmarshal(body, &transactionIDs)
	if err != nil {
		panic(err)
	}
	return transactionIDs
}

func runWebServer() {
	r := mux.NewRouter()
	r.HandleFunc("/status", Status).Methods("POST")
	r.HandleFunc("/transactions", Transactions).Methods("POST")
	r.HandleFunc("/ws/transactions", wsTransactions).Methods("GET")
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
	initConfig(os.Args[1:])

	if config.webServer {
		runWebServer()
		return
	}

	client := getClientWithProgressOutput()
	transactions := getAndFilterTransactions(client, config.days)
	if len(transactions) > 0 {
		fmt.Printf("Found %d transactions from N26 in the last %d days:\n", config.days, len(transactions))
		for _, transaction := range transactions {
			fmt.Printf("\t%s\t%s\n", transaction.Date, transaction.Payee)
		}
		transactions := uploadTransactions(uploadTransactionsDTO{transactions, true, true, true})
		fmt.Printf("Inserted %d new transactions into LunchMoney\n", len(transactions.IDs))
		return
	}
	fmt.Printf("No transactions found within the last %d days\n", config.days)
}
