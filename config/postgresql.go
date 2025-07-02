package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func ConnectDB() (*pgx.Conn, context.Context, context.CancelFunc) {
	logger := GetLog()
	var username string = os.Getenv("DB_USER")
	var password string = os.Getenv("DB_PASS")
	var host string = os.Getenv("DB_HOST")
	var database string = os.Getenv("DB_NAME")
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s", username, password, host, database)
	logger.Debug("Connection String : " + connStr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		cancel()
		logger.DPanic("Unable to connect to database : " + err.Error())
	}

	return conn, ctx, cancel
}
