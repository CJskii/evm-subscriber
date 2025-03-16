# Ethereum Pending Transactions Subscriber

A simple Dockerized Go application that connects to an Ethereum node over WebSockets with JWT authentication to subscribe to pending transactions (or any other JSON-RPC methods).

## 📂 Project Structure

```
evm-subscriber/
│── sub/          # EVM subscriber container
│── nethermind/   # EVM node setup
│── .env
```

## ✅ Prerequisites

- **Docker** and **Docker Compose** installed on your machine.

## ⚙️ Setup

### 1️⃣ Clone this repository:

```bash
git clone https://github.com/youruser/evm-subscriber.git
cd evm-subscriber
```

### 2️⃣ Navigate to the `sub/` directory and create a `.env` file:

```bash
cd sub
touch .env
```

Inside `.env`, add:

```ini
SECRET_KEY=<hex-encoded-secret>
ETH_HOST=<local-address-of-evm-node>
```

### 3️⃣ (Optional) Generate `go.sum` if it’s missing or out of date:

```bash
docker run --rm -v "$(pwd)":/app -w /app golang:1.24 go mod tidy
```

## 🚀 Usage

Run your own node using Docker Compose from the root `nethermind` directory:

```bash
docker compose up --build
```

Run the subscribe application using Docker Compose from the root `sub` directory:

```bash
docker compose up --build
```

### This will:

✅ Read the JWT secret from `SECRET_KEY`.  
✅ Connect to the Ethereum WebSocket host from `ETH_HOST`.  
✅ Generate a JWT token and authenticate over WebSocket.  
✅ Send a JSON-RPC request (`eth_subscribe` by default).  
✅ Print out any messages received.

## 🏗️ Notes on Project Structure

- **`sub/`**: Holds the Ethereum subscriber container.
- **`nethermind/`**: Designed to build and run an Ethereum node (Nethermind) locally.
- **`.env`**: Keep secrets in this file, never commit it.
- **Docker Compose**: Reads `.env` and injects variables at runtime.
- **No local Go installation needed**: Everything runs inside a one-off Docker container.

## 📜 License

MIT
