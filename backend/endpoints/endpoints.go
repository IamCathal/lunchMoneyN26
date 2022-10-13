package endpoints

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/iamcathal/lunchMoneyN26/dtos"
	"github.com/iamcathal/lunchMoneyN26/util"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/status", status).Methods("POST", "OPTIONS")
	r.HandleFunc("/transactions", transactions).Methods("POST")
	r.HandleFunc("/ws/transactions", wsTransactions).Methods("GET")
	r.Use(logMiddleware)
	return r
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
	wsSummaryStats := dtos.WsTransactionStats{
		DaysLookedUp: daysToLookup,
		CurrTime:     time.Now(),
	}
	defer wsFinish(ws, &wsSummaryStats)

	wsMsg("Waiting for N26 2FA authorization", ws)
	client := util.GetClient()
	wsMsg("N26 has been authorized", ws)

	wsMsg("Retrieving transactions from N26", ws)
	transactions := util.GetAndFilterTransactions(client, daysToLookup)
	wsMsg("Retrieved transactions from N26", ws)
	wsSummaryStats.N26FoundTransactions = len(transactions)

	wsMsg("Uploading transactions to LunchMoney", ws)
	newTransactions := util.UploadTransactions(util.DefaultUploadTransactionsDTO(transactions))
	wsMsg("transactions inserted into LunchMoney", ws)
	wsSummaryStats.LunchMoneyInsertedTranscations = len(newTransactions.IDs)
}

func transactions(w http.ResponseWriter, r *http.Request) {
	fullDaysToLookupString := r.URL.Query().Get("days")
	daysToLookup, err := strconv.Atoi(fullDaysToLookupString)
	if err != nil {
		panic(err)
	}

	client := util.GetClient()
	transactions := util.GetAndFilterTransactions(client, daysToLookup)
	util.UploadTransactions(util.DefaultUploadTransactionsDTO(transactions))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

func status(w http.ResponseWriter, r *http.Request) {
	req := dtos.UptimeResponse{
		Status:      "operational",
		Uptime:      time.Duration(time.Since(appConfig.ApplicationStartUpTime).Milliseconds()),
		StartUpTime: appConfig.ApplicationStartUpTime.Unix(),
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
		setupCORS(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		if isAuthRequiredEndpoint(r.URL.Path) {
			if !verifyPassword(r) {
				fmt.Printf("ip: %s with user-agent: '%s' wasn't authorized to access %s. Attempted to use API_KEY: '%s'\n",
					r.RemoteAddr, r.Header.Get("User-Agent"), r.URL.Path, r.Header.Get("API_KEY"))
				w.WriteHeader(http.StatusForbidden)
				response := struct {
					Error string `json:"error"`
				}{
					"You are not authorized to access this endpoint",
				}
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		fmt.Printf("%v %+v\n", time.Now().Format(time.RFC3339), r)
		next.ServeHTTP(w, r)
	})
}
