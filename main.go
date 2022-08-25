package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/guitmz/n26"
	"github.com/joho/godotenv"
)

type filteredTransaction struct {
	ID           string    `json:"id"`
	VisibleTS    time.Time `json:"visibleTS"`
	Payee        string    `json:"payee"`
	Amount       float64   `json:"amount"`
	CurrencyCode string    `json:"currencyCode"`
	Category     string    `json:"category"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	fullDaysToLookup := 0

	cliArgs := os.Args[1:]
	if len(cliArgs) == 1 {
		days, err := strconv.ParseInt(cliArgs[0], 10, 32)
		if err != nil {
			panic("duhhh")
		}
		fullDaysToLookup = int(days)
	}

	endTime := n26.TimeStamp{Time: time.Now()}
	startTime := n26.TimeStamp{Time: endTime.Time.Add((-time.Hour * 24) * time.Duration(fullDaysToLookup))}

	fmt.Println("waiting for 2FA confirmation in app")
	client, err := n26.NewClient(n26.Auth{
		UserName:    os.Getenv("USERNAME"),
		Password:    os.Getenv("PASSWORD"),
		DeviceToken: os.Getenv("DEVICE_TOKEN"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("auth complete")

	fmt.Println("auth woked")
	transactions, err := client.GetTransactions(startTime, endTime, "12")
	if err != nil {
		panic(err)
	}

	filteredTransactions := []filteredTransaction{}

	for _, transaction := range *transactions {
		currTransaction := filteredTransaction{
			ID:       transaction.ID,
			Date:     transaction.VisibleTS.Time.Format(time.RFC3339),
			Amount:   transaction.Amount,
			Currency: transaction.OriginalCurrency,
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

	fmt.Println(string(jsonString))
}
