package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

func wsTransactions(w http.ResponseWriter, r *http.Request) {
	ws := setupWebSocket(w, r)
	if ws == nil {
		SendBasicInvalidResponse(w, r, "unable to upgrade websocket", http.StatusBadRequest)
		return
	}
	fullDaysToLookupString := r.URL.Query().Get("days")
	daysToLookup, err := strconv.Atoi(fullDaysToLookupString)
	if err != nil {
		panic(err)
	}
	wsSummaryStats := wsTransactionStats{
		DaysLookedUp: daysToLookup,
		CurrTime:     time.Now(),
	}
	defer wsFinish(ws, &wsSummaryStats)

	wsMsg("Waiting for N26 2FA authorization", ws)
	client := getClient()
	wsMsg("N26 has been authorized", ws)

	wsMsg("Retrieving transactions from N26", ws)
	transactions := getAndFilterTransactions(client, daysToLookup)
	wsMsg("Retrieved transactions from N26", ws)
	wsSummaryStats.N26FoundTransactions = len(transactions)

	wsMsg("Uploading transactions to LunchMoney", ws)
	newTransactions := uploadTransactions(uploadTransactionsDTO{transactions, true, true, true})
	wsMsg("transactions inserted into LunchMoney", ws)
	wsSummaryStats.LunchMoneyInsertedTranscations = len(newTransactions.IDs)
}

func Transactions(w http.ResponseWriter, r *http.Request) {
	fullDaysToLookupString := r.URL.Query().Get("days")
	daysToLookup, err := strconv.Atoi(fullDaysToLookupString)
	if err != nil {
		panic(err)
	}

	client := getClient()
	transactions := getAndFilterTransactions(client, daysToLookup)
	uploadTransactions(uploadTransactionsDTO{transactions, true, true, true})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

func Status(w http.ResponseWriter, r *http.Request) {
	req := uptimeResponse{
		Status:      "operational",
		Uptime:      time.Duration(time.Since(ApplicationStartUpTime).Milliseconds()),
		StartUpTime: ApplicationStartUpTime.Unix(),
	}
	jsonObj, err := json.MarshalIndent(req, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(string(jsonObj))
}

func Online(w http.ResponseWriter, r *http.Request) {
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
