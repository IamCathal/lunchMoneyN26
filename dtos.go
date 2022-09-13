package main

type appConfig struct {
	webServer bool
	port      int

	n26Username     string
	n26Password     string
	n26DeviceToken  string
	lunchMoneyToken string

	// Only to be used when the application
	// is running in offline mode
	days int
}

// UptimeResponse is the standard response
// for any service's /status endpoint
type uptimeResponse struct {
	Status string `json:"status"`
}

type uploadTransactionsDTO struct {
	Transactions      []filteredTransaction `json:"transactions"`
	ApplyRules        bool                  `json:"apply_rules"`
	SkipDuplicates    bool                  `json:"skip_duplicates"`
	CheckForRecurring bool                  `json:"check_for_recurring"`
}

type filteredTransaction struct {
	ID       string  `json:"id"`
	Date     string  `json:"date"`
	Payee    string  `json:"payee"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Category string  `json:"category"`
}

type lunchMoneyInsertTransactionResponse struct {
	IDs []int `json:"ids"`
}
