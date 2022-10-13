package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/guitmz/n26"
)

func isAuthRequiredEndpoint(urlPath string) bool {
	authRequiredEndpoint := make(map[string]bool)
	authRequiredEndpoint["/status"] = true
	authRequiredEndpoint["/transactions"] = true
	authRequiredEndpoint["/ws/transactions"] = true

	requiresAuth := authRequiredEndpoint[urlPath]
	return requiresAuth
}

func verifyPassword(password string) bool {
	return password == config.APIPassword
}

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
	return newClient
}

func setupWebSocket(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// if _, ok := err.(websocket.HandshakeError); !ok {
		// 	return nil
		// }
		return nil
	}
	return ws
}

func wsMsg(msg string, ws *websocket.Conn) {
	newWsMsg := websocketMessage{
		Msg:      msg,
		Finished: false,
	}
	jsonStr, err := json.Marshal(newWsMsg)
	if err != nil {
		panic(err)
	}
	err = ws.WriteMessage(1, jsonStr)
	if err != nil {
		panic(err)
	}
}

func wsFinish(ws *websocket.Conn, summaryStats *wsTransactionStats) {
	newWsMsg := websocketMessage{
		Msg:          "transaction finished",
		Finished:     true,
		SummaryStats: *summaryStats,
	}
	jsonStr, err := json.Marshal(newWsMsg)
	if err != nil {
		panic(err)
	}
	err = ws.WriteMessage(1, jsonStr)
	if err != nil {
		panic(err)
	}
	ws.Close()
}

func SendBasicInvalidResponse(w http.ResponseWriter, req *http.Request, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	response := struct {
		Error string `json:"error"`
	}{
		msg,
	}
	json.NewEncoder(w).Encode(response)
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetupCORS(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		if isAuthRequiredEndpoint(r.URL.Path) {
			if !verifyPassword(r.Header.Get("API_KEY")) {
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

func ensureAllEnvVarsAreSet() {
	requiredEnvVars := []string{
		"N26_USERNAME",
		"N26_PASSWORD",
		"N26_DEVICE_TOKEN",
		"LUNCHMONEY_TOKEN",
		"API_KEY",
	}

	for _, envVar := range requiredEnvVars {
		if isSet := envVarIsSet(envVar); !isSet {
			panic(fmt.Sprint("env var " + envVar + " is not set"))
		}
	}
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

func SetupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, API_KEY")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
}
