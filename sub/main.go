package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
)

type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method,omitempty"`
	Params  struct {
		Subscription string      `json:"subscription"`
		Result       interface{} `json:"result"`
	} `json:"params"`
}

func main() {
	log.Println("===== Starting Ethereum Pending Transactions Subscriber =====")

	secretKey := os.Getenv("SECRET_KEY")
	host := os.Getenv("ETH_HOST")
	wsPort, rpcPort := os.Getenv("WS_PORT"), os.Getenv("RPC_PORT")

	if secretKey == "" || host == "" {
		log.Fatal("SECRET_KEY or ETH_HOST not set in environment.")
	}

	tokenString, err := decodeHexToBytes(secretKey)
	if err != nil {
		log.Fatalf("Failed to decode secret key: %v", err)
	}

	conn, err := subscribeToWebsocket(string(tokenString), host, wsPort)
	if err != nil {
		log.Fatalf("Failed to subscribe to websocket: %v", err)
	}

	checkSyncStatus(secretKey, host, rpcPort)
	subscribeToNewPendingTransactions(conn)
	listenToWebsocket(conn)
}
func checkSyncStatus(secretKey, host, rpcPort string) {
	request := RPCRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "eth_syncing",
		Params:  []interface{}{},
	}
	log.Printf("[ 👉 ] Sending request to check sync status: %+v", request)

	// Send the JSON-RPC request.
	response, err := sendRPCRequest(request, secretKey, host, rpcPort)
	if err != nil {
		log.Fatalf("Failed to check sync status: %v", err)
	}

	// Log the response.
	log.Printf("[ 👈 ] Received sync status: %+v", response)
}

func sendRPCRequest(request RPCRequest, secretKey, host string, rpcPort string) (RPCResponse, error) {
	rpcUrl := "http://" + host + ":" + rpcPort + "/"
	log.Printf("%s", "[ 🚀 ] RPC host: " + rpcUrl)
	client := &http.Client{}
	reqBody, err := json.Marshal(request)
	if err != nil {
		return RPCResponse{}, err
	}

	req, err := http.NewRequest("POST", rpcUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return RPCResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secretKey)

	resp, err := client.Do(req)
	if err != nil {
		return RPCResponse{}, err
	}
	defer resp.Body.Close()

	var rpcResponse RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return RPCResponse{}, err
	}

	return rpcResponse, nil
}

func subscribeToWebsocket(tokenString, host string, wsPort string) (*websocket.Conn, error) {
	websocketHost := host + ":" + wsPort
	u := url.URL{Scheme: "ws", Host: websocketHost, Path: "/"}
	log.Printf("[ 🔌 ] Connecting to WebSocket URL: %s", u.String())

	header := http.Header{}
	header.Add("Authorization", "Bearer "+tokenString)

	dialer := websocket.Dialer{
		Subprotocols: []string{"jsonrpc"},
	}

	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}
	log.Println("[ ✅ ] Connected to Ethereum WebSocket.")

	return conn, nil
}

func subscribeToNewPendingTransactions(conn *websocket.Conn) {
	subscribeRequest := RPCRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "eth_subscribe",
		Params:  []interface{}{"newPendingTransactions"},
	}
	log.Printf("[ 👉 ] Sending subscription request: %+v", subscribeRequest)

	if err := conn.WriteJSON(subscribeRequest); err != nil {
		log.Fatalf("Error sending subscribe request: %v", err)
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("Error reading subscription response: %v", err)
	}
	log.Printf("[ 👈 ] Received subscription response: %s", message)
}

func listenToWebsocket(conn *websocket.Conn) {
	log.Println("[ 🔔 ] Listening to websocket...")
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Raw message received: %s", msg)

		// Try to decode the JSON-RPC message.
		var response RPCResponse
		if err := json.Unmarshal(msg, &response); err != nil {
			log.Printf("Error decoding message: %v", err)
			continue
		}
		log.Printf("Decoded message: %+v", response)

		// Log the pending transaction (usually a transaction hash).
		log.Printf("[ ℹ️ ] New Pending Transaction: %v", response.Params.Result)
	}
}

func decodeHexToBytes(secretKey string) ([]byte, error) {
	// Decode the hex string into raw bytes.
	secretBytes, err := hex.DecodeString(secretKey)
	if err != nil {
		return nil, err
	}

	// Generate JWT token using HS256.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(10 * time.Minute).Unix(),
	})
	tokenString, err := token.SignedString(secretBytes)
	if err != nil {
		return nil, err
	}

	return []byte(tokenString), nil
}