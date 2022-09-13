package main

import (
	"encoding/json"
	"fmt"
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
	writeMessageToWs(fmt.Sprintf("Retrieved %d transactions from the last %d days from N26", len(transactions), daysToLookup), ws)

	writeMessageToWs(fmt.Sprintf("Uploading %d transactions to LunchMoney", len(transactions)), ws)
	newTransactions := uploadTransactions(uploadTransactionsDTO{transactions, true, true, true})
	writeMessageToWs(fmt.Sprintf("%d unique transactions were created in LunchMoney", len(newTransactions.IDs)), ws)

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
