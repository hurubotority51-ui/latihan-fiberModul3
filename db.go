package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn
var studentRepo StudentRepository

func connectDB() (*pgx.Conn, error) {
	connString := os.Getenv("DATABASE_URL")

	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL belum diatur")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return nil, err
	}

	fmt.Println("Database connected successfully")

	return conn, nil
}
