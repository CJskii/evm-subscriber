package main

import (
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
	if secretKey == "" || host == "" {
		log.Fatal("SECRET_KEY or ETH_HOST not set in environment.")
	}

	tokenString, err := decodeHexToBytes(secretKey)
	if err != nil {
		log.Fatalf("Failed to decode secret key: %v", err)
	}

	conn, err := subscribeToWebsocket(string(tokenString), host)
	if err != nil {
		log.Fatalf("Failed to subscribe to websocket: %v", err)
	}

	subscribeToNewPendingTransactions(conn)
	listenToWebsocket(conn)
}

func subscribeToWebsocket(tokenString, host string) (*websocket.Conn, error) {
	log.Printf("[ 🔌 ] Using Ethereum WebSocket host: %s", host)
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
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