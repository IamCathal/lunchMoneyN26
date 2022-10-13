package dtos

import "time"

type AppConfig struct {
	WebServer bool
	Port      int

	N26Username     string
	N26Password     string
	N26DeviceToken  string
	LunchMoneyToken string
	APIPassword     string

	// Only to be used when the application
	// is running in offline mode
	Days int

	// Global variables
	ApplicationStartUpTime time.Time
}

// UptimeResponse is the standard response
// for any service's /status endpoint
type UptimeResponse struct {
	Status      string        `json:"status,omitempty"`
	Uptime      time.Duration `json:"uptime,omitempty"`
	StartUpTime int64         `json:"startuptime,omitempty"`
}

type UploadTransactionsDTO struct {
	Transactions      []FilteredTransaction `json:"transactions"`
	ApplyRules        bool                  `json:"apply_rules"`
	SkipDuplicates    bool                  `json:"skip_duplicates"`
	CheckForRecurring bool                  `json:"check_for_recurring"`
}

type FilteredTransaction struct {
	ID       string  `json:"id"`
	Date     string  `json:"date"`
	Payee    string  `json:"payee"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Category string  `json:"category"`
}

type LunchMoneyInsertTransactionResponse struct {
	IDs []int `json:"ids"`
}

type WebsocketMessage struct {
	Msg          string             `json:"msg"`
	Finished     bool               `json:"finished"`
	SummaryStats WsTransactionStats `json:"summarystats,omitempty"`
}

type WsTransactionStats struct {
	N26FoundTransactions           int       `json:"n26FoundTransactions"`
	LunchMoneyInsertedTranscations int       `json:"lunchMoneyInseredTransactions"`
	DaysLookedUp                   int       `json:"daysLookedUp"`
	CurrTime                       time.Time `json:"currTime"`
}
