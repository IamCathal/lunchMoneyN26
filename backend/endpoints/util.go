package endpoints

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/lunchMoneyN26/dtos"
)

var (
	appConfig dtos.AppConfig
)

func SetConfig(conf dtos.AppConfig) {
	appConfig = conf
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
	newWsMsg := dtos.WebsocketMessage{
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

func wsFinish(ws *websocket.Conn, summaryStats *dtos.WsTransactionStats) {
	newWsMsg := dtos.WebsocketMessage{
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

func isAuthRequiredEndpoint(urlPath string) bool {
	authRequiredEndpoint := make(map[string]bool)
	authRequiredEndpoint["/status"] = true
	authRequiredEndpoint["/transactions"] = true
	authRequiredEndpoint["/ws/transactions"] = true

	requiresAuth := authRequiredEndpoint[urlPath]
	return requiresAuth
}

func verifyPassword(r *http.Request) bool {
	if isWsRequest := strings.HasPrefix(r.URL.Path, "/ws/"); isWsRequest {
		apiKey := r.URL.Query().Get("apikey")
		return apiKey == appConfig.APIPassword
	}
	return r.Header.Get("API_KEY") == appConfig.APIPassword
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

func setupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, API_KEY")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
}
