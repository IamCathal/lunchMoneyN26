package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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

	writeMessageToWs("Waiting for N26 2FA authorization", ws)
	client := getClient()
	writeMessageToWs("N26 has been authorized", ws)

	writeMessageToWs("Retrieving transactions from N26", ws)
	transactions := getAndFilterTransactions(client, daysToLookup)

	writeMessageToWs("Uploading transactions to LunchMoney", ws)
	uploadTransactions(uploadTransactionsDTO{transactions, true, true, true})
	writeMessageToWs("Transactions uploaded", ws)

	ws.Close()
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
