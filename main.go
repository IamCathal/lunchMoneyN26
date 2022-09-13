package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/guitmz/n26"
	"github.com/joho/godotenv"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	config appConfig
)

func getClientWithProgressOutput() *n26.Client {
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
		UserName:    config.n26Username,
		Password:    config.n26Password,
		DeviceToken: config.n26DeviceToken,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("\nYou've successfully authenticated")
	authenticatedInApp = true

	return newClient
}

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
	transactionIDs := lunchMoneyInsertTransactionResponse{}
	err = json.Unmarshal(body, &transactionIDs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inserted %d new transactions into LunchMoney\n", len(transactionIDs.IDs))
}

func initConfig(args []string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ensureAllEnvVarsAreSet()

	appConfig := appConfig{
		webServer:       false,
		port:            2944,
		n26Username:     os.Getenv("N26_USERNAME"),
		n26Password:     os.Getenv("N26_PASSWORD"),
		n26DeviceToken:  os.Getenv("N26_DEVICE_TOKEN"),
		lunchMoneyToken: os.Getenv("LUNCHMONEY_TOKEN"),

		days: 0,
	}
	if os.Getenv("API_PORT") != "" {
		intAPIPort, err := strconv.Atoi(os.Getenv("API_PORT"))
		if err != nil {
			panic(err)
		}
		appConfig.port = intAPIPort
	}

	webServer := flag.Bool("webserver", false, "Run the application as a webserver")
	days := flag.Int("d", 1, "Search for the last n days for any transactions (if not in webserver mode)")
	flag.Parse()

	if *webServer {
		appConfig.webServer = true
	} else {
		appConfig.days = *days
	}
	config = appConfig
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
		uploadTransactions(uploadTransactionsDTO{transactions, true, true, true})
		return
	}
	fmt.Printf("No transactions found within the last %d days\n", config.days)
}
