package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Set default environment variables for supplier peer
	if os.Getenv("PEER_NODE_TYPE") == "" {
		os.Setenv("PEER_NODE_TYPE", "supplier")
	}
	if os.Getenv("PEER_NODE_ID") == "" {
		os.Setenv("PEER_NODE_ID", "supplier-peer-1")
	}
	if os.Getenv("PEER_PORT") == "" {
		os.Setenv("PEER_PORT", "8083")
	}

	port := os.Getenv("PEER_PORT")
	log.Printf("Supplier Peer placeholder running on port %s", port)

	// Simple health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "peer": "supplier", "message": "Peer is running (placeholder)"}`)
	})

	// CORS middleware
	handler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "*")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	log.Printf("Supplier Peer listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler(http.DefaultServeMux)))
}
