package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/guitmz/n26"
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

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%v %+v\n", time.Now().Format(time.RFC3339), r)
		next.ServeHTTP(w, r)
	})
}

func ensureAllEnvVarsAreSet() {
	requiredEnvVars := []string{
		"N26_USERNAME",
		"N26_PASSWORD",
		"N26_DEVICE_TOKEN",
		"LUNCHMONEY_TOKEN",
	}

	for _, envVar := range requiredEnvVars {
		if isSet := envVarIsSet(envVar); !isSet {
			panic(fmt.Sprint("env var " + envVar + " is not set"))
		}
	}
}

func getEnvVarOrDefault(varName, defaultVal string) string {
	if isSet := envVarIsSet(varName); isSet {
		return os.Getenv(varName)
	}
	return defaultVal
}

func envVarIsSet(varName string) bool {
	if _, exists := os.LookupEnv(varName); exists {
		return true
	}
	return false
}

func getMinutesAndSecondsLeft(totalSeconds int) string {
	totalMinutesLeft, secondsLeftRemainder := divmod(totalSeconds, 60)

	if totalMinutesLeft != 0 {
		return fmt.Sprintf("%dm %ds", totalMinutesLeft, secondsLeftRemainder)
	} else {
		return fmt.Sprintf("%ds", secondsLeftRemainder)
	}
}

func divmod(big, little int) (int, int) {
	quotient := big / little
	remainder := big % little
	return quotient, remainder
}
