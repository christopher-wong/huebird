package main

import (
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func setupNatsServer() (*server.Server, error) {
	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      4222,
		JetStream: true,              // Enable JetStream
		StoreDir:  "/tmp/nats-store", // Required for JetStream
	}

	// Create the server with the options
	ns, err := server.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS server: %v", err)
	}

	// Start the server
	go ns.Start()

	// Wait for server to be ready
	if !ns.ReadyForConnections(4 * time.Second) {
		return nil, fmt.Errorf("nats server failed to start")
	}

	// Wait for JetStream to be ready
	timeout := time.Now().Add(4 * time.Second)
	for time.Now().Before(timeout) {
		if ns.JetStreamEnabled() {
			return ns, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return nil, fmt.Errorf("jetstream failed to start")
}

func connectToNats() (*nats.Conn, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}
	return nc, nil
}

func setupKVStore(nc *nats.Conn) (nats.KeyValue, error) {
	// Create or get the KV store
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get jetstream context: %v", err)
	}

	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket:  kvsName,
		History: 10, // Keep last 10 values
	})
	if err != nil {
		// If it already exists, try to get it
		kv, err = js.KeyValue(kvsName)
		if err != nil {
			return nil, fmt.Errorf("failed to create/get KV store: %v", err)
		}
	}
	return kv, nil
}
