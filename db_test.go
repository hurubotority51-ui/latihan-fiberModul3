package main

import (
	"context"
	"testing"
	"time"
)

func TestConnectDB(t *testing.T) {
	conn, err := connectDB()
	if err != nil {
		t.Fatalf("gagal koneksi database: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		t.Fatalf("database tidak merespons: %v", err)
	}
}
