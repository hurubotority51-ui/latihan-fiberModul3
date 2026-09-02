package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool
var studentRepo StudentRepository

func connectDB() (*pgxpool.Pool, error) {
	connString := os.Getenv("DATABASE_URL")

	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL belum diatur")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	fmt.Println("Database connected successfully")

	return pool, nil
}
