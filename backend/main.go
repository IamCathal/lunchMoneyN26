package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/lunchMoneyN26/config"
	"github.com/iamcathal/lunchMoneyN26/dtos"
	"github.com/iamcathal/lunchMoneyN26/endpoints"
	"github.com/iamcathal/lunchMoneyN26/util"
	"github.com/joho/godotenv"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func runWebServer(appConfig dtos.AppConfig) {
	r := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + fmt.Sprint(appConfig.Port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	fmt.Printf("Serving requests on :%d\n", appConfig.Port)
	log.Fatal(srv.ListenAndServe())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	appConfig := config.InitConfig(os.Args[1:])
	endpoints.SetConfig(appConfig)
	util.SetConfig(appConfig)

	if appConfig.WebServer {
		runWebServer(appConfig)
		return
	}

	client := util.GetClientWithProgressOutput()
	transactions := util.GetAndFilterTransactions(client, appConfig.Days)
	if len(transactions) > 0 {
		fmt.Printf("Found %d transactions from N26 in the last %d days:\n", appConfig.Days, len(transactions))
		for _, transaction := range transactions {
			fmt.Printf("\t%s\t%s\n", transaction.Date, transaction.Payee)
		}
		transactions := util.UploadTransactions(util.DefaultUploadTransactionsDTO(transactions))
		fmt.Printf("Inserted %d new transactions into LunchMoney\n", len(transactions.IDs))
		return
	}
	fmt.Printf("No transactions found within the last %d days\n", appConfig.Days)
}
