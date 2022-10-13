package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/guitmz/n26"
	"github.com/iamcathal/lunchMoneyN26/dtos"
)

var (
	appConfig dtos.AppConfig
)

func SetConfig(conf dtos.AppConfig) {
	appConfig = conf
}

func UploadTransactions(transactions dtos.UploadTransactionsDTO) dtos.LunchMoneyInsertTransactionResponse {
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
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	transactionIDs := dtos.LunchMoneyInsertTransactionResponse{}
	err = json.Unmarshal(body, &transactionIDs)
	if err != nil {
		panic(err)
	}
	return transactionIDs
}

func GetAndFilterTransactions(client *n26.Client, daysToLookup int) []dtos.FilteredTransaction {
	endTime := n26.TimeStamp{Time: time.Now()}
	startTime := n26.TimeStamp{Time: endTime.Time.Add((-time.Hour * 24) * time.Duration(daysToLookup))}

	transactions, err := client.GetTransactions(startTime, endTime, fmt.Sprint(daysToLookup))
	if err != nil {
		panic(err)
	}

	filteredTransactions := []dtos.FilteredTransaction{}
	for _, transaction := range *transactions {
		currTransaction := dtos.FilteredTransaction{
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
			// or from a business/*  */
			currTransaction.Payee = transaction.MerchantName
			currTransaction.Category = transaction.Category
		}
		filteredTransactions = append(filteredTransactions, currTransaction)
	}
	return filteredTransactions
}

func GetClientWithProgressOutput() *n26.Client {
	authenticatedInApp := false
	waitTimeRemaining := 300

	go func() {
		for {
			if authenticatedInApp {
				break
			}
			if waitTimeRemaining == 0 {
				fmt.Println("Maximum allowed wait time of 5m exceeded")
				os.Exit(1)
			}

			fmt.Printf("\r2FA Confirmation required in your N26 app within the next: %v", getMinutesAndSecondsLeft(waitTimeRemaining))
			time.Sleep(1 * time.Second)
			waitTimeRemaining -= 1
		}
	}()

	newClient, err := n26.NewClient(n26.Auth{
		UserName:    appConfig.N26Username,
		Password:    appConfig.N26Password,
		DeviceToken: appConfig.N26DeviceToken,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("\nYou've successfully authenticated")
	authenticatedInApp = true

	return newClient
}

func GetClient() *n26.Client {
	fmt.Println("waiting for 2FA confirmation in app")
	newClient, err := n26.NewClient(n26.Auth{
		UserName:    appConfig.N26Username,
		Password:    appConfig.N26Password,
		DeviceToken: appConfig.N26DeviceToken,
	})
	if err != nil {
		panic(err)
	}
	return newClient
}

func DefaultUploadTransactionsDTO(transactions []dtos.FilteredTransaction) dtos.UploadTransactionsDTO {
	return dtos.UploadTransactionsDTO{
		Transactions:      transactions,
		ApplyRules:        true,
		SkipDuplicates:    true,
		CheckForRecurring: true,
	}
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
